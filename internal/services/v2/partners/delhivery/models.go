package delhivery

// ServiceabilityRequest represents a request to Delhivery's serviceability API
type ServiceabilityRequest struct {
	OriginPin      string `json:"origin_pin"`      // Source pincode
	DestinationPin string `json:"destination_pin"` // Destination pincode
	MOT            string `json:"mot"`            // Mode of transport (S for surface, A for air)
}

// ServiceabilityResponse represents a response from Delhivery's serviceability API
type ServiceabilityResponse struct {
	Success bool              `json:"success"`
	Msg     string            `json:"msg"`
	Data    ServiceabilityData `json:"data"`
}

// ServiceabilityData represents the serviceability data in the response
type ServiceabilityData struct {
	TAT int `json:"tat"` // Turnaround time in days
}

