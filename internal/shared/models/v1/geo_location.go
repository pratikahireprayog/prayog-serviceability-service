package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// GeoLocation represents a geographic location with detailed metadata
type GeoLocation struct {
	PostalCode       string     `json:"postal_code" gorm:"primaryKey;column:postal_code"`
	Name             string     `json:"name" gorm:"not null;size:200" validate:"required,min=1,max=200"`
	AsciiName        string     `json:"ascii_name" gorm:"not null;size:200;column:asciiname" validate:"required,min=1,max=200"`
	AlternateNames   *string    `json:"alternate_names,omitempty" gorm:"type:text;column:alternatenames"`
	Latitude         float64    `json:"latitude" gorm:"not null;type:decimal(10,8)" validate:"required,min=-90,max=90"`
	Longitude        float64    `json:"longitude" gorm:"not null;type:decimal(11,8)" validate:"required,min=-180,max=180"`
	FeatureClass     string     `json:"feature_class" gorm:"not null;size:1;column:feature_class" validate:"required,len=1"`
	FeatureCode      string     `json:"feature_code" gorm:"not null;size:10;column:feature_code" validate:"required,min=1,max=10"`
	CountryCode      string     `json:"country_code" gorm:"not null;size:2;column:country_code;index" validate:"required,len=2"`
	CC2              *string    `json:"cc2,omitempty" gorm:"size:200;column:cc2"`
	Admin1Code       *string    `json:"admin1_code,omitempty" gorm:"size:20;column:admin1_code;index"`
	Admin2Code       *string    `json:"admin2_code,omitempty" gorm:"size:80;column:admin2_code;index"`
	Admin3Code       *string    `json:"admin3_code,omitempty" gorm:"size:20;column:admin3_code"`
	Admin4Code       *string    `json:"admin4_code,omitempty" gorm:"size:20;column:admin4_code"`
	Population       *int64     `json:"population,omitempty" gorm:"column:population"`
	Elevation        *int       `json:"elevation,omitempty" gorm:"column:elevation"`
	DEM              *int       `json:"dem,omitempty" gorm:"column:dem"`
	Timezone         *string    `json:"timezone,omitempty" gorm:"size:40;column:timezone"`
	ModificationDate time.Time  `json:"modification_date" gorm:"column:modification_date;default:CURRENT_TIMESTAMP"`
	CreatedAt        time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for GeoLocation
func (GeoLocation) TableName() string {
	return "geo_locations"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (GeoLocation) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_at IS NULL")
}

// SoftDelete marks the geo location as deleted without removing it from database
func (g *GeoLocation) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	g.DeletedAt = &now
	g.ModificationDate = now
	g.UpdatedAt = now
	return db.Save(g).Error
}

// Restore un-deletes a soft-deleted geo location
func (g *GeoLocation) Restore(db *gorm.DB) error {
	now := time.Now()
	g.DeletedAt = nil
	g.ModificationDate = now
	g.UpdatedAt = now
	return db.Save(g).Error
}

// IsDeletedRecord checks if the geo location is soft-deleted
func (g *GeoLocation) IsDeletedRecord() bool {
	return g.DeletedAt != nil
}

// ValidateBusinessRules performs custom business rule validation for GeoLocation
func (g *GeoLocation) ValidateBusinessRules() error {
	// Validate latitude range
	if g.Latitude < -90 || g.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}

	// Validate longitude range
	if g.Longitude < -180 || g.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	// Validate feature class (should be single character)
	if len(g.FeatureClass) != 1 {
		return fmt.Errorf("feature class must be exactly 1 character")
	}

	// Validate country code (should be 2 characters)
	if len(g.CountryCode) != 2 {
		return fmt.Errorf("country code must be exactly 2 characters")
	}

	// Validate population (should be non-negative if provided)
	if g.Population != nil && *g.Population < 0 {
		return fmt.Errorf("population must be non-negative")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if g.IsDeletedRecord() {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted geo location")
	}

	return nil
}

// BeforeCreate hook to set timestamps
func (g *GeoLocation) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = now
	}
	if g.UpdatedAt.IsZero() {
		g.UpdatedAt = now
	}
	if g.ModificationDate.IsZero() {
		g.ModificationDate = now
	}
	return nil
}

// BeforeUpdate hook to update timestamps
func (g *GeoLocation) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	g.UpdatedAt = now
	g.ModificationDate = now
	return nil
}
