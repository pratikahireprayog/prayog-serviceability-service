package fedex

import (
	"time"
)

// FedExConfig represents the configuration for FedEx adapter
type FedExConfig struct {
	Enabled       bool          `json:"enabled"`
	BaseURL       string        `json:"base_url"`
	ClientID      string        `json:"client_id"`
	ClientSecret  string        `json:"client_secret"`
	AccountNumber string        `json:"account_number"`
	Timeout       time.Duration `json:"timeout"`
}

// DefaultFedExConfig returns default configuration for FedEx
func DefaultFedExConfig() FedExConfig {
	return FedExConfig{
		Enabled:       true,
		BaseURL:       "https://apis-sandbox.fedex.com",
		ClientID:      "your_client_id",
		ClientSecret:  "your_client_secret",
		AccountNumber: "510087020",
		Timeout:       30 * time.Second,
	}
}

// RateRequest represents a request to FedEx's rates API
type RateRequest struct {
	AccountNumber     AccountNumber     `json:"accountNumber"`
	RequestedShipment RequestedShipment `json:"requestedShipment"`
}

type AccountNumber struct {
	Value string `json:"value"`
}

type RequestedShipment struct {
	Shipper                   Shipper                   `json:"shipper"`
	Recipient                 Recipient                 `json:"recipient"`
	PickupType                string                    `json:"pickupType"`
	RateRequestType           []string                  `json:"rateRequestType"`
	RequestedPackageLineItems []RequestedPackageLineItem `json:"requestedPackageLineItems"`
	ShipDate                  string                    `json:"shipDate,omitempty"`
	TotalWeight               *Weight                   `json:"totalWeight,omitempty"`
	PackageCount              int                       `json:"packageCount,omitempty"`
}

type Shipper struct {
	Address Address `json:"address"`
}

type Recipient struct {
	Address Address `json:"address"`
}

type Address struct {
	StreetLines         []string `json:"streetLines,omitempty"`
	City                string   `json:"city"`
	StateOrProvinceCode string   `json:"stateOrProvinceCode,omitempty"`
	PostalCode          string   `json:"postalCode"`
	CountryCode         string   `json:"countryCode"`
	Residential         bool     `json:"residential,omitempty"`
}

type RequestedPackageLineItem struct {
	Weight       Weight      `json:"weight"`
	Dimensions   Dimensions  `json:"dimensions,omitempty"`
	CustomerReferences []CustomerReference `json:"customerReferences,omitempty"`
	Description  string      `json:"description,omitempty"`
}

type Weight struct {
	Units  string  `json:"units"`
	Value  float64 `json:"value"`
}

type Dimensions struct {
	Length int    `json:"length"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Units  string `json:"units"`
}

type CustomerReference struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// RateResponse represents a response from FedEx's rates API
type RateResponse struct {
	TransactionID string `json:"transactionId"`
	Output        Output `json:"output"`
}

type Output struct {
	RateReplyDetails []RateReplyDetail `json:"rateReplyDetails"`
	Quota            []interface{}     `json:"quota,omitempty"`
	Alerts           []Alert           `json:"alerts,omitempty"`
	Errors           []APIError        `json:"errors,omitempty"`
}

type RateReplyDetail struct {
	ServiceType               string                    `json:"serviceType"`
	ServiceName               string                    `json:"serviceName"`
	PackagingType             string                    `json:"packagingType"`
	CustomerMessages          []CustomerMessage         `json:"customerMessages,omitempty"`
	RatedShipmentDetails      []RatedShipmentDetail     `json:"ratedShipmentDetails"`
	OperationalDetail         *OperationalDetail        `json:"operationalDetail,omitempty"`
	SignatureOptionType       string                    `json:"signatureOptionType,omitempty"`
	ServiceDescription        *ServiceDescription       `json:"serviceDescription,omitempty"`
	CommitedDetail            *CommitedDetail           `json:"committedDetail,omitempty"`
}

type CustomerMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type RatedShipmentDetail struct {
	RateType           string    `json:"rateType"`
	TotalNetCharge     *Charge   `json:"totalNetCharge"`
	TotalBaseCharge    *Charge   `json:"totalBaseCharge"`
	TotalNetFedExCharge *Charge  `json:"totalNetFedExCharge,omitempty"`
}

type Charge struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type OperationalDetail struct {
	OriginLocationID       string `json:"originLocationId"`
	DeliveryDay            string `json:"deliveryDay,omitempty"`
	TransitTime            string `json:"transitTime,omitempty"`
	IneligibleForMoneyBack bool   `json:"ineligibleForMoneyBack"`
}

type ServiceDescription struct {
	ServiceType    string `json:"serviceType"`
	Code           string `json:"code"`
	Names          []Name `json:"names"`
	Description    string `json:"description"`
	AstraDescription string `json:"astraDescription"`
}

type Name struct {
	Type string `json:"type"`
	Encoding string `json:"encoding"`
	Value string `json:"value"`
}

type CommitedDetail struct {
	Date               string `json:"date"`
	DayOfWeek          int    `json:"dayOfWeek"`
	TransitTime        string `json:"transitTime"`
	CommitTimestamp    string `json:"commitTimestamp,omitempty"`
	DerivedDestination string `json:"derivedDestination,omitempty"`
}

type Alert struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	AlertType   string `json:"alertType"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Parameter []string `json:"parameter,omitempty"`
}

// ServiceabilityRequest for serviceability check
type ServiceabilityRequest struct {
	OriginAddress      Address `json:"origin_address"`
	DestinationAddress Address `json:"destination_address"`
	Weight             float64 `json:"weight,omitempty"`
	ShipDate           string  `json:"ship_date,omitempty"`
}

// ServiceabilityResponse for serviceability check
type ServiceabilityResponse struct {
	Serviceable        bool      `json:"serviceable"`
	AvailableServices  []Service `json:"available_services,omitempty"`
	Restrictions       []string  `json:"restrictions,omitempty"`
	Errors             []Error   `json:"errors,omitempty"`
}

type Service struct {
	ServiceType   string  `json:"service_type"`
	ServiceName   string  `json:"service_name"`
	TransitDays   int     `json:"transit_days"`
	DeliveryDate  string  `json:"delivery_date,omitempty"`
	Cost          float64 `json:"cost,omitempty"`
	Currency      string  `json:"currency,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TokenResponse for OAuth authentication
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}