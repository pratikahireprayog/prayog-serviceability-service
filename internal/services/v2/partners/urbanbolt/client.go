package urbanbolt

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
)

// Client implements the common.HTTPClient interface for UrbanBolt
type Client struct {
	httpClient   *http.Client
	config       config.UrbanBoltConfig
	authenticator common.Authenticator
}

// NewClient creates a new UrbanBolt HTTP client
func NewClient(cfg config.UrbanBoltConfig, auth common.Authenticator) *Client {
	// Create HTTP client with SSL verification disabled for UAT environment
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Bypass SSL certificate verification for UAT
		},
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: tr,
		},
		config:       cfg,
		authenticator: auth,
	}
}

// CheckServiceability makes a serviceability check request to UrbanBolt
func (c *Client) CheckServiceability(ctx context.Context, pincodes []string) (*ServiceabilityResponse, error) {
	// Ensure we're authenticated
	if !c.authenticator.IsAuthenticated() {
		if err := c.authenticator.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Build query parameters - pincodes should be comma-separated
	pincodeStr := strings.Join(pincodes, ",")
	
	// Create URL with query parameters
	apiURL := c.config.BaseURL + c.config.ServiceabilityURL
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	
	q := u.Query()
	q.Set("pincodes", pincodeStr)
	u.RawQuery = q.Encode()

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers
	authHeaders := c.authenticator.GetAuthHeaders()
	for key, value := range authHeaders {
		httpReq.Header.Set(key, value)
	}

	// Add other headers
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

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// If unauthorized, try to re-authenticate once
		if resp.StatusCode == http.StatusUnauthorized {
			if err := c.authenticator.RefreshAuth(ctx); err == nil {
				// Retry the request with new token
				authHeaders = c.authenticator.GetAuthHeaders()
				for key, value := range authHeaders {
					httpReq.Header.Set(key, value)
				}
				
				retryResp, retryErr := c.httpClient.Do(httpReq)
				if retryErr != nil {
					return nil, fmt.Errorf("retry request failed: %w", retryErr)
				}
				defer retryResp.Body.Close()
				
				body, err = io.ReadAll(retryResp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read retry response: %w", err)
				}
				
				if retryResp.StatusCode < 200 || retryResp.StatusCode >= 300 {
					return nil, fmt.Errorf("serviceability API returned status %d: %s", retryResp.StatusCode, string(body))
				}
				
				// Parse successful retry response
				var serviceabilityResp ServiceabilityResponse
				if err := json.Unmarshal(body, &serviceabilityResp); err != nil {
					return nil, fmt.Errorf("failed to parse response: %w", err)
				}
				return &serviceabilityResp, nil
			}
		}
		return nil, fmt.Errorf("serviceability API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var serviceabilityResp ServiceabilityResponse
	if err := json.Unmarshal(body, &serviceabilityResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &serviceabilityResp, nil
}

// executeWithRetry executes the request with retry logic
func (c *Client) executeWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.config.RetryDelay * time.Duration(attempt)):
			}
		}
		
		resp, err := c.httpClient.Do(req)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
	}
	
	return nil, fmt.Errorf("request failed after %d retries: %w", c.config.MaxRetries, lastErr)
}

// Get implements common.HTTPClient.Get
func (c *Client) Get(ctx context.Context, endpoint string, params map[string]string) (*common.HTTPResponse, error) {
	// Ensure authentication
	if !c.authenticator.IsAuthenticated() {
		if err := c.authenticator.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Build URL with query parameters
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if len(params) > 0 {
		q := u.Query()
		for key, value := range params {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	authHeaders := c.authenticator.GetAuthHeaders()
	for key, value := range authHeaders {
		req.Header.Set(key, value)
	}

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

// Post implements common.HTTPClient.Post
func (c *Client) Post(ctx context.Context, endpoint string, body interface{}) (*common.HTTPResponse, error) {
	// Ensure authentication
	if !c.authenticator.IsAuthenticated() {
		if err := c.authenticator.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

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
	authHeaders := c.authenticator.GetAuthHeaders()
	for key, value := range authHeaders {
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
	// Ensure authentication
	if !c.authenticator.IsAuthenticated() {
		if err := c.authenticator.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

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
	authHeaders := c.authenticator.GetAuthHeaders()
	for key, value := range authHeaders {
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
	// Ensure authentication
	if !c.authenticator.IsAuthenticated() {
		if err := c.authenticator.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	authHeaders := c.authenticator.GetAuthHeaders()
	for key, value := range authHeaders {
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
	// Store retry policy if needed
	// For now, retry logic is handled in executeWithRetry
}

