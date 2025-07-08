package dhl

import (
	"context"
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// Adapter implements the PartnerAdapter interface for DHL international shipping
type Adapter struct {
	client *DHLClient
	config config.DHLConfig
}

// NewAdapter creates a new DHL adapter instance
func NewAdapter(config config.DHLConfig) *Adapter {
	return &Adapter{
		client: NewDHLClient(config),
		config: config,
	}
}

// GetPartnerCode returns the partner code
func (a *Adapter) GetPartnerCode() string {
	return "dhl"
}

// GetPartnerName returns the partner name
func (a *Adapter) GetPartnerName() string {
	return "DHL"
}

// GetAdapterType returns the adapter type
func (a *Adapter) GetAdapterType() common.AdapterType {
	return common.AdapterTypeHTTP
}

// IsEnabled returns whether the adapter is enabled
func (a *Adapter) IsEnabled() bool {
	return a.config.Enabled
}

// GetRating returns the partner rating
func (a *Adapter) GetRating() float64 {
	return a.config.Rating
}

// SupportsRequest checks if DHL supports the given request
func (a *Adapter) SupportsRequest(ctx context.Context, request *models.ServiceabilityV2Request) bool {
	// DHL specializes in international shipping
	if !a.IsInternationalRequest(request) {
		return false
	}

	// Check if request has required fields for DHL
	if request.CountryCode == nil || *request.CountryCode == "" {
		return false
	}

	// DHL handles international parcels - check if it's a supported category
	if request.ParcelCategory != nil {
		category := strings.ToLower(*request.ParcelCategory)
		if category != "international" && category != "ecomm" && category != "courier" {
			return false
		}
	}

	return true
}

// CheckServiceability checks if DHL can service the given request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request) (*common.PartnerServiceabilityResult, error) {
	if !a.SupportsRequest(ctx, request) {
		return &common.PartnerServiceabilityResult{
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			Metadata: map[string]interface{}{
				"reason": "DHL does not support this request type",
			},
		}, nil
	}

	// Convert request to DHL format
	dhlRequest := a.convertToServiceabilityRequest(request)

	// Make API call
	response, err := a.client.CheckServiceability(ctx, dhlRequest)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			Error:         err,
			ErrorMessage:  &[]string{fmt.Sprintf("DHL API call failed: %v", err)}[0],
		}, nil
	}

	// Convert response and return
	// The orchestrator will set PartnerCode and PartnerName from database
	return a.convertServiceabilityResponse(response), nil
}

// validateDHLRequirements validates DHL-specific requirements
func (a *Adapter) validateDHLRequirements(request *models.ServiceabilityV2Request) error {
	// DHL requires both source and destination postal codes
	if request.SourcePostalCode == nil || *request.SourcePostalCode == "" {
		return fmt.Errorf("source postal code is required for DHL shipments")
	}

	if request.DestinationPostalCode == nil || *request.DestinationPostalCode == "" {
		return fmt.Errorf("destination postal code is required for DHL shipments")
	}

	// DHL requires package information for international shipments
	if a.IsInternationalRequest(request) {
		if request.Package == nil {
			return fmt.Errorf("package information is required for international shipments via DHL")
		}

		// Validate weight is provided
		if request.Package.Weight == nil {
			return fmt.Errorf("weight information is required for international shipments via DHL")
		}

		// Validate dimensions are provided
		if request.Package.Dimensions == nil {
			return fmt.Errorf("dimensions information is required for international shipments via DHL")
		}
	}

	return nil
}

// convertToServiceabilityRequest converts v2 request to DHL format
func (a *Adapter) convertToServiceabilityRequest(request *models.ServiceabilityV2Request) ServiceabilityRequest {
	dhlReq := ServiceabilityRequest{}

	// Set country codes
	if request.CountryCode != nil {
		dhlReq.DestinationCountryCode = *request.CountryCode
	}
	dhlReq.OriginCountryCode = "IN" // Default origin

	// Add postal codes if available
	if request.SourcePostalCode != nil {
		dhlReq.OriginPostalCode = *request.SourcePostalCode
	}
	if request.DestinationPostalCode != nil {
		dhlReq.DestinationPostalCode = *request.DestinationPostalCode
	} else if request.PostalCode != nil {
		dhlReq.DestinationPostalCode = *request.PostalCode
	}

	// Set product type based on parcel category
	if request.ParcelCategory != nil {
		switch strings.ToLower(*request.ParcelCategory) {
		case "international":
			dhlReq.ProductType = "EXPRESS"
		case "ecomm":
			dhlReq.ProductType = "ECONOMY"
		case "courier":
			dhlReq.ProductType = "EXPRESS"
		default:
			dhlReq.ProductType = "EXPRESS" // Default to express
		}
	} else {
		dhlReq.ProductType = "EXPRESS"
	}

	// Set weight from package information if available, otherwise use default
	if request.Package != nil && request.Package.Weight != nil {
		dhlReq.Weight = a.convertWeightToKg(request.Package.Weight)
	} else {
		dhlReq.Weight = 1.0 // 1kg default
	}

	dhlReq.Currency = "INR"

	return dhlReq
}

// convertWeightToKg converts weight from different units to kilograms
func (a *Adapter) convertWeightToKg(weight *models.Weight) float64 {
	switch strings.ToLower(weight.Unit) {
	case "kg":
		return weight.Value
	case "g":
		return weight.Value / 1000.0
	case "lb":
		return weight.Value * 0.453592
	case "oz":
		return weight.Value * 0.0283495
	default:
		// Default to kg if unit is not recognized
		return weight.Value
	}
}

// convertDimensionToCm converts dimension from different units to centimeters
func (a *Adapter) convertDimensionToCm(value float64, unit string) float64 {
	switch strings.ToLower(unit) {
	case "cm":
		return value
	case "mm":
		return value / 10.0
	case "in":
		return value * 2.54
	case "ft":
		return value * 30.48
	default:
		// Default to cm if unit is not recognized
		return value
	}
}

// convertServiceabilityResponse converts DHL response to common format
func (a *Adapter) convertServiceabilityResponse(response *ServiceabilityResponse) *common.PartnerServiceabilityResult {
	// TODO: Implement new DHL capabilities structure as per final payload format:
	// capabilities: {
	//   "pickup_capabilities": {
	//     "next_business_day": false,
	//     "local_cutoff_date_and_time": "2025-05-05T12:30:00",
	//     "pickup_earliest": "10:00:00",
	//     "pickup_latest": "20:30:00",
	//     "pickup_cutoff_same_day_outbound_processing": "14:30:00",
	//     "origin_service_area_code": "BLR",
	//     "origin_facility_area_code": "YPU",
	//     "pickup_additional_days": 0,
	//     "pickup_day_of_week": 1
	//   },
	//   "delivery_capabilities": {
	//     "delivery_type_code": "QDDC",
	//     "estimated_delivery_date_and_time": "2025-05-12T23:59:00",
	//     "destination_service_area_code": "TAO",
	//     "destination_facility_area_code": "QDN",
	//     "delivery_additional_days": 0,
	//     "delivery_day_of_week": 1,
	//     "total_transit_days": 7
	//   }
	// }

	result := &common.PartnerServiceabilityResult{
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	if !response.Success || response.Data == nil {
		result.IsServiceable = false
		reason := "Unknown error"
		if response.Error != nil {
			reason = response.Error.Message
		}
		result.Metadata["reason"] = reason
		return result
	}

	data := response.Data
	result.IsServiceable = data.IsServiceable

	if !data.IsServiceable {
		reason := "Not serviceable by DHL"
		if len(data.Restrictions) > 0 {
			reason = strings.Join(data.Restrictions, "; ")
		}
		result.Metadata["reason"] = reason
		return result
	}

	// Convert services
	for _, service := range data.Services {
		serviceV2 := models.ServiceV2{
			ServiceCode: service.ProductCode,
			ServiceName: service.ProductName,
			TATDays:     service.TransitDays,
			IsCOD:       false, // DHL typically doesn't do COD internationally
			Pickup:      true,
			Delivery:    true,
			Insurance:   true,
			ProductTypes: map[string]bool{
				"international": true,
				"express":       strings.Contains(strings.ToLower(service.ServiceType), "express"),
				"economy":       strings.Contains(strings.ToLower(service.ServiceType), "economy"),
			},
			DeliveryModes: map[string]bool{
				"international": true,
				"express":       strings.Contains(strings.ToLower(service.ServiceType), "express"),
			},
		}

		// Add pricing if available
		if service.Pricing != nil {
			serviceV2.Pricing = &models.ServicePricingV2{
				BaseCost:      service.Pricing.BaseCost,
				Currency:      service.Pricing.Currency,
				CODCharges:    0.0,
				FuelSurcharge: service.Pricing.FuelSurcharge,
			}
		}

		result.Services = append(result.Services, serviceV2)
	}

	// Set capabilities
	result.Capabilities["international"] = true
	result.Capabilities["tracking"] = true
	result.Capabilities["express_delivery"] = true
	result.Capabilities["customs_clearance"] = true

	// Set reason for successful response
	if len(result.Services) > 0 {
		result.Metadata["reason"] = fmt.Sprintf("DHL offers %d services", len(result.Services))
	} else {
		result.Metadata["reason"] = "No DHL services available"
	}

	return result
}

// IsInternationalRequest checks if the request is for international shipping
func (a *Adapter) IsInternationalRequest(request *models.ServiceabilityV2Request) bool {
	if request.CountryCode == nil {
		return false
	}

	originCountry := "IN" // Default origin
	destCountry := strings.ToUpper(*request.CountryCode)

	// Consider it international if countries are different
	return originCountry != destCountry && destCountry != ""
}

// Initialize implements PartnerAdapter interface
func (a *Adapter) Initialize(ctx context.Context) error {
	// Test authentication if credentials are provided
	if a.config.Username != "" && a.config.Password != "" {
		if _, err := a.client.auth.GetAuthToken(ctx); err != nil {
			return fmt.Errorf("DHL authentication failed: %w", err)
		}
	}
	return nil
}

// IsHealthy implements PartnerAdapter interface
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	if !a.config.Enabled {
		return false
	}

	// TODO: Add actual health check by calling DHL API
	return true
}

// GetMetrics implements PartnerAdapter interface
func (a *Adapter) GetMetrics() *common.PartnerMetrics {
	return &common.PartnerMetrics{
		PartnerCode:         a.GetPartnerCode(),
		TotalRequests:       0, // TODO: Implement actual metrics
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

// GetQuote gets a quote for international shipping (optional method for enhanced functionality)
func (a *Adapter) GetQuote(ctx context.Context, request *models.ServiceabilityV2Request) (*QuoteResponse, error) {
	if !a.SupportsRequest(ctx, request) {
		return nil, fmt.Errorf("quote request not supported by DHL")
	}

	quoteRequest := QuoteRequest{
		DestinationCountryCode: *request.CountryCode,
		OriginCountryCode:      "IN",
		ServiceType:            "EXPRESS",
	}

	if request.DestinationPostalCode != nil {
		quoteRequest.DestinationPostalCode = *request.DestinationPostalCode
	} else if request.PostalCode != nil {
		quoteRequest.DestinationPostalCode = *request.PostalCode
	}

	if request.SourcePostalCode != nil {
		quoteRequest.OriginPostalCode = *request.SourcePostalCode
	}

	// Add package information for quote
	pkg := DHLPackage{
		DeclaredValue: 100.0, // default value
		Currency:      "INR",
	}

	// Use package information from request if available
	if request.Package != nil {
		if request.Package.Weight != nil {
			pkg.Weight = a.convertWeightToKg(request.Package.Weight)
		} else {
			pkg.Weight = 1.0 // default 1kg
		}

		if request.Package.Dimensions != nil {
			pkg.Length = a.convertDimensionToCm(request.Package.Dimensions.Length, request.Package.Dimensions.Unit)
			pkg.Width = a.convertDimensionToCm(request.Package.Dimensions.Width, request.Package.Dimensions.Unit)
			pkg.Height = a.convertDimensionToCm(request.Package.Dimensions.Height, request.Package.Dimensions.Unit)
		} else {
			// Use default dimensions
			pkg.Length = 30 // default 30cm
			pkg.Width = 20  // default 20cm
			pkg.Height = 15 // default 15cm
		}
	} else {
		// Use all default values
		pkg.Weight = 1.0 // default 1kg
		pkg.Length = 30  // default 30cm
		pkg.Width = 20   // default 20cm
		pkg.Height = 15  // default 15cm
	}

	quoteRequest.Packages = []DHLPackage{pkg}

	return a.client.GetQuote(ctx, quoteRequest)
}
