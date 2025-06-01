package repositories

import (

	"gorm.io/gorm"
)

// RepositoryFactory provides access to all repositories
type RepositoryFactory struct {
	db                                *gorm.DB
	countryRepository                 CountryRepository
	regionRepository                  RegionRepository
	cityRepository                    CityRepository
	areaRepository                    AreaRepository
	postalCodeRepository              PostalCodeRepository
	postalCodeAliasRepository         PostalCodeAliasRepository
	locationTypeRepository            LocationTypeRepository
	locationAliasRepository           LocationAliasRepository
	partnerLocationCoverageRepository PartnerLocationCoverageRepository
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
	factory.regionRepository = NewRegionRepository(db)
	factory.cityRepository = NewCityRepository(db)
	factory.areaRepository = NewAreaRepository(db)
	factory.postalCodeRepository = NewPostalCodeRepository(db)
	factory.postalCodeAliasRepository = NewPostalCodeAliasRepository(db)
	factory.locationTypeRepository = NewLocationTypeRepository(db)
	factory.locationAliasRepository = NewLocationAliasRepository(db)
	factory.partnerLocationCoverageRepository = NewPartnerLocationCoverageRepository(db)
	factory.locationRepository = NewLocationRepository(factory)
	factory.locationSearchRepository = NewLocationSearchRepository(db)

	return factory
}

// GetCountryRepository returns the country repository
func (f *RepositoryFactory) GetCountryRepository() CountryRepository {
	return f.countryRepository
}

// GetRegionRepository returns the region repository
func (f *RepositoryFactory) GetRegionRepository() RegionRepository {
	return f.regionRepository
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

// GetLocationRepository returns the unified location repository
func (f *RepositoryFactory) GetLocationRepository() LocationRepository {
	return f.locationRepository
}

// GetLocationSearchRepository returns the location search repository
func (f *RepositoryFactory) GetLocationSearchRepository() LocationSearchRepository {
	return f.locationSearchRepository
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

// Regions returns the region repository
func (r *locationRepository) Regions() RegionRepository {
	return r.factory.GetRegionRepository()
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
