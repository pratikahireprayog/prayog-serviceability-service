package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Hub represents a hub in the serviceability network
type Hub struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code          *string    `json:"code,omitempty" gorm:"size:50;unique"`
	Name          *string    `json:"name,omitempty" gorm:"size:100"`
	ContactName   *string    `json:"contact_name,omitempty" gorm:"size:100"`
	ContactNumber *string    `json:"contact_number,omitempty" gorm:"size:20"`
	Address       *string    `json:"address,omitempty" gorm:"type:text"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted     bool       `json:"is_deleted" gorm:"default:false;index"`

	// Relationships
	LocationCoverages []HubLocationCoverage `json:"location_coverages,omitempty" gorm:"foreignKey:HubID"`
	Specifications    []HubSpecification    `json:"specifications,omitempty" gorm:"foreignKey:HubID"`
}

// TableName returns the table name for Hub
func (Hub) TableName() string {
	return "hub"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (Hub) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the hub as deleted without removing it from database
func (h *Hub) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	h.DeletedAt = &now
	h.IsDeleted = true
	return db.Save(h).Error
}

// Restore un-deletes a soft-deleted hub
func (h *Hub) Restore(db *gorm.DB) error {
	h.DeletedAt = nil
	h.IsDeleted = false
	return db.Save(h).Error
}

// IsDeletedRecord checks if the hub is soft-deleted
func (h *Hub) IsDeletedRecord() bool {
	return h.IsDeleted
}

// Validate performs custom business rule validation for Hub
func (h *Hub) ValidateBusinessRules() error {
	// Validate hub code format if provided
	if h.Code != nil && !isSnakeCase(*h.Code) {
		return fmt.Errorf("hub code must be in snake_case format")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if h.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted hub")
	}

	return nil
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
	DeletedAt     *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted     bool       `json:"is_deleted" gorm:"default:false;index"`

	// Relationships
	Hub *Hub `json:"hub,omitempty" gorm:"foreignKey:HubID;references:ID;constraint:OnDelete:SET NULL"`
}

// TableName returns the table name for HubLocationCoverage
func (HubLocationCoverage) TableName() string {
	return "hub_location_coverage"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (HubLocationCoverage) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the hub location coverage as deleted without removing it from database
func (hlc *HubLocationCoverage) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	hlc.DeletedAt = &now
	hlc.IsDeleted = true
	return db.Save(hlc).Error
}

// Restore un-deletes a soft-deleted hub location coverage
func (hlc *HubLocationCoverage) Restore(db *gorm.DB) error {
	hlc.DeletedAt = nil
	hlc.IsDeleted = false
	return db.Save(hlc).Error
}

// IsDeletedRecord checks if the hub location coverage is soft-deleted
func (hlc *HubLocationCoverage) IsDeletedRecord() bool {
	return hlc.IsDeleted
}

// Validate performs custom business rule validation for HubLocationCoverage
func (hlc *HubLocationCoverage) ValidateBusinessRules() error {
	// Validate that either both or neither hub ID and code are provided
	if (hlc.HubID == nil) != (hlc.HubCode == nil) {
		return fmt.Errorf("hub ID and hub code must both be provided or both be nil")
	}

	// Validate that either both or neither location ID and code are provided
	if (hlc.LocationID == nil) != (hlc.LocationCode == nil) {
		return fmt.Errorf("location ID and location code must both be provided or both be nil")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if hlc.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted hub location coverage")
	}

	return nil
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
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted bool       `json:"is_deleted" gorm:"default:false;index"`

	// Relationships
	Hub *Hub `json:"hub,omitempty" gorm:"foreignKey:HubID;references:ID;constraint:OnDelete:SET NULL"`
}

// TableName returns the table name for HubSpecification
func (HubSpecification) TableName() string {
	return "hub_specification"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (HubSpecification) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the hub specification as deleted without removing it from database
func (hs *HubSpecification) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	hs.DeletedAt = &now
	hs.IsDeleted = true
	return db.Save(hs).Error
}

// Restore un-deletes a soft-deleted hub specification
func (hs *HubSpecification) Restore(db *gorm.DB) error {
	hs.DeletedAt = nil
	hs.IsDeleted = false
	return db.Save(hs).Error
}

// IsDeletedRecord checks if the hub specification is soft-deleted
func (hs *HubSpecification) IsDeletedRecord() bool {
	return hs.IsDeleted
}

// Validate performs custom business rule validation for HubSpecification
func (hs *HubSpecification) ValidateBusinessRules() error {
	// Validate that either both or neither hub ID and code are provided
	if (hs.HubID == nil) != (hs.HubCode == nil) {
		return fmt.Errorf("hub ID and hub code must both be provided or both be nil")
	}

	// Validate that either both or neither spec ID and code are provided
	if (hs.SpecID == nil) != (hs.SpecCode == nil) {
		return fmt.Errorf("spec ID and spec code must both be provided or both be nil")
	}

	// Validate spec code format if provided
	if hs.SpecCode != nil && !isSnakeCase(*hs.SpecCode) {
		return fmt.Errorf("spec code must be in snake_case format")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if hs.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted hub specification")
	}

	return nil
}
