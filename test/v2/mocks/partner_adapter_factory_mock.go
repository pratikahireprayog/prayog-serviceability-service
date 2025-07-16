package mocks

import (
	"errors"
	"fmt"
	"time"

	models "prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/services/v2/partners/factory"
)

// MockPartnerAdapterFactory implements the factory.PartnerAdapterFactory interface for testing
type MockPartnerAdapterFactory struct {
	// Configuration
	Adapters          map[string]common.PartnerAdapter
	SupportedPartners []string

	// Error configuration
	CreateAdapterError    error
	ValidateConfigError   error
	GetAdapterConfigError error

	// Call tracking
	CreateAdapterCalled          bool
	IsPartnerSupportedCalled     bool
	GetSupportedPartnersCalled   bool
	GetAdapterCalled             bool
	GetAllAdaptersCalled         bool
	RefreshAdaptersCalled        bool
	GetAdapterConfigCalled       bool
	ValidateConfigurationsCalled bool

	// Last call parameters
	LastPartnerCode string
}

// NewMockPartnerAdapterFactory creates a new mock partner adapter factory
func NewMockPartnerAdapterFactory() *MockPartnerAdapterFactory {
	return &MockPartnerAdapterFactory{
		Adapters:          make(map[string]common.PartnerAdapter),
		SupportedPartners: []string{"dhl", "shipyaari", "smile_ecom"},
	}
}

// CreateAdapter creates a mock adapter for the given partner code
func (m *MockPartnerAdapterFactory) CreateAdapter(partnerCode string) (common.PartnerAdapter, error) {
	m.CreateAdapterCalled = true
	m.LastPartnerCode = partnerCode

	if m.CreateAdapterError != nil {
		return nil, m.CreateAdapterError
	}

	if adapter, exists := m.Adapters[partnerCode]; exists {
		return adapter, nil
	}

	// Create a default mock adapter
	mockAdapter := NewMockPartnerAdapter(partnerCode)
	m.Adapters[partnerCode] = mockAdapter
	return mockAdapter, nil
}

// IsPartnerSupported checks if a partner is supported
func (m *MockPartnerAdapterFactory) IsPartnerSupported(partnerCode string) bool {
	m.IsPartnerSupportedCalled = true
	m.LastPartnerCode = partnerCode

	for _, supported := range m.SupportedPartners {
		if supported == partnerCode {
			return true
		}
	}

	return false
}

// GetSupportedPartners returns the list of supported partners
func (m *MockPartnerAdapterFactory) GetSupportedPartners() []string {
	m.GetSupportedPartnersCalled = true
	return m.SupportedPartners
}

// GetAdapter returns an adapter for the given partner code
func (m *MockPartnerAdapterFactory) GetAdapter(partnerCode string) (common.PartnerAdapter, bool) {
	m.GetAdapterCalled = true
	m.LastPartnerCode = partnerCode

	if adapter, exists := m.Adapters[partnerCode]; exists {
		return adapter, true
	}

	// Create a default mock adapter
	mockAdapter := NewMockPartnerAdapter(partnerCode)
	m.Adapters[partnerCode] = mockAdapter
	return mockAdapter, true
}

// GetAllAdapters returns all adapters
func (m *MockPartnerAdapterFactory) GetAllAdapters() map[string]common.PartnerAdapter {
	m.GetAllAdaptersCalled = true
	return m.Adapters
}

// RefreshAdapters refreshes all adapters
func (m *MockPartnerAdapterFactory) RefreshAdapters() {
	m.RefreshAdaptersCalled = true
	// Reset adapters for testing
	m.Adapters = make(map[string]common.PartnerAdapter)
}

// GetAdapterConfig returns configuration for a specific adapter
func (m *MockPartnerAdapterFactory) GetAdapterConfig(partnerCode string) (interface{}, error) {
	m.GetAdapterConfigCalled = true
	m.LastPartnerCode = partnerCode

	if m.GetAdapterConfigError != nil {
		return nil, m.GetAdapterConfigError
	}

	// Return mock configuration
	return map[string]interface{}{
		"partner_code": partnerCode,
		"enabled":      true,
		"mock":         true,
	}, nil
}

// ValidateConfigurations validates all adapter configurations
func (m *MockPartnerAdapterFactory) ValidateConfigurations() error {
	m.ValidateConfigurationsCalled = true
	return m.ValidateConfigError
}

// SetAdapter sets a specific adapter for a partner code
func (m *MockPartnerAdapterFactory) SetAdapter(partnerCode string, adapter common.PartnerAdapter) {
	m.Adapters[partnerCode] = adapter
}

// SetSupportedPartners sets the list of supported partners
func (m *MockPartnerAdapterFactory) SetSupportedPartners(partners []string) {
	m.SupportedPartners = partners
}

// SetCreateAdapterError sets the error to return from CreateAdapter
func (m *MockPartnerAdapterFactory) SetCreateAdapterError(err error) {
	m.CreateAdapterError = err
}

// SetValidateConfigError sets the error to return from ValidateConfigurations
func (m *MockPartnerAdapterFactory) SetValidateConfigError(err error) {
	m.ValidateConfigError = err
}

// SetGetAdapterConfigError sets the error to return from GetAdapterConfig
func (m *MockPartnerAdapterFactory) SetGetAdapterConfigError(err error) {
	m.GetAdapterConfigError = err
}

// Reset resets all call tracking
func (m *MockPartnerAdapterFactory) Reset() {
	m.CreateAdapterCalled = false
	m.IsPartnerSupportedCalled = false
	m.GetSupportedPartnersCalled = false
	m.GetAdapterCalled = false
	m.GetAllAdaptersCalled = false
	m.RefreshAdaptersCalled = false
	m.GetAdapterConfigCalled = false
	m.ValidateConfigurationsCalled = false
	m.LastPartnerCode = ""
}

// SetupDefaultAdapters sets up default adapters for testing
func (m *MockPartnerAdapterFactory) SetupDefaultAdapters() {
	// DHL adapter (serviceable)
	dhlAdapter := NewMockPartnerAdapter("dhl")
	dhlAdapter.SetServiceabilityResult(dhlAdapter.CreateServiceableResult())
	m.SetAdapter("dhl", dhlAdapter)

	// Shipyaari adapter (serviceable)
	shipyaariAdapter := NewMockPartnerAdapter("shipyaari")
	shipyaariAdapter.SetServiceabilityResult(shipyaariAdapter.CreateServiceableResult())
	m.SetAdapter("shipyaari", shipyaariAdapter)

	// Smile Ecom adapter (not serviceable)
	smileEcomAdapter := NewMockPartnerAdapter("smile_ecom")
	smileEcomAdapter.SetServiceabilityResult(smileEcomAdapter.CreateNonServiceableResult())
	m.SetAdapter("smile_ecom", smileEcomAdapter)

	// Smile Courier adapter (error)
	smileCourierAdapter := NewMockPartnerAdapter("smile_courier")
	smileCourierAdapter.SetServiceabilityError(errors.New("partner temporarily unavailable"))
	m.SetAdapter("smile_courier", smileCourierAdapter)
}

// SetupDHLAdapters sets up DHL-specific adapters for comprehensive testing
func (m *MockPartnerAdapterFactory) SetupDHLAdapters() {
	// DHL serviceable with flattened capabilities
	dhlAdapter := NewMockPartnerAdapter("dhl")
	dhlAdapter.SetServiceabilityResult(dhlAdapter.CreateDHLServiceableResult())
	m.SetAdapter("dhl", dhlAdapter)
}

// SetupDHLValidationErrorAdapters sets up DHL adapters that return validation errors
func (m *MockPartnerAdapterFactory) SetupDHLValidationErrorAdapters() {
	dhlAdapter := NewMockPartnerAdapter("dhl")
	dhlAdapter.SetServiceabilityResult(dhlAdapter.CreateDHLNonServiceableResult())
	m.SetAdapter("dhl", dhlAdapter)
}

// SetupDHLAuthenticationErrorAdapters sets up DHL adapters that return authentication errors
func (m *MockPartnerAdapterFactory) SetupDHLAuthenticationErrorAdapters() {
	dhlAdapter := NewMockPartnerAdapter("dhl")
	dhlAdapter.SetServiceabilityError(fmt.Errorf("DHL authentication failed: invalid credentials"))
	m.SetAdapter("dhl", dhlAdapter)
}

// SetupDHLHubLookupErrorAdapters sets up DHL adapters that return hub lookup errors
func (m *MockPartnerAdapterFactory) SetupDHLHubLookupErrorAdapters() {
	dhlAdapter := NewMockPartnerAdapter("dhl")
	hubLookupError := &common.PartnerServiceabilityResult{
		PartnerID:    nil, // Will be set by adapter
		PartnerCode:  "dhl",
		PartnerName:  "DHL",
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		ErrorMessage: &[]string{"Hub location not found: failed to find nearest hub for postal code 999999"}[0],
		ResponseTime: 75 * time.Millisecond,
		Metadata: map[string]interface{}{
			"reason":         "Hub location not found",
			"source_pincode": "999999",
			"step":           "hub_lookup",
		},
	}
	dhlAdapter.SetServiceabilityResult(hubLookupError)
	m.SetAdapter("dhl", dhlAdapter)
}

// SetupDHLCountryCodeErrorAdapters sets up DHL adapters that return country code resolution errors
func (m *MockPartnerAdapterFactory) SetupDHLCountryCodeErrorAdapters() {
	dhlAdapter := NewMockPartnerAdapter("dhl")
	countryCodeError := &common.PartnerServiceabilityResult{
		PartnerID:    nil, // Will be set by adapter
		PartnerCode:  "dhl",
		PartnerName:  "DHL",
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		ErrorMessage: &[]string{"Country code resolution failed: failed to get source country code for postal code INVALID"}[0],
		ResponseTime: 50 * time.Millisecond,
		Metadata: map[string]interface{}{
			"reason":              "Country code resolution failed",
			"source_pincode":      "INVALID",
			"destination_pincode": "110001",
			"step":                "country_code_resolution",
		},
	}
	dhlAdapter.SetServiceabilityResult(countryCodeError)
	m.SetAdapter("dhl", dhlAdapter)
}

// SetupDHLAPIErrorAdapters sets up DHL adapters that return API call errors
func (m *MockPartnerAdapterFactory) SetupDHLAPIErrorAdapters() {
	dhlAdapter := NewMockPartnerAdapter("dhl")
	apiError := &common.PartnerServiceabilityResult{
		PartnerID:    nil, // Will be set by adapter
		PartnerCode:  "dhl",
		PartnerName:  "DHL",
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		ErrorMessage: &[]string{"DHL API call failed: HTTP 500 Internal Server Error"}[0],
		ResponseTime: 200 * time.Millisecond,
		Metadata: map[string]interface{}{
			"reason":                   "DHL API call failed",
			"source_country_code":      "IN",
			"destination_country_code": "US",
			"step":                     "dhl_api_call",
		},
	}
	dhlAdapter.SetServiceabilityResult(apiError)
	m.SetAdapter("dhl", dhlAdapter)
}

// SetupDHLInternationalFlowAdapters sets up DHL adapters for international flow testing
func (m *MockPartnerAdapterFactory) SetupDHLInternationalFlowAdapters() {
	dhlAdapter := NewMockPartnerAdapter("dhl")

	// Create international serviceability result with detailed metadata
	internationalResult := &common.PartnerServiceabilityResult{
		PartnerID:   nil, // Will be set by adapter
		PartnerCode: "dhl",
		PartnerName: "DHL",
		Services:    []models.ServiceV2{}, // DHL doesn't return services in serviceability check
		Capabilities: map[string]interface{}{
			// Flattened pickup capabilities
			"next_business_day":                          true,
			"local_cutoff_date_and_time":                 "2024-01-15T15:00:00GMT+05:30",
			"pickup_earliest":                            "2024-01-15T09:00:00GMT+05:30",
			"pickup_latest":                              "2024-01-15T18:00:00GMT+05:30",
			"pickup_cutoff_same_day_outbound_processing": "2024-01-15T15:00:00GMT+05:30",
			"origin_service_area_code":                   "BLR",
			"origin_facility_area_code":                  "BLR01",
			"pickup_additional_days":                     0,
			"pickup_day_of_week":                         1,
			// Flattened delivery capabilities
			"delivery_type_code":               "QDDF",
			"estimated_delivery_date_and_time": "2024-01-18T14:00:00GMT-05:00",
			"destination_service_area_code":    "NYC",
			"destination_facility_area_code":   "NYC01",
			"delivery_additional_days":         0,
			"delivery_day_of_week":             4,
			"total_transit_days":               3,
		},
		ResponseTime: 250 * time.Millisecond,
		Metadata: map[string]interface{}{
			"reason":                   "DHL serviceability check completed",
			"product_count":            1,
			"exchange_rates":           2,
			"flow":                     "international",
			"source_country_code":      "IN",
			"destination_country_code": "US",
			"hub_postal_code":          "560001",
			"international_hub_code":   "BLR01",
		},
	}

	dhlAdapter.SetServiceabilityResult(internationalResult)
	m.SetAdapter("dhl", dhlAdapter)
}

// SetupMixedCapabilitiesAdapters sets up adapters with different capability structures for testing response transformation
func (m *MockPartnerAdapterFactory) SetupMixedCapabilitiesAdapters() {
	// DHL with flattened capabilities
	dhlAdapter := NewMockPartnerAdapter("dhl")
	dhlAdapter.SetServiceabilityResult(dhlAdapter.CreateDHLServiceableResult())
	m.SetAdapter("dhl", dhlAdapter)

	// Traditional partner with nested capabilities
	shipyaariAdapter := NewMockPartnerAdapter("shipyaari")
	traditionalResult := shipyaariAdapter.CreateServiceableResult()
	traditionalResult.Capabilities = map[string]interface{}{
		"pickup": map[string]interface{}{
			"available": true,
			"cutoff":    "18:00",
		},
		"delivery": map[string]interface{}{
			"same_day":   true,
			"next_day":   true,
			"hyperlocal": true,
		},
		"cod_available": true,
		"tracking":      true,
	}
	shipyaariAdapter.SetServiceabilityResult(traditionalResult)
	m.SetAdapter("shipyaari", shipyaariAdapter)
}

// Compile time check to ensure MockPartnerAdapterFactory implements the interface
var _ factory.PartnerAdapterFactory = (*MockPartnerAdapterFactory)(nil)
