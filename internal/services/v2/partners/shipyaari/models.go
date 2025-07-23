package shipyaari

import "time"

// ServiceabilityRequest represents a request to Shipyaari's serviceability API
type ServiceabilityRequest struct {
	FromPincode string  `json:"from_pincode"`
	ToPincode   string  `json:"to_pincode"`
	OrderValue  float64 `json:"order_value"`
	PaymentMode string  `json:"payment_mode"`
	ProductType string  `json:"product_type"`
	Weight      float64 `json:"weight,omitempty"`
	Length      float64 `json:"length,omitempty"`
	Breadth     float64 `json:"breadth,omitempty"`
	Height      float64 `json:"height,omitempty"`
}

// ServiceabilityResponse represents a response from Shipyaari's serviceability API
type ServiceabilityResponse struct {
	IsServiceable        bool               `json:"is_serviceable"`
	ResponseID           string             `json:"response_id"`
	Zone                 string             `json:"zone"`
	Services             []ShipyaariService `json:"services"`
	CODAvailable         bool               `json:"cod_available"`
	PickupAvailable      bool               `json:"pickup_available"`
	ReversePickup        bool               `json:"reverse_pickup"`
	InsuranceAvailable   bool               `json:"insurance_available"`
	ErrorMessage         string             `json:"error_message,omitempty"`
	ExpectedDeliveryDate time.Time          `json:"expected_delivery_date,omitempty"`
	Charges              *ShipyaariCharges  `json:"charges,omitempty"`
}

// ShipyaariService represents a shipping service offered by Shipyaari
type ShipyaariService struct {
	ServiceID    string          `json:"service_id"`
	ServiceName  string          `json:"service_name"`
	ServiceType  string          `json:"service_type"`
	DeliveryTime string          `json:"delivery_time"`
	Cost         float64         `json:"cost"`
	Currency     string          `json:"currency"`
	IsActive     bool            `json:"is_active"`
	Description  string          `json:"description,omitempty"`
	Features     map[string]bool `json:"features,omitempty"`
}

// ShipyaariCharges represents detailed pricing information
type ShipyaariCharges struct {
	BaseCharge    float64 `json:"base_charge"`
	FuelSurcharge float64 `json:"fuel_surcharge"`
	ServiceTax    float64 `json:"service_tax"`
	CODCharges    float64 `json:"cod_charges,omitempty"`
	TotalCharge   float64 `json:"total_charge"`
	Currency      string  `json:"currency"`
}

// AuthRequest represents the authentication request to Shipyaari
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response from Shipyaari
type AuthResponse struct {
	Token        string    `json:"token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Success      bool      `json:"success"`
	Message      string    `json:"message,omitempty"`
}

// ErrorResponse represents an error response from Shipyaari API
type ErrorResponse struct {
	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Details      string `json:"details,omitempty"`
}
