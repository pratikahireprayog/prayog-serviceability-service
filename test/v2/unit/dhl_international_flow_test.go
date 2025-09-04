package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"
)

// TestDHLInternationalFlow tests the complete DHL international serviceability flow
func TestDHLInternationalFlow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		request                *models.ServiceabilityV2Request
		setupMocks             func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectServiceable      bool
		expectedMetadataFields []string
		description            string
	}{
		{
			name: "DHL_CompleteInternationalFlow_Success",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("10001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 2.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 30.0, Width: 20.0, Height: 15.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLInternationalFlowAdapters()
			},
			expectServiceable: true,
			expectedMetadataFields: []string{
				"reason",
				"product_count",
				"exchange_rates",
				"flow",
				"source_country_code",
				"destination_country_code",
				"hub_postal_code",
				"international_hub_code",
			},
			description: "Complete DHL international flow should succeed with all steps",
		},
		{
			name: "DHL_HubLookupFailure",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("999999"), // Invalid postal code for hub lookup
				DestinationPostalCode: stringPtr("10001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLHubLookupErrorAdapters()
			},
			expectServiceable: false,
			expectedMetadataFields: []string{
				"reason",
				"source_pincode",
				"step",
			},
			description: "DHL should handle hub lookup failures gracefully",
		},
		{
			name: "DHL_CountryCodeResolutionFailure",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("INVALID"), // Invalid postal code for country resolution
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLCountryCodeErrorAdapters()
			},
			expectServiceable: false,
			expectedMetadataFields: []string{
				"reason",
				"source_pincode",
				"destination_pincode",
				"step",
			},
			description: "DHL should handle country code resolution failures",
		},
		{
			name: "DHL_APICallFailure",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLAPIErrorAdapters()
			},
			expectServiceable: false,
			expectedMetadataFields: []string{
				"reason",
				"source_country_code",
				"destination_country_code",
				"step",
			},
			description: "DHL should handle API call failures",
		},
		{
			name: "DHL_MultiPackageInternationalFlow",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("10001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.0, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 20.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
					{
						Weight:     &models.Weight{Value: 2.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 35.0, Width: 25.0, Height: 20.0, Unit: "cm"},
					},
					{
						Weight:     &models.Weight{Value: 0.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 15.0, Width: 10.0, Height: 5.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLInternationalFlowAdapters()
			},
			expectServiceable: true,
			expectedMetadataFields: []string{
				"reason",
				"product_count",
				"flow",
				"source_country_code",
				"destination_country_code",
			},
			description: "DHL should handle multiple packages in international flow",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			factory := mocks.NewMockPartnerAdapterFactory()
			repo := mocks.NewMockPartnerAttributeMapRepository()
			tt.setupMocks(factory, repo)

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				factory,
				repo,
				30*time.Second,
				false, // Return all partners to see error details
			)

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, tt.request)

			// Verify
			require.NoError(t, err, "Should not return orchestrator error: %s", tt.description)
			require.NotNil(t, response, "Response should not be nil")
			require.NotEmpty(t, response.Partners, "Should have DHL partner in response")

			// Find DHL partner in response
			var dhlPartner *models.PartnerV2Response
			for i := range response.Partners {
				if response.Partners[i].PartnerCode == "dhl" {
					dhlPartner = &response.Partners[i]
					break
				}
			}

			require.NotNil(t, dhlPartner, "DHL partner should be in response")

			if tt.expectServiceable {
				// Verify successful serviceability
				if dhlPartner.Error != nil {
					assert.Nil(t, dhlPartner.Error, "DHL should not have error for serviceable request: %s (got: %s)",
						tt.description, *dhlPartner.Error)
				}

				// Verify capabilities are present (even if services are empty for DHL)
				assert.NotEmpty(t, dhlPartner.Capabilities, "DHL should have capabilities for serviceable request")

				// Verify flattened capabilities structure
				assert.Contains(t, dhlPartner.Capabilities, "total_transit_days", "Should have flattened delivery capabilities")
				assert.Contains(t, dhlPartner.Capabilities, "next_business_day", "Should have flattened pickup capabilities")

			} else {
				// Verify error is present
				assert.NotNil(t, dhlPartner.Error, "DHL should have error for non-serviceable request: %s", tt.description)
				assert.Empty(t, dhlPartner.Services, "DHL should have no services for non-serviceable request")
			}

			// Verify response time is reasonable
			assert.Greater(t, dhlPartner.ResponseTime, time.Duration(0), "Should have positive response time")
			assert.Less(t, dhlPartner.ResponseTime, 5*time.Second, "Response time should be reasonable")

			// Verify partner information
			assert.Equal(t, "dhl", dhlPartner.PartnerCode, "Partner code should be dhl")
			assert.NotEmpty(t, dhlPartner.PartnerID, "Partner ID should be set")
		})
	}
}

// TestDHLInternationalFlowStepByStep tests individual steps of the DHL international flow
func TestDHLInternationalFlowStepByStep(t *testing.T) {
	t.Parallel()

	// Test each step failure individually
	stepTests := []struct {
		name         string
		setupMocks   func(*mocks.MockPartnerAdapterFactory)
		expectedStep string
		description  string
	}{
		{
			name: "Step2_HubLookupFailure",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory) {
				factory.SetupDHLHubLookupErrorAdapters()
			},
			expectedStep: "hub_lookup",
			description:  "Hub lookup should fail gracefully",
		},
		{
			name: "Step3_CountryCodeResolutionFailure",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory) {
				factory.SetupDHLCountryCodeErrorAdapters()
			},
			expectedStep: "country_code_resolution",
			description:  "Country code resolution should fail gracefully",
		},
		{
			name: "Step5_DHLAPICallFailure",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory) {
				factory.SetupDHLAPIErrorAdapters()
			},
			expectedStep: "dhl_api_call",
			description:  "DHL API call should fail gracefully",
		},
	}

	for _, tt := range stepTests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			factory := mocks.NewMockPartnerAdapterFactory()
			repo := mocks.NewMockPartnerAttributeMapRepository()
			factory.SetSupportedPartners([]string{"dhl"})
			tt.setupMocks(factory)

			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				factory,
				repo,
				30*time.Second,
				false,
			)

			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			}

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, request)

			// Verify
			require.NoError(t, err)
			require.NotEmpty(t, response.Partners)

			// Find DHL partner
			var dhlPartner *models.PartnerV2Response
			for i := range response.Partners {
				if response.Partners[i].PartnerCode == "dhl" {
					dhlPartner = &response.Partners[i]
					break
				}
			}

			require.NotNil(t, dhlPartner)
			require.NotNil(t, dhlPartner.Error, "Should have error for step failure: %s", tt.description)

			// Verify error contains step information
			assert.Contains(t, *dhlPartner.Error, tt.expectedStep,
				"Error should indicate which step failed: %s", tt.description)
		})
	}
}

// TestDHLInternationalFlowMetadata tests metadata structure for international flow
func TestDHLInternationalFlowMetadata(t *testing.T) {
	t.Parallel()

	// Setup successful international flow
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()
	factory.SetSupportedPartners([]string{"dhl"})
	factory.SetupDHLInternationalFlowAdapters()

	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		factory,
		repo,
		30*time.Second,
		false,
	)

	request := &models.ServiceabilityV2Request{
		SourcePostalCode:      stringPtr("560001"),
		DestinationPostalCode: stringPtr("10001"),
		CountryCode:           stringPtr("IN"),
		Packages: []models.Package{
			{
				Weight:     &models.Weight{Value: 2.5, Unit: "kg"},
				Dimensions: &models.Dimensions{Length: 30.0, Width: 20.0, Height: 15.0, Unit: "cm"},
			},
		},
	}

	// Execute
	ctx := context.Background()
	response, err := orchestrator.CheckServiceability(ctx, request)

	// Verify
	require.NoError(t, err)
	require.NotEmpty(t, response.Partners)

	// Find DHL partner
	var dhlPartner *models.PartnerV2Response
	for i := range response.Partners {
		if response.Partners[i].PartnerCode == "dhl" {
			dhlPartner = &response.Partners[i]
			break
		}
	}

	require.NotNil(t, dhlPartner)
	require.NotNil(t, dhlPartner.Capabilities)

	// Verify international flow metadata in capabilities or through error structure
	// Since this is a successful flow, check capabilities structure
	metadataFields := map[string]bool{
		"next_business_day":                false,
		"local_cutoff_date_and_time":       false,
		"pickup_earliest":                  false,
		"pickup_latest":                    false,
		"delivery_type_code":               false,
		"estimated_delivery_date_and_time": false,
		"total_transit_days":               false,
		"origin_service_area_code":         false,
		"destination_service_area_code":    false,
	}

	// Check that all expected capability fields are present
	for field := range metadataFields {
		_, exists := dhlPartner.Capabilities[field]
		assert.True(t, exists, "DHL international flow should include capability: %s", field)
	}

	// Verify capability types are correct for international flow
	transitDays, ok := dhlPartner.Capabilities["total_transit_days"]
	assert.True(t, ok, "Should have total_transit_days")
	assert.IsType(t, int(0), transitDays, "total_transit_days should be integer")

	nextBusinessDay, ok := dhlPartner.Capabilities["next_business_day"]
	assert.True(t, ok, "Should have next_business_day")
	assert.IsType(t, true, nextBusinessDay, "next_business_day should be boolean")

	deliveryType, ok := dhlPartner.Capabilities["delivery_type_code"]
	assert.True(t, ok, "Should have delivery_type_code")
	assert.IsType(t, "", deliveryType, "delivery_type_code should be string")
}

// TestDHLAuthenticationFlow tests DHL authentication in the international flow
func TestDHLAuthenticationFlow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockPartnerAdapterFactory)
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name: "DHL_AuthenticationSuccess",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory) {
				factory.SetupDHLInternationalFlowAdapters()
			},
			expectError: false,
			description: "DHL authentication should succeed for valid credentials",
		},
		{
			name: "DHL_AuthenticationFailure",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory) {
				factory.SetupDHLAuthenticationErrorAdapters()
			},
			expectError:   true,
			errorContains: "authentication failed",
			description:   "DHL should handle authentication failures",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			factory := mocks.NewMockPartnerAdapterFactory()
			repo := mocks.NewMockPartnerAttributeMapRepository()
			factory.SetSupportedPartners([]string{"dhl"})
			tt.setupMocks(factory)

			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				factory,
				repo,
				30*time.Second,
				false,
			)

			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			}

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, request)

			// Verify
			require.NoError(t, err)
			require.NotEmpty(t, response.Partners)

			// Find DHL partner
			var dhlPartner *models.PartnerV2Response
			for i := range response.Partners {
				if response.Partners[i].PartnerCode == "dhl" {
					dhlPartner = &response.Partners[i]
					break
				}
			}

			require.NotNil(t, dhlPartner)

			if tt.expectError {
				assert.NotNil(t, dhlPartner.Error, "Should have authentication error: %s", tt.description)
				if tt.errorContains != "" {
					assert.Contains(t, *dhlPartner.Error, tt.errorContains,
						"Error should contain expected authentication failure message")
				}
			} else {
				if dhlPartner.Error != nil {
					assert.Nil(t, dhlPartner.Error, "Should not have authentication error: %s (got: %s)",
						tt.description, *dhlPartner.Error)
				}
			}
		})
	}
}
