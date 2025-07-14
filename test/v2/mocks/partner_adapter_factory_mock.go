package mocks

import (
	"errors"
	"fmt"

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

// SetupDefaultAdapters sets up common adapters for testing
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

// SetupErrorAdapters sets up adapters that return errors for testing
func (m *MockPartnerAdapterFactory) SetupErrorAdapters() {
	for _, partnerCode := range m.SupportedPartners {
		adapter := NewMockPartnerAdapter(partnerCode)
		adapter.SetServiceabilityError(fmt.Errorf("partner %s temporarily unavailable", partnerCode))
		m.SetAdapter(partnerCode, adapter)
	}
}

// SetupNonServiceableAdapters sets up adapters that return non-serviceable results
func (m *MockPartnerAdapterFactory) SetupNonServiceableAdapters() {
	for _, partnerCode := range m.SupportedPartners {
		adapter := NewMockPartnerAdapter(partnerCode)
		adapter.SetServiceabilityResult(adapter.CreateNonServiceableResult())
		m.SetAdapter(partnerCode, adapter)
	}
}

// SetupServiceableAdapters sets up adapters that return serviceable results
func (m *MockPartnerAdapterFactory) SetupServiceableAdapters() {
	for _, partnerCode := range m.SupportedPartners {
		adapter := NewMockPartnerAdapter(partnerCode)
		adapter.SetServiceabilityResult(adapter.CreateServiceableResult())
		m.SetAdapter(partnerCode, adapter)
	}
}

// Compile time check to ensure MockPartnerAdapterFactory implements the interface
var _ factory.PartnerAdapterFactory = (*MockPartnerAdapterFactory)(nil)
