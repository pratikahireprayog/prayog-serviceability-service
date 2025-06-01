package utils

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"prayog-serviceability-service/internal/infrastructure/errors"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// PartnerCapabilityFallbackManager manages fallback strategies for partner capability data
type PartnerCapabilityFallbackManager interface {
	// Partner capability fallbacks
	GetPartnerCapabilityWithFallback(ctx context.Context, partnerID uint, primaryFetch func(ctx context.Context) (*interfaces.PartnerCapability, error)) (*interfaces.PartnerCapability, *errors.FallbackResult)
	GetPartnerServiceCapabilitiesWithFallback(ctx context.Context, partnerID uint, filters *interfaces.ServiceFilters, primaryFetch func(ctx context.Context) (*interfaces.PartnerServiceCapabilities, error)) (*interfaces.PartnerServiceCapabilities, *errors.FallbackResult)
	GetPartnersByLocationWithFallback(ctx context.Context, locationHierarchy *models.LocationHierarchy, primaryFetch func(ctx context.Context) ([]interfaces.PartnerCapability, error)) ([]interfaces.PartnerCapability, *errors.FallbackResult)

	// Service definition fallbacks
	GetServiceDefinitionsWithFallback(ctx context.Context, serviceTypes, parcelCategories []string, primaryFetch func(ctx context.Context) ([]interfaces.ServiceDefinition, error)) ([]interfaces.ServiceDefinition, *errors.FallbackResult)

	// Batch operation fallbacks
	GetMultiplePartnerCapabilitiesWithFallback(ctx context.Context, partnerIDs []uint, primaryFetch func(ctx context.Context) (map[uint]*interfaces.PartnerCapability, error)) (map[uint]*interfaces.PartnerCapability, *errors.FallbackResult)

	// Retry operations
	RetryWithBackoff(ctx context.Context, operation func(ctx context.Context) error) error

	// Metrics and management
	GetMetrics() *PartnerFallbackMetrics
	ResetMetrics()
	IsHealthy() bool
	GetHealth() PartnerFallbackHealth
}

// Implementation
type partnerCapabilityFallbackManager struct {
	cache         PartnerCapabilityCacheManager
	defaultData   *PartnerDefaultDataProvider
	config        PartnerFallbackConfig
	logger        *log.Logger
	metrics       *PartnerFallbackMetrics
	mu            sync.RWMutex
	healthTracker *FallbackHealthTracker
}

// Configuration for fallback behavior
type PartnerFallbackConfig struct {
	EnableStaleCache        bool          `yaml:"enable_stale_cache" json:"enable_stale_cache"`
	StaleDataTTL            time.Duration `yaml:"stale_data_ttl" json:"stale_data_ttl"`
	EnableDefaultData       bool          `yaml:"enable_default_data" json:"enable_default_data"`
	MaxRetryAttempts        int           `yaml:"max_retry_attempts" json:"max_retry_attempts"`
	RetryDelay              time.Duration `yaml:"retry_delay" json:"retry_delay"`
	FallbackTimeout         time.Duration `yaml:"fallback_timeout" json:"fallback_timeout"`
	EnableMetrics           bool          `yaml:"enable_metrics" json:"enable_metrics"`
	EnableHealthTracking    bool          `yaml:"enable_health_tracking" json:"enable_health_tracking"`
	CircuitBreakerThreshold int           `yaml:"circuit_breaker_threshold" json:"circuit_breaker_threshold"`
	PartialFailureThreshold float64       `yaml:"partial_failure_threshold" json:"partial_failure_threshold"`
	PreferCacheOverDefault  bool          `yaml:"prefer_cache_over_default" json:"prefer_cache_over_default"`
	EnableDegradedMode      bool          `yaml:"enable_degraded_mode" json:"enable_degraded_mode"`
}

// Metrics for tracking fallback usage
type PartnerFallbackMetrics struct {
	// Basic fallback counts
	CacheFallbacks   int64 `json:"cache_fallbacks"`
	DefaultFallbacks int64 `json:"default_fallbacks"`
	StaleFallbacks   int64 `json:"stale_fallbacks"`
	EmptyFallbacks   int64 `json:"empty_fallbacks"`
	RetryFallbacks   int64 `json:"retry_fallbacks"`
	TotalFallbacks   int64 `json:"total_fallbacks"`

	// Success/failure tracking
	SuccessfulRetries int64 `json:"successful_retries"`
	FailedRetries     int64 `json:"failed_retries"`
	PrimarySuccesses  int64 `json:"primary_successes"`
	PrimaryFailures   int64 `json:"primary_failures"`

	// Operation-specific metrics
	PartnerCapabilityFallbacks   int64 `json:"partner_capability_fallbacks"`
	ServiceCapabilitiesFallbacks int64 `json:"service_capabilities_fallbacks"`
	LocationPartnersFallbacks    int64 `json:"location_partners_fallbacks"`
	ServiceDefinitionsFallbacks  int64 `json:"service_definitions_fallbacks"`
	BatchOperationFallbacks      int64 `json:"batch_operation_fallbacks"`

	// Performance metrics
	AverageFallbackDuration time.Duration `json:"average_fallback_duration"`
	LastFallbackTime        time.Time     `json:"last_fallback_time"`
	FallbackSuccessRate     float64       `json:"fallback_success_rate"`

	// Health indicators
	ConsecutiveFailures int64     `json:"consecutive_failures"`
	LastHealthCheck     time.Time `json:"last_health_check"`
	HealthStatus        string    `json:"health_status"`
}

// Health information
type PartnerFallbackHealth struct {
	IsHealthy               bool      `json:"is_healthy"`
	Status                  string    `json:"status"`
	LastSuccessfulOperation time.Time `json:"last_successful_operation"`
	ConsecutiveErrors       int       `json:"consecutive_errors"`
	ErrorRate               float64   `json:"error_rate"`
	FallbackRate            float64   `json:"fallback_rate"`
	RecommendedActions      []string  `json:"recommended_actions"`
	Circuit                 string    `json:"circuit_status"`
	Degraded                bool      `json:"degraded_mode"`
}

// Health tracker for monitoring fallback health
type FallbackHealthTracker struct {
	lastSuccess       time.Time
	consecutiveErrors int
	errorCount        int64
	totalOperations   int64
	mu                sync.RWMutex
}

// Default data provider for fallback scenarios
type PartnerDefaultDataProvider struct {
	defaultPartners     []interfaces.PartnerCapability
	defaultCapabilities map[uint]*interfaces.PartnerServiceCapabilities
	defaultDefinitions  []interfaces.ServiceDefinition
	mu                  sync.RWMutex
}

// NewPartnerCapabilityFallbackManager creates a new fallback manager
func NewPartnerCapabilityFallbackManager(cache PartnerCapabilityCacheManager, config PartnerFallbackConfig, logger *log.Logger) PartnerCapabilityFallbackManager {
	manager := &partnerCapabilityFallbackManager{
		cache:       cache,
		defaultData: NewPartnerDefaultDataProvider(),
		config:      config,
		logger:      logger,
		metrics: &PartnerFallbackMetrics{
			LastHealthCheck: time.Now(),
			HealthStatus:    "healthy",
		},
		healthTracker: &FallbackHealthTracker{
			lastSuccess: time.Now(),
		},
	}

	return manager
}

// GetPartnerCapabilityWithFallback retrieves partner capability with fallback strategies
func (pfm *partnerCapabilityFallbackManager) GetPartnerCapabilityWithFallback(ctx context.Context, partnerID uint, primaryFetch func(ctx context.Context) (*interfaces.PartnerCapability, error)) (*interfaces.PartnerCapability, *errors.FallbackResult) {
	// Try primary fetch first
	capability, err := primaryFetch(ctx)
	if err == nil && capability != nil {
		pfm.recordSuccess()
		pfm.metrics.PrimarySuccesses++
		return capability, nil
	}

	pfm.metrics.PrimaryFailures++
	pfm.logError("Primary fetch failed for partner capability", err)

	// Try cache fallback
	if pfm.cache != nil {
		if capability, found := pfm.cache.GetPartnerCapability(ctx, partnerID); found {
			pfm.metrics.CacheFallbacks++
			pfm.metrics.TotalFallbacks++
			pfm.recordSuccess()
			pfm.logInfo(fmt.Sprintf("Cache fallback successful for partner %d", partnerID))

			return capability, errors.NewFallbackResult(errors.FallbackStrategyCache, true, capability, fmt.Sprintf("Cache fallback successful for partner %d", partnerID), err)
		}
	}

	// Try default data fallback
	if pfm.config.EnableDefaultData && pfm.defaultData != nil {
		capability := pfm.defaultData.GetDefaultPartnerCapability(partnerID)
		if capability != nil {
			pfm.metrics.DefaultFallbacks++
			pfm.metrics.TotalFallbacks++
			pfm.logInfo(fmt.Sprintf("Default data fallback used for partner %d", partnerID))

			return capability, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, capability, fmt.Sprintf("Default data fallback used for partner %d", partnerID), err)
		}
	}

	// All fallbacks failed
	pfm.recordFailure()
	pfm.metrics.EmptyFallbacks++
	pfm.metrics.TotalFallbacks++

	return nil, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, fmt.Sprintf("All fallbacks failed for partner %d", partnerID), err)
}

// GetPartnerServiceCapabilitiesWithFallback retrieves service capabilities with fallback
func (pfm *partnerCapabilityFallbackManager) GetPartnerServiceCapabilitiesWithFallback(ctx context.Context, partnerID uint, filters *interfaces.ServiceFilters, primaryFetch func(ctx context.Context) (*interfaces.PartnerServiceCapabilities, error)) (*interfaces.PartnerServiceCapabilities, *errors.FallbackResult) {
	// Try primary fetch
	capabilities, err := primaryFetch(ctx)
	if err == nil && capabilities != nil {
		pfm.recordSuccess()
		pfm.metrics.PrimarySuccesses++
		return capabilities, nil
	}

	pfm.metrics.PrimaryFailures++
	pfm.metrics.ServiceCapabilitiesFallbacks++

	// Try cache fallback
	if pfm.cache != nil {
		if capabilities, found := pfm.cache.GetPartnerServiceCapabilities(ctx, partnerID, filters); found {
			pfm.metrics.CacheFallbacks++
			pfm.recordSuccess()

			return capabilities, errors.NewFallbackResult(errors.FallbackStrategyCache, true, capabilities, fmt.Sprintf("Cache fallback successful for partner %d", partnerID), err)
		}
	}

	// Try default data
	if pfm.config.EnableDefaultData {
		capabilities := pfm.defaultData.GetDefaultPartnerServiceCapabilities(partnerID, filters)
		if capabilities != nil {
			pfm.metrics.DefaultFallbacks++

			return capabilities, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, capabilities, fmt.Sprintf("Default data fallback used for partner %d", partnerID), err)
		}
	}

	pfm.recordFailure()
	return nil, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, fmt.Sprintf("All fallbacks failed for partner %d", partnerID), err)
}

// GetPartnersByLocationWithFallback retrieves partners by location with fallback
func (pfm *partnerCapabilityFallbackManager) GetPartnersByLocationWithFallback(ctx context.Context, locationHierarchy *models.LocationHierarchy, primaryFetch func(ctx context.Context) ([]interfaces.PartnerCapability, error)) ([]interfaces.PartnerCapability, *errors.FallbackResult) {
	// Try primary fetch
	partners, err := primaryFetch(ctx)
	if err == nil && len(partners) > 0 {
		pfm.recordSuccess()
		pfm.metrics.PrimarySuccesses++
		return partners, nil
	}

	pfm.metrics.PrimaryFailures++
	pfm.metrics.LocationPartnersFallbacks++

	// Try cache fallback
	if pfm.cache != nil {
		if partners, found := pfm.cache.GetPartnersByLocation(ctx, locationHierarchy); found && len(partners) > 0 {
			pfm.metrics.CacheFallbacks++
			pfm.recordSuccess()

			return partners, errors.NewFallbackResult(errors.FallbackStrategyCache, true, partners, "Cache fallback successful for location", err)
		}
	}

	// Try default data
	if pfm.config.EnableDefaultData {
		partners := pfm.defaultData.GetDefaultPartnersByLocation(locationHierarchy)
		if len(partners) > 0 {
			pfm.metrics.DefaultFallbacks++

			return partners, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, partners, "Default data fallback used for location", err)
		}
	}

	pfm.recordFailure()
	return nil, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "All fallbacks failed for location", err)
}

// GetServiceDefinitionsWithFallback retrieves service definitions with fallback
func (pfm *partnerCapabilityFallbackManager) GetServiceDefinitionsWithFallback(ctx context.Context, serviceTypes, parcelCategories []string, primaryFetch func(ctx context.Context) ([]interfaces.ServiceDefinition, error)) ([]interfaces.ServiceDefinition, *errors.FallbackResult) {
	// Try primary fetch
	definitions, err := primaryFetch(ctx)
	if err == nil && len(definitions) > 0 {
		pfm.recordSuccess()
		pfm.metrics.PrimarySuccesses++
		return definitions, nil
	}

	pfm.metrics.PrimaryFailures++
	pfm.metrics.ServiceDefinitionsFallbacks++

	// Try cache fallback
	if pfm.cache != nil {
		if definitions, found := pfm.cache.GetServiceDefinitions(ctx, serviceTypes, parcelCategories); found && len(definitions) > 0 {
			pfm.metrics.CacheFallbacks++
			pfm.recordSuccess()

			return definitions, errors.NewFallbackResult(errors.FallbackStrategyCache, true, definitions, "Cache fallback successful for service definitions", err)
		}
	}

	// Try default data
	if pfm.config.EnableDefaultData {
		definitions := pfm.defaultData.GetDefaultServiceDefinitions(serviceTypes, parcelCategories)
		if len(definitions) > 0 {
			pfm.metrics.DefaultFallbacks++

			return definitions, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, definitions, "Default data fallback used for service definitions", err)
		}
	}

	pfm.recordFailure()
	return nil, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "All fallbacks failed for service definitions", err)
}

// GetMultiplePartnerCapabilitiesWithFallback retrieves multiple partner capabilities with fallback
func (pfm *partnerCapabilityFallbackManager) GetMultiplePartnerCapabilitiesWithFallback(ctx context.Context, partnerIDs []uint, primaryFetch func(ctx context.Context) (map[uint]*interfaces.PartnerCapability, error)) (map[uint]*interfaces.PartnerCapability, *errors.FallbackResult) {
	// Try primary fetch
	capabilities, err := primaryFetch(ctx)
	if err == nil && len(capabilities) > 0 {
		pfm.recordSuccess()
		pfm.metrics.PrimarySuccesses++
		return capabilities, nil
	}

	pfm.metrics.PrimaryFailures++
	pfm.metrics.BatchOperationFallbacks++

	// Try cache fallback
	if pfm.cache != nil {
		if capabilities, _ := pfm.cache.GetMultiplePartnerCapabilities(ctx, partnerIDs); len(capabilities) > 0 {
			pfm.metrics.CacheFallbacks++
			pfm.recordSuccess()

			return capabilities, errors.NewFallbackResult(errors.FallbackStrategyCache, true, capabilities, "Cache fallback successful for multiple partners", err)
		}
	}

	// Try default data
	if pfm.config.EnableDefaultData {
		capabilities := pfm.defaultData.GetDefaultMultiplePartnerCapabilities(partnerIDs)
		if len(capabilities) > 0 {
			pfm.metrics.DefaultFallbacks++

			return capabilities, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, capabilities, "Default data fallback used for multiple partners", err)
		}
	}

	pfm.recordFailure()
	return map[uint]*interfaces.PartnerCapability{}, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "All fallbacks failed for multiple partners", err)
}

// RetryWithBackoff performs retry with exponential backoff
func (pfm *partnerCapabilityFallbackManager) RetryWithBackoff(ctx context.Context, operation func(ctx context.Context) error) error {
	maxAttempts := pfm.config.MaxRetryAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := operation(ctx)
		if err == nil {
			pfm.metrics.SuccessfulRetries++
			return nil
		}

		lastErr = err
		if !pfm.isRetryableError(err) {
			break
		}

		if attempt < maxAttempts-1 {
			delay := pfm.config.RetryDelay * time.Duration(1<<attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	pfm.metrics.FailedRetries++
	return lastErr
}

// GetMetrics returns current fallback metrics
func (pfm *partnerCapabilityFallbackManager) GetMetrics() *PartnerFallbackMetrics {
	pfm.mu.RLock()
	defer pfm.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *pfm.metrics
	return &metrics
}

// ResetMetrics resets all metrics counters
func (pfm *partnerCapabilityFallbackManager) ResetMetrics() {
	pfm.mu.Lock()
	defer pfm.mu.Unlock()

	pfm.metrics = &PartnerFallbackMetrics{
		LastHealthCheck: time.Now(),
		HealthStatus:    "healthy",
	}
}

// IsHealthy returns current health status
func (pfm *partnerCapabilityFallbackManager) IsHealthy() bool {
	pfm.mu.RLock()
	defer pfm.mu.RUnlock()

	// Check various health indicators
	if pfm.metrics.ConsecutiveFailures > int64(pfm.config.CircuitBreakerThreshold) {
		return false
	}

	// Check error rate
	total := pfm.metrics.PrimarySuccesses + pfm.metrics.PrimaryFailures
	if total > 0 {
		errorRate := float64(pfm.metrics.PrimaryFailures) / float64(total)
		if errorRate > pfm.config.PartialFailureThreshold {
			return false
		}
	}

	return true
}

// GetHealth returns detailed health information
func (pfm *partnerCapabilityFallbackManager) GetHealth() PartnerFallbackHealth {
	pfm.mu.RLock()
	defer pfm.mu.RUnlock()

	total := pfm.metrics.PrimarySuccesses + pfm.metrics.PrimaryFailures
	errorRate := 0.0
	fallbackRate := 0.0

	if total > 0 {
		errorRate = float64(pfm.metrics.PrimaryFailures) / float64(total)
		fallbackRate = float64(pfm.metrics.TotalFallbacks) / float64(total)
	}

	isHealthy := pfm.IsHealthy()
	status := "healthy"
	if !isHealthy {
		status = "unhealthy"
	}

	var recommendations []string
	if errorRate > 0.5 {
		recommendations = append(recommendations, "High error rate detected")
	}
	if pfm.metrics.ConsecutiveFailures > 5 {
		recommendations = append(recommendations, "Multiple consecutive failures")
	}
	if fallbackRate > 0.8 {
		recommendations = append(recommendations, "High fallback usage")
	}

	return PartnerFallbackHealth{
		IsHealthy:               isHealthy,
		Status:                  status,
		LastSuccessfulOperation: pfm.healthTracker.lastSuccess,
		ConsecutiveErrors:       int(pfm.metrics.ConsecutiveFailures),
		ErrorRate:               errorRate,
		FallbackRate:            fallbackRate,
		RecommendedActions:      recommendations,
		Circuit:                 "closed",
		Degraded:                errorRate > pfm.config.PartialFailureThreshold,
	}
}

// Helper methods

func (pfm *partnerCapabilityFallbackManager) recordSuccess() {
	pfm.mu.Lock()
	defer pfm.mu.Unlock()

	pfm.healthTracker.lastSuccess = time.Now()
	pfm.healthTracker.consecutiveErrors = 0
	pfm.metrics.ConsecutiveFailures = 0
	pfm.healthTracker.totalOperations++
}

func (pfm *partnerCapabilityFallbackManager) recordFailure() {
	pfm.mu.Lock()
	defer pfm.mu.Unlock()

	pfm.healthTracker.consecutiveErrors++
	pfm.healthTracker.errorCount++
	pfm.healthTracker.totalOperations++
	pfm.metrics.ConsecutiveFailures++
}

func (pfm *partnerCapabilityFallbackManager) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	retryablePatterns := []string{
		"timeout", "connection", "network", "unavailable", "5", "503", "502", "504",
	}

	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

func (pfm *partnerCapabilityFallbackManager) logInfo(message string) {
	if pfm.logger != nil {
		pfm.logger.Printf("[PARTNER_FALLBACK] INFO: %s", message)
	}
}

func (pfm *partnerCapabilityFallbackManager) logError(message string, err error) {
	if pfm.logger != nil {
		if err != nil {
			pfm.logger.Printf("[PARTNER_FALLBACK] ERROR: %s - %v", message, err)
		} else {
			pfm.logger.Printf("[PARTNER_FALLBACK] ERROR: %s", message)
		}
	}
}

// Helper function to check if string contains substrings (case-insensitive)
func contains(str string, substrings ...string) bool {
	for _, substring := range substrings {
		if len(str) >= len(substring) {
			for i := 0; i <= len(str)-len(substring); i++ {
				match := true
				for j := 0; j < len(substring); j++ {
					if str[i+j] != substring[j] && str[i+j] != substring[j]-32 && str[i+j] != substring[j]+32 {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}

// NewPartnerDefaultDataProvider creates a new partner default data provider
func NewPartnerDefaultDataProvider() *PartnerDefaultDataProvider {
	provider := &PartnerDefaultDataProvider{
		defaultCapabilities: make(map[uint]*interfaces.PartnerServiceCapabilities),
	}
	provider.initializeDefaultData()
	return provider
}

// GetDefaultPartnerCapability returns default partner capability
func (pddp *PartnerDefaultDataProvider) GetDefaultPartnerCapability(partnerID uint) *interfaces.PartnerCapability {
	pddp.mu.RLock()
	defer pddp.mu.RUnlock()

	for _, partner := range pddp.defaultPartners {
		if partner.PartnerID == partnerID {
			return &partner
		}
	}

	// Return a generic default if specific partner not found
	return &interfaces.PartnerCapability{
		PartnerID:        partnerID,
		PartnerName:      fmt.Sprintf("Partner %d", partnerID),
		IsActive:         true,
		ServiceTypes:     []string{"Standard"},
		ParcelCategories: []string{"courier"},
		OperationTypes:   []string{"pickup", "delivery"},
		PaymentModes:     []string{"ONLINE", "COD"},
		DeliveryModes:    []string{"AIR", "SURFACE"},
	}
}

// GetDefaultPartnerServiceCapabilities returns default partner service capabilities
func (pddp *PartnerDefaultDataProvider) GetDefaultPartnerServiceCapabilities(partnerID uint, filters *interfaces.ServiceFilters) *interfaces.PartnerServiceCapabilities {
	pddp.mu.RLock()
	defer pddp.mu.RUnlock()

	if capability, exists := pddp.defaultCapabilities[partnerID]; exists {
		return capability
	}

	// Return generic default
	return &interfaces.PartnerServiceCapabilities{
		PartnerID: partnerID,
		Capabilities: []interfaces.ServiceCapability{
			{
				ServiceType:    "Standard",
				ParcelCategory: "courier",
				OperationTypes: []string{"pickup", "delivery"},
				PaymentModes:   []string{"ONLINE", "COD"},
				DeliveryModes:  []string{"AIR", "SURFACE"},
				IsActive:       true,
				Rating:         5.0,
			},
		},
		Preferences: []interfaces.PartnerPreference{
			{
				ServiceType:     "Standard",
				ParcelCategory:  "courier",
				IsPreferred:     true,
				Priority:        3,
				EffectiveRating: 5.0,
			},
		},
	}
}

// GetDefaultPartnersByLocation returns default partners for a location
func (pddp *PartnerDefaultDataProvider) GetDefaultPartnersByLocation(locationHierarchy *models.LocationHierarchy) []interfaces.PartnerCapability {
	pddp.mu.RLock()
	defer pddp.mu.RUnlock()

	// Return a subset of default partners for any location
	if len(pddp.defaultPartners) > 0 {
		return pddp.defaultPartners[:min(len(pddp.defaultPartners), 3)]
	}

	return []interfaces.PartnerCapability{}
}

// GetDefaultServiceDefinitions returns default service definitions
func (pddp *PartnerDefaultDataProvider) GetDefaultServiceDefinitions(serviceTypes, parcelCategories []string) []interfaces.ServiceDefinition {
	pddp.mu.RLock()
	defer pddp.mu.RUnlock()

	return pddp.defaultDefinitions
}

// GetDefaultMultiplePartnerCapabilities returns default capabilities for multiple partners
func (pddp *PartnerDefaultDataProvider) GetDefaultMultiplePartnerCapabilities(partnerIDs []uint) map[uint]*interfaces.PartnerCapability {
	pddp.mu.RLock()
	defer pddp.mu.RUnlock()

	result := make(map[uint]*interfaces.PartnerCapability)
	for _, partnerID := range partnerIDs {
		if capability := pddp.GetDefaultPartnerCapability(partnerID); capability != nil {
			result[partnerID] = capability
		}
	}

	return result
}

// initializeDefaultData sets up default fallback data
func (pddp *PartnerDefaultDataProvider) initializeDefaultData() {
	// Initialize default partners
	pddp.defaultPartners = []interfaces.PartnerCapability{
		{
			PartnerID:        1,
			PartnerName:      "Default Courier Partner",
			IsActive:         true,
			ServiceTypes:     []string{"Standard", "Express"},
			ParcelCategories: []string{"courier"},
			OperationTypes:   []string{"pickup", "delivery"},
			PaymentModes:     []string{"ONLINE", "COD"},
			DeliveryModes:    []string{"AIR", "SURFACE"},
		},
		{
			PartnerID:        2,
			PartnerName:      "Default Ecom Partner",
			IsActive:         true,
			ServiceTypes:     []string{"SDD", "NDD"},
			ParcelCategories: []string{"ecom"},
			OperationTypes:   []string{"pickup", "delivery"},
			PaymentModes:     []string{"ONLINE", "COD"},
			DeliveryModes:    []string{"AIR", "SURFACE"},
		},
		{
			PartnerID:        3,
			PartnerName:      "Default Cargo Partner",
			IsActive:         true,
			ServiceTypes:     []string{"vayuquick", "vayuquickpro"},
			ParcelCategories: []string{"cargo"},
			OperationTypes:   []string{"pickup", "delivery"},
			PaymentModes:     []string{"ONLINE"},
			DeliveryModes:    []string{"SURFACE", "RAIL"},
		},
	}

	// Initialize default service definitions
	pddp.defaultDefinitions = []interfaces.ServiceDefinition{
		{
			ServiceType:       "Standard",
			ParcelCategory:    "courier",
			DefaultOperations: []string{"pickup", "delivery"},
			DefaultPayments:   []string{"ONLINE", "COD"},
			DefaultDelivery:   []string{"AIR", "SURFACE"},
			Description:       "Standard courier service",
		},
		{
			ServiceType:       "Express",
			ParcelCategory:    "courier",
			DefaultOperations: []string{"pickup", "delivery"},
			DefaultPayments:   []string{"ONLINE", "COD"},
			DefaultDelivery:   []string{"AIR"},
			Description:       "Express courier service",
		},
		{
			ServiceType:       "SDD",
			ParcelCategory:    "ecom",
			DefaultOperations: []string{"pickup", "delivery"},
			DefaultPayments:   []string{"ONLINE", "COD"},
			DefaultDelivery:   []string{"AIR"},
			Description:       "Same day delivery",
		},
		{
			ServiceType:       "NDD",
			ParcelCategory:    "ecom",
			DefaultOperations: []string{"pickup", "delivery"},
			DefaultPayments:   []string{"ONLINE", "COD"},
			DefaultDelivery:   []string{"AIR", "SURFACE"},
			Description:       "Next day delivery",
		},
	}

	// Initialize default capabilities for known partners
	for _, partner := range pddp.defaultPartners {
		pddp.defaultCapabilities[partner.PartnerID] = &interfaces.PartnerServiceCapabilities{
			PartnerID: partner.PartnerID,
			Capabilities: []interfaces.ServiceCapability{
				{
					ServiceType:    partner.ServiceTypes[0],
					ParcelCategory: partner.ParcelCategories[0],
					OperationTypes: partner.OperationTypes,
					PaymentModes:   partner.PaymentModes,
					DeliveryModes:  partner.DeliveryModes,
					IsActive:       true,
					Rating:         5.0,
				},
			},
			Preferences: []interfaces.PartnerPreference{
				{
					ServiceType:     partner.ServiceTypes[0],
					ParcelCategory:  partner.ParcelCategories[0],
					IsPreferred:     true,
					Priority:        3,
					EffectiveRating: 5.0,
				},
			},
		}
	}
}

// Helper function to find minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
