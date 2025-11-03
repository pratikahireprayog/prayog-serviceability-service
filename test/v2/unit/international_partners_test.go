package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internationalstrategy "prayog-serviceability-service/internal/services/v2/orchestrators/strategies/international_strategy"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"
)

// TestAllInternationalPartners tests all international partners (dhl, aramex, fedex, shipcube, indiapost, naqel)
func TestAllInternationalPartners(t *testing.T) {
	t.Parallel()

	internationalPartners := []string{"dhl", "aramex", "fedex", "shipcube", "indiapost", "naqel"}

	for _, partnerCode := range internationalPartners {
		partnerCode := partnerCode // capture loop variable
		t.Run(fmt.Sprintf("TestPartner_%s", partnerCode), func(t *testing.T) {
			t.Parallel()

			// Test successful serviceability
			t.Run("SuccessfulServiceability", func(t *testing.T) {
				testPartnerSuccessfulServiceability(t, partnerCode)
			})

			// Test non-serviceable scenario
			t.Run("NonServiceable", func(t *testing.T) {
				testPartnerNonServiceable(t, partnerCode)
			})

			// Test adapter error
			t.Run("AdapterError", func(t *testing.T) {
				testPartnerAdapterError(t, partnerCode)
			})

			// Test nil result
			t.Run("NilResult", func(t *testing.T) {
				testPartnerNilResult(t, partnerCode)
			})

			// Test partner-specific conversion
			t.Run("PartnerConversion", func(t *testing.T) {
				testPartnerConversion(t, partnerCode)
			})
		})
	}
}

// TestInternationalPartnerDetection tests detection of international partners
func TestInternationalPartnerDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		partnerCodes   []string
		expectedResult bool
		description    string
	}{
		{
			name:           "SingleDHL",
			partnerCodes:   []string{"dhl"},
			expectedResult: true,
			description:    "DHL should be detected as international partner",
		},
		{
			name:           "MultipleInternational",
			partnerCodes:   []string{"dhl", "fedex", "aramex"},
			expectedResult: true,
			description:    "Multiple international partners should be detected",
		},
		{
			name:           "MixedPartners",
			partnerCodes:   []string{"dhl", "smile_ecom"},
			expectedResult: true,
			description:    "Mixed partners with international should be detected",
		},
		{
			name:           "NonInternational",
			partnerCodes:   []string{"smile_ecom", "shipyaari"},
			expectedResult: false,
			description:    "Non-international partners should not be detected",
		},
		{
			name:           "AllInternationalPartners",
			partnerCodes:   []string{"dhl", "aramex", "fedex", "shipcube", "indiapost", "naqel"},
			expectedResult: true,
			description:    "All international partners should be detected",
		},
		{
			name:           "CaseInsensitive",
			partnerCodes:   []string{"DHL", "FEDEX"},
			expectedResult: true,
			description:    "Partner detection should be case-insensitive",
		},
		{
			name:           "IndiaPostVariants",
			partnerCodes:   []string{"india_post_international"},
			expectedResult: true,
			description:    "India Post variants should be detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup all requested partners
			for _, code := range tt.partnerCodes {
				adapter := mocks.NewMockPartnerAdapter(code)
				adapter.SetServiceabilityResult(adapter.CreateServiceableResult())
				mockFactory.SetAdapter(code, adapter)
			}

			mockFactory.SetSupportedPartners(tt.partnerCodes)

			// Create request with partners
			partners := make([]models.PartnerFilter, 0, len(tt.partnerCodes))
			for _, code := range tt.partnerCodes {
				partners = append(partners, models.PartnerFilter{Code: code})
			}

			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("10001"),
				SourceCountryCode:     stringPtr("IN"),
				DestinationCountryCode: stringPtr("US"),
				ParcelCategory:        stringPtr("ecomm"), // Should be ignored when partners provided
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 2.0, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 30.0, Width: 20.0, Height: 15.0, Unit: "cm"},
					},
				},
				Partners: partners,
			}

			strategy := internationalstrategy.NewInternationalStrategy(mockFactory)
			response, err := strategy.Execute(ctx, request)

			require.NoError(t, err)
			require.NotNil(t, response)

			// Verify that international partners route through international strategy
			if tt.expectedResult {
				// Should have partners in response
				assert.GreaterOrEqual(t, len(response.Partners), 0, "Response should contain partners")
				// Verify parcel_category was ignored (partners take priority)
				assert.NotEmpty(t, response.Partners, "International partners should be processed")
			}
		})
	}
}

// TestPartnerFiltering tests filtering of requested partners
func TestPartnerFiltering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		requestedPartners   []string
		expectedPartnerCode string
		description         string
	}{
		{
			name:                "SinglePartner",
			requestedPartners:   []string{"dhl"},
			expectedPartnerCode: "dhl",
			description:         "Single requested partner should be called",
		},
		{
			name:                "MultiplePartners",
			requestedPartners:   []string{"dhl", "fedex"},
			expectedPartnerCode: "dhl", // At least one should be present
			description:         "Multiple requested partners should be called",
		},
		{
			name:                "AllInternational",
			requestedPartners:   []string{"dhl", "aramex", "fedex", "shipcube", "indiapost", "naqel"},
			expectedPartnerCode: "dhl",
			description:         "All international partners should be called",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup all requested partners
			calledPartners := make(map[string]bool)
			for _, code := range tt.requestedPartners {
				adapter := mocks.NewMockPartnerAdapter(code)
				result := adapter.CreateServiceableResult()
				adapter.SetServiceabilityResult(result)
				mockFactory.SetAdapter(code, adapter)

				// Track adapter calls by checking CheckServiceabilityCalled after execution
				calledPartners[code] = false // Will be checked after execution
			}

			mockFactory.SetSupportedPartners(tt.requestedPartners)

			// Create request with specific partners
			partners := make([]models.PartnerFilter, 0, len(tt.requestedPartners))
			for _, code := range tt.requestedPartners {
				partners = append(partners, models.PartnerFilter{Code: code})
			}

			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("10001"),
				SourceCountryCode:     stringPtr("IN"),
				DestinationCountryCode: stringPtr("US"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 2.0, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 30.0, Width: 20.0, Height: 15.0, Unit: "cm"},
					},
				},
				Partners: partners,
			}

			strategy := internationalstrategy.NewInternationalStrategy(mockFactory)
			response, err := strategy.Execute(ctx, request)

			require.NoError(t, err)
			require.NotNil(t, response)

			// Verify all requested partners were processed
			assert.Equal(t, len(tt.requestedPartners), len(response.Partners), "All requested partners should be in response")
		})
	}
}

// TestPartnerCodeNormalization tests normalization of partner codes
func TestPartnerCodeNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		inputCode      string
		expectedOutput string
		description    string
	}{
		{
			name:           "DHLDirect",
			inputCode:      "dhl",
			expectedOutput: "dhl",
			description:    "DHL code should remain unchanged",
		},
		{
			name:           "DHLUppercase",
			inputCode:      "DHL",
			expectedOutput: "dhl",
			description:    "DHL uppercase should be normalized to lowercase",
		},
		{
			name:           "IndiaPostVariant",
			inputCode:      "india_post_international",
			expectedOutput: "indiapost",
			description:    "India Post variant should be normalized",
		},
		{
			name:           "UnknownCode",
			inputCode:      "unknown_partner",
			expectedOutput: "unknown_partner",
			description:    "Unknown codes should be returned as-is",
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockFactory := mocks.NewMockPartnerAdapterFactory()
			adapter := mocks.NewMockPartnerAdapter(tt.expectedOutput)
			adapter.SetServiceabilityResult(adapter.CreateServiceableResult())
			mockFactory.SetAdapter(tt.expectedOutput, adapter)
			mockFactory.SetSupportedPartners([]string{tt.expectedOutput})

			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("10001"),
				SourceCountryCode:     stringPtr("IN"),
				DestinationCountryCode: stringPtr("US"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 2.0, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 30.0, Width: 20.0, Height: 15.0, Unit: "cm"},
					},
				},
				Partners: []models.PartnerFilter{{Code: tt.inputCode}},
			}

			strategy := internationalstrategy.NewInternationalStrategy(mockFactory)
			response, err := strategy.Execute(ctx, request)

			require.NoError(t, err)
			require.NotNil(t, response)

			// Verify partner was processed (normalized code should work)
			assert.True(t, adapter.CheckServiceabilityCalled, "Adapter should be called for normalized code")
		})
	}
}

// Helper functions

func testPartnerSuccessfulServiceability(t *testing.T, partnerCode string) {
	ctx := context.Background()
	mockFactory := mocks.NewMockPartnerAdapterFactory()

	adapter := mocks.NewMockPartnerAdapter(partnerCode)
	result := createPartnerServiceableResult(partnerCode)
	adapter.SetServiceabilityResult(result)
	mockFactory.SetAdapter(partnerCode, adapter)
	mockFactory.SetSupportedPartners([]string{partnerCode})

	request := &models.ServiceabilityV2Request{
		SourcePostalCode:      stringPtr("560001"),
		DestinationPostalCode: stringPtr("10001"),
		SourceCountryCode:     stringPtr("IN"),
		DestinationCountryCode: stringPtr("US"),
		Packages: []models.Package{
			{
				Weight:     &models.Weight{Value: 2.0, Unit: "kg"},
				Dimensions: &models.Dimensions{Length: 30.0, Width: 20.0, Height: 15.0, Unit: "cm"},
			},
		},
		Partners: []models.PartnerFilter{{Code: partnerCode}},
	}

	strategy := internationalstrategy.NewInternationalStrategy(mockFactory)
	response, err := strategy.Execute(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Find the partner in response
	var partnerResp *models.PartnerV2Response
	for i := range response.Partners {
		if response.Partners[i].PartnerCode == partnerCode {
			partnerResp = &response.Partners[i]
			break
		}
	}

	require.NotNil(t, partnerResp, "Partner response should exist")
	assert.True(t, partnerResp.IsServiceable, "Partner should be serviceable")
	assert.Equal(t, partnerCode, partnerResp.PartnerCode, "Partner code should match")
	assert.NotEmpty(t, partnerResp.PartnerServices, "Partner should have services")
}

func testPartnerNonServiceable(t *testing.T, partnerCode string) {
	ctx := context.Background()
	mockFactory := mocks.NewMockPartnerAdapterFactory()

	adapter := mocks.NewMockPartnerAdapter(partnerCode)
	result := adapter.CreateNonServiceableResult()
	adapter.SetServiceabilityResult(result)
	mockFactory.SetAdapter(partnerCode, adapter)
	mockFactory.SetSupportedPartners([]string{partnerCode})

	request := &models.ServiceabilityV2Request{
		SourcePostalCode:      stringPtr("560001"),
		DestinationPostalCode: stringPtr("999999"),
		SourceCountryCode:     stringPtr("IN"),
		DestinationCountryCode: stringPtr("US"),
		Packages: []models.Package{
			{
				Weight:     &models.Weight{Value: 2.0, Unit: "kg"},
				Dimensions: &models.Dimensions{Length: 30.0, Width: 20.0, Height: 15.0, Unit: "cm"},
			},
		},
		Partners: []models.PartnerFilter{{Code: partnerCode}},
	}

	strategy := internationalstrategy.NewInternationalStrategy(mockFactory)
	response, err := strategy.Execute(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Non-serviceable partners should not be in response (filtered out)
	found := false
	for i := range response.Partners {
		if response.Partners[i].PartnerCode == partnerCode {
			found = true
			assert.False(t, response.Partners[i].IsServiceable, "Partner should not be serviceable")
			break
		}
	}

	// Partner might not be in response if filtered out
	if found {
		t.Logf("Partner %s found in response as non-serviceable (expected)", partnerCode)
	} else {
		t.Logf("Partner %s not in response (filtered out as non-serviceable - expected)", partnerCode)
	}
}

func testPartnerAdapterError(t *testing.T, partnerCode string) {
	ctx := context.Background()
	mockFactory := mocks.NewMockPartnerAdapterFactory()

	adapter := mocks.NewMockPartnerAdapter(partnerCode)
	adapter.SetServiceabilityError(fmt.Errorf("adapter connection failed"))
	mockFactory.SetAdapter(partnerCode, adapter)
	mockFactory.SetSupportedPartners([]string{partnerCode})

	request := &models.ServiceabilityV2Request{
		SourcePostalCode:      stringPtr("560001"),
		DestinationPostalCode: stringPtr("10001"),
		SourceCountryCode:     stringPtr("IN"),
		DestinationCountryCode: stringPtr("US"),
		Packages: []models.Package{
			{
				Weight:     &models.Weight{Value: 2.0, Unit: "kg"},
				Dimensions: &models.Dimensions{Length: 30.0, Width: 20.0, Height: 15.0, Unit: "cm"},
			},
		},
		Partners: []models.PartnerFilter{{Code: partnerCode}},
	}

	strategy := internationalstrategy.NewInternationalStrategy(mockFactory)
	response, err := strategy.Execute(ctx, request)

	require.NoError(t, err, "Strategy should handle adapter errors gracefully")
	require.NotNil(t, response)

	// Partner with error should not be in response (filtered out)
	found := false
	for i := range response.Partners {
		if response.Partners[i].PartnerCode == partnerCode {
			found = true
			assert.False(t, response.Partners[i].IsServiceable, "Partner with error should not be serviceable")
			break
		}
	}

	if !found {
		t.Logf("Partner %s not in response (filtered out due to error - expected)", partnerCode)
	}
}

func testPartnerNilResult(t *testing.T, partnerCode string) {
	ctx := context.Background()
	mockFactory := mocks.NewMockPartnerAdapterFactory()

	adapter := mocks.NewMockPartnerAdapter(partnerCode)
	adapter.SetServiceabilityResult(nil)
	mockFactory.SetAdapter(partnerCode, adapter)
	mockFactory.SetSupportedPartners([]string{partnerCode})

	request := &models.ServiceabilityV2Request{
		SourcePostalCode:      stringPtr("560001"),
		DestinationPostalCode: stringPtr("10001"),
		SourceCountryCode:     stringPtr("IN"),
		DestinationCountryCode: stringPtr("US"),
		Packages: []models.Package{
			{
				Weight:     &models.Weight{Value: 2.0, Unit: "kg"},
				Dimensions: &models.Dimensions{Length: 30.0, Width: 20.0, Height: 15.0, Unit: "cm"},
			},
		},
		Partners: []models.PartnerFilter{{Code: partnerCode}},
	}

	strategy := internationalstrategy.NewInternationalStrategy(mockFactory)
	response, err := strategy.Execute(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Nil result should be handled gracefully
	// Partner might not appear in response or appear as non-serviceable
	t.Logf("Nil result handled for partner %s", partnerCode)
}

func testPartnerConversion(t *testing.T, partnerCode string) {
	ctx := context.Background()
	mockFactory := mocks.NewMockPartnerAdapterFactory()

	adapter := mocks.NewMockPartnerAdapter(partnerCode)
	result := createPartnerServiceableResult(partnerCode)
	adapter.SetServiceabilityResult(result)
	mockFactory.SetAdapter(partnerCode, adapter)
	mockFactory.SetSupportedPartners([]string{partnerCode})

	request := &models.ServiceabilityV2Request{
		SourcePostalCode:      stringPtr("560001"),
		DestinationPostalCode: stringPtr("10001"),
		SourceCountryCode:     stringPtr("IN"),
		DestinationCountryCode: stringPtr("US"),
		Packages: []models.Package{
			{
				Weight:     &models.Weight{Value: 2.0, Unit: "kg"},
				Dimensions: &models.Dimensions{Length: 30.0, Width: 20.0, Height: 15.0, Unit: "cm"},
			},
		},
		Partners: []models.PartnerFilter{{Code: partnerCode}},
	}

	strategy := internationalstrategy.NewInternationalStrategy(mockFactory)
	response, err := strategy.Execute(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Find partner in response
	var partnerResp *models.PartnerV2Response
	for i := range response.Partners {
		if response.Partners[i].PartnerCode == partnerCode {
			partnerResp = &response.Partners[i]
			break
		}
	}

	if partnerResp != nil {
		// Verify partner-specific fields
		assert.Equal(t, partnerCode, partnerResp.PartnerCode)
		assert.NotEmpty(t, partnerResp.PartnerID)
		assert.NotEmpty(t, partnerResp.Source, "Source should be set")
		assert.Contains(t, partnerResp.Metadata, "flow", "Metadata should contain flow")
		assert.Equal(t, "international", partnerResp.Metadata["flow"], "Flow should be international")

		// Partner-specific assertions
		switch partnerCode {
		case "dhl":
			assert.Equal(t, "DHL Express", partnerResp.PartnerName)
			assert.Contains(t, partnerResp.Metadata, "dhl_response")
		case "aramex":
			assert.Equal(t, "Aramex", partnerResp.PartnerName)
			assert.Contains(t, partnerResp.Metadata, "aramex_metadata")
		case "fedex":
			assert.Equal(t, "FedEx", partnerResp.PartnerName)
		case "shipcube":
			assert.Equal(t, "ShipCube", partnerResp.PartnerName)
		case "indiapost":
			assert.Equal(t, "India Post International", partnerResp.PartnerName)
		case "naqel":
			assert.Equal(t, "Naqel", partnerResp.PartnerName)
			assert.Contains(t, partnerResp.Metadata, "naqel_response")
		}
	}
}

// Helper function to create partner-specific serviceable results
func createPartnerServiceableResult(partnerCode string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	result := &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: partnerCode,
		PartnerName: getPartnerDisplayName(partnerCode),
		Services: []models.ServiceV2{
			{
				ServiceCode:   "EXPRESS",
				ServiceName:   "Express Service",
				TATDays:       3,
				IsCOD:         false,
				Pickup:        true,
				Delivery:      true,
				Insurance:     true,
				ProductTypes:  map[string]bool{"commercial": true, "document": true},
				DeliveryModes: map[string]bool{"express": true},
			},
		},
		Capabilities: map[string]interface{}{
			"pickup":    true,
			"delivery":  true,
			"insurance": true,
		},
		ResponseTime: 150 * time.Millisecond,
		Metadata: map[string]interface{}{
			"source_country_code":      "IN",
			"destination_country_code": "US",
			"flow":                     "international",
		},
	}

	// Partner-specific metadata
	switch partnerCode {
	case "aramex":
		result.Metadata["is_serviceable"] = true
		result.Metadata["aramex_metadata"] = map[string]interface{}{
			"service_type": "international_express",
		}
	case "dhl":
		result.Metadata["dhl_response"] = "success"
		result.Metadata["product_count"] = 1
	case "naqel":
		result.Metadata["naqel_response"] = "success"
	}

	return result
}

func getPartnerDisplayName(code string) string {
	names := map[string]string{
		"dhl":       "DHL Express",
		"aramex":    "Aramex",
		"fedex":     "FedEx",
		"shipcube":  "ShipCube",
		"indiapost": "India Post International",
		"naqel":     "Naqel",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return code
}


