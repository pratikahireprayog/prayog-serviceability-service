package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stringPtr helper function is defined in other test files

// TestIntegration_EndToEndFlow_InternationalRequest tests full flow for international requests
func TestIntegration_EndToEndFlow_InternationalRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		request       *models.ServiceabilityV2Request
		setupMocks    func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectedValid bool
		verify        func(t *testing.T, response *models.ServiceabilityV2Response)
	}{
		{
			name: "International request with package information",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("US"),
				SourcePostalCode:      stringPtr("10001"),
				DestinationPostalCode: stringPtr("90210"),
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{{
					Weight: &models.Weight{
						Value: 2.5,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 30.0,
						Width:  20.0,
						Height: 15.0,
						Unit:   "cm",
					},
				}},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl", "shipyaari"})

				// DHL supports international
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerCode: "dhl",
					Services: []models.ServiceV2{
						{ServiceCode: "EXPRESS", ServiceName: "DHL Express"},
						{ServiceCode: "ECONOMY", ServiceName: "DHL Economy"},
					},
					Capabilities: map[string]interface{}{
						"international": true,
						"tracking":      true,
						"insurance":     true,
					},
				})
				factory.SetAdapter("dhl", dhlAdapter)

				// Shipyaari doesn't support international
				shipyaariAdapter := mocks.NewMockPartnerAdapter("shipyaari")
				shipyaariAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerCode: "shipyaari",

					Services: []models.ServiceV2{},
					Capabilities: map[string]interface{}{
						"international": false,
					},
				})
				factory.SetAdapter("shipyaari", shipyaariAdapter)
			},
			expectedValid: true,
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.True(t, response.Success)
				assert.Len(t, response.Partners, 1) // Only DHL should be serviceable

				dhlPartner := response.Partners[0]
				assert.Equal(t, "dhl", dhlPartner.PartnerCode)
				assert.True(t, len(dhlPartner.Services) > 0)
				assert.Len(t, dhlPartner.Services, 2)
				assert.NotEmpty(t, dhlPartner.Capabilities)

				// Verify metadata
				assert.NotNil(t, response.Metadata)
				assert.Equal(t, 2, response.Metadata.TotalPartners)
				assert.Equal(t, 1, response.Metadata.ServiceableCount)
				assert.Equal(t, "US", *response.Metadata.Filters.CountryCode)
				assert.Equal(t, "international", *response.Metadata.Filters.ParcelCategory)
			},
		},
		{
			name: "International request without package info",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("US"),
				SourcePostalCode:      stringPtr("10001"),
				DestinationPostalCode: stringPtr("90210"),
				ParcelCategory:        stringPtr("international"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})

				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerCode: "dhl",

					Services: []models.ServiceV2{
						{ServiceCode: "EXPRESS", ServiceName: "DHL Express"},
					},
					Capabilities: map[string]interface{}{
						"international":         true,
						"requires_package_info": false,
					},
				})
				factory.SetAdapter("dhl", dhlAdapter)
			},
			expectedValid: true,
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.True(t, response.Success)
				assert.Len(t, response.Partners, 1)
				assert.Equal(t, "dhl", response.Partners[0].PartnerCode)
				assert.True(t, len(response.Partners[0].Services) > 0)
			},
		},
		{
			name: "International request with no supporting partners",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("ZZ"),
				SourcePostalCode:      stringPtr("00000"),
				DestinationPostalCode: stringPtr("00001"),
				ParcelCategory:        stringPtr("international"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"shipyaari", "smile_ecom"})

				// Both partners don't support this international route
				for _, partnerCode := range []string{"shipyaari", "smile_ecom"} {
					adapter := mocks.NewMockPartnerAdapter(partnerCode)
					adapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
						PartnerCode: partnerCode,

						Services: []models.ServiceV2{},
						Capabilities: map[string]interface{}{
							"international":       false,
							"unsupported_country": "ZZ",
						},
					})
					factory.SetAdapter(partnerCode, adapter)
				}
			},
			expectedValid: true,
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.False(t, response.Success)   // No partners serviceable
				assert.Len(t, response.Partners, 2) // All partners returned when success=false

				for _, partner := range response.Partners {
					assert.False(t, len(partner.Services) > 0)
				}

				assert.NotNil(t, response.Metadata)
				assert.Equal(t, 2, response.Metadata.TotalPartners)
				assert.Equal(t, 0, response.Metadata.ServiceableCount)
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
			tt.setupMocks(mockFactory, mockRepo)

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				10*time.Second, // Longer timeout for international requests
				false,
			)

			// Execute
			response, err := orchestrator.CheckServiceability(context.Background(), tt.request)

			if tt.expectedValid {
				// Verify successful response
				require.NoError(t, err)
				require.NotNil(t, response)
				tt.verify(t, response)
			} else {
				// Verify error response
				require.Error(t, err)
				require.Nil(t, response)
			}
		})
	}
}

// TestIntegration_CompleteWorkflow tests the complete workflow from request to response
func TestIntegration_CompleteWorkflow(t *testing.T) {
	t.Parallel()

	// Setup comprehensive mock environment
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()

	// Setup multiple partners with different characteristics
	mockFactory.SetSupportedPartners([]string{"dhl", "shipyaari", "smile_ecom", "smile_courier"})

	// DHL - International specialist
	dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
	dhlAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
		PartnerCode: "dhl",

		Services: []models.ServiceV2{
			{ServiceCode: "EXPRESS", ServiceName: "DHL Express"},
			{ServiceCode: "ECONOMY", ServiceName: "DHL Economy"},
		},
		Capabilities: map[string]interface{}{
			"international": true,
			"domestic":      true,
			"tracking":      true,
			"insurance":     true,
		},
		ResponseTime: 150 * time.Millisecond,
	})
	mockFactory.SetAdapter("dhl", dhlAdapter)

	// Shipyaari - Domestic specialist
	shipyaariAdapter := mocks.NewMockPartnerAdapter("shipyaari")
	shipyaariAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
		PartnerCode: "shipyaari",

		Services: []models.ServiceV2{
			{ServiceCode: "SDD", ServiceName: "Same Day Delivery"},
			{ServiceCode: "NDD", ServiceName: "Next Day Delivery"},
		},
		Capabilities: map[string]interface{}{
			"domestic":   true,
			"cod":        true,
			"hyperlocal": true,
		},
		ResponseTime: 100 * time.Millisecond,
	})
	mockFactory.SetAdapter("shipyaari", shipyaariAdapter)

	// Smile Ecom - E-commerce focused
	smileEcomAdapter := mocks.NewMockPartnerAdapter("smile_ecom")
	smileEcomAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
		PartnerCode: "smile_ecom",

		Services: []models.ServiceV2{
			{ServiceCode: "STANDARD", ServiceName: "Standard Delivery"},
			{ServiceCode: "EXPRESS", ServiceName: "Express Delivery"},
		},
		Capabilities: map[string]interface{}{
			"ecommerce": true,
			"bulk":      true,
			"returns":   true,
		},
		ResponseTime: 80 * time.Millisecond,
	})
	mockFactory.SetAdapter("smile_ecom", smileEcomAdapter)

	// Smile Courier - General courier
	smileCourierAdapter := mocks.NewMockPartnerAdapter("smile_courier")
	smileCourierAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
		PartnerCode: "smile_courier",

		Services: []models.ServiceV2{
			{ServiceCode: "REGULAR", ServiceName: "Regular Delivery"},
		},
		Capabilities: map[string]interface{}{
			"documents": true,
			"fragile":   true,
		},
		ResponseTime: 120 * time.Millisecond,
	})
	mockFactory.SetAdapter("smile_courier", smileCourierAdapter)

	// Create orchestrator
	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		mockFactory,
		mockRepo,
		5*time.Second,
		false,
	)

	// Test complex request
	request := &models.ServiceabilityV2Request{
		CountryCode:           stringPtr("IN"),
		SourcePostalCode:      stringPtr("110001"),
		DestinationPostalCode: stringPtr("400001"),
		ParcelCategory:        stringPtr("ecomm"),
		ProductType:           stringPtr("electronics"),
		Packages: []models.Package{{
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
		}},
	}

	// Execute
	start := time.Now()
	response, err := orchestrator.CheckServiceability(context.Background(), request)
	duration := time.Since(start)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, response)

	// Verify response structure
	assert.True(t, response.Success)
	assert.Len(t, response.Partners, 4) // All partners serviceable
	assert.Nil(t, response.Error)

	// Verify metadata
	assert.NotNil(t, response.Metadata)
	assert.Equal(t, 4, response.Metadata.TotalPartners)
	assert.Equal(t, 4, response.Metadata.ServiceableCount)
	assert.Equal(t, "IN", *response.Metadata.Filters.CountryCode)
	assert.Equal(t, "ecomm", *response.Metadata.Filters.ParcelCategory)
	assert.Equal(t, "electronics", *response.Metadata.Filters.ProductType)

	// Verify partner responses
	partnerCodes := make(map[string]bool)
	for _, partner := range response.Partners {
		partnerCodes[partner.PartnerCode] = true
		assert.True(t, len(partner.Services) > 0)
		assert.NotEmpty(t, partner.Services)
		assert.NotEmpty(t, partner.Capabilities)
		assert.Greater(t, partner.ResponseTime, time.Duration(0))
	}

	// Verify all expected partners are present
	expectedPartners := []string{"dhl", "shipyaari", "smile_ecom", "smile_courier"}
	for _, expected := range expectedPartners {
		assert.True(t, partnerCodes[expected], "Partner %s should be present", expected)
	}

	// Performance verification
	assert.Less(t, duration, 1*time.Second, "Operation should complete quickly")
}

// TestIntegration_ErrorRecoveryAndResilience tests error recovery and resilience
func TestIntegration_ErrorRecoveryAndResilience(t *testing.T) {
	t.Parallel()

	// Setup mocks with mixed scenarios
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()

	mockFactory.SetSupportedPartners([]string{"working_partner", "error_partner", "slow_partner"})

	// Working partner
	workingAdapter := mocks.NewMockPartnerAdapter("working_partner")
	workingAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
		PartnerCode: "working_partner",

		Services: []models.ServiceV2{{ServiceCode: "STANDARD", ServiceName: "Standard"}},
	})
	mockFactory.SetAdapter("working_partner", workingAdapter)

	// Error partner
	errorAdapter := mocks.NewMockPartnerAdapter("error_partner")
	errorAdapter.SetHealthy(false)
	mockFactory.SetAdapter("error_partner", errorAdapter)

	// Slow partner (will be handled by timeout)
	slowAdapter := mocks.NewMockPartnerAdapter("slow_partner")
	slowAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
		PartnerCode: "slow_partner",

		Services:     []models.ServiceV2{{ServiceCode: "SLOW", ServiceName: "Slow Service"}},
		ResponseTime: 10 * time.Second, // Very slow
	})
	mockFactory.SetAdapter("slow_partner", slowAdapter)

	// Create orchestrator with short timeout
	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		mockFactory,
		mockRepo,
		2*time.Second, // Short timeout
		false,
	)

	// Test request
	request := &models.ServiceabilityV2Request{
		CountryCode: stringPtr("IN"),
		PostalCode:  stringPtr("110001"),
	}

	// Execute
	response, err := orchestrator.CheckServiceability(context.Background(), request)

	// Verify - should succeed with partial results
	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return all partners (including errors) when success depends on working partners
	assert.NotEmpty(t, response.Partners)
	assert.NotNil(t, response.Metadata)
	assert.Equal(t, 3, response.Metadata.TotalPartners)

	// Verify that at least the working partner is included
	workingPartnerFound := false
	for _, partner := range response.Partners {
		if partner.PartnerCode == "working_partner" && len(partner.Services) > 0 {
			workingPartnerFound = true
			break
		}
	}
	assert.True(t, workingPartnerFound, "Working partner should be found and serviceable")
}

// TestIntegration_ConcurrentRequests tests concurrent request handling
func TestIntegration_ConcurrentRequests(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()

	mockFactory.SetSupportedPartners([]string{"partner1", "partner2"})

	// Setup partners
	for _, partnerCode := range []string{"partner1", "partner2"} {
		adapter := mocks.NewMockPartnerAdapter(partnerCode)
		adapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
			PartnerCode: partnerCode,

			Services: []models.ServiceV2{{ServiceCode: "STANDARD", ServiceName: "Standard"}},
		})
		mockFactory.SetAdapter(partnerCode, adapter)
	}

	// Create orchestrator
	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		mockFactory,
		mockRepo,
		5*time.Second,
		false,
	)

	// Test concurrent requests
	concurrency := 10
	results := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			request := &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr("110001"),
			}

			response, err := orchestrator.CheckServiceability(context.Background(), request)

			success := err == nil && response != nil && response.Success
			results <- success
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < concurrency; i++ {
		if <-results {
			successCount++
		}
	}

	// Verify all requests succeeded
	assert.Equal(t, concurrency, successCount, "All concurrent requests should succeed")
}

// TestIntegration_ReturnOnlyServiceablePartners_EndToEnd tests the complete workflow with returnOnlyServiceable=true
func TestIntegration_ReturnOnlyServiceablePartners_EndToEnd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		returnOnlyServiceable bool
		setupMocks            func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectedSuccess       bool
		expectedPartnersCount int
		expectedErrorCode     string
		description           string
	}{
		{
			name:                  "ReturnOnlyServiceable_True_ValidationError_EndToEnd",
			returnOnlyServiceable: true,
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})

				// DHL adapter with validation error (postal code not found)
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityError(fmt.Errorf("Failed to determine destination country: country code not found for postal code 266001: POSTAL_CODE_NOT_FOUND: Postal code not found (Postal code '266001' does not exist in our database)"))
				factory.SetAdapter("dhl", dhlAdapter)

				repo.SetGetPartnerInfoByAttributeError(fmt.Errorf("not filtered"))
			},
			expectedSuccess:       false,
			expectedPartnersCount: 0,
			expectedErrorCode:     "POSTAL_CODE_NOT_FOUND",
			description:           "End-to-end test with validation error should return empty partners array and POSTAL_CODE_NOT_FOUND",
		},
		{
			name:                  "ReturnOnlyServiceable_True_NoServiceablePartners_EndToEnd",
			returnOnlyServiceable: true,
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl", "shipyaari"})

				// Both partners non-serviceable (no errors)
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityResult(dhlAdapter.CreateCleanNonServiceableResult())
				factory.SetAdapter("dhl", dhlAdapter)

				shipyaariAdapter := mocks.NewMockPartnerAdapter("shipyaari")
				shipyaariAdapter.SetServiceabilityResult(shipyaariAdapter.CreateCleanNonServiceableResult())
				factory.SetAdapter("shipyaari", shipyaariAdapter)

				repo.SetGetPartnerInfoByAttributeError(fmt.Errorf("not filtered"))
			},
			expectedSuccess:       false,
			expectedPartnersCount: 0,
			expectedErrorCode:     "NO_SERVICEABLE_PARTNERS",
			description:           "End-to-end test with no serviceable partners should return empty partners array and NO_SERVICEABLE_PARTNERS error",
		},
		{
			name:                  "ReturnOnlyServiceable_True_SuccessfulServiceable_EndToEnd",
			returnOnlyServiceable: true,
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl", "shipyaari"})

				// One serviceable, one non-serviceable (no errors)
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityResult(dhlAdapter.CreateServiceableResult())
				factory.SetAdapter("dhl", dhlAdapter)

				shipyaariAdapter := mocks.NewMockPartnerAdapter("shipyaari")
				shipyaariAdapter.SetServiceabilityResult(shipyaariAdapter.CreateCleanNonServiceableResult())
				factory.SetAdapter("shipyaari", shipyaariAdapter)

				repo.SetGetPartnerInfoByAttributeError(fmt.Errorf("not filtered"))
			},
			expectedSuccess:       true,
			expectedPartnersCount: 1,
			expectedErrorCode:     "",
			description:           "End-to-end test with serviceable partners should return only serviceable partners",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()
			tt.setupMocks(mockFactory, mockRepo)

			// Create orchestrator with specific returnOnlyServiceable setting
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				10*time.Second, // Longer timeout for integration tests
				tt.returnOnlyServiceable,
			)

			// Create test request
			request := &models.ServiceabilityV2Request{
				PostalCode:     stringPtr("266001"),
				CountryCode:    stringPtr("IN"),
				ParcelCategory: stringPtr("international"),
				Packages: []models.Package{{
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
				}},
			}

			// Execute the complete workflow
			startTime := time.Now()
			response, err := orchestrator.CheckServiceability(context.Background(), request)
			processingTime := time.Since(startTime)

			// Verify no system errors
			require.NoError(t, err, "Should not return system error")
			require.NotNil(t, response, "Response should not be nil")

			// Verify response structure
			assert.Equal(t, tt.expectedSuccess, response.Success, "Success should match expected")
			assert.Len(t, response.Partners, tt.expectedPartnersCount, "Partners count should match expected")

			// Verify metadata
			assert.NotNil(t, response.Metadata, "Metadata should be present")
			assert.Equal(t, "international", *response.Metadata.Filters.ParcelCategory, "Parcel category should be preserved")
			assert.Equal(t, "IN", *response.Metadata.Filters.CountryCode, "Country code should be preserved")

			// Verify error handling
			if tt.expectedErrorCode != "" {
				assert.NotNil(t, response.Error, "Should have error when expected")
				assert.Equal(t, tt.expectedErrorCode, response.Error.Code, "Error code should match expected")
				assert.NotEmpty(t, response.Error.Message, "Error message should not be empty")
			} else {
				assert.Nil(t, response.Error, "Should not have error when not expected")
			}

			// Verify performance
			assert.Less(t, processingTime, 5*time.Second, "Processing should complete within reasonable time")

			// Verify partner responses when present
			if tt.expectedPartnersCount > 0 {
				for _, partner := range response.Partners {
					assert.True(t, len(partner.Services) > 0, "All returned partners should be serviceable when returnOnlyServiceable=true")
					assert.Nil(t, partner.Error, "Serviceable partners should not have errors")
					assert.NotEmpty(t, partner.PartnerCode, "Partner code should be present")
				}
			}

			t.Logf("✅ %s: Processing time: %v", tt.description, processingTime)
		})
	}
}
