package dtos

import (
	"time"

	"github.com/google/uuid"
)

// CreatePartnerLocationCoverageRequest represents the request to create a partner location coverage
type CreatePartnerLocationCoverageRequest struct {
	LocationScope *string    `json:"location_scope,omitempty" validate:"omitempty,oneof=POSTAL_CODE AREA CITY REGION COUNTRY" example:"CITY"`
	LocationID    *uuid.UUID `json:"location_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	LocationCode  *string    `json:"location_code,omitempty" validate:"omitempty,max=20" example:"NYC"`
	ZoneType      *string    `json:"zone_type,omitempty" validate:"omitempty,oneof=PRIMARY SECONDARY BUFFER" example:"PRIMARY"`
	IsActive      *bool      `json:"is_active,omitempty" validate:"omitempty" example:"true"`
}

// UpdatePartnerLocationCoverageRequest represents the request to update a partner location coverage
type UpdatePartnerLocationCoverageRequest struct {
	LocationScope *string    `json:"location_scope,omitempty" validate:"omitempty,oneof=POSTAL_CODE AREA CITY REGION COUNTRY" example:"CITY"`
	LocationID    *uuid.UUID `json:"location_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	LocationCode  *string    `json:"location_code,omitempty" validate:"omitempty,max=20" example:"NYC"`
	ZoneType      *string    `json:"zone_type,omitempty" validate:"omitempty,oneof=PRIMARY SECONDARY BUFFER" example:"PRIMARY"`
	IsActive      *bool      `json:"is_active,omitempty" validate:"omitempty" example:"true"`
}

// PartnerLocationCoverageResponse represents the response for a partner location coverage
type PartnerLocationCoverageResponse struct {
	ID            uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PartnerID     *uuid.UUID `json:"partner_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	PartnerCode   *string    `json:"partner_code,omitempty" example:"PARTNER_001"`
	LocationScope *string    `json:"location_scope,omitempty" example:"CITY"`
	LocationID    *uuid.UUID `json:"location_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	LocationCode  *string    `json:"location_code,omitempty" example:"NYC"`
	ZoneType      *string    `json:"zone_type,omitempty" example:"PRIMARY"`
	IsActive      bool       `json:"is_active" example:"true"`
	CreatedAt     time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt     time.Time  `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// PartnerLocationCoverageListResponse represents the response for listing partner location coverages
type PartnerLocationCoverageListResponse struct {
	Data       []PartnerLocationCoverageResponse `json:"data"`
	Pagination PaginationResponse                `json:"pagination"`
}

// PartnerLocationCoverageFiltersRequest represents filters for querying partner location coverages
type PartnerLocationCoverageFiltersRequest struct {
	LocationScope *string    `json:"location_scope,omitempty" validate:"omitempty,oneof=POSTAL_CODE AREA CITY REGION COUNTRY" example:"CITY"`
	LocationID    *uuid.UUID `json:"location_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	ZoneType      *string    `json:"zone_type,omitempty" validate:"omitempty,oneof=PRIMARY SECONDARY BUFFER" example:"PRIMARY"`
	IsActive      *bool      `json:"is_active,omitempty" validate:"omitempty" example:"true"`
	Limit         *int       `json:"limit,omitempty" validate:"omitempty,min=1,max=100" example:"10"`
	Offset        *int       `json:"offset,omitempty" validate:"omitempty,min=0" example:"0"`
}

// BulkCreatePartnerLocationCoverageRequest represents the request to create multiple partner location coverages
type BulkCreatePartnerLocationCoverageRequest struct {
	Coverages []CreatePartnerLocationCoverageRequest `json:"coverages" validate:"required,min=1,max=100"`
}

// BulkCreatePartnerLocationCoverageResponse represents the response for bulk creating partner location coverages
type BulkCreatePartnerLocationCoverageResponse struct {
	Created []PartnerLocationCoverageResponse `json:"created"`
	Failed  []BulkCreateError                 `json:"failed,omitempty"`
}

// BulkCreateError represents an error that occurred during bulk creation
type BulkCreateError struct {
	Index   int                                  `json:"index"`
	Error   string                               `json:"error"`
	Request CreatePartnerLocationCoverageRequest `json:"request"`
}

// PartnerLocationCoverageResultResponse represents partner coverage with location details
type PartnerLocationCoverageResultResponse struct {
	PartnerID         uuid.UUID                  `json:"partner_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	LocationScope     string                     `json:"location_scope" example:"CITY"`
	LocationID        uuid.UUID                  `json:"location_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ZoneType          string                     `json:"zone_type" example:"PRIMARY"`
	LocationHierarchy *LocationHierarchyResponse `json:"location_hierarchy,omitempty"`
}

// LocationHierarchyResponse represents location hierarchy information
type LocationHierarchyResponse struct {
	PostalCode  string `json:"postal_code,omitempty" example:"10001"`
	Country     string `json:"country,omitempty" example:"United States"`
	CountryCode string `json:"country_code,omitempty" example:"US"`
	Region      string `json:"region,omitempty" example:"New York"`
	RegionCode  string `json:"region_code,omitempty" example:"NY"`
	City        string `json:"city,omitempty" example:"New York"`
	CityCode    string `json:"city_code,omitempty" example:"NYC"`
	Area        string `json:"area,omitempty" example:"Manhattan"`
	AreaCode    string `json:"area_code,omitempty" example:"MAN"`
}
