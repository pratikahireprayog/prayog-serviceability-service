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
}

type Shipper struct {
	Address Address `json:"address"`
}

type Recipient struct {
	Address Address `json:"address"`
}

type Address struct {
	PostalCode  string `json:"postalCode"`
	CountryCode string `json:"countryCode"`
}

type RequestedPackageLineItem struct {
	Weight Weight `json:"weight"`
}
type FedExAlert struct {
    Code      string `json:"code"`
    Message   string `json:"message"`
    AlertType string `json:"alertType"`
}

type ShipmentRate struct {
    RateZone         string       `json:"rateZone"`
    DimDivisor       int          `json:"dimDivisor"`
    FuelSurchargePercent float64  `json:"fuelSurchargePercent"`
    TotalSurcharges  float64      `json:"totalSurcharges"`
    TotalFreightDiscount float64   `json:"totalFreightDiscount"`
    SurCharges       []Surcharge  `json:"surCharges"`
    PricingCode      string       `json:"pricingCode"`
    TotalBillingWeight Weight      `json:"totalBillingWeight"`
    Currency         string       `json:"currency"`
    RateScale        string       `json:"rateScale"`
}

type RatedPackage struct {
    GroupNumber            int          `json:"groupNumber"`
    EffectiveNetDiscount   float64      `json:"effectiveNetDiscount"`
    PackageRateDetail      PackageRate  `json:"packageRateDetail"`
}

type PackageRate struct {
    RateType            string   `json:"rateType"`
    RatedWeightMethod   string   `json:"ratedWeightMethod"`
    BaseCharge          float64  `json:"baseCharge"`
    NetFreight          float64  `json:"netFreight"`
    TotalSurcharges     float64  `json:"totalSurcharges"`
    NetFedExCharge      float64  `json:"netFedExCharge"`
    TotalTaxes          float64  `json:"totalTaxes"`
    NetCharge           float64  `json:"netCharge"`
    BillingWeight       Weight   `json:"billingWeight"`
    Currency            string   `json:"currency"`
    Surcharges          []Surcharge `json:"surcharges"`
    TotalFreightDiscounts float64 `json:"totalFreightDiscounts"`
}

type Weight struct {
    Units string  `json:"units"`
    Value float64 `json:"value"`
}

type Surcharge struct {
    Type        string  `json:"type"`
    Description string  `json:"description"`
    Amount      float64 `json:"amount"`
}

type OperationalDetail struct {
    IneligibleForMoneyBackGuarantee bool   `json:"ineligibleForMoneyBackGuarantee"`
    AstraDescription                string `json:"astraDescription"`
    AirportID                       string `json:"airportId"`
    ServiceCode                      string `json:"serviceCode"`
}

type ServiceDescription struct {
    ServiceID       string `json:"serviceId"`
    ServiceType     string `json:"serviceType"`
    Code            string `json:"code"`
    Description     string `json:"description"`
    AstraDescription string `json:"astraDescription"`
}


type Charge struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type Alert struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	AlertType string `json:"alertType"`
}

type APIError struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Parameter []string `json:"parameter,omitempty"`
}

// ServiceabilityRequest for serviceability check
type ServiceabilityRequest struct {
	OriginAddress      Address `json:"origin_address"`
	DestinationAddress Address `json:"destination_address"`
	Weight             float64 `json:"weight"`
}

// ServiceabilityResponse for serviceability check
type ServiceabilityResponse struct {
	Serviceable       bool      `json:"serviceable"`
	AvailableServices []Service `json:"available_services,omitempty"`
	Errors            []Error   `json:"errors,omitempty"`
}

type Service struct {
	ServiceType string  `json:"service_type"`
	ServiceName string  `json:"service_name"`
	Cost        float64 `json:"cost,omitempty"`
	Currency    string  `json:"currency,omitempty"`
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

// TransitTimeRequest represents a request to FedEx's transit times API
type TransitTimeRequest struct {
	RequestedShipment TransitTimeShipment `json:"requestedShipment"`
	CarrierCodes      []string           `json:"carrierCodes"`
}

type TransitTimeShipment struct {
	Shipper                   Shipper                   `json:"shipper"`
	Recipients                []Recipient              `json:"recipients"`
	PackagingType             string                   `json:"packagingType"`
	CustomsClearanceDetail    *CustomsClearanceDetail  `json:"customsClearanceDetail,omitempty"`
	RequestedPackageLineItems []RequestedPackageLineItem `json:"requestedPackageLineItems"`
}

type CustomsClearanceDetail struct {
	Commodities []Commodity `json:"commodities"`
}

type Commodity struct {
	Description   string      `json:"description"`
	CustomsValue  CustomsValue `json:"customsValue"`
	NumberOfPieces int        `json:"numberOfPieces"`
}

type CustomsValue struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// TransitTimeResponse represents the response from FedEx's transit times API
type TransitTimeResponse struct {
	TransactionID string          `json:"transactionId"`
	Output        TransitOutput   `json:"output"`
}

type TransitOutput struct {
	TransitTimes []TransitTime `json:"transitTimes"`
	Encoded      bool          `json:"encoded"`
}

type TransitTime struct {
	Alerts            []TransitAlert        `json:"alerts"`
	TransitTimeDetails []TransitTimeDetail  `json:"transitTimeDetails"`
}

type TransitAlert struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	AlertType string `json:"alertType"`
}

type TransitTimeDetail struct {
	ServiceType        string             `json:"serviceType"`
	ServiceName        string             `json:"serviceName"`
	Commit             *CommitDetail      `json:"commit,omitempty"`
	CustomerMessages   []CustomerMessage  `json:"customerMessages"`
}

type CommitDetail struct {
	DateDetail                 DateDetail                 `json:"dateDetail"`
	CommodityName             string                     `json:"commodityName"`
	DerivedDestinationDetail  DerivedDestinationDetail   `json:"derivedDestinationDetail"`
	CutOffTime                string                     `json:"cutOffTime"`
	AccessTime                AccessTime                 `json:"accessTime"`
}

type DateDetail struct {
	DayOfWeek string `json:"dayOfWeek"`
	Time      string `json:"time"`
	Day       string `json:"day"`
}

type DerivedDestinationDetail struct {
	CountryCode       string `json:"countryCode"`
	StateOrProvinceCode string `json:"stateOrProvinceCode"`
	PostalCode        string `json:"postalCode"`
	ServiceArea       string `json:"serviceArea"`
	LocationId        string `json:"locationId"`
	LocationNumber    int    `json:"locationNumber"`
	AirportId         string `json:"airportId"`
}

type AccessTime struct {
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
}

type CustomerMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}