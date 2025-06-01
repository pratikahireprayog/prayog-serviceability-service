package services

import (
	"context"
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/infrastructure/api/http/outbound"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// PartnerIntegrationService implements business logic for Partner Service integration
type PartnerIntegrationService struct {
	transport *outbound.HTTPTransport
	config    config.PartnerServiceConfig
}

// NewPartnerIntegrationService creates a new Partner Service integration
func NewPartnerIntegrationService(config config.PartnerServiceConfig) interfaces.PartnerServiceClient {
	transport := outbound.NewHTTPTransport(config.ToHTTPTransportConfig())

	return &PartnerIntegrationService{
		transport: transport,
		config:    config,
	}
}

// GetPartnersByLocation fetches partners based on location criteria
// Implements interfaces.PartnerServiceClient interface method
func (s *PartnerIntegrationService) GetPartnersByLocation(ctx context.Context, locationType, locationID string) ([]interfaces.PartnerInfo, error) {
	// Build request path with query parameters
	path := s.config.Endpoints.GetPartnersByLocation
	req := outbound.Request{
		Method: "GET",
		Path:   path,
		Query: map[string]string{
			"location_type": locationType,
			"location_id":   locationID,
		},
	}

	// Make the request
	var response dtos.PartnersLocationResponse
	if err := s.transport.Execute(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to get partners by location: %w", err)
	}

	// Convert response to interface models
	partners := make([]interfaces.PartnerInfo, len(response.Partners))
	for i, partnerResp := range response.Partners {
		partners[i] = s.mapPartnerSummaryToInfo(partnerResp)
	}

	return partners, nil
}

// GetPartnerEffectiveDetails fetches partner effective details by partner ID and entity
// Implements interfaces.PartnerServiceClient interface method
func (s *PartnerIntegrationService) GetPartnerEffectiveDetails(ctx context.Context, partnerID uint, entityType, entityID string) (*interfaces.PartnerEffectiveDetails, error) {
	// Build path and request
	path := fmt.Sprintf("/api/v1/partners/%d/effective-details", partnerID)
	req := outbound.Request{
		Method: "GET",
		Path:   path,
		Query: map[string]string{
			"entity_type": entityType,
			"entity_id":   entityID,
		},
	}

	var response struct {
		PartnerID   uint                          `json:"partner_id"`
		Preferences []EffectivePreferenceResponse `json:"preferences"`
		Ratings     []EffectiveRatingResponse     `json:"ratings"`
	}

	if err := s.transport.Execute(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to get partner effective details for partner %d: %w", partnerID, err)
	}

	// Convert to interface model
	result := &interfaces.PartnerEffectiveDetails{
		PartnerID:   response.PartnerID,
		Preferences: make([]interfaces.EffectivePreference, len(response.Preferences)),
		Ratings:     make([]interfaces.EffectiveRating, len(response.Ratings)),
	}

	for i, pref := range response.Preferences {
		result.Preferences[i] = interfaces.EffectivePreference{
			EntityType:      pref.EntityType,
			EntityID:        pref.EntityID,
			PreferenceValue: pref.PreferenceValue,
			EffectiveRating: pref.EffectiveRating,
		}
	}

	for i, rating := range response.Ratings {
		result.Ratings[i] = interfaces.EffectiveRating{
			EntityType: rating.EntityType,
			EntityID:   rating.EntityID,
			Rating:     rating.Rating,
		}
	}

	return result, nil
}

// Additional business methods (not part of interface but useful for internal operations)

// GetPartnerCapabilities fetches partner capabilities by partner ID
func (s *PartnerIntegrationService) GetPartnerCapabilities(ctx context.Context, partnerID string) (*dtos.PartnerCapabilitiesResponse, error) {
	// Replace path parameter
	path := strings.Replace(s.config.Endpoints.GetPartnerCapabilities, ":id", partnerID, 1)

	req := outbound.Request{
		Method: "GET",
		Path:   path,
	}

	var response dtos.PartnerCapabilitiesResponse
	if err := s.transport.Execute(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to get partner capabilities for partner %s: %w", partnerID, err)
	}

	return &response, nil
}

// ValidatePartnerAvailability validates if a partner is available for a location
func (s *PartnerIntegrationService) ValidatePartnerAvailability(ctx context.Context, partnerID string, request *dtos.PartnerAvailabilityRequest) (*dtos.PartnerAvailabilityResponse, error) {
	// Replace path parameter
	path := strings.Replace(s.config.Endpoints.ValidatePartnerAvailability, ":id", partnerID, 1)

	req := outbound.Request{
		Method: "POST",
		Path:   path,
		Body:   request,
	}

	var response dtos.PartnerAvailabilityResponse
	if err := s.transport.Execute(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to validate partner availability for partner %s: %w", partnerID, err)
	}

	return &response, nil
}

// GetPartnerDetails fetches detailed information about a specific partner
func (s *PartnerIntegrationService) GetPartnerDetails(ctx context.Context, partnerID string) (*dtos.PartnerDetailsResponse, error) {
	// Replace path parameter
	path := strings.Replace(s.config.Endpoints.GetPartnerDetails, ":id", partnerID, 1)

	req := outbound.Request{
		Method: "GET",
		Path:   path,
	}

	var response dtos.PartnerDetailsResponse
	if err := s.transport.Execute(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to get partner details for partner %s: %w", partnerID, err)
	}

	return &response, nil
}

// GetCircuitBreakerState returns the current state of the circuit breaker
func (s *PartnerIntegrationService) GetCircuitBreakerState() string {
	return s.transport.GetCircuitBreakerState().String()
}

// GetMetrics returns integration metrics
func (s *PartnerIntegrationService) GetMetrics() map[string]interface{} {
	return s.transport.GetMetrics()
}

// UpdateConfig updates the service configuration
func (s *PartnerIntegrationService) UpdateConfig(newConfig config.PartnerServiceConfig) {
	s.config = newConfig
	s.transport.UpdateConfig(newConfig.ToHTTPTransportConfig())
}

// mapPartnerSummaryToInfo converts DTO response to interface model
func (s *PartnerIntegrationService) mapPartnerSummaryToInfo(summary dtos.PartnerSummary) interfaces.PartnerInfo {
	// Convert string ID to uint - in real implementation, ensure proper conversion
	var partnerID uint
	fmt.Sscanf(summary.ID, "%d", &partnerID)

	return interfaces.PartnerInfo{
		ID:       partnerID,
		Name:     summary.Name,
		IsActive: summary.Status == "active",
		Type:     summary.Type,
	}
}

// Response DTOs for internal API conversion

// EffectivePreferenceResponse represents effective preference from Partner Service
type EffectivePreferenceResponse struct {
	EntityType      string  `json:"entity_type"`
	EntityID        string  `json:"entity_id"`
	PreferenceValue string  `json:"preference_value"`
	EffectiveRating float64 `json:"effective_rating"`
}

// EffectiveRatingResponse represents effective rating from Partner Service
type EffectiveRatingResponse struct {
	EntityType string  `json:"entity_type"`
	EntityID   string  `json:"entity_id"`
	Rating     float64 `json:"rating"`
}
