package dtos

import (
	"time"

	"github.com/google/uuid"
)

// CreatePartnerLocationCoverageRequest represents the request to create a partner location coverage
type CreatePartnerLocationCoverageRequest struct {
	PartnerCode  *string    `json:"partner_code,omitempty" validate:"omitempty,max=50" example:"shipyaari"`
	PostalCode   *string    `json:"postal_code,omitempty" validate:"omitempty,max=20" example:"110001"`
	PostalCodeID *uuid.UUID `json:"postal_code_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	ZoneType     *string    `json:"zone_type,omitempty" validate:"omitempty,oneof=PRIMARY SECONDARY BUFFER" example:"PRIMARY"`

	// Geographic context
	CountryCode *string `json:"country_code,omitempty" validate:"omitempty,max=10" example:"IN"`

	// Core serviceability attributes
	ProductType    *string `json:"product_type,omitempty" validate:"omitempty,max=50" example:"travel_free"`
	ParcelCategory *string `json:"parcel_category,omitempty" validate:"omitempty,max=50" example:"ecomm"`
	ServiceType    *string `json:"service_type,omitempty" validate:"omitempty,max=50" example:"sdd"`
	TATDays        *int    `json:"tat_days,omitempty" validate:"omitempty,min=0,max=365" example:"1"`

	// Service capabilities
	Pickup       *bool   `json:"pickup,omitempty" validate:"omitempty" example:"true"`
	Delivery     *bool   `json:"delivery,omitempty" validate:"omitempty" example:"true"`
	DeliveryMode *string `json:"delivery_mode,omitempty" validate:"omitempty,max=50,oneof=air surface rail" example:"surface"`
	CODAvailable *bool   `json:"cod_available,omitempty" validate:"omitempty" example:"true"`
	Insurance    *bool   `json:"insurance,omitempty" validate:"omitempty" example:"true"`

	// Weight constraints
	MinWeightKG *float64 `json:"min_weight_kg,omitempty" validate:"omitempty,min=0" example:"0.1"`
	MaxWeightKG *float64 `json:"max_weight_kg,omitempty" validate:"omitempty,min=0" example:"50.0"`

	IsActive *bool `json:"is_active,omitempty" validate:"omitempty" example:"true"`
}

// UpdatePartnerLocationCoverageRequest represents the request to update a partner location coverage
type UpdatePartnerLocationCoverageRequest struct {
	PartnerCode  *string    `json:"partner_code,omitempty" validate:"omitempty,max=50" example:"shipyaari"`
	PostalCode   *string    `json:"postal_code,omitempty" validate:"omitempty,max=20" example:"110001"`
	PostalCodeID *uuid.UUID `json:"postal_code_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	ZoneType     *string    `json:"zone_type,omitempty" validate:"omitempty,oneof=PRIMARY SECONDARY BUFFER" example:"PRIMARY"`

	// Geographic context
	CountryCode *string `json:"country_code,omitempty" validate:"omitempty,max=10" example:"IN"`

	// Core serviceability attributes
	ProductType    *string `json:"product_type,omitempty" validate:"omitempty,max=50" example:"travel_free"`
	ParcelCategory *string `json:"parcel_category,omitempty" validate:"omitempty,max=50" example:"ecomm"`
	ServiceType    *string `json:"service_type,omitempty" validate:"omitempty,max=50" example:"sdd"`
	TATDays        *int    `json:"tat_days,omitempty" validate:"omitempty,min=0,max=365" example:"1"`

	// Service capabilities
	Pickup       *bool   `json:"pickup,omitempty" validate:"omitempty" example:"true"`
	Delivery     *bool   `json:"delivery,omitempty" validate:"omitempty" example:"true"`
	DeliveryMode *string `json:"delivery_mode,omitempty" validate:"omitempty,max=50,oneof=air surface rail" example:"surface"`
	CODAvailable *bool   `json:"cod_available,omitempty" validate:"omitempty" example:"true"`
	Insurance    *bool   `json:"insurance,omitempty" validate:"omitempty" example:"true"`

	// Weight constraints
	MinWeightKG *float64 `json:"min_weight_kg,omitempty" validate:"omitempty,min=0" example:"0.1"`
	MaxWeightKG *float64 `json:"max_weight_kg,omitempty" validate:"omitempty,min=0" example:"50.0"`

	IsActive *bool `json:"is_active,omitempty" validate:"omitempty" example:"true"`
}

// PartnerLocationCoverageResponse represents the response for a partner location coverage
type PartnerLocationCoverageResponse struct {
	ID           uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PartnerID    *uuid.UUID `json:"partner_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	PartnerCode  *string    `json:"partner_code,omitempty" example:"PARTNER_001"`
	PostalCode   *string    `json:"postal_code,omitempty" example:"110001"`
	PostalCodeID *uuid.UUID `json:"postal_code_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	ZoneType     *string    `json:"zone_type,omitempty" example:"PRIMARY"`

	// Geographic context
	CountryCode *string `json:"country_code,omitempty" example:"IN"`

	// Core serviceability attributes
	ProductType    *string `json:"product_type,omitempty" example:"travel_free"`
	ParcelCategory *string `json:"parcel_category,omitempty" example:"ecomm"`
	ServiceType    *string `json:"service_type,omitempty" example:"sdd"`
	TATDays        *int    `json:"tat_days,omitempty" example:"1"`

	// Service capabilities
	Pickup       bool    `json:"pickup" example:"true"`
	Delivery     bool    `json:"delivery" example:"true"`
	DeliveryMode *string `json:"delivery_mode,omitempty" example:"surface"`
	CODAvailable bool    `json:"cod_available" example:"true"`
	Insurance    bool    `json:"insurance" example:"true"`

	// Weight constraints
	MinWeightKG *float64 `json:"min_weight_kg,omitempty" example:"0.1"`
	MaxWeightKG *float64 `json:"max_weight_kg,omitempty" example:"50.0"`

	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// PartnerLocationCoverageListResponse represents the response for listing partner location coverages
type PartnerLocationCoverageListResponse struct {
	Data       []PartnerLocationCoverageResponse `json:"data"`
	Pagination PaginationResponse                `json:"pagination"`
}

// PartnerLocationCoverageFiltersRequest represents filters for querying partner location coverages
type PartnerLocationCoverageFiltersRequest struct {
	PostalCode   *string    `json:"postal_code,omitempty" validate:"omitempty,max=20" example:"110001"`
	PostalCodeID *uuid.UUID `json:"postal_code_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	ZoneType     *string    `json:"zone_type,omitempty" validate:"omitempty,oneof=PRIMARY SECONDARY BUFFER" example:"PRIMARY"`

	// Geographic filters
	CountryCode *string `json:"country_code,omitempty" validate:"omitempty,max=10" example:"IN"`

	// Serviceability filters
	ProductType    *string  `json:"product_type,omitempty" validate:"omitempty,max=50" example:"travel_free"`
	ParcelCategory *string  `json:"parcel_category,omitempty" validate:"omitempty,max=50" example:"ecomm"`
	ServiceType    *string  `json:"service_type,omitempty" validate:"omitempty,max=50" example:"sdd"`
	TATDays        *int     `json:"tat_days,omitempty" validate:"omitempty,min=0,max=365" example:"1"`
	Pickup         *bool    `json:"pickup,omitempty" validate:"omitempty" example:"true"`
	Delivery       *bool    `json:"delivery,omitempty" validate:"omitempty" example:"true"`
	DeliveryMode   *string  `json:"delivery_mode,omitempty" validate:"omitempty,max=50,oneof=air surface rail" example:"surface"`
	CODAvailable   *bool    `json:"cod_available,omitempty" validate:"omitempty" example:"true"`
	Insurance      *bool    `json:"insurance,omitempty" validate:"omitempty" example:"true"`
	MinWeightKG    *float64 `json:"min_weight_kg,omitempty" validate:"omitempty,min=0" example:"0.1"`
	MaxWeightKG    *float64 `json:"max_weight_kg,omitempty" validate:"omitempty,min=0" example:"50.0"`

	IsActive *bool `json:"is_active,omitempty" validate:"omitempty" example:"true"`
	Limit    *int  `json:"limit,omitempty" validate:"omitempty,min=1,max=100" example:"10"`
	Offset   *int  `json:"offset,omitempty" validate:"omitempty,min=0" example:"0"`
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
	PostalCode        string                     `json:"postal_code" example:"110001"`
	PostalCodeID      uuid.UUID                  `json:"postal_code_id" example:"550e8400-e29b-41d4-a716-446655440000"`
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

// CoverageCheckResponse represents the response for checking coverage availability
type CoverageCheckResponse struct {
	Covered             bool                         `json:"covered" example:"true"`
	PartnerCode         *string                      `json:"partner_code,omitempty" example:"SHIPYAARI_001"`
	PostalCode          string                       `json:"postal_code" example:"110001"`
	ServiceCapabilities *CoverageServiceCapabilities `json:"service_capabilities,omitempty"`
	Message             *string                      `json:"message,omitempty" example:"Coverage available with all requested services"`
}

// CoverageServiceCapabilities represents the service capabilities for a coverage check
type CoverageServiceCapabilities struct {
	Pickup       bool                 `json:"pickup" example:"true"`
	Delivery     bool                 `json:"delivery" example:"true"`
	CODAvailable bool                 `json:"cod_available" example:"true"`
	Insurance    bool                 `json:"insurance" example:"true"`
	ServiceType  *string              `json:"service_type,omitempty" example:"sdd"`
	TATDays      *int                 `json:"tat_days,omitempty" example:"1"`
	DeliveryMode *string              `json:"delivery_mode,omitempty" example:"surface"`
	WeightRange  *CoverageWeightRange `json:"weight_range,omitempty"`
}

// CoverageWeightRange represents the weight range for coverage
type CoverageWeightRange struct {
	MinKG *float64 `json:"min_kg,omitempty" example:"0.1"`
	MaxKG *float64 `json:"max_kg,omitempty" example:"50.0"`
}
