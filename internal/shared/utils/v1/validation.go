package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"prayog-serviceability-service/internal/shared/constants/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// PostalCodeValidator provides postal code validation functionality
type PostalCodeValidator struct {
	countryPatterns map[string]*regexp.Regexp
}

// NewPostalCodeValidator creates a new postal code validator
func NewPostalCodeValidator() *PostalCodeValidator {
	return &PostalCodeValidator{
		countryPatterns: map[string]*regexp.Regexp{
			"IN": regexp.MustCompile(`^[1-9][0-9]{5}$`),                                                  // India: 6 digits, first digit 1-9
			"US": regexp.MustCompile(`^\d{5}(-\d{4})?$`),                                                 // USA: 5 digits or 5+4 format
			"CA": regexp.MustCompile(`^[ABCEGHJ-NPRSTVXY]\d[ABCEGHJ-NPRSTV-Z] ?\d[ABCEGHJ-NPRSTV-Z]\d$`), // Canada
			"GB": regexp.MustCompile(`^[A-Z]{1,2}\d[A-Z\d]? ?\d[A-Z]{2}$`),                               // UK
			"AU": regexp.MustCompile(`^\d{4}$`),                                                          // Australia: 4 digits
		},
	}
}

// ValidatePostalCode validates postal code format for a given country
func (v *PostalCodeValidator) ValidatePostalCode(postalCode, countryCode string) error {
	if len(postalCode) < constants.MinPostalCodeLength || len(postalCode) > constants.MaxPostalCodeLength {
		return fmt.Errorf("postal code length must be between %d and %d characters",
			constants.MinPostalCodeLength, constants.MaxPostalCodeLength)
	}

	pattern, exists := v.countryPatterns[strings.ToUpper(countryCode)]
	if !exists {
		// For countries without specific patterns, use generic validation
		return v.validateGenericPostalCode(postalCode)
	}

	if !pattern.MatchString(strings.ToUpper(postalCode)) {
		return fmt.Errorf("invalid postal code format for country %s", countryCode)
	}

	return nil
}

// validateGenericPostalCode provides generic postal code validation
func (v *PostalCodeValidator) validateGenericPostalCode(postalCode string) error {
	if strings.TrimSpace(postalCode) == "" {
		return fmt.Errorf("postal code cannot be empty")
	}

	// Check for valid alphanumeric characters and common postal code symbols
	validChars := regexp.MustCompile(`^[A-Za-z0-9\s\-]+$`)
	if !validChars.MatchString(postalCode) {
		return fmt.Errorf("postal code contains invalid characters")
	}

	return nil
}

// CountryCodeValidator provides country code validation
type CountryCodeValidator struct {
	validCodes map[string]bool
}

// NewCountryCodeValidator creates a new country code validator
func NewCountryCodeValidator() *CountryCodeValidator {
	// Common country codes - in production this would be loaded from database or config
	validCodes := map[string]bool{
		"IN": true, "US": true, "CA": true, "GB": true, "AU": true,
		"DE": true, "FR": true, "JP": true, "CN": true, "BR": true,
		"MX": true, "IT": true, "ES": true, "NL": true, "SE": true,
		"NO": true, "DK": true, "FI": true, "BE": true, "CH": true,
	}

	return &CountryCodeValidator{validCodes: validCodes}
}

// ValidateCountryCode validates if the country code is valid
func (v *CountryCodeValidator) ValidateCountryCode(countryCode string) error {
	if len(countryCode) != constants.CountryCodeLength {
		return fmt.Errorf("country code must be exactly %d characters", constants.CountryCodeLength)
	}

	upperCode := strings.ToUpper(countryCode)
	if !v.validCodes[upperCode] {
		return fmt.Errorf("invalid country code: %s", countryCode)
	}

	return nil
}

// ServiceabilityRequestValidator validates serviceability requests
type ServiceabilityRequestValidator struct {
	postalCodeValidator  *PostalCodeValidator
	countryCodeValidator *CountryCodeValidator
}

// NewServiceabilityRequestValidator creates a new request validator
func NewServiceabilityRequestValidator() *ServiceabilityRequestValidator {
	return &ServiceabilityRequestValidator{
		postalCodeValidator:  NewPostalCodeValidator(),
		countryCodeValidator: NewCountryCodeValidator(),
	}
}

// ValidateServiceabilityRequest validates a serviceability check request
func (v *ServiceabilityRequestValidator) ValidateServiceabilityRequest(req *models.ServiceabilityCheckRequest) error {
	// Validate country code
	if err := v.countryCodeValidator.ValidateCountryCode(req.CountryCode); err != nil {
		return err
	}

	// Check request type and validate accordingly
	isSingleLocation := req.PostalCode != nil
	isRoute := req.PickupPostalCode != nil && req.DeliveryPostalCode != nil

	if !isSingleLocation && !isRoute {
		return fmt.Errorf("request must specify either postal_code or both pickup_postal_code and delivery_postal_code")
	}

	if isSingleLocation && isRoute {
		return fmt.Errorf("request cannot specify both postal_code and pickup/delivery postal codes")
	}

	// Validate postal codes
	if isSingleLocation {
		if err := v.postalCodeValidator.ValidatePostalCode(*req.PostalCode, req.CountryCode); err != nil {
			return fmt.Errorf("invalid postal_code: %w", err)
		}
	}

	if isRoute {
		if err := v.postalCodeValidator.ValidatePostalCode(*req.PickupPostalCode, req.CountryCode); err != nil {
			return fmt.Errorf("invalid pickup_postal_code: %w", err)
		}
		if err := v.postalCodeValidator.ValidatePostalCode(*req.DeliveryPostalCode, req.CountryCode); err != nil {
			return fmt.Errorf("invalid delivery_postal_code: %w", err)
		}
	}

	return nil
}

// ValidateBulkRequest validates a bulk serviceability request
func (v *ServiceabilityRequestValidator) ValidateBulkRequest(req *models.BulkServiceabilityRequest) error {
	if len(req.Requests) == 0 {
		return fmt.Errorf("bulk request must contain at least one request")
	}

	if len(req.Requests) > constants.MaxBulkRequestSize {
		return fmt.Errorf("bulk request cannot contain more than %d requests", constants.MaxBulkRequestSize)
	}

	for i, subReq := range req.Requests {
		if err := v.ValidateServiceabilityRequest(&subReq); err != nil {
			return fmt.Errorf("request at index %d: %w", i, err)
		}
	}

	return nil
}

// StringUtils provides string utility functions
type StringUtils struct{}

// ToUpperCase converts string to uppercase
func (StringUtils) ToUpperCase(s string) string {
	return strings.ToUpper(s)
}

// ToTitleCase converts string to title case
func (StringUtils) ToTitleCase(s string) string {
	return strings.Title(strings.ToLower(s))
}

// NormalizePostalCode normalizes postal code format
func (StringUtils) NormalizePostalCode(postalCode string) string {
	// Remove extra spaces and convert to uppercase
	normalized := strings.TrimSpace(postalCode)
	normalized = strings.ToUpper(normalized)

	// Remove multiple consecutive spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	normalized = spaceRegex.ReplaceAllString(normalized, " ")

	return normalized
}

// SanitizeString removes potentially harmful characters
func (StringUtils) SanitizeString(s string) string {
	// Remove control characters
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

// Contains checks if slice contains a string
func (StringUtils) Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase checks if slice contains a string (case insensitive)
func (StringUtils) ContainsIgnoreCase(slice []string, item string) bool {
	lowerItem := strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == lowerItem {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from slice
func (StringUtils) RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ArrayUtils provides array utility functions
type ArrayUtils struct{}

// ChunkStrings splits a string slice into chunks of specified size
func (ArrayUtils) ChunkStrings(slice []string, chunkSize int) [][]string {
	if chunkSize <= 0 {
		return nil
	}

	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// FilterStrings filters string slice based on predicate function
func (ArrayUtils) FilterStrings(slice []string, predicate func(string) bool) []string {
	var result []string
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// MapStrings applies a function to each element in string slice
func (ArrayUtils) MapStrings(slice []string, mapper func(string) string) []string {
	result := make([]string, len(slice))
	for i, item := range slice {
		result[i] = mapper(item)
	}
	return result
}

// Convenience validation functions

// IsValidCountryCode validates a country code using the default validator
func IsValidCountryCode(countryCode string) bool {
	validator := NewCountryCodeValidator()
	return validator.ValidateCountryCode(countryCode) == nil
}

// IsValidPostalCodeForCountry validates a postal code for a specific country using the default validator
func IsValidPostalCodeForCountry(postalCode, countryCode string) bool {
	validator := NewPostalCodeValidator()
	return validator.ValidatePostalCode(postalCode, countryCode) == nil
}
