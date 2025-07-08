package dhl

// ServiceabilityRequest represents a request to DHL's serviceability API
type ServiceabilityRequest struct {
	OriginCountryCode      string  `json:"origin_country_code"`
	OriginPostalCode       string  `json:"origin_postal_code"`
	DestinationCountryCode string  `json:"destination_country_code"`
	DestinationPostalCode  string  `json:"destination_postal_code"`
	ShipmentDate           string  `json:"shipment_date,omitempty"`
	ProductType            string  `json:"product_type,omitempty"` // EXPRESS, ECONOMY, etc.
	Weight                 float64 `json:"weight,omitempty"`       // in KG
	DeclaredValue          float64 `json:"declared_value,omitempty"`
	Currency               string  `json:"currency,omitempty"`
}

// ServiceabilityResponse represents a response from DHL's serviceability API
type ServiceabilityResponse struct {
	Success bool                   `json:"success"`
	Data    *DHLServiceabilityData `json:"data,omitempty"`
	Error   *DHLError              `json:"error,omitempty"`
}

// DHLServiceabilityData contains the actual serviceability information
type DHLServiceabilityData struct {
	IsServiceable       bool                `json:"is_serviceable"`
	Services            []DHLService        `json:"services"`
	TransitTime         *DHLTransitTime     `json:"transit_time,omitempty"`
	ServiceCapabilities map[string]bool     `json:"service_capabilities"`
	Restrictions        []string            `json:"restrictions,omitempty"`
	CountryPairInfo     *DHLCountryPairInfo `json:"country_pair_info,omitempty"`
}

// DHLService represents a shipping service offered by DHL
type DHLService struct {
	ProductCode             string                 `json:"product_code"` // EXPRESS_WORLDWIDE, ECONOMY_SELECT, etc.
	ProductName             string                 `json:"product_name"`
	ServiceType             string                 `json:"service_type"`       // EXPRESS, ECONOMY, GROUND
	EstimatedDelivery       string                 `json:"estimated_delivery"` // Date string
	TransitDays             int                    `json:"transit_days"`
	IsDocumentsSupported    bool                   `json:"documents_supported"`
	IsNonDocumentsSupported bool                   `json:"non_documents_supported"`
	MaxWeight               float64                `json:"max_weight"`
	TrackingSupported       bool                   `json:"tracking_supported"`
	SignatureRequired       bool                   `json:"signature_required"`
	Pricing                 *DHLServicePricing     `json:"pricing,omitempty"`
	DeliveryCommitment      *DHLDeliveryCommitment `json:"delivery_commitment,omitempty"`
}

// DHLServicePricing represents pricing information for DHL services
type DHLServicePricing struct {
	BaseCost            float64               `json:"base_cost"`
	Currency            string                `json:"currency"`
	FuelSurcharge       float64               `json:"fuel_surcharge"`
	SecuritySurcharge   float64               `json:"security_surcharge"`
	RemoteAreaSurcharge float64               `json:"remote_area_surcharge"`
	DutiesAndTaxes      float64               `json:"duties_and_taxes,omitempty"`
	AdditionalCharges   []DHLAdditionalCharge `json:"additional_charges,omitempty"`
	TotalCost           float64               `json:"total_cost"`
}

// DHLAdditionalCharge represents additional charges
type DHLAdditionalCharge struct {
	ChargeType  string  `json:"charge_type"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

// DHLDeliveryCommitment represents delivery commitment information
type DHLDeliveryCommitment struct {
	DeliveryDate string `json:"delivery_date"`
	DeliveryTime string `json:"delivery_time,omitempty"`
	CutoffTime   string `json:"cutoff_time,omitempty"`
}

// DHLTransitTime represents transit time information
type DHLTransitTime struct {
	MinDays int `json:"min_days"`
	MaxDays int `json:"max_days"`
}

// DHLCountryPairInfo represents information about origin-destination country pair
type DHLCountryPairInfo struct {
	OriginCountry      string   `json:"origin_country"`
	DestinationCountry string   `json:"destination_country"`
	IsInternational    bool     `json:"is_international"`
	TimeZoneDifference int      `json:"timezone_difference"` // in hours
	SupportedProducts  []string `json:"supported_products"`
}

// DHLError represents an error response from DHL API
type DHLError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	Field   string `json:"field,omitempty"`
}

// DHLAuthRequest represents authentication request for DHL API
type DHLAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// DHLAuthResponse represents authentication response from DHL API
type DHLAuthResponse struct {
	Success     bool   `json:"success"`
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"` // in seconds
	Error       string `json:"error,omitempty"`
}
