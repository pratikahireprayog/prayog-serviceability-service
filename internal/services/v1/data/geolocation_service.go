package services

import (
	"context"
	"strings"
	"time"

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
	geoLocationRepo repositories.GeoLocationRepository
}

// NewGeolocationService creates a new geolocation service instance
func NewGeolocationService(geoLocationRepo repositories.GeoLocationRepository) GeolocationService {
	return &geolocationService{
		geoLocationRepo: geoLocationRepo,
	}
}

// GetCountryCodeByPostalCode retrieves the country code for a given postal code
func (g *geolocationService) GetCountryCodeByPostalCode(ctx context.Context, postalCode string) (*string, error) {
	if strings.TrimSpace(postalCode) == "" {
		return nil, errors.ErrValidationFailed("postal_code", "cannot be empty")
	}

	// Check if geolocation repository is available
	if g.geoLocationRepo == nil {
		return nil, errors.ErrServiceUnavailable("geolocation repository")
	}

	// Create a timeout context specifically for database operations (5 minutes)
	// This is longer than the default context to handle potential database performance issues
	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get geolocation data by postal code with extended timeout
	geoLocation, err := g.geoLocationRepo.GetByID(dbCtx, postalCode)
	if err != nil {
		// Properly propagate structured errors from repository
		return nil, err
	}

	// Get country code from the GeoLocation model
	if geoLocation.CountryCode != "" {
		return &geoLocation.CountryCode, nil
	}

	// If no country code found, return error
	return nil, errors.ErrInternalError("postal code found but country code is missing", nil)
}

// GetLocationHierarchy retrieves the complete location hierarchy for a postal code
func (g *geolocationService) GetLocationHierarchy(ctx context.Context, postalCode string) (*models.LocationHierarchy, error) {
	if strings.TrimSpace(postalCode) == "" {
		return nil, errors.ErrValidationFailed("postal_code", "cannot be empty")
	}

	// Check if geolocation repository is available
	if g.geoLocationRepo == nil {
		return nil, errors.ErrServiceUnavailable("geolocation repository")
	}

	// Get country code for the postal code
	countryCode, err := g.GetCountryCodeByPostalCode(ctx, postalCode)
	if err != nil {
		// Propagate the error from GetCountryCodeByPostalCode
		return nil, err
	}

	// For now, return a basic hierarchy with country code
	// This can be enhanced later to build a complete hierarchy from geolocation data
	hierarchy := &models.LocationHierarchy{
		CountryCode: *countryCode,
		PostalCode:  postalCode,
	}

	return hierarchy, nil
}

// ValidateInternationalRequest validates if a request is international and returns country code
func (g *geolocationService) ValidateInternationalRequest(ctx context.Context, postalCode string) (bool, *string, error) {
	if strings.TrimSpace(postalCode) == "" {
		return false, nil, errors.ErrValidationFailed("postal_code", "cannot be empty")
	}

	// Check if geolocation repository is available
	if g.geoLocationRepo == nil {
		return false, nil, errors.ErrServiceUnavailable("geolocation repository")
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
