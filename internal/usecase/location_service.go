package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/prayog/serviceability/internal/domain"
)

// LocationHierarchy represents a complete location hierarchy from postal code to country
type LocationHierarchy struct {
	PostalCode *domain.PostalCode
	Area       *domain.Area
	City       *domain.City
	Region     *domain.AdministrativeRegion
	Country    *domain.Country

	// IDs for easy reference
	PostalCodeID uint
	AreaID       uint
	CityID       uint
	RegionID     uint
	CountryID    uint
}

// LocationType enum for location types
const (
	LocationTypePostalCode = "postal_code"
	LocationTypeArea       = "area"
	LocationTypeCity       = "city"
	LocationTypeRegion     = "region"
	LocationTypeCountry    = "country"
)

// Location interface for common location properties
type Location interface {
	GetID() uint
	GetType() string
}

// LocationEntries represents a list of locations with their types for serviceability checks
type LocationEntries struct {
	Entries []LocationEntry
}

// LocationEntry holds a specific location type and its ID for serviceability queries
type LocationEntry struct {
	LocationType string
	LocationID   uint
}

// LocationService defines operations for resolving and managing location hierarchies
type LocationService interface {
	// ResolvePostalCode resolves a postal code string to its full location hierarchy
	ResolvePostalCode(postalCode string) (*LocationHierarchy, error)

	// GetLocationHierarchyByID retrieves the location hierarchy for any type of location by its ID
	GetLocationHierarchy(locationType string, locationID uint) (*LocationHierarchy, error)

	// GetLocationEntries returns an ordered list of locations to check for serviceability
	GetLocationEntries(postalCode string) (*LocationEntries, error)
}

// locationService implements the LocationService interface
type locationService struct {
	postalCodeRepo domain.PostalCodeRepository
	areaRepo       domain.AreaRepository
	cityRepo       domain.CityRepository
	regionRepo     domain.RegionRepository
	countryRepo    domain.CountryRepository
	ctx            context.Context
}

// NewLocationService creates a new location service
func NewLocationService(
	postalCodeRepo domain.PostalCodeRepository,
	areaRepo domain.AreaRepository,
	cityRepo domain.CityRepository,
	regionRepo domain.RegionRepository,
	countryRepo domain.CountryRepository,
) LocationService {
	return &locationService{
		postalCodeRepo: postalCodeRepo,
		areaRepo:       areaRepo,
		cityRepo:       cityRepo,
		regionRepo:     regionRepo,
		countryRepo:    countryRepo,
		ctx:            context.Background(),
	}
}

// ResolvePostalCode resolves a postal code string to its full location hierarchy
func (s *locationService) ResolvePostalCode(postalCode string) (*LocationHierarchy, error) {
	// 1. Get postal code
	pc, err := s.postalCodeRepo.GetByCode(s.ctx, postalCode)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve postal code: %w", err)
	}
	if pc == nil {
		return nil, errors.New("postal code not found")
	}

	// Create basic hierarchy with postal code
	hierarchy := &LocationHierarchy{
		PostalCode:   pc,
		PostalCodeID: pc.ID,
	}

	// Try to get area
	area, err := s.areaRepo.GetByID(s.ctx, pc.AreaID)
	if err == nil && area != nil {
		hierarchy.Area = area
		hierarchy.AreaID = area.ID

		// Try to get city
		city, err := s.cityRepo.GetByID(s.ctx, area.CityID)
		if err == nil && city != nil {
			hierarchy.City = city
			hierarchy.CityID = city.ID

			// Try to get region
			region, err := s.regionRepo.GetByID(s.ctx, city.AdministrativeRegionID)
			if err == nil && region != nil {
				hierarchy.Region = region
				hierarchy.RegionID = region.ID

				// Try to get country
				country, err := s.countryRepo.GetByID(s.ctx, region.CountryID)
				if err == nil && country != nil {
					hierarchy.Country = country
					hierarchy.CountryID = country.ID
				}
			}
		}
	}

	return hierarchy, nil
}

// GetLocationHierarchy retrieves the location hierarchy for any type of location by its ID
func (s *locationService) GetLocationHierarchy(locationType string, locationID uint) (*LocationHierarchy, error) {
	switch locationType {
	case LocationTypePostalCode:
		postalCode, err := s.postalCodeRepo.GetByID(s.ctx, locationID)
		if err != nil {
			return nil, err
		}
		return s.ResolvePostalCode(postalCode.Code)

	case LocationTypeArea:
		area, err := s.areaRepo.GetByID(s.ctx, locationID)
		if err != nil {
			return nil, err
		}

		city, err := s.cityRepo.GetByID(s.ctx, area.CityID)
		if err != nil {
			return nil, err
		}

		region, err := s.regionRepo.GetByID(s.ctx, city.AdministrativeRegionID)
		if err != nil {
			return nil, err
		}

		country, err := s.countryRepo.GetByID(s.ctx, region.CountryID)
		if err != nil {
			return nil, err
		}

		return &LocationHierarchy{
			Area:      area,
			City:      city,
			Region:    region,
			Country:   country,
			AreaID:    area.ID,
			CityID:    city.ID,
			RegionID:  region.ID,
			CountryID: region.CountryID,
		}, nil

	case LocationTypeCity:
		city, err := s.cityRepo.GetByID(s.ctx, locationID)
		if err != nil {
			return nil, err
		}

		region, err := s.regionRepo.GetByID(s.ctx, city.AdministrativeRegionID)
		if err != nil {
			return nil, err
		}

		country, err := s.countryRepo.GetByID(s.ctx, region.CountryID)
		if err != nil {
			return nil, err
		}

		return &LocationHierarchy{
			City:      city,
			Region:    region,
			Country:   country,
			CityID:    city.ID,
			RegionID:  region.ID,
			CountryID: region.CountryID,
		}, nil

	case LocationTypeRegion:
		region, err := s.regionRepo.GetByID(s.ctx, locationID)
		if err != nil {
			return nil, err
		}

		country, err := s.countryRepo.GetByID(s.ctx, region.CountryID)
		if err != nil {
			return nil, err
		}

		return &LocationHierarchy{
			Region:    region,
			Country:   country,
			RegionID:  region.ID,
			CountryID: region.CountryID,
		}, nil

	case LocationTypeCountry:
		country, err := s.countryRepo.GetByID(s.ctx, locationID)
		if err != nil {
			return nil, err
		}

		return &LocationHierarchy{
			Country:   country,
			CountryID: country.ID,
		}, nil

	default:
		return nil, fmt.Errorf("invalid location type: %s", locationType)
	}
}

// GetLocationEntries returns an ordered list of locations to check for serviceability
func (s *locationService) GetLocationEntries(postalCode string) (*LocationEntries, error) {
	// Resolve the full hierarchy
	hierarchy, err := s.ResolvePostalCode(postalCode)
	if err != nil {
		return nil, err
	}

	// Create the location entries in order from most specific to most general
	entries := &LocationEntries{
		Entries: []LocationEntry{},
	}

	// Only add entries for which we have valid IDs
	if hierarchy.PostalCodeID > 0 {
		entries.Entries = append(entries.Entries,
			LocationEntry{LocationType: LocationTypePostalCode, LocationID: hierarchy.PostalCodeID})
	}

	if hierarchy.AreaID > 0 {
		entries.Entries = append(entries.Entries,
			LocationEntry{LocationType: LocationTypeArea, LocationID: hierarchy.AreaID})
	}

	if hierarchy.CityID > 0 {
		entries.Entries = append(entries.Entries,
			LocationEntry{LocationType: LocationTypeCity, LocationID: hierarchy.CityID})
	}

	if hierarchy.RegionID > 0 {
		entries.Entries = append(entries.Entries,
			LocationEntry{LocationType: LocationTypeRegion, LocationID: hierarchy.RegionID})
	}

	if hierarchy.CountryID > 0 {
		entries.Entries = append(entries.Entries,
			LocationEntry{LocationType: LocationTypeCountry, LocationID: hierarchy.CountryID})
	}

	return entries, nil
}
