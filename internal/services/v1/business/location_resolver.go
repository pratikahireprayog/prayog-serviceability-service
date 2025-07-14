package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// locationResolver implements the LocationResolver interface
type locationResolver struct {
	partnerService interfaces.PartnerServiceClient
}

// NewLocationResolver creates a new LocationResolver instance
func NewLocationResolver(partnerService interfaces.PartnerServiceClient) interfaces.LocationResolver {
	return &locationResolver{
		partnerService: partnerService,
	}
}

// ValidatePostalCode validates a postal code and returns location validation result
func (lr *locationResolver) ValidatePostalCode(ctx context.Context, postalCode, countryCode string) (*models.LocationValidationResult, error) {
	if postalCode == "" || countryCode == "" {
		return &models.LocationValidationResult{
			IsValid: false,
			Error:   "postal_code and country_code are required",
		}, nil
	}

	// Validate postal code format based on country
	isValid, reason := lr.validatePostalCodeFormat(postalCode, countryCode)
	if !isValid {
		return &models.LocationValidationResult{
			IsValid: false,
			Error:   reason,
		}, nil
	}

	// Check if we have any partners serving this location
	// Use postal code as location ID for partner lookup
	partners, err := lr.partnerService.GetPartnersByLocation(ctx, "postal_code", postalCode)
	if err != nil {
		// If partner service is down, we can still validate the format
		return &models.LocationValidationResult{
			IsValid: true,
			Error:   "format_valid_partner_service_unavailable",
		}, nil
	}

	// If we have partners serving this location, it's valid
	if len(partners) > 0 {
		return &models.LocationValidationResult{
			IsValid: true,
			Error:   "valid_with_partner_coverage",
		}, nil
	}

	return &models.LocationValidationResult{
		IsValid: true,
		Error:   "valid_format_no_partner_coverage",
	}, nil
}

// GetLocationHierarchy resolves a postal code to a location hierarchy
func (lr *locationResolver) GetLocationHierarchy(ctx context.Context, postalCode, countryCode string) (*models.LocationHierarchy, error) {
	// Validate first
	validation, err := lr.ValidatePostalCode(ctx, postalCode, countryCode)
	if err != nil {
		return nil, fmt.Errorf("failed to validate postal code: %w", err)
	}

	if !validation.IsValid {
		return nil, fmt.Errorf("invalid postal code: %s", validation.Error)
	}

	// Create location hierarchy based on postal code patterns
	// This is a simplified implementation - in production, you'd have a proper location database
	locationHierarchy := &models.LocationHierarchy{
		PostalCode:  postalCode,
		CountryCode: countryCode,
	}

	// Add location details based on known patterns for different countries
	lr.enrichLocationHierarchy(locationHierarchy)

	return locationHierarchy, nil
}

// IsPostalCodeActive checks if a postal code is active for delivery services
func (lr *locationResolver) IsPostalCodeActive(ctx context.Context, postalCode, countryCode string) (bool, error) {
	// Check if postal code is valid
	validation, err := lr.ValidatePostalCode(ctx, postalCode, countryCode)
	if err != nil {
		return false, fmt.Errorf("failed to validate postal code: %w", err)
	}

	// A postal code is considered active if it's valid and has partner coverage
	if validation.IsValid && strings.Contains(validation.Error, "partner_coverage") {
		return true, nil
	}

	// For now, consider all valid postal codes as active
	return validation.IsValid, nil
}

// validatePostalCodeFormat validates postal code format by country
func (lr *locationResolver) validatePostalCodeFormat(postalCode, countryCode string) (bool, string) {
	countryCode = strings.ToUpper(countryCode)

	switch countryCode {
	case "IN":
		// Indian postal codes: 6 digits
		matched, _ := regexp.MatchString(`^\d{6}$`, postalCode)
		if !matched {
			return false, "invalid_indian_postal_code_format"
		}
	case "US":
		// US postal codes: 5 digits or 5+4 format
		matched, _ := regexp.MatchString(`^\d{5}(-\d{4})?$`, postalCode)
		if !matched {
			return false, "invalid_us_postal_code_format"
		}
	case "GB", "UK":
		// UK postal codes: complex format
		matched, _ := regexp.MatchString(`^[A-Z]{1,2}[0-9]{1,2}[A-Z]?\s?[0-9][A-Z]{2}$`, strings.ToUpper(postalCode))
		if !matched {
			return false, "invalid_uk_postal_code_format"
		}
	default:
		// For other countries, just check it's not empty and has reasonable length
		if len(postalCode) < 3 || len(postalCode) > 10 {
			return false, "invalid_postal_code_length"
		}
	}

	return true, "valid_format"
}

// enrichLocationHierarchy adds location details based on postal code patterns
func (lr *locationResolver) enrichLocationHierarchy(location *models.LocationHierarchy) {
	switch location.CountryCode {
	case "IN":
		lr.enrichIndianLocation(location)
	case "US":
		lr.enrichUSLocation(location)
	default:
		// Generic enrichment
		location.CountryName = location.CountryCode
	}
}

// enrichIndianLocation adds Indian location details based on postal code
func (lr *locationResolver) enrichIndianLocation(location *models.LocationHierarchy) {
	location.CountryName = "India"

	// Simple mapping based on postal code prefixes - this would be from a database in production
	switch {
	case strings.HasPrefix(location.PostalCode, "110"):
		location.RegionName = "Delhi"
		location.CityName = "New Delhi"
		location.RegionCode = "DL"
		location.CityCode = "ND"
	case strings.HasPrefix(location.PostalCode, "400"):
		location.RegionName = "Maharashtra"
		location.CityName = "Mumbai"
		location.RegionCode = "MH"
		location.CityCode = "MUM"
	case strings.HasPrefix(location.PostalCode, "560"):
		location.RegionName = "Karnataka"
		location.CityName = "Bangalore"
		location.RegionCode = "KA"
		location.CityCode = "BLR"
	case strings.HasPrefix(location.PostalCode, "700"):
		location.RegionName = "West Bengal"
		location.CityName = "Kolkata"
		location.RegionCode = "WB"
		location.CityCode = "KOL"
	case strings.HasPrefix(location.PostalCode, "600"):
		location.RegionName = "Tamil Nadu"
		location.CityName = "Chennai"
		location.RegionCode = "TN"
		location.CityCode = "CHE"
	case strings.HasPrefix(location.PostalCode, "500"):
		location.RegionName = "Telangana"
		location.CityName = "Hyderabad"
		location.RegionCode = "TS"
		location.CityCode = "HYD"
	default:
		// Generic Indian location
		location.RegionName = "Unknown"
		location.CityName = "Unknown"
		location.RegionCode = "UK"
		location.CityCode = "UK"
	}
}

// enrichUSLocation adds US location details based on postal code
func (lr *locationResolver) enrichUSLocation(location *models.LocationHierarchy) {
	location.CountryName = "United States"

	// Extract ZIP code (first 5 digits)
	zipCode := location.PostalCode
	if len(zipCode) > 5 {
		zipCode = zipCode[:5]
	}

	// Simple mapping based on ZIP code ranges - this would be from a database in production
	switch {
	case zipCode >= "10000" && zipCode <= "14999":
		location.RegionName = "New York"
		location.RegionCode = "NY"
	case zipCode >= "90000" && zipCode <= "96699":
		location.RegionName = "California"
		location.RegionCode = "CA"
	case zipCode >= "60000" && zipCode <= "60999":
		location.RegionName = "Illinois"
		location.CityName = "Chicago"
		location.RegionCode = "IL"
		location.CityCode = "CHI"
	default:
		location.RegionName = "Unknown"
		location.RegionCode = "UK"
	}
}
