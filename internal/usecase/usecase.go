// Package usecase contains application business rules and use cases
package usecase

import (
	"github.com/prayog/serviceability/internal/domain"
)

// GeoService defines the interface for geography-related operations
type GeoService interface {
	// Country operations
	GetCountryByID(id uint) (*domain.Country, error)
	GetCountryByCode(code string) (*domain.Country, error)
	ListCountries() ([]*domain.Country, error)
	CreateCountry(country *domain.Country) error
	UpdateCountry(country *domain.Country) error
	DeleteCountry(id uint) error

	// Region operations
	GetRegionByID(id uint) (*domain.AdministrativeRegion, error)
	GetRegionByCode(code string) (*domain.AdministrativeRegion, error)
	GetRegionsByCountryID(countryID uint) ([]*domain.AdministrativeRegion, error)
	ListRegions() ([]*domain.AdministrativeRegion, error)
	CreateRegion(region *domain.AdministrativeRegion) error
	UpdateRegion(region *domain.AdministrativeRegion) error
	DeleteRegion(id uint) error

	// City operations
	GetCityByID(id uint) (*domain.City, error)
	GetCitiesByRegionID(regionID uint) ([]*domain.City, error)
	ListCities() ([]*domain.City, error)
	CreateCity(city *domain.City) error
	UpdateCity(city *domain.City) error
	DeleteCity(id uint) error

	// Area operations
	GetAreaByID(id uint) (*domain.Area, error)
	GetAreasByCityID(cityID uint) ([]*domain.Area, error)
	ListAreas() ([]*domain.Area, error)
	CreateArea(area *domain.Area) error
	UpdateArea(area *domain.Area) error
	DeleteArea(id uint) error

	// Postal code operations
	GetPostalCodeByID(id uint) (*domain.PostalCode, error)
	GetPostalCodeByCode(code string) (*domain.PostalCode, error)
	GetPostalCodesByAreaID(areaID uint) ([]*domain.PostalCode, error)
	ListPostalCodes() ([]*domain.PostalCode, error)
	CreatePostalCode(postalCode *domain.PostalCode) error
	UpdatePostalCode(postalCode *domain.PostalCode) error
	DeletePostalCode(id uint) error
}

// OrderTypeService defines the interface for order type operations
type OrderTypeService interface {
	GetByID(id uint) (*domain.OrderType, error)
	GetByCode(code string) (*domain.OrderType, error)
	List() ([]*domain.OrderType, error)
	Create(orderType *domain.OrderType) error
	Update(orderType *domain.OrderType) error
	Delete(id uint) error
}

// ServiceTypeService defines the interface for service type operations
type ServiceTypeService interface {
	GetByID(id uint) (*domain.ServiceType, error)
	GetByCode(code string) (*domain.ServiceType, error)
	List() ([]*domain.ServiceType, error)
	Create(serviceType *domain.ServiceType) error
	Update(serviceType *domain.ServiceType) error
	Delete(id uint) error
}

// ServiceabilityService defines the interface for serviceability operations
type ServiceabilityService interface {
	CheckServiceability(request domain.ServiceabilityRequest) (*domain.ServiceabilityResult, error)
	BulkCheckServiceability(requests []domain.ServiceabilityRequest) ([]domain.ServiceabilityResult, error)

	// Service availability management
	GetServiceAvailabilityByID(id uint) (*domain.ServiceAvailability, error)
	GetServiceAvailabilitiesByLocation(locationType string, locationID uint) ([]*domain.ServiceAvailability, error)
	GetServiceAvailabilitiesByService(serviceTypeID uint) ([]*domain.ServiceAvailability, error)
	GetServiceAvailabilitiesByOrderType(orderTypeID uint) ([]*domain.ServiceAvailability, error)
	ListServiceAvailabilities() ([]*domain.ServiceAvailability, error)
	CreateServiceAvailability(serviceAvailability *domain.ServiceAvailability) error
	UpdateServiceAvailability(serviceAvailability *domain.ServiceAvailability) error
	DeleteServiceAvailability(id uint) error
}

// UseCaseFactory is a factory for creating use case services
type UseCaseFactory struct {
	geoService            GeoService
	orderTypeService      OrderTypeService
	serviceTypeService    ServiceTypeService
	serviceabilityService ServiceabilityService
}

// NewUseCaseFactory creates a new use case factory
func NewUseCaseFactory(
	geoService GeoService,
	orderTypeService OrderTypeService,
	serviceTypeService ServiceTypeService,
	serviceabilityService ServiceabilityService,
) *UseCaseFactory {
	return &UseCaseFactory{
		geoService:            geoService,
		orderTypeService:      orderTypeService,
		serviceTypeService:    serviceTypeService,
		serviceabilityService: serviceabilityService,
	}
}

// GeoService returns the geo service implementation
func (f *UseCaseFactory) GeoService() GeoService {
	return f.geoService
}

// OrderTypeService returns the order type service implementation
func (f *UseCaseFactory) OrderTypeService() OrderTypeService {
	return f.orderTypeService
}

// ServiceTypeService returns the service type service implementation
func (f *UseCaseFactory) ServiceTypeService() ServiceTypeService {
	return f.serviceTypeService
}

// ServiceabilityService returns the serviceability service implementation
func (f *UseCaseFactory) ServiceabilityService() ServiceabilityService {
	return f.serviceabilityService
}
