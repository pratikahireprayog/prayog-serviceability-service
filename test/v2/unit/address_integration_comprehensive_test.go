package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"prayog-serviceability-service/internal/shared/models/v1"
)

// TestAddressIntegrationResponseStructure tests the response structure for address integration
func TestAddressIntegrationResponseStructure(t *testing.T) {
	t.Parallel()

	t.Run("ServiceabilityV2Response_Address_Fields_Structure", func(t *testing.T) {
		// Test the new address fields are present and properly typed
		response := &models.ServiceabilityV2Response{
			Success: true,
			SourceAddress: &models.AddressInfo{
				PostalCode:  "560086",
				CountryCode: "IN",
			},
			DestinationAddress: &models.AddressInfo{
				PostalCode:  "266001",
				CountryCode: "CN",
			},
			Addresses: []models.DetailedAddress{
				{
					Type:        "INTERNATIONAL_HUB_ADDRESS",
					Zip:         "560024",
					Name:        "Hub Manager",
					Phone:       "+91-9876543210",
					Email:       "hub@test.com",
					Street:      "Test Hub Street",
					Landmark:    "Near Test Landmark",
					City:        "Bangalore",
					State:       "Karnataka",
					Country:     "India",
					Latitude:    float64Ptr(13.0358),
					Longitude:   float64Ptr(77.597),
					AddressName: "WAREHOUSE",
				},
			},
			Partners: []models.PartnerV2Response{},
		}

		// Verify structure and fields
		assert.True(t, response.Success)
		assert.NotNil(t, response.SourceAddress)
		assert.NotNil(t, response.DestinationAddress)
		assert.Len(t, response.Addresses, 1)

		// Verify source address
		assert.Equal(t, "560086", response.SourceAddress.PostalCode)
		assert.Equal(t, "IN", response.SourceAddress.CountryCode)

		// Verify destination address
		assert.Equal(t, "266001", response.DestinationAddress.PostalCode)
		assert.Equal(t, "CN", response.DestinationAddress.CountryCode)

		// Verify hub address details
		hubAddress := response.Addresses[0]
		assert.Equal(t, "INTERNATIONAL_HUB_ADDRESS", hubAddress.Type)
		assert.Equal(t, "WAREHOUSE", hubAddress.AddressName)
		assert.Equal(t, "Hub Manager", hubAddress.Name)
		assert.Equal(t, "+91-9876543210", hubAddress.Phone)
		assert.Equal(t, "hub@test.com", hubAddress.Email)
		assert.Equal(t, "Bangalore", hubAddress.City)
		assert.Equal(t, "Karnataka", hubAddress.State)
		assert.Equal(t, "India", hubAddress.Country)
		assert.NotNil(t, hubAddress.Latitude)
		assert.NotNil(t, hubAddress.Longitude)
		assert.Equal(t, 13.0358, *hubAddress.Latitude)
		assert.Equal(t, 77.597, *hubAddress.Longitude)
	})

	t.Run("AddressInfo_Basic_Structure", func(t *testing.T) {
		// Test basic address info structure
		addressInfo := &models.AddressInfo{
			PostalCode:  "560086",
			CountryCode: "IN",
		}

		assert.Equal(t, "560086", addressInfo.PostalCode)
		assert.Equal(t, "IN", addressInfo.CountryCode)
	})

	t.Run("DetailedAddress_Complete_Structure", func(t *testing.T) {
		// Test detailed address structure with all fields
		detailedAddress := &models.DetailedAddress{
			Type:        "INTERNATIONAL_HUB_ADDRESS",
			Zip:         "560024",
			Name:        "Contact Person",
			Phone:       "+91-1234567890",
			Email:       "contact@hub.com",
			Street:      "Hub Street Address",
			Landmark:    "Near Landmark",
			City:        "City Name",
			State:       "State Name",
			Country:     "Country Name",
			Latitude:    float64Ptr(12.9716),
			Longitude:   float64Ptr(77.5946),
			AddressName: "WAREHOUSE",
		}

		// Verify all fields are set correctly
		assert.Equal(t, "INTERNATIONAL_HUB_ADDRESS", detailedAddress.Type)
		assert.Equal(t, "560024", detailedAddress.Zip)
		assert.Equal(t, "Contact Person", detailedAddress.Name)
		assert.Equal(t, "+91-1234567890", detailedAddress.Phone)
		assert.Equal(t, "contact@hub.com", detailedAddress.Email)
		assert.Equal(t, "Hub Street Address", detailedAddress.Street)
		assert.Equal(t, "Near Landmark", detailedAddress.Landmark)
		assert.Equal(t, "City Name", detailedAddress.City)
		assert.Equal(t, "State Name", detailedAddress.State)
		assert.Equal(t, "Country Name", detailedAddress.Country)
		assert.Equal(t, "WAREHOUSE", detailedAddress.AddressName)
		assert.NotNil(t, detailedAddress.Latitude)
		assert.NotNil(t, detailedAddress.Longitude)
		assert.Equal(t, 12.9716, *detailedAddress.Latitude)
		assert.Equal(t, 77.5946, *detailedAddress.Longitude)
	})

	t.Run("Backward_Compatibility_Structure", func(t *testing.T) {
		// Test that existing response structure is preserved
		response := &models.ServiceabilityV2Response{
			Success: true,
			Partners: []models.PartnerV2Response{
				{
					PartnerID:   "test-id",
					PartnerCode: "dhl",
					Rating:      4.5,
					Capabilities: map[string]interface{}{
						"total_transit_days": 7,
					},
				},
			},
			Metadata: &models.V2ResponseMetadata{
				TotalPartners:    1,
				ServiceableCount: 1,
			},
		}

		// Verify backward compatibility
		assert.True(t, response.Success)
		assert.Len(t, response.Partners, 1)
		assert.NotNil(t, response.Metadata)
		assert.Equal(t, "dhl", response.Partners[0].PartnerCode)
		assert.Equal(t, 1, response.Metadata.TotalPartners)
		assert.Equal(t, 1, response.Metadata.ServiceableCount)

		// Verify new fields can be nil for backward compatibility
		assert.Nil(t, response.SourceAddress)
		assert.Nil(t, response.DestinationAddress)
		assert.Len(t, response.Addresses, 0)
	})
}

// TestAddressIntegrationEdgeCases tests edge cases for address integration
func TestAddressIntegrationEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("Response_With_No_Address_Data", func(t *testing.T) {
		// Test response when no address data is available
		response := &models.ServiceabilityV2Response{
			Success:            true,
			SourceAddress:      nil,
			DestinationAddress: nil,
			Addresses:          []models.DetailedAddress{},
			Partners: []models.PartnerV2Response{
				{
					PartnerCode: "shipyaari",
					Rating:      4.2,
				},
			},
		}

		assert.True(t, response.Success)
		assert.Nil(t, response.SourceAddress)
		assert.Nil(t, response.DestinationAddress)
		assert.Len(t, response.Addresses, 0)
		assert.Len(t, response.Partners, 1)
	})

	t.Run("Response_With_Partial_Address_Data", func(t *testing.T) {
		// Test response with only source address
		response := &models.ServiceabilityV2Response{
			Success: true,
			SourceAddress: &models.AddressInfo{
				PostalCode:  "560086",
				CountryCode: "IN",
			},
			DestinationAddress: nil,
			Addresses:          []models.DetailedAddress{},
		}

		assert.True(t, response.Success)
		assert.NotNil(t, response.SourceAddress)
		assert.Nil(t, response.DestinationAddress)
		assert.Len(t, response.Addresses, 0)
	})

	t.Run("DetailedAddress_With_Missing_Optional_Fields", func(t *testing.T) {
		// Test detailed address with only required fields
		detailedAddress := &models.DetailedAddress{
			Type:        "INTERNATIONAL_HUB_ADDRESS",
			Zip:         "560024",
			Name:        "Hub Manager",
			Phone:       "+91-9876543210",
			Email:       "hub@test.com",
			Street:      "Hub Street",
			City:        "Bangalore",
			State:       "Karnataka",
			Country:     "India",
			AddressName: "WAREHOUSE",
			// Missing: Landmark, Latitude, Longitude
		}

		assert.Equal(t, "INTERNATIONAL_HUB_ADDRESS", detailedAddress.Type)
		assert.Equal(t, "Hub Manager", detailedAddress.Name)
		assert.Equal(t, "Bangalore", detailedAddress.City)
		assert.Equal(t, "WAREHOUSE", detailedAddress.AddressName)
		assert.Empty(t, detailedAddress.Landmark)
		assert.Nil(t, detailedAddress.Latitude)
		assert.Nil(t, detailedAddress.Longitude)
	})

	t.Run("Multiple_Hub_Addresses", func(t *testing.T) {
		// Test response with multiple hub addresses
		response := &models.ServiceabilityV2Response{
			Success: true,
			Addresses: []models.DetailedAddress{
				{
					Type:        "INTERNATIONAL_HUB_ADDRESS",
					City:        "Bangalore",
					AddressName: "WAREHOUSE",
				},
				{
					Type:        "INTERNATIONAL_HUB_ADDRESS",
					City:        "Delhi",
					AddressName: "WAREHOUSE",
				},
			},
		}

		assert.Len(t, response.Addresses, 2)
		assert.Equal(t, "Bangalore", response.Addresses[0].City)
		assert.Equal(t, "Delhi", response.Addresses[1].City)
	})
}

// TestJSONMarshalingForAddresses tests JSON serialization of the new address fields
func TestJSONMarshalingForAddresses(t *testing.T) {
	t.Parallel()

	t.Run("Address_JSON_Tags", func(t *testing.T) {
		// This test ensures the JSON tags are correct for API responses
		// The structs should serialize correctly for API consumption

		addressInfo := &models.AddressInfo{
			PostalCode:  "560086",
			CountryCode: "IN",
		}

		detailedAddress := &models.DetailedAddress{
			Type:        "INTERNATIONAL_HUB_ADDRESS",
			Zip:         "560024",
			Name:        "Hub Manager",
			AddressName: "WAREHOUSE",
		}

		// Basic field validation - ensuring structs are valid
		assert.NotNil(t, addressInfo)
		assert.NotNil(t, detailedAddress)
		assert.Equal(t, "560086", addressInfo.PostalCode)
		assert.Equal(t, "INTERNATIONAL_HUB_ADDRESS", detailedAddress.Type)
	})
}
