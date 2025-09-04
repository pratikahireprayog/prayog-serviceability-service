package smile_courier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
)

// Client implements the common.HTTPClient interface for Smile Courier
type Client struct {
	httpClient *http.Client
	config     config.SmileCourierConfig
}

// NewClient creates a new Smile Courier HTTP client
func NewClient(cfg config.SmileCourierConfig) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
	}
}

// CheckServiceability makes a serviceability check request to Smile Courier
func (c *Client) CheckServiceability(ctx context.Context, req *ServiceabilityRequest) (*ServiceabilityResponse, error) {
	// Prepare request body
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	apiURL := c.config.BaseURL + c.config.CheckServiceURL
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers (no authentication required for Smile Courier)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Make the request with retry logic
	resp, err := c.executeWithRetry(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

    // Parse response first
	var serviceabilityResp ServiceabilityResponse
	if err := json.Unmarshal(body, &serviceabilityResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

    // Debug: log raw HTTP status and body, and parsed summary
    fmt.Printf("[smile_courier] API response http_status=%d raw_body=%s\n", resp.StatusCode, string(body))
    // Note: serviceabilityResp.Data is RawMessage; serviceable logged by adapter after parsing

	// Check for HTTP errors (but allow 400 for business logic errors)
	if resp.StatusCode < 200 || (resp.StatusCode >= 300 && resp.StatusCode != 400) {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Note: We don't throw an error for status 400 here
	// The adapter will handle the business logic and set appropriate error messages

	return &serviceabilityResp, nil
}

// executeWithRetry executes HTTP request with retry logic
func (c *Client) executeWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error

	maxRetries := c.config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3 // Default retry count
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			retryDelay := c.config.RetryDelay
			if retryDelay <= 0 {
				retryDelay = time.Second // Default retry delay
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt)):
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Check if response indicates we should retry
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: server error", resp.StatusCode)
			continue
		}

		// Success or client error (don't retry)
		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %v", maxRetries, lastErr)
}

// Get implements common.HTTPClient.Get
func (c *Client) Get(ctx context.Context, endpoint string, params map[string]string) (*common.HTTPResponse, error) {
	// Build URL with query parameters
	url := endpoint
	if len(params) > 0 {
		url += "?"
		for key, value := range params {
			url += fmt.Sprintf("%s=%s&", key, value)
		}
		url = url[:len(url)-1] // Remove trailing &
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set standard headers (no auth needed)
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &common.HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
		Duration:   0, // Could be calculated if needed
	}, nil
}

// Post implements common.HTTPClient.Post
func (c *Client) Post(ctx context.Context, endpoint string, body interface{}) (*common.HTTPResponse, error) {
	// Marshal body
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Set headers (no auth needed)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &common.HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
		Duration:   0,
	}, nil
}

// Put implements common.HTTPClient.Put
func (c *Client) Put(ctx context.Context, endpoint string, body interface{}) (*common.HTTPResponse, error) {
	// Marshal body
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Set headers (no auth needed)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &common.HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
		Duration:   0,
	}, nil
}

// Delete implements common.HTTPClient.Delete
func (c *Client) Delete(ctx context.Context, endpoint string) (*common.HTTPResponse, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Set headers (no auth needed)
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &common.HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
		Duration:   0,
	}, nil
}

// GetBaseURL implements common.HTTPClient.GetBaseURL
func (c *Client) GetBaseURL() string {
	return c.config.BaseURL
}

// GetTimeout implements common.HTTPClient.GetTimeout
func (c *Client) GetTimeout() time.Duration {
	return c.config.Timeout
}

// SetRetryPolicy implements common.HTTPClient.SetRetryPolicy
func (c *Client) SetRetryPolicy(policy common.RetryPolicy) {
	// Could implement retry policy configuration here if needed
} 