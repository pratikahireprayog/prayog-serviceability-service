package mocks

import (
	"context"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// MockPartnerAdapter implements the common.PartnerAdapter interface for testing
type MockPartnerAdapter struct {
	// Configuration
	PartnerCode     string
	PartnerName     string
	AdapterType     common.AdapterType
	Enabled         bool
	Healthy         bool
	InitializeError error
	ShutdownError   error

	// Response configuration
	ServiceabilityResult *common.PartnerServiceabilityResult
	ServiceabilityError  error

	// Metrics
	Metrics *common.PartnerMetrics

	// Call tracking
	InitializeCalled          bool
	ShutdownCalled            bool
	CheckServiceabilityCalled bool
	IsHealthyCalled           bool
	GetMetricsCalled          bool

	// Last call parameters
	LastContext context.Context
	LastRequest *models.ServiceabilityV2Request
}

// NewMockPartnerAdapter creates a new mock partner adapter
func NewMockPartnerAdapter(partnerCode string) *MockPartnerAdapter {
	return &MockPartnerAdapter{
		PartnerCode: partnerCode,
		PartnerName: "Mock " + partnerCode,
		AdapterType: common.AdapterTypeHTTP,
		Enabled:     true,
		Healthy:     true,
		Metrics: &common.PartnerMetrics{
			PartnerCode:         partnerCode,
			TotalRequests:       0,
			SuccessfulRequests:  0,
			FailedRequests:      0,
			AverageResponseTime: 100 * time.Millisecond,
			HealthStatus:        "healthy",
			ErrorRate:           0.0,
		},
	}
}

// GetPartnerCode returns the partner code
func (m *MockPartnerAdapter) GetPartnerCode() string {
	return m.PartnerCode
}

// GetPartnerName returns the partner name
func (m *MockPartnerAdapter) GetPartnerName() string {
	return m.PartnerName
}

// GetAdapterType returns the adapter type
func (m *MockPartnerAdapter) GetAdapterType() common.AdapterType {
	return m.AdapterType
}

// CheckServiceability mocks the serviceability check
func (m *MockPartnerAdapter) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	m.CheckServiceabilityCalled = true
	m.LastContext = ctx
	m.LastRequest = req

	if m.ServiceabilityError != nil {
		return nil, m.ServiceabilityError
	}

	if m.ServiceabilityResult != nil {
		return m.ServiceabilityResult, nil
	}

	// Default response
	return &common.PartnerServiceabilityResult{
		PartnerID:     partnerInfo.PartnerID,
		PartnerCode:   partnerInfo.PartnerCode,
		
		Services:      []models.ServiceV2{},
		Capabilities:  make(map[string]interface{}),
		ResponseTime:  50 * time.Millisecond,
		Metadata: map[string]interface{}{
			"mock": true,
		},
	}, nil
}

// IsHealthy returns the health status
func (m *MockPartnerAdapter) IsHealthy(ctx context.Context) bool {
	m.IsHealthyCalled = true
	m.LastContext = ctx
	return m.Healthy
}

// GetMetrics returns the metrics
func (m *MockPartnerAdapter) GetMetrics() *common.PartnerMetrics {
	m.GetMetricsCalled = true
	return m.Metrics
}

// Initialize mocks the initialization
func (m *MockPartnerAdapter) Initialize(ctx context.Context) error {
	m.InitializeCalled = true
	m.LastContext = ctx
	return m.InitializeError
}

// Shutdown mocks the shutdown
func (m *MockPartnerAdapter) Shutdown(ctx context.Context) error {
	m.ShutdownCalled = true
	m.LastContext = ctx
	return m.ShutdownError
}

// SetServiceabilityResult sets the result to return from CheckServiceability
func (m *MockPartnerAdapter) SetServiceabilityResult(result *common.PartnerServiceabilityResult) {
	m.ServiceabilityResult = result
}

// SetServiceabilityError sets the error to return from CheckServiceability
func (m *MockPartnerAdapter) SetServiceabilityError(err error) {
	m.ServiceabilityError = err
}

// SetHealthy sets the health status
func (m *MockPartnerAdapter) SetHealthy(healthy bool) {
	m.Healthy = healthy
}

// Reset resets all call tracking
func (m *MockPartnerAdapter) Reset() {
	m.InitializeCalled = false
	m.ShutdownCalled = false
	m.CheckServiceabilityCalled = false
	m.IsHealthyCalled = false
	m.GetMetricsCalled = false
	m.LastContext = nil
	m.LastRequest = nil
}

// CreateServiceableResult creates a serviceable result for testing
func (m *MockPartnerAdapter) CreateServiceableResult() *common.PartnerServiceabilityResult {
	return &common.PartnerServiceabilityResult{
		PartnerCode:   m.PartnerCode,
		
		Services: []models.ServiceV2{
			{
				ServiceCode: "express",
				ServiceName: "Express Delivery",
				TATDays:     1,
				IsCOD:       true,
				Pickup:      true,
				Delivery:    true,
				Insurance:   false,
				ProductTypes: map[string]bool{
					"ecommerce": true,
				},
				DeliveryModes: map[string]bool{
					"standard": true,
				},
				Pricing: &models.ServicePricingV2{
					BaseCost: 100.0,
					Currency: "INR",
				},
			},
		},
		Capabilities: map[string]interface{}{
			"cod_available": true,
			"tracking":      true,
		},
		ResponseTime: 100 * time.Millisecond,
		Metadata: map[string]interface{}{
			"mock": true,
		},
	}
}

// CreateNonServiceableResult creates a non-serviceable result for testing
func (m *MockPartnerAdapter) CreateNonServiceableResult() *common.PartnerServiceabilityResult {
	errorMsg := "Not serviceable in this location"
	return &common.PartnerServiceabilityResult{
		PartnerCode:   m.PartnerCode,
		
		Services:      []models.ServiceV2{},
		Capabilities:  make(map[string]interface{}),
		ErrorMessage:  &errorMsg,
		ResponseTime:  50 * time.Millisecond,
		Metadata: map[string]interface{}{
			"mock": true,
		},
	}
}

// CreateCleanNonServiceableResult creates a non-serviceable result without error messages for testing
func (m *MockPartnerAdapter) CreateCleanNonServiceableResult() *common.PartnerServiceabilityResult {
	return &common.PartnerServiceabilityResult{
		PartnerCode:   m.PartnerCode,
		
		Services:      []models.ServiceV2{},
		Capabilities:  make(map[string]interface{}),
		ErrorMessage:  nil, // No error message - truly non-serviceable without validation errors
		ResponseTime:  50 * time.Millisecond,
		Metadata: map[string]interface{}{
			"mock": true,
		},
	}
}

// CreateErrorResult creates an error result for testing
func (m *MockPartnerAdapter) CreateErrorResult(err error) *common.PartnerServiceabilityResult {
	errorMsg := err.Error()
	return &common.PartnerServiceabilityResult{
		PartnerCode:   m.PartnerCode,
		
		Services:      []models.ServiceV2{},
		Capabilities:  make(map[string]interface{}),
		Error:         err,
		ErrorMessage:  &errorMsg,
		ResponseTime:  25 * time.Millisecond,
		Metadata: map[string]interface{}{
			"mock": true,
		},
	}
}
