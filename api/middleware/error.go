package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"prayog-serviceability-service/internal/shared/constants/v1"
	"prayog-serviceability-service/internal/shared/dtos/v1"
)

// ErrorType represents different types of errors that can occur
type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeAuthentication
	ErrorTypeAuthorization
	ErrorTypeNotFound
	ErrorTypeConflict
	ErrorTypeRateLimit
	ErrorTypeTimeout
	ErrorTypeDatabase
	ErrorTypeExternalService
	ErrorTypeInternal
	ErrorTypePanic
)

// AppError represents a structured application error
type AppError struct {
	Type       ErrorType   `json:"type"`
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	StatusCode int         `json:"status_code"`
	Cause      error       `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, code, message string, statusCode int) *AppError {
	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, details interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Code:       constants.ErrorCodeValidationError,
		Message:    message,
		Details:    details,
		StatusCode: constants.StatusBadRequest,
	}
}

// NewAuthenticationError creates an authentication error
func NewAuthenticationError(code, message string) *AppError {
	return &AppError{
		Type:       ErrorTypeAuthentication,
		Code:       code,
		Message:    message,
		StatusCode: constants.StatusUnauthorized,
	}
}

// NewAuthorizationError creates an authorization error
func NewAuthorizationError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeAuthorization,
		Code:       constants.ErrorCodeInsufficientPermissions,
		Message:    message,
		StatusCode: constants.StatusForbidden,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Code:       constants.ErrorCodeInvalidRequest,
		Message:    message,
		StatusCode: constants.StatusNotFound,
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Code:       constants.ErrorCodeInternalServerError,
		Message:    message,
		StatusCode: constants.StatusInternalServerError,
		Cause:      cause,
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(message string, cause error) *AppError {
	return &AppError{
		Type:       ErrorTypeDatabase,
		Code:       constants.ErrorCodeDatabaseError,
		Message:    message,
		StatusCode: constants.StatusInternalServerError,
		Cause:      cause,
	}
}

// NewExternalServiceError creates an external service error
func NewExternalServiceError(service, message string, cause error) *AppError {
	var code string
	switch strings.ToLower(service) {
	case "partner":
		code = constants.ErrorCodePartnerServiceError
	case "specification":
		code = constants.ErrorCodeSpecificationServiceError
	default:
		code = constants.ErrorCodeServiceUnavailable
	}

	return &AppError{
		Type:       ErrorTypeExternalService,
		Code:       code,
		Message:    message,
		StatusCode: constants.StatusBadGateway,
		Cause:      cause,
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeTimeout,
		Code:       constants.ErrorCodeTimeoutError,
		Message:    message,
		StatusCode: constants.StatusGatewayTimeout,
	}
}

// ErrorHandler creates middleware that handles all types of errors
func ErrorHandler(logger *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a custom response writer to capture errors
			ew := &errorResponseWriter{
				ResponseWriter: w,
				logger:         logger,
			}

			// Handle panics
			defer func() {
				if err := recover(); err != nil {
					ew.handlePanic(err)
				}
			}()

			// Store the current request in the error writer for context tracking
			ew.currentRequest = r

			// Add the error writer to the request context so helper functions can access it
			ctx := context.WithValue(r.Context(), ErrorWriterKey, ew)
			rWithErrorWriter := r.WithContext(ctx)
			ew.currentRequest = rWithErrorWriter

			// Call the next handler
			next.ServeHTTP(ew, rWithErrorWriter)

			// Check if an error was set in context after handler completion
			if appErr := GetErrorFromContext(ew.currentRequest.Context()); appErr != nil {
				ew.handleAppError(appErr)
			}
		})
	}
}

// errorResponseWriter wraps http.ResponseWriter to handle errors
type errorResponseWriter struct {
	http.ResponseWriter
	currentRequest *http.Request
	logger         *log.Logger
	headerWritten  bool
}

// updateRequestContext updates the current request context (used by error setters)
func (ew *errorResponseWriter) updateRequestContext(newRequest *http.Request) {
	ew.currentRequest = newRequest
}

// WriteHeader captures the status code and handles error responses
func (ew *errorResponseWriter) WriteHeader(code int) {
	if ew.headerWritten {
		return
	}
	ew.headerWritten = true

	// If it's an error status code and no error was explicitly handled
	if code >= 400 && GetErrorFromContext(ew.currentRequest.Context()) == nil {
		ew.handleHTTPError(code)
		return
	}

	ew.ResponseWriter.WriteHeader(code)
}

// handlePanic handles panic recovery
func (ew *errorResponseWriter) handlePanic(panicValue interface{}) {
	if ew.logger != nil {
		ew.logger.Printf("Panic recovered: %v\nStack trace:\n%s", panicValue, debug.Stack())
	}

	appErr := &AppError{
		Type:       ErrorTypePanic,
		Code:       constants.ErrorCodeInternalServerError,
		Message:    constants.MsgInternalServerError,
		StatusCode: constants.StatusInternalServerError,
		Details:    fmt.Sprintf("Panic: %v", panicValue),
	}

	ew.sendErrorResponse(appErr)
}

// handleAppError handles structured application errors
func (ew *errorResponseWriter) handleAppError(appErr *AppError) {
	if ew.headerWritten {
		return
	}

	if ew.logger != nil && appErr.Type == ErrorTypeInternal {
		ew.logger.Printf("Application error: %s (Code: %s, Type: %d)", appErr.Message, appErr.Code, appErr.Type)
		if appErr.Cause != nil {
			ew.logger.Printf("Caused by: %v", appErr.Cause)
		}
	}

	ew.sendErrorResponse(appErr)
}

// handleHTTPError handles standard HTTP error status codes
func (ew *errorResponseWriter) handleHTTPError(statusCode int) {
	var appErr *AppError

	switch statusCode {
	case constants.StatusBadRequest:
		appErr = NewValidationError(constants.MsgValidationError, nil)
	case constants.StatusUnauthorized:
		appErr = NewAuthenticationError(constants.ErrorCodeAuthFailed, constants.MsgAuthFailed)
	case constants.StatusForbidden:
		appErr = NewAuthorizationError(constants.MsgInsufficientPermissions)
	case constants.StatusNotFound:
		appErr = NewNotFoundError("Resource not found")
	case constants.StatusMethodNotAllowed:
		appErr = NewAppError(ErrorTypeValidation, constants.ErrorCodeInvalidRequest, "Method not allowed", statusCode)
	case constants.StatusConflict:
		appErr = NewAppError(ErrorTypeConflict, constants.ErrorCodeInvalidRequest, "Resource conflict", statusCode)
	case constants.StatusTooManyRequests:
		appErr = NewAppError(ErrorTypeRateLimit, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded", statusCode)
	case constants.StatusInternalServerError:
		appErr = NewInternalError(constants.MsgInternalServerError, nil)
	case constants.StatusBadGateway:
		appErr = NewExternalServiceError("external", constants.MsgServiceUnavailable, nil)
	case constants.StatusServiceUnavailable:
		appErr = NewAppError(ErrorTypeExternalService, constants.ErrorCodeServiceUnavailable, constants.MsgServiceUnavailable, statusCode)
	case constants.StatusGatewayTimeout:
		appErr = NewTimeoutError(constants.MsgTimeoutError)
	default:
		appErr = NewInternalError("Unknown error", nil)
		appErr.StatusCode = statusCode
	}

	ew.sendErrorResponse(appErr)
}

// sendErrorResponse sends a standardized error response
func (ew *errorResponseWriter) sendErrorResponse(appErr *AppError) {
	response := dtos.StandardErrorResponse{
		Success: false,
		Message: appErr.Message,
		Error: dtos.ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	}

	ew.Header().Set("Content-Type", "application/json")
	ew.WriteHeader(appErr.StatusCode)

	// Use JSON encoder to write response
	if err := json.NewEncoder(ew).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// Context key for storing errors
type errorContextKey struct{}

var ErrorContextKey = &errorContextKey{}

// ErrorWriterKey is used to store the error writer in context
type errorWriterKey struct{}

var ErrorWriterKey = &errorWriterKey{}

// SetErrorInContext sets an error in the request context
func SetErrorInContext(ctx context.Context, err *AppError) context.Context {
	return context.WithValue(ctx, ErrorContextKey, err)
}

// GetErrorFromContext gets an error from the request context
func GetErrorFromContext(ctx context.Context) *AppError {
	if err, ok := ctx.Value(ErrorContextKey).(*AppError); ok {
		return err
	}
	return nil
}

// Helper functions for handlers to set errors in context

// SetValidationError sets a validation error in context
func SetValidationError(r *http.Request, message string, details interface{}) *http.Request {
	err := NewValidationError(message, details)
	ctx := SetErrorInContext(r.Context(), err)
	newRequest := r.WithContext(ctx)

	// Update the error writer's current request if available
	if ew, ok := r.Context().Value(ErrorWriterKey).(*errorResponseWriter); ok {
		ew.updateRequestContext(newRequest)
	}

	return newRequest
}

// SetAuthenticationError sets an authentication error in context
func SetAuthenticationError(r *http.Request, code, message string) *http.Request {
	err := NewAuthenticationError(code, message)
	ctx := SetErrorInContext(r.Context(), err)
	newRequest := r.WithContext(ctx)

	// Update the error writer's current request if available
	if ew, ok := r.Context().Value(ErrorWriterKey).(*errorResponseWriter); ok {
		ew.updateRequestContext(newRequest)
	}

	return newRequest
}

// SetAuthorizationError sets an authorization error in context
func SetAuthorizationError(r *http.Request, message string) *http.Request {
	err := NewAuthorizationError(message)
	ctx := SetErrorInContext(r.Context(), err)
	newRequest := r.WithContext(ctx)

	// Update the error writer's current request if available
	if ew, ok := r.Context().Value(ErrorWriterKey).(*errorResponseWriter); ok {
		ew.updateRequestContext(newRequest)
	}

	return newRequest
}

// SetNotFoundError sets a not found error in context
func SetNotFoundError(r *http.Request, message string) *http.Request {
	err := NewNotFoundError(message)
	ctx := SetErrorInContext(r.Context(), err)
	newRequest := r.WithContext(ctx)

	// Update the error writer's current request if available
	if ew, ok := r.Context().Value(ErrorWriterKey).(*errorResponseWriter); ok {
		ew.updateRequestContext(newRequest)
	}

	return newRequest
}

// SetInternalError sets an internal error in context
func SetInternalError(r *http.Request, message string, cause error) *http.Request {
	err := NewInternalError(message, cause)
	ctx := SetErrorInContext(r.Context(), err)
	newRequest := r.WithContext(ctx)

	// Update the error writer's current request if available
	if ew, ok := r.Context().Value(ErrorWriterKey).(*errorResponseWriter); ok {
		ew.updateRequestContext(newRequest)
	}

	return newRequest
}

// SetDatabaseError sets a database error in context
func SetDatabaseError(r *http.Request, message string, cause error) *http.Request {
	err := NewDatabaseError(message, cause)
	ctx := SetErrorInContext(r.Context(), err)
	newRequest := r.WithContext(ctx)

	// Update the error writer's current request if available
	if ew, ok := r.Context().Value(ErrorWriterKey).(*errorResponseWriter); ok {
		ew.updateRequestContext(newRequest)
	}

	return newRequest
}

// SetExternalServiceError sets an external service error in context
func SetExternalServiceError(r *http.Request, service, message string, cause error) *http.Request {
	err := NewExternalServiceError(service, message, cause)
	ctx := SetErrorInContext(r.Context(), err)
	newRequest := r.WithContext(ctx)

	// Update the error writer's current request if available
	if ew, ok := r.Context().Value(ErrorWriterKey).(*errorResponseWriter); ok {
		ew.updateRequestContext(newRequest)
	}

	return newRequest
}

// SetTimeoutError sets a timeout error in context
func SetTimeoutError(r *http.Request, message string) *http.Request {
	err := NewTimeoutError(message)
	ctx := SetErrorInContext(r.Context(), err)
	newRequest := r.WithContext(ctx)

	// Update the error writer's current request if available
	if ew, ok := r.Context().Value(ErrorWriterKey).(*errorResponseWriter); ok {
		ew.updateRequestContext(newRequest)
	}

	return newRequest
}
