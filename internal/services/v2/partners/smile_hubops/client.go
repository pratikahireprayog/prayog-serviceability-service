package smile_hubops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// HubOpsClient handles HTTP communication with the Smile HubOps API
type HubOpsClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewHubOpsClient creates a new Smile HubOps client
func NewHubOpsClient(baseURL string, httpClient *http.Client) *HubOpsClient {
	return &HubOpsClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logrus.New(),
	}
}

// CheckServiceability checks serviceability between source and destination postal codes
func (c *HubOpsClient) CheckServiceability(ctx context.Context, sourcePostalCode, destinationPostalCode string) (*HubOpsResponse, error) {
	// Convert postal codes to integers
	sourceCode, err := strconv.Atoi(sourcePostalCode)
	if err != nil {
		return nil, fmt.Errorf("invalid source postal code: %w", err)
	}

	destCode, err := strconv.Atoi(destinationPostalCode)
	if err != nil {
		return nil, fmt.Errorf("invalid destination postal code: %w", err)
	}

	// Create request payload
	request := HubOpsRequest{
		SourcePostalCode:      sourceCode,
		DestinationPostalCode: destCode,
	}

	c.logger.WithFields(logrus.Fields{
		"component":           "smile_hubops_client",
		"action":              "check_serviceability",
		"source_postal_code":  sourcePostalCode,
		"dest_postal_code":    destinationPostalCode,
		"request_payload":     request,
	}).Info("Starting HubOps serviceability check")

	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_client",
			"error":     err.Error(),
		}).Error("Failed to marshal request")
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/smcs-webapp/hubops-serviceability/by-route", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_client",
			"error":     err.Error(),
		}).Error("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Execute request with timeout
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	responseTime := time.Since(startTime)

	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"component":    "smile_hubops_client",
			"error":        err.Error(),
			"response_time": responseTime,
		}).Error("HTTP request failed")
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	c.logger.WithFields(logrus.Fields{
		"component":     "smile_hubops_client",
		"status_code":   resp.StatusCode,
		"response_time": responseTime,
	}).Info("Received HubOps API response")

	// Check for timeout errors
	if ctx.Err() == context.DeadlineExceeded {
		c.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_client",
			"error":     "request timeout",
		}).Error("HubOps API request timed out")
		return nil, fmt.Errorf("HubOps API request timed out")
	}

	// Read the response body once
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_client",
			"error":     err.Error(),
		}).Error("Failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code first
	if resp.StatusCode == 404 {
		// Parse the error response body
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &errorResponse); err != nil {
			c.logger.WithFields(logrus.Fields{
				"component":   "smile_hubops_client",
				"status_code": resp.StatusCode,
				"error":       "failed to parse 404 response",
			}).Error("Failed to parse 404 response body")
		} else {
			c.logger.WithFields(logrus.Fields{
				"component":   "smile_hubops_client",
				"status_code": resp.StatusCode,
				"detail":      errorResponse["detail"],
			}).Info("HubOps API returned 404 - service not available")
		}
		return nil, fmt.Errorf("service not available (404)")
	}
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.WithFields(logrus.Fields{
			"component":   "smile_hubops_client",
			"status_code": resp.StatusCode,
		}).Error("HubOps API returned error status")
		return nil, fmt.Errorf("HubOps API returned status %d", resp.StatusCode)
	}

	// Parse the raw response
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawResponse); err != nil {
		c.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_client",
			"error":     err.Error(),
		}).Error("Failed to parse raw response")
		return nil, fmt.Errorf("failed to parse raw response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"component":        "smile_hubops_client",
		"raw_response_keys": len(rawResponse),
		"raw_response":     rawResponse,
	}).Info("Parsed raw response successfully")

	// Also parse into structured response for internal use
	var hubOpsResponse HubOpsResponse
	if err := json.Unmarshal(bodyBytes, &hubOpsResponse); err != nil {
		c.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_client",
			"error":     err.Error(),
		}).Warn("Failed to parse structured response, using raw response")
	}

	// Store raw response in the structured response
	hubOpsResponse.RawResponse = rawResponse

	c.logger.WithFields(logrus.Fields{
		"component":        "smile_hubops_client",
		"structured_parse_success": err == nil,
		"delivery_available":       hubOpsResponse.DeliveryAvailable,
		"route_count":              len(hubOpsResponse.Routes),
		"has_raw_response":         hubOpsResponse.RawResponse != nil,
	}).Info("Parsed structured response")

	c.logger.WithFields(logrus.Fields{
		"component":           "smile_hubops_client",
		"delivery_available":  hubOpsResponse.DeliveryAvailable,
		"route_count":         len(hubOpsResponse.Routes),
		"response_time":       responseTime,
		"raw_response_keys":   len(rawResponse),
		"has_raw_response":    hubOpsResponse.RawResponse != nil,
	}).Info("HubOps serviceability check completed successfully")

	return &hubOpsResponse, nil
}

// IsHealthy checks if the HubOps API is healthy
func (c *HubOpsClient) IsHealthy(ctx context.Context) bool {
	// Create a simple health check request
	request := HubOpsRequest{
		SourcePostalCode:      110001, // Delhi
		DestinationPostalCode: 110001, // Delhi (same as source for health check)
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_client",
			"error":     err.Error(),
		}).Error("Failed to marshal health check request")
		return false
	}

	// Create HTTP request with short timeout for health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/smcs-webapp/hubops-serviceability/by-route", c.baseURL)
	req, err := http.NewRequestWithContext(healthCtx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_client",
			"error":     err.Error(),
		}).Error("Failed to create health check request")
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_client",
			"error":     err.Error(),
		}).Error("Health check request failed")
		return false
	}
	defer resp.Body.Close()

	// Consider healthy if we get any response (even error responses indicate the service is reachable)
	isHealthy := resp.StatusCode >= 200 && resp.StatusCode < 500

	c.logger.WithFields(logrus.Fields{
		"component":   "smile_hubops_client",
		"status_code": resp.StatusCode,
		"is_healthy":  isHealthy,
	}).Info("HubOps health check completed")

	return isHealthy
}
