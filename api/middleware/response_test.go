package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"prayog-serviceability-service/internal/shared/constants/v1"
)

func TestSuccessResponseHandler(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "JSON response gets formatted",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"id": 1, "name": "test"}`))
			}),
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response StandardResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if !response.Success {
					t.Error("Expected success to be true")
				}
				if response.Data == nil {
					t.Error("Expected data to be present")
				}
			},
		},
		{
			name: "String response gets formatted",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello World"))
			}),
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response StandardResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if !response.Success {
					t.Error("Expected success to be true")
				}
				if response.Data != "Hello World" {
					t.Errorf("Expected data to be 'Hello World', got %v", response.Data)
				}
			},
		},
		{
			name: "Already formatted response is not double-formatted",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				RespondWithJSON(w, http.StatusOK, map[string]string{"test": "data"})
			}),
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response StandardResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if !response.Success {
					t.Error("Expected success to be true")
				}
				// Should not be double-wrapped
				if dataMap, ok := response.Data.(map[string]interface{}); ok {
					if dataMap["test"] != "data" {
						t.Error("Data was incorrectly formatted")
					}
				} else {
					t.Error("Expected data to be a map")
				}
			},
		},
		{
			name: "Error status codes are not formatted",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Bad request"))
			}),
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				if body != "Bad request" {
					t.Errorf("Expected 'Bad request', got %s", body)
				}
			},
		},
		{
			name: "Empty response is handled",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				// Empty responses should not be formatted since nothing was written
				if body != "" {
					t.Errorf("Expected empty body, got %s", body)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := SuccessResponseHandler()
			wrappedHandler := middleware(tt.handler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			tt.checkResponse(t, rr.Body.String())
		})
	}
}

func TestResponseContextHelpers(t *testing.T) {
	t.Run("GetResponseData", func(t *testing.T) {
		middleware := SuccessResponseHandler()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			respData := GetResponseData(r)
			if respData == nil {
				t.Error("Expected response data to be available")
			}
			if respData.Status != "success" {
				t.Errorf("Expected status 'success', got %s", respData.Status)
			}
		})

		wrappedHandler := middleware(handler)
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)
	})

	t.Run("SetResponseData", func(t *testing.T) {
		middleware := SuccessResponseHandler()
		testData := map[string]string{"key": "value"}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			SetResponseData(r, testData)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test"))
		})

		wrappedHandler := middleware(handler)
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		var response StandardResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if dataMap, ok := response.Data.(map[string]interface{}); ok {
			if dataMap["key"] != "value" {
				t.Error("SetResponseData did not work correctly")
			}
		} else {
			t.Error("Expected data to be set from context")
		}
	})

	t.Run("SetResponseMessage", func(t *testing.T) {
		middleware := SuccessResponseHandler()
		testMessage := "Operation completed successfully"

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			SetResponseMessage(r, testMessage)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test"))
		})

		wrappedHandler := middleware(handler)
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		var response StandardResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Message != testMessage {
			t.Errorf("Expected message '%s', got '%s'", testMessage, response.Message)
		}
	})
}

func TestRespondWithJSON(t *testing.T) {
	testData := map[string]interface{}{
		"id":   1,
		"name": "test",
	}

	rr := httptest.NewRecorder()

	RespondWithJSON(rr, constants.StatusOK, testData)

	if rr.Code != constants.StatusOK {
		t.Errorf("Expected status %d, got %d", constants.StatusOK, rr.Code)
	}

	var response StandardResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Data == nil {
		t.Error("Expected data to be present")
	}
}

func TestRespondWithSuccess(t *testing.T) {
	testData := map[string]string{"result": "success"}
	testMessage := "Operation completed"

	rr := httptest.NewRecorder()

	RespondWithSuccess(rr, constants.StatusCreated, testMessage, testData)

	if rr.Code != constants.StatusCreated {
		t.Errorf("Expected status %d, got %d", constants.StatusCreated, rr.Code)
	}

	var response StandardResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Message != testMessage {
		t.Errorf("Expected message '%s', got '%s'", testMessage, response.Message)
	}

	if response.Data == nil {
		t.Error("Expected data to be present")
	}
}

func TestRespondWithError(t *testing.T) {
	testMessage := "Something went wrong"
	testErrorCode := constants.ErrorCodeValidationError

	rr := httptest.NewRecorder()

	RespondWithError(rr, constants.StatusBadRequest, testMessage, testErrorCode)

	if rr.Code != constants.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", constants.StatusBadRequest, rr.Code)
	}

	var response StandardResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected success to be false")
	}

	if response.Message != testMessage {
		t.Errorf("Expected message '%s', got '%s'", testMessage, response.Message)
	}

	if response.Error == nil {
		t.Error("Expected error to be present")
	} else {
		if response.Error.Code != testErrorCode {
			t.Errorf("Expected error code '%s', got '%s'", testErrorCode, response.Error.Code)
		}
		if response.Error.Message != testMessage {
			t.Errorf("Expected error message '%s', got '%s'", testMessage, response.Error.Message)
		}
	}
}

func TestRespondWithErrorDetails(t *testing.T) {
	testMessage := "Validation failed"
	testErrorCode := constants.ErrorCodeValidationError
	testDetails := map[string]string{
		"field": "email",
		"issue": "invalid format",
	}

	rr := httptest.NewRecorder()

	RespondWithErrorDetails(rr, constants.StatusBadRequest, testMessage, testErrorCode, testDetails)

	if rr.Code != constants.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", constants.StatusBadRequest, rr.Code)
	}

	var response StandardResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected success to be false")
	}

	if response.Error == nil {
		t.Error("Expected error to be present")
	} else {
		if response.Error.Details == nil {
			t.Error("Expected error details to be present")
		}
	}
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name           string
		helperFunc     func(http.ResponseWriter, string)
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "RespondWithValidationError",
			helperFunc:     RespondWithValidationError,
			expectedStatus: constants.StatusBadRequest,
			expectedCode:   constants.ErrorCodeValidationError,
		},
		{
			name:           "RespondWithNotFound",
			helperFunc:     RespondWithNotFound,
			expectedStatus: constants.StatusNotFound,
			expectedCode:   constants.ErrorCodeInvalidRequest,
		},
		{
			name:           "RespondWithInternalError",
			helperFunc:     RespondWithInternalError,
			expectedStatus: constants.StatusInternalServerError,
			expectedCode:   constants.ErrorCodeInternalServerError,
		},
		{
			name:           "RespondWithUnauthorized",
			helperFunc:     RespondWithUnauthorized,
			expectedStatus: constants.StatusUnauthorized,
			expectedCode:   constants.ErrorCodeAuthFailed,
		},
		{
			name:           "RespondWithForbidden",
			helperFunc:     RespondWithForbidden,
			expectedStatus: constants.StatusForbidden,
			expectedCode:   constants.ErrorCodeInsufficientPermissions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			testMessage := "Test error message"

			tt.helperFunc(rr, testMessage)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response StandardResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Success {
				t.Error("Expected success to be false")
			}

			if response.Error == nil {
				t.Error("Expected error to be present")
			} else if response.Error.Code != tt.expectedCode {
				t.Errorf("Expected error code '%s', got '%s'", tt.expectedCode, response.Error.Code)
			}
		})
	}
}

func TestIsAlreadyFormatted(t *testing.T) {
	tests := []struct {
		name     string
		body     []byte
		expected bool
	}{
		{
			name:     "Empty body",
			body:     []byte{},
			expected: false,
		},
		{
			name:     "Standard success response",
			body:     []byte(`{"success": true, "data": {"test": "value"}}`),
			expected: true,
		},
		{
			name:     "Standard error response",
			body:     []byte(`{"success": false, "error": {"code": "TEST", "message": "test"}}`),
			expected: true,
		},
		{
			name:     "Non-standard JSON",
			body:     []byte(`{"id": 1, "name": "test"}`),
			expected: false,
		},
		{
			name:     "Invalid JSON",
			body:     []byte(`{invalid json`),
			expected: false,
		},
		{
			name:     "Plain text",
			body:     []byte(`Hello World`),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlreadyFormatted(tt.body)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
