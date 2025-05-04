package repository

import (
	"context"
	"time"

	"github.com/prayog/serviceability/internal/domain"
)

// CountryRepository defines the operations for Country entities
type CountryRepository interface {
	Create(ctx context.Context, country *domain.Country) error
	GetByID(ctx context.Context, id uint) (*domain.Country, error)
	GetByCode(ctx context.Context, code string) (*domain.Country, error)
	GetAll(ctx context.Context) ([]*domain.Country, error)
	Update(ctx context.Context, country *domain.Country) error
	Delete(ctx context.Context, id uint) error
}

// AdministrativeRegionRepository defines the operations for AdministrativeRegion entities
type AdministrativeRegionRepository interface {
	Create(ctx context.Context, region *domain.AdministrativeRegion) error
	GetByID(ctx context.Context, id uint) (*domain.AdministrativeRegion, error)
	GetByCode(ctx context.Context, code string) (*domain.AdministrativeRegion, error)
	GetByCountryID(ctx context.Context, countryID uint) ([]*domain.AdministrativeRegion, error)
	GetAll(ctx context.Context) ([]*domain.AdministrativeRegion, error)
	Update(ctx context.Context, region *domain.AdministrativeRegion) error
	Delete(ctx context.Context, id uint) error
}

// CityRepository defines the operations for City entities
type CityRepository interface {
	Create(ctx context.Context, city *domain.City) error
	GetByID(ctx context.Context, id uint) (*domain.City, error)
	GetByRegionID(ctx context.Context, regionID uint) ([]*domain.City, error)
	GetAll(ctx context.Context) ([]*domain.City, error)
	Update(ctx context.Context, city *domain.City) error
	Delete(ctx context.Context, id uint) error
}

// AreaRepository defines the operations for Area entities
type AreaRepository interface {
	Create(ctx context.Context, area *domain.Area) error
	GetByID(ctx context.Context, id uint) (*domain.Area, error)
	GetByCityID(ctx context.Context, cityID uint) ([]*domain.Area, error)
	GetAll(ctx context.Context) ([]*domain.Area, error)
	Update(ctx context.Context, area *domain.Area) error
	Delete(ctx context.Context, id uint) error
}

// PostalCodeRepository defines the operations for PostalCode entities
type PostalCodeRepository interface {
	Create(ctx context.Context, postalCode *domain.PostalCode) error
	GetByID(ctx context.Context, id uint) (*domain.PostalCode, error)
	GetByCode(ctx context.Context, code string) (*domain.PostalCode, error)
	GetByAreaID(ctx context.Context, areaID uint) ([]*domain.PostalCode, error)
	GetAll(ctx context.Context) ([]*domain.PostalCode, error)
	Update(ctx context.Context, postalCode *domain.PostalCode) error
	Delete(ctx context.Context, id uint) error
}

// OrderTypeRepository defines the operations for OrderType entities
type OrderTypeRepository interface {
	Create(ctx context.Context, orderType *domain.OrderType) error
	GetByID(ctx context.Context, id uint) (*domain.OrderType, error)
	GetByCode(ctx context.Context, code string) (*domain.OrderType, error)
	GetAll(ctx context.Context) ([]*domain.OrderType, error)
	Update(ctx context.Context, orderType *domain.OrderType) error
	Delete(ctx context.Context, id uint) error
}

// ServiceTypeRepository defines the operations for ServiceType entities
type ServiceTypeRepository interface {
	Create(ctx context.Context, serviceType *domain.ServiceType) error
	GetByID(ctx context.Context, id uint) (*domain.ServiceType, error)
	GetByCode(ctx context.Context, code string) (*domain.ServiceType, error)
	GetAll(ctx context.Context) ([]*domain.ServiceType, error)
	Update(ctx context.Context, serviceType *domain.ServiceType) error
	Delete(ctx context.Context, id uint) error
}

// ServiceAvailabilityRepository defines the operations for ServiceAvailability entities
type ServiceAvailabilityRepository interface {
	Create(ctx context.Context, availability *domain.ServiceAvailability) error
	GetByID(ctx context.Context, id uint) (*domain.ServiceAvailability, error)

	// Find service availability based on location, order type, and service type
	FindByLocation(ctx context.Context, locationType string, locationID uint, orderTypeID, serviceTypeID uint) (*domain.ServiceAvailability, error)

	// Get all service availabilities for a specific location
	GetByLocation(ctx context.Context, locationType string, locationID uint) ([]*domain.ServiceAvailability, error)

	// Get all service availabilities for a specific order type
	GetByOrderType(ctx context.Context, orderTypeID uint) ([]*domain.ServiceAvailability, error)

	// Get all service availabilities for a specific service type
	GetByServiceType(ctx context.Context, serviceTypeID uint) ([]*domain.ServiceAvailability, error)

	// Update a service availability record
	Update(ctx context.Context, availability *domain.ServiceAvailability) error

	// Delete a service availability record
	Delete(ctx context.Context, id uint) error

	// Bulk check service availability for multiple locations, order types, and service types
	BulkCheck(ctx context.Context, requests []ServiceAvailabilityRequest) ([]ServiceAvailabilityResponse, error)
}

// ServiceAvailabilityRequest represents a request to check service availability
type ServiceAvailabilityRequest struct {
	LocationType  string `json:"location_type"`
	LocationID    uint   `json:"location_id"`
	OrderTypeID   uint   `json:"order_type_id"`
	ServiceTypeID uint   `json:"service_type_id"`
}

// ServiceAvailabilityResponse represents the response to a service availability check
type ServiceAvailabilityResponse struct {
	Request        ServiceAvailabilityRequest `json:"request"`
	IsAvailable    bool                       `json:"is_available"`
	EffectiveFrom  time.Time                  `json:"effective_from,omitempty"`
	EffectiveTo    time.Time                  `json:"effective_to,omitempty"`
	AdditionalData string                     `json:"additional_data,omitempty"`
}
