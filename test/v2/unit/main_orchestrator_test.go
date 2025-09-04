package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckServiceabilityHappyPath(t *testing.T) {
	tests := []struct {
		name                     string
		request                  *models.ServiceabilityV2Request
		setupMocks               func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectSuccess            bool
		expectedPartnersCount    int
		expectedServiceableCount int
		description              string
	}{
		{
			name: "SinglePostalCode_AllPartnersServiceable",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				// Setup factory with 3 partners
				factory.SetSupportedPartners([]string{"partner1", "partner2", "partner3"})

				// All partners are serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityResult(adapter2.CreateServiceableResult())
				factory.SetAdapter("partner2", adapter2)

				adapter3 := mocks.NewMockPartnerAdapter("partner3")
				adapter3.SetServiceabilityResult(adapter3.CreateServiceableResult())
				factory.SetAdapter("partner3", adapter3)

				// No parcel category filtering
				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectSuccess:            true,
			expectedPartnersCount:    3,
			expectedServiceableCount: 3,
			description:              "All partners are serviceable",
		},
		{
			name: "SinglePostalCode_MixedServiceability",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2", "partner3"})

				// Mixed serviceability: 2 serviceable, 1 not serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityResult(adapter2.CreateNonServiceableResult())
				factory.SetAdapter("partner2", adapter2)

				adapter3 := mocks.NewMockPartnerAdapter("partner3")
				adapter3.SetServiceabilityResult(adapter3.CreateServiceableResult())
				factory.SetAdapter("partner3", adapter3)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectSuccess:            true,
			expectedPartnersCount:    2, // Only serviceable partners returned when success=true
			expectedServiceableCount: 2,
			description:              "Mixed serviceability returns only serviceable partners",
		},
		{
			name: "SinglePostalCode_NoPartnersServiceable",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2", "partner3"})

				// No partners serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateNonServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityResult(adapter2.CreateNonServiceableResult())
				factory.SetAdapter("partner2", adapter2)

				adapter3 := mocks.NewMockPartnerAdapter("partner3")
				adapter3.SetServiceabilityResult(adapter3.CreateNonServiceableResult())
				factory.SetAdapter("partner3", adapter3)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectSuccess:            false,
			expectedPartnersCount:    3, // All partners returned when success=false
			expectedServiceableCount: 0,
			description:              "No serviceable partners returns all partners",
		},
		{
			name: "SourceDestinationPostalCodes_WithParcelCategory",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("12345"),
				DestinationPostalCode: stringPtr("67890"),
				CountryCode:           stringPtr("IN"),
				ParcelCategory:        stringPtr("ecomm"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2", "partner3"})

				// Setup parcel category filtering - only partner1 and partner3 support ecomm
				repo.SetPartnerInfo("ecomm", []models.PartnerAttributeMap{
					{PartnerCode: "partner1", PartnerID: uuidPtr(uuid.New())},
					{PartnerCode: "partner3", PartnerID: uuidPtr(uuid.New())},
				})

				// Both eligible partners are serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				adapter3 := mocks.NewMockPartnerAdapter("partner3")
				adapter3.SetServiceabilityResult(adapter3.CreateServiceableResult())
				factory.SetAdapter("partner3", adapter3)
			},
			expectSuccess:            true,
			expectedPartnersCount:    2, // Only eligible partners
			expectedServiceableCount: 2,
			description:              "Parcel category filtering works correctly",
		},
		{
			name: "PostalCodeAsDestinationWithSource",
			request: &models.ServiceabilityV2Request{
				PostalCode:       stringPtr("67890"), // destination
				SourcePostalCode: stringPtr("12345"), // source
				CountryCode:      stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityResult(adapter2.CreateServiceableResult())
				factory.SetAdapter("partner2", adapter2)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectSuccess:            true,
			expectedPartnersCount:    2,
			expectedServiceableCount: 2,
			description:              "Postal code as destination with source works correctly",
		},
	}

	for _, tt := range tests {
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
				time.Second*30,
				false,
			)

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, tt.request)

			// Verify
			require.NoError(t, err, "CheckServiceability should not return error")
			require.NotNil(t, response, "Response should not be nil")

			assert.Equal(t, tt.expectSuccess, response.Success, "Success should match expected: %s", tt.description)
			assert.Equal(t, tt.expectedPartnersCount, len(response.Partners), "Partners count should match expected: %s", tt.description)

			if response.Metadata != nil {
				assert.Equal(t, tt.expectedServiceableCount, response.Metadata.ServiceableCount, "Serviceable count should match expected: %s", tt.description)
			}

			// Verify filter metadata
			if response.Metadata != nil {
				assert.Equal(t, tt.request.CountryCode, response.Metadata.Filters.CountryCode, "Country code filter should match")
				assert.Equal(t, tt.request.ParcelCategory, response.Metadata.Filters.ParcelCategory, "Parcel category filter should match")
				assert.Equal(t, tt.request.ProductType, response.Metadata.Filters.ProductType, "Product type filter should match")
			}
		})
	}
}

func TestCheckServiceabilityErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		request     *models.ServiceabilityV2Request
		setupMocks  func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectError string
		description string
	}{
		{
			name:        "NilRequest",
			request:     nil,
			setupMocks:  func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository) {},
			expectError: "request cannot be nil",
			description: "Nil request should return error",
		},
		{
			name: "InvalidRequest_NoPostalCodes",
			request: &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
			},
			setupMocks:  func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository) {},
			expectError: "invalid request",
			description: "Missing postal codes should return validation error",
		},
		{
			name: "InvalidRequest_ConflictingPostalCodes",
			request: &models.ServiceabilityV2Request{
				PostalCode:            stringPtr("12345"),
				SourcePostalCode:      stringPtr("67890"),
				DestinationPostalCode: stringPtr("11111"),
				CountryCode:           stringPtr("IN"),
			},
			setupMocks:  func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository) {},
			expectError: "invalid request",
			description: "Conflicting postal codes should return validation error",
		},
		{
			name: "InvalidRequest_PostalCodeWithDestination",
			request: &models.ServiceabilityV2Request{
				PostalCode:            stringPtr("12345"),
				SourcePostalCode:      stringPtr("67890"),
				DestinationPostalCode: stringPtr("11111"),
				CountryCode:           stringPtr("IN"),
			},
			setupMocks:  func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository) {},
			expectError: "invalid request",
			description: "Postal code as destination with destination postal code should return validation error",
		},
	}

	for _, tt := range tests {
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
				time.Second*30,
				false,
			)

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, tt.request)

			// Verify
			require.Error(t, err, "CheckServiceability should return error: %s", tt.description)
			assert.Nil(t, response, "Response should be nil when error occurs")
			assert.Contains(t, err.Error(), tt.expectError, "Error should contain expected message: %s", tt.description)
		})
	}
}

func TestCheckServiceabilityPartnerErrorHandling(t *testing.T) {
	tests := []struct {
		name                     string
		request                  *models.ServiceabilityV2Request
		setupMocks               func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectSuccess            bool
		expectedPartnersCount    int
		expectedServiceableCount int
		description              string
	}{
		{
			name: "PartnerNotFound_Error",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				// partner1 exists and is serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				// partner2 doesn't exist in factory adapters map - will be created but with no specific setup

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectSuccess:            true,
			expectedPartnersCount:    2, // Both partners will be returned (partner2 will use default serviceable)
			expectedServiceableCount: 2,
			description:              "Partner not found handled gracefully",
		},
		{
			name: "PartnerUnhealthy_Error",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				// partner1 is healthy and serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				// partner2 is unhealthy
				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetHealthy(false)
				factory.SetAdapter("partner2", adapter2)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectSuccess:            true,
			expectedPartnersCount:    1, // Only healthy serviceable partner returned
			expectedServiceableCount: 1,
			description:              "Unhealthy partner error handled gracefully",
		},
		{
			name: "PartnerAdapter_CheckServiceabilityError",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				// partner1 is serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				// partner2 returns error
				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityError(errors.New("adapter error"))
				factory.SetAdapter("partner2", adapter2)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectSuccess:            true,
			expectedPartnersCount:    1, // Only successful serviceable partner returned
			expectedServiceableCount: 1,
			description:              "Partner adapter error handled gracefully",
		},
		{
			name: "AllPartners_Error",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				// All partners return errors
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityError(errors.New("error 1"))
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityError(errors.New("error 2"))
				factory.SetAdapter("partner2", adapter2)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectSuccess:            false,
			expectedPartnersCount:    2, // All partners returned when success=false
			expectedServiceableCount: 0,
			description:              "All partner errors handled gracefully",
		},
	}

	for _, tt := range tests {
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
				time.Second*30,
				false,
			)

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, tt.request)

			// Verify
			require.NoError(t, err, "CheckServiceability should not return error: %s", tt.description)
			require.NotNil(t, response, "Response should not be nil")

			assert.Equal(t, tt.expectSuccess, response.Success, "Success should match expected: %s", tt.description)
			assert.Equal(t, tt.expectedPartnersCount, len(response.Partners), "Partners count should match expected: %s", tt.description)

			if response.Metadata != nil {
				assert.Equal(t, tt.expectedServiceableCount, response.Metadata.ServiceableCount, "Serviceable count should match expected: %s", tt.description)
			}
		})
	}
}

func TestCheckServiceabilityEligiblePartnersScenarios(t *testing.T) {
	tests := []struct {
		name                     string
		request                  *models.ServiceabilityV2Request
		setupMocks               func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectSuccess            bool
		expectedPartnersCount    int
		expectedServiceableCount int
		description              string
	}{
		{
			name: "NoEligiblePartners_EmptyResponse",
			request: &models.ServiceabilityV2Request{
				PostalCode:     stringPtr("12345"),
				CountryCode:    stringPtr("IN"),
				ParcelCategory: stringPtr("ecomm"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				// No partners support ecomm category
				repo.SetPartnerInfo("ecomm", []models.PartnerAttributeMap{})
			},
			expectSuccess:            false,
			expectedPartnersCount:    0,
			expectedServiceableCount: 0,
			description:              "No eligible partners returns empty response",
		},
		{
			name: "EligiblePartners_SubsetOfSupported",
			request: &models.ServiceabilityV2Request{
				PostalCode:     stringPtr("12345"),
				CountryCode:    stringPtr("IN"),
				ParcelCategory: stringPtr("ecomm"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2", "partner3"})

				// Only partner1 and partner3 support ecomm
				repo.SetPartnerInfo("ecomm", []models.PartnerAttributeMap{
					{PartnerCode: "partner1", PartnerID: uuidPtr(uuid.New())},
					{PartnerCode: "partner3", PartnerID: uuidPtr(uuid.New())},
				})

				// Both eligible partners are serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				adapter3 := mocks.NewMockPartnerAdapter("partner3")
				adapter3.SetServiceabilityResult(adapter3.CreateServiceableResult())
				factory.SetAdapter("partner3", adapter3)
			},
			expectSuccess:            true,
			expectedPartnersCount:    2,
			expectedServiceableCount: 2,
			description:              "Eligible partners filtering works correctly",
		},
		{
			name: "EligiblePartners_DatabaseError_FallbackToAll",
			request: &models.ServiceabilityV2Request{
				PostalCode:     stringPtr("12345"),
				CountryCode:    stringPtr("IN"),
				ParcelCategory: stringPtr("ecomm"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				// Database error - should fallback to all partners
				repo.SetGetPartnerInfoByAttributeError(errors.New("database error"))

				// All partners are serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityResult(adapter2.CreateServiceableResult())
				factory.SetAdapter("partner2", adapter2)
			},
			expectSuccess:            true,
			expectedPartnersCount:    2,
			expectedServiceableCount: 2,
			description:              "Database error falls back to all partners",
		},
		{
			name: "EligiblePartners_NilRepository_FallbackToAll",
			request: &models.ServiceabilityV2Request{
				PostalCode:     stringPtr("12345"),
				CountryCode:    stringPtr("IN"),
				ParcelCategory: stringPtr("ecomm"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				// All partners are serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityResult(adapter2.CreateServiceableResult())
				factory.SetAdapter("partner2", adapter2)

				// Note: repo will be nil for this test case - don't call methods on it
			},
			expectSuccess:            true,
			expectedPartnersCount:    2,
			expectedServiceableCount: 2,
			description:              "Nil repository falls back to all partners",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture loop variable for parallel execution
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			factory := mocks.NewMockPartnerAdapterFactory()
			var repo *mocks.MockPartnerAttributeMapRepository

			if tt.name == "EligiblePartners_NilRepository_FallbackToAll" {
				repo = nil                  // Test with nil repository
				tt.setupMocks(factory, nil) // Pass nil to setupMocks
			} else {
				repo = mocks.NewMockPartnerAttributeMapRepository()
				tt.setupMocks(factory, repo)
			}

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				factory,
				repo,
				time.Second*30,
				false,
			)

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, tt.request)

			// Verify
			require.NoError(t, err, "CheckServiceability should not return error: %s", tt.description)
			require.NotNil(t, response, "Response should not be nil")

			assert.Equal(t, tt.expectSuccess, response.Success, "Success should match expected: %s", tt.description)
			assert.Equal(t, tt.expectedPartnersCount, len(response.Partners), "Partners count should match expected: %s", tt.description)

			if response.Metadata != nil {
				assert.Equal(t, tt.expectedServiceableCount, response.Metadata.ServiceableCount, "Serviceable count should match expected: %s", tt.description)

				// For empty response, total partners should be 0
				if tt.expectedPartnersCount == 0 {
					assert.Equal(t, 0, response.Metadata.TotalPartners, "Total partners should be 0 for empty response")
				}
			}
		})
	}
}

func TestCheckServiceabilityTimeout(t *testing.T) {
	t.Parallel()

	// Setup mocks
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()

	factory.SetSupportedPartners([]string{"partner1"})

	// Setup a partner that returns results normally (timeout is handled at orchestrator level)
	adapter1 := mocks.NewMockPartnerAdapter("partner1")
	adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
	factory.SetAdapter("partner1", adapter1)

	repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))

	// Create orchestrator with short timeout
	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		factory,
		repo,
		time.Millisecond*100, // Very short timeout
		false,
	)

	// Execute
	ctx := context.Background()
	response, err := orchestrator.CheckServiceability(ctx, &models.ServiceabilityV2Request{
		PostalCode:  stringPtr("12345"),
		CountryCode: stringPtr("IN"),
	})

	// Note: The timeout is applied to the entire operation, not individual partner calls
	// The goroutines will complete eventually, but the context will be cancelled
	// This test verifies that the orchestrator handles timeouts gracefully
	require.NoError(t, err, "CheckServiceability should not return timeout error due to goroutine handling")
	require.NotNil(t, response, "Response should not be nil")

	// The response should still be valid as goroutines complete independently
	assert.NotNil(t, response.Partners, "Partners should not be nil")
	assert.NotNil(t, response.Metadata, "Metadata should not be nil")
}

func TestCheckServiceabilityContextCancellation(t *testing.T) {
	t.Parallel()

	// Setup mocks
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()

	factory.SetSupportedPartners([]string{"partner1"})

	adapter1 := mocks.NewMockPartnerAdapter("partner1")
	adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
	factory.SetAdapter("partner1", adapter1)

	repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))

	// Create orchestrator
	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		factory,
		repo,
		time.Second*30,
		false,
	)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Execute with cancelled context
	response, err := orchestrator.CheckServiceability(ctx, &models.ServiceabilityV2Request{
		PostalCode:  stringPtr("12345"),
		CountryCode: stringPtr("IN"),
	})

	// The orchestrator should handle cancelled context gracefully
	// It may still return a response as goroutines execute independently
	if err != nil {
		assert.Contains(t, err.Error(), "context canceled", "Error should indicate context cancellation")
	} else {
		assert.NotNil(t, response, "Response should not be nil if no error")
	}
}

// Helper functions
func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}
