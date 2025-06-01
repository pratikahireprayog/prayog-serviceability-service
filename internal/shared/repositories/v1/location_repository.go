package repositories

import (
	"context"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// CountryRepository defines the interface for country operations
type CountryRepository interface {
	GetByCode(ctx context.Context, code string) (*models.Country, error)
	GetAll(ctx context.Context) ([]models.Country, error)
	Create(ctx context.Context, country *models.Country) error
	Update(ctx context.Context, country *models.Country) error
	Delete(ctx context.Context, id uint) error
}

// RegionRepository defines the interface for region operations
type RegionRepository interface {
	GetByID(ctx context.Context, id uint) (*models.Region, error)
	GetByCountryID(ctx context.Context, countryID uint) ([]models.Region, error)
	GetByCode(ctx context.Context, code string, countryID uint) (*models.Region, error)
	Create(ctx context.Context, region *models.Region) error
	Update(ctx context.Context, region *models.Region) error
	Delete(ctx context.Context, id uint) error
}

// CityRepository defines the interface for city operations
type CityRepository interface {
	GetByID(ctx context.Context, id uint) (*models.City, error)
	GetByRegionID(ctx context.Context, regionID uint) ([]models.City, error)
	GetByCode(ctx context.Context, code string, regionID uint) (*models.City, error)
	Create(ctx context.Context, city *models.City) error
	Update(ctx context.Context, city *models.City) error
	Delete(ctx context.Context, id uint) error
}

// AreaRepository defines the interface for area operations
type AreaRepository interface {
	GetByID(ctx context.Context, id uint) (*models.Area, error)
	GetByCityID(ctx context.Context, cityID uint) ([]models.Area, error)
	GetByCode(ctx context.Context, code string, cityID uint) (*models.Area, error)
	Create(ctx context.Context, area *models.Area) error
	Update(ctx context.Context, area *models.Area) error
	Delete(ctx context.Context, id uint) error
}

// PostalCodeRepository defines the interface for postal code operations
type PostalCodeRepository interface {
	GetByCode(ctx context.Context, code string) (*models.PostalCode, error)
	GetByCodeAndCountry(ctx context.Context, code, countryCode string) (*models.PostalCode, error)
	GetByAreaID(ctx context.Context, areaID uint) ([]models.PostalCode, error)
	GetLocationHierarchy(ctx context.Context, postalCode, countryCode string) (*models.LocationHierarchy, error)
	ValidatePostalCode(ctx context.Context, postalCode, countryCode string) (*models.LocationValidationResult, error)
	IsActive(ctx context.Context, postalCode, countryCode string) (bool, error)
	Create(ctx context.Context, postalCode *models.PostalCode) error
	Update(ctx context.Context, postalCode *models.PostalCode) error
	Delete(ctx context.Context, id uint) error
	BulkCreate(ctx context.Context, postalCodes []models.PostalCode) error
}

// PostalCodeAliasRepository defines the interface for postal code alias operations
type PostalCodeAliasRepository interface {
	GetByAliasCode(ctx context.Context, aliasCode string) (*models.PostalCodeAlias, error)
	GetByPostalCodeID(ctx context.Context, postalCodeID uint) ([]models.PostalCodeAlias, error)
	Create(ctx context.Context, alias *models.PostalCodeAlias) error
	Update(ctx context.Context, alias *models.PostalCodeAlias) error
	Delete(ctx context.Context, id uint) error
}

// LocationTypeRepository defines the interface for location type operations
type LocationTypeRepository interface {
	GetByCode(ctx context.Context, code string) (*models.LocationType, error)
	GetAll(ctx context.Context) ([]models.LocationType, error)
	Create(ctx context.Context, locationType *models.LocationType) error
	Update(ctx context.Context, locationType *models.LocationType) error
	Delete(ctx context.Context, code string) error
}

// LocationAliasRepository defines the interface for location alias operations
type LocationAliasRepository interface {
	GetByID(ctx context.Context, id uint) (*models.LocationAlias, error)
	GetByEntityTypeAndID(ctx context.Context, entityType string, entityID uint) ([]models.LocationAlias, error)
	GetByAliasName(ctx context.Context, aliasName string) (*models.LocationAlias, error)
	GetPrimaryAlias(ctx context.Context, entityType string, entityID uint) (*models.LocationAlias, error)
	Create(ctx context.Context, alias *models.LocationAlias) error
	Update(ctx context.Context, alias *models.LocationAlias) error
	Delete(ctx context.Context, id uint) error
}

// PartnerLocationCoverageRepository defines the interface for partner location coverage operations
type PartnerLocationCoverageRepository interface {
	GetByPartnerID(ctx context.Context, partnerID uint) ([]models.PartnerLocationCoverage, error)
	GetByLocationScopeAndID(ctx context.Context, locationScope string, locationID uint) ([]models.PartnerLocationCoverage, error)
	GetByFilters(ctx context.Context, filters *models.PartnerLocationCoverageFilters) ([]models.PartnerLocationCoverageResult, error)
	GetPartnersForLocation(ctx context.Context, locationScope string, locationID uint, zoneTypes []string) ([]models.PartnerLocationCoverageResult, error)
	GetCoverageForPartner(ctx context.Context, partnerID uint, isActive bool) ([]models.PartnerLocationCoverageResult, error)
	Create(ctx context.Context, coverage *models.PartnerLocationCoverage) error
	Update(ctx context.Context, coverage *models.PartnerLocationCoverage) error
	Delete(ctx context.Context, id uint) error
	BulkCreate(ctx context.Context, coverages []models.PartnerLocationCoverage) error
}

// LocationRepository provides a unified interface for location operations
type LocationRepository interface {
	Countries() CountryRepository
	Regions() RegionRepository
	Cities() CityRepository
	Areas() AreaRepository
	PostalCodes() PostalCodeRepository
	PostalCodeAliases() PostalCodeAliasRepository
	LocationTypes() LocationTypeRepository
	LocationAliases() LocationAliasRepository
	PartnerLocationCoverages() PartnerLocationCoverageRepository
}

// LocationSearchFilters represents search filters for location queries
type LocationSearchFilters struct {
	CountryCode string `json:"country_code,omitempty"`
	RegionCode  string `json:"region_code,omitempty"`
	CityCode    string `json:"city_code,omitempty"`
	AreaCode    string `json:"area_code,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// LocationSearchResult represents the result of a location search
type LocationSearchResult struct {
	Countries   []models.Country    `json:"countries,omitempty"`
	Regions     []models.Region     `json:"regions,omitempty"`
	Cities      []models.City       `json:"cities,omitempty"`
	Areas       []models.Area       `json:"areas,omitempty"`
	PostalCodes []models.PostalCode `json:"postal_codes,omitempty"`
	TotalCount  int64               `json:"total_count"`
}

// LocationSearchRepository defines advanced search operations
type LocationSearchRepository interface {
	SearchLocations(ctx context.Context, filters *LocationSearchFilters) (*LocationSearchResult, error)
	GetLocationsByPostalCodePattern(ctx context.Context, pattern, countryCode string, limit int) ([]models.LocationHierarchy, error)
	GetNearbyPostalCodes(ctx context.Context, latitude, longitude float64, radiusKm int, limit int) ([]models.PostalCode, error)
}
