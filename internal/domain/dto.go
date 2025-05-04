package domain

import "time"

// APIRequest is a base type for all API requests
type APIRequest struct {
	RequestID string    `json:"request_id,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// APIResponse is a base type for all API responses
type APIResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// APIServiceabilityRequest represents a request to check service availability via the API
type APIServiceabilityRequest struct {
	APIRequest
	Location    string `json:"location"`     // Postal code or other location identifier
	ServiceType string `json:"service_type"` // Type of service requested
	OrderType   string `json:"order_type"`   // Type of order
}

// APIServiceabilityResponse represents the response to a serviceability check via the API
type APIServiceabilityResponse struct {
	APIResponse
	IsServiceable bool   `json:"is_serviceable"`
	Location      string `json:"location"`
	ServiceType   string `json:"service_type"`
	OrderType     string `json:"order_type"`
}

// APIBulkServiceabilityRequest represents multiple serviceability check requests via the API
type APIBulkServiceabilityRequest struct {
	APIRequest
	Requests []APIServiceabilityRequest `json:"requests"`
}

// APIBulkServiceabilityResponse represents multiple serviceability check responses via the API
type APIBulkServiceabilityResponse struct {
	APIResponse
	Results []APIServiceabilityResponse `json:"results"`
}

// APIServiceAvailabilityRequest represents a request to check service availability for a specific location via the API
type APIServiceAvailabilityRequest struct {
	APIRequest
	LocationType  string `json:"location_type"`   // COUNTRY, REGION, CITY, AREA, POSTAL_CODE
	LocationID    string `json:"location_id"`     // ID of the location (UUID as string)
	OrderTypeID   string `json:"order_type_id"`   // ID of the order type (UUID as string)
	ServiceTypeID string `json:"service_type_id"` // ID of the service type (UUID as string)
}

// APIServiceAvailabilityResponse represents the response to a service availability check via the API
type APIServiceAvailabilityResponse struct {
	APIResponse
	Request        APIServiceAvailabilityRequest `json:"request"`
	IsAvailable    bool                          `json:"is_available"`
	EffectiveFrom  time.Time                     `json:"effective_from,omitempty"`
	EffectiveTo    time.Time                     `json:"effective_to,omitempty"`
	AdditionalData string                        `json:"additional_data,omitempty"`
}

// APIGeoResponse represents a generic response for geo-related operations via the API
type APIGeoResponse struct {
	APIResponse
	ID       string `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code,omitempty"`
	IsActive bool   `json:"is_active"`
}

// LocationEntry represents an entry in the location hierarchy
type LocationEntry struct {
	Type     string `json:"type"`      // COUNTRY, REGION, CITY, AREA, POSTAL_CODE
	ID       string `json:"id"`        // ID of the location
	Name     string `json:"name"`      // Name of the location
	Code     string `json:"code"`      // Code for the location
	Priority int    `json:"priority"`  // Priority for serviceability rules (higher is more specific)
	IsActive bool   `json:"is_active"` // Whether the location is active
}

// LocationHierarchy represents a hierarchical view of locations
type LocationHierarchy struct {
	PostalCode string          `json:"postal_code"` // The postal code
	Entries    []LocationEntry `json:"entries"`     // Entries in the hierarchy
}

// ServiceabilityLocationHierarchy represents a complete location hierarchy from postal code to country
type ServiceabilityLocationHierarchy struct {
	PostalCode *PostalCode           `json:"postal_code,omitempty"`
	Area       *Area                 `json:"area,omitempty"`
	City       *City                 `json:"city,omitempty"`
	Region     *AdministrativeRegion `json:"region,omitempty"`
	Country    *Country              `json:"country,omitempty"`

	// IDs for easy reference
	PostalCodeID string `json:"postal_code_id,omitempty"`
	AreaID       string `json:"area_id,omitempty"`
	CityID       string `json:"city_id,omitempty"`
	RegionID     string `json:"region_id,omitempty"`
	CountryID    string `json:"country_id,omitempty"`
}

// ServiceabilityLocationEntry holds a specific location type and its ID for serviceability queries
type ServiceabilityLocationEntry struct {
	LocationType string `json:"location_type"`
	LocationID   string `json:"location_id"`
}

// ServiceabilityLocationEntries represents a list of locations with their types for serviceability checks
type ServiceabilityLocationEntries struct {
	Entries []ServiceabilityLocationEntry `json:"entries"`
}
