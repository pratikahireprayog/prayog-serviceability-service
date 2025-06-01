package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"prayog-serviceability-service/internal/infrastructure/api/http/outbound"
	"prayog-serviceability-service/internal/infrastructure/resilience"

	"github.com/joho/godotenv"
)

// IntegrationConfig holds configuration for all service integrations
type IntegrationConfig struct {
	Partner       PartnerServiceConfig       `yaml:"partner" json:"partner"`
	Specification SpecificationServiceConfig `yaml:"specification" json:"specification"`
}

// PartnerServiceConfig holds configuration for Partner Service integration
type PartnerServiceConfig struct {
	BaseURL        string            `yaml:"base_url" json:"base_url"`
	APIKey         string            `yaml:"api_key" json:"api_key"`
	Timeout        time.Duration     `yaml:"timeout" json:"timeout"`
	EnableLogging  bool              `yaml:"enable_logging" json:"enable_logging"`
	DefaultHeaders map[string]string `yaml:"default_headers" json:"default_headers"`

	// Resilience settings
	CircuitBreaker resilience.Config        `yaml:"circuit_breaker" json:"circuit_breaker"`
	RetryPolicy    resilience.RetryPolicy   `yaml:"retry_policy" json:"retry_policy"`
	TimeoutConfig  resilience.TimeoutConfig `yaml:"timeout_config" json:"timeout_config"`

	// Business-specific settings
	Endpoints PartnerEndpoints `yaml:"endpoints" json:"endpoints"`
}

// PartnerEndpoints defines Partner Service API endpoints
type PartnerEndpoints struct {
	GetPartnersByLocation       string `yaml:"get_partners_by_location" json:"get_partners_by_location"`
	GetPartnerDetails           string `yaml:"get_partner_details" json:"get_partner_details"`
	GetPartnerCapabilities      string `yaml:"get_partner_capabilities" json:"get_partner_capabilities"`
	ValidatePartnerAvailability string `yaml:"validate_partner_availability" json:"validate_partner_availability"`
}

// SpecificationServiceConfig holds configuration for Specification Service integration
type SpecificationServiceConfig struct {
	BaseURL        string            `yaml:"base_url" json:"base_url"`
	APIKey         string            `yaml:"api_key" json:"api_key"`
	Timeout        time.Duration     `yaml:"timeout" json:"timeout"`
	EnableLogging  bool              `yaml:"enable_logging" json:"enable_logging"`
	DefaultHeaders map[string]string `yaml:"default_headers" json:"default_headers"`

	// Resilience settings
	CircuitBreaker resilience.Config        `yaml:"circuit_breaker" json:"circuit_breaker"`
	RetryPolicy    resilience.RetryPolicy   `yaml:"retry_policy" json:"retry_policy"`
	TimeoutConfig  resilience.TimeoutConfig `yaml:"timeout_config" json:"timeout_config"`

	// Business-specific settings
	Endpoints SpecificationEndpoints `yaml:"endpoints" json:"endpoints"`
}

// SpecificationEndpoints defines Specification Service API endpoints
type SpecificationEndpoints struct {
	GetCatalogs             string `yaml:"get_catalogs" json:"get_catalogs"`
	GetSpecificationsByType string `yaml:"get_specifications_by_type" json:"get_specifications_by_type"`
	ValidateSpecification   string `yaml:"validate_specification" json:"validate_specification"`
}

// ServiceDefinitionResolverConfig holds configuration for the Service Definition Resolver
type ServiceDefinitionResolverConfig struct {
	// Cache configuration
	CacheEnabled bool          `yaml:"cache_enabled" json:"cache_enabled"`
	CacheTTL     time.Duration `yaml:"cache_ttl" json:"cache_ttl"`

	// Circuit breaker configuration
	CircuitBreakerEnabled bool          `yaml:"circuit_breaker_enabled" json:"circuit_breaker_enabled"`
	FailureThreshold      int           `yaml:"failure_threshold" json:"failure_threshold"`
	RecoveryTimeout       time.Duration `yaml:"recovery_timeout" json:"recovery_timeout"`

	// Validation configuration
	ValidateResponses bool `yaml:"validate_responses" json:"validate_responses"`
	StrictValidation  bool `yaml:"strict_validation" json:"strict_validation"`

	// Transformation configuration
	EnableTransformation    bool `yaml:"enable_transformation" json:"enable_transformation"`
	TransformCatalogToLists bool `yaml:"transform_catalog_to_lists" json:"transform_catalog_to_lists"`

	// Concurrency configuration
	MaxConcurrentRequests int           `yaml:"max_concurrent_requests" json:"max_concurrent_requests"`
	RequestTimeout        time.Duration `yaml:"request_timeout" json:"request_timeout"`

	// Category cache configuration
	CategoryCacheTTL time.Duration `yaml:"category_cache_ttl" json:"category_cache_ttl"`
}

// LoadIntegrationConfig loads integration configuration from environment variables
func LoadIntegrationConfig() IntegrationConfig {
	// Load .env file
	godotenv.Load()

	partnerConfig := PartnerServiceConfig{
		BaseURL:       getEnvOrDefault("PARTNER_SERVICE_BASE_URL", "http://localhost:9024"),
		APIKey:        getEnvOrDefault("PARTNER_SERVICE_API_KEY", ""),
		Timeout:       getEnvAsDurationOrDefault("PARTNER_SERVICE_TIMEOUT", 30*time.Second),
		EnableLogging: getEnvAsBoolOrDefault("PARTNER_SERVICE_ENABLE_LOGGING", true),
		DefaultHeaders: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
		CircuitBreaker: resilience.Config{
			Name:             "partner-service",
			MaxRequests:      10,
			FailureThreshold: 5,
			Timeout:          60 * time.Second,
			Interval:         5 * time.Second,
		},
		RetryPolicy: resilience.RetryPolicy{
			MaxRetries:        getEnvAsIntOrDefault("PARTNER_SERVICE_MAX_RETRIES", 3),
			InitialDelay:      getEnvAsDurationOrDefault("PARTNER_SERVICE_RETRY_INITIAL_DELAY", 1*time.Second),
			MaxDelay:          getEnvAsDurationOrDefault("PARTNER_SERVICE_RETRY_MAX_DELAY", 30*time.Second),
			BackoffMultiplier: 2.0,
			Jitter:            true,
		},
		TimeoutConfig: resilience.TimeoutConfig{
			RequestTimeout:    getEnvAsDurationOrDefault("PARTNER_SERVICE_TIMEOUT", 30*time.Second),
			ConnectionTimeout: getEnvAsDurationOrDefault("PARTNER_SERVICE_CONNECTION_TIMEOUT", 10*time.Second),
			IdleTimeout:       90 * time.Second,
		},
		Endpoints: PartnerEndpoints{
			GetPartnersByLocation:       "/api/v1/partners",
			GetPartnerDetails:           "/api/v1/partners/:id",
			GetPartnerCapabilities:      "/api/v1/partners/:id/capabilities",
			ValidatePartnerAvailability: "/api/v1/partners/:id/availability",
		},
	}

	specConfig := SpecificationServiceConfig{
		BaseURL:       getEnvOrDefault("SPECIFICATION_SERVICE_BASE_URL", "http://localhost:9023"),
		APIKey:        getEnvOrDefault("SPECIFICATION_SERVICE_API_KEY", ""),
		Timeout:       getEnvAsDurationOrDefault("SPECIFICATION_SERVICE_TIMEOUT", 30*time.Second),
		EnableLogging: getEnvAsBoolOrDefault("SPECIFICATION_SERVICE_ENABLE_LOGGING", true),
		DefaultHeaders: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
		CircuitBreaker: resilience.Config{
			Name:             "specification-service",
			MaxRequests:      10,
			FailureThreshold: 5,
			Timeout:          60 * time.Second,
			Interval:         5 * time.Second,
		},
		RetryPolicy: resilience.RetryPolicy{
			MaxRetries:        getEnvAsIntOrDefault("SPECIFICATION_SERVICE_MAX_RETRIES", 3),
			InitialDelay:      getEnvAsDurationOrDefault("SPECIFICATION_SERVICE_RETRY_INITIAL_DELAY", 1*time.Second),
			MaxDelay:          getEnvAsDurationOrDefault("SPECIFICATION_SERVICE_RETRY_MAX_DELAY", 30*time.Second),
			BackoffMultiplier: 2.0,
			Jitter:            true,
		},
		TimeoutConfig: resilience.TimeoutConfig{
			RequestTimeout:    getEnvAsDurationOrDefault("SPECIFICATION_SERVICE_TIMEOUT", 30*time.Second),
			ConnectionTimeout: getEnvAsDurationOrDefault("SPECIFICATION_SERVICE_CONNECTION_TIMEOUT", 10*time.Second),
			IdleTimeout:       90 * time.Second,
		},
		Endpoints: SpecificationEndpoints{
			GetCatalogs:             "/specification/api/v1/catalogs",
			GetSpecificationsByType: "/specification/api/v1/specifications/type/:type",
			ValidateSpecification:   "/specification/api/v1/specifications/validate",
		},
	}

	return IntegrationConfig{
		Partner:       partnerConfig,
		Specification: specConfig,
	}
}

// ToHTTPTransportConfig converts PartnerServiceConfig to HTTP transport config
func (c PartnerServiceConfig) ToHTTPTransportConfig() outbound.Config {
	return outbound.Config{
		ServiceName:    "partner-service",
		BaseURL:        c.BaseURL,
		Timeout:        c.Timeout,
		DefaultHeaders: c.DefaultHeaders,
		EnableLogging:  c.EnableLogging,
		CircuitBreaker: c.CircuitBreaker,
		RetryPolicy:    c.RetryPolicy,
		TimeoutConfig:  c.TimeoutConfig,
	}
}

// ToHTTPTransportConfig converts SpecificationServiceConfig to HTTP transport config
func (c SpecificationServiceConfig) ToHTTPTransportConfig() outbound.Config {
	return outbound.Config{
		ServiceName:    "specification-service",
		BaseURL:        c.BaseURL,
		Timeout:        c.Timeout,
		DefaultHeaders: c.DefaultHeaders,
		EnableLogging:  c.EnableLogging,
		CircuitBreaker: c.CircuitBreaker,
		RetryPolicy:    c.RetryPolicy,
		TimeoutConfig:  c.TimeoutConfig,
	}
}

// Validate validates the integration configuration
func (c IntegrationConfig) Validate() error {
	if err := c.Partner.Validate(); err != nil {
		return err
	}
	return c.Specification.Validate()
}

// Validate validates the Partner Service configuration
func (c PartnerServiceConfig) Validate() error {
	if c.BaseURL == "" {
		return NewConfigError("partner.base_url is required")
	}
	if c.Timeout <= 0 {
		return NewConfigError("partner.timeout must be positive")
	}
	return nil
}

// Validate validates the Specification Service configuration
func (c SpecificationServiceConfig) Validate() error {
	if c.BaseURL == "" {
		return NewConfigError("specification.base_url is required")
	}
	if c.Timeout <= 0 {
		return NewConfigError("specification.timeout must be positive")
	}
	return nil
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e ConfigError) Error() string {
	return fmt.Sprintf("config error [%s]: %s", e.Field, e.Message)
}

// NewConfigError creates a new configuration error
func NewConfigError(message string) ConfigError {
	return ConfigError{
		Message: message,
	}
}

// Helper functions using os.Getenv() directly
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := time.ParseDuration(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}
