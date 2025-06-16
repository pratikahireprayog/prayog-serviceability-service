package utils

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestSnakeCaseValidation(t *testing.T) {
	// Set up validator with custom validations
	validatorSetup := NewValidatorSetup()
	v := validatorSetup.GetValidator()

	// Test struct for snake_case validation
	type TestStruct struct {
		Code string `validate:"snake_case"`
	}

	tests := []struct {
		name      string
		code      string
		expectErr bool
	}{
		// Valid cases - lowercase single word
		{"lowercase single word", "state", false},
		{"lowercase with number", "state1", false},
		{"lowercase with multiple numbers", "state123", false},

		// Valid cases - snake_case format
		{"snake_case format", "state_province", false},
		{"snake_case with numbers", "state_1", false},
		{"multiple underscores", "state_province_territory", false},

		// Invalid cases - uppercase
		{"uppercase single", "STATE", true},
		{"uppercase snake_case", "STATE_PROVINCE", true},
		{"mixed case", "State", true},
		{"mixed case snake", "State_Province", true},

		// Invalid cases - invalid characters
		{"with hyphen", "state-province", true},
		{"with space", "state province", true},
		{"with special chars", "state@province", true},

		// Invalid cases - format issues
		{"starting with underscore", "_state", true},
		{"ending with underscore", "state_", true},
		{"consecutive underscores", "state__province", true},
		{"empty string", "", true},
		{"only underscore", "_", true},
		{"starting with number", "1state", true}, // Numbers at start should be invalid for snake_case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStruct{Code: tt.code}
			err := v.Struct(testStruct)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected validation error for code '%s', but got none", tt.code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error for code '%s', but got: %v", tt.code, err)
				}
			}
		})
	}
}

func TestCountryCodeISOValidation(t *testing.T) {
	// Set up validator with custom validations
	validatorSetup := NewValidatorSetup()
	v := validatorSetup.GetValidator()

	// Test struct for country_code_iso validation
	type TestStruct struct {
		CountryCode string `validate:"country_code_iso"`
	}

	tests := []struct {
		name      string
		code      string
		expectErr bool
	}{
		// Valid cases - ISO 3166 A-2 format
		{"US uppercase", "US", false},
		{"IN uppercase", "IN", false},
		{"GB uppercase", "GB", false},
		{"DE uppercase", "DE", false},

		// Invalid cases - lowercase
		{"us lowercase", "us", true},
		{"in lowercase", "in", true},

		// Invalid cases - wrong length
		{"too short", "U", true},
		{"too long", "USA", true},
		{"empty", "", true},

		// Invalid cases - special characters
		{"with number", "U1", true},
		{"with hyphen", "U-S", true},
		{"with space", "U S", true},
		{"with special char", "U@", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStruct{CountryCode: tt.code}
			err := v.Struct(testStruct)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected validation error for country code '%s', but got none", tt.code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error for country code '%s', but got: %v", tt.code, err)
				}
			}
		})
	}
}

func TestRegionCodeISOValidation(t *testing.T) {
	// Set up validator with custom validations
	validatorSetup := NewValidatorSetup()
	v := validatorSetup.GetValidator()

	// Test struct for region_code_iso validation
	type TestStruct struct {
		RegionCode string `validate:"region_code_iso"`
	}

	tests := []struct {
		name      string
		code      string
		expectErr bool
	}{
		// Valid cases - ISO 3166-2 format
		{"US-CA (California)", "US-CA", false},
		{"US-NY (New York)", "US-NY", false},
		{"IN-MH (Maharashtra)", "IN-MH", false},
		{"IN-DL (Delhi)", "IN-DL", false},
		{"GB-ENG (England)", "GB-ENG", false},
		{"AU-NSW (New South Wales)", "AU-NSW", false},
		{"CA-ON (Ontario)", "CA-ON", false},
		{"FR-75 (Paris with number)", "FR-75", false},
		{"US-1 (single digit subdivision)", "US-1", false},

		// Invalid cases - wrong format
		{"missing hyphen", "USCA", true},
		{"multiple hyphens", "US-CA-LA", true},
		{"only country code", "US", true},
		{"only subdivision", "CA", true},
		{"empty", "", true},

		// Invalid cases - country code part
		{"lowercase country", "us-CA", true},
		{"mixed case country", "Us-CA", true},
		{"single letter country", "U-CA", true},
		{"three letter country", "USA-CA", true},
		{"number in country", "U1-CA", true},

		// Invalid cases - subdivision part
		{"lowercase subdivision", "US-ca", true},
		{"mixed case subdivision", "US-Ca", true},
		{"empty subdivision", "US-", true},
		{"too long subdivision", "US-ABCD", true},
		{"special char in subdivision", "US-C@", true},
		{"space in subdivision", "US-C A", true},
		{"hyphen in subdivision", "US-C-A", true},

		// Invalid cases - general format
		{"starts with hyphen", "-US-CA", true},
		{"ends with hyphen", "US-CA-", true},
		{"only hyphen", "-", true},
		{"space instead of hyphen", "US CA", true},
		{"underscore instead of hyphen", "US_CA", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStruct{RegionCode: tt.code}
			err := v.Struct(testStruct)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected validation error for region code '%s', but got none", tt.code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error for region code '%s', but got: %v", tt.code, err)
				}
			}
		})
	}
}

func TestValidatorSetupIntegration(t *testing.T) {
	// Test that the validator setup correctly registers all custom validations
	validatorSetup := NewValidatorSetup()
	v := validatorSetup.GetValidator()

	// Test a complex struct with multiple validation tags
	type ComplexStruct struct {
		RegionCode  string  `validate:"required,region_code_iso"`
		CountryCode string  `validate:"required,country_code_iso"`
		PostalCode  string  `validate:"required,postal_code"`
		PartnerName string  `validate:"required,partner_name"`
		ServiceType string  `validate:"required,service_type"`
		ParcelCat   string  `validate:"required,parcel_category"`
		Rating      float64 `validate:"required,rating"`
	}

	// Valid case
	validStruct := ComplexStruct{
		RegionCode:  "US-CA",
		CountryCode: "US",
		PostalCode:  "12345",
		PartnerName: "Test Partner Corp",
		ServiceType: "Standard",
		ParcelCat:   "ecom",
		Rating:      8.5,
	}

	err := v.Struct(validStruct)
	if err != nil {
		t.Errorf("Expected no validation error for valid struct, but got: %v", err)
	}

	// Invalid case
	invalidStruct := ComplexStruct{
		RegionCode:  "STATE_PROVINCE", // Invalid - not ISO 3166-2 format
		CountryCode: "us",             // Invalid - lowercase
		PostalCode:  "abc",            // Invalid - letters only (postal_code validation might accept this)
		PartnerName: "",               // Invalid - empty
		ServiceType: "INVALID",        // Invalid - not in allowed list
		ParcelCat:   "invalid",        // Invalid - not in allowed list
		Rating:      15.0,             // Invalid - out of range
	}

	err = v.Struct(invalidStruct)
	if err == nil {
		t.Error("Expected validation errors for invalid struct, but got none")
	}

	// Check that validation errors are present - count may vary based on actual validation logic
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		expectedFields := []string{"RegionCode", "CountryCode", "PartnerName", "ServiceType", "ParcelCat", "Rating"}
		if len(validationErrors) < 6 {
			t.Errorf("Expected at least 6 validation errors for fields %v, but got %d", expectedFields, len(validationErrors))
		}

		// Log the actual errors for debugging
		t.Logf("Validation errors found: %d", len(validationErrors))
		for _, err := range validationErrors {
			t.Logf("  - Field: %s, Tag: %s, Value: %v", err.Field(), err.Tag(), err.Value())
		}
	} else {
		t.Error("Expected validator.ValidationErrors type")
	}
}

// TestOriginalErrorScenario tests the specific scenario that caused the original 500 error:
// Creating a region-type with code "state" should now work without the "Undefined validation function 'snake_case'" error
func TestOriginalErrorScenario(t *testing.T) {
	validatorSetup := NewValidatorSetup()
	v := validatorSetup.GetValidator()

	// Simulate the original CreateRegionTypeRequest that was failing
	type CreateRegionTypeRequest struct {
		Code        string  `json:"code" validate:"required,min=2,max=20,snake_case"`
		Name        string  `json:"name" validate:"required,min=2,max=100"`
		Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
		IsActive    *bool   `json:"is_active,omitempty"`
	}

	// This was the exact request that caused the original 500 error
	description := "State level administrative division"
	isActive := true

	regionTypeRequest := CreateRegionTypeRequest{
		Code:        "state", // This was failing with "Undefined validation function 'snake_case'"
		Name:        "State",
		Description: &description,
		IsActive:    &isActive,
	}

	// This should now pass validation without any "undefined validation function" errors
	err := v.Struct(regionTypeRequest)
	if err != nil {
		t.Errorf("Expected no validation error for region-type with code 'state', but got: %v", err)
		t.Errorf("This indicates the original 500 error scenario is not fully resolved")
	}

	// Also test the snake_case variant to ensure both formats work
	regionTypeRequestSnakeCase := CreateRegionTypeRequest{
		Code:        "state_province",
		Name:        "State Province",
		Description: &description,
		IsActive:    &isActive,
	}

	err = v.Struct(regionTypeRequestSnakeCase)
	if err != nil {
		t.Errorf("Expected no validation error for region-type with code 'state_province', but got: %v", err)
	}

	// Test that invalid cases still properly fail validation (not with "undefined function" but with proper validation messages)
	invalidRegionTypeRequest := CreateRegionTypeRequest{
		Code:        "STATE", // Should fail snake_case validation
		Name:        "State",
		Description: &description,
		IsActive:    &isActive,
	}

	err = v.Struct(invalidRegionTypeRequest)
	if err == nil {
		t.Error("Expected validation error for uppercase code 'STATE', but got none")
	} else {
		// Ensure it's a proper validation error, not an "undefined function" error
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			found := false
			for _, fieldError := range validationErrors {
				if fieldError.Field() == "Code" && fieldError.Tag() == "snake_case" {
					found = true
					break
				}
			}
			if !found {
				t.Error("Expected snake_case validation error for Code field, but didn't find it")
			}
		} else {
			t.Errorf("Expected validator.ValidationErrors type, but got: %T", err)
		}
	}

	// Test new ISO 3166-2 validation for regions (future-proofing)
	type CreateRegionRequest struct {
		Code string `json:"code" validate:"required,region_code_iso"`
		Name string `json:"name" validate:"required,min=2,max=100"`
	}

	// Valid ISO 3166-2 region codes
	validRegionCodes := []string{"US-CA", "IN-MH", "GB-ENG", "AU-NSW"}
	for _, code := range validRegionCodes {
		regionRequest := CreateRegionRequest{
			Code: code,
			Name: "Test Region",
		}

		err = v.Struct(regionRequest)
		if err != nil {
			t.Errorf("Expected no validation error for valid ISO 3166-2 region code '%s', but got: %v", code, err)
		}
	}

	// Invalid region codes for ISO 3166-2
	invalidRegionCodes := []string{"state", "STATE", "US_CA", "USCA", "us-ca"}
	for _, code := range invalidRegionCodes {
		regionRequest := CreateRegionRequest{
			Code: code,
			Name: "Test Region",
		}

		err = v.Struct(regionRequest)
		if err == nil {
			t.Errorf("Expected validation error for invalid ISO 3166-2 region code '%s', but got none", code)
		}
	}
}
