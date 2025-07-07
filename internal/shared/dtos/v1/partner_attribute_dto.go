package dtos

import (
	"time"

	"github.com/google/uuid"
)

// AttributeCategory DTOs

// CreateAttributeCategoryRequest represents request to create a new attribute category
type CreateAttributeCategoryRequest struct {
	Code     string `json:"code" validate:"required,min=2,max=50,snake_case" example:"parcel_category"`
	Name     string `json:"name" validate:"required,min=2,max=100" example:"Parcel Category"`
	IsActive *bool  `json:"is_active,omitempty" example:"true"`
}

// UpdateAttributeCategoryRequest represents request to update an attribute category
type UpdateAttributeCategoryRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=2,max=100" example:"Parcel Category"`
	IsActive *bool   `json:"is_active,omitempty" example:"true"`
}

// AttributeCategoryResponse represents attribute category response
type AttributeCategoryResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Code      string    `json:"code" example:"parcel_category"`
	Name      string    `json:"name" example:"Parcel Category"`
	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// AttributeCategoryListResponse represents list of attribute categories with pagination
type AttributeCategoryListResponse = StandardListResponse[AttributeCategoryResponse]

// Attribute DTOs

// CreateAttributeRequest represents request to create a new attribute
type CreateAttributeRequest struct {
	CategoryID    uuid.UUID `json:"category_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Code          string    `json:"code" validate:"required,min=2,max=50,snake_case" example:"ecom"`
	AttributeCode *string   `json:"attribute_code,omitempty" validate:"omitempty,min=2,max=50,snake_case" example:"ecom"`
	Name          string    `json:"name" validate:"required,min=2,max=100" example:"E-commerce"`
	IsActive      *bool     `json:"is_active,omitempty" example:"true"`
}

// UpdateAttributeRequest represents request to update an attribute
type UpdateAttributeRequest struct {
	CategoryID    *uuid.UUID `json:"category_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Code          *string    `json:"code,omitempty" validate:"omitempty,min=2,max=50,snake_case" example:"ecom"`
	AttributeCode *string    `json:"attribute_code,omitempty" validate:"omitempty,min=2,max=50,snake_case" example:"ecom"`
	Name          *string    `json:"name,omitempty" validate:"omitempty,min=2,max=100" example:"E-commerce"`
	IsActive      *bool      `json:"is_active,omitempty" example:"true"`
}

// AttributeResponse represents attribute response
type AttributeResponse struct {
	ID            uuid.UUID                  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CategoryID    uuid.UUID                  `json:"category_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Code          string                     `json:"code" example:"ecom"`
	AttributeCode string                     `json:"attribute_code" example:"ecom"`
	Name          string                     `json:"name" example:"E-commerce"`
	IsActive      bool                       `json:"is_active" example:"true"`
	CreatedAt     time.Time                  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt     time.Time                  `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	Category      *AttributeCategoryResponse `json:"category,omitempty"`
}

// AttributeListResponse represents list of attributes with pagination
type AttributeListResponse = StandardListResponse[AttributeResponse]

// PartnerAttributeMap DTOs

// CreatePartnerAttributeMapRequest represents request to create a new partner attribute mapping
type CreatePartnerAttributeMapRequest struct {
	PartnerID   *uuid.UUID `json:"partner_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	PartnerCode string     `json:"partner_code" validate:"required,min=1,max=50" example:"partner_xyz"`
	AttributeID uuid.UUID  `json:"attribute_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	IsActive    *bool      `json:"is_active,omitempty" example:"true"`
}

// UpdatePartnerAttributeMapRequest represents request to update a partner attribute mapping
type UpdatePartnerAttributeMapRequest struct {
	PartnerID   *uuid.UUID `json:"partner_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	PartnerCode *string    `json:"partner_code,omitempty" validate:"omitempty,min=1,max=50" example:"partner_xyz"`
	AttributeID *uuid.UUID `json:"attribute_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	IsActive    *bool      `json:"is_active,omitempty" example:"true"`
}

// PartnerAttributeMapResponse represents partner attribute mapping response
type PartnerAttributeMapResponse struct {
	ID            uuid.UUID          `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PartnerID     *uuid.UUID         `json:"partner_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	PartnerCode   string             `json:"partner_code" example:"partner_xyz"`
	AttributeID   uuid.UUID          `json:"attribute_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	AttributeCode string             `json:"attribute_code" example:"ecom"`
	IsActive      bool               `json:"is_active" example:"true"`
	CreatedAt     time.Time          `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt     time.Time          `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	Attribute     *AttributeResponse `json:"attribute,omitempty"`
}

// PartnerAttributeMapListResponse represents list of partner attribute mappings with pagination
type PartnerAttributeMapListResponse = StandardListResponse[PartnerAttributeMapResponse]

// Filter DTOs for Partner Attribute Map

// PartnerAttributeMapFilterRequest represents filtering options for partner attribute mappings
type PartnerAttributeMapFilterRequest struct {
	PartnerID     *uuid.UUID `json:"partner_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	PartnerCode   *string    `json:"partner_code,omitempty" validate:"omitempty,min=1,max=50" example:"partner_xyz"`
	AttributeCode *string    `json:"attribute_code,omitempty" validate:"omitempty,min=2,max=50" example:"ecom"`
	AttributeID   *uuid.UUID `json:"attribute_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	CategoryID    *uuid.UUID `json:"category_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	IsActive      *bool      `json:"is_active,omitempty" example:"true"`
	Limit         *int       `json:"limit,omitempty" validate:"omitempty,min=1,max=100" example:"10"`
	Offset        *int       `json:"offset,omitempty" validate:"omitempty,min=0" example:"0"`
}

// Partner Codes by Attribute Response

// PartnerCodesByAttributeResponse represents partners mapped to a specific attribute
type PartnerCodesByAttributeResponse struct {
	AttributeCode    string    `json:"attribute_code" example:"ecom"`
	AttributeID      uuid.UUID `json:"attribute_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	AttributeName    string    `json:"attribute_name" example:"E-commerce"`
	PartnerCodes     []string  `json:"partner_codes" example:"[\"partner_xyz\",\"partner_abc\"]"`
	TotalPartners    int       `json:"total_partners" example:"2"`
	ActivePartners   int       `json:"active_partners" example:"2"`
	InactivePartners int       `json:"inactive_partners" example:"0"`
}

// Attributes by Partner Response

// AttributesByPartnerResponse represents attributes mapped to a specific partner
type AttributesByPartnerResponse struct {
	PartnerCode        string              `json:"partner_code" example:"partner_xyz"`
	Attributes         []AttributeResponse `json:"attributes"`
	TotalAttributes    int                 `json:"total_attributes" example:"3"`
	ActiveAttributes   int                 `json:"active_attributes" example:"3"`
	InactiveAttributes int                 `json:"inactive_attributes" example:"0"`
}

// Batch operations DTOs

// BulkCreatePartnerAttributeMapRequest represents request to create multiple partner attribute mappings
type BulkCreatePartnerAttributeMapRequest struct {
	Mappings []CreatePartnerAttributeMapRequest `json:"mappings" validate:"required,min=1,max=100"`
}

// BulkCreatePartnerAttributeMapResponse represents response for bulk creating partner attribute mappings
type BulkCreatePartnerAttributeMapResponse struct {
	Created []PartnerAttributeMapResponse     `json:"created"`
	Failed  []BulkCreatePartnerAttributeError `json:"failed,omitempty"`
}

// BulkCreatePartnerAttributeError represents an error during bulk creation
type BulkCreatePartnerAttributeError struct {
	Index   int                              `json:"index" example:"0"`
	Error   string                           `json:"error" example:"partner code already mapped to this attribute"`
	Request CreatePartnerAttributeMapRequest `json:"request"`
}

// BulkDeletePartnerAttributeMapRequest represents request to delete multiple partner attribute mappings
type BulkDeletePartnerAttributeMapRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,min=1,max=100"`
}

// BulkDeletePartnerAttributeMapResponse represents response for bulk deleting partner attribute mappings
type BulkDeletePartnerAttributeMapResponse struct {
	Deleted []uuid.UUID                       `json:"deleted"`
	Failed  []BulkDeletePartnerAttributeError `json:"failed,omitempty"`
}

// BulkDeletePartnerAttributeError represents an error during bulk deletion
type BulkDeletePartnerAttributeError struct {
	ID    uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Error string    `json:"error" example:"mapping not found"`
}

// Statistics and Analytics DTOs

// AttributeCategoryStatsResponse represents statistics for attribute categories
type AttributeCategoryStatsResponse struct {
	ID               uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Code             string    `json:"code" example:"parcel_category"`
	Name             string    `json:"name" example:"Parcel Category"`
	TotalAttributes  int       `json:"total_attributes" example:"5"`
	ActiveAttributes int       `json:"active_attributes" example:"4"`
	TotalMappings    int       `json:"total_mappings" example:"150"`
	ActiveMappings   int       `json:"active_mappings" example:"120"`
	UniquePartners   int       `json:"unique_partners" example:"25"`
}

// AttributeStatsResponse represents statistics for individual attributes
type AttributeStatsResponse struct {
	ID             uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Code           string     `json:"code" example:"ecom"`
	Name           string     `json:"name" example:"E-commerce"`
	CategoryCode   string     `json:"category_code" example:"parcel_category"`
	CategoryName   string     `json:"category_name" example:"Parcel Category"`
	TotalMappings  int        `json:"total_mappings" example:"50"`
	ActiveMappings int        `json:"active_mappings" example:"45"`
	UniquePartners int        `json:"unique_partners" example:"15"`
	CreatedAt      time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	LastMappingAt  *time.Time `json:"last_mapping_at,omitempty" example:"2023-12-01T10:30:00Z"`
}

// PartnerAttributeStatsResponse represents partner attribute mapping statistics
type PartnerAttributeStatsResponse struct {
	PartnerCode       string     `json:"partner_code" example:"partner_xyz"`
	TotalMappings     int        `json:"total_mappings" example:"8"`
	ActiveMappings    int        `json:"active_mappings" example:"7"`
	UniqueCategories  int        `json:"unique_categories" example:"3"`
	FirstMappingAt    *time.Time `json:"first_mapping_at,omitempty" example:"2023-01-15T08:00:00Z"`
	LastMappingAt     *time.Time `json:"last_mapping_at,omitempty" example:"2023-12-01T10:30:00Z"`
	MostUsedCategory  *string    `json:"most_used_category,omitempty" example:"parcel_category"`
	MostUsedAttribute *string    `json:"most_used_attribute,omitempty" example:"ecom"`
}

// Summary DTOs for dashboard/overview endpoints

// AttributeSystemSummaryResponse represents overall system summary
type AttributeSystemSummaryResponse struct {
	TotalCategories      int        `json:"total_categories" example:"5"`
	ActiveCategories     int        `json:"active_categories" example:"4"`
	TotalAttributes      int        `json:"total_attributes" example:"25"`
	ActiveAttributes     int        `json:"active_attributes" example:"20"`
	TotalMappings        int        `json:"total_mappings" example:"500"`
	ActiveMappings       int        `json:"active_mappings" example:"450"`
	UniquePartners       int        `json:"unique_partners" example:"75"`
	LastCategoryCreated  *time.Time `json:"last_category_created,omitempty" example:"2023-11-15T14:20:00Z"`
	LastAttributeCreated *time.Time `json:"last_attribute_created,omitempty" example:"2023-12-01T09:15:00Z"`
	LastMappingCreated   *time.Time `json:"last_mapping_created,omitempty" example:"2023-12-15T16:45:00Z"`
}

// Validation and Health Check DTOs

// AttributeValidationRequest represents request to validate attribute configurations
type AttributeValidationRequest struct {
	CategoryID  *uuid.UUID `json:"category_id,omitempty" validate:"omitempty,uuid"`
	AttributeID *uuid.UUID `json:"attribute_id,omitempty" validate:"omitempty,uuid"`
	PartnerCode *string    `json:"partner_code,omitempty" validate:"omitempty,min=1,max=50"`
}

// AttributeValidationResponse represents validation results
type AttributeValidationResponse struct {
	IsValid          bool                          `json:"is_valid" example:"true"`
	Errors           []string                      `json:"errors,omitempty" example:"[]"`
	Warnings         []string                      `json:"warnings,omitempty" example:"[]"`
	OrphanedMappings []PartnerAttributeMapResponse `json:"orphaned_mappings,omitempty"`
	EmptyCategories  []AttributeCategoryResponse   `json:"empty_categories,omitempty"`
	UnusedAttributes []AttributeResponse           `json:"unused_attributes,omitempty"`
}
