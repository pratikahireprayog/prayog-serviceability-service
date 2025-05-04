package models

// ServiceabilityRequest represents a request to check if a location is serviceable.
type ServiceabilityRequest struct {
	PostalCode      string `json:"postal_code"`
	CountryCode     string `json:"country_code,omitempty"`
	ServiceTypeCode string `json:"service_type_code,omitempty"`
	OrderTypeCode   string `json:"order_type_code,omitempty"`
}

// ServiceabilityResponse represents the response to a serviceability check.
type ServiceabilityResponse struct {
	PostalCode    string             `json:"postal_code"`
	IsServiceable bool               `json:"is_serviceable"`
	Services      []ServiceAvailable `json:"services,omitempty"`
	Message       string             `json:"message,omitempty"`
}

// ServiceAvailable represents the availability status of a service.
type ServiceAvailable struct {
	ServiceTypeCode string `json:"service_type_code"`
	ServiceTypeName string `json:"service_type_name"`
	IsAvailable     bool   `json:"is_available"`
}

// BulkServiceabilityRequest represents a request to check serviceability for multiple postal codes.
type BulkServiceabilityRequest struct {
	Requests []ServiceabilityRequest `json:"requests"`
}

// BulkServiceabilityResponse represents the response to a bulk serviceability check.
type BulkServiceabilityResponse struct {
	Results []ServiceabilityResponse `json:"results"`
}
