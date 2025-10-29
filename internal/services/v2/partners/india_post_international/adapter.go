package india_post_international

import (
	"context"
	"fmt"
	"time"

	services "prayog-serviceability-service/internal/services/v1/data"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/sirupsen/logrus"
)

// Adapter implements the PartnerAdapter interface for India Post International shipping
type Adapter struct {
	client             *IndiaPostClient
	config             config.IndiaPostConfig
	geolocationService services.GeolocationService
	hubLocationService services.HubLocationService
	logger             *logrus.Logger
}

// NewAdapter creates a new India Post International adapter instance
func NewAdapter(config config.IndiaPostConfig, geolocationService services.GeolocationService, hubLocationService services.HubLocationService) *Adapter {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Log configuration with redacted sensitive fields
	logger.WithFields(logrus.Fields{
		"partner":  "IndiaPostInternational",
		"base_url": config.BaseURL,
		"username": redactField(config.Username),
		"enabled":  config.Enabled,
	}).Info("Creating India Post International adapter")

	return &Adapter{
		client:             NewIndiaPostClient(config),
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

// CheckServiceability checks if India Post International can service the given request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()
	
	// Extract partner ID safely
	partnerID := ""
	if partnerInfo.PartnerID != nil {
		partnerID = partnerInfo.PartnerID.String()
	}

	a.logger.WithFields(logrus.Fields{
		"component":    "india_post_international_adapter",
		"action":       "check_serviceability_start",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
	}).Info("Starting India Post International adapter serviceability check")

	// Validate requirements
	if err := a.validateRequirements(request); err != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "india_post_international_adapter",
			"event":        "validation_failed",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   partnerID,
			"error":        err.Error(),
		}).Warn("India Post International validation failed")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("India Post International validation failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "validation_failed",
			},
		}, nil
	}

	// Use international flow
	return a.checkInternationalServiceability(ctx, request, startTime, partnerInfo)
}

// checkInternationalServiceability implements the international serviceability flow
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

	// Determine destination country code
	destinationCountryCode := ""
	if request.DestinationCountryCode != nil && *request.DestinationCountryCode != "" {
		destinationCountryCode = *request.DestinationCountryCode
	} else if request.CountryCode != nil && *request.CountryCode != "" {
		destinationCountryCode = *request.CountryCode
	}

	a.logger.WithFields(logrus.Fields{
		"component":                "india_post_international_adapter",
		"step":                     1,
		"partner_code":             partnerInfo.PartnerCode,
		"partner_id":               pid,
		"source_pincode":           sourcePincode,
		"destination_pincode":      destinationPincode,
		"destination_country_code": destinationCountryCode,
		"flow":                     "international",
	}).Info("Starting India Post International serviceability check")

	// Calculate weight from packages (default to 50g if not provided)
	weight := 50 // in grams
	if request.Packages != nil && len(request.Packages) > 0 {
		totalWeight := 0.0
		for _, pkg := range request.Packages {
			if pkg.Weight != nil {
				// Convert weight to grams
				weightValue := pkg.Weight.Value
				if pkg.Weight.Unit == "kg" {
					weightValue *= 1000
				}
				totalWeight += weightValue
			}
		}
		if totalWeight > 0 {
			weight = int(totalWeight)
		}
	}

	// Build India Post tariff request
	tariffReq := TariffRequest{
		ProductType:        "FGN_LETTER", // Default product type
		Weight:             weight,
		CountryCode:        destinationCountryCode,
		Registration:       true,
		Insurance:          true,
		InsAmount:          1000,
		AdviceOfDelivery:   true,
		ModeOfTransmission: "AMS",
		SourcePincode:      sourcePincode,
	}

	a.logger.WithFields(logrus.Fields{
		"component":    "india_post_international_adapter",
		"step":         2,
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   pid,
		"request":      tariffReq,
	}).Info("Calling India Post tariff API")

	// Call tariff API
	tariffResp, err := a.client.CalculateTariff(ctx, tariffReq)
	if err != nil {
		errMsg := fmt.Sprintf("India Post tariff API failed: %v", err)
		a.logger.WithFields(logrus.Fields{
			"component":    "india_post_international_adapter",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   pid,
			"error":        err.Error(),
		}).Error("India Post tariff API call failed")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &errMsg,
			Metadata: map[string]interface{}{
				"reason": "api_call_failed",
			},
		}, nil
	}

	// Check if serviceable based on success flag inside data
	// India Post returns success in the data object, not at the root level
	isServiceable := false
	if tariffResp.Data != nil {
		if successVal, ok := tariffResp.Data["success"].(bool); ok {
			isServiceable = successVal
		}
	}

	a.logger.WithFields(logrus.Fields{
		"component":           "india_post_international_adapter",
		"step":                3,
		"partner_code":        partnerInfo.PartnerCode,
		"partner_id":          pid,
		"is_serviceable":      isServiceable,
		"message":             tariffResp.Message,
		"outer_success":       tariffResp.Success,
		"has_tariff_data":     tariffResp.Data != nil,
	}).Info("India Post serviceability result")

	// Build result
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		ResponseTime: time.Since(startTime),
		Services:     make([]models.ServiceV2, 0),
		Metadata: map[string]interface{}{
			"flow":                     "international",
			"destination_country_code": destinationCountryCode,
			"is_serviceable":           isServiceable,
			"message":                  tariffResp.Message,
			"tariff_data":              tariffResp.Data,
		},
	}

	// If serviceable, add service details
	if isServiceable {
		serviceName := "India Post International"
		// Try to extract product name from tariff data
		if tariffResp.Data != nil {
			if productName, ok := tariffResp.Data["productName"].(string); ok && productName != "" {
				serviceName = productName
			}
		}

		service := models.ServiceV2{
			ServiceName: serviceName,
			TATDays:     7, // Default TAT for international
			Pickup:      false,
			Delivery:    true,
			Insurance:   true,
			ProductTypes: map[string]bool{
				"document":     true,
				"non_document": true,
			},
			DeliveryModes: map[string]bool{
				"standard": true,
			},
		}
		result.Services = append(result.Services, service)

		a.logger.WithFields(logrus.Fields{
			"component":    "india_post_international_adapter",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   pid,
			"service_name": serviceName,
		}).Info("Created service for India Post International")
	}

	return result, nil
}

// validateRequirements validates India Post International-specific requirements
func (a *Adapter) validateRequirements(request *models.ServiceabilityV2Request) error {
	// India Post requires source and destination postal codes
	if request.SourcePostalCode == nil || *request.SourcePostalCode == "" {
		return fmt.Errorf("source postal code is required for India Post International shipments")
	}

	if request.DestinationPostalCode == nil || *request.DestinationPostalCode == "" {
		return fmt.Errorf("destination postal code is required for India Post International shipments")
	}

	// Validate destination country code
	if request.DestinationCountryCode == nil || *request.DestinationCountryCode == "" {
		if request.CountryCode == nil || *request.CountryCode == "" {
			return fmt.Errorf("destination country code is required for India Post International shipments")
		}
	}

	return nil
}

// Initialize implements PartnerAdapter interface
func (a *Adapter) Initialize(ctx context.Context) error {
	// Pre-authenticate to verify credentials
	return a.client.Authenticate(ctx)
}

// IsHealthy implements PartnerAdapter interface
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	// Only check if enabled, don't authenticate during health check
	// Authentication will happen lazily when needed
	return a.config.Enabled
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

