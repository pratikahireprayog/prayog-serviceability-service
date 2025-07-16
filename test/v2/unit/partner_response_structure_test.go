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

// TestPartnerV2ResponseStructure tests the updated PartnerV2Response structure
func TestPartnerV2ResponseStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		setupMocks            func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectedPartnersCount int
		validatePartnerFields func(*testing.T, *models.PartnerV2Response)
		description           string
	}{
		{
			name: "PartnerV2Response_AllFieldsPresent",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl", "shipyaari"})
				factory.SetupDefaultAdapters()
			},
			expectedPartnersCount: 2,
			validatePartnerFields: func(t *testing.T, partner *models.PartnerV2Response) {
				// Verify all required fields are present
				assert.NotEmpty(t, partner.PartnerID, "PartnerID should be present")
				assert.NotEmpty(t, partner.PartnerCode, "PartnerCode should be present")
				assert.NotEmpty(t, partner.PartnerName, "PartnerName should be present")
				assert.GreaterOrEqual(t, partner.Rating, float64(0), "Rating should be non-negative")
				assert.NotNil(t, partner.Services, "Services should be initialized")
				assert.NotNil(t, partner.Capabilities, "Capabilities should be initialized")
				assert.Greater(t, partner.ResponseTime, time.Duration(0), "ResponseTime should be positive")

				// Verify PartnerID is string format (UUID string)
				assert.IsType(t, "", partner.PartnerID, "PartnerID should be string")
				assert.Len(t, partner.PartnerID, 36, "PartnerID should be valid UUID string length")
				assert.Contains(t, partner.PartnerID, "-", "PartnerID should contain UUID hyphens")

				// Verify PartnerName follows expected format
				assert.Contains(t, partner.PartnerName, partner.PartnerCode, "PartnerName should contain PartnerCode")
			},
			description: "PartnerV2Response should have all required fields in correct format",
		},
		{
			name: "PartnerV2Response_DHLSpecificStructure",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLAdapters()
			},
			expectedPartnersCount: 1,
			validatePartnerFields: func(t *testing.T, partner *models.PartnerV2Response) {
				if partner.PartnerCode == "dhl" {
					// DHL-specific validations
					assert.Equal(t, "dhl", partner.PartnerCode, "DHL partner code should be correct")
					assert.Contains(t, partner.PartnerName, "DHL", "DHL partner name should contain DHL")

					// Verify DHL capabilities structure (flattened)
					assert.Contains(t, partner.Capabilities, "next_business_day", "DHL should have flattened pickup capabilities")
					assert.Contains(t, partner.Capabilities, "total_transit_days", "DHL should have flattened delivery capabilities")

					// Verify no nested capabilities
					_, hasPickupObject := partner.Capabilities["pickup"]
					_, hasDeliveryObject := partner.Capabilities["delivery"]
					assert.False(t, hasPickupObject, "DHL should not have nested pickup object")
					assert.False(t, hasDeliveryObject, "DHL should not have nested delivery object")

					// Services should be empty for DHL (only provides serviceability status)
					assert.Empty(t, partner.Services, "DHL should not return detailed services in serviceability check")
				}
			},
			description: "DHL partner should have correct structure with flattened capabilities",
		},
		{
			name: "PartnerV2Response_TraditionalPartnerStructure",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"shipyaari"})
				factory.SetupDefaultAdapters()
			},
			expectedPartnersCount: 1,
			validatePartnerFields: func(t *testing.T, partner *models.PartnerV2Response) {
				if partner.PartnerCode == "shipyaari" {
					// Traditional partner validations
					assert.Equal(t, "shipyaari", partner.PartnerCode, "Shipyaari partner code should be correct")
					assert.Contains(t, partner.PartnerName, "shipyaari", "Shipyaari partner name should contain partner code")

					// Verify traditional capabilities structure
					assert.Contains(t, partner.Capabilities, "cod_available", "Traditional partner should have COD capability")
					assert.Contains(t, partner.Capabilities, "tracking", "Traditional partner should have tracking capability")

					// Services should be present for traditional partners
					assert.NotEmpty(t, partner.Services, "Traditional partner should return services")

					// Verify service structure
					for _, service := range partner.Services {
						assert.NotEmpty(t, service.ServiceCode, "Service should have code")
						assert.NotEmpty(t, service.ServiceName, "Service should have name")
						assert.GreaterOrEqual(t, service.TATDays, 0, "TAT days should be non-negative")
					}
				}
			},
			description: "Traditional partner should have correct structure with services",
		},
		{
			name: "PartnerV2Response_ErrorScenario",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			expectedPartnersCount: 1,
			validatePartnerFields: func(t *testing.T, partner *models.PartnerV2Response) {
				if partner.PartnerCode == "dhl" {
					// Error scenario validations
					assert.NotNil(t, partner.Error, "Partner should have error for validation failure")
					assert.NotEmpty(t, *partner.Error, "Error message should not be empty")
					assert.Contains(t, *partner.Error, "DHL validation failed", "Error should indicate DHL validation failure")

					// Structure should still be valid
					assert.NotEmpty(t, partner.PartnerID, "PartnerID should be present even with error")
					assert.NotEmpty(t, partner.PartnerCode, "PartnerCode should be present even with error")
					assert.NotEmpty(t, partner.PartnerName, "PartnerName should be present even with error")

					// Services and capabilities should be empty for errors
					assert.Empty(t, partner.Services, "Services should be empty for validation errors")
					assert.Empty(t, partner.Capabilities, "Capabilities should be empty for validation errors")
				}
			},
			description: "Partner with error should maintain structure integrity",
		},
		{
			name: "PartnerV2Response_MixedResults",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl", "shipyaari", "smile_ecom"})
				factory.SetupMixedCapabilitiesAdapters()
				// Add a non-serviceable partner
				smileEcomAdapter := mocks.NewMockPartnerAdapter("smile_ecom")
				smileEcomAdapter.SetServiceabilityResult(smileEcomAdapter.CreateNonServiceableResult())
				factory.SetAdapter("smile_ecom", smileEcomAdapter)
			},
			expectedPartnersCount: 3,
			validatePartnerFields: func(t *testing.T, partner *models.PartnerV2Response) {
				// All partners should have valid structure regardless of serviceability
				assert.NotEmpty(t, partner.PartnerID, "PartnerID should be present")
				assert.NotEmpty(t, partner.PartnerCode, "PartnerCode should be present")
				assert.NotEmpty(t, partner.PartnerName, "PartnerName should be present")

				switch partner.PartnerCode {
				case "dhl":
					// DHL with flattened capabilities
					assert.Contains(t, partner.Capabilities, "next_business_day", "DHL should have flattened capabilities")
				case "shipyaari":
					// Traditional partner with nested capabilities
					assert.Contains(t, partner.Capabilities, "cod_available", "Shipyaari should have traditional capabilities")
				case "smile_ecom":
					// Non-serviceable partner
					assert.NotNil(t, partner.Error, "Non-serviceable partner should have error message")
					assert.Contains(t, *partner.Error, "Not serviceable", "Error should indicate non-serviceability")
				}
			},
			description: "Mixed partner results should all maintain proper structure",
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
				false, // Return all partners to validate structure
			)

			// Create test request
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

			// Verify response structure
			require.NoError(t, err, "Should not return error: %s", tt.description)
			require.NotNil(t, response, "Response should not be nil")
			assert.Len(t, response.Partners, tt.expectedPartnersCount, "Should have expected number of partners: %s", tt.description)

			// Validate each partner's structure
			for i := range response.Partners {
				partner := &response.Partners[i]
				tt.validatePartnerFields(t, partner)
			}
		})
	}
}

// TestPartnerIDMapping tests that PartnerID is correctly mapped from UUID to string
func TestPartnerIDMapping(t *testing.T) {
	t.Parallel()

	// Setup
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()
	factory.SetSupportedPartners([]string{"dhl", "shipyaari"})
	factory.SetupDefaultAdapters()

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

	// Verify PartnerID mapping for each partner
	for _, partner := range response.Partners {
		// PartnerID should be string representation of UUID
		assert.IsType(t, "", partner.PartnerID, "PartnerID should be string type")
		assert.NotEmpty(t, partner.PartnerID, "PartnerID should not be empty")

		// Should be valid UUID format (36 characters with hyphens)
		assert.Len(t, partner.PartnerID, 36, "PartnerID should be valid UUID string length")
		assert.Contains(t, partner.PartnerID, "-", "PartnerID should be valid UUID format with hyphens")

		// Should be different for each partner
		for _, otherPartner := range response.Partners {
			if partner.PartnerCode != otherPartner.PartnerCode {
				assert.NotEqual(t, partner.PartnerID, otherPartner.PartnerID,
					"Different partners should have different PartnerIDs")
			}
		}
	}
}

// TestPartnerNameField tests the new PartnerName field in responses
func TestPartnerNameField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupMocks   func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		partnerCode  string
		expectedName string
		description  string
	}{
		{
			name: "DHL_PartnerName",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLAdapters()
			},
			partnerCode:  "dhl",
			expectedName: "Mock dhl",
			description:  "DHL should have correct partner name",
		},
		{
			name: "Shipyaari_PartnerName",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"shipyaari"})
				factory.SetupDefaultAdapters()
			},
			partnerCode:  "shipyaari",
			expectedName: "Mock shipyaari",
			description:  "Shipyaari should have correct partner name",
		},
		{
			name: "SmileEcom_PartnerName",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"smile_ecom"})
				factory.SetupDefaultAdapters()
			},
			partnerCode:  "smile_ecom",
			expectedName: "Mock smile_ecom",
			description:  "Smile Ecom should have correct partner name",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			factory := mocks.NewMockPartnerAdapterFactory()
			repo := mocks.NewMockPartnerAttributeMapRepository()
			tt.setupMocks(factory, repo)

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

			// Find the target partner
			var targetPartner *models.PartnerV2Response
			for i := range response.Partners {
				if response.Partners[i].PartnerCode == tt.partnerCode {
					targetPartner = &response.Partners[i]
					break
				}
			}

			require.NotNil(t, targetPartner, "Target partner should be found: %s", tt.description)

			// Verify PartnerName field
			assert.Equal(t, tt.expectedName, targetPartner.PartnerName,
				"PartnerName should match expected: %s", tt.description)
			assert.NotEmpty(t, targetPartner.PartnerName, "PartnerName should not be empty")
			assert.Contains(t, targetPartner.PartnerName, tt.partnerCode,
				"PartnerName should contain partner code")
		})
	}
}

// TestPartnerResponseConsistency tests consistency of partner response structure across different scenarios
func TestPartnerResponseConsistency(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name        string
		setupMocks  func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		description string
	}{
		{
			name: "ServiceablePartners",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl", "shipyaari"})
				factory.SetupDefaultAdapters()
			},
			description: "Serviceable partners should have consistent structure",
		},
		{
			name: "NonServiceablePartners",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"smile_ecom"})
				smileEcomAdapter := mocks.NewMockPartnerAdapter("smile_ecom")
				smileEcomAdapter.SetServiceabilityResult(smileEcomAdapter.CreateNonServiceableResult())
				factory.SetAdapter("smile_ecom", smileEcomAdapter)
			},
			description: "Non-serviceable partners should have consistent structure",
		},
		{
			name: "ErrorPartners",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			description: "Error partners should have consistent structure",
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			factory := mocks.NewMockPartnerAdapterFactory()
			repo := mocks.NewMockPartnerAttributeMapRepository()
			scenario.setupMocks(factory, repo)

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

			// Verify structure consistency for all partners
			for _, partner := range response.Partners {
				// Core fields should always be present
				assert.NotEmpty(t, partner.PartnerID, "PartnerID should always be present: %s", scenario.description)
				assert.NotEmpty(t, partner.PartnerCode, "PartnerCode should always be present: %s", scenario.description)
				assert.NotEmpty(t, partner.PartnerName, "PartnerName should always be present: %s", scenario.description)
				assert.GreaterOrEqual(t, partner.Rating, float64(0), "Rating should be non-negative: %s", scenario.description)
				assert.Greater(t, partner.ResponseTime, time.Duration(0), "ResponseTime should be positive: %s", scenario.description)

				// Arrays should be initialized (not nil)
				assert.NotNil(t, partner.Services, "Services should be initialized: %s", scenario.description)
				assert.NotNil(t, partner.Capabilities, "Capabilities should be initialized: %s", scenario.description)
			}
		})
	}
}
