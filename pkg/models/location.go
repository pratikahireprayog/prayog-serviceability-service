package models

// Location represents a geographic location.
type Location struct {
	ID         string `json:"id"`
	PostalCode string `json:"postal_code"`
	City       string `json:"city"`
	Region     string `json:"region"`
	Country    string `json:"country"`
}

// ServiceabilityResult represents the result of a serviceability check.
type ServiceabilityResult struct {
	Location      *Location `json:"location"`
	IsServiceable bool      `json:"is_serviceable"`
	ServiceType   string    `json:"service_type,omitempty"`
	OrderType     string    `json:"order_type,omitempty"`
	Message       string    `json:"message,omitempty"`
}
