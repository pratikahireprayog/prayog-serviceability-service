package smile_courier

// ServiceabilityRequest represents a request to Smile Courier's serviceability API
type ServiceabilityRequest struct {
	FromPincode string `json:"from_pincode,omitempty"`
	ToPincode   string `json:"to_pincode"`
	CountryCode string `json:"country_code"`
}

// ServiceabilityResponse represents a response from Smile Courier's serviceability API
type ServiceabilityResponse struct {
	Success bool                `json:"success"`
	Data    *ServiceabilityData `json:"data,omitempty"`
	Error   *SmileCourierError  `json:"error,omitempty"`
}

// ServiceabilityData contains the actual serviceability information
type ServiceabilityData struct {
	IsServiceable  bool                  `json:"is_serviceable"`
	Services       []SmileCourierService `json:"services"`
	AvailableModes []string              `json:"available_modes"`
	PaymentModes   []string              `json:"payment_modes"`
	MaxWeight      float64               `json:"max_weight"`
	ServiceZones   []string              `json:"service_zones"`
}

// SmileCourierService represents a shipping service offered by Smile Courier
type SmileCourierService struct {
	ServiceCode        string                      `json:"service_code"`
	ServiceName        string                      `json:"service_name"`
	DeliveryDays       int                         `json:"delivery_days"`
	CODSupported       bool                        `json:"cod_supported"`
	PickupAvailable    bool                        `json:"pickup_available"`
	DeliveryAvailable  bool                        `json:"delivery_available"`
	InsuranceAvailable bool                        `json:"insurance_available"`
	DocumentsSupported bool                        `json:"documents_supported"`
	ExpressAvailable   bool                        `json:"express_available"`
	Pricing            *SmileCourierServicePricing `json:"pricing,omitempty"`
}

// SmileCourierServicePricing represents pricing information
type SmileCourierServicePricing struct {
	BaseCost      float64 `json:"base_cost"`
	Currency      string  `json:"currency"`
	CODCharges    float64 `json:"cod_charges"`
	FuelSurcharge float64 `json:"fuel_surcharge"`
}

// SmileCourierError represents an error response from Smile Courier API
type SmileCourierError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
