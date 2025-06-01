package fallback

import (
	"context"
	"log"
	"time"

	"prayog-serviceability-service/internal/infrastructure/cache"
	"prayog-serviceability-service/internal/infrastructure/errors"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// FallbackManager manages fallback strategies for service definition resolution
type FallbackManager struct {
	cache       *cache.ServiceDefinitionCache
	defaultData *DefaultDataProvider
	config      FallbackConfig
	logger      *log.Logger
	metrics     *FallbackMetrics
}

// FallbackConfig holds configuration for fallback strategies
type FallbackConfig struct {
	EnableStaleCache  bool          `yaml:"enable_stale_cache" json:"enable_stale_cache"`
	StaleDataTTL      time.Duration `yaml:"stale_data_ttl" json:"stale_data_ttl"`
	EnableDefaultData bool          `yaml:"enable_default_data" json:"enable_default_data"`
	MaxRetryAttempts  int           `yaml:"max_retry_attempts" json:"max_retry_attempts"`
	RetryDelay        time.Duration `yaml:"retry_delay" json:"retry_delay"`
	FallbackTimeout   time.Duration `yaml:"fallback_timeout" json:"fallback_timeout"`
	EnableMetrics     bool          `yaml:"enable_metrics" json:"enable_metrics"`
}

// FallbackMetrics tracks fallback usage statistics
type FallbackMetrics struct {
	CacheFallbacks    int64 `json:"cache_fallbacks"`
	DefaultFallbacks  int64 `json:"default_fallbacks"`
	StaleFallbacks    int64 `json:"stale_fallbacks"`
	EmptyFallbacks    int64 `json:"empty_fallbacks"`
	RetryFallbacks    int64 `json:"retry_fallbacks"`
	SuccessfulRetries int64 `json:"successful_retries"`
	FailedRetries     int64 `json:"failed_retries"`
	TotalFallbacks    int64 `json:"total_fallbacks"`
}

// NewFallbackManager creates a new fallback manager
func NewFallbackManager(cache *cache.ServiceDefinitionCache, config FallbackConfig, logger *log.Logger) *FallbackManager {
	// Set defaults
	if config.StaleDataTTL <= 0 {
		config.StaleDataTTL = 24 * time.Hour
	}
	if config.MaxRetryAttempts <= 0 {
		config.MaxRetryAttempts = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 1 * time.Second
	}
	if config.FallbackTimeout <= 0 {
		config.FallbackTimeout = 5 * time.Second
	}

	return &FallbackManager{
		cache:       cache,
		defaultData: NewDefaultDataProvider(),
		config:      config,
		logger:      logger,
		metrics:     &FallbackMetrics{},
	}
}

// GetSpecDefinitionsWithFallback attempts to get spec definitions with fallback strategies
func (fm *FallbackManager) GetSpecDefinitionsWithFallback(ctx context.Context, primaryFetch func(ctx context.Context) ([]interfaces.SpecDefinition, error)) ([]interfaces.SpecDefinition, *errors.FallbackResult) {
	// Try primary fetch first
	data, err := primaryFetch(ctx)
	if err == nil {
		return data, nil
	}

	fm.logError("Primary fetch failed for spec definitions", err)

	// Try cache fallback
	if cachedData, found := fm.cache.GetSpecDefinitions(ctx); found {
		fm.metrics.CacheFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using cached spec definitions as fallback")
		return cachedData, errors.NewFallbackResult(errors.FallbackStrategyCache, true, cachedData, "Used cached data", err)
	}

	// Try default data fallback
	if fm.config.EnableDefaultData {
		defaultData := fm.defaultData.GetDefaultSpecDefinitions()
		fm.metrics.DefaultFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using default spec definitions as fallback")
		return defaultData, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, defaultData, "Used default data", err)
	}

	// Return empty result as last resort
	fm.metrics.EmptyFallbacks++
	fm.metrics.TotalFallbacks++
	fm.logError("All fallback strategies failed for spec definitions", err)
	return []interfaces.SpecDefinition{}, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "No fallback data available", err)
}

// GetCatalogsWithFallback attempts to get catalogs with fallback strategies
func (fm *FallbackManager) GetCatalogsWithFallback(ctx context.Context, primaryFetch func(ctx context.Context) ([]interfaces.Catalog, error)) ([]interfaces.Catalog, *errors.FallbackResult) {
	// Try primary fetch first
	data, err := primaryFetch(ctx)
	if err == nil {
		return data, nil
	}

	fm.logError("Primary fetch failed for catalogs", err)

	// Try cache fallback
	if cachedData, found := fm.cache.GetCatalogs(ctx); found {
		fm.metrics.CacheFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using cached catalogs as fallback")
		return cachedData, errors.NewFallbackResult(errors.FallbackStrategyCache, true, cachedData, "Used cached data", err)
	}

	// Try default data fallback
	if fm.config.EnableDefaultData {
		defaultData := fm.defaultData.GetDefaultCatalogs()
		fm.metrics.DefaultFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using default catalogs as fallback")
		return defaultData, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, defaultData, "Used default data", err)
	}

	// Return empty result as last resort
	fm.metrics.EmptyFallbacks++
	fm.metrics.TotalFallbacks++
	fm.logError("All fallback strategies failed for catalogs", err)
	return []interfaces.Catalog{}, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "No fallback data available", err)
}

// GetEntitySpecificationsWithFallback attempts to get entity specifications with fallback strategies
func (fm *FallbackManager) GetEntitySpecificationsWithFallback(ctx context.Context, entityType, entityID string, primaryFetch func(ctx context.Context) ([]interfaces.EntitySpecification, error)) ([]interfaces.EntitySpecification, *errors.FallbackResult) {
	// Try primary fetch first
	data, err := primaryFetch(ctx)
	if err == nil {
		return data, nil
	}

	fm.logError("Primary fetch failed for entity specifications", err)

	// Try cache fallback
	if cachedData, found := fm.cache.GetEntitySpecifications(ctx, entityType, entityID); found {
		fm.metrics.CacheFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using cached entity specifications as fallback")
		return cachedData, errors.NewFallbackResult(errors.FallbackStrategyCache, true, cachedData, "Used cached data", err)
	}

	// Try default data fallback
	if fm.config.EnableDefaultData {
		defaultData := fm.defaultData.GetDefaultEntitySpecifications(entityType, entityID)
		fm.metrics.DefaultFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using default entity specifications as fallback")
		return defaultData, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, defaultData, "Used default data", err)
	}

	// Return empty result as last resort
	fm.metrics.EmptyFallbacks++
	fm.metrics.TotalFallbacks++
	fm.logError("All fallback strategies failed for entity specifications", err)
	return []interfaces.EntitySpecification{}, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "No fallback data available", err)
}

// GetServiceDefinitionsWithFallback attempts to get service definitions with fallback strategies
func (fm *FallbackManager) GetServiceDefinitionsWithFallback(ctx context.Context, serviceTypes, parcelCategories []string, primaryFetch func(ctx context.Context) ([]interfaces.ServiceDefinition, error)) ([]interfaces.ServiceDefinition, *errors.FallbackResult) {
	// Try primary fetch first
	data, err := primaryFetch(ctx)
	if err == nil {
		return data, nil
	}

	fm.logError("Primary fetch failed for service definitions", err)

	// Try cache fallback
	if cachedData, found := fm.cache.GetServiceDefinitions(ctx, serviceTypes, parcelCategories); found {
		fm.metrics.CacheFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using cached service definitions as fallback")
		return cachedData, errors.NewFallbackResult(errors.FallbackStrategyCache, true, cachedData, "Used cached data", err)
	}

	// Try default data fallback
	if fm.config.EnableDefaultData {
		defaultData := fm.defaultData.GetDefaultServiceDefinitions(serviceTypes, parcelCategories)
		fm.metrics.DefaultFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using default service definitions as fallback")
		return defaultData, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, defaultData, "Used default data", err)
	}

	// Return empty result as last resort
	fm.metrics.EmptyFallbacks++
	fm.metrics.TotalFallbacks++
	fm.logError("All fallback strategies failed for service definitions", err)
	return []interfaces.ServiceDefinition{}, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "No fallback data available", err)
}

// GetParcelCategoriesWithFallback attempts to get parcel categories with fallback strategies
func (fm *FallbackManager) GetParcelCategoriesWithFallback(ctx context.Context, primaryFetch func(ctx context.Context) ([]string, error)) ([]string, *errors.FallbackResult) {
	// Try primary fetch first
	data, err := primaryFetch(ctx)
	if err == nil {
		return data, nil
	}

	fm.logError("Primary fetch failed for parcel categories", err)

	// Try default data fallback
	if fm.config.EnableDefaultData {
		defaultData := []string{"ecom", "courier", "cargo", "documents", "packages"}
		fm.metrics.DefaultFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using default parcel categories as fallback")
		return defaultData, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, defaultData, "Used default data", err)
	}

	// Return empty result as last resort
	fm.metrics.EmptyFallbacks++
	fm.metrics.TotalFallbacks++
	fm.logError("All fallback strategies failed for parcel categories", err)
	return []string{}, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "No fallback data available", err)
}

// GetServiceTypesWithFallback attempts to get service types with fallback strategies
func (fm *FallbackManager) GetServiceTypesWithFallback(ctx context.Context, parcelCategory string, primaryFetch func(ctx context.Context) ([]string, error)) ([]string, *errors.FallbackResult) {
	// Try primary fetch first
	data, err := primaryFetch(ctx)
	if err == nil {
		return data, nil
	}

	fm.logError("Primary fetch failed for service types", err)

	// Try default data fallback
	if fm.config.EnableDefaultData {
		defaultData := []string{"SDD", "NDD", "Standard", "Express"}
		fm.metrics.DefaultFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using default service types as fallback")
		return defaultData, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, defaultData, "Used default data", err)
	}

	// Return empty result as last resort
	fm.metrics.EmptyFallbacks++
	fm.metrics.TotalFallbacks++
	fm.logError("All fallback strategies failed for service types", err)
	return []string{}, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "No fallback data available", err)
}

// GetOperationTypesWithFallback attempts to get operation types with fallback strategies
func (fm *FallbackManager) GetOperationTypesWithFallback(ctx context.Context, primaryFetch func(ctx context.Context) ([]string, error)) ([]string, *errors.FallbackResult) {
	// Try primary fetch first
	data, err := primaryFetch(ctx)
	if err == nil {
		return data, nil
	}

	fm.logError("Primary fetch failed for operation types", err)

	// Try default data fallback
	if fm.config.EnableDefaultData {
		defaultData := []string{"pickup", "delivery", "tracking"}
		fm.metrics.DefaultFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using default operation types as fallback")
		return defaultData, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, defaultData, "Used default data", err)
	}

	// Return empty result as last resort
	fm.metrics.EmptyFallbacks++
	fm.metrics.TotalFallbacks++
	fm.logError("All fallback strategies failed for operation types", err)
	return []string{}, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "No fallback data available", err)
}

// GetPaymentModesWithFallback attempts to get payment modes with fallback strategies
func (fm *FallbackManager) GetPaymentModesWithFallback(ctx context.Context, primaryFetch func(ctx context.Context) ([]string, error)) ([]string, *errors.FallbackResult) {
	// Try primary fetch first
	data, err := primaryFetch(ctx)
	if err == nil {
		return data, nil
	}

	fm.logError("Primary fetch failed for payment modes", err)

	// Try default data fallback
	if fm.config.EnableDefaultData {
		defaultData := []string{"COD", "ONLINE", "PREPAID"}
		fm.metrics.DefaultFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using default payment modes as fallback")
		return defaultData, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, defaultData, "Used default data", err)
	}

	// Return empty result as last resort
	fm.metrics.EmptyFallbacks++
	fm.metrics.TotalFallbacks++
	fm.logError("All fallback strategies failed for payment modes", err)
	return []string{}, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "No fallback data available", err)
}

// GetDeliveryModesWithFallback attempts to get delivery modes with fallback strategies
func (fm *FallbackManager) GetDeliveryModesWithFallback(ctx context.Context, primaryFetch func(ctx context.Context) ([]string, error)) ([]string, *errors.FallbackResult) {
	// Try primary fetch first
	data, err := primaryFetch(ctx)
	if err == nil {
		return data, nil
	}

	fm.logError("Primary fetch failed for delivery modes", err)

	// Try default data fallback
	if fm.config.EnableDefaultData {
		defaultData := []string{"AIR", "SURFACE", "RAIL"}
		fm.metrics.DefaultFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using default delivery modes as fallback")
		return defaultData, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, defaultData, "Used default data", err)
	}

	// Return empty result as last resort
	fm.metrics.EmptyFallbacks++
	fm.metrics.TotalFallbacks++
	fm.logError("All fallback strategies failed for delivery modes", err)
	return []string{}, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "No fallback data available", err)
}

// GetServiceDefinitionWithFallback attempts to get a single service definition with fallback strategies
func (fm *FallbackManager) GetServiceDefinitionWithFallback(ctx context.Context, serviceType string, primaryFetch func(ctx context.Context) (*interfaces.ServiceDefinition, error)) (*interfaces.ServiceDefinition, *errors.FallbackResult) {
	// Try primary fetch first
	data, err := primaryFetch(ctx)
	if err == nil {
		return data, nil
	}

	fm.logError("Primary fetch failed for service definition", err)

	// Try default data fallback
	if fm.config.EnableDefaultData {
		defaultData := &interfaces.ServiceDefinition{
			ServiceType:       serviceType,
			ParcelCategory:    "ecom",
			DefaultOperations: []string{"pickup", "delivery"},
			DefaultPayments:   []string{"COD", "ONLINE"},
			DefaultDelivery:   []string{"SURFACE"},
			Description:       "Default service definition for " + serviceType,
		}
		fm.metrics.DefaultFallbacks++
		fm.metrics.TotalFallbacks++
		fm.logInfo("Using default service definition as fallback")
		return defaultData, errors.NewFallbackResult(errors.FallbackStrategyDefault, true, defaultData, "Used default data", err)
	}

	// Return empty result as last resort
	fm.metrics.EmptyFallbacks++
	fm.metrics.TotalFallbacks++
	fm.logError("All fallback strategies failed for service definition", err)
	return nil, errors.NewFallbackResult(errors.FallbackStrategyEmpty, false, nil, "No fallback data available", err)
}

// RetryWithBackoff retries an operation with exponential backoff
func (fm *FallbackManager) RetryWithBackoff(ctx context.Context, operation func(ctx context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt < fm.config.MaxRetryAttempts; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := time.Duration(attempt) * fm.config.RetryDelay
			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := operation(ctx)
		if err == nil {
			if attempt > 0 {
				fm.metrics.SuccessfulRetries++
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if sdErr, ok := errors.GetServiceDefinitionError(err); ok && !sdErr.IsRetryable() {
			fm.metrics.FailedRetries++
			return lastErr
		}
	}

	fm.metrics.FailedRetries++
	fm.metrics.RetryFallbacks++
	fm.metrics.TotalFallbacks++
	return lastErr
}

// GetMetrics returns fallback metrics
func (fm *FallbackManager) GetMetrics() *FallbackMetrics {
	return fm.metrics
}

// ResetMetrics resets fallback metrics
func (fm *FallbackManager) ResetMetrics() {
	fm.metrics = &FallbackMetrics{}
}

// Helper methods for logging
func (fm *FallbackManager) logInfo(message string) {
	if fm.logger != nil {
		fm.logger.Printf("[FALLBACK] INFO: %s", message)
	}
}

func (fm *FallbackManager) logError(message string, err error) {
	if fm.logger != nil {
		if err != nil {
			fm.logger.Printf("[FALLBACK] ERROR: %s - %v", message, err)
		} else {
			fm.logger.Printf("[FALLBACK] ERROR: %s", message)
		}
	}
}
