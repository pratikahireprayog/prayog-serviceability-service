package delcaper

// LoginRequest represents the login request to Delcaper API
type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	VendorType string `json:"vendorType"`
}

// LoginResponse represents the login response from Delcaper API
type LoginResponse struct {
	Data   LoginData `json:"data"`
	Status int       `json:"status"`
}

// LoginData contains the login response data
type LoginData struct {
	UserDto           UserDto `json:"userDto"`
	ExpiresIn         string  `json:"expiresIn"`
	AccessToken       string  `json:"accessToken"`
	RefreshTokenExpiresIn string `json:"refreshTokenExpiresIn"`
	RefreshToken      string  `json:"refreshToken"`
}

// UserDto contains user information
type UserDto struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	VendorCode  string `json:"vendorCode"`
	VendorType  string `json:"vendorType"`
	Mobile      string `json:"mobile"`
	IsActive    bool   `json:"isActive"`
	AuthType    string `json:"authType"`
	KycStatus   string `json:"kycStatus"`
	PaymentType string `json:"paymentType"`
}

// CheckFeasibleRequest represents the serviceability check request
type CheckFeasibleRequest struct {
	OrderType      string        `json:"orderType"`
	PickupAddress  AddressInfo   `json:"pickupAddress"`
	ShippingAddress AddressInfo  `json:"shippingAddress"`
}

// AddressInfo represents address information
type AddressInfo struct {
	Zip string `json:"zip"`
}

// CheckFeasibleResponse represents the serviceability check response
type CheckFeasibleResponse struct {
	Data   FeasibleData `json:"data"`
	Status int          `json:"status"`
}

// FeasibleData contains the feasibility check response data
type FeasibleData struct {
	IsFeasible bool     `json:"isFeasible"`
	Services   []Service `json:"services,omitempty"`
	Message    string   `json:"message,omitempty"`
}

// Service represents a delivery service
type Service struct {
	ServiceCode string  `json:"serviceCode"`
	ServiceName string  `json:"serviceName"`
	TATDays     int     `json:"tatDays"`
	IsCOD       bool    `json:"isCod"`
	Pickup      bool    `json:"pickup"`
	Delivery    bool    `json:"delivery"`
	Insurance   bool    `json:"insurance"`
	Pricing     *Pricing `json:"pricing,omitempty"`
}

// Pricing represents service pricing information
type Pricing struct {
	BaseCost      float64 `json:"baseCost"`
	Currency      string  `json:"currency"`
	CODCharges    float64 `json:"codCharges"`
	FuelSurcharge float64 `json:"fuelSurcharge"`
} 