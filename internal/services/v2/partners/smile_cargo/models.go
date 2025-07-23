package smile_cargo

// ServiceAvailabilityRequest represents request to Smile Cargo service availability API
type ServiceAvailabilityRequest struct {
	FromPincode string `json:"fromPincode" validate:"required"`
	ToPincode   string `json:"toPincode" validate:"required"`
	// Remove VendorCode as per new business requirement
	// VendorCode  string `json:"vendorCode" validate:"required"`
}

// ServiceAvailabilityResponse represents response from Smile Cargo API
type ServiceAvailabilityResponse struct {
	Data []ServiceAvailabilityData `json:"data"`
}

// ServiceAvailabilityData represents individual data entry in the response
type ServiceAvailabilityData struct {
	FromPincode     *int             `json:"fromPincode,omitempty"`
	ToPincode       *int             `json:"toPincode,omitempty"`
	BigShipPartners []BigShipPartner `json:"bigShipPartners"`
	SmilePartner    *SmilePartner    `json:"smilePartner"`
}

// BigShipPartner represents big ship partner information (ignored if smilePartner is null)
type BigShipPartner struct {
	PartnerName string `json:"partnerName"`
	CourierType string `json:"courierType"` // "Surface", "Air"
	CourierID   int    `json:"courierId"`
	Prepaid     bool   `json:"prepaid"`
	COD         bool   `json:"cod"`
	Delivery    bool   `json:"delivery"`
	Pickup      bool   `json:"pickup"`
	Status      bool   `json:"status"`
}

// SmilePartner represents Smile partner information (null means not serviceable)
type SmilePartner struct {
	Status           bool   `json:"status"`
	City             string `json:"city"`
	CityName         string `json:"cityName"`
	Zone             string `json:"zone"`
	AreaAvailability string `json:"areaAvailability"`
	LastMile         bool   `json:"lastMile"`
	FirstMile        bool   `json:"firstMile"`
	COD              bool   `json:"cod"`
	ToPay            bool   `json:"toPay"`
}

// SmileCargoConfig represents configuration for Smile Cargo adapter
type SmileCargoConfig struct {
	BaseURL     string  `json:"base_url"`
	ServiceURL  string  `json:"service_url"`
	VendorCode  string  `json:"vendor_code"`
	Timeout     int     `json:"timeout"`
	MaxRetries  int     `json:"max_retries"`
	RetryDelay  int     `json:"retry_delay"`
	Enabled     bool    `json:"enabled"`
	Rating      float64 `json:"rating"`
	MinWeightKG float64 `json:"min_weight_kg"`
	MaxWeightKG float64 `json:"max_weight_kg"`
}
