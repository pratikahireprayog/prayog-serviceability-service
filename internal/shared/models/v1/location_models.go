package models

import (
	"time"
)

// LocationType represents the type of location entity
type LocationType struct {
	Code        string    `json:"code" gorm:"primaryKey;size:50"`
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName returns the table name for LocationType
func (LocationType) TableName() string {
	return "location_types"
}

// LocationAlias represents alias names for location entities
type LocationAlias struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	EntityType string    `json:"entity_type" gorm:"size:50;not null"`
	EntityID   uint      `json:"entity_id" gorm:"not null"`
	AliasName  string    `json:"alias_name" gorm:"size:255;not null"`
	IsPrimary  bool      `json:"is_primary" gorm:"default:false"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName returns the table name for LocationAlias
func (LocationAlias) TableName() string {
	return "location_aliases"
}

// PartnerLocationCoverage represents partner coverage mapping to locations
type PartnerLocationCoverage struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	PartnerID     uint      `json:"partner_id" gorm:"not null;comment:Ref to Partner Service"`
	LocationScope string    `json:"location_scope" gorm:"size:20;not null;comment:POSTAL_CODE|AREA|CITY|REGION|COUNTRY"`
	LocationID    uint      `json:"location_id" gorm:"not null"`
	ZoneType      string    `json:"zone_type" gorm:"size:20;not null;comment:PRIMARY|SECONDARY|BUFFER"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName returns the table name for PartnerLocationCoverage
func (PartnerLocationCoverage) TableName() string {
	return "partner_location_coverages"
}

// PartnerLocationCoverageFilters represents filters for partner location coverage queries
type PartnerLocationCoverageFilters struct {
	PartnerID     *uint  `json:"partner_id,omitempty"`
	LocationScope string `json:"location_scope,omitempty"`
	LocationID    *uint  `json:"location_id,omitempty"`
	ZoneType      string `json:"zone_type,omitempty"`
	IsActive      *bool  `json:"is_active,omitempty"`
}

// PartnerLocationCoverageResult represents partner coverage with location details
type PartnerLocationCoverageResult struct {
	PartnerID         uint               `json:"partner_id"`
	LocationScope     string             `json:"location_scope"`
	LocationID        uint               `json:"location_id"`
	ZoneType          string             `json:"zone_type"`
	LocationHierarchy *LocationHierarchy `json:"location_hierarchy,omitempty"`
}
