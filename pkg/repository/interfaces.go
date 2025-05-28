package repository

import (
	"context"
	"prayog-serviceability-service/pkg/domain"

	"github.com/google/uuid"
)

// Repository defines generic CRUD operations for any entity type
type Repository[T any] interface {
	GetByID(ctx context.Context, id uuid.UUID) (T, error)
	List(ctx context.Context) ([]T, error)
	Create(ctx context.Context, entity T) error
	Update(ctx context.Context, entity T) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CountryRepository defines operations for country entity
type CountryRepository interface {
	Repository[domain.Country]
	GetByCode(ctx context.Context, code string) (domain.Country, error)
}

// RegionRepository defines operations for region entity
type RegionRepository interface {
	Repository[domain.AdministrativeRegion]
	GetByCode(ctx context.Context, code string) (domain.AdministrativeRegion, error)
	GetByCountryID(ctx context.Context, countryID uuid.UUID) ([]domain.AdministrativeRegion, error)
}

// CityRepository defines operations for city entity
type CityRepository interface {
	Repository[domain.City]
	GetByRegionID(ctx context.Context, regionID uuid.UUID) ([]domain.City, error)
}

// AreaRepository defines operations for area entity
type AreaRepository interface {
	Repository[domain.Area]
	GetByCityID(ctx context.Context, cityID uuid.UUID) ([]domain.Area, error)
}

// PostalCodeRepository defines operations for postal code entity
type PostalCodeRepository interface {
	Repository[domain.PostalCode]
	GetByCode(ctx context.Context, code string) (domain.PostalCode, error)
	GetByAreaID(ctx context.Context, areaID uuid.UUID) ([]domain.PostalCode, error)
}

// OrderTypeRepository defines operations for order type entity
type OrderTypeRepository interface {
	Repository[domain.OrderType]
	GetByCode(ctx context.Context, code string) (domain.OrderType, error)
}

// ServiceTypeRepository defines operations for service type entity
type ServiceTypeRepository interface {
	Repository[domain.ServiceType]
	GetByCode(ctx context.Context, code string) (domain.ServiceType, error)
}

// ServiceAvailabilityRepository defines operations for service availability entity
type ServiceAvailabilityRepository interface {
	Repository[domain.ServiceAvailability]
	GetByLocationAndTypes(ctx context.Context, locationType string, locationID uuid.UUID, orderTypeID, serviceTypeID uuid.UUID) ([]domain.ServiceAvailability, error)
}
