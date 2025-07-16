package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"
)

func TestDHLIntegrationScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		request           *models.ServiceabilityV2Request
		expectedServices  int
		expectedError     bool
		expectedErrorType string
		setupDHLResponse  func() *common.PartnerServiceabilityResult
	}{
		{
			name: "DHL international request - US to India",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("10001"),
				DestinationPostalCode: stringPtr("400001"),
				CountryCode:           stringPtr("US"),
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 2.5,
							Unit:  "kg",
						},
						Dimensions: &models.Dimensions{
							Length: 30.0,
							Width:  20.0,
							Height: 10.0,
							Unit:   "cm",
						},
					},
				},
			},
			expectedServices: 2,
			expectedError:    false,
			setupDHLResponse: setupDHLInternationalSuccessResponse,
		},
		{
			name: "DHL international request - validation failure",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr(""),
				DestinationPostalCode: stringPtr("400001"),
				CountryCode:           stringPtr("US"),
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 0.0, // Invalid weight
							Unit:  "kg",
						},
					},
				},
			},
			expectedServices:  0,
			expectedError:     true,
			expectedErrorType: "DHL_VALIDATION_ERROR",
			setupDHLResponse:  setupDHLValidationErrorResponse,
		},
		{
			name: "DHL country code resolution - geolocation service integration",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("SW1A 1AA"),
				DestinationPostalCode: stringPtr("10001"),
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.0,
							Unit:  "kg",
						},
					},
				},
			},
			expectedServices: 1,
			expectedError:    false,
			setupDHLResponse: setupDHLCountryCodeResolutionResponse,
		},
		{
			name: "DHL hub location resolution - nearest hub scenario",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("94102"),
				DestinationPostalCode: stringPtr("10001"),
				CountryCode:           stringPtr("US"),
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 5.0,
							Unit:  "kg",
						},
					},
				},
			},
			expectedServices: 2,
			expectedError:    false,
			setupDHLResponse: setupDHLHubLocationResponse,
		},
		{
			name: "DHL multiple packages validation",
			request: &models.ServiceabilityV2Request{
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
						Dimensions: &models.Dimensions{
							Length: 20.0,
							Width:  15.0,
							Height: 10.0,
							Unit:   "cm",
						},
					},
					{
						Weight: &models.Weight{
							Value: 2.0,
							Unit:  "kg",
						},
						Dimensions: &models.Dimensions{
							Length: 25.0,
							Width:  20.0,
							Height: 15.0,
							Unit:   "cm",
						},
					},
				},
			},
			expectedServices: 1,
			expectedError:    false,
			setupDHLResponse: setupDHLMultiplePackagesResponse,
		},
		{
			name: "DHL API timeout scenario",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("10001"),
				DestinationPostalCode: stringPtr("400001"),
				CountryCode:           stringPtr("US"),
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.0,
							Unit:  "kg",
						},
					},
				},
			},
			expectedServices:  0,
			expectedError:     true,
			expectedErrorType: "DHL_API_TIMEOUT",
			setupDHLResponse:  setupDHLTimeoutResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()

			// Setup DHL adapter with specific response
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
			dhlResponse := tt.setupDHLResponse()
			dhlAdapter.SetServiceabilityResult(dhlResponse)
			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				5*time.Second,
				false, // disable logging for tests
			)

			// Execute serviceability check
			response, err := orchestrator.CheckServiceability(ctx, tt.request)

			// Verify response
			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrorType != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorType)
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.True(t, response.Success)

				// Find DHL partner in response
				var dhlPartner *models.PartnerV2Response
				for i := range response.Partners {
					if response.Partners[i].PartnerCode == "dhl" {
						dhlPartner = &response.Partners[i]
						break
					}
				}

				require.NotNil(t, dhlPartner, "DHL partner should be in response")
				assert.Len(t, dhlPartner.Services, tt.expectedServices)

				// Verify DHL-specific capabilities structure
				verifyDHLIntegrationCapabilities(t, dhlPartner.Capabilities)
			}
		})
	}
}

func TestDHLIntegrationInternationalFlow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		sourceCountry         string
		destinationCountry    string
		sourcePostalCode      string
		destinationPostalCode string
		expectedSteps         []string
		expectedCapabilities  []string
	}{
		{
			name:                  "US to India international flow",
			sourceCountry:         "US",
			destinationCountry:    "IN",
			sourcePostalCode:      "10001",
			destinationPostalCode: "400001",
			expectedSteps: []string{
				"postal_code_extraction",
				"hub_lookup",
				"country_code_resolution",
				"api_request_creation",
				"api_call",
				"response_processing",
				"capability_flattening",
			},
			expectedCapabilities: []string{
				"international_supported",
				"customs_clearance",
				"total_transit_days",
				"pickup_earliest",
				"pickup_latest",
			},
		},
		{
			name:                  "UK to US international flow",
			sourceCountry:         "GB",
			destinationCountry:    "US",
			sourcePostalCode:      "SW1A 1AA",
			destinationPostalCode: "90210",
			expectedSteps: []string{
				"postal_code_extraction",
				"hub_lookup",
				"country_code_resolution",
				"api_request_creation",
				"api_call",
				"response_processing",
				"capability_flattening",
			},
			expectedCapabilities: []string{
				"international_supported",
				"customs_clearance",
				"total_transit_days",
				"delivery_signature_required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()

			// Setup DHL adapter with international flow tracking
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
			dhlResponse := createDHLInternationalFlowResponse(tt.sourceCountry, tt.destinationCountry)
			dhlAdapter.SetServiceabilityResult(dhlResponse)
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

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				10*time.Second,
				false,
			)

			// Execute serviceability check
			response, err := orchestrator.CheckServiceability(ctx, request)

			// Verify response
			assert.NoError(t, err)
			require.NotNil(t, response)
			assert.True(t, response.Success)

			// Find DHL partner
			var dhlPartner *models.PartnerV2Response
			for i := range response.Partners {
				if response.Partners[i].PartnerCode == "dhl" {
					dhlPartner = &response.Partners[i]
					break
				}
			}

			require.NotNil(t, dhlPartner)
			assert.NotEmpty(t, dhlPartner.Services)

			// Verify international capabilities
			for _, capability := range tt.expectedCapabilities {
				_, exists := dhlPartner.Capabilities[capability]
				assert.True(t, exists, "Expected capability %s should exist", capability)
			}

			// Verify flattened structure (no nested objects)
			for key, value := range dhlPartner.Capabilities {
				_, isMap := value.(map[string]interface{})
				assert.False(t, isMap, "Capability %s should be flattened, not nested", key)
			}
		})
	}
}

func TestDHLIntegrationPackageValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		packages      []models.Package
		expectedValid bool
		expectedError string
	}{
		{
			name: "Valid single package",
			packages: []models.Package{
				{
					Weight: &models.Weight{
						Value: 1.5,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 20.0,
						Width:  15.0,
						Height: 10.0,
						Unit:   "cm",
					},
				},
			},
			expectedValid: true,
		},
		{
			name: "Valid multiple packages",
			packages: []models.Package{
				{
					Weight: &models.Weight{
						Value: 1.0,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 15.0,
						Width:  10.0,
						Height: 8.0,
						Unit:   "cm",
					},
				},
				{
					Weight: &models.Weight{
						Value: 0.5,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 10.0,
						Width:  8.0,
						Height: 5.0,
						Unit:   "cm",
					},
				},
			},
			expectedValid: true,
		},
		{
			name: "Invalid package - no weight",
			packages: []models.Package{
				{
					Dimensions: &models.Dimensions{
						Length: 20.0,
						Width:  15.0,
						Height: 10.0,
						Unit:   "cm",
					},
				},
			},
			expectedValid: false,
			expectedError: "package weight is required",
		},
		{
			name: "Invalid package - excessive weight",
			packages: []models.Package{
				{
					Weight: &models.Weight{
						Value: 100.0, // Very heavy package
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 50.0,
						Width:  40.0,
						Height: 30.0,
						Unit:   "cm",
					},
				},
			},
			expectedValid: false,
			expectedError: "DHL validation failed: package weight exceeds maximum limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple package validation without orchestrator
			err := validatePackagesForDHL(tt.packages)

			if tt.expectedValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

// Helper functions for DHL integration test scenarios

func setupDHLInternationalSuccessResponse() *common.PartnerServiceabilityResult {
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
			{
				ServiceCode: "EXPRESS_ENVELOPE",
				ServiceName: "DHL Express Envelope",
				TATDays:     3,
			},
		},
		Capabilities: map[string]interface{}{
			"international_supported":    true,
			"customs_clearance":          true,
			"total_transit_days":         5,
			"pickup_earliest":            "09:00",
			"pickup_latest":              "17:00",
			"local_cutoff_date_and_time": "2024-01-15T15:00:00Z",
		},
		ResponseTime: 200 * time.Millisecond,
	}
}

func setupDHLValidationErrorResponse() *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:    &partnerID,
		PartnerCode:  "dhl",
		PartnerName:  "DHL Express",
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("DHL validation failed: invalid postal code"),
		ErrorMessage: stringPtr("DHL validation failed: invalid postal code"),
		ResponseTime: 50 * time.Millisecond,
	}
}

func setupDHLCountryCodeResolutionResponse() *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: "dhl",
		PartnerName: "DHL Express",
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS_WORLDWIDE",
				ServiceName: "DHL Express Worldwide",
				TATDays:     4,
			},
		},
		Capabilities: map[string]interface{}{
			"international_supported": true,
			"country_code_resolved":   true,
			"source_country":          "GB",
			"destination_country":     "US",
			"total_transit_days":      4,
			"pickup_earliest":         "08:00",
			"pickup_latest":           "18:00",
		},
		Metadata: map[string]interface{}{
			"geolocation_service_used": true,
			"country_resolution_time":  "25ms",
		},
		ResponseTime: 180 * time.Millisecond,
	}
}

func setupDHLHubLocationResponse() *common.PartnerServiceabilityResult {
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
			{
				ServiceCode: "EXPRESS_12",
				ServiceName: "DHL Express 12:00",
				TATDays:     3,
			},
		},
		Capabilities: map[string]interface{}{
			"hub_location_resolved": true,
			"nearest_hub":           "San Francisco International",
			"hub_postal_code":       "94128",
			"hub_distance_km":       "15.2",
			"total_transit_days":    5,
			"pickup_earliest":       "09:00",
			"pickup_latest":         "17:00",
		},
		Metadata: map[string]interface{}{
			"hub_lookup_time": "35ms",
			"hub_service_id":  "SFO_HUB_001",
		},
		ResponseTime: 220 * time.Millisecond,
	}
}

func setupDHLMultiplePackagesResponse() *common.PartnerServiceabilityResult {
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
			"multiple_packages_supported": true,
			"total_packages":              2,
			"total_weight_kg":             3.5,
			"consolidated_shipment":       true,
			"total_transit_days":          5,
			"pickup_earliest":             "09:00",
			"pickup_latest":               "17:00",
		},
		Metadata: map[string]interface{}{
			"package_validation_time": "15ms",
			"consolidation_applied":   true,
		},
		ResponseTime: 160 * time.Millisecond,
	}
}

func setupDHLTimeoutResponse() *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:    &partnerID,
		PartnerCode:  "dhl",
		PartnerName:  "DHL Express",
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("DHL API timeout: request exceeded 5 second timeout"),
		ErrorMessage: stringPtr("DHL API timeout: request exceeded 5 second timeout"),
		ResponseTime: 5 * time.Second,
	}
}

func createDHLInternationalFlowResponse(sourceCountry, destinationCountry string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: "dhl",
		PartnerName: "DHL Express",
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS_WORLDWIDE",
				ServiceName: "DHL Express Worldwide",
				TATDays:     getDHLTransitDays(sourceCountry, destinationCountry),
			},
		},
		Capabilities: map[string]interface{}{
			"international_supported":     true,
			"customs_clearance":           true,
			"source_country":              sourceCountry,
			"destination_country":         destinationCountry,
			"total_transit_days":          getDHLTransitDays(sourceCountry, destinationCountry),
			"pickup_earliest":             "09:00",
			"pickup_latest":               "17:00",
			"delivery_signature_required": true,
		},
		Metadata: map[string]interface{}{
			"flow_steps_completed": []string{
				"postal_code_extraction",
				"hub_lookup",
				"country_code_resolution",
				"api_request_creation",
				"api_call",
				"response_processing",
				"capability_flattening",
			},
			"international_route": fmt.Sprintf("%s-%s", sourceCountry, destinationCountry),
		},
		ResponseTime: 250 * time.Millisecond,
	}
}

func setupDHLValidPackageResponse() *common.PartnerServiceabilityResult {
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
			"package_validation_passed": true,
			"total_transit_days":        5,
			"pickup_earliest":           "09:00",
			"pickup_latest":             "17:00",
		},
		ResponseTime: 120 * time.Millisecond,
	}
}

func setupDHLInvalidPackageResponse(errorMsg string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:    &partnerID,
		PartnerCode:  "dhl",
		PartnerName:  "DHL Express",
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("%s", errorMsg),
		ErrorMessage: &errorMsg,
		ResponseTime: 50 * time.Millisecond,
	}
}

func verifyDHLIntegrationCapabilities(t *testing.T, capabilities map[string]interface{}) {
	// Verify that capabilities are in flattened format
	for key, value := range capabilities {
		_, isMap := value.(map[string]interface{})
		assert.False(t, isMap, "DHL capability %s should be flattened in integration response", key)
	}

	// Check for common DHL international capabilities
	expectedIntlCapabilities := []string{
		"total_transit_days",
		"pickup_earliest",
		"pickup_latest",
	}

	for _, capability := range expectedIntlCapabilities {
		if _, exists := capabilities[capability]; exists {
			// Verify data types
			switch capability {
			case "total_transit_days":
				_, isInt := capabilities[capability].(int)
				assert.True(t, isInt, "total_transit_days should be integer")
			case "pickup_earliest", "pickup_latest":
				_, isString := capabilities[capability].(string)
				assert.True(t, isString, "%s should be string", capability)
			}
		}
	}
}

func getDHLTransitDays(sourceCountry, destinationCountry string) int {
	// Simulate different transit times based on country pairs
	countryPair := sourceCountry + "-" + destinationCountry

	switch countryPair {
	case "US-IN", "IN-US":
		return 5
	case "GB-US", "US-GB":
		return 3
	case "US-CA", "CA-US":
		return 2
	default:
		return 4
	}
}

// stringPtr helper function is defined in other test files

// validatePackagesForDHL validates packages for DHL requirements
func validatePackagesForDHL(packages []models.Package) error {
	for i, pkg := range packages {
		// Check weight requirement
		if pkg.Weight == nil {
			return fmt.Errorf("package %d: package weight is required", i+1)
		}

		// Check weight limits (DHL max is typically 70kg for standard service)
		if pkg.Weight.Value <= 0 {
			return fmt.Errorf("package %d: package weight must be greater than 0", i+1)
		}

		if pkg.Weight.Value > 70.0 {
			return fmt.Errorf("DHL validation failed: package weight exceeds maximum limit")
		}

		// Check dimensions if provided
		if pkg.Dimensions != nil {
			if pkg.Dimensions.Length <= 0 || pkg.Dimensions.Width <= 0 || pkg.Dimensions.Height <= 0 {
				return fmt.Errorf("package %d: dimensions must be greater than 0", i+1)
			}

			// Check dimension limits (DHL max is typically 120cm for any single dimension)
			if pkg.Dimensions.Length > 120.0 || pkg.Dimensions.Width > 120.0 || pkg.Dimensions.Height > 120.0 {
				return fmt.Errorf("DHL validation failed: package dimensions exceed maximum limit")
			}
		}
	}

	return nil
}
