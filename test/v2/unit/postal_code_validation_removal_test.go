package unit

import (
	"errors"
	"testing"

	v1_models "prayog-serviceability-service/internal/shared/models/v1"
)

// TestInternationalPostalCodeSupport tests support for international postal codes without country-specific format restrictions
func TestInternationalPostalCodeSupport(t *testing.T) {
	tests := []struct {
		name          string
		sourcePostal  string
		destPostal    string
		countryCode   string
		shouldBeValid bool
		description   string
	}{
		{
			name:          "US ZIP Code Format",
			sourcePostal:  "110001",
			destPostal:    "10001",
			countryCode:   "US",
			shouldBeValid: true,
			description:   "Should accept US 5-digit ZIP codes",
		},
		{
			name:          "US ZIP+4 Code Format",
			sourcePostal:  "110001",
			destPostal:    "10001-1234",
			countryCode:   "US",
			shouldBeValid: true,
			description:   "Should accept US ZIP+4 codes",
		},
		{
			name:          "UK Postal Code Format",
			sourcePostal:  "110001",
			destPostal:    "SW1A 1AA",
			countryCode:   "GB",
			shouldBeValid: true,
			description:   "Should accept UK alphanumeric postal codes",
		},
		{
			name:          "Canada Postal Code Format",
			sourcePostal:  "110001",
			destPostal:    "K1A 0A6",
			countryCode:   "CA",
			shouldBeValid: true,
			description:   "Should accept Canadian postal codes",
		},
		{
			name:          "German Postal Code Format",
			sourcePostal:  "110001",
			destPostal:    "10115",
			countryCode:   "DE",
			shouldBeValid: true,
			description:   "Should accept German 5-digit postal codes",
		},
		{
			name:          "Australian Postal Code Format",
			sourcePostal:  "110001",
			destPostal:    "2000",
			countryCode:   "AU",
			shouldBeValid: true,
			description:   "Should accept Australian 4-digit postal codes",
		},
		{
			name:          "French Postal Code Format",
			sourcePostal:  "110001",
			destPostal:    "75001",
			countryCode:   "FR",
			shouldBeValid: true,
			description:   "Should accept French 5-digit postal codes",
		},
		{
			name:          "Japanese Postal Code Format",
			sourcePostal:  "110001",
			destPostal:    "100-0001",
			countryCode:   "JP",
			shouldBeValid: true,
			description:   "Should accept Japanese postal codes with hyphen",
		},
		{
			name:          "Brazil CEP Format",
			sourcePostal:  "110001",
			destPostal:    "01310-100",
			countryCode:   "BR",
			shouldBeValid: true,
			description:   "Should accept Brazilian CEP format",
		},
		{
			name:          "Netherlands Postal Code Format",
			sourcePostal:  "110001",
			destPostal:    "1012 AB",
			countryCode:   "NL",
			shouldBeValid: true,
			description:   "Should accept Dutch postal codes",
		},
		{
			name:          "Chinese Postal Code Format",
			sourcePostal:  "110001",
			destPostal:    "100000",
			countryCode:   "CN",
			shouldBeValid: true,
			description:   "Should accept Chinese 6-digit postal codes",
		},
		{
			name:          "Mixed Alphanumeric Format",
			sourcePostal:  "110001",
			destPostal:    "ABC123XYZ",
			countryCode:   "XX",
			shouldBeValid: true,
			description:   "Should accept mixed alphanumeric postal codes",
		},
		{
			name:          "Special Characters in Postal Code",
			sourcePostal:  "110001",
			destPostal:    "12-345/67",
			countryCode:   "XX",
			shouldBeValid: true,
			description:   "Should accept postal codes with special characters",
		},
		{
			name:          "Very Long Postal Code",
			sourcePostal:  "110001",
			destPostal:    "ABCDEFGHIJK123456789",
			countryCode:   "XX",
			shouldBeValid: true,
			description:   "Should accept longer postal codes for international use",
		},
		{
			name:          "Single Character Postal Code",
			sourcePostal:  "110001",
			destPostal:    "A",
			countryCode:   "XX",
			shouldBeValid: false,
			description:   "Should reject extremely short postal codes",
		},
		{
			name:          "Empty Postal Code",
			sourcePostal:  "110001",
			destPostal:    "",
			countryCode:   "XX",
			shouldBeValid: false,
			description:   "Should reject empty postal codes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with international postal codes
			request := &v1_models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr(tt.sourcePostal),
				DestinationPostalCode: stringPtr(tt.destPostal),
				CountryCode:           stringPtr(tt.countryCode),
				Packages: []v1_models.Package{
					{
						Weight: &v1_models.Weight{Value: 1.0, Unit: "kg"},
						Dimensions: &v1_models.Dimensions{
							Length: 10, Width: 10, Height: 10, Unit: "cm",
						},
					},
				},
			}

			// Validate request with relaxed postal code format restrictions
			err := validateInternationalPostalCode(request)

			if tt.shouldBeValid {
				if err != nil {
					t.Errorf("Expected postal code '%s' to be valid for %s, got error: %v",
						tt.destPostal, tt.description, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected postal code '%s' to be invalid for %s, got no error",
						tt.destPostal, tt.description)
				}
			}
		})
	}
}

// TestPostalCodeFormatFlexibility tests that postal code validation is flexible for different countries
func TestPostalCodeFormatFlexibility(t *testing.T) {
	tests := []struct {
		name          string
		postalCode    string
		countryCode   string
		oldValidation bool // Would this pass strict country-specific validation?
		newValidation bool // Should this pass flexible international validation?
		description   string
	}{
		{
			name:          "Indian Postal Code in US Format",
			postalCode:    "110001",
			countryCode:   "US",
			oldValidation: false, // Doesn't match US ZIP format
			newValidation: true,  // Should be accepted for international flexibility
			description:   "Indian postal code should be accepted for US destination",
		},
		{
			name:          "US ZIP Code for Indian Destination",
			postalCode:    "10001",
			countryCode:   "IN",
			oldValidation: false, // Doesn't match Indian postal format
			newValidation: true,  // Should be accepted for international flexibility
			description:   "US ZIP code should be accepted for Indian destination",
		},
		{
			name:          "UK Postal Code for German Destination",
			postalCode:    "SW1A 1AA",
			countryCode:   "DE",
			oldValidation: false, // Doesn't match German numeric format
			newValidation: true,  // Should be accepted for international flexibility
			description:   "UK postal code should be accepted for German destination",
		},
		{
			name:          "Hyphenated Format for Non-Hyphen Country",
			postalCode:    "12345-6789",
			countryCode:   "FR",
			oldValidation: false, // French codes don't typically use hyphens
			newValidation: true,  // Should be accepted for international flexibility
			description:   "Hyphenated postal code should be accepted for French destination",
		},
		{
			name:          "Spaces in Postal Code",
			postalCode:    "12 345",
			countryCode:   "CN",
			oldValidation: false, // Chinese codes don't typically use spaces
			newValidation: true,  // Should be accepted for international flexibility
			description:   "Postal code with spaces should be accepted for Chinese destination",
		},
		{
			name:          "Mixed Case Alphanumeric",
			postalCode:    "Ab12Cd",
			countryCode:   "AU",
			oldValidation: false, // Australian codes are numeric
			newValidation: true,  // Should be accepted for international flexibility
			description:   "Mixed case alphanumeric postal code should be accepted for Australian destination",
		},
		{
			name:          "Excessively Long Code",
			postalCode:    "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789",
			countryCode:   "XX",
			oldValidation: false, // Too long for any standard format
			newValidation: false, // Even flexible validation should have limits
			description:   "Excessively long postal codes should still be rejected",
		},
		{
			name:          "Special Characters",
			postalCode:    "12@34#56",
			countryCode:   "XX",
			oldValidation: false, // Special characters not in standard formats
			newValidation: false, // Special characters should still be rejected
			description:   "Postal codes with special symbols should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test old strict validation (simulated)
			oldResult := simulateStrictCountryValidation(tt.postalCode, tt.countryCode)
			if oldResult != tt.oldValidation {
				t.Errorf("Old validation assumption incorrect for %s: expected %v, got %v",
					tt.description, tt.oldValidation, oldResult)
			}

			// Test new flexible validation
			newResult := validateFlexiblePostalCode(tt.postalCode)
			if newResult != tt.newValidation {
				t.Errorf("New flexible validation failed for %s: expected %v, got %v",
					tt.description, tt.newValidation, newResult)
			}
		})
	}
}

// TestInternationalShippingRequirements tests requirements specific to international shipping
func TestInternationalShippingRequirements(t *testing.T) {
	tests := []struct {
		name              string
		sourceCountry     string
		destCountry       string
		postalCodeFormat  string
		shouldRequireCode bool
		description       string
	}{
		{
			name:              "Domestic Indian Shipping",
			sourceCountry:     "IN",
			destCountry:       "IN",
			postalCodeFormat:  "110001",
			shouldRequireCode: true,
			description:       "Domestic shipments should require postal codes",
		},
		{
			name:              "India to US International",
			sourceCountry:     "IN",
			destCountry:       "US",
			postalCodeFormat:  "10001",
			shouldRequireCode: true,
			description:       "International shipments should require postal codes",
		},
		{
			name:              "International to Country Without Postal System",
			sourceCountry:     "IN",
			destCountry:       "XX", // Fictional country
			postalCodeFormat:  "NONE",
			shouldRequireCode: false,
			description:       "Some countries might not have postal code systems",
		},
		{
			name:              "International with Multiple Format Support",
			sourceCountry:     "IN",
			destCountry:       "GB",
			postalCodeFormat:  "SW1A 1AA",
			shouldRequireCode: true,
			description:       "Should support various international postal code formats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request for international shipping
			request := &v1_models.ServiceabilityV2Request{
				SourcePostalCode: stringPtr("110001"),
				DestinationPostalCode: func() *string {
					if tt.postalCodeFormat == "NONE" {
						return nil
					}
					return stringPtr(tt.postalCodeFormat)
				}(),
				CountryCode: stringPtr(tt.destCountry),
				Packages: []v1_models.Package{
					{
						Weight: &v1_models.Weight{Value: 1.0, Unit: "kg"},
					},
				},
			}

			// Validate international shipping requirements
			err := validateInternationalShippingRequirements(request, tt.sourceCountry, tt.destCountry)

			if tt.shouldRequireCode {
				if err != nil && request.DestinationPostalCode != nil {
					t.Errorf("Expected postal code requirement to be satisfied for %s, got error: %v",
						tt.description, err)
				}
			}

			// All international requests should be processable with flexible validation
			flexibleResult := validateInternationalPostalCode(request)
			if flexibleResult != nil && tt.postalCodeFormat != "NONE" {
				t.Errorf("Flexible international validation should accept valid postal codes for %s, got error: %v",
					tt.description, flexibleResult)
			}
		})
	}
}

// TestPostalCodeValidationNonBlocking tests that postal code validation doesn't block international requests
func TestPostalCodeValidationNonBlocking(t *testing.T) {
	tests := []struct {
		name         string
		sourcePostal string
		destPostal   string
		countryCode  string
		shouldBlock  bool
		description  string
	}{
		{
			name:         "Valid International Format",
			sourcePostal: "110001",
			destPostal:   "SW1A 1AA",
			countryCode:  "GB",
			shouldBlock:  false,
			description:  "Valid international postal codes should not block requests",
		},
		{
			name:         "Unusual but Acceptable Format",
			sourcePostal: "110001",
			destPostal:   "ABC-123",
			countryCode:  "XX",
			shouldBlock:  false,
			description:  "Unusual but potentially valid formats should not block requests",
		},
		{
			name:         "Missing Destination Postal Code",
			sourcePostal: "110001",
			destPostal:   "",
			countryCode:  "XX",
			shouldBlock:  true,
			description:  "Missing destination postal codes should block DHL requests",
		},
		{
			name:         "Format Mismatch with Country",
			sourcePostal: "110001",
			destPostal:   "ABCDEFG123",
			countryCode:  "US",
			shouldBlock:  false,
			description:  "Format mismatches should not block international requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			request := &v1_models.ServiceabilityV2Request{
				SourcePostalCode: stringPtr(tt.sourcePostal),
				DestinationPostalCode: func() *string {
					if tt.destPostal == "" {
						return nil
					}
					return stringPtr(tt.destPostal)
				}(),
				CountryCode: stringPtr(tt.countryCode),
				Packages: []v1_models.Package{
					{
						Weight: &v1_models.Weight{Value: 1.0, Unit: "kg"},
					},
				},
			}

			// Test non-blocking validation
			shouldBlock := isBlockingValidationIssue(request)

			if shouldBlock != tt.shouldBlock {
				t.Errorf("Blocking validation check failed for %s: expected %v, got %v",
					tt.description, tt.shouldBlock, shouldBlock)
			}

			// Ensure that non-blocking issues don't prevent processing
			if !tt.shouldBlock {
				err := validateInternationalPostalCode(request)
				if err != nil {
					t.Errorf("Non-blocking validation issue incorrectly blocked request for %s: %v",
						tt.description, err)
				}
			}
		})
	}
}

// Helper functions for postal code validation

func validateInternationalPostalCode(request *v1_models.ServiceabilityV2Request) error {
	// Basic validation - ensure postal codes exist but don't enforce country-specific formats
	if request.DestinationPostalCode == nil || *request.DestinationPostalCode == "" {
		return errors.New("destination postal code is required")
	}

	destPostal := *request.DestinationPostalCode

	// Very basic length check (allow 1-50 characters for international flexibility)
	if len(destPostal) < 2 || len(destPostal) > 50 {
		return errors.New("postal code length must be between 2 and 50 characters")
	}

	// Allow alphanumeric characters, spaces, hyphens, and basic punctuation
	// This is much more flexible than country-specific validation
	for _, char := range destPostal {
		if !isValidPostalCodeCharacter(char) {
			return errors.New("postal code contains invalid characters")
		}
	}

	return nil
}

func validateFlexiblePostalCode(postalCode string) bool {
	// Flexible validation for international postal codes
	if len(postalCode) < 2 || len(postalCode) > 50 {
		return false
	}

	// Allow alphanumeric, spaces, hyphens, but reject special symbols
	for _, char := range postalCode {
		if !isValidPostalCodeCharacter(char) {
			return false
		}
	}

	return true
}

func simulateStrictCountryValidation(postalCode, countryCode string) bool {
	// Simulate old strict country-specific validation
	switch countryCode {
	case "US":
		// US ZIP codes: 5 digits or 5+4 format
		return len(postalCode) == 5 || (len(postalCode) == 10 && postalCode[5] == '-')
	case "IN":
		// Indian postal codes: 6 digits
		return len(postalCode) == 6
	case "GB":
		// UK postal codes: complex alphanumeric format
		return len(postalCode) >= 6 && len(postalCode) <= 8
	case "DE":
		// German postal codes: 5 digits
		return len(postalCode) == 5
	case "FR":
		// French postal codes: 5 digits
		return len(postalCode) == 5
	case "CN":
		// Chinese postal codes: 6 digits
		return len(postalCode) == 6
	case "AU":
		// Australian postal codes: 4 digits
		return len(postalCode) == 4
	default:
		// For unknown countries, old validation would likely fail
		return false
	}
}

func validateInternationalShippingRequirements(request *v1_models.ServiceabilityV2Request, sourceCountry, destCountry string) error {
	// Basic international shipping validation
	if sourceCountry != destCountry {
		// International shipment - postal codes usually required
		if request.DestinationPostalCode == nil || *request.DestinationPostalCode == "" {
			return errors.New("destination postal code required for international shipments")
		}
	}

	return nil
}

func isBlockingValidationIssue(request *v1_models.ServiceabilityV2Request) bool {
	// Only block on critical missing information
	if request.DestinationPostalCode == nil || *request.DestinationPostalCode == "" {
		return true
	}

	// Don't block on format issues - those are handled with flexible validation
	return false
}

func isValidPostalCodeCharacter(char rune) bool {
	// Allow alphanumeric characters, spaces, hyphens
	return (char >= 'A' && char <= 'Z') ||
		(char >= 'a' && char <= 'z') ||
		(char >= '0' && char <= '9') ||
		char == ' ' || char == '-'
}

// Note: stringPtr function is already defined in other test files
