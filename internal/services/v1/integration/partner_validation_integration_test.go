package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"prayog-serviceability-service/internal/shared/constants/v1"
)

// PartnerAPIResponse represents the response from the Partner API
type PartnerAPIResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ErrorResponse represents an error response from the Partner API
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// TestPartnerValidationService_Integration tests the complete partner validation flow with realistic HTTP mocking
func TestPartnerValidationService_Integration(t *testing.T) {
	tests := []struct {
		name             string
		partnerID        string
		mockHandler      http.HandlerFunc
		expectedError    bool
		expectedCode     string
		expectValidation bool
	}{
		{
			name:      "Valid Partner",
			partnerID: "valid-partner-123",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.Path, "/partner/v1/partners/valid-partner-123") {
					t.Errorf("Expected URL path to end with /partner/v1/partners/valid-partner-123, got %s", r.URL.Path)
				}

				// Check headers
				if r.Header.Get("Accept") != "application/json" {
					t.Errorf("Expected Accept header to be application/json, got %s", r.Header.Get("Accept"))
				}

				response := PartnerAPIResponse{
					ID:     "valid-partner-123",
					Name:   "Test Partner",
					Status: "active",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			},
			expectedError:    false,
			expectValidation: true,
		},
		{
			name:      "Partner Not Found",
			partnerID: "nonexistent-partner",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				errorResponse := ErrorResponse{
					Error:   "Partner not found",
					Message: "The requested partner does not exist",
					Code:    "PARTNER_NOT_FOUND",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(errorResponse)
			},
			expectedError:    true,
			expectedCode:     constants.ErrorCodeInvalidPartner,
			expectValidation: false,
		},
		{
			name:      "Unauthorized Access",
			partnerID: "partner-123",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				errorResponse := ErrorResponse{
					Error:   "Unauthorized",
					Message: "Invalid or missing authentication credentials",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(errorResponse)
			},
			expectedError:    true,
			expectedCode:     constants.ErrorCodeAuthFailed,
			expectValidation: false,
		},
		{
			name:      "Server Error",
			partnerID: "partner-456",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				errorResponse := ErrorResponse{
					Error:   "Internal Server Error",
					Message: "An unexpected error occurred",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(errorResponse)
			},
			expectedError:    true,
			expectedCode:     constants.ErrorCodePartnerServiceError,
			expectValidation: false,
		},
		{
			name:      "Service Unavailable with Retry",
			partnerID: "partner-789",
			mockHandler: func() http.HandlerFunc {
				callCount := 0
				return func(w http.ResponseWriter, r *http.Request) {
					callCount++
					if callCount <= 2 {
						// First two calls return service unavailable
						w.WriteHeader(http.StatusServiceUnavailable)
						w.Write([]byte("Service temporarily unavailable"))
						return
					}
					// Third call succeeds
					response := PartnerAPIResponse{
						ID:     "partner-789",
						Name:   "Retry Test Partner",
						Status: "active",
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(response)
				}
			}(),
			expectedError:    false,
			expectValidation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(tt.mockHandler)
			defer server.Close()

			// Create configuration with test server URL
			config := PartnerValidationConfig{
				BaseURL:        server.URL,
				RequestTimeout: 5 * time.Second,
				RetryAttempts:  3,
				RetryDelay:     100 * time.Millisecond, // Faster for tests
			}

			// Create HTTP client and service
			logger := log.New(os.Stdout, fmt.Sprintf("TEST[%s]: ", tt.name), log.LstdFlags)
			httpClient := NewPartnerHTTPClient(config, logger)
			service := NewPartnerValidationService(config, httpClient, logger)

			// Test validation
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := service.ValidatePartner(ctx, tt.partnerID)

			// Verify results
			if tt.expectedError {
				if err == nil {
					t.Fatalf("Expected error for test case %s, got nil", tt.name)
				}

				partnerErr, ok := err.(*PartnerValidationError)
				if !ok {
					t.Fatalf("Expected PartnerValidationError for test case %s, got %T", tt.name, err)
				}

				if partnerErr.GetErrorCode() != tt.expectedCode {
					t.Errorf("Expected error code %s for test case %s, got %s", tt.expectedCode, tt.name, partnerErr.GetErrorCode())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error for test case %s, got: %v", tt.name, err)
				}
			}
		})
	}
}

// TestPartnerValidationService_Timeout tests timeout handling with slow server
func TestPartnerValidationService_Timeout(t *testing.T) {
	// Create a slow server that takes longer than the client timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than client timeout
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "slow-partner"}`))
	}))
	defer server.Close()

	config := PartnerValidationConfig{
		BaseURL:        server.URL,
		RequestTimeout: 500 * time.Millisecond, // Short timeout
		RetryAttempts:  1,
		RetryDelay:     100 * time.Millisecond,
	}

	logger := log.New(os.Stdout, "TEST[Timeout]: ", log.LstdFlags)
	httpClient := NewPartnerHTTPClient(config, logger)
	service := NewPartnerValidationService(config, httpClient, logger)

	ctx := context.Background()
	err := service.ValidatePartner(ctx, "slow-partner")

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	partnerErr, ok := err.(*PartnerValidationError)
	if !ok {
		t.Fatalf("Expected PartnerValidationError, got %T", err)
	}

	if partnerErr.GetErrorCode() != constants.ErrorCodePartnerServiceError {
		t.Errorf("Expected error code %s, got %s", constants.ErrorCodePartnerServiceError, partnerErr.GetErrorCode())
	}
}

// TestPartnerValidationService_ContextCancellation tests context cancellation
func TestPartnerValidationService_ContextCancellation(t *testing.T) {
	// Create a server that responds slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "cancelled-partner"}`))
	}))
	defer server.Close()

	config := PartnerValidationConfig{
		BaseURL:        server.URL,
		RequestTimeout: 5 * time.Second,
		RetryAttempts:  1,
		RetryDelay:     100 * time.Millisecond,
	}

	logger := log.New(os.Stdout, "TEST[Cancellation]: ", log.LstdFlags)
	httpClient := NewPartnerHTTPClient(config, logger)
	service := NewPartnerValidationService(config, httpClient, logger)

	// Create context that gets cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := service.ValidatePartner(ctx, "cancelled-partner")

	if err == nil {
		t.Fatal("Expected cancellation error, got nil")
	}

	// Should be either a context error or partner service error wrapping the context error
	if err != context.DeadlineExceeded {
		partnerErr, ok := err.(*PartnerValidationError)
		if !ok {
			t.Fatalf("Expected context error or PartnerValidationError, got %T", err)
		}
		if partnerErr.GetErrorCode() != constants.ErrorCodePartnerServiceError {
			t.Errorf("Expected error code %s, got %s", constants.ErrorCodePartnerServiceError, partnerErr.GetErrorCode())
		}
	}
}

// TestPartnerValidationService_EmptyResponse tests handling of empty/invalid responses
func TestPartnerValidationService_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Send empty response body
	}))
	defer server.Close()

	config := PartnerValidationConfig{
		BaseURL:        server.URL,
		RequestTimeout: 5 * time.Second,
		RetryAttempts:  1,
		RetryDelay:     100 * time.Millisecond,
	}

	logger := log.New(os.Stdout, "TEST[EmptyResponse]: ", log.LstdFlags)
	httpClient := NewPartnerHTTPClient(config, logger)
	service := NewPartnerValidationService(config, httpClient, logger)

	ctx := context.Background()
	err := service.ValidatePartner(ctx, "empty-response-partner")

	// Should succeed since we only check status code, not response body content
	if err != nil {
		t.Fatalf("Expected no error for empty response with 200 status, got: %v", err)
	}
}

// BenchmarkPartnerValidation benchmarks the partner validation performance
func BenchmarkPartnerValidation(b *testing.B) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PartnerAPIResponse{
			ID:     "benchmark-partner",
			Name:   "Benchmark Partner",
			Status: "active",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := PartnerValidationConfig{
		BaseURL:        server.URL,
		RequestTimeout: 5 * time.Second,
		RetryAttempts:  1,
		RetryDelay:     100 * time.Millisecond,
	}

	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	httpClient := NewPartnerHTTPClient(config, logger)
	service := NewPartnerValidationService(config, httpClient, logger)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.ValidatePartner(ctx, "benchmark-partner")
		if err != nil {
			b.Fatalf("Benchmark failed with error: %v", err)
		}
	}
}
