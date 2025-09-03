package smile_cargo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// Adapter implements the PartnerAdapter interface for Smile Cargo
type Adapter struct {
	client *SmileCargoClient
	config config.SmileCargoConfig
	logger *logrus.Logger
}

// NewAdapter creates a new Smile Cargo adapter instance
func NewAdapter(config config.SmileCargoConfig) *Adapter {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel) // Consistent with other adapters
	
	// Log configuration
	logger.WithFields(logrus.Fields{
		"partner":      "Smile Cargo",
		"base_url":     config.BaseURL,
		"service_url":  config.ServiceURL,
		"enabled":      config.Enabled,
		"timeout":      config.Timeout,
		"max_retries":  config.MaxRetries,
		"retry_delay":  config.RetryDelay,
	}).Info("🚚 Creating Smile Cargo adapter with configuration")
	
	return &Adapter{
		client: NewSmileCargoClient(config),
		config: config,
		logger: logger,
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
	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "SupportsRequest",
		"source_pincode": request.SourcePostalCode,
		"destination_pincode": request.DestinationPostalCode,
		"postal_code": request.PostalCode,
		"parcel_category": request.ParcelCategory,
	}).Info("🎯 Smile Cargo SupportsRequest called - checking support")

	// Check if we have required postal codes
	if !a.hasValidPincodes(request) {
		a.logger.WithFields(logrus.Fields{
			"component": "smile_cargo_adapter",
			"method":    "SupportsRequest",
			"reason":    "Invalid pincodes",
		}).Info("Request not supported: invalid pincodes")
		return false
	}

	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "SupportsRequest",
		"supported": true,
	}).Info("Request supported by Smile Cargo")

	// Partner attribute mapping in database determines supported parcel categories
	// No hardcoded category filtering needed here
	return true
}

// CheckServiceability checks serviceability for the request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()
	
	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "CheckServiceability",
		"partner_id": partnerInfo.PartnerID,
		"partner_code": partnerInfo.PartnerCode,
		"source_pincode": request.SourcePostalCode,
		"destination_pincode": request.DestinationPostalCode,
		"postal_code": request.PostalCode,
	}).Info("🚚 Smile Cargo CheckServiceability STARTED - making API call")

	if !a.SupportsRequest(ctx, request) {
		a.logger.WithFields(logrus.Fields{
			"component": "smile_cargo_adapter",
			"method":    "CheckServiceability",
			"reason":    "Request not supported",
		}).Warn("Request not supported by Smile Cargo")
		
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
	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "CheckServiceability",
		"smile_cargo_request": smileCargoRequest,
	}).Info("Converted request to Smile Cargo format")

	// Make API call
	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "CheckServiceability",
		"api_url": a.config.ServiceURL,
		"full_url": a.config.BaseURL + a.config.ServiceURL,
	}).Info("🌐 Making Smile Cargo API call to external service")
	
	response, err := a.client.CheckServiceAvailability(ctx, smileCargoRequest)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"component": "smile_cargo_adapter",
			"method":    "CheckServiceability",
			"error": err.Error(),
		}).Error("Smile Cargo API call failed")
		
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Smile Cargo API call failed: %v", err)}[0],
		}, nil
	}

	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "CheckServiceability",
		"response_status": response.Status,
		"data_count": len(response.Data),
		"response_time_ms": time.Since(startTime).Milliseconds(),
	}).Info("✅ Smile Cargo API call SUCCESSFUL - processing response")

	// Convert response and return
	// The orchestrator will set PartnerCode and PartnerName from database
	result := a.convertServiceAvailabilityResponse(response, partnerInfo)
	
	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "CheckServiceability",
		"total_services": len(result.Services),
		"total_capabilities": len(result.Capabilities),
		"has_error": result.Error != nil,
		"total_time_ms": time.Since(startTime).Milliseconds(),
		"serviceable": len(result.Services) > 0,
	}).Info("🚚 Smile Cargo serviceability check COMPLETED")
	
	return result, nil
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

	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "convertToServiceAvailabilityRequest",
		"from_pincode": req.FromPincode,
		"to_pincode": req.ToPincode,
		"vendor_code": a.config.VendorCode,
	}).Info("Converted request to Smile Cargo format")

	return req
}

// convertServiceAvailabilityResponse converts Smile Cargo response to common format
func (a *Adapter) convertServiceAvailabilityResponse(response *ServiceAvailabilityResponse, partnerInfo common.PartnerInfo) *common.PartnerServiceabilityResult {
	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "convertServiceAvailabilityResponse",
		"response_status": response.Status,
		"data_count": len(response.Data),
		"partner_id": partnerInfo.PartnerID,
		"partner_code": partnerInfo.PartnerCode,
	}).Info("Starting response conversion")

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
			a.logger.WithFields(logrus.Fields{
				"component": "smile_cargo_adapter",
				"method":    "convertServiceAvailabilityResponse",
				"from_pincode": *data.FromPincode,
				"from_active_partners": len(data.ActivePartners),
			}).Info("Found from pincode data")
		}
		if data.ToPincode != nil {
			toPincodeData = &data
			hasToData = true
			a.logger.WithFields(logrus.Fields{
				"component": "smile_cargo_adapter",
				"method":    "convertServiceAvailabilityResponse",
				"to_pincode": *data.ToPincode,
				"to_active_partners": len(data.ActivePartners),
			}).Info("Found to pincode data")
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
		
		a.logger.WithFields(logrus.Fields{
			"component": "smile_cargo_adapter",
			"method":    "convertServiceAvailabilityResponse",
			"has_from_data": hasFromData,
			"has_to_data": hasToData,
			"reason": reason,
		}).Warn("Service not available: missing pincode data")
		
		result.ErrorMessage = &reason
		return result
	}

	// Check if both pincodes have multiple active partners (length > 1)
	fromPartnersCount := len(fromPincodeData.ActivePartners)
	toPartnersCount := len(toPincodeData.ActivePartners)

	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "convertServiceAvailabilityResponse",
		"from_partners_count": fromPartnersCount,
		"to_partners_count": toPartnersCount,
		"from_pincode": *fromPincodeData.FromPincode,
		"to_pincode": *toPincodeData.ToPincode,
	}).Info("Checking active partners count")

	if fromPartnersCount <= 1 || toPartnersCount <= 1 {
		reason := fmt.Sprintf("Smile Cargo not serviceable: insufficient active partners. From: %d, To: %d (need >1 for both)", fromPartnersCount, toPartnersCount)
		
		a.logger.WithFields(logrus.Fields{
			"component": "smile_cargo_adapter",
			"method":    "convertServiceAvailabilityResponse",
			"from_partners_count": fromPartnersCount,
			"to_partners_count": toPartnersCount,
			"reason": reason,
		}).Warn("Service not available: insufficient active partners")
		
		result.ErrorMessage = &reason
		return result
	}

	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "convertServiceAvailabilityResponse",
		"from_partners_count": fromPartnersCount,
		"to_partners_count": toPartnersCount,
	}).Info("Service available: sufficient active partners found")

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
		
		a.logger.WithFields(logrus.Fields{
			"component": "smile_cargo_adapter",
			"method":    "convertServiceAvailabilityResponse",
			"partner_code": partner.PartnerCode,
			"pincode_type": "source",
			"city": partner.CityName,
			"first_mile": partner.FirstMile,
			"last_mile": partner.LastMile,
		}).Info("Added source partner capability")
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
		
		a.logger.WithFields(logrus.Fields{
			"component": "smile_cargo_adapter",
			"method":    "convertServiceAvailabilityResponse",
			"partner_code": partner.PartnerCode,
			"pincode_type": "destination",
			"city": partner.CityName,
			"first_mile": partner.FirstMile,
			"last_mile": partner.LastMile,
		}).Info("Added destination partner capability")
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
		
		a.logger.WithFields(logrus.Fields{
			"component": "smile_cargo_adapter",
			"method":    "convertServiceAvailabilityResponse",
			"service_code": service.ServiceCode,
			"service_name": service.ServiceName,
			"delivery_modes_count": len(service.DeliveryModes),
		}).Info("Created service for partner")
	}

	result.Services = services

	// Set metadata
	result.Metadata["reason"] = fmt.Sprintf("Smile Cargo serviceable: %d active partners at source, %d at destination", fromPartnersCount, toPartnersCount)
	result.Metadata["from_city"] = fromPincodeData.ActivePartners[0].CityName
	result.Metadata["to_city"] = toPincodeData.ActivePartners[0].CityName
	result.Metadata["serviceable"] = true

	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "convertServiceAvailabilityResponse",
		"total_services": len(services),
		"total_capabilities": len(result.Capabilities),
		"from_city": result.Metadata["from_city"],
		"to_city": result.Metadata["to_city"],
		"serviceable": result.Metadata["serviceable"],
	}).Info("Response conversion completed successfully")

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
	
	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "getDeliveryModes",
		"partner_code": partner["partner_code"],
		"surface": partner["surface"],
		"air": partner["air"],
		"rail": partner["rail"],
		"extracted_modes": modes,
	}).Info("Extracted delivery modes from partner")
	
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
	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "hasValidPincodes",
		"source_postal_code": request.SourcePostalCode,
		"destination_postal_code": request.DestinationPostalCode,
		"postal_code": request.PostalCode,
	}).Info("Validating pincodes for Smile Cargo")

	// Need source pincode
	if request.SourcePostalCode == nil || *request.SourcePostalCode == "" {
		a.logger.WithFields(logrus.Fields{
			"component": "smile_cargo_adapter",
			"method":    "hasValidPincodes",
			"reason":    "Missing source postal code",
		}).Info("Pincode validation failed: missing source")
		return false
	}

	// Need destination pincode
	hasDestination := (request.DestinationPostalCode != nil && *request.DestinationPostalCode != "") ||
		(request.PostalCode != nil && *request.PostalCode != "")

	if !hasDestination {
		a.logger.WithFields(logrus.Fields{
			"component": "smile_cargo_adapter",
			"method":    "hasValidPincodes",
			"reason":    "Missing destination postal code",
		}).Info("Pincode validation failed: missing destination")
		return false
	}

	a.logger.WithFields(logrus.Fields{
		"component": "smile_cargo_adapter",
		"method":    "hasValidPincodes",
		"valid":     true,
		"source":    *request.SourcePostalCode,
		"destination": func() string {
			if request.DestinationPostalCode != nil {
				return *request.DestinationPostalCode
			}
			return *request.PostalCode
		}(),
	}).Info("Pincode validation successful")

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
