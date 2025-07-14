package dhl

// RatesRequest represents a request to DHL's rates API
type RatesRequest struct {
	CustomerDetails            CustomerDetails       `json:"customerDetails"`
	Accounts                   []Account             `json:"accounts"`
	ProductsAndServices        []ProductAndService   `json:"productsAndServices"`
	PayerCountryCode           string                `json:"payerCountryCode"`
	PlannedShippingDateAndTime string                `json:"plannedShippingDateAndTime"`
	UnitOfMeasurement          string                `json:"unitOfMeasurement"`
	IsCustomsDeclarable        bool                  `json:"isCustomsDeclarable"`
	EstimatedDeliveryDate      EstimatedDeliveryDate `json:"estimatedDeliveryDate"`
	ReturnStandardProductsOnly bool                  `json:"returnStandardProductsOnly"`
	Packages                   []Package             `json:"packages"`
}

// CustomerDetails represents customer information
type CustomerDetails struct {
	ShipperDetails  ShipperDetails  `json:"shipperDetails"`
	ReceiverDetails ReceiverDetails `json:"receiverDetails"`
}

// ShipperDetails represents shipper information
type ShipperDetails struct {
	PostalCode  string `json:"postalCode"`
	CityName    string `json:"cityName"`
	CountryCode string `json:"countryCode"`
}

// ReceiverDetails represents receiver information
type ReceiverDetails struct {
	PostalCode  string `json:"postalCode"`
	CityName    string `json:"cityName"`
	CountryCode string `json:"countryCode"`
}

// Account represents account information
type Account struct {
	TypeCode string `json:"typeCode"`
	Number   string `json:"number"`
}

// ProductAndService represents product and service information
type ProductAndService struct {
	ProductCode      string `json:"productCode"`
	LocalProductCode string `json:"localProductCode"`
}

// EstimatedDeliveryDate represents delivery date request
type EstimatedDeliveryDate struct {
	IsRequested bool   `json:"isRequested"`
	TypeCode    string `json:"typeCode"`
}

// Package represents package information
type Package struct {
	Weight     float64    `json:"weight"`
	Dimensions Dimensions `json:"dimensions"`
}

// Dimensions represents package dimensions
type Dimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// RatesResponse represents a response from DHL's rates API
type RatesResponse struct {
	Products      []Product      `json:"products"`
	ExchangeRates []ExchangeRate `json:"exchangeRates"`
}

// Product represents a shipping product from DHL
type Product struct {
	ProductName             string                   `json:"productName"`
	ProductCode             string                   `json:"productCode"`
	LocalProductCode        string                   `json:"localProductCode"`
	LocalProductCountryCode string                   `json:"localProductCountryCode"`
	NetworkTypeCode         string                   `json:"networkTypeCode"`
	IsCustomerAgreement     bool                     `json:"isCustomerAgreement"`
	Weight                  WeightInfo               `json:"weight"`
	TotalPrice              []TotalPrice             `json:"totalPrice"`
	TotalPriceBreakdown     []TotalPriceBreakdown    `json:"totalPriceBreakdown"`
	DetailedPriceBreakdown  []DetailedPriceBreakdown `json:"detailedPriceBreakdown"`
	PickupCapabilities      PickupCapabilities       `json:"pickupCapabilities"`
	DeliveryCapabilities    DeliveryCapabilities     `json:"deliveryCapabilities"`
	PricingDate             string                   `json:"pricingDate"`
}

// WeightInfo represents weight information
type WeightInfo struct {
	Volumetric        float64 `json:"volumetric"`
	Provided          float64 `json:"provided"`
	UnitOfMeasurement string  `json:"unitOfMeasurement"`
}

// TotalPrice represents total price information
type TotalPrice struct {
	CurrencyType  string  `json:"currencyType"`
	PriceCurrency string  `json:"priceCurrency"`
	Price         float64 `json:"price"`
}

// TotalPriceBreakdown represents total price breakdown
type TotalPriceBreakdown struct {
	CurrencyType   string           `json:"currencyType"`
	PriceCurrency  string           `json:"priceCurrency"`
	PriceBreakdown []PriceBreakdown `json:"priceBreakdown"`
}

// PriceBreakdown represents price breakdown details
type PriceBreakdown struct {
	TypeCode string  `json:"typeCode"`
	Price    float64 `json:"price"`
}

// DetailedPriceBreakdown represents detailed price breakdown
type DetailedPriceBreakdown struct {
	CurrencyType  string      `json:"currencyType"`
	PriceCurrency string      `json:"priceCurrency"`
	Breakdown     []Breakdown `json:"breakdown"`
}

// Breakdown represents breakdown details
type Breakdown struct {
	Name                string                   `json:"name"`
	ServiceCode         string                   `json:"serviceCode,omitempty"`
	LocalServiceCode    string                   `json:"localServiceCode,omitempty"`
	ServiceTypeCode     string                   `json:"serviceTypeCode,omitempty"`
	Price               float64                  `json:"price"`
	IsCustomerAgreement bool                     `json:"isCustomerAgreement,omitempty"`
	IsMarketedService   bool                     `json:"isMarketedService,omitempty"`
	PriceBreakdown      []DetailedPriceBreakdown `json:"priceBreakdown,omitempty"`
}

// PickupCapabilities represents pickup capabilities
type PickupCapabilities struct {
	NextBusinessDay                       bool   `json:"nextBusinessDay"`
	LocalCutoffDateAndTime                string `json:"localCutoffDateAndTime"`
	PickupEarliest                        string `json:"pickupEarliest"`
	PickupLatest                          string `json:"pickupLatest"`
	PickupCutoffSameDayOutboundProcessing string `json:"pickupCutoffSameDayOutboundProcessing"`
	OriginServiceAreaCode                 string `json:"originServiceAreaCode"`
	OriginFacilityAreaCode                string `json:"originFacilityAreaCode"`
	PickupAdditionalDays                  int    `json:"pickupAdditionalDays"`
	PickupDayOfWeek                       int    `json:"pickupDayOfWeek"`
}

// DeliveryCapabilities represents delivery capabilities
type DeliveryCapabilities struct {
	DeliveryTypeCode             string `json:"deliveryTypeCode"`
	EstimatedDeliveryDateAndTime string `json:"estimatedDeliveryDateAndTime"`
	DestinationServiceAreaCode   string `json:"destinationServiceAreaCode"`
	DestinationFacilityAreaCode  string `json:"destinationFacilityAreaCode"`
	DeliveryAdditionalDays       int    `json:"deliveryAdditionalDays"`
	DeliveryDayOfWeek            int    `json:"deliveryDayOfWeek"`
	TotalTransitDays             int    `json:"totalTransitDays"`
}

// ExchangeRate represents currency exchange rate
type ExchangeRate struct {
	CurrentExchangeRate float64 `json:"currentExchangeRate"`
	Currency            string  `json:"currency"`
	BaseCurrency        string  `json:"baseCurrency"`
}

// ServiceabilityCapabilities represents the expected response format
type ServiceabilityCapabilities struct {
	PartnerID    string       `json:"partner_id"`
	PartnerCode  string       `json:"partner_code"`
	Capabilities Capabilities `json:"capabilities"`
}

// Capabilities represents capabilities structure
type Capabilities struct {
	PickupCapabilities   PickupCapabilitiesV2   `json:"pickup_capabilities"`
	DeliveryCapabilities DeliveryCapabilitiesV2 `json:"delivery_capabilities"`
}

// PickupCapabilitiesV2 represents pickup capabilities in the expected format
type PickupCapabilitiesV2 struct {
	NextBusinessDay                       bool   `json:"next_business_day"`
	LocalCutoffDateAndTime                string `json:"local_cutoff_date_and_time"`
	PickupEarliest                        string `json:"pickup_earliest"`
	PickupLatest                          string `json:"pickup_latest"`
	PickupCutoffSameDayOutboundProcessing string `json:"pickup_cutoff_same_day_outbound_processing"`
	OriginServiceAreaCode                 string `json:"origin_service_area_code"`
	OriginFacilityAreaCode                string `json:"origin_facility_area_code"`
	PickupAdditionalDays                  int    `json:"pickup_additional_days"`
	PickupDayOfWeek                       int    `json:"pickup_day_of_week"`
}

// DeliveryCapabilitiesV2 represents delivery capabilities in the expected format
type DeliveryCapabilitiesV2 struct {
	DeliveryTypeCode             string `json:"delivery_type_code"`
	EstimatedDeliveryDateAndTime string `json:"estimated_delivery_date_and_time"`
	DestinationServiceAreaCode   string `json:"destination_service_area_code"`
	DestinationFacilityAreaCode  string `json:"destination_facility_area_code"`
	DeliveryAdditionalDays       int    `json:"delivery_additional_days"`
	DeliveryDayOfWeek            int    `json:"delivery_day_of_week"`
	TotalTransitDays             int    `json:"total_transit_days"`
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

// Legacy models for backward compatibility
type ServiceabilityRequest struct {
	OriginCountryCode      string  `json:"origin_country_code"`
	OriginPostalCode       string  `json:"origin_postal_code"`
	DestinationCountryCode string  `json:"destination_country_code"`
	DestinationPostalCode  string  `json:"destination_postal_code"`
	ShipmentDate           string  `json:"shipment_date,omitempty"`
	ProductType            string  `json:"product_type,omitempty"`
	Weight                 float64 `json:"weight,omitempty"`
	DeclaredValue          float64 `json:"declared_value,omitempty"`
	Currency               string  `json:"currency,omitempty"`
}

type ServiceabilityResponse struct {
	Success bool                   `json:"success"`
	Data    *DHLServiceabilityData `json:"data,omitempty"`
	Error   *DHLError              `json:"error,omitempty"`
}

type DHLServiceabilityData struct {
	IsServiceable       bool                `json:"is_serviceable"`
	Services            []DHLService        `json:"services"`
	TransitTime         *DHLTransitTime     `json:"transit_time,omitempty"`
	ServiceCapabilities map[string]bool     `json:"service_capabilities"`
	Restrictions        []string            `json:"restrictions,omitempty"`
	CountryPairInfo     *DHLCountryPairInfo `json:"country_pair_info,omitempty"`
}

type DHLService struct {
	ProductCode             string                 `json:"product_code"`
	ProductName             string                 `json:"product_name"`
	ServiceType             string                 `json:"service_type"`
	EstimatedDelivery       string                 `json:"estimated_delivery"`
	TransitDays             int                    `json:"transit_days"`
	IsDocumentsSupported    bool                   `json:"documents_supported"`
	IsNonDocumentsSupported bool                   `json:"non_documents_supported"`
	MaxWeight               float64                `json:"max_weight"`
	TrackingSupported       bool                   `json:"tracking_supported"`
	SignatureRequired       bool                   `json:"signature_required"`
	Pricing                 *DHLServicePricing     `json:"pricing,omitempty"`
	DeliveryCommitment      *DHLDeliveryCommitment `json:"delivery_commitment,omitempty"`
}

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

type DHLAdditionalCharge struct {
	ChargeType  string  `json:"charge_type"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

type DHLDeliveryCommitment struct {
	DeliveryDate string `json:"delivery_date"`
	DeliveryTime string `json:"delivery_time,omitempty"`
	CutoffTime   string `json:"cutoff_time,omitempty"`
}

type DHLTransitTime struct {
	MinDays int `json:"min_days"`
	MaxDays int `json:"max_days"`
}

type DHLCountryPairInfo struct {
	OriginCountry      string   `json:"origin_country"`
	DestinationCountry string   `json:"destination_country"`
	IsInternational    bool     `json:"is_international"`
	TimeZoneDifference int      `json:"timezone_difference"`
	SupportedProducts  []string `json:"supported_products"`
}
