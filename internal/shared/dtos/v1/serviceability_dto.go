package dtos

import (
	"prayog-serviceability-service/internal/shared/models/v1"
)

// SingleLocationServiceabilityRequest represents a request for single postal code serviceability
type SingleLocationServiceabilityRequest struct {
	PostalCode      string  `json:"postal_code" validate:"required"`
	CountryCode     string  `json:"country_code" validate:"required,len=2"`
	ServiceTypeCode *string `json:"service_type_code,omitempty"`
	ParcelCategory  *string `json:"parcel_category,omitempty"`
}

// OriginDestinationServiceabilityRequest represents a request for origin-destination serviceability
type OriginDestinationServiceabilityRequest struct {
	PickupPostalCode   string  `json:"pickup_postal_code" validate:"required"`
	DeliveryPostalCode string  `json:"delivery_postal_code" validate:"required"`
	CountryCode        string  `json:"country_code" validate:"required,len=2"`
	ServiceTypeCode    *string `json:"service_type_code,omitempty"`
	ParcelCategory     *string `json:"parcel_category,omitempty"`
}

// BulkServiceabilityRequest represents a request for multiple serviceability checks
type BulkServiceabilityRequest struct {
	Requests []ServiceabilityRequestItem `json:"requests" validate:"required,min=1,max=100"`
}

// ServiceabilityRequestItem represents an item in bulk request
type ServiceabilityRequestItem struct {
	PostalCode         *string `json:"postal_code,omitempty"`
	PickupPostalCode   *string `json:"pickup_postal_code,omitempty"`
	DeliveryPostalCode *string `json:"delivery_postal_code,omitempty"`
	CountryCode        string  `json:"country_code" validate:"required,len=2"`
	ServiceTypeCode    *string `json:"service_type_code,omitempty"`
	ParcelCategory     *string `json:"parcel_category,omitempty"`
}

// ServiceabilityResponse represents the main response structure
type ServiceabilityResponse struct {
	Success bool               `json:"success"`
	Data    ServiceabilityData `json:"data,omitempty"`
	Error   *ErrorResponse     `json:"error,omitempty"`
}

// ServiceabilityData contains the serviceability information
type ServiceabilityData struct {
	QueryType        string        `json:"query_type"` // "generic_location" or "origin_destination"
	Location         *LocationData `json:"location,omitempty"`
	PickupLocation   *LocationData `json:"pickup_location,omitempty"`
	DeliveryLocation *LocationData `json:"delivery_location,omitempty"`
}

// LocationData represents location information with serviceability
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

// Service represents a service offering
type Service struct {
	ServiceType    string   `json:"service_type"`              // SDD, NDD, Standard, Express, etc.
	OperationTypes []string `json:"operation_types,omitempty"` // pickup, delivery
	PaymentModes   []string `json:"payment_modes"`             // ONLINE, COD
	DeliveryModes  []string `json:"delivery_modes"`            // AIR, SURFACE, RAIL
}

// BulkServiceabilityResponse represents bulk response
type BulkServiceabilityResponse struct {
	Success bool                     `json:"success"`
	Data    []ServiceabilityResponse `json:"data,omitempty"`
	Error   *ErrorResponse           `json:"error,omitempty"`
}

// ErrorResponse represents error information
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Legacy DTOs for backward compatibility

// LegacyServiceabilityRequest represents the legacy request format
type LegacyServiceabilityRequest struct {
	PostalCode      string `json:"postal_code" validate:"required"`
	CountryCode     string `json:"country_code,omitempty"`
	ServiceTypeCode string `json:"service_type_code,omitempty"`
	OrderTypeCode   string `json:"order_type_code,omitempty"`
}

// LegacyServiceabilityResponse represents the legacy response format
type LegacyServiceabilityResponse struct {
	PostalCode    string                   `json:"postal_code"`
	IsServiceable bool                     `json:"is_serviceable"`
	Services      []LegacyServiceAvailable `json:"services,omitempty"`
	Message       string                   `json:"message,omitempty"`
}

// LegacyServiceAvailable represents legacy service availability
type LegacyServiceAvailable struct {
	ServiceTypeCode string `json:"service_type_code"`
	ServiceTypeName string `json:"service_type_name"`
	IsAvailable     bool   `json:"is_available"`
}

// LegacyBulkServiceabilityRequest represents bulk legacy request
type LegacyBulkServiceabilityRequest struct {
	Requests []LegacyServiceabilityRequest `json:"requests"`
}

// LegacyBulkServiceabilityResponse represents bulk legacy response
type LegacyBulkServiceabilityResponse struct {
	Results []LegacyServiceabilityResponse `json:"results"`
}

// Enhanced DTOs

// ServiceabilityCheckRequestDTO represents the enhanced request DTO
type ServiceabilityCheckRequestDTO struct {
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

// ToModel converts DTO to model
func (dto *ServiceabilityCheckRequestDTO) ToModel() *models.ServiceabilityCheckRequest {
	return &models.ServiceabilityCheckRequest{
		PostalCode:         dto.PostalCode,
		PickupPostalCode:   dto.PickupPostalCode,
		DeliveryPostalCode: dto.DeliveryPostalCode,
		CountryCode:        dto.CountryCode,
		ServiceTypeCode:    dto.ServiceTypeCode,
		ParcelCategory:     dto.ParcelCategory,
	}
}

// BulkServiceabilityRequestDTO represents the enhanced bulk request DTO
type BulkServiceabilityRequestDTO struct {
	Requests []ServiceabilityCheckRequestDTO `json:"requests" validate:"required,min=1,max=100,dive"`
}

// ToModel converts DTO to model
func (dto *BulkServiceabilityRequestDTO) ToModel() *models.BulkServiceabilityRequest {
	requests := make([]models.ServiceabilityCheckRequest, len(dto.Requests))
	for i, req := range dto.Requests {
		requests[i] = *req.ToModel()
	}
	return &models.BulkServiceabilityRequest{
		Requests: requests,
	}
}

// ServiceabilityResponseDTO represents the enhanced response DTO
type ServiceabilityResponseDTO struct {
	Success bool                   `json:"success"`
	Data    *ServiceabilityDataDTO `json:"data,omitempty"`
	Error   *ErrorResponseDTO      `json:"error,omitempty"`
}

// FromModel converts model to DTO
func ServiceabilityResponseDTOFromModel(model *models.ServiceabilityResponse) *ServiceabilityResponseDTO {
	dto := &ServiceabilityResponseDTO{
		Success: model.Success,
	}

	if model.Data != nil {
		dto.Data = ServiceabilityDataDTOFromModel(model.Data)
	}

	if model.Error != nil {
		dto.Error = ErrorResponseDTOFromModel(model.Error)
	}

	return dto
}

// ServiceabilityDataDTO contains the serviceability information
type ServiceabilityDataDTO struct {
	QueryType        string           `json:"query_type"` // "generic_location" or "origin_destination"
	Location         *LocationDataDTO `json:"location,omitempty"`
	PickupLocation   *LocationDataDTO `json:"pickup_location,omitempty"`
	DeliveryLocation *LocationDataDTO `json:"delivery_location,omitempty"`
}

// FromModel converts model to DTO
func ServiceabilityDataDTOFromModel(model *models.ServiceabilityData) *ServiceabilityDataDTO {
	dto := &ServiceabilityDataDTO{
		QueryType: model.QueryType,
	}

	if model.Location != nil {
		dto.Location = LocationDataDTOFromModel(model.Location)
	}

	if model.PickupLocation != nil {
		dto.PickupLocation = LocationDataDTOFromModel(model.PickupLocation)
	}

	if model.DeliveryLocation != nil {
		dto.DeliveryLocation = LocationDataDTOFromModel(model.DeliveryLocation)
	}

	return dto
}

// LocationDataDTO represents location information with serviceability
type LocationDataDTO struct {
	PostalCode     string                     `json:"postal_code"`
	CountryCode    string                     `json:"country_code"`
	Serviceability []ParcelCategoryServiceDTO `json:"serviceability"`
}

// FromModel converts model to DTO
func LocationDataDTOFromModel(model *models.LocationData) *LocationDataDTO {
	serviceability := make([]ParcelCategoryServiceDTO, len(model.Serviceability))
	for i, pcs := range model.Serviceability {
		serviceability[i] = *ParcelCategoryServiceDTOFromModel(&pcs)
	}

	return &LocationDataDTO{
		PostalCode:     model.PostalCode,
		CountryCode:    model.CountryCode,
		Serviceability: serviceability,
	}
}

// ParcelCategoryServiceDTO groups services by parcel category
type ParcelCategoryServiceDTO struct {
	ParcelCategory string       `json:"parcel_category"` // ecom, courier, cargo
	Services       []ServiceDTO `json:"services"`
}

// FromModel converts model to DTO
func ParcelCategoryServiceDTOFromModel(model *models.ParcelCategoryService) *ParcelCategoryServiceDTO {
	services := make([]ServiceDTO, len(model.Services))
	for i, svc := range model.Services {
		services[i] = *ServiceDTOFromModel(&svc)
	}

	return &ParcelCategoryServiceDTO{
		ParcelCategory: model.ParcelCategory,
		Services:       services,
	}
}

// ServiceDTO represents a service offering
type ServiceDTO struct {
	ServiceType    string   `json:"service_type"`              // SDD, NDD, Standard, Express, etc.
	OperationTypes []string `json:"operation_types,omitempty"` // pickup, delivery
	PaymentModes   []string `json:"payment_modes"`             // ONLINE, COD
	DeliveryModes  []string `json:"delivery_modes"`            // AIR, SURFACE, RAIL
}

// FromModel converts model to DTO
func ServiceDTOFromModel(model *models.Service) *ServiceDTO {
	return &ServiceDTO{
		ServiceType:    model.ServiceType,
		OperationTypes: model.OperationTypes,
		PaymentModes:   model.PaymentModes,
		DeliveryModes:  model.DeliveryModes,
	}
}

// BulkServiceabilityResponseDTO represents bulk response
type BulkServiceabilityResponseDTO struct {
	Success bool                        `json:"success"`
	Data    []ServiceabilityResponseDTO `json:"data,omitempty"`
	Error   *ErrorResponseDTO           `json:"error,omitempty"`
}

// FromModel converts model to DTO
func BulkServiceabilityResponseDTOFromModel(model *models.BulkServiceabilityResponse) *BulkServiceabilityResponseDTO {
	dto := &BulkServiceabilityResponseDTO{
		Success: model.Success,
	}

	if model.Data != nil {
		data := make([]ServiceabilityResponseDTO, len(model.Data))
		for i, resp := range model.Data {
			data[i] = *ServiceabilityResponseDTOFromModel(&resp)
		}
		dto.Data = data
	}

	if model.Error != nil {
		dto.Error = ErrorResponseDTOFromModel(model.Error)
	}

	return dto
}

// ErrorResponseDTO represents error information
type ErrorResponseDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// FromModel converts model to DTO
func ErrorResponseDTOFromModel(model *models.ErrorResponse) *ErrorResponseDTO {
	return &ErrorResponseDTO{
		Code:    model.Code,
		Message: model.Message,
		Details: model.Details,
	}
}

// Legacy DTOs for backward compatibility

// LegacyServiceabilityRequestDTO represents the legacy request format
type LegacyServiceabilityRequestDTO struct {
	PostalCode      string `json:"postal_code" validate:"required,min=1,max=20"`
	CountryCode     string `json:"country_code,omitempty" validate:"omitempty,len=2"`
	ServiceTypeCode string `json:"service_type_code,omitempty"`
	OrderTypeCode   string `json:"order_type_code,omitempty"`
}

// ToModel converts legacy DTO to model
func (dto *LegacyServiceabilityRequestDTO) ToModel() *models.ServiceabilityRequest {
	return &models.ServiceabilityRequest{
		PostalCode:      dto.PostalCode,
		CountryCode:     dto.CountryCode,
		ServiceTypeCode: dto.ServiceTypeCode,
		OrderTypeCode:   dto.OrderTypeCode,
	}
}

// ToEnhancedModel converts legacy DTO to enhanced model
func (dto *LegacyServiceabilityRequestDTO) ToEnhancedModel() *models.ServiceabilityCheckRequest {
	return dto.ToModel().ToEnhancedRequest()
}

// LegacyServiceabilityResponseDTO represents the legacy response format
type LegacyServiceabilityResponseDTO struct {
	PostalCode    string                      `json:"postal_code"`
	IsServiceable bool                        `json:"is_serviceable"`
	Services      []LegacyServiceAvailableDTO `json:"services,omitempty"`
	Message       string                      `json:"message,omitempty"`
}

// LegacyServiceAvailableDTO represents legacy service availability
type LegacyServiceAvailableDTO struct {
	ServiceTypeCode string `json:"service_type_code"`
	ServiceTypeName string `json:"service_type_name"`
	IsAvailable     bool   `json:"is_available"`
}

// LegacyBulkServiceabilityRequestDTO represents bulk legacy request
type LegacyBulkServiceabilityRequestDTO struct {
	Requests []LegacyServiceabilityRequestDTO `json:"requests" validate:"required,min=1,max=100,dive"`
}

// LegacyBulkServiceabilityResponseDTO represents bulk legacy response
type LegacyBulkServiceabilityResponseDTO struct {
	Results []LegacyServiceabilityResponseDTO `json:"results"`
}

// Converter functions for legacy compatibility

// NEW POSTAL CODE BASED SERVICEABILITY DTOS

// PostalCodeServiceabilityRequest represents the request for postal code based serviceability API
type PostalCodeServiceabilityRequest struct {
	SourcePostalCode      *string `json:"source_postal_code,omitempty" validate:"omitempty,min=1,max=20"`
	DestinationPostalCode string  `json:"destination_postal_code" validate:"required,min=1,max=20"`
	ParcelCategory        *string `json:"parcel_category,omitempty" validate:"omitempty,oneof=ecomm cargo courier"`
	ProductType           *string `json:"product_type,omitempty"`
}

// PostalCodeServiceabilityResponse represents the response for postal code based serviceability API
type PostalCodeServiceabilityResponse struct {
	Success       bool                           `json:"success"`
	IsServiceable *bool                          `json:"is_serviceable"`
	Data          interface{}                    `json:"data,omitempty"`
	Error         *PostalCodeServiceabilityError `json:"error,omitempty"`
}

// PostalCodeServiceabilityData contains the main serviceability data
type PostalCodeServiceabilityData struct {
	SourcePostalCode      *string                           `json:"source_postal_code,omitempty"`
	DestinationPostalCode string                            `json:"destination_postal_code"`
	Serviceability        []PostalCodeParcelCategoryService `json:"serviceability"`
}

// PostalCodeParcelCategoryService groups services by parcel category
type PostalCodeParcelCategoryService struct {
	ParcelCategoryCode string                  `json:"parcel_category_code"` // ecomm, courier, cargo
	IsServiceable      bool                    `json:"is_serviceable"`
	Services           []PostalCodeServiceInfo `json:"services"`
}

// PostalCodeServiceInfo represents an individual service offering
type PostalCodeServiceInfo struct {
	ServiceCode   string                         `json:"service_code"` // sdd, ndd, express, etc.
	TATDays       *int                           `json:"tat_days,omitempty"`
	IsCOD         bool                           `json:"is_cod"`
	Pickup        *bool                          `json:"pickup,omitempty"`
	Delivery      *bool                          `json:"delivery,omitempty"`
	Insurance     *bool                          `json:"insurance,omitempty"`
	ProductTypes  map[string]bool                `json:"product_types,omitempty"`
	DeliveryModes PostalCodeServiceDeliveryModes `json:"delivery_modes"`
}

// PostalCodeServiceDeliveryModes represents delivery mode availability
type PostalCodeServiceDeliveryModes struct {
	Air     bool `json:"air"`
	Surface bool `json:"surface"`
}

// PostalCodeServiceabilityError represents error information
type PostalCodeServiceabilityError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ConvertEnhancedToLegacyResponse converts enhanced response to legacy format
func ConvertEnhancedToLegacyResponse(enhanced *ServiceabilityResponseDTO, postalCode string) *LegacyServiceabilityResponseDTO {
	legacy := &LegacyServiceabilityResponseDTO{
		PostalCode:    postalCode,
		IsServiceable: enhanced.Success && enhanced.Data != nil,
		Services:      []LegacyServiceAvailableDTO{},
	}

	if enhanced.Error != nil {
		legacy.Message = enhanced.Error.Message
		legacy.IsServiceable = false
		return legacy
	}

	if enhanced.Data != nil && enhanced.Data.Location != nil {
		// Convert serviceability data to legacy format
		for _, parcelCategoryService := range enhanced.Data.Location.Serviceability {
			for _, service := range parcelCategoryService.Services {
				legacy.Services = append(legacy.Services, LegacyServiceAvailableDTO{
					ServiceTypeCode: service.ServiceType,
					ServiceTypeName: service.ServiceType, // Use same value for legacy
					IsAvailable:     true,
				})
			}
		}
	}

	return legacy
}
