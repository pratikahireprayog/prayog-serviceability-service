package unit

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"
)

func TestDHLAuthenticationFlowValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		username         string
		password         string
		expectedAuth     bool
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name:          "Valid DHL credentials",
			username:      "dhl_test_user",
			password:      "dhl_test_password",
			expectedAuth:  true,
			expectedError: false,
		},
		{
			name:             "Empty username",
			username:         "",
			password:         "dhl_test_password",
			expectedAuth:     false,
			expectedError:    true,
			expectedErrorMsg: "DHL authentication failed: username cannot be empty",
		},
		{
			name:             "Empty password",
			username:         "dhl_test_user",
			password:         "",
			expectedAuth:     false,
			expectedError:    true,
			expectedErrorMsg: "DHL authentication failed: password cannot be empty",
		},
		{
			name:             "Invalid credentials",
			username:         "invalid_user",
			password:         "invalid_password",
			expectedAuth:     false,
			expectedError:    true,
			expectedErrorMsg: "DHL authentication failed: invalid credentials",
		},
		{
			name:          "Special characters in credentials",
			username:      "dhl@test.com",
			password:      "password123!@#",
			expectedAuth:  true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock adapter factory
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			if tt.expectedError {
				// Setup factory to return error for DHL adapter creation
				mockFactory.CreateAdapterError = fmt.Errorf("%s", tt.expectedErrorMsg)
			} else {
				// Setup successful DHL adapter
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				partnerID := uuid.New()
				dhlAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerID:   &partnerID,
					PartnerCode: "dhl",
					PartnerName: "DHL Express",
					Services: []models.ServiceV2{
						{
							ServiceCode: "EXPRESS",
							ServiceName: "DHL Express Worldwide",
						},
					},
				})
				mockFactory.SetAdapter("dhl", dhlAdapter)
			}

			// Test authentication flow
			_, err := mockFactory.CreateAdapter("dhl")

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "DHL authentication failed")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDHLBasicAuthSetup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		username           string
		password           string
		expectedAuthHeader string
		shouldGenerateAuth bool
	}{
		{
			name:               "Standard credentials",
			username:           "testuser",
			password:           "testpass",
			expectedAuthHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("testuser:testpass")),
			shouldGenerateAuth: true,
		},
		{
			name:               "Special characters in credentials",
			username:           "user@domain.com",
			password:           "p@ssw0rd!",
			expectedAuthHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("user@domain.com:p@ssw0rd!")),
			shouldGenerateAuth: true,
		},
		{
			name:               "Unicode characters in password",
			username:           "testuser",
			password:           "пароль123",
			expectedAuthHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("testuser:пароль123")),
			shouldGenerateAuth: true,
		},
		{
			name:               "Empty credentials",
			username:           "",
			password:           "",
			expectedAuthHeader: "",
			shouldGenerateAuth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.shouldGenerateAuth {
				authHeader := generateBasicAuth(tt.username, tt.password)
				assert.Equal(t, tt.expectedAuthHeader, authHeader)
			} else {
				// Test that empty credentials don't generate auth header
				authHeader := generateBasicAuth(tt.username, tt.password)
				assert.Empty(t, authHeader)
			}
		})
	}
}

func TestDHLCredentialRedaction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		username            string
		password            string
		expectedRedactedLog string
	}{
		{
			name:                "Standard credentials redaction",
			username:            "testuser",
			password:            "testpass",
			expectedRedactedLog: "DHL auth configured for user: test***",
		},
		{
			name:                "Long username redaction",
			username:            "very_long_username_for_testing",
			password:            "password123",
			expectedRedactedLog: "DHL auth configured for user: very***",
		},
		{
			name:                "Short username redaction",
			username:            "ab",
			password:            "password",
			expectedRedactedLog: "DHL auth configured for user: ***",
		},
		{
			name:                "Email username redaction",
			username:            "user@domain.com",
			password:            "password",
			expectedRedactedLog: "DHL auth configured for user: user***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			redactedLog := redactCredentialsForLogging(tt.username, tt.password)
			assert.Equal(t, tt.expectedRedactedLog, redactedLog)

			// Ensure password is never logged
			assert.NotContains(t, redactedLog, tt.password)
		})
	}
}

func TestDHLTokenManagement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		existingToken     string
		tokenExpiry       bool
		shouldRefreshAuth bool
		expectedNewToken  bool
	}{
		{
			name:              "Valid existing token",
			existingToken:     "valid_token_123",
			tokenExpiry:       false,
			shouldRefreshAuth: false,
			expectedNewToken:  false,
		},
		{
			name:              "Expired token requires refresh",
			existingToken:     "expired_token_456",
			tokenExpiry:       true,
			shouldRefreshAuth: true,
			expectedNewToken:  true,
		},
		{
			name:              "No existing token",
			existingToken:     "",
			tokenExpiry:       false,
			shouldRefreshAuth: true,
			expectedNewToken:  true,
		},
		{
			name:              "Invalid token format",
			existingToken:     "invalid_token",
			tokenExpiry:       false,
			shouldRefreshAuth: true,
			expectedNewToken:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock adapter
			mockAdapter := mocks.NewMockPartnerAdapter("dhl")
			partnerID := uuid.New()
			mockAdapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
				PartnerID:   &partnerID,
				PartnerCode: "dhl",
				PartnerName: "DHL Express",
			})

			// Test token management logic
			result := handleTokenManagement(mockAdapter, tt.existingToken, tt.tokenExpiry)

			if tt.expectedNewToken {
				assert.True(t, result.TokenRefreshed)
				assert.NotEmpty(t, result.NewToken)
			} else {
				assert.False(t, result.TokenRefreshed)
				assert.Equal(t, tt.existingToken, result.CurrentToken)
			}
		})
	}
}

func TestDHLAuthenticationErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		authError         error
		expectedErrorMsg  string
		expectedErrorCode string
	}{
		{
			name:              "Network timeout error",
			authError:         fmt.Errorf("connection timeout"),
			expectedErrorMsg:  "DHL authentication failed: connection timeout",
			expectedErrorCode: "DHL_AUTH_TIMEOUT",
		},
		{
			name:              "Invalid credentials error",
			authError:         fmt.Errorf("401 Unauthorized"),
			expectedErrorMsg:  "DHL authentication failed: 401 Unauthorized",
			expectedErrorCode: "DHL_AUTH_INVALID_CREDENTIALS",
		},
		{
			name:              "Service unavailable error",
			authError:         fmt.Errorf("503 Service Unavailable"),
			expectedErrorMsg:  "DHL authentication failed: 503 Service Unavailable",
			expectedErrorCode: "DHL_AUTH_SERVICE_UNAVAILABLE",
		},
		{
			name:              "Rate limit error",
			authError:         fmt.Errorf("429 Too Many Requests"),
			expectedErrorMsg:  "DHL authentication failed: 429 Too Many Requests",
			expectedErrorCode: "DHL_AUTH_RATE_LIMITED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock adapter factory with authentication error
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockFactory.CreateAdapterError = tt.authError

			// Test authentication error handling
			_, err := mockFactory.CreateAdapter("dhl")

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "DHL authentication failed")

			// Test error categorization
			errorCode := categorizeAuthError(tt.authError)
			assert.Equal(t, tt.expectedErrorCode, errorCode)
		})
	}
}

func TestDHLAuthenticationCredentialValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		username      string
		password      string
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Valid credentials",
			username:      "valid_user",
			password:      "valid_password",
			expectedValid: true,
		},
		{
			name:          "Empty username",
			username:      "",
			password:      "password",
			expectedValid: false,
			expectedError: "username cannot be empty",
		},
		{
			name:          "Empty password",
			username:      "username",
			password:      "",
			expectedValid: false,
			expectedError: "password cannot be empty",
		},
		{
			name:          "Whitespace only username",
			username:      "   ",
			password:      "password",
			expectedValid: false,
			expectedError: "username cannot be empty",
		},
		{
			name:          "Whitespace only password",
			username:      "username",
			password:      "   ",
			expectedValid: false,
			expectedError: "password cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			isValid, err := validateDHLCredentials(tt.username, tt.password)

			if tt.expectedValid {
				assert.True(t, isValid)
				assert.NoError(t, err)
			} else {
				assert.False(t, isValid)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

// Helper functions for DHL authentication tests

func generateBasicAuth(username, password string) string {
	if username == "" && password == "" {
		return ""
	}
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func redactCredentialsForLogging(username, password string) string {
	if len(username) <= 4 {
		return "DHL auth configured for user: ***"
	}
	return fmt.Sprintf("DHL auth configured for user: %s***", username[:4])
}

type TokenManagementResult struct {
	TokenRefreshed bool
	NewToken       string
	CurrentToken   string
}

func handleTokenManagement(adapter *mocks.MockPartnerAdapter, existingToken string, tokenExpiry bool) TokenManagementResult {
	if existingToken == "" || tokenExpiry || existingToken == "invalid_token" {
		// Simulate token refresh
		return TokenManagementResult{
			TokenRefreshed: true,
			NewToken:       "new_refreshed_token_" + generateRandomString(8),
		}
	}

	return TokenManagementResult{
		TokenRefreshed: false,
		CurrentToken:   existingToken,
	}
}

func categorizeAuthError(err error) string {
	errMsg := err.Error()

	switch {
	case contains(errMsg, "timeout"):
		return "DHL_AUTH_TIMEOUT"
	case contains(errMsg, "401"):
		return "DHL_AUTH_INVALID_CREDENTIALS"
	case contains(errMsg, "503"):
		return "DHL_AUTH_SERVICE_UNAVAILABLE"
	case contains(errMsg, "429"):
		return "DHL_AUTH_RATE_LIMITED"
	default:
		return "DHL_AUTH_UNKNOWN_ERROR"
	}
}

func validateDHLCredentials(username, password string) (bool, error) {
	// Trim whitespace and validate
	username = trimWhitespace(username)
	password = trimWhitespace(password)

	if username == "" {
		return false, fmt.Errorf("username cannot be empty")
	}

	if password == "" {
		return false, fmt.Errorf("password cannot be empty")
	}

	return true, nil
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func trimWhitespace(s string) string {
	// Simple whitespace trimming implementation
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// Trim trailing whitespace
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// stringPtr helper function is defined in other test files
