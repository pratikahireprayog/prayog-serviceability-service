// Package domain contains the core business entities and interfaces
package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Point represents a geographical point (latitude, longitude)
type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Value implements the driver.Valuer interface for Point
func (p Point) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements the sql.Scanner interface for Point
func (p *Point) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &p)
}

// Constraints represents JSON constraints for service availability
type Constraints map[string]interface{}

// Value implements the driver.Valuer interface for Constraints
func (c Constraints) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for Constraints
func (c *Constraints) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &c)
}

// Region type constants
const (
	RegionTypeState     = "STATE"
	RegionTypeProvince  = "PROVINCE"
	RegionTypeCounty    = "COUNTY"
	RegionTypeRegion    = "REGION"
	RegionTypeTerritory = "TERRITORY"
)

// Area type constants
const (
	AreaTypeNeighborhood = "NEIGHBORHOOD"
	AreaTypeLocality     = "LOCALITY"
	AreaTypeWard         = "WARD"
	AreaTypeZone         = "ZONE"
)

// LocationType enum for service availability
const (
	LocationTypeCountry    = "COUNTRY"
	LocationTypeRegion     = "REGION"
	LocationTypeCity       = "CITY"
	LocationTypeArea       = "AREA"
	LocationTypePostalCode = "POSTAL_CODE"
)

// Country represents a country entity
type Country struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	Code         string         `gorm:"size:10;uniqueIndex;not null" json:"code"`
	CurrencyCode string         `gorm:"size:3;index" json:"currency_code"`
	PhoneCode    string         `gorm:"size:10;index" json:"phone_code"`
	IsActive     bool           `gorm:"default:true;not null" json:"is_active"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	Aliases      []CountryAlias `gorm:"foreignKey:CountryID" json:"aliases,omitempty"`
}

// CountryAlias represents an alternative name for a country in various languages
type CountryAlias struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CountryID    uuid.UUID `gorm:"type:uuid;index;not null" json:"country_id"`
	Country      Country   `gorm:"foreignKey:CountryID;constraint:OnDelete:CASCADE" json:"-"`
	AliasName    string    `gorm:"size:100;not null" json:"alias_name"`
	LanguageCode string    `gorm:"size:10;index;not null" json:"language_code"`
	IsPrimary    bool      `gorm:"default:false;not null" json:"is_primary"`
	IsActive     bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// AdministrativeRegion represents a state/province/county/region within a country
type AdministrativeRegion struct {
	ID         uuid.UUID                   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CountryID  uuid.UUID                   `gorm:"type:uuid;index;not null" json:"country_id"`
	Country    Country                     `gorm:"foreignKey:CountryID;constraint:OnDelete:CASCADE" json:"-"`
	Name       string                      `gorm:"size:100;not null" json:"name"`
	Code       string                      `gorm:"size:20;index;not null" json:"code"`
	RegionType string                      `gorm:"size:20;index;comment:'STATE|PROVINCE|COUNTY|REGION|TERRITORY'" json:"region_type"`
	IsActive   bool                        `gorm:"default:true;not null" json:"is_active"`
	CreatedAt  time.Time                   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time                   `gorm:"autoUpdateTime" json:"updated_at"`
	Aliases    []AdministrativeRegionAlias `gorm:"foreignKey:RegionID" json:"aliases,omitempty"`
}

// AdministrativeRegionAlias represents an alternative name for a region in various languages
type AdministrativeRegionAlias struct {
	ID           uuid.UUID            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RegionID     uuid.UUID            `gorm:"type:uuid;index;not null" json:"region_id"`
	Region       AdministrativeRegion `gorm:"foreignKey:RegionID;constraint:OnDelete:CASCADE" json:"-"`
	AliasName    string               `gorm:"size:100;not null" json:"alias_name"`
	LanguageCode string               `gorm:"size:10;index;not null" json:"language_code"`
	IsPrimary    bool                 `gorm:"default:false;not null" json:"is_primary"`
	IsActive     bool                 `gorm:"default:true;not null" json:"is_active"`
	CreatedAt    time.Time            `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time            `gorm:"autoUpdateTime" json:"updated_at"`
}

// City represents a city entity
type City struct {
	ID        uuid.UUID            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CountryID uuid.UUID            `gorm:"type:uuid;index;comment:'For countries without regions'" json:"country_id"`
	Country   Country              `gorm:"foreignKey:CountryID" json:"-"`
	RegionID  uuid.UUID            `gorm:"type:uuid;index;comment:'Optional - for hierarchical countries'" json:"region_id"`
	Region    AdministrativeRegion `gorm:"foreignKey:RegionID" json:"-"`
	Name      string               `gorm:"size:100;not null" json:"name"`
	Code      string               `gorm:"size:20;index" json:"code"`
	IsActive  bool                 `gorm:"default:true;not null" json:"is_active"`
	CreatedAt time.Time            `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time            `gorm:"autoUpdateTime" json:"updated_at"`
	Aliases   []CityAlias          `gorm:"foreignKey:CityID" json:"aliases,omitempty"`
}

// CityAlias represents an alternative name for a city in various languages
type CityAlias struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CityID       uuid.UUID `gorm:"type:uuid;index;not null" json:"city_id"`
	City         City      `gorm:"foreignKey:CityID;constraint:OnDelete:CASCADE" json:"-"`
	AliasName    string    `gorm:"size:100;not null" json:"alias_name"`
	LanguageCode string    `gorm:"size:10;index;not null" json:"language_code"`
	IsPrimary    bool      `gorm:"default:false;not null" json:"is_primary"`
	IsActive     bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Area represents a specific area or locality within a city
type Area struct {
	ID          uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CityID      uuid.UUID   `gorm:"type:uuid;index;not null" json:"city_id"`
	City        City        `gorm:"foreignKey:CityID;constraint:OnDelete:CASCADE" json:"-"`
	Name        string      `gorm:"size:100;not null" json:"name"`
	AreaType    string      `gorm:"size:20;index;comment:'NEIGHBORHOOD|LOCALITY|WARD|ZONE'" json:"area_type"`
	GeoLocation Point       `gorm:"type:jsonb" json:"geo_location"`
	IsActive    bool        `gorm:"default:true;not null" json:"is_active"`
	CreatedAt   time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
	Aliases     []AreaAlias `gorm:"foreignKey:AreaID" json:"aliases,omitempty"`
}

// AreaAlias represents an alternative name for an area in various languages
type AreaAlias struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AreaID       uuid.UUID `gorm:"type:uuid;index;not null" json:"area_id"`
	Area         Area      `gorm:"foreignKey:AreaID;constraint:OnDelete:CASCADE" json:"-"`
	AliasName    string    `gorm:"size:100;not null" json:"alias_name"`
	LanguageCode string    `gorm:"size:10;index;not null" json:"language_code"`
	IsPrimary    bool      `gorm:"default:false;not null" json:"is_primary"`
	IsActive     bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// PostalCode represents a postal/zip code
type PostalCode struct {
	ID          uuid.UUID            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CountryID   uuid.UUID            `gorm:"type:uuid;index;not null;comment:'Always required'" json:"country_id"`
	Country     Country              `gorm:"foreignKey:CountryID" json:"-"`
	RegionID    uuid.UUID            `gorm:"type:uuid;index;comment:'Optional'" json:"region_id"`
	Region      AdministrativeRegion `gorm:"foreignKey:RegionID" json:"-"`
	CityID      uuid.UUID            `gorm:"type:uuid;index;comment:'Optional'" json:"city_id"`
	City        City                 `gorm:"foreignKey:CityID" json:"-"`
	AreaID      uuid.UUID            `gorm:"type:uuid;index;comment:'Optional'" json:"area_id"`
	Area        Area                 `gorm:"foreignKey:AreaID" json:"-"`
	Code        string               `gorm:"size:20;index;not null" json:"code"`
	GeoLocation Point                `gorm:"type:jsonb" json:"geo_location"`
	IsActive    bool                 `gorm:"default:true;not null" json:"is_active"`
	CreatedAt   time.Time            `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time            `gorm:"autoUpdateTime" json:"updated_at"`
}

// OrderType represents a type of order
type OrderType struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Code        string    `gorm:"size:20;uniqueIndex;not null" json:"code"`
	Description string    `gorm:"size:500" json:"description"`
	IsActive    bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// ServiceType represents a type of service offered
type ServiceType struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Code        string    `gorm:"size:20;uniqueIndex;not null" json:"code"`
	SLAHours    int       `gorm:"not null" json:"sla_hours"`
	Description string    `gorm:"size:500" json:"description"`
	IsActive    bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// ServiceAvailability represents the availability of a service in a specific location
type ServiceAvailability struct {
	ID            uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LocationType  string      `gorm:"size:20;index;not null;comment:'COUNTRY|REGION|CITY|AREA|POSTAL_CODE'" json:"location_type"`
	LocationID    uuid.UUID   `gorm:"type:uuid;index;not null" json:"location_id"`
	OrderTypeID   uuid.UUID   `gorm:"type:uuid;index;not null" json:"order_type_id"`
	OrderType     OrderType   `gorm:"foreignKey:OrderTypeID;constraint:OnDelete:CASCADE" json:"-"`
	ServiceTypeID uuid.UUID   `gorm:"type:uuid;index;not null" json:"service_type_id"`
	ServiceType   ServiceType `gorm:"foreignKey:ServiceTypeID;constraint:OnDelete:CASCADE" json:"-"`
	IsAvailable   bool        `gorm:"default:false;not null" json:"is_available"`
	Constraints   Constraints `gorm:"type:jsonb" json:"constraints"`
	EffectiveFrom time.Time   `gorm:"not null;index" json:"effective_from"`
	EffectiveTo   time.Time   `gorm:"not null;index" json:"effective_to"`
	IsActive      bool        `gorm:"default:true;not null" json:"is_active"`
	CreatedAt     time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
}

// ServiceabilityRule represents a rule defining whether a service is available for a specific location and order type
// NOTE: This is a legacy compatibility type - prefer ServiceAvailability for new code
type ServiceabilityRule struct {
	ID            uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PostalCodeID  uuid.UUID   `gorm:"type:uuid;index;not null" json:"postal_code_id"`
	PostalCode    PostalCode  `gorm:"foreignKey:PostalCodeID" json:"-"`
	ServiceTypeID uuid.UUID   `gorm:"type:uuid;index;not null" json:"service_type_id"`
	ServiceType   ServiceType `gorm:"foreignKey:ServiceTypeID" json:"-"`
	OrderTypeID   uuid.UUID   `gorm:"type:uuid;index;not null" json:"order_type_id"`
	OrderType     OrderType   `gorm:"foreignKey:OrderTypeID" json:"-"`
	IsServiceable bool        `gorm:"default:false;not null" json:"is_serviceable"`
	EffectiveFrom time.Time   `gorm:"not null;index" json:"effective_from"`
	EffectiveTo   time.Time   `gorm:"not null;index" json:"effective_to"`
	CreatedAt     time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
}

// TimeRule represents a time-based rule for service availability
type TimeRule struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LocationID uuid.UUID `gorm:"type:uuid;index;not null" json:"location_id"`
	StartTime  time.Time `gorm:"not null;index" json:"start_time"`
	EndTime    time.Time `gorm:"not null;index" json:"end_time"`
	Available  bool      `gorm:"default:false;not null" json:"available"`
	Message    string    `gorm:"size:255" json:"message"`
	Priority   int       `gorm:"default:100;not null" json:"priority"`
	IsActive   bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
