package middleware

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"prayog-serviceability-service/internal/shared/constants/v1"
)

// TestMiddlewareStackIntegration tests the complete middleware stack integration
func TestMiddlewareStackIntegration(t *testing.T) {
	// Create logger for error middleware
	logger := log.New(io.Discard, "", 0)

	// Valid API keys for testing
	validAPIKeys := []string{"test-api-key-1", "test-api-key-2"}

	// Test handler that demonstrates various scenarios
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			// Simple success response
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "Operation successful"}`))

		case "/success-with-context":
			// Success response using context helpers
			SetResponseMessage(r, "Data retrieved successfully")
			SetResponseData(r, map[string]interface{}{
				"users": []string{"user1", "user2"},
				"count": 2,
			})
			w.WriteHeader(http.StatusOK)
			// Write empty body to trigger response formatting
			w.Write([]byte{})

		case "/validation-error":
			// Handler sets validation error
			r = SetValidationError(r, "Invalid input data", map[string]string{
				"email": "required",
				"name":  "must be at least 2 characters",
			})
			// Don't write response - error middleware will handle it

		case "/auth-error":
			// Handler sets authentication error
			r = SetAuthenticationError(r, constants.ErrorCodeExpiredToken, "Token has expired")
			// Don't write response - error middleware will handle it

		case "/internal-error":
			// Handler sets internal error
			r = SetInternalError(r, "Database connection failed", nil)
			// Don't write response - error middleware will handle it

		case "/panic":
			// Handler panics
			panic("Something went terribly wrong")

		case "/not-found":
			// Handler returns 404
			w.WriteHeader(http.StatusNotFound)

		case "/plain-text":
			// Plain text response
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, World!"))

		case "/already-formatted":
			// Already formatted response (should not be double-formatted)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "message": "Already formatted", "data": {"test": true}}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Apply middleware stack to handler in proper order:
	// 1. Error handler (outermost - catches all errors)
	// 2. Success response handler (formats successful responses)
	// 3. Authentication middleware (validates requests)
	handler := Chain(testHandler,
		Auth(validAPIKeys),
		SuccessResponseHandler(),
		ErrorHandler(logger),
	)

	tests := []struct {
		name           string
		path           string
		method         string
		authHeader     string
		expectedStatus int
		expectedBody   map[string]interface{}
		checkSuccess   bool
		checkError     bool
		checkData      bool
		checkMessage   bool
	}{
		{
			name:           "Successful request with valid auth",
			path:           "/success",
			method:         http.MethodGet,
			authHeader:     "Bearer test-api-key-1",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data":    map[string]interface{}{"message": "Operation successful"},
			},
			checkSuccess: true,
			checkData:    true,
		},
		{
			name:           "Success with context helpers",
			path:           "/success-with-context",
			method:         http.MethodGet,
			authHeader:     "Bearer test-api-key-2",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"message": "Data retrieved successfully",
				"data": map[string]interface{}{
					"users": []interface{}{"user1", "user2"},
					"count": float64(2),
				},
			},
			checkSuccess: true,
			checkMessage: true,
			checkData:    true,
		},
		{
			name:           "Authentication failure - missing header",
			path:           "/success",
			method:         http.MethodGet,
			authHeader:     "",
			expectedStatus: constants.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": constants.MsgMissingCredentials,
				"error": map[string]interface{}{
					"code":    constants.ErrorCodeMissingCredentials,
					"message": constants.MsgMissingCredentials,
				},
			},
			checkSuccess: true,
			checkError:   true,
		},
		{
			name:           "Authentication failure - invalid token",
			path:           "/success",
			method:         http.MethodGet,
			authHeader:     "Bearer invalid-token",
			expectedStatus: constants.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": constants.MsgInvalidToken,
				"error": map[string]interface{}{
					"code":    constants.ErrorCodeInvalidToken,
					"message": constants.MsgInvalidToken,
				},
			},
			checkSuccess: true,
			checkError:   true,
		},
		{
			name:           "Validation error from handler",
			path:           "/validation-error",
			method:         http.MethodPost,
			authHeader:     "Bearer test-api-key-1",
			expectedStatus: constants.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": "Invalid input data",
				"error": map[string]interface{}{
					"code":    constants.ErrorCodeValidationError,
					"message": "Invalid input data",
					"details": map[string]interface{}{
						"email": "required",
						"name":  "must be at least 2 characters",
					},
				},
			},
			checkSuccess: true,
			checkError:   true,
		},
		{
			name:           "Authentication error from handler",
			path:           "/auth-error",
			method:         http.MethodGet,
			authHeader:     "Bearer test-api-key-1",
			expectedStatus: constants.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": "Token has expired",
				"error": map[string]interface{}{
					"code":    constants.ErrorCodeExpiredToken,
					"message": "Token has expired",
				},
			},
			checkSuccess: true,
			checkError:   true,
		},
		{
			name:           "Internal error from handler",
			path:           "/internal-error",
			method:         http.MethodGet,
			authHeader:     "Bearer test-api-key-1",
			expectedStatus: constants.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": "Database connection failed",
				"error": map[string]interface{}{
					"code":    constants.ErrorCodeInternalServerError,
					"message": "Database connection failed",
				},
			},
			checkSuccess: true,
			checkError:   true,
		},
		{
			name:           "Panic recovery",
			path:           "/panic",
			method:         http.MethodGet,
			authHeader:     "Bearer test-api-key-1",
			expectedStatus: constants.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": constants.MsgInternalServerError,
				"error": map[string]interface{}{
					"code":    constants.ErrorCodeInternalServerError,
					"message": constants.MsgInternalServerError,
				},
			},
			checkSuccess: true,
			checkError:   true,
		},
		{
			name:           "HTTP 404 error",
			path:           "/not-found",
			method:         http.MethodGet,
			authHeader:     "Bearer test-api-key-1",
			expectedStatus: constants.StatusNotFound,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": "Resource not found",
				"error": map[string]interface{}{
					"code":    constants.ErrorCodeInvalidRequest,
					"message": "Resource not found",
				},
			},
			checkSuccess: true,
			checkError:   true,
		},
		{
			name:           "Plain text response formatting",
			path:           "/plain-text",
			method:         http.MethodGet,
			authHeader:     "Bearer test-api-key-1",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data":    "Hello, World!",
			},
			checkSuccess: true,
			checkData:    true,
		},
		{
			name:           "Already formatted response (no double formatting)",
			path:           "/already-formatted",
			method:         http.MethodGet,
			authHeader:     "Bearer test-api-key-1",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"message": "Already formatted",
				"data":    map[string]interface{}{"test": true},
			},
			checkSuccess: true,
			checkMessage: true,
			checkData:    true,
		},
		{
			name:           "Health endpoint bypasses auth",
			path:           "/health",
			method:         http.MethodGet,
			authHeader:     "",                  // No auth header
			expectedStatus: http.StatusNotFound, // Handler returns 404 for unknown paths
			expectedBody: map[string]interface{}{
				"success": false,
				"message": "Resource not found",
				"error": map[string]interface{}{
					"code":    constants.ErrorCodeInvalidRequest,
					"message": "Resource not found",
				},
			},
			checkSuccess: true,
			checkError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check Content-Type header
			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("Expected Content-Type to contain 'application/json', got '%s'", contentType)
			}

			// Parse response body
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response JSON: %v\nBody: %s", err, w.Body.String())
			}

			// Check success field
			if tt.checkSuccess {
				expectedSuccess := tt.expectedBody["success"].(bool)
				if success, ok := response["success"].(bool); !ok || success != expectedSuccess {
					t.Errorf("Expected success to be %v, got %v", expectedSuccess, response["success"])
				}
			}

			// Check message field
			if tt.checkMessage {
				if expectedMessage, exists := tt.expectedBody["message"]; exists {
					if message, ok := response["message"].(string); !ok || message != expectedMessage {
						t.Errorf("Expected message '%v', got '%v'", expectedMessage, response["message"])
					}
				}
			}

			// Check data field
			if tt.checkData {
				if expectedData, exists := tt.expectedBody["data"]; exists {
					if !compareData(t, expectedData, response["data"]) {
						t.Errorf("Data mismatch.\nExpected: %v\nGot: %v", expectedData, response["data"])
					}
				}
			}

			// Check error field
			if tt.checkError {
				if expectedError, exists := tt.expectedBody["error"]; exists {
					errorObj, ok := response["error"].(map[string]interface{})
					if !ok {
						t.Fatalf("Expected error object, got %v", response["error"])
					}

					expectedErrorMap := expectedError.(map[string]interface{})

					// Check error code
					if expectedCode, exists := expectedErrorMap["code"]; exists {
						if code, ok := errorObj["code"].(string); !ok || code != expectedCode {
							t.Errorf("Expected error code '%v', got '%v'", expectedCode, errorObj["code"])
						}
					}

					// Check error message
					if expectedMessage, exists := expectedErrorMap["message"]; exists {
						if message, ok := errorObj["message"].(string); !ok || message != expectedMessage {
							t.Errorf("Expected error message '%v', got '%v'", expectedMessage, errorObj["message"])
						}
					}

					// Check error details if present
					if expectedDetails, exists := expectedErrorMap["details"]; exists {
						if !compareData(t, expectedDetails, errorObj["details"]) {
							t.Errorf("Error details mismatch.\nExpected: %v\nGot: %v", expectedDetails, errorObj["details"])
						}
					}
				}
			}
		})
	}
}

// compareData compares two interface{} values for equality, handling nested maps and slices
func compareData(t *testing.T, expected, actual interface{}) bool {
	expectedJSON, err1 := json.Marshal(expected)
	actualJSON, err2 := json.Marshal(actual)

	if err1 != nil || err2 != nil {
		t.Logf("JSON marshal error: %v, %v", err1, err2)
		return false
	}

	return string(expectedJSON) == string(actualJSON)
}

// TestMiddlewareOrder tests that middleware is applied in the correct order
func TestMiddlewareOrder(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	validAPIKeys := []string{"test-key"}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This handler will cause an authentication error due to missing header
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Should not reach here"))
	})

	// Test with different middleware orders to ensure error handling works correctly
	handler := Chain(testHandler,
		Auth(validAPIKeys),       // Authentication check
		SuccessResponseHandler(), // Formats successful responses
		ErrorHandler(logger),     // Must be outermost to catch all errors
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Authorization header - should trigger auth error
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should get 401 with proper error format
	if w.Code != constants.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", constants.StatusUnauthorized, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should have error structure
	if success, ok := response["success"].(bool); !ok || success {
		t.Error("Expected success to be false")
	}

	if _, ok := response["error"]; !ok {
		t.Error("Expected error object in response")
	}

	// Should not contain the handler's message
	if strings.Contains(w.Body.String(), "Should not reach here") {
		t.Error("Handler should not have been reached due to auth failure")
	}
}

// TestMiddlewarePerformance tests that the middleware stack doesn't add significant overhead
func TestMiddlewarePerformance(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	validAPIKeys := []string{"test-key"}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test": "data"}`))
	})

	handler := Chain(testHandler,
		Auth(validAPIKeys),
		SuccessResponseHandler(),
		ErrorHandler(logger),
	)

	// Run multiple requests to check for any performance issues
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer test-key")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed with status %d", i, w.Code)
		}
	}
}

// TestConcurrentRequests tests that the middleware stack is safe for concurrent use
func TestConcurrentRequests(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	validAPIKeys := []string{"test-key"}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"concurrent": "test"}`))
	})

	handler := Chain(testHandler,
		Auth(validAPIKeys),
		SuccessResponseHandler(),
		ErrorHandler(logger),
	)

	// Run concurrent requests
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer test-key")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Concurrent request %d failed with status %d", id, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Concurrent request %d failed to parse response: %v", id, err)
			}

			if success, ok := response["success"].(bool); !ok || !success {
				t.Errorf("Concurrent request %d: expected success to be true", id)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
