package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// ServiceDefinitionCache provides type-safe caching for service definitions
type ServiceDefinitionCache struct {
	cache  Cache
	config ServiceDefinitionCacheConfig
}

// ServiceDefinitionCacheConfig holds configuration for service definition cache
type ServiceDefinitionCacheConfig struct {
	SpecDefinitionsTTL time.Duration `yaml:"spec_definitions_ttl" json:"spec_definitions_ttl"`
	CatalogsTTL        time.Duration `yaml:"catalogs_ttl" json:"catalogs_ttl"`
	EntitySpecsTTL     time.Duration `yaml:"entity_specs_ttl" json:"entity_specs_ttl"`
	ServiceDefsTTL     time.Duration `yaml:"service_defs_ttl" json:"service_defs_ttl"`
	KeyPrefix          string        `yaml:"key_prefix" json:"key_prefix"`
}

// NewServiceDefinitionCache creates a new service definition cache
func NewServiceDefinitionCache(cache Cache, config ServiceDefinitionCacheConfig) *ServiceDefinitionCache {
	// Set defaults if not provided
	if config.SpecDefinitionsTTL <= 0 {
		config.SpecDefinitionsTTL = 2 * time.Hour
	}
	if config.CatalogsTTL <= 0 {
		config.CatalogsTTL = 1 * time.Hour
	}
	if config.EntitySpecsTTL <= 0 {
		config.EntitySpecsTTL = 30 * time.Minute
	}
	if config.ServiceDefsTTL <= 0 {
		config.ServiceDefsTTL = 1 * time.Hour
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "servicedef"
	}

	return &ServiceDefinitionCache{
		cache:  cache,
		config: config,
	}
}

// GetSpecDefinitions retrieves cached specification definitions
func (sdc *ServiceDefinitionCache) GetSpecDefinitions(ctx context.Context) ([]interfaces.SpecDefinition, bool) {
	key := sdc.buildKey("spec_definitions", "all")

	value, found := sdc.cache.Get(ctx, key)
	if !found {
		return nil, false
	}

	specDefs, ok := value.([]interfaces.SpecDefinition)
	if !ok {
		// Invalid type, remove from cache
		sdc.cache.Delete(ctx, key)
		return nil, false
	}

	return specDefs, true
}

// SetSpecDefinitions caches specification definitions
func (sdc *ServiceDefinitionCache) SetSpecDefinitions(ctx context.Context, specDefs []interfaces.SpecDefinition) error {
	key := sdc.buildKey("spec_definitions", "all")
	return sdc.cache.Set(ctx, key, specDefs, sdc.config.SpecDefinitionsTTL)
}

// GetCatalogs retrieves cached catalogs
func (sdc *ServiceDefinitionCache) GetCatalogs(ctx context.Context) ([]interfaces.Catalog, bool) {
	key := sdc.buildKey("catalogs", "all")

	value, found := sdc.cache.Get(ctx, key)
	if !found {
		return nil, false
	}

	catalogs, ok := value.([]interfaces.Catalog)
	if !ok {
		// Invalid type, remove from cache
		sdc.cache.Delete(ctx, key)
		return nil, false
	}

	return catalogs, true
}

// SetCatalogs caches catalogs
func (sdc *ServiceDefinitionCache) SetCatalogs(ctx context.Context, catalogs []interfaces.Catalog) error {
	key := sdc.buildKey("catalogs", "all")
	return sdc.cache.Set(ctx, key, catalogs, sdc.config.CatalogsTTL)
}

// GetEntitySpecifications retrieves cached entity specifications
func (sdc *ServiceDefinitionCache) GetEntitySpecifications(ctx context.Context, entityType, entityID string) ([]interfaces.EntitySpecification, bool) {
	key := sdc.buildKey("entity_specs", fmt.Sprintf("%s:%s", entityType, entityID))

	value, found := sdc.cache.Get(ctx, key)
	if !found {
		return nil, false
	}

	entitySpecs, ok := value.([]interfaces.EntitySpecification)
	if !ok {
		// Invalid type, remove from cache
		sdc.cache.Delete(ctx, key)
		return nil, false
	}

	return entitySpecs, true
}

// SetEntitySpecifications caches entity specifications
func (sdc *ServiceDefinitionCache) SetEntitySpecifications(ctx context.Context, entityType, entityID string, entitySpecs []interfaces.EntitySpecification) error {
	key := sdc.buildKey("entity_specs", fmt.Sprintf("%s:%s", entityType, entityID))
	return sdc.cache.Set(ctx, key, entitySpecs, sdc.config.EntitySpecsTTL)
}

// GetServiceDefinitions retrieves cached service definitions
func (sdc *ServiceDefinitionCache) GetServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string) ([]interfaces.ServiceDefinition, bool) {
	key := sdc.buildServiceDefinitionsKey(serviceTypes, parcelCategories)

	value, found := sdc.cache.Get(ctx, key)
	if !found {
		return nil, false
	}

	serviceDefs, ok := value.([]interfaces.ServiceDefinition)
	if !ok {
		// Invalid type, remove from cache
		sdc.cache.Delete(ctx, key)
		return nil, false
	}

	return serviceDefs, true
}

// SetServiceDefinitions caches service definitions
func (sdc *ServiceDefinitionCache) SetServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string, serviceDefs []interfaces.ServiceDefinition) error {
	key := sdc.buildServiceDefinitionsKey(serviceTypes, parcelCategories)
	return sdc.cache.Set(ctx, key, serviceDefs, sdc.config.ServiceDefsTTL)
}

// InvalidateSpecDefinitions removes cached specification definitions
func (sdc *ServiceDefinitionCache) InvalidateSpecDefinitions(ctx context.Context) error {
	key := sdc.buildKey("spec_definitions", "all")
	return sdc.cache.Delete(ctx, key)
}

// InvalidateCatalogs removes cached catalogs
func (sdc *ServiceDefinitionCache) InvalidateCatalogs(ctx context.Context) error {
	key := sdc.buildKey("catalogs", "all")
	return sdc.cache.Delete(ctx, key)
}

// InvalidateEntitySpecifications removes cached entity specifications for a specific entity
func (sdc *ServiceDefinitionCache) InvalidateEntitySpecifications(ctx context.Context, entityType, entityID string) error {
	key := sdc.buildKey("entity_specs", fmt.Sprintf("%s:%s", entityType, entityID))
	return sdc.cache.Delete(ctx, key)
}

// InvalidateServiceDefinitions removes cached service definitions for specific criteria
func (sdc *ServiceDefinitionCache) InvalidateServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string) error {
	key := sdc.buildServiceDefinitionsKey(serviceTypes, parcelCategories)
	return sdc.cache.Delete(ctx, key)
}

// InvalidateAll removes all cached service definition data
func (sdc *ServiceDefinitionCache) InvalidateAll(ctx context.Context) error {
	// This is a simple implementation that clears the entire cache
	// In a more sophisticated implementation, we could track keys by prefix
	return sdc.cache.Clear(ctx)
}

// GetStats returns cache statistics
func (sdc *ServiceDefinitionCache) GetStats() CacheStats {
	return sdc.cache.GetStats()
}

// GetTTL returns the remaining TTL for a specific cache entry
func (sdc *ServiceDefinitionCache) GetTTL(ctx context.Context, cacheType, identifier string) (time.Duration, bool) {
	var key string
	switch cacheType {
	case "spec_definitions":
		key = sdc.buildKey("spec_definitions", "all")
	case "catalogs":
		key = sdc.buildKey("catalogs", "all")
	case "entity_specs":
		key = sdc.buildKey("entity_specs", identifier)
	case "service_definitions":
		key = sdc.buildKey("service_definitions", identifier)
	default:
		return 0, false
	}

	return sdc.cache.GetTTL(ctx, key)
}

// buildKey creates a cache key with the configured prefix
func (sdc *ServiceDefinitionCache) buildKey(category, identifier string) string {
	return fmt.Sprintf("%s:%s:%s", sdc.config.KeyPrefix, category, identifier)
}

// buildServiceDefinitionsKey creates a cache key for service definitions based on query parameters
func (sdc *ServiceDefinitionCache) buildServiceDefinitionsKey(serviceTypes, parcelCategories []string) string {
	// Create a deterministic key based on the query parameters
	queryData := map[string]interface{}{
		"service_types":     serviceTypes,
		"parcel_categories": parcelCategories,
	}

	// Convert to JSON for consistent ordering
	jsonData, err := json.Marshal(queryData)
	if err != nil {
		// Fallback to simple concatenation
		return sdc.buildKey("service_definitions", fmt.Sprintf("%s_%s",
			strings.Join(serviceTypes, ","),
			strings.Join(parcelCategories, ",")))
	}

	// Create hash of the JSON data for a shorter key
	hash := md5.Sum(jsonData)
	hashStr := fmt.Sprintf("%x", hash)

	return sdc.buildKey("service_definitions", hashStr)
}

// CacheMetrics provides detailed metrics about the service definition cache
type CacheMetrics struct {
	Stats                 CacheStats `json:"stats"`
	SpecDefinitionsCached bool       `json:"spec_definitions_cached"`
	CatalogsCached        bool       `json:"catalogs_cached"`
	EntitySpecsCacheCount int        `json:"entity_specs_cache_count"`
	ServiceDefsCacheCount int        `json:"service_defs_cache_count"`
}

// GetDetailedMetrics returns detailed metrics about cached data
func (sdc *ServiceDefinitionCache) GetDetailedMetrics(ctx context.Context) CacheMetrics {
	stats := sdc.cache.GetStats()

	// Check if main data types are cached
	specDefsCached := sdc.cache.Exists(ctx, sdc.buildKey("spec_definitions", "all"))
	catalogsCached := sdc.cache.Exists(ctx, sdc.buildKey("catalogs", "all"))

	return CacheMetrics{
		Stats:                 stats,
		SpecDefinitionsCached: specDefsCached,
		CatalogsCached:        catalogsCached,
		// Note: EntitySpecsCacheCount and ServiceDefsCacheCount would require
		// additional tracking or cache key enumeration to implement accurately
	}
}
