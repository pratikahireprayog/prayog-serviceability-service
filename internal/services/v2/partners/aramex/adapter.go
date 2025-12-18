package aramex

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	services "prayog-serviceability-service/internal/services/v1/data"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	tenantcontext "prayog-serviceability-service/internal/shared/context"
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
type AramexServiceabilityResp struct {
    XMLName   xml.Name `xml:"AddressServiceabilityResponse"`
    HasErrors bool     `xml:"HasErrors"`
    IsServiced bool    `xml:"IsServiced"`
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
func (a *Adapter) checkInternationalServiceability(
	ctx context.Context,
	request *models.ServiceabilityV2Request,
	startTime time.Time,
	partnerInfo common.PartnerInfo,
) (*common.PartnerServiceabilityResult, error) {
	// Safe extraction of postal codes
	sourcePincode := ""
	if request.SourcePostalCode != nil {
		sourcePincode = *request.SourcePostalCode
	}
	destinationPincode := ""
	if request.DestinationPostalCode != nil {
		destinationPincode = *request.DestinationPostalCode
	}

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

	// Determine country codes safely
	sourceCountryCode := ""
	if request.SourceCountryCode != nil && *request.SourceCountryCode != "" {
		sourceCountryCode = *request.SourceCountryCode
	} else if request.CountryCode != nil && *request.CountryCode != "" {
		sourceCountryCode = *request.CountryCode
	}

	destinationCountryCode := ""
	if request.DestinationCountryCode != nil && *request.DestinationCountryCode != "" {
		destinationCountryCode = *request.DestinationCountryCode
	} else if request.CountryCode != nil && *request.CountryCode != "" {
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

	// Extract tenant-specific credentials from context if available
	credentials, hasCredentials := tenantcontext.GetPartnerCredentials(ctx, partnerInfo.PartnerCode)
	if hasCredentials && credentials != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "aramex_adapter",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   pid,
		}).Info("Using tenant-specific credentials for Aramex API call")
	} else {
		a.logger.WithFields(logrus.Fields{
			"component":    "aramex_adapter",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   pid,
		}).Debug("Using default credentials from env for Aramex API call")
	}

	// Build Aramex request
	aramexRequest := a.createServiceabilityRequest(ctx, request, sourceCountryCode, destinationCountryCode, credentials)

	url := a.config.BaseURL + "/ShippingAPI.V2/Location/Service_1_0.svc/json/IsAddressServiced"
	bodyBytes, err := json.Marshal(aramexRequest)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to marshal Aramex request: %v", err)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			ErrorMessage: &errMsg,
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create Aramex request: %v", err)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			ErrorMessage: &errMsg,
		}, nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/xml") // Aramex responds in XML

	resp, err := a.client.httpClient.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("Aramex request failed: %v", err)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			ErrorMessage: &errMsg,
		}, nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read Aramex response: %v", err)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			ErrorMessage: &errMsg,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Aramex API returned status %d: %s", resp.StatusCode, string(respBody))
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			ErrorMessage: &errMsg,
			Metadata: map[string]interface{}{
				"status_code": resp.StatusCode,
				"raw_body":    string(respBody),
			},
		}, nil
	}

	// Parse XML response
	var arResp AramexServiceabilityResp
	if err := xml.Unmarshal(respBody, &arResp); err != nil {
		errMsg := fmt.Sprintf("Failed to parse Aramex XML response: %v", err)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			ErrorMessage: &errMsg,
		}, nil
	}

	a.logger.WithField("aramex_response", arResp).Info("Aramex serviceability response received")

	// Build PartnerServiceabilityResult
	result := &common.PartnerServiceabilityResult{
		PartnerID:     partnerInfo.PartnerID,
		PartnerCode:   partnerInfo.PartnerCode,
		ResponseTime:  time.Since(startTime),
		Metadata: map[string]interface{}{
			"flow":                     "international",
			"source_country_code":      sourceCountryCode,
			"destination_country_code": destinationCountryCode,
			"aramex_response":          arResp,
			"is_serviceable":           arResp.IsServiced, 
		},
	}

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
// credentials is optional - if provided, uses tenant-specific credentials, otherwise uses default config
func (a *Adapter) createServiceabilityRequest(ctx context.Context, request *models.ServiceabilityV2Request, sourceCountryCode, destinationCountryCode string, credentials map[string]string) ServiceabilityRequest {
	// Use tenant credentials if provided, otherwise fallback to default config
	username := a.config.Username
	password := a.config.Password
	accountNumber := a.config.AccountNumber
	accountPin := a.config.AccountPin
	accountEntity := a.config.AccountEntity
	accountCountryCode := a.config.AccountCountryCode
	source := a.config.Source

	if credentials != nil {
		if u, ok := credentials["username"]; ok && u != "" {
			username = u
		}
		if p, ok := credentials["password"]; ok && p != "" {
			password = p
		}
		if an, ok := credentials["account_number"]; ok && an != "" {
			accountNumber = an
		}
		if ap, ok := credentials["account_pin"]; ok && ap != "" {
			accountPin = ap
		}
		if ae, ok := credentials["account_entity"]; ok && ae != "" {
			accountEntity = ae
		}
		if acc, ok := credentials["account_country_code"]; ok && acc != "" {
			accountCountryCode = acc
		}
		if s, ok := credentials["source"]; ok && s != "" {
			if sourceInt, err := strconv.Atoi(s); err == nil {
				source = sourceInt
			}
		}
	}

	// Get city name for destination
	destinationCity := ""
	return ServiceabilityRequest{
		ClientInfo: ClientInfo{
			UserName:           username,
			Password:           password,
			Version:            "v1.0",
			AccountNumber:      accountNumber,
			AccountPin:         accountPin,
			AccountEntity:      accountEntity,
			AccountCountryCode: accountCountryCode,
			Source:             source,
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
			ProductGroup: "EXP",
			ProductType:  "PPX",
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
