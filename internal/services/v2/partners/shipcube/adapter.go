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
	if a.geolocationService == nil {
		a.logger.Warn("geolocationService is nil — skipping validation")
	} else if req.DestinationPostalCode != nil {
		countryCode, err := a.geolocationService.GetCountryCodeByPostalCode(ctx, *req.DestinationPostalCode)
		a.logger.Info("countryCode", countryCode);
		if err != nil {
			a.logger.WithError(err).Warn("Failed to get country code for destination ZIP")
		} else if countryCode == nil || *countryCode != "US" {
			a.logger.WithFields(logrus.Fields{
				"postal_code": *req.DestinationPostalCode,
				"country":     countryCode,
			}).Warn("Destination postal code is not US — skipping ShipCube")
			return &common.PartnerServiceabilityResult{
				PartnerID:    partnerInfo.PartnerID,
				PartnerCode:  partnerInfo.PartnerCode,
				ResponseTime: time.Since(start),
				Services:     []models.ServiceV2{},
				Metadata: map[string]interface{}{
					"validated_zip": *req.DestinationPostalCode,
					"note":          "Non-US postal code, skipped ShipCube",
				},
			}, nil
		}
	}

	// --- Step 2: Mock response for now (since ShipCube API not ready) ---
	service := models.ServiceV2{
		ServiceCode: "STANDARD",
		ServiceName: "Standard Delivery",
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
