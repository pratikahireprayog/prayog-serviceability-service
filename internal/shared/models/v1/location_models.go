package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LocationType represents the type of location entity
type LocationType struct {
	Code        string     `json:"code" gorm:"primaryKey;size:20"`
	Description *string    `json:"description,omitempty" gorm:"type:text"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted   bool       `json:"is_deleted" gorm:"default:false;index"`
}

// TableName returns the table name for LocationType
func (LocationType) TableName() string {
	return "location_type"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (LocationType) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the location type as deleted without removing it from database
func (lt *LocationType) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	lt.DeletedAt = &now
	lt.IsDeleted = true
	return db.Save(lt).Error
}

// Restore un-deletes a soft-deleted location type
func (lt *LocationType) Restore(db *gorm.DB) error {
	lt.DeletedAt = nil
	lt.IsDeleted = false
	return db.Save(lt).Error
}

// IsDeletedRecord checks if the location type is soft-deleted
func (lt *LocationType) IsDeletedRecord() bool {
	return lt.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for LocationType
func (lt *LocationType) ValidateBusinessRules() error {
	// Validate snake_case format for code
	if !isSnakeCase(lt.Code) {
		return fmt.Errorf("location type code must be in snake_case format")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if lt.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted location type")
	}

	return nil
}

// LocationAlias represents alias names for location entities
type LocationAlias struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	EntityTypeCode *string    `json:"entity_type_code,omitempty" gorm:"size:20;index"`
	EntityType     *string    `json:"entity_type,omitempty" gorm:"size:20;index"`
	EntityID       uuid.UUID  `json:"entity_id" gorm:"type:uuid;not null"`
	EntityCode     *string    `json:"entity_code,omitempty" gorm:"size:20"`
	AliasName      string     `json:"alias_name" gorm:"size:100;not null"`
	IsPrimary      bool       `json:"is_primary" gorm:"default:false"`
	IsActive       bool       `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted      bool       `json:"is_deleted" gorm:"default:false;index"`

	// Relationships
	LocationType *LocationType `json:"location_type,omitempty" gorm:"foreignKey:EntityTypeCode;references:Code;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for LocationAlias
func (LocationAlias) TableName() string {
	return "location_alias"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (LocationAlias) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the location alias as deleted without removing it from database
func (la *LocationAlias) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	la.DeletedAt = &now
	la.IsDeleted = true
	return db.Save(la).Error
}

// Restore un-deletes a soft-deleted location alias
func (la *LocationAlias) Restore(db *gorm.DB) error {
	la.DeletedAt = nil
	la.IsDeleted = false
	return db.Save(la).Error
}

// IsDeletedRecord checks if the location alias is soft-deleted
func (la *LocationAlias) IsDeletedRecord() bool {
	return la.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for LocationAlias
func (la *LocationAlias) ValidateBusinessRules() error {
	// Validate alias name is not empty
	if len(la.AliasName) < 1 || len(la.AliasName) > 100 {
		return fmt.Errorf("alias name must be between 1 and 100 characters")
	}

	// Validate entity ID is provided
	if la.EntityID == (uuid.UUID{}) {
		return fmt.Errorf("entity ID must be provided")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if la.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted location alias")
	}

	return nil
}

// PartnerLocationCoverage represents partner coverage mapping to locations with serviceability details
type PartnerLocationCoverage struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	PartnerID    *uuid.UUID `json:"partner_id,omitempty" gorm:"type:uuid"`
	PartnerCode  *string    `json:"partner_code,omitempty" gorm:"size:50"`
	PostalCode   *string    `json:"postal_code,omitempty" gorm:"size:20"`
	PostalCodeID *uuid.UUID `json:"postal_code_id,omitempty" gorm:"type:uuid"`
	ZoneType     *string    `json:"zone_type,omitempty" gorm:"size:20"`

	// Geographic context
	CountryCode *string `json:"country_code,omitempty" gorm:"size:10"`

	// Core serviceability attributes
	ProductType    *string `json:"product_type,omitempty" gorm:"size:50"`
	ParcelCategory *string `json:"parcel_category,omitempty" gorm:"size:50"`
	ServiceType    *string `json:"service_type,omitempty" gorm:"size:50"`
	TATDays        *int    `json:"tat_days,omitempty"`

	// Service capabilities
	Pickup       bool    `json:"pickup" gorm:"default:false"`
	Delivery     bool    `json:"delivery" gorm:"default:false"`
	DeliveryMode *string `json:"delivery_mode,omitempty" gorm:"size:50"`
	CODAvailable bool    `json:"cod_available" gorm:"default:false"`
	Insurance    bool    `json:"insurance" gorm:"default:false"`

	// Weight constraints
	MinWeightKG *float64 `json:"min_weight_kg,omitempty" gorm:"type:numeric(10,3)"`
	MaxWeightKG *float64 `json:"max_weight_kg,omitempty" gorm:"type:numeric(10,3)"`

	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted bool       `json:"is_deleted" gorm:"default:false;index"`
}

// TableName returns the table name for PartnerLocationCoverage
func (PartnerLocationCoverage) TableName() string {
	return "partner_location_coverage"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (PartnerLocationCoverage) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the partner location coverage as deleted without removing it from database
func (plc *PartnerLocationCoverage) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	plc.DeletedAt = &now
	plc.IsDeleted = true
	return db.Save(plc).Error
}

// Restore un-deletes a soft-deleted partner location coverage
func (plc *PartnerLocationCoverage) Restore(db *gorm.DB) error {
	plc.DeletedAt = nil
	plc.IsDeleted = false
	return db.Save(plc).Error
}

// IsDeletedRecord checks if the partner location coverage is soft-deleted
func (plc *PartnerLocationCoverage) IsDeletedRecord() bool {
	return plc.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for PartnerLocationCoverage
func (plc *PartnerLocationCoverage) ValidateBusinessRules() error {
	// Note: Removed partner ID/code consistency validation to allow flexible API usage
	// The API can now work with just partner ID (from URL path) without requiring partner code

	// Note: Removed postal code ID consistency validation to allow working with postal code only
	// The API can now work with just postal code without requiring postal code ID

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if plc.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted partner location coverage")
	}

	return nil
}

// PartnerLocationCoverageFilters represents filters for partner location coverage queries
type PartnerLocationCoverageFilters struct {
	PartnerID    *uuid.UUID `json:"partner_id,omitempty"`
	PartnerCode  *string    `json:"partner_code,omitempty"`
	PostalCode   *string    `json:"postal_code,omitempty"`
	PostalCodeID *uuid.UUID `json:"postal_code_id,omitempty"`
	ZoneType     string     `json:"zone_type,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`

	// Serviceability filters
	CountryCode    *string  `json:"country_code,omitempty"`
	ProductType    *string  `json:"product_type,omitempty"`
	ParcelCategory *string  `json:"parcel_category,omitempty"`
	ServiceType    *string  `json:"service_type,omitempty"`
	TATDays        *int     `json:"tat_days,omitempty"`
	Pickup         *bool    `json:"pickup,omitempty"`
	Delivery       *bool    `json:"delivery,omitempty"`
	DeliveryMode   *string  `json:"delivery_mode,omitempty"`
	CODAvailable   *bool    `json:"cod_available,omitempty"`
	Insurance      *bool    `json:"insurance,omitempty"`
	MinWeightKG    *float64 `json:"min_weight_kg,omitempty"`
	MaxWeightKG    *float64 `json:"max_weight_kg,omitempty"`
}

// PartnerLocationCoverageResult represents partner coverage with location details
type PartnerLocationCoverageResult struct {
	PartnerID         uuid.UUID          `json:"partner_id"`
	PostalCode        string             `json:"postal_code"`
	PostalCodeID      uuid.UUID          `json:"postal_code_id"`
	ZoneType          string             `json:"zone_type"`
	LocationHierarchy *LocationHierarchy `json:"location_hierarchy,omitempty"`
}
