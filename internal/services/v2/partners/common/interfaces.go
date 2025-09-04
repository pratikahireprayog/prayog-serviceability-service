package common

import (
	"context"
	"time"

	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
)

// PartnerAdapter defines the main interface that all partner adapters must implement
type PartnerAdapter interface {
	// Main functionality
	CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request, partnerInfo PartnerInfo) (*PartnerServiceabilityResult, error)

	// Health and monitoring
	IsHealthy(ctx context.Context) bool
	GetMetrics() *PartnerMetrics

	// Lifecycle management
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// HTTPPartnerAdapter extends PartnerAdapter for HTTP-based partners
type HTTPPartnerAdapter interface {
	PartnerAdapter
	HTTPClient
	Authenticator
}

// DatabasePartnerAdapter extends PartnerAdapter for database-based partners
type DatabasePartnerAdapter interface {
	PartnerAdapter
	DatabaseClient
}

// HTTPClient interface for HTTP-based partner communication
type HTTPClient interface {
	// Core HTTP operations
	Get(ctx context.Context, endpoint string, params map[string]string) (*HTTPResponse, error)
	Post(ctx context.Context, endpoint string, body interface{}) (*HTTPResponse, error)
	Put(ctx context.Context, endpoint string, body interface{}) (*HTTPResponse, error)
	Delete(ctx context.Context, endpoint string) (*HTTPResponse, error)

	// Configuration
	GetBaseURL() string
	GetTimeout() time.Duration
	SetRetryPolicy(policy RetryPolicy)
}

// Authenticator interface for handling partner authentication
type Authenticator interface {
	// Authentication methods
	Authenticate(ctx context.Context) error
	GetAuthHeaders() map[string]string
	IsAuthenticated() bool
	RefreshAuth(ctx context.Context) error

	// Token management (for JWT-based auth)
	GetToken() string
	IsTokenExpired() bool
}

// DatabaseClient interface for database-based partner operations
type DatabaseClient interface {
	// Query operations
	Query(ctx context.Context, query string, args ...interface{}) (*DatabaseResult, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (*DatabaseRow, error)

	// Transaction support
	BeginTx(ctx context.Context) (DatabaseTransaction, error)

	// Health check
	Ping(ctx context.Context) error
}

// DataTransformer interface for converting between partner and standard formats
type DataTransformer interface {
	// Request transformation
	TransformRequest(req *models.ServiceabilityV2Request) (interface{}, error)

	// Response transformation
	TransformResponse(partnerResponse interface{}) (*PartnerServiceabilityResult, error)

	// Error transformation
	TransformError(partnerError error) error
}

// RateLimiter interface for controlling request rates to partners
type RateLimiter interface {
	// Rate limiting
	Allow() bool
	Wait(ctx context.Context) error

	// Configuration
	SetRate(requestsPerSecond int)
	SetBurst(burst int)
}

// AdapterType represents the type of partner adapter
type AdapterType string

const (
	AdapterTypeHTTP     AdapterType = "http"
	AdapterTypeDatabase AdapterType = "database"
	AdapterTypeGRPC     AdapterType = "grpc"
	AdapterTypeHybrid   AdapterType = "hybrid"
)

// PartnerServiceabilityResult represents the result from a partner serviceability check
type PartnerServiceabilityResult struct {
	PartnerID       *uuid.UUID             `json:"partner_id,omitempty"`
	PartnerCode     string                 `json:"partner_code"`
	PartnerName     string                 `json:"partner_name,omitempty"`
	Services        []models.ServiceV2     `json:"standard_services,omitempty"`
	PartnerServices interface{}            `json:"services,omitempty"`
	Capabilities    map[string]interface{} `json:"capabilities,omitempty"`
	Error           error                  `json:"-"`
	ErrorMessage    *string                `json:"error,omitempty"`
	ResponseTime    time.Duration          `json:"response_time"`
	Rating          float64                `json:"rating,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PartnerMetrics represents performance and health metrics for a partner
type PartnerMetrics struct {
	PartnerCode         string        `json:"partner_code"`
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastRequestTime     time.Time     `json:"last_request_time"`
	HealthStatus        string        `json:"health_status"`
	ErrorRate           float64       `json:"error_rate"`
	Uptime              time.Duration `json:"uptime"`
}

// HTTPResponse represents an HTTP response from a partner
type HTTPResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
	Duration   time.Duration       `json:"duration"`
}

// DatabaseResult represents a database query result
type DatabaseResult struct {
	Rows     []map[string]interface{} `json:"rows"`
	RowCount int64                    `json:"row_count"`
	Duration time.Duration            `json:"duration"`
}

// DatabaseRow represents a single database row
type DatabaseRow struct {
	Data     map[string]interface{} `json:"data"`
	Duration time.Duration          `json:"duration"`
}

// DatabaseTransaction represents a database transaction
type DatabaseTransaction interface {
	Commit() error
	Rollback() error
	Query(ctx context.Context, query string, args ...interface{}) (*DatabaseResult, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (*DatabaseRow, error)
}

// RetryPolicy defines retry behavior for failed requests
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// PartnerConfig represents configuration for a partner adapter
type PartnerConfig struct {
	Code         string                 `json:"code"`
	Name         string                 `json:"name"`
	Type         AdapterType            `json:"type"`
	Enabled      bool                   `json:"enabled"`
	Priority     int                    `json:"priority"`
	Rating       float64                `json:"rating"`
	Timeout      time.Duration          `json:"timeout"`
	RetryPolicy  RetryPolicy            `json:"retry_policy"`
	RateLimit    RateLimitConfig        `json:"rate_limit"`
	Auth         AuthConfig             `json:"auth"`
	Endpoints    map[string]string      `json:"endpoints"`
	CustomConfig map[string]interface{} `json:"custom_config"`
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	Timeout           time.Duration `json:"timeout"`
}

// AuthConfig defines authentication configuration
type AuthConfig struct {
	Type        AuthType          `json:"type"`
	Credentials map[string]string `json:"credentials"`
	TokenURL    string            `json:"token_url,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// AuthType represents the type of authentication
type AuthType string

const (
	AuthTypeNone   AuthType = "none"
	AuthTypeJWT    AuthType = "jwt"
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeBasic  AuthType = "basic"
	AuthTypeOAuth2 AuthType = "oauth2"
)

// PartnerInfo holds partner information
type PartnerInfo struct {
	PartnerID   *uuid.UUID `json:"partner_id"`
	PartnerCode string     `json:"partner_code"`
}
