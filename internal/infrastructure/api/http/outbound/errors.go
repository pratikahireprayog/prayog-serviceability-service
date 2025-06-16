package outbound

import (
	"fmt"
	"net/http"
)

// HTTPError represents an HTTP-specific error
type HTTPError struct {
	StatusCode int    `json:"status_code"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	Cause      error  `json:"-"`
}

// Error implements error interface
func (e *HTTPError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("HTTP %d %s: %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error unwrapping
func (e *HTTPError) Unwrap() error {
	return e.Cause
}

// IsRetryable determines if the HTTP error should be retried
func (e *HTTPError) IsRetryable() bool {
	return IsRetryableStatusCode(e.StatusCode)
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(statusCode int, code, message string, cause error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Cause:      cause,
	}
}

// NewHTTPErrorFromResponse creates an HTTP error from response
func NewHTTPErrorFromResponse(statusCode int, body string) *HTTPError {
	code := GetErrorCodeFromStatus(statusCode)
	message := GetErrorMessageFromStatus(statusCode)

	return &HTTPError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Details:    body,
	}
}

// IsRetryableStatusCode determines if an HTTP status code should be retried
func IsRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusRequestTimeout, // 408
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

// GetErrorCodeFromStatus returns error code based on HTTP status
func GetErrorCodeFromStatus(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case http.StatusRequestTimeout:
		return "REQUEST_TIMEOUT"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	case http.StatusTooManyRequests:
		return "TOO_MANY_REQUESTS"
	case http.StatusInternalServerError:
		return "INTERNAL_SERVER_ERROR"
	case http.StatusBadGateway:
		return "BAD_GATEWAY"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case http.StatusGatewayTimeout:
		return "GATEWAY_TIMEOUT"
	default:
		if statusCode >= 400 && statusCode < 500 {
			return "CLIENT_ERROR"
		} else if statusCode >= 500 {
			return "SERVER_ERROR"
		}
		return "UNKNOWN_ERROR"
	}
}

// GetErrorMessageFromStatus returns error message based on HTTP status
func GetErrorMessageFromStatus(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "Bad request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Forbidden"
	case http.StatusNotFound:
		return "Not found"
	case http.StatusMethodNotAllowed:
		return "Method not allowed"
	case http.StatusRequestTimeout:
		return "Request timeout"
	case http.StatusConflict:
		return "Conflict"
	case http.StatusUnprocessableEntity:
		return "Unprocessable entity"
	case http.StatusTooManyRequests:
		return "Too many requests"
	case http.StatusInternalServerError:
		return "Internal server error"
	case http.StatusBadGateway:
		return "Bad gateway"
	case http.StatusServiceUnavailable:
		return "Service unavailable"
	case http.StatusGatewayTimeout:
		return "Gateway timeout"
	default:
		return http.StatusText(statusCode)
	}
}

// IsHTTPError checks if error is an HTTP error
func IsHTTPError(err error) bool {
	_, ok := err.(*HTTPError)
	return ok
}

// GetHTTPStatusCode extracts HTTP status code from error
func GetHTTPStatusCode(err error) int {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode
	}
	return 0
}
