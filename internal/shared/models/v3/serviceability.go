package models

import (
	"time"

	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"
)

// ServiceabilityV3Request represents the V3 request structure with source_location and destination_location
type ServiceabilityV3Request struct {
	SourceLocation      *LocationInput            `json:"source_location" validate:"required"`
	DestinationLocation *LocationInput            `json:"destination_location" validate:"required"`
	ParcelCategory      *string                   `json:"parcel_category,omitempty" validate:"omitempty,oneof=ecomm courier cargo international hyperlocal"`
	ProductType         *string                   `json:"product_type,omitempty"`
	Packages            []modelsv1.Package         `json:"packages,omitempty" validate:"omitempty,dive"`
	Partners            []modelsv1.PartnerFilter   `json:"partners,omitempty" validate:"omitempty,dive"`
}

// LocationInput represents the structured location input in V3 request
type LocationInput struct {
	PostalCode  *string `json:"postal_code" validate:"required,min=3,max=10"`
	CountryCode *string `json:"country_code,omitempty" validate:"omitempty,len=2"`
}

// ServiceabilityV3Response represents the V3 response structure
type ServiceabilityV3Response struct {
	Success             bool                    `json:"success"`
	Message             string                  `json:"message,omitempty"`
	SourceLocation      *LocationInfo           `json:"source_location,omitempty"`
	DestinationLocation *LocationInfo           `json:"destination_location,omitempty"`
	HubDetails          interface{}             `json:"hub_details,omitempty"`
	Partners            []PartnerV3Response     `json:"partners"`
	Error               *modelsv1.ErrorResponse `json:"error,omitempty"`
	Metadata            *V3ResponseMetadata     `json:"metadata,omitempty"`
}

type LocationInfo struct {
	PostalCode  string `json:"postal_code,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
}

// PartnerV3Response represents individual partner response in V3 format
type PartnerV3Response struct {
	PartnerID      string                 `json:"partner_id"`
	PartnerCode    string                 `json:"partner_code"`
	PartnerName    string                 `json:"partner_name,omitempty"`
	Rating         float64                `json:"rating"`
	Source         string                 `json:"source,omitempty"`
	IsServiceable  bool                   `json:"is_serviceable"`
	Services       interface{}            `json:"services,omitempty"`
	Capabilities   interface{}            `json:"capabilities,omitempty"`
	Error          *string                `json:"error,omitempty"`
	ResponseTime   time.Duration          `json:"-"` // Internal use
	ResponseTimeMs int64                  `json:"response_time_ms,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	HubDetails     interface{}            `json:"hub_details,omitempty"`
}

// V3ResponseMetadata contains metadata about the V3 response
type V3ResponseMetadata struct {
	RequestID                string             `json:"request_id,omitempty"`
	ResponseTimeMs           int64              `json:"response_time_ms,omitempty"`
	PartnersQueried          int                `json:"partners_queried,omitempty"`
	PartnersSucceeded        int                `json:"partners_succeeded,omitempty"`
	PartnersFailed           int                `json:"partners_failed,omitempty"`
	TotalServiceablePartners int                `json:"total_serviceable_partners,omitempty"`
	RatesIncluded            bool               `json:"rates_included,omitempty"`
	TotalPartners            int                `json:"total_partners"`
	ServiceableCount         int                `json:"serviceable_count"`
	ProcessingTime           time.Duration      `json:"-"`
	Filters                  modelsv1.V2Filters `json:"filters"`
}

// ToV2Request converts V3 request to V2 request format
func (req *ServiceabilityV3Request) ToV2Request() *modelsv1.ServiceabilityV2Request {
	if req == nil {
		return nil
	}

	v2Req := &modelsv1.ServiceabilityV2Request{
		ParcelCategory: req.ParcelCategory,
		ProductType:    req.ProductType,
		Packages:       req.Packages,
		Partners:       req.Partners, // Preserve partners from request
	}

	if req.SourceLocation != nil {
		v2Req.SourcePostalCode = req.SourceLocation.PostalCode
		v2Req.SourceCountryCode = req.SourceLocation.CountryCode
	}

	if req.DestinationLocation != nil {
		v2Req.DestinationPostalCode = req.DestinationLocation.PostalCode
		v2Req.DestinationCountryCode = req.DestinationLocation.CountryCode
	}

	return v2Req
}

// PartnerV3ResponseFromV2 converts V2 partner response to V3 format
func PartnerV3ResponseFromV2(v2Partner *modelsv1.PartnerV2Response) *PartnerV3Response {
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

	// Use PartnerServices if available, otherwise use Services
	if v2Partner.PartnerServices != nil {
		v3Partner.Services = v2Partner.PartnerServices
	} else if v2Partner.Services != nil {
		v3Partner.Services = v2Partner.Services
	}

	if v2Partner.Error != nil {
		errorStr := *v2Partner.Error
		v3Partner.Error = &errorStr
	}

	return v3Partner
}

// ServiceabilityV3ResponseFromV2 converts V2 response to V3 format
func ServiceabilityV3ResponseFromV2(v2Response *modelsv1.ServiceabilityV2Response) *ServiceabilityV3Response {
	if v2Response == nil {
		return nil
	}

	v3Response := &ServiceabilityV3Response{
		Success:    v2Response.Success,
		Message:    v2Response.Message,
		HubDetails: v2Response.HubDetails,
		Error:      v2Response.Error,
	}

	if v2Response.SourceAddress != nil {
		v3Response.SourceLocation = &LocationInfo{
			PostalCode:  v2Response.SourceAddress.PostalCode,
			CountryCode: v2Response.SourceAddress.CountryCode,
		}
	}

	if v2Response.DestinationAddress != nil {
		v3Response.DestinationLocation = &LocationInfo{
			PostalCode:  v2Response.DestinationAddress.PostalCode,
			CountryCode: v2Response.DestinationAddress.CountryCode,
		}
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
