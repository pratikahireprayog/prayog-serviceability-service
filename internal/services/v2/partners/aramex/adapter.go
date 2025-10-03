package aramex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	services "prayog-serviceability-service/internal/services/v1/data"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/sirupsen/logrus"
)

// Adapter implements the PartnerAdapter interface for Aramex shipping
type Adapter struct {
	client             *AramexClient
	config             config.AramexConfig
	geolocationService services.GeolocationService
	hubLocationService services.HubLocationService
	logger             *logrus.Logger
}

// NewAdapter creates a new Aramex adapter instance
func NewAdapter(config config.AramexConfig, geolocationService services.GeolocationService, hubLocationService services.HubLocationService) *Adapter {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Log configuration with redacted sensitive fields
	logger.WithFields(logrus.Fields{
		"partner":        "Aramex",
		"base_url":       config.BaseURL,
		"username":       redactField(config.Username),
		"enabled":        config.Enabled,
		"account_number": redactField(config.AccountNumber),
	}).Info("Creating Aramex adapter")

	return &Adapter{
		client:             NewAramexClient(config),
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

// CheckServiceability checks if Aramex can service the given request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()
	// Structured start log
	partnerID := ""
	if partnerInfo.PartnerID != nil {
		partnerID = partnerInfo.PartnerID.String()
	}
	a.logger.WithFields(logrus.Fields{
		"component":    "aramex_adapter",
		"action":       "check_serviceability_start",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
	}).Info("Starting Aramex adapter serviceability check")

	// Validate Aramex-specific requirements
	if err := a.validateAramexRequirements(request); err != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "aramex_adapter",
			"event":        "validation_failed",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   partnerID,
			"error":        err.Error(),
		}).Warn("Aramex validation failed")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Aramex validation failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "Aramex validation failed",
			},
		}, nil
	}

	// Use international flow for Aramex
	return a.checkInternationalServiceability(ctx, request, startTime, partnerInfo)
}

// checkInternationalServiceability implements the international serviceability flow for Aramex
func (a *Adapter) checkInternationalServiceability(ctx context.Context, request *models.ServiceabilityV2Request, startTime time.Time, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	// Step 1: Extract postal codes from request
	sourcePincode := *request.SourcePostalCode
	destinationPincode := *request.DestinationPostalCode

	pid := ""
	if partnerInfo.PartnerID != nil {
		pid = partnerInfo.PartnerID.String()
	}
	a.logger.WithFields(logrus.Fields{
		"component":           "aramex_adapter",
		"step":                1,
		"partner_code":        partnerInfo.PartnerCode,
		"partner_id":          pid,
		"partner":             "Aramex",
		"source_pincode":      sourcePincode,
		"destination_pincode": destinationPincode,
		"flow":                "international",
	}).Info("Starting Aramex international serviceability check")

	// Step 2: Set country codes from request
	sourceCountryCode := "IN"      // Default to India
	destinationCountryCode := "US" // Default to USA

	// Use the new explicit source_country_code field if available
	if request.SourceCountryCode != nil && *request.SourceCountryCode != "" {
		sourceCountryCode = *request.SourceCountryCode
	} else if request.CountryCode != nil && *request.CountryCode != "" {
		// Fallback to generic country_code for source
		sourceCountryCode = *request.CountryCode
	}

	// Use the new explicit destination_country_code field if available
	if request.DestinationCountryCode != nil && *request.DestinationCountryCode != "" {
		destinationCountryCode = *request.DestinationCountryCode
	} else if request.CountryCode != nil && *request.CountryCode != "" {
		// Fallback to generic country_code for destination
		destinationCountryCode = *request.CountryCode
	}

	a.logger.WithFields(logrus.Fields{
		"component":                "aramex_adapter",
		"step":                     2,
		"partner_code":             partnerInfo.PartnerCode,
		"partner_id":               pid,
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
	}).Info("Using request-provided country codes")

	// Step 3: Create Aramex API request
	aramexRequest := a.createServiceabilityRequest(ctx, request, sourceCountryCode, destinationCountryCode)
	a.logger.WithFields(logrus.Fields{	
		"aramexRequest": aramexRequest,
	}).Info("Before Hitting aramex api")

	a.logger.WithFields(logrus.Fields{
		"component":                "aramex_adapter",
		"step":                     3,
		"partner_code":             partnerInfo.PartnerCode,
		"partner_id":               pid,
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
	}).Info("Built Aramex serviceability request")

	// Log full Aramex request payload at INFO level for diagnostics
	if payload, err := json.Marshal(aramexRequest); err == nil {
		a.logger.WithFields(logrus.Fields{
			"component":      "aramex_adapter",
			"step":           3,
			"partner_code":   partnerInfo.PartnerCode,
			"partner_id":     pid,
			"aramex_request": string(payload),
		}).Info("Aramex serviceability request payload")
	} else {
		a.logger.WithFields(logrus.Fields{
			"component":    "aramex_adapter",
			"step":         3,
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   pid,
			"error":        err.Error(),
		}).Warn("Failed to marshal Aramex request payload for logging")
	}

	// Step 4: Call Aramex API
	a.logger.WithFields(logrus.Fields{
		"component":    "aramex_adapter",
		"step":         4,
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   pid,
	}).Info("Calling Aramex serviceability API")

	var response *ServiceabilityResponse
	{
		url := a.config.BaseURL + "/ShippingAPI.V2/Location/Service_1_0.svc/json/IsAddressServiced"
		body, mErr := json.Marshal(aramexRequest)
		if mErr != nil {
			return &common.PartnerServiceabilityResult{
				PartnerID:    partnerInfo.PartnerID,
				PartnerCode:  partnerInfo.PartnerCode,
				ResponseTime: time.Since(startTime),
				Error:        mErr,
				ErrorMessage: &[]string{fmt.Sprintf("Failed to marshal Aramex request: %v", mErr)}[0],
			}, nil
		}

		req, rErr := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
		if rErr != nil {
			return &common.PartnerServiceabilityResult{
				PartnerID:    partnerInfo.PartnerID,
				PartnerCode:  partnerInfo.PartnerCode,
				ResponseTime: time.Since(startTime),
				Error:        rErr,
				ErrorMessage: &[]string{fmt.Sprintf("Failed to create Aramex request: %v", rErr)}[0],
			}, nil
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, doErr := a.client.httpClient.Do(req)
		if doErr != nil {
			return &common.PartnerServiceabilityResult{
				PartnerID:    partnerInfo.PartnerID,
				PartnerCode:  partnerInfo.PartnerCode,
				ResponseTime: time.Since(startTime),
				Error:        doErr,
				ErrorMessage: &[]string{fmt.Sprintf("Aramex request failed: %v", doErr)}[0],
			}, nil
		}
		defer resp.Body.Close()

		respBody, rdErr := io.ReadAll(resp.Body)
		if rdErr != nil {
			return &common.PartnerServiceabilityResult{
				PartnerID:    partnerInfo.PartnerID,
				PartnerCode:  partnerInfo.PartnerCode,
				ResponseTime: time.Since(startTime),
				Error:        rdErr,
				ErrorMessage: &[]string{fmt.Sprintf("Failed to read Aramex response: %v", rdErr)}[0],
			}, nil
		}

		if resp.StatusCode != http.StatusOK {
			apiErr := &AramexAPIError{
				StatusCode: resp.StatusCode,
				RawBody:    string(respBody),
			}
			return &common.PartnerServiceabilityResult{
				PartnerID:    partnerInfo.PartnerID,
				PartnerCode:  partnerInfo.PartnerCode,
				ResponseTime: time.Since(startTime),
				Error:        apiErr,
				ErrorMessage: &[]string{fmt.Sprintf("Aramex API call failed: %v", apiErr)}[0],
				Metadata:     map[string]interface{}{"status_code": resp.StatusCode, "raw_body": string(respBody)},
			}, nil
		}

		var parsed ServiceabilityResponse
		if uErr := json.Unmarshal(respBody, &parsed); uErr != nil {
			return &common.PartnerServiceabilityResult{
				PartnerID:    partnerInfo.PartnerID,
				PartnerCode:  partnerInfo.PartnerCode,
				ResponseTime: time.Since(startTime),
				Error:        uErr,
				ErrorMessage: &[]string{fmt.Sprintf("Failed to decode Aramex response: %v", uErr)}[0],
			}, nil
		}
		response = &parsed
	}

	// Step 5: Process response
	result := a.convertServiceabilityResponse(response, partnerInfo)
	result.ResponseTime = time.Since(startTime)
	result.Metadata["flow"] = "international"
	result.Metadata["source_country_code"] = sourceCountryCode
	result.Metadata["destination_country_code"] = destinationCountryCode

	// Attach full Aramex response for diagnostics
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["aramex_response"] = response

	a.logger.WithFields(logrus.Fields{
		"component":                "aramex_adapter",
		"step":                     5,
		"partner_code":             partnerInfo.PartnerCode,
		"partner_id":               pid,
		"partner":                  "Aramex",
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
		"is_serviceable":           response.IsAddressServiced,
		"flow":                     "international",
	}).Info("Aramex international serviceability check completed")

	return result, nil
}

// validateAramexRequirements validates Aramex-specific requirements
func (a *Adapter) validateAramexRequirements(request *models.ServiceabilityV2Request) error {
	// Aramex requires both source and destination postal codes
	if request.SourcePostalCode == nil || *request.SourcePostalCode == "" {
		return fmt.Errorf("source postal code is required for Aramex shipments")
	}

	if request.DestinationPostalCode == nil || *request.DestinationPostalCode == "" {
		return fmt.Errorf("destination postal code is required for Aramex shipments")
	}

	// Validate country codes
	if request.SourceCountryCode == nil || *request.SourceCountryCode == "" {
		if request.CountryCode == nil || *request.CountryCode == "" {
			return fmt.Errorf("country code is required for Aramex shipments")
		}
	}

	return nil
}

// createServiceabilityRequest creates an Aramex serviceability request
func (a *Adapter) createServiceabilityRequest(ctx context.Context, request *models.ServiceabilityV2Request, sourceCountryCode, destinationCountryCode string) ServiceabilityRequest {
	// Get city name for destination
	destinationCity := a.getCityName(ctx, *request.DestinationPostalCode)

	return ServiceabilityRequest{
		ClientInfo: ClientInfo{
			UserName:           a.config.Username,
			Password:           a.config.Password,
			Version:            "v1.0",
			AccountNumber:      a.config.AccountNumber,
			AccountPin:         a.config.AccountPin,
			AccountEntity:      a.config.AccountEntity,
			AccountCountryCode: a.config.AccountCountryCode,
			Source:             a.config.Source,
		},
		Address: Address{
			Line1:               "", 
			Line2:               "",     
			Line3:               "",           
			City:                destinationCity,
			StateOrProvinceCode: "",
			PostCode:            *request.DestinationPostalCode,
			CountryCode:         destinationCountryCode,
			Longitude:           0,
			Latitude:            0,
			BuildingNumber:      nil,
			BuildingName:        nil,
			Floor:               nil,
			Apartment:           nil,
			POBox:               nil,
			Description:         nil,
		},
		ServiceDetails: ServiceDetails{
			// ProductGroup: "EXP", // Express
			// ProductType:  "PDX", // Priority Document Express
			ServiceMode:  1,
		},
		Transaction: Transaction{
			Reference1: "",
			Reference2: "",
			Reference3: "",
			Reference4: "",
			Reference5: "",
		},
	}
}

// getCityName gets the city name from geolocation service
func (a *Adapter) getCityName(ctx context.Context, postalCode string) string {
	if a.geolocationService == nil {
		a.logger.WithFields(logrus.Fields{
			"partner":     "Aramex",
			"postal_code": postalCode,
		}).Warn("Geolocation service not available for city lookup")
		return "Unknown City"
	}

	cityName, err := a.geolocationService.GetCityNameByPostalCode(ctx, postalCode)
	if err != nil || cityName == nil || *cityName == "" {
		a.logger.WithFields(logrus.Fields{
			"partner":     "Aramex",
			"postal_code": postalCode,
			"error":       err,
		}).Warn("Failed to get city name, using fallback")
		return "Unknown City"
	}

	return *cityName
}

// convertServiceabilityResponse converts Aramex response to common format
func (a *Adapter) convertServiceabilityResponse(response *ServiceabilityResponse, partnerInfo common.PartnerInfo) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// Set basic serviceability information
	result.Metadata["is_serviceable"] = response.IsAddressServiced
	result.Metadata["has_errors"] = response.HasErrors
	result.Metadata["reason"] = "Aramex serviceability check completed"

	if response.HasErrors {
		result.Metadata["errors"] = response.Errors
	}
	fmt.Print("RSULR of arameexxxx", response);
	// If address is serviced, create a basic service entry
	if response.IsAddressServiced {
		service := models.ServiceV2{
			// ServiceID:   "ARAMEX_EXP",
			ServiceName: "Aramex Express",
			// ServiceType: "EXPRESS",
			// Carrier:     "ARAMEX",
			TATDays:   3, // Default TAT for international
			Pickup:    true,
			Delivery:  true,
			Insurance: true,
			ProductTypes: map[string]bool{
				"document":     true,
				"non_document": true,
				"commercial":   true,
			},
			DeliveryModes: map[string]bool{
				"express":  true,
				"standard": false,
			},
		}
		result.Services = append(result.Services, service)
	}

	return result
}

// Initialize implements PartnerAdapter interface
func (a *Adapter) Initialize(ctx context.Context) error {
	// No authentication required for Aramex serviceability API
	return nil
}

// IsHealthy implements PartnerAdapter interface
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	if !a.config.Enabled {
		return false
	}

	// Simple health check - try to make a test request
	testRequest := ServiceabilityRequest{
		ClientInfo: ClientInfo{
			UserName:           a.config.Username,
			Password:           a.config.Password,
			Version:            "v1.0",
			AccountNumber:      a.config.AccountNumber,
			AccountPin:         a.config.AccountPin,
			AccountEntity:      a.config.AccountEntity,
			AccountCountryCode: a.config.AccountCountryCode,
			Source:             a.config.Source,
		},
		Address: Address{
			Line1:       "Test Address",
			City:        "Amman",
			PostCode:    "11821",
			CountryCode: "JO",
		},
		ServiceDetails: ServiceDetails{
			ProductGroup: "EXP",
			ProductType:  "PDX",
			ServiceMode:  1,
		},
	}

	// Make a quick test call (with timeout)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := a.client.CheckServiceability(ctx, testRequest)
	return err == nil
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
