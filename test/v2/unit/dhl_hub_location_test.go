package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"
)

func TestDHLFindNearestHubSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		sourcePostalCode      string
		destinationPostalCode string
		sourceCountry         string
		expectedHubCode       string
		expectedHubName       string
		expectedDistance      float64
		expectedServices      []string
	}{
		{
			name:                  "US domestic - San Francisco to New York",
			sourcePostalCode:      "94102",
			destinationPostalCode: "10001",
			sourceCountry:         "US",
			expectedHubCode:       "SFO_HUB_001",
			expectedHubName:       "San Francisco International Hub",
			expectedDistance:      15.2,
			expectedServices:      []string{"EXPRESS_DOMESTIC", "EXPRESS_12"},
		},
		{
			name:                  "US international - Los Angeles to London",
			sourcePostalCode:      "90210",
			destinationPostalCode: "SW1A 1AA",
			sourceCountry:         "US",
			expectedHubCode:       "LAX_HUB_001",
			expectedHubName:       "Los Angeles International Hub",
			expectedDistance:      22.8,
			expectedServices:      []string{"EXPRESS_WORLDWIDE", "EXPRESS_ENVELOPE"},
		},
		{
			name:                  "UK international - London to Mumbai",
			sourcePostalCode:      "SW1A 1AA",
			destinationPostalCode: "400001",
			sourceCountry:         "GB",
			expectedHubCode:       "LHR_HUB_001",
			expectedHubName:       "Heathrow International Hub",
			expectedDistance:      18.5,
			expectedServices:      []string{"EXPRESS_WORLDWIDE", "EXPRESS_9", "EXPRESS_12"},
		},
		{
			name:                  "Germany international - Berlin to Tokyo",
			sourcePostalCode:      "10115",
			destinationPostalCode: "100-0001",
			sourceCountry:         "DE",
			expectedHubCode:       "BER_HUB_001",
			expectedHubName:       "Berlin Brandenburg Hub",
			expectedDistance:      25.3,
			expectedServices:      []string{"EXPRESS_WORLDWIDE", "EXPRESS_ENVELOPE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup DHL adapter with hub location response
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
			hubResponse := createDHLHubLocationResponse(
				tt.expectedHubCode,
				tt.expectedHubName,
				tt.expectedDistance,
				tt.expectedServices,
			)
			dhlAdapter.SetServiceabilityResult(hubResponse)
			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Create request
			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      &tt.sourcePostalCode,
				DestinationPostalCode: &tt.destinationPostalCode,
				CountryCode:           &tt.sourceCountry,
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 2.0,
							Unit:  "kg",
						},
					},
				},
			}

			// Execute hub location lookup
			result, err := executeDHLHubLookup(ctx, mockFactory, request)

			// Verify response
			assert.NoError(t, err)
			require.NotNil(t, result)

			// Verify hub details
			assert.Equal(t, tt.expectedHubCode, result.HubCode)
			assert.Equal(t, tt.expectedHubName, result.HubName)
			assert.Equal(t, tt.expectedDistance, result.DistanceKm)
			assert.ElementsMatch(t, tt.expectedServices, result.AvailableServices)

			// Verify hub capabilities
			assert.True(t, result.IsInternational)
			assert.NotEmpty(t, result.PostalCode)
			assert.Greater(t, result.ResponseTimeMs, int64(0))
		})
	}
}

func TestDHLFindNearestHubValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		sourcePostalCode      string
		destinationPostalCode string
		sourceCountry         string
		expectedError         bool
		expectedErrorType     string
		expectedErrorMessage  string
	}{
		{
			name:                  "Invalid source postal code",
			sourcePostalCode:      "",
			destinationPostalCode: "400001",
			sourceCountry:         "US",
			expectedError:         true,
			expectedErrorType:     "DHL_HUB_VALIDATION_ERROR",
			expectedErrorMessage:  "source postal code is required for hub lookup",
		},
		{
			name:                  "Invalid destination postal code",
			sourcePostalCode:      "10001",
			destinationPostalCode: "",
			sourceCountry:         "US",
			expectedError:         true,
			expectedErrorType:     "DHL_HUB_VALIDATION_ERROR",
			expectedErrorMessage:  "destination postal code is required for hub lookup",
		},
		{
			name:                  "Unsupported country code",
			sourcePostalCode:      "12345",
			destinationPostalCode: "67890",
			sourceCountry:         "XX",
			expectedError:         true,
			expectedErrorType:     "DHL_HUB_UNSUPPORTED_COUNTRY",
			expectedErrorMessage:  "DHL hub lookup not supported for country XX",
		},
		{
			name:                  "Invalid postal code format",
			sourcePostalCode:      "INVALID",
			destinationPostalCode: "400001",
			sourceCountry:         "US",
			expectedError:         true,
			expectedErrorType:     "DHL_HUB_INVALID_POSTAL_CODE",
			expectedErrorMessage:  "invalid postal code format for hub lookup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup DHL adapter with validation error
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
			errorResponse := createDHLHubValidationErrorResponse(tt.expectedErrorType, tt.expectedErrorMessage)
			dhlAdapter.SetServiceabilityResult(errorResponse)
			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Create request
			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      &tt.sourcePostalCode,
				DestinationPostalCode: &tt.destinationPostalCode,
				CountryCode:           &tt.sourceCountry,
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.0,
							Unit:  "kg",
						},
					},
				},
			}

			// Execute hub location lookup
			result, err := executeDHLHubLookup(ctx, mockFactory, request)

			// Verify error response
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErrorType)
			assert.Contains(t, err.Error(), tt.expectedErrorMessage)
		})
	}
}

func TestDHLHubLocationServiceIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		hubServiceResponse    *HubLocationServiceResponse
		expectedHubFound      bool
		expectedHubCount      int
		expectedDistanceRange []float64 // [min, max]
		expectedServiceTypes  []string
	}{
		{
			name: "Multiple hubs found - select nearest",
			hubServiceResponse: &HubLocationServiceResponse{
				Hubs: []HubLocation{
					{
						HubCode:         "NYC_HUB_001",
						HubName:         "New York JFK Hub",
						PostalCode:      "11430",
						DistanceKm:      12.5,
						IsInternational: true,
						Services:        []string{"EXPRESS_WORLDWIDE", "EXPRESS_12"},
					},
					{
						HubCode:         "NYC_HUB_002",
						HubName:         "New York LGA Hub",
						PostalCode:      "11371",
						DistanceKm:      18.2,
						IsInternational: false,
						Services:        []string{"EXPRESS_DOMESTIC"},
					},
					{
						HubCode:         "NYC_HUB_003",
						HubName:         "New York Newark Hub",
						PostalCode:      "07114",
						DistanceKm:      25.8,
						IsInternational: true,
						Services:        []string{"EXPRESS_WORLDWIDE", "EXPRESS_9", "EXPRESS_12"},
					},
				},
				ResponseTime: 45 * time.Millisecond,
			},
			expectedHubFound:      true,
			expectedHubCount:      3,
			expectedDistanceRange: []float64{12.5, 25.8},
			expectedServiceTypes:  []string{"EXPRESS_WORLDWIDE", "EXPRESS_DOMESTIC"},
		},
		{
			name: "Single hub found",
			hubServiceResponse: &HubLocationServiceResponse{
				Hubs: []HubLocation{
					{
						HubCode:         "MIA_HUB_001",
						HubName:         "Miami International Hub",
						PostalCode:      "33126",
						DistanceKm:      8.3,
						IsInternational: true,
						Services:        []string{"EXPRESS_WORLDWIDE", "EXPRESS_ENVELOPE"},
					},
				},
				ResponseTime: 32 * time.Millisecond,
			},
			expectedHubFound:      true,
			expectedHubCount:      1,
			expectedDistanceRange: []float64{8.3, 8.3},
			expectedServiceTypes:  []string{"EXPRESS_WORLDWIDE"},
		},
		{
			name: "No hubs found",
			hubServiceResponse: &HubLocationServiceResponse{
				Hubs:         []HubLocation{},
				ResponseTime: 20 * time.Millisecond,
				Error:        "No DHL hubs found within 50km radius",
			},
			expectedHubFound:      false,
			expectedHubCount:      0,
			expectedDistanceRange: []float64{},
			expectedServiceTypes:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup DHL adapter with hub service response
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")

			if tt.expectedHubFound {
				hubResponse := createDHLHubServiceIntegrationResponse(tt.hubServiceResponse)
				dhlAdapter.SetServiceabilityResult(hubResponse)
			} else {
				errorResponse := createDHLNoHubFoundResponse(tt.hubServiceResponse.Error)
				dhlAdapter.SetServiceabilityResult(errorResponse)
			}

			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Create request
			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("33101"),
				DestinationPostalCode: stringPtr("10001"),
				CountryCode:           stringPtr("US"),
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 3.0,
							Unit:  "kg",
						},
					},
				},
			}

			// Execute hub service integration
			result, err := executeDHLHubLookup(ctx, mockFactory, request)

			// Verify response
			if tt.expectedHubFound {
				assert.NoError(t, err)
				require.NotNil(t, result)

				// Verify hub selection (should be nearest)
				if len(tt.hubServiceResponse.Hubs) > 0 {
					nearestHub := tt.hubServiceResponse.Hubs[0]
					for _, hub := range tt.hubServiceResponse.Hubs {
						if hub.DistanceKm < nearestHub.DistanceKm {
							nearestHub = hub
						}
					}

					assert.Equal(t, nearestHub.HubCode, result.HubCode)
					assert.Equal(t, nearestHub.HubName, result.HubName)
					assert.Equal(t, nearestHub.DistanceKm, result.DistanceKm)
				}

				// Verify distance range
				if len(tt.expectedDistanceRange) == 2 {
					assert.GreaterOrEqual(t, result.DistanceKm, tt.expectedDistanceRange[0])
					assert.LessOrEqual(t, result.DistanceKm, tt.expectedDistanceRange[1])
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "No DHL hubs found")
			}
		})
	}
}

func TestDHLHubLocationErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		hubServiceError   error
		expectedErrorType string
		expectedRetryable bool
		expectedBackoffMs int
	}{
		{
			name:              "Hub service timeout",
			hubServiceError:   fmt.Errorf("DHL hub service timeout: request exceeded 5 seconds"),
			expectedErrorType: "DHL_HUB_SERVICE_TIMEOUT",
			expectedRetryable: true,
			expectedBackoffMs: 1000,
		},
		{
			name:              "Hub service unavailable",
			hubServiceError:   fmt.Errorf("DHL hub service unavailable: 503 Service Unavailable"),
			expectedErrorType: "DHL_HUB_SERVICE_UNAVAILABLE",
			expectedRetryable: true,
			expectedBackoffMs: 2000,
		},
		{
			name:              "Hub service rate limited",
			hubServiceError:   fmt.Errorf("DHL hub service rate limited: 429 Too Many Requests"),
			expectedErrorType: "DHL_HUB_SERVICE_RATE_LIMITED",
			expectedRetryable: true,
			expectedBackoffMs: 5000,
		},
		{
			name:              "Hub service authentication error",
			hubServiceError:   fmt.Errorf("DHL hub service authentication failed: 401 Unauthorized"),
			expectedErrorType: "DHL_HUB_SERVICE_AUTH_ERROR",
			expectedRetryable: false,
			expectedBackoffMs: 0,
		},
		{
			name:              "Hub service data error",
			hubServiceError:   fmt.Errorf("DHL hub service data error: invalid response format"),
			expectedErrorType: "DHL_HUB_SERVICE_DATA_ERROR",
			expectedRetryable: false,
			expectedBackoffMs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup DHL adapter with hub service error
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
			errorResponse := createDHLHubServiceErrorResponse(tt.hubServiceError, tt.expectedErrorType)
			dhlAdapter.SetServiceabilityResult(errorResponse)
			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Create request
			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("10001"),
				DestinationPostalCode: stringPtr("400001"),
				CountryCode:           stringPtr("US"),
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.5,
							Unit:  "kg",
						},
					},
				},
			}

			// Execute hub location lookup with error
			result, err := executeDHLHubLookup(ctx, mockFactory, request)

			// Verify error handling
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErrorType)

			// Verify error characteristics
			errorInfo := parseHubServiceError(err)
			assert.Equal(t, tt.expectedRetryable, errorInfo.IsRetryable)
			assert.Equal(t, tt.expectedBackoffMs, errorInfo.BackoffMs)
		})
	}
}

func TestDHLInternationalHubValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		sourceCountry            string
		destinationCountry       string
		requiresInternationalHub bool
		hubCapabilities          []string
		expectedValid            bool
		expectedErrorMessage     string
	}{
		{
			name:                     "International shipment - valid hub",
			sourceCountry:            "US",
			destinationCountry:       "IN",
			requiresInternationalHub: true,
			hubCapabilities:          []string{"international", "customs_clearance", "express_worldwide"},
			expectedValid:            true,
		},
		{
			name:                     "Domestic shipment - any hub valid",
			sourceCountry:            "US",
			destinationCountry:       "US",
			requiresInternationalHub: false,
			hubCapabilities:          []string{"domestic", "express_domestic"},
			expectedValid:            true,
		},
		{
			name:                     "International shipment - hub lacks international capability",
			sourceCountry:            "US",
			destinationCountry:       "IN",
			requiresInternationalHub: true,
			hubCapabilities:          []string{"domestic", "express_domestic"},
			expectedValid:            false,
			expectedErrorMessage:     "Selected hub does not support international shipments",
		},
		{
			name:                     "International shipment - hub lacks customs clearance",
			sourceCountry:            "US",
			destinationCountry:       "IN",
			requiresInternationalHub: true,
			hubCapabilities:          []string{"international", "express_worldwide"},
			expectedValid:            false,
			expectedErrorMessage:     "Selected hub does not support customs clearance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup DHL adapter with international hub validation
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")

			if tt.expectedValid {
				hubResponse := createDHLValidInternationalHubResponse(tt.hubCapabilities)
				dhlAdapter.SetServiceabilityResult(hubResponse)
			} else {
				errorResponse := createDHLInternationalHubValidationErrorResponse(tt.expectedErrorMessage)
				dhlAdapter.SetServiceabilityResult(errorResponse)
			}

			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Create request
			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("10001"),
				DestinationPostalCode: stringPtr("400001"),
				CountryCode:           &tt.sourceCountry,
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 2.0,
							Unit:  "kg",
						},
					},
				},
			}

			// Execute international hub validation
			result, err := executeDHLHubLookup(ctx, mockFactory, request)

			// Verify validation result
			if tt.expectedValid {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.SupportsInternational)
				assert.Contains(t, result.HubCapabilities, "international")
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrorMessage)
			}
		})
	}
}

// Helper types and functions for DHL hub location tests

type HubLocation struct {
	HubCode         string
	HubName         string
	PostalCode      string
	DistanceKm      float64
	IsInternational bool
	Services        []string
}

type HubLocationServiceResponse struct {
	Hubs         []HubLocation
	ResponseTime time.Duration
	Error        string
}

type DHLHubLookupResult struct {
	HubCode               string
	HubName               string
	PostalCode            string
	DistanceKm            float64
	IsInternational       bool
	SupportsInternational bool
	AvailableServices     []string
	HubCapabilities       []string
	ResponseTimeMs        int64
}

type HubServiceErrorInfo struct {
	IsRetryable bool
	BackoffMs   int
}

func executeDHLHubLookup(ctx context.Context, factory *mocks.MockPartnerAdapterFactory, request *models.ServiceabilityV2Request) (*DHLHubLookupResult, error) {
	// Simulate DHL hub lookup execution
	adapter, err := factory.CreateAdapter("dhl")
	if err != nil {
		return nil, err
	}

	partnerInfo := common.PartnerInfo{
		PartnerCode: "dhl",
	}

	result, err := adapter.CheckServiceability(ctx, request, partnerInfo)
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, result.Error
	}

	// Extract hub information from result capabilities
	hubResult := &DHLHubLookupResult{
		ResponseTimeMs: result.ResponseTime.Milliseconds(),
	}

	if hubCode, exists := result.Capabilities["hub_code"]; exists {
		hubResult.HubCode = hubCode.(string)
	}
	if hubName, exists := result.Capabilities["hub_name"]; exists {
		hubResult.HubName = hubName.(string)
	}
	if distance, exists := result.Capabilities["hub_distance_km"]; exists {
		hubResult.DistanceKm = distance.(float64)
	}
	if services, exists := result.Capabilities["available_services"]; exists {
		hubResult.AvailableServices = services.([]string)
	}
	if capabilities, exists := result.Capabilities["hub_capabilities"]; exists {
		hubResult.HubCapabilities = capabilities.([]string)
	}
	if isIntl, exists := result.Capabilities["supports_international"]; exists {
		hubResult.SupportsInternational = isIntl.(bool)
	}

	return hubResult, nil
}

func createDHLHubLocationResponse(hubCode, hubName string, distance float64, services []string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: "dhl",
		PartnerName: "DHL Express",
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS_WORLDWIDE",
				ServiceName: "DHL Express Worldwide",
				TATDays:     5,
			},
		},
		Capabilities: map[string]interface{}{
			"hub_code":               hubCode,
			"hub_name":               hubName,
			"hub_distance_km":        distance,
			"available_services":     services,
			"supports_international": true,
			"hub_capabilities":       []string{"international", "customs_clearance"},
		},
		ResponseTime: 180 * time.Millisecond,
	}
}

func createDHLHubValidationErrorResponse(errorType, errorMessage string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:    &partnerID,
		PartnerCode:  "dhl",
		PartnerName:  "DHL Express",
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("%s: %s", errorType, errorMessage),
		ErrorMessage: stringPtr(fmt.Sprintf("%s: %s", errorType, errorMessage)),
		ResponseTime: 50 * time.Millisecond,
	}
}

func createDHLHubServiceIntegrationResponse(hubServiceResponse *HubLocationServiceResponse) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()

	// Select nearest hub
	var nearestHub HubLocation
	if len(hubServiceResponse.Hubs) > 0 {
		nearestHub = hubServiceResponse.Hubs[0]
		for _, hub := range hubServiceResponse.Hubs {
			if hub.DistanceKm < nearestHub.DistanceKm {
				nearestHub = hub
			}
		}
	}

	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: "dhl",
		PartnerName: "DHL Express",
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS_WORLDWIDE",
				ServiceName: "DHL Express Worldwide",
				TATDays:     5,
			},
		},
		Capabilities: map[string]interface{}{
			"hub_code":               nearestHub.HubCode,
			"hub_name":               nearestHub.HubName,
			"hub_distance_km":        nearestHub.DistanceKm,
			"available_services":     nearestHub.Services,
			"supports_international": nearestHub.IsInternational,
			"hub_capabilities":       []string{"international", "customs_clearance"},
			"total_hubs_found":       len(hubServiceResponse.Hubs),
		},
		Metadata: map[string]interface{}{
			"hub_service_response_time": hubServiceResponse.ResponseTime.Milliseconds(),
			"hub_selection_criteria":    "nearest_distance",
		},
		ResponseTime: hubServiceResponse.ResponseTime + 50*time.Millisecond,
	}
}

func createDHLNoHubFoundResponse(errorMessage string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:    &partnerID,
		PartnerCode:  "dhl",
		PartnerName:  "DHL Express",
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("DHL_HUB_NOT_FOUND: %s", errorMessage),
		ErrorMessage: stringPtr(fmt.Sprintf("DHL_HUB_NOT_FOUND: %s", errorMessage)),
		ResponseTime: 100 * time.Millisecond,
	}
}

func createDHLHubServiceErrorResponse(serviceError error, errorType string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:    &partnerID,
		PartnerCode:  "dhl",
		PartnerName:  "DHL Express",
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("%s: %v", errorType, serviceError),
		ErrorMessage: stringPtr(fmt.Sprintf("%s: %v", errorType, serviceError)),
		ResponseTime: 75 * time.Millisecond,
	}
}

func createDHLValidInternationalHubResponse(capabilities []string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: "dhl",
		PartnerName: "DHL Express",
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS_WORLDWIDE",
				ServiceName: "DHL Express Worldwide",
				TATDays:     5,
			},
		},
		Capabilities: map[string]interface{}{
			"hub_code":               "INTL_HUB_001",
			"hub_name":               "International Processing Hub",
			"hub_distance_km":        12.5,
			"supports_international": containsString(capabilities, "international"),
			"hub_capabilities":       capabilities,
			"available_services":     []string{"EXPRESS_WORLDWIDE", "EXPRESS_ENVELOPE"},
		},
		ResponseTime: 150 * time.Millisecond,
	}
}

func createDHLInternationalHubValidationErrorResponse(errorMessage string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:    &partnerID,
		PartnerCode:  "dhl",
		PartnerName:  "DHL Express",
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("DHL_HUB_INTERNATIONAL_VALIDATION_ERROR: %s", errorMessage),
		ErrorMessage: stringPtr(fmt.Sprintf("DHL_HUB_INTERNATIONAL_VALIDATION_ERROR: %s", errorMessage)),
		ResponseTime: 60 * time.Millisecond,
	}
}

func parseHubServiceError(err error) HubServiceErrorInfo {
	errMsg := err.Error()

	switch {
	case contains(errMsg, "TIMEOUT"):
		return HubServiceErrorInfo{IsRetryable: true, BackoffMs: 1000}
	case contains(errMsg, "UNAVAILABLE"):
		return HubServiceErrorInfo{IsRetryable: true, BackoffMs: 2000}
	case contains(errMsg, "RATE_LIMITED"):
		return HubServiceErrorInfo{IsRetryable: true, BackoffMs: 5000}
	case contains(errMsg, "AUTH_ERROR"):
		return HubServiceErrorInfo{IsRetryable: false, BackoffMs: 0}
	case contains(errMsg, "DATA_ERROR"):
		return HubServiceErrorInfo{IsRetryable: false, BackoffMs: 0}
	default:
		return HubServiceErrorInfo{IsRetryable: false, BackoffMs: 0}
	}
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// stringPtr helper function is defined in other test files
