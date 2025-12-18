package fedex

import (
	"context"
	"fmt"
	"time"

	services "prayog-serviceability-service/internal/services/v1/data"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	tenantcontext "prayog-serviceability-service/internal/shared/context"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/sirupsen/logrus"
)

// Adapter implements the PartnerAdapter interface for FedEx shipping
type Adapter struct {
	client             *FedExClient
	config             config.FedExConfig
	geolocationService services.GeolocationService
	hubLocationService services.HubLocationService
	logger             *logrus.Logger
}

// NewAdapter creates a new FedEx adapter instance
func NewAdapter(config config.FedExConfig, geolocationService services.GeolocationService, hubLocationService services.HubLocationService) *Adapter {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Adapter{
		client:             NewFedExClient(config),
		config:             config,
		geolocationService: geolocationService,
		hubLocationService: hubLocationService,
		logger:             logger,
	}
}

// GetAdapterType returns the adapter type
func (a *Adapter) GetAdapterType() common.AdapterType {
	return common.AdapterTypeHTTP
}

// IsEnabled returns whether the adapter is enabled
func (a *Adapter) IsEnabled() bool {
	return a.config.Enabled
}

// CheckServiceability checks if FedEx can service the given request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	partnerID := ""
	if partnerInfo.PartnerID != nil {
		partnerID = partnerInfo.PartnerID.String()
	}
	a.logger.WithFields(logrus.Fields{
		"component":    "fedex_adapter",
		"action":       "check_serviceability_start",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
	}).Info("Starting FedEx adapter serviceability check")

	// Validate FedEx-specific requirements
	if err := a.validateFedExRequirements(request); err != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "fedex_adapter",
			"event":        "validation_failed",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   partnerID,
			"error":        err.Error(),
		}).Warn("FedEx validation failed")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("FedEx validation failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "FedEx validation failed",
			},
		}, nil
	}

	// Use international flow for FedEx
	return a.checkInternationalServiceability(ctx, request, startTime, partnerInfo)
}

// checkInternationalServiceability implements the international serviceability flow for FedEx
func (a *Adapter) checkInternationalServiceability(ctx context.Context, request *models.ServiceabilityV2Request, startTime time.Time, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	// Extract postal codes from request
	sourcePincode := *request.SourcePostalCode
	destinationPincode := *request.DestinationPostalCode

	pid := ""
	if partnerInfo.PartnerID != nil {
		pid = partnerInfo.PartnerID.String()
	}
	a.logger.WithFields(logrus.Fields{
		"component":           "fedex_adapter",
		"partner_code":        partnerInfo.PartnerCode,
		"partner_id":          pid,
		"partner":             "FedEx",
		"source_pincode":      sourcePincode,
		"destination_pincode": destinationPincode,
		"flow":                "international",
	}).Info("Starting FedEx international serviceability check")

	// Set country codes from request
	sourceCountryCode := ""
	destinationCountryCode := ""

	// Use the new explicit source_country_code field if available
	if request.SourceCountryCode != nil && *request.SourceCountryCode != "" {
		sourceCountryCode = *request.SourceCountryCode
	} else if request.CountryCode != nil && *request.CountryCode != "" {
		sourceCountryCode = *request.CountryCode
	}

	// Use the new explicit destination_country_code field if available
	if request.DestinationCountryCode != nil && *request.DestinationCountryCode != "" {
		destinationCountryCode = *request.DestinationCountryCode
	} else if request.CountryCode != nil && *request.CountryCode != "" {
		destinationCountryCode = *request.CountryCode
	}

	a.logger.WithFields(logrus.Fields{
		"component":                "fedex_adapter",
		"partner_code":             partnerInfo.PartnerCode,
		"partner_id":               pid,
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
	}).Info("Using request-provided country codes")

	// Create FedEx serviceability request
	fedexRequest := a.createServiceabilityRequest(request, sourceCountryCode, destinationCountryCode)

	a.logger.WithFields(logrus.Fields{
		"component":                "fedex_adapter",
		"partner_code":             partnerInfo.PartnerCode,
		"partner_id":               pid,
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
	}).Info("Built FedEx serviceability request")

	// Extract tenant-specific credentials from context if available
	credentials, hasCredentials := tenantcontext.GetPartnerCredentials(ctx, partnerInfo.PartnerCode)
	if hasCredentials && credentials != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "fedex_adapter",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   pid,
		}).Info("Using tenant-specific credentials for FedEx API call")
	} else {
		a.logger.WithFields(logrus.Fields{
			"component":    "fedex_adapter",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   pid,
		}).Debug("Using default credentials from env for FedEx API call")
	}

	// Call FedEx API
	a.logger.WithFields(logrus.Fields{
		"component":    "fedex_adapter",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   pid,
	}).Info("Calling FedEx serviceability API")

	serviceabilityResp, err := a.client.CheckServiceability(ctx, fedexRequest, credentials)
	if err != nil {
		a.logger.WithError(err).WithFields(logrus.Fields{
			"component":    "fedex_adapter",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   pid,
		}).Warn("FedEx serviceability API call failed")

		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("FedEx API call failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "FedEx API call failed",
			},
		}, nil
	}

	// Process response
	result := a.convertServiceabilityResponse(serviceabilityResp, partnerInfo)
	result.ResponseTime = time.Since(startTime)
	result.Metadata["flow"] = "international"
	result.Metadata["source_country_code"] = sourceCountryCode
	result.Metadata["destination_country_code"] = destinationCountryCode

	// Attach full FedEx response for diagnostics
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["fedex_response"] = serviceabilityResp

	a.logger.WithFields(logrus.Fields{
		"component":                "fedex_adapter",
		"partner_code":             partnerInfo.PartnerCode,
		"partner_id":               pid,
		"partner":                  "FedEx",
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
		"is_serviceable":           serviceabilityResp.Serviceable,
		"available_services":       len(serviceabilityResp.AvailableServices),
		"flow":                     "international",
	}).Info("FedEx international serviceability check completed")

	return result, nil
}

// validateFedExRequirements validates FedEx-specific requirements
func (a *Adapter) validateFedExRequirements(request *models.ServiceabilityV2Request) error {
	// FedEx requires both source and destination postal codes
	if request.SourcePostalCode == nil || *request.SourcePostalCode == "" {
		return fmt.Errorf("source postal code is required for FedEx shipments")
	}

	if request.DestinationPostalCode == nil || *request.DestinationPostalCode == "" {
		return fmt.Errorf("destination postal code is required for FedEx shipments")
	}

	// Validate country codes
	if request.SourceCountryCode == nil || *request.SourceCountryCode == "" {
		if request.CountryCode == nil || *request.CountryCode == "" {
			return fmt.Errorf("country code is required for FedEx shipments")
		}
	}

	return nil
}

// createServiceabilityRequest creates a FedEx serviceability request
func (a *Adapter) createServiceabilityRequest(request *models.ServiceabilityV2Request, sourceCountryCode, destinationCountryCode string) ServiceabilityRequest {
	// Calculate weight from request
	return ServiceabilityRequest{
		OriginAddress: Address{
			PostalCode:  *request.SourcePostalCode,
			CountryCode: sourceCountryCode,
		},
		DestinationAddress: Address{
			PostalCode:  *request.DestinationPostalCode,
			CountryCode: destinationCountryCode,
		},
		Weight: 1,
	}
}

// convertServiceabilityResponse converts FedEx response to common format
func (a *Adapter) convertServiceabilityResponse(response *ServiceabilityResponse, partnerInfo common.PartnerInfo) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// Set basic serviceability information
	result.Metadata["is_serviceable"] = response.Serviceable
	result.Metadata["available_services_count"] = len(response.AvailableServices)
	result.Metadata["reason"] = "FedEx serviceability check completed"

	if len(response.Errors) > 0 {
		result.Metadata["errors"] = response.Errors
	}

	// If serviceable, create service entries
	if response.Serviceable {
		for _, fedexService := range response.AvailableServices {
			service := models.ServiceV2{
				ServiceName: fedexService.ServiceName,
				ServiceCode: fedexService.ServiceType,
				TATDays:     3, // Default TAT for international
				Pickup:      true,
				Delivery:    true,
				Insurance:   true,
				ProductTypes: map[string]bool{
					"document":     true,
					"non_document": true,
				},
				DeliveryModes: map[string]bool{
					"express":  true,
					"standard": fedexService.ServiceType != "FEDEX_INTERNATIONAL_PRIORITY",
				},
			}
			result.Services = append(result.Services, service)
		}
	}

	return result
}

// Initialize implements PartnerAdapter interface
func (a *Adapter) Initialize(ctx context.Context) error {
	a.logger.WithFields(logrus.Fields{
		"partner": "FedEx",
	}).Info("Initializing FedEx adapter")

	// Test authentication with minimal request
	testRequest := ServiceabilityRequest{
		OriginAddress: Address{
			PostalCode:  "10001",
			CountryCode: "US",
		},
		DestinationAddress: Address{
			PostalCode:  "90210",
			CountryCode: "US",
		},
		Weight: 0.5,
	}

	_, err := a.client.CheckServiceability(ctx, testRequest, nil)
	if err != nil {
		a.logger.WithError(err).Warn("FedEx adapter initialization test failed")
	}

	a.logger.Info("FedEx adapter initialized successfully")
	return nil
}

// IsHealthy implements PartnerAdapter interface
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	if !a.config.Enabled {
		return false
	}

	// Simple health check with minimal request
	testRequest := ServiceabilityRequest{
		OriginAddress: Address{
			PostalCode:  "10001",
			CountryCode: "US",
		},
		DestinationAddress: Address{
			PostalCode:  "90210",
			CountryCode: "US",
		},
		Weight: 0.5,
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	_, err := a.client.CheckServiceability(ctx, testRequest, nil)
	return err == nil
}

// GetMetrics implements PartnerAdapter interface
func (a *Adapter) GetMetrics() *common.PartnerMetrics {
	return &common.PartnerMetrics{
		PartnerCode:         "",
		TotalRequests:       0,
		SuccessfulRequests:  0,
		FailedRequests:      0,
		AverageResponseTime: 0,
		HealthStatus:        "healthy",
		ErrorRate:           0.0,
	}
}

// Shutdown implements PartnerAdapter interface
func (a *Adapter) Shutdown(ctx context.Context) error {
	a.logger.Info("FedEx adapter shutdown completed")
	return nil
}