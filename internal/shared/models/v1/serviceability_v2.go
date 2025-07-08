package models

import (
	"time"
)

// V2 Request and Response Types for Multi-Partner Serviceability

// ServiceabilityV2Request represents the request structure for v2 serviceability checks
type ServiceabilityV2Request struct {
	PostalCode         *string `json:"postal_code,omitempty" validate:"omitempty,min=3,max=10"`
	PickupPostalCode   *string `json:"pickup_postal_code,omitempty" validate:"omitempty,min=3,max=10"`
	DeliveryPostalCode *string `json:"delivery_postal_code,omitempty" validate:"omitempty,min=3,max=10"`
	CountryCode        string  `json:"country_code" validate:"required,len=2"`
	ParcelCategory     *string `json:"parcel_category,omitempty" validate:"omitempty,oneof=ecomm courier cargo international hyperlocal"`
	ProductType        *string `json:"product_type,omitempty"`
}

// ServiceabilityV2Response represents the aggregated response structure for v2
type ServiceabilityV2Response struct {
	Success  bool                `json:"success"`
	Partners []PartnerV2Response `json:"partners,omitempty"`
	Error    *ErrorResponse      `json:"error,omitempty"`
	Metadata *V2ResponseMetadata `json:"metadata,omitempty"`
}

// PartnerV2Response represents individual partner response in v2 format
type PartnerV2Response struct {
	PartnerID     string                 `json:"partner_id"`
	PartnerCode   string                 `json:"partner_code"`
	PartnerName   string                 `json:"partner_name"`
	Rating        float64                `json:"rating"`
	IsServiceable bool                   `json:"is_serviceable"`
	Services      []ServiceV2            `json:"services,omitempty"`
	Capabilities  map[string]interface{} `json:"capabilities,omitempty"`
	Error         *string                `json:"error,omitempty"`
	ResponseTime  time.Duration          `json:"response_time"`
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
	TotalPartners    int           `json:"total_partners"`
	ServiceableCount int           `json:"serviceable_count"`
	ProcessingTime   time.Duration `json:"processing_time"`
	Filters          V2Filters     `json:"filters"`
	EligiblePartners []string      `json:"eligible_partners"`
}

// V2Filters represents the filters applied to determine eligible partners
type V2Filters struct {
	ParcelCategory *string `json:"parcel_category,omitempty"`
	ProductType    *string `json:"product_type,omitempty"`
	CountryCode    string  `json:"country_code"`
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
