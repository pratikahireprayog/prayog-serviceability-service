package repositories

import (
	"gorm.io/gorm"
)

// RepositoryFactory provides access to all repositories
type RepositoryFactory struct {
	db                                *gorm.DB
	countryRepository                 CountryRepository
	regionTypeRepository              RegionTypeRepository
	regionRepository                  RegionRepository
	districtRepository                DistrictRepository
	cityRepository                    CityRepository
	areaRepository                    AreaRepository
	postalCodeRepository              PostalCodeRepository
	postalCodeAliasRepository         PostalCodeAliasRepository
	locationTypeRepository            LocationTypeRepository
	locationAliasRepository           LocationAliasRepository
	partnerLocationCoverageRepository PartnerLocationCoverageRepository
	attributeCategoryRepository       AttributeCategoryRepository
	attributeRepository               AttributeRepository
	partnerAttributeMapRepository     PartnerAttributeMapRepository
	geoLocationRepository             GeoLocationRepository
	nearestHubLocationRepository      NearestHubLocationRepository
	locationRepository                LocationRepository
	locationSearchRepository          LocationSearchRepository
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(db *gorm.DB) *RepositoryFactory {
	factory := &RepositoryFactory{
		db: db,
	}

	// Initialize all repositories
	factory.countryRepository = NewCountryRepository(db)
	factory.regionTypeRepository = NewRegionTypeRepository(db)
	factory.regionRepository = NewRegionRepository(db)
	factory.districtRepository = NewDistrictRepository(db)
	factory.cityRepository = NewCityRepository(db)
	factory.areaRepository = NewAreaRepository(db)
	factory.postalCodeRepository = NewPostalCodeRepository(db)
	factory.postalCodeAliasRepository = NewPostalCodeAliasRepository(db)
	factory.locationTypeRepository = NewLocationTypeRepository(db)
	factory.locationAliasRepository = NewLocationAliasRepository(db)
	factory.partnerLocationCoverageRepository = NewPartnerLocationCoverageRepository(db)
	factory.attributeCategoryRepository = NewAttributeCategoryRepository(db)
	factory.attributeRepository = NewAttributeRepository(db)
	factory.partnerAttributeMapRepository = NewPartnerAttributeMapRepository(db)
	factory.geoLocationRepository = NewGeoLocationRepository(db)
	factory.nearestHubLocationRepository = NewNearestHubLocationRepository(db)
	factory.locationRepository = NewLocationRepository(factory)
	factory.locationSearchRepository = NewLocationSearchRepository(db)

	return factory
}

// GetCountryRepository returns the country repository
func (f *RepositoryFactory) GetCountryRepository() CountryRepository {
	return f.countryRepository
}

// GetRegionTypeRepository returns the region type repository
func (f *RepositoryFactory) GetRegionTypeRepository() RegionTypeRepository {
	return f.regionTypeRepository
}

// GetRegionRepository returns the region repository
func (f *RepositoryFactory) GetRegionRepository() RegionRepository {
	return f.regionRepository
}

// GetDistrictRepository returns the district repository
func (f *RepositoryFactory) GetDistrictRepository() DistrictRepository {
	return f.districtRepository
}

// GetCityRepository returns the city repository
func (f *RepositoryFactory) GetCityRepository() CityRepository {
	return f.cityRepository
}

// GetAreaRepository returns the area repository
func (f *RepositoryFactory) GetAreaRepository() AreaRepository {
	return f.areaRepository
}

// GetPostalCodeRepository returns the postal code repository
func (f *RepositoryFactory) GetPostalCodeRepository() PostalCodeRepository {
	return f.postalCodeRepository
}

// GetPostalCodeAliasRepository returns the postal code alias repository
func (f *RepositoryFactory) GetPostalCodeAliasRepository() PostalCodeAliasRepository {
	return f.postalCodeAliasRepository
}

// GetLocationTypeRepository returns the location type repository
func (f *RepositoryFactory) GetLocationTypeRepository() LocationTypeRepository {
	return f.locationTypeRepository
}

// GetLocationAliasRepository returns the location alias repository
func (f *RepositoryFactory) GetLocationAliasRepository() LocationAliasRepository {
	return f.locationAliasRepository
}

// GetPartnerLocationCoverageRepository returns the partner location coverage repository
func (f *RepositoryFactory) GetPartnerLocationCoverageRepository() PartnerLocationCoverageRepository {
	return f.partnerLocationCoverageRepository
}

// GetGeoLocationRepository returns the geo location repository
func (f *RepositoryFactory) GetGeoLocationRepository() GeoLocationRepository {
	return f.geoLocationRepository
}

// GetNearestHubLocationRepository returns the nearest hub location repository
func (f *RepositoryFactory) GetNearestHubLocationRepository() NearestHubLocationRepository {
	return f.nearestHubLocationRepository
}

// GetLocationRepository returns the unified location repository
func (f *RepositoryFactory) GetLocationRepository() LocationRepository {
	return f.locationRepository
}

// GetLocationSearchRepository returns the location search repository
func (f *RepositoryFactory) GetLocationSearchRepository() LocationSearchRepository {
	return f.locationSearchRepository
}

// GetAttributeCategoryRepository returns the attribute category repository
func (f *RepositoryFactory) GetAttributeCategoryRepository() AttributeCategoryRepository {
	return f.attributeCategoryRepository
}

// GetAttributeRepository returns the attribute repository
func (f *RepositoryFactory) GetAttributeRepository() AttributeRepository {
	return f.attributeRepository
}

// GetPartnerAttributeMapRepository returns the partner attribute map repository
func (f *RepositoryFactory) GetPartnerAttributeMapRepository() PartnerAttributeMapRepository {
	return f.partnerAttributeMapRepository
}

// DB returns the underlying database connection
func (f *RepositoryFactory) DB() *gorm.DB {
	return f.db
}

// locationRepository implements the unified LocationRepository interface
type locationRepository struct {
	factory *RepositoryFactory
}

// NewLocationRepository creates a new unified location repository
func NewLocationRepository(factory *RepositoryFactory) LocationRepository {
	return &locationRepository{factory: factory}
}

// Countries returns the country repository
func (r *locationRepository) Countries() CountryRepository {
	return r.factory.GetCountryRepository()
}

// RegionTypes returns the region type repository
func (r *locationRepository) RegionTypes() RegionTypeRepository {
	return r.factory.GetRegionTypeRepository()
}

// Regions returns the region repository
func (r *locationRepository) Regions() RegionRepository {
	return r.factory.GetRegionRepository()
}

// Districts returns the district repository
func (r *locationRepository) Districts() DistrictRepository {
	return r.factory.GetDistrictRepository()
}

// Cities returns the city repository
func (r *locationRepository) Cities() CityRepository {
	return r.factory.GetCityRepository()
}

// Areas returns the area repository
func (r *locationRepository) Areas() AreaRepository {
	return r.factory.GetAreaRepository()
}

// PostalCodes returns the postal code repository
func (r *locationRepository) PostalCodes() PostalCodeRepository {
	return r.factory.GetPostalCodeRepository()
}

// PostalCodeAliases returns the postal code alias repository
func (r *locationRepository) PostalCodeAliases() PostalCodeAliasRepository {
	return r.factory.GetPostalCodeAliasRepository()
}

// LocationTypes returns the location type repository
func (r *locationRepository) LocationTypes() LocationTypeRepository {
	return r.factory.GetLocationTypeRepository()
}

// LocationAliases returns the location alias repository
func (r *locationRepository) LocationAliases() LocationAliasRepository {
	return r.factory.GetLocationAliasRepository()
}

// PartnerLocationCoverages returns the partner location coverage repository
func (r *locationRepository) PartnerLocationCoverages() PartnerLocationCoverageRepository {
	return r.factory.GetPartnerLocationCoverageRepository()
}
