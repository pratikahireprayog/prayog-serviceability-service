package smile_cargo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"prayog-serviceability-service/internal/shared/config"
)

// SmileCargoClient handles HTTP communication with Smile Cargo API
type SmileCargoClient struct {
	httpClient *http.Client
	config     config.SmileCargoConfig
}

// NewSmileCargoClient creates a new Smile Cargo HTTP client
func NewSmileCargoClient(config config.SmileCargoConfig) *SmileCargoClient {
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	return &SmileCargoClient{
		httpClient: httpClient,
		config:     config,
	}
}

// CheckServiceAvailability calls Smile Cargo service availability API
func (c *SmileCargoClient) CheckServiceAvailability(ctx context.Context, request ServiceAvailabilityRequest) (*ServiceAvailabilityResponse, error) {
	// Build query parameters
	params := url.Values{}
	params.Add("fromPincode", request.FromPincode)
	params.Add("toPincode", request.ToPincode)
	// Remove vendorCode as per business requirement
	// params.Add("vendorCode", request.VendorCode)

	// Construct URL
	apiURL := c.config.BaseURL + c.config.ServiceURL + "?" + params.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for URL %s: %w", apiURL, err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Prayog-Serviceability-Service/1.0")

	// Execute request with retry logic
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.config.RetryDelay * time.Duration(attempt))
		}

		resp, lastErr = c.httpClient.Do(req)
		if lastErr == nil && resp != nil && resp.StatusCode < 500 {
			break // Success or client error (don't retry)
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.config.MaxRetries, lastErr)
	}

	// Additional safety check (though this should not happen if lastErr == nil)
	if resp == nil {
		return nil, fmt.Errorf("received nil response despite no error after %d retries", c.config.MaxRetries)
	}

	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var serviceResp ServiceAvailabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&serviceResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &serviceResp, nil
}

// ConvertPincodeToString safely converts pincode to string
func ConvertPincodeToString(pincode interface{}) string {
	switch v := pincode.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case *int:
		if v != nil {
			return strconv.Itoa(*v)
		}
	case float64:
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	return ""
}

// ConvertStringToPincode safely converts string to pincode
func ConvertStringToPincode(pincodeStr string) (string, error) {
	if pincodeStr == "" {
		return "", fmt.Errorf("pincode cannot be empty")
	}

	// Validate pincode format (assuming 6 digits for Indian pincodes)
	if len(pincodeStr) != 6 {
		return "", fmt.Errorf("invalid pincode format: %s", pincodeStr)
	}

	// Check if it's numeric
	if _, err := strconv.Atoi(pincodeStr); err != nil {
		return "", fmt.Errorf("pincode must be numeric: %s", pincodeStr)
	}

	return pincodeStr, nil
}
