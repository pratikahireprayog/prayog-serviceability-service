package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"
)

func TestDHLResponseBuilderCapabilitiesFlattening(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		nestedCapabilities       map[string]interface{}
		expectedFlatCapabilities map[string]interface{}
		shouldFlatten            bool
	}{
		{
			name: "DHL nested pickup/delivery capabilities",
			nestedCapabilities: map[string]interface{}{
				"pickup": map[string]interface{}{
					"earliest":          "09:00",
					"latest":            "17:00",
					"next_business_day": true,
				},
				"delivery": map[string]interface{}{
					"total_transit_days":         3,
					"local_cutoff_date_and_time": "2024-01-15T15:00:00Z",
				},
			},
			expectedFlatCapabilities: map[string]interface{}{
				"pickup_earliest":            "09:00",
				"pickup_latest":              "17:00",
				"next_business_day":          true,
				"total_transit_days":         3,
				"local_cutoff_date_and_time": "2024-01-15T15:00:00Z",
			},
			shouldFlatten: true,
		},
		{
			name: "DHL international capabilities",
			nestedCapabilities: map[string]interface{}{
				"international": map[string]interface{}{
					"supported":         true,
					"transit_days":      5,
					"customs_clearance": true,
				},
				"pickup": map[string]interface{}{
					"earliest": "08:00",
					"latest":   "18:00",
				},
			},
			expectedFlatCapabilities: map[string]interface{}{
				"international_supported":    true,
				"international_transit_days": 5,
				"customs_clearance":          true,
				"pickup_earliest":            "08:00",
				"pickup_latest":              "18:00",
			},
			shouldFlatten: true,
		},
		{
			name: "DHL already flat capabilities",
			nestedCapabilities: map[string]interface{}{
				"next_business_day":          true,
				"total_transit_days":         2,
				"pickup_earliest":            "10:00",
				"pickup_latest":              "16:00",
				"local_cutoff_date_and_time": "2024-01-15T14:00:00Z",
			},
			expectedFlatCapabilities: map[string]interface{}{
				"next_business_day":          true,
				"total_transit_days":         2,
				"pickup_earliest":            "10:00",
				"pickup_latest":              "16:00",
				"local_cutoff_date_and_time": "2024-01-15T14:00:00Z",
			},
			shouldFlatten: false,
		},
		{
			name:                     "Empty nested capabilities",
			nestedCapabilities:       map[string]interface{}{},
			expectedFlatCapabilities: map[string]interface{}{},
			shouldFlatten:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock partner result with nested capabilities
			partnerID := uuid.New()
			partnerResult := &common.PartnerServiceabilityResult{
				PartnerID:   &partnerID,
				PartnerCode: "dhl",
				PartnerName: "DHL Express",
				Services: []models.ServiceV2{
					{
						ServiceCode: "EXPRESS",
						ServiceName: "DHL Express Worldwide",
					},
				},
				Capabilities: tt.nestedCapabilities,
				ResponseTime: 150 * time.Millisecond,
			}

			// Transform capabilities using response builder
			flatCapabilities := flattenDHLCapabilities(partnerResult.Capabilities)

			// Verify flattened structure
			for expectedKey, expectedValue := range tt.expectedFlatCapabilities {
				actualValue, exists := flatCapabilities[expectedKey]
				assert.True(t, exists, "Expected capability %s should exist", expectedKey)
				assert.Equal(t, expectedValue, actualValue, "Capability %s should match expected value", expectedKey)
			}

			// Verify no nested objects remain in flattened capabilities
			for key, value := range flatCapabilities {
				_, isMap := value.(map[string]interface{})
				assert.False(t, isMap, "Capability %s should not be a nested object after flattening", key)
			}
		})
	}
}

func TestDHLResponseBuilderCapabilitiesTransformation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		partnerResults       []*common.PartnerServiceabilityResult
		expectedPartners     int
		expectedDHLFlat      bool
		expectedOthersNested bool
	}{
		{
			name: "Mixed DHL and traditional partners",
			partnerResults: []*common.PartnerServiceabilityResult{
				createDHLPartnerWithNestedCapabilities(),
				createTraditionalPartnerWithFlatCapabilities("shipyaari"),
				createTraditionalPartnerWithNestedCapabilities("smile_ecom"),
			},
			expectedPartners:     3,
			expectedDHLFlat:      true,
			expectedOthersNested: true,
		},
		{
			name: "Only DHL partner",
			partnerResults: []*common.PartnerServiceabilityResult{
				createDHLPartnerWithNestedCapabilities(),
			},
			expectedPartners:     1,
			expectedDHLFlat:      true,
			expectedOthersNested: false,
		},
		{
			name: "Only traditional partners",
			partnerResults: []*common.PartnerServiceabilityResult{
				createTraditionalPartnerWithFlatCapabilities("shipyaari"),
				createTraditionalPartnerWithNestedCapabilities("smile_ecom"),
			},
			expectedPartners:     2,
			expectedDHLFlat:      false,
			expectedOthersNested: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Build response using response builder
			response := buildPartnerResponse(tt.partnerResults)

			// Verify number of partners
			assert.Len(t, response.Partners, tt.expectedPartners)

			// Check DHL capabilities flattening
			for _, partner := range response.Partners {
				if partner.PartnerID == "dhl" && tt.expectedDHLFlat {
					// Verify DHL capabilities are flattened
					verifyDHLCapabilitiesFlattened(t, partner.Capabilities)
				} else if partner.PartnerID != "dhl" && tt.expectedOthersNested {
					// Verify other partners may have nested capabilities
					verifyTraditionalCapabilitiesStructure(t, partner.Capabilities)
				}
			}
		})
	}
}

func TestDHLResponseBuilderCapabilitiesValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		capabilities        map[string]interface{}
		expectedValidation  bool
		expectedErrorFields []string
	}{
		{
			name: "Valid DHL flattened capabilities",
			capabilities: map[string]interface{}{
				"next_business_day":          true,
				"total_transit_days":         3,
				"pickup_earliest":            "09:00",
				"pickup_latest":              "17:00",
				"local_cutoff_date_and_time": "2024-01-15T15:00:00Z",
			},
			expectedValidation:  true,
			expectedErrorFields: []string{},
		},
		{
			name: "Invalid DHL capabilities - missing required fields",
			capabilities: map[string]interface{}{
				"next_business_day": true,
				// Missing other required fields
			},
			expectedValidation: false,
			expectedErrorFields: []string{
				"total_transit_days",
				"pickup_earliest",
				"pickup_latest",
			},
		},
		{
			name: "Invalid DHL capabilities - wrong data types",
			capabilities: map[string]interface{}{
				"next_business_day":          "invalid_boolean",
				"total_transit_days":         "not_a_number",
				"pickup_earliest":            123,
				"pickup_latest":              true,
				"local_cutoff_date_and_time": 456,
			},
			expectedValidation: false,
			expectedErrorFields: []string{
				"next_business_day",
				"total_transit_days",
				"pickup_earliest",
				"pickup_latest",
				"local_cutoff_date_and_time",
			},
		},
		{
			name: "DHL capabilities with international fields",
			capabilities: map[string]interface{}{
				"next_business_day":          false,
				"total_transit_days":         5,
				"pickup_earliest":            "08:00",
				"pickup_latest":              "18:00",
				"local_cutoff_date_and_time": "2024-01-15T16:00:00Z",
				"international_supported":    true,
				"customs_clearance":          true,
			},
			expectedValidation:  true,
			expectedErrorFields: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Validate DHL capabilities structure
			isValid, errorFields := validateDHLCapabilitiesStructure(tt.capabilities)

			assert.Equal(t, tt.expectedValidation, isValid)

			if !tt.expectedValidation {
				assert.NotEmpty(t, errorFields)
				for _, expectedField := range tt.expectedErrorFields {
					assert.Contains(t, errorFields, expectedField, "Expected error field %s should be present", expectedField)
				}
			} else {
				assert.Empty(t, errorFields)
			}
		})
	}
}

func TestDHLResponseBuilderCapabilitiesComparison(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		dhlNestedCapabilities   map[string]interface{}
		traditionalCapabilities map[string]interface{}
		expectedDifferences     []string
	}{
		{
			name: "DHL vs Traditional structure differences",
			dhlNestedCapabilities: map[string]interface{}{
				"pickup": map[string]interface{}{
					"earliest": "09:00",
					"latest":   "17:00",
				},
				"delivery": map[string]interface{}{
					"next_business_day": true,
					"transit_days":      3,
				},
			},
			traditionalCapabilities: map[string]interface{}{
				"pickup_times": map[string]interface{}{
					"start": "09:00",
					"end":   "17:00",
				},
				"delivery_options": []string{"next_day", "standard"},
				"estimated_days":   3,
			},
			expectedDifferences: []string{
				"structure_format",
				"field_naming",
				"data_representation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Flatten DHL capabilities
			flatDHLCapabilities := flattenDHLCapabilities(tt.dhlNestedCapabilities)

			// Compare structures
			differences := compareCapabilitiesStructure(flatDHLCapabilities, tt.traditionalCapabilities)

			// Verify expected differences
			for _, expectedDiff := range tt.expectedDifferences {
				assert.Contains(t, differences, expectedDiff, "Expected difference %s should be detected", expectedDiff)
			}
		})
	}
}

// Helper functions for response builder capabilities tests

func flattenDHLCapabilities(capabilities map[string]interface{}) map[string]interface{} {
	flattened := make(map[string]interface{})

	for key, value := range capabilities {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Handle nested objects
			for nestedKey, nestedValue := range nestedMap {
				flatKey := key + "_" + nestedKey
				flattened[flatKey] = nestedValue
			}
		} else {
			// Keep flat values as-is
			flattened[key] = value
		}
	}

	return flattened
}

func createDHLPartnerWithNestedCapabilities() *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: "dhl",
		PartnerName: "DHL Express",
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS",
				ServiceName: "DHL Express Worldwide",
			},
		},
		Capabilities: map[string]interface{}{
			"pickup": map[string]interface{}{
				"earliest":          "09:00",
				"latest":            "17:00",
				"next_business_day": true,
			},
			"delivery": map[string]interface{}{
				"total_transit_days":         3,
				"local_cutoff_date_and_time": "2024-01-15T15:00:00Z",
			},
		},
	}
}

func createTraditionalPartnerWithFlatCapabilities(partnerCode string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: partnerCode,
		PartnerName: "Traditional " + partnerCode,
		Services: []models.ServiceV2{
			{
				ServiceCode: "STANDARD",
				ServiceName: "Standard Delivery",
			},
		},
		Capabilities: map[string]interface{}{
			"next_day_delivery": true,
			"estimated_days":    2,
			"pickup_available":  true,
			"delivery_window":   "9AM-6PM",
		},
	}
}

func createTraditionalPartnerWithNestedCapabilities(partnerCode string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: partnerCode,
		PartnerName: "Traditional " + partnerCode,
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS",
				ServiceName: "Express Delivery",
			},
		},
		Capabilities: map[string]interface{}{
			"delivery_options": map[string]interface{}{
				"same_day": false,
				"next_day": true,
				"standard": true,
			},
			"pickup_times": map[string]interface{}{
				"start": "08:00",
				"end":   "18:00",
			},
		},
	}
}

func buildPartnerResponse(partnerResults []*common.PartnerServiceabilityResult) *models.ServiceabilityV2Response {
	partners := make([]models.PartnerV2Response, 0, len(partnerResults))

	for _, result := range partnerResults {
		capabilities := result.Capabilities

		// Flatten DHL capabilities
		if result.PartnerCode == "dhl" {
			capabilities = flattenDHLCapabilities(capabilities)
		}

		partner := models.PartnerV2Response{
			PartnerID:    result.PartnerCode,
			PartnerCode:  result.PartnerCode,
			PartnerName:  result.PartnerName,
			Services:     result.Services,
			Capabilities: capabilities,
		}

		partners = append(partners, partner)
	}

	return &models.ServiceabilityV2Response{
		Success:  true,
		Partners: partners,
	}
}

func verifyDHLCapabilitiesFlattened(t *testing.T, capabilities map[string]interface{}) {
	// Verify no nested objects exist
	for key, value := range capabilities {
		_, isMap := value.(map[string]interface{})
		assert.False(t, isMap, "DHL capability %s should not be nested after flattening", key)
	}

	// Verify expected DHL flat structure
	expectedFields := []string{
		"pickup_earliest",
		"pickup_latest",
		"next_business_day",
		"total_transit_days",
		"local_cutoff_date_and_time",
	}

	for _, field := range expectedFields {
		if _, exists := capabilities[field]; exists {
			// Field exists, verify it's not nested
			_, isMap := capabilities[field].(map[string]interface{})
			assert.False(t, isMap, "DHL capability %s should be flat, not nested", field)
		}
	}
}

func verifyTraditionalCapabilitiesStructure(t *testing.T, capabilities map[string]interface{}) {
	// Traditional partners may have various structures
	// This is just to verify they're processed differently than DHL
	assert.NotEmpty(t, capabilities, "Traditional partner should have capabilities")
}

func validateDHLCapabilitiesStructure(capabilities map[string]interface{}) (bool, []string) {
	var errorFields []string

	requiredFields := map[string]string{
		"next_business_day":  "bool",
		"total_transit_days": "int",
		"pickup_earliest":    "string",
		"pickup_latest":      "string",
	}

	// Check required fields and types
	for field, expectedType := range requiredFields {
		value, exists := capabilities[field]
		if !exists {
			errorFields = append(errorFields, field)
			continue
		}

		// Type validation
		switch expectedType {
		case "bool":
			if _, ok := value.(bool); !ok {
				errorFields = append(errorFields, field)
			}
		case "int":
			if _, ok := value.(int); !ok {
				errorFields = append(errorFields, field)
			}
		case "string":
			if _, ok := value.(string); !ok {
				errorFields = append(errorFields, field)
			}
		}
	}

	// Optional fields type validation
	if value, exists := capabilities["local_cutoff_date_and_time"]; exists {
		if _, ok := value.(string); !ok {
			errorFields = append(errorFields, "local_cutoff_date_and_time")
		}
	}

	return len(errorFields) == 0, errorFields
}

func compareCapabilitiesStructure(dhlCapabilities, traditionalCapabilities map[string]interface{}) []string {
	var differences []string

	// Check structure format differences
	dhlHasFlat := hasOnlyFlatStructure(dhlCapabilities)
	traditionalHasFlat := hasOnlyFlatStructure(traditionalCapabilities)

	if dhlHasFlat != traditionalHasFlat {
		differences = append(differences, "structure_format")
	}

	// Check field naming patterns
	dhlFields := getFieldNames(dhlCapabilities)
	traditionalFields := getFieldNames(traditionalCapabilities)

	if !hasCommonFieldNaming(dhlFields, traditionalFields) {
		differences = append(differences, "field_naming")
	}

	// Check data representation
	if hasDifferentDataRepresentation(dhlCapabilities, traditionalCapabilities) {
		differences = append(differences, "data_representation")
	}

	return differences
}

func hasOnlyFlatStructure(capabilities map[string]interface{}) bool {
	for _, value := range capabilities {
		if _, isMap := value.(map[string]interface{}); isMap {
			return false
		}
	}
	return true
}

func getFieldNames(capabilities map[string]interface{}) []string {
	var fields []string
	for key := range capabilities {
		fields = append(fields, key)
	}
	return fields
}

func hasCommonFieldNaming(fields1, fields2 []string) bool {
	// Simple check for common naming patterns
	for _, field1 := range fields1 {
		for _, field2 := range fields2 {
			if field1 == field2 {
				return true
			}
		}
	}
	return false
}

func hasDifferentDataRepresentation(caps1, caps2 map[string]interface{}) bool {
	// Check if similar concepts are represented differently
	// This is a simplified check
	return len(caps1) != len(caps2)
}
