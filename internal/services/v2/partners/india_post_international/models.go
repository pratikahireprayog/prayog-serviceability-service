package india_post_international

// LoginRequest represents the India Post login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the India Post login response (nested under "data")
type LoginResponse struct {
	Success bool `json:"success"`
	Data    struct {
		AccessToken      string `json:"access_token"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
		RefreshToken     string `json:"refresh_token"`
		TokenType        string `json:"token_type"`
		IDToken          string `json:"id_token"`
		SessionState     string `json:"session_state"`
		Scope            string `json:"scope"`
	} `json:"data"`
}

// TariffRequest represents the India Post international tariff calculation request
// According to API documentation: POST /beextcustomer/v1/international-tariff/itps
// Request body: { "weight": 800, "countryCode": "DE", "sourcePincode": "110001" }
type TariffRequest struct {
	Weight        int    `json:"weight"`        // Weight in grams (required)
	CountryCode   string `json:"countryCode"`    // ISO country code of destination (required)
	SourcePincode string `json:"sourcePincode"` // Postal code of origin location (required)
}

// TariffResponse represents the India Post tariff response
// According to API documentation, response format:
// { "tariffAmount": 1500, "currency": "EUR", "deliveryTime": "5-7 business days", "status": "success" }
type TariffResponse struct {
	TariffAmount float64 `json:"tariffAmount"` // The calculated tariff amount
	Currency     string  `json:"currency"`       // Currency in which tariff is expressed
	DeliveryTime string  `json:"deliveryTime"`  // Estimated delivery time frame
	Status       string  `json:"status"`       // Status of the request (e.g., "success")
	
	// Additional fields that might be present in wrapped responses
	Success bool                   `json:"success,omitempty"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// IndiaPostAPIError represents a structured India Post API error
type IndiaPostAPIError struct {
	StatusCode int
	Message    string
	RawBody    string
}

func (e *IndiaPostAPIError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

