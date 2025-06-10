package dtos

import (
	"time"

	"github.com/google/uuid"
)

// Common DTOs

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Offset int `json:"offset" query:"offset" validate:"min=0"`
	Limit  int `json:"limit" query:"limit" validate:"min=1,max=100"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Offset      int   `json:"offset"`
	Limit       int   `json:"limit"`
	Total       int64 `json:"total"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

// Country DTOs

// CreateCountryRequest represents request to create a new country
type CreateCountryRequest struct {
	Code         string  `json:"code" validate:"required,min=2,max=10,alpha"`
	Name         string  `json:"name" validate:"required,min=2,max=100"`
	CurrencyCode *string `json:"currency_code,omitempty" validate:"omitempty,len=3,alpha"`
	PhoneCode    *string `json:"phone_code,omitempty" validate:"omitempty,min=1,max=10"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// UpdateCountryRequest represents request to update a country
type UpdateCountryRequest struct {
	Name         *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	CurrencyCode *string `json:"currency_code,omitempty" validate:"omitempty,len=3,alpha"`
	PhoneCode    *string `json:"phone_code,omitempty" validate:"omitempty,min=1,max=10"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// CountryResponse represents country response
type CountryResponse struct {
	ID           uuid.UUID `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	CurrencyCode *string   `json:"currency_code,omitempty"`
	PhoneCode    *string   `json:"phone_code,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CountryListResponse represents list of countries with pagination
// Deprecated: Use StandardListResponse[CountryResponse] instead
type CountryListResponse = StandardListResponse[CountryResponse]

// RegionType DTOs

// CreateRegionTypeRequest represents request to create a new region type
type CreateRegionTypeRequest struct {
	Code        string  `json:"code" validate:"required,min=2,max=20,snake_case"`
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// UpdateRegionTypeRequest represents request to update a region type
type UpdateRegionTypeRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// RegionTypeResponse represents region type response
type RegionTypeResponse struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RegionTypeListResponse represents list of region types with pagination
// Deprecated: Use StandardListResponse[RegionTypeResponse] instead
type RegionTypeListResponse = StandardListResponse[RegionTypeResponse]

// Region DTOs

// CreateRegionRequest represents request to create a new region
type CreateRegionRequest struct {
	Code           string     `json:"code" validate:"required,min=2,max=20,snake_case"`
	CountryID      *uuid.UUID `json:"country_id,omitempty"`
	CountryCode    *string    `json:"country_code,omitempty" validate:"omitempty,min=2,max=10"`
	RegionTypeCode *string    `json:"region_type_code,omitempty" validate:"omitempty,min=2,max=20"`
	Name           string     `json:"name" validate:"required,min=2,max=100"`
	IsActive       *bool      `json:"is_active,omitempty"`
}

// UpdateRegionRequest represents request to update a region
type UpdateRegionRequest struct {
	CountryID      *uuid.UUID `json:"country_id,omitempty"`
	CountryCode    *string    `json:"country_code,omitempty" validate:"omitempty,min=2,max=10"`
	RegionTypeCode *string    `json:"region_type_code,omitempty" validate:"omitempty,min=2,max=20"`
	Name           *string    `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	IsActive       *bool      `json:"is_active,omitempty"`
}

// RegionResponse represents region response
type RegionResponse struct {
	ID             uuid.UUID           `json:"id"`
	Code           string              `json:"code"`
	CountryID      *uuid.UUID          `json:"country_id,omitempty"`
	CountryCode    *string             `json:"country_code,omitempty"`
	RegionTypeCode *string             `json:"region_type_code,omitempty"`
	Name           string              `json:"name"`
	IsActive       bool                `json:"is_active"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	Country        *CountryResponse    `json:"country,omitempty"`
	RegionType     *RegionTypeResponse `json:"region_type,omitempty"`
}

// RegionListResponse represents list of regions with pagination
// Deprecated: Use StandardListResponse[RegionResponse] instead
type RegionListResponse = StandardListResponse[RegionResponse]

// District DTOs

// CreateDistrictRequest represents request to create a new district
type CreateDistrictRequest struct {
	Code        string     `json:"code" validate:"required,min=2,max=20,snake_case"`
	RegionID    *uuid.UUID `json:"region_id,omitempty"`
	RegionCode  *string    `json:"region_code,omitempty" validate:"omitempty,min=2,max=20"`
	CountryID   *uuid.UUID `json:"country_id,omitempty"`
	CountryCode *string    `json:"country_code,omitempty" validate:"omitempty,min=2,max=10"`
	Name        string     `json:"name" validate:"required,min=2,max=100"`
	IsActive    *bool      `json:"is_active,omitempty"`
}

// UpdateDistrictRequest represents request to update a district
type UpdateDistrictRequest struct {
	RegionID    *uuid.UUID `json:"region_id,omitempty"`
	RegionCode  *string    `json:"region_code,omitempty" validate:"omitempty,min=2,max=20"`
	CountryID   *uuid.UUID `json:"country_id,omitempty"`
	CountryCode *string    `json:"country_code,omitempty" validate:"omitempty,min=2,max=10"`
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	IsActive    *bool      `json:"is_active,omitempty"`
}

// DistrictResponse represents district response
type DistrictResponse struct {
	ID          uuid.UUID        `json:"id"`
	Code        string           `json:"code"`
	RegionID    *uuid.UUID       `json:"region_id,omitempty"`
	RegionCode  *string          `json:"region_code,omitempty"`
	CountryID   *uuid.UUID       `json:"country_id,omitempty"`
	CountryCode *string          `json:"country_code,omitempty"`
	Name        string           `json:"name"`
	IsActive    bool             `json:"is_active"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Region      *RegionResponse  `json:"region,omitempty"`
	Country     *CountryResponse `json:"country,omitempty"`
}

// DistrictListResponse represents list of districts with pagination
// Deprecated: Use StandardListResponse[DistrictResponse] instead
type DistrictListResponse = StandardListResponse[DistrictResponse]

// City DTOs

// CreateCityRequest represents request to create a new city
type CreateCityRequest struct {
	Code         string     `json:"code" validate:"required,min=2,max=20"`
	RegionID     *uuid.UUID `json:"region_id,omitempty"`
	RegionCode   *string    `json:"region_code,omitempty" validate:"omitempty,min=2,max=20"`
	CountryID    *uuid.UUID `json:"country_id,omitempty"`
	CountryCode  *string    `json:"country_code,omitempty" validate:"omitempty,min=2,max=10"`
	DistrictID   *uuid.UUID `json:"district_id,omitempty"`
	DistrictCode *string    `json:"district_code,omitempty" validate:"omitempty,min=2,max=20"`
	Name         string     `json:"name" validate:"required,min=2,max=100"`
	IsActive     *bool      `json:"is_active,omitempty"`
}

// UpdateCityRequest represents request to update a city
type UpdateCityRequest struct {
	RegionID     *uuid.UUID `json:"region_id,omitempty"`
	RegionCode   *string    `json:"region_code,omitempty" validate:"omitempty,min=2,max=20"`
	CountryID    *uuid.UUID `json:"country_id,omitempty"`
	CountryCode  *string    `json:"country_code,omitempty" validate:"omitempty,min=2,max=10"`
	DistrictID   *uuid.UUID `json:"district_id,omitempty"`
	DistrictCode *string    `json:"district_code,omitempty" validate:"omitempty,min=2,max=20"`
	Name         *string    `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	IsActive     *bool      `json:"is_active,omitempty"`
}

// CityResponse represents city response
type CityResponse struct {
	ID           uuid.UUID         `json:"id"`
	Code         string            `json:"code"`
	RegionID     *uuid.UUID        `json:"region_id,omitempty"`
	RegionCode   *string           `json:"region_code,omitempty"`
	CountryID    *uuid.UUID        `json:"country_id,omitempty"`
	CountryCode  *string           `json:"country_code,omitempty"`
	DistrictID   *uuid.UUID        `json:"district_id,omitempty"`
	DistrictCode *string           `json:"district_code,omitempty"`
	Name         string            `json:"name"`
	IsActive     bool              `json:"is_active"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Region       *RegionResponse   `json:"region,omitempty"`
	Country      *CountryResponse  `json:"country,omitempty"`
	District     *DistrictResponse `json:"district,omitempty"`
}

// CityListResponse represents list of cities with pagination
// Deprecated: Use StandardListResponse[CityResponse] instead
type CityListResponse = StandardListResponse[CityResponse]

// Area DTOs

// CreateAreaRequest represents request to create a new area
type CreateAreaRequest struct {
	Code     string     `json:"code" validate:"required,min=2,max=20"`
	CityID   *uuid.UUID `json:"city_id,omitempty"`
	CityCode *string    `json:"city_code,omitempty" validate:"omitempty,min=2,max=20"`
	Name     string     `json:"name" validate:"required,min=2,max=100"`
	IsActive *bool      `json:"is_active,omitempty"`
}

// UpdateAreaRequest represents request to update an area
type UpdateAreaRequest struct {
	CityID   *uuid.UUID `json:"city_id,omitempty"`
	CityCode *string    `json:"city_code,omitempty" validate:"omitempty,min=2,max=20"`
	Name     *string    `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	IsActive *bool      `json:"is_active,omitempty"`
}

// AreaResponse represents area response
type AreaResponse struct {
	ID        uuid.UUID     `json:"id"`
	Code      string        `json:"code"`
	CityID    *uuid.UUID    `json:"city_id,omitempty"`
	CityCode  *string       `json:"city_code,omitempty"`
	Name      string        `json:"name"`
	IsActive  bool          `json:"is_active"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	City      *CityResponse `json:"city,omitempty"`
}

// AreaListResponse represents list of areas with pagination
// Deprecated: Use StandardListResponse[AreaResponse] instead
type AreaListResponse = StandardListResponse[AreaResponse]

// PostalCode DTOs

// CreatePostalCodeRequest represents request to create a new postal code
type CreatePostalCodeRequest struct {
	Code          string     `json:"code" validate:"required,min=3,max=20"`
	CountryID     *uuid.UUID `json:"country_id,omitempty"`
	CountryCode   *string    `json:"country_code,omitempty" validate:"omitempty,min=2,max=10"`
	RegionID      *uuid.UUID `json:"region_id,omitempty"`
	RegionCode    *string    `json:"region_code,omitempty" validate:"omitempty,min=2,max=20"`
	CityID        *uuid.UUID `json:"city_id,omitempty"`
	CityCode      *string    `json:"city_code,omitempty" validate:"omitempty,min=2,max=20"`
	AreaID        *uuid.UUID `json:"area_id,omitempty"`
	AreaCode      *string    `json:"area_code,omitempty" validate:"omitempty,min=2,max=20"`
	LocationScope *string    `json:"location_scope,omitempty" validate:"omitempty,oneof=country region city area"`
	IsActive      *bool      `json:"is_active,omitempty"`
}

// UpdatePostalCodeRequest represents request to update a postal code
type UpdatePostalCodeRequest struct {
	CountryID     *uuid.UUID `json:"country_id,omitempty"`
	CountryCode   *string    `json:"country_code,omitempty" validate:"omitempty,min=2,max=10"`
	RegionID      *uuid.UUID `json:"region_id,omitempty"`
	RegionCode    *string    `json:"region_code,omitempty" validate:"omitempty,min=2,max=20"`
	CityID        *uuid.UUID `json:"city_id,omitempty"`
	CityCode      *string    `json:"city_code,omitempty" validate:"omitempty,min=2,max=20"`
	AreaID        *uuid.UUID `json:"area_id,omitempty"`
	AreaCode      *string    `json:"area_code,omitempty" validate:"omitempty,min=2,max=20"`
	LocationScope *string    `json:"location_scope,omitempty" validate:"omitempty,oneof=country region city area"`
	IsActive      *bool      `json:"is_active,omitempty"`
}

// PostalCodeResponse represents postal code response
type PostalCodeResponse struct {
	ID            uuid.UUID        `json:"id"`
	Code          string           `json:"code"`
	CountryID     *uuid.UUID       `json:"country_id,omitempty"`
	CountryCode   *string          `json:"country_code,omitempty"`
	RegionID      *uuid.UUID       `json:"region_id,omitempty"`
	RegionCode    *string          `json:"region_code,omitempty"`
	CityID        *uuid.UUID       `json:"city_id,omitempty"`
	CityCode      *string          `json:"city_code,omitempty"`
	AreaID        *uuid.UUID       `json:"area_id,omitempty"`
	AreaCode      *string          `json:"area_code,omitempty"`
	LocationScope *string          `json:"location_scope,omitempty"`
	IsActive      bool             `json:"is_active"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	Country       *CountryResponse `json:"country,omitempty"`
	Region        *RegionResponse  `json:"region,omitempty"`
	City          *CityResponse    `json:"city,omitempty"`
	Area          *AreaResponse    `json:"area,omitempty"`
}

// PostalCodeListResponse represents list of postal codes with pagination
// Deprecated: Use StandardListResponse[PostalCodeResponse] instead
type PostalCodeListResponse = StandardListResponse[PostalCodeResponse]

// LocationType DTOs

// CreateLocationTypeRequest represents request to create a new location type
type CreateLocationTypeRequest struct {
	Code        string  `json:"code" validate:"required,min=2,max=20,snake_case"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// UpdateLocationTypeRequest represents request to update a location type
type UpdateLocationTypeRequest struct {
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// LocationTypeResponse represents location type response
type LocationTypeResponse struct {
	Code        string  `json:"code"`
	Description *string `json:"description,omitempty"`
}

// LocationTypeListResponse represents list of location types with pagination
// Deprecated: Use StandardListResponse[LocationTypeResponse] instead
type LocationTypeListResponse = StandardListResponse[LocationTypeResponse]

// LocationAlias DTOs

// CreateLocationAliasRequest represents request to create a new location alias
type CreateLocationAliasRequest struct {
	EntityTypeCode *string   `json:"entity_type_code,omitempty" validate:"omitempty,min=2,max=20"`
	EntityType     *string   `json:"entity_type,omitempty" validate:"omitempty,min=2,max=20"`
	EntityID       uuid.UUID `json:"entity_id" validate:"required"`
	EntityCode     *string   `json:"entity_code,omitempty" validate:"omitempty,min=2,max=20"`
	AliasName      string    `json:"alias_name" validate:"required,min=1,max=100"`
	IsPrimary      *bool     `json:"is_primary,omitempty"`
	IsActive       *bool     `json:"is_active,omitempty"`
}

// UpdateLocationAliasRequest represents request to update a location alias
type UpdateLocationAliasRequest struct {
	EntityTypeCode *string `json:"entity_type_code,omitempty" validate:"omitempty,min=2,max=20"`
	EntityType     *string `json:"entity_type,omitempty" validate:"omitempty,min=2,max=20"`
	EntityCode     *string `json:"entity_code,omitempty" validate:"omitempty,min=2,max=20"`
	AliasName      *string `json:"alias_name,omitempty" validate:"omitempty,min=1,max=100"`
	IsPrimary      *bool   `json:"is_primary,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
}

// LocationAliasResponse represents location alias response
type LocationAliasResponse struct {
	ID             uuid.UUID             `json:"id"`
	EntityTypeCode *string               `json:"entity_type_code,omitempty"`
	EntityType     *string               `json:"entity_type,omitempty"`
	EntityID       uuid.UUID             `json:"entity_id"`
	EntityCode     *string               `json:"entity_code,omitempty"`
	AliasName      string                `json:"alias_name"`
	IsPrimary      bool                  `json:"is_primary"`
	IsActive       bool                  `json:"is_active"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	LocationType   *LocationTypeResponse `json:"location_type,omitempty"`
}

// LocationAliasListResponse represents list of location aliases with pagination
// Deprecated: Use StandardListResponse[LocationAliasResponse] instead
type LocationAliasListResponse = StandardListResponse[LocationAliasResponse]

// ValidationErrorResponse represents validation error details
type ValidationErrorResponse struct {
	Error  string            `json:"error"`
	Code   string            `json:"code"`
	Fields map[string]string `json:"fields"`
}
