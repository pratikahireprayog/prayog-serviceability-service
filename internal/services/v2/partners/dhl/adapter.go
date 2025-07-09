package dhl

import (
	"context"
	"fmt"
	"strings"
	"time"

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

// GetPartnerCode returns the partner code (interface compatibility only)
// TODO: This should be removed when all code uses database values
func (a *Adapter) GetPartnerCode() string {
	return "dhl"
}

// GetPartnerName returns the partner name (interface compatibility only)
// TODO: This should be removed when all code uses database values
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

	// Partner attribute mapping in database determines supported parcel categories
	// No hardcoded category filtering needed here
	return true
}

// CheckServiceability checks if DHL can service the given request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	if !a.SupportsRequest(ctx, request) {
		return &common.PartnerServiceabilityResult{
			PartnerCode:   a.GetPartnerCode(),
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			ResponseTime:  time.Since(startTime),
			Metadata: map[string]interface{}{
				"reason": "DHL does not support this request type",
			},
		}, nil
	}

	// Convert request to DHL rates format
	dhlRequest := a.convertToRatesRequest(request)

	// Make API call to DHL rates endpoint
	response, err := a.client.CheckRates(ctx, dhlRequest)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerCode:   a.GetPartnerCode(),
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			ResponseTime:  time.Since(startTime),
			Error:         err,
			ErrorMessage:  &[]string{fmt.Sprintf("DHL API call failed: %v", err)}[0],
		}, nil
	}

	// Convert response to the expected format
	// The orchestrator will set PartnerCode and PartnerName from database
	result := a.convertRatesResponse(response)
	result.ResponseTime = time.Since(startTime)

	return result, nil
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

// convertToRatesRequest converts v2 request to DHL rates format
func (a *Adapter) convertToRatesRequest(request *models.ServiceabilityV2Request) RatesRequest {
	// Extract postal codes
	sourcePincode := a.getSourcePincode(request)
	destinationPincode := a.getDestinationPincode(request)

	// Create rates request
	dhlReq := RatesRequest{
		CustomerDetails: CustomerDetails{
			ShipperDetails: ShipperDetails{
				PostalCode:  sourcePincode,
				CityName:    "Origin City", // Could be enhanced with actual city lookup
				CountryCode: "IN",          // Default origin country
			},
			ReceiverDetails: ReceiverDetails{
				PostalCode:  destinationPincode,
				CityName:    "Destination City", // Could be enhanced with actual city lookup
				CountryCode: *request.CountryCode,
			},
		},
		Accounts: []Account{
			{
				TypeCode: "shipper",
				Number:   a.config.AccountNumber,
			},
		},
		ProductsAndServices: []ProductAndService{
			{
				ProductCode:      "P", // Default to express worldwide
				LocalProductCode: "P",
			},
		},
		PayerCountryCode:           "IN",
		PlannedShippingDateAndTime: a.getPlannedShippingDateTime(),
		UnitOfMeasurement:          "metric",
		IsCustomsDeclarable:        true,
		EstimatedDeliveryDate: EstimatedDeliveryDate{
			IsRequested: true,
			TypeCode:    "QDDC",
		},
		ReturnStandardProductsOnly: true,
		Packages:                   a.getPackages(request),
	}

	return dhlReq
}

// convertRatesResponse converts DHL rates response to common format with capabilities
func (a *Adapter) convertRatesResponse(response *RatesResponse) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerCode:   a.GetPartnerCode(),
		IsServiceable: len(response.Products) > 0,
		Services:      make([]models.ServiceV2, 0),
		Capabilities:  make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
	}

	if len(response.Products) == 0 {
		result.Metadata["reason"] = "No DHL products available for the requested route"
		return result
	}

	// Use the first product for capabilities (DHL typically returns one main product)
	product := response.Products[0]

	// Build capabilities structure as per the expected format
	capabilities := map[string]interface{}{
		"pickup_capabilities": map[string]interface{}{
			"next_business_day":                          product.PickupCapabilities.NextBusinessDay,
			"local_cutoff_date_and_time":                 product.PickupCapabilities.LocalCutoffDateAndTime,
			"pickup_earliest":                            product.PickupCapabilities.PickupEarliest,
			"pickup_latest":                              product.PickupCapabilities.PickupLatest,
			"pickup_cutoff_same_day_outbound_processing": product.PickupCapabilities.PickupCutoffSameDayOutboundProcessing,
			"origin_service_area_code":                   product.PickupCapabilities.OriginServiceAreaCode,
			"origin_facility_area_code":                  product.PickupCapabilities.OriginFacilityAreaCode,
			"pickup_additional_days":                     product.PickupCapabilities.PickupAdditionalDays,
			"pickup_day_of_week":                         product.PickupCapabilities.PickupDayOfWeek,
		},
		"delivery_capabilities": map[string]interface{}{
			"delivery_type_code":               product.DeliveryCapabilities.DeliveryTypeCode,
			"estimated_delivery_date_and_time": product.DeliveryCapabilities.EstimatedDeliveryDateAndTime,
			"destination_service_area_code":    product.DeliveryCapabilities.DestinationServiceAreaCode,
			"destination_facility_area_code":   product.DeliveryCapabilities.DestinationFacilityAreaCode,
			"delivery_additional_days":         product.DeliveryCapabilities.DeliveryAdditionalDays,
			"delivery_day_of_week":             product.DeliveryCapabilities.DeliveryDayOfWeek,
			"total_transit_days":               product.DeliveryCapabilities.TotalTransitDays,
		},
	}

	result.Capabilities = capabilities

	// Add services for backward compatibility
	for _, product := range response.Products {
		serviceV2 := models.ServiceV2{
			ServiceCode: product.ProductCode,
			ServiceName: product.ProductName,
			TATDays:     product.DeliveryCapabilities.TotalTransitDays,
			IsCOD:       false, // DHL typically doesn't do COD internationally
			Pickup:      true,
			Delivery:    true,
			Insurance:   true,
			ProductTypes: map[string]bool{
				"international": true,
				"express":       strings.Contains(strings.ToLower(product.ProductName), "express"),
				"economy":       strings.Contains(strings.ToLower(product.ProductName), "economy"),
			},
			DeliveryModes: map[string]bool{
				"international": true,
				"express":       strings.Contains(strings.ToLower(product.ProductName), "express"),
			},
		}

		// Add pricing if available
		if len(product.TotalPrice) > 0 {
			serviceV2.Pricing = &models.ServicePricingV2{
				BaseCost:      product.TotalPrice[0].Price,
				Currency:      product.TotalPrice[0].PriceCurrency,
				CODCharges:    0.0,
				FuelSurcharge: a.extractFuelSurcharge(product.DetailedPriceBreakdown),
			}
		}

		result.Services = append(result.Services, serviceV2)
	}

	// Set metadata
	result.Metadata["reason"] = fmt.Sprintf("DHL offers %d services", len(result.Services))
	result.Metadata["product_count"] = len(response.Products)
	result.Metadata["exchange_rates"] = len(response.ExchangeRates)

	return result
}

// extractFuelSurcharge extracts fuel surcharge from detailed price breakdown
func (a *Adapter) extractFuelSurcharge(breakdowns []DetailedPriceBreakdown) float64 {
	for _, breakdown := range breakdowns {
		for _, item := range breakdown.Breakdown {
			if strings.Contains(strings.ToLower(item.Name), "fuel") {
				return item.Price
			}
		}
	}
	return 0.0
}

// getSourcePincode extracts source pincode from request
func (a *Adapter) getSourcePincode(request *models.ServiceabilityV2Request) string {
	if request.SourcePostalCode != nil {
		return *request.SourcePostalCode
	}
	return "110001" // Default to Delhi if no source specified
}

// getDestinationPincode extracts destination pincode from request
func (a *Adapter) getDestinationPincode(request *models.ServiceabilityV2Request) string {
	if request.DestinationPostalCode != nil {
		return *request.DestinationPostalCode
	}
	if request.PostalCode != nil {
		return *request.PostalCode
	}
	return ""
}

// getPlannedShippingDateTime gets the planned shipping date and time
func (a *Adapter) getPlannedShippingDateTime() string {
	// Default to next business day at 1 PM IST
	now := time.Now()
	nextDay := now.Add(24 * time.Hour)
	return nextDay.Format("2006-01-02T15:04:05GMT+05:30")
}

// getPackages creates packages from request
func (a *Adapter) getPackages(request *models.ServiceabilityV2Request) []Package {
	if request.Package != nil {
		return []Package{
			{
				Weight:     a.getWeight(request.Package.Weight),
				Dimensions: a.getDimensions(request.Package.Dimensions),
			},
		}
	}

	// Default package
	return []Package{
		{
			Weight: 0.5, // Default 0.5 kg
			Dimensions: Dimensions{
				Length: 30,
				Width:  20,
				Height: 15,
			},
		},
	}
}

// getWeight converts weight to kg
func (a *Adapter) getWeight(weight *models.Weight) float64 {
	if weight == nil {
		return 0.5 // Default weight
	}

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
		return weight.Value // Assume kg if unit unknown
	}
}

// getDimensions converts dimensions to cm
func (a *Adapter) getDimensions(dimensions *models.Dimensions) Dimensions {
	if dimensions == nil {
		return Dimensions{
			Length: 30,
			Width:  20,
			Height: 15,
		}
	}

	return Dimensions{
		Length: a.convertToCm(dimensions.Length, dimensions.Unit),
		Width:  a.convertToCm(dimensions.Width, dimensions.Unit),
		Height: a.convertToCm(dimensions.Height, dimensions.Unit),
	}
}

// convertToCm converts dimension to cm
func (a *Adapter) convertToCm(value float64, unit string) float64 {
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
		return value // Assume cm if unit unknown
	}
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
	// Test authentication
	if err := a.client.auth.Authenticate(ctx); err != nil {
		return fmt.Errorf("DHL authentication failed: %w", err)
	}
	return nil
}

// IsHealthy implements PartnerAdapter interface
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	if !a.config.Enabled {
		return false
	}

	// Check if credentials are available
	return a.client.auth.IsAuthenticated()
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

	// Convert to quote request format
	quoteRequest := QuoteRequest{
		DestinationCountryCode: *request.CountryCode,
		OriginCountryCode:      "IN",
		ServiceType:            "P", // Express
		DestinationPostalCode:  a.getDestinationPincode(request),
		OriginPostalCode:       a.getSourcePincode(request),
		Packages:               a.convertToQuotePackages(request),
	}

	return a.client.GetQuote(ctx, quoteRequest)
}

// convertToQuotePackages converts request packages to quote format
func (a *Adapter) convertToQuotePackages(request *models.ServiceabilityV2Request) []DHLPackage {
	if request.Package == nil {
		return []DHLPackage{
			{
				Weight:        0.5,
				Length:        30,
				Width:         20,
				Height:        15,
				DeclaredValue: 100.0,
				Currency:      "INR",
			},
		}
	}

	return []DHLPackage{
		{
			Weight:        a.getWeight(request.Package.Weight),
			Length:        a.convertToCm(request.Package.Dimensions.Length, request.Package.Dimensions.Unit),
			Width:         a.convertToCm(request.Package.Dimensions.Width, request.Package.Dimensions.Unit),
			Height:        a.convertToCm(request.Package.Dimensions.Height, request.Package.Dimensions.Unit),
			DeclaredValue: 100.0, // Default value
			Currency:      "INR",
		},
	}
}
