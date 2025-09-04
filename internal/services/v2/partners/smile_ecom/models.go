package smile_ecom

// ServiceabilityData represents database query result for Smile Ecom serviceability
type ServiceabilityData struct {
	IsServiceable bool                   `json:"is_serviceable"`
	Services      []SmileEcomService     `json:"services"`
	Capabilities  map[string]interface{} `json:"capabilities"`
}

// SmileEcomService represents a service record from the database
type SmileEcomService struct {
	ServiceCode        string  `json:"service_code" db:"service_code"`
	ServiceName        string  `json:"service_name" db:"service_name"`
	TATDays            int     `json:"tat_days" db:"tat_days"`
	CODAvailable       bool    `json:"cod_available" db:"cod_available"`
	PickupAvailable    bool    `json:"pickup_available" db:"pickup_available"`
	DeliveryAvailable  bool    `json:"delivery_available" db:"delivery_available"`
	InsuranceAvailable bool    `json:"insurance_available" db:"insurance_available"`
	BaseCost           float64 `json:"base_cost" db:"base_cost"`
	Currency           string  `json:"currency" db:"currency"`
	CODCharges         float64 `json:"cod_charges" db:"cod_charges"`
	FuelSurcharge      float64 `json:"fuel_surcharge" db:"fuel_surcharge"`
	IsActive           bool    `json:"is_active" db:"is_active"`
}

// ServiceabilityQuery represents parameters for database queries
type ServiceabilityQuery struct {
	FromPincode string `json:"from_pincode"`
	ToPincode   string `json:"to_pincode"`
	CountryCode string `json:"country_code"`
	TableName   string `json:"table_name"`
}

// CacheKey represents a cache key for Smile Ecom serviceability
type CacheKey struct {
	FromPincode string `json:"from_pincode"`
	ToPincode   string `json:"to_pincode"`
	CountryCode string `json:"country_code"`
}

// DatabaseQueryResult represents the raw result from database query
type DatabaseQueryResult struct {
	ServiceCode        string  `db:"service_code"`
	ServiceName        string  `db:"service_name"`
	TATDays            int     `db:"tat_days"`
	CODAvailable       bool    `db:"cod_available"`
	PickupAvailable    bool    `db:"pickup_available"`
	DeliveryAvailable  bool    `db:"delivery_available"`
	InsuranceAvailable bool    `db:"insurance_available"`
	BaseCost           float64 `db:"base_cost"`
	Currency           string  `db:"currency"`
	CODCharges         float64 `db:"cod_charges"`
	FuelSurcharge      float64 `db:"fuel_surcharge"`
	IsActive           bool    `db:"is_active"`
}
