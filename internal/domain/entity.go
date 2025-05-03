// Package domain contains the core business entities and interfaces
package domain

import (
	"time"
)

// Country represents a country entity
type Country struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AdministrativeRegion represents a state/province/region within a country
type AdministrativeRegion struct {
	ID        uint      `json:"id"`
	CountryID uint      `json:"country_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// City represents a city entity
type City struct {
	ID                     uint      `json:"id"`
	AdministrativeRegionID uint      `json:"administrative_region_id"`
	Name                   string    `json:"name"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// Area represents a specific area or locality within a city
type Area struct {
	ID        uint      `json:"id"`
	CityID    uint      `json:"city_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PostalCode represents a postal/zip code
type PostalCode struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	AreaID    uint      `json:"area_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderType represents different types of orders (e.g., delivery, pickup)
type OrderType struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ServiceType represents different types of services offered
type ServiceType struct {
	ID          uint      `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ServiceAvailability represents the availability of a service in a specific location
type ServiceAvailability struct {
	ID             uint      `json:"id"`
	LocationType   string    `json:"location_type"` // Can be 'country', 'region', 'city', 'area', or 'postal_code'
	LocationID     uint      `json:"location_id"`
	OrderTypeID    uint      `json:"order_type_id"`
	ServiceTypeID  uint      `json:"service_type_id"`
	IsAvailable    bool      `json:"is_available"`
	EffectiveFrom  time.Time `json:"effective_from"`
	EffectiveTo    time.Time `json:"effective_to"`
	AdditionalData string    `json:"additional_data,omitempty"` // JSON string for flexible additional properties
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
