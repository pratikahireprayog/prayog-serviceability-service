package models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// Country represents a country in the location hierarchy
type Country struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code         string    `json:"code" gorm:"unique;not null;size:10" validate:"required,min=2,max=10,alpha"`
	Name         string    `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	CurrencyCode *string   `json:"currency_code,omitempty" gorm:"size:3" validate:"omitempty,len=3,alpha"`
	PhoneCode    *string   `json:"phone_code,omitempty" gorm:"size:10" validate:"omitempty,min=1,max=10"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	Regions []Region `json:"regions,omitempty" gorm:"foreignKey:CountryID"`
}

// TableName returns the table name for Country
func (Country) TableName() string {
	return "country"
}

// Validate performs custom business rule validation for Country
func (c *Country) Validate() error {
	// Validate country code format (ISO standards)
	if len(c.Code) < 2 || len(c.Code) > 10 {
		return fmt.Errorf("country code must be between 2 and 10 characters")
	}

	// Validate currency code format if provided
	if c.CurrencyCode != nil && len(*c.CurrencyCode) != 3 {
		return fmt.Errorf("currency code must be exactly 3 characters")
	}

	return nil
}

// RegionType represents the type of region (state, province, etc.)
type RegionType struct {
	Code        string    `json:"code" gorm:"primaryKey;size:20" validate:"required,min=2,max=20,snake_case"`
	Name        string    `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	Description *string   `json:"description,omitempty" gorm:"type:text" validate:"omitempty,max=500"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	Regions []Region `json:"regions,omitempty" gorm:"foreignKey:RegionTypeCode"`
}

// TableName returns the table name for RegionType
func (RegionType) TableName() string {
	return "region_type"
}

// Validate performs custom business rule validation for RegionType
func (rt *RegionType) Validate() error {
	// Validate snake_case format for code
	if !isSnakeCase(rt.Code) {
		return fmt.Errorf("region type code must be in snake_case format")
	}
	return nil
}

// Region represents a state/province in the location hierarchy
type Region struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code           string     `json:"code" gorm:"unique;not null;size:20;index" validate:"required,min=2,max=20,snake_case"`
	CountryID      *uuid.UUID `json:"country_id,omitempty" gorm:"type:uuid;index"`
	CountryCode    *string    `json:"country_code,omitempty" gorm:"size:10;index" validate:"omitempty,min=2,max=10"`
	RegionTypeCode *string    `json:"region_type_code,omitempty" gorm:"size:20;index" validate:"omitempty,min=2,max=20"`
	RegionTypeID   *uuid.UUID `json:"region_type_id,omitempty" gorm:"type:uuid"`
	Name           string     `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	IsActive       bool       `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	Country    *Country    `json:"country,omitempty" gorm:"foreignKey:CountryID;references:ID;constraint:OnDelete:CASCADE"`
	RegionType *RegionType `json:"region_type,omitempty" gorm:"foreignKey:RegionTypeCode;references:Code;constraint:OnDelete:SET NULL"`
	Districts  []District  `json:"districts,omitempty" gorm:"foreignKey:RegionID"`
	Cities     []City      `json:"cities,omitempty" gorm:"foreignKey:RegionID"`
}

// TableName returns the table name for Region
func (Region) TableName() string {
	return "region"
}

// Validate performs custom business rule validation for Region
func (r *Region) Validate() error {
	// Validate snake_case format for code
	if !isSnakeCase(r.Code) {
		return fmt.Errorf("region code must be in snake_case format")
	}

	// Validate that either both or neither country ID and code are provided
	if (r.CountryID == nil) != (r.CountryCode == nil) {
		return fmt.Errorf("country ID and country code must both be provided or both be nil")
	}

	return nil
}

// District represents a district in the location hierarchy
type District struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code        string     `json:"code" gorm:"unique;not null;size:20;index" validate:"required,min=2,max=20,snake_case"`
	RegionID    *uuid.UUID `json:"region_id,omitempty" gorm:"type:uuid;index"`
	RegionCode  *string    `json:"region_code,omitempty" gorm:"size:20;index" validate:"omitempty,min=2,max=20"`
	CountryID   *uuid.UUID `json:"country_id,omitempty" gorm:"type:uuid;index"`
	CountryCode *string    `json:"country_code,omitempty" gorm:"size:10;index" validate:"omitempty,min=2,max=10"`
	Name        string     `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	Region  *Region  `json:"region,omitempty" gorm:"foreignKey:RegionID;references:ID;constraint:OnDelete:CASCADE"`
	Country *Country `json:"country,omitempty" gorm:"foreignKey:CountryID;references:ID;constraint:OnDelete:CASCADE"`
	Cities  []City   `json:"cities,omitempty" gorm:"foreignKey:DistrictID"`
}

// TableName returns the table name for District
func (District) TableName() string {
	return "district"
}

// City represents a city in the location hierarchy
type City struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code         string     `json:"code" gorm:"unique;not null;size:20;index"`
	RegionID     *uuid.UUID `json:"region_id,omitempty" gorm:"type:uuid;index"`
	RegionCode   *string    `json:"region_code,omitempty" gorm:"size:20;index"`
	CountryID    *uuid.UUID `json:"country_id,omitempty" gorm:"type:uuid;index"`
	CountryCode  *string    `json:"country_code,omitempty" gorm:"size:10;index"`
	DistrictID   *uuid.UUID `json:"district_id,omitempty" gorm:"type:uuid;index"`
	DistrictCode *string    `json:"district_code,omitempty" gorm:"size:20;index"`
	Name         string     `json:"name" gorm:"not null;size:100"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	Region   *Region   `json:"region,omitempty" gorm:"foreignKey:RegionID;references:ID;constraint:OnDelete:CASCADE"`
	Country  *Country  `json:"country,omitempty" gorm:"foreignKey:CountryID;references:ID;constraint:OnDelete:CASCADE"`
	District *District `json:"district,omitempty" gorm:"foreignKey:DistrictID;references:ID;constraint:OnDelete:SET NULL"`
	Areas    []Area    `json:"areas,omitempty" gorm:"foreignKey:CityID"`
}

// TableName returns the table name for City
func (City) TableName() string {
	return "city"
}

// Area represents an area/district in the location hierarchy
type Area struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code      string     `json:"code" gorm:"unique;not null;size:20;index"`
	CityID    *uuid.UUID `json:"city_id,omitempty" gorm:"type:uuid;index"`
	CityCode  *string    `json:"city_code,omitempty" gorm:"size:20;index"`
	Name      string     `json:"name" gorm:"not null;size:100"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	City        *City        `json:"city,omitempty" gorm:"foreignKey:CityID;references:ID;constraint:OnDelete:CASCADE"`
	PostalCodes []PostalCode `json:"postal_codes,omitempty" gorm:"foreignKey:AreaID"`
}

// TableName returns the table name for Area
func (Area) TableName() string {
	return "area"
}

// PostalCode represents postal code information
type PostalCode struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code          string     `json:"code" gorm:"unique;not null;size:20;index"`
	CountryID     *uuid.UUID `json:"country_id,omitempty" gorm:"type:uuid;index"`
	CountryCode   *string    `json:"country_code,omitempty" gorm:"size:10;index"`
	RegionID      *uuid.UUID `json:"region_id,omitempty" gorm:"type:uuid;index"`
	RegionCode    *string    `json:"region_code,omitempty" gorm:"size:20;index"`
	CityID        *uuid.UUID `json:"city_id,omitempty" gorm:"type:uuid;index"`
	CityCode      *string    `json:"city_code,omitempty" gorm:"size:20;index"`
	AreaID        *uuid.UUID `json:"area_id,omitempty" gorm:"type:uuid;index"`
	AreaCode      *string    `json:"area_code,omitempty" gorm:"size:20;index"`
	LocationScope *string    `json:"location_scope,omitempty" gorm:"size:20"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	Country *Country `json:"country,omitempty" gorm:"foreignKey:CountryID;references:ID;constraint:OnDelete:CASCADE"`
	Region  *Region  `json:"region,omitempty" gorm:"foreignKey:RegionID;references:ID;constraint:OnDelete:SET NULL"`
	City    *City    `json:"city,omitempty" gorm:"foreignKey:CityID;references:ID;constraint:OnDelete:SET NULL"`
	Area    *Area    `json:"area,omitempty" gorm:"foreignKey:AreaID;references:ID;constraint:OnDelete:SET NULL"`
}

// TableName returns the table name for PostalCode
func (PostalCode) TableName() string {
	return "postal_code"
}

// PostalCodeAlias represents alias names for postal codes
type PostalCodeAlias struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	PostalCodeID *uuid.UUID `json:"postal_code_id,omitempty" gorm:"type:uuid;index"`
	AliasCode    string     `json:"alias_code" gorm:"size:20;not null;index"`
	IsPrimary    bool       `json:"is_primary" gorm:"default:false"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	PostalCode *PostalCode `json:"postal_code,omitempty" gorm:"foreignKey:PostalCodeID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for PostalCodeAlias
func (PostalCodeAlias) TableName() string {
	return "postal_code_alias"
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

// Helper function to validate snake_case format
func isSnakeCase(s string) bool {
	// Snake case: lowercase letters, numbers, and underscores only
	// Must start with letter, no consecutive underscores, no trailing underscore
	pattern := `^[a-z][a-z0-9_]*[a-z0-9]$|^[a-z]$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}
