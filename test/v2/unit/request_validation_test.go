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

// TestValidateV2Request tests the validateV2Request method
func TestValidateV2Request(t *testing.T) {
	t.Parallel()

	// Setup orchestrator with mocks
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()
	mockFactory.SetupDefaultAdapters()
	mockRepo.SetupDefaultData()

	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		mockFactory,
		mockRepo,
		30*time.Second,
		true,
	)

	tests := []struct {
		name          string
		request       *models.ServiceabilityV2Request
		expectedError bool
		errorContains string
		description   string
	}{
		// Valid cases
		{
			name: "Valid single postal code",
			request: &models.ServiceabilityV2Request{
				PostalCode: stringPtr("110001"),
			},
			expectedError: false,
			description:   "Should accept single postal code",
		},
		{
			name: "Valid source and destination postal codes",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("560001"),
			},
			expectedError: false,
			description:   "Should accept source and destination postal codes",
		},
		{
			name: "Valid postal code as destination with source",
			request: &models.ServiceabilityV2Request{
				PostalCode:       stringPtr("560001"),
				SourcePostalCode: stringPtr("110001"),
			},
			expectedError: false,
			description:   "Should accept postal code as destination with source postal code",
		},

		// Invalid cases - Missing postal codes
		{
			name:          "No postal codes provided",
			request:       &models.ServiceabilityV2Request{},
			expectedError: true,
			errorContains: "postal_code or source_postal_code/destination_postal_code",
			description:   "Should reject request with no postal codes",
		},
		{
			name: "Empty postal code",
			request: &models.ServiceabilityV2Request{
				PostalCode: stringPtr(""),
			},
			expectedError: true,
			errorContains: "postal_code or source_postal_code/destination_postal_code",
			description:   "Should reject request with empty postal code",
		},
		{
			name: "Only source postal code",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode: stringPtr("110001"),
			},
			expectedError: true,
			errorContains: "postal_code or source_postal_code/destination_postal_code",
			description:   "Should reject request with only source postal code",
		},
		{
			name: "Only destination postal code",
			request: &models.ServiceabilityV2Request{
				DestinationPostalCode: stringPtr("560001"),
			},
			expectedError: true,
			errorContains: "postal_code or source_postal_code/destination_postal_code",
			description:   "Should reject request with only destination postal code",
		},
		{
			name: "Empty source postal code",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr(""),
				DestinationPostalCode: stringPtr("560001"),
			},
			expectedError: true,
			errorContains: "postal_code or source_postal_code/destination_postal_code",
			description:   "Should reject request with empty source postal code",
		},
		{
			name: "Empty destination postal code",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr(""),
			},
			expectedError: true,
			errorContains: "postal_code or source_postal_code/destination_postal_code",
			description:   "Should reject request with empty destination postal code",
		},

		// Invalid cases - Conflicting configurations
		{
			name: "Both single and source/destination postal codes",
			request: &models.ServiceabilityV2Request{
				PostalCode:            stringPtr("110001"),
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("560001"),
			},
			expectedError: true,
			errorContains: "provide either postal_code only or source/destination postal codes, not both",
			description:   "Should reject request with both single and source/destination postal codes",
		},
		{
			name: "Postal code as destination with source and destination",
			request: &models.ServiceabilityV2Request{
				PostalCode:            stringPtr("560001"),
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("560001"),
			},
			expectedError: true,
			errorContains: "provide either postal_code only or source/destination postal codes, not both",
			description:   "Should reject postal code as destination with both source and destination postal codes",
		},

		// Edge cases
		{
			name:          "Nil request",
			request:       nil,
			expectedError: true,
			description:   "Should handle nil request gracefully",
		},
		{
			name: "All postal codes nil",
			request: &models.ServiceabilityV2Request{
				PostalCode:            nil,
				SourcePostalCode:      nil,
				DestinationPostalCode: nil,
			},
			expectedError: true,
			errorContains: "postal_code or source_postal_code/destination_postal_code",
			description:   "Should reject request with all postal codes nil",
		},
		{
			name: "Valid request with additional fields",
			request: &models.ServiceabilityV2Request{
				PostalCode:     stringPtr("110001"),
				ParcelCategory: stringPtr("international"),
				CountryCode:    stringPtr("US"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.0,
							Unit:  "kg",
						},
						Dimensions: &models.Dimensions{
							Length: 10.0,
							Width:  10.0,
							Height: 10.0,
							Unit:   "cm",
						},
					},
				},
			},
			expectedError: false,
			description:   "Should accept valid request with additional fields",
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call validation through CheckServiceability (since validateV2Request is not exposed)
			// We'll create a minimal context for testing
			ctx := context.Background()

			// Call the orchestrator method that uses validateV2Request
			response, err := orchestrator.CheckServiceability(ctx, tt.request)

			if tt.expectedError {
				assert.Error(t, err, "Expected error for: %s", tt.description)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains, "Error should contain expected message")
				}
				assert.Nil(t, response, "Response should be nil when error occurs")
			} else {
				// For valid requests, we might get other errors (like partner processing errors)
				// But we shouldn't get validation errors
				if err != nil {
					// If there's an error, it shouldn't be a validation error
					assert.NotContains(t, err.Error(), "postal_code or source_postal_code/destination_postal_code",
						"Should not get postal code validation error for valid request")
					assert.NotContains(t, err.Error(), "provide either postal_code only or source/destination postal codes",
						"Should not get conflicting postal code error for valid request")
				}
				// Response may be nil due to other processing errors, but that's fine for validation testing
			}
		})
	}
}

// TestValidateV2RequestPostalCodeCombinations tests various postal code combinations
func TestValidateV2RequestPostalCodeCombinations(t *testing.T) {
	t.Parallel()

	// Setup orchestrator with mocks
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()
	mockFactory.SetupDefaultAdapters()
	mockRepo.SetupDefaultData()

	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		mockFactory,
		mockRepo,
		30*time.Second,
		true,
	)

	ctx := context.Background()

	// Test all possible combinations of postal code fields
	combinations := []struct {
		name                  string
		postalCode            *string
		sourcePostalCode      *string
		destinationPostalCode *string
		expectedValid         bool
		description           string
	}{
		// Valid combinations
		{
			name:          "Single postal code",
			postalCode:    stringPtr("110001"),
			expectedValid: true,
			description:   "Only postal_code provided",
		},
		{
			name:                  "Source and destination",
			sourcePostalCode:      stringPtr("110001"),
			destinationPostalCode: stringPtr("560001"),
			expectedValid:         true,
			description:           "Both source_postal_code and destination_postal_code provided",
		},
		{
			name:             "Postal code as destination with source",
			postalCode:       stringPtr("560001"),
			sourcePostalCode: stringPtr("110001"),
			expectedValid:    true,
			description:      "postal_code as destination with source_postal_code",
		},

		// Invalid combinations
		{
			name:          "No postal codes",
			expectedValid: false,
			description:   "No postal codes provided",
		},
		{
			name:             "Only source",
			sourcePostalCode: stringPtr("110001"),
			expectedValid:    false,
			description:      "Only source_postal_code provided",
		},
		{
			name:                  "Only destination",
			destinationPostalCode: stringPtr("560001"),
			expectedValid:         false,
			description:           "Only destination_postal_code provided",
		},
		{
			name:                  "All three postal codes",
			postalCode:            stringPtr("110001"),
			sourcePostalCode:      stringPtr("110001"),
			destinationPostalCode: stringPtr("560001"),
			expectedValid:         false,
			description:           "All three postal codes provided (conflicting)",
		},
		{
			name:          "Empty postal code",
			postalCode:    stringPtr(""),
			expectedValid: false,
			description:   "Empty postal_code",
		},
		{
			name:             "Empty source postal code",
			sourcePostalCode: stringPtr(""),
			expectedValid:    false,
			description:      "Empty source_postal_code",
		},
		{
			name:                  "Empty destination postal code",
			destinationPostalCode: stringPtr(""),
			expectedValid:         false,
			description:           "Empty destination_postal_code",
		},
	}

	for _, combo := range combinations {
		combo := combo // capture loop variable
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			request := &models.ServiceabilityV2Request{
				PostalCode:            combo.postalCode,
				SourcePostalCode:      combo.sourcePostalCode,
				DestinationPostalCode: combo.destinationPostalCode,
			}

			_, err := orchestrator.CheckServiceability(ctx, request)

			if combo.expectedValid {
				// Should not get validation error for valid postal code combinations
				if err != nil {
					assert.NotContains(t, err.Error(), "postal_code or source_postal_code/destination_postal_code",
						"Should not get postal code validation error for valid combination: %s", combo.description)
					assert.NotContains(t, err.Error(), "provide either postal_code only or source/destination postal codes",
						"Should not get conflicting postal code error for valid combination: %s", combo.description)
					assert.NotContains(t, err.Error(), "when using postal_code as destination with source_postal_code",
						"Should not get destination conflict error for valid combination: %s", combo.description)
				}
			} else {
				// Should get validation error for invalid postal code combinations
				require.Error(t, err, "Expected validation error for invalid combination: %s", combo.description)
			}
		})
	}
}

// TestValidateV2RequestNilValues tests validation with nil values
func TestValidateV2RequestNilValues(t *testing.T) {
	t.Parallel()

	// Setup orchestrator with mocks
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()
	mockFactory.SetupDefaultAdapters()
	mockRepo.SetupDefaultData()

	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		mockFactory,
		mockRepo,
		30*time.Second,
		true,
	)

	ctx := context.Background()

	// Test nil request
	t.Run("Nil request", func(t *testing.T) {
		_, err := orchestrator.CheckServiceability(ctx, nil)
		assert.Error(t, err, "Should return error for nil request")
	})

	// Test request with all nil postal codes
	t.Run("All nil postal codes", func(t *testing.T) {
		request := &models.ServiceabilityV2Request{
			PostalCode:            nil,
			SourcePostalCode:      nil,
			DestinationPostalCode: nil,
		}

		_, err := orchestrator.CheckServiceability(ctx, request)
		assert.Error(t, err, "Should return error for request with all nil postal codes")
		assert.Contains(t, err.Error(), "postal_code or source_postal_code/destination_postal_code",
			"Should return postal code validation error")
	})
}

// Helper function to create string pointers
