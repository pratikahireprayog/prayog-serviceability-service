package services

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"prayog-serviceability-service/internal/shared/constants/v1"
)

// MockHTTPClient implements HTTPClient for testing
type MockHTTPClient struct {
	response *HTTPResponse
	err      error
}

func (m *MockHTTPClient) Get(ctx context.Context, url string) (*HTTPResponse, error) {
	return m.response, m.err
}

func TestLoadPartnerValidationConfig(t *testing.T) {
	// Test with environment variable set
	originalBaseURL := os.Getenv(constants.EnvPartnerServiceBaseURL)
	defer func() {
		if originalBaseURL == "" {
			os.Unsetenv(constants.EnvPartnerServiceBaseURL)
		} else {
			os.Setenv(constants.EnvPartnerServiceBaseURL, originalBaseURL)
		}
	}()

	testBaseURL := "https://test.example.com"
	os.Setenv(constants.EnvPartnerServiceBaseURL, testBaseURL)

	config := LoadPartnerValidationConfig()

	if config.BaseURL != testBaseURL {
		t.Errorf("Expected BaseURL %s, got %s", testBaseURL, config.BaseURL)
	}

	if config.RequestTimeout != 30*time.Second {
		t.Errorf("Expected RequestTimeout 30s, got %v", config.RequestTimeout)
	}

	if config.RetryAttempts != 3 {
		t.Errorf("Expected RetryAttempts 3, got %d", config.RetryAttempts)
	}

	if config.RetryDelay != 1*time.Second {
		t.Errorf("Expected RetryDelay 1s, got %v", config.RetryDelay)
	}
}

func TestLoadPartnerValidationConfigWithoutEnv(t *testing.T) {
	// Test without environment variable (should use default)
	originalBaseURL := os.Getenv(constants.EnvPartnerServiceBaseURL)
	defer func() {
		if originalBaseURL == "" {
			os.Unsetenv(constants.EnvPartnerServiceBaseURL)
		} else {
			os.Setenv(constants.EnvPartnerServiceBaseURL, originalBaseURL)
		}
	}()

	os.Unsetenv(constants.EnvPartnerServiceBaseURL)

	config := LoadPartnerValidationConfig()

	expectedDefault := "http://localhost:9024"
	if config.BaseURL != expectedDefault {
		t.Errorf("Expected default BaseURL %s, got %s", expectedDefault, config.BaseURL)
	}
}

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	expectedBaseURL := "http://localhost:9024"
	if config.BaseURL != expectedBaseURL {
		t.Errorf("Expected BaseURL %s, got %s", expectedBaseURL, config.BaseURL)
	}

	if config.RequestTimeout != 30*time.Second {
		t.Errorf("Expected RequestTimeout 30s, got %v", config.RequestTimeout)
	}

	if config.RetryAttempts != 3 {
		t.Errorf("Expected RetryAttempts 3, got %d", config.RetryAttempts)
	}

	if config.RetryDelay != 1*time.Second {
		t.Errorf("Expected RetryDelay 1s, got %v", config.RetryDelay)
	}
}

func TestNewPartnerValidationService(t *testing.T) {
	config := GetDefaultConfig()
	mockClient := &MockHTTPClient{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

	service := NewPartnerValidationService(config, mockClient, logger)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	// Check if service implements the interface
	_, ok := service.(PartnerValidationService)
	if !ok {
		t.Error("Service does not implement PartnerValidationService interface")
	}
}

func TestPartnerValidationError(t *testing.T) {
	// Test error creation and methods
	testCode := constants.ErrorCodeInvalidPartner
	testMessage := "Test error message"
	testCause := context.DeadlineExceeded

	err := NewPartnerValidationError(testCode, testMessage, testCause)

	if err.GetErrorCode() != testCode {
		t.Errorf("Expected error code %s, got %s", testCode, err.GetErrorCode())
	}

	expectedError := testMessage + ": " + testCause.Error()
	if err.Error() != expectedError {
		t.Errorf("Expected error message %s, got %s", expectedError, err.Error())
	}

	if err.Unwrap() != testCause {
		t.Errorf("Expected unwrapped error to be %v, got %v", testCause, err.Unwrap())
	}
}

func TestPartnerValidationErrorWithoutCause(t *testing.T) {
	// Test error creation without cause
	testCode := constants.ErrorCodeInvalidPartner
	testMessage := "Test error message"

	err := NewPartnerValidationError(testCode, testMessage, nil)

	if err.Error() != testMessage {
		t.Errorf("Expected error message %s, got %s", testMessage, err.Error())
	}

	if err.Unwrap() != nil {
		t.Errorf("Expected unwrapped error to be nil, got %v", err.Unwrap())
	}
}

func TestValidatePartner_Success(t *testing.T) {
	mockClient := &MockHTTPClient{
		response: &HTTPResponse{
			StatusCode: constants.StatusOK,
			Body:       []byte(`{"id": "partner123", "status": "active"}`),
			Headers:    map[string]string{"Content-Type": "application/json"},
		},
		err: nil,
	}

	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := NewPartnerValidationService(config, mockClient, logger)

	ctx := context.Background()
	err := service.ValidatePartner(ctx, "partner123")

	if err != nil {
		t.Errorf("Expected no error for valid partner, got: %v", err)
	}
}

func TestValidatePartner_NotFound(t *testing.T) {
	mockClient := &MockHTTPClient{
		response: &HTTPResponse{
			StatusCode: constants.StatusNotFound,
			Body:       []byte(`{"error": "Partner not found"}`),
			Headers:    map[string]string{"Content-Type": "application/json"},
		},
		err: nil,
	}

	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := NewPartnerValidationService(config, mockClient, logger)

	ctx := context.Background()
	err := service.ValidatePartner(ctx, "nonexistent")

	if err == nil {
		t.Fatal("Expected error for non-existent partner, got nil")
	}

	partnerErr, ok := err.(*PartnerValidationError)
	if !ok {
		t.Fatalf("Expected PartnerValidationError, got %T", err)
	}

	if partnerErr.GetErrorCode() != constants.ErrorCodeInvalidPartner {
		t.Errorf("Expected error code %s, got %s", constants.ErrorCodeInvalidPartner, partnerErr.GetErrorCode())
	}
}

func TestValidatePartner_EmptyPartnerID(t *testing.T) {
	mockClient := &MockHTTPClient{}
	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := NewPartnerValidationService(config, mockClient, logger)

	ctx := context.Background()
	err := service.ValidatePartner(ctx, "")

	if err == nil {
		t.Fatal("Expected error for empty partner ID, got nil")
	}

	partnerErr, ok := err.(*PartnerValidationError)
	if !ok {
		t.Fatalf("Expected PartnerValidationError, got %T", err)
	}

	if partnerErr.GetErrorCode() != constants.ErrorCodeInvalidPartner {
		t.Errorf("Expected error code %s, got %s", constants.ErrorCodeInvalidPartner, partnerErr.GetErrorCode())
	}
}

func TestValidatePartner_Unauthorized(t *testing.T) {
	mockClient := &MockHTTPClient{
		response: &HTTPResponse{
			StatusCode: constants.StatusUnauthorized,
			Body:       []byte(`{"error": "Unauthorized"}`),
			Headers:    map[string]string{"Content-Type": "application/json"},
		},
		err: nil,
	}

	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := NewPartnerValidationService(config, mockClient, logger)

	ctx := context.Background()
	err := service.ValidatePartner(ctx, "partner123")

	if err == nil {
		t.Fatal("Expected error for unauthorized request, got nil")
	}

	partnerErr, ok := err.(*PartnerValidationError)
	if !ok {
		t.Fatalf("Expected PartnerValidationError, got %T", err)
	}

	if partnerErr.GetErrorCode() != constants.ErrorCodeAuthFailed {
		t.Errorf("Expected error code %s, got %s", constants.ErrorCodeAuthFailed, partnerErr.GetErrorCode())
	}
}

func TestValidatePartner_ServerError(t *testing.T) {
	mockClient := &MockHTTPClient{
		response: &HTTPResponse{
			StatusCode: constants.StatusInternalServerError,
			Body:       []byte(`{"error": "Internal server error"}`),
			Headers:    map[string]string{"Content-Type": "application/json"},
		},
		err: nil,
	}

	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := NewPartnerValidationService(config, mockClient, logger)

	ctx := context.Background()
	err := service.ValidatePartner(ctx, "partner123")

	if err == nil {
		t.Fatal("Expected error for server error, got nil")
	}

	partnerErr, ok := err.(*PartnerValidationError)
	if !ok {
		t.Fatalf("Expected PartnerValidationError, got %T", err)
	}

	if partnerErr.GetErrorCode() != constants.ErrorCodePartnerServiceError {
		t.Errorf("Expected error code %s, got %s", constants.ErrorCodePartnerServiceError, partnerErr.GetErrorCode())
	}
}

func TestValidatePartner_NetworkError(t *testing.T) {
	mockClient := &MockHTTPClient{
		response: nil,
		err:      errors.New("network connection failed"),
	}

	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := NewPartnerValidationService(config, mockClient, logger)

	ctx := context.Background()
	err := service.ValidatePartner(ctx, "partner123")

	if err == nil {
		t.Fatal("Expected error for network failure, got nil")
	}

	partnerErr, ok := err.(*PartnerValidationError)
	if !ok {
		t.Fatalf("Expected PartnerValidationError, got %T", err)
	}

	if partnerErr.GetErrorCode() != constants.ErrorCodePartnerServiceError {
		t.Errorf("Expected error code %s, got %s", constants.ErrorCodePartnerServiceError, partnerErr.GetErrorCode())
	}

	if partnerErr.Unwrap() == nil {
		t.Error("Expected wrapped error for network failure")
	}
}

func TestValidatePartner_ContextTimeout(t *testing.T) {
	mockClient := &MockHTTPClient{
		response: nil,
		err:      context.DeadlineExceeded,
	}

	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := NewPartnerValidationService(config, mockClient, logger)

	ctx := context.Background()
	err := service.ValidatePartner(ctx, "partner123")

	if err == nil {
		t.Fatal("Expected error for context timeout, got nil")
	}

	partnerErr, ok := err.(*PartnerValidationError)
	if !ok {
		t.Fatalf("Expected PartnerValidationError, got %T", err)
	}

	if partnerErr.GetErrorCode() != constants.ErrorCodePartnerServiceError {
		t.Errorf("Expected error code %s, got %s", constants.ErrorCodePartnerServiceError, partnerErr.GetErrorCode())
	}
}

// URLCapturingHTTPClient implements HTTPClient and captures the URL for testing
type URLCapturingHTTPClient struct {
	response    *HTTPResponse
	err         error
	capturedURL string
}

func (m *URLCapturingHTTPClient) Get(ctx context.Context, url string) (*HTTPResponse, error) {
	m.capturedURL = url
	return m.response, m.err
}

func TestValidatePartner_URLConstruction(t *testing.T) {
	mockClient := &URLCapturingHTTPClient{
		response: &HTTPResponse{
			StatusCode: constants.StatusOK,
			Body:       []byte(`{"id": "partner123", "status": "active"}`),
			Headers:    map[string]string{"Content-Type": "application/json"},
		},
		err: nil,
	}

	config := PartnerValidationConfig{
		BaseURL:        "https://api.example.com",
		RequestTimeout: 30 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
	}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := NewPartnerValidationService(config, mockClient, logger)

	ctx := context.Background()
	err := service.ValidatePartner(ctx, "partner123")

	if err != nil {
		t.Errorf("Expected no error for valid partner, got: %v", err)
	}

	expectedURL := "https://api.example.com/partner/v1/partners/partner123"
	if mockClient.capturedURL != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, mockClient.capturedURL)
	}
}

// Additional tests for improved coverage

func TestGetEnvInt_WithValidValue(t *testing.T) {
	originalValue := os.Getenv("TEST_INT_VAR")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_INT_VAR")
		} else {
			os.Setenv("TEST_INT_VAR", originalValue)
		}
	}()

	os.Setenv("TEST_INT_VAR", "42")
	result := getEnvInt("TEST_INT_VAR", 10)

	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

func TestGetEnvInt_WithInvalidValue(t *testing.T) {
	originalValue := os.Getenv("TEST_INT_VAR")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_INT_VAR")
		} else {
			os.Setenv("TEST_INT_VAR", originalValue)
		}
	}()

	os.Setenv("TEST_INT_VAR", "invalid")
	result := getEnvInt("TEST_INT_VAR", 10)

	if result != 10 {
		t.Errorf("Expected default value 10, got %d", result)
	}
}

func TestGetEnvDuration_WithValidValue(t *testing.T) {
	originalValue := os.Getenv("TEST_DURATION_VAR")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_DURATION_VAR")
		} else {
			os.Setenv("TEST_DURATION_VAR", originalValue)
		}
	}()

	os.Setenv("TEST_DURATION_VAR", "5s")
	result := getEnvDuration("TEST_DURATION_VAR", 10*time.Second)

	if result != 5*time.Second {
		t.Errorf("Expected 5s, got %v", result)
	}
}

func TestGetEnvDuration_WithInvalidValue(t *testing.T) {
	originalValue := os.Getenv("TEST_DURATION_VAR")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_DURATION_VAR")
		} else {
			os.Setenv("TEST_DURATION_VAR", originalValue)
		}
	}()

	os.Setenv("TEST_DURATION_VAR", "invalid")
	result := getEnvDuration("TEST_DURATION_VAR", 10*time.Second)

	if result != 10*time.Second {
		t.Errorf("Expected default value 10s, got %v", result)
	}
}

func TestNewPartnerValidationService_WithNilLogger(t *testing.T) {
	config := GetDefaultConfig()
	mockClient := &MockHTTPClient{}

	service := NewPartnerValidationService(config, mockClient, nil)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	// Verify that the service still works with nil logger
	ctx := context.Background()
	err := service.ValidatePartner(ctx, "")

	if err == nil {
		t.Error("Expected error for empty partner ID, got nil")
	}
}

// Test for HTTP client creation
func TestPartnerHTTPClient_Creation(t *testing.T) {
	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	client := NewPartnerHTTPClient(config, logger)

	if client == nil {
		t.Error("Expected client to be created, got nil")
	}

	// Test that client implements HTTPClient interface
	_, ok := client.(HTTPClient)
	if !ok {
		t.Error("Client does not implement HTTPClient interface")
	}
}

// Test HTTP client behavior indirectly through integration tests
func TestPartnerHTTPClient_RetryBehavior(t *testing.T) {
	// This test verifies retry behavior through actual HTTP calls
	config := GetDefaultConfig()
	config.RetryAttempts = 2
	config.RetryDelay = 50 * time.Millisecond // Faster for tests

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 2 {
			// First two calls return server error (retryable)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Third call succeeds
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "test-partner"}`))
	}))
	defer server.Close()

	config.BaseURL = server.URL
	client := NewPartnerHTTPClient(config, logger)

	ctx := context.Background()
	response, err := client.Get(ctx, server.URL+"/partner/v1/partners/test")

	if err != nil {
		t.Fatalf("Expected request to succeed after retries, got: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	// Verify that retries actually happened (should be called 3 times: initial + 2 retries)
	if callCount != 3 {
		t.Errorf("Expected 3 calls (with retries), got %d", callCount)
	}
}
