package models

import (
	"time"
)

// Country represents a country in the location hierarchy
type Country struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Code      string    `json:"code" gorm:"uniqueIndex;not null;size:2"` // ISO 2-letter code
	Name      string    `json:"name" gorm:"not null;size:100"`
	PhoneCode string    `json:"phone_code" gorm:"size:10"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Regions []Region `json:"regions,omitempty" gorm:"foreignKey:CountryID"`
}

// Region represents a state/province in the location hierarchy
type Region struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CountryID uint      `json:"country_id" gorm:"not null;index"`
	Code      string    `json:"code" gorm:"not null;size:10;index:idx_region_code"`
	Name      string    `json:"name" gorm:"not null;size:100"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Country Country `json:"country,omitempty" gorm:"foreignKey:CountryID"`
	Cities  []City  `json:"cities,omitempty" gorm:"foreignKey:RegionID"`
}

// City represents a city in the location hierarchy
type City struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	RegionID  uint      `json:"region_id" gorm:"not null;index"`
	Code      string    `json:"code" gorm:"not null;size:10;index:idx_city_code"`
	Name      string    `json:"name" gorm:"not null;size:100"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Region Region `json:"region,omitempty" gorm:"foreignKey:RegionID"`
	Areas  []Area `json:"areas,omitempty" gorm:"foreignKey:CityID"`
}

// Area represents an area/district in the location hierarchy
type Area struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CityID    uint      `json:"city_id" gorm:"not null;index"`
	Code      string    `json:"code" gorm:"not null;size:10;index:idx_area_code"`
	Name      string    `json:"name" gorm:"not null;size:100"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	City        City         `json:"city,omitempty" gorm:"foreignKey:CityID"`
	PostalCodes []PostalCode `json:"postal_codes,omitempty" gorm:"foreignKey:AreaID"`
}

// PostalCode represents postal code information
type PostalCode struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AreaID      uint      `json:"area_id" gorm:"not null;index"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null;size:20"`
	Description string    `json:"description" gorm:"size:255"`
	Latitude    *float64  `json:"latitude,omitempty" gorm:"type:decimal(10,8)"`
	Longitude   *float64  `json:"longitude,omitempty" gorm:"type:decimal(11,8)"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Area Area `json:"area,omitempty" gorm:"foreignKey:AreaID"`
}

// PostalCodeAlias represents alternative postal codes or aliases
type PostalCodeAlias struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	PostalCodeID uint      `json:"postal_code_id" gorm:"not null;index"`
	AliasCode    string    `json:"alias_code" gorm:"uniqueIndex;not null;size:20"`
	Description  string    `json:"description" gorm:"size:255"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	PostalCode PostalCode `json:"postal_code,omitempty" gorm:"foreignKey:PostalCodeID"`
}

// LocationHierarchy represents the complete location path for a postal code
type LocationHierarchy struct {
	PostalCode  string   `json:"postal_code"`
	CountryCode string   `json:"country_code"`
	CountryName string   `json:"country_name"`
	RegionCode  string   `json:"region_code"`
	RegionName  string   `json:"region_name"`
	CityCode    string   `json:"city_code"`
	CityName    string   `json:"city_name"`
	AreaCode    string   `json:"area_code"`
	AreaName    string   `json:"area_name"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
}

// LocationValidationResult represents the result of postal code validation
type LocationValidationResult struct {
	IsValid   bool               `json:"is_valid"`
	Hierarchy *LocationHierarchy `json:"hierarchy,omitempty"`
	Error     string             `json:"error,omitempty"`
}
