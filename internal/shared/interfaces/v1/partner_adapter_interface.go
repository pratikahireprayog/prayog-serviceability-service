package interfaces

import (
	"context"
	"time"

	"prayog-serviceability-service/internal/shared/models/v1"
)

// PartnerAdapter defines the interface that all partner adapters must implement
type PartnerAdapter interface {
	// GetPartnerCode returns the unique code for this partner
	GetPartnerCode() string

	// GetPartnerName returns the display name for this partner
	GetPartnerName() string

	// CheckServiceability checks if the partner can service the given request
	CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request) (*PartnerServiceabilityResult, error)

	// IsHealthy returns the health status of the partner adapter
	IsHealthy(ctx context.Context) bool

	// GetAdapterType returns the type of adapter (http, database, etc.)
	GetAdapterType() AdapterType
}

// AdapterType represents the type of partner adapter
type AdapterType string

const (
	AdapterTypeHTTP     AdapterType = "http"
	AdapterTypeDatabase AdapterType = "database"
	AdapterTypeGRPC     AdapterType = "grpc"
)

// PartnerServiceabilityResult represents the result from a partner serviceability check
type PartnerServiceabilityResult struct {
	PartnerCode   string                 `json:"partner_code"`
	PartnerName   string                 `json:"partner_name"`
	IsServiceable bool                   `json:"is_serviceable"`
	Services      []models.ServiceV2     `json:"services,omitempty"`
	Capabilities  map[string]interface{} `json:"capabilities,omitempty"`
	Error         error                  `json:"-"`
	ErrorMessage  *string                `json:"error,omitempty"`
	ResponseTime  time.Duration          `json:"response_time"`
	Rating        float64                `json:"rating"`
}

// PartnerAdapterFactory creates partner adapters based on partner codes
type PartnerAdapterFactory interface {
	// CreateAdapter creates a partner adapter for the given partner code
	CreateAdapter(partnerCode string) (PartnerAdapter, error)

	// GetSupportedPartners returns a list of supported partner codes
	GetSupportedPartners() []string

	// IsPartnerSupported checks if a partner code is supported
	IsPartnerSupported(partnerCode string) bool
}

// TokenManager interface for managing JWT tokens (used by Shipyaari adapter)
type TokenManager interface {
	// GetToken retrieves a valid token, refreshing if necessary
	GetToken(ctx context.Context) (string, error)

	// RefreshToken forcefully refreshes the token
	RefreshToken(ctx context.Context) (string, error)

	// IsTokenValid checks if the current token is valid
	IsTokenValid() bool

	// ClearToken clears the stored token
	ClearToken()
}

// HTTPPartnerAdapter represents an HTTP-based partner adapter
type HTTPPartnerAdapter interface {
	PartnerAdapter

	// GetBaseURL returns the base URL for the partner API
	GetBaseURL() string

	// GetTimeout returns the request timeout
	GetTimeout() time.Duration

	// GetRetryConfig returns retry configuration
	GetRetryConfig() RetryConfig
}

// DatabasePartnerAdapter represents a database-based partner adapter
type DatabasePartnerAdapter interface {
	PartnerAdapter

	// GetTableName returns the primary table name used by this adapter
	GetTableName() string

	// GetConnectionInfo returns connection information (for health checks)
	GetConnectionInfo() map[string]interface{}
}

// RetryConfig defines retry configuration for HTTP adapters
type RetryConfig struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// PartnerAuthConfig represents authentication configuration for partner adapters
type PartnerAuthConfig struct {
	Type         AuthType          `json:"type"`
	Credentials  map[string]string `json:"credentials"`
	TokenURL     string            `json:"token_url,omitempty"`
	ExpiryBuffer time.Duration     `json:"expiry_buffer,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
}

// AuthType represents the type of authentication
type AuthType string

const (
	AuthTypeNone   AuthType = "none"
	AuthTypeJWT    AuthType = "jwt"
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeBasic  AuthType = "basic"
)

// PartnerConfig represents configuration for a partner adapter
type PartnerConfig struct {
	Code      string            `json:"code"`
	Name      string            `json:"name"`
	BaseURL   string            `json:"base_url"`
	Auth      PartnerAuthConfig `json:"auth"`
	Retry     RetryConfig       `json:"retry"`
	Timeout   time.Duration     `json:"timeout"`
	RateLimit RateLimitConfig   `json:"rate_limit"`
	Enabled   bool              `json:"enabled"`
	Priority  int               `json:"priority"`
	Rating    float64           `json:"rating"`
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	Timeout           time.Duration `json:"timeout"`
}

// PartnerAdapterMetrics represents metrics for partner adapters
type PartnerAdapterMetrics struct {
	PartnerCode     string        `json:"partner_code"`
	TotalRequests   int64         `json:"total_requests"`
	SuccessfulReqs  int64         `json:"successful_requests"`
	FailedRequests  int64         `json:"failed_requests"`
	AverageLatency  time.Duration `json:"average_latency"`
	LastRequestTime time.Time     `json:"last_request_time"`
	HealthStatus    string        `json:"health_status"`
	ErrorRate       float64       `json:"error_rate"`
}

// MetricsCollector interface for collecting adapter metrics
type MetricsCollector interface {
	// RecordRequest records a request attempt
	RecordRequest(partnerCode string, duration time.Duration, success bool)

	// GetMetrics returns metrics for a specific partner
	GetMetrics(partnerCode string) *PartnerAdapterMetrics

	// GetAllMetrics returns metrics for all partners
	GetAllMetrics() map[string]*PartnerAdapterMetrics

	// Reset resets metrics for a partner
	Reset(partnerCode string)
}
