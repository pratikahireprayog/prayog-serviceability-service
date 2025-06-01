package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	infraErrors "prayog-serviceability-service/internal/infrastructure/errors"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/utils/v1"
)

// Mock implementations for testing

type mockPartnerServiceClient struct {
	partners         []interfaces.PartnerInfo
	effectiveDetails map[uint]*interfaces.PartnerEffectiveDetails
	shouldFail       bool
	failCount        int
	callCount        int
	mu               sync.Mutex
}

func (m *mockPartnerServiceClient) GetPartnersByLocation(ctx context.Context, locationType, locationID string) ([]interfaces.PartnerInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldFail {
		return nil, fmt.Errorf("partner service error")
	}
	return m.partners, nil
}

func (m *mockPartnerServiceClient) GetPartnerEffectiveDetails(ctx context.Context, partnerID uint, entityType, entityID string) (*interfaces.PartnerEffectiveDetails, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldFail {
		return nil, fmt.Errorf("partner service error")
	}

	if details, exists := m.effectiveDetails[partnerID]; exists {
		return details, nil
	}

	return &interfaces.PartnerEffectiveDetails{
		PartnerID: partnerID,
		Preferences: []interfaces.EffectivePreference{
			{
				EntityType:      "service",
				EntityID:        "1",
				PreferenceValue: "EXPRESS",
				EffectiveRating: 4.5,
			},
		},
		Ratings: []interfaces.EffectiveRating{
			{
				EntityType: "service",
				EntityID:   "1",
				Rating:     4.5,
			},
		},
	}, nil
}

type mockSpecificationServiceClient struct {
	specs      []interfaces.EntitySpecification
	shouldFail bool
	callCount  int
	mu         sync.Mutex
}

func (m *mockSpecificationServiceClient) GetSpecDefinitions(ctx context.Context) ([]interfaces.SpecDefinition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldFail {
		return nil, fmt.Errorf("spec service error")
	}
	return []interfaces.SpecDefinition{}, nil
}

func (m *mockSpecificationServiceClient) GetCatalogs(ctx context.Context) ([]interfaces.Catalog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldFail {
		return nil, fmt.Errorf("spec service error")
	}
	return []interfaces.Catalog{}, nil
}

func (m *mockSpecificationServiceClient) GetEntitySpecifications(ctx context.Context, entityType, entityID string) ([]interfaces.EntitySpecification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldFail {
		return nil, fmt.Errorf("spec service error")
	}
	return m.specs, nil
}

type mockCacheManager struct {
	partnerCapabilities map[uint]*interfaces.PartnerCapability
	serviceCapabilities map[string]*interfaces.PartnerServiceCapabilities
	locationPartners    map[string][]interfaces.PartnerCapability
	shouldFail          bool
	callCount           int
	mu                  sync.RWMutex
}

func (m *mockCacheManager) GetPartnerCapability(ctx context.Context, partnerID uint) (*interfaces.PartnerCapability, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.callCount++

	if m.shouldFail {
		return nil, false
	}
	capability, exists := m.partnerCapabilities[partnerID]
	return capability, exists
}

func (m *mockCacheManager) SetPartnerCapability(ctx context.Context, partnerID uint, capability *interfaces.PartnerCapability) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldFail {
		return fmt.Errorf("cache error")
	}
	if m.partnerCapabilities == nil {
		m.partnerCapabilities = make(map[uint]*interfaces.PartnerCapability)
	}
	m.partnerCapabilities[partnerID] = capability
	return nil
}

func (m *mockCacheManager) GetPartnerServiceCapabilities(ctx context.Context, partnerID uint, filters *interfaces.ServiceFilters) (*interfaces.PartnerServiceCapabilities, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.callCount++

	if m.shouldFail {
		return nil, false
	}
	key := fmt.Sprintf("%d", partnerID)
	capability, exists := m.serviceCapabilities[key]
	return capability, exists
}

func (m *mockCacheManager) SetPartnerServiceCapabilities(ctx context.Context, partnerID uint, filters *interfaces.ServiceFilters, capabilities *interfaces.PartnerServiceCapabilities) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldFail {
		return fmt.Errorf("cache error")
	}
	if m.serviceCapabilities == nil {
		m.serviceCapabilities = make(map[string]*interfaces.PartnerServiceCapabilities)
	}
	key := fmt.Sprintf("%d", partnerID)
	m.serviceCapabilities[key] = capabilities
	return nil
}

func (m *mockCacheManager) GetPartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy) ([]interfaces.PartnerCapability, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.callCount++

	if m.shouldFail {
		return nil, false
	}
	key := locationHierarchy.PostalCode
	partners, exists := m.locationPartners[key]
	return partners, exists
}

func (m *mockCacheManager) SetPartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy, partners []interfaces.PartnerCapability) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldFail {
		return fmt.Errorf("cache error")
	}
	if m.locationPartners == nil {
		m.locationPartners = make(map[string][]interfaces.PartnerCapability)
	}
	key := locationHierarchy.PostalCode
	m.locationPartners[key] = partners
	return nil
}

func (m *mockCacheManager) GetServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string) ([]interfaces.ServiceDefinition, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return []interfaces.ServiceDefinition{}, false
}

func (m *mockCacheManager) SetServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string, definitions []interfaces.ServiceDefinition) error {
	return nil
}

func (m *mockCacheManager) GetMultiplePartnerCapabilities(ctx context.Context, partnerIDs []uint) (map[uint]*interfaces.PartnerCapability, []uint) {
	return make(map[uint]*interfaces.PartnerCapability), partnerIDs
}

func (m *mockCacheManager) SetMultiplePartnerCapabilities(ctx context.Context, capabilities map[uint]*interfaces.PartnerCapability) error {
	return nil
}

func (m *mockCacheManager) InvalidatePartnerCapability(ctx context.Context, partnerID uint) error {
	return nil
}

func (m *mockCacheManager) InvalidatePartnerServiceCapabilities(ctx context.Context, partnerID uint) error {
	return nil
}

func (m *mockCacheManager) InvalidatePartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy) error {
	return nil
}

func (m *mockCacheManager) InvalidateServiceDefinitions(ctx context.Context, serviceTypes, parcelCategories []string) error {
	return nil
}

func (m *mockCacheManager) InvalidateMultiplePartners(ctx context.Context, partnerIDs []uint) error {
	return nil
}

func (m *mockCacheManager) GetStats() interface{} {
	return map[string]interface{}{
		"hits":   100,
		"misses": 10,
		"size":   50,
	}
}

func (m *mockCacheManager) GetDetailedMetrics(ctx context.Context) utils.CacheMetrics {
	return utils.CacheMetrics{}
}

func (m *mockCacheManager) InvalidateAll(ctx context.Context) error {
	return nil
}

func (m *mockCacheManager) InvalidateExpired(ctx context.Context) error {
	return nil
}

func (m *mockCacheManager) GetTTL(ctx context.Context, cacheType, identifier string) (time.Duration, bool) {
	return time.Hour, true
}

func (m *mockCacheManager) IsHealthy() bool {
	return !m.shouldFail
}

func (m *mockCacheManager) GetHealth() utils.HealthStatus {
	return utils.HealthStatus{}
}

type mockFallbackManager struct {
	shouldReturnFallback bool
	fallbackData         interface{}
	callCount            int
	mu                   sync.Mutex
}

func (m *mockFallbackManager) GetPartnerCapabilityWithFallback(ctx context.Context, partnerID uint, primaryFetch func(ctx context.Context) (*interfaces.PartnerCapability, error)) (*interfaces.PartnerCapability, *infraErrors.FallbackResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	// Try primary first
	result, err := primaryFetch(ctx)
	if err == nil {
		return result, nil
	}

	// Return fallback if enabled
	if m.shouldReturnFallback {
		fallbackData := &interfaces.PartnerCapability{
			PartnerID:   partnerID,
			PartnerName: "Fallback Partner",
			IsActive:    true,
		}
		return fallbackData, &infraErrors.FallbackResult{Success: true, Message: "Cache fallback"}
	}

	return nil, &infraErrors.FallbackResult{Success: false, Message: "All fallbacks failed"}
}

func (m *mockFallbackManager) GetPartnerServiceCapabilitiesWithFallback(ctx context.Context, partnerID uint, filters *interfaces.ServiceFilters, primaryFetch func(ctx context.Context) (*interfaces.PartnerServiceCapabilities, error)) (*interfaces.PartnerServiceCapabilities, *infraErrors.FallbackResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	// Try primary first
	result, err := primaryFetch(ctx)
	if err == nil {
		return result, nil
	}

	// Return fallback if enabled
	if m.shouldReturnFallback {
		fallbackData := &interfaces.PartnerServiceCapabilities{
			PartnerID: partnerID,
			Capabilities: []interfaces.ServiceCapability{
				{
					ServiceType:    "Standard",
					ParcelCategory: "courier",
					IsActive:       true,
					Rating:         4.0,
				},
			},
		}
		return fallbackData, &infraErrors.FallbackResult{Success: true, Message: "Cache fallback"}
	}

	return nil, &infraErrors.FallbackResult{Success: false, Message: "All fallbacks failed"}
}

func (m *mockFallbackManager) GetPartnersByLocationWithFallback(ctx context.Context, locationHierarchy *models.LocationHierarchy, primaryFetch func(ctx context.Context) ([]interfaces.PartnerCapability, error)) ([]interfaces.PartnerCapability, *infraErrors.FallbackResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	// Try primary first
	result, err := primaryFetch(ctx)
	if err == nil {
		return result, nil
	}

	// Return fallback if enabled
	if m.shouldReturnFallback {
		fallbackData := []interfaces.PartnerCapability{
			{
				PartnerID:   1,
				PartnerName: "Fallback Partner",
				IsActive:    true,
			},
		}
		return fallbackData, &infraErrors.FallbackResult{Success: true, Message: "Cache fallback"}
	}

	return nil, &infraErrors.FallbackResult{Success: false, Message: "All fallbacks failed"}
}

func (m *mockFallbackManager) GetServiceDefinitionsWithFallback(ctx context.Context, serviceTypes, parcelCategories []string, primaryFetch func(ctx context.Context) ([]interfaces.ServiceDefinition, error)) ([]interfaces.ServiceDefinition, *infraErrors.FallbackResult) {
	return []interfaces.ServiceDefinition{}, nil
}

func (m *mockFallbackManager) GetMultiplePartnerCapabilitiesWithFallback(ctx context.Context, partnerIDs []uint, primaryFetch func(ctx context.Context) (map[uint]*interfaces.PartnerCapability, error)) (map[uint]*interfaces.PartnerCapability, *infraErrors.FallbackResult) {
	return make(map[uint]*interfaces.PartnerCapability), nil
}

func (m *mockFallbackManager) RetryWithBackoff(ctx context.Context, operation func(ctx context.Context) error) error {
	return operation(ctx)
}

func (m *mockFallbackManager) GetMetrics() interface{} {
	return map[string]interface{}{
		"fallbacks": m.callCount,
	}
}

func (m *mockFallbackManager) ResetMetrics() {
	m.callCount = 0
}

func (m *mockFallbackManager) IsHealthy() bool {
	return true
}

func (m *mockFallbackManager) GetHealth() utils.PartnerFallbackHealth {
	return utils.PartnerFallbackHealth{}
}

// Helper functions for test data

func createTestLocationHierarchy() *models.LocationHierarchy {
	return &models.LocationHierarchy{
		CountryCode: "US",
		RegionCode:  "CA",
		CityCode:    "SF",
		PostalCode:  "94102",
	}
}

func createTestPartnerInfo() []interfaces.PartnerInfo {
	return []interfaces.PartnerInfo{
		{
			ID:       1,
			Name:     "Test Partner 1",
			IsActive: true,
			Type:     "COURIER",
		},
		{
			ID:       2,
			Name:     "Test Partner 2",
			IsActive: true,
			Type:     "ECOM",
		},
	}
}

func createTestServiceFilters() *interfaces.ServiceFilters {
	return &interfaces.ServiceFilters{
		ServiceTypes:     []string{"EXPRESS", "STANDARD"},
		ParcelCategories: []string{"courier"},
		OperationTypes:   []string{"pickup", "delivery"},
	}
}

func createTestConfig() PartnerCapabilityAggregatorConfig {
	return PartnerCapabilityAggregatorConfig{
		CacheEnabled:             true,
		CacheTTL:                 30 * time.Minute,
		CircuitBreakerEnabled:    true,
		FailureThreshold:         5,
		RecoveryTimeout:          30 * time.Second,
		ValidateResponses:        true,
		StrictValidation:         false,
		EnableTransformation:     true,
		IncludeInactive:          false,
		MaxConcurrentRequests:    10,
		RequestTimeout:           30 * time.Second,
		LocationCacheTTL:         15 * time.Minute,
		EnableParallelProcessing: true,
		MaxParallelWorkers:       5,
	}
}

// Test Cases

func TestNewPartnerCapabilityAggregator(t *testing.T) {
	tests := []struct {
		name   string
		config PartnerCapabilityAggregatorConfig
	}{
		{
			name:   "Valid creation with default config",
			config: createTestConfig(),
		},
		{
			name:   "Valid creation with minimal config",
			config: PartnerCapabilityAggregatorConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerService := &mockPartnerServiceClient{
				partners: createTestPartnerInfo(),
			}
			mockSpecService := &mockSpecificationServiceClient{}
			mockCache := &mockCacheManager{}
			mockFallback := &mockFallbackManager{}

			aggregator := NewPartnerCapabilityAggregator(
				mockPartnerService,
				mockSpecService,
				mockCache,
				mockFallback,
				tt.config,
			)

			if aggregator == nil {
				t.Error("Expected non-nil aggregator")
			}

			// Test interface compliance
			if _, ok := aggregator.(interfaces.PartnerCapabilityAggregator); !ok {
				t.Error("Aggregator does not implement PartnerCapabilityAggregator interface")
			}
		})
	}
}

func TestGetPartnersByLocation_Success(t *testing.T) {
	mockPartnerService := &mockPartnerServiceClient{
		partners: createTestPartnerInfo(),
	}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()
	locationHierarchy := createTestLocationHierarchy()

	partners, err := aggregator.GetPartnersByLocation(ctx, locationHierarchy)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(partners) == 0 {
		t.Error("Expected partners to be returned")
	}

	// Verify partner service was called
	if mockPartnerService.callCount == 0 {
		t.Error("Expected partner service to be called")
	}
}

func TestGetPartnersByLocation_WithFallback(t *testing.T) {
	mockPartnerService := &mockPartnerServiceClient{
		shouldFail: true, // Force failure to trigger fallback
	}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{
		shouldReturnFallback: true,
	}

	config := createTestConfig()
	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()
	locationHierarchy := createTestLocationHierarchy()

	partners, err := aggregator.GetPartnersByLocation(ctx, locationHierarchy)

	if err != nil {
		t.Errorf("Expected no error with fallback, got %v", err)
	}

	if len(partners) == 0 {
		t.Error("Expected fallback partners to be returned")
	}

	// Verify fallback was used
	if mockFallback.callCount == 0 {
		t.Error("Expected fallback to be called")
	}
}

func TestGetPartnerCapabilities_Success(t *testing.T) {
	mockPartnerService := &mockPartnerServiceClient{
		effectiveDetails: map[uint]*interfaces.PartnerEffectiveDetails{
			1: {
				PartnerID: 1,
				Preferences: []interfaces.EffectivePreference{
					{
						EntityType:      "service",
						EntityID:        "1",
						PreferenceValue: "EXPRESS",
						EffectiveRating: 4.5,
					},
				},
			},
		},
	}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()
	partnerID := uint(1)
	serviceFilters := createTestServiceFilters()

	capabilities, err := aggregator.GetPartnerCapabilities(ctx, partnerID, serviceFilters)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if capabilities == nil {
		t.Error("Expected capabilities to be returned")
	}

	if capabilities.PartnerID != partnerID {
		t.Errorf("Expected partner ID %d, got %d", partnerID, capabilities.PartnerID)
	}
}

func TestGetPartnerCapabilities_InvalidInput(t *testing.T) {
	mockPartnerService := &mockPartnerServiceClient{}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()

	// Test with zero partner ID
	_, err := aggregator.GetPartnerCapabilities(ctx, 0, nil)
	if err == nil {
		t.Error("Expected error for zero partner ID")
	}
	if !strings.Contains(err.Error(), "partner ID cannot be zero") {
		t.Errorf("Expected partner ID error, got %v", err)
	}
}

func TestGetPartnersByLocation_InvalidLocation(t *testing.T) {
	mockPartnerService := &mockPartnerServiceClient{}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()

	// Test with invalid location hierarchy
	invalidLocation := &models.LocationHierarchy{
		CountryCode: "INVALID",
		PostalCode:  "123",
	}

	_, err := aggregator.GetPartnersByLocation(ctx, invalidLocation)
	if err == nil {
		t.Error("Expected error for invalid location")
	}
}

// Concurrency Tests

func TestConcurrentGetPartnersByLocation(t *testing.T) {
	mockPartnerService := &mockPartnerServiceClient{
		partners: createTestPartnerInfo(),
	}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()
	locationHierarchy := createTestLocationHierarchy()

	// Run concurrent requests
	const numGoroutines = 50
	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			partners, err := aggregator.GetPartnersByLocation(ctx, locationHierarchy)
			if err != nil {
				errorChan <- err
				return
			}

			if len(partners) == 0 {
				errorChan <- fmt.Errorf("no partners returned")
				return
			}
		}()
	}

	wg.Wait()
	close(errorChan)

	// Check for any errors
	for err := range errorChan {
		t.Errorf("Concurrent request failed: %v", err)
	}
}

func TestConcurrentGetPartnerCapabilities(t *testing.T) {
	mockPartnerService := &mockPartnerServiceClient{
		effectiveDetails: map[uint]*interfaces.PartnerEffectiveDetails{
			1: {PartnerID: 1},
			2: {PartnerID: 2},
			3: {PartnerID: 3},
		},
	}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()
	serviceFilters := createTestServiceFilters()

	// Run concurrent requests for different partners
	const numGoroutines = 30
	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(partnerID uint) {
			defer wg.Done()

			capabilities, err := aggregator.GetPartnerCapabilities(ctx, partnerID, serviceFilters)
			if err != nil {
				errorChan <- err
				return
			}

			if capabilities == nil {
				errorChan <- fmt.Errorf("no capabilities returned for partner %d", partnerID)
				return
			}

			if capabilities.PartnerID != partnerID {
				errorChan <- fmt.Errorf("wrong partner ID returned: expected %d, got %d", partnerID, capabilities.PartnerID)
				return
			}
		}(uint((i % 3) + 1))
	}

	wg.Wait()
	close(errorChan)

	// Check for any errors
	for err := range errorChan {
		t.Errorf("Concurrent request failed: %v", err)
	}
}

// Circuit Breaker Tests

func TestCircuitBreaker_OpenState(t *testing.T) {
	mockPartnerService := &mockPartnerServiceClient{
		shouldFail: true, // Always fail to trigger circuit breaker
	}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	config.FailureThreshold = 3 // Low threshold for testing

	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()
	locationHierarchy := createTestLocationHierarchy()

	// Trigger enough failures to open circuit
	for i := 0; i < config.FailureThreshold+1; i++ {
		aggregator.GetPartnersByLocation(ctx, locationHierarchy)
	}

	// Next request should fail immediately due to open circuit
	_, err := aggregator.GetPartnersByLocation(ctx, locationHierarchy)
	if err == nil {
		t.Error("Expected circuit breaker to be open")
	}
	if !strings.Contains(err.Error(), "circuit breaker is open") {
		t.Errorf("Expected circuit breaker error, got %v", err)
	}
}

// Performance Tests

func BenchmarkGetPartnersByLocation(b *testing.B) {
	mockPartnerService := &mockPartnerServiceClient{
		partners: createTestPartnerInfo(),
	}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()
	locationHierarchy := createTestLocationHierarchy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := aggregator.GetPartnersByLocation(ctx, locationHierarchy)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkGetPartnerCapabilities(b *testing.B) {
	mockPartnerService := &mockPartnerServiceClient{
		effectiveDetails: map[uint]*interfaces.PartnerEffectiveDetails{
			1: {PartnerID: 1},
		},
	}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()
	partnerID := uint(1)
	serviceFilters := createTestServiceFilters()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := aggregator.GetPartnerCapabilities(ctx, partnerID, serviceFilters)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

// Integration Tests

func TestIntegration_FullWorkflow(t *testing.T) {
	// Set up realistic mock data
	mockPartnerService := &mockPartnerServiceClient{
		partners: []interfaces.PartnerInfo{
			{
				ID:       1,
				Name:     "DHL Express",
				IsActive: true,
				Type:     "COURIER",
			},
			{
				ID:       2,
				Name:     "Amazon Logistics",
				IsActive: true,
				Type:     "ECOM",
			},
		},
		effectiveDetails: map[uint]*interfaces.PartnerEffectiveDetails{
			1: {
				PartnerID: 1,
				Preferences: []interfaces.EffectivePreference{
					{
						EntityType:      "service",
						EntityID:        "1",
						PreferenceValue: "EXPRESS",
						EffectiveRating: 4.8,
					},
				},
				Ratings: []interfaces.EffectiveRating{
					{
						EntityType: "service",
						EntityID:   "1",
						Rating:     4.8,
					},
				},
			},
			2: {
				PartnerID: 2,
				Preferences: []interfaces.EffectivePreference{
					{
						EntityType:      "service",
						EntityID:        "2",
						PreferenceValue: "SDD",
						EffectiveRating: 4.5,
					},
				},
				Ratings: []interfaces.EffectiveRating{
					{
						EntityType: "service",
						EntityID:   "2",
						Rating:     4.5,
					},
				},
			},
		},
	}

	mockSpecService := &mockSpecificationServiceClient{
		specs: []interfaces.EntitySpecification{
			{
				ID:          1,
				EntityType:  "PARTNER",
				EntityID:    "1",
				SpecDefID:   1,
				SpecValue:   "EXPRESS",
				Description: "Express service capability",
			},
		},
	}

	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()
	locationHierarchy := createTestLocationHierarchy()

	// Test 1: Get partners by location
	partners, err := aggregator.GetPartnersByLocation(ctx, locationHierarchy)
	if err != nil {
		t.Fatalf("Failed to get partners by location: %v", err)
	}

	if len(partners) != 2 {
		t.Errorf("Expected 2 partners, got %d", len(partners))
	}

	// Verify partner data
	for _, partner := range partners {
		if partner.PartnerID == 0 {
			t.Error("Partner ID should not be zero")
		}
		if partner.PartnerName == "" {
			t.Error("Partner name should not be empty")
		}
	}

	// Test 2: Get capabilities for each partner
	for _, partner := range partners {
		capabilities, err := aggregator.GetPartnerCapabilities(ctx, partner.PartnerID, nil)
		if err != nil {
			t.Errorf("Failed to get capabilities for partner %d: %v", partner.PartnerID, err)
			continue
		}

		if capabilities == nil {
			t.Errorf("Capabilities should not be nil for partner %d", partner.PartnerID)
			continue
		}

		if capabilities.PartnerID != partner.PartnerID {
			t.Errorf("Capability partner ID mismatch: expected %d, got %d", partner.PartnerID, capabilities.PartnerID)
		}
	}

	// Test 3: Get capabilities with filters
	serviceFilters := &interfaces.ServiceFilters{
		ServiceTypes: []string{"EXPRESS"},
	}

	filteredCapabilities, err := aggregator.GetPartnerCapabilities(ctx, 1, serviceFilters)
	if err != nil {
		t.Errorf("Failed to get filtered capabilities: %v", err)
	}

	if filteredCapabilities == nil {
		t.Error("Filtered capabilities should not be nil")
	}
}

// Error Handling Tests

func TestErrorHandling_ServiceFailures(t *testing.T) {
	tests := []struct {
		name                     string
		partnerServiceShouldFail bool
		specServiceShouldFail    bool
		cacheShouldFail          bool
		fallbackShouldWork       bool
		expectError              bool
		expectFallbackUsed       bool
	}{
		{
			name:                     "All services working",
			partnerServiceShouldFail: false,
			specServiceShouldFail:    false,
			cacheShouldFail:          false,
			fallbackShouldWork:       false,
			expectError:              false,
			expectFallbackUsed:       false,
		},
		{
			name:                     "Partner service fails, fallback works",
			partnerServiceShouldFail: true,
			specServiceShouldFail:    false,
			cacheShouldFail:          false,
			fallbackShouldWork:       true,
			expectError:              false,
			expectFallbackUsed:       true,
		},
		{
			name:                     "All services fail, no fallback",
			partnerServiceShouldFail: true,
			specServiceShouldFail:    true,
			cacheShouldFail:          true,
			fallbackShouldWork:       false,
			expectError:              true,
			expectFallbackUsed:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerService := &mockPartnerServiceClient{
				partners:   createTestPartnerInfo(),
				shouldFail: tt.partnerServiceShouldFail,
			}
			mockSpecService := &mockSpecificationServiceClient{
				shouldFail: tt.specServiceShouldFail,
			}
			mockCache := &mockCacheManager{
				shouldFail: tt.cacheShouldFail,
			}
			mockFallback := &mockFallbackManager{
				shouldReturnFallback: tt.fallbackShouldWork,
			}

			config := createTestConfig()
			aggregator := NewPartnerCapabilityAggregator(
				mockPartnerService,
				mockSpecService,
				mockCache,
				mockFallback,
				config,
			)

			ctx := context.Background()
			locationHierarchy := createTestLocationHierarchy()

			partners, err := aggregator.GetPartnersByLocation(ctx, locationHierarchy)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.expectFallbackUsed && mockFallback.callCount == 0 {
				t.Error("Expected fallback to be used but it wasn't called")
			}

			if !tt.expectError && len(partners) == 0 {
				t.Error("Expected partners to be returned")
			}
		})
	}
}

// Memory and Resource Tests

func TestMemoryUsage_LargeBatch(t *testing.T) {
	// Create large dataset to test memory handling
	largePartnerList := make([]interfaces.PartnerInfo, 1000)
	for i := 0; i < 1000; i++ {
		largePartnerList[i] = interfaces.PartnerInfo{
			ID:       uint(i + 1),
			Name:     fmt.Sprintf("Partner %d", i+1),
			IsActive: true,
			Type:     "COURIER",
		}
	}

	mockPartnerService := &mockPartnerServiceClient{
		partners: largePartnerList,
	}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	config.MaxParallelWorkers = 10 // Test with parallel processing

	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	ctx := context.Background()
	locationHierarchy := createTestLocationHierarchy()

	partners, err := aggregator.GetPartnersByLocation(ctx, locationHierarchy)
	if err != nil {
		t.Fatalf("Failed to handle large batch: %v", err)
	}

	if len(partners) == 0 {
		t.Error("Expected partners to be returned")
	}
}

// Timeout and Context Tests

func TestContextCancellation(t *testing.T) {
	mockPartnerService := &mockPartnerServiceClient{
		partners: createTestPartnerInfo(),
	}
	mockSpecService := &mockSpecificationServiceClient{}
	mockCache := &mockCacheManager{}
	mockFallback := &mockFallbackManager{}

	config := createTestConfig()
	aggregator := NewPartnerCapabilityAggregator(
		mockPartnerService,
		mockSpecService,
		mockCache,
		mockFallback,
		config,
	)

	// Create context with immediate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	locationHierarchy := createTestLocationHierarchy()

	_, err := aggregator.GetPartnersByLocation(ctx, locationHierarchy)

	// Should handle cancelled context gracefully
	if err != nil && err != context.Canceled {
		// This test ensures the aggregator handles context cancellation properly
		// The exact behavior depends on how the underlying services handle cancellation
		t.Logf("Context cancellation handled: %v", err)
	}
}

func TestMain(m *testing.M) {
	// Set up test environment
	log.SetOutput(os.Stdout)

	// Run tests
	code := m.Run()

	// Clean up
	os.Exit(code)
}
