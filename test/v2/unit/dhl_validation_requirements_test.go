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

// TestDHLValidationRequirements tests DHL-specific validation requirements
func TestDHLValidationRequirements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		request       *models.ServiceabilityV2Request
		setupMocks    func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name: "DHL_ValidInternationalRequest",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:          stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.5,
							Unit:  "kg",
						},
						Dimensions: &models.Dimensions{
							Length: 25.0,
							Width:  15.0,
							Height: 10.0,
							Unit:   "cm",
						},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLAdapters()
			},
			expectError: false,
			description: "Valid DHL request with all required fields should succeed",
		},
		{
			name: "DHL_MissingSourcePostalCode",
			request: &models.ServiceabilityV2Request{
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:          stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			expectError:   true,
			errorContains: "source postal code is required for DHL shipments",
			description:   "DHL requires source postal code",
		},
		{
			name: "DHL_MissingDestinationPostalCode",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode: stringPtr("560001"),
				CountryCode:     stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			expectError:   true,
			errorContains: "destination postal code is required for DHL shipments",
			description:   "DHL requires destination postal code",
		},
		{
			name: "DHL_EmptySourcePostalCode",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr(""),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:          stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			expectError:   true,
			errorContains: "source postal code is required for DHL shipments",
			description:   "DHL requires non-empty source postal code",
		},
		{
			name: "DHL_EmptyDestinationPostalCode",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr(""),
				CountryCode:          stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			expectError:   true,
			errorContains: "destination postal code is required for DHL shipments",
			description:   "DHL requires non-empty destination postal code",
		},
		{
			name: "DHL_MissingPackages",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:          stringPtr("IN"),
				Packages:             []models.Package{}, // Empty packages
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			expectError:   true,
			errorContains: "at least one package is required for DHL shipments",
			description:   "DHL requires at least one package",
		},
		{
			name: "DHL_MissingPackageWeight",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:          stringPtr("IN"),
				Packages: []models.Package{
					{
						// Weight: nil, // Missing weight
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			expectError:   true,
			errorContains: "package weight is required for package 1 in DHL shipments",
			description:   "DHL requires package weight",
		},
		{
			name: "DHL_InvalidPackageWeight",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:          stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{Value: 0, Unit: "kg"}, // Invalid weight
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			expectError:   true,
			errorContains: "package weight must be greater than 0 for package 1 in DHL shipments",
			description:   "DHL requires positive package weight",
		},
		{
			name: "DHL_MissingPackageDimensions",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:          stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{Value: 1.5, Unit: "kg"},
						// Dimensions: nil, // Missing dimensions
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			expectError:   true,
			errorContains: "package dimensions are required for package 1 in DHL shipments",
			description:   "DHL requires package dimensions",
		},
		{
			name: "DHL_InvalidPackageDimensions",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:          stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 0, Width: 15.0, Height: 10.0, Unit: "cm"}, // Invalid length
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			expectError:   true,
			errorContains: "package dimensions must be greater than 0 for package 1 in DHL shipments",
			description:   "DHL requires positive package dimensions",
		},
		{
			name: "DHL_MultiplePackagesValidation",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:          stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
					{
						Weight: &models.Weight{Value: 2.0, Unit: "kg"},
						// Dimensions: nil, // Second package missing dimensions
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLValidationErrorAdapters()
			},
			expectError:   true,
			errorContains: "package dimensions are required for package 2 in DHL shipments",
			description:   "DHL validates all packages in request",
		},
		{
			name: "DHL_MultiplePackagesAllValid",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:          stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
					{
						Weight: &models.Weight{Value: 2.0, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 30.0, Width: 20.0, Height: 15.0, Unit: "cm"},
					},
					{
						Weight: &models.Weight{Value: 0.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 15.0, Width: 10.0, Height: 5.0, Unit: "cm"},
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLAdapters()
			},
			expectError: false,
			description: "DHL should accept multiple valid packages",
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
				false, // Return all partners to see errors
			)

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, tt.request)

			// Verify
			require.NoError(t, err, "Orchestrator should not return error: %s", tt.description)
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

			if tt.expectError {
				// Verify error is present in partner response
				assert.NotNil(t, dhlPartner.Error, "DHL partner should have error: %s", tt.description)
				if tt.errorContains != "" {
					assert.Contains(t, *dhlPartner.Error, tt.errorContains, 
						"Error should contain expected message: %s", tt.description)
				}
				// Services should be empty for validation errors
				assert.Empty(t, dhlPartner.Services, "DHL should have no services for validation errors")
			} else {
				// Verify no error
				if dhlPartner.Error != nil {
					assert.Nil(t, dhlPartner.Error, "DHL partner should not have error: %s (got: %s)", 
						tt.description, *dhlPartner.Error)
				}
				// Should have capabilities for successful validation
				assert.NotEmpty(t, dhlPartner.Capabilities, "DHL should have capabilities for valid requests")
			}
		})
	}
}

// TestDHLValidationErrorMetadata tests that DHL validation errors include proper metadata
func TestDHLValidationErrorMetadata(t *testing.T) {
	t.Parallel()

	// Setup
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()
	factory.SetSupportedPartners([]string{"dhl"})
	factory.SetupDHLValidationErrorAdapters()

	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		factory,
		repo,
		30*time.Second,
		false,
	)

	// Request with validation error
	request := &models.ServiceabilityV2Request{
		SourcePostalCode: stringPtr("560001"),
		// DestinationPostalCode: nil, // Missing destination
		CountryCode: stringPtr("IN"),
		Packages: []models.Package{
			{
				Weight: &models.Weight{Value: 1.5, Unit: "kg"},
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
	require.NotNil(t, dhlPartner.Error)

	// Verify error message format
	assert.Contains(t, *dhlPartner.Error, "DHL validation failed", "Error should indicate DHL validation failure")
	assert.Contains(t, *dhlPartner.Error, "destination postal code is required", "Error should specify validation issue")

	// Verify response structure for validation errors
	assert.Empty(t, dhlPartner.Services, "Should have no services for validation errors")
	assert.Empty(t, dhlPartner.Capabilities, "Should have no capabilities for validation errors")
	assert.Greater(t, dhlPartner.ResponseTime, time.Duration(0), "Should have positive response time")
} 