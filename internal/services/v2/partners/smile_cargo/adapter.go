package smile_cargo

import (
	"context"
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// Adapter implements the PartnerAdapter interface for Smile Cargo
type Adapter struct {
	client *SmileCargoClient
	config config.SmileCargoConfig
}

// NewAdapter creates a new Smile Cargo adapter instance
func NewAdapter(config config.SmileCargoConfig) *Adapter {
	return &Adapter{
		client: NewSmileCargoClient(config),
		config: config,
	}
}

// GetPartnerCode returns the partner code (interface compatibility only)
// TODO: This should be removed when all code uses database values
func (a *Adapter) GetPartnerCode() string {
	return "smile_cargo"
}

// GetPartnerName returns the partner name (interface compatibility only)
// TODO: This should be removed when all code uses database values
func (a *Adapter) GetPartnerName() string {
	return "Smile Cargo"
}

// GetAdapterType returns the adapter type
func (a *Adapter) GetAdapterType() common.AdapterType {
	return common.AdapterTypeHTTP
}

// IsEnabled returns whether the adapter is enabled
func (a *Adapter) IsEnabled() bool {
	return a.config.Enabled
}

// SupportsRequest checks if Smile Cargo supports the given request
func (a *Adapter) SupportsRequest(ctx context.Context, request *models.ServiceabilityV2Request) bool {
	// Check if we have required postal codes
	if !a.hasValidPincodes(request) {
		return false
	}

	// Partner attribute mapping in database determines supported parcel categories
	// No hardcoded category filtering needed here
	return true
}

// CheckServiceability checks serviceability for the request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request) (*common.PartnerServiceabilityResult, error) {
	if !a.SupportsRequest(ctx, request) {
		return &common.PartnerServiceabilityResult{
			PartnerCode:   a.GetPartnerCode(),
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			ResponseTime:  0,
			Metadata: map[string]interface{}{
				"reason": "Request not supported by Smile Cargo (not a cargo request or missing pincodes)",
			},
		}, nil
	}

	// Convert request to Smile Cargo format
	smileCargoRequest := a.convertToServiceAvailabilityRequest(request)

	// Make API call
	response, err := a.client.CheckServiceAvailability(ctx, smileCargoRequest)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerCode:   a.GetPartnerCode(),
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			Error:         err,
			ErrorMessage:  &[]string{fmt.Sprintf("Smile Cargo API call failed: %v", err)}[0],
		}, nil
	}

	// Convert response and return
	// The orchestrator will set PartnerCode and PartnerName from database
	return a.convertServiceAvailabilityResponse(response), nil
}

// convertToServiceAvailabilityRequest converts v2 request to Smile Cargo format
func (a *Adapter) convertToServiceAvailabilityRequest(request *models.ServiceabilityV2Request) ServiceAvailabilityRequest {
	req := ServiceAvailabilityRequest{
		// Remove vendorCode as per new business requirement
		// VendorCode: a.config.VendorCode,
	}

	// Set from pincode (source)
	if request.SourcePostalCode != nil {
		req.FromPincode = *request.SourcePostalCode
	}

	// Set to pincode (destination)
	if request.DestinationPostalCode != nil {
		req.ToPincode = *request.DestinationPostalCode
	} else if request.PostalCode != nil {
		req.ToPincode = *request.PostalCode
	}

	return req
}

// convertServiceAvailabilityResponse converts Smile Cargo response to common format
func (a *Adapter) convertServiceAvailabilityResponse(response *ServiceAvailabilityResponse) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerCode:  a.GetPartnerCode(),
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// CRITICAL BUSINESS RULE: SmilePartner must exist in BOTH "from" and "to" objects
	// for the service to be considered serviceable
	var fromSmilePartner *SmilePartner
	var toSmilePartner *SmilePartner
	var hasFromPartner = false
	var hasToPartner = false

	// Check all data entries and separate "from" and "to" entries
	for _, data := range response.Data {
		// Check if this is a "from" entry (has fromPincode)
		if data.FromPincode != nil && data.SmilePartner != nil && data.SmilePartner.Status {
			fromSmilePartner = data.SmilePartner
			hasFromPartner = true
		}

		// Check if this is a "to" entry (has toPincode)
		if data.ToPincode != nil && data.SmilePartner != nil && data.SmilePartner.Status {
			toSmilePartner = data.SmilePartner
			hasToPartner = true
		}
	}

	// Service is only available if BOTH from and to have valid SmilePartners
	if !hasFromPartner || !hasToPartner {
		result.IsServiceable = false
		reason := "SmilePartner missing in "
		if !hasFromPartner && !hasToPartner {
			reason += "both 'from' and 'to' objects"
		} else if !hasFromPartner {
			reason += "'from' object"
		} else {
			reason += "'to' object"
		}
		result.Metadata["reason"] = reason + " - service not available"
		result.Metadata["has_from_partner"] = hasFromPartner
		result.Metadata["has_to_partner"] = hasToPartner
		result.Metadata["bigship_partners_found"] = len(response.Data) > 0 && len(response.Data[0].BigShipPartners) > 0

		// Return empty capabilities structure for non-serviceable requests
		result.Capabilities = map[string]interface{}{
			"source_postal_code":      nil,
			"destination_postal_code": nil,
		}
		return result
	}

	// Both SmilePartners found and active
	result.IsServiceable = true

	// Create the new capabilities structure based on the final payload format
	capabilities := map[string]interface{}{
		"source_postal_code": map[string]interface{}{
			"status":    fromSmilePartner.Status,
			"lastMile":  fromSmilePartner.LastMile,
			"firstMile": fromSmilePartner.FirstMile,
			"cod":       fromSmilePartner.COD,
			"toPay":     fromSmilePartner.ToPay,
		},
		"destination_postal_code": map[string]interface{}{
			"status":    toSmilePartner.Status,
			"lastMile":  toSmilePartner.LastMile,
			"firstMile": toSmilePartner.FirstMile,
			"cod":       toSmilePartner.COD,
			"toPay":     toSmilePartner.ToPay,
		},
	}

	result.Capabilities = capabilities

	// Smile Cargo doesn't have individual services, only capabilities
	// Remove the services array as per user request
	result.Services = []models.ServiceV2{} // Empty services array

	// Set metadata
	result.Metadata["reason"] = "Smile Cargo service available in both from and to locations"
	result.Metadata["from_city"] = fromSmilePartner.CityName
	result.Metadata["to_city"] = toSmilePartner.CityName
	result.Metadata["area_availability"] = toSmilePartner.AreaAvailability

	return result
}

// IsCargoRequest determines if this is a cargo/freight request
// TODO: This method is deprecated - parcel categories should be determined by partner attribute mapping
func (a *Adapter) IsCargoRequest(request *models.ServiceabilityV2Request) bool {
	if request.ParcelCategory == nil {
		return false
	}

	category := strings.ToLower(*request.ParcelCategory)
	return category == "cargo" || category == "freight"
}

// hasValidPincodes checks if request has valid pincodes for Smile Cargo
func (a *Adapter) hasValidPincodes(request *models.ServiceabilityV2Request) bool {
	// Need source pincode
	if request.SourcePostalCode == nil || *request.SourcePostalCode == "" {
		return false
	}

	// Need destination pincode
	hasDestination := (request.DestinationPostalCode != nil && *request.DestinationPostalCode != "") ||
		(request.PostalCode != nil && *request.PostalCode != "")

	return hasDestination
}

// Initialize implements PartnerAdapter interface
func (a *Adapter) Initialize(ctx context.Context) error {
	// Validate required configuration
	if a.config.VendorCode == "" {
		return fmt.Errorf("Smile Cargo vendor code is required")
	}

	if a.config.BaseURL == "" {
		return fmt.Errorf("Smile Cargo base URL is required")
	}

	if a.config.ServiceURL == "" {
		return fmt.Errorf("Smile Cargo service URL is required")
	}

	return nil
}

// IsHealthy implements PartnerAdapter interface
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	if !a.config.Enabled {
		return false
	}

	// TODO: Add actual health check by calling Smile Cargo API
	return true
}

// GetMetrics implements PartnerAdapter interface
func (a *Adapter) GetMetrics() *common.PartnerMetrics {
	return &common.PartnerMetrics{
		PartnerCode:         "smile_cargo", // Use static value for metrics
		TotalRequests:       0,             // TODO: Implement actual metrics
		SuccessfulRequests:  0,
		FailedRequests:      0,
		AverageResponseTime: 0,
		HealthStatus:        "healthy",
		ErrorRate:           0.0,
	}
}

// Shutdown implements PartnerAdapter interface
func (a *Adapter) Shutdown(ctx context.Context) error {
	// No cleanup needed for HTTP client
	return nil
}
