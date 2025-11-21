package shipcube

import (
	"context"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	models "prayog-serviceability-service/internal/shared/models/v1"

	services "prayog-serviceability-service/internal/services/v1/data"

	"github.com/sirupsen/logrus"
)

type Adapter struct {
	client             *ShipCubeClient
	config             config.ShipCubeConfig
	geolocationService services.GeolocationService
	hubLocationService services.HubLocationService
	logger             *logrus.Logger
}

func NewAdapter(config config.ShipCubeConfig, geolocationService services.GeolocationService, hubLocationService services.HubLocationService) *Adapter {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.WithFields(logrus.Fields{
		"partner":  "shipcube",
		"base_url": config.BaseURL,
		"enabled":  config.Enabled,
	}).Info("Creating ShipCube adapter")

	return &Adapter{
		client:             NewShipcubeClient(config),
		config:             config,
		geolocationService: geolocationService,
		hubLocationService: hubLocationService,
		logger:             logger,
	}
}

func (a *Adapter) Initialize(ctx context.Context) error {
	return nil
}

func (a *Adapter) IsHealthy(ctx context.Context) bool {
	if !a.config.Enabled {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return true
}

func (a *Adapter) Shutdown(ctx context.Context) error {
	return nil
}

func (a *Adapter) CheckServiceability(
	ctx context.Context,
	req *models.ServiceabilityV2Request,
	partnerInfo common.PartnerInfo,
) (*common.PartnerServiceabilityResult, error) {

	start := time.Now()
	partnerID := ""
	if partnerInfo.PartnerID != nil {
		partnerID = partnerInfo.PartnerID.String()
	}

	a.logger.WithFields(logrus.Fields{
		"component":    "shipcube_adapter",
		"action":       "check_serviceability_start",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
	}).Info("Starting ShipCube adapter serviceability check")

	// --- Step 1: Validate destination ZIP using geolocationService ---
	// ShipCube requires a destination postal code to validate serviceability
	if req.DestinationPostalCode == nil || *req.DestinationPostalCode == "" {
		errMsg := "Destination postal code is required for ShipCube serviceability check"
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(start),
			Services:     []models.ServiceV2{},
			ErrorMessage: &errMsg,
			Metadata: map[string]interface{}{
				"reason": "Destination postal code is required",
			},
		}, nil
	}

	if a.geolocationService == nil {
		a.logger.Warn("geolocationService is nil — skipping validation")
	} else {
		countryCode, err := a.geolocationService.GetCountryCodeByPostalCode(ctx, *req.DestinationPostalCode)
		if err != nil {
			// Geolocation lookup failed - cannot validate postal code, so ShipCube is not serviceable
			a.logger.WithError(err).Warn("Failed to get country code for destination ZIP")
			errMsg := "Failed to validate postal code: " + err.Error()
			return &common.PartnerServiceabilityResult{
				PartnerID:    partnerInfo.PartnerID,
				PartnerCode:  partnerInfo.PartnerCode,
				ResponseTime: time.Since(start),
				Services:     []models.ServiceV2{},
				ErrorMessage: &errMsg,
				Metadata: map[string]interface{}{
					"validated_zip": *req.DestinationPostalCode,
					"reason":        "Geolocation lookup failed, cannot validate postal code",
					"error":         err.Error(),
				},
			}, nil
		} else if countryCode == nil || *countryCode != "US" {
			a.logger.WithFields(logrus.Fields{
				"postal_code": *req.DestinationPostalCode,
				"country":     countryCode,
			}).Warn("Destination postal code is not US — skipping ShipCube")
			errMsg := "ShipCube only services US destinations"
			return &common.PartnerServiceabilityResult{
				PartnerID:    partnerInfo.PartnerID,
				PartnerCode:  partnerInfo.PartnerCode,
				ResponseTime: time.Since(start),
				Services:     []models.ServiceV2{},
				ErrorMessage: &errMsg,
				Metadata: map[string]interface{}{
					"validated_zip": *req.DestinationPostalCode,
					"country_code":  countryCode,
					"reason":        "Non-US postal code, ShipCube only services US destinations",
				},
			}, nil
		}
	}

	service := models.ServiceV2{
		ServiceCode: "STANDARD",
		ServiceName: "Standard Deliver",
		Pickup:      true,
		Delivery:    true,
		Insurance:   true,
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



	a.logger.WithFields(logrus.Fields{
		"component": "shipcube_adapter",
		"zip":       req.DestinationPostalCode,
	}).Info("Returning static ShipCube serviceability result")
	return &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		ResponseTime: time.Since(start),
		Services:     []models.ServiceV2{service},
		Metadata: map[string]interface{}{
			"validated_zip": req.DestinationPostalCode,
			"note":          "validated using geolocationService",
		},
	}, nil
}

func (a *Adapter) GetMetrics() *common.PartnerMetrics {
	return &common.PartnerMetrics{
		PartnerCode:         "shipcube",
		TotalRequests:       0,
		SuccessfulRequests:  0,
		FailedRequests:      0,
		AverageResponseTime: 0,
		HealthStatus:        "healthy",
		ErrorRate:           0.0,
	}
}
