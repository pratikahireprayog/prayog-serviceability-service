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

func TestBuildV2Response(t *testing.T) {
	tests := []struct {
		name                     string
		request                  *models.ServiceabilityV2Request
		setupMocks               func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectedSuccess          bool
		expectedPartnersCount    int
		expectedServiceableCount int
		expectedTotalPartners    int
		verifyPartnerContent     func(t *testing.T, partners []models.PartnerV2Response)
		description              string
	}{
		{
			name: "AllPartnersServiceable_ReturnsOnlyServiceable",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2", "partner3"})

				// All partners serviceable with services
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				result1 := adapter1.CreateServiceableResult()
				result1.Services = []models.ServiceV2{
					{ServiceCode: "SDD", ServiceName: "Same Day Delivery", TATDays: 0},
				}
				result1.Capabilities = map[string]interface{}{
					"cod":       true,
					"insurance": true,
				}
				adapter1.SetServiceabilityResult(result1)
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				result2 := adapter2.CreateServiceableResult()
				result2.Services = []models.ServiceV2{
					{ServiceCode: "NDD", ServiceName: "Next Day Delivery", TATDays: 1},
				}
				adapter2.SetServiceabilityResult(result2)
				factory.SetAdapter("partner2", adapter2)

				adapter3 := mocks.NewMockPartnerAdapter("partner3")
				adapter3.SetServiceabilityResult(adapter3.CreateServiceableResult())
				factory.SetAdapter("partner3", adapter3)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedSuccess:          true,
			expectedPartnersCount:    3, // All serviceable partners returned
			expectedServiceableCount: 3,
			expectedTotalPartners:    3,
			verifyPartnerContent: func(t *testing.T, partners []models.PartnerV2Response) {
				// All partners should be serviceable
				for _, partner := range partners {
					assert.True(t, partner.IsServiceable, "Partner %s should be serviceable", partner.PartnerCode)
					assert.Nil(t, partner.Error, "Partner %s should not have error", partner.PartnerCode)
				}

				// Find partner1 and verify it has services
				partner1Found := false
				for _, partner := range partners {
					if partner.PartnerCode == "partner1" {
						partner1Found = true
						assert.Len(t, partner.Services, 1, "Partner1 should have services")
						assert.Equal(t, "SDD", partner.Services[0].ServiceCode)
						assert.Equal(t, "Same Day Delivery", partner.Services[0].ServiceName)
						assert.Equal(t, 0, partner.Services[0].TATDays)
						assert.True(t, partner.Capabilities["cod"].(bool))
						assert.True(t, partner.Capabilities["insurance"].(bool))
					}
				}
				assert.True(t, partner1Found, "Partner1 should be in response")
			},
			description: "When all partners are serviceable, returns only serviceable partners with services",
		},
		{
			name: "MixedServiceability_ReturnsOnlyServiceable",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2", "partner3"})

				// Mixed serviceability
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
			expectedSuccess:          true,
			expectedPartnersCount:    2, // Only serviceable partners returned
			expectedServiceableCount: 2,
			expectedTotalPartners:    3, // Total partners processed
			verifyPartnerContent: func(t *testing.T, partners []models.PartnerV2Response) {
				// Only serviceable partners should be returned
				for _, partner := range partners {
					assert.True(t, partner.IsServiceable, "Partner %s should be serviceable", partner.PartnerCode)
					assert.Nil(t, partner.Error, "Partner %s should not have error", partner.PartnerCode)
				}

				// Should not contain partner2 (non-serviceable)
				for _, partner := range partners {
					assert.NotEqual(t, "partner2", partner.PartnerCode, "Non-serviceable partner2 should not be in response")
				}
			},
			description: "When mixed serviceability, returns only serviceable partners",
		},
		{
			name: "NoPartnersServiceable_ReturnsAllPartners",
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
			expectedSuccess:          false,
			expectedPartnersCount:    3, // All partners returned when success=false
			expectedServiceableCount: 0,
			expectedTotalPartners:    3,
			verifyPartnerContent: func(t *testing.T, partners []models.PartnerV2Response) {
				// All partners should be non-serviceable
				for _, partner := range partners {
					assert.False(t, partner.IsServiceable, "Partner %s should not be serviceable", partner.PartnerCode)
					// Non-serviceable partners may have informational error messages
					if partner.Error != nil {
						assert.Contains(t, *partner.Error, "Not serviceable", "Error message should indicate non-serviceability")
					}
				}
			},
			description: "When no partners are serviceable, returns all partners",
		},
		{
			name: "PartnersWithErrors_ReturnsAllWithErrors",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2", "partner3"})

				// One serviceable, one error, one non-serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityError(errors.New("adapter timeout"))
				factory.SetAdapter("partner2", adapter2)

				adapter3 := mocks.NewMockPartnerAdapter("partner3")
				adapter3.SetServiceabilityResult(adapter3.CreateNonServiceableResult())
				factory.SetAdapter("partner3", adapter3)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedSuccess:          true,
			expectedPartnersCount:    1, // Only serviceable partner returned
			expectedServiceableCount: 1,
			expectedTotalPartners:    3,
			verifyPartnerContent: func(t *testing.T, partners []models.PartnerV2Response) {
				// Only serviceable partner should be returned
				assert.Len(t, partners, 1, "Should return only serviceable partner")
				assert.Equal(t, "partner1", partners[0].PartnerCode)
				assert.True(t, partners[0].IsServiceable)
				assert.Nil(t, partners[0].Error)
			},
			description: "When partners have errors, returns only serviceable partners",
		},
		{
			name: "AllPartnersError_ReturnsAllWithErrors",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				// All partners have errors
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityError(errors.New("timeout error"))
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityError(errors.New("network error"))
				factory.SetAdapter("partner2", adapter2)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedSuccess:          false,
			expectedPartnersCount:    2, // All partners returned when success=false
			expectedServiceableCount: 0,
			expectedTotalPartners:    2,
			verifyPartnerContent: func(t *testing.T, partners []models.PartnerV2Response) {
				// All partners should have errors
				for _, partner := range partners {
					assert.False(t, partner.IsServiceable, "Partner %s should not be serviceable", partner.PartnerCode)
					assert.NotNil(t, partner.Error, "Partner %s should have error", partner.PartnerCode)
				}

				// Verify specific error messages
				errorMessages := make(map[string]string)
				for _, partner := range partners {
					errorMessages[partner.PartnerCode] = *partner.Error
				}

				assert.Contains(t, errorMessages["partner1"], "timeout error")
				assert.Contains(t, errorMessages["partner2"], "network error")
			},
			description: "When all partners have errors, returns all partners with errors",
		},
		{
			name: "WithDatabasePartnerInfo_UsesPartnerID",
			request: &models.ServiceabilityV2Request{
				PostalCode:     stringPtr("12345"),
				CountryCode:    stringPtr("IN"),
				ParcelCategory: stringPtr("ecomm"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				// Setup partner database info
				partner1ID := uuid.New()
				partner2ID := uuid.New()

				repo.SetPartnerInfo("ecomm", []models.PartnerAttributeMap{
					{PartnerCode: "partner1", PartnerID: &partner1ID},
					{PartnerCode: "partner2", PartnerID: &partner2ID},
				})

				// Partners are serviceable
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityResult(adapter2.CreateServiceableResult())
				factory.SetAdapter("partner2", adapter2)
			},
			expectedSuccess:          true,
			expectedPartnersCount:    2,
			expectedServiceableCount: 2,
			expectedTotalPartners:    2,
			verifyPartnerContent: func(t *testing.T, partners []models.PartnerV2Response) {
				// Partners should use UUID as PartnerID instead of PartnerCode
				for _, partner := range partners {
					assert.True(t, partner.IsServiceable, "Partner %s should be serviceable", partner.PartnerCode)

					// PartnerID should be UUID string, not the partner code
					assert.NotEqual(t, partner.PartnerCode, partner.PartnerID, "PartnerID should be UUID, not code")

					// Verify it's a valid UUID format
					_, err := uuid.Parse(partner.PartnerID)
					assert.NoError(t, err, "PartnerID should be valid UUID")
				}
			},
			description: "When database partner info is available, uses PartnerID from database",
		},
		{
			name: "WithoutDatabasePartnerInfo_UsesPartnerCode",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				// No database partner info
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter1)

				adapter2 := mocks.NewMockPartnerAdapter("partner2")
				adapter2.SetServiceabilityResult(adapter2.CreateServiceableResult())
				factory.SetAdapter("partner2", adapter2)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedSuccess:          true,
			expectedPartnersCount:    2,
			expectedServiceableCount: 2,
			expectedTotalPartners:    2,
			verifyPartnerContent: func(t *testing.T, partners []models.PartnerV2Response) {
				// Partners should use PartnerCode as PartnerID
				for _, partner := range partners {
					assert.True(t, partner.IsServiceable, "Partner %s should be serviceable", partner.PartnerCode)
					assert.Equal(t, partner.PartnerCode, partner.PartnerID, "PartnerID should be same as PartnerCode when no database info")
				}
			},
			description: "When no database partner info is available, uses PartnerCode as PartnerID",
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

			assert.Equal(t, tt.expectedSuccess, response.Success, "Success should match expected: %s", tt.description)
			assert.Equal(t, tt.expectedPartnersCount, len(response.Partners), "Partners count should match expected: %s", tt.description)

			// Verify metadata
			require.NotNil(t, response.Metadata, "Metadata should not be nil")
			assert.Equal(t, tt.expectedServiceableCount, response.Metadata.ServiceableCount, "Serviceable count should match expected: %s", tt.description)
			assert.Equal(t, tt.expectedTotalPartners, response.Metadata.TotalPartners, "Total partners should match expected: %s", tt.description)

			// Verify filter metadata
			assert.Equal(t, tt.request.CountryCode, response.Metadata.Filters.CountryCode, "Country code filter should match")
			assert.Equal(t, tt.request.ParcelCategory, response.Metadata.Filters.ParcelCategory, "Parcel category filter should match")
			assert.Equal(t, tt.request.ProductType, response.Metadata.Filters.ProductType, "Product type filter should match")

			// Verify partner content if specified
			if tt.verifyPartnerContent != nil {
				tt.verifyPartnerContent(t, response.Partners)
			}
		})
	}
}

func TestBuildV2ResponseEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		request     *models.ServiceabilityV2Request
		setupMocks  func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		verify      func(t *testing.T, response *models.ServiceabilityV2Response)
		description string
	}{
		{
			name: "EmptyPartnersResponse",
			request: &models.ServiceabilityV2Request{
				PostalCode:     stringPtr("12345"),
				CountryCode:    stringPtr("IN"),
				ParcelCategory: stringPtr("ecomm"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1", "partner2"})

				// No partners support ecomm
				repo.SetPartnerInfo("ecomm", []models.PartnerAttributeMap{})
			},
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.False(t, response.Success, "Should not be successful")
				assert.Empty(t, response.Partners, "Should have no partners")
				assert.Equal(t, 0, response.Metadata.ServiceableCount, "Should have 0 serviceable partners")
				assert.Equal(t, 0, response.Metadata.TotalPartners, "Should have 0 total partners")
			},
			description: "When no partners are eligible, returns empty response",
		},
		{
			name: "ResponseTimeTracking",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1"})

				// Partner with custom response time
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				result := adapter1.CreateServiceableResult()
				result.ResponseTime = 150 * time.Millisecond
				adapter1.SetServiceabilityResult(result)
				factory.SetAdapter("partner1", adapter1)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.True(t, response.Success, "Should be successful")
				assert.Len(t, response.Partners, 1, "Should have one partner")
				assert.Equal(t, 150*time.Millisecond, response.Partners[0].ResponseTime, "Should track response time")
			},
			description: "Response time is properly tracked for partners",
		},
		{
			name: "ComplexCapabilitiesMapping",
			request: &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1"})

				// Partner with complex capabilities
				adapter1 := mocks.NewMockPartnerAdapter("partner1")
				result := adapter1.CreateServiceableResult()
				result.Capabilities = map[string]interface{}{
					"cod":             true,
					"insurance":       false,
					"max_weight":      25.5,
					"supported_zones": []string{"zone1", "zone2"},
					"pricing": map[string]interface{}{
						"base_rate": 100.0,
						"per_kg":    25.0,
					},
				}
				adapter1.SetServiceabilityResult(result)
				factory.SetAdapter("partner1", adapter1)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.True(t, response.Success, "Should be successful")
				assert.Len(t, response.Partners, 1, "Should have one partner")

				capabilities := response.Partners[0].Capabilities
				assert.Equal(t, true, capabilities["cod"], "COD capability should be preserved")
				assert.Equal(t, false, capabilities["insurance"], "Insurance capability should be preserved")
				assert.Equal(t, 25.5, capabilities["max_weight"], "Max weight should be preserved")

				zones := capabilities["supported_zones"].([]string)
				assert.Len(t, zones, 2, "Should have 2 zones")
				assert.Contains(t, zones, "zone1")
				assert.Contains(t, zones, "zone2")

				pricing := capabilities["pricing"].(map[string]interface{})
				assert.Equal(t, 100.0, pricing["base_rate"])
				assert.Equal(t, 25.0, pricing["per_kg"])
			},
			description: "Complex capabilities mapping is preserved correctly",
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

			// Run custom verification
			tt.verify(t, response)
		})
	}
}

func TestBuildV2ResponseMetadata(t *testing.T) {
	t.Parallel()

	// Setup mocks
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()

	factory.SetSupportedPartners([]string{"partner1", "partner2"})

	// Mixed results
	adapter1 := mocks.NewMockPartnerAdapter("partner1")
	adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
	factory.SetAdapter("partner1", adapter1)

	adapter2 := mocks.NewMockPartnerAdapter("partner2")
	adapter2.SetServiceabilityResult(adapter2.CreateNonServiceableResult())
	factory.SetAdapter("partner2", adapter2)

	repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))

	// Create orchestrator
	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		factory,
		repo,
		time.Second*30,
		false,
	)

	// Execute
	ctx := context.Background()
	response, err := orchestrator.CheckServiceability(ctx, &models.ServiceabilityV2Request{
		PostalCode:     stringPtr("12345"),
		CountryCode:    stringPtr("IN"),
		ParcelCategory: stringPtr("ecomm"),
		ProductType:    stringPtr("electronics"),
	})

	// Verify
	require.NoError(t, err, "CheckServiceability should not return error")
	require.NotNil(t, response, "Response should not be nil")
	require.NotNil(t, response.Metadata, "Metadata should not be nil")

	// Verify metadata structure
	assert.Equal(t, 2, response.Metadata.TotalPartners, "Total partners should be 2")
	assert.Equal(t, 1, response.Metadata.ServiceableCount, "Serviceable count should be 1")
	// Processing time might be 0 for fast operations in tests
	assert.GreaterOrEqual(t, response.Metadata.ProcessingTime, time.Duration(0), "Processing time should be >= 0")

	// Verify filters
	assert.Equal(t, stringPtr("IN"), response.Metadata.Filters.CountryCode)
	assert.Equal(t, stringPtr("ecomm"), response.Metadata.Filters.ParcelCategory)
	assert.Equal(t, stringPtr("electronics"), response.Metadata.Filters.ProductType)
}
