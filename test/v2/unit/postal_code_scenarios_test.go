package unit

import (
	"context"
	"testing"
	"time"

	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostalCodeScenarios tests different postal code configurations
func TestPostalCodeScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		request       *models.ServiceabilityV2Request
		expectedValid bool
		expectedError string
	}{
		{
			name: "Single postal code - valid",
			request: &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr("110001"),
			},
			expectedValid: true,
		},
		{
			name: "Source and destination postal codes - valid",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("110002"),
			},
			expectedValid: true,
		},
		{
			name: "Postal code as destination with source - valid",
			request: &models.ServiceabilityV2Request{
				CountryCode:      stringPtr("IN"),
				PostalCode:       stringPtr("110002"),
				SourcePostalCode: stringPtr("110001"),
			},
			expectedValid: true,
		},
		{
			name: "No postal codes - invalid",
			request: &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
			},
			expectedValid: false,
			expectedError: "postal_code or source_postal_code/destination_postal_code",
		},
		{
			name: "Empty postal code - invalid",
			request: &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr(""),
			},
			expectedValid: false,
			expectedError: "postal_code or source_postal_code/destination_postal_code",
		},
		{
			name: "Single postal code with source and destination - invalid",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				PostalCode:            stringPtr("110001"),
				SourcePostalCode:      stringPtr("110002"),
				DestinationPostalCode: stringPtr("110003"),
			},
			expectedValid: false,
			expectedError: "provide either postal_code only or source/destination postal codes, not both",
		},
		{
			name: "Postal code as destination with source and destination - invalid",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				PostalCode:            stringPtr("110001"),
				SourcePostalCode:      stringPtr("110002"),
				DestinationPostalCode: stringPtr("110003"),
			},
			expectedValid: false,
			expectedError: "provide either postal_code only or source/destination postal codes, not both",
		},

		{
			name: "Only source postal code - invalid",
			request: &models.ServiceabilityV2Request{
				CountryCode:      stringPtr("IN"),
				SourcePostalCode: stringPtr("110001"),
			},
			expectedValid: false,
			expectedError: "postal_code or source_postal_code/destination_postal_code",
		},
		{
			name: "Only destination postal code - invalid",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				DestinationPostalCode: stringPtr("110001"),
			},
			expectedValid: false,
			expectedError: "postal_code or source_postal_code/destination_postal_code",
		},
		{
			name: "Empty source postal code - invalid",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				SourcePostalCode:      stringPtr(""),
				DestinationPostalCode: stringPtr("110001"),
			},
			expectedValid: false,
			expectedError: "postal_code or source_postal_code/destination_postal_code",
		},
		{
			name: "Empty destination postal code - invalid",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr(""),
			},
			expectedValid: false,
			expectedError: "postal_code or source_postal_code/destination_postal_code",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()

			// Setup a simple serviceable partner
			if tt.expectedValid {
				mockFactory.SetSupportedPartners([]string{"shipyaari"})
				adapter := mocks.NewMockPartnerAdapter("shipyaari")
				adapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerCode:   "shipyaari",
					
					Services:      []models.ServiceV2{{ServiceCode: "SDD", ServiceName: "Same Day Delivery"}},
				})
				mockFactory.SetAdapter("shipyaari", adapter)
			}

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				5*time.Second,
				false,
			)

			// Execute
			response, err := orchestrator.CheckServiceability(context.Background(), tt.request)

			if tt.expectedValid {
				// Valid request should succeed
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.True(t, response.Success)
				assert.Len(t, response.Partners, 1)
				assert.Equal(t, "shipyaari", response.Partners[0].PartnerCode)
				assert.True(t, len(response.Partners[0].Services) > 0)
			} else {
				// Invalid request should fail with validation error
				require.Error(t, err)
				require.Nil(t, response)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

// TestPostalCodeWithPackageInformation tests postal code scenarios with package information
func TestPostalCodeWithPackageInformation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request *models.ServiceabilityV2Request
		verify  func(t *testing.T, response *models.ServiceabilityV2Response)
	}{
		{
			name: "Single postal code with package info",
			request: &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr("110001"),
				Package: &models.Package{
					Weight: &models.Weight{
						Value: 1.5,
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
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.True(t, response.Success)
				assert.Len(t, response.Partners, 1)
				assert.True(t, len(response.Partners[0].Services) > 0)
			},
		},
		{
			name: "Source/destination postal codes with package info",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("110002"),
				Package: &models.Package{
					Weight: &models.Weight{
						Value: 2.0,
						Unit:  "kg",
					},
				},
			},
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.True(t, response.Success)
				assert.Len(t, response.Partners, 1)
				assert.True(t, len(response.Partners[0].Services) > 0)
			},
		},
		{
			name: "Postal code as destination with package info",
			request: &models.ServiceabilityV2Request{
				CountryCode:      stringPtr("IN"),
				PostalCode:       stringPtr("110002"),
				SourcePostalCode: stringPtr("110001"),
				Package: &models.Package{
					Dimensions: &models.Dimensions{
						Length: 15.0,
						Width:  12.0,
						Height: 8.0,
						Unit:   "cm",
					},
				},
			},
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.True(t, response.Success)
				assert.Len(t, response.Partners, 1)
				assert.True(t, len(response.Partners[0].Services) > 0)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()

			// Setup a serviceable partner
			mockFactory.SetSupportedPartners([]string{"shipyaari"})
			adapter := mocks.NewMockPartnerAdapter("shipyaari")
			adapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
				PartnerCode:   "shipyaari",
				
				Services:      []models.ServiceV2{{ServiceCode: "SDD", ServiceName: "Same Day Delivery"}},
			})
			mockFactory.SetAdapter("shipyaari", adapter)

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				5*time.Second,
				false,
			)

			// Execute
			response, err := orchestrator.CheckServiceability(context.Background(), tt.request)

			// Verify
			require.NoError(t, err)
			require.NotNil(t, response)
			tt.verify(t, response)
		})
	}
}

// TestPostalCodeWithParcelCategory tests postal code scenarios with parcel category filtering
func TestPostalCodeWithParcelCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		request         *models.ServiceabilityV2Request
		setupMocks      func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectedCount   int
		expectedSuccess bool
	}{
		{
			name: "Single postal code with parcel category",
			request: &models.ServiceabilityV2Request{
				CountryCode:    stringPtr("IN"),
				PostalCode:     stringPtr("110001"),
				ParcelCategory: stringPtr("ecomm"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"shipyaari", "smile_ecom"})

				// Setup serviceable partners
				adapter1 := mocks.NewMockPartnerAdapter("shipyaari")
				adapter1.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerCode:   "shipyaari",
					
					Services:      []models.ServiceV2{{ServiceCode: "SDD", ServiceName: "Same Day Delivery"}},
				})
				factory.SetAdapter("shipyaari", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("smile_ecom")
				adapter2.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerCode:   "smile_ecom",
					
					Services:      []models.ServiceV2{{ServiceCode: "NDD", ServiceName: "Next Day Delivery"}},
				})
				factory.SetAdapter("smile_ecom", adapter2)
			},
			expectedCount:   2,
			expectedSuccess: true,
		},
		{
			name: "Source/destination postal codes with parcel category",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("110002"),
				ParcelCategory:        stringPtr("courier"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"smile_courier"})

				adapter := mocks.NewMockPartnerAdapter("smile_courier")
				adapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerCode:   "smile_courier",
					
					Services:      []models.ServiceV2{{ServiceCode: "NDD", ServiceName: "Next Day Delivery"}},
				})
				factory.SetAdapter("smile_courier", adapter)
			},
			expectedCount:   1,
			expectedSuccess: true,
		},
		{
			name: "No serviceable partners for parcel category",
			request: &models.ServiceabilityV2Request{
				CountryCode:    stringPtr("IN"),
				PostalCode:     stringPtr("110001"),
				ParcelCategory: stringPtr("hyperlocal"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"shipyaari"})

				adapter := mocks.NewMockPartnerAdapter("shipyaari")
				adapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerCode:   "shipyaari",
					
					Services:      []models.ServiceV2{},
				})
				factory.SetAdapter("shipyaari", adapter)
			},
			expectedCount:   0,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()
			tt.setupMocks(mockFactory, mockRepo)

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				5*time.Second,
				false,
			)

			// Execute
			response, err := orchestrator.CheckServiceability(context.Background(), tt.request)

			// Verify
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tt.expectedSuccess, response.Success)

			if tt.expectedSuccess {
				assert.Len(t, response.Partners, tt.expectedCount)
				for _, partner := range response.Partners {
					assert.True(t, len(partner.Services) > 0)
				}
			} else {
				// When success=false, check that partners are returned but not serviceable
				assert.NotNil(t, response.Metadata)
				assert.Equal(t, 0, response.Metadata.ServiceableCount)
			}
		})
	}
}

// TestPostalCodeEdgeCases tests edge cases for postal code scenarios
func TestPostalCodeEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		request       *models.ServiceabilityV2Request
		expectedError string
	}{
		{
			name: "Nil postal code pointers",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				PostalCode:            nil,
				SourcePostalCode:      nil,
				DestinationPostalCode: nil,
			},
			expectedError: "postal_code or source_postal_code/destination_postal_code",
		},
		{
			name: "Mixed nil and empty postal codes",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				PostalCode:            nil,
				SourcePostalCode:      stringPtr(""),
				DestinationPostalCode: stringPtr("110001"),
			},
			expectedError: "postal_code or source_postal_code/destination_postal_code",
		},
		{
			name: "All empty postal codes",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				PostalCode:            stringPtr(""),
				SourcePostalCode:      stringPtr(""),
				DestinationPostalCode: stringPtr(""),
			},
			expectedError: "postal_code or source_postal_code/destination_postal_code",
		},
		{
			name: "Whitespace-only postal codes",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				PostalCode:            stringPtr("   "),
				SourcePostalCode:      stringPtr("  "),
				DestinationPostalCode: stringPtr(" "),
			},
			expectedError: "postal_code or source_postal_code/destination_postal_code",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				5*time.Second,
				false,
			)

			// Execute
			response, err := orchestrator.CheckServiceability(context.Background(), tt.request)

			// Verify error
			require.Error(t, err)
			require.Nil(t, response)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// Helper function to create string pointers (duplicate removed - using shared function)
