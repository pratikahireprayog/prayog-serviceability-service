package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		version        string
		expectedStatus int
	}{
		{
			name:           "Valid Health Check",
			version:        "1.0.0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty Version",
			version:        "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("GET", "/health", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler := HealthHandler(tc.version)
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}

			// Check Content-Type
			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Handler returned wrong content type: got %v want %v", contentType, "application/json")
			}

			// Parse response
			var response HealthResponse
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Errorf("Error decoding response: %v", err)
				return
			}

			// Check response fields
			if response.Status != "ok" {
				t.Errorf("Expected status 'ok', got '%s'", response.Status)
			}

			if response.Version != tc.version {
				t.Errorf("Expected version '%s', got '%s'", tc.version, response.Version)
			}

			// Check if timestamp is set (not zero)
			if response.Timestamp.IsZero() {
				t.Error("Expected non-zero timestamp")
			}
		})
	}
}
