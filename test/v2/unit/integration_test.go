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
				Package: &models.Package{
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
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl", "shipyaari"})

				// DHL supports international
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerCode:   "dhl",
					IsServiceable: true,
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
					PartnerCode:   "shipyaari",
					IsServiceable: false,
					Services:      []models.ServiceV2{},
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
				assert.True(t, dhlPartner.IsServiceable)
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
					PartnerCode:   "dhl",
					IsServiceable: true,
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
				assert.True(t, response.Partners[0].IsServiceable)
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
						PartnerCode:   partnerCode,
						IsServiceable: false,
						Services:      []models.ServiceV2{},
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
					assert.False(t, partner.IsServiceable)
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
		PartnerCode:   "dhl",
		IsServiceable: true,
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
		PartnerCode:   "shipyaari",
		IsServiceable: true,
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
		PartnerCode:   "smile_ecom",
		IsServiceable: true,
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
		PartnerCode:   "smile_courier",
		IsServiceable: true,
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
		Package: &models.Package{
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
		assert.True(t, partner.IsServiceable)
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
		PartnerCode:   "working_partner",
		IsServiceable: true,
		Services:      []models.ServiceV2{{ServiceCode: "STANDARD", ServiceName: "Standard"}},
	})
	mockFactory.SetAdapter("working_partner", workingAdapter)

	// Error partner
	errorAdapter := mocks.NewMockPartnerAdapter("error_partner")
	errorAdapter.SetHealthy(false)
	mockFactory.SetAdapter("error_partner", errorAdapter)

	// Slow partner (will be handled by timeout)
	slowAdapter := mocks.NewMockPartnerAdapter("slow_partner")
	slowAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
		PartnerCode:   "slow_partner",
		IsServiceable: true,
		Services:      []models.ServiceV2{{ServiceCode: "SLOW", ServiceName: "Slow Service"}},
		ResponseTime:  10 * time.Second, // Very slow
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
		if partner.PartnerCode == "working_partner" && partner.IsServiceable {
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
			PartnerCode:   partnerCode,
			IsServiceable: true,
			Services:      []models.ServiceV2{{ServiceCode: "STANDARD", ServiceName: "Standard"}},
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
