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
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	if !a.SupportsRequest(ctx, request) {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: 0,
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
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Smile Cargo API call failed: %v", err)}[0],
		}, nil
	}

	// Convert response and return
	// The orchestrator will set PartnerCode and PartnerName from database
	return a.convertServiceAvailabilityResponse(response, partnerInfo), nil
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
func (a *Adapter) convertServiceAvailabilityResponse(response *ServiceAvailabilityResponse, partnerInfo common.PartnerInfo) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// NEW BUSINESS RULE: Check if activePartners length > 1 in BOTH from and to pincodes
	var fromPincodeData *ServiceAvailabilityData
	var toPincodeData *ServiceAvailabilityData
	var hasFromData = false
	var hasToData = false

	// Separate "from" and "to" entries from API response
	for _, data := range response.Data {
		if data.FromPincode != nil {
			fromPincodeData = &data
			hasFromData = true
		}
		if data.ToPincode != nil {
			toPincodeData = &data
			hasToData = true
		}
	}

	// Service is only available if BOTH from and to have data
	if !hasFromData || !hasToData {
		reason := "Smile Cargo not serviceable: missing data for "
		if !hasFromData && !hasToData {
			reason += "both 'from' and 'to' pincodes"
		} else if !hasFromData {
			reason += "'from' pincode"
		} else {
			reason += "'to' pincode"
		}
		result.ErrorMessage = &reason
		return result
	}

	// Check if both pincodes have multiple active partners (length > 1)
	fromPartnersCount := len(fromPincodeData.ActivePartners)
	toPartnersCount := len(toPincodeData.ActivePartners)

	if fromPartnersCount <= 1 || toPartnersCount <= 1 {
		reason := fmt.Sprintf("Smile Cargo not serviceable: insufficient active partners. From: %d, To: %d (need >1 for both)", fromPartnersCount, toPartnersCount)
		result.ErrorMessage = &reason
		return result
	}

	// SERVICEABLE: Both pincodes have multiple active partners
	// Build capabilities for all active partners
	allPartners := make([]map[string]interface{}, 0)
	
	// Add from pincode partners
	for _, partner := range fromPincodeData.ActivePartners {
		partnerCapability := map[string]interface{}{
			"pincode_type": "source",
			"partner_code": partner.PartnerCode,
			"is_active":    partner.IsActive,
			"city":         partner.CityName,
			"district":     partner.DistrictName,
			"zone":         partner.Zone,
			"first_mile":   partner.FirstMile,
			"last_mile":    partner.LastMile,
			"hub_code":     partner.HubCode,
			"cod":          partner.COD,
			"to_pay":       partner.ToPay,
			"surface":      partner.Surface,
			"air":          partner.Air,
			"rail":         partner.Rail,
			"vendor":       partner.Vendor,
		}
		allPartners = append(allPartners, partnerCapability)
	}

	// Add to pincode partners
	for _, partner := range toPincodeData.ActivePartners {
		partnerCapability := map[string]interface{}{
			"pincode_type": "destination",
			"partner_code": partner.PartnerCode,
			"is_active":    partner.IsActive,
			"city":         partner.CityName,
			"district":     partner.DistrictName,
			"zone":         partner.Zone,
			"first_mile":   partner.FirstMile,
			"last_mile":    partner.LastMile,
			"hub_code":     partner.HubCode,
			"cod":          partner.COD,
			"to_pay":       partner.ToPay,
			"surface":      partner.Surface,
			"air":          partner.Air,
			"rail":         partner.Rail,
			"vendor":       partner.Vendor,
		}
		allPartners = append(allPartners, partnerCapability)
	}

	result.Capabilities = map[string]interface{}{
		"active_partners": allPartners,
		"from_partners_count": fromPartnersCount,
		"to_partners_count":   toPartnersCount,
		"total_partners":      len(allPartners),
	}

	// Create individual services for each active partner
	services := make([]models.ServiceV2, 0)
	for _, partner := range allPartners {
		// Convert delivery modes to map[string]bool format
		deliveryModesMap := make(map[string]bool)
		deliveryModes := a.getDeliveryModes(partner)
		for _, mode := range deliveryModes {
			deliveryModesMap[mode] = true
		}
		
		service := models.ServiceV2{
			ServiceCode:   partner["partner_code"].(string),
			ServiceName:   fmt.Sprintf("Cargo Service - %s", partner["city"]),
			TATDays:       0, // Default TAT for cargo
			IsCOD:         partner["cod"].(bool),
			Pickup:        partner["first_mile"].(bool),
			Delivery:      partner["last_mile"].(bool),
			Insurance:     false, // Default for cargo
			ProductTypes:  map[string]bool{"cargo": true},
			DeliveryModes: deliveryModesMap,
		}
		services = append(services, service)
	}

	result.Services = services

	// Set metadata
	result.Metadata["reason"] = fmt.Sprintf("Smile Cargo serviceable: %d active partners at source, %d at destination", fromPartnersCount, toPartnersCount)
	result.Metadata["from_city"] = fromPincodeData.ActivePartners[0].CityName
	result.Metadata["to_city"] = toPincodeData.ActivePartners[0].CityName
	result.Metadata["serviceable"] = true

	return result
}

// getDeliveryModes extracts delivery modes from partner capabilities
func (a *Adapter) getDeliveryModes(partner map[string]interface{}) []string {
	modes := make([]string, 0)
	
	if surface, ok := partner["surface"].(bool); ok && surface {
		modes = append(modes, "SURFACE")
	}
	if air, ok := partner["air"].(bool); ok && air {
		modes = append(modes, "AIR")
	}
	if rail, ok := partner["rail"].(bool); ok && rail {
		modes = append(modes, "RAIL")
	}
	
	return modes
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
