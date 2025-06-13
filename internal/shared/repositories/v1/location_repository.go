package repositories

import (
	"context"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// CountryRepository defines the interface for country operations
type CountryRepository interface {
	GetByID(ctx context.Context, id string) (*models.Country, error)
	GetByIDWithDeleted(ctx context.Context, id string) (*models.Country, error)
	GetByCode(ctx context.Context, code string) (*models.Country, error)
	GetByCodeWithDeleted(ctx context.Context, code string) (*models.Country, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.Country, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.Country, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.Country, int64, error)
	Create(ctx context.Context, country *models.Country) error
	Update(ctx context.Context, id string, country *models.Country) error
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
}

// RegionTypeRepository defines the interface for region type operations
type RegionTypeRepository interface {
	GetByCode(ctx context.Context, code string) (*models.RegionType, error)
	GetByCodeWithDeleted(ctx context.Context, code string) (*models.RegionType, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.RegionType, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.RegionType, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.RegionType, int64, error)
	Create(ctx context.Context, regionType *models.RegionType) error
	Update(ctx context.Context, code string, regionType *models.RegionType) error
	Delete(ctx context.Context, code string) error
	Restore(ctx context.Context, code string) error
	ForceDelete(ctx context.Context, code string) error
}

// RegionRepository defines the interface for region operations
type RegionRepository interface {
	GetByID(ctx context.Context, id string) (*models.Region, error)
	GetByIDWithDeleted(ctx context.Context, id string) (*models.Region, error)
	GetByCode(ctx context.Context, code string) (*models.Region, error)
	GetByCodeWithDeleted(ctx context.Context, code string) (*models.Region, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.Region, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.Region, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.Region, int64, error)
	GetByCountryID(ctx context.Context, countryID string) ([]models.Region, error)
	GetByCountryIDWithDeleted(ctx context.Context, countryID string) ([]models.Region, error)
	Create(ctx context.Context, region *models.Region) error
	Update(ctx context.Context, id string, region *models.Region) error
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
}

// DistrictRepository defines the interface for district operations
type DistrictRepository interface {
	GetByID(ctx context.Context, id string) (*models.District, error)
	GetByIDWithDeleted(ctx context.Context, id string) (*models.District, error)
	GetByCode(ctx context.Context, code string) (*models.District, error)
	GetByCodeWithDeleted(ctx context.Context, code string) (*models.District, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.District, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.District, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.District, int64, error)
	GetByRegionID(ctx context.Context, regionID string) ([]models.District, error)
	GetByRegionIDWithDeleted(ctx context.Context, regionID string) ([]models.District, error)
	Create(ctx context.Context, district *models.District) error
	Update(ctx context.Context, id string, district *models.District) error
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
}

// CityRepository defines the interface for city operations
type CityRepository interface {
	GetByID(ctx context.Context, id string) (*models.City, error)
	GetByIDWithDeleted(ctx context.Context, id string) (*models.City, error)
	GetByCode(ctx context.Context, code string) (*models.City, error)
	GetByCodeWithDeleted(ctx context.Context, code string) (*models.City, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.City, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.City, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.City, int64, error)
	GetByRegionID(ctx context.Context, regionID string) ([]models.City, error)
	GetByRegionIDWithDeleted(ctx context.Context, regionID string) ([]models.City, error)
	Create(ctx context.Context, city *models.City) error
	Update(ctx context.Context, id string, city *models.City) error
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
}

// AreaRepository defines the interface for area operations
type AreaRepository interface {
	GetByID(ctx context.Context, id string) (*models.Area, error)
	GetByIDWithDeleted(ctx context.Context, id string) (*models.Area, error)
	GetByCode(ctx context.Context, code string) (*models.Area, error)
	GetByCodeWithDeleted(ctx context.Context, code string) (*models.Area, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.Area, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.Area, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.Area, int64, error)
	GetByCityID(ctx context.Context, cityID string) ([]models.Area, error)
	GetByCityIDWithDeleted(ctx context.Context, cityID string) ([]models.Area, error)
	Create(ctx context.Context, area *models.Area) error
	Update(ctx context.Context, id string, area *models.Area) error
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
}

// PostalCodeRepository defines the interface for postal code operations
type PostalCodeRepository interface {
	GetByID(ctx context.Context, id string) (*models.PostalCode, error)
	GetByIDWithDeleted(ctx context.Context, id string) (*models.PostalCode, error)
	GetByCode(ctx context.Context, code string) (*models.PostalCode, error)
	GetByCodeWithDeleted(ctx context.Context, code string) (*models.PostalCode, error)
	GetByCodeAndCountry(ctx context.Context, code, countryCode string) (*models.PostalCode, error)
	GetByCodeAndCountryWithDeleted(ctx context.Context, code, countryCode string) (*models.PostalCode, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.PostalCode, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.PostalCode, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.PostalCode, int64, error)
	GetByLocation(ctx context.Context, countryCode, regionCode, cityCode, areaCode string) ([]models.PostalCode, error)
	GetByLocationWithDeleted(ctx context.Context, countryCode, regionCode, cityCode, areaCode string) ([]models.PostalCode, error)
	GetByAreaID(ctx context.Context, areaID string) ([]models.PostalCode, error)
	GetByAreaIDWithDeleted(ctx context.Context, areaID string) ([]models.PostalCode, error)
	GetLocationHierarchy(ctx context.Context, postalCode, countryCode string) (*models.LocationHierarchy, error)
	ValidatePostalCode(ctx context.Context, postalCode, countryCode string) (*models.LocationValidationResult, error)
	IsActive(ctx context.Context, postalCode, countryCode string) (bool, error)
	Create(ctx context.Context, postalCode *models.PostalCode) error
	Update(ctx context.Context, postalCode *models.PostalCode) error
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
	BulkCreate(ctx context.Context, postalCodes []models.PostalCode) error
}

// PostalCodeAliasRepository defines the interface for postal code alias operations
type PostalCodeAliasRepository interface {
	GetByAliasCode(ctx context.Context, aliasCode string) (*models.PostalCodeAlias, error)
	GetByAliasCodeWithDeleted(ctx context.Context, aliasCode string) (*models.PostalCodeAlias, error)
	GetByPostalCodeID(ctx context.Context, postalCodeID uint) ([]models.PostalCodeAlias, error)
	GetByPostalCodeIDWithDeleted(ctx context.Context, postalCodeID uint) ([]models.PostalCodeAlias, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.PostalCodeAlias, int64, error)
	Create(ctx context.Context, alias *models.PostalCodeAlias) error
	Update(ctx context.Context, alias *models.PostalCodeAlias) error
	Delete(ctx context.Context, id uint) error
	Restore(ctx context.Context, id uint) error
	ForceDelete(ctx context.Context, id uint) error
}

// LocationTypeRepository defines the interface for location type operations
type LocationTypeRepository interface {
	GetByCode(ctx context.Context, code string) (*models.LocationType, error)
	GetByCodeWithDeleted(ctx context.Context, code string) (*models.LocationType, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.LocationType, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.LocationType, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.LocationType, int64, error)
	Create(ctx context.Context, locationType *models.LocationType) error
	Update(ctx context.Context, locationType *models.LocationType) error
	Delete(ctx context.Context, code string) error
	Restore(ctx context.Context, code string) error
	ForceDelete(ctx context.Context, code string) error
}

// LocationAliasRepository defines the interface for location alias operations
type LocationAliasRepository interface {
	GetByID(ctx context.Context, id string) (*models.LocationAlias, error)
	GetByIDWithDeleted(ctx context.Context, id string) (*models.LocationAlias, error)
	GetByEntityTypeAndID(ctx context.Context, entityType string, entityID string) ([]models.LocationAlias, error)
	GetByEntityTypeAndIDWithDeleted(ctx context.Context, entityType string, entityID string) ([]models.LocationAlias, error)
	GetByAliasName(ctx context.Context, aliasName string) (*models.LocationAlias, error)
	GetByAliasNameWithDeleted(ctx context.Context, aliasName string) (*models.LocationAlias, error)
	GetPrimaryAlias(ctx context.Context, entityType string, entityID string) (*models.LocationAlias, error)
	GetPrimaryAliasWithDeleted(ctx context.Context, entityType string, entityID string) (*models.LocationAlias, error)
	GetAllByEntityID(ctx context.Context, entityID string) ([]models.LocationAlias, error)
	GetAllByEntityIDWithDeleted(ctx context.Context, entityID string) ([]models.LocationAlias, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.LocationAlias, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.LocationAlias, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.LocationAlias, int64, error)
	Create(ctx context.Context, alias *models.LocationAlias) error
	Update(ctx context.Context, alias *models.LocationAlias) error
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
}

// PartnerLocationCoverageRepository defines the interface for partner location coverage operations
type PartnerLocationCoverageRepository interface {
	GetByPartnerID(ctx context.Context, partnerID string) ([]models.PartnerLocationCoverage, error)
	GetByPartnerIDWithDeleted(ctx context.Context, partnerID string) ([]models.PartnerLocationCoverage, error)
	GetByPartnerCode(ctx context.Context, partnerCode string) ([]models.PartnerLocationCoverage, error)
	GetByLocationScopeAndID(ctx context.Context, locationScope string, locationID uint) ([]models.PartnerLocationCoverage, error)
	GetByLocationScopeAndIDWithDeleted(ctx context.Context, locationScope string, locationID uint) ([]models.PartnerLocationCoverage, error)
	GetByFilters(ctx context.Context, filters *models.PartnerLocationCoverageFilters) ([]models.PartnerLocationCoverageResult, error)
	GetByFiltersWithDeleted(ctx context.Context, filters *models.PartnerLocationCoverageFilters) ([]models.PartnerLocationCoverageResult, error)
	GetPartnersForLocation(ctx context.Context, locationScope string, locationID uint, zoneTypes []string) ([]models.PartnerLocationCoverageResult, error)
	GetPartnersForLocationWithDeleted(ctx context.Context, locationScope string, locationID uint, zoneTypes []string) ([]models.PartnerLocationCoverageResult, error)
	GetCoverageForPartner(ctx context.Context, partnerID string, isActive bool) ([]models.PartnerLocationCoverageResult, error)
	GetCoverageForPartnerWithDeleted(ctx context.Context, partnerID string, isActive bool) ([]models.PartnerLocationCoverageResult, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.PartnerLocationCoverage, int64, error)
	Create(ctx context.Context, coverage *models.PartnerLocationCoverage) error
	Update(ctx context.Context, coverage *models.PartnerLocationCoverage) error
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
	BulkCreate(ctx context.Context, coverages []models.PartnerLocationCoverage) error
}

// LocationRepository provides a unified interface for location operations
type LocationRepository interface {
	Countries() CountryRepository
	RegionTypes() RegionTypeRepository
	Regions() RegionRepository
	Districts() DistrictRepository
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
