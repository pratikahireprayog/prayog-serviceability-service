package utils

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"prayog-serviceability-service/internal/infrastructure/cache"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// PartnerCapabilityCacheManager provides type-safe caching for partner capability data
type PartnerCapabilityCacheManager interface {
	// Partner Capability methods
	GetPartnerCapability(ctx context.Context, partnerID uint) (*interfaces.PartnerCapability, bool)
	SetPartnerCapability(ctx context.Context, partnerID uint, capability *interfaces.PartnerCapability) error
	InvalidatePartnerCapability(ctx context.Context, partnerID uint) error

	// Partner Service Capabilities methods
	GetPartnerServiceCapabilities(ctx context.Context, partnerID uint, filters *interfaces.ServiceFilters) (*interfaces.PartnerServiceCapabilities, bool)
	SetPartnerServiceCapabilities(ctx context.Context, partnerID uint, filters *interfaces.ServiceFilters, capabilities *interfaces.PartnerServiceCapabilities) error
	InvalidatePartnerServiceCapabilities(ctx context.Context, partnerID uint) error

	// Partners by Location methods
	GetPartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy) ([]interfaces.PartnerCapability, bool)
	SetPartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy, partners []interfaces.PartnerCapability) error
	InvalidatePartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy) error

	// Service Definitions cache
	GetServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string) ([]interfaces.ServiceDefinition, bool)
	SetServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string, definitions []interfaces.ServiceDefinition) error
	InvalidateServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string) error

	// Batch operations
	GetMultiplePartnerCapabilities(ctx context.Context, partnerIDs []uint) (map[uint]*interfaces.PartnerCapability, []uint)
	SetMultiplePartnerCapabilities(ctx context.Context, capabilities map[uint]*interfaces.PartnerCapability) error
	InvalidateMultiplePartners(ctx context.Context, partnerIDs []uint) error

	// Cache management
	GetStats() cache.CacheStats
	GetDetailedMetrics(ctx context.Context) CacheMetrics
	InvalidateAll(ctx context.Context) error
	InvalidateExpired(ctx context.Context) error
	GetTTL(ctx context.Context, cacheType, identifier string) (time.Duration, bool)

	// Health and monitoring
	IsHealthy() bool
	GetHealth() HealthStatus
}

// partnerCapabilityCacheManager implements the PartnerCapabilityCacheManager interface
type partnerCapabilityCacheManager struct {
	cache        cache.Cache
	config       PartnerCacheConfig
	keyGenerator *CacheKeyGenerator
	metrics      *CacheMetrics
	mu           sync.RWMutex
}

// PartnerCacheConfig holds configuration for partner capability cache
type PartnerCacheConfig struct {
	PartnerCapabilityTTL       time.Duration `yaml:"partner_capability_ttl" json:"partner_capability_ttl"`
	ServiceCapabilitiesTTL     time.Duration `yaml:"service_capabilities_ttl" json:"service_capabilities_ttl"`
	LocationPartnersTTL        time.Duration `yaml:"location_partners_ttl" json:"location_partners_ttl"`
	ServiceDefinitionsTTL      time.Duration `yaml:"service_definitions_ttl" json:"service_definitions_ttl"`
	KeyPrefix                  string        `yaml:"key_prefix" json:"key_prefix"`
	EnableCompression          bool          `yaml:"enable_compression" json:"enable_compression"`
	MaxValueSize               int           `yaml:"max_value_size" json:"max_value_size"`
	BatchSize                  int           `yaml:"batch_size" json:"batch_size"`
	EnableMetrics              bool          `yaml:"enable_metrics" json:"enable_metrics"`
	StaleDataTolerance         time.Duration `yaml:"stale_data_tolerance" json:"stale_data_tolerance"`
	PreloadPopularData         bool          `yaml:"preload_popular_data" json:"preload_popular_data"`
	PopularPartnersRefreshRate time.Duration `yaml:"popular_partners_refresh_rate" json:"popular_partners_refresh_rate"`
	EnableInvalidationTracking bool          `yaml:"enable_invalidation_tracking" json:"enable_invalidation_tracking"`
}

// CacheKeyGenerator generates cache keys with consistent patterns
type CacheKeyGenerator struct {
	prefix string
}

// CacheMetrics provides detailed metrics about partner capability cache
type CacheMetrics struct {
	BaseStats                   cache.CacheStats `json:"base_stats"`
	PartnerCapabilitiesCached   int              `json:"partner_capabilities_cached"`
	ServiceCapabilitiesCached   int              `json:"service_capabilities_cached"`
	LocationPartnersCached      int              `json:"location_partners_cached"`
	ServiceDefinitionsCached    int              `json:"service_definitions_cached"`
	BatchOperationsCount        int64            `json:"batch_operations_count"`
	InvalidationCount           int64            `json:"invalidation_count"`
	CompressionSavings          int64            `json:"compression_savings_bytes"`
	AverageValueSize            float64          `json:"average_value_size_bytes"`
	PopularDataHitRatio         float64          `json:"popular_data_hit_ratio"`
	StaleDataServedCount        int64            `json:"stale_data_served_count"`
	LastCleanupTime             time.Time        `json:"last_cleanup_time"`
	NextScheduledCleanup        time.Time        `json:"next_scheduled_cleanup"`
	MemoryUsageEstimate         int64            `json:"memory_usage_estimate_bytes"`
	InvalidationTrackingEnabled bool             `json:"invalidation_tracking_enabled"`
}

// HealthStatus represents cache health information
type HealthStatus struct {
	IsHealthy               bool      `json:"is_healthy"`
	LastSuccessfulOperation time.Time `json:"last_successful_operation"`
	ConsecutiveErrors       int       `json:"consecutive_errors"`
	MemoryPressure          float64   `json:"memory_pressure"`
	ResponseTimeP95         float64   `json:"response_time_p95_ms"`
	ErrorRate               float64   `json:"error_rate"`
	UpstreamHealthy         bool      `json:"upstream_healthy"`
	CacheEfficiency         float64   `json:"cache_efficiency"`
	RecommendedActions      []string  `json:"recommended_actions"`
}

// NewPartnerCapabilityCacheManager creates a new partner capability cache manager
func NewPartnerCapabilityCacheManager(cache cache.Cache, config PartnerCacheConfig) PartnerCapabilityCacheManager {
	// Set defaults if not provided
	if config.PartnerCapabilityTTL <= 0 {
		config.PartnerCapabilityTTL = 30 * time.Minute
	}
	if config.ServiceCapabilitiesTTL <= 0 {
		config.ServiceCapabilitiesTTL = 15 * time.Minute
	}
	if config.LocationPartnersTTL <= 0 {
		config.LocationPartnersTTL = 10 * time.Minute
	}
	if config.ServiceDefinitionsTTL <= 0 {
		config.ServiceDefinitionsTTL = 2 * time.Hour
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "partner_capability"
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 50
	}
	if config.MaxValueSize <= 0 {
		config.MaxValueSize = 1024 * 1024 // 1MB
	}
	if config.StaleDataTolerance <= 0 {
		config.StaleDataTolerance = 5 * time.Minute
	}
	if config.PopularPartnersRefreshRate <= 0 {
		config.PopularPartnersRefreshRate = 1 * time.Hour
	}

	return &partnerCapabilityCacheManager{
		cache:        cache,
		config:       config,
		keyGenerator: &CacheKeyGenerator{prefix: config.KeyPrefix},
		metrics: &CacheMetrics{
			BaseStats:                   cache.GetStats(),
			InvalidationTrackingEnabled: config.EnableInvalidationTracking,
		},
	}
}

// GetPartnerCapability retrieves cached partner capability
func (pcm *partnerCapabilityCacheManager) GetPartnerCapability(ctx context.Context, partnerID uint) (*interfaces.PartnerCapability, bool) {
	key := pcm.keyGenerator.PartnerCapabilityKey(partnerID)

	value, found := pcm.cache.Get(ctx, key)
	if !found {
		pcm.updateMetrics(func(m *CacheMetrics) {
			m.BaseStats.Misses++
		})
		return nil, false
	}

	capability, ok := value.(*interfaces.PartnerCapability)
	if !ok {
		// Invalid type, remove from cache
		pcm.cache.Delete(ctx, key)
		pcm.updateMetrics(func(m *CacheMetrics) {
			m.BaseStats.Misses++
			m.InvalidationCount++
		})
		return nil, false
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BaseStats.Hits++
		m.BaseStats.HitRatio = float64(m.BaseStats.Hits) / float64(m.BaseStats.Hits+m.BaseStats.Misses)
	})

	return capability, true
}

// SetPartnerCapability caches partner capability
func (pcm *partnerCapabilityCacheManager) SetPartnerCapability(ctx context.Context, partnerID uint, capability *interfaces.PartnerCapability) error {
	if capability == nil {
		return fmt.Errorf("cannot cache nil partner capability")
	}

	key := pcm.keyGenerator.PartnerCapabilityKey(partnerID)

	// Check value size if compression is disabled
	if !pcm.config.EnableCompression {
		if err := pcm.validateValueSize(capability); err != nil {
			return fmt.Errorf("partner capability too large: %w", err)
		}
	}

	err := pcm.cache.Set(ctx, key, capability, pcm.config.PartnerCapabilityTTL)
	if err != nil {
		return fmt.Errorf("failed to cache partner capability: %w", err)
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BaseStats.Sets++
		m.PartnerCapabilitiesCached++
	})

	return nil
}

// InvalidatePartnerCapability removes cached partner capability
func (pcm *partnerCapabilityCacheManager) InvalidatePartnerCapability(ctx context.Context, partnerID uint) error {
	key := pcm.keyGenerator.PartnerCapabilityKey(partnerID)

	err := pcm.cache.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to invalidate partner capability: %w", err)
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BaseStats.Deletes++
		m.InvalidationCount++
	})

	return nil
}

// GetPartnerServiceCapabilities retrieves cached partner service capabilities
func (pcm *partnerCapabilityCacheManager) GetPartnerServiceCapabilities(ctx context.Context, partnerID uint, filters *interfaces.ServiceFilters) (*interfaces.PartnerServiceCapabilities, bool) {
	key := pcm.keyGenerator.ServiceCapabilitiesKey(partnerID, filters)

	value, found := pcm.cache.Get(ctx, key)
	if !found {
		pcm.updateMetrics(func(m *CacheMetrics) {
			m.BaseStats.Misses++
		})
		return nil, false
	}

	capabilities, ok := value.(*interfaces.PartnerServiceCapabilities)
	if !ok {
		// Invalid type, remove from cache
		pcm.cache.Delete(ctx, key)
		pcm.updateMetrics(func(m *CacheMetrics) {
			m.BaseStats.Misses++
			m.InvalidationCount++
		})
		return nil, false
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BaseStats.Hits++
		m.BaseStats.HitRatio = float64(m.BaseStats.Hits) / float64(m.BaseStats.Hits+m.BaseStats.Misses)
	})

	return capabilities, true
}

// SetPartnerServiceCapabilities caches partner service capabilities
func (pcm *partnerCapabilityCacheManager) SetPartnerServiceCapabilities(ctx context.Context, partnerID uint, filters *interfaces.ServiceFilters, capabilities *interfaces.PartnerServiceCapabilities) error {
	if capabilities == nil {
		return fmt.Errorf("cannot cache nil partner service capabilities")
	}

	key := pcm.keyGenerator.ServiceCapabilitiesKey(partnerID, filters)

	if !pcm.config.EnableCompression {
		if err := pcm.validateValueSize(capabilities); err != nil {
			return fmt.Errorf("service capabilities too large: %w", err)
		}
	}

	err := pcm.cache.Set(ctx, key, capabilities, pcm.config.ServiceCapabilitiesTTL)
	if err != nil {
		return fmt.Errorf("failed to cache service capabilities: %w", err)
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BaseStats.Sets++
		m.ServiceCapabilitiesCached++
	})

	return nil
}

// InvalidatePartnerServiceCapabilities removes cached service capabilities for a partner
func (pcm *partnerCapabilityCacheManager) InvalidatePartnerServiceCapabilities(ctx context.Context, partnerID uint) error {
	// In a more sophisticated implementation, we would track all keys for a partner
	// For now, we'll use a pattern-based invalidation
	keyPattern := pcm.keyGenerator.ServiceCapabilitiesPrefix(partnerID)

	// Note: This would require a cache implementation that supports pattern-based deletion
	// For now, we'll implement a simplified version
	err := pcm.invalidateByPattern(ctx, keyPattern)
	if err != nil {
		return fmt.Errorf("failed to invalidate service capabilities: %w", err)
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.InvalidationCount++
	})

	return nil
}

// GetPartnersByLocation retrieves cached partners by location
func (pcm *partnerCapabilityCacheManager) GetPartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy) ([]interfaces.PartnerCapability, bool) {
	key := pcm.keyGenerator.LocationPartnersKey(locationHierarchy)

	value, found := pcm.cache.Get(ctx, key)
	if !found {
		pcm.updateMetrics(func(m *CacheMetrics) {
			m.BaseStats.Misses++
		})
		return nil, false
	}

	partners, ok := value.([]interfaces.PartnerCapability)
	if !ok {
		// Invalid type, remove from cache
		pcm.cache.Delete(ctx, key)
		pcm.updateMetrics(func(m *CacheMetrics) {
			m.BaseStats.Misses++
			m.InvalidationCount++
		})
		return nil, false
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BaseStats.Hits++
		m.BaseStats.HitRatio = float64(m.BaseStats.Hits) / float64(m.BaseStats.Hits+m.BaseStats.Misses)
	})

	return partners, true
}

// SetPartnersByLocation caches partners by location
func (pcm *partnerCapabilityCacheManager) SetPartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy, partners []interfaces.PartnerCapability) error {
	if partners == nil {
		return fmt.Errorf("cannot cache nil partners list")
	}

	key := pcm.keyGenerator.LocationPartnersKey(locationHierarchy)

	if !pcm.config.EnableCompression {
		if err := pcm.validateValueSize(partners); err != nil {
			return fmt.Errorf("partners list too large: %w", err)
		}
	}

	err := pcm.cache.Set(ctx, key, partners, pcm.config.LocationPartnersTTL)
	if err != nil {
		return fmt.Errorf("failed to cache location partners: %w", err)
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BaseStats.Sets++
		m.LocationPartnersCached++
	})

	return nil
}

// InvalidatePartnersByLocation removes cached partners for a location
func (pcm *partnerCapabilityCacheManager) InvalidatePartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy) error {
	key := pcm.keyGenerator.LocationPartnersKey(locationHierarchy)

	err := pcm.cache.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to invalidate location partners: %w", err)
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BaseStats.Deletes++
		m.InvalidationCount++
	})

	return nil
}

// GetServiceDefinitions retrieves cached service definitions
func (pcm *partnerCapabilityCacheManager) GetServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string) ([]interfaces.ServiceDefinition, bool) {
	key := pcm.keyGenerator.ServiceDefinitionsKey(serviceTypes, parcelCategories)

	value, found := pcm.cache.Get(ctx, key)
	if !found {
		pcm.updateMetrics(func(m *CacheMetrics) {
			m.BaseStats.Misses++
		})
		return nil, false
	}

	definitions, ok := value.([]interfaces.ServiceDefinition)
	if !ok {
		// Invalid type, remove from cache
		pcm.cache.Delete(ctx, key)
		pcm.updateMetrics(func(m *CacheMetrics) {
			m.BaseStats.Misses++
			m.InvalidationCount++
		})
		return nil, false
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BaseStats.Hits++
		m.BaseStats.HitRatio = float64(m.BaseStats.Hits) / float64(m.BaseStats.Hits+m.BaseStats.Misses)
	})

	return definitions, true
}

// SetServiceDefinitions caches service definitions
func (pcm *partnerCapabilityCacheManager) SetServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string, definitions []interfaces.ServiceDefinition) error {
	if definitions == nil {
		return fmt.Errorf("cannot cache nil service definitions")
	}

	key := pcm.keyGenerator.ServiceDefinitionsKey(serviceTypes, parcelCategories)

	if !pcm.config.EnableCompression {
		if err := pcm.validateValueSize(definitions); err != nil {
			return fmt.Errorf("service definitions too large: %w", err)
		}
	}

	err := pcm.cache.Set(ctx, key, definitions, pcm.config.ServiceDefinitionsTTL)
	if err != nil {
		return fmt.Errorf("failed to cache service definitions: %w", err)
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BaseStats.Sets++
		m.ServiceDefinitionsCached++
	})

	return nil
}

// InvalidateServiceDefinitions removes cached service definitions
func (pcm *partnerCapabilityCacheManager) InvalidateServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string) error {
	key := pcm.keyGenerator.ServiceDefinitionsKey(serviceTypes, parcelCategories)

	err := pcm.cache.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to invalidate service definitions: %w", err)
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BaseStats.Deletes++
		m.InvalidationCount++
	})

	return nil
}

// GetMultiplePartnerCapabilities retrieves multiple partner capabilities in batch
func (pcm *partnerCapabilityCacheManager) GetMultiplePartnerCapabilities(ctx context.Context, partnerIDs []uint) (map[uint]*interfaces.PartnerCapability, []uint) {
	found := make(map[uint]*interfaces.PartnerCapability)
	missing := make([]uint, 0)

	for _, partnerID := range partnerIDs {
		capability, exists := pcm.GetPartnerCapability(ctx, partnerID)
		if exists {
			found[partnerID] = capability
		} else {
			missing = append(missing, partnerID)
		}
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BatchOperationsCount++
	})

	return found, missing
}

// SetMultiplePartnerCapabilities caches multiple partner capabilities in batch
func (pcm *partnerCapabilityCacheManager) SetMultiplePartnerCapabilities(ctx context.Context, capabilities map[uint]*interfaces.PartnerCapability) error {
	var errors []string

	for partnerID, capability := range capabilities {
		if err := pcm.SetPartnerCapability(ctx, partnerID, capability); err != nil {
			errors = append(errors, fmt.Sprintf("partner %d: %v", partnerID, err))
		}
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BatchOperationsCount++
	})

	if len(errors) > 0 {
		return fmt.Errorf("batch set errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// InvalidateMultiplePartners removes multiple partner capabilities
func (pcm *partnerCapabilityCacheManager) InvalidateMultiplePartners(ctx context.Context, partnerIDs []uint) error {
	var errors []string

	for _, partnerID := range partnerIDs {
		if err := pcm.InvalidatePartnerCapability(ctx, partnerID); err != nil {
			errors = append(errors, fmt.Sprintf("partner %d: %v", partnerID, err))
		}
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.BatchOperationsCount++
	})

	if len(errors) > 0 {
		return fmt.Errorf("batch invalidation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetStats returns cache statistics
func (pcm *partnerCapabilityCacheManager) GetStats() cache.CacheStats {
	return pcm.cache.GetStats()
}

// GetDetailedMetrics returns detailed cache metrics
func (pcm *partnerCapabilityCacheManager) GetDetailedMetrics(ctx context.Context) CacheMetrics {
	pcm.mu.RLock()
	defer pcm.mu.RUnlock()

	metrics := *pcm.metrics
	metrics.BaseStats = pcm.cache.GetStats()
	metrics.LastCleanupTime = time.Now()
	metrics.NextScheduledCleanup = time.Now().Add(1 * time.Hour)

	// Calculate additional metrics
	if metrics.BaseStats.Size > 0 {
		metrics.MemoryUsageEstimate = int64(metrics.BaseStats.Size * 1024) // Rough estimate
	}

	return metrics
}

// InvalidateAll removes all cached data
func (pcm *partnerCapabilityCacheManager) InvalidateAll(ctx context.Context) error {
	err := pcm.cache.Clear(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	pcm.updateMetrics(func(m *CacheMetrics) {
		m.InvalidationCount++
	})

	return nil
}

// InvalidateExpired removes expired entries (if supported by cache implementation)
func (pcm *partnerCapabilityCacheManager) InvalidateExpired(ctx context.Context) error {
	// This would depend on the underlying cache implementation
	// For now, return nil as most cache implementations handle this automatically
	return nil
}

// GetTTL returns the remaining TTL for a specific cache entry
func (pcm *partnerCapabilityCacheManager) GetTTL(ctx context.Context, cacheType, identifier string) (time.Duration, bool) {
	var key string

	switch cacheType {
	case "partner_capability":
		if partnerID, err := parsePartnerID(identifier); err == nil {
			key = pcm.keyGenerator.PartnerCapabilityKey(partnerID)
		}
	case "service_capabilities":
		if partnerID, err := parsePartnerID(identifier); err == nil {
			key = pcm.keyGenerator.ServiceCapabilitiesPrefix(partnerID) + ":" + identifier
		}
	case "location_partners":
		key = pcm.keyGenerator.buildKey("location_partners", identifier)
	case "service_definitions":
		key = pcm.keyGenerator.buildKey("service_definitions", identifier)
	default:
		return 0, false
	}

	return pcm.cache.GetTTL(ctx, key)
}

// IsHealthy returns whether the cache is healthy
func (pcm *partnerCapabilityCacheManager) IsHealthy() bool {
	stats := pcm.cache.GetStats()

	// Consider healthy if hit ratio is reasonable and no recent errors
	return stats.HitRatio > 0.1 && time.Since(stats.LastUpdated) < 5*time.Minute
}

// GetHealth returns detailed health status
func (pcm *partnerCapabilityCacheManager) GetHealth() HealthStatus {
	stats := pcm.cache.GetStats()

	health := HealthStatus{
		IsHealthy:               pcm.IsHealthy(),
		LastSuccessfulOperation: stats.LastUpdated,
		MemoryPressure:          float64(stats.Size) / float64(stats.MaxSize),
		CacheEfficiency:         stats.HitRatio,
		UpstreamHealthy:         true, // Assume healthy for cache
	}

	// Add recommendations based on metrics
	if health.MemoryPressure > 0.9 {
		health.RecommendedActions = append(health.RecommendedActions, "Consider increasing cache size or adjusting TTL values")
	}
	if stats.HitRatio < 0.5 {
		health.RecommendedActions = append(health.RecommendedActions, "Low hit ratio - review caching strategy")
	}

	return health
}

// Helper methods

// updateMetrics safely updates cache metrics
func (pcm *partnerCapabilityCacheManager) updateMetrics(update func(*CacheMetrics)) {
	if !pcm.config.EnableMetrics {
		return
	}

	pcm.mu.Lock()
	defer pcm.mu.Unlock()
	update(pcm.metrics)
}

// validateValueSize checks if value is within size limits
func (pcm *partnerCapabilityCacheManager) validateValueSize(value interface{}) error {
	// Rough size estimation - in production, you might want more accurate sizing
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value for size check: %w", err)
	}

	if len(data) > pcm.config.MaxValueSize {
		return fmt.Errorf("value size %d exceeds maximum %d", len(data), pcm.config.MaxValueSize)
	}

	return nil
}

// invalidateByPattern removes cache entries matching a pattern (simplified implementation)
func (pcm *partnerCapabilityCacheManager) invalidateByPattern(ctx context.Context, pattern string) error {
	// This is a simplified implementation
	// In a real scenario, you would need cache support for pattern-based operations
	// or maintain a registry of keys
	return nil
}

// CacheKeyGenerator methods

// PartnerCapabilityKey generates cache key for partner capability
func (kg *CacheKeyGenerator) PartnerCapabilityKey(partnerID uint) string {
	return kg.buildKey("partner_capability", fmt.Sprintf("%d", partnerID))
}

// ServiceCapabilitiesKey generates cache key for service capabilities
func (kg *CacheKeyGenerator) ServiceCapabilitiesKey(partnerID uint, filters *interfaces.ServiceFilters) string {
	filterHash := kg.hashServiceFilters(filters)
	return kg.buildKey("service_capabilities", fmt.Sprintf("%d:%s", partnerID, filterHash))
}

// ServiceCapabilitiesPrefix generates prefix for service capabilities keys
func (kg *CacheKeyGenerator) ServiceCapabilitiesPrefix(partnerID uint) string {
	return kg.buildKey("service_capabilities", fmt.Sprintf("%d", partnerID))
}

// LocationPartnersKey generates cache key for location partners
func (kg *CacheKeyGenerator) LocationPartnersKey(locationHierarchy *models.LocationHierarchy) string {
	locationHash := kg.hashLocationHierarchy(locationHierarchy)
	return kg.buildKey("location_partners", locationHash)
}

// ServiceDefinitionsKey generates cache key for service definitions
func (kg *CacheKeyGenerator) ServiceDefinitionsKey(serviceTypes, parcelCategories []string) string {
	queryHash := kg.hashServiceQuery(serviceTypes, parcelCategories)
	return kg.buildKey("service_definitions", queryHash)
}

// buildKey creates a cache key with the configured prefix
func (kg *CacheKeyGenerator) buildKey(category, identifier string) string {
	return fmt.Sprintf("%s:%s:%s", kg.prefix, category, identifier)
}

// hashServiceFilters creates a hash for service filters
func (kg *CacheKeyGenerator) hashServiceFilters(filters *interfaces.ServiceFilters) string {
	if filters == nil {
		return "no_filters"
	}

	data := map[string]interface{}{
		"service_types":     filters.ServiceTypes,
		"parcel_categories": filters.ParcelCategories,
		"operation_types":   filters.OperationTypes,
	}

	return kg.hashData(data)
}

// hashLocationHierarchy creates a hash for location hierarchy
func (kg *CacheKeyGenerator) hashLocationHierarchy(hierarchy *models.LocationHierarchy) string {
	if hierarchy == nil {
		return "no_location"
	}

	data := map[string]string{
		"postal_code":  hierarchy.PostalCode,
		"city_code":    hierarchy.CityCode,
		"region_code":  hierarchy.RegionCode,
		"country_code": hierarchy.CountryCode,
	}

	return kg.hashData(data)
}

// hashServiceQuery creates a hash for service query parameters
func (kg *CacheKeyGenerator) hashServiceQuery(serviceTypes, parcelCategories []string) string {
	data := map[string]interface{}{
		"service_types":     serviceTypes,
		"parcel_categories": parcelCategories,
	}

	return kg.hashData(data)
}

// hashData creates an MD5 hash of the given data
func (kg *CacheKeyGenerator) hashData(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Fallback to string representation
		return fmt.Sprintf("%v", data)
	}

	hash := md5.Sum(jsonData)
	return fmt.Sprintf("%x", hash)
}

// Helper functions

// parsePartnerID parses partner ID from string
func parsePartnerID(identifier string) (uint, error) {
	id := 0
	_, err := fmt.Sscanf(identifier, "%d", &id)
	return uint(id), err
}
