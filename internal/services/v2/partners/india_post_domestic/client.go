package india_post_domestic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"

	"github.com/sirupsen/logrus"
)

// Client implements the common.HTTPClient interface for India Post Domestic
type Client struct {
	httpClient *http.Client
	config     config.IndiaPostDomesticConfig
	auth       common.Authenticator
	logger     *logrus.Logger
}

// NewClient creates a new India Post Domestic HTTP client
func NewClient(cfg config.IndiaPostDomesticConfig, auth common.Authenticator) *Client {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
		auth:   auth,
		logger: logger,
	}
}

// SearchPincode searches for postal offices by pincode
func (c *Client) SearchPincode(ctx context.Context, pincode string) (*PincodeSearchResponse, error) {
	// Ensure we're authenticated
	if !c.auth.IsAuthenticated() {
		if err := c.auth.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Build API URL with query parameters
	apiURL := fmt.Sprintf("%s%s?pincode=%s&limit=%d&office-type=%s",
		c.config.BaseURL,
		c.config.PincodeSearchURL,
		pincode,
		c.config.SearchLimit,
		c.config.OfficeType,
	)

	c.logger.WithFields(logrus.Fields{
		"partner": "IndiaPostDomestic",
		"action":  "search_pincode",
		"pincode": pincode,
		"url":     apiURL,
	}).Info("Searching pincode in India Post Domestic")

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication headers
	for key, value := range c.auth.GetAuthHeaders() {
		req.Header.Set(key, value)
	}

	// Make the request with retry logic
	resp, err := c.executeWithRetry(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"partner":     "IndiaPostDomestic",
		"status_code": resp.StatusCode,
		"body_size":   len(body),
	}).Debug("Received India Post Domestic API response")

	// Check for HTTP errors
	if resp.StatusCode == http.StatusUnauthorized {
		// Token might have expired, try to refresh and retry once
		c.logger.Warn("Received 401, attempting to re-authenticate")
		if err := c.auth.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("re-authentication failed: %w", err)
		}
		
		// Retry the request once with new token
		for key, value := range c.auth.GetAuthHeaders() {
			req.Header.Set(key, value)
		}
		
		resp2, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("retry request failed: %w", err)
		}
		defer resp2.Body.Close()
		
		body, err = io.ReadAll(resp2.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read retry response: %w", err)
		}
		resp = resp2
	}

	// Parse response
	var pincodeResp PincodeSearchResponse
	if err := json.Unmarshal(body, &pincodeResp); err != nil {
		c.logger.WithFields(logrus.Fields{
			"partner":     "IndiaPostDomestic",
			"error":       err.Error(),
			"body":        string(body),
			"status_code": resp.StatusCode,
		}).Error("Failed to parse India Post Domestic response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check response success flag
	if !pincodeResp.Success {
	c.logger.WithFields(logrus.Fields{
		"partner":     "IndiaPostDomestic",
		"pincode":     pincode,
		"message":     pincodeResp.Message,
		"status_code": pincodeResp.StatusCode,
	}).Warn("India Post Domestic API returned unsuccessful response")
		
		// This is not an error - just means no data found
		// Return empty response
		return &pincodeResp, nil
	}

	c.logger.WithFields(logrus.Fields{
		"partner":       "IndiaPostDomestic",
		"pincode":       pincode,
		"offices_found": pincodeResp.ReturnedRecordsCount,
	}).Info("Successfully retrieved India Post Domestic data")

	return &pincodeResp, nil
}

// GetRouteByPincode calls the HubOps serviceability by route API
func (c *Client) GetRouteByPincode(ctx context.Context, sourcePostalCode, destinationPostalCode string) (map[string]interface{}, error) {
	// Build API URL
	apiURL := "https://apis-hubops.innofulfill.com/smcs-webapp/hubops-serviceability/by-route"

	// Create request body - postal codes should be integers
	sourcePincode, err := strconv.Atoi(sourcePostalCode)
	if err != nil {
		return nil, fmt.Errorf("invalid source postal code: %w", err)
	}
	
	destPincode, err := strconv.Atoi(destinationPostalCode)
	if err != nil {
		return nil, fmt.Errorf("invalid destination postal code: %w", err)
	}

	reqBody := map[string]interface{}{
		"sourcePostalCode":      sourcePincode,
		"destinationPostalCode": destPincode,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"partner":                "IndiaPostDomestic",
		"action":                 "get_route_by_pincode",
		"source_postal_code":     sourcePostalCode,
		"destination_postal_code": destinationPostalCode,
		"url":                    apiURL,
	}).Info("Calling HubOps serviceability by route API")

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"partner":     "IndiaPostDomestic",
		"status_code": resp.StatusCode,
		"body_size":   len(body),
	}).Debug("Received HubOps API response")

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.WithFields(logrus.Fields{
			"partner":     "IndiaPostDomestic",
			"status_code": resp.StatusCode,
			"body":        string(body),
		}).Warn("HubOps API returned error status")
		// Don't fail the entire request if hubops API fails, just log and return nil
		return nil, fmt.Errorf("HubOps API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response as map to preserve structure
	var routeResp map[string]interface{}
	if err := json.Unmarshal(body, &routeResp); err != nil {
		c.logger.WithFields(logrus.Fields{
			"partner":     "IndiaPostDomestic",
			"error":       err.Error(),
			"body":        string(body),
			"status_code": resp.StatusCode,
		}).Error("Failed to parse HubOps response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"partner":                "IndiaPostDomestic",
		"source_postal_code":     sourcePostalCode,
		"destination_postal_code": destinationPostalCode,
		"has_destination_3pl_hub": routeResp["destination3PLHub"] != nil,
	}).Info("Successfully retrieved HubOps route data")

	return routeResp, nil
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

			c.logger.WithFields(logrus.Fields{
				"partner": "IndiaPostDomestic",
				"attempt": attempt,
			}).Debug("Retrying India Post Domestic request")
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

		// Success or client error (don't retry on client errors except 401)
		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %v", maxRetries, lastErr)
}

// Get implements common.HTTPClient.Get
func (c *Client) Get(ctx context.Context, endpoint string, params map[string]string) (*common.HTTPResponse, error) {
	// Ensure authentication
	if !c.auth.IsAuthenticated() {
		if err := c.auth.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

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
		Duration:   0,
	}, nil
}

// Post implements common.HTTPClient.Post
func (c *Client) Post(ctx context.Context, endpoint string, body interface{}) (*common.HTTPResponse, error) {
	// Ensure authentication
	if !c.auth.IsAuthenticated() {
		if err := c.auth.Authenticate(ctx); err != nil {
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
	// Similar to Post but with PUT method
	return c.Post(ctx, endpoint, body) // Simplified for now
}

// Delete implements common.HTTPClient.Delete
func (c *Client) Delete(ctx context.Context, endpoint string) (*common.HTTPResponse, error) {
	// Ensure authentication
	if !c.auth.IsAuthenticated() {
		if err := c.auth.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

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
	// Could implement retry policy configuration here if needed
}

