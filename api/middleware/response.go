package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"prayog-serviceability-service/internal/shared/constants/v1"
)

// EnhancedResponseWriter is an enhanced http.ResponseWriter that captures response data
type EnhancedResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	body          []byte
	written       bool
	headerWritten bool
}

// WriteHeader captures the status code but delays writing to underlying writer
func (rw *EnhancedResponseWriter) WriteHeader(code int) {
	if rw.headerWritten {
		return
	}
	rw.statusCode = code
	rw.headerWritten = true
}

// Write captures the response body but delays writing to underlying writer
func (rw *EnhancedResponseWriter) Write(b []byte) (int, error) {
	if !rw.headerWritten {
		rw.WriteHeader(constants.StatusOK)
	}
	rw.body = append(rw.body, b...)
	rw.written = true
	return len(b), nil
}

// StatusCode returns the HTTP status code
func (rw *EnhancedResponseWriter) StatusCode() int {
	if rw.statusCode == 0 {
		return constants.StatusOK
	}
	return rw.statusCode
}

// Body returns the captured response body
func (rw *EnhancedResponseWriter) Body() []byte {
	return rw.body
}

// Flush writes the captured response to the underlying ResponseWriter
func (rw *EnhancedResponseWriter) Flush() {
	// Write headers first
	for k, v := range rw.Header() {
		rw.ResponseWriter.Header()[k] = v
	}

	// Write status code
	rw.ResponseWriter.WriteHeader(rw.StatusCode())

	// Write body
	if len(rw.body) > 0 {
		rw.ResponseWriter.Write(rw.body)
	}
}

// ResponseContext is the key type for the response context
type ResponseContext string

// Response context keys
const (
	ResponseCtxKey ResponseContext = "response"
)

// StandardResponse represents the standardized API response structure
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error details in the response
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ResponseData represents internal response data for context
type ResponseData struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// SuccessResponseHandler creates middleware that automatically formats successful responses
func SuccessResponseHandler() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create enhanced response writer to capture response
			rw := &EnhancedResponseWriter{
				ResponseWriter: w,
			}

			// Create context with response data
			respData := &ResponseData{
				Status: "success",
			}
			ctx := context.WithValue(r.Context(), ResponseCtxKey, respData)

			// Call next handler with enhanced context
			next.ServeHTTP(rw, r.WithContext(ctx))

			// Format and flush response
			if rw.written {
				// Only format response if it's a success status and not already formatted
				if rw.StatusCode() >= 200 && rw.StatusCode() < 300 {
					if !isAlreadyFormatted(rw.Body()) {
						formatSuccessResponse(rw, respData)
					}
				}
			}

			// Always flush the response
			rw.Flush()
		})
	}
}

// isAlreadyFormatted checks if the response is already in our standard format
func isAlreadyFormatted(body []byte) bool {
	if len(body) == 0 {
		return false
	}

	var response StandardResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return false
	}

	// Check if it has our standard structure (success field exists)
	return json.Valid(body) && (strings.Contains(string(body), `"success":`) || strings.Contains(string(body), `"error":`))
}

// formatSuccessResponse formats the captured response into standard format
func formatSuccessResponse(rw *EnhancedResponseWriter, respData *ResponseData) {
	var data interface{}

	// Try to parse existing response body as JSON
	if len(rw.Body()) > 0 {
		if err := json.Unmarshal(rw.Body(), &data); err != nil {
			// If not valid JSON, treat as string
			data = string(rw.Body())
		}
	}

	// Use data from context if available, otherwise use captured body
	if respData.Data != nil {
		data = respData.Data
	}

	// Create standardized response
	response := StandardResponse{
		Success: true,
		Message: respData.Message,
		Data:    data,
	}

	// Marshal the standardized response
	if jsonData, err := json.Marshal(response); err == nil {
		// Create a new response writer to avoid double writing
		rw.ResponseWriter.Header().Set("Content-Type", "application/json")
		// Clear the body buffer and replace with formatted response
		rw.body = jsonData
	}
}

// Response middleware standardizes all API responses (legacy function for backward compatibility)
func Response() Middleware {
	return SuccessResponseHandler()
}

// Helper functions for handlers to use

// GetResponseData gets the response data from context
func GetResponseData(r *http.Request) *ResponseData {
	if respData, ok := r.Context().Value(ResponseCtxKey).(*ResponseData); ok {
		return respData
	}
	// Fallback if middleware is not used
	return &ResponseData{Status: "success"}
}

// SetResponseData sets data in the response context
func SetResponseData(r *http.Request, data interface{}) {
	if respData := GetResponseData(r); respData != nil {
		respData.Data = data
	}
}

// SetResponseMessage sets message in the response context
func SetResponseMessage(r *http.Request, message string) {
	if respData := GetResponseData(r); respData != nil {
		respData.Message = message
	}
}

// RespondWithJSON sends a standardized JSON success response
func RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	response := StandardResponse{
		Success: true,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RespondWithSuccess sends a standardized JSON success response with message
func RespondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	response := StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RespondWithError sends a standardized JSON error response
func RespondWithError(w http.ResponseWriter, code int, message string, errorCode string) {
	response := StandardResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RespondWithErrorDetails sends a standardized JSON error response with details
func RespondWithErrorDetails(w http.ResponseWriter, code int, message string, errorCode string, details interface{}) {
	response := StandardResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// Common error response helpers using constants

// RespondWithValidationError responds with a 400 Bad Request
func RespondWithValidationError(w http.ResponseWriter, message string) {
	RespondWithError(w, constants.StatusBadRequest, message, constants.ErrorCodeValidationError)
}

// RespondWithNotFound responds with a 404 Not Found
func RespondWithNotFound(w http.ResponseWriter, message string) {
	RespondWithError(w, constants.StatusNotFound, message, constants.ErrorCodeInvalidRequest)
}

// RespondWithInternalError responds with a 500 Internal Server Error
func RespondWithInternalError(w http.ResponseWriter, message string) {
	RespondWithError(w, constants.StatusInternalServerError, message, constants.ErrorCodeInternalServerError)
}

// RespondWithUnauthorized responds with a 401 Unauthorized
func RespondWithUnauthorized(w http.ResponseWriter, message string) {
	RespondWithError(w, constants.StatusUnauthorized, message, constants.ErrorCodeAuthFailed)
}

// RespondWithForbidden responds with a 403 Forbidden
func RespondWithForbidden(w http.ResponseWriter, message string) {
	RespondWithError(w, constants.StatusForbidden, message, constants.ErrorCodeInsufficientPermissions)
}
