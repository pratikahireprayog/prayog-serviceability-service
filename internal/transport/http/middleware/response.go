package middleware

import (
	"context"
	"encoding/json"
	"net/http"
)

// EnhancedResponseWriter is an enhanced http.ResponseWriter that captures response data
type EnhancedResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

// WriteHeader captures the status code and calls the wrapped ResponseWriter
func (rw *EnhancedResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body and calls the wrapped ResponseWriter
func (rw *EnhancedResponseWriter) Write(b []byte) (int, error) {
	rw.body = b
	return rw.ResponseWriter.Write(b)
}

// StatusCode returns the HTTP status code
func (rw *EnhancedResponseWriter) StatusCode() int {
	return rw.statusCode
}

// ResponseContext is the key type for the response context
type ResponseContext string

// Response context keys
const (
	ResponseCtxKey ResponseContext = "response"
)

// ResponseData represents standardized response data
type ResponseData struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// Response middleware standardizes all API responses
func Response() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create enhanced response writer
			rw := &EnhancedResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Create context with response data
			respData := &ResponseData{
				Status: "success",
			}
			ctx := context.WithValue(r.Context(), ResponseCtxKey, respData)

			// Call next handler with enhanced context
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
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

// RespondWithJSON sends a standardized JSON success response
func RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	response := &ResponseData{
		Status: "success",
		Data:   data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RespondWithError sends a standardized JSON error response
func RespondWithError(w http.ResponseWriter, code int, message string, errorCode string) {
	response := &ResponseData{
		Status:  "error",
		Message: message,
		Code:    errorCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// Common error response helpers

// RespondWithValidationError responds with a 400 Bad Request
func RespondWithValidationError(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusBadRequest, message, "validation_error")
}

// RespondWithNotFound responds with a 404 Not Found
func RespondWithNotFound(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusNotFound, message, "not_found")
}

// RespondWithInternalError responds with a 500 Internal Server Error
func RespondWithInternalError(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusInternalServerError, message, "internal_error")
}

// RespondWithUnauthorized responds with a 401 Unauthorized
func RespondWithUnauthorized(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusUnauthorized, message, "unauthorized")
}

// RespondWithForbidden responds with a 403 Forbidden
func RespondWithForbidden(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusForbidden, message, "forbidden")
}
