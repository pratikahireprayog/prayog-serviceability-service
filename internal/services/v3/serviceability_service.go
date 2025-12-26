package service

import (
	"context"

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
	// Convert V3 request to V2 request format
	v2Request := request.ToV2Request()

	var partners []partner_service.PartnerInfo
	
	if len(v2Request.Partners) > 0 {
		// Partners specified in request - use them directly
		// v2Request.Partners is already set from ToV2Request(), so we can proceed
		s.logger.WithFields(logrus.Fields{
			"tenant_id":      tenantID,
			"user_id":        userID,
			"request_partners_count": len(v2Request.Partners),
		}).Info("Using partners specified in request payload")
	} else {
		// No partners in request - try to fetch from user-partners API
		// If API fails or returns empty data, fall back to default flow (all eligible partners)
		s.logger.WithFields(logrus.Fields{
			"tenant_id":      tenantID,
			"user_id":        userID,
			"partner_client": s.partnerClient != nil,
		}).Info("No partners in request, attempting to fetch user partners from partner service API")
		
		if s.partnerClient != nil {
			var err error
			partners, err = s.partnerClient.GetUserPartners(ctx, tenantID, userID)
			if err != nil {
				// API call failed - log warning and fall back to default flow
				s.logger.WithFields(logrus.Fields{
					"tenant_id": tenantID,
					"user_id":   userID,
					"error":     err.Error(),
				}).Warn("Failed to fetch user partners from partner service, falling back to default flow (all eligible partners)")
				// Continue with default flow - v2Request.Partners remains empty, so V2 orchestrator will check all eligible partners
			} else if len(partners) == 0 {
				// API returned empty data - log info and fall back to default flow
				s.logger.WithFields(logrus.Fields{
					"tenant_id": tenantID,
					"user_id":   userID,
				}).Info("Partner service API returned empty data, falling back to default flow (all eligible partners)")
				// Continue with default flow - v2Request.Partners remains empty, so V2 orchestrator will check all eligible partners
			} else {
				// API returned data - use only these partners
				s.logger.WithFields(logrus.Fields{
					"tenant_id":     tenantID,
					"user_id":       userID,
					"partners_count": len(partners),
				}).Info("Received partners from partner service API, using only these partners for serviceability")

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
			}
		} else {
			// Partner client is nil - log warning and fall back to default flow
			s.logger.WithFields(logrus.Fields{
				"tenant_id": tenantID,
				"user_id":   userID,
			}).Warn("Partner client is not configured, falling back to default flow (all eligible partners)")
			// Continue with default flow - v2Request.Partners remains empty, so V2 orchestrator will check all eligible partners
		}
	}

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

	// Enrich V3 response with capabilities from partner service
	s.enrichWithCapabilities(v3Response, partners)

	return v3Response, nil
}

// enrichWithCapabilities enriches V3 response partners with capabilities from partner service
func (s *ServiceabilityService) enrichWithCapabilities(v3Response *modelsv3.ServiceabilityV3Response, partners []partner_service.PartnerInfo) {
	if v3Response == nil || len(partners) == 0 {
		return
	}

	// Create a map of partner code to capabilities for quick lookup
	partnerCapabilitiesMap := make(map[string][]partner_service.Capability)
	for _, p := range partners {
		if p.Code != "" && len(p.Capabilities) > 0 {
			partnerCapabilitiesMap[p.Code] = p.Capabilities
		}
	}

	// Enrich each partner in the response with capabilities
	for i := range v3Response.Partners {
		partnerCode := v3Response.Partners[i].PartnerCode
		if capabilities, exists := partnerCapabilitiesMap[partnerCode]; exists {
			// Convert capabilities to array format (without id, capability_id, and code fields)
			capabilitiesList := make([]map[string]interface{}, 0, len(capabilities))
			for _, cap := range capabilities {
				capMap := map[string]interface{}{
					"name":         cap.Name,
					"category":     cap.Category,
					"is_supported": cap.IsSupported,
				}
				if len(cap.Metadata) > 0 {
					capMap["metadata"] = cap.Metadata
				}
				capabilitiesList = append(capabilitiesList, capMap)
			}
			
			// Set capabilities as a direct array
			v3Response.Partners[i].Capabilities = capabilitiesList
			
			s.logger.WithFields(logrus.Fields{
				"partner_code":      partnerCode,
				"capabilities_count": len(capabilities),
			}).Debug("Enriched partner with capabilities from partner service")
		}
	}
}
