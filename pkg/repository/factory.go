package repository

import (
	database "prayog-serviceability-service/pkg/infrastructure/db"
	gormRepo "prayog-serviceability-service/pkg/repository/gorm"
)

// RepositoryProvider defines an interface for repository providers
// This makes it easier to inject repositories into services
type RepositoryProvider interface {
	CountryRepository() CountryRepository
	RegionRepository() RegionRepository
	CityRepository() CityRepository
	AreaRepository() AreaRepository
	PostalCodeRepository() PostalCodeRepository
	OrderTypeRepository() OrderTypeRepository
	ServiceTypeRepository() ServiceTypeRepository
	ServiceAvailabilityRepository() ServiceAvailabilityRepository
}

// RepositoryFactory provides a factory for creating repositories
type RepositoryFactory struct {
	db *database.DB

	// Repositories instances (lazy initialization)
	countryRepo             CountryRepository
	regionRepo              RegionRepository
	cityRepo                CityRepository
	areaRepo                AreaRepository
	postalCodeRepo          PostalCodeRepository
	orderTypeRepo           OrderTypeRepository
	serviceTypeRepo         ServiceTypeRepository
	serviceAvailabilityRepo ServiceAvailabilityRepository
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(db *database.DB) *RepositoryFactory {
	return &RepositoryFactory{
		db: db,
	}
}

// New creates a new repository factory - convenience function
func New(db *database.DB) RepositoryProvider {
	return NewRepositoryFactory(db)
}

// CountryRepository returns the country repository
func (f *RepositoryFactory) CountryRepository() CountryRepository {
	if f.countryRepo == nil {
		f.countryRepo = gormRepo.NewCountryRepository(f.db)
	}
	return f.countryRepo
}

// RegionRepository returns the region repository
func (f *RepositoryFactory) RegionRepository() RegionRepository {
	if f.regionRepo == nil {
		f.regionRepo = gormRepo.NewRegionRepository(f.db)
	}
	return f.regionRepo
}

// CityRepository returns the city repository
func (f *RepositoryFactory) CityRepository() CityRepository {
	if f.cityRepo == nil {
		f.cityRepo = gormRepo.NewCityRepository(f.db)
	}
	return f.cityRepo
}

// AreaRepository returns the area repository
func (f *RepositoryFactory) AreaRepository() AreaRepository {
	if f.areaRepo == nil {
		f.areaRepo = gormRepo.NewAreaRepository(f.db)
	}
	return f.areaRepo
}

// PostalCodeRepository returns the postal code repository
func (f *RepositoryFactory) PostalCodeRepository() PostalCodeRepository {
	if f.postalCodeRepo == nil {
		f.postalCodeRepo = gormRepo.NewPostalCodeRepository(f.db)
	}
	return f.postalCodeRepo
}

// OrderTypeRepository returns the order type repository
func (f *RepositoryFactory) OrderTypeRepository() OrderTypeRepository {
	if f.orderTypeRepo == nil {
		f.orderTypeRepo = gormRepo.NewOrderTypeRepository(f.db)
	}
	return f.orderTypeRepo
}

// ServiceTypeRepository returns the service type repository
func (f *RepositoryFactory) ServiceTypeRepository() ServiceTypeRepository {
	if f.serviceTypeRepo == nil {
		f.serviceTypeRepo = gormRepo.NewServiceTypeRepository(f.db)
	}
	return f.serviceTypeRepo
}

// ServiceAvailabilityRepository returns the service availability repository
func (f *RepositoryFactory) ServiceAvailabilityRepository() ServiceAvailabilityRepository {
	if f.serviceAvailabilityRepo == nil {
		f.serviceAvailabilityRepo = gormRepo.NewServiceAvailabilityRepository(f.db)
	}
	return f.serviceAvailabilityRepo
}
