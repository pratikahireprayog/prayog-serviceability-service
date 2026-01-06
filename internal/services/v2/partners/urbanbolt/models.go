package urbanbolt

// AuthRequest represents the authentication request to UrbanBolt
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response from UrbanBolt
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"` // in seconds
	TokenType   string `json:"token_type,omitempty"`
	Expires     string `json:"expires,omitempty"`     // ISO timestamp when token expires
	Status      string `json:"status,omitempty"`     // Response status (e.g., "Success")
}

// ServiceabilityRequest represents a request to UrbanBolt's serviceability API
type ServiceabilityRequest struct {
	Pincodes []string `json:"pincodes"` // Comma-separated pincodes as query parameter
}

// ServiceabilityResponse represents a response from UrbanBolt's serviceability API
type ServiceabilityResponse struct {
	Status        string                `json:"status"`
	Message       string                `json:"message"`
	Data          []PincodeServiceability `json:"data"`
	ErrorPincodes []string              `json:"errorPincodes"`
}

// PincodeServiceability represents serviceability data for a pincode
type PincodeServiceability struct {
	ID            int    `json:"id"`
	Pincode       int    `json:"pincode"`
	Inbound       bool   `json:"inbound"`
	Outbound      bool   `json:"outbound"`
	RTN           bool   `json:"rtn"`
	IsActive      bool   `json:"isActive"`
	ServiceCenter string `json:"serviceCenter"`
	City          string `json:"city"`
	State         string `json:"state"`
	Region        string `json:"region"`
	Zone          string `json:"zone"`
	RouteCode     string `json:"routeCode"`
	ServiceType   string `json:"serviceType"` // Comma-separated service types like "SDD,NDD,ATA,PTP,2HR,IMP"
}

