package unit

import (
	"context"
	"testing"
	"time"

	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPartnerIDUUIDToStringMapping tests the conversion of PartnerID from UUID to string
func TestPartnerIDUUIDToStringMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupAdapter func() (*mocks.MockPartnerAdapter, *common.PartnerServiceabilityResult)
		validateID   func(t *testing.T, partnerID string, expectedUUID uuid.UUID)
		description  string
	}{
		{
			name: "Valid UUID conversion",
			setupAdapter: func() (*mocks.MockPartnerAdapter, *common.PartnerServiceabilityResult) {
				adapter := mocks.NewMockPartnerAdapter("test_partner")
				expectedUUID := uuid.New()

				result := &common.PartnerServiceabilityResult{
					PartnerID:    &expectedUUID,
					PartnerCode:  "test_partner",
					PartnerName:  "Test Partner",
					Services:     []models.ServiceV2{},
					Capabilities: make(map[string]interface{}),
					ResponseTime: 100 * time.Millisecond,
				}

				adapter.SetServiceabilityResult(result)
				return adapter, result
			},
			validateID: func(t *testing.T, partnerID string, expectedUUID uuid.UUID) {
				assert.Equal(t, expectedUUID.String(), partnerID, "PartnerID should be string representation of UUID")
				assert.Len(t, partnerID, 36, "PartnerID should be 36 characters (UUID string format)")
				assert.Contains(t, partnerID, "-", "PartnerID should contain UUID hyphens")

				// Verify it can be parsed back to UUID
				parsedUUID, err := uuid.Parse(partnerID)
				assert.NoError(t, err, "PartnerID should be parseable as UUID")
				assert.Equal(t, expectedUUID, parsedUUID, "Parsed UUID should match original")
			},
			description: "PartnerID should be correctly converted from UUID to string",
		},
		{
			name: "Nil UUID handling",
			setupAdapter: func() (*mocks.MockPartnerAdapter, *common.PartnerServiceabilityResult) {
				adapter := mocks.NewMockPartnerAdapter("test_partner")

				result := &common.PartnerServiceabilityResult{
					PartnerID:    nil, // Nil UUID pointer
					PartnerCode:  "test_partner",
					PartnerName:  "Test Partner",
					Services:     []models.ServiceV2{},
					Capabilities: make(map[string]interface{}),
					ResponseTime: 100 * time.Millisecond,
				}

				adapter.SetServiceabilityResult(result)
				return adapter, result
			},
			validateID: func(t *testing.T, partnerID string, expectedUUID uuid.UUID) {
				// When UUID is nil, should use database info or fallback to "unknown"
				if partnerID != "unknown" {
					// If not "unknown", should be a valid UUID from database
					_, err := uuid.Parse(partnerID)
					assert.NoError(t, err, "PartnerID should be valid UUID even when result UUID is nil")
				} else {
					assert.Equal(t, "unknown", partnerID, "PartnerID should be 'unknown' when UUID is nil and no database info")
				}
			},
			description: "Nil PartnerID should be handled gracefully",
		},
		{
			name: "Zero UUID handling",
			setupAdapter: func() (*mocks.MockPartnerAdapter, *common.PartnerServiceabilityResult) {
				adapter := mocks.NewMockPartnerAdapter("test_partner")
				zeroUUID := uuid.UUID{} // Zero UUID

				result := &common.PartnerServiceabilityResult{
					PartnerID:    &zeroUUID,
					PartnerCode:  "test_partner",
					PartnerName:  "Test Partner",
					Services:     []models.ServiceV2{},
					Capabilities: make(map[string]interface{}),
					ResponseTime: 100 * time.Millisecond,
				}

				adapter.SetServiceabilityResult(result)
				return adapter, result
			},
			validateID: func(t *testing.T, partnerID string, expectedUUID uuid.UUID) {
				assert.Equal(t, "00000000-0000-0000-0000-000000000000", partnerID, "Zero UUID should be converted to zero UUID string")

				// Verify it can be parsed back to UUID
				parsedUUID, err := uuid.Parse(partnerID)
				assert.NoError(t, err, "Zero UUID string should be parseable")
				assert.Equal(t, uuid.UUID{}, parsedUUID, "Parsed UUID should be zero UUID")
			},
			description: "Zero UUID should be correctly handled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			factory := mocks.NewMockPartnerAdapterFactory()
			repo := mocks.NewMockPartnerAttributeMapRepository()
			adapter, result := tt.setupAdapter()

			factory.SetSupportedPartners([]string{"test_partner"})
			factory.SetAdapter("test_partner", adapter)

			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				factory,
				repo,
				30*time.Second,
				false,
			)

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			})

			// Verify
			require.NoError(t, err, "Should not return error")
			require.NotEmpty(t, response.Partners, "Should have partners in response")

			// Find test partner
			var testPartner *models.PartnerV2Response
			for i := range response.Partners {
				if response.Partners[i].PartnerCode == "test_partner" {
					testPartner = &response.Partners[i]
					break
				}
			}

			require.NotNil(t, testPartner, "Test partner should be in response")

			// Validate UUID conversion
			var expectedUUID uuid.UUID
			if result.PartnerID != nil {
				expectedUUID = *result.PartnerID
			}
			tt.validateID(t, testPartner.PartnerID, expectedUUID)
		})
	}
}

// TestPartnerNameFieldMapping tests the PartnerName field mapping
func TestPartnerNameFieldMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		partnerCode  string
		partnerName  string
		expectedName string
		description  string
	}{
		{
			name:         "DHL partner name mapping",
			partnerCode:  "dhl",
			partnerName:  "DHL Express",
			expectedName: "DHL Express",
			description:  "DHL should use full partner name",
		},
		{
			name:         "Shipyaari partner name mapping",
			partnerCode:  "shipyaari",
			partnerName:  "Shipyaari Logistics",
			expectedName: "Shipyaari Logistics",
			description:  "Shipyaari should use full partner name",
		},
		{
			name:         "SmileCargo partner name mapping",
			partnerCode:  "smile_cargo",
			partnerName:  "Smile Cargo Services",
			expectedName: "Smile Cargo Services",
			description:  "SmileCargo should use full partner name",
		},
		{
			name:         "Empty partner name handling",
			partnerCode:  "test_partner",
			partnerName:  "",
			expectedName: "", // Empty name should remain empty in response
			description:  "Empty partner name should be handled correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			factory := mocks.NewMockPartnerAdapterFactory()
			repo := mocks.NewMockPartnerAttributeMapRepository()

			adapter := mocks.NewMockPartnerAdapter(tt.partnerCode)
			adapter.PartnerName = tt.partnerName

			result := adapter.CreateServiceableResult()
			result.PartnerName = tt.partnerName
			adapter.SetServiceabilityResult(result)

			factory.SetSupportedPartners([]string{tt.partnerCode})
			factory.SetAdapter(tt.partnerCode, adapter)

			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				factory,
				repo,
				30*time.Second,
				false,
			)

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			})

			// Verify
			require.NoError(t, err, "Should not return error")
			require.NotEmpty(t, response.Partners, "Should have partners in response")

			// Find target partner
			var targetPartner *models.PartnerV2Response
			for i := range response.Partners {
				if response.Partners[i].PartnerCode == tt.partnerCode {
					targetPartner = &response.Partners[i]
					break
				}
			}

			require.NotNil(t, targetPartner, "Target partner should be in response")
			assert.Equal(t, tt.expectedName, targetPartner.PartnerName, tt.description)
		})
	}
}

// TestDatabasePartnerInfoPriority tests that database partner info takes priority over adapter results
func TestDatabasePartnerInfoPriority(t *testing.T) {
	t.Parallel()

	// Setup
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()

	// Create database partner info
	databasePartnerID := uuid.New()
	adapterPartnerID := uuid.New()

	factory.SetSupportedPartners([]string{"test_partner"})

	// Setup database info (should take priority)
	repo.SetPartnerInfo("ecomm", []models.PartnerAttributeMap{
		{PartnerCode: "test_partner", PartnerID: &databasePartnerID},
	})

	// Setup adapter with different UUID (should be overridden)
	adapter := mocks.NewMockPartnerAdapter("test_partner")
	result := adapter.CreateServiceableResult()
	result.PartnerID = &adapterPartnerID // Different from database
	adapter.SetServiceabilityResult(result)
	factory.SetAdapter("test_partner", adapter)

	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		factory,
		repo,
		30*time.Second,
		false,
	)

	// Execute
	ctx := context.Background()
	response, err := orchestrator.CheckServiceability(ctx, &models.ServiceabilityV2Request{
		PostalCode:     stringPtr("12345"),
		CountryCode:    stringPtr("IN"),
		ParcelCategory: stringPtr("ecomm"),
		Packages: []models.Package{
			{
				Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
				Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
			},
		},
	})

	// Verify
	require.NoError(t, err, "Should not return error")
	require.NotEmpty(t, response.Partners, "Should have partners in response")

	// Find test partner
	var testPartner *models.PartnerV2Response
	for i := range response.Partners {
		if response.Partners[i].PartnerCode == "test_partner" {
			testPartner = &response.Partners[i]
			break
		}
	}

	require.NotNil(t, testPartner, "Test partner should be in response")

	// Database partner ID should take priority
	assert.Equal(t, databasePartnerID.String(), testPartner.PartnerID,
		"Should use database PartnerID, not adapter PartnerID")
	assert.NotEqual(t, adapterPartnerID.String(), testPartner.PartnerID,
		"Should not use adapter PartnerID when database info is available")
}

// TestPartnerInfoMappingWithErrors tests partner info mapping when errors occur
func TestPartnerInfoMappingWithErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupAdapter func() *mocks.MockPartnerAdapter
		verifyResult func(t *testing.T, partner *models.PartnerV2Response)
		description  string
	}{
		{
			name: "Partner with validation error",
			setupAdapter: func() *mocks.MockPartnerAdapter {
				adapter := mocks.NewMockPartnerAdapter("dhl")
				adapter.PartnerName = "DHL Express"

				// Create result with error
				result := adapter.CreateDHLNonServiceableResult()
				adapter.SetServiceabilityResult(result)
				return adapter
			},
			verifyResult: func(t *testing.T, partner *models.PartnerV2Response) {
				// Even with error, partner info should be mapped correctly
				assert.NotEmpty(t, partner.PartnerID, "PartnerID should be present even with error")
				assert.Equal(t, "dhl", partner.PartnerCode, "PartnerCode should be correct")
				assert.NotEmpty(t, partner.PartnerName, "PartnerName should be present even with error")
				assert.NotNil(t, partner.Error, "Error should be present")
				assert.Contains(t, *partner.Error, "DHL validation failed", "Error should contain validation message")

				// Verify PartnerID is valid UUID format
				_, err := uuid.Parse(partner.PartnerID)
				assert.NoError(t, err, "PartnerID should be valid UUID format even with error")
			},
			description: "Partner with validation error should still have correct info mapping",
		},
		{
			name: "Partner with API error",
			setupAdapter: func() *mocks.MockPartnerAdapter {
				adapter := mocks.NewMockPartnerAdapter("test_partner")
				adapter.PartnerName = "Test Partner"

				// Create result with API error
				partnerID := uuid.New()
				errorMsg := "API request failed with status 500"
				result := &common.PartnerServiceabilityResult{
					PartnerID:    &partnerID,
					PartnerCode:  "test_partner",
					PartnerName:  "Test Partner",
					Services:     []models.ServiceV2{},
					Capabilities: make(map[string]interface{}),
					ErrorMessage: &errorMsg,
					ResponseTime: 1000 * time.Millisecond,
				}

				adapter.SetServiceabilityResult(result)
				return adapter
			},
			verifyResult: func(t *testing.T, partner *models.PartnerV2Response) {
				assert.NotEmpty(t, partner.PartnerID, "PartnerID should be present with API error")
				assert.Equal(t, "test_partner", partner.PartnerCode, "PartnerCode should be correct")
				assert.NotEmpty(t, partner.PartnerName, "PartnerName should be present with API error")
				assert.NotNil(t, partner.Error, "Error should be present")
				assert.Contains(t, *partner.Error, "API request failed", "Error should contain API error message")

				// Verify PartnerID format
				_, err := uuid.Parse(partner.PartnerID)
				assert.NoError(t, err, "PartnerID should be valid UUID format with API error")
			},
			description: "Partner with API error should maintain proper info mapping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			factory := mocks.NewMockPartnerAdapterFactory()
			repo := mocks.NewMockPartnerAttributeMapRepository()

			adapter := tt.setupAdapter()
			factory.SetSupportedPartners([]string{adapter.PartnerCode})
			factory.SetAdapter(adapter.PartnerCode, adapter)

			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				factory,
				repo,
				30*time.Second,
				false, // Include error partners in response
			)

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			})

			// Verify
			require.NoError(t, err, "Should not return orchestrator error")
			require.NotEmpty(t, response.Partners, "Should have partners in response")

			// Find target partner
			var targetPartner *models.PartnerV2Response
			for i := range response.Partners {
				if response.Partners[i].PartnerCode == adapter.PartnerCode {
					targetPartner = &response.Partners[i]
					break
				}
			}

			require.NotNil(t, targetPartner, "Target partner should be in response")
			tt.verifyResult(t, targetPartner)
		})
	}
}

// TestPartnerInfoMappingConsistency tests consistency of partner info mapping across different scenarios
func TestPartnerInfoMappingConsistency(t *testing.T) {
	t.Parallel()

	// Setup multiple partners with different configurations
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()

	// Setup database info for some partners
	partner1ID := uuid.New()
	partner2ID := uuid.New()

	repo.SetPartnerInfo("ecomm", []models.PartnerAttributeMap{
		{PartnerCode: "partner1", PartnerID: &partner1ID},
		{PartnerCode: "partner2", PartnerID: &partner2ID},
		// partner3 has no database info
	})

	factory.SetSupportedPartners([]string{"partner1", "partner2", "partner3"})

	// Partner1: Has database info, serviceable
	adapter1 := mocks.NewMockPartnerAdapter("partner1")
	adapter1.PartnerName = "Partner One"
	result1 := adapter1.CreateServiceableResult()
	adapter1.SetServiceabilityResult(result1)
	factory.SetAdapter("partner1", adapter1)

	// Partner2: Has database info, non-serviceable
	adapter2 := mocks.NewMockPartnerAdapter("partner2")
	adapter2.PartnerName = "Partner Two"
	result2 := adapter2.CreateNonServiceableResult()
	adapter2.SetServiceabilityResult(result2)
	factory.SetAdapter("partner2", adapter2)

	// Partner3: No database info, serviceable
	adapter3 := mocks.NewMockPartnerAdapter("partner3")
	adapter3.PartnerName = "Partner Three"
	result3 := adapter3.CreateServiceableResult()
	adapter3.SetServiceabilityResult(result3)
	factory.SetAdapter("partner3", adapter3)

	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		factory,
		repo,
		30*time.Second,
		false, // Include all partners
	)

	// Execute
	ctx := context.Background()
	response, err := orchestrator.CheckServiceability(ctx, &models.ServiceabilityV2Request{
		PostalCode:     stringPtr("12345"),
		CountryCode:    stringPtr("IN"),
		ParcelCategory: stringPtr("ecomm"),
		Packages: []models.Package{
			{
				Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
				Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
			},
		},
	})

	// Verify
	require.NoError(t, err, "Should not return error")
	require.Len(t, response.Partners, 3, "Should have all three partners")

	// Create map for easier verification
	partnerMap := make(map[string]*models.PartnerV2Response)
	for i := range response.Partners {
		partnerMap[response.Partners[i].PartnerCode] = &response.Partners[i]
	}

	// Verify Partner1 (database info, serviceable)
	partner1 := partnerMap["partner1"]
	require.NotNil(t, partner1, "Partner1 should be in response")
	assert.Equal(t, partner1ID.String(), partner1.PartnerID, "Partner1 should use database ID")
	assert.Equal(t, "Partner One", partner1.PartnerName, "Partner1 name should be correct")
	assert.NotEmpty(t, partner1.Services, "Partner1 should have services")

	// Verify Partner2 (database info, non-serviceable)
	partner2 := partnerMap["partner2"]
	require.NotNil(t, partner2, "Partner2 should be in response")
	assert.Equal(t, partner2ID.String(), partner2.PartnerID, "Partner2 should use database ID")
	assert.Equal(t, "Partner Two", partner2.PartnerName, "Partner2 name should be correct")
	assert.Empty(t, partner2.Services, "Partner2 should have no services")

	// Verify Partner3 (no database info, serviceable)
	partner3 := partnerMap["partner3"]
	require.NotNil(t, partner3, "Partner3 should be in response")
	assert.NotEqual(t, "unknown", partner3.PartnerID, "Partner3 should have UUID from adapter")
	_, err = uuid.Parse(partner3.PartnerID)
	assert.NoError(t, err, "Partner3 PartnerID should be valid UUID")
	assert.Equal(t, "Partner Three", partner3.PartnerName, "Partner3 name should be correct")
	assert.NotEmpty(t, partner3.Services, "Partner3 should have services")

	// Verify all partners have unique IDs
	partnerIDs := make(map[string]bool)
	for _, partner := range response.Partners {
		assert.False(t, partnerIDs[partner.PartnerID], "Each partner should have unique PartnerID")
		partnerIDs[partner.PartnerID] = true
	}
}
