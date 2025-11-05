package models

import (
	"time"
)

// V2 Request and Response Types for Multi-Partner Serviceability

// Weight represents weight information for packages
type Weight struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// Dimensions represents dimension information for packages
type Dimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"`
}

// Package represents package information for serviceability checks
type Package struct {
	Weight     *Weight     `json:"weight,omitempty" validate:"omitempty"`
	Dimensions *Dimensions `json:"dimensions,omitempty" validate:"omitempty"`
}

// PartnerFilter represents a partner to be queried for serviceability
type PartnerFilter struct {
	ID   *string `json:"id,omitempty"`
	Code string  `json:"code" validate:"required"`
}

// ServiceabilityV2Request represents the request structure for v2 serviceability checks
type ServiceabilityV2Request struct {
	PostalCode            *string         `json:"postal_code,omitempty" validate:"omitempty,min=3,max=10"`
	SourcePostalCode      *string         `json:"source_postal_code,omitempty" validate:"omitempty,min=3,max=10"`
	DestinationPostalCode *string         `json:"destination_postal_code,omitempty" validate:"omitempty,min=3,max=10"`
	SourceLatitude        *float64        `json:"source_latitude,omitempty"`
	SourceLongitude       *float64        `json:"source_longitude,omitempty"`
	DestinationLatitude   *float64        `json:"destination_latitude,omitempty"`
	DestinationLongitude  *float64        `json:"destination_longitude,omitempty"`
	CountryCode           *string         `json:"country_code,omitempty" validate:"omitempty,len=2"`
	SourceCountryCode     *string         `json:"source_country_code,omitempty" validate:"omitempty,len=2"`
	DestinationCountryCode *string        `json:"destination_country_code,omitempty" validate:"omitempty,len=2"`
	ParcelCategory        *string         `json:"parcel_category,omitempty" validate:"omitempty,oneof=ecomm courier cargo international hyperlocal"`
	ProductType           *string         `json:"product_type,omitempty"`
	Packages              []Package       `json:"packages,omitempty" validate:"omitempty,dive"`
	Partners              []PartnerFilter `json:"partners,omitempty" validate:"omitempty,dive"`
}

// AddressInfo represents address information in the V2 response
type AddressInfo struct {
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

// DetailedAddress represents detailed address information for hubs
type DetailedAddress struct {
	Type        string   `json:"type"` // e.g., "INTERNATIONAL_HUB_ADDRESS"
	Zip         string   `json:"zip"`
	Name        string   `json:"name"`
	Phone       string   `json:"phone"`
	Email       string   `json:"email"`
	Street      string   `json:"street"`
	Landmark    string   `json:"landmark"`
	City        string   `json:"city"`
	State       string   `json:"state"`
	Country     string   `json:"country"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	AddressName string   `json:"addressName"` // e.g., "WAREHOUSE"
}

// ServiceabilityV2Response represents the aggregated response structure for v2
type ServiceabilityV2Response struct {
	Success            bool                `json:"success"`
	Message            string              `json:"message,omitempty"`
	SourceAddress      *AddressInfo        `json:"source_address,omitempty"`
	DestinationAddress *AddressInfo        `json:"destination_address,omitempty"`
	Addresses          []DetailedAddress   `json:"addresses,omitempty"`
    HubDetails         interface{}         `json:"hub_details,omitempty"`
	Partners           []PartnerV2Response `json:"partners"` // Remove omitempty to always include partners array
	Error              *ErrorResponse      `json:"error,omitempty"`
	Metadata           *V2ResponseMetadata `json:"metadata,omitempty"`
}

// PartnerV2Response represents individual partner response in v2 format
type PartnerV2Response struct {
	PartnerID       string                 `json:"partner_id"`
	PartnerCode     string                 `json:"partner_code"`
	PartnerName     string                 `json:"partner_name,omitempty"`
	Rating          float64                `json:"rating"`
	Source          string                 `json:"source,omitempty"` // "real_time", "cache", etc.
	IsServiceable   bool                   `json:"is_serviceable"`
	Services        []ServiceV2            `json:"standard_services,omitempty"`
	PartnerServices interface{}            `json:"services,omitempty"`
	Capabilities    map[string]interface{} `json:"capabilities,omitempty"`
	Error           *string                `json:"error,omitempty"`
	ResponseTime    time.Duration          `json:"response_time,omitempty"`
	ResponseTimeMs  int64                  `json:"response_time_ms,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	HubDetails      interface{}            `json:"hub_details,omitempty"`
}

// ServiceV2 represents service information for v2 responses
type ServiceV2 struct {
	ServiceCode   string            `json:"service_code"`
	ServiceName   string            `json:"service_name"`
	TATDays       int               `json:"tat_days"`
	IsCOD         bool              `json:"is_cod"`
	Pickup        bool              `json:"pickup"`
	Delivery      bool              `json:"delivery"`
	Insurance     bool              `json:"insurance"`
	ProductTypes  map[string]bool   `json:"product_types"`
	DeliveryModes map[string]bool   `json:"delivery_modes"`
	Pricing       *ServicePricingV2 `json:"pricing,omitempty"`
	Rate          *Rate             `json:"rate,omitempty"`
}

// ServicePricingV2 represents pricing information
type ServicePricingV2 struct {
	BaseCost      float64 `json:"base_cost"`
	Currency      string  `json:"currency"`
	CODCharges    float64 `json:"cod_charges,omitempty"`
	FuelSurcharge float64 `json:"fuel_surcharge,omitempty"`
}

// V2ResponseMetadata contains metadata about the v2 response
type V2ResponseMetadata struct {
	RequestID                string        `json:"request_id,omitempty"`
	ResponseTimeMs           int64         `json:"response_time_ms,omitempty"`
	PartnersQueried          int           `json:"partners_queried,omitempty"`
	PartnersSucceeded        int           `json:"partners_succeeded,omitempty"`
	PartnersFailed           int           `json:"partners_failed,omitempty"`
	TotalServiceablePartners int           `json:"total_serviceable_partners,omitempty"`
	RatesIncluded            bool          `json:"rates_included,omitempty"`
	TotalPartners            int           `json:"total_partners"`
	ServiceableCount         int           `json:"serviceable_count"`
	ProcessingTime           time.Duration `json:"processing_time"`
	Filters                  V2Filters     `json:"filters"`
}

// V2Filters represents the filters applied to determine eligible partners
type V2Filters struct {
	ParcelCategory    *string         `json:"parcel_category,omitempty"`
	ProductType       *string         `json:"product_type,omitempty"`
	CountryCode       *string         `json:"country_code,omitempty"`
	RequestedPartners []PartnerFilter `json:"requested_partners,omitempty"`
}

// BulkServiceabilityV2Request represents bulk requests for v2
type BulkServiceabilityV2Request struct {
	Requests []ServiceabilityV2Request `json:"requests" validate:"required,min=1,max=100,dive"`
}

// BulkServiceabilityV2Response represents bulk responses for v2
type BulkServiceabilityV2Response struct {
	Success  bool                       `json:"success"`
	Data     []ServiceabilityV2Response `json:"data,omitempty"`
	Error    *ErrorResponse             `json:"error,omitempty"`
	Metadata *BulkV2Metadata            `json:"metadata,omitempty"`
}

// BulkV2Metadata contains metadata for bulk v2 operations
type BulkV2Metadata struct {
	TotalRequests   int           `json:"total_requests"`
	SuccessfulCount int           `json:"successful_count"`
	FailedCount     int           `json:"failed_count"`
	ProcessingTime  time.Duration `json:"processing_time"`
}

// Partner-specific capability structures for V2 responses

// SmileCargoLocationCapability represents location-specific capabilities for Smile Cargo
type SmileCargoLocationCapability struct {
	Status    bool `json:"status"`
	LastMile  bool `json:"lastMile"`
	FirstMile bool `json:"firstMile"`
	COD       bool `json:"cod"`
	ToPay     bool `json:"toPay"`
}

// SmileCargoCapabilities represents the capabilities structure for Smile Cargo
type SmileCargoCapabilities struct {
	SourcePostalCode      *SmileCargoLocationCapability `json:"source_postal_code,omitempty"`
	DestinationPostalCode *SmileCargoLocationCapability `json:"destination_postal_code,omitempty"`
}

// DHLPickupCapabilities represents pickup capabilities for DHL
// TODO: Implement when DHL API is integrated
type DHLPickupCapabilities struct {
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

// DHLDeliveryCapabilities represents delivery capabilities for DHL
// TODO: Implement when DHL API is integrated
type DHLDeliveryCapabilities struct {
	DeliveryTypeCode             string `json:"delivery_type_code"`
	EstimatedDeliveryDateAndTime string `json:"estimated_delivery_date_and_time"`
	DestinationServiceAreaCode   string `json:"destination_service_area_code"`
	DestinationFacilityAreaCode  string `json:"destination_facility_area_code"`
	DeliveryAdditionalDays       int    `json:"delivery_additional_days"`
	DeliveryDayOfWeek            int    `json:"delivery_day_of_week"`
	TotalTransitDays             int    `json:"total_transit_days"`
}

// DHLCapabilities represents the capabilities structure for DHL
// TODO: Implement when DHL API is integrated
type DHLCapabilities struct {
	PickupCapabilities   *DHLPickupCapabilities   `json:"pickup_capabilities,omitempty"`
	DeliveryCapabilities *DHLDeliveryCapabilities `json:"delivery_capabilities,omitempty"`
}

// ShipyaariService represents a service offered by Shipyaari
// TODO: Implement when Shipyaari API is integrated
type ShipyaariService struct {
	PartnerServiceID    string  `json:"partner_service_id"`
	PartnerServiceName  string  `json:"partner_service_name"`
	CompanyServiceID    string  `json:"company_service_id"`
	CompanyServiceName  string  `json:"company_service_name"`
	PartnerName         string  `json:"partner_name"`
	ServiceMode         string  `json:"service_mode"`
	AppliedWeight       float64 `json:"applied_weight"`
	InvoiceValue        float64 `json:"invoice_value"`
	CollectableAmount   float64 `json:"collectable_amount"`
	Insurance           float64 `json:"insurance"`
	Base                float64 `json:"base"`
	Add                 float64 `json:"add"`
	Variables           float64 `json:"variables"`
	MinChargeableAmount float64 `json:"minchargeableamount"`
	VariableServices    float64 `json:"variable_services"`
	COD                 float64 `json:"cod"`
	Tax                 float64 `json:"tax"`
	Total               float64 `json:"total"`
	ZoneName            string  `json:"zone_name"`
	EDT                 int     `json:"edt"`
	SortType            string  `json:"sort_type"`
	PriorityRank        *int    `json:"priority_rank"`
	CheapestRank        int     `json:"cheapest_rank"`
}

// Rate represents rate information from rate service
type Rate struct {
	RateID string `json:"rate_id,omitempty"`
	Price  Price  `json:"price"`
}

// Price represents price details
type Price struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	Type     string  `json:"type"`
}

// PartnerError represents error information for a partner
type PartnerError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error code constants
const (
	ErrorCodeAdapterNotFound         = "ADAPTER_NOT_FOUND"
	ErrorCodeServiceabilityFailed    = "SERVICEABILITY_FAILED"
	ErrorCodeInvalidRequest          = "INVALID_REQUEST"
	ErrorCodePartnerUnavailable      = "PARTNER_UNAVAILABLE"
)

// HubOpsRequest represents request to HubOps API
type HubOpsRequest struct {
	PostalCode string `json:"postalCode"`
}

// HubOpsResponse represents response from HubOps API
type HubOpsResponse struct {
	NearestInternationalHub *HubInfoData `json:"nearestInternationalHub,omitempty"`
}

// HubInfoData represents hub information
type HubInfoData struct {
	Pincode         *int        `json:"pincode,omitempty"`
	PremiseName     *string     `json:"premiseName,omitempty"`
	AddressLine1    *string     `json:"addressLine1,omitempty"`
	AddressLine2    *string     `json:"addressLine2,omitempty"`
	Address         *string     `json:"address,omitempty"`
	City            *string     `json:"city,omitempty"`
	State           *string     `json:"state,omitempty"`
	Latitude        *string     `json:"latitude,omitempty"`
	Longitude       *string     `json:"longitude,omitempty"`
	PersonalEmailId *string     `json:"personalEmailId,omitempty"`
	OfficialEmailId *string     `json:"officialEmailId,omitempty"`
	PersonalNumber  interface{} `json:"personalNumber,omitempty"`
	OfficialNumber  interface{} `json:"officialNumber,omitempty"`
}
