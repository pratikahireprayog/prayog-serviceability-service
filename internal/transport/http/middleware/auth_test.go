package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuth(t *testing.T) {
	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Setup API keys
	config := AuthConfig{
		APIKeys: []string{"valid-key-1", "valid-key-2"},
	}
	authMiddleware := Auth(config)
	wrappedHandler := authMiddleware(handler)

	// Test cases
	testCases := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Valid Bearer token",
			authHeader:     "Bearer valid-key-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid ApiKey token",
			authHeader:     "ApiKey valid-key-2",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid API key",
			authHeader:     "Bearer invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing Authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid header format",
			authHeader:     "invalid-format",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Unsupported scheme",
			authHeader:     "Basic some-credentials",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rr.Code)
			}
		})
	}
}

func TestExtractAPIKey(t *testing.T) {
	// Test cases
	testCases := []struct {
		name        string
		authHeader  string
		expectedKey string
		expectError bool
	}{
		{
			name:        "Valid Bearer token",
			authHeader:  "Bearer test-key",
			expectedKey: "test-key",
			expectError: false,
		},
		{
			name:        "Valid ApiKey token",
			authHeader:  "ApiKey test-key",
			expectedKey: "test-key",
			expectError: false,
		},
		{
			name:        "Missing Authorization header",
			authHeader:  "",
			expectedKey: "",
			expectError: true,
		},
		{
			name:        "Invalid header format",
			authHeader:  "invalid-format",
			expectedKey: "",
			expectError: true,
		},
		{
			name:        "Unsupported scheme",
			authHeader:  "Basic credentials",
			expectedKey: "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			key, err := extractAPIKey(req)

			// Check if error matches expectation
			if tc.expectError && err == nil {
				t.Error("Expected error but got nil")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check extracted key
			if key != tc.expectedKey {
				t.Errorf("Expected key %q, got %q", tc.expectedKey, key)
			}
		})
	}
}
