package dhl

import (
	"context"
	"fmt"
	"strings"
	"time"

	services "prayog-serviceability-service/internal/services/v1/data"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/sirupsen/logrus"
)

// Adapter implements the PartnerAdapter interface for DHL international shipping
type Adapter struct {
	client             *DHLClient
	config             config.DHLConfig
	geolocationService services.GeolocationService
	hubLocationService services.HubLocationService
	logger             *logrus.Logger
}

// NewAdapter creates a new DHL adapter instance
func NewAdapter(config config.DHLConfig, geolocationService services.GeolocationService, hubLocationService services.HubLocationService) *Adapter {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Log configuration with redacted sensitive fields
	logger.WithFields(logrus.Fields{
		"partner":    "DHL",
		"base_url":   config.BaseURL,
		"username":   redactField(config.Username),
		"enabled":    config.Enabled,
		"basic_auth": "[REDACTED]", // Never log credentials
	}).Info("Creating DHL adapter")

	return &Adapter{
		client:             NewDHLClient(config),
		config:             config,
		geolocationService: geolocationService,
		hubLocationService: hubLocationService,
		logger:             logger,
	}
}

// redactField safely redacts sensitive fields for logging
func redactField(field string) string {
	if field == "" {
		return "[EMPTY]"
	}
	if len(field) <= 3 {
		return "[REDACTED]"
	}
	return field[:3] + "***"
}

// GetAdapterType returns the adapter type
func (a *Adapter) GetAdapterType() common.AdapterType {
	return common.AdapterTypeHTTP
}

// IsEnabled returns whether the adapter is enabled
func (a *Adapter) IsEnabled() bool {
	return a.config.Enabled
}

// CheckServiceability checks if DHL can service the given request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	// Validate DHL-specific requirements
	if err := a.validateDHLRequirements(request); err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:     partnerInfo.PartnerID,
			PartnerCode:   partnerInfo.PartnerCode,
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			ResponseTime:  time.Since(startTime),
			Error:         err,
			ErrorMessage:  &[]string{fmt.Sprintf("DHL validation failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "DHL validation failed",
			},
		}, nil
	}

	// Use international flow for DHL
	return a.checkInternationalServiceability(ctx, request, startTime, partnerInfo)
}

// checkInternationalServiceability implements the international serviceability flow for DHL
func (a *Adapter) checkInternationalServiceability(ctx context.Context, request *models.ServiceabilityV2Request, startTime time.Time, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	// Step 1: Extract postal codes from request
	sourcePincode := a.getSourcePincode(request)
	destinationPincode := a.getDestinationPincode(request)

	a.logger.WithFields(logrus.Fields{
		"partner":             "DHL",
		"source_pincode":      sourcePincode,
		"destination_pincode": destinationPincode,
		"flow":                "international",
	}).Info("Starting DHL international serviceability check")

	// Step 2: Find nearest hub for source postal code
	hubLocation, err := a.findNearestHub(ctx, sourcePincode)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:     partnerInfo.PartnerID,
			PartnerCode:   partnerInfo.PartnerCode,
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			ResponseTime:  time.Since(startTime),
			Error:         err,
			ErrorMessage:  &[]string{fmt.Sprintf("Hub location not found: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason":         "Hub location not found",
				"source_pincode": sourcePincode,
				"step":           "hub_lookup",
			},
		}, nil
	}

	a.logger.WithFields(logrus.Fields{
		"partner":                       "DHL",
		"source_pincode":                sourcePincode,
		"hub_postal_code":               hubLocation.PostalCode,
		"international_hub_postal_code": hubLocation.InternationalHub.PostalCode,
		"international_hub_city_code":   hubLocation.InternationalHub.CityCode,
	}).Info("Found nearest hub for source postal code")

	// Step 3: Get country codes for source and destination
	sourceCountryCode, destinationCountryCode, err := a.resolveCountryCodes(ctx, sourcePincode, destinationPincode)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:     partnerInfo.PartnerID,
			PartnerCode:   partnerInfo.PartnerCode,
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			ResponseTime:  time.Since(startTime),
			Error:         err,
			ErrorMessage:  &[]string{fmt.Sprintf("Country code resolution failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason":              "Country code resolution failed",
				"source_pincode":      sourcePincode,
				"destination_pincode": destinationPincode,
				"step":                "country_code_resolution",
			},
		}, nil
	}

	a.logger.WithFields(logrus.Fields{
		"partner":                  "DHL",
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
		"flow":                     "international",
	}).Info("Resolved country codes for international flow")

	// Step 4: Create DHL API request with hardcoded values
	dhlRequest := a.createInternationalRatesRequest(request, sourceCountryCode, destinationCountryCode)

	// Step 5: Call DHL API
	response, err := a.client.CheckRates(ctx, dhlRequest)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:     partnerInfo.PartnerID,
			PartnerCode:   partnerInfo.PartnerCode,
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			ResponseTime:  time.Since(startTime),
			Error:         err,
			ErrorMessage:  &[]string{fmt.Sprintf("DHL API call failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason":                   "DHL API call failed",
				"source_country_code":      sourceCountryCode,
				"destination_country_code": destinationCountryCode,
				"step":                     "dhl_api_call",
			},
		}, nil
	}

	// Step 6: Process response
	if len(response.Products) == 0 {
		return &common.PartnerServiceabilityResult{
			PartnerID:     partnerInfo.PartnerID,
			PartnerCode:   partnerInfo.PartnerCode,
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			ResponseTime:  time.Since(startTime),
			Metadata: map[string]interface{}{
				"reason":                   "No DHL products available",
				"source_country_code":      sourceCountryCode,
				"destination_country_code": destinationCountryCode,
				"hub_info":                 hubLocation,
			},
		}, nil
	}

	// Step 7: Convert response to serviceability result
	result := a.convertRatesResponse(response, partnerInfo)
	result.ResponseTime = time.Since(startTime)
	result.Metadata["flow"] = "international"
	result.Metadata["source_country_code"] = sourceCountryCode
	result.Metadata["destination_country_code"] = destinationCountryCode
	result.Metadata["hub_info"] = hubLocation
	result.Metadata["product_code_used"] = "P" // Hardcoded as per requirements

	a.logger.WithFields(logrus.Fields{
		"partner":                  "DHL",
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
		"services":                 len(result.Services),
		"is_serviceable":           result.IsServiceable,
		"flow":                     "international",
	}).Info("DHL international serviceability check completed")

	return result, nil
}

// findNearestHub finds the nearest hub for a given postal code
func (a *Adapter) findNearestHub(ctx context.Context, postalCode string) (*models.HubLocationInfo, error) {
	a.logger.WithFields(logrus.Fields{
		"partner":     "DHL",
		"postal_code": postalCode,
		"method":      "findNearestHub",
	}).Debug("Looking up nearest hub")

	// Check if hub location service is available
	if a.hubLocationService == nil {
		return nil, fmt.Errorf("hub location service is not available")
	}

	// Get nearest hub from service
	hubLocation, err := a.hubLocationService.GetNearestHubByPostalCode(ctx, postalCode)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": postalCode,
			"error":       err.Error(),
		}).Error("Failed to find nearest hub")
		return nil, fmt.Errorf("failed to find nearest hub for postal code %s: %w", postalCode, err)
	}

	// Validate that international hub information is available
	if hubLocation.InternationalHub == nil {
		return nil, fmt.Errorf("no international hub found for postal code %s", postalCode)
	}

	a.logger.WithFields(logrus.Fields{
		"partner":                       "DHL",
		"postal_code":                   postalCode,
		"hub_postal_code":               hubLocation.PostalCode,
		"international_hub_postal_code": hubLocation.InternationalHub.PostalCode,
		"international_hub_city_code":   hubLocation.InternationalHub.CityCode,
	}).Debug("Successfully found nearest hub")

	return hubLocation, nil
}

// resolveCountryCodes resolves country codes for source and destination postal codes
func (a *Adapter) resolveCountryCodes(ctx context.Context, sourcePincode, destinationPincode string) (string, string, error) {
	a.logger.WithFields(logrus.Fields{
		"partner":             "DHL",
		"source_pincode":      sourcePincode,
		"destination_pincode": destinationPincode,
		"method":              "resolveCountryCodes",
	}).Debug("Resolving country codes")

	// Check if geolocation service is available
	if a.geolocationService == nil {
		return "", "", fmt.Errorf("geolocation service is not available")
	}

	// Get source country code
	sourceCountryCode, err := a.geolocationService.GetCountryCodeByPostalCode(ctx, sourcePincode)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"partner":        "DHL",
			"source_pincode": sourcePincode,
			"error":          err.Error(),
		}).Error("Failed to get source country code")
		return "", "", fmt.Errorf("failed to get source country code for postal code %s: %w", sourcePincode, err)
	}

	// Get destination country code
	destinationCountryCode, err := a.geolocationService.GetCountryCodeByPostalCode(ctx, destinationPincode)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"partner":             "DHL",
			"destination_pincode": destinationPincode,
			"error":               err.Error(),
		}).Error("Failed to get destination country code")
		return "", "", fmt.Errorf("failed to get destination country code for postal code %s: %w", destinationPincode, err)
	}

	a.logger.WithFields(logrus.Fields{
		"partner":                  "DHL",
		"source_country_code":      *sourceCountryCode,
		"destination_country_code": *destinationCountryCode,
	}).Debug("Successfully resolved country codes")

	return *sourceCountryCode, *destinationCountryCode, nil
}

// createInternationalRatesRequest creates a DHL rates request with hardcoded values for international flow
func (a *Adapter) createInternationalRatesRequest(request *models.ServiceabilityV2Request, sourceCountryCode, destinationCountryCode string) RatesRequest {
	// Extract postal codes
	sourcePincode := a.getSourcePincode(request)
	destinationPincode := a.getDestinationPincode(request)

	a.logger.WithFields(logrus.Fields{
		"partner":                  "DHL",
		"source_pincode":           sourcePincode,
		"destination_pincode":      destinationPincode,
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
		"method":                   "createInternationalRatesRequest",
	}).Debug("Creating DHL rates request with hardcoded values")

	// Create rates request with hardcoded values as per requirements
	dhlReq := RatesRequest{
		CustomerDetails: CustomerDetails{
			ShipperDetails: ShipperDetails{
				PostalCode:  sourcePincode,
				CityName:    "Bangalore", // Enhanced with actual city
				CountryCode: sourceCountryCode,
			},
			ReceiverDetails: ReceiverDetails{
				PostalCode:  destinationPincode,
				CityName:    "Destination City", // Could be enhanced with actual city lookup
				CountryCode: destinationCountryCode,
			},
		},
		// Hardcoded account information as per requirements
		Accounts: []Account{
			{
				TypeCode: "shipper",
				Number:   "533748932",
			},
		},
		// Hardcoded product information as per requirements
		ProductsAndServices: []ProductAndService{
			{
				ProductCode:      "P",
				LocalProductCode: "P",
			},
		},
		PayerCountryCode: "IN",
		// Hardcoded shipping date as per requirements
		PlannedShippingDateAndTime: "2025-05-05T13:00:00GMT+05:30",
		// Hardcoded unit of measurement as per requirements
		UnitOfMeasurement: "metric",
		// Hardcoded customs declarable flag as per requirements
		IsCustomsDeclarable: true,
		// Hardcoded estimated delivery date configuration as per requirements
		EstimatedDeliveryDate: EstimatedDeliveryDate{
			IsRequested: true,
			TypeCode:    "QDDC",
		},
		// Hardcoded return standard products only flag as per requirements
		ReturnStandardProductsOnly: true,
		Packages:                   a.getPackages(request),
	}

	a.logger.WithFields(logrus.Fields{
		"partner":                  "DHL",
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
		"product_code":             "P",
		"account_number":           "533748932",
		"unit_of_measurement":      "metric",
		"customs_declarable":       true,
	}).Debug("Created DHL rates request with hardcoded values")

	return dhlReq
}

// checkServiceabilityWithFallback attempts multiple product codes
func (a *Adapter) checkServiceabilityWithFallback(ctx context.Context, request *models.ServiceabilityV2Request, startTime time.Time, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	// Determine product codes to try based on ACTUAL route geography, not just parcel_category
	destinationCountry, err := a.getDestinationCountryCode(ctx, request, a.getDestinationPincode(request))
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:     partnerInfo.PartnerID,
			PartnerCode:   partnerInfo.PartnerCode,
			IsServiceable: false,
			Services:      make([]models.ServiceV2, 0),
			ResponseTime:  time.Since(startTime),
			Error:         err,
			ErrorMessage:  &[]string{fmt.Sprintf("Failed to determine destination country: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "Failed to determine destination country",
			},
		}, nil
	}
	sourceCountry := "IN" // Source is always India

	// Use actual route geography for product code selection
	isActuallyInternational := destinationCountry != sourceCountry

	a.logger.WithFields(logrus.Fields{
		"partner":                "DHL",
		"source_country":         sourceCountry,
		"destination_country":    destinationCountry,
		"actually_international": isActuallyInternational,
	}).Debug("Route analysis completed")

	var productCodes []string
	if isActuallyInternational {
		// International product codes (in order of preference)
		productCodes = []string{"U", "P", "D", "G", "J"}
		a.logger.WithFields(logrus.Fields{
			"partner":       "DHL",
			"route_type":    "international",
			"product_codes": productCodes,
		}).Debug("Using international product codes")
	} else {
		// Domestic product codes for India (in order of preference)
		productCodes = []string{"B", "D", "E", "G", "H"}
		a.logger.WithFields(logrus.Fields{
			"partner":       "DHL",
			"route_type":    "domestic",
			"product_codes": productCodes,
		}).Debug("Using domestic product codes")
	}

	var lastError error
	var attempts []string

	// Try each product code
	for _, productCode := range productCodes {
		attempts = append(attempts, productCode)
		a.logger.WithFields(logrus.Fields{
			"partner":      "DHL",
			"product_code": productCode,
		}).Debug("Trying DHL product code")

		// Create request with this product code
		dhlRequest := a.convertToRatesRequestWithProductCode(request, productCode)

		// Make API call
		response, err := a.client.CheckRates(ctx, dhlRequest)
		if err != nil {
			lastError = err
			a.logger.WithFields(logrus.Fields{
				"partner":      "DHL",
				"product_code": productCode,
				"error":        err.Error(),
			}).Debug("DHL product code failed")
			continue
		}

		// If we got a valid response with products, return success
		if len(response.Products) > 0 {
			result := a.convertRatesResponse(response, partnerInfo)
			result.ResponseTime = time.Since(startTime)
			result.Metadata["product_code_used"] = productCode
			result.Metadata["product_codes_attempted"] = attempts
			result.Metadata["actual_route_type"] = map[string]bool{"international": isActuallyInternational}
			a.logger.WithFields(logrus.Fields{
				"partner":      "DHL",
				"product_code": productCode,
				"services":     len(result.Services),
			}).Debug("Success with product code")
			return result, nil
		}

		// No products in response, try next product code
		lastError = fmt.Errorf("no products available for product code %s", productCode)
		a.logger.WithFields(logrus.Fields{
			"partner":      "DHL",
			"product_code": productCode,
		}).Debug("Product code returned no products")
	}

	// All product codes failed
	return &common.PartnerServiceabilityResult{
		PartnerID:     partnerInfo.PartnerID,
		PartnerCode:   partnerInfo.PartnerCode,
		IsServiceable: false,
		Services:      make([]models.ServiceV2, 0),
		ResponseTime:  time.Since(startTime),
		Error:         lastError,
		ErrorMessage:  &[]string{fmt.Sprintf("DHL API call failed: %v", lastError)}[0],
		Metadata: map[string]interface{}{
			"reason":                  "All product codes failed",
			"product_codes_attempted": attempts,
			"actual_route_type":       map[string]bool{"international": isActuallyInternational},
			"source_country":          sourceCountry,
			"destination_country":     destinationCountry,
		},
	}, nil
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

	// DHL will use default package information if not provided
	// No additional restrictions based on request type

	return nil
}

// convertToRatesRequest converts v2 request to DHL rates format
func (a *Adapter) convertToRatesRequest(request *models.ServiceabilityV2Request) RatesRequest {
	// Extract postal codes
	sourcePincode := a.getSourcePincode(request)
	destinationPincode := a.getDestinationPincode(request)

	// For international requests, we need to determine the destination country
	destinationCountry, err := a.getDestinationCountryCode(context.Background(), request, destinationPincode)
	if err != nil {
		// Log error and use default country
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": destinationPincode,
			"error":       err.Error(),
		}).Debug("Failed to get destination country code, using default")
		destinationCountry = "CN"
	}

	// Create rates request with static values
	dhlReq := RatesRequest{
		CustomerDetails: CustomerDetails{
			ShipperDetails: ShipperDetails{
				PostalCode:  sourcePincode,
				CityName:    "Bangalore", // Enhanced with actual city
				CountryCode: "IN",        // Source is always India
			},
			ReceiverDetails: ReceiverDetails{
				PostalCode:  destinationPincode,
				CityName:    "Destination City", // Could be enhanced with actual city lookup
				CountryCode: destinationCountry,
			},
		},
		// Static account information
		Accounts: []Account{
			{
				TypeCode: "shipper",
				Number:   "533748932",
			},
		},
		// Static product and service information
		ProductsAndServices: []ProductAndService{
			{
				ProductCode:      "P",
				LocalProductCode: "P",
			},
		},
		PayerCountryCode:           "IN",
		PlannedShippingDateAndTime: a.getPlannedShippingDateTime(),
		// Static unit of measurement
		UnitOfMeasurement: "metric",
		// Static customs declarable flag
		IsCustomsDeclarable: true,
		// Static estimated delivery date configuration
		EstimatedDeliveryDate: EstimatedDeliveryDate{
			IsRequested: true,
			TypeCode:    "QDDC",
		},
		// Static return standard products only flag
		ReturnStandardProductsOnly: true,
		Packages:                   a.getPackages(request),
	}

	return dhlReq
}

// convertToRatesRequestWithProductCode converts v2 request to DHL rates format with specific product code
func (a *Adapter) convertToRatesRequestWithProductCode(request *models.ServiceabilityV2Request, productCode string) RatesRequest {
	// Extract postal codes
	sourcePincode := a.getSourcePincode(request)
	destinationPincode := a.getDestinationPincode(request)

	// For international requests, we need to determine the destination country
	destinationCountry, err := a.getDestinationCountryCode(context.Background(), request, destinationPincode)
	if err != nil {
		// Log error and use default country
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": destinationPincode,
			"error":       err.Error(),
		}).Debug("Failed to get destination country code, using default")
		destinationCountry = "CN"
	}

	// Create rates request with static values and specific product code
	dhlReq := RatesRequest{
		CustomerDetails: CustomerDetails{
			ShipperDetails: ShipperDetails{
				PostalCode:  sourcePincode,
				CityName:    "Bangalore", // Enhanced with actual city
				CountryCode: "IN",        // Source is always India
			},
			ReceiverDetails: ReceiverDetails{
				PostalCode:  destinationPincode,
				CityName:    "Destination City", // Could be enhanced with actual city lookup
				CountryCode: destinationCountry,
			},
		},
		// Static account information
		Accounts: []Account{
			{
				TypeCode: "shipper",
				Number:   "533748932",
			},
		},
		// Dynamic product code but static structure
		ProductsAndServices: []ProductAndService{
			{
				ProductCode:      productCode,
				LocalProductCode: productCode,
			},
		},
		PayerCountryCode:           "IN",
		PlannedShippingDateAndTime: a.getPlannedShippingDateTime(),
		// Static unit of measurement
		UnitOfMeasurement: "metric",
		// Static customs declarable flag
		IsCustomsDeclarable: true,
		// Static estimated delivery date configuration
		EstimatedDeliveryDate: EstimatedDeliveryDate{
			IsRequested: true,
			TypeCode:    "QDDC",
		},
		// Static return standard products only flag
		ReturnStandardProductsOnly: true,
		Packages:                   a.getPackages(request),
	}

	return dhlReq
}

// convertRatesResponse converts DHL rates response to common format with capabilities
func (a *Adapter) convertRatesResponse(response *RatesResponse, partnerInfo common.PartnerInfo) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:     partnerInfo.PartnerID,
		PartnerCode:   partnerInfo.PartnerCode,
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

	// Services array is intentionally left empty as we only provide serviceability status
	// The detailed services information is not real-time and should not be included

	// Set metadata
	result.Metadata["reason"] = "DHL serviceability check completed"
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
	// TODO: Remove this once we have a proper source pincode
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

// getCountryCode extracts country code from request with default
func (a *Adapter) getCountryCode(request *models.ServiceabilityV2Request) string {
	if request.CountryCode != nil {
		return *request.CountryCode
	}
	// TODO: Remove this once we have a proper country code
	return "IN" // Default to India if no country code specified
}

// getPlannedShippingDateTime gets the planned shipping date and time
func (a *Adapter) getPlannedShippingDateTime() string {
	// Use current time + 2 days, formatted as "YYYY-MM-DDTHH:MM:SSGMT+05:30"
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("GMT+05:30", 5*60*60+30*60)
	}
	plannedTime := time.Now().In(loc).Add(48 * time.Hour)
	return plannedTime.Format("2006-01-02T15:04:05") + "GMT+05:30"
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

	// For basic health check, we only require that the adapter is enabled
	// Authentication will be checked during actual API calls
	return true
}

// GetMetrics implements PartnerAdapter interface
func (a *Adapter) GetMetrics() *common.PartnerMetrics {
	return &common.PartnerMetrics{
		PartnerCode:         "", // Will be set by orchestrator from database
		TotalRequests:       0,  // TODO: Implement actual metrics
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

	// Convert to quote request format
	quoteRequest := QuoteRequest{
		DestinationCountryCode: *request.CountryCode,
		// TODO: Remove this once we have a proper country code
		OriginCountryCode:     "IN",
		ServiceType:           "P", // Express
		DestinationPostalCode: a.getDestinationPincode(request),
		OriginPostalCode:      a.getSourcePincode(request),
		Packages:              a.convertToQuotePackages(request),
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
				// TODO: Remove this once we have a proper currency
				Currency: "INR",
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
			// TODO: Remove this once we have a proper currency
			Currency: "INR",
		},
	}
}

// getDestinationCountryCode determines the destination country code using geolocation service
func (a *Adapter) getDestinationCountryCode(ctx context.Context, request *models.ServiceabilityV2Request, destinationPincode string) (string, error) {
	// If country code is explicitly provided, use it
	if request.CountryCode != nil {
		return *request.CountryCode, nil
	}

	// If no destination pincode, return error
	if destinationPincode == "" {
		return "", fmt.Errorf("destination postal code is required to determine country code")
	}

	// Check if geolocation service is available
	if a.geolocationService == nil {
		return "", fmt.Errorf("geolocation service is not available")
	}

	// Get country code from geolocation service
	countryCode, err := a.geolocationService.GetCountryCodeByPostalCode(ctx, destinationPincode)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": destinationPincode,
			"error":       err.Error(),
		}).Debug("Failed to get destination country code, using default")
		return "", fmt.Errorf("country code not found for postal code %s: %w", destinationPincode, err)
	}

	if countryCode == nil {
		return "", fmt.Errorf("country code not found for postal code %s", destinationPincode)
	}

	return *countryCode, nil
}
