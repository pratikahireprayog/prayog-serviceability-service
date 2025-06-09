package models

import (
	"time"

	"github.com/google/uuid"
)

// LocationType represents the type of location entity
type LocationType struct {
	Code        string  `json:"code" gorm:"primaryKey;size:20"`
	Description *string `json:"description,omitempty" gorm:"type:text"`
}

// TableName returns the table name for LocationType
func (LocationType) TableName() string {
	return "location_type"
}

// LocationAlias represents alias names for location entities
type LocationAlias struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	EntityTypeCode *string   `json:"entity_type_code,omitempty" gorm:"size:20;index"`
	EntityType     *string   `json:"entity_type,omitempty" gorm:"size:20;index"`
	EntityID       uuid.UUID `json:"entity_id" gorm:"type:uuid;not null"`
	EntityCode     *string   `json:"entity_code,omitempty" gorm:"size:20"`
	AliasName      string    `json:"alias_name" gorm:"size:100;not null"`
	IsPrimary      bool      `json:"is_primary" gorm:"default:false"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	LocationType *LocationType `json:"location_type,omitempty" gorm:"foreignKey:EntityTypeCode;references:Code;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for LocationAlias
func (LocationAlias) TableName() string {
	return "location_alias"
}

// PartnerLocationCoverage represents partner coverage mapping to locations
type PartnerLocationCoverage struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	PartnerID     *uuid.UUID `json:"partner_id,omitempty" gorm:"type:uuid"`
	PartnerCode   *string    `json:"partner_code,omitempty" gorm:"size:50"`
	LocationScope *string    `json:"location_scope,omitempty" gorm:"size:20"`
	LocationID    *uuid.UUID `json:"location_id,omitempty" gorm:"type:uuid"`
	LocationCode  *string    `json:"location_code,omitempty" gorm:"size:20"`
	ZoneType      *string    `json:"zone_type,omitempty" gorm:"size:20"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for PartnerLocationCoverage
func (PartnerLocationCoverage) TableName() string {
	return "partner_location_coverage"
}

// PartnerLocationCoverageFilters represents filters for partner location coverage queries
type PartnerLocationCoverageFilters struct {
	PartnerID     *uuid.UUID `json:"partner_id,omitempty"`
	LocationScope string     `json:"location_scope,omitempty"`
	LocationID    *uuid.UUID `json:"location_id,omitempty"`
	ZoneType      string     `json:"zone_type,omitempty"`
	IsActive      *bool      `json:"is_active,omitempty"`
}

// PartnerLocationCoverageResult represents partner coverage with location details
type PartnerLocationCoverageResult struct {
	PartnerID         uuid.UUID          `json:"partner_id"`
	LocationScope     string             `json:"location_scope"`
	LocationID        uuid.UUID          `json:"location_id"`
	ZoneType          string             `json:"zone_type"`
	LocationHierarchy *LocationHierarchy `json:"location_hierarchy,omitempty"`
}
