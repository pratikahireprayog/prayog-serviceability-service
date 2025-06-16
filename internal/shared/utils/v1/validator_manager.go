package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/go-playground/validator/v10"
)

// PartnerDataValidator defines the interface for validating partner-related data
type PartnerDataValidator interface {
	ValidatePartnerCapability(ctx context.Context, capability *interfaces.PartnerCapability) error
	ValidatePartnerServiceCapabilities(ctx context.Context, capabilities *interfaces.PartnerServiceCapabilities) error
	ValidateServiceFilters(ctx context.Context, filters *interfaces.ServiceFilters) error
	ValidatePartnerInfo(ctx context.Context, partnerInfo *interfaces.PartnerInfo) error
	ValidateLocationHierarchy(ctx context.Context, hierarchy *models.LocationHierarchy) error
	ValidateServiceabilityRequest(ctx context.Context, req *models.ServiceabilityCheckRequest) error
	ValidateBulkServiceabilityRequest(ctx context.Context, req *models.BulkServiceabilityRequest) error
	ValidatePartnerRating(rating float64) error
	ValidateServiceType(serviceType string) error
	ValidateParcelCategory(category string) error
	ValidateOperationType(operationType string) error
	ValidatePaymentMode(mode string) error
	ValidateDeliveryMode(mode string) error
}

// partnerDataValidator implements the PartnerDataValidator interface
type partnerDataValidator struct {
	validator               *validator.Validate
	allowedServiceTypes     []string
	allowedParcelCategories []string
	allowedOperationTypes   []string
	allowedPaymentModes     []string
	allowedDeliveryModes    []string
	timeout                 time.Duration
}

// ValidationConfig holds configuration for the validator
type ValidationConfig struct {
	AllowedServiceTypes     []string
	AllowedParcelCategories []string
	AllowedOperationTypes   []string
	AllowedPaymentModes     []string
	AllowedDeliveryModes    []string
	Timeout                 time.Duration
}

// ValidationError represents a validation error with field-specific details
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
	Tag     string      `json:"tag,omitempty"`
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface for ValidationErrors
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		if err.Value != nil {
			messages = append(messages, fmt.Sprintf("%s: %s (value: %v)", err.Field, err.Message, err.Value))
		} else {
			messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
		}
	}
	return "validation failed: " + strings.Join(messages, ", ")
}

// Regular expressions for validation
var (
	postalCodeRegex  = regexp.MustCompile(`^[A-Z0-9\-\s]{3,20}$`)
	countryCodeRegex = regexp.MustCompile(`^[A-Z]{2}$`)
	partnerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_\.&]+$`)
)

// NewPartnerDataValidator creates a new instance of the partner data validator
func NewPartnerDataValidator(config *ValidationConfig) PartnerDataValidator {
	v := validator.New()

	// Register custom validation functions
	v.RegisterValidation("postal_code", validatePostalCode)
	v.RegisterValidation("country_code", validateCountryCode)
	v.RegisterValidation("partner_name", validatePartnerName)
	v.RegisterValidation("service_type", validateServiceTypeTag)
	v.RegisterValidation("parcel_category", validateParcelCategoryTag)
	v.RegisterValidation("rating", validateRatingTag)

	if config == nil {
		config = &ValidationConfig{
			Timeout: 30 * time.Second,
		}
	}

	// Set default allowed values if not provided
	if config.AllowedServiceTypes == nil {
		config.AllowedServiceTypes = []string{"SDD", "NDD", "Standard", "Express", "vayuquick", "vayuquickpro"}
	}
	if config.AllowedParcelCategories == nil {
		config.AllowedParcelCategories = []string{"ecom", "courier", "cargo"}
	}
	if config.AllowedOperationTypes == nil {
		config.AllowedOperationTypes = []string{"pickup", "delivery"}
	}
	if config.AllowedPaymentModes == nil {
		config.AllowedPaymentModes = []string{"ONLINE", "COD"}
	}
	if config.AllowedDeliveryModes == nil {
		config.AllowedDeliveryModes = []string{"AIR", "SURFACE", "RAIL"}
	}

	return &partnerDataValidator{
		validator:               v,
		allowedServiceTypes:     config.AllowedServiceTypes,
		allowedParcelCategories: config.AllowedParcelCategories,
		allowedOperationTypes:   config.AllowedOperationTypes,
		allowedPaymentModes:     config.AllowedPaymentModes,
		allowedDeliveryModes:    config.AllowedDeliveryModes,
		timeout:                 config.Timeout,
	}
}

// ValidatePartnerCapability validates a partner capability structure
func (v *partnerDataValidator) ValidatePartnerCapability(ctx context.Context, capability *interfaces.PartnerCapability) error {
	if capability == nil {
		return ValidationErrors{
			Errors: []ValidationError{
				{Field: "capability", Message: "Partner capability cannot be nil"},
			},
		}
	}

	var errors []ValidationError

	// Validate partner ID
	if capability.PartnerID == 0 {
		errors = append(errors, ValidationError{
			Field:   "partner_id",
			Message: "Partner ID must be greater than 0",
			Value:   capability.PartnerID,
		})
	}

	// Validate partner name
	if err := v.validatePartnerNameField(capability.PartnerName); err != nil {
		errors = append(errors, ValidationError{
			Field:   "partner_name",
			Message: err.Error(),
			Value:   capability.PartnerName,
		})
	}

	// Validate service types
	if len(capability.ServiceTypes) == 0 {
		errors = append(errors, ValidationError{
			Field:   "service_types",
			Message: "At least one service type is required",
			Value:   capability.ServiceTypes,
		})
	} else {
		for i, serviceType := range capability.ServiceTypes {
			if err := v.ValidateServiceType(serviceType); err != nil {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("service_types[%d]", i),
					Message: err.Error(),
					Value:   serviceType,
				})
			}
		}
	}

	// Validate parcel categories
	if len(capability.ParcelCategories) == 0 {
		errors = append(errors, ValidationError{
			Field:   "parcel_categories",
			Message: "At least one parcel category is required",
			Value:   capability.ParcelCategories,
		})
	} else {
		for i, category := range capability.ParcelCategories {
			if err := v.ValidateParcelCategory(category); err != nil {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("parcel_categories[%d]", i),
					Message: err.Error(),
					Value:   category,
				})
			}
		}
	}

	// Validate operation types
	for i, opType := range capability.OperationTypes {
		if err := v.ValidateOperationType(opType); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("operation_types[%d]", i),
				Message: err.Error(),
				Value:   opType,
			})
		}
	}

	// Validate payment modes
	for i, mode := range capability.PaymentModes {
		if err := v.ValidatePaymentMode(mode); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("payment_modes[%d]", i),
				Message: err.Error(),
				Value:   mode,
			})
		}
	}

	// Validate delivery modes
	for i, mode := range capability.DeliveryModes {
		if err := v.ValidateDeliveryMode(mode); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("delivery_modes[%d]", i),
				Message: err.Error(),
				Value:   mode,
			})
		}
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidatePartnerServiceCapabilities validates partner service capabilities
func (v *partnerDataValidator) ValidatePartnerServiceCapabilities(ctx context.Context, capabilities *interfaces.PartnerServiceCapabilities) error {
	if capabilities == nil {
		return ValidationErrors{
			Errors: []ValidationError{
				{Field: "capabilities", Message: "Partner service capabilities cannot be nil"},
			},
		}
	}

	var errors []ValidationError

	// Validate partner ID
	if capabilities.PartnerID == 0 {
		errors = append(errors, ValidationError{
			Field:   "partner_id",
			Message: "Partner ID must be greater than 0",
			Value:   capabilities.PartnerID,
		})
	}

	// Validate individual capabilities
	for i, capability := range capabilities.Capabilities {
		if err := v.validateServiceCapability(capability, fmt.Sprintf("capabilities[%d]", i)); err != nil {
			if validationErrors, ok := err.(ValidationErrors); ok {
				errors = append(errors, validationErrors.Errors...)
			}
		}
	}

	// Validate preferences
	for i, preference := range capabilities.Preferences {
		if err := v.validatePartnerPreference(preference, fmt.Sprintf("preferences[%d]", i)); err != nil {
			if validationErrors, ok := err.(ValidationErrors); ok {
				errors = append(errors, validationErrors.Errors...)
			}
		}
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateServiceFilters validates service filter parameters
func (v *partnerDataValidator) ValidateServiceFilters(ctx context.Context, filters *interfaces.ServiceFilters) error {
	if filters == nil {
		return nil // Filters are optional
	}

	var errors []ValidationError

	// Validate service types
	for i, serviceType := range filters.ServiceTypes {
		if err := v.ValidateServiceType(serviceType); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("service_types[%d]", i),
				Message: err.Error(),
				Value:   serviceType,
			})
		}
	}

	// Validate parcel categories
	for i, category := range filters.ParcelCategories {
		if err := v.ValidateParcelCategory(category); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("parcel_categories[%d]", i),
				Message: err.Error(),
				Value:   category,
			})
		}
	}

	// Validate operation types
	for i, opType := range filters.OperationTypes {
		if err := v.ValidateOperationType(opType); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("operation_types[%d]", i),
				Message: err.Error(),
				Value:   opType,
			})
		}
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidatePartnerInfo validates partner information
func (v *partnerDataValidator) ValidatePartnerInfo(ctx context.Context, partnerInfo *interfaces.PartnerInfo) error {
	if partnerInfo == nil {
		return ValidationErrors{
			Errors: []ValidationError{
				{Field: "partner_info", Message: "Partner info cannot be nil"},
			},
		}
	}

	var errors []ValidationError

	// Validate ID
	if partnerInfo.ID == 0 {
		errors = append(errors, ValidationError{
			Field:   "id",
			Message: "Partner ID must be greater than 0",
			Value:   partnerInfo.ID,
		})
	}

	// Validate name
	if err := v.validatePartnerNameField(partnerInfo.Name); err != nil {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: err.Error(),
			Value:   partnerInfo.Name,
		})
	}

	// Validate type
	if strings.TrimSpace(partnerInfo.Type) == "" {
		errors = append(errors, ValidationError{
			Field:   "type",
			Message: "Partner type cannot be empty",
			Value:   partnerInfo.Type,
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateLocationHierarchy validates location hierarchy data
func (v *partnerDataValidator) ValidateLocationHierarchy(ctx context.Context, hierarchy *models.LocationHierarchy) error {
	if hierarchy == nil {
		return ValidationErrors{
			Errors: []ValidationError{
				{Field: "location_hierarchy", Message: "Location hierarchy cannot be nil"},
			},
		}
	}

	var errors []ValidationError

	// Validate postal code
	if !postalCodeRegex.MatchString(hierarchy.PostalCode) {
		errors = append(errors, ValidationError{
			Field:   "postal_code",
			Message: "Invalid postal code format",
			Value:   hierarchy.PostalCode,
		})
	}

	// Validate country code
	if !countryCodeRegex.MatchString(hierarchy.CountryCode) {
		errors = append(errors, ValidationError{
			Field:   "country_code",
			Message: "Country code must be 2 characters long",
			Value:   hierarchy.CountryCode,
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateServiceabilityRequest validates serviceability check request
func (v *partnerDataValidator) ValidateServiceabilityRequest(ctx context.Context, req *models.ServiceabilityCheckRequest) error {
	if req == nil {
		return ValidationErrors{
			Errors: []ValidationError{
				{Field: "request", Message: "Serviceability request cannot be nil"},
			},
		}
	}

	// Use struct validation with custom tags
	if err := v.validator.Struct(req); err != nil {
		return v.convertValidationErrors(err)
	}

	// Additional business logic validation
	var errors []ValidationError

	// Validate query type logic
	queryType := req.GetQueryType()
	if queryType == "generic_location" {
		if req.PickupPostalCode != nil || req.DeliveryPostalCode != nil {
			errors = append(errors, ValidationError{
				Field:   "query_type",
				Message: "Cannot specify pickup/delivery postal codes with generic postal code",
			})
		}
	} else if queryType == "origin_destination" {
		if req.PostalCode != nil {
			errors = append(errors, ValidationError{
				Field:   "query_type",
				Message: "Cannot specify generic postal code with pickup/delivery postal codes",
			})
		}
		if req.PickupPostalCode == nil || req.DeliveryPostalCode == nil {
			errors = append(errors, ValidationError{
				Field:   "query_type",
				Message: "Both pickup and delivery postal codes are required for origin-destination queries",
			})
		}
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateBulkServiceabilityRequest validates bulk serviceability request
func (v *partnerDataValidator) ValidateBulkServiceabilityRequest(ctx context.Context, req *models.BulkServiceabilityRequest) error {
	if req == nil {
		return ValidationErrors{
			Errors: []ValidationError{
				{Field: "request", Message: "Bulk serviceability request cannot be nil"},
			},
		}
	}

	// Use struct validation
	if err := v.validator.Struct(req); err != nil {
		return v.convertValidationErrors(err)
	}

	// Validate individual requests
	for i, individualReq := range req.Requests {
		if err := v.ValidateServiceabilityRequest(ctx, &individualReq); err != nil {
			if validationErrors, ok := err.(ValidationErrors); ok {
				for _, validationErr := range validationErrors.Errors {
					validationErr.Field = fmt.Sprintf("requests[%d].%s", i, validationErr.Field)
				}
				return ValidationErrors{Errors: validationErrors.Errors}
			}
		}
	}

	return nil
}

// ValidatePartnerRating validates partner rating value
func (v *partnerDataValidator) ValidatePartnerRating(rating float64) error {
	if rating < 0 || rating > 10 {
		return fmt.Errorf("rating must be between 0 and 10, got %.2f", rating)
	}
	return nil
}

// ValidateServiceType validates service type
func (v *partnerDataValidator) ValidateServiceType(serviceType string) error {
	if strings.TrimSpace(serviceType) == "" {
		return fmt.Errorf("service type cannot be empty")
	}

	for _, allowed := range v.allowedServiceTypes {
		if serviceType == allowed {
			return nil
		}
	}

	return fmt.Errorf("invalid service type '%s', allowed values: %s",
		serviceType, strings.Join(v.allowedServiceTypes, ", "))
}

// ValidateParcelCategory validates parcel category
func (v *partnerDataValidator) ValidateParcelCategory(category string) error {
	if strings.TrimSpace(category) == "" {
		return fmt.Errorf("parcel category cannot be empty")
	}

	for _, allowed := range v.allowedParcelCategories {
		if category == allowed {
			return nil
		}
	}

	return fmt.Errorf("invalid parcel category '%s', allowed values: %s",
		category, strings.Join(v.allowedParcelCategories, ", "))
}

// ValidateOperationType validates operation type
func (v *partnerDataValidator) ValidateOperationType(operationType string) error {
	if strings.TrimSpace(operationType) == "" {
		return fmt.Errorf("operation type cannot be empty")
	}

	for _, allowed := range v.allowedOperationTypes {
		if operationType == allowed {
			return nil
		}
	}

	return fmt.Errorf("invalid operation type '%s', allowed values: %s",
		operationType, strings.Join(v.allowedOperationTypes, ", "))
}

// ValidatePaymentMode validates payment mode
func (v *partnerDataValidator) ValidatePaymentMode(mode string) error {
	if strings.TrimSpace(mode) == "" {
		return fmt.Errorf("payment mode cannot be empty")
	}

	for _, allowed := range v.allowedPaymentModes {
		if mode == allowed {
			return nil
		}
	}

	return fmt.Errorf("invalid payment mode '%s', allowed values: %s",
		mode, strings.Join(v.allowedPaymentModes, ", "))
}

// ValidateDeliveryMode validates delivery mode
func (v *partnerDataValidator) ValidateDeliveryMode(mode string) error {
	if strings.TrimSpace(mode) == "" {
		return fmt.Errorf("delivery mode cannot be empty")
	}

	for _, allowed := range v.allowedDeliveryModes {
		if mode == allowed {
			return nil
		}
	}

	return fmt.Errorf("invalid delivery mode '%s', allowed values: %s",
		mode, strings.Join(v.allowedDeliveryModes, ", "))
}

// Helper methods

// validateServiceCapability validates individual service capability
func (v *partnerDataValidator) validateServiceCapability(capability interfaces.ServiceCapability, prefix string) error {
	var errors []ValidationError

	// Validate service type
	if err := v.ValidateServiceType(capability.ServiceType); err != nil {
		errors = append(errors, ValidationError{
			Field:   prefix + ".service_type",
			Message: err.Error(),
			Value:   capability.ServiceType,
		})
	}

	// Validate parcel category
	if err := v.ValidateParcelCategory(capability.ParcelCategory); err != nil {
		errors = append(errors, ValidationError{
			Field:   prefix + ".parcel_category",
			Message: err.Error(),
			Value:   capability.ParcelCategory,
		})
	}

	// Validate rating
	if err := v.ValidatePartnerRating(capability.Rating); err != nil {
		errors = append(errors, ValidationError{
			Field:   prefix + ".rating",
			Message: err.Error(),
			Value:   capability.Rating,
		})
	}

	// Validate operation types, payment modes, delivery modes
	for i, opType := range capability.OperationTypes {
		if err := v.ValidateOperationType(opType); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.operation_types[%d]", prefix, i),
				Message: err.Error(),
				Value:   opType,
			})
		}
	}

	for i, mode := range capability.PaymentModes {
		if err := v.ValidatePaymentMode(mode); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.payment_modes[%d]", prefix, i),
				Message: err.Error(),
				Value:   mode,
			})
		}
	}

	for i, mode := range capability.DeliveryModes {
		if err := v.ValidateDeliveryMode(mode); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.delivery_modes[%d]", prefix, i),
				Message: err.Error(),
				Value:   mode,
			})
		}
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// validatePartnerPreference validates partner preference
func (v *partnerDataValidator) validatePartnerPreference(preference interfaces.PartnerPreference, prefix string) error {
	var errors []ValidationError

	// Validate service type
	if err := v.ValidateServiceType(preference.ServiceType); err != nil {
		errors = append(errors, ValidationError{
			Field:   prefix + ".service_type",
			Message: err.Error(),
			Value:   preference.ServiceType,
		})
	}

	// Validate parcel category
	if err := v.ValidateParcelCategory(preference.ParcelCategory); err != nil {
		errors = append(errors, ValidationError{
			Field:   prefix + ".parcel_category",
			Message: err.Error(),
			Value:   preference.ParcelCategory,
		})
	}

	// Validate effective rating
	if err := v.ValidatePartnerRating(preference.EffectiveRating); err != nil {
		errors = append(errors, ValidationError{
			Field:   prefix + ".effective_rating",
			Message: err.Error(),
			Value:   preference.EffectiveRating,
		})
	}

	// Validate priority
	if preference.Priority < 1 || preference.Priority > 10 {
		errors = append(errors, ValidationError{
			Field:   prefix + ".priority",
			Message: "Priority must be between 1 and 10",
			Value:   preference.Priority,
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// validatePartnerNameField validates partner name field
func (v *partnerDataValidator) validatePartnerNameField(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("partner name cannot be empty")
	}

	if len(name) < 2 || len(name) > 100 {
		return fmt.Errorf("partner name must be between 2 and 100 characters")
	}

	if !partnerNameRegex.MatchString(name) {
		return fmt.Errorf("partner name contains invalid characters")
	}

	return nil
}

// convertValidationErrors converts validator errors to our custom format
func (v *partnerDataValidator) convertValidationErrors(err error) error {
	var errors []ValidationError

	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			errors = append(errors, ValidationError{
				Field:   v.getJSONFieldName(fieldError.Field()),
				Message: v.getValidationErrorMessage(fieldError),
				Value:   fieldError.Value(),
				Tag:     fieldError.Tag(),
			})
		}
	}

	return ValidationErrors{Errors: errors}
}

// getJSONFieldName converts struct field name to JSON field name
func (v *partnerDataValidator) getJSONFieldName(fieldName string) string {
	// Convert CamelCase to snake_case for JSON field names
	switch fieldName {
	case "PostalCode":
		return "postal_code"
	case "PickupPostalCode":
		return "pickup_postal_code"
	case "DeliveryPostalCode":
		return "delivery_postal_code"
	case "CountryCode":
		return "country_code"
	case "ServiceTypeCode":
		return "service_type_code"
	case "ParcelCategory":
		return "parcel_category"
	default:
		return strings.ToLower(fieldName)
	}
}

// getValidationErrorMessage returns user-friendly validation error message
func (v *partnerDataValidator) getValidationErrorMessage(fieldError validator.FieldError) string {
	field := v.getJSONFieldName(fieldError.Field())

	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, fieldError.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, fieldError.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, fieldError.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fieldError.Param())
	case "postal_code":
		return fmt.Sprintf("%s has invalid format", field)
	case "country_code":
		return fmt.Sprintf("%s must be a valid 2-character country code", field)
	case "partner_name":
		return fmt.Sprintf("%s contains invalid characters", field)
	case "service_type":
		return fmt.Sprintf("%s contains invalid service type", field)
	case "parcel_category":
		return fmt.Sprintf("%s contains invalid parcel category", field)
	case "rating":
		return fmt.Sprintf("%s must be between 0 and 10", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// Custom validation functions for validator tags

func validatePostalCode(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return postalCodeRegex.MatchString(strings.ToUpper(value))
}

func validateCountryCode(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return countryCodeRegex.MatchString(strings.ToUpper(value))
}

func validatePartnerName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return partnerNameRegex.MatchString(value)
}

func validateServiceTypeTag(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	allowedTypes := []string{"SDD", "NDD", "Standard", "Express", "vayuquick", "vayuquickpro"}
	for _, allowed := range allowedTypes {
		if value == allowed {
			return true
		}
	}
	return false
}

func validateParcelCategoryTag(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	allowedCategories := []string{"ecom", "courier", "cargo"}
	for _, allowed := range allowedCategories {
		if value == allowed {
			return true
		}
	}
	return false
}

func validateRatingTag(fl validator.FieldLevel) bool {
	value := fl.Field().Float()
	return value >= 0 && value <= 10
}
