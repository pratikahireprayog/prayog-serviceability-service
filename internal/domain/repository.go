// Package domain contains the core business entities and interfaces
package domain

import (
	"context"

	"github.com/google/uuid"
)

// CountryRepository defines the interface for country data operations
type CountryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (Country, error)
	GetByCode(ctx context.Context, code string) (Country, error)
	List(ctx context.Context) ([]Country, error)
	Create(ctx context.Context, country Country) error
	Update(ctx context.Context, country Country) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RegionRepository defines the interface for administrative region data operations
type RegionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (AdministrativeRegion, error)
	GetByCode(ctx context.Context, code string) (AdministrativeRegion, error)
	GetByCountryID(ctx context.Context, countryID uuid.UUID) ([]AdministrativeRegion, error)
	List(ctx context.Context) ([]AdministrativeRegion, error)
	Create(ctx context.Context, region AdministrativeRegion) error
	Update(ctx context.Context, region AdministrativeRegion) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CityRepository defines the interface for city data operations
type CityRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (City, error)
	GetByRegionID(ctx context.Context, regionID uuid.UUID) ([]City, error)
	List(ctx context.Context) ([]City, error)
	Create(ctx context.Context, city City) error
	Update(ctx context.Context, city City) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// AreaRepository defines the interface for area data operations
type AreaRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (Area, error)
	GetByCityID(ctx context.Context, cityID uuid.UUID) ([]Area, error)
	List(ctx context.Context) ([]Area, error)
	Create(ctx context.Context, area Area) error
	Update(ctx context.Context, area Area) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PostalCodeRepository defines the interface for postal code data operations
type PostalCodeRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (PostalCode, error)
	GetByCode(ctx context.Context, code string) (PostalCode, error)
	GetByAreaID(ctx context.Context, areaID uuid.UUID) ([]PostalCode, error)
	List(ctx context.Context) ([]PostalCode, error)
	Create(ctx context.Context, postalCode PostalCode) error
	Update(ctx context.Context, postalCode PostalCode) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// OrderTypeRepository defines the interface for order type data operations
type OrderTypeRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (OrderType, error)
	GetByCode(ctx context.Context, code string) (OrderType, error)
	List(ctx context.Context) ([]OrderType, error)
	Create(ctx context.Context, orderType OrderType) error
	Update(ctx context.Context, orderType OrderType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ServiceTypeRepository defines the interface for service type data operations
type ServiceTypeRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (ServiceType, error)
	GetByCode(ctx context.Context, code string) (ServiceType, error)
	List(ctx context.Context) ([]ServiceType, error)
	Create(ctx context.Context, serviceType ServiceType) error
	Update(ctx context.Context, serviceType ServiceType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ServiceAvailabilityRepository defines the interface for service availability data operations
type ServiceAvailabilityRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (ServiceAvailability, error)

	// Find service availability based on location, order type, and service type
	FindByLocation(ctx context.Context, locationType string, locationID uuid.UUID, orderTypeID, serviceTypeID uuid.UUID) (ServiceAvailability, error)

	// Check if a service is available for a specific location
	CheckAvailability(ctx context.Context, postalCode string, serviceTypeCode, orderTypeCode string) (bool, error)

	// Get all service availabilities for a specific location
	GetByLocation(ctx context.Context, locationType string, locationID uuid.UUID) ([]ServiceAvailability, error)

	// Get all service availabilities for a specific order type
	GetByOrderType(ctx context.Context, orderTypeID uuid.UUID) ([]ServiceAvailability, error)

	// Get all service availabilities for a specific service type
	GetByServiceType(ctx context.Context, serviceTypeID uuid.UUID) ([]ServiceAvailability, error)

	// List all service availabilities
	List(ctx context.Context) ([]ServiceAvailability, error)

	// Create a new service availability
	Create(ctx context.Context, serviceAvailability ServiceAvailability) error

	// Update a service availability
	Update(ctx context.Context, serviceAvailability ServiceAvailability) error

	// Delete a service availability
	Delete(ctx context.Context, id uuid.UUID) error

	// Bulk check availability for multiple requests
	BulkCheckAvailability(ctx context.Context, requests []APIServiceabilityRequest) ([]APIServiceabilityResponse, error)
}

// RepositoryFactory creates and provides repositories
type RepositoryFactory interface {
	CountryRepository() CountryRepository
	RegionRepository() RegionRepository
	CityRepository() CityRepository
	AreaRepository() AreaRepository
	PostalCodeRepository() PostalCodeRepository
	OrderTypeRepository() OrderTypeRepository
	ServiceTypeRepository() ServiceTypeRepository
	ServiceAvailabilityRepository() ServiceAvailabilityRepository
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
