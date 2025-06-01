package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"prayog-serviceability-service/internal/infrastructure/cache"
	"prayog-serviceability-service/internal/infrastructure/errors"
	"prayog-serviceability-service/internal/infrastructure/fallback"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// ServiceDefinitionResolver implements the ServiceDefinitionResolver interface
// with caching, fallback support, and thread-safe operations
type ServiceDefinitionResolver struct {
	specService     interfaces.SpecificationServiceClient
	cache           *cache.ServiceDefinitionCache
	fallbackManager *fallback.FallbackManager
	validator       *ServiceDefinitionValidator
	transformer     *ServiceDefinitionTransformer
	config          config.ServiceDefinitionResolverConfig

	// Thread-safety
	mu                  sync.RWMutex
	categoryCache       map[string][]string // parcel category -> service types mapping
	categoryCacheExpiry time.Time
	categoryMu          sync.RWMutex

	// Circuit breaker state
	circuitState *CircuitBreakerState
	circuitMu    sync.RWMutex

	// Health tracking
	lastSuccessTime   time.Time
	consecutiveErrors int
	healthMu          sync.RWMutex
}

// ServiceDefinitionResolverConfig holds configuration for the resolver
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

// CircuitBreakerState tracks the state of the circuit breaker
type CircuitBreakerState struct {
	State           CircuitState `json:"state"`
	FailureCount    int          `json:"failure_count"`
	LastFailureTime time.Time    `json:"last_failure_time"`
	NextRetryTime   time.Time    `json:"next_retry_time"`
}

type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half_open"
)

// NewServiceDefinitionResolver creates a new service definition resolver
func NewServiceDefinitionResolver(
	specService interfaces.SpecificationServiceClient,
	cache *cache.ServiceDefinitionCache,
	fallbackManager *fallback.FallbackManager,
	config config.ServiceDefinitionResolverConfig,
) interfaces.ServiceDefinitionResolver {

	// Set defaults
	if config.CacheTTL <= 0 {
		config.CacheTTL = 30 * time.Minute
	}
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 5
	}
	if config.RecoveryTimeout <= 0 {
		config.RecoveryTimeout = 30 * time.Second
	}
	if config.MaxConcurrentRequests <= 0 {
		config.MaxConcurrentRequests = 10
	}
	if config.RequestTimeout <= 0 {
		config.RequestTimeout = 30 * time.Second
	}
	if config.CategoryCacheTTL <= 0 {
		config.CategoryCacheTTL = 15 * time.Minute
	}

	return &ServiceDefinitionResolver{
		specService:     specService,
		cache:           cache,
		fallbackManager: fallbackManager,
		validator:       NewServiceDefinitionValidator(config),
		transformer:     NewServiceDefinitionTransformer(config),
		config:          config,
		categoryCache:   make(map[string][]string),
		circuitState: &CircuitBreakerState{
			State: CircuitClosed,
		},
		lastSuccessTime: time.Now(),
	}
}

// GetParcelCategories retrieves all available parcel categories
func (sdr *ServiceDefinitionResolver) GetParcelCategories(ctx context.Context) ([]string, error) {
	// Check circuit breaker
	if sdr.isCircuitOpen() {
		return sdr.handleCircuitOpen(ctx, "GetParcelCategories")
	}

	// Try cache first
	if sdr.config.CacheEnabled {
		if categories, found := sdr.getCachedParcelCategories(ctx); found {
			return categories, nil
		}
	}

	// Fetch from service with fallback
	categories, err := sdr.fetchParcelCategoriesWithFallback(ctx)
	if err != nil {
		sdr.recordFailure(err)
		return nil, err
	}

	// Validate and transform
	if sdr.config.ValidateResponses {
		if err := sdr.validator.ValidateParcelCategories(categories); err != nil {
			sdr.recordFailure(err)
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	transformedCategories := sdr.transformer.TransformParcelCategories(categories)

	// Cache the result
	if sdr.config.CacheEnabled {
		sdr.cacheParcelCategories(ctx, transformedCategories)
	}

	sdr.recordSuccess()
	return transformedCategories, nil
}

// GetServiceTypes retrieves service types for a specific parcel category
func (sdr *ServiceDefinitionResolver) GetServiceTypes(ctx context.Context, parcelCategory string) ([]string, error) {
	// Check circuit breaker
	if sdr.isCircuitOpen() {
		return sdr.handleCircuitOpen(ctx, "GetServiceTypes")
	}

	// Validate input
	if err := sdr.validator.ValidateParcelCategory(parcelCategory); err != nil {
		return nil, fmt.Errorf("invalid parcel category: %w", err)
	}

	// Check category cache
	if serviceTypes, found := sdr.getCachedServiceTypes(parcelCategory); found {
		return serviceTypes, nil
	}

	// Fetch from service with fallback
	serviceTypes, err := sdr.fetchServiceTypesWithFallback(ctx, parcelCategory)
	if err != nil {
		sdr.recordFailure(err)
		return nil, err
	}

	// Validate and transform
	if sdr.config.ValidateResponses {
		if err := sdr.validator.ValidateServiceTypes(serviceTypes); err != nil {
			sdr.recordFailure(err)
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	transformedServiceTypes := sdr.transformer.TransformServiceTypes(serviceTypes)

	// Cache the result
	sdr.cacheServiceTypes(parcelCategory, transformedServiceTypes)

	sdr.recordSuccess()
	return transformedServiceTypes, nil
}

// GetOperationTypes retrieves all available operation types
func (sdr *ServiceDefinitionResolver) GetOperationTypes(ctx context.Context) ([]string, error) {
	// Check circuit breaker
	if sdr.isCircuitOpen() {
		return sdr.handleCircuitOpen(ctx, "GetOperationTypes")
	}

	// Fetch from service with fallback
	operationTypes, err := sdr.fetchOperationTypesWithFallback(ctx)
	if err != nil {
		sdr.recordFailure(err)
		return nil, err
	}

	// Validate and transform
	if sdr.config.ValidateResponses {
		if err := sdr.validator.ValidateOperationTypes(operationTypes); err != nil {
			sdr.recordFailure(err)
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	transformedOperationTypes := sdr.transformer.TransformOperationTypes(operationTypes)

	sdr.recordSuccess()
	return transformedOperationTypes, nil
}

// GetPaymentModes retrieves all available payment modes
func (sdr *ServiceDefinitionResolver) GetPaymentModes(ctx context.Context) ([]string, error) {
	// Check circuit breaker
	if sdr.isCircuitOpen() {
		return sdr.handleCircuitOpen(ctx, "GetPaymentModes")
	}

	// Fetch from service with fallback
	paymentModes, err := sdr.fetchPaymentModesWithFallback(ctx)
	if err != nil {
		sdr.recordFailure(err)
		return nil, err
	}

	// Validate and transform
	if sdr.config.ValidateResponses {
		if err := sdr.validator.ValidatePaymentModes(paymentModes); err != nil {
			sdr.recordFailure(err)
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	transformedPaymentModes := sdr.transformer.TransformPaymentModes(paymentModes)

	sdr.recordSuccess()
	return transformedPaymentModes, nil
}

// GetDeliveryModes retrieves all available delivery modes
func (sdr *ServiceDefinitionResolver) GetDeliveryModes(ctx context.Context) ([]string, error) {
	// Check circuit breaker
	if sdr.isCircuitOpen() {
		return sdr.handleCircuitOpen(ctx, "GetDeliveryModes")
	}

	// Fetch from service with fallback
	deliveryModes, err := sdr.fetchDeliveryModesWithFallback(ctx)
	if err != nil {
		sdr.recordFailure(err)
		return nil, err
	}

	// Validate and transform
	if sdr.config.ValidateResponses {
		if err := sdr.validator.ValidateDeliveryModes(deliveryModes); err != nil {
			sdr.recordFailure(err)
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	transformedDeliveryModes := sdr.transformer.TransformDeliveryModes(deliveryModes)

	sdr.recordSuccess()
	return transformedDeliveryModes, nil
}

// GetServiceDefinition retrieves a specific service definition
func (sdr *ServiceDefinitionResolver) GetServiceDefinition(ctx context.Context, serviceType string) (*interfaces.ServiceDefinition, error) {
	// Check circuit breaker
	if sdr.isCircuitOpen() {
		return sdr.handleCircuitOpenForServiceDef(ctx, serviceType)
	}

	// Validate input
	if err := sdr.validator.ValidateServiceType(serviceType); err != nil {
		return nil, fmt.Errorf("invalid service type: %w", err)
	}

	// Fetch from service with fallback
	serviceDef, err := sdr.fetchServiceDefinitionWithFallback(ctx, serviceType)
	if err != nil {
		sdr.recordFailure(err)
		return nil, err
	}

	// Validate and transform
	if sdr.config.ValidateResponses {
		if err := sdr.validator.ValidateServiceDefinition(serviceDef); err != nil {
			sdr.recordFailure(err)
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	transformedServiceDef := sdr.transformer.TransformServiceDefinition(serviceDef)

	sdr.recordSuccess()
	return transformedServiceDef, nil
}

// Private helper methods for fetching data with fallback

func (sdr *ServiceDefinitionResolver) fetchParcelCategoriesWithFallback(ctx context.Context) ([]string, error) {
	result, fallbackResult := sdr.fallbackManager.GetParcelCategoriesWithFallback(ctx, func(ctx context.Context) ([]string, error) {
		return sdr.extractParcelCategoriesFromCatalogs(ctx)
	})

	if fallbackResult != nil {
		// Log fallback usage
		sdr.logFallbackUsage("parcel_categories", fallbackResult)
	}

	return result, nil
}

func (sdr *ServiceDefinitionResolver) fetchServiceTypesWithFallback(ctx context.Context, parcelCategory string) ([]string, error) {
	result, fallbackResult := sdr.fallbackManager.GetServiceTypesWithFallback(ctx, parcelCategory, func(ctx context.Context) ([]string, error) {
		return sdr.extractServiceTypesFromCatalogs(ctx, parcelCategory)
	})

	if fallbackResult != nil {
		sdr.logFallbackUsage("service_types", fallbackResult)
	}

	return result, nil
}

func (sdr *ServiceDefinitionResolver) fetchOperationTypesWithFallback(ctx context.Context) ([]string, error) {
	result, fallbackResult := sdr.fallbackManager.GetOperationTypesWithFallback(ctx, func(ctx context.Context) ([]string, error) {
		return sdr.extractOperationTypesFromCatalogs(ctx)
	})

	if fallbackResult != nil {
		sdr.logFallbackUsage("operation_types", fallbackResult)
	}

	return result, nil
}

func (sdr *ServiceDefinitionResolver) fetchPaymentModesWithFallback(ctx context.Context) ([]string, error) {
	result, fallbackResult := sdr.fallbackManager.GetPaymentModesWithFallback(ctx, func(ctx context.Context) ([]string, error) {
		return sdr.extractPaymentModesFromCatalogs(ctx)
	})

	if fallbackResult != nil {
		sdr.logFallbackUsage("payment_modes", fallbackResult)
	}

	return result, nil
}

func (sdr *ServiceDefinitionResolver) fetchDeliveryModesWithFallback(ctx context.Context) ([]string, error) {
	result, fallbackResult := sdr.fallbackManager.GetDeliveryModesWithFallback(ctx, func(ctx context.Context) ([]string, error) {
		return sdr.extractDeliveryModesFromCatalogs(ctx)
	})

	if fallbackResult != nil {
		sdr.logFallbackUsage("delivery_modes", fallbackResult)
	}

	return result, nil
}

func (sdr *ServiceDefinitionResolver) fetchServiceDefinitionWithFallback(ctx context.Context, serviceType string) (*interfaces.ServiceDefinition, error) {
	result, fallbackResult := sdr.fallbackManager.GetServiceDefinitionWithFallback(ctx, serviceType, func(ctx context.Context) (*interfaces.ServiceDefinition, error) {
		return sdr.buildServiceDefinitionFromCatalogs(ctx, serviceType)
	})

	if fallbackResult != nil {
		sdr.logFallbackUsage("service_definition", fallbackResult)
	}

	return result, nil
}

// Thread-safe caching methods

func (sdr *ServiceDefinitionResolver) getCachedParcelCategories(ctx context.Context) ([]string, bool) {
	// This would integrate with the cache system
	// Implementation depends on your specific cache structure
	return nil, false
}

func (sdr *ServiceDefinitionResolver) cacheParcelCategories(ctx context.Context, categories []string) {
	// Cache implementation
}

func (sdr *ServiceDefinitionResolver) getCachedServiceTypes(parcelCategory string) ([]string, bool) {
	sdr.categoryMu.RLock()
	defer sdr.categoryMu.RUnlock()

	if time.Now().After(sdr.categoryCacheExpiry) {
		return nil, false
	}

	serviceTypes, found := sdr.categoryCache[parcelCategory]
	return serviceTypes, found
}

func (sdr *ServiceDefinitionResolver) cacheServiceTypes(parcelCategory string, serviceTypes []string) {
	sdr.categoryMu.Lock()
	defer sdr.categoryMu.Unlock()

	sdr.categoryCache[parcelCategory] = serviceTypes
	sdr.categoryCacheExpiry = time.Now().Add(sdr.config.CategoryCacheTTL)
}

// Circuit breaker methods

func (sdr *ServiceDefinitionResolver) isCircuitOpen() bool {
	if !sdr.config.CircuitBreakerEnabled {
		return false
	}

	sdr.circuitMu.RLock()
	defer sdr.circuitMu.RUnlock()

	switch sdr.circuitState.State {
	case CircuitOpen:
		// Check if it's time to try again
		if time.Now().After(sdr.circuitState.NextRetryTime) {
			sdr.circuitMu.RUnlock()
			sdr.circuitMu.Lock()
			sdr.circuitState.State = CircuitHalfOpen
			sdr.circuitMu.Unlock()
			sdr.circuitMu.RLock()
			return false
		}
		return true
	case CircuitHalfOpen:
		return false
	default:
		return false
	}
}

func (sdr *ServiceDefinitionResolver) recordSuccess() {
	sdr.healthMu.Lock()
	defer sdr.healthMu.Unlock()

	sdr.lastSuccessTime = time.Now()
	sdr.consecutiveErrors = 0

	if sdr.config.CircuitBreakerEnabled {
		sdr.circuitMu.Lock()
		defer sdr.circuitMu.Unlock()

		sdr.circuitState.State = CircuitClosed
		sdr.circuitState.FailureCount = 0
	}
}

func (sdr *ServiceDefinitionResolver) recordFailure(err error) {
	sdr.healthMu.Lock()
	defer sdr.healthMu.Unlock()

	sdr.consecutiveErrors++

	if sdr.config.CircuitBreakerEnabled {
		sdr.circuitMu.Lock()
		defer sdr.circuitMu.Unlock()

		sdr.circuitState.FailureCount++
		sdr.circuitState.LastFailureTime = time.Now()

		if sdr.circuitState.FailureCount >= sdr.config.FailureThreshold {
			sdr.circuitState.State = CircuitOpen
			sdr.circuitState.NextRetryTime = time.Now().Add(sdr.config.RecoveryTimeout)
		}
	}
}

func (sdr *ServiceDefinitionResolver) handleCircuitOpen(ctx context.Context, operation string) ([]string, error) {
	// Return fallback data or error
	return nil, errors.NewServiceDefinitionError(
		errors.ErrorTypeCircuitOpen,
		"CIRCUIT_OPEN",
		fmt.Sprintf("Circuit breaker is open for operation: %s", operation),
		nil,
	)
}

func (sdr *ServiceDefinitionResolver) handleCircuitOpenForServiceDef(ctx context.Context, serviceType string) (*interfaces.ServiceDefinition, error) {
	// Return fallback data or error
	return nil, errors.NewServiceDefinitionError(
		errors.ErrorTypeCircuitOpen,
		"CIRCUIT_OPEN",
		fmt.Sprintf("Circuit breaker is open for service definition: %s", serviceType),
		nil,
	)
}

// Data extraction methods (to be implemented based on your catalog structure)

func (sdr *ServiceDefinitionResolver) extractParcelCategoriesFromCatalogs(ctx context.Context) ([]string, error) {
	// Implementation will depend on your specific catalog structure
	// This is a placeholder for the actual implementation
	catalogs, err := sdr.specService.GetCatalogs(ctx)
	if err != nil {
		return nil, err
	}

	return sdr.transformer.ExtractParcelCategoriesFromCatalogs(catalogs), nil
}

func (sdr *ServiceDefinitionResolver) extractServiceTypesFromCatalogs(ctx context.Context, parcelCategory string) ([]string, error) {
	catalogs, err := sdr.specService.GetCatalogs(ctx)
	if err != nil {
		return nil, err
	}

	return sdr.transformer.ExtractServiceTypesFromCatalogs(catalogs, parcelCategory), nil
}

func (sdr *ServiceDefinitionResolver) extractOperationTypesFromCatalogs(ctx context.Context) ([]string, error) {
	catalogs, err := sdr.specService.GetCatalogs(ctx)
	if err != nil {
		return nil, err
	}

	return sdr.transformer.ExtractOperationTypesFromCatalogs(catalogs), nil
}

func (sdr *ServiceDefinitionResolver) extractPaymentModesFromCatalogs(ctx context.Context) ([]string, error) {
	catalogs, err := sdr.specService.GetCatalogs(ctx)
	if err != nil {
		return nil, err
	}

	return sdr.transformer.ExtractPaymentModesFromCatalogs(catalogs), nil
}

func (sdr *ServiceDefinitionResolver) extractDeliveryModesFromCatalogs(ctx context.Context) ([]string, error) {
	catalogs, err := sdr.specService.GetCatalogs(ctx)
	if err != nil {
		return nil, err
	}

	return sdr.transformer.ExtractDeliveryModesFromCatalogs(catalogs), nil
}

func (sdr *ServiceDefinitionResolver) buildServiceDefinitionFromCatalogs(ctx context.Context, serviceType string) (*interfaces.ServiceDefinition, error) {
	catalogs, err := sdr.specService.GetCatalogs(ctx)
	if err != nil {
		return nil, err
	}

	return sdr.transformer.BuildServiceDefinitionFromCatalogs(catalogs, serviceType), nil
}

// Utility methods

func (sdr *ServiceDefinitionResolver) logFallbackUsage(operation string, result *errors.FallbackResult) {
	// Log fallback usage for monitoring
	fmt.Printf("Fallback used for %s: strategy=%s, success=%t, message=%s\n",
		operation, result.Strategy, result.Success, result.Message)
}

// Health and metrics methods

func (sdr *ServiceDefinitionResolver) GetHealth() map[string]interface{} {
	sdr.healthMu.RLock()
	defer sdr.healthMu.RUnlock()

	sdr.circuitMu.RLock()
	defer sdr.circuitMu.RUnlock()

	return map[string]interface{}{
		"last_success_time":     sdr.lastSuccessTime,
		"consecutive_errors":    sdr.consecutiveErrors,
		"circuit_breaker_state": sdr.circuitState.State,
		"circuit_failure_count": sdr.circuitState.FailureCount,
		"is_healthy":            sdr.consecutiveErrors < sdr.config.FailureThreshold,
	}
}

func (sdr *ServiceDefinitionResolver) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"cache_stats":    sdr.cache.GetStats(),
		"fallback_stats": sdr.fallbackManager.GetMetrics(),
		"health":         sdr.GetHealth(),
	}
}
