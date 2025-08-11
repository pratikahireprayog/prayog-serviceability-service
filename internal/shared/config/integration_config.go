package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"prayog-serviceability-service/internal/infrastructure/api/http/outbound"
	"prayog-serviceability-service/internal/infrastructure/resilience"
)

// IntegrationConfig holds configuration for all service integrations
type IntegrationConfig struct {
	Partner         PartnerServiceConfig       `yaml:"partner" json:"partner"`
	Specification   SpecificationServiceConfig `yaml:"specification" json:"specification"`
	PartnerAdapters PartnerAdaptersConfig      `yaml:"partner_adapters" json:"partner_adapters"`
}

// PartnerAdaptersConfig holds configuration for all partner adapters
type PartnerAdaptersConfig struct {
	Shipyaari    ShipyaariConfig    `yaml:"shipyaari" json:"shipyaari"`
	SmileCourier SmileCourierConfig `yaml:"smile_courier" json:"smile_courier"`
	SmileEcom    SmileEcomConfig    `yaml:"smile_ecom" json:"smile_ecom"`
	DHL          DHLConfig          `yaml:"dhl" json:"dhl"`
	SmileCargo   SmileCargoConfig   `yaml:"smile_cargo" json:"smile_cargo"`
	Delcaper     DelcaperConfig     `yaml:"delcaper" json:"delcaper"`
	SmileHubOps  SmileHubOpsConfig  `yaml:"smile_hubops" json:"smile_hubops"`
}

// ShipyaariConfig configuration for Shipyaari partner adapter
type ShipyaariConfig struct {
	BaseURL           string        `yaml:"base_url" json:"base_url"`
	Email             string        `yaml:"email" json:"email"`
	Password          string        `yaml:"password" json:"password"`
	TokenURL          string        `yaml:"token_url" json:"token_url"`
	CheckServiceURL   string        `yaml:"check_service_url" json:"check_service_url"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout"`
	TokenExpiryBuffer time.Duration `yaml:"token_expiry_buffer" json:"token_expiry_buffer"`
	MaxRetries        int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay        time.Duration `yaml:"retry_delay" json:"retry_delay"`
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	Rating            float64       `yaml:"rating" json:"rating"`
}

// SmileCourierConfig configuration for Smile Courier partner adapter
type SmileCourierConfig struct {
	BaseURL         string        `yaml:"base_url" json:"base_url"`
	CheckServiceURL string        `yaml:"check_service_url" json:"check_service_url"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	MaxRetries      int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay      time.Duration `yaml:"retry_delay" json:"retry_delay"`
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	Rating          float64       `yaml:"rating" json:"rating"`
}

// SmileEcomConfig configuration for Smile Ecom partner adapter (database-based)
type SmileEcomConfig struct {
	TableName    string        `yaml:"table_name" json:"table_name"`
	Enabled      bool          `yaml:"enabled" json:"enabled"`
	Rating       float64       `yaml:"rating" json:"rating"`
	CacheEnabled bool          `yaml:"cache_enabled" json:"cache_enabled"`
	CacheTTL     time.Duration `yaml:"cache_ttl" json:"cache_ttl"`
}

// DHLConfig configuration for DHL partner adapter (international shipping)
type DHLConfig struct {
	BaseURL       string        `yaml:"base_url" json:"base_url"`
	Username      string        `yaml:"username" json:"username"`
	Password      string        `yaml:"password" json:"password"`
	BasicAuth     string        `yaml:"basic_auth" json:"basic_auth"`
	APIKey        string        `yaml:"api_key" json:"api_key"`
	AccountNumber string        `yaml:"account_number" json:"account_number"`
	AuthURL       string        `yaml:"auth_url" json:"auth_url"`
	ServiceURL    string        `yaml:"service_url" json:"service_url"`
	TrackingURL   string        `yaml:"tracking_url" json:"tracking_url"`
	Timeout       time.Duration `yaml:"timeout" json:"timeout"`
	MaxRetries    int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay    time.Duration `yaml:"retry_delay" json:"retry_delay"`
	Enabled       bool          `yaml:"enabled" json:"enabled"`
	Rating        float64       `yaml:"rating" json:"rating"`
	SandboxMode   bool          `yaml:"sandbox_mode" json:"sandbox_mode"`
}

// SmileCargoConfig configuration for Smile Cargo partner adapter (cargo/freight)
type SmileCargoConfig struct {
	BaseURL     string        `yaml:"base_url" json:"base_url"`
	APIKey      string        `yaml:"api_key" json:"api_key"`
	VendorCode  string        `yaml:"vendor_code" json:"vendor_code"`
	ServiceURL  string        `yaml:"service_url" json:"service_url"`
	QuoteURL    string        `yaml:"quote_url" json:"quote_url"`
	TrackingURL string        `yaml:"tracking_url" json:"tracking_url"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	MaxRetries  int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay  time.Duration `yaml:"retry_delay" json:"retry_delay"`
	Enabled     bool          `yaml:"enabled" json:"enabled"`
	Rating      float64       `yaml:"rating" json:"rating"`
	MinWeightKG float64       `yaml:"min_weight_kg" json:"min_weight_kg"`
	MaxWeightKG float64       `yaml:"max_weight_kg" json:"max_weight_kg"`
}

// DelcaperConfig configuration for Delcaper partner adapter (hyperlocal delivery)
type DelcaperConfig struct {
	BaseURL         string        `yaml:"base_url" json:"base_url"`
	Email           string        `yaml:"email" json:"email"`
	Password        string        `yaml:"password" json:"password"`
	VendorType      string        `yaml:"vendor_type" json:"vendor_type"`
	LoginURL        string        `yaml:"login_url" json:"login_url"`
	CheckServiceURL string        `yaml:"check_service_url" json:"check_service_url"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	TokenExpiryBuffer time.Duration `yaml:"token_expiry_buffer" json:"token_expiry_buffer"`
	MaxRetries      int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay      time.Duration `yaml:"retry_delay" json:"retry_delay"`
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	Rating          float64       `yaml:"rating" json:"rating"`
}

// SmileHubOpsConfig configuration for Smile HubOps partner adapter (hub operations)
type SmileHubOpsConfig struct {
	BaseURL     string        `yaml:"base_url" json:"base_url"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	MaxRetries  int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay  time.Duration `yaml:"retry_delay" json:"retry_delay"`
	Enabled     bool          `yaml:"enabled" json:"enabled"`
	Rating      float64       `yaml:"rating" json:"rating"`
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
	// Environment variables are loaded and prioritized in app_config.go

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

	// Partner Adapters Configuration
	partnerAdaptersConfig := PartnerAdaptersConfig{
		Shipyaari: ShipyaariConfig{
			BaseURL:           getEnvOrDefault("SHIPYAARI_BASE_URL", "https://api-seller.shipyaari.com"),
			Email:             getEnvOrDefault("SHIPYAARI_EMAIL", "palak.parikh@shreemaruti.com"),
			Password:          getEnvOrDefault("SHIPYAARI_PASSWORD", "Test@123"),
			TokenURL:          "/api/v1/seller/signIn",
			CheckServiceURL:   "/api/v1/order/checkServiceabilityV2",
			Timeout:           getEnvAsDurationOrDefault("SHIPYAARI_TIMEOUT", 30*time.Second),
			TokenExpiryBuffer: getEnvAsDurationOrDefault("SHIPYAARI_TOKEN_EXPIRY_BUFFER", 5*time.Minute),
			MaxRetries:        getEnvAsIntOrDefault("SHIPYAARI_MAX_RETRIES", 3),
			RetryDelay:        getEnvAsDurationOrDefault("SHIPYAARI_RETRY_DELAY", 1*time.Second),
			Enabled:           getEnvAsBoolOrDefault("SHIPYAARI_ENABLED", true),
			Rating:            4.5,
		},
		SmileCourier: SmileCourierConfig{
			BaseURL:         getEnvOrDefault("SMILE_COURIER_BASE_URL", "https://apis.delcaper.com"),
			CheckServiceURL: "/serviceability/courier",
			Timeout:         getEnvAsDurationOrDefault("SMILE_COURIER_TIMEOUT", 30*time.Second),
			MaxRetries:      getEnvAsIntOrDefault("SMILE_COURIER_MAX_RETRIES", 3),
			RetryDelay:      getEnvAsDurationOrDefault("SMILE_COURIER_RETRY_DELAY", 1*time.Second),
			Enabled:         getEnvAsBoolOrDefault("SMILE_COURIER_ENABLED", true),
			Rating:          4.2,
		},
		SmileEcom: SmileEcomConfig{
			TableName:    getEnvOrDefault("SMILE_ECOM_TABLE_NAME", "smile_ecom_serviceability"),
			Enabled:      getEnvAsBoolOrDefault("SMILE_ECOM_ENABLED", true), // Enabled by default
			Rating:       4.0,
			CacheEnabled: getEnvAsBoolOrDefault("SMILE_ECOM_CACHE_ENABLED", true),
			CacheTTL:     getEnvAsDurationOrDefault("SMILE_ECOM_CACHE_TTL", 15*time.Minute),
		},
		DHL: DHLConfig{
			BaseURL:       getEnvOrDefault("DHL_BASE_URL", ""),
			Username:      getEnvOrDefault("DHL_USERNAME", ""),
			Password:      getEnvOrDefault("DHL_PASSWORD", ""),
			BasicAuth:     getEnvOrDefault("DHL_BASIC_AUTH", ""),
			APIKey:        getEnvOrDefault("DHL_API_KEY", ""),
			AccountNumber: getEnvOrDefault("DHL_ACCOUNT_NUMBER", ""),
			AuthURL:       getEnvOrDefault("DHL_AUTH_URL", "/v1/auth/login"),
			ServiceURL:    getEnvOrDefault("DHL_SERVICE_URL", "/v1/serviceability"),
			TrackingURL:   getEnvOrDefault("DHL_TRACKING_URL", "/v1/tracking"),
			Timeout:       getEnvAsDurationOrDefault("DHL_TIMEOUT", 30*time.Second),
			MaxRetries:    getEnvAsIntOrDefault("DHL_MAX_RETRIES", 3),
			RetryDelay:    getEnvAsDurationOrDefault("DHL_RETRY_DELAY", 2*time.Second),
			Enabled:       getEnvAsBoolOrDefault("DHL_ENABLED", true), // Disabled by default
			Rating:        4.7,
			SandboxMode:   getEnvAsBoolOrDefault("DHL_SANDBOX_MODE", true),
		},
		SmileCargo: SmileCargoConfig{
			BaseURL:     getEnvOrDefault("SMILE_CARGO_BASE_URL", "https://apis.delcaper.com"),
			APIKey:      getEnvOrDefault("SMILE_CARGO_API_KEY", ""),
			VendorCode:  getEnvOrDefault("SMILE_CARGO_VENDOR_CODE", "bhav19"),
			ServiceURL:  getEnvOrDefault("SMILE_CARGO_SERVICE_URL", "/delivery-orchestrator/service-availability/v3"),
			QuoteURL:    getEnvOrDefault("SMILE_CARGO_QUOTE_URL", "/api/v1/quote"),
			TrackingURL: getEnvOrDefault("SMILE_CARGO_TRACKING_URL", "/api/v1/tracking"),
			Timeout:     getEnvAsDurationOrDefault("SMILE_CARGO_TIMEOUT", 45*time.Second),
			MaxRetries:  getEnvAsIntOrDefault("SMILE_CARGO_MAX_RETRIES", 3),
			RetryDelay:  getEnvAsDurationOrDefault("SMILE_CARGO_RETRY_DELAY", 2*time.Second),
			Enabled:     getEnvAsBoolOrDefault("SMILE_CARGO_ENABLED", true), // Disabled by default
			Rating:      4.3,
			MinWeightKG: getEnvAsFloatOrDefault("SMILE_CARGO_MIN_WEIGHT_KG", 25.0),   // 25kg minimum for cargo
			MaxWeightKG: getEnvAsFloatOrDefault("SMILE_CARGO_MAX_WEIGHT_KG", 5000.0), // 5 ton maximum
		},
		Delcaper: DelcaperConfig{
			BaseURL:           getEnvOrDefault("DELCAPER_BASE_URL", "https://apis.delcaper.com"),
			Email:             getEnvOrDefault("DELCAPER_EMAIL", "atharva.bodke@shreemaruti.com"),
			Password:          getEnvOrDefault("DELCAPER_PASSWORD", "Atharva@PRS2024"),
			VendorType:        getEnvOrDefault("DELCAPER_VENDOR_TYPE", "SELLER"),
			LoginURL:          getEnvOrDefault("DELCAPER_LOGIN_URL", "/auth/login"),
			CheckServiceURL:   getEnvOrDefault("DELCAPER_CHECK_SERVICE_URL", "/fulfillment/public/seller/order/check-feasible"),
			Timeout:           getEnvAsDurationOrDefault("DELCAPER_TIMEOUT", 30*time.Second),
			TokenExpiryBuffer: getEnvAsDurationOrDefault("DELCAPER_TOKEN_EXPIRY_BUFFER", 5*time.Minute),
			MaxRetries:        getEnvAsIntOrDefault("DELCAPER_MAX_RETRIES", 3),
			RetryDelay:        getEnvAsDurationOrDefault("DELCAPER_RETRY_DELAY", 1*time.Second),
			Enabled:           getEnvAsBoolOrDefault("DELCAPER_ENABLED", true),
			Rating:            4.0,
		},
		SmileHubOps: SmileHubOpsConfig{
			BaseURL:    getEnvOrDefault("SMILE_HUBOPS_BASE_URL", "https://devhubopsapis.innofulfill.com"),
			Timeout:    getEnvAsDurationOrDefault("SMILE_HUBOPS_TIMEOUT", 30*time.Second),
			MaxRetries: getEnvAsIntOrDefault("SMILE_HUBOPS_MAX_RETRIES", 3),
			RetryDelay: getEnvAsDurationOrDefault("SMILE_HUBOPS_RETRY_DELAY", 1*time.Second),
			Enabled:    getEnvAsBoolOrDefault("SMILE_HUBOPS_ENABLED", true),
			Rating:     4.5,
		},
	}

	return IntegrationConfig{
		Partner:         partnerConfig,
		Specification:   specConfig,
		PartnerAdapters: partnerAdaptersConfig,
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

// Note: Helper functions (getEnvOrDefault, getEnvAsIntOrDefault, etc.)
// are now defined in app_config.go and shared across configuration files

// getEnvAsFloatOrDefault gets environment variable as float64 or returns default
func getEnvAsFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}
