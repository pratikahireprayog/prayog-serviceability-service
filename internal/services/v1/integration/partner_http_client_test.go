package services

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewPartnerHTTPClient(t *testing.T) {
	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

	client := NewPartnerHTTPClient(config, logger)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	// Check if client implements HTTPClient interface
	_, ok := client.(HTTPClient)
	if !ok {
		t.Error("Client does not implement HTTPClient interface")
	}

	// Check metrics
	metrics := client.(*PartnerHTTPClient).GetMetrics()
	if metrics["retry_attempts"] != config.RetryAttempts {
		t.Errorf("Expected retry_attempts %d, got %v", config.RetryAttempts, metrics["retry_attempts"])
	}
}

func TestHTTPClientSuccessfulRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("User-Agent") != "prayog-serviceability-service/1.0" {
			t.Errorf("Expected User-Agent header, got %s", r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept header, got %s", r.Header.Get("Accept"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"partner_id": "123", "status": "active"}`))
	}))
	defer server.Close()

	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	client := NewPartnerHTTPClient(config, logger)

	ctx := context.Background()
	response, err := client.Get(ctx, server.URL)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	expectedBody := `{"partner_id": "123", "status": "active"}`
	if string(response.Body) != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, string(response.Body))
	}

	if response.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header, got %s", response.Headers["Content-Type"])
	}
}

func TestHTTPClientRetryOnServerError(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			// Return server error for first 2 attempts
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		} else {
			// Success on 3rd attempt
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		}
	}))
	defer server.Close()

	config := PartnerValidationConfig{
		BaseURL:        server.URL,
		RequestTimeout: 10 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     10 * time.Millisecond, // Short delay for testing
	}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	client := NewPartnerHTTPClient(config, logger)

	ctx := context.Background()
	response, err := client.Get(ctx, server.URL)

	if err != nil {
		t.Fatalf("Expected no error after retries, got %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

func TestHTTPClientRetryExhaustion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return server error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	config := PartnerValidationConfig{
		BaseURL:        server.URL,
		RequestTimeout: 5 * time.Second,
		RetryAttempts:  2,
		RetryDelay:     10 * time.Millisecond,
	}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	client := NewPartnerHTTPClient(config, logger)

	ctx := context.Background()
	response, err := client.Get(ctx, server.URL)

	if err == nil {
		t.Fatal("Expected error after retry exhaustion, got nil")
	}

	if response != nil {
		t.Error("Expected no response on retry exhaustion, got response")
	}

	if !strings.Contains(err.Error(), "failed after") {
		t.Errorf("Expected retry exhaustion error, got %v", err)
	}
}

func TestHTTPClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	client := NewPartnerHTTPClient(config, logger)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Get(ctx, server.URL)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "cancelled") && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected cancellation/timeout error, got %v", err)
	}
}

func TestHTTPClientNonRetryableError(t *testing.T) {
	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	client := NewPartnerHTTPClient(config, logger)

	ctx := context.Background()
	// Use invalid URL to trigger non-retryable error
	_, err := client.Get(ctx, "http://invalid-url-that-does-not-exist.local")

	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}

	// Should not contain retry exhaustion message for non-retryable errors
	if strings.Contains(err.Error(), "failed after") {
		t.Errorf("Expected non-retryable error, but got retry exhaustion: %v", err)
	}
}

func TestHTTPClientRetryableStatusCodes(t *testing.T) {
	testCases := []struct {
		statusCode  int
		shouldRetry bool
		description string
	}{
		{http.StatusOK, false, "200 OK"},
		{http.StatusNotFound, false, "404 Not Found"},
		{http.StatusBadRequest, false, "400 Bad Request"},
		{http.StatusTooManyRequests, true, "429 Too Many Requests"},
		{http.StatusInternalServerError, true, "500 Internal Server Error"},
		{http.StatusBadGateway, true, "502 Bad Gateway"},
		{http.StatusServiceUnavailable, true, "503 Service Unavailable"},
		{http.StatusGatewayTimeout, true, "504 Gateway Timeout"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			config := GetDefaultConfig()
			logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
			client := NewPartnerHTTPClient(config, logger).(*PartnerHTTPClient)

			result := client.shouldRetryStatus(tc.statusCode)
			if result != tc.shouldRetry {
				t.Errorf("Status %d: expected shouldRetry=%v, got %v", tc.statusCode, tc.shouldRetry, result)
			}
		})
	}
}

func TestHTTPClientRetryDelayCalculation(t *testing.T) {
	config := PartnerValidationConfig{
		RetryDelay: 100 * time.Millisecond,
	}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	client := NewPartnerHTTPClient(config, logger).(*PartnerHTTPClient)

	// Test exponential backoff
	delay0 := client.calculateRetryDelay(0)
	delay1 := client.calculateRetryDelay(1)
	delay2 := client.calculateRetryDelay(2)

	// Delays should generally increase (accounting for jitter)
	if delay0 > delay1*2 || delay1 > delay2*2 {
		t.Errorf("Delays not following exponential backoff pattern: %v, %v, %v", delay0, delay1, delay2)
	}

	// All delays should be at least 100ms
	minDelay := 100 * time.Millisecond
	if delay0 < minDelay || delay1 < minDelay || delay2 < minDelay {
		t.Errorf("Delays below minimum: %v, %v, %v", delay0, delay1, delay2)
	}

	// Test max delay cap
	delay10 := client.calculateRetryDelay(10)
	maxDelay := 30 * time.Second
	if delay10 > maxDelay*2 { // Allow some jitter
		t.Errorf("Delay exceeds maximum: %v", delay10)
	}
}

func TestHTTPClientRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check required headers
		userAgent := r.Header.Get("User-Agent")
		accept := r.Header.Get("Accept")

		if userAgent != "prayog-serviceability-service/1.0" {
			t.Errorf("Expected User-Agent 'prayog-serviceability-service/1.0', got '%s'", userAgent)
		}

		if accept != "application/json" {
			t.Errorf("Expected Accept 'application/json', got '%s'", accept)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	client := NewPartnerHTTPClient(config, logger)

	ctx := context.Background()
	_, err := client.Get(ctx, server.URL)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHTTPClientMetrics(t *testing.T) {
	config := PartnerValidationConfig{
		BaseURL:        "http://test.com",
		RequestTimeout: 15 * time.Second,
		RetryAttempts:  5,
		RetryDelay:     2 * time.Second,
	}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	client := NewPartnerHTTPClient(config, logger).(*PartnerHTTPClient)

	metrics := client.GetMetrics()

	if metrics["retry_attempts"] != 5 {
		t.Errorf("Expected retry_attempts 5, got %v", metrics["retry_attempts"])
	}

	if metrics["retry_delay"] != "2s" {
		t.Errorf("Expected retry_delay '2s', got %v", metrics["retry_delay"])
	}

	if metrics["timeout"] != "15s" {
		t.Errorf("Expected timeout '15s', got %v", metrics["timeout"])
	}
}

func TestHTTPClientConnectionPooling(t *testing.T) {
	// Test that the client properly configures connection pooling
	config := GetDefaultConfig()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	client := NewPartnerHTTPClient(config, logger).(*PartnerHTTPClient)

	// Check that transport is properly configured
	transport, ok := client.client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport, got different type")
	}

	if transport.MaxIdleConns != 100 {
		t.Errorf("Expected MaxIdleConns 100, got %d", transport.MaxIdleConns)
	}

	if transport.MaxIdleConnsPerHost != 10 {
		t.Errorf("Expected MaxIdleConnsPerHost 10, got %d", transport.MaxIdleConnsPerHost)
	}

	if transport.IdleConnTimeout != 90*time.Second {
		t.Errorf("Expected IdleConnTimeout 90s, got %v", transport.IdleConnTimeout)
	}
}
