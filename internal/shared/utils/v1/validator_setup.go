package utils

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidatorSetup provides centralized validator configuration
type ValidatorSetup struct {
	validator *validator.Validate
}

// NewValidatorSetup creates a new validator with all custom validation functions registered
func NewValidatorSetup() *ValidatorSetup {
	v := validator.New()

	// Register all custom validation functions
	registerCustomValidations(v)

	return &ValidatorSetup{
		validator: v,
	}
}

// GetValidator returns the configured validator instance
func (vs *ValidatorSetup) GetValidator() *validator.Validate {
	return vs.validator
}

// registerCustomValidations registers all custom validation functions
func registerCustomValidations(v *validator.Validate) {
	// Register snake_case validation for local location codes (districts, cities, areas)
	// Accepts: lowercase single words (e.g., "downtown") OR snake_case (e.g., "metro_area")
	v.RegisterValidation("snake_case", validateSnakeCase)

	// Register ISO 3166-1 country code validation (A-2 standard)
	// Accepts: exactly 2 uppercase letters (e.g., "US", "IN", "GB")
	v.RegisterValidation("country_code_iso", validateCountryCodeISO)

	// Register ISO 3166-2 region code validation (subdivision codes)
	// Accepts: CC-XXX format (e.g., "US-CA", "IN-MH", "GB-ENG")
	v.RegisterValidation("region_code_iso", validateRegionCodeISO)

	// Register existing validations from validator_manager.go
	v.RegisterValidation("postal_code", validatePostalCode)
	v.RegisterValidation("country_code", validateCountryCode)
	v.RegisterValidation("partner_name", validatePartnerName)
	v.RegisterValidation("service_type", validateServiceTypeTag)
	v.RegisterValidation("parcel_category", validateParcelCategoryTag)
	v.RegisterValidation("rating", validateRatingTag)
}

// validateSnakeCase validates that a code is either:
// 1. All lowercase (e.g., "state", "region", "city123")
// 2. Snake case format (e.g., "state_province", "metro_city", "region_1")
//
// This function accepts both formats as requested by the user.
//
// Valid examples: "state", "region", "city1", "state_province", "metro_city", "area_123"
// Invalid examples: "State", "STATE", "state-province", "state province", "_state", "state_"
func validateSnakeCase(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())
	if value == "" {
		return false
	}

	// Check if it's all lowercase (single word with numbers allowed)
	if isAllLowercase(value) {
		return true
	}

	// Check if it's proper snake_case
	return isValidSnakeCase(value)
}

// isAllLowercase checks if the string contains only lowercase letters and numbers
// Used for single-word location codes like "state", "region", "city1"
// Must start with a lowercase letter for consistency with snake_case rules
func isAllLowercase(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Must start with lowercase letter (consistent with snake_case)
	if s[0] < 'a' || s[0] > 'z' {
		return false
	}

	for _, char := range s {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}

// isValidSnakeCase checks if a string is in valid snake_case format
// Rules:
// - Must start with lowercase letter
// - Cannot end with underscore
// - Cannot have consecutive underscores
// - Can contain lowercase letters, numbers, and single underscores as separators
func isValidSnakeCase(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Must start with lowercase letter
	if s[0] < 'a' || s[0] > 'z' {
		return false
	}

	// Cannot end with underscore
	if strings.HasSuffix(s, "_") {
		return false
	}

	// Cannot have consecutive underscores
	if strings.Contains(s, "__") {
		return false
	}

	prevUnderscore := false
	for i, char := range s {
		// Allow lowercase letters, numbers, and underscores
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			prevUnderscore = false
			continue
		} else if char == '_' {
			// No consecutive underscores or leading underscore
			if prevUnderscore || i == 0 {
				return false
			}
			prevUnderscore = true
		} else {
			// Invalid character (uppercase, special chars, etc.)
			return false
		}
	}

	return true
}

// validateCountryCodeISO validates country codes according to ISO 3166-1 A-2 standard
// Country codes are exactly 2 uppercase letters (e.g., "US", "IN", "GB", "DE")
//
// This validation is specifically for country codes and should not be used for other location codes.
func validateCountryCodeISO(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())
	if len(value) != 2 {
		return false
	}

	// Must be exactly 2 uppercase letters for ISO 3166-1 A-2
	for _, char := range value {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}

// validateRegionCodeISO validates region codes according to ISO 3166-2 standard
// Region codes follow CC-XXX format where:
// - CC: ISO 3166-1 alpha-2 country code (2 uppercase letters)
// - XXX: subdivision code (1-3 characters, letters/numbers, usually uppercase)
// Examples: "US-CA", "IN-MH", "GB-ENG", "AU-NSW", "CA-ON"
//
// This validation is specifically for administrative regions (states, provinces, counties, etc.)
func validateRegionCodeISO(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())

	// Must contain exactly one hyphen
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return false
	}

	countryCode := parts[0]
	subdivisionCode := parts[1]

	// Validate country code part (ISO 3166-1 A-2)
	if len(countryCode) != 2 {
		return false
	}
	for _, char := range countryCode {
		if char < 'A' || char > 'Z' {
			return false
		}
	}

	// Validate subdivision code part (1-3 characters, alphanumeric, usually uppercase)
	if len(subdivisionCode) < 1 || len(subdivisionCode) > 3 {
		return false
	}
	for _, char := range subdivisionCode {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}

	return true
}
