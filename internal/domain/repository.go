package domain

import (
	"context"
	"time"
)

// CountryRepository defines the operations for Country entities
type CountryRepository interface {
	Create(ctx context.Context, country *Country) error
	GetByID(ctx context.Context, id uint) (*Country, error)
	GetByCode(ctx context.Context, code string) (*Country, error)
	GetAll(ctx context.Context) ([]*Country, error)
	Update(ctx context.Context, country *Country) error
	Delete(ctx context.Context, id uint) error
}

// AdministrativeRegionRepository defines the operations for AdministrativeRegion entities
type AdministrativeRegionRepository interface {
	Create(ctx context.Context, region *AdministrativeRegion) error
	GetByID(ctx context.Context, id uint) (*AdministrativeRegion, error)
	GetByCode(ctx context.Context, code string) (*AdministrativeRegion, error)
	GetByCountryID(ctx context.Context, countryID uint) ([]*AdministrativeRegion, error)
	GetAll(ctx context.Context) ([]*AdministrativeRegion, error)
	Update(ctx context.Context, region *AdministrativeRegion) error
	Delete(ctx context.Context, id uint) error
}

// CityRepository defines the operations for City entities
type CityRepository interface {
	Create(ctx context.Context, city *City) error
	GetByID(ctx context.Context, id uint) (*City, error)
	GetByRegionID(ctx context.Context, regionID uint) ([]*City, error)
	GetAll(ctx context.Context) ([]*City, error)
	Update(ctx context.Context, city *City) error
	Delete(ctx context.Context, id uint) error
}

// AreaRepository defines the operations for Area entities
type AreaRepository interface {
	Create(ctx context.Context, area *Area) error
	GetByID(ctx context.Context, id uint) (*Area, error)
	GetByCityID(ctx context.Context, cityID uint) ([]*Area, error)
	GetAll(ctx context.Context) ([]*Area, error)
	Update(ctx context.Context, area *Area) error
	Delete(ctx context.Context, id uint) error
}

// PostalCodeRepository defines the operations for PostalCode entities
type PostalCodeRepository interface {
	Create(ctx context.Context, postalCode *PostalCode) error
	GetByID(ctx context.Context, id uint) (*PostalCode, error)
	GetByCode(ctx context.Context, code string) (*PostalCode, error)
	GetByAreaID(ctx context.Context, areaID uint) ([]*PostalCode, error)
	GetAll(ctx context.Context) ([]*PostalCode, error)
	Update(ctx context.Context, postalCode *PostalCode) error
	Delete(ctx context.Context, id uint) error
}

// OrderTypeRepository defines the operations for OrderType entities
type OrderTypeRepository interface {
	Create(ctx context.Context, orderType *OrderType) error
	GetByID(ctx context.Context, id uint) (*OrderType, error)
	GetByCode(ctx context.Context, code string) (*OrderType, error)
	GetAll(ctx context.Context) ([]*OrderType, error)
	Update(ctx context.Context, orderType *OrderType) error
	Delete(ctx context.Context, id uint) error
}

// ServiceTypeRepository defines the operations for ServiceType entities
type ServiceTypeRepository interface {
	Create(ctx context.Context, serviceType *ServiceType) error
	GetByID(ctx context.Context, id uint) (*ServiceType, error)
	GetByCode(ctx context.Context, code string) (*ServiceType, error)
	GetAll(ctx context.Context) ([]*ServiceType, error)
	Update(ctx context.Context, serviceType *ServiceType) error
	Delete(ctx context.Context, id uint) error
}

// ServiceAvailabilityRepository defines the operations for ServiceAvailability entities
type ServiceAvailabilityRepository interface {
	Create(ctx context.Context, availability *ServiceAvailability) error
	GetByID(ctx context.Context, id uint) (*ServiceAvailability, error)

	// Find service availability based on location, order type, and service type
	FindByLocation(ctx context.Context, locationType string, locationID uint, orderTypeID, serviceTypeID uint) (*ServiceAvailability, error)

	// Get all service availabilities for a specific location
	GetByLocation(ctx context.Context, locationType string, locationID uint) ([]*ServiceAvailability, error)

	// Get all service availabilities for a specific order type
	GetByOrderType(ctx context.Context, orderTypeID uint) ([]*ServiceAvailability, error)

	// Get all service availabilities for a specific service type
	GetByServiceType(ctx context.Context, serviceTypeID uint) ([]*ServiceAvailability, error)

	// Update a service availability record
	Update(ctx context.Context, availability *ServiceAvailability) error

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
