package unit

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartnerProcessingConcurrency(t *testing.T) {
	tests := []struct {
		name               string
		partnerCount       int
		setupPartners      func(factory *mocks.MockPartnerAdapterFactory, partnerCodes []string)
		expectedCallCounts map[string]int
		description        string
	}{
		{
			name:         "SinglePartner_ProcessedCorrectly",
			partnerCount: 1,
			setupPartners: func(factory *mocks.MockPartnerAdapterFactory, partnerCodes []string) {
				adapter := mocks.NewMockPartnerAdapter(partnerCodes[0])
				adapter.SetServiceabilityResult(adapter.CreateServiceableResult())
				factory.SetAdapter(partnerCodes[0], adapter)
			},
			expectedCallCounts: map[string]int{"partner1": 1},
			description:        "Single partner is processed correctly",
		},
		{
			name:         "MultiplePartners_ProcessedConcurrently",
			partnerCount: 5,
			setupPartners: func(factory *mocks.MockPartnerAdapterFactory, partnerCodes []string) {
				for _, code := range partnerCodes {
					adapter := mocks.NewMockPartnerAdapter(code)
					adapter.SetServiceabilityResult(adapter.CreateServiceableResult())
					factory.SetAdapter(code, adapter)
				}
			},
			expectedCallCounts: map[string]int{
				"partner1": 1, "partner2": 1, "partner3": 1, "partner4": 1, "partner5": 1,
			},
			description: "Multiple partners are processed concurrently",
		},
		{
			name:         "LargeNumberOfPartners_ProcessedConcurrently",
			partnerCount: 10,
			setupPartners: func(factory *mocks.MockPartnerAdapterFactory, partnerCodes []string) {
				for _, code := range partnerCodes {
					adapter := mocks.NewMockPartnerAdapter(code)
					adapter.SetServiceabilityResult(adapter.CreateServiceableResult())
					factory.SetAdapter(code, adapter)
				}
			},
			expectedCallCounts: map[string]int{
				"partner1": 1, "partner2": 1, "partner3": 1, "partner4": 1, "partner5": 1,
				"partner6": 1, "partner7": 1, "partner8": 1, "partner9": 1, "partner10": 1,
			},
			description: "Large number of partners are processed concurrently",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Generate partner codes
			partnerCodes := make([]string, tt.partnerCount)
			for i := 0; i < tt.partnerCount; i++ {
				partnerCodes[i] = fmt.Sprintf("partner%d", i+1)
			}

			// Setup mocks
			factory := mocks.NewMockPartnerAdapterFactory()
			repo := mocks.NewMockPartnerAttributeMapRepository()

			factory.SetSupportedPartners(partnerCodes)
			tt.setupPartners(factory, partnerCodes)

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
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			})

			// Verify no errors
			require.NoError(t, err, "CheckServiceability should not return error")
			require.NotNil(t, response, "Response should not be nil")

			// Verify concurrent processing worked
			assert.Equal(t, tt.partnerCount, len(response.Partners), "Should process all partners")
			assert.True(t, response.Success, "Should be successful with all serviceable partners")

			// Verify all partners were called exactly once
			for partnerCode := range tt.expectedCallCounts {
				adapter, exists := factory.GetAdapter(partnerCode)
				require.True(t, exists, "Partner %s should exist", partnerCode)
				mockAdapter := adapter.(*mocks.MockPartnerAdapter)
				assert.True(t, mockAdapter.CheckServiceabilityCalled, "Partner %s should be called", partnerCode)
			}
		})
	}
}

func TestPartnerAdapterCallPattern(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		verifyAdapter func(t *testing.T, adapter *mocks.MockPartnerAdapter, partnerCode string)
		description   string
	}{
		{
			name: "PartnerAdapter_ReceivesCorrectRequest",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1"})

				adapter := mocks.NewMockPartnerAdapter("partner1")
				adapter.SetServiceabilityResult(adapter.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			verifyAdapter: func(t *testing.T, adapter *mocks.MockPartnerAdapter, partnerCode string) {
				assert.True(t, adapter.CheckServiceabilityCalled, "CheckServiceability should be called")
				assert.NotNil(t, adapter.LastRequest, "Should have received request")
				assert.Equal(t, stringPtr("12345"), adapter.LastRequest.PostalCode)
				assert.Equal(t, stringPtr("IN"), adapter.LastRequest.CountryCode)
			},
			description: "Partner adapter receives correct request parameters",
		},
		{
			name: "PartnerAdapter_ReceivesCorrectContext",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1"})

				adapter := mocks.NewMockPartnerAdapter("partner1")
				adapter.SetServiceabilityResult(adapter.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			verifyAdapter: func(t *testing.T, adapter *mocks.MockPartnerAdapter, partnerCode string) {
				assert.NotNil(t, adapter.LastContext, "Should have received context")
				// The context should have a timeout (wrapped by orchestrator)
				deadline, hasDeadline := adapter.LastContext.Deadline()
				assert.True(t, hasDeadline, "Context should have timeout deadline")
				assert.False(t, deadline.IsZero(), "Deadline should not be zero")
			},
			description: "Partner adapter receives correct context",
		},
		{
			name: "PartnerAdapter_HealthCheckCalled",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1"})

				adapter := mocks.NewMockPartnerAdapter("partner1")
				adapter.SetServiceabilityResult(adapter.CreateServiceableResult())
				factory.SetAdapter("partner1", adapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			verifyAdapter: func(t *testing.T, adapter *mocks.MockPartnerAdapter, partnerCode string) {
				assert.True(t, adapter.IsHealthyCalled, "IsHealthy should be called")
			},
			description: "Partner adapter health check is called",
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
			response, err := orchestrator.CheckServiceability(ctx, &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			})

			// Verify no errors
			require.NoError(t, err, "CheckServiceability should not return error")
			require.NotNil(t, response, "Response should not be nil")

			// Verify adapter calls
			adapter, exists := factory.GetAdapter("partner1")
			require.True(t, exists, "Partner should exist")
			mockAdapter := adapter.(*mocks.MockPartnerAdapter)

			tt.verifyAdapter(t, mockAdapter, "partner1")
		})
	}
}

func TestPartnerErrorHandling(t *testing.T) {
	tests := []struct {
		name                string
		setupMocks          func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectedSuccess     bool
		expectedPartners    int
		expectedServiceable int
		verifyPartnerErrors func(t *testing.T, partners []models.PartnerV2Response)
		description         string
	}{
		{
			name: "PartnerNotFound_HandledGracefully",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				// Partner codes in supported list but not in adapters map
				factory.SetSupportedPartners([]string{"missing_partner"})

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedSuccess:     true,
			expectedPartners:    1, // The GetAdapter will create a new adapter
			expectedServiceable: 1, // Default adapter is serviceable
			verifyPartnerErrors: func(t *testing.T, partners []models.PartnerV2Response) {
				assert.Equal(t, "missing_partner", partners[0].PartnerCode)
				assert.True(t, len(partners[0].Services) > 0, "Default adapter should be serviceable")
			},
			description: "Missing partner is handled gracefully by creating default adapter",
		},
		{
			name: "UnhealthyPartner_SkippedCorrectly",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"healthy_partner", "unhealthy_partner"})

				healthyAdapter := mocks.NewMockPartnerAdapter("healthy_partner")
				healthyAdapter.SetServiceabilityResult(healthyAdapter.CreateServiceableResult())
				factory.SetAdapter("healthy_partner", healthyAdapter)

				unhealthyAdapter := mocks.NewMockPartnerAdapter("unhealthy_partner")
				unhealthyAdapter.SetHealthy(false)
				factory.SetAdapter("unhealthy_partner", unhealthyAdapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedSuccess:     true,
			expectedPartners:    1, // Only healthy partner in response
			expectedServiceable: 1,
			verifyPartnerErrors: func(t *testing.T, partners []models.PartnerV2Response) {
				assert.Len(t, partners, 1, "Should only return healthy partner")
				assert.Equal(t, "healthy_partner", partners[0].PartnerCode)
				assert.True(t, len(partners[0].Services) > 0)
			},
			description: "Unhealthy partners are skipped correctly",
		},
		{
			name: "PartnerServiceError_HandledCorrectly",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"working_partner", "error_partner"})

				workingAdapter := mocks.NewMockPartnerAdapter("working_partner")
				workingAdapter.SetServiceabilityResult(workingAdapter.CreateServiceableResult())
				factory.SetAdapter("working_partner", workingAdapter)

				errorAdapter := mocks.NewMockPartnerAdapter("error_partner")
				errorAdapter.SetServiceabilityError(errors.New("service timeout"))
				factory.SetAdapter("error_partner", errorAdapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedSuccess:     true,
			expectedPartners:    1, // Only working partner in success response
			expectedServiceable: 1,
			verifyPartnerErrors: func(t *testing.T, partners []models.PartnerV2Response) {
				assert.Len(t, partners, 1, "Should only return working partner")
				assert.Equal(t, "working_partner", partners[0].PartnerCode)
				assert.True(t, len(partners[0].Services) > 0)
				assert.Nil(t, partners[0].Error)
			},
			description: "Partner service errors are handled correctly",
		},
		{
			name: "AllPartnersError_ReturnsErrorResponse",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"error_partner1", "error_partner2"})

				errorAdapter1 := mocks.NewMockPartnerAdapter("error_partner1")
				errorAdapter1.SetServiceabilityError(errors.New("timeout error"))
				factory.SetAdapter("error_partner1", errorAdapter1)

				errorAdapter2 := mocks.NewMockPartnerAdapter("error_partner2")
				errorAdapter2.SetServiceabilityError(errors.New("network error"))
				factory.SetAdapter("error_partner2", errorAdapter2)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedSuccess:     false,
			expectedPartners:    2, // All partners in error response
			expectedServiceable: 0,
			verifyPartnerErrors: func(t *testing.T, partners []models.PartnerV2Response) {
				assert.Len(t, partners, 2, "Should return all partners with errors")
				for _, partner := range partners {
					assert.False(t, partner.IsServiceable, "Partner %s should not be serviceable", partner.PartnerCode)
					assert.NotNil(t, partner.Error, "Partner %s should have error", partner.PartnerCode)
				}
			},
			description: "When all partners error, returns error response with all partners",
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
			response, err := orchestrator.CheckServiceability(ctx, &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			})

			// Verify no errors at orchestrator level
			require.NoError(t, err, "CheckServiceability should not return error")
			require.NotNil(t, response, "Response should not be nil")

			// Verify response structure
			assert.Equal(t, tt.expectedSuccess, response.Success, "Success should match expected: %s", tt.description)
			assert.Equal(t, tt.expectedPartners, len(response.Partners), "Partners count should match expected: %s", tt.description)

			if response.Metadata != nil {
				assert.Equal(t, tt.expectedServiceable, response.Metadata.ServiceableCount, "Serviceable count should match expected: %s", tt.description)
			}

			// Run custom verification
			if tt.verifyPartnerErrors != nil {
				tt.verifyPartnerErrors(t, response.Partners)
			}
		})
	}
}

func TestPartnerResponseMapping(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		verify      func(t *testing.T, response *models.ServiceabilityV2Response)
		description string
	}{
		{
			name: "PartnerServices_MappedCorrectly",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1"})

				adapter := mocks.NewMockPartnerAdapter("partner1")
				result := adapter.CreateServiceableResult()
				result.Services = []models.ServiceV2{
					{
						ServiceCode: "SDD",
						ServiceName: "Same Day Delivery",
						TATDays:     0,
						IsCOD:       true,
						Pickup:      true,
						Delivery:    true,
						Insurance:   false,
					},
					{
						ServiceCode: "NDD",
						ServiceName: "Next Day Delivery",
						TATDays:     1,
						IsCOD:       false,
						Pickup:      false,
						Delivery:    true,
						Insurance:   true,
					},
				}
				adapter.SetServiceabilityResult(result)
				factory.SetAdapter("partner1", adapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.True(t, response.Success, "Should be successful")
				assert.Len(t, response.Partners, 1, "Should have one partner")

				partner := response.Partners[0]
				assert.Len(t, partner.Services, 2, "Should have 2 services")

				// Verify SDD service
				sdd := partner.Services[0]
				assert.Equal(t, "SDD", sdd.ServiceCode)
				assert.Equal(t, "Same Day Delivery", sdd.ServiceName)
				assert.Equal(t, 0, sdd.TATDays)
				assert.True(t, sdd.IsCOD)
				assert.True(t, sdd.Pickup)
				assert.True(t, sdd.Delivery)
				assert.False(t, sdd.Insurance)

				// Verify NDD service
				ndd := partner.Services[1]
				assert.Equal(t, "NDD", ndd.ServiceCode)
				assert.Equal(t, "Next Day Delivery", ndd.ServiceName)
				assert.Equal(t, 1, ndd.TATDays)
				assert.False(t, ndd.IsCOD)
				assert.False(t, ndd.Pickup)
				assert.True(t, ndd.Delivery)
				assert.True(t, ndd.Insurance)
			},
			description: "Partner services are mapped correctly to response",
		},
		{
			name: "PartnerCapabilities_MappedCorrectly",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"partner1"})

				adapter := mocks.NewMockPartnerAdapter("partner1")
				result := adapter.CreateServiceableResult()
				result.Capabilities = map[string]interface{}{
					"max_weight": 25.5,
					"max_dimensions": map[string]float64{
						"length": 100.0,
						"width":  80.0,
						"height": 60.0,
					},
					"supported_payment_modes": []string{"ONLINE", "COD", "UPI"},
					"delivery_time_slots": map[string]interface{}{
						"morning":   "09:00-12:00",
						"afternoon": "12:00-18:00",
						"evening":   "18:00-21:00",
					},
				}
				adapter.SetServiceabilityResult(result)
				factory.SetAdapter("partner1", adapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.True(t, response.Success, "Should be successful")
				assert.Len(t, response.Partners, 1, "Should have one partner")

				capabilities := response.Partners[0].Capabilities
				assert.Equal(t, 25.5, capabilities["max_weight"])

				dimensions := capabilities["max_dimensions"].(map[string]float64)
				assert.Equal(t, 100.0, dimensions["length"])
				assert.Equal(t, 80.0, dimensions["width"])
				assert.Equal(t, 60.0, dimensions["height"])

				paymentModes := capabilities["supported_payment_modes"].([]string)
				assert.Len(t, paymentModes, 3)
				assert.Contains(t, paymentModes, "ONLINE")
				assert.Contains(t, paymentModes, "COD")
				assert.Contains(t, paymentModes, "UPI")

				timeSlots := capabilities["delivery_time_slots"].(map[string]interface{})
				assert.Equal(t, "09:00-12:00", timeSlots["morning"])
				assert.Equal(t, "12:00-18:00", timeSlots["afternoon"])
				assert.Equal(t, "18:00-21:00", timeSlots["evening"])
			},
			description: "Partner capabilities are mapped correctly to response",
		},
		{
			name: "PartnerResponseTime_TrackedCorrectly",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"fast_partner", "slow_partner"})

				fastAdapter := mocks.NewMockPartnerAdapter("fast_partner")
				fastResult := fastAdapter.CreateServiceableResult()
				fastResult.ResponseTime = 50 * time.Millisecond
				fastAdapter.SetServiceabilityResult(fastResult)
				factory.SetAdapter("fast_partner", fastAdapter)

				slowAdapter := mocks.NewMockPartnerAdapter("slow_partner")
				slowResult := slowAdapter.CreateServiceableResult()
				slowResult.ResponseTime = 300 * time.Millisecond
				slowAdapter.SetServiceabilityResult(slowResult)
				factory.SetAdapter("slow_partner", slowAdapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.True(t, response.Success, "Should be successful")
				assert.Len(t, response.Partners, 2, "Should have two partners")

				responseTimeMap := make(map[string]time.Duration)
				for _, partner := range response.Partners {
					responseTimeMap[partner.PartnerCode] = partner.ResponseTime
				}

				assert.Equal(t, 50*time.Millisecond, responseTimeMap["fast_partner"])
				assert.Equal(t, 300*time.Millisecond, responseTimeMap["slow_partner"])
			},
			description: "Partner response times are tracked correctly",
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
			response, err := orchestrator.CheckServiceability(ctx, &models.ServiceabilityV2Request{
				PostalCode:  stringPtr("12345"),
				CountryCode: stringPtr("IN"),
			})

			// Verify no errors
			require.NoError(t, err, "CheckServiceability should not return error")
			require.NotNil(t, response, "Response should not be nil")

			// Run custom verification
			tt.verify(t, response)
		})
	}
}

func TestPartnerProcessingWithDatabaseInfo(t *testing.T) {
	t.Parallel()

	// Setup mocks
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()

	factory.SetSupportedPartners([]string{"partner1", "partner2"})

	// Setup database partner info
	partner1ID := uuid.New()
	partner2ID := uuid.New()

	repo.SetPartnerInfo("ecomm", []models.PartnerAttributeMap{
		{PartnerCode: "partner1", PartnerID: &partner1ID},
		{PartnerCode: "partner2", PartnerID: &partner2ID},
	})

	// Setup adapters
	adapter1 := mocks.NewMockPartnerAdapter("partner1")
	adapter1.SetServiceabilityResult(adapter1.CreateServiceableResult())
	factory.SetAdapter("partner1", adapter1)

	adapter2 := mocks.NewMockPartnerAdapter("partner2")
	adapter2.SetServiceabilityResult(adapter2.CreateServiceableResult())
	factory.SetAdapter("partner2", adapter2)

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
	})

	// Verify
	require.NoError(t, err, "CheckServiceability should not return error")
	require.NotNil(t, response, "Response should not be nil")

	assert.True(t, response.Success, "Should be successful")
	assert.Len(t, response.Partners, 2, "Should have two partners")

	// Verify database partner IDs are used
	partnerIDMap := make(map[string]string)
	for _, partner := range response.Partners {
		partnerIDMap[partner.PartnerCode] = partner.PartnerID
	}

	assert.Equal(t, partner1ID.String(), partnerIDMap["partner1"], "Partner1 should use database ID")
	assert.Equal(t, partner2ID.String(), partnerIDMap["partner2"], "Partner2 should use database ID")
}
