package models

import (
	"errors"
	"strings"
	"time"

	"prayog-serviceability-service/internal/shared/constants/v1"
)

// Enhanced Request Models

// ServiceabilityCheckRequest represents a unified request for serviceability checks
// It supports both single location and origin-destination queries
type ServiceabilityCheckRequest struct {
	// Single location query fields
	PostalCode *string `json:"postal_code,omitempty" validate:"omitempty,min=1,max=20"`

	// Origin-destination query fields
	PickupPostalCode   *string `json:"pickup_postal_code,omitempty" validate:"omitempty,min=1,max=20"`
	DeliveryPostalCode *string `json:"delivery_postal_code,omitempty" validate:"omitempty,min=1,max=20"`

	// Common fields
	CountryCode     string  `json:"country_code" validate:"required,len=2"`
	ServiceTypeCode *string `json:"service_type_code,omitempty" validate:"omitempty,oneof=SDD NDD Standard Express vayuquick vayuquickpro"`
	ParcelCategory  *string `json:"parcel_category,omitempty" validate:"omitempty,oneof=ecom courier cargo"`
}

// GetQueryType returns the type of query (single location or origin-destination)
func (r *ServiceabilityCheckRequest) GetQueryType() string {
	if r.PostalCode != nil && *r.PostalCode != "" {
		return constants.QueryTypeGenericLocation
	}
	return constants.QueryTypeOriginDestination
}

// GetPostalCodes returns all postal codes involved in the request
func (r *ServiceabilityCheckRequest) GetPostalCodes() []string {
	var codes []string

	if r.PostalCode != nil && *r.PostalCode != "" {
		codes = append(codes, *r.PostalCode)
	}

	if r.PickupPostalCode != nil && *r.PickupPostalCode != "" {
		codes = append(codes, *r.PickupPostalCode)
	}

	if r.DeliveryPostalCode != nil && *r.DeliveryPostalCode != "" {
		codes = append(codes, *r.DeliveryPostalCode)
	}

	return codes
}

// BulkServiceabilityRequest represents a request for multiple serviceability checks
type BulkServiceabilityRequest struct {
	Requests []ServiceabilityCheckRequest `json:"requests" validate:"required,min=1,max=100,dive"`
}

// Enhanced Response Models

// ServiceabilityResponse represents the response structure for serviceability checks
type ServiceabilityResponse struct {
	Success bool                `json:"success"`
	Data    *ServiceabilityData `json:"data,omitempty"`
	Error   *ErrorResponse      `json:"error,omitempty"`
}

// ServiceabilityData contains the core serviceability information
type ServiceabilityData struct {
	QueryType        string        `json:"query_type"` // "generic_location" or "origin_destination"
	Location         *LocationData `json:"location,omitempty"`
	PickupLocation   *LocationData `json:"pickup_location,omitempty"`
	DeliveryLocation *LocationData `json:"delivery_location,omitempty"`
}

// LocationData represents location information with its serviceability options
type LocationData struct {
	PostalCode     string                  `json:"postal_code"`
	CountryCode    string                  `json:"country_code"`
	Serviceability []ParcelCategoryService `json:"serviceability"`
}

// ParcelCategoryService groups services by parcel category
type ParcelCategoryService struct {
	ParcelCategory string    `json:"parcel_category"` // ecom, courier, cargo
	Services       []Service `json:"services"`
}

// Service represents an individual service offering
type Service struct {
	ServiceType    string   `json:"service_type"`              // SDD, NDD, Standard, Express, vayuquick, etc.
	OperationTypes []string `json:"operation_types,omitempty"` // pickup, delivery
	PaymentModes   []string `json:"payment_modes"`             // ONLINE, COD
	DeliveryModes  []string `json:"delivery_modes"`            // AIR, SURFACE, RAIL
}

// BulkServiceabilityResponse represents the response for bulk serviceability checks
type BulkServiceabilityResponse struct {
	Success bool                     `json:"success"`
	Data    []ServiceabilityResponse `json:"data,omitempty"`
	Error   *ErrorResponse           `json:"error,omitempty"`
}


// Helper Models for Business Logic

// PartnerCapability represents a partner's service capability
type PartnerCapability struct {
	PartnerID      uint     `json:"partner_id"`
	ServiceType    string   `json:"service_type"`
	ParcelCategory string   `json:"parcel_category"`
	OperationTypes []string `json:"operation_types"`
	PaymentModes   []string `json:"payment_modes"`
	DeliveryModes  []string `json:"delivery_modes"`
	IsActive       bool     `json:"is_active"`
	Rating         float64  `json:"rating"`
	Priority       int      `json:"priority"`
}

// ServiceabilityFilters represents filters for serviceability queries
type ServiceabilityFilters struct {
	ServiceTypes     []string `json:"service_types,omitempty"`
	ParcelCategories []string `json:"parcel_categories,omitempty"`
	OperationTypes   []string `json:"operation_types,omitempty"`
	PaymentModes     []string `json:"payment_modes,omitempty"`
	DeliveryModes    []string `json:"delivery_modes,omitempty"`
}

// Legacy models for backward compatibility (to be deprecated)

// Location represents a geographic location (legacy)
type Location struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	PostalCode string    `json:"postal_code" gorm:"not null;index"`
	City       string    `json:"city"`
	Region     string    `json:"region"`
	Country    string    `json:"country"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ServiceabilityResult represents the result of a serviceability check (legacy)
type ServiceabilityResult struct {
	Location      *Location `json:"location"`
	IsServiceable bool      `json:"is_serviceable"`
	ServiceType   string    `json:"service_type,omitempty"`
	OrderType     string    `json:"order_type,omitempty"`
	Message       string    `json:"message,omitempty"`
}

// ServiceabilityRequest represents a request to check if a location is serviceable (legacy)
type ServiceabilityRequest struct {
	PostalCode      string `json:"postal_code"`
	CountryCode     string `json:"country_code,omitempty"`
	ServiceTypeCode string `json:"service_type_code,omitempty"`
	OrderTypeCode   string `json:"order_type_code,omitempty"`
}

// ServiceabilityResponseLegacy represents the legacy response format
type ServiceabilityResponseLegacy struct {
	PostalCode    string             `json:"postal_code"`
	IsServiceable bool               `json:"is_serviceable"`
	Services      []ServiceAvailable `json:"services,omitempty"`
	Message       string             `json:"message,omitempty"`
}

// ServiceAvailable represents the availability status of a service (legacy)
type ServiceAvailable struct {
	ServiceTypeCode string `json:"service_type_code"`
	ServiceTypeName string `json:"service_type_name"`
	IsAvailable     bool   `json:"is_available"`
}

// ToEnhancedRequest converts legacy request to enhanced request format
func (r *ServiceabilityRequest) ToEnhancedRequest() *ServiceabilityCheckRequest {
	enhanced := &ServiceabilityCheckRequest{
		PostalCode:  &r.PostalCode,
		CountryCode: r.CountryCode,
	}

	if r.CountryCode == "" {
		enhanced.CountryCode = "IN" // Default to India for legacy requests
	}

	if r.ServiceTypeCode != "" {
		enhanced.ServiceTypeCode = &r.ServiceTypeCode
	}

	// Map legacy order type to parcel category if needed
	if r.OrderTypeCode != "" {
		category := mapOrderTypeToParcelCategory(r.OrderTypeCode)
		if category != "" {
			enhanced.ParcelCategory = &category
		}
	}

	return enhanced
}

// Helper function to map legacy order types to parcel categories
func mapOrderTypeToParcelCategory(orderType string) string {
	switch strings.ToLower(orderType) {
	case "ecommerce", "ecom":
		return "ecom"
	case "courier":
		return "courier"
	case "cargo":
		return "cargo"
	default:
		return ""
	}
}

// Validation helper methods

// ValidateEnhancedRequest performs validation on ServiceabilityCheckRequest
func (r *ServiceabilityCheckRequest) ValidateEnhancedRequest() error {
	// Ensure we have either a single postal code OR both pickup and delivery postal codes
	singleLocation := r.PostalCode != nil && *r.PostalCode != ""
	originDestination := r.PickupPostalCode != nil && *r.PickupPostalCode != "" &&
		r.DeliveryPostalCode != nil && *r.DeliveryPostalCode != ""

	if !singleLocation && !originDestination {
		return errors.New("either postal_code or both pickup_postal_code and delivery_postal_code must be provided")
	}

	if singleLocation && originDestination {
		return errors.New("cannot provide both postal_code and pickup/delivery postal codes in the same request")
	}

	// Basic country code validation
	if len(r.CountryCode) != 2 {
		return errors.New("country_code must be exactly 2 characters")
	}

	return nil
}

// ValidateBulkRequest performs validation on BulkServiceabilityRequest
func (r *BulkServiceabilityRequest) ValidateBulkRequest() error {
	if len(r.Requests) == 0 {
		return errors.New("at least one request is required")
	}

	if len(r.Requests) > 100 {
		return errors.New("maximum 100 requests allowed in bulk operation")
	}

	// Validate each individual request
	for i, req := range r.Requests {
		if err := req.ValidateEnhancedRequest(); err != nil {
			return errors.New("request " + string(rune(i+1)) + ": " + err.Error())
		}
	}

	return nil
}
