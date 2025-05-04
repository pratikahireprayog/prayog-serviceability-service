package usecase

import (
	"github.com/prayog/serviceability/internal/domain"
)

// RepositoryFactory defines an interface for creating repositories
type RepositoryFactory interface {
	Country() domain.CountryRepository
	Region() domain.RegionRepository
	City() domain.CityRepository
	Area() domain.AreaRepository
	PostalCode() domain.PostalCodeRepository
	OrderType() domain.OrderTypeRepository
	ServiceType() domain.ServiceTypeRepository
	ServiceAvailability() domain.ServiceAvailabilityRepository
}

// Factory creates all the use case services
type Factory struct {
	repoFactory RepositoryFactory
}

// NewFactory creates a new use case factory
func NewFactory(repoFactory RepositoryFactory) *UseCaseFactory {
	// Create the geo service
	geoService := NewGeoService(
		repoFactory.Country(),
		repoFactory.Region(),
		repoFactory.City(),
		repoFactory.Area(),
		repoFactory.PostalCode(),
	)

	// Create the order type service
	orderTypeService := NewOrderTypeService(
		repoFactory.OrderType(),
	)

	// Create the service type service
	serviceTypeService := NewServiceTypeService(
		repoFactory.ServiceType(),
	)

	// Create the serviceability service
	serviceabilityService := NewServiceabilityService(
		repoFactory.ServiceAvailability(),
		repoFactory.PostalCode(),
		repoFactory.ServiceType(),
		repoFactory.OrderType(),
		repoFactory.Country(),
	)

	// Return the use case factory with all services
	return NewUseCaseFactory(
		geoService,
		orderTypeService,
		serviceTypeService,
		serviceabilityService,
	)
}
