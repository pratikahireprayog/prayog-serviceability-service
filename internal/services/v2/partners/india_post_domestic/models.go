package india_post_domestic

import "time"

// PincodeSearchRequest represents a request to India Post Domestic pincode search API
type PincodeSearchRequest struct {
	Pincode    string `json:"pincode"`
	Limit      int    `json:"limit"`
	OfficeType string `json:"office_type"` // "post" for postal offices
}

// PincodeSearchResponse represents the response from India Post Domestic API
type PincodeSearchResponse struct {
	StatusCode           int            `json:"status_code"`
	Success              bool           `json:"success"`
	Message              string         `json:"message"`
	Skip                 int            `json:"skip"`
	Limit                int            `json:"limit"`
	ReturnedRecordsCount int            `json:"returned_records_count"`
	Data                 []PostalOffice `json:"data"`
}

// PostalOffice represents a postal office in India Post Domestic system
type PostalOffice struct {
	Pincode            int    `json:"pincode"`
	OfficeName         string `json:"office_name"`
	OfficeID           string `json:"office_id"`
	OfficeTypeCode     string `json:"office_type_code"`
	StateName          string `json:"state_name"`
	DeliveryOfficeFlag bool   `json:"delivery_office_flag"`
	CityName           string `json:"city_name"`
	TalukName          string `json:"taluk_name"`
	VillageName        string `json:"village_name"`
	IsRolledOut        bool   `json:"is_rolled_out"`
	SpdsID             string `json:"spds_id"`
	IdcID              int    `json:"idc_id"`
}

// AuthRequest represents authentication request to India Post Domestic
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	ClientID string `json:"client_id"`
}

// AuthResponse represents authentication response from India Post Domestic
type AuthResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"` // in seconds
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
	GrantType    string `json:"grant_type"` // "refresh_token"
}

// TokenInfo holds token information
type TokenInfo struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string
}

// IsExpired checks if the token is expired
func (t *TokenInfo) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return true
	}
	// Add 5 minute buffer before actual expiry
	return time.Now().Add(5 * time.Minute).After(t.ExpiresAt)
}

// ErrorResponse represents an error response from India Post Domestic API
type ErrorResponse struct {
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Error      string `json:"error,omitempty"`
}

