package shipyaari

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

// Client implements the common.HTTPClient interface for Shipyaari
type Client struct {
	httpClient *http.Client
	config     config.ShipyaariConfig
	auth       common.Authenticator
}

// NewClient creates a new Shipyaari HTTP client
func NewClient(cfg config.ShipyaariConfig, auth common.Authenticator) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
		auth:   auth,
	}
}

// CheckServiceability makes a serviceability check request to Shipyaari
func (c *Client) CheckServiceability(ctx context.Context, req *ServiceabilityRequest) (*ServiceabilityResponse, error) {
	// Ensure we're authenticated
	if !c.auth.IsAuthenticated() {
		if err := c.auth.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Prepare request body
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with full URL
	apiURL := c.config.BaseURL + c.config.CheckServiceURL
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Add authentication headers
	for key, value := range c.auth.GetAuthHeaders() {
		httpReq.Header.Set(key, value)
	}

	// Make the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors (but allow 500 for business logic errors)
	if resp.StatusCode < 200 || (resp.StatusCode >= 300 && resp.StatusCode != 500) {
		var errorResp ErrorResponse
		if json.Unmarshal(body, &errorResp) == nil {
			return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, errorResp.ErrorMessage)
		}
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var serviceabilityResp ServiceabilityResponse
	if err := json.Unmarshal(body, &serviceabilityResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Note: We don't throw an error for success: false here
	// The adapter will handle the business logic and set appropriate error messages

	return &serviceabilityResp, nil
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

	// Add authentication headers
	for key, value := range c.auth.GetAuthHeaders() {
		req.Header.Set(key, value)
	}

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

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Add authentication headers
	for key, value := range c.auth.GetAuthHeaders() {
		req.Header.Set(key, value)
	}

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

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Add authentication headers
	for key, value := range c.auth.GetAuthHeaders() {
		req.Header.Set(key, value)
	}

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

	// Add authentication headers
	for key, value := range c.auth.GetAuthHeaders() {
		req.Header.Set(key, value)
	}

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
	// Could implement retry logic here if needed
}
