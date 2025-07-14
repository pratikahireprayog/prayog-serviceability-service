package services

import (
	"context"
	"strings"

	"prayog-serviceability-service/internal/shared/errors"
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
		return nil, errors.ErrValidationFailed("postal_code", "cannot be empty")
	}

	// Check if postal code repository is available
	if g.postalCodeRepo == nil {
		return nil, errors.ErrServiceUnavailable("postal code repository")
	}

	// Get postal code with preloaded country information
	postal, err := g.postalCodeRepo.GetByCode(ctx, postalCode)
	if err != nil {
		// Properly propagate structured errors from repository
		return nil, err
	}

	// First, try to get country code from the Country relationship
	if postal.Country != nil && postal.Country.Code != "" {
		return &postal.Country.Code, nil
	}

	// If Country relationship is not available, try the direct country_code field
	if postal.CountryCode != nil && *postal.CountryCode != "" {
		return postal.CountryCode, nil
	}

	// If no country code found in either place, return error
	return nil, errors.ErrInternalError("postal code found but country code is missing", nil)
}

// GetLocationHierarchy retrieves the complete location hierarchy for a postal code
func (g *geolocationService) GetLocationHierarchy(ctx context.Context, postalCode string) (*models.LocationHierarchy, error) {
	if strings.TrimSpace(postalCode) == "" {
		return nil, errors.ErrValidationFailed("postal_code", "cannot be empty")
	}

	// Check if postal code repository is available
	if g.postalCodeRepo == nil {
		return nil, errors.ErrServiceUnavailable("postal code repository")
	}

	// Get country code for the postal code
	countryCode, err := g.GetCountryCodeByPostalCode(ctx, postalCode)
	if err != nil {
		// Propagate the error from GetCountryCodeByPostalCode
		return nil, err
	}

	// Get complete location hierarchy
	hierarchy, err := g.postalCodeRepo.GetLocationHierarchy(ctx, postalCode, *countryCode)
	if err != nil {
		// Propagate structured errors from repository
		return nil, err
	}

	return hierarchy, nil
}

// ValidateInternationalRequest validates if a request is international and returns country code
func (g *geolocationService) ValidateInternationalRequest(ctx context.Context, postalCode string) (bool, *string, error) {
	if strings.TrimSpace(postalCode) == "" {
		return false, nil, errors.ErrValidationFailed("postal_code", "cannot be empty")
	}

	// Check if postal code repository is available
	if g.postalCodeRepo == nil {
		return false, nil, errors.ErrServiceUnavailable("postal code repository")
	}

	// Get country code for the postal code
	countryCode, err := g.GetCountryCodeByPostalCode(ctx, postalCode)
	if err != nil {
		// Propagate the error from GetCountryCodeByPostalCode
		return false, nil, err
	}

	// Define default origin country (India)
	originCountry := "IN"

	// Check if destination country is different from origin country
	isInternational := strings.ToUpper(*countryCode) != originCountry

	return isInternational, countryCode, nil
}
