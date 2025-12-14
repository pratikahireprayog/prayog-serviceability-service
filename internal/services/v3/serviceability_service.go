package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/infrastructure/external/partner_service"
	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/shared/errors"
	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"
	modelsv3 "prayog-serviceability-service/internal/shared/models/v3"
)

// ServiceabilityService handles V3 serviceability logic
type ServiceabilityService struct {
	v2Orchestrator orchestrators.ServiceabilityOrchestrator
	partnerClient  *partner_service.PartnerServiceClient
	logger         *logrus.Logger
}

// NewServiceabilityService creates a new V3 service
func NewServiceabilityService(
	v2Orchestrator orchestrators.ServiceabilityOrchestrator,
	partnerClient *partner_service.PartnerServiceClient,
	logger *logrus.Logger,
) *ServiceabilityService {
	return &ServiceabilityService{
		v2Orchestrator: v2Orchestrator,
		partnerClient:  partnerClient,
		logger:         logger,
	}
}

// CheckServiceability orchestrates the V3 serviceability check
func (s *ServiceabilityService) CheckServiceability(ctx context.Context, request *modelsv3.ServiceabilityV3Request, tenantID, userID string) (*modelsv3.ServiceabilityV3Response, error) {
	// Fetch user partners
	partners, err := s.partnerClient.GetUserPartners(ctx, tenantID, userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch user partners")
		return nil, fmt.Errorf("failed to fetch user partners: %w", err)
	}

	if len(partners) == 0 {
		return &modelsv3.ServiceabilityV3Response{
			Success:  true,
			Message:  "No partners found for user",
			Partners: []modelsv3.PartnerV3Response{},
		}, nil
	}

	// Convert V3 request to V2 request format
	v2Request := request.ToV2Request()

	// Map fetched partners to V2 request partners
	v2Partners := make([]modelsv1.PartnerFilter, len(partners))
	for i, p := range partners {
		var id string
		if p.ID != "" {
			id = p.ID
		}

		v2Partners[i] = modelsv1.PartnerFilter{
			ID:   &id,
			Code: p.Code,
		}
	}
	v2Request.Partners = v2Partners

	s.logger.WithFields(logrus.Fields{
		"source":         v2Request.SourcePostalCode,
		"dest":           v2Request.DestinationPostalCode,
		"partners_count": len(v2Request.Partners),
	}).Debug("Calling V2 orchestrator")

	v2Response, err := s.v2Orchestrator.CheckServiceability(ctx, v2Request)
	if err != nil {
		// If it's a known service error, return it so handler can format it
		if _, ok := err.(*errors.ServiceError); ok {
			return nil, err
		}
		s.logger.WithError(err).Error("V2 orchestrator failed")
		return nil, err
	}

	// Convert V2 response to V3 format
	v3Response := modelsv3.ServiceabilityV3ResponseFromV2(v2Response)

	return v3Response, nil
}
