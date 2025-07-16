package unit

import (
	"context"
	"testing"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/utils/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDHLCapabilitiesValidation tests validation of flattened DHL capabilities structure
func TestDHLCapabilitiesValidation(t *testing.T) {
	t.Parallel()

	// Setup validator
	config := &utils.ValidationConfig{
		AllowedServiceTypes:     []string{"Express", "Standard", "SDD", "NDD"},
		AllowedParcelCategories: []string{"courier", "ecom", "cargo"},
		AllowedOperationTypes:   []string{"pickup", "delivery"},
		AllowedPaymentModes:     []string{"ONLINE", "COD"},
		AllowedDeliveryModes:    []string{"AIR", "SURFACE", "RAIL"},
	}
	validator := utils.NewPartnerDataValidator(config)

	tests := []struct {
		name        string
		capability  *interfaces.PartnerCapability
		expectError bool
		errorField  string
	}{
		{
			name: "Valid DHL flattened capabilities",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "DHL Express",
				IsActive:         true,
				ServiceTypes:     []string{"Express"},
				ParcelCategories: []string{"courier"},
				OperationTypes:   []string{"pickup", "delivery"},
				PaymentModes:     []string{"ONLINE"},
				DeliveryModes:    []string{"AIR"},
			},
			expectError: false,
		},
		{
			name: "Invalid partner ID",
			capability: &interfaces.PartnerCapability{
				PartnerID:        0, // Invalid
				PartnerName:      "DHL Express",
				IsActive:         true,
				ServiceTypes:     []string{"Express"},
				ParcelCategories: []string{"courier"},
				OperationTypes:   []string{"pickup", "delivery"},
				PaymentModes:     []string{"ONLINE"},
				DeliveryModes:    []string{"AIR"},
			},
			expectError: true,
			errorField:  "partner_id",
		},
		{
			name: "Empty partner name",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "", // Invalid
				IsActive:         true,
				ServiceTypes:     []string{"Express"},
				ParcelCategories: []string{"courier"},
				OperationTypes:   []string{"pickup", "delivery"},
				PaymentModes:     []string{"ONLINE"},
				DeliveryModes:    []string{"AIR"},
			},
			expectError: true,
			errorField:  "partner_name",
		},
		{
			name: "Empty service types",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "DHL Express",
				IsActive:         true,
				ServiceTypes:     []string{}, // Invalid
				ParcelCategories: []string{"courier"},
				OperationTypes:   []string{"pickup", "delivery"},
				PaymentModes:     []string{"ONLINE"},
				DeliveryModes:    []string{"AIR"},
			},
			expectError: true,
			errorField:  "service_types",
		},
		{
			name: "Invalid service type",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "DHL Express",
				IsActive:         true,
				ServiceTypes:     []string{"InvalidType"}, // Invalid
				ParcelCategories: []string{"courier"},
				OperationTypes:   []string{"pickup", "delivery"},
				PaymentModes:     []string{"ONLINE"},
				DeliveryModes:    []string{"AIR"},
			},
			expectError: true,
			errorField:  "service_types[0]",
		},
		{
			name: "Empty parcel categories",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "DHL Express",
				IsActive:         true,
				ServiceTypes:     []string{"Express"},
				ParcelCategories: []string{}, // Invalid
				OperationTypes:   []string{"pickup", "delivery"},
				PaymentModes:     []string{"ONLINE"},
				DeliveryModes:    []string{"AIR"},
			},
			expectError: true,
			errorField:  "parcel_categories",
		},
		{
			name: "Invalid operation type",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "DHL Express",
				IsActive:         true,
				ServiceTypes:     []string{"Express"},
				ParcelCategories: []string{"courier"},
				OperationTypes:   []string{"invalid_operation"}, // Invalid
				PaymentModes:     []string{"ONLINE"},
				DeliveryModes:    []string{"AIR"},
			},
			expectError: true,
			errorField:  "operation_types[0]",
		},
		{
			name: "Invalid payment mode",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "DHL Express",
				IsActive:         true,
				ServiceTypes:     []string{"Express"},
				ParcelCategories: []string{"courier"},
				OperationTypes:   []string{"pickup", "delivery"},
				PaymentModes:     []string{"INVALID_PAYMENT"}, // Invalid
				DeliveryModes:    []string{"AIR"},
			},
			expectError: true,
			errorField:  "payment_modes[0]",
		},
		{
			name: "Invalid delivery mode",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "DHL Express",
				IsActive:         true,
				ServiceTypes:     []string{"Express"},
				ParcelCategories: []string{"courier"},
				OperationTypes:   []string{"pickup", "delivery"},
				PaymentModes:     []string{"ONLINE"},
				DeliveryModes:    []string{"INVALID_MODE"}, // Invalid
			},
			expectError: true,
			errorField:  "delivery_modes[0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			err := validator.ValidatePartnerCapability(ctx, tt.capability)

			if tt.expectError {
				require.Error(t, err)
				validationErr, ok := err.(utils.ValidationErrors)
				require.True(t, ok, "Expected ValidationErrors type")
				require.NotEmpty(t, validationErr.Errors)

				// Check if the expected field is in the validation errors
				found := false
				for _, validationError := range validationErr.Errors {
					if validationError.Field == tt.errorField {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error field %s not found in validation errors", tt.errorField)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDHLFlattenedCapabilitiesStructure tests specific DHL flattened capabilities format
func TestDHLFlattenedCapabilitiesStructure(t *testing.T) {
	t.Parallel()

	// Test DHL capabilities with individual fields vs nested structure
	dhlCapabilities := map[string]interface{}{
		// DHL-style flattened capabilities
		"next_business_day":          true,
		"total_transit_days":         2,
		"pickup_earliest":            "09:00",
		"pickup_latest":              "17:00",
		"local_cutoff_date_and_time": "2024-01-15T15:00:00Z",
		"guaranteed_delivery":        true,
		"service_name":               "EXPRESS WORLDWIDE",
		"category_name":              "EXPRESS",
		"dimension_based_pricing":    true,
		"weight_based_pricing":       true,
		"max_weight_kg":              70.0,
		"max_dimension_cm":           "120x80x80",
		"insurance_available":        true,
		"signature_required":         true,
		"tracking_available":         true,
		"cod_available":              false,
	}

	// Validate that all expected DHL fields are present and have correct types
	assert.IsType(t, true, dhlCapabilities["next_business_day"])
	assert.IsType(t, 0, dhlCapabilities["total_transit_days"])
	assert.IsType(t, "", dhlCapabilities["pickup_earliest"])
	assert.IsType(t, "", dhlCapabilities["pickup_latest"])
	assert.IsType(t, "", dhlCapabilities["local_cutoff_date_and_time"])
	assert.IsType(t, true, dhlCapabilities["guaranteed_delivery"])
	assert.IsType(t, "", dhlCapabilities["service_name"])
	assert.IsType(t, "", dhlCapabilities["category_name"])
	assert.IsType(t, true, dhlCapabilities["dimension_based_pricing"])
	assert.IsType(t, true, dhlCapabilities["weight_based_pricing"])
	assert.IsType(t, 0.0, dhlCapabilities["max_weight_kg"])
	assert.IsType(t, "", dhlCapabilities["max_dimension_cm"])
	assert.IsType(t, true, dhlCapabilities["insurance_available"])
	assert.IsType(t, true, dhlCapabilities["signature_required"])
	assert.IsType(t, true, dhlCapabilities["tracking_available"])
	assert.IsType(t, false, dhlCapabilities["cod_available"])

	// Test that all fields have valid values
	assert.Equal(t, true, dhlCapabilities["next_business_day"])
	assert.Equal(t, 2, dhlCapabilities["total_transit_days"])
	assert.Equal(t, "09:00", dhlCapabilities["pickup_earliest"])
	assert.Equal(t, "17:00", dhlCapabilities["pickup_latest"])
	assert.Equal(t, "2024-01-15T15:00:00Z", dhlCapabilities["local_cutoff_date_and_time"])
	assert.Equal(t, true, dhlCapabilities["guaranteed_delivery"])
	assert.Equal(t, "EXPRESS WORLDWIDE", dhlCapabilities["service_name"])
	assert.Equal(t, "EXPRESS", dhlCapabilities["category_name"])
	assert.Equal(t, 70.0, dhlCapabilities["max_weight_kg"])
	assert.Equal(t, "120x80x80", dhlCapabilities["max_dimension_cm"])
}

// TestDHLVsTraditionalCapabilitiesComparison tests differences between DHL flattened and traditional nested capabilities
func TestDHLVsTraditionalCapabilitiesComparison(t *testing.T) {
	t.Parallel()

	// Traditional nested capabilities structure (like other partners)
	traditionalCapabilities := map[string]interface{}{
		"pickup": map[string]interface{}{
			"available": true,
			"earliest":  "09:00",
			"latest":    "17:00",
			"same_day":  false,
			"next_day":  true,
		},
		"delivery": map[string]interface{}{
			"available":    true,
			"guaranteed":   false,
			"transit_days": 3,
			"cutoff_time":  "15:00",
		},
		"services": []map[string]interface{}{
			{
				"name":     "Standard",
				"category": "SURFACE",
				"cod":      true,
				"tracking": true,
			},
		},
	}

	// DHL flattened capabilities structure
	dhlCapabilities := map[string]interface{}{
		"pickup_available":           true,
		"pickup_earliest":            "09:00",
		"pickup_latest":              "17:00",
		"pickup_same_day":            false,
		"pickup_next_day":            true,
		"delivery_available":         true,
		"guaranteed_delivery":        true, // DHL usually guarantees
		"total_transit_days":         2,    // DHL is faster
		"local_cutoff_date_and_time": "2024-01-15T15:00:00Z",
		"service_name":               "EXPRESS WORLDWIDE",
		"category_name":              "EXPRESS", // Different category
		"cod_available":              false,     // DHL typically doesn't support COD
		"tracking_available":         true,
	}

	// Test structure differences
	t.Run("Traditional nested structure", func(t *testing.T) {
		// Traditional structure has nested objects
		pickup, exists := traditionalCapabilities["pickup"]
		require.True(t, exists)
		pickupMap, ok := pickup.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, pickupMap, "available")
		assert.Contains(t, pickupMap, "earliest")

		delivery, exists := traditionalCapabilities["delivery"]
		require.True(t, exists)
		deliveryMap, ok := delivery.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, deliveryMap, "available")
		assert.Contains(t, deliveryMap, "guaranteed")
	})

	t.Run("DHL flattened structure", func(t *testing.T) {
		// DHL structure has flattened fields
		assert.Contains(t, dhlCapabilities, "pickup_available")
		assert.Contains(t, dhlCapabilities, "pickup_earliest")
		assert.Contains(t, dhlCapabilities, "delivery_available")
		assert.Contains(t, dhlCapabilities, "guaranteed_delivery")

		// No nested objects - all fields are at top level
		for key, value := range dhlCapabilities {
			// Ensure no nested maps (all primitives)
			switch value.(type) {
			case map[string]interface{}:
				t.Errorf("Field %s should not be a nested object in DHL flattened structure", key)
			case []map[string]interface{}:
				t.Errorf("Field %s should not be an array of objects in DHL flattened structure", key)
			}
		}
	})

	t.Run("Capability mapping validation", func(t *testing.T) {
		// Test that we can map between structures

		// Traditional pickup.available -> DHL pickup_available
		traditionalPickup := traditionalCapabilities["pickup"].(map[string]interface{})
		assert.Equal(t, traditionalPickup["available"], dhlCapabilities["pickup_available"])

		// Traditional pickup.earliest -> DHL pickup_earliest
		assert.Equal(t, traditionalPickup["earliest"], dhlCapabilities["pickup_earliest"])

		// Traditional delivery.guaranteed -> DHL guaranteed_delivery
		traditionalDelivery := traditionalCapabilities["delivery"].(map[string]interface{})
		assert.NotEqual(t, traditionalDelivery["guaranteed"], dhlCapabilities["guaranteed_delivery"]) // DHL guarantees, others may not

		// Traditional delivery.transit_days -> DHL total_transit_days
		assert.NotEqual(t, traditionalDelivery["transit_days"], dhlCapabilities["total_transit_days"]) // DHL is typically faster
	})
}

// TestDHLCapabilityFieldValidation tests individual DHL capability field validation
func TestDHLCapabilityFieldValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		field       string
		value       interface{}
		expectValid bool
	}{
		// Boolean fields
		{"Valid next_business_day true", "next_business_day", true, true},
		{"Valid next_business_day false", "next_business_day", false, true},
		{"Invalid next_business_day string", "next_business_day", "true", false},

		// Integer fields
		{"Valid total_transit_days", "total_transit_days", 2, true},
		{"Valid total_transit_days zero", "total_transit_days", 0, true},
		{"Invalid total_transit_days negative", "total_transit_days", -1, false},
		{"Invalid total_transit_days string", "total_transit_days", "2", false},

		// String fields
		{"Valid pickup_earliest", "pickup_earliest", "09:00", true},
		{"Valid pickup_latest", "pickup_latest", "17:00", true},
		{"Invalid pickup_earliest empty", "pickup_earliest", "", false},
		{"Invalid pickup_latest non-time", "pickup_latest", "not-a-time", false},

		// Float fields
		{"Valid max_weight_kg", "max_weight_kg", 70.0, true},
		{"Valid max_weight_kg zero", "max_weight_kg", 0.0, true},
		{"Invalid max_weight_kg negative", "max_weight_kg", -10.0, false},
		{"Invalid max_weight_kg string", "max_weight_kg", "70", false},

		// Service name validation
		{"Valid service_name", "service_name", "EXPRESS WORLDWIDE", true},
		{"Invalid service_name empty", "service_name", "", false},
		{"Invalid service_name number", "service_name", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			capabilities := map[string]interface{}{
				tt.field: tt.value,
			}

			isValid := validateDHLCapabilityField(tt.field, tt.value)
			if tt.expectValid {
				assert.True(t, isValid, "Field %s with value %v should be valid", tt.field, tt.value)
			} else {
				assert.False(t, isValid, "Field %s with value %v should be invalid", tt.field, tt.value)
			}

			// Ensure the capability map contains the field
			assert.Contains(t, capabilities, tt.field)
		})
	}
}

// TestPartnerServiceabilityResultCapabilitiesValidation tests validation of PartnerServiceabilityResult capabilities
func TestPartnerServiceabilityResultCapabilitiesValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		capabilities map[string]interface{}
		expectValid  bool
		description  string
	}{
		{
			name: "Valid DHL capabilities",
			capabilities: map[string]interface{}{
				"next_business_day":          true,
				"total_transit_days":         2,
				"pickup_earliest":            "09:00",
				"pickup_latest":              "17:00",
				"local_cutoff_date_and_time": "2024-01-15T15:00:00Z",
				"guaranteed_delivery":        true,
				"service_name":               "EXPRESS WORLDWIDE",
				"category_name":              "EXPRESS",
				"max_weight_kg":              70.0,
				"tracking_available":         true,
			},
			expectValid: true,
			description: "All DHL capability fields are valid",
		},
		{
			name: "Mixed valid and invalid capabilities",
			capabilities: map[string]interface{}{
				"next_business_day":  true,
				"total_transit_days": -1, // Invalid
				"pickup_earliest":    "", // Invalid
				"service_name":       "EXPRESS WORLDWIDE",
				"max_weight_kg":      70.0,
			},
			expectValid: false,
			description: "Some capability fields are invalid",
		},
		{
			name:         "Empty capabilities",
			capabilities: map[string]interface{}{},
			expectValid:  true, // Empty is valid for optional capabilities
			description:  "Empty capabilities should be valid",
		},
		{
			name: "Traditional nested capabilities",
			capabilities: map[string]interface{}{
				"pickup": map[string]interface{}{
					"available": true,
					"earliest":  "09:00",
				},
				"delivery": map[string]interface{}{
					"available":  true,
					"guaranteed": false,
				},
			},
			expectValid: true,
			description: "Traditional nested structure should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := &interfaces.PartnerServiceabilityResult{
				PartnerCode:   "DHL",
				PartnerName:   "DHL Express",
				IsServiceable: true,
				Capabilities:  tt.capabilities,
			}

			isValid := validatePartnerServiceabilityResultCapabilities(result)
			if tt.expectValid {
				assert.True(t, isValid, tt.description)
			} else {
				assert.False(t, isValid, tt.description)
			}
		})
	}
}

// Helper function to validate individual DHL capability fields
func validateDHLCapabilityField(field string, value interface{}) bool {
	switch field {
	case "next_business_day", "guaranteed_delivery", "dimension_based_pricing",
		"weight_based_pricing", "insurance_available", "signature_required",
		"tracking_available", "cod_available":
		_, ok := value.(bool)
		return ok

	case "total_transit_days":
		if val, ok := value.(int); ok {
			return val >= 0
		}
		return false

	case "pickup_earliest", "pickup_latest", "local_cutoff_date_and_time",
		"service_name", "category_name", "max_dimension_cm":
		if val, ok := value.(string); ok {
			return val != ""
		}
		return false

	case "max_weight_kg":
		if val, ok := value.(float64); ok {
			return val >= 0
		}
		return false

	default:
		// Unknown field - accept as valid for extensibility
		return true
	}
}

// Helper function to validate PartnerServiceabilityResult capabilities
func validatePartnerServiceabilityResultCapabilities(result *interfaces.PartnerServiceabilityResult) bool {
	if result.Capabilities == nil {
		return true // Nil capabilities are valid
	}

	for field, value := range result.Capabilities {
		// For DHL, validate flattened structure
		if !validateDHLCapabilityField(field, value) {
			return false
		}
	}

	return true
}
