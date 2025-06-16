package utils

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"prayog-serviceability-service/internal/infrastructure/cache"
	infraErrors "prayog-serviceability-service/internal/infrastructure/errors"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// Mock implementations for testing
type mockCacheManager struct {
	partnerCapabilities  map[uint]*interfaces.PartnerCapability
	serviceCapabilities  map[uint]*interfaces.PartnerServiceCapabilities
	locationPartners     map[string][]interfaces.PartnerCapability
	serviceDefinitions   map[string][]interfaces.ServiceDefinition
	multipleCapabilities map[uint]*interfaces.PartnerCapability
	shouldFail           bool
	missingPartners      []uint
}

func (m *mockCacheManager) GetPartnerCapability(ctx context.Context, partnerID uint) (*interfaces.PartnerCapability, bool) {
	if m.shouldFail {
		return nil, false
	}
	capability, exists := m.partnerCapabilities[partnerID]
	return capability, exists
}

func (m *mockCacheManager) SetPartnerCapability(ctx context.Context, partnerID uint, capability *interfaces.PartnerCapability) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	if m.partnerCapabilities == nil {
		m.partnerCapabilities = make(map[uint]*interfaces.PartnerCapability)
	}
	m.partnerCapabilities[partnerID] = capability
	return nil
}

func (m *mockCacheManager) GetPartnerServiceCapabilities(ctx context.Context, partnerID uint, filters *interfaces.ServiceFilters) (*interfaces.PartnerServiceCapabilities, bool) {
	if m.shouldFail {
		return nil, false
	}
	capabilities, exists := m.serviceCapabilities[partnerID]
	return capabilities, exists
}

func (m *mockCacheManager) SetPartnerServiceCapabilities(ctx context.Context, partnerID uint, filters *interfaces.ServiceFilters, capabilities *interfaces.PartnerServiceCapabilities) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	if m.serviceCapabilities == nil {
		m.serviceCapabilities = make(map[uint]*interfaces.PartnerServiceCapabilities)
	}
	m.serviceCapabilities[partnerID] = capabilities
	return nil
}

func (m *mockCacheManager) GetPartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy) ([]interfaces.PartnerCapability, bool) {
	if m.shouldFail {
		return nil, false
	}
	key := locationHierarchy.CountryCode + "-" + locationHierarchy.RegionCode
	partners, exists := m.locationPartners[key]
	return partners, exists
}

func (m *mockCacheManager) SetPartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy, partners []interfaces.PartnerCapability) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	if m.locationPartners == nil {
		m.locationPartners = make(map[string][]interfaces.PartnerCapability)
	}
	key := locationHierarchy.CountryCode + "-" + locationHierarchy.RegionCode
	m.locationPartners[key] = partners
	return nil
}

func (m *mockCacheManager) GetServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string) ([]interfaces.ServiceDefinition, bool) {
	if m.shouldFail {
		return nil, false
	}
	key := strings.Join(serviceTypes, ",") + "-" + strings.Join(parcelCategories, ",")
	definitions, exists := m.serviceDefinitions[key]
	return definitions, exists
}

func (m *mockCacheManager) SetServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string, definitions []interfaces.ServiceDefinition) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	if m.serviceDefinitions == nil {
		m.serviceDefinitions = make(map[string][]interfaces.ServiceDefinition)
	}
	key := strings.Join(serviceTypes, ",") + "-" + strings.Join(parcelCategories, ",")
	m.serviceDefinitions[key] = definitions
	return nil
}

func (m *mockCacheManager) GetMultiplePartnerCapabilities(ctx context.Context, partnerIDs []uint) (map[uint]*interfaces.PartnerCapability, []uint) {
	if m.shouldFail {
		return make(map[uint]*interfaces.PartnerCapability), partnerIDs
	}

	result := make(map[uint]*interfaces.PartnerCapability)
	missing := []uint{}

	for _, id := range partnerIDs {
		if capability, exists := m.multipleCapabilities[id]; exists {
			result[id] = capability
		} else {
			missing = append(missing, id)
		}
	}

	return result, missing
}

func (m *mockCacheManager) SetMultiplePartnerCapabilities(ctx context.Context, capabilities map[uint]*interfaces.PartnerCapability) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	if m.multipleCapabilities == nil {
		m.multipleCapabilities = make(map[uint]*interfaces.PartnerCapability)
	}
	for id, capability := range capabilities {
		m.multipleCapabilities[id] = capability
	}
	return nil
}

func (m *mockCacheManager) InvalidatePartnerCapability(ctx context.Context, partnerID uint) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	delete(m.partnerCapabilities, partnerID)
	return nil
}

func (m *mockCacheManager) InvalidatePartnerServiceCapabilities(ctx context.Context, partnerID uint) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	delete(m.serviceCapabilities, partnerID)
	return nil
}

func (m *mockCacheManager) InvalidatePartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	key := locationHierarchy.CountryCode + "-" + locationHierarchy.RegionCode
	delete(m.locationPartners, key)
	return nil
}

func (m *mockCacheManager) InvalidateServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	key := strings.Join(serviceTypes, ",") + "-" + strings.Join(parcelCategories, ",")
	delete(m.serviceDefinitions, key)
	return nil
}

func (m *mockCacheManager) InvalidateMultiplePartners(ctx context.Context, partnerIDs []uint) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	for _, id := range partnerIDs {
		delete(m.multipleCapabilities, id)
	}
	return nil
}

// Additional interface methods
func (m *mockCacheManager) GetStats() cache.CacheStats {
	return cache.CacheStats{
		Hits:   100,
		Misses: 10,
		Size:   50,
	}
}

func (m *mockCacheManager) GetDetailedMetrics(ctx context.Context) CacheMetrics {
	return CacheMetrics{
		BaseStats: cache.CacheStats{
			Hits:   100,
			Misses: 10,
			Size:   50,
		},
		PartnerCapabilitiesCached: 25,
		ServiceCapabilitiesCached: 15,
		LocationPartnersCached:    5,
		ServiceDefinitionsCached:  5,
	}
}

func (m *mockCacheManager) InvalidateAll(ctx context.Context) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	m.partnerCapabilities = make(map[uint]*interfaces.PartnerCapability)
	m.serviceCapabilities = make(map[uint]*interfaces.PartnerServiceCapabilities)
	m.locationPartners = make(map[string][]interfaces.PartnerCapability)
	m.serviceDefinitions = make(map[string][]interfaces.ServiceDefinition)
	m.multipleCapabilities = make(map[uint]*interfaces.PartnerCapability)
	return nil
}

func (m *mockCacheManager) InvalidateExpired(ctx context.Context) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	return nil
}

func (m *mockCacheManager) GetTTL(ctx context.Context, cacheType, identifier string) (time.Duration, bool) {
	if m.shouldFail {
		return 0, false
	}
	return time.Hour, true
}

func (m *mockCacheManager) IsHealthy() bool {
	return !m.shouldFail
}

func (m *mockCacheManager) GetHealth() HealthStatus {
	return HealthStatus{
		IsHealthy:               !m.shouldFail,
		LastSuccessfulOperation: time.Now(),
		ConsecutiveErrors:       0,
		MemoryPressure:          0.1,
		ResponseTimeP95:         100,
		ErrorRate:               0.1,
		UpstreamHealthy:         true,
		CacheEfficiency:         0.9,
		RecommendedActions:      []string{},
	}
}

// Test data helpers
func createTestPartnerCapability(partnerID uint) *interfaces.PartnerCapability {
	return &interfaces.PartnerCapability{
		PartnerID:        partnerID,
		PartnerName:      "Test Partner",
		IsActive:         true,
		ServiceTypes:     []string{"Standard"},
		ParcelCategories: []string{"ecom"},
		OperationTypes:   []string{"pickup", "delivery"},
		PaymentModes:     []string{"ONLINE", "COD"},
		DeliveryModes:    []string{"SURFACE"},
	}
}

func createTestServiceCapabilities(partnerID uint) *interfaces.PartnerServiceCapabilities {
	return &interfaces.PartnerServiceCapabilities{
		PartnerID: partnerID,
		Capabilities: []interfaces.ServiceCapability{
			{
				ServiceType:    "Express",
				ParcelCategory: "courier",
				OperationTypes: []string{"pickup", "delivery"},
				PaymentModes:   []string{"ONLINE"},
				DeliveryModes:  []string{"AIR"},
				IsActive:       true,
				Rating:         4.5,
			},
		},
		Preferences: []interfaces.PartnerPreference{
			{
				ServiceType:     "Express",
				ParcelCategory:  "courier",
				IsPreferred:     true,
				Priority:        1,
				EffectiveRating: 4.5,
			},
		},
	}
}

func createTestLocationHierarchy() *models.LocationHierarchy {
	return &models.LocationHierarchy{
		CountryCode: "IN",
		RegionCode:  "KA",
		CityCode:    "BLR",
		PostalCode:  "560001",
	}
}

func createTestServiceDefinitions() []interfaces.ServiceDefinition {
	return []interfaces.ServiceDefinition{
		{
			ServiceType:       "Standard",
			ParcelCategory:    "ecom",
			DefaultOperations: []string{"pickup", "delivery"},
			DefaultPayments:   []string{"ONLINE", "COD"},
			DefaultDelivery:   []string{"SURFACE"},
			Description:       "Standard ecommerce delivery",
		},
		{
			ServiceType:       "Express",
			ParcelCategory:    "courier",
			DefaultOperations: []string{"pickup", "delivery"},
			DefaultPayments:   []string{"ONLINE"},
			DefaultDelivery:   []string{"AIR"},
			Description:       "Express courier service",
		},
	}
}

func createTestConfig() PartnerFallbackConfig {
	return PartnerFallbackConfig{
		EnableStaleCache:        true,
		StaleDataTTL:            time.Hour,
		EnableDefaultData:       true,
		MaxRetryAttempts:        3,
		RetryDelay:              time.Millisecond * 100,
		FallbackTimeout:         time.Second * 5,
		EnableMetrics:           true,
		EnableHealthTracking:    true,
		CircuitBreakerThreshold: 5,
		PartialFailureThreshold: 0.5,
		PreferCacheOverDefault:  true,
		EnableDegradedMode:      false,
	}
}

// Test cases start here
func TestNewPartnerCapabilityFallbackManager(t *testing.T) {
	tests := []struct {
		name   string
		cache  PartnerCapabilityCacheManager
		config PartnerFallbackConfig
		logger *log.Logger
	}{
		{
			name:   "Valid creation with cache",
			cache:  &mockCacheManager{},
			config: createTestConfig(),
			logger: log.New(os.Stdout, "test: ", log.LstdFlags),
		},
		{
			name:   "Valid creation without cache",
			cache:  nil,
			config: createTestConfig(),
			logger: log.New(os.Stdout, "test: ", log.LstdFlags),
		},
		{
			name:   "Valid creation with minimal config",
			cache:  &mockCacheManager{},
			config: PartnerFallbackConfig{},
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPartnerCapabilityFallbackManager(tt.cache, tt.config, tt.logger)
			if manager == nil {
				t.Error("Expected non-nil manager")
			}

			// Test interface compliance
			if _, ok := manager.(PartnerCapabilityFallbackManager); !ok {
				t.Error("Manager does not implement PartnerCapabilityFallbackManager interface")
			}
		})
	}
}

func TestGetPartnerCapabilityWithFallback_PrimarySuccess(t *testing.T) {
	mockCache := &mockCacheManager{}
	config := createTestConfig()
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	manager := NewPartnerCapabilityFallbackManager(mockCache, config, logger)

	ctx := context.Background()
	partnerID := uint(123)
	expectedCapability := createTestPartnerCapability(partnerID)

	primaryFetch := func(ctx context.Context) (*interfaces.PartnerCapability, error) {
		return expectedCapability, nil
	}

	capability, fallbackResult := manager.GetPartnerCapabilityWithFallback(ctx, partnerID, primaryFetch)

	if capability == nil {
		t.Error("Expected non-nil capability")
	}
	if capability.PartnerID != partnerID {
		t.Errorf("Expected partner ID %d, got %d", partnerID, capability.PartnerID)
	}
	if fallbackResult != nil {
		t.Error("Expected nil fallback result for successful primary fetch")
	}

	metrics := manager.GetMetrics()
	if metrics.PrimarySuccesses != 1 {
		t.Errorf("Expected 1 primary success, got %d", metrics.PrimarySuccesses)
	}
}

func TestGetPartnerCapabilityWithFallback_CacheFallback(t *testing.T) {
	partnerID := uint(123)
	expectedCapability := createTestPartnerCapability(partnerID)

	mockCache := &mockCacheManager{
		partnerCapabilities: map[uint]*interfaces.PartnerCapability{
			partnerID: expectedCapability,
		},
	}

	config := createTestConfig()
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	manager := NewPartnerCapabilityFallbackManager(mockCache, config, logger)

	ctx := context.Background()

	primaryFetch := func(ctx context.Context) (*interfaces.PartnerCapability, error) {
		return nil, errors.New("primary fetch failed")
	}

	capability, fallbackResult := manager.GetPartnerCapabilityWithFallback(ctx, partnerID, primaryFetch)

	if capability == nil {
		t.Error("Expected non-nil capability from cache fallback")
	}
	if capability.PartnerID != partnerID {
		t.Errorf("Expected partner ID %d, got %d", partnerID, capability.PartnerID)
	}
	if fallbackResult == nil {
		t.Error("Expected non-nil fallback result for cache fallback")
	}
	if fallbackResult.Strategy != infraErrors.FallbackStrategyCache {
		t.Errorf("Expected cache fallback strategy, got %v", fallbackResult.Strategy)
	}
	if !fallbackResult.Success {
		t.Error("Expected successful fallback result")
	}

	metrics := manager.GetMetrics()
	if metrics.PrimaryFailures != 1 {
		t.Errorf("Expected 1 primary failure, got %d", metrics.PrimaryFailures)
	}
	if metrics.CacheFallbacks != 1 {
		t.Errorf("Expected 1 cache fallback, got %d", metrics.CacheFallbacks)
	}
}

func TestGetPartnerCapabilityWithFallback_DefaultFallback(t *testing.T) {
	mockCache := &mockCacheManager{
		partnerCapabilities: map[uint]*interfaces.PartnerCapability{},
	}

	config := createTestConfig()
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	manager := NewPartnerCapabilityFallbackManager(mockCache, config, logger)

	ctx := context.Background()
	partnerID := uint(123)

	primaryFetch := func(ctx context.Context) (*interfaces.PartnerCapability, error) {
		return nil, errors.New("primary fetch failed")
	}

	capability, fallbackResult := manager.GetPartnerCapabilityWithFallback(ctx, partnerID, primaryFetch)

	if capability == nil {
		t.Error("Expected non-nil capability from default fallback")
	}
	if fallbackResult == nil {
		t.Error("Expected non-nil fallback result for default fallback")
	}
	if fallbackResult.Strategy != infraErrors.FallbackStrategyDefault {
		t.Errorf("Expected default fallback strategy, got %v", fallbackResult.Strategy)
	}

	metrics := manager.GetMetrics()
	if metrics.DefaultFallbacks != 1 {
		t.Errorf("Expected 1 default fallback, got %d", metrics.DefaultFallbacks)
	}
}

func TestGetPartnerCapabilityWithFallback_EmptyFallback(t *testing.T) {
	mockCache := &mockCacheManager{
		partnerCapabilities: map[uint]*interfaces.PartnerCapability{},
	}

	config := PartnerFallbackConfig{
		EnableDefaultData: false,
	}
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	manager := NewPartnerCapabilityFallbackManager(mockCache, config, logger)

	ctx := context.Background()
	partnerID := uint(123)

	primaryFetch := func(ctx context.Context) (*interfaces.PartnerCapability, error) {
		return nil, errors.New("primary fetch failed")
	}

	capability, fallbackResult := manager.GetPartnerCapabilityWithFallback(ctx, partnerID, primaryFetch)

	if capability != nil {
		t.Error("Expected nil capability when all fallbacks fail")
	}
	if fallbackResult == nil {
		t.Error("Expected non-nil fallback result for empty fallback")
	}
	if fallbackResult.Success {
		t.Error("Expected unsuccessful fallback result")
	}

	metrics := manager.GetMetrics()
	if metrics.EmptyFallbacks != 1 {
		t.Errorf("Expected 1 empty fallback, got %d", metrics.EmptyFallbacks)
	}
}

func TestGetPartnerServiceCapabilitiesWithFallback(t *testing.T) {
	partnerID := uint(123)
	expectedCapabilities := createTestServiceCapabilities(partnerID)

	mockCache := &mockCacheManager{
		serviceCapabilities: map[uint]*interfaces.PartnerServiceCapabilities{
			partnerID: expectedCapabilities,
		},
	}

	config := createTestConfig()
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	manager := NewPartnerCapabilityFallbackManager(mockCache, config, logger)

	ctx := context.Background()
	filters := &interfaces.ServiceFilters{
		ServiceTypes:     []string{"Express"},
		ParcelCategories: []string{"courier"},
	}

	primaryFetch := func(ctx context.Context) (*interfaces.PartnerServiceCapabilities, error) {
		return nil, errors.New("primary fetch failed")
	}

	capabilities, fallbackResult := manager.GetPartnerServiceCapabilitiesWithFallback(ctx, partnerID, filters, primaryFetch)

	if capabilities == nil {
		t.Error("Expected non-nil capabilities from cache fallback")
	}
	if capabilities.PartnerID != partnerID {
		t.Errorf("Expected partner ID %d, got %d", partnerID, capabilities.PartnerID)
	}
	if fallbackResult.Strategy != infraErrors.FallbackStrategyCache {
		t.Errorf("Expected cache fallback strategy, got %v", fallbackResult.Strategy)
	}

	metrics := manager.GetMetrics()
	if metrics.ServiceCapabilitiesFallbacks != 1 {
		t.Errorf("Expected 1 service capabilities fallback, got %d", metrics.ServiceCapabilitiesFallbacks)
	}
}

func TestGetPartnersByLocationWithFallback(t *testing.T) {
	location := createTestLocationHierarchy()
	expectedPartners := []interfaces.PartnerCapability{
		*createTestPartnerCapability(123),
		*createTestPartnerCapability(456),
	}

	mockCache := &mockCacheManager{
		locationPartners: map[string][]interfaces.PartnerCapability{
			"IN-KA": expectedPartners,
		},
	}

	config := createTestConfig()
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	manager := NewPartnerCapabilityFallbackManager(mockCache, config, logger)

	ctx := context.Background()

	primaryFetch := func(ctx context.Context) ([]interfaces.PartnerCapability, error) {
		return nil, errors.New("primary fetch failed")
	}

	partners, fallbackResult := manager.GetPartnersByLocationWithFallback(ctx, location, primaryFetch)

	if partners == nil {
		t.Error("Expected non-nil partners from cache fallback")
	}
	if len(partners) != 2 {
		t.Errorf("Expected 2 partners, got %d", len(partners))
	}
	if fallbackResult.Strategy != infraErrors.FallbackStrategyCache {
		t.Errorf("Expected cache fallback strategy, got %v", fallbackResult.Strategy)
	}

	metrics := manager.GetMetrics()
	if metrics.LocationPartnersFallbacks != 1 {
		t.Errorf("Expected 1 location partners fallback, got %d", metrics.LocationPartnersFallbacks)
	}
}

func TestRetryWithBackoff(t *testing.T) {
	config := createTestConfig()
	config.MaxRetryAttempts = 3
	config.RetryDelay = time.Millisecond * 10

	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	manager := NewPartnerCapabilityFallbackManager(nil, config, logger)

	ctx := context.Background()

	t.Run("Successful operation", func(t *testing.T) {
		attempts := 0
		operation := func(ctx context.Context) error {
			attempts++
			return nil
		}

		err := manager.RetryWithBackoff(ctx, operation)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("Retryable error with eventual success", func(t *testing.T) {
		attempts := 0
		operation := func(ctx context.Context) error {
			attempts++
			if attempts < 3 {
				return errors.New("retryable error")
			}
			return nil
		}

		err := manager.RetryWithBackoff(ctx, operation)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("Max retries exceeded", func(t *testing.T) {
		attempts := 0
		operation := func(ctx context.Context) error {
			attempts++
			return errors.New("persistent error")
		}

		err := manager.RetryWithBackoff(ctx, operation)
		if err == nil {
			t.Error("Expected error after max retries")
		}
		if attempts != 4 { // 1 initial + 3 retries
			t.Errorf("Expected 4 attempts, got %d", attempts)
		}
	})
}

func TestMetricsAndHealth(t *testing.T) {
	config := createTestConfig()
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	manager := NewPartnerCapabilityFallbackManager(nil, config, logger)

	// Test initial metrics
	metrics := manager.GetMetrics()
	if metrics == nil {
		t.Error("Expected non-nil metrics")
	}
	if metrics.HealthStatus != "healthy" {
		t.Errorf("Expected healthy status, got %s", metrics.HealthStatus)
	}

	// Test health check
	if !manager.IsHealthy() {
		t.Error("Expected manager to be healthy initially")
	}

	health := manager.GetHealth()
	if !health.IsHealthy {
		t.Error("Expected health to be healthy initially")
	}
	if health.Status != "healthy" {
		t.Errorf("Expected healthy status, got %s", health.Status)
	}

	// Test metrics reset
	manager.ResetMetrics()
	newMetrics := manager.GetMetrics()
	if newMetrics.TotalFallbacks != 0 {
		t.Errorf("Expected 0 total fallbacks after reset, got %d", newMetrics.TotalFallbacks)
	}
}
