package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"prayog-serviceability-service/internal/shared/constants/v1"
)

// Helper function to create a test logger
func createTestLogger() *log.Logger {
	return log.New(io.Discard, "", 0) // Discard output during tests
}

// Helper function to parse error response
func parseErrorResponse(t *testing.T, body string) map[string]interface{} {
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}
	return response
}

// Helper function to validate error response structure
func validateErrorResponse(t *testing.T, response map[string]interface{}, expectedCode string, expectedMessage string, expectedStatus int) {
	// Check success field
	if success, ok := response["success"].(bool); !ok || success {
		t.Errorf("Expected success to be false, got %v", response["success"])
	}

	// Check message field
	if message, ok := response["message"].(string); !ok || message != expectedMessage {
		t.Errorf("Expected message '%s', got '%v'", expectedMessage, response["message"])
	}

	// Check error object
	errorObj, ok := response["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected error object, got %v", response["error"])
	}

	// Check error code
	if code, ok := errorObj["code"].(string); !ok || code != expectedCode {
		t.Errorf("Expected error code '%s', got '%v'", expectedCode, errorObj["code"])
	}

	// Check error message
	if errMsg, ok := errorObj["message"].(string); !ok || errMsg != expectedMessage {
		t.Errorf("Expected error message '%s', got '%v'", expectedMessage, errorObj["message"])
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "Error without cause",
			appErr: &AppError{
				Message: "Test error message",
			},
			expected: "Test error message",
		},
		{
			name: "Error with cause",
			appErr: &AppError{
				Message: "Test error message",
				Cause:   errors.New("underlying error"),
			},
			expected: "Test error message: underlying error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appErr.Error()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNewAppError(t *testing.T) {
	err := NewAppError(ErrorTypeValidation, "TEST_CODE", "Test message", 400)

	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected type %v, got %v", ErrorTypeValidation, err.Type)
	}
	if err.Code != "TEST_CODE" {
		t.Errorf("Expected code 'TEST_CODE', got '%s'", err.Code)
	}
	if err.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", err.Message)
	}
	if err.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", err.StatusCode)
	}
}

func TestNewValidationError(t *testing.T) {
	details := map[string]string{"field": "required"}
	err := NewValidationError("Validation failed", details)

	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected type %v, got %v", ErrorTypeValidation, err.Type)
	}
	if err.Code != constants.ErrorCodeValidationError {
		t.Errorf("Expected code '%s', got '%s'", constants.ErrorCodeValidationError, err.Code)
	}
	if err.StatusCode != constants.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", constants.StatusBadRequest, err.StatusCode)
	}
	// Check that details are properly set (can't directly compare maps)
	if err.Details == nil {
		t.Error("Expected details to be set")
	} else if detailsMap, ok := err.Details.(map[string]string); !ok {
		t.Error("Expected details to be a map[string]string")
	} else if detailsMap["field"] != "required" {
		t.Errorf("Expected field detail 'required', got '%v'", detailsMap["field"])
	}
}

func TestNewAuthenticationError(t *testing.T) {
	err := NewAuthenticationError(constants.ErrorCodeInvalidToken, "Invalid token")

	if err.Type != ErrorTypeAuthentication {
		t.Errorf("Expected type %v, got %v", ErrorTypeAuthentication, err.Type)
	}
	if err.Code != constants.ErrorCodeInvalidToken {
		t.Errorf("Expected code '%s', got '%s'", constants.ErrorCodeInvalidToken, err.Code)
	}
	if err.StatusCode != constants.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", constants.StatusUnauthorized, err.StatusCode)
	}
}

func TestNewAuthorizationError(t *testing.T) {
	err := NewAuthorizationError("Insufficient permissions")

	if err.Type != ErrorTypeAuthorization {
		t.Errorf("Expected type %v, got %v", ErrorTypeAuthorization, err.Type)
	}
	if err.Code != constants.ErrorCodeInsufficientPermissions {
		t.Errorf("Expected code '%s', got '%s'", constants.ErrorCodeInsufficientPermissions, err.Code)
	}
	if err.StatusCode != constants.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", constants.StatusForbidden, err.StatusCode)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("Resource not found")

	if err.Type != ErrorTypeNotFound {
		t.Errorf("Expected type %v, got %v", ErrorTypeNotFound, err.Type)
	}
	if err.StatusCode != constants.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", constants.StatusNotFound, err.StatusCode)
	}
}

func TestNewInternalError(t *testing.T) {
	cause := errors.New("database connection failed")
	err := NewInternalError("Internal server error", cause)

	if err.Type != ErrorTypeInternal {
		t.Errorf("Expected type %v, got %v", ErrorTypeInternal, err.Type)
	}
	if err.Cause != cause {
		t.Errorf("Expected cause %v, got %v", cause, err.Cause)
	}
	if err.StatusCode != constants.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", constants.StatusInternalServerError, err.StatusCode)
	}
}

func TestNewDatabaseError(t *testing.T) {
	cause := errors.New("connection timeout")
	err := NewDatabaseError("Database operation failed", cause)

	if err.Type != ErrorTypeDatabase {
		t.Errorf("Expected type %v, got %v", ErrorTypeDatabase, err.Type)
	}
	if err.Code != constants.ErrorCodeDatabaseError {
		t.Errorf("Expected code '%s', got '%s'", constants.ErrorCodeDatabaseError, err.Code)
	}
}

func TestNewExternalServiceError(t *testing.T) {
	tests := []struct {
		name           string
		service        string
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "Partner service error",
			service:        "partner",
			expectedCode:   constants.ErrorCodePartnerServiceError,
			expectedStatus: constants.StatusBadGateway,
		},
		{
			name:           "Specification service error",
			service:        "specification",
			expectedCode:   constants.ErrorCodeSpecificationServiceError,
			expectedStatus: constants.StatusBadGateway,
		},
		{
			name:           "Unknown service error",
			service:        "unknown",
			expectedCode:   constants.ErrorCodeServiceUnavailable,
			expectedStatus: constants.StatusBadGateway,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cause := errors.New("service unavailable")
			err := NewExternalServiceError(tt.service, "Service error", cause)

			if err.Type != ErrorTypeExternalService {
				t.Errorf("Expected type %v, got %v", ErrorTypeExternalService, err.Type)
			}
			if err.Code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, err.Code)
			}
			if err.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, err.StatusCode)
			}
		})
	}
}

func TestNewTimeoutError(t *testing.T) {
	err := NewTimeoutError("Request timeout")

	if err.Type != ErrorTypeTimeout {
		t.Errorf("Expected type %v, got %v", ErrorTypeTimeout, err.Type)
	}
	if err.Code != constants.ErrorCodeTimeoutError {
		t.Errorf("Expected code '%s', got '%s'", constants.ErrorCodeTimeoutError, err.Code)
	}
	if err.StatusCode != constants.StatusGatewayTimeout {
		t.Errorf("Expected status code %d, got %d", constants.StatusGatewayTimeout, err.StatusCode)
	}
}

func TestErrorHandler_WithAppError(t *testing.T) {
	logger := createTestLogger()
	middleware := ErrorHandler(logger)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set an error in context
		r = SetValidationError(r, "Invalid input", map[string]string{"field": "required"})
		// Handler completes without writing response (error middleware will handle it)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != constants.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", constants.StatusBadRequest, w.Code)
	}

	response := parseErrorResponse(t, w.Body.String())
	validateErrorResponse(t, response, constants.ErrorCodeValidationError, "Invalid input", constants.StatusBadRequest)

	// Check if details are included
	errorObj := response["error"].(map[string]interface{})
	if details, ok := errorObj["details"]; !ok {
		t.Error("Expected error details to be present")
	} else if detailsMap, ok := details.(map[string]interface{}); !ok {
		t.Error("Expected error details to be a map")
	} else if detailsMap["field"] != "required" {
		t.Errorf("Expected field detail 'required', got '%v'", detailsMap["field"])
	}
}

func TestErrorHandler_WithPanic(t *testing.T) {
	logger := createTestLogger()
	middleware := ErrorHandler(logger)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != constants.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", constants.StatusInternalServerError, w.Code)
	}

	response := parseErrorResponse(t, w.Body.String())
	if !strings.Contains(response["message"].(string), "Internal server error") {
		t.Error("Expected panic message to contain 'Internal server error'")
	}
}

func TestErrorHandler_WithHTTPError(t *testing.T) {
	logger := createTestLogger()
	middleware := ErrorHandler(logger)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write an error status code directly
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != constants.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", constants.StatusNotFound, w.Code)
	}

	response := parseErrorResponse(t, w.Body.String())
	validateErrorResponse(t, response, constants.ErrorCodeInvalidRequest, "Resource not found", constants.StatusNotFound)
}

func TestErrorHandler_NoError(t *testing.T) {
	logger := createTestLogger()
	middleware := ErrorHandler(logger)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message": "OK"}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	if !strings.Contains(w.Body.String(), "success") {
		t.Error("Expected success response to pass through unchanged")
	}
}

func TestContextErrorHelpers(t *testing.T) {
	tests := []struct {
		name           string
		setError       func(*http.Request) *http.Request
		expectedCode   string
		expectedStatus int
	}{
		{
			name: "SetValidationError",
			setError: func(r *http.Request) *http.Request {
				return SetValidationError(r, "Validation failed", nil)
			},
			expectedCode:   constants.ErrorCodeValidationError,
			expectedStatus: constants.StatusBadRequest,
		},
		{
			name: "SetAuthenticationError",
			setError: func(r *http.Request) *http.Request {
				return SetAuthenticationError(r, constants.ErrorCodeInvalidToken, "Invalid token")
			},
			expectedCode:   constants.ErrorCodeInvalidToken,
			expectedStatus: constants.StatusUnauthorized,
		},
		{
			name: "SetAuthorizationError",
			setError: func(r *http.Request) *http.Request {
				return SetAuthorizationError(r, "Insufficient permissions")
			},
			expectedCode:   constants.ErrorCodeInsufficientPermissions,
			expectedStatus: constants.StatusForbidden,
		},
		{
			name: "SetNotFoundError",
			setError: func(r *http.Request) *http.Request {
				return SetNotFoundError(r, "Resource not found")
			},
			expectedCode:   constants.ErrorCodeInvalidRequest,
			expectedStatus: constants.StatusNotFound,
		},
		{
			name: "SetInternalError",
			setError: func(r *http.Request) *http.Request {
				return SetInternalError(r, "Internal error", errors.New("cause"))
			},
			expectedCode:   constants.ErrorCodeInternalServerError,
			expectedStatus: constants.StatusInternalServerError,
		},
		{
			name: "SetDatabaseError",
			setError: func(r *http.Request) *http.Request {
				return SetDatabaseError(r, "Database error", errors.New("connection failed"))
			},
			expectedCode:   constants.ErrorCodeDatabaseError,
			expectedStatus: constants.StatusInternalServerError,
		},
		{
			name: "SetExternalServiceError",
			setError: func(r *http.Request) *http.Request {
				return SetExternalServiceError(r, "partner", "Service error", errors.New("unavailable"))
			},
			expectedCode:   constants.ErrorCodePartnerServiceError,
			expectedStatus: constants.StatusBadGateway,
		},
		{
			name: "SetTimeoutError",
			setError: func(r *http.Request) *http.Request {
				return SetTimeoutError(r, "Request timeout")
			},
			expectedCode:   constants.ErrorCodeTimeoutError,
			expectedStatus: constants.StatusGatewayTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = tt.setError(req)

			appErr := GetErrorFromContext(req.Context())
			if appErr == nil {
				t.Fatal("Expected error to be set in context")
			}

			if appErr.Code != tt.expectedCode {
				t.Errorf("Expected error code '%s', got '%s'", tt.expectedCode, appErr.Code)
			}
			if appErr.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, appErr.StatusCode)
			}
		})
	}
}

func TestSetErrorInContext_GetErrorFromContext(t *testing.T) {
	appErr := NewValidationError("Test error", nil)
	ctx := context.Background()

	// Test setting error in context
	newCtx := SetErrorInContext(ctx, appErr)
	if newCtx == ctx {
		t.Error("Expected new context to be different from original")
	}

	// Test getting error from context
	retrievedErr := GetErrorFromContext(newCtx)
	if retrievedErr != appErr {
		t.Error("Expected retrieved error to match original error")
	}

	// Test getting error from context without error
	noErr := GetErrorFromContext(ctx)
	if noErr != nil {
		t.Error("Expected no error from original context")
	}
}

func TestErrorResponseWriter_MultipleWriteHeader(t *testing.T) {
	logger := createTestLogger()
	middleware := ErrorHandler(logger)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to write header multiple times
		w.WriteHeader(http.StatusOK)
		w.WriteHeader(http.StatusBadRequest) // This should be ignored
		w.Write([]byte("test"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

// Integration test with all middleware components
func TestErrorHandler_Integration(t *testing.T) {
	logger := createTestLogger()

	// Create a middleware chain with error handler
	errorMiddleware := ErrorHandler(logger)

	// Test handler that uses various error types
	handler := errorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/validation":
			r = SetValidationError(r, "Invalid input", map[string]string{"field": "required"})
		case "/auth":
			r = SetAuthenticationError(r, constants.ErrorCodeInvalidToken, "Token expired")
		case "/panic":
			panic("Something went wrong")
		case "/success":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "data": "test"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Validation error",
			path:           "/validation",
			expectedStatus: constants.StatusBadRequest,
			expectedCode:   constants.ErrorCodeValidationError,
		},
		{
			name:           "Authentication error",
			path:           "/auth",
			expectedStatus: constants.StatusUnauthorized,
			expectedCode:   constants.ErrorCodeInvalidToken,
		},
		{
			name:           "Panic handling",
			path:           "/panic",
			expectedStatus: constants.StatusInternalServerError,
			expectedCode:   constants.ErrorCodeInternalServerError,
		},
		{
			name:           "Success response",
			path:           "/success",
			expectedStatus: http.StatusOK,
			expectedCode:   "", // No error code for success
		},
		{
			name:           "HTTP not found",
			path:           "/notfound",
			expectedStatus: constants.StatusNotFound,
			expectedCode:   constants.ErrorCodeInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedCode == "" {
				// Success case - check for success response
				if !strings.Contains(w.Body.String(), "success") {
					t.Error("Expected success response")
				}
			} else {
				// Error case - validate error structure
				response := parseErrorResponse(t, w.Body.String())
				if errorObj, ok := response["error"].(map[string]interface{}); ok {
					if code, ok := errorObj["code"].(string); !ok || code != tt.expectedCode {
						t.Errorf("Expected error code '%s', got '%v'", tt.expectedCode, errorObj["code"])
					}
				} else {
					t.Error("Expected error object in response")
				}
			}
		})
	}
}
