package errors

import (
	"fmt"
	"net/http"
)

// ServiceError represents a structured error with HTTP status information
type ServiceError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	HTTPStatus int    `json:"-"`
	Cause      error  `json:"-"`
}

// Error implements the error interface
func (e *ServiceError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *ServiceError) Unwrap() error {
	return e.Cause
}

// NewServiceError creates a new service error
func NewServiceError(code, message string, httpStatus int, cause error) *ServiceError {
	return &ServiceError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Cause:      cause,
	}
}

// WithDetails adds details to the error
func (e *ServiceError) WithDetails(details string) *ServiceError {
	e.Details = details
	return e
}

// Predefined error constructors for common cases

// ErrNotFound creates a 404 Not Found error
func ErrNotFound(resource, identifier string) *ServiceError {
	return &ServiceError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s not found", resource),
		Details:    fmt.Sprintf("%s with identifier '%s' does not exist", resource, identifier),
		HTTPStatus: http.StatusNotFound,
	}
}

// ErrPostalCodeNotFound creates a 404 error for postal codes
func ErrPostalCodeNotFound(postalCode string) *ServiceError {
	return &ServiceError{
		Code:       "POSTAL_CODE_NOT_FOUND",
		Message:    "Postal code not found",
		Details:    fmt.Sprintf("Postal code '%s' does not exist in our database", postalCode),
		HTTPStatus: http.StatusNotFound,
	}
}

// ErrPostalCodeInactive creates a 422 error for inactive postal codes
func ErrPostalCodeInactive(postalCode string) *ServiceError {
	return &ServiceError{
		Code:       "POSTAL_CODE_INACTIVE",
		Message:    "Postal code is not serviceable",
		Details:    fmt.Sprintf("Postal code '%s' is inactive and not serviceable", postalCode),
		HTTPStatus: http.StatusUnprocessableEntity,
	}
}

// ErrInvalidRequest creates a 400 Bad Request error
func ErrInvalidRequest(message string) *ServiceError {
	return &ServiceError{
		Code:       "INVALID_REQUEST",
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrValidationFailed creates a 400 Bad Request error for validation failures
func ErrValidationFailed(field, message string) *ServiceError {
	return &ServiceError{
		Code:       "VALIDATION_FAILED",
		Message:    "Request validation failed",
		Details:    fmt.Sprintf("Field '%s': %s", field, message),
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrMissingRequiredField creates a 400 Bad Request error for missing required fields
func ErrMissingRequiredField(field string) *ServiceError {
	return &ServiceError{
		Code:       "MISSING_REQUIRED_FIELD",
		Message:    "Required field missing",
		Details:    fmt.Sprintf("Field '%s' is required", field),
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrInvalidCountryCode creates a 400 Bad Request error for invalid country codes
func ErrInvalidCountryCode(countryCode string) *ServiceError {
	return &ServiceError{
		Code:       "INVALID_COUNTRY_CODE",
		Message:    "Invalid country code",
		Details:    fmt.Sprintf("Country code '%s' is not valid", countryCode),
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrInvalidParcelCategory creates a 400 Bad Request error for invalid parcel categories
func ErrInvalidParcelCategory(category string) *ServiceError {
	return &ServiceError{
		Code:       "INVALID_PARCEL_CATEGORY",
		Message:    "Invalid parcel category",
		Details:    fmt.Sprintf("Parcel category '%s' is not supported", category),
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrServiceUnavailable creates a 503 Service Unavailable error
func ErrServiceUnavailable(service string) *ServiceError {
	return &ServiceError{
		Code:       "SERVICE_UNAVAILABLE",
		Message:    "Service temporarily unavailable",
		Details:    fmt.Sprintf("Service '%s' is currently unavailable", service),
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

// ErrPartnerUnavailable creates a 503 error for partner unavailability
func ErrPartnerUnavailable(partnerCode string) *ServiceError {
	return &ServiceError{
		Code:       "PARTNER_UNAVAILABLE",
		Message:    "Partner service unavailable",
		Details:    fmt.Sprintf("Partner '%s' is currently unavailable", partnerCode),
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

// ErrInternalError creates a 500 Internal Server Error
func ErrInternalError(message string, cause error) *ServiceError {
	return &ServiceError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Cause:      cause,
	}
}

// ErrDatabaseError creates a 500 error for database issues
func ErrDatabaseError(operation string, cause error) *ServiceError {
	return &ServiceError{
		Code:       "DATABASE_ERROR",
		Message:    "Database operation failed",
		Details:    fmt.Sprintf("Failed to %s", operation),
		HTTPStatus: http.StatusInternalServerError,
		Cause:      cause,
	}
}

// ErrTimeout creates a 504 Gateway Timeout error
func ErrTimeout(operation string) *ServiceError {
	return &ServiceError{
		Code:       "TIMEOUT",
		Message:    "Operation timed out",
		Details:    fmt.Sprintf("Operation '%s' exceeded timeout limit", operation),
		HTTPStatus: http.StatusGatewayTimeout,
	}
}

// ErrNoServiceablePartners creates a 422 error when no partners can service a request
func ErrNoServiceablePartners() *ServiceError {
	return &ServiceError{
		Code:       "NO_SERVICEABLE_PARTNERS",
		Message:    "No partners available to service this request",
		HTTPStatus: http.StatusUnprocessableEntity,
	}
}

// ErrPartnerNotFound creates a 404 error for partner not found
func ErrPartnerNotFound(partnerCode string) *ServiceError {
	return &ServiceError{
		Code:       "PARTNER_NOT_FOUND",
		Message:    "Partner not found",
		Details:    fmt.Sprintf("Partner with code '%s' does not exist or is inactive", partnerCode),
		HTTPStatus: http.StatusNotFound,
	}
}

// ErrInvalidPostalCodeFormat creates a 400 error for invalid postal code format
func ErrInvalidPostalCodeFormat(postalCode string) *ServiceError {
	return &ServiceError{
		Code:       "INVALID_POSTAL_CODE_FORMAT",
		Message:    "Invalid postal code format",
		Details:    fmt.Sprintf("Postal code '%s' has invalid format", postalCode),
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrInternationalPackageRequired creates a 400 error for missing package info in international requests
func ErrInternationalPackageRequired() *ServiceError {
	return &ServiceError{
		Code:       "INTERNATIONAL_PACKAGE_REQUIRED",
		Message:    "Package information required for international shipments",
		Details:    "Weight and dimensions are mandatory for international shipping",
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrInvalidPackageWeight creates a 400 error for invalid package weight
func ErrInvalidPackageWeight(weight float64) *ServiceError {
	return &ServiceError{
		Code:       "INVALID_PACKAGE_WEIGHT",
		Message:    "Invalid package weight",
		Details:    fmt.Sprintf("Package weight %.2f is invalid, must be greater than 0", weight),
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrInvalidPackageDimensions creates a 400 error for invalid package dimensions
func ErrInvalidPackageDimensions() *ServiceError {
	return &ServiceError{
		Code:       "INVALID_PACKAGE_DIMENSIONS",
		Message:    "Invalid package dimensions",
		Details:    "Package dimensions must be greater than 0",
		HTTPStatus: http.StatusBadRequest,
	}
}

// Helper functions to check error types

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	if serviceErr, ok := err.(*ServiceError); ok {
		return serviceErr.HTTPStatus == http.StatusNotFound
	}
	return false
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if serviceErr, ok := err.(*ServiceError); ok {
		return serviceErr.HTTPStatus == http.StatusBadRequest
	}
	return false
}

// IsInternalError checks if an error is an internal server error
func IsInternalError(err error) bool {
	if serviceErr, ok := err.(*ServiceError); ok {
		return serviceErr.HTTPStatus == http.StatusInternalServerError
	}
	return false
}

// IsServiceUnavailable checks if an error is a service unavailable error
func IsServiceUnavailable(err error) bool {
	if serviceErr, ok := err.(*ServiceError); ok {
		return serviceErr.HTTPStatus == http.StatusServiceUnavailable
	}
	return false
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	if serviceErr, ok := err.(*ServiceError); ok {
		return serviceErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetErrorCode returns the error code for an error
func GetErrorCode(err error) string {
	if serviceErr, ok := err.(*ServiceError); ok {
		return serviceErr.Code
	}
	return "UNKNOWN_ERROR"
}
