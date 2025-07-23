package common

import (
	"context"
	"fmt"
	"sync"
	"time"

	"prayog-serviceability-service/internal/shared/models/v1"
)

// BaseAdapter provides common functionality for all partner adapters
type BaseAdapter struct {
	config      *PartnerConfig
	metrics     *PartnerMetrics
	initialized bool
	startTime   time.Time
	mu          sync.RWMutex
}

// NewBaseAdapter creates a new base adapter
func NewBaseAdapter(config *PartnerConfig) *BaseAdapter {
	return &BaseAdapter{
		config:    config,
		startTime: time.Now(),
		metrics: &PartnerMetrics{
			PartnerCode:         config.Code,
			TotalRequests:       0,
			SuccessfulRequests:  0,
			FailedRequests:      0,
			AverageResponseTime: 0,
			LastRequestTime:     time.Time{},
			HealthStatus:        "healthy",
			ErrorRate:           0.0,
			Uptime:              0,
		},
	}
}

// GetPartnerCode returns the partner code (interface compatibility only)
// TODO: This should be removed when all code uses database values
func (b *BaseAdapter) GetPartnerCode() string {
	return b.config.Code
}

// GetPartnerName returns the partner name (interface compatibility only)
// TODO: This should be removed when all code uses database values
func (b *BaseAdapter) GetPartnerName() string {
	return b.config.Name
}

// GetAdapterType returns the adapter type
func (b *BaseAdapter) GetAdapterType() AdapterType {
	return b.config.Type
}

// GetMetrics returns current metrics
func (b *BaseAdapter) GetMetrics() *PartnerMetrics {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Update uptime
	b.metrics.Uptime = time.Since(b.startTime)

	// Calculate error rate
	if b.metrics.TotalRequests > 0 {
		b.metrics.ErrorRate = float64(b.metrics.FailedRequests) / float64(b.metrics.TotalRequests)
	}

	return b.metrics
}

// IsHealthy returns the health status
func (b *BaseAdapter) IsHealthy(ctx context.Context) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.metrics.HealthStatus == "healthy"
}

// Initialize initializes the adapter
func (b *BaseAdapter) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	// Basic initialization
	b.initialized = true
	b.metrics.HealthStatus = "healthy"

	return nil
}

// Shutdown shuts down the adapter
func (b *BaseAdapter) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	b.metrics.HealthStatus = "shutdown"

	return nil
}

// RecordRequest records a request in the metrics
func (b *BaseAdapter) RecordRequest(duration time.Duration, success bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.metrics.TotalRequests++
	b.metrics.LastRequestTime = time.Now()

	if success {
		b.metrics.SuccessfulRequests++
	} else {
		b.metrics.FailedRequests++
	}

	// Update average response time
	if b.metrics.TotalRequests == 1 {
		b.metrics.AverageResponseTime = duration
	} else {
		// Calculate moving average
		totalTime := b.metrics.AverageResponseTime * time.Duration(b.metrics.TotalRequests-1)
		b.metrics.AverageResponseTime = (totalTime + duration) / time.Duration(b.metrics.TotalRequests)
	}
}

// SetHealthStatus sets the health status
func (b *BaseAdapter) SetHealthStatus(status string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.metrics.HealthStatus = status
}

// GetConfig returns the partner configuration
func (b *BaseAdapter) GetConfig() *PartnerConfig {
	return b.config
}

// GenericAdapter implements a basic adapter for partners without specific implementations
type GenericAdapter struct {
	*BaseAdapter
}

// NewGenericAdapter creates a new generic adapter
func NewGenericAdapter(partnerCode, partnerName string) PartnerAdapter {
	config := &PartnerConfig{
		Code:    partnerCode,
		Name:    partnerName,
		Type:    AdapterTypeHTTP,
		Enabled: true,
		// Rating comes from database, not hardcoded
		Timeout: 30 * time.Second,
	}

	return &GenericAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}
}

// CheckServiceability implements a generic serviceability check
func (g *GenericAdapter) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request, partnerInfo PartnerInfo) (*PartnerServiceabilityResult, error) {
	startTime := time.Now()

	// Record the request
	defer func() {
		duration := time.Since(startTime)
		g.RecordRequest(duration, false) // Always false for generic adapter
	}()

	// Generic adapters always return "implementation pending"
	errorMsg := "implementation pending"

	return &PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     []models.ServiceV2{},
		Capabilities: make(map[string]interface{}),
		ErrorMessage: &errorMsg,
		ResponseTime: time.Since(startTime),
		// Rating comes from database, not hardcoded
		Metadata: map[string]interface{}{
			"adapter_type": "generic",
			"status":       "pending_implementation",
		},
	}, nil
}

// PartnerCodeAdapter wraps another adapter and preserves the original partner code
type PartnerCodeAdapter struct {
	PartnerAdapter
	originalInfo PartnerInfo
}

// NewPartnerCodeAdapter creates a wrapper that preserves the original partner code
func NewPartnerCodeAdapter(originalInfo PartnerInfo, wrappedAdapter PartnerAdapter) PartnerAdapter {
	return &PartnerCodeAdapter{
		PartnerAdapter: wrappedAdapter,
		originalInfo:   originalInfo,
	}
}

// GetPartnerCode returns the original partner code (from database)
func (p *PartnerCodeAdapter) GetPartnerCode() string {
	return p.originalInfo.PartnerCode
}

// CheckServiceability wraps the underlying adapter but preserves the original partner code
func (p *PartnerCodeAdapter) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request, partnerInfo PartnerInfo) (*PartnerServiceabilityResult, error) {
	result, err := p.PartnerAdapter.CheckServiceability(ctx, req, partnerInfo)
	if err != nil {
		return nil, err
	}

	// Preserve the original partner code in the result
	if result != nil {
		result.PartnerID = p.originalInfo.PartnerID
		result.PartnerCode = p.originalInfo.PartnerCode
	}

	return result, nil
}

// HTTPBaseAdapter provides common HTTP functionality
type HTTPBaseAdapter struct {
	*BaseAdapter
	httpClient HTTPClient
	auth       Authenticator
}

// NewHTTPBaseAdapter creates a new HTTP base adapter
func NewHTTPBaseAdapter(config *PartnerConfig) *HTTPBaseAdapter {
	return &HTTPBaseAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}
}

// SetHTTPClient sets the HTTP client
func (h *HTTPBaseAdapter) SetHTTPClient(client HTTPClient) {
	h.httpClient = client
}

// SetAuthenticator sets the authenticator
func (h *HTTPBaseAdapter) SetAuthenticator(auth Authenticator) {
	h.auth = auth
}

// GetHTTPClient returns the HTTP client
func (h *HTTPBaseAdapter) GetHTTPClient() HTTPClient {
	return h.httpClient
}

// GetAuthenticator returns the authenticator
func (h *HTTPBaseAdapter) GetAuthenticator() Authenticator {
	return h.auth
}

// DatabaseBaseAdapter provides common database functionality
type DatabaseBaseAdapter struct {
	*BaseAdapter
	dbClient DatabaseClient
}

// NewDatabaseBaseAdapter creates a new database base adapter
func NewDatabaseBaseAdapter(config *PartnerConfig) *DatabaseBaseAdapter {
	return &DatabaseBaseAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}
}

// SetDatabaseClient sets the database client
func (d *DatabaseBaseAdapter) SetDatabaseClient(client DatabaseClient) {
	d.dbClient = client
}

// GetDatabaseClient returns the database client
func (d *DatabaseBaseAdapter) GetDatabaseClient() DatabaseClient {
	return d.dbClient
}

// Validate validates a serviceability request
func ValidateServiceabilityRequest(req *models.ServiceabilityV2Request) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Check postal code requirements
	hasSingleCode := req.PostalCode != nil && *req.PostalCode != ""
	hasSourceDestination := req.SourcePostalCode != nil && *req.SourcePostalCode != "" &&
		req.DestinationPostalCode != nil && *req.DestinationPostalCode != ""

	// Handle case where postal_code is used as destination with source_postal_code
	hasPostalCodeAsDestination := req.PostalCode != nil && *req.PostalCode != "" &&
		req.SourcePostalCode != nil && *req.SourcePostalCode != ""

	if !hasSingleCode && !hasSourceDestination && !hasPostalCodeAsDestination {
		return fmt.Errorf("must provide either: postal_code only, or both source_postal_code and destination_postal_code, or postal_code (destination) with source_postal_code")
	}

	// Don't allow conflicting postal code configurations
	if hasSingleCode && hasSourceDestination {
		return fmt.Errorf("provide either postal_code only or source/destination postal codes, not both")
	}

	if hasPostalCodeAsDestination && req.DestinationPostalCode != nil && *req.DestinationPostalCode != "" {
		return fmt.Errorf("when using postal_code as destination with source_postal_code, do not provide destination_postal_code")
	}

	return nil
}

// GetPartnerConfigDefaults returns default configuration for a partner
func GetPartnerConfigDefaults(partnerCode, partnerName string, adapterType AdapterType) *PartnerConfig {
	return &PartnerConfig{
		Code:     partnerCode,
		Name:     partnerName,
		Type:     adapterType,
		Enabled:  true,
		Priority: 100,
		// Rating comes from database, not hardcoded
		Timeout: 30 * time.Second,
		RetryPolicy: RetryPolicy{
			MaxRetries:    3,
			InitialDelay:  1 * time.Second,
			MaxDelay:      10 * time.Second,
			BackoffFactor: 2.0,
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: 10,
			BurstSize:         20,
			Timeout:           5 * time.Second,
		},
		Auth: AuthConfig{
			Type: AuthTypeNone,
		},
		Endpoints:    make(map[string]string),
		CustomConfig: make(map[string]interface{}),
	}
}
