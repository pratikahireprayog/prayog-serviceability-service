package services

import (
	"context"
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"
)

// GeolocationService defines the interface for geolocation operations
type GeolocationService interface {
	GetCountryCodeByPostalCode(ctx context.Context, postalCode string) (*string, error)
	GetLocationHierarchy(ctx context.Context, postalCode string) (*models.LocationHierarchy, error)
	ValidateInternationalRequest(ctx context.Context, postalCode string) (bool, *string, error)
}

// geolocationService implements the GeolocationService interface
type geolocationService struct {
	postalCodeRepo repositories.PostalCodeRepository
}

// NewGeolocationService creates a new geolocation service instance
func NewGeolocationService(postalCodeRepo repositories.PostalCodeRepository) GeolocationService {
	return &geolocationService{
		postalCodeRepo: postalCodeRepo,
	}
}

// GetCountryCodeByPostalCode retrieves the country code for a given postal code
func (g *geolocationService) GetCountryCodeByPostalCode(ctx context.Context, postalCode string) (*string, error) {
	if strings.TrimSpace(postalCode) == "" {
		return nil, fmt.Errorf("postal code cannot be empty")
	}

	// Check if postal code repository is available
	if g.postalCodeRepo == nil {
		// Graceful degradation: return default country code
		defaultCountry := "IN"
		return &defaultCountry, nil
	}

	// First try to get postal code with preloaded country information
	postal, err := g.postalCodeRepo.GetByCode(ctx, postalCode)
	if err != nil {
		// If we can't get postal code, return default country code
		defaultCountry := "IN"
		return &defaultCountry, fmt.Errorf("failed to get postal code details, defaulting to IN: %w", err)
	}

	// Return country code from the postal code data
	if postal.Country != nil && postal.Country.Code != "" {
		return &postal.Country.Code, nil
	}

	// If no country code found, return default
	defaultCountry := "IN"
	return &defaultCountry, nil
}

// GetLocationHierarchy retrieves the complete location hierarchy for a postal code
func (g *geolocationService) GetLocationHierarchy(ctx context.Context, postalCode string) (*models.LocationHierarchy, error) {
	if strings.TrimSpace(postalCode) == "" {
		return nil, fmt.Errorf("postal code cannot be empty")
	}

	// Check if postal code repository is available
	if g.postalCodeRepo == nil {
		// Return a default hierarchy for graceful degradation
		defaultHierarchy := &models.LocationHierarchy{
			CountryCode: "IN",
			CountryName: "India",
			PostalCode:  postalCode,
		}
		return defaultHierarchy, nil
	}

	// First try to get country code for the postal code
	countryCode, err := g.GetCountryCodeByPostalCode(ctx, postalCode)
	if err != nil {
		// Return default hierarchy if we can't get country code
		defaultHierarchy := &models.LocationHierarchy{
			CountryCode: "IN",
			CountryName: "India",
			PostalCode:  postalCode,
		}
		return defaultHierarchy, fmt.Errorf("failed to get country code, returning default hierarchy: %w", err)
	}

	// Get complete location hierarchy
	hierarchy, err := g.postalCodeRepo.GetLocationHierarchy(ctx, postalCode, *countryCode)
	if err != nil {
		// Return default hierarchy if we can't get full hierarchy
		defaultHierarchy := &models.LocationHierarchy{
			CountryCode: *countryCode,
			CountryName: "Unknown",
			PostalCode:  postalCode,
		}
		return defaultHierarchy, fmt.Errorf("failed to get location hierarchy, returning default: %w", err)
	}

	return hierarchy, nil
}

// ValidateInternationalRequest validates if a request is international and returns country code
func (g *geolocationService) ValidateInternationalRequest(ctx context.Context, postalCode string) (bool, *string, error) {
	if strings.TrimSpace(postalCode) == "" {
		return false, nil, fmt.Errorf("postal code cannot be empty")
	}

	// Check if postal code repository is available
	if g.postalCodeRepo == nil {
		// Graceful degradation: assume domestic if we can't check
		// This allows the system to continue working even without geolocation
		defaultCountry := "IN"
		return false, &defaultCountry, nil
	}

	// Get country code for the postal code
	countryCode, err := g.GetCountryCodeByPostalCode(ctx, postalCode)
	if err != nil {
		// If we can't get country code, assume domestic for safety
		defaultCountry := "IN"
		return false, &defaultCountry, fmt.Errorf("failed to get country code, defaulting to domestic: %w", err)
	}

	// Define default origin country (India)
	originCountry := "IN"

	// Check if destination country is different from origin country
	isInternational := strings.ToUpper(*countryCode) != originCountry

	return isInternational, countryCode, nil
}
