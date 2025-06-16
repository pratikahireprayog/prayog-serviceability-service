package errors

import (
	"fmt"
	"time"
)

// ServiceDefinitionError represents errors related to service definition resolution
type ServiceDefinitionError struct {
	Type        ErrorType `json:"type"`
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	Cause       error     `json:"-"`
	Timestamp   time.Time `json:"timestamp"`
	Retryable   bool      `json:"retryable"`
	ServiceName string    `json:"service_name,omitempty"`
}

// ErrorType represents the category of error
type ErrorType string

const (
	// API-related errors
	ErrorTypeAPIConnection   ErrorType = "API_CONNECTION"
	ErrorTypeAPITimeout      ErrorType = "API_TIMEOUT"
	ErrorTypeAPIUnauthorized ErrorType = "API_UNAUTHORIZED"
	ErrorTypeAPINotFound     ErrorType = "API_NOT_FOUND"
	ErrorTypeAPIRateLimit    ErrorType = "API_RATE_LIMIT"
	ErrorTypeAPIServerError  ErrorType = "API_SERVER_ERROR"
	ErrorTypeAPIBadResponse  ErrorType = "API_BAD_RESPONSE"

	// Cache-related errors
	ErrorTypeCacheConnection ErrorType = "CACHE_CONNECTION"
	ErrorTypeCacheTimeout    ErrorType = "CACHE_TIMEOUT"
	ErrorTypeCacheCorrupted  ErrorType = "CACHE_CORRUPTED"
	ErrorTypeCacheEvicted    ErrorType = "CACHE_EVICTED"

	// Data-related errors
	ErrorTypeDataValidation   ErrorType = "DATA_VALIDATION"
	ErrorTypeDataTransform    ErrorType = "DATA_TRANSFORM"
	ErrorTypeDataIncomplete   ErrorType = "DATA_INCOMPLETE"
	ErrorTypeDataInconsistent ErrorType = "DATA_INCONSISTENT"

	// Configuration errors
	ErrorTypeConfiguration  ErrorType = "CONFIGURATION"
	ErrorTypeInitialization ErrorType = "INITIALIZATION"

	// Circuit breaker errors
	ErrorTypeCircuitOpen    ErrorType = "CIRCUIT_OPEN"
	ErrorTypeCircuitTimeout ErrorType = "CIRCUIT_TIMEOUT"

	// Unknown errors
	ErrorTypeUnknown ErrorType = "UNKNOWN"
)

// Error implements the error interface
func (e *ServiceDefinitionError) Error() string {
	if e.ServiceName != "" {
		return fmt.Sprintf("[%s] %s (%s): %s", e.ServiceName, e.Type, e.Code, e.Message)
	}
	return fmt.Sprintf("%s (%s): %s", e.Type, e.Code, e.Message)
}

// Unwrap returns the underlying error for error unwrapping
func (e *ServiceDefinitionError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns whether the error should be retried
func (e *ServiceDefinitionError) IsRetryable() bool {
	return e.Retryable
}

// NewServiceDefinitionError creates a new service definition error
func NewServiceDefinitionError(errorType ErrorType, code, message string, cause error) *ServiceDefinitionError {
	return &ServiceDefinitionError{
		Type:      errorType,
		Code:      code,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
		Retryable: isRetryableErrorType(errorType),
	}
}

// NewAPIError creates an API-related error
func NewAPIError(code, message string, cause error) *ServiceDefinitionError {
	errorType := categorizeAPIError(cause)
	return &ServiceDefinitionError{
		Type:        errorType,
		Code:        code,
		Message:     message,
		Cause:       cause,
		Timestamp:   time.Now(),
		Retryable:   isRetryableErrorType(errorType),
		ServiceName: "specification-service",
	}
}

// NewCacheError creates a cache-related error
func NewCacheError(code, message string, cause error) *ServiceDefinitionError {
	errorType := categorizeCacheError(cause)
	return &ServiceDefinitionError{
		Type:      errorType,
		Code:      code,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
		Retryable: isRetryableErrorType(errorType),
	}
}

// NewDataError creates a data-related error
func NewDataError(code, message string, cause error) *ServiceDefinitionError {
	return &ServiceDefinitionError{
		Type:      ErrorTypeDataValidation,
		Code:      code,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
		Retryable: false, // Data errors are typically not retryable
	}
}

// isRetryableErrorType determines if an error type should be retried
func isRetryableErrorType(errorType ErrorType) bool {
	switch errorType {
	case ErrorTypeAPIConnection,
		ErrorTypeAPITimeout,
		ErrorTypeAPIRateLimit,
		ErrorTypeAPIServerError,
		ErrorTypeCacheConnection,
		ErrorTypeCacheTimeout,
		ErrorTypeCircuitTimeout:
		return true
	case ErrorTypeAPIUnauthorized,
		ErrorTypeAPINotFound,
		ErrorTypeAPIBadResponse,
		ErrorTypeCacheCorrupted,
		ErrorTypeDataValidation,
		ErrorTypeDataTransform,
		ErrorTypeDataIncomplete,
		ErrorTypeDataInconsistent,
		ErrorTypeConfiguration,
		ErrorTypeInitialization,
		ErrorTypeCircuitOpen:
		return false
	default:
		return false
	}
}

// categorizeAPIError categorizes API errors based on the underlying error
func categorizeAPIError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errStr := err.Error()
	switch {
	case contains(errStr, "timeout", "deadline"):
		return ErrorTypeAPITimeout
	case contains(errStr, "connection", "network", "dial"):
		return ErrorTypeAPIConnection
	case contains(errStr, "unauthorized", "401"):
		return ErrorTypeAPIUnauthorized
	case contains(errStr, "not found", "404"):
		return ErrorTypeAPINotFound
	case contains(errStr, "rate limit", "429"):
		return ErrorTypeAPIRateLimit
	case contains(errStr, "server error", "500", "502", "503", "504"):
		return ErrorTypeAPIServerError
	case contains(errStr, "bad response", "invalid response"):
		return ErrorTypeAPIBadResponse
	default:
		return ErrorTypeAPIServerError
	}
}

// categorizeCacheError categorizes cache errors based on the underlying error
func categorizeCacheError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errStr := err.Error()
	switch {
	case contains(errStr, "timeout", "deadline"):
		return ErrorTypeCacheTimeout
	case contains(errStr, "connection", "network"):
		return ErrorTypeCacheConnection
	case contains(errStr, "corrupted", "invalid", "unmarshal"):
		return ErrorTypeCacheCorrupted
	case contains(errStr, "evicted", "expired"):
		return ErrorTypeCacheEvicted
	default:
		return ErrorTypeCacheConnection
	}
}

// contains checks if any of the substrings exist in the main string (case-insensitive)
func contains(str string, substrings ...string) bool {
	str = fmt.Sprintf("%s", str) // Convert to lowercase
	for _, substring := range substrings {
		if len(str) >= len(substring) {
			for i := 0; i <= len(str)-len(substring); i++ {
				match := true
				for j := 0; j < len(substring); j++ {
					if str[i+j] != substring[j] && str[i+j] != substring[j]-32 && str[i+j] != substring[j]+32 {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}

// FallbackStrategy represents different fallback strategies
type FallbackStrategy string

const (
	FallbackStrategyCache   FallbackStrategy = "CACHE"
	FallbackStrategyDefault FallbackStrategy = "DEFAULT"
	FallbackStrategyStale   FallbackStrategy = "STALE"
	FallbackStrategyEmpty   FallbackStrategy = "EMPTY"
	FallbackStrategyRetry   FallbackStrategy = "RETRY"
)

// FallbackResult represents the result of a fallback operation
type FallbackResult struct {
	Strategy    FallbackStrategy `json:"strategy"`
	Success     bool             `json:"success"`
	Data        interface{}      `json:"data,omitempty"`
	Message     string           `json:"message"`
	Timestamp   time.Time        `json:"timestamp"`
	OriginalErr error            `json:"-"`
}

// NewFallbackResult creates a new fallback result
func NewFallbackResult(strategy FallbackStrategy, success bool, data interface{}, message string, originalErr error) *FallbackResult {
	return &FallbackResult{
		Strategy:    strategy,
		Success:     success,
		Data:        data,
		Message:     message,
		Timestamp:   time.Now(),
		OriginalErr: originalErr,
	}
}

// IsServiceDefinitionError checks if an error is a ServiceDefinitionError
func IsServiceDefinitionError(err error) bool {
	_, ok := err.(*ServiceDefinitionError)
	return ok
}

// GetServiceDefinitionError extracts ServiceDefinitionError from error
func GetServiceDefinitionError(err error) (*ServiceDefinitionError, bool) {
	if sdErr, ok := err.(*ServiceDefinitionError); ok {
		return sdErr, true
	}
	return nil, false
}

// WrapError wraps an existing error as a ServiceDefinitionError
func WrapError(err error, errorType ErrorType, code, message string) *ServiceDefinitionError {
	return &ServiceDefinitionError{
		Type:      errorType,
		Code:      code,
		Message:   message,
		Cause:     err,
		Timestamp: time.Now(),
		Retryable: isRetryableErrorType(errorType),
	}
}
