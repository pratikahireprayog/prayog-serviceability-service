package shipyaari

import "time"

// ServiceabilityRequest represents a request to Shipyaari's serviceability API
type ServiceabilityRequest struct {
	PickupPincode    int     `json:"pickupPincode"`
	DeliveryPincode  int     `json:"deliveryPincode"`
	InvoiceValue     float64 `json:"invoiceValue"`
	PaymentMode      string  `json:"paymentMode"`
	Weight           float64 `json:"weight"`
	OrderType        string  `json:"orderType"`
	Dimension        Dimension `json:"dimension"`
}

// Dimension represents package dimensions
type Dimension struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
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
	PartnerServiceID   string  `json:"partnerServiceId"`
	PartnerServiceName string  `json:"partnerServiceName"`
	CompanyServiceID   string  `json:"companyServiceId"`
	CompanyServiceName string  `json:"companyServiceName"`
	PartnerName        string  `json:"partnerName"`
	ServiceMode        string  `json:"serviceMode"`
	AppliedWeight      float64 `json:"appliedWeight"`
	InvoiceValue       float64 `json:"invoiceValue"`
	CollectableAmount  float64 `json:"collectableAmount"`
	Insurance          float64 `json:"insurance"`
	Base               float64 `json:"base"`
	Add                float64 `json:"add"`
	Variables          float64 `json:"variables"`
	MinChargeableAmount float64 `json:"minchargeableamount"`
	VariableServices   float64 `json:"variableServices"`
	COD                float64 `json:"cod"`
	Tax                float64 `json:"tax"`
	Total              float64 `json:"total"`
	ZoneName           string  `json:"zoneName"`
	EDT                int     `json:"EDT"`
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
	Success    bool      `json:"success"`
	Data       []AuthData `json:"data"`
	Message    string    `json:"message"`
	StatusCode int       `json:"statusCode"`
}

// AuthData represents the authentication data
type AuthData struct {
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	SellerID         int       `json:"sellerId"`
	CompanyID        string    `json:"companyId"`
	PrivateCompanyID int       `json:"privateCompanyId"`
	Token            string    `json:"token"`
	JWT              string    `json:"jwt"`
	PrivateCompany   PrivateCompany `json:"privateCompany"`
	NextStep         NextStep  `json:"nextStep"`
	ContactNumber    int       `json:"contactNumber"`
	IsWalletRecharge bool      `json:"isWalletRechage"`
	IsReturningUser  bool      `json:"isReturningUser"`
	IsMigrated       bool      `json:"isMigrated"`
	PHPUserID        int       `json:"phpUserId"`
	PHPParentID      int       `json:"phpParentId"`
	BusinessType     string    `json:"businessType"`
	KYCDetails       KYCDetails `json:"kycDetails"`
	IsPostpaid       bool      `json:"isPostpaid"`
	IsMaskedUser     bool      `json:"isMaskedUser"`
	IsWalletBlackListed bool   `json:"isWalletBlackListed"`
}

// PrivateCompany represents company information
type PrivateCompany struct {
	Name        string `json:"name"`
	CompanyID   int    `json:"companyId"`
	Address     string `json:"address"`
	Pincode     int    `json:"pincode"`
	City        string `json:"city"`
	State       string `json:"state"`
	LogoURL     string `json:"logoUrl"`
	BrandName   string `json:"brandName"`
	BusinessType string `json:"businessType"`
	Website     string `json:"webSite"`
	FacebookURL string `json:"facebookUrl"`
	InstagramURL string `json:"instagramUrl"`
	WhatsappURL string `json:"whatsappUrl"`
}

// NextStep represents next steps for the user
type NextStep struct {
	QNA                bool `json:"qna"`
	KYC                bool `json:"kyc"`
	Bank               bool `json:"bank"`
	IsChannelIntegrated bool `json:"isChannelIntegrated"`
}

// KYCDetails represents KYC information
type KYCDetails struct {
	GSTNumber     string `json:"gstNumber"`
	GSTVerified   bool   `json:"gstVerified"`
	GSTFile       string `json:"gstFile"`
	PANNumber     string `json:"panNumber"`
	PANVerified   bool   `json:"panVerified"`
	PANFile       string `json:"panFile"`
	AadharNumber  int    `json:"aadharNumber"`
	AadharVerified bool  `json:"aadharVerified"`
	AadharFile    string `json:"aadharFile"`
	Address       Address `json:"address"`
	FullAddress   string `json:"fullAddress"`
	IsKYCDone     bool   `json:"isKYCDone"`
	FullName      string `json:"fullName"`
}

// Address represents address information
type Address struct {
	PlotNumber string `json:"plotNumber"`
	Locality   string `json:"locality"`
	City       string `json:"city"`
	District   string `json:"district"`
	Pincode    int    `json:"pincode"`
	State      string `json:"state"`
	Country    string `json:"country"`
}

// ErrorResponse represents an error response from Shipyaari API
type ErrorResponse struct {
	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Details      string `json:"details,omitempty"`
}
