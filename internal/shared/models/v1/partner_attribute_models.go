package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AttributeCategory represents a category that groups attributes (e.g., 'Parcel Category', 'Service Type')
type AttributeCategory struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code      string     `json:"code" gorm:"unique;not null;size:50;index" validate:"required,min=2,max=50"`
	Name      string     `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted bool       `json:"is_deleted" gorm:"default:false;index"`

	// Relationships
	Attributes []Attribute `json:"attributes,omitempty" gorm:"foreignKey:CategoryID"`
}

// TableName returns the table name for AttributeCategory
func (AttributeCategory) TableName() string {
	return "attribute_category"
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (AttributeCategory) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the attribute category as deleted without removing it from database
func (ac *AttributeCategory) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	ac.DeletedAt = &now
	ac.IsDeleted = true
	return db.Save(ac).Error
}

// Restore un-deletes a soft-deleted attribute category
func (ac *AttributeCategory) Restore(db *gorm.DB) error {
	ac.DeletedAt = nil
	ac.IsDeleted = false
	return db.Save(ac).Error
}

// IsDeletedRecord checks if the attribute category is soft-deleted
func (ac *AttributeCategory) IsDeletedRecord() bool {
	return ac.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for AttributeCategory
func (ac *AttributeCategory) ValidateBusinessRules() error {
	// Validate snake_case format for code
	if !isSnakeCase(ac.Code) {
		return fmt.Errorf("attribute category code must be in snake_case format")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if ac.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted attribute category")
	}

	return nil
}

// CanBeDeleted checks if the attribute category can be safely deleted (no active attributes)
func (ac *AttributeCategory) CanBeDeleted(db *gorm.DB) (bool, error) {
	var count int64
	err := db.Model(&Attribute{}).Where("category_id = ? AND is_deleted = ?", ac.ID, false).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// Attribute represents individual attributes that can be assigned (e.g., 'E-commerce', 'Express')
type Attribute struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CategoryID uuid.UUID  `json:"category_id" gorm:"type:uuid;not null;index" validate:"required"`
	Code       string     `json:"code" gorm:"not null;size:50;index" validate:"required,min=2,max=50"`
	Name       string     `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted  bool       `json:"is_deleted" gorm:"default:false;index"`

	// Relationships
	Category             *AttributeCategory    `json:"category,omitempty" gorm:"foreignKey:CategoryID;references:ID;constraint:OnDelete:CASCADE"`
	PartnerAttributeMaps []PartnerAttributeMap `json:"partner_attribute_maps,omitempty" gorm:"foreignKey:AttributeID"`
}

// TableName returns the table name for Attribute
func (Attribute) TableName() string {
	return "attribute"
}

// BeforeCreate ensures unique code within category
func (a *Attribute) BeforeCreate(tx *gorm.DB) error {
	var count int64
	err := tx.Model(&Attribute{}).Where("category_id = ? AND code = ? AND is_deleted = ?", a.CategoryID, a.Code, false).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("attribute code '%s' already exists in this category", a.Code)
	}
	return nil
}

// BeforeUpdate ensures unique code within category during updates
func (a *Attribute) BeforeUpdate(tx *gorm.DB) error {
	var count int64
	err := tx.Model(&Attribute{}).Where("category_id = ? AND code = ? AND id != ? AND is_deleted = ?", a.CategoryID, a.Code, a.ID, false).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("attribute code '%s' already exists in this category", a.Code)
	}
	return nil
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (Attribute) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the attribute as deleted without removing it from database
func (a *Attribute) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	a.DeletedAt = &now
	a.IsDeleted = true
	return db.Save(a).Error
}

// Restore un-deletes a soft-deleted attribute
func (a *Attribute) Restore(db *gorm.DB) error {
	a.DeletedAt = nil
	a.IsDeleted = false
	return db.Save(a).Error
}

// IsDeletedRecord checks if the attribute is soft-deleted
func (a *Attribute) IsDeletedRecord() bool {
	return a.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for Attribute
func (a *Attribute) ValidateBusinessRules() error {
	// Validate snake_case format for code
	if !isSnakeCase(a.Code) {
		return fmt.Errorf("attribute code must be in snake_case format")
	}

	// Validate that category ID is provided
	if a.CategoryID == (uuid.UUID{}) {
		return fmt.Errorf("category ID must be provided")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if a.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted attribute")
	}

	return nil
}

// CanBeDeleted checks if the attribute can be safely deleted (no active partner mappings)
func (a *Attribute) CanBeDeleted(db *gorm.DB) (bool, error) {
	var count int64
	err := db.Model(&PartnerAttributeMap{}).Where("attribute_id = ? AND is_deleted = ?", a.ID, false).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// PartnerAttributeMap represents the mapping between partners and attributes
type PartnerAttributeMap struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	PartnerCode   string     `json:"partner_code" gorm:"not null;size:50;index" validate:"required,min=1,max=50"`
	AttributeID   uuid.UUID  `json:"attribute_id" gorm:"type:uuid;not null;index" validate:"required"`
	AttributeCode string     `json:"attribute_code" gorm:"not null;size:50;index" validate:"required,min=2,max=50"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	IsDeleted     bool       `json:"is_deleted" gorm:"default:false;index"`

	// Relationships
	Attribute *Attribute `json:"attribute,omitempty" gorm:"foreignKey:AttributeID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for PartnerAttributeMap
func (PartnerAttributeMap) TableName() string {
	return "partner_attribute_map"
}

// BeforeCreate ensures unique partner-attribute mapping
func (pam *PartnerAttributeMap) BeforeCreate(tx *gorm.DB) error {
	var count int64
	err := tx.Model(&PartnerAttributeMap{}).Where("partner_code = ? AND attribute_id = ? AND is_deleted = ?", pam.PartnerCode, pam.AttributeID, false).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("partner '%s' is already mapped to this attribute", pam.PartnerCode)
	}
	return nil
}

// BeforeUpdate ensures unique partner-attribute mapping during updates
func (pam *PartnerAttributeMap) BeforeUpdate(tx *gorm.DB) error {
	var count int64
	err := tx.Model(&PartnerAttributeMap{}).Where("partner_code = ? AND attribute_id = ? AND id != ? AND is_deleted = ?", pam.PartnerCode, pam.AttributeID, pam.ID, false).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("partner '%s' is already mapped to this attribute", pam.PartnerCode)
	}
	return nil
}

// DefaultScope applies default query conditions to filter out soft-deleted records
func (PartnerAttributeMap) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// SoftDelete marks the partner attribute map as deleted without removing it from database
func (pam *PartnerAttributeMap) SoftDelete(db *gorm.DB) error {
	now := time.Now()
	pam.DeletedAt = &now
	pam.IsDeleted = true
	return db.Save(pam).Error
}

// Restore un-deletes a soft-deleted partner attribute map
func (pam *PartnerAttributeMap) Restore(db *gorm.DB) error {
	pam.DeletedAt = nil
	pam.IsDeleted = false
	return db.Save(pam).Error
}

// IsDeletedRecord checks if the partner attribute map is soft-deleted
func (pam *PartnerAttributeMap) IsDeletedRecord() bool {
	return pam.IsDeleted
}

// ValidateBusinessRules performs custom business rule validation for PartnerAttributeMap
func (pam *PartnerAttributeMap) ValidateBusinessRules() error {
	// Validate partner code format
	if len(pam.PartnerCode) < 1 || len(pam.PartnerCode) > 50 {
		return fmt.Errorf("partner code must be between 1 and 50 characters")
	}

	// Validate attribute code format
	if !isSnakeCase(pam.AttributeCode) {
		return fmt.Errorf("attribute code must be in snake_case format")
	}

	// Validate that attribute ID is provided
	if pam.AttributeID == (uuid.UUID{}) {
		return fmt.Errorf("attribute ID must be provided")
	}

	// Soft delete validation: cannot update essential fields of soft-deleted records
	if pam.IsDeleted {
		return fmt.Errorf("cannot modify essential fields of a soft-deleted partner attribute map")
	}

	return nil
}

// BeforeSave ensures attribute_code is populated from the related attribute
func (pam *PartnerAttributeMap) BeforeSave(tx *gorm.DB) error {
	if pam.AttributeCode == "" && pam.AttributeID != (uuid.UUID{}) {
		var attribute Attribute
		if err := tx.First(&attribute, pam.AttributeID).Error; err != nil {
			return fmt.Errorf("failed to find attribute: %w", err)
		}
		pam.AttributeCode = attribute.Code
	}
	return nil
}
