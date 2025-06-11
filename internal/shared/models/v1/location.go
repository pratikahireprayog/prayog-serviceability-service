package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SoftDeleteModel defines the interface for models with soft delete capability
type SoftDeleteModel interface {
	SoftDelete(db *gorm.DB) error
	Restore(db *gorm.DB) error
	IsDeletedRecord() bool
}

// Country represents a country in the location hierarchy
type Country struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code         string     `json:"code" gorm:"unique;not null;size:2" validate:"required,len=2,alpha"`
	Name         string     `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	CurrencyCode *string    `json:"currency_code,omitempty" gorm:"size:3" validate:"omitempty,len=3,alpha"`
	PhoneCode    *string    `json:"phone_code,omitempty" gorm:"size:10" validate:"omitempty,min=1,max=10"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted    bool       `json:"is_deleted" gorm:"default:false;index"`

	// Relationships
	Regions []Region `json:"regions,omitempty" gorm:"foreignKey:CountryID"`
}

// TableName returns the table name for Country
func (Country) TableName() string {
	return "country"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (Country) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the country as deleted without removing it from database
func (c *Country) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	c.DeletedAt = &now
	c.IsDeleted = true
	return db.Save(c).Error
}

// Restore un-deletes a soft-deleted country
func (c *Country) Restore(db *gorm.DB) error {
	c.DeletedAt = nil
	c.IsDeleted = false
	return db.Save(c).Error
}

// IsDeletedRecord checks if the country is soft-deleted
func (c *Country) IsDeletedRecord() bool {
	return c.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for Country
func (c *Country) ValidateBusinessRules() error {
	// Validate snake_case format for code (lowercase letters only for consistency with project conventions)
	if !isSnakeCase(c.Code) {
		return fmt.Errorf("country code must be in snake_case format (lowercase, 2 characters)")
	}

	// Validate country code length (ISO 3166 A-2 standard, but lowercase)
	if len(c.Code) != 2 {
		return fmt.Errorf("country code must be exactly 2 characters")
	}

	// Validate currency code format if provided
	if c.CurrencyCode != nil && len(*c.CurrencyCode) != 3 {
		return fmt.Errorf("currency code must be exactly 3 characters")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if c.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted country")
	}

	return nil
}

// RegionType represents the type of region (state, province, etc.)
type RegionType struct {
	Code        string     `json:"code" gorm:"primaryKey;size:20" validate:"required,min=2,max=20"`
	Name        string     `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	Description *string    `json:"description,omitempty" gorm:"type:text" validate:"omitempty,max=500"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted   bool       `json:"is_deleted" gorm:"default:false;index"`

	// Relationships
	Regions []Region `json:"regions,omitempty" gorm:"foreignKey:RegionTypeCode"`
}

// TableName returns the table name for RegionType
func (RegionType) TableName() string {
	return "region_type"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (RegionType) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the region type as deleted without removing it from database
func (rt *RegionType) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	rt.DeletedAt = &now
	rt.IsDeleted = true
	return db.Save(rt).Error
}

// Restore un-deletes a soft-deleted region type
func (rt *RegionType) Restore(db *gorm.DB) error {
	rt.DeletedAt = nil
	rt.IsDeleted = false
	return db.Save(rt).Error
}

// IsDeletedRecord checks if the region type is soft-deleted
func (rt *RegionType) IsDeletedRecord() bool {
	return rt.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for RegionType
func (rt *RegionType) ValidateBusinessRules() error {
	// Validate snake_case format for code
	if !isSnakeCase(rt.Code) {
		return fmt.Errorf("region type code must be in snake_case format")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if rt.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted region type")
	}

	return nil
}

// Region represents a state/province in the location hierarchy
type Region struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code           string     `json:"code" gorm:"unique;not null;size:20;index" validate:"required,min=2,max=20"`
	CountryID      *uuid.UUID `json:"country_id,omitempty" gorm:"type:uuid;index"`
	CountryCode    *string    `json:"country_code,omitempty" gorm:"size:2;index" validate:"omitempty,len=2"`
	RegionTypeCode *string    `json:"region_type_code,omitempty" gorm:"size:20;index" validate:"omitempty,min=2,max=20"`
	RegionTypeID   *uuid.UUID `json:"region_type_id,omitempty" gorm:"type:uuid"`
	Name           string     `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	IsActive       bool       `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted      bool       `json:"is_deleted" gorm:"default:false;index"`

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

// DefaultScope applies default query conditions to filter out soft-deleted records
func (Region) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the region as deleted without removing it from database
func (r *Region) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	r.DeletedAt = &now
	r.IsDeleted = true
	return db.Save(r).Error
}

// Restore un-deletes a soft-deleted region
func (r *Region) Restore(db *gorm.DB) error {
	r.DeletedAt = nil
	r.IsDeleted = false
	return db.Save(r).Error
}

// IsDeletedRecord checks if the region is soft-deleted
func (r *Region) IsDeletedRecord() bool {
	return r.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for Region
func (r *Region) ValidateBusinessRules() error {
	// Validate snake_case format for code
	if !isSnakeCase(r.Code) {
		return fmt.Errorf("region code must be in snake_case format")
	}

	// Validate that either both or neither country ID and code are provided
	if (r.CountryID == nil) != (r.CountryCode == nil) {
		return fmt.Errorf("country ID and country code must both be provided or both be nil")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if r.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted region")
	}

	return nil
}

// District represents a district in the location hierarchy
type District struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code        string     `json:"code" gorm:"unique;not null;size:20;index" validate:"required,min=2,max=20"`
	RegionID    *uuid.UUID `json:"region_id,omitempty" gorm:"type:uuid;index"`
	RegionCode  *string    `json:"region_code,omitempty" gorm:"size:20;index" validate:"omitempty,min=2,max=20"`
	CountryID   *uuid.UUID `json:"country_id,omitempty" gorm:"type:uuid;index"`
	CountryCode *string    `json:"country_code,omitempty" gorm:"size:2;index" validate:"omitempty,len=2"`
	Name        string     `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted   bool       `json:"is_deleted" gorm:"default:false;index"`

	// Relationships
	Region  *Region  `json:"region,omitempty" gorm:"foreignKey:RegionID;references:ID;constraint:OnDelete:CASCADE"`
	Country *Country `json:"country,omitempty" gorm:"foreignKey:CountryID;references:ID;constraint:OnDelete:CASCADE"`
	Cities  []City   `json:"cities,omitempty" gorm:"foreignKey:DistrictID"`
}

// TableName returns the table name for District
func (District) TableName() string {
	return "district"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (District) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the district as deleted without removing it from database
func (d *District) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	d.DeletedAt = &now
	d.IsDeleted = true
	return db.Save(d).Error
}

// Restore un-deletes a soft-deleted district
func (d *District) Restore(db *gorm.DB) error {
	d.DeletedAt = nil
	d.IsDeleted = false
	return db.Save(d).Error
}

// IsDeletedRecord checks if the district is soft-deleted
func (d *District) IsDeletedRecord() bool {
	return d.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for District
func (d *District) ValidateBusinessRules() error {
	// Validate snake_case format for code
	if !isSnakeCase(d.Code) {
		return fmt.Errorf("district code must be in snake_case format")
	}

	// Validate that either both or neither region ID and code are provided
	if (d.RegionID == nil) != (d.RegionCode == nil) {
		return fmt.Errorf("region ID and region code must both be provided or both be nil")
	}

	// Validate that either both or neither country ID and code are provided
	if (d.CountryID == nil) != (d.CountryCode == nil) {
		return fmt.Errorf("country ID and country code must both be provided or both be nil")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if d.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted district")
	}

	return nil
}

// City represents a city in the location hierarchy
type City struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code         string     `json:"code" gorm:"unique;not null;size:20;index"`
	RegionID     *uuid.UUID `json:"region_id,omitempty" gorm:"type:uuid;index"`
	RegionCode   *string    `json:"region_code,omitempty" gorm:"size:20;index"`
	CountryID    *uuid.UUID `json:"country_id,omitempty" gorm:"type:uuid;index"`
	CountryCode  *string    `json:"country_code,omitempty" gorm:"size:2;index"`
	DistrictID   *uuid.UUID `json:"district_id,omitempty" gorm:"type:uuid;index"`
	DistrictCode *string    `json:"district_code,omitempty" gorm:"size:20;index"`
	Name         string     `json:"name" gorm:"not null;size:100"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted    bool       `json:"is_deleted" gorm:"default:false;index"`

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

// DefaultScope applies default query conditions to filter out soft-deleted records
func (City) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the city as deleted without removing it from database
func (c *City) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	c.DeletedAt = &now
	c.IsDeleted = true
	return db.Save(c).Error
}

// Restore un-deletes a soft-deleted city
func (c *City) Restore(db *gorm.DB) error {
	c.DeletedAt = nil
	c.IsDeleted = false
	return db.Save(c).Error
}

// IsDeletedRecord checks if the city is soft-deleted
func (c *City) IsDeletedRecord() bool {
	return c.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for City
func (c *City) ValidateBusinessRules() error {
	// Validate snake_case format for code
	if !isSnakeCase(c.Code) {
		return fmt.Errorf("city code must be in snake_case format")
	}

	// Validate that either both or neither region ID and code are provided
	if (c.RegionID == nil) != (c.RegionCode == nil) {
		return fmt.Errorf("region ID and region code must both be provided or both be nil")
	}

	// Validate that either both or neither country ID and code are provided
	if (c.CountryID == nil) != (c.CountryCode == nil) {
		return fmt.Errorf("country ID and country code must both be provided or both be nil")
	}

	// Validate that either both or neither district ID and code are provided
	if (c.DistrictID == nil) != (c.DistrictCode == nil) {
		return fmt.Errorf("district ID and district code must both be provided or both be nil")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if c.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted city")
	}

	return nil
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
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted bool       `json:"is_deleted" gorm:"default:false;index"`

	// Relationships
	City        *City        `json:"city,omitempty" gorm:"foreignKey:CityID;references:ID;constraint:OnDelete:CASCADE"`
	PostalCodes []PostalCode `json:"postal_codes,omitempty" gorm:"foreignKey:AreaID"`
}

// TableName returns the table name for Area
func (Area) TableName() string {
	return "area"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (Area) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the area as deleted without removing it from database
func (a *Area) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	a.DeletedAt = &now
	a.IsDeleted = true
	return db.Save(a).Error
}

// Restore un-deletes a soft-deleted area
func (a *Area) Restore(db *gorm.DB) error {
	a.DeletedAt = nil
	a.IsDeleted = false
	return db.Save(a).Error
}

// IsDeletedRecord checks if the area is soft-deleted
func (a *Area) IsDeletedRecord() bool {
	return a.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for Area
func (a *Area) ValidateBusinessRules() error {
	// Validate snake_case format for code
	if !isSnakeCase(a.Code) {
		return fmt.Errorf("area code must be in snake_case format")
	}

	// Validate that either both or neither city ID and code are provided
	if (a.CityID == nil) != (a.CityCode == nil) {
		return fmt.Errorf("city ID and city code must both be provided or both be nil")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if a.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted area")
	}

	return nil
}

// PostalCode represents postal code information
type PostalCode struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code          string     `json:"code" gorm:"unique;not null;size:20;index" validate:"required,min=3,max=20"`
	CountryID     *uuid.UUID `json:"country_id,omitempty" gorm:"type:uuid;index"`
	CountryCode   *string    `json:"country_code,omitempty" gorm:"size:2;index" validate:"omitempty,len=2"` // ISO 3166 A-2
	RegionID      *uuid.UUID `json:"region_id,omitempty" gorm:"type:uuid;index"`
	RegionCode    *string    `json:"region_code,omitempty" gorm:"size:20;index" validate:"omitempty,min=2,max=20"`
	CityID        *uuid.UUID `json:"city_id,omitempty" gorm:"type:uuid;index"`
	CityCode      *string    `json:"city_code,omitempty" gorm:"size:20;index" validate:"omitempty,min=2,max=20"`
	AreaID        *uuid.UUID `json:"area_id,omitempty" gorm:"type:uuid;index"`
	AreaCode      *string    `json:"area_code,omitempty" gorm:"size:20;index" validate:"omitempty,min=2,max=20"`
	LocationScope *string    `json:"location_scope,omitempty" gorm:"size:20" validate:"omitempty,oneof=country region city area"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted     bool       `json:"is_deleted" gorm:"default:false;index"`

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

// DefaultScope applies default query conditions to filter out soft-deleted records
func (PostalCode) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the postal code as deleted without removing it from database
func (p *PostalCode) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	p.DeletedAt = &now
	p.IsDeleted = true
	return db.Save(p).Error
}

// Restore un-deletes a soft-deleted postal code
func (p *PostalCode) Restore(db *gorm.DB) error {
	p.DeletedAt = nil
	p.IsDeleted = false
	return db.Save(p).Error
}

// IsDeletedRecord checks if the postal code is soft-deleted
func (p *PostalCode) IsDeletedRecord() bool {
	return p.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for PostalCode
func (p *PostalCode) ValidateBusinessRules() error {
	// Basic format validation for postal code
	if len(p.Code) < 1 || len(p.Code) > 20 {
		return fmt.Errorf("postal code must be between 1 and 20 characters")
	}

	// Validate that either both or neither country ID and code are provided
	if (p.CountryID == nil) != (p.CountryCode == nil) {
		return fmt.Errorf("country ID and country code must both be provided or both be nil")
	}

	// Validate that either both or neither region ID and code are provided
	if (p.RegionID == nil) != (p.RegionCode == nil) {
		return fmt.Errorf("region ID and region code must both be provided or both be nil")
	}

	// Validate that either both or neither city ID and code are provided
	if (p.CityID == nil) != (p.CityCode == nil) {
		return fmt.Errorf("city ID and city code must both be provided or both be nil")
	}

	// Validate that either both or neither area ID and code are provided
	if (p.AreaID == nil) != (p.AreaCode == nil) {
		return fmt.Errorf("area ID and area code must both be provided or both be nil")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if p.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted postal code")
	}

	return nil
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
	if len(s) == 0 {
		return false
	}

	// Must start with lowercase letter
	if s[0] < 'a' || s[0] > 'z' {
		return false
	}

	prevUnderscore := false
	for i, char := range s {
		// Allow lowercase letters, numbers, and underscores
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			prevUnderscore = false
			continue
		} else if char == '_' {
			// No consecutive underscores or trailing underscore
			if prevUnderscore || i == len(s)-1 {
				return false
			}
			prevUnderscore = true
		} else {
			// Invalid character
			return false
		}
	}

	return true
}
