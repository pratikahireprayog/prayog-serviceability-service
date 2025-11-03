package models

import (
	"time"
)

// ServiceabilityV3Request represents the V3 request structure with addresses object
type ServiceabilityV3Request struct {
	Addresses         *AddressesInput    `json:"addresses" validate:"required"`
	ParcelCategory    *string            `json:"parcel_category,omitempty" validate:"omitempty,oneof=ecomm courier cargo international hyperlocal"`
	ProductType       *string            `json:"product_type,omitempty"`
	Packages          []Package          `json:"packages,omitempty" validate:"omitempty,dive"`
	Partners          []PartnerFilter    `json:"partners,omitempty" validate:"omitempty,dive"`
}

// AddressesInput represents the addresses object in V3 request
type AddressesInput struct {
	SourcePostalCode      *string `json:"source_postal_code" validate:"required,min=3,max=10"`
	DestinationPostalCode *string `json:"destination_postal_code" validate:"required,min=3,max=10"`
	SourceCountryCode     *string `json:"source_country_code,omitempty" validate:"omitempty,len=2"`
	DestinationCountryCode *string `json:"destination_country_code,omitempty" validate:"omitempty,len=2"`
}

// ServiceabilityV3Response represents the V3 response structure
type ServiceabilityV3Response struct {
	Success            bool                `json:"success"`
	Message            string              `json:"message,omitempty"`
	SourceAddress      *AddressInfo        `json:"source_address,omitempty"`
	DestinationAddress *AddressInfo        `json:"destination_address,omitempty"`
	Addresses          []DetailedAddress   `json:"addresses,omitempty"`
	HubDetails         interface{}         `json:"hub_details,omitempty"`
	Partners           []PartnerV3Response `json:"partners"`
	Error              *ErrorResponse      `json:"error,omitempty"`
	Metadata           *V3ResponseMetadata `json:"metadata,omitempty"`
}

// PartnerV3Response represents individual partner response in V3 format (services only, no standard_services)
type PartnerV3Response struct {
	PartnerID       string                 `json:"partner_id"`
	PartnerCode     string                 `json:"partner_code"`
	PartnerName     string                 `json:"partner_name,omitempty"`
	Rating          float64                `json:"rating"`
	Source          string                 `json:"source,omitempty"`
	IsServiceable   bool                   `json:"is_serviceable"`
	Services        interface{}            `json:"services,omitempty"` // Only services field, no standard_services
	Capabilities    map[string]interface{} `json:"capabilities,omitempty"`
	Error           *string                `json:"error,omitempty"`
	ResponseTime    time.Duration          `json:"response_time"`
	ResponseTimeMs  int64                  `json:"response_time_ms,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	HubDetails      interface{}            `json:"hub_details,omitempty"`
}

// V3ResponseMetadata contains metadata about the V3 response
type V3ResponseMetadata struct {
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

// ToV2Request converts V3 request to V2 request format
func (req *ServiceabilityV3Request) ToV2Request() *ServiceabilityV2Request {
	if req == nil || req.Addresses == nil {
		return nil
	}

	return &ServiceabilityV2Request{
		SourcePostalCode:       req.Addresses.SourcePostalCode,
		DestinationPostalCode:  req.Addresses.DestinationPostalCode,
		SourceCountryCode:      req.Addresses.SourceCountryCode,
		DestinationCountryCode: req.Addresses.DestinationCountryCode,
		ParcelCategory:         req.ParcelCategory,
		ProductType:            req.ProductType,
		Packages:               req.Packages,
		Partners:               req.Partners,
	}
}

// PartnerV3ResponseFromV2 converts V2 partner response to V3 format
func PartnerV3ResponseFromV2(v2Partner *PartnerV2Response) *PartnerV3Response {
	if v2Partner == nil {
		return nil
	}

	v3Partner := &PartnerV3Response{
		PartnerID:      v2Partner.PartnerID,
		PartnerCode:    v2Partner.PartnerCode,
		PartnerName:    v2Partner.PartnerName,
		Rating:         v2Partner.Rating,
		Source:         v2Partner.Source,
		IsServiceable:  v2Partner.IsServiceable,
		Capabilities:   v2Partner.Capabilities,
		ResponseTime:   v2Partner.ResponseTime,
		ResponseTimeMs: v2Partner.ResponseTimeMs,
		Metadata:       v2Partner.Metadata,
		HubDetails:     v2Partner.HubDetails,
	}

	// Use PartnerServices if available, otherwise use Services (but not standard_services)
	if v2Partner.PartnerServices != nil {
		v3Partner.Services = v2Partner.PartnerServices
	} else if len(v2Partner.Services) > 0 {
		v3Partner.Services = v2Partner.Services
	}

	// Set IsServiceable based on whether services are available
	if v3Partner.Services != nil {
		if services, ok := v3Partner.Services.([]ServiceV2); ok {
			v3Partner.IsServiceable = len(services) > 0
		} else {
			// For other types (like []interface{}), check if not empty
			v3Partner.IsServiceable = true
		}
	}

	if v2Partner.Error != nil {
		errorStr := *v2Partner.Error
		v3Partner.Error = &errorStr
	}

	return v3Partner
}

// ServiceabilityV3ResponseFromV2 converts V2 response to V3 format
func ServiceabilityV3ResponseFromV2(v2Response *ServiceabilityV2Response) *ServiceabilityV3Response {
	if v2Response == nil {
		return nil
	}

	v3Response := &ServiceabilityV3Response{
		Success:            v2Response.Success,
		Message:            v2Response.Message,
		SourceAddress:      v2Response.SourceAddress,
		DestinationAddress: v2Response.DestinationAddress,
		Addresses:          v2Response.Addresses,
		HubDetails:         v2Response.HubDetails,
		Error:              v2Response.Error,
	}

	// Convert partners
	v3Partners := make([]PartnerV3Response, 0, len(v2Response.Partners))
	for _, v2Partner := range v2Response.Partners {
		v3Partner := PartnerV3ResponseFromV2(&v2Partner)
		if v3Partner != nil {
			v3Partners = append(v3Partners, *v3Partner)
		}
	}
	v3Response.Partners = v3Partners

	// Convert metadata
	if v2Response.Metadata != nil {
		v3Response.Metadata = &V3ResponseMetadata{
			RequestID:                v2Response.Metadata.RequestID,
			ResponseTimeMs:           v2Response.Metadata.ResponseTimeMs,
			PartnersQueried:          v2Response.Metadata.PartnersQueried,
			PartnersSucceeded:        v2Response.Metadata.PartnersSucceeded,
			PartnersFailed:           v2Response.Metadata.PartnersFailed,
			TotalServiceablePartners: v2Response.Metadata.TotalServiceablePartners,
			RatesIncluded:            v2Response.Metadata.RatesIncluded,
			TotalPartners:            v2Response.Metadata.TotalPartners,
			ServiceableCount:         v2Response.Metadata.ServiceableCount,
			ProcessingTime:           v2Response.Metadata.ProcessingTime,
			Filters:                  v2Response.Metadata.Filters,
		}
	}

	return v3Response
}

