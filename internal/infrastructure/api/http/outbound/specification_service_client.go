package outbound

import (
	"context"
	"fmt"
	"log"
	"time"

	"prayog-serviceability-service/internal/infrastructure/resilience"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// SpecificationServiceClient implements the SpecificationServiceClient interface
type SpecificationServiceClient struct {
	transport *HTTPTransport
	config    SpecificationServiceConfig
	logger    *log.Logger
}

// SpecificationServiceConfig holds configuration for Specification Service client
type SpecificationServiceConfig struct {
	ServiceName    string                   `yaml:"service_name" json:"service_name"`
	BaseURL        string                   `yaml:"base_url" json:"base_url"`
	Timeout        time.Duration            `yaml:"timeout" json:"timeout"`
	DefaultHeaders map[string]string        `yaml:"default_headers" json:"default_headers"`
	EnableLogging  bool                     `yaml:"enable_logging" json:"enable_logging"`
	CircuitBreaker resilience.Config        `yaml:"circuit_breaker" json:"circuit_breaker"`
	RetryPolicy    resilience.RetryPolicy   `yaml:"retry_policy" json:"retry_policy"`
	TimeoutConfig  resilience.TimeoutConfig `yaml:"timeout_config" json:"timeout_config"`
	APIVersion     string                   `yaml:"api_version" json:"api_version"`
	AuthToken      string                   `yaml:"auth_token" json:"auth_token"`
}

// NewSpecificationServiceClient creates a new Specification Service client
func NewSpecificationServiceClient(config SpecificationServiceConfig, logger *log.Logger) *SpecificationServiceClient {
	// Prepare default headers
	defaultHeaders := make(map[string]string)
	if config.DefaultHeaders != nil {
		for k, v := range config.DefaultHeaders {
			defaultHeaders[k] = v
		}
	}

	// Add authentication header if token is provided
	if config.AuthToken != "" {
		defaultHeaders["Authorization"] = fmt.Sprintf("Bearer %s", config.AuthToken)
	}

	// Add content type and accept headers
	defaultHeaders["Content-Type"] = "application/json"
	defaultHeaders["Accept"] = "application/json"

	// Add API version header if specified
	if config.APIVersion != "" {
		defaultHeaders["X-API-Version"] = config.APIVersion
	}

	// Create transport configuration
	transportConfig := Config{
		ServiceName:    config.ServiceName,
		BaseURL:        config.BaseURL,
		Timeout:        config.Timeout,
		DefaultHeaders: defaultHeaders,
		EnableLogging:  config.EnableLogging,
		CircuitBreaker: config.CircuitBreaker,
		RetryPolicy:    config.RetryPolicy,
		TimeoutConfig:  config.TimeoutConfig,
	}

	// Create HTTP transport
	transport := NewHTTPTransport(transportConfig)

	return &SpecificationServiceClient{
		transport: transport,
		config:    config,
		logger:    logger,
	}
}

// GetSpecDefinitions fetches all specification definitions
func (c *SpecificationServiceClient) GetSpecDefinitions(ctx context.Context) ([]interfaces.SpecDefinition, error) {
	var response dtos.SpecDefinitionsResponse

	err := c.transport.GET(ctx, "/api/v1/spec-definitions", &response)
	if err != nil {
		c.logError("Failed to fetch spec definitions", err)
		return nil, fmt.Errorf("failed to fetch spec definitions: %w", err)
	}

	// Convert DTOs to interface types
	specDefinitions := make([]interfaces.SpecDefinition, len(response.SpecDefinitions))
	for i, spec := range response.SpecDefinitions {
		specDefinitions[i] = interfaces.SpecDefinition{
			ID:          spec.ID,
			SpecType:    spec.SpecType,
			Name:        spec.Name,
			Description: spec.Description,
		}
	}

	c.logInfo(fmt.Sprintf("Successfully fetched %d spec definitions", len(specDefinitions)))
	return specDefinitions, nil
}

// GetCatalogs fetches all catalogs
func (c *SpecificationServiceClient) GetCatalogs(ctx context.Context) ([]interfaces.Catalog, error) {
	var response dtos.CatalogsResponse

	err := c.transport.GET(ctx, "/api/v1/catalogs", &response)
	if err != nil {
		c.logError("Failed to fetch catalogs", err)
		return nil, fmt.Errorf("failed to fetch catalogs: %w", err)
	}

	// Convert DTOs to interface types
	catalogs := make([]interfaces.Catalog, len(response.Catalogs))
	for i, catalog := range response.Catalogs {
		// Convert catalog items
		catalogItems := make([]interfaces.CatalogItem, len(catalog.CatalogItems))
		for j, item := range catalog.CatalogItems {
			catalogItems[j] = interfaces.CatalogItem{
				ID:          item.ID,
				Code:        item.Code,
				Name:        item.Name,
				Description: item.Description,
			}
		}

		catalogs[i] = interfaces.Catalog{
			ID:           catalog.ID,
			SpecDefID:    catalog.SpecDefID,
			Name:         catalog.Name,
			Description:  catalog.Description,
			CatalogItems: catalogItems,
		}
	}

	c.logInfo(fmt.Sprintf("Successfully fetched %d catalogs", len(catalogs)))
	return catalogs, nil
}

// GetEntitySpecifications fetches entity specifications for a specific entity
func (c *SpecificationServiceClient) GetEntitySpecifications(ctx context.Context, entityType, entityID string) ([]interfaces.EntitySpecification, error) {
	// Prepare request body
	requestBody := dtos.EntitySpecificationsRequest{
		EntityType:   entityType,
		EntityID:     entityID,
		OnlyActive:   true,
		IncludeItems: true,
	}

	var response dtos.EntitySpecificationsResponse

	err := c.transport.POST(ctx, "/api/v1/entity-specifications", requestBody, &response)
	if err != nil {
		c.logError(fmt.Sprintf("Failed to fetch entity specifications for %s:%s", entityType, entityID), err)
		return nil, fmt.Errorf("failed to fetch entity specifications for %s:%s: %w", entityType, entityID, err)
	}

	// Convert DTOs to interface types
	entitySpecs := make([]interfaces.EntitySpecification, len(response.EntitySpecifications))
	for i, spec := range response.EntitySpecifications {
		entitySpecs[i] = interfaces.EntitySpecification{
			ID:         spec.ID,
			EntityType: spec.EntityType,
			EntityID:   spec.EntityID,
			SpecDefID:  spec.SpecDefID,
		}
	}

	c.logInfo(fmt.Sprintf("Successfully fetched %d entity specifications for %s:%s", len(entitySpecs), entityType, entityID))
	return entitySpecs, nil
}

// GetServiceDefinitions fetches service definitions based on query parameters
func (c *SpecificationServiceClient) GetServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string) ([]interfaces.ServiceDefinition, error) {
	// Prepare request body
	requestBody := dtos.ServiceDefinitionQueryRequest{
		ServiceTypes:     serviceTypes,
		ParcelCategories: parcelCategories,
		OnlyActive:       true,
		IncludeDefaults:  true,
	}

	var response dtos.ServiceDefinitionQueryResponse

	err := c.transport.POST(ctx, "/api/v1/service-definitions/query", requestBody, &response)
	if err != nil {
		c.logError("Failed to fetch service definitions", err)
		return nil, fmt.Errorf("failed to fetch service definitions: %w", err)
	}

	// Convert DTOs to interface types
	serviceDefinitions := make([]interfaces.ServiceDefinition, len(response.ServiceDefinitions))
	for i, def := range response.ServiceDefinitions {
		serviceDefinitions[i] = interfaces.ServiceDefinition{
			ServiceType:       def.ServiceType,
			ParcelCategory:    def.ParcelCategory,
			DefaultOperations: def.DefaultOperations,
			DefaultPayments:   def.DefaultPayments,
			DefaultDelivery:   def.DefaultDelivery,
			Description:       def.Description,
		}
	}

	c.logInfo(fmt.Sprintf("Successfully fetched %d service definitions", len(serviceDefinitions)))
	return serviceDefinitions, nil
}

// HealthCheck performs a health check on the Specification Service
func (c *SpecificationServiceClient) HealthCheck(ctx context.Context) error {
	var response dtos.HealthCheckResponse

	err := c.transport.GET(ctx, "/health", &response)
	if err != nil {
		c.logError("Health check failed", err)
		return fmt.Errorf("specification service health check failed: %w", err)
	}

	if response.Status != "healthy" && response.Status != "ok" {
		c.logError(fmt.Sprintf("Service health check returned status: %s", response.Status), nil)
		return fmt.Errorf("specification service is not healthy: status=%s", response.Status)
	}

	c.logInfo("Health check passed")
	return nil
}

// GetMetrics returns metrics from the underlying transport
func (c *SpecificationServiceClient) GetMetrics() map[string]interface{} {
	metrics := c.transport.GetMetrics()
	metrics["service_name"] = c.config.ServiceName
	metrics["base_url"] = c.config.BaseURL
	return metrics
}

// UpdateConfig updates the client configuration
func (c *SpecificationServiceClient) UpdateConfig(config SpecificationServiceConfig) {
	c.config = config

	// Update transport configuration
	transportConfig := Config{
		ServiceName:    config.ServiceName,
		BaseURL:        config.BaseURL,
		Timeout:        config.Timeout,
		DefaultHeaders: config.DefaultHeaders,
		EnableLogging:  config.EnableLogging,
		CircuitBreaker: config.CircuitBreaker,
		RetryPolicy:    config.RetryPolicy,
		TimeoutConfig:  config.TimeoutConfig,
	}

	c.transport.UpdateConfig(transportConfig)
	c.logInfo("Configuration updated")
}

// Helper methods for logging
func (c *SpecificationServiceClient) logInfo(message string) {
	if c.logger != nil && c.config.EnableLogging {
		c.logger.Printf("[SPEC-SERVICE] INFO: %s", message)
	}
}

func (c *SpecificationServiceClient) logError(message string, err error) {
	if c.logger != nil {
		if err != nil {
			c.logger.Printf("[SPEC-SERVICE] ERROR: %s - %v", message, err)
		} else {
			c.logger.Printf("[SPEC-SERVICE] ERROR: %s", message)
		}
	}
}

// Verify that SpecificationServiceClient implements the interface
var _ interfaces.SpecificationServiceClient = (*SpecificationServiceClient)(nil)
