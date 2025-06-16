package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"prayog-serviceability-service/internal/shared/constants/v1"
)

func TestAuth(t *testing.T) {
	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Setup API keys
	apiKeys := []string{"valid-key-1", "valid-key-2"}
	authMiddleware := Auth(apiKeys)
	wrappedHandler := authMiddleware(handler)

	// Test cases
	testCases := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Valid Bearer token",
			path:           "/test",
			authHeader:     "Bearer valid-key-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid ApiKey token",
			path:           "/test",
			authHeader:     "ApiKey valid-key-2",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid API key",
			path:           "/test",
			authHeader:     "Bearer invalid-key",
			expectedStatus: constants.StatusUnauthorized,
		},
		{
			name:           "Missing Authorization header",
			path:           "/test",
			authHeader:     "",
			expectedStatus: constants.StatusUnauthorized,
		},
		{
			name:           "Invalid header format",
			path:           "/test",
			authHeader:     "invalid-format",
			expectedStatus: constants.StatusUnauthorized,
		},
		{
			name:           "Unsupported scheme",
			path:           "/test",
			authHeader:     "Basic some-credentials",
			expectedStatus: constants.StatusUnauthorized,
		},
		{
			name:           "Skip auth for health endpoint",
			path:           "/health",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Skip auth for ready endpoint",
			path:           "/ready",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Skip auth for version endpoint",
			path:           "/version",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.path, nil)
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
			name:        "Valid bearer token (lowercase)",
			authHeader:  "bearer test-key",
			expectedKey: "test-key",
			expectError: false,
		},
		{
			name:        "Valid apikey token (lowercase)",
			authHeader:  "apikey test-key",
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
			name:        "Invalid header format - no space",
			authHeader:  "invalid-format",
			expectedKey: "",
			expectError: true,
		},
		{
			name:        "Invalid header format - only scheme",
			authHeader:  "Bearer",
			expectedKey: "",
			expectError: true,
		},
		{
			name:        "Unsupported scheme",
			authHeader:  "Basic credentials",
			expectedKey: "",
			expectError: true,
		},
		{
			name:        "Empty token",
			authHeader:  "Bearer ",
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

func TestIsValidAPIKey(t *testing.T) {
	validKeys := []string{"key1", "key2", "key3"}

	testCases := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{
			name:     "Valid key 1",
			apiKey:   "key1",
			expected: true,
		},
		{
			name:     "Valid key 2",
			apiKey:   "key2",
			expected: true,
		},
		{
			name:     "Valid key 3",
			apiKey:   "key3",
			expected: true,
		},
		{
			name:     "Invalid key",
			apiKey:   "invalid-key",
			expected: false,
		},
		{
			name:     "Empty key",
			apiKey:   "",
			expected: false,
		},
		{
			name:     "Case sensitive - wrong case",
			apiKey:   "KEY1",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidAPIKey(tc.apiKey, validKeys)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
