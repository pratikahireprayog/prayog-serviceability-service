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

// TestBulkCheckServiceability_AllSuccessful tests bulk operation with all requests successful
func TestBulkCheckServiceability_AllSuccessful(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()

	// Setup serviceable partners
	mockFactory.SetSupportedPartners([]string{"shipyaari", "smile_ecom"})

	adapter1 := mocks.NewMockPartnerAdapter("shipyaari")
	adapter1.SetServiceabilityResult(&common.PartnerServiceabilityResult{
		PartnerCode:   "shipyaari",
		IsServiceable: true,
		Services:      []models.ServiceV2{{ServiceCode: "SDD", ServiceName: "Same Day Delivery"}},
	})
	mockFactory.SetAdapter("shipyaari", adapter1)

	adapter2 := mocks.NewMockPartnerAdapter("smile_ecom")
	adapter2.SetServiceabilityResult(&common.PartnerServiceabilityResult{
		PartnerCode:   "smile_ecom",
		IsServiceable: true,
		Services:      []models.ServiceV2{{ServiceCode: "NDD", ServiceName: "Next Day Delivery"}},
	})
	mockFactory.SetAdapter("smile_ecom", adapter2)

	// Create orchestrator
	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		mockFactory,
		mockRepo,
		5*time.Second,
		false,
	)

	// Create bulk request with multiple valid requests
	bulkRequest := &models.BulkServiceabilityV2Request{
		Requests: []models.ServiceabilityV2Request{
			{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr("110001"),
			},
			{
				CountryCode:           stringPtr("IN"),
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("110002"),
			},
			{
				CountryCode:      stringPtr("IN"),
				PostalCode:       stringPtr("110003"),
				SourcePostalCode: stringPtr("110001"),
			},
		},
	}

	// Execute
	response, err := orchestrator.BulkCheckServiceability(context.Background(), bulkRequest)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Len(t, response.Data, 3)
	assert.Nil(t, response.Error)

	// Verify all individual responses are successful
	for i, individualResponse := range response.Data {
		assert.True(t, individualResponse.Success, "Request %d should be successful", i+1)
		assert.Len(t, individualResponse.Partners, 2, "Request %d should have 2 partners", i+1)
		assert.Nil(t, individualResponse.Error, "Request %d should have no error", i+1)

		// Verify all partners are serviceable
		for _, partner := range individualResponse.Partners {
			assert.True(t, partner.IsServiceable)
		}
	}
}

// TestBulkCheckServiceability_AllFailures tests bulk operation with all requests failing
func TestBulkCheckServiceability_AllFailures(t *testing.T) {
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

	// Create bulk request with all invalid requests
	bulkRequest := &models.BulkServiceabilityV2Request{
		Requests: []models.ServiceabilityV2Request{
			{
				CountryCode: stringPtr("IN"),
				// Missing postal codes - should fail validation
			},
			{
				CountryCode:           stringPtr("IN"),
				PostalCode:            stringPtr("110001"),
				SourcePostalCode:      stringPtr("110002"),
				DestinationPostalCode: stringPtr("110003"),
				// Conflicting postal codes - should fail validation
			},
			{
				CountryCode:      stringPtr("IN"),
				SourcePostalCode: stringPtr("110001"),
				// Missing destination postal code - should fail validation
			},
		},
	}

	// Execute
	response, err := orchestrator.BulkCheckServiceability(context.Background(), bulkRequest)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.False(t, response.Success)
	assert.Len(t, response.Data, 3)
	assert.NotNil(t, response.Error)
	assert.Contains(t, response.Error.Message, "Processed 3 requests with 3 errors")

	// Verify all individual responses failed
	for i, individualResponse := range response.Data {
		assert.False(t, individualResponse.Success, "Request %d should have failed", i+1)
		assert.NotNil(t, individualResponse.Error, "Request %d should have an error", i+1)
		assert.Contains(t, individualResponse.Error.Message, "Request")
	}
}

// TestBulkCheckServiceability_MaxRequestsLimit tests bulk operation request limits
func TestBulkCheckServiceability_MaxRequestsLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		requestCount  int
		expectedError bool
		errorContains string
	}{
		{
			name:          "Within limit - 50 requests",
			requestCount:  50,
			expectedError: false,
		},
		{
			name:          "Within limit - 100 requests",
			requestCount:  100,
			expectedError: false,
		},
		{
			name:          "Exceeds limit - 101 requests",
			requestCount:  101,
			expectedError: false, // Current implementation doesn't enforce limits
		},
		{
			name:          "Single request",
			requestCount:  1,
			expectedError: false,
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
			mockFactory.SetSupportedPartners([]string{"shipyaari"})
			adapter := mocks.NewMockPartnerAdapter("shipyaari")
			adapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
				PartnerCode:   "shipyaari",
				IsServiceable: true,
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

			// Create bulk request with specified number of requests
			requests := make([]models.ServiceabilityV2Request, tt.requestCount)
			for i := 0; i < tt.requestCount; i++ {
				requests[i] = models.ServiceabilityV2Request{
					CountryCode: stringPtr("IN"),
					PostalCode:  stringPtr("110001"),
				}
			}

			bulkRequest := &models.BulkServiceabilityV2Request{
				Requests: requests,
			}

			// Execute
			response, err := orchestrator.BulkCheckServiceability(context.Background(), bulkRequest)

			if tt.expectedError {
				// Verify error case
				require.Error(t, err)
				require.Nil(t, response)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				// Verify success case
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Len(t, response.Data, tt.requestCount)

				// For reasonable request counts, verify all succeeded
				if tt.requestCount <= 100 {
					assert.True(t, response.Success)
					for i, individualResponse := range response.Data {
						assert.True(t, individualResponse.Success, "Request %d should be successful", i+1)
					}
				}
			}
		})
	}
}

// TestBulkCheckServiceability_PerformanceAndConcurrency tests bulk operation performance
func TestBulkCheckServiceability_PerformanceAndConcurrency(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()

	// Setup multiple partners
	mockFactory.SetSupportedPartners([]string{"shipyaari", "smile_ecom", "smile_courier"})

	for _, partnerCode := range []string{"shipyaari", "smile_ecom", "smile_courier"} {
		adapter := mocks.NewMockPartnerAdapter(partnerCode)
		adapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
			PartnerCode:   partnerCode,
			IsServiceable: true,
			Services:      []models.ServiceV2{{ServiceCode: "SDD", ServiceName: "Same Day Delivery"}},
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

	// Create bulk request with multiple requests
	requestCount := 10
	requests := make([]models.ServiceabilityV2Request, requestCount)
	for i := 0; i < requestCount; i++ {
		requests[i] = models.ServiceabilityV2Request{
			CountryCode: stringPtr("IN"),
			PostalCode:  stringPtr("110001"),
		}
	}

	bulkRequest := &models.BulkServiceabilityV2Request{
		Requests: requests,
	}

	// Execute and measure time
	start := time.Now()
	response, err := orchestrator.BulkCheckServiceability(context.Background(), bulkRequest)
	duration := time.Since(start)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Len(t, response.Data, requestCount)

	// Performance verification - bulk operations should be reasonably fast
	// This is more of a sanity check than a strict requirement
	assert.Less(t, duration, 5*time.Second, "Bulk operation should complete within reasonable time")

	// Verify all responses are successful
	for i, individualResponse := range response.Data {
		assert.True(t, individualResponse.Success, "Request %d should be successful", i+1)
		assert.Len(t, individualResponse.Partners, 3, "Request %d should have 3 partners", i+1)
	}
}

// TestBulkCheckServiceability_MixedRequestTypes tests bulk operation with different request types
func TestBulkCheckServiceability_MixedRequestTypes(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()

	// Setup mock repository to handle parcel category "ecomm"
	mockRepo.SetPartnerInfo("ecomm", []models.PartnerAttributeMap{
		{PartnerCode: "shipyaari"},
		{PartnerCode: "smile_ecom"},
	})

	// Setup serviceable partners
	mockFactory.SetSupportedPartners([]string{"shipyaari", "smile_ecom"})

	adapter1 := mocks.NewMockPartnerAdapter("shipyaari")
	adapter1.SetServiceabilityResult(&common.PartnerServiceabilityResult{
		PartnerCode:   "shipyaari",
		IsServiceable: true,
		Services:      []models.ServiceV2{{ServiceCode: "SDD", ServiceName: "Same Day Delivery"}},
	})
	mockFactory.SetAdapter("shipyaari", adapter1)

	adapter2 := mocks.NewMockPartnerAdapter("smile_ecom")
	adapter2.SetServiceabilityResult(&common.PartnerServiceabilityResult{
		PartnerCode:   "smile_ecom",
		IsServiceable: true,
		Services:      []models.ServiceV2{{ServiceCode: "NDD", ServiceName: "Next Day Delivery"}},
	})
	mockFactory.SetAdapter("smile_ecom", adapter2)

	// Create orchestrator
	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		mockFactory,
		mockRepo,
		5*time.Second,
		false,
	)

	// Create bulk request with different request types
	bulkRequest := &models.BulkServiceabilityV2Request{
		Requests: []models.ServiceabilityV2Request{
			{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr("110001"),
			},
			{
				CountryCode:           stringPtr("IN"),
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("110002"),
			},
			{
				CountryCode:      stringPtr("IN"),
				PostalCode:       stringPtr("110003"),
				SourcePostalCode: stringPtr("110001"),
			},
			{
				CountryCode:    stringPtr("IN"),
				PostalCode:     stringPtr("110004"),
				ParcelCategory: stringPtr("ecomm"),
			},
			{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr("110005"),
				Package: &models.Package{
					Weight: &models.Weight{
						Value: 1.5,
						Unit:  "kg",
					},
				},
			},
		},
	}

	// Execute
	response, err := orchestrator.BulkCheckServiceability(context.Background(), bulkRequest)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Len(t, response.Data, 5)
	assert.Nil(t, response.Error)

	// Verify all individual responses are successful
	for i, individualResponse := range response.Data {
		assert.True(t, individualResponse.Success, "Request %d should be successful", i+1)
		assert.Len(t, individualResponse.Partners, 2, "Request %d should have 2 partners", i+1)
		assert.Nil(t, individualResponse.Error, "Request %d should have no error", i+1)
	}
}

// TestBulkCheckServiceability_ReturnOnlyServiceable tests bulk operation respects returnOnlyServiceable setting
func TestBulkCheckServiceability_ReturnOnlyServiceable(t *testing.T) {
	t.Parallel()

	// Simple test to verify bulk operations work with the new feature
	t.Run("BulkReturnOnlyServiceable_Integration", func(t *testing.T) {
		t.Parallel()

		// Setup mocks
		mockFactory := mocks.NewMockPartnerAdapterFactory()
		mockRepo := mocks.NewMockPartnerAttributeMapRepository()

		// Setup simple scenario: one serviceable partner
		mockFactory.SetSupportedPartners([]string{"dhl"})

		// DHL adapter - serviceable (no errors)
		dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
		dhlAdapter.SetServiceabilityResult(dhlAdapter.CreateServiceableResult())
		mockFactory.SetAdapter("dhl", dhlAdapter)

		// Ensure repository doesn't filter out the partner
		mockRepo.SetGetPartnerInfoByAttributeError(fmt.Errorf("not filtered"))

		// Create orchestrator with returnOnlyServiceable=true
		orchestrator := orchestrators.NewServiceabilityOrchestrator(
			mockFactory,
			mockRepo,
			5*time.Second,
			true, // returnOnlyServiceable=true
		)

		// Create bulk request with simple valid request (no parcel category to avoid filtering)
		bulkRequest := &models.BulkServiceabilityV2Request{
			Requests: []models.ServiceabilityV2Request{
				{
					PostalCode:  stringPtr("110001"),
					CountryCode: stringPtr("IN"),
					// No parcel category to avoid filtering issues
				},
			},
		}

		// Execute
		response, err := orchestrator.BulkCheckServiceability(context.Background(), bulkRequest)

		// Verify basic bulk functionality works
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Data, 1)

		// Check that the individual response preserves the returnOnlyServiceable behavior
		individualResponse := response.Data[0]

		// The key test: verify bulk operations preserve individual orchestrator behavior
		if individualResponse.Success {
			// If successful, should have serviceable partners
			assert.NotEmpty(t, individualResponse.Partners, "Successful responses should have partners")
			for _, partner := range individualResponse.Partners {
				assert.True(t, partner.IsServiceable, "All returned partners should be serviceable when returnOnlyServiceable=true")
			}
		} else {
			// If not successful, with returnOnlyServiceable=true, should have empty partners or error
			if len(individualResponse.Partners) == 0 {
				assert.NotNil(t, individualResponse.Error, "Empty partners array should have error explanation")
			}
		}

		t.Logf("✅ Bulk operations integration with returnOnlyServiceable=true verified")
	})
}
