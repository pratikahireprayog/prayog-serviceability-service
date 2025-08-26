package dhl

import (
	"context"
	"fmt"
	"strconv"
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

	// Structured start log similar to pickup strategy
	partnerID := ""
	if partnerInfo.PartnerID != nil { partnerID = partnerInfo.PartnerID.String() }
	a.logger.WithFields(logrus.Fields{
		"component":    "dhl_adapter",
		"action":       "check_serviceability_start",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
	}).Info("Starting DHL adapter serviceability check")

	// Validate DHL-specific requirements
	if err := a.validateDHLRequirements(request); err != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "dhl_adapter",
			"event":        "validation_failed",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   partnerID,
			"error":        err.Error(),
		}).Warn("DHL validation failed")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("DHL validation failed: %v", err)}[0],
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
	sourcePincode := *request.SourcePostalCode
	destinationPincode := *request.DestinationPostalCode

	pid := ""
	if partnerInfo.PartnerID != nil { pid = partnerInfo.PartnerID.String() }
	a.logger.WithFields(logrus.Fields{
		"component":           "dhl_adapter",
		"step":                1,
		"partner_code":        partnerInfo.PartnerCode,
		"partner_id":          pid,
		"partner":             "DHL",
		"source_pincode":      sourcePincode,
		"destination_pincode": destinationPincode,
		"flow":                "international",
	}).Info("Starting DHL international serviceability check")

	// Step 2: Find nearest hub for source postal code
	hubLocation, err := a.findNearestHub(ctx, sourcePincode)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Hub location not found: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason":         "Hub location not found",
				"source_pincode": sourcePincode,
				"step":           "hub_lookup",
			},
		}, nil
	}

	a.logger.WithFields(logrus.Fields{
		"component":            "dhl_adapter",
		"step":                 2,
		"partner_code":         partnerInfo.PartnerCode,
		"partner_id":           pid,
		"partner":              "DHL",
		"source_pincode":       sourcePincode,
		"hub_postal_code":      hubLocation.PostalCode,
		"hub_info_postal_code": hubLocation.HubInfo.PostalCode,
		"hub_city_code":        hubLocation.HubInfo.CityCode,
	}).Info("Found nearest hub for source postal code")

	// Step 3: Get country codes for source and destination
	sourceCountryCode, destinationCountryCode, err := a.resolveCountryCodes(ctx, sourcePincode, destinationPincode)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Country code resolution failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason":              "Country code resolution failed",
				"source_pincode":      sourcePincode,
				"destination_pincode": destinationPincode,
				"step":                "country_code_resolution",
			},
		}, nil
	}

	a.logger.WithFields(logrus.Fields{
		"component":                "dhl_adapter",
		"step":                     3,
		"partner_code":             partnerInfo.PartnerCode,
		"partner_id":               pid,
		"partner":                  "DHL",
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
		"flow":                     "international",
	}).Info("Resolved country codes for international flow")

	// Step 4: Create DHL API request with dynamic values
	dhlRequest := a.createInternationalRatesRequest(ctx, request, sourceCountryCode, destinationCountryCode, hubLocation)

	a.logger.WithFields(logrus.Fields{
		"component":                "dhl_adapter",
		"step":                     4,
		"partner_code":             partnerInfo.PartnerCode,
		"partner_id":               pid,
		"product_code":             "P",
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
	}).Info("Built DHL rates request")

	// Step 5: Call DHL API
	a.logger.WithFields(logrus.Fields{
		"component":    "dhl_adapter",
		"step":         5,
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   pid,
	}).Info("Calling DHL rates API")
	response, err := a.client.CheckRates(ctx, dhlRequest)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("DHL API call failed: %v", err)}[0],
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
			a.logger.WithFields(logrus.Fields{
				"component":                "dhl_adapter",
				"step":                     6,
				"partner_code":             partnerInfo.PartnerCode,
				"partner_id":               pid,
				"source_country_code":      sourceCountryCode,
				"destination_country_code": destinationCountryCode,
			}).Info("No DHL products available for requested route")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
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
		"component":                "dhl_adapter",
		"step":                     7,
		"partner_code":             partnerInfo.PartnerCode,
		"partner_id":               pid,
		"partner":                  "DHL",
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
		"services":                 len(result.Services),
		"is_serviceable":           len(result.Services) > 0,
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

	// Validate that hub information is available
	if hubLocation.HubInfo == nil {
		return nil, fmt.Errorf("no hub found for postal code %s", postalCode)
	}

	a.logger.WithFields(logrus.Fields{
		"partner":              "DHL",
		"postal_code":          postalCode,
		"hub_postal_code":      hubLocation.PostalCode,
		"hub_info_postal_code": hubLocation.HubInfo.PostalCode,
		"hub_city_code":        hubLocation.HubInfo.CityCode,
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

// getShipperCityName gets the shipper city name from hub location info
func (a *Adapter) getShipperCityName(hubLocation *models.HubLocationInfo) string {
	if hubLocation != nil && hubLocation.HubInfo != nil && hubLocation.HubInfo.CityCode != nil {
		return *hubLocation.HubInfo.CityCode
	}
	return "Unknown City" // Fallback if hub info is not available
}

// getReceiverCityName gets the receiver city name from geolocation service using asciiname
func (a *Adapter) getReceiverCityName(ctx context.Context, postalCode string) string {
	if a.geolocationService == nil {
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": postalCode,
		}).Warn("Geolocation service not available for receiver city lookup")
		return "Unknown City"
	}

	a.logger.WithFields(logrus.Fields{
		"partner":     "DHL",
		"postal_code": postalCode,
	}).Debug("Attempting to get city name from geo_locations table")

	// Try direct asciiname lookup first
	cityName, err := a.geolocationService.GetCityNameByPostalCode(ctx, postalCode)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": postalCode,
			"error":       err.Error(),
		}).Warn("Failed to get asciiname from geo_locations table, trying fallback method")

		// Fallback to location hierarchy method
		return a.getReceiverCityNameFallback(ctx, postalCode)
	}

	if cityName == nil {
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": postalCode,
		}).Warn("City name returned as nil from geo_locations table, trying fallback method")

		// Fallback to location hierarchy method
		return a.getReceiverCityNameFallback(ctx, postalCode)
	}

	if *cityName == "" {
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": postalCode,
		}).Warn("City name returned as empty string from geo_locations table, trying fallback method")

		// Fallback to location hierarchy method
		return a.getReceiverCityNameFallback(ctx, postalCode)
	}

	a.logger.WithFields(logrus.Fields{
		"partner":     "DHL",
		"postal_code": postalCode,
		"city_name":   *cityName,
	}).Info("Successfully retrieved asciiname from geo_locations table")
	return *cityName
}

// getReceiverCityNameFallback tries to get city name using location hierarchy as fallback
func (a *Adapter) getReceiverCityNameFallback(ctx context.Context, postalCode string) string {
	a.logger.WithFields(logrus.Fields{
		"partner":     "DHL",
		"postal_code": postalCode,
	}).Debug("Using fallback method to get city name from location hierarchy")

	// Get location hierarchy to find city name as fallback
	hierarchy, err := a.geolocationService.GetLocationHierarchy(ctx, postalCode)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": postalCode,
			"error":       err.Error(),
		}).Debug("Failed to get location hierarchy for receiver city fallback")
		return "Unknown City"
	}

	// Try to get city name from different levels of hierarchy
	if hierarchy.CityName != "" {
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": postalCode,
			"city_name":   hierarchy.CityName,
			"method":      "fallback-city",
		}).Info("Retrieved city name from location hierarchy fallback")
		return hierarchy.CityName
	}
	if hierarchy.RegionName != "" {
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": postalCode,
			"city_name":   hierarchy.RegionName,
			"method":      "fallback-region",
		}).Info("Retrieved region name as city from location hierarchy fallback")
		return hierarchy.RegionName
	}
	if hierarchy.CountryName != "" {
		a.logger.WithFields(logrus.Fields{
			"partner":     "DHL",
			"postal_code": postalCode,
			"city_name":   hierarchy.CountryName,
			"method":      "fallback-country",
		}).Info("Retrieved country name as city from location hierarchy fallback")
		return hierarchy.CountryName
	}

	a.logger.WithFields(logrus.Fields{
		"partner":     "DHL",
		"postal_code": postalCode,
	}).Warn("All fallback methods failed to retrieve city name")
	return "Unknown City"
}

// createInternationalRatesRequest creates a DHL rates request with dynamic values for international flow
func (a *Adapter) createInternationalRatesRequest(ctx context.Context, request *models.ServiceabilityV2Request, sourceCountryCode, destinationCountryCode string, hubLocation *models.HubLocationInfo) RatesRequest {
	// Extract postal codes
	sourcePincode := *request.SourcePostalCode
	destinationPincode := *request.DestinationPostalCode

	a.logger.WithFields(logrus.Fields{
		"partner":                  "DHL",
		"source_pincode":           sourcePincode,
		"destination_pincode":      destinationPincode,
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
		"method":                   "createInternationalRatesRequest",
	}).Debug("Creating DHL rates request with dynamic values")

	// Get dynamic city names
	shipperCityName := a.getShipperCityName(hubLocation)
	receiverCityName := a.getReceiverCityName(ctx, destinationPincode)

	// Get hub postal code for shipper details
	shipperPostalCode := sourcePincode // fallback to source if hub info not available
	if hubLocation != nil && hubLocation.HubInfo != nil && hubLocation.HubInfo.PostalCode != nil {
		shipperPostalCode = strconv.Itoa(*hubLocation.HubInfo.PostalCode)
	}

	// Create rates request with dynamic values
	dhlReq := RatesRequest{
		CustomerDetails: CustomerDetails{
			ShipperDetails: ShipperDetails{
				PostalCode:  shipperPostalCode,
				CityName:    shipperCityName,
				CountryCode: sourceCountryCode,
			},
			ReceiverDetails: ReceiverDetails{
				PostalCode:  destinationPincode,
				CityName:    receiverCityName,
				CountryCode: destinationCountryCode,
			},
		},
		// Static account information as per business requirements
		Accounts: []Account{
			{
				TypeCode: "shipper",
				Number:   "533748932",
			},
		},
		// Static product information - hardcoded as per requirements
		ProductsAndServices: []ProductAndService{
			{
				ProductCode:      "P",
				LocalProductCode: "P",
			},
		},
		PayerCountryCode: sourceCountryCode,
		// Current time for shipping date
		PlannedShippingDateAndTime: a.getPlannedShippingDateTime(),
		// Static configuration as per business requirements
		UnitOfMeasurement:   "metric",
		IsCustomsDeclarable: true,
		EstimatedDeliveryDate: EstimatedDeliveryDate{
			IsRequested: true,
			TypeCode:    "QDDC",
		},
		ReturnStandardProductsOnly: true,
		Packages:                   a.getPackages(request),
	}

	a.logger.WithFields(logrus.Fields{
		"partner":                  "DHL",
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
		"shipper_city":             shipperCityName,
		"receiver_city":            receiverCityName,
		"product_code":             "P",
	}).Debug("Created DHL rates request with dynamic values")

	return dhlReq
}

// This function was removed as it's dead code - not called anywhere and we're using hardcoded product code "P"

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
	if len(request.Packages) == 0 {
		return fmt.Errorf("at least one package is required for DHL shipments")
	}

	// Validate each package in the array
	for i, pkg := range request.Packages {
		// Validate package weight
		if pkg.Weight == nil {
			return fmt.Errorf("package weight is required for package %d in DHL shipments", i+1)
		}

		if pkg.Weight.Value <= 0 {
			return fmt.Errorf("package weight must be greater than 0 for package %d in DHL shipments", i+1)
		}

		// Validate package dimensions
		if pkg.Dimensions == nil {
			return fmt.Errorf("package dimensions are required for package %d in DHL shipments", i+1)
		}

		if pkg.Dimensions.Length <= 0 || pkg.Dimensions.Width <= 0 || pkg.Dimensions.Height <= 0 {
			return fmt.Errorf("package dimensions must be greater than 0 for package %d in DHL shipments", i+1)
		}
	}

	return nil
}

// These functions are removed as they were duplicates and we now use createInternationalRatesRequest

// convertRatesResponse converts DHL rates response to common format with capabilities
func (a *Adapter) convertRatesResponse(response *RatesResponse, partnerInfo common.PartnerInfo) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	if len(response.Products) == 0 {
		result.Metadata["reason"] = "No DHL products available for the requested route"
		return result
	}

	// Use the first product for capabilities (DHL typically returns one main product)
	product := response.Products[0]

	// Build flattened capabilities structure as per the expected format
	capabilities := map[string]interface{}{
		// Pickup capabilities (flattened)
		"next_business_day":                          product.PickupCapabilities.NextBusinessDay,
		"local_cutoff_date_and_time":                 product.PickupCapabilities.LocalCutoffDateAndTime,
		"pickup_earliest":                            product.PickupCapabilities.PickupEarliest,
		"pickup_latest":                              product.PickupCapabilities.PickupLatest,
		"pickup_cutoff_same_day_outbound_processing": product.PickupCapabilities.PickupCutoffSameDayOutboundProcessing,
		"origin_service_area_code":                   product.PickupCapabilities.OriginServiceAreaCode,
		"origin_facility_area_code":                  product.PickupCapabilities.OriginFacilityAreaCode,
		"pickup_additional_days":                     product.PickupCapabilities.PickupAdditionalDays,
		"pickup_day_of_week":                         product.PickupCapabilities.PickupDayOfWeek,
		// Delivery capabilities (flattened)
		"delivery_type_code":               product.DeliveryCapabilities.DeliveryTypeCode,
		"estimated_delivery_date_and_time": product.DeliveryCapabilities.EstimatedDeliveryDateAndTime,
		"destination_service_area_code":    product.DeliveryCapabilities.DestinationServiceAreaCode,
		"destination_facility_area_code":   product.DeliveryCapabilities.DestinationFacilityAreaCode,
		"delivery_additional_days":         product.DeliveryCapabilities.DeliveryAdditionalDays,
		"delivery_day_of_week":             product.DeliveryCapabilities.DeliveryDayOfWeek,
		"total_transit_days":               product.DeliveryCapabilities.TotalTransitDays,
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

// getPlannedShippingDateTime gets the planned shipping date and time
func (a *Adapter) getPlannedShippingDateTime() string {
	// Use next business day at 1 PM IST to ensure DHL services are available
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("GMT+05:30", 5*60*60+30*60)
	}

	now := time.Now().In(loc)

	// Calculate next business day
	nextBusinessDay := a.getNextBusinessDay(now)

	// Set time to 1 PM (13:00) to match working curl example
	plannedTime := time.Date(
		nextBusinessDay.Year(),
		nextBusinessDay.Month(),
		nextBusinessDay.Day(),
		13, 0, 0, 0, // 1 PM
		loc,
	)

	return plannedTime.Format("2006-01-02T15:04:05") + "GMT+05:30"
}

// getNextBusinessDay calculates the next business day (Monday-Friday)
func (a *Adapter) getNextBusinessDay(from time.Time) time.Time {
	nextDay := from.Add(24 * time.Hour)

	// If it's Saturday (6), add 2 days to get Monday
	// If it's Sunday (0), add 1 day to get Monday
	for nextDay.Weekday() == time.Saturday || nextDay.Weekday() == time.Sunday {
		nextDay = nextDay.Add(24 * time.Hour)
	}

	return nextDay
}

// getPackages creates packages from request
func (a *Adapter) getPackages(request *models.ServiceabilityV2Request) []Package {
	if len(request.Packages) == 0 {
		// This should not happen as we validate package existence in validateDHLRequirements
		a.logger.Error("Package information is missing - this should have been caught in validation")
		return []Package{}
	}

	packages := make([]Package, 0, len(request.Packages))
	for _, pkg := range request.Packages {
		packages = append(packages, Package{
			Weight:     a.getWeight(pkg.Weight),
			Dimensions: a.getDimensions(pkg.Dimensions),
		})
	}

	return packages
}

// getWeight converts weight to kg
func (a *Adapter) getWeight(weight *models.Weight) float64 {
	if weight == nil {
		// This should not happen as we validate weight existence in validateDHLRequirements
		a.logger.Error("Weight information is missing - this should have been caught in validation")
		return 0.0
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
		// This should not happen as we validate dimensions existence in validateDHLRequirements
		a.logger.Error("Dimensions information is missing - this should have been caught in validation")
		return Dimensions{
			Length: 0,
			Width:  0,
			Height: 0,
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
	// Get dynamic country codes
	sourceCountryCode, destinationCountryCode, err := a.resolveCountryCodes(ctx, *request.SourcePostalCode, *request.DestinationPostalCode)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve country codes: %w", err)
	}

	quoteRequest := QuoteRequest{
		DestinationCountryCode: destinationCountryCode,
		OriginCountryCode:      sourceCountryCode,
		ServiceType:            "P", // Express
		DestinationPostalCode:  *request.DestinationPostalCode,
		OriginPostalCode:       *request.SourcePostalCode,
		Packages:               a.convertToQuotePackages(request),
	}

	return a.client.GetQuote(ctx, quoteRequest)
}

// convertToQuotePackages converts request packages to quote format
func (a *Adapter) convertToQuotePackages(request *models.ServiceabilityV2Request) []DHLPackage {
	if len(request.Packages) == 0 {
		// This should not happen as we validate package existence in validateDHLRequirements
		a.logger.Error("Package information is missing - this should have been caught in validation")
		return []DHLPackage{}
	}

	packages := make([]DHLPackage, 0, len(request.Packages))
	for _, pkg := range request.Packages {
		packages = append(packages, DHLPackage{
			Weight:        a.getWeight(pkg.Weight),
			Length:        a.convertToCm(pkg.Dimensions.Length, pkg.Dimensions.Unit),
			Width:         a.convertToCm(pkg.Dimensions.Width, pkg.Dimensions.Unit),
			Height:        a.convertToCm(pkg.Dimensions.Height, pkg.Dimensions.Unit),
			DeclaredValue: 100.0, // Default value
			// TODO: Remove this once we have a proper currency
			Currency: "INR",
		})
	}

	return packages
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
