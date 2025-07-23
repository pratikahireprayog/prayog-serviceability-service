package mocks

import (
	"context"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
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
	LastContext     context.Context
	LastRequest     *models.ServiceabilityV2Request
	LastPartnerInfo common.PartnerInfo
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
func (m *MockPartnerAdapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	m.CheckServiceabilityCalled = true
	m.LastContext = ctx
	m.LastRequest = request
	m.LastPartnerInfo = partnerInfo

	if m.ServiceabilityError != nil {
		return nil, m.ServiceabilityError
	}

	if m.ServiceabilityResult != nil {
		// Copy the result and update with actual partnerInfo
		result := *m.ServiceabilityResult
		result.PartnerID = partnerInfo.PartnerID
		result.PartnerCode = partnerInfo.PartnerCode
		if partnerInfo.PartnerCode != "" {
			result.PartnerName = "Mock " + partnerInfo.PartnerCode
		}
		return &result, nil
	}

	// Default response
	return m.CreateServiceableResult(), nil
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

// SetServiceabilityResult sets the serviceability result to return
func (m *MockPartnerAdapter) SetServiceabilityResult(result *common.PartnerServiceabilityResult) {
	m.ServiceabilityResult = result
}

// SetServiceabilityError sets the serviceability error to return
func (m *MockPartnerAdapter) SetServiceabilityError(err error) {
	m.ServiceabilityError = err
}

// SetHealthy sets the health status
func (m *MockPartnerAdapter) SetHealthy(healthy bool) {
	m.Healthy = healthy
	if healthy {
		m.Metrics.HealthStatus = "healthy"
	} else {
		m.Metrics.HealthStatus = "unhealthy"
	}
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
	// Create a mock UUID for testing
	partnerID := uuid.New()

	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: m.PartnerCode,
		PartnerName: m.PartnerName,
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
	partnerID := uuid.New()
	errorMsg := "Not serviceable in this location"

	return &common.PartnerServiceabilityResult{
		PartnerID:    &partnerID,
		PartnerCode:  m.PartnerCode,
		PartnerName:  m.PartnerName,
		Services:     []models.ServiceV2{},
		Capabilities: make(map[string]interface{}),
		ErrorMessage: &errorMsg,
		ResponseTime: 50 * time.Millisecond,
		Metadata: map[string]interface{}{
			"mock": true,
		},
	}
}

// CreateDHLServiceableResult creates a DHL-specific serviceable result with flattened capabilities
func (m *MockPartnerAdapter) CreateDHLServiceableResult() *common.PartnerServiceabilityResult {
	partnerID := uuid.New()

	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: m.PartnerCode,
		PartnerName: m.PartnerName,
		Services:    []models.ServiceV2{}, // DHL doesn't return detailed services in serviceability check
		// Flattened DHL capabilities structure
		Capabilities: map[string]interface{}{
			// Pickup capabilities (flattened)
			"next_business_day":                          true,
			"local_cutoff_date_and_time":                 "2024-01-15T12:00:00GMT+05:30",
			"pickup_earliest":                            "2024-01-15T09:00:00GMT+05:30",
			"pickup_latest":                              "2024-01-15T17:00:00GMT+05:30",
			"pickup_cutoff_same_day_outbound_processing": "2024-01-15T12:00:00GMT+05:30",
			"origin_service_area_code":                   "DEL",
			"origin_facility_area_code":                  "DEL01",
			"pickup_additional_days":                     0,
			"pickup_day_of_week":                         2,
			// Delivery capabilities (flattened)
			"delivery_type_code":               "QDDC",
			"estimated_delivery_date_and_time": "2024-01-17T10:00:00GMT+05:30",
			"destination_service_area_code":    "BOM",
			"destination_facility_area_code":   "BOM01",
			"delivery_additional_days":         0,
			"delivery_day_of_week":             4,
			"total_transit_days":               2,
		},
		ResponseTime: 150 * time.Millisecond,
		Metadata: map[string]interface{}{
			"mock":          true,
			"partner":       "DHL",
			"product_count": 1,
			"flow":          "international",
		},
	}
}

// CreateDHLNonServiceableResult creates a DHL-specific non-serviceable result
func (m *MockPartnerAdapter) CreateDHLNonServiceableResult() *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	errorMsg := "DHL validation failed: destination postal code is required for DHL shipments"

	return &common.PartnerServiceabilityResult{
		PartnerID:    &partnerID,
		PartnerCode:  m.PartnerCode,
		PartnerName:  m.PartnerName,
		Services:     []models.ServiceV2{},
		Capabilities: make(map[string]interface{}),
		ErrorMessage: &errorMsg,
		ResponseTime: 50 * time.Millisecond,
		Metadata: map[string]interface{}{
			"mock":   true,
			"reason": "DHL validation failed",
			"step":   "validation",
		},
	}
}

// CreateCleanNonServiceableResult creates a non-serviceable result without error messages for testing
func (m *MockPartnerAdapter) CreateCleanNonServiceableResult() *common.PartnerServiceabilityResult {
	return &common.PartnerServiceabilityResult{
		PartnerCode: m.PartnerCode,

		Services:     []models.ServiceV2{},
		Capabilities: make(map[string]interface{}),
		ErrorMessage: nil, // No error message - truly non-serviceable without validation errors
		ResponseTime: 50 * time.Millisecond,
		Metadata: map[string]interface{}{
			"mock": true,
		},
	}
}

// CreateErrorResult creates an error result for testing
func (m *MockPartnerAdapter) CreateErrorResult(err error) *common.PartnerServiceabilityResult {
	errorMsg := err.Error()
	return &common.PartnerServiceabilityResult{
		PartnerCode: m.PartnerCode,

		Services:     []models.ServiceV2{},
		Capabilities: make(map[string]interface{}),
		Error:        err,
		ErrorMessage: &errorMsg,
		ResponseTime: 25 * time.Millisecond,
		Metadata: map[string]interface{}{
			"mock": true,
		},
	}
}
