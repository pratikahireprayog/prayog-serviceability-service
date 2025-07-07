package repositories

import (
	"context"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
)

// AttributeCategoryRepository defines the interface for attribute category operations
type AttributeCategoryRepository interface {
	GetByID(ctx context.Context, id string) (*models.AttributeCategory, error)
	GetByIDWithDeleted(ctx context.Context, id string) (*models.AttributeCategory, error)
	GetByCode(ctx context.Context, code string) (*models.AttributeCategory, error)
	GetByCodeWithDeleted(ctx context.Context, code string) (*models.AttributeCategory, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.AttributeCategory, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.AttributeCategory, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.AttributeCategory, int64, error)
	Create(ctx context.Context, category *models.AttributeCategory) error
	Update(ctx context.Context, id string, category *models.AttributeCategory) error
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
	CanBeDeleted(ctx context.Context, id string) (bool, error)
	GetStats(ctx context.Context, id string) (*AttributeCategoryStats, error)
}

// AttributeRepository defines the interface for attribute operations
type AttributeRepository interface {
	GetByID(ctx context.Context, id string) (*models.Attribute, error)
	GetByIDWithDeleted(ctx context.Context, id string) (*models.Attribute, error)
	GetByCode(ctx context.Context, code string) (*models.Attribute, error)
	GetByCodeWithDeleted(ctx context.Context, code string) (*models.Attribute, error)
	GetByCategoryID(ctx context.Context, categoryID string) ([]models.Attribute, error)
	GetByCategoryIDWithDeleted(ctx context.Context, categoryID string) ([]models.Attribute, error)
	GetByCategoryCode(ctx context.Context, categoryCode string) ([]models.Attribute, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.Attribute, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.Attribute, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.Attribute, int64, error)
	Create(ctx context.Context, attribute *models.Attribute) error
	Update(ctx context.Context, id string, attribute *models.Attribute) error
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
	CanBeDeleted(ctx context.Context, id string) (bool, error)
	GetStats(ctx context.Context, id string) (*AttributeStats, error)
}

// PartnerAttributeMapRepository defines the interface for partner attribute mapping operations
type PartnerAttributeMapRepository interface {
	GetByID(ctx context.Context, id string) (*models.PartnerAttributeMap, error)
	GetByIDWithDeleted(ctx context.Context, id string) (*models.PartnerAttributeMap, error)
	GetByPartnerCode(ctx context.Context, partnerCode string) ([]models.PartnerAttributeMap, error)
	GetByPartnerCodeWithDeleted(ctx context.Context, partnerCode string) ([]models.PartnerAttributeMap, error)
	GetByAttributeID(ctx context.Context, attributeID string) ([]models.PartnerAttributeMap, error)
	GetByAttributeIDWithDeleted(ctx context.Context, attributeID string) ([]models.PartnerAttributeMap, error)
	GetByAttributeCode(ctx context.Context, attributeCode string) ([]models.PartnerAttributeMap, error)
	GetByAttributeCodeWithDeleted(ctx context.Context, attributeCode string) ([]models.PartnerAttributeMap, error)
	GetByPartnerAndAttribute(ctx context.Context, partnerCode, attributeID string) (*models.PartnerAttributeMap, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.PartnerAttributeMap, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.PartnerAttributeMap, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.PartnerAttributeMap, int64, error)
	GetWithFilters(ctx context.Context, filters *PartnerAttributeMapFilters) ([]models.PartnerAttributeMap, int64, error)
	GetPartnerCodesByAttribute(ctx context.Context, attributeCode string) ([]string, error)
	GetAttributesByPartner(ctx context.Context, partnerCode string) ([]models.Attribute, error)
	Create(ctx context.Context, mapping *models.PartnerAttributeMap) error
	Update(ctx context.Context, id string, mapping *models.PartnerAttributeMap) error
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
	BulkCreate(ctx context.Context, mappings []models.PartnerAttributeMap) error
	BulkDelete(ctx context.Context, ids []uuid.UUID) error
	GetStats(ctx context.Context, partnerCode string) (*PartnerAttributeStats, error)
}

// Filter structs for complex queries

// PartnerAttributeMapFilters represents filtering options for partner attribute mappings
type PartnerAttributeMapFilters struct {
	PartnerCode   *string    `json:"partner_code,omitempty"`
	AttributeCode *string    `json:"attribute_code,omitempty"`
	AttributeID   *uuid.UUID `json:"attribute_id,omitempty"`
	CategoryID    *uuid.UUID `json:"category_id,omitempty"`
	IsActive      *bool      `json:"is_active,omitempty"`
	Limit         int        `json:"limit"`
	Offset        int        `json:"offset"`
}

// Statistics structs for analytics

// AttributeCategoryStats represents statistics for attribute categories
type AttributeCategoryStats struct {
	ID               uuid.UUID `json:"id"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	TotalAttributes  int64     `json:"total_attributes"`
	ActiveAttributes int64     `json:"active_attributes"`
	TotalMappings    int64     `json:"total_mappings"`
	ActiveMappings   int64     `json:"active_mappings"`
	UniquePartners   int64     `json:"unique_partners"`
}

// AttributeStats represents statistics for individual attributes
type AttributeStats struct {
	ID             uuid.UUID `json:"id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	CategoryCode   string    `json:"category_code"`
	CategoryName   string    `json:"category_name"`
	TotalMappings  int64     `json:"total_mappings"`
	ActiveMappings int64     `json:"active_mappings"`
	UniquePartners int64     `json:"unique_partners"`
	LastMappingAt  *string   `json:"last_mapping_at,omitempty"`
}

// PartnerAttributeStats represents partner attribute mapping statistics
type PartnerAttributeStats struct {
	PartnerCode       string  `json:"partner_code"`
	TotalMappings     int64   `json:"total_mappings"`
	ActiveMappings    int64   `json:"active_mappings"`
	UniqueCategories  int64   `json:"unique_categories"`
	FirstMappingAt    *string `json:"first_mapping_at,omitempty"`
	LastMappingAt     *string `json:"last_mapping_at,omitempty"`
	MostUsedCategory  *string `json:"most_used_category,omitempty"`
	MostUsedAttribute *string `json:"most_used_attribute,omitempty"`
}
