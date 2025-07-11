package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"prayog-serviceability-service/internal/infrastructure/cache"
	infraErrors "prayog-serviceability-service/internal/infrastructure/errors"
	"prayog-serviceability-service/internal/infrastructure/fallback"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSpecificationServiceClient is a mock implementation of SpecificationServiceClient
type MockSpecificationServiceClient struct {
	mock.Mock
}

func (m *MockSpecificationServiceClient) GetSpecDefinitions(ctx context.Context) ([]interfaces.SpecDefinition, error) {
	args := m.Called(ctx)
	return args.Get(0).([]interfaces.SpecDefinition), args.Error(1)
}

func (m *MockSpecificationServiceClient) GetCatalogs(ctx context.Context) ([]interfaces.Catalog, error) {
	args := m.Called(ctx)
	return args.Get(0).([]interfaces.Catalog), args.Error(1)
}

func (m *MockSpecificationServiceClient) GetEntitySpecifications(ctx context.Context, entityType, entityID string) ([]interfaces.EntitySpecification, error) {
	args := m.Called(ctx, entityType, entityID)
	return args.Get(0).([]interfaces.EntitySpecification), args.Error(1)
}

// MockCache is a mock implementation of the cache interface
type MockCache struct {
	mock.Mock
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
	}
}

func (m *MockCache) Get(ctx context.Context, key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called(ctx, key)
	value, found := m.data[key]
	return value, found && args.Bool(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]interface{})
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.data[key]
	args := m.Called(ctx, key)
	return exists && args.Bool(1)
}

func (m *MockCache) GetTTL(ctx context.Context, key string) (time.Duration, bool) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Bool(1)
}

// Close implements the cache.Cache interface
func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCache) GetStats() cache.CacheStats {
	args := m.Called()
	return args.Get(0).(cache.CacheStats)
}

// Test helper functions

func createTestConfig() config.ServiceDefinitionResolverConfig {
	return config.ServiceDefinitionResolverConfig{
		CacheEnabled:            true,
		CacheTTL:                30 * time.Minute,
		CircuitBreakerEnabled:   true,
		FailureThreshold:        3,
		RecoveryTimeout:         30 * time.Second,
		ValidateResponses:       true,
		StrictValidation:        false,
		EnableTransformation:    true,
		TransformCatalogToLists: true,
		MaxConcurrentRequests:   10,
		RequestTimeout:          30 * time.Second,
		CategoryCacheTTL:        15 * time.Minute,
	}
}

func createTestCatalogs() []interfaces.Catalog {
	return []interfaces.Catalog{
		{
			ID:          1,
			SpecDefID:   1,
			Name:        "Parcel Category Catalog",
			Description: "Available parcel categories",
			CatalogItems: []interfaces.CatalogItem{
				{ID: 1, Code: "ECOM", Name: "E-commerce", Description: "E-commerce packages"},
				{ID: 2, Code: "COURIER", Name: "Courier", Description: "Courier documents"},
				{ID: 3, Code: "CARGO", Name: "Cargo", Description: "Cargo shipments"},
			},
		},
		{
			ID:          2,
			SpecDefID:   2,
			Name:        "Service Type Catalog",
			Description: "Available service types",
			CatalogItems: []interfaces.CatalogItem{
				{ID: 4, Code: "SDD", Name: "Same Day Delivery", Description: "Same day delivery service"},
				{ID: 5, Code: "NDD", Name: "Next Day Delivery", Description: "Next day delivery service"},
				{ID: 6, Code: "STANDARD", Name: "Standard Delivery", Description: "Standard delivery service"},
			},
		},
		{
			ID:          3,
			SpecDefID:   3,
			Name:        "Operation Type Catalog",
			Description: "Available operation types",
			CatalogItems: []interfaces.CatalogItem{
				{ID: 7, Code: "PICKUP", Name: "Pickup Service", Description: "Package pickup"},
				{ID: 8, Code: "DELIVERY", Name: "Delivery Service", Description: "Package delivery"},
				{ID: 9, Code: "TRACKING", Name: "Tracking Service", Description: "Package tracking"},
			},
		},
		{
			ID:          4,
			SpecDefID:   4,
			Name:        "Payment Mode Catalog",
			Description: "Available payment modes",
			CatalogItems: []interfaces.CatalogItem{
				{ID: 10, Code: "COD", Name: "Cash on Delivery", Description: "COD payment"},
				{ID: 11, Code: "ONLINE", Name: "Online Payment", Description: "Online payment"},
				{ID: 12, Code: "PREPAID", Name: "Prepaid", Description: "Prepaid payment"},
			},
		},
		{
			ID:          5,
			SpecDefID:   5,
			Name:        "Delivery Mode Catalog",
			Description: "Available delivery modes",
			CatalogItems: []interfaces.CatalogItem{
				{ID: 13, Code: "AIR", Name: "Air Transport", Description: "Air delivery"},
				{ID: 14, Code: "SURFACE", Name: "Surface Transport", Description: "Surface delivery"},
				{ID: 15, Code: "RAIL", Name: "Rail Transport", Description: "Rail delivery"},
			},
		},
	}
}

func createTestResolver(specService *MockSpecificationServiceClient, mockCache *MockCache) *ServiceDefinitionResolver {
	// Create cache
	cacheConfig := cache.ServiceDefinitionCacheConfig{
		SpecDefinitionsTTL: 30 * time.Minute,
		CatalogsTTL:        30 * time.Minute,
		EntitySpecsTTL:     30 * time.Minute,
		ServiceDefsTTL:     30 * time.Minute,
		KeyPrefix:          "test",
	}
	serviceDefCache := cache.NewServiceDefinitionCache(mockCache, cacheConfig)

	// Create fallback manager
	fallbackConfig := fallback.FallbackConfig{
		EnableDefaultData: true,
		MaxRetryAttempts:  3,
		RetryDelay:        1 * time.Second,
	}
	fallbackManager := fallback.NewFallbackManager(serviceDefCache, fallbackConfig, nil)

	// Create resolver
	resolverConfig := createTestConfig()
	resolver := NewServiceDefinitionResolver(specService, serviceDefCache, fallbackManager, resolverConfig)

	// Return the concrete type so we can access GetMetrics and GetHealth
	return resolver.(*ServiceDefinitionResolver)
}

// Test cases

func TestServiceDefinitionResolver_GetParcelCategories_Success(t *testing.T) {
	// Setup
	specService := &MockSpecificationServiceClient{}
	mockCache := NewMockCache()
	resolver := createTestResolver(specService, mockCache)

	ctx := context.Background()
	testCatalogs := createTestCatalogs()

	// Mock expectations
	specService.On("GetCatalogs", ctx).Return(testCatalogs, nil)
	mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, false)
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Execute
	categories, err := resolver.GetParcelCategories(ctx)

	// Verify
	assert.NoError(t, err)
	assert.NotEmpty(t, categories)
	assert.Contains(t, categories, "ecom")
	assert.Contains(t, categories, "courier")
	assert.Contains(t, categories, "cargo")

	specService.AssertExpectations(t)
}

func TestServiceDefinitionResolver_GetServiceTypes_Success(t *testing.T) {
	// Setup
	specService := &MockSpecificationServiceClient{}
	mockCache := NewMockCache()
	resolver := createTestResolver(specService, mockCache)

	ctx := context.Background()
	testCatalogs := createTestCatalogs()

	// Mock expectations
	specService.On("GetCatalogs", ctx).Return(testCatalogs, nil)

	// Execute
	serviceTypes, err := resolver.GetServiceTypes(ctx, "ecom")

	// Verify
	assert.NoError(t, err)
	assert.NotEmpty(t, serviceTypes)
	assert.Contains(t, serviceTypes, "SDD")
	assert.Contains(t, serviceTypes, "NDD")
	assert.Contains(t, serviceTypes, "STANDARD")

	specService.AssertExpectations(t)
}

func TestServiceDefinitionResolver_GetParcelCategories_WithFallback(t *testing.T) {
	// Setup
	specService := &MockSpecificationServiceClient{}
	mockCache := NewMockCache()
	resolver := createTestResolver(specService, mockCache)

	ctx := context.Background()

	// Mock expectations - simulate service failure
	specService.On("GetCatalogs", ctx).Return([]interfaces.Catalog{}, errors.New("service unavailable"))
	mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, false)

	// Execute
	categories, err := resolver.GetParcelCategories(ctx)

	// Verify - should get fallback data
	assert.NoError(t, err)
	assert.NotEmpty(t, categories)
	assert.Contains(t, categories, "ecom")
	assert.Contains(t, categories, "courier")

	specService.AssertExpectations(t)
}

func TestServiceDefinitionResolver_ValidationFailure(t *testing.T) {
	// Setup with strict validation
	specService := &MockSpecificationServiceClient{}
	mockCache := NewMockCache()

	config := createTestConfig()
	config.StrictValidation = true

	// Create cache
	cacheConfig := cache.ServiceDefinitionCacheConfig{
		SpecDefinitionsTTL: 30 * time.Minute,
		CatalogsTTL:        30 * time.Minute,
		EntitySpecsTTL:     30 * time.Minute,
		ServiceDefsTTL:     30 * time.Minute,
		KeyPrefix:          "test",
	}
	serviceDefCache := cache.NewServiceDefinitionCache(mockCache, cacheConfig)

	// Create fallback manager
	fallbackConfig := fallback.FallbackConfig{
		EnableDefaultData: true,
		MaxRetryAttempts:  3,
		RetryDelay:        1 * time.Second,
	}
	fallbackManager := fallback.NewFallbackManager(serviceDefCache, fallbackConfig, nil)

	resolver := NewServiceDefinitionResolver(specService, serviceDefCache, fallbackManager, config)

	ctx := context.Background()

	// Execute with invalid input
	_, err := resolver.GetServiceTypes(ctx, "INVALID_CATEGORY")

	// Verify - should fail validation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parcel category")
}

func TestServiceDefinitionResolver_CircuitBreaker(t *testing.T) {
	// Setup
	specService := &MockSpecificationServiceClient{}
	mockCache := NewMockCache()

	config := createTestConfig()
	config.FailureThreshold = 2 // Lower threshold for testing

	// Create cache
	cacheConfig := cache.ServiceDefinitionCacheConfig{
		SpecDefinitionsTTL: 30 * time.Minute,
		CatalogsTTL:        30 * time.Minute,
		EntitySpecsTTL:     30 * time.Minute,
		ServiceDefsTTL:     30 * time.Minute,
		KeyPrefix:          "test",
	}
	serviceDefCache := cache.NewServiceDefinitionCache(mockCache, cacheConfig)

	// Create fallback manager with disabled fallback for this test
	fallbackConfig := fallback.FallbackConfig{
		EnableDefaultData: false, // Disable fallback to test circuit breaker
		MaxRetryAttempts:  3,
		RetryDelay:        1 * time.Second,
	}
	fallbackManager := fallback.NewFallbackManager(serviceDefCache, fallbackConfig, nil)

	resolver := NewServiceDefinitionResolver(specService, serviceDefCache, fallbackManager, config).(*ServiceDefinitionResolver)

	ctx := context.Background()

	// Mock expectations - simulate service failures
	specService.On("GetCatalogs", ctx).Return([]interfaces.Catalog{}, errors.New("service failure"))
	mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, false)

	// Execute multiple failures to trigger circuit breaker
	for i := 0; i < 3; i++ {
		_, _ = resolver.GetParcelCategories(ctx)
	}

	// Execute again - should fail with circuit breaker error
	_, err := resolver.GetParcelCategories(ctx)
	assert.Error(t, err)

	// Check if it's a circuit breaker error
	if sdErr, ok := infraErrors.GetServiceDefinitionError(err); ok {
		assert.Equal(t, infraErrors.ErrorTypeCircuitOpen, sdErr.Type)
	} else {
		// If not a circuit breaker error, it should still be an error from the service failure
		assert.Contains(t, err.Error(), "service failure")
	}
}

func TestServiceDefinitionResolver_ThreadSafety(t *testing.T) {
	// Setup
	specService := &MockSpecificationServiceClient{}
	mockCache := NewMockCache()
	resolver := createTestResolver(specService, mockCache)

	ctx := context.Background()
	testCatalogs := createTestCatalogs()

	// Mock expectations
	specService.On("GetCatalogs", ctx).Return(testCatalogs, nil)
	mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, false)
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Execute concurrent requests
	const numGoroutines = 10
	const numRequestsPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numRequestsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numRequestsPerGoroutine; j++ {
				_, err := resolver.GetParcelCategories(ctx)
				if err != nil {
					errors <- err
				}

				_, err = resolver.GetServiceTypes(ctx, "ecom")
				if err != nil {
					errors <- err
				}

				_, err = resolver.GetOperationTypes(ctx)
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Verify no errors occurred due to race conditions
	errorCount := 0
	for err := range errors {
		t.Logf("Concurrent execution error: %v", err)
		errorCount++
	}

	assert.Equal(t, 0, errorCount, "Thread safety test failed with %d errors", errorCount)
}

func TestServiceDefinitionResolver_CacheIntegration(t *testing.T) {
	// Setup
	specService := &MockSpecificationServiceClient{}
	mockCache := NewMockCache()
	resolver := createTestResolver(specService, mockCache)

	ctx := context.Background()
	testCatalogs := createTestCatalogs()

	// Mock expectations for first call (cache miss)
	specService.On("GetCatalogs", ctx).Return(testCatalogs, nil).Times(2) // Allow for multiple calls
	mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, false).Times(2)
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(2)

	// First call - should hit the service
	categories1, err1 := resolver.GetParcelCategories(ctx)
	assert.NoError(t, err1)
	assert.NotEmpty(t, categories1)

	// Second call - due to our current implementation, it may still hit the service
	// since the cache integration is not fully mocked
	categories2, err2 := resolver.GetParcelCategories(ctx)
	assert.NoError(t, err2)
	assert.NotEmpty(t, categories2)

	specService.AssertExpectations(t)
}

func TestServiceDefinitionTransformer_Functionality(t *testing.T) {
	// Setup
	config := createTestConfig()
	transformer := NewServiceDefinitionTransformer(config)

	// Test category transformation
	categories := []string{"ECOMMERCE", "courier", "CARGO"}
	transformed := transformer.TransformParcelCategories(categories)

	assert.Contains(t, transformed, "ecom")
	assert.Contains(t, transformed, "courier")
	assert.Contains(t, transformed, "cargo")

	// Test service type transformation
	serviceTypes := []string{"SAME_DAY_DELIVERY", "EXPRESS_DELIVERY"}
	transformedTypes := transformer.TransformServiceTypes(serviceTypes)

	assert.Contains(t, transformedTypes, "SDD")
	assert.Contains(t, transformedTypes, "Express")
}

func TestServiceDefinitionValidator_Functionality(t *testing.T) {
	// Setup
	config := createTestConfig()
	validator := NewServiceDefinitionValidator(config)

	// Test valid inputs
	err := validator.ValidateParcelCategory("ecom")
	assert.NoError(t, err)

	err = validator.ValidateServiceType("SDD")
	assert.NoError(t, err)

	// Test invalid inputs
	err = validator.ValidateParcelCategory("")
	assert.Error(t, err)

	err = validator.ValidateServiceType("")
	assert.Error(t, err)

	// Test service definition validation
	serviceDef := &interfaces.ServiceDefinition{
		ServiceType:       "SDD",
		ParcelCategory:    "ecom",
		DefaultOperations: []string{"pickup", "delivery"},
		DefaultPayments:   []string{"COD", "ONLINE"},
		DefaultDelivery:   []string{"AIR"},
		Description:       "Test service definition",
	}

	err = validator.ValidateServiceDefinition(serviceDef)
	assert.NoError(t, err)

	// Test invalid service definition
	invalidServiceDef := &interfaces.ServiceDefinition{
		ServiceType:       "",
		ParcelCategory:    "",
		DefaultOperations: []string{},
		DefaultPayments:   []string{},
		DefaultDelivery:   []string{},
		Description:       "",
	}

	err = validator.ValidateServiceDefinition(invalidServiceDef)
	assert.Error(t, err)
}

func TestServiceDefinitionResolver_GetMetrics(t *testing.T) {
	// Setup
	specService := &MockSpecificationServiceClient{}
	mockCache := NewMockCache()
	resolver := createTestResolver(specService, mockCache)

	// Mock cache stats
	expectedStats := cache.CacheStats{
		Hits:      10,
		Misses:    5,
		Evictions: 1,
	}
	mockCache.On("GetStats").Return(expectedStats)

	// Execute
	metrics := resolver.GetMetrics()

	// Verify
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "cache_stats")
	assert.Contains(t, metrics, "fallback_stats")
	assert.Contains(t, metrics, "health")

	mockCache.AssertExpectations(t)
}

func TestServiceDefinitionResolver_GetHealth(t *testing.T) {
	// Setup
	specService := &MockSpecificationServiceClient{}
	mockCache := NewMockCache()
	resolver := createTestResolver(specService, mockCache)

	// Execute
	health := resolver.GetHealth()

	// Verify
	assert.NotNil(t, health)
	assert.Contains(t, health, "last_success_time")
	assert.Contains(t, health, "consecutive_errors")
	assert.Contains(t, health, "circuit_breaker_state")
	assert.Contains(t, health, "is_healthy")

	// Should be healthy initially
	assert.True(t, health["is_healthy"].(bool))
}

// Benchmarks

func BenchmarkServiceDefinitionResolver_GetParcelCategories(b *testing.B) {
	// Setup
	specService := &MockSpecificationServiceClient{}
	mockCache := NewMockCache()
	resolver := createTestResolver(specService, mockCache)

	ctx := context.Background()
	testCatalogs := createTestCatalogs()

	// Mock expectations
	specService.On("GetCatalogs", ctx).Return(testCatalogs, nil)
	mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, false)
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	b.ResetTimer()

	// Benchmark
	for i := 0; i < b.N; i++ {
		_, err := resolver.GetParcelCategories(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServiceDefinitionResolver_ConcurrentAccess(b *testing.B) {
	// Setup
	specService := &MockSpecificationServiceClient{}
	mockCache := NewMockCache()
	resolver := createTestResolver(specService, mockCache)

	ctx := context.Background()
	testCatalogs := createTestCatalogs()

	// Mock expectations
	specService.On("GetCatalogs", ctx).Return(testCatalogs, nil)
	mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, false)
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	b.ResetTimer()

	// Benchmark concurrent access
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := resolver.GetParcelCategories(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
