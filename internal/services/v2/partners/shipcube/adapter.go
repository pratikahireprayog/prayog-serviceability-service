package shipcube

import (
	"context"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	models "prayog-serviceability-service/internal/shared/models/v1"

	"github.com/sirupsen/logrus"
)

type Adapter struct {
	client *ShipCubeClient
	config config.ShipCubeConfig
	logger *logrus.Logger
}

func NewAdapter(cfg config.ShipCubeConfig) *Adapter {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Adapter{
		client: NewShipcubeClient(cfg),
		config: cfg,
		logger: logger,
	}
}

// Initialize implements PartnerAdapter interface
func (a *Adapter) Initialize(ctx context.Context) error {
	// No authentication required for ShipCube
	return nil
}

// IsHealthy implements PartnerAdapter interface
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	if !a.config.Enabled {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// Optional: perform lightweight ping here
	return true
}

// Shutdown implements PartnerAdapter interface
func (a *Adapter) Shutdown(ctx context.Context) error {
	// No cleanup required
	return nil
}

// CheckServiceability implements PartnerAdapter interface
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

	// --- Assume validation already handled in strategy ---
 	destPin := req.DestinationCountryCode

	// Mock ShipCube response since no direct API for now
	service := models.ServiceV2{
		ServiceCode: "STANDARD",
		ServiceName: "Standard Delivery",
	}

	a.logger.WithFields(logrus.Fields{
		"component": "shipcube_adapter",
		"zip":       destPin,
	}).Info("Returning static serviceability result (mock)")

	return &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		ResponseTime: time.Since(start),
		Services:     []models.ServiceV2{service},
		Metadata: map[string]interface{}{
			"validated_zip": destPin,
			"note":          "validated at strategy layer",
		},
	}, nil
}

// GetMetrics implements PartnerAdapter interface
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
