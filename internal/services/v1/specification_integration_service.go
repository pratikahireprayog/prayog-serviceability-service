package services

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/infrastructure/api/http/outbound"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// SpecificationIntegrationService implements business logic for Specification Service integration
type SpecificationIntegrationService struct {
	transport *outbound.HTTPTransport
	config    config.SpecificationServiceConfig
}

// NewSpecificationIntegrationService creates a new Specification Service integration
func NewSpecificationIntegrationService(config config.SpecificationServiceConfig) interfaces.SpecificationServiceClient {
	transport := outbound.NewHTTPTransport(config.ToHTTPTransportConfig())

	return &SpecificationIntegrationService{
		transport: transport,
		config:    config,
	}
}

// GetSpecDefinitions fetches all specification definitions
// Implements interfaces.SpecificationServiceClient interface method
func (s *SpecificationIntegrationService) GetSpecDefinitions(ctx context.Context) ([]interfaces.SpecDefinition, error) {
	req := outbound.Request{
		Method: "GET",
		Path:   "/api/v1/spec-definitions",
	}

	var response struct {
		SpecDefinitions []SpecDefinitionResponse `json:"spec_definitions"`
		Total           int                      `json:"total"`
	}

	if err := s.transport.Execute(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to get specification definitions: %w", err)
	}

	// Convert response to interface models
	specDefs := make([]interfaces.SpecDefinition, len(response.SpecDefinitions))
	for i, specResp := range response.SpecDefinitions {
		specDefs[i] = interfaces.SpecDefinition{
			ID:          specResp.ID,
			SpecType:    specResp.SpecType,
			Name:        specResp.Name,
			Description: specResp.Description,
		}
	}

	return specDefs, nil
}

// GetCatalogs fetches all catalogs
// Implements interfaces.SpecificationServiceClient interface method
func (s *SpecificationIntegrationService) GetCatalogs(ctx context.Context) ([]interfaces.Catalog, error) {
	req := outbound.Request{
		Method: "GET",
		Path:   s.config.Endpoints.GetCatalogs,
	}

	var response struct {
		Catalogs []CatalogResponse `json:"catalogs"`
		Total    int               `json:"total"`
	}

	if err := s.transport.Execute(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to get catalogs: %w", err)
	}

	// Convert response to interface models
	catalogs := make([]interfaces.Catalog, len(response.Catalogs))
	for i, catResp := range response.Catalogs {
		catalogItems := make([]interfaces.CatalogItem, len(catResp.CatalogItems))
		for j, itemResp := range catResp.CatalogItems {
			catalogItems[j] = interfaces.CatalogItem{
				ID:          itemResp.ID,
				Code:        itemResp.Code,
				Name:        itemResp.Name,
				Description: itemResp.Description,
			}
		}

		catalogs[i] = interfaces.Catalog{
			ID:           catResp.ID,
			SpecDefID:    catResp.SpecDefID,
			Name:         catResp.Name,
			Description:  catResp.Description,
			CatalogItems: catalogItems,
		}
	}

	return catalogs, nil
}

// GetEntitySpecifications fetches entity specifications by type and ID
// Implements interfaces.SpecificationServiceClient interface method
func (s *SpecificationIntegrationService) GetEntitySpecifications(ctx context.Context, entityType, entityID string) ([]interfaces.EntitySpecification, error) {
	req := outbound.Request{
		Method: "GET",
		Path:   "/api/v1/entity-specifications",
		Query: map[string]string{
			"entity_type": entityType,
			"entity_id":   entityID,
		},
	}

	var response struct {
		EntitySpecifications []EntitySpecificationResponse `json:"entity_specifications"`
		Total                int                           `json:"total"`
	}

	if err := s.transport.Execute(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to get entity specifications: %w", err)
	}

	// Convert response to interface models
	entitySpecs := make([]interfaces.EntitySpecification, len(response.EntitySpecifications))
	for i, specResp := range response.EntitySpecifications {
		entitySpecs[i] = interfaces.EntitySpecification{
			ID:          specResp.ID,
			EntityType:  specResp.EntityType,
			EntityID:    specResp.EntityID,
			SpecDefID:   specResp.SpecDefID,
			SpecValue:   specResp.SpecValue,
			Description: specResp.Description,
		}
	}

	return entitySpecs, nil
}

// Additional business methods (not part of interface but useful for internal operations)

// ValidateSpecification validates a specification request
func (s *SpecificationIntegrationService) ValidateSpecification(ctx context.Context, request interface{}) (bool, error) {
	req := outbound.Request{
		Method: "POST",
		Path:   s.config.Endpoints.ValidateSpecification,
		Body:   request,
	}

	var response struct {
		IsValid bool   `json:"is_valid"`
		Reason  string `json:"reason,omitempty"`
	}

	if err := s.transport.Execute(ctx, req, &response); err != nil {
		return false, fmt.Errorf("failed to validate specification: %w", err)
	}

	return response.IsValid, nil
}

// GetCircuitBreakerState returns the current state of the circuit breaker
func (s *SpecificationIntegrationService) GetCircuitBreakerState() string {
	return s.transport.GetCircuitBreakerState().String()
}

// GetMetrics returns integration metrics
func (s *SpecificationIntegrationService) GetMetrics() map[string]interface{} {
	return s.transport.GetMetrics()
}

// UpdateConfig updates the service configuration
func (s *SpecificationIntegrationService) UpdateConfig(newConfig config.SpecificationServiceConfig) {
	s.config = newConfig
	s.transport.UpdateConfig(newConfig.ToHTTPTransportConfig())
}

// Response DTOs for internal API conversion

// SpecDefinitionResponse represents specification definition from API
type SpecDefinitionResponse struct {
	ID          uint   `json:"id"`
	SpecType    string `json:"spec_type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CatalogResponse represents catalog from API
type CatalogResponse struct {
	ID           uint                  `json:"id"`
	SpecDefID    uint                  `json:"spec_def_id"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	CatalogItems []CatalogItemResponse `json:"catalog_items"`
}

// CatalogItemResponse represents catalog item from API
type CatalogItemResponse struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// EntitySpecificationResponse represents entity specification from API
type EntitySpecificationResponse struct {
	ID          uint   `json:"id"`
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	SpecDefID   uint   `json:"spec_def_id"`
	SpecValue   string `json:"spec_value"`
	Description string `json:"description"`
}
