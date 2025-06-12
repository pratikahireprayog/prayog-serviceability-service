package handlers

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/shared/dtos/v1"
)

// ErrorHandler provides centralized error handling for HTTP responses
type ErrorHandler struct {
	logger *logrus.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *logrus.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrorCodeInvalidRequestBody    ErrorCode = "INVALID_REQUEST_BODY"
	ErrorCodeValidationError       ErrorCode = "VALIDATION_ERROR"
	ErrorCodeInvalidRequestType    ErrorCode = "INVALID_REQUEST_TYPE"
	ErrorCodeMissingRequiredFields ErrorCode = "MISSING_REQUIRED_FIELDS"
	ErrorCodeInvalidPostalCode     ErrorCode = "INVALID_POSTAL_CODE"
	ErrorCodeInvalidCountryCode    ErrorCode = "INVALID_COUNTRY_CODE"
	ErrorCodePostalCodeInactive    ErrorCode = "POSTAL_CODE_INACTIVE"
	ErrorCodeRequestTooLarge       ErrorCode = "REQUEST_TOO_LARGE"
	ErrorCodeRateLimitExceeded     ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Server errors (5xx)
	ErrorCodeServiceabilityCheckFailed     ErrorCode = "SERVICEABILITY_CHECK_FAILED"
	ErrorCodeBulkServiceabilityCheckFailed ErrorCode = "BULK_SERVICEABILITY_CHECK_FAILED"
	ErrorCodeInternalServerError           ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorCodeServiceUnavailable            ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeDependencyFailure             ErrorCode = "DEPENDENCY_FAILURE"
	ErrorCodePartialFailure                ErrorCode = "PARTIAL_FAILURE"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code       ErrorCode         `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	RequestID  string            `json:"request_id,omitempty"`
	Timestamp  string            `json:"timestamp,omitempty"`
	Validation []ValidationError `json:"validation_errors,omitempty"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
}

// HandleParsingError handles request body parsing errors
func (eh *ErrorHandler) HandleParsingError(c *fiber.Ctx, err error) error {
	requestID := c.Get("X-Request-ID", "unknown")

	eh.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"method":     c.Method(),
		"path":       c.Path(),
		"error":      err.Error(),
	}).Error("Request body parsing failed")

	return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
		Success: false,
		Message: "Invalid request body",
		Error: dtos.ErrorInfo{
			Code:    "INVALID_REQUEST",
			Message: "Failed to parse request body",
			Details: eh.sanitizeErrorMessage(err.Error()),
		},
	})
}

// HandleValidationError handles struct validation errors
func (eh *ErrorHandler) HandleValidationError(c *fiber.Ctx, err error) error {
	requestID := c.Get("X-Request-ID", "unknown")

	eh.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"method":     c.Method(),
		"path":       c.Path(),
		"error":      err.Error(),
	}).Error("Request validation failed")

	// Parse validation errors for better user experience
	validationErrors := eh.parseValidationErrors(err)

	return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
		Success: false,
		Message: "Invalid request body",
		Error: dtos.ErrorInfo{
			Code:    "INVALID_REQUEST",
			Message: "Request validation failed",
			Details: validationErrors,
		},
	})
}

// HandleBusinessLogicError handles business logic validation errors
func (eh *ErrorHandler) HandleBusinessLogicError(c *fiber.Ctx, code ErrorCode, message string, err error) error {
	requestID := c.Get("X-Request-ID", "unknown")

	eh.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"method":     c.Method(),
		"path":       c.Path(),
		"error_code": code,
		"error":      err.Error(),
	}).Error("Business logic validation failed")

	return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
		Success: false,
		Message: message,
		Error: dtos.ErrorInfo{
			Code:    string(code),
			Message: eh.sanitizeErrorMessage(err.Error()),
		},
	})
}

// HandleServiceError handles service layer errors
func (eh *ErrorHandler) HandleServiceError(c *fiber.Ctx, code ErrorCode, message string, err error) error {
	requestID := c.Get("X-Request-ID", "unknown")

	eh.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"method":     c.Method(),
		"path":       c.Path(),
		"error_code": code,
		"error":      err.Error(),
	}).Error("Service error occurred")

	statusCode := fiber.StatusInternalServerError
	if code == ErrorCodeServiceUnavailable {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(dtos.StandardErrorResponse{
		Success: false,
		Message: message,
		Error: dtos.ErrorInfo{
			Code:    string(code),
			Message: eh.sanitizeErrorMessage(err.Error()),
		},
	})
}

// HandleBulkServiceError handles bulk service errors
func (eh *ErrorHandler) HandleBulkServiceError(c *fiber.Ctx, code ErrorCode, message string, err error) error {
	requestID := c.Get("X-Request-ID", "unknown")

	eh.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"method":     c.Method(),
		"path":       c.Path(),
		"error_code": code,
		"error":      err.Error(),
	}).Error("Bulk service error occurred")

	statusCode := fiber.StatusInternalServerError
	if code == ErrorCodePartialFailure {
		statusCode = fiber.StatusMultiStatus
	}

	return c.Status(statusCode).JSON(dtos.StandardErrorResponse{
		Success: false,
		Message: message,
		Error: dtos.ErrorInfo{
			Code:    string(code),
			Message: eh.sanitizeErrorMessage(err.Error()),
		},
	})
}

// parseValidationErrors parses validator errors into structured format
func (eh *ErrorHandler) parseValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			validationError := ValidationError{
				Field:   eh.getJSONFieldName(fieldError.Field()),
				Value:   fmt.Sprintf("%v", fieldError.Value()),
				Tag:     fieldError.Tag(),
				Message: eh.getValidationErrorMessage(fieldError),
			}
			validationErrors = append(validationErrors, validationError)
		}
	}

	return validationErrors
}

// getValidationErrorMessage returns a user-friendly validation error message
func (eh *ErrorHandler) getValidationErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", eh.getJSONFieldName(fieldError.Field()))
	case "email":
		return fmt.Sprintf("%s must be a valid email address", eh.getJSONFieldName(fieldError.Field()))
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", eh.getJSONFieldName(fieldError.Field()), fieldError.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", eh.getJSONFieldName(fieldError.Field()), fieldError.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", eh.getJSONFieldName(fieldError.Field()), fieldError.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", eh.getJSONFieldName(fieldError.Field()), fieldError.Param())
	default:
		return fmt.Sprintf("%s is invalid", eh.getJSONFieldName(fieldError.Field()))
	}
}

// getJSONFieldName converts struct field name to JSON field name
func (eh *ErrorHandler) getJSONFieldName(fieldName string) string {
	// Convert CamelCase to snake_case for JSON field names
	// This is a simple implementation - could be enhanced with reflection
	switch fieldName {
	case "PostalCode":
		return "postal_code"
	case "PickupPostalCode":
		return "pickup_postal_code"
	case "DeliveryPostalCode":
		return "delivery_postal_code"
	case "CountryCode":
		return "country_code"
	case "ServiceTypes":
		return "service_types"
	case "PartnerIds":
		return "partner_ids"
	default:
		return strings.ToLower(fieldName)
	}
}

// formatValidationErrors formats validation errors into a readable string
func (eh *ErrorHandler) formatValidationErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	var messages []string
	for _, err := range errors {
		messages = append(messages, err.Message)
	}

	return strings.Join(messages, "; ")
}

// sanitizeErrorMessage removes sensitive information from error messages
func (eh *ErrorHandler) sanitizeErrorMessage(message string) string {
	// Remove potential sensitive information from error messages
	// This is a basic implementation - enhance based on security requirements

	// Remove file paths
	if strings.Contains(message, "/") {
		parts := strings.Split(message, ":")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[len(parts)-1])
		}
	}

	return message
}

// GetStatusCodeForError returns appropriate HTTP status code for error code
func (eh *ErrorHandler) GetStatusCodeForError(code ErrorCode) int {
	switch code {
	case ErrorCodeInvalidRequestBody, ErrorCodeValidationError, ErrorCodeInvalidRequestType,
		ErrorCodeMissingRequiredFields, ErrorCodeInvalidPostalCode, ErrorCodeInvalidCountryCode:
		return fiber.StatusBadRequest
	case ErrorCodePostalCodeInactive:
		return fiber.StatusUnprocessableEntity
	case ErrorCodeRequestTooLarge:
		return fiber.StatusRequestEntityTooLarge
	case ErrorCodeRateLimitExceeded:
		return fiber.StatusTooManyRequests
	case ErrorCodeServiceUnavailable:
		return fiber.StatusServiceUnavailable
	case ErrorCodePartialFailure:
		return fiber.StatusMultiStatus
	default:
		return fiber.StatusInternalServerError
	}
}
