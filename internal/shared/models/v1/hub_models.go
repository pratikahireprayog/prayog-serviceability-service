package models

import (
	"time"

	"github.com/google/uuid"
)

// Hub represents a hub in the serviceability network
type Hub struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code          *string   `json:"code,omitempty" gorm:"size:50;unique"`
	Name          *string   `json:"name,omitempty" gorm:"size:100"`
	ContactName   *string   `json:"contact_name,omitempty" gorm:"size:100"`
	ContactNumber *string   `json:"contact_number,omitempty" gorm:"size:20"`
	Address       *string   `json:"address,omitempty" gorm:"type:text"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	LocationCoverages []HubLocationCoverage `json:"location_coverages,omitempty" gorm:"foreignKey:HubID"`
	Specifications    []HubSpecification    `json:"specifications,omitempty" gorm:"foreignKey:HubID"`
}

// TableName returns the table name for Hub
func (Hub) TableName() string {
	return "hub"
}

// HubLocationCoverage represents the location coverage mapping for a hub
type HubLocationCoverage struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	HubID         *uuid.UUID `json:"hub_id,omitempty" gorm:"type:uuid;index"`
	HubCode       *string    `json:"hub_code,omitempty" gorm:"size:50;index"`
	LocationScope *string    `json:"location_scope,omitempty" gorm:"size:20"`
	LocationID    *uuid.UUID `json:"location_id,omitempty" gorm:"type:uuid"`
	LocationCode  *string    `json:"location_code,omitempty" gorm:"size:20"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	Hub *Hub `json:"hub,omitempty" gorm:"foreignKey:HubID;references:ID;constraint:OnDelete:SET NULL"`
}

// TableName returns the table name for HubLocationCoverage
func (HubLocationCoverage) TableName() string {
	return "hub_location_coverage"
}

// HubSpecification represents specifications for a hub
type HubSpecification struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	HubID     *uuid.UUID `json:"hub_id,omitempty" gorm:"type:uuid;index"`
	HubCode   *string    `json:"hub_code,omitempty" gorm:"size:50;index"`
	SpecID    *uuid.UUID `json:"spec_id,omitempty" gorm:"type:uuid"`
	SpecCode  *string    `json:"spec_code,omitempty" gorm:"size:50"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	Hub *Hub `json:"hub,omitempty" gorm:"foreignKey:HubID;references:ID;constraint:OnDelete:SET NULL"`
}

// TableName returns the table name for HubSpecification
func (HubSpecification) TableName() string {
	return "hub_specification"
}
