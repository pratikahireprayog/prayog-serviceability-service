package repository

import (
	"context"
	"prayog-serviceability-service/pkg/domain"

	"github.com/google/uuid"
)


// Basic CRUD operations
type Repository[T any] interface {
	GetByID(ctx context.Context, id uuid.UUID) (T, error)
	List(ctx context.Context) ([]T, error)
	Create(ctx context.Context, entity T) error
	Update(ctx context.Context, entity T) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// Only define specific interfaces when needed
type CountryFinder interface {
	GetByCode(ctx context.Context, code string) (domain.Country, error)
}

// Service only depends on what it needs
type CountryService struct {
	repo interface {
		Repository[domain.Country]
		CountryFinder
	}
}

// CountryRepository defines the contract for country repository operations
type CountryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.Country, error)
	GetByCode(ctx context.Context, code string) (domain.Country, error)
	List(ctx context.Context) ([]domain.Country, error)
	Create(ctx context.Context, country domain.Country) error
	Update(ctx context.Context, country domain.Country) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RegionRepository defines the contract for region repository operations
type RegionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.AdministrativeRegion, error)
	GetByCode(ctx context.Context, code string) (domain.AdministrativeRegion, error)
	GetByCountryID(ctx context.Context, countryID uuid.UUID) ([]domain.AdministrativeRegion, error)
	List(ctx context.Context) ([]domain.AdministrativeRegion, error)
	Create(ctx context.Context, region domain.AdministrativeRegion) error
	Update(ctx context.Context, region domain.AdministrativeRegion) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CityRepository defines the contract for city repository operations
type CityRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.City, error)
	GetByRegionID(ctx context.Context, regionID uuid.UUID) ([]domain.City, error)
	List(ctx context.Context) ([]domain.City, error)
	Create(ctx context.Context, city domain.City) error
	Update(ctx context.Context, city domain.City) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// AreaRepository defines the contract for area repository operations
type AreaRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.Area, error)
	GetByCityID(ctx context.Context, cityID uuid.UUID) ([]domain.Area, error)
	List(ctx context.Context) ([]domain.Area, error)
	Create(ctx context.Context, area domain.Area) error
	Update(ctx context.Context, area domain.Area) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PostalCodeRepository defines the contract for postal code repository operations
type PostalCodeRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.PostalCode, error)
	GetByCode(ctx context.Context, code string) (domain.PostalCode, error)
	GetByAreaID(ctx context.Context, areaID uuid.UUID) ([]domain.PostalCode, error)
	List(ctx context.Context) ([]domain.PostalCode, error)
	Create(ctx context.Context, postalCode domain.PostalCode) error
	Update(ctx context.Context, postalCode domain.PostalCode) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// OrderTypeRepository defines the contract for order type repository operations
type OrderTypeRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.OrderType, error)
	GetByCode(ctx context.Context, code string) (domain.OrderType, error)
	List(ctx context.Context) ([]domain.OrderType, error)
	Create(ctx context.Context, orderType domain.OrderType) error
	Update(ctx context.Context, orderType domain.OrderType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ServiceTypeRepository defines the contract for service type repository operations
type ServiceTypeRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.ServiceType, error)
	GetByCode(ctx context.Context, code string) (domain.ServiceType, error)
	List(ctx context.Context) ([]domain.ServiceType, error)
	Create(ctx context.Context, serviceType domain.ServiceType) error
	Update(ctx context.Context, serviceType domain.ServiceType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ServiceAvailabilityRepository defines the contract for service availability repository operations
type ServiceAvailabilityRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.ServiceAvailability, error)
	GetByLocationAndTypes(ctx context.Context, locationType string, locationID uuid.UUID, orderTypeID, serviceTypeID uuid.UUID) ([]domain.ServiceAvailability, error)
	List(ctx context.Context) ([]domain.ServiceAvailability, error)
	Create(ctx context.Context, serviceAvailability domain.ServiceAvailability) error
	Update(ctx context.Context, serviceAvailability domain.ServiceAvailability) error
	Delete(ctx context.Context, id uuid.UUID) error
}
