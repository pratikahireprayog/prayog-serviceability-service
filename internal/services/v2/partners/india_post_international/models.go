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
type TariffRequest struct {
	ProductType         string `json:"productType"`         // e.g., "FGN_LETTER"
	Weight              int    `json:"weight"`              // in grams
	CountryCode         string `json:"countryCode"`         // destination country code
	Registration        bool   `json:"registration"`        // registration required
	Insurance           bool   `json:"insurance"`           // insurance required
	InsAmount           int    `json:"insAmount"`           // insurance amount
	AdviceOfDelivery    bool   `json:"adviceOfDelivery"`    // advice of delivery
	ModeOfTransmission  string `json:"modeOfTransmission"`  // e.g., "AMS"
	SourcePincode       string `json:"sourcePincode"`       // source pincode
}

// TariffResponse represents the India Post tariff response
type TariffResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
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

