package gorm

import (
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"gorm.io/gorm"
)

// RepositoryFactory provides access to all repositories
type RepositoryFactory struct {
	db                                *gorm.DB
	countryRepository                 repositories.CountryRepository
	regionRepository                  repositories.RegionRepository
	cityRepository                    repositories.CityRepository
	areaRepository                    repositories.AreaRepository
	postalCodeRepository              repositories.PostalCodeRepository
	postalCodeAliasRepository         repositories.PostalCodeAliasRepository
	locationTypeRepository            repositories.LocationTypeRepository
	locationAliasRepository           repositories.LocationAliasRepository
	partnerLocationCoverageRepository repositories.PartnerLocationCoverageRepository
	locationRepository                repositories.LocationRepository
	locationSearchRepository          repositories.LocationSearchRepository
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
func (f *RepositoryFactory) GetCountryRepository() repositories.CountryRepository {
	return f.countryRepository
}

// GetRegionRepository returns the region repository
func (f *RepositoryFactory) GetRegionRepository() repositories.RegionRepository {
	return f.regionRepository
}

// GetCityRepository returns the city repository
func (f *RepositoryFactory) GetCityRepository() repositories.CityRepository {
	return f.cityRepository
}

// GetAreaRepository returns the area repository
func (f *RepositoryFactory) GetAreaRepository() repositories.AreaRepository {
	return f.areaRepository
}

// GetPostalCodeRepository returns the postal code repository
func (f *RepositoryFactory) GetPostalCodeRepository() repositories.PostalCodeRepository {
	return f.postalCodeRepository
}

// GetPostalCodeAliasRepository returns the postal code alias repository
func (f *RepositoryFactory) GetPostalCodeAliasRepository() repositories.PostalCodeAliasRepository {
	return f.postalCodeAliasRepository
}

// GetLocationTypeRepository returns the location type repository
func (f *RepositoryFactory) GetLocationTypeRepository() repositories.LocationTypeRepository {
	return f.locationTypeRepository
}

// GetLocationAliasRepository returns the location alias repository
func (f *RepositoryFactory) GetLocationAliasRepository() repositories.LocationAliasRepository {
	return f.locationAliasRepository
}

// GetPartnerLocationCoverageRepository returns the partner location coverage repository
func (f *RepositoryFactory) GetPartnerLocationCoverageRepository() repositories.PartnerLocationCoverageRepository {
	return f.partnerLocationCoverageRepository
}

// GetLocationRepository returns the unified location repository
func (f *RepositoryFactory) GetLocationRepository() repositories.LocationRepository {
	return f.locationRepository
}

// GetLocationSearchRepository returns the location search repository
func (f *RepositoryFactory) GetLocationSearchRepository() repositories.LocationSearchRepository {
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
func NewLocationRepository(factory *RepositoryFactory) repositories.LocationRepository {
	return &locationRepository{factory: factory}
}

// Countries returns the country repository
func (r *locationRepository) Countries() repositories.CountryRepository {
	return r.factory.GetCountryRepository()
}

// Regions returns the region repository
func (r *locationRepository) Regions() repositories.RegionRepository {
	return r.factory.GetRegionRepository()
}

// Cities returns the city repository
func (r *locationRepository) Cities() repositories.CityRepository {
	return r.factory.GetCityRepository()
}

// Areas returns the area repository
func (r *locationRepository) Areas() repositories.AreaRepository {
	return r.factory.GetAreaRepository()
}

// PostalCodes returns the postal code repository
func (r *locationRepository) PostalCodes() repositories.PostalCodeRepository {
	return r.factory.GetPostalCodeRepository()
}

// PostalCodeAliases returns the postal code alias repository
func (r *locationRepository) PostalCodeAliases() repositories.PostalCodeAliasRepository {
	return r.factory.GetPostalCodeAliasRepository()
}

// LocationTypes returns the location type repository
func (r *locationRepository) LocationTypes() repositories.LocationTypeRepository {
	return r.factory.GetLocationTypeRepository()
}

// LocationAliases returns the location alias repository
func (r *locationRepository) LocationAliases() repositories.LocationAliasRepository {
	return r.factory.GetLocationAliasRepository()
}

// PartnerLocationCoverages returns the partner location coverage repository
func (r *locationRepository) PartnerLocationCoverages() repositories.PartnerLocationCoverageRepository {
	return r.factory.GetPartnerLocationCoverageRepository()
}
