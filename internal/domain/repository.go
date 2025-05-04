// Package domain contains the core business entities and interfaces
package domain

import (
	"context"
)

// CountryRepository defines the interface for country data operations
type CountryRepository interface {
	GetByID(ctx context.Context, id uint) (*Country, error)
	GetByCode(ctx context.Context, code string) (*Country, error)
	List(ctx context.Context) ([]*Country, error)
	Create(ctx context.Context, country *Country) error
	Update(ctx context.Context, country *Country) error
	Delete(ctx context.Context, id uint) error
}

// RegionRepository defines the interface for administrative region data operations
type RegionRepository interface {
	GetByID(ctx context.Context, id uint) (*AdministrativeRegion, error)
	GetByCode(ctx context.Context, code string) (*AdministrativeRegion, error)
	GetByCountryID(ctx context.Context, countryID uint) ([]*AdministrativeRegion, error)
	List(ctx context.Context) ([]*AdministrativeRegion, error)
	Create(ctx context.Context, region *AdministrativeRegion) error
	Update(ctx context.Context, region *AdministrativeRegion) error
	Delete(ctx context.Context, id uint) error
}

// CityRepository defines the interface for city data operations
type CityRepository interface {
	GetByID(ctx context.Context, id uint) (*City, error)
	GetByRegionID(ctx context.Context, regionID uint) ([]*City, error)
	List(ctx context.Context) ([]*City, error)
	Create(ctx context.Context, city *City) error
	Update(ctx context.Context, city *City) error
	Delete(ctx context.Context, id uint) error
}

// AreaRepository defines the interface for area data operations
type AreaRepository interface {
	GetByID(ctx context.Context, id uint) (*Area, error)
	GetByCityID(ctx context.Context, cityID uint) ([]*Area, error)
	List(ctx context.Context) ([]*Area, error)
	Create(ctx context.Context, area *Area) error
	Update(ctx context.Context, area *Area) error
	Delete(ctx context.Context, id uint) error
}

// PostalCodeRepository defines the interface for postal code data operations
type PostalCodeRepository interface {
	GetByID(ctx context.Context, id uint) (*PostalCode, error)
	GetByCode(ctx context.Context, code string) (*PostalCode, error)
	GetByAreaID(ctx context.Context, areaID uint) ([]*PostalCode, error)
	List(ctx context.Context) ([]*PostalCode, error)
	Create(ctx context.Context, postalCode *PostalCode) error
	Update(ctx context.Context, postalCode *PostalCode) error
	Delete(ctx context.Context, id uint) error
}

// OrderTypeRepository defines the interface for order type data operations
type OrderTypeRepository interface {
	GetByID(ctx context.Context, id uint) (*OrderType, error)
	GetByCode(ctx context.Context, code string) (*OrderType, error)
	List(ctx context.Context) ([]*OrderType, error)
	Create(ctx context.Context, orderType *OrderType) error
	Update(ctx context.Context, orderType *OrderType) error
	Delete(ctx context.Context, id uint) error
}

// ServiceTypeRepository defines the interface for service type data operations
type ServiceTypeRepository interface {
	GetByID(ctx context.Context, id uint) (*ServiceType, error)
	GetByCode(ctx context.Context, code string) (*ServiceType, error)
	List(ctx context.Context) ([]*ServiceType, error)
	Create(ctx context.Context, serviceType *ServiceType) error
	Update(ctx context.Context, serviceType *ServiceType) error
	Delete(ctx context.Context, id uint) error
}

// ServiceAvailabilityRepository defines the interface for service availability data operations
type ServiceAvailabilityRepository interface {
	GetByID(ctx context.Context, id uint) (*ServiceAvailability, error)
	GetByLocation(ctx context.Context, locationType string, locationID uint) ([]*ServiceAvailability, error)
	GetByService(ctx context.Context, serviceTypeID uint) ([]*ServiceAvailability, error)
	GetByOrderType(ctx context.Context, orderTypeID uint) ([]*ServiceAvailability, error)
	CheckAvailability(ctx context.Context, postalCode string, orderTypeCode string, serviceTypeCode string) (bool, error)
	BulkCheckAvailability(ctx context.Context, requests []ServiceabilityRequest) ([]ServiceabilityResult, error)
	List(ctx context.Context) ([]*ServiceAvailability, error)
	Create(ctx context.Context, serviceAvailability *ServiceAvailability) error
	Update(ctx context.Context, serviceAvailability *ServiceAvailability) error
	Delete(ctx context.Context, id uint) error
}

// ServiceabilityRequest represents a request to check service availability
type ServiceabilityRequest struct {
	PostalCode      string `json:"postal_code"`
	CountryCode     string `json:"country_code"`
	OrderTypeCode   string `json:"order_type_code"`
	ServiceTypeCode string `json:"service_type_code,omitempty"` // Optional, if not provided, all services are checked
}

// ServiceabilityResult represents the result of a serviceability check
type ServiceabilityResult struct {
	PostalCode    string             `json:"postal_code"`
	CountryCode   string             `json:"country_code"`
	OrderTypeCode string             `json:"order_type_code"`
	Services      []ServiceAvailable `json:"services"`
	IsServiceable bool               `json:"is_serviceable"`
}

// ServiceAvailable represents the availability of a specific service
type ServiceAvailable struct {
	ServiceTypeCode string `json:"service_type_code"`
	ServiceTypeName string `json:"service_type_name"`
	IsAvailable     bool   `json:"is_available"`
}
