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

// TestDHLCapabilitiesStructure tests the flattened DHL capabilities structure
func TestDHLCapabilitiesStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		setupMocks           func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectedCapabilities []string
		description          string
	}{
		{
			name: "DHL_FlattenedPickupCapabilities",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLAdapters()
			},
			expectedCapabilities: []string{
				"next_business_day",
				"local_cutoff_date_and_time",
				"pickup_earliest",
				"pickup_latest",
				"pickup_cutoff_same_day_outbound_processing",
				"origin_service_area_code",
				"origin_facility_area_code",
				"pickup_additional_days",
				"pickup_day_of_week",
			},
			description: "DHL should return flattened pickup capabilities as individual fields",
		},
		{
			name: "DHL_FlattenedDeliveryCapabilities",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLAdapters()
			},
			expectedCapabilities: []string{
				"delivery_type_code",
				"estimated_delivery_date_and_time",
				"destination_service_area_code",
				"destination_facility_area_code",
				"delivery_additional_days",
				"delivery_day_of_week",
				"total_transit_days",
			},
			description: "DHL should return flattened delivery capabilities as individual fields",
		},
		{
			name: "DHL_InternationalFlowCapabilities",
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})
				factory.SetupDHLInternationalFlowAdapters()
			},
			expectedCapabilities: []string{
				"next_business_day",
				"local_cutoff_date_and_time",
				"pickup_earliest",
				"pickup_latest",
				"delivery_type_code",
				"estimated_delivery_date_and_time",
				"total_transit_days",
			},
			description: "DHL international flow should return comprehensive flattened capabilities",
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
				false, // Return all partners for capability testing
			)

			// Create test request with DHL requirements
			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("110001"),
				CountryCode:           stringPtr("IN"),
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
			}

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, request)

			// Verify
			require.NoError(t, err, "Should not return error: %s", tt.description)
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
			assert.NotNil(t, dhlPartner.Capabilities, "DHL should have capabilities")

			// Verify expected capabilities are present and flattened
			for _, capability := range tt.expectedCapabilities {
				assert.Contains(t, dhlPartner.Capabilities, capability,
					"DHL capabilities should contain %s: %s", capability, tt.description)

				// Verify capability is not nested (should be primitive value, not map)
				capValue := dhlPartner.Capabilities[capability]
				assert.NotNil(t, capValue, "Capability %s should have a value", capability)

				// Verify it's not a nested map (flattened structure)
				_, isMap := capValue.(map[string]interface{})
				assert.False(t, isMap, "Capability %s should be flattened, not nested: %s", capability, tt.description)
			}

			// Verify no nested pickup/delivery objects exist
			_, hasPickupObject := dhlPartner.Capabilities["pickup"]
			_, hasDeliveryObject := dhlPartner.Capabilities["delivery"]
			assert.False(t, hasPickupObject, "DHL capabilities should not have nested pickup object")
			assert.False(t, hasDeliveryObject, "DHL capabilities should not have nested delivery object")
		})
	}
}

// TestDHLCapabilitiesDataTypes tests that DHL capabilities have correct data types
func TestDHLCapabilitiesDataTypes(t *testing.T) {
	t.Parallel()

	// Setup
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()
	factory.SetSupportedPartners([]string{"dhl"})
	factory.SetupDHLAdapters()

	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		factory,
		repo,
		30*time.Second,
		false,
	)

	request := &models.ServiceabilityV2Request{
		SourcePostalCode:      stringPtr("560001"),
		DestinationPostalCode: stringPtr("110001"),
		CountryCode:           stringPtr("IN"),
		Packages: []models.Package{
			{
				Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
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
	require.NotNil(t, dhlPartner.Capabilities)

	// Test specific data types
	testCases := []struct {
		capability   string
		expectedType string
	}{
		{"next_business_day", "bool"},
		{"pickup_additional_days", "int"},
		{"pickup_day_of_week", "int"},
		{"delivery_additional_days", "int"},
		{"delivery_day_of_week", "int"},
		{"total_transit_days", "int"},
		{"local_cutoff_date_and_time", "string"},
		{"pickup_earliest", "string"},
		{"pickup_latest", "string"},
		{"estimated_delivery_date_and_time", "string"},
		{"origin_service_area_code", "string"},
		{"destination_service_area_code", "string"},
	}

	for _, tc := range testCases {
		t.Run("DataType_"+tc.capability, func(t *testing.T) {
			value, exists := dhlPartner.Capabilities[tc.capability]
			assert.True(t, exists, "Capability %s should exist", tc.capability)

			switch tc.expectedType {
			case "bool":
				_, ok := value.(bool)
				assert.True(t, ok, "Capability %s should be boolean, got %T", tc.capability, value)
			case "int":
				_, ok := value.(int)
				assert.True(t, ok, "Capability %s should be integer, got %T", tc.capability, value)
			case "string":
				_, ok := value.(string)
				assert.True(t, ok, "Capability %s should be string, got %T", tc.capability, value)
			}
		})
	}
}

// TestDHLCapabilitiesVsTraditionalPartners compares DHL flattened capabilities with traditional partners
func TestDHLCapabilitiesVsTraditionalPartners(t *testing.T) {
	t.Parallel()

	// Setup mixed adapters
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()
	factory.SetSupportedPartners([]string{"dhl", "shipyaari"})
	factory.SetupMixedCapabilitiesAdapters()

	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		factory,
		repo,
		30*time.Second,
		false,
	)

	request := &models.ServiceabilityV2Request{
		SourcePostalCode:      stringPtr("560001"),
		DestinationPostalCode: stringPtr("110001"),
		CountryCode:           stringPtr("IN"),
		Packages: []models.Package{
			{
				Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
				Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
			},
		},
	}

	// Execute
	ctx := context.Background()
	response, err := orchestrator.CheckServiceability(ctx, request)

	// Verify
	require.NoError(t, err)
	require.Len(t, response.Partners, 2, "Should have both DHL and Shipyaari")

	var dhlPartner, shipyaariPartner *models.PartnerV2Response
	for i := range response.Partners {
		switch response.Partners[i].PartnerCode {
		case "dhl":
			dhlPartner = &response.Partners[i]
		case "shipyaari":
			shipyaariPartner = &response.Partners[i]
		}
	}

	require.NotNil(t, dhlPartner, "DHL partner should be present")
	require.NotNil(t, shipyaariPartner, "Shipyaari partner should be present")

	// Verify DHL has flattened capabilities
	assert.Contains(t, dhlPartner.Capabilities, "next_business_day", "DHL should have flattened pickup capabilities")
	assert.Contains(t, dhlPartner.Capabilities, "total_transit_days", "DHL should have flattened delivery capabilities")

	// Verify traditional partner has nested capabilities
	pickup, hasPickup := shipyaariPartner.Capabilities["pickup"]
	assert.True(t, hasPickup, "Traditional partner should have pickup object")

	pickupMap, isMap := pickup.(map[string]interface{})
	assert.True(t, isMap, "Traditional partner pickup should be nested object")
	assert.Contains(t, pickupMap, "available", "Traditional partner pickup should have nested fields")

	// Verify DHL doesn't have nested structures
	_, hasDHLPickupObject := dhlPartner.Capabilities["pickup"]
	assert.False(t, hasDHLPickupObject, "DHL should not have nested pickup object")
}
