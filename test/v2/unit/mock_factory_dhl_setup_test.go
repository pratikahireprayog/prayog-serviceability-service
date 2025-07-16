package unit

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"
)

func TestMockFactoryDHLSetup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		setupServiceable     bool
		setupError           error
		expectedAdapterCount int
		expectedPartners     []string
	}{
		{
			name:                 "Setup DHL serviceable adapter",
			setupServiceable:     true,
			setupError:           nil,
			expectedAdapterCount: 1,
			expectedPartners:     []string{"dhl"},
		},
		{
			name:                 "Setup DHL non-serviceable adapter",
			setupServiceable:     false,
			setupError:           nil,
			expectedAdapterCount: 1,
			expectedPartners:     []string{"dhl"},
		},
		{
			name:                 "Setup DHL adapter with error",
			setupServiceable:     true,
			setupError:           fmt.Errorf("DHL authentication failed"),
			expectedAdapterCount: 0,
			expectedPartners:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock factory
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			if tt.setupError != nil {
				// Setup factory to return error
				mockFactory.CreateAdapterError = tt.setupError
			} else if tt.setupServiceable {
				// Setup serviceable DHL adapter
				dhlAdapter := setupDHLServiceableAdapter()
				mockFactory.SetAdapter("dhl", dhlAdapter)
			} else {
				// Setup non-serviceable DHL adapter
				dhlAdapter := setupDHLNonServiceableAdapter()
				mockFactory.SetAdapter("dhl", dhlAdapter)
			}

			// Test adapter creation
			if tt.setupError != nil {
				_, err := mockFactory.CreateAdapter("dhl")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "DHL authentication failed")
			} else {
				adapter, err := mockFactory.CreateAdapter("dhl")
				assert.NoError(t, err)
				assert.NotNil(t, adapter)
			}

			// Verify supported partners
			supportedPartners := mockFactory.GetSupportedPartners()
			assert.Contains(t, supportedPartners, "dhl")
		})
	}
}

func TestMockFactoryDHLConfigurationMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		configMethod   string
		expectedResult bool
		expectedError  bool
	}{
		{
			name:           "Create DHL serviceable result",
			configMethod:   "CreateDHLServiceableResult",
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "Create DHL non-serviceable result",
			configMethod:   "CreateDHLNonServiceableResult",
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:           "Create DHL validation error result",
			configMethod:   "CreateDHLValidationErrorResult",
			expectedResult: false,
			expectedError:  true,
		},
		{
			name:           "Create DHL API error result",
			configMethod:   "CreateDHLAPIErrorResult",
			expectedResult: false,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			partnerID := uuid.New()

			var result *common.PartnerServiceabilityResult

			switch tt.configMethod {
			case "CreateDHLServiceableResult":
				result = createDHLServiceableResult(&partnerID, "DHL Express")
			case "CreateDHLNonServiceableResult":
				result = createDHLNonServiceableResult(&partnerID, "DHL Express")
			case "CreateDHLValidationErrorResult":
				result = createDHLValidationErrorResult(&partnerID, "DHL Express", "Invalid postal code")
			case "CreateDHLAPIErrorResult":
				result = createDHLAPIErrorResult(&partnerID, "DHL Express", "API rate limit exceeded")
			}

			assert.NotNil(t, result)
			assert.Equal(t, &partnerID, result.PartnerID)
			assert.Equal(t, "dhl", result.PartnerCode)
			assert.Equal(t, "DHL Express", result.PartnerName)

			if tt.expectedResult {
				assert.NotEmpty(t, result.Services)
			} else {
				if tt.expectedError {
					assert.Error(t, result.Error)
				}
			}
		})
	}
}

func TestMockFactoryDHLScenarioSetup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		scenario             string
		expectedServices     int
		expectedError        bool
		expectedCapabilities map[string]interface{}
	}{
		{
			name:             "DHL international scenario",
			scenario:         "international",
			expectedServices: 2,
			expectedError:    false,
			expectedCapabilities: map[string]interface{}{
				"next_business_day":          false,
				"total_transit_days":         5,
				"pickup_earliest":            "09:00",
				"pickup_latest":              "17:00",
				"local_cutoff_date_and_time": "2024-01-15T15:00:00Z",
				"supports_international":     true,
			},
		},
		{
			name:             "DHL domestic scenario",
			scenario:         "domestic",
			expectedServices: 1,
			expectedError:    false,
			expectedCapabilities: map[string]interface{}{
				"next_business_day":          true,
				"total_transit_days":         1,
				"pickup_earliest":            "08:00",
				"pickup_latest":              "18:00",
				"local_cutoff_date_and_time": "2024-01-15T16:00:00Z",
				"supports_international":     false,
			},
		},
		{
			name:                 "DHL validation error scenario",
			scenario:             "validation_error",
			expectedServices:     0,
			expectedError:        true,
			expectedCapabilities: nil,
		},
		{
			name:                 "DHL timeout scenario",
			scenario:             "timeout",
			expectedServices:     0,
			expectedError:        true,
			expectedCapabilities: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			partnerID := uuid.New()

			var result *common.PartnerServiceabilityResult

			switch tt.scenario {
			case "international":
				result = createDHLInternationalResult(&partnerID, "DHL Express")
			case "domestic":
				result = createDHLDomesticResult(&partnerID, "DHL Express")
			case "validation_error":
				result = createDHLValidationErrorResult(&partnerID, "DHL Express", "DHL validation failed: missing postal code")
			case "timeout":
				result = createDHLTimeoutErrorResult(&partnerID, "DHL Express")
			}

			assert.NotNil(t, result)
			assert.Equal(t, &partnerID, result.PartnerID)
			assert.Equal(t, "dhl", result.PartnerCode)
			assert.Equal(t, "DHL Express", result.PartnerName)

			if tt.expectedError {
				assert.Error(t, result.Error)
				assert.NotNil(t, result.ErrorMessage)
			} else {
				assert.NoError(t, result.Error)
				assert.Len(t, result.Services, tt.expectedServices)

				if tt.expectedCapabilities != nil {
					for key, expectedValue := range tt.expectedCapabilities {
						actualValue, exists := result.Capabilities[key]
						assert.True(t, exists, "Capability %s should exist", key)
						assert.Equal(t, expectedValue, actualValue, "Capability %s should match", key)
					}
				}
			}
		})
	}
}

func TestMockFactoryDHLAdapterTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		adapterType          common.AdapterType
		expectedHTTPFeatures bool
		expectedDBFeatures   bool
	}{
		{
			name:                 "DHL HTTP adapter",
			adapterType:          common.AdapterTypeHTTP,
			expectedHTTPFeatures: true,
			expectedDBFeatures:   false,
		},
		{
			name:                 "DHL database adapter",
			adapterType:          common.AdapterTypeDatabase,
			expectedHTTPFeatures: false,
			expectedDBFeatures:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock factory
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")

			// Configure adapter type
			dhlAdapter.AdapterType = tt.adapterType
			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Test adapter creation
			adapter, err := mockFactory.CreateAdapter("dhl")
			assert.NoError(t, err)
			assert.NotNil(t, adapter)

			mockAdapter := adapter.(*mocks.MockPartnerAdapter)
			assert.Equal(t, tt.adapterType, mockAdapter.AdapterType)

			if tt.expectedHTTPFeatures {
				assert.Equal(t, common.AdapterTypeHTTP, mockAdapter.AdapterType)
			}

			if tt.expectedDBFeatures {
				assert.Equal(t, common.AdapterTypeDatabase, mockAdapter.AdapterType)
			}
		})
	}
}

func TestMockFactoryDHLErrorScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		errorType    string
		errorMessage string
		expectedCode string
	}{
		{
			name:         "DHL authentication error",
			errorType:    "authentication",
			errorMessage: "DHL authentication failed: invalid credentials",
			expectedCode: "DHL_AUTH_ERROR",
		},
		{
			name:         "DHL validation error",
			errorType:    "validation",
			errorMessage: "DHL validation failed: missing required fields",
			expectedCode: "DHL_VALIDATION_ERROR",
		},
		{
			name:         "DHL API error",
			errorType:    "api",
			errorMessage: "DHL API error: service unavailable",
			expectedCode: "DHL_API_ERROR",
		},
		{
			name:         "DHL network error",
			errorType:    "network",
			errorMessage: "DHL network error: connection timeout",
			expectedCode: "DHL_NETWORK_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock factory with specific error
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockFactory.CreateAdapterError = fmt.Errorf("%s", tt.errorMessage)

			// Test error handling
			_, err := mockFactory.CreateAdapter("dhl")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMessage)

			// Test error categorization
			errorCode := categorizeDHLError(err)
			assert.Equal(t, tt.expectedCode, errorCode)
		})
	}
}

// Helper functions for DHL mock factory setup

func setupDHLServiceableAdapter() *mocks.MockPartnerAdapter {
	partnerID := uuid.New()
	adapter := mocks.NewMockPartnerAdapter("dhl")
	adapter.SetServiceabilityResult(createDHLServiceableResult(&partnerID, "DHL Express"))
	return adapter
}

func setupDHLNonServiceableAdapter() *mocks.MockPartnerAdapter {
	partnerID := uuid.New()
	adapter := mocks.NewMockPartnerAdapter("dhl")
	adapter.SetServiceabilityResult(createDHLNonServiceableResult(&partnerID, "DHL Express"))
	return adapter
}

func createDHLServiceableResult(partnerID *uuid.UUID, partnerName string) *common.PartnerServiceabilityResult {
	return &common.PartnerServiceabilityResult{
		PartnerID:   partnerID,
		PartnerCode: "dhl",
		PartnerName: partnerName,
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS",
				ServiceName: "DHL Express Worldwide",
			},
		},
		Capabilities: map[string]interface{}{
			"next_business_day":          true,
			"total_transit_days":         3,
			"pickup_earliest":            "09:00",
			"pickup_latest":              "17:00",
			"local_cutoff_date_and_time": "2024-01-15T15:00:00Z",
		},
	}
}

func createDHLNonServiceableResult(partnerID *uuid.UUID, partnerName string) *common.PartnerServiceabilityResult {
	return &common.PartnerServiceabilityResult{
		PartnerID:    partnerID,
		PartnerCode:  "dhl",
		PartnerName:  partnerName,
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("DHL service not available for this route"),
		ErrorMessage: stringPtr("Service not available for this postal code combination"),
	}
}

func createDHLValidationErrorResult(partnerID *uuid.UUID, partnerName string, errorMsg string) *common.PartnerServiceabilityResult {
	return &common.PartnerServiceabilityResult{
		PartnerID:    partnerID,
		PartnerCode:  "dhl",
		PartnerName:  partnerName,
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("DHL validation failed: %s", errorMsg),
		ErrorMessage: stringPtr(fmt.Sprintf("DHL validation failed: %s", errorMsg)),
	}
}

func createDHLAPIErrorResult(partnerID *uuid.UUID, partnerName string, errorMsg string) *common.PartnerServiceabilityResult {
	return &common.PartnerServiceabilityResult{
		PartnerID:    partnerID,
		PartnerCode:  "dhl",
		PartnerName:  partnerName,
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("DHL API error: %s", errorMsg),
		ErrorMessage: stringPtr(fmt.Sprintf("DHL API error: %s", errorMsg)),
	}
}

func createDHLInternationalResult(partnerID *uuid.UUID, partnerName string) *common.PartnerServiceabilityResult {
	return &common.PartnerServiceabilityResult{
		PartnerID:   partnerID,
		PartnerCode: "dhl",
		PartnerName: partnerName,
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS_WORLDWIDE",
				ServiceName: "DHL Express Worldwide",
			},
			{
				ServiceCode: "EXPRESS_ENVELOPE",
				ServiceName: "DHL Express Envelope",
			},
		},
		Capabilities: map[string]interface{}{
			"next_business_day":          false,
			"total_transit_days":         5,
			"pickup_earliest":            "09:00",
			"pickup_latest":              "17:00",
			"local_cutoff_date_and_time": "2024-01-15T15:00:00Z",
			"supports_international":     true,
		},
	}
}

func createDHLDomesticResult(partnerID *uuid.UUID, partnerName string) *common.PartnerServiceabilityResult {
	return &common.PartnerServiceabilityResult{
		PartnerID:   partnerID,
		PartnerCode: "dhl",
		PartnerName: partnerName,
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS_DOMESTIC",
				ServiceName: "DHL Express Domestic",
			},
		},
		Capabilities: map[string]interface{}{
			"next_business_day":          true,
			"total_transit_days":         1,
			"pickup_earliest":            "08:00",
			"pickup_latest":              "18:00",
			"local_cutoff_date_and_time": "2024-01-15T16:00:00Z",
			"supports_international":     false,
		},
	}
}

func createDHLTimeoutErrorResult(partnerID *uuid.UUID, partnerName string) *common.PartnerServiceabilityResult {
	return &common.PartnerServiceabilityResult{
		PartnerID:    partnerID,
		PartnerCode:  "dhl",
		PartnerName:  partnerName,
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("DHL network error: connection timeout"),
		ErrorMessage: stringPtr("DHL network error: connection timeout"),
	}
}

func categorizeDHLError(err error) string {
	errMsg := err.Error()

	switch {
	case contains(errMsg, "authentication"):
		return "DHL_AUTH_ERROR"
	case contains(errMsg, "validation"):
		return "DHL_VALIDATION_ERROR"
	case contains(errMsg, "API"):
		return "DHL_API_ERROR"
	case contains(errMsg, "network") || contains(errMsg, "timeout"):
		return "DHL_NETWORK_ERROR"
	default:
		return "DHL_UNKNOWN_ERROR"
	}
}

// stringPtr helper function is defined in request_validation_test.go
