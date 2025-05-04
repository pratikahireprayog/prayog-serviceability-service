package gorm

import (
	"prayog-serviceability-service/internal/domain"
	"prayog-serviceability-service/pkg/database"
)

// RepositoryFactory creates and provides access to all repositories
type RepositoryFactory struct {
	db                            *database.DB
	countryRepository             domain.CountryRepository
	regionRepository              domain.RegionRepository
	cityRepository                domain.CityRepository
	areaRepository                domain.AreaRepository
	postalCodeRepository          domain.PostalCodeRepository
	orderTypeRepository           domain.OrderTypeRepository
	serviceTypeRepository         domain.ServiceTypeRepository
	serviceAvailabilityRepository domain.ServiceAvailabilityRepository
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(db *database.DB) *RepositoryFactory {
	return &RepositoryFactory{
		db: db,
	}
}

// Country returns the country repository
func (f *RepositoryFactory) Country() domain.CountryRepository {
	if f.countryRepository == nil {
		f.countryRepository = NewCountryRepository(f.db)
	}
	return f.countryRepository
}

// Region returns the region repository
func (f *RepositoryFactory) Region() domain.RegionRepository {
	if f.regionRepository == nil {
		f.regionRepository = NewRegionRepository(f.db)
	}
	return f.regionRepository
}

// City returns the city repository
func (f *RepositoryFactory) City() domain.CityRepository {
	if f.cityRepository == nil {
		f.cityRepository = NewCityRepository(f.db)
	}
	return f.cityRepository
}

// Area returns the area repository
func (f *RepositoryFactory) Area() domain.AreaRepository {
	if f.areaRepository == nil {
		f.areaRepository = NewAreaRepository(f.db)
	}
	return f.areaRepository
}

// PostalCode returns the postal code repository
func (f *RepositoryFactory) PostalCode() domain.PostalCodeRepository {
	if f.postalCodeRepository == nil {
		f.postalCodeRepository = NewPostalCodeRepository(f.db)
	}
	return f.postalCodeRepository
}

// OrderType returns the order type repository
func (f *RepositoryFactory) OrderType() domain.OrderTypeRepository {
	if f.orderTypeRepository == nil {
		f.orderTypeRepository = NewOrderTypeRepository(f.db)
	}
	return f.orderTypeRepository
}

// ServiceType returns the service type repository
func (f *RepositoryFactory) ServiceType() domain.ServiceTypeRepository {
	if f.serviceTypeRepository == nil {
		f.serviceTypeRepository = NewServiceTypeRepository(f.db)
	}
	return f.serviceTypeRepository
}

// ServiceAvailability returns the service availability repository
func (f *RepositoryFactory) ServiceAvailability() domain.ServiceAvailabilityRepository {
	if f.serviceAvailabilityRepository == nil {
		f.serviceAvailabilityRepository = NewServiceAvailabilityRepository(f.db)
	}
	return f.serviceAvailabilityRepository
}
