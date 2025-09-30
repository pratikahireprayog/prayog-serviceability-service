package aramex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"prayog-serviceability-service/internal/shared/config"

	"github.com/sirupsen/logrus"
)

// AramexClient handles HTTP communication with Aramex API
type AramexClient struct {
	httpClient *http.Client
	config     config.AramexConfig
	logger     *logrus.Logger
}

// AramexAPIError represents a structured Aramex API error
type AramexAPIError struct {
	StatusCode int
	RawBody    string
}

func (e *AramexAPIError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("Aramex API error [%d]: %s", e.StatusCode, e.RawBody)
}

// NewAramexClient creates a new Aramex HTTP client
func NewAramexClient(config config.AramexConfig) *AramexClient {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	// Log configuration with redacted sensitive fields
	logger.WithFields(logrus.Fields{
		"partner":              "Aramex",
		"base_url":             config.BaseURL,
		"username":             redactCredentialField(config.Username),
		"account_number":       redactCredentialField(config.AccountNumber),
		"account_entity":       config.AccountEntity,
		"account_country_code": config.AccountCountryCode,
		"timeout":              config.Timeout,
	}).Info("Creating Aramex client")

	return &AramexClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// redactCredentialField safely redacts sensitive credential fields for logging
func redactCredentialField(field string) string {
	if field == "" {
		return "[EMPTY]"
	}
	if len(field) <= 3 {
		return "[REDACTED]"
	}
	return field[:3] + "***"
}

// truncateForLog safely truncates large strings for logging
func truncateForLog(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	if limit < 0 {
		return ""
	}
	return s[:limit] + "...(truncated)"
}

// CheckServiceability calls Aramex serviceability API
func (c *AramexClient) CheckServiceability(ctx context.Context, request ServiceabilityRequest) (*ServiceabilityResponse, error) {
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Log request with appropriate level of detail
	url := c.config.BaseURL + "/ShippingAPI.V2/Location/Service_1_0.svc/json/IsAddressServiced"
	c.logger.WithFields(logrus.Fields{
		"partner":   "Aramex",
		"method":    "POST",
		"url":       url,
		"body_size": len(reqBody),
	}).Info("Making Aramex API request")

	// Log request body (truncated)
	c.logger.WithFields(logrus.Fields{
		"partner":      "Aramex",
		"event":        "request_body",
		"body_preview": truncateForLog(string(reqBody), 2048),
		"body_size":    len(reqBody),
	}).Info("Aramex serviceability request body")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add retry logic
	var resp *http.Response
	var lastErr error
	maxRetries := 3
	retryDelay := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay * time.Duration(attempt))
		}

		resp, lastErr = c.httpClient.Do(req)
		if lastErr == nil && resp.StatusCode < 500 {
			break // Success or client error (don't retry)
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response details safely
	c.logger.WithFields(logrus.Fields{
		"partner":     "Aramex",
		"status_code": resp.StatusCode,
		"body_size":   len(body),
	}).Info("Received Aramex API response")

	// Log response body (truncated)
	c.logger.WithFields(logrus.Fields{
		"partner":      "Aramex",
		"event":        "response_body",
		"status_code":  resp.StatusCode,
		"body_preview": truncateForLog(string(body), 4096),
		"body_size":    len(body),
	}).Info("Aramex serviceability response body")

	if resp.StatusCode != http.StatusOK {
		return nil, &AramexAPIError{
			StatusCode: resp.StatusCode,
			RawBody:    string(body),
		}
	}

	var serviceabilityResp ServiceabilityResponse
	if err := json.Unmarshal(body, &serviceabilityResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &serviceabilityResp, nil
}
