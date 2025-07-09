package dtos

import (
	"time"
)

// GeoLocation DTOs

// CreateGeoLocationRequest represents request to create a new geo location
type CreateGeoLocationRequest struct {
	PostalCode       string     `json:"postal_code" validate:"required"`
	Name             string     `json:"name" validate:"required,min=1,max=200"`
	AsciiName        string     `json:"ascii_name" validate:"required,min=1,max=200"`
	AlternateNames   *string    `json:"alternate_names,omitempty"`
	Latitude         float64    `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude        float64    `json:"longitude" validate:"required,min=-180,max=180"`
	FeatureClass     string     `json:"feature_class" validate:"required,len=1"`
	FeatureCode      string     `json:"feature_code" validate:"required,min=1,max=10"`
	CountryCode      string     `json:"country_code" validate:"required,len=2"`
	CC2              *string    `json:"cc2,omitempty"`
	Admin1Code       *string    `json:"admin1_code,omitempty"`
	Admin2Code       *string    `json:"admin2_code,omitempty"`
	Admin3Code       *string    `json:"admin3_code,omitempty"`
	Admin4Code       *string    `json:"admin4_code,omitempty"`
	Population       *int64     `json:"population,omitempty" validate:"omitempty,min=0"`
	Elevation        *int       `json:"elevation,omitempty"`
	DEM              *int       `json:"dem,omitempty"`
	Timezone         *string    `json:"timezone,omitempty"`
	ModificationDate *time.Time `json:"modification_date,omitempty"`
}

// UpdateGeoLocationRequest represents request to update a geo location
type UpdateGeoLocationRequest struct {
	Name           *string  `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	AsciiName      *string  `json:"ascii_name,omitempty" validate:"omitempty,min=1,max=200"`
	AlternateNames *string  `json:"alternate_names,omitempty"`
	Latitude       *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude      *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	FeatureClass   *string  `json:"feature_class,omitempty" validate:"omitempty,len=1"`
	FeatureCode    *string  `json:"feature_code,omitempty" validate:"omitempty,min=1,max=10"`
	CountryCode    *string  `json:"country_code,omitempty" validate:"omitempty,len=2"`
	CC2            *string  `json:"cc2,omitempty"`
	Admin1Code     *string  `json:"admin1_code,omitempty"`
	Admin2Code     *string  `json:"admin2_code,omitempty"`
	Admin3Code     *string  `json:"admin3_code,omitempty"`
	Admin4Code     *string  `json:"admin4_code,omitempty"`
	Population     *int64   `json:"population,omitempty" validate:"omitempty,min=0"`
	Elevation      *int     `json:"elevation,omitempty"`
	DEM            *int     `json:"dem,omitempty"`
	Timezone       *string  `json:"timezone,omitempty"`
}

// GeoLocationResponse represents geo location response
type GeoLocationResponse struct {
	PostalCode       string    `json:"postal_code"`
	Name             string    `json:"name"`
	AsciiName        string    `json:"ascii_name"`
	AlternateNames   *string   `json:"alternate_names,omitempty"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	FeatureClass     string    `json:"feature_class"`
	FeatureCode      string    `json:"feature_code"`
	CountryCode      string    `json:"country_code"`
	CC2              *string   `json:"cc2,omitempty"`
	Admin1Code       *string   `json:"admin1_code,omitempty"`
	Admin2Code       *string   `json:"admin2_code,omitempty"`
	Admin3Code       *string   `json:"admin3_code,omitempty"`
	Admin4Code       *string   `json:"admin4_code,omitempty"`
	Population       *int64    `json:"population,omitempty"`
	Elevation        *int      `json:"elevation,omitempty"`
	DEM              *int      `json:"dem,omitempty"`
	Timezone         *string   `json:"timezone,omitempty"`
	ModificationDate time.Time `json:"modification_date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// GeoLocationFilters represents basic filters for geo location queries
type GeoLocationFilters struct {
	CountryCode *string  `json:"country_code,omitempty" query:"country_code"`
	PostalCodes []string `json:"postal_codes,omitempty" query:"postal_code"`
	Name        *string  `json:"name,omitempty" query:"name"`
	FeatureCode *string  `json:"feature_code,omitempty" query:"feature_code"`
}

// GeoLocationListResponse represents list of geo locations with pagination
type GeoLocationListResponse = StandardListResponse[GeoLocationResponse]

// GeoLocationSingleResponse represents single geo location response
type GeoLocationSingleResponse = StandardSingleResponse[GeoLocationResponse]

// GeoLocationSearchFilters represents search filters for geo location queries
type GeoLocationSearchFilters struct {
	CountryCode   *string  `json:"country_code,omitempty" query:"country_code"`
	FeatureClass  *string  `json:"feature_class,omitempty" query:"feature_class"`
	FeatureCode   *string  `json:"feature_code,omitempty" query:"feature_code"`
	Admin1Code    *string  `json:"admin1_code,omitempty" query:"admin1_code"`
	Admin2Code    *string  `json:"admin2_code,omitempty" query:"admin2_code"`
	Admin3Code    *string  `json:"admin3_code,omitempty" query:"admin3_code"`
	Admin4Code    *string  `json:"admin4_code,omitempty" query:"admin4_code"`
	MinLatitude   *float64 `json:"min_latitude,omitempty" query:"min_latitude"`
	MaxLatitude   *float64 `json:"max_latitude,omitempty" query:"max_latitude"`
	MinLongitude  *float64 `json:"min_longitude,omitempty" query:"min_longitude"`
	MaxLongitude  *float64 `json:"max_longitude,omitempty" query:"max_longitude"`
	MinPopulation *int64   `json:"min_population,omitempty" query:"min_population"`
	MaxPopulation *int64   `json:"max_population,omitempty" query:"max_population"`
	MinElevation  *int     `json:"min_elevation,omitempty" query:"min_elevation"`
	MaxElevation  *int     `json:"max_elevation,omitempty" query:"max_elevation"`
	Timezone      *string  `json:"timezone,omitempty" query:"timezone"`
	NameSearch    *string  `json:"name_search,omitempty" query:"name_search"`
	Offset        int      `json:"offset" query:"offset" validate:"min=0"`
	Limit         int      `json:"limit" query:"limit" validate:"min=1,max=100"`
}

// GeoLocationSearchResponse represents search results for geo locations
type GeoLocationSearchResponse struct {
	Data       []GeoLocationResponse `json:"data"`
	Pagination PaginationResponse    `json:"pagination"`
}

// GeoLocationStatsResponse represents statistics for geo locations
type GeoLocationStatsResponse struct {
	TotalCount        int64                    `json:"total_count"`
	CountryStats      map[string]int64         `json:"country_stats"`
	FeatureClassStats map[string]int64         `json:"feature_class_stats"`
	FeatureCodeStats  map[string]int64         `json:"feature_code_stats"`
	TimezoneStats     map[string]int64         `json:"timezone_stats"`
	AverageLatitude   float64                  `json:"average_latitude"`
	AverageLongitude  float64                  `json:"average_longitude"`
	PopulationStats   *PopulationStatsResponse `json:"population_stats,omitempty"`
	ElevationStats    *ElevationStatsResponse  `json:"elevation_stats,omitempty"`
}

// PopulationStatsResponse represents population statistics
type PopulationStatsResponse struct {
	Total    int64   `json:"total"`
	Average  float64 `json:"average"`
	Min      int64   `json:"min"`
	Max      int64   `json:"max"`
	WithData int64   `json:"with_data"`
}

// ElevationStatsResponse represents elevation statistics
type ElevationStatsResponse struct {
	Average  float64 `json:"average"`
	Min      int     `json:"min"`
	Max      int     `json:"max"`
	WithData int64   `json:"with_data"`
}
