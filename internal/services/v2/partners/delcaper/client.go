package delcaper

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

// DelcaperClient handles HTTP communication with Delcaper API
type DelcaperClient struct {
	config      config.DelcaperConfig
	httpClient  *http.Client
	tokenManager *TokenManager
	logger      *logrus.Logger
}

// NewDelcaperClient creates a new Delcaper client
func NewDelcaperClient(cfg config.DelcaperConfig, httpClient *http.Client) *DelcaperClient {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create token manager
	tokenManager := NewTokenManager(cfg, httpClient)

	return &DelcaperClient{
		config:      cfg,
		httpClient:  httpClient,
		tokenManager: tokenManager,
		logger:      logger,
	}
}

// CheckFeasible checks serviceability for the given request
func (c *DelcaperClient) CheckFeasible(ctx context.Context, req *CheckFeasibleRequest) (*CheckFeasibleResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"order_type": req.OrderType,
		"pickup_zip": req.PickupAddress.Zip,
		"shipping_zip": req.ShippingAddress.Zip,
	}).Info("Starting Delcaper feasibility check")
	
	// Get valid token (this will perform login if needed)
	tokenCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()
	
	c.logger.Info("Getting valid token for feasibility check")
	token, err := c.tokenManager.GetToken(tokenCtx)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to get valid token")
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"token_length": len(token),
	}).Debug("Successfully obtained token for feasibility check")

	jsonData, err := json.Marshal(req)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to marshal feasibility request")
		return nil, fmt.Errorf("failed to marshal feasibility request: %w", err)
	}

	url := c.config.BaseURL + c.config.CheckServiceURL
	c.logger.WithFields(logrus.Fields{
		"url":           url,
		"order_type":    req.OrderType,
		"request_size":  len(jsonData),
		"content_type":  "application/json",
	}).Info("Making Delcaper feasibility request")

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"url":   url,
			"error": err.Error(),
		}).Error("Failed to create feasibility request")
		return nil, fmt.Errorf("failed to create feasibility request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	// Use a timeout context for the HTTP request
	requestCtx, requestCancel := context.WithTimeout(ctx, c.config.Timeout)
	defer requestCancel()
	httpReq = httpReq.WithContext(requestCtx)

	startTime := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	responseTime := time.Since(startTime)
	
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"url":           url,
			"response_time": responseTime,
			"error":         err.Error(),
		}).Error("Failed to execute feasibility request")
		return nil, fmt.Errorf("failed to execute feasibility request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"status_code":   resp.StatusCode,
			"response_time": responseTime,
			"error":         err.Error(),
		}).Error("Failed to read feasibility response body")
		return nil, fmt.Errorf("failed to read feasibility response body: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"status_code":   resp.StatusCode,
		"response_size": len(body),
		"response_time": responseTime,
	}).Info("Received feasibility response")

	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
			"url":        url,
		}).Error("Feasibility check failed with non-OK status")
		return nil, fmt.Errorf("feasibility check failed with status %d: %s", resp.StatusCode, string(body))
	}

	var feasibleResp CheckFeasibleResponse
	if err := json.Unmarshal(body, &feasibleResp); err != nil {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
			"error":       err.Error(),
		}).Error("Failed to unmarshal feasibility response")
		return nil, fmt.Errorf("failed to unmarshal feasibility response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"feasible":       feasibleResp.Data.Feasible,
		"distance":       feasibleResp.Data.Distance,
		"services_count": len(feasibleResp.Data.Services),
		"response_time":  responseTime,
		"status":         "success",
	}).Info("Delcaper feasibility check completed")

	return &feasibleResp, nil
}

// IsHealthy checks if the client can successfully authenticate
func (c *DelcaperClient) IsHealthy(ctx context.Context) bool {
	// For now, just check if the configuration is valid
	// The actual authentication will happen when making API calls
	isHealthy := c.config.Enabled && c.config.BaseURL != "" && c.config.Email != "" && c.config.Password != ""
	
	c.logger.WithFields(logrus.Fields{
		"enabled":   c.config.Enabled,
		"base_url":  c.config.BaseURL != "",
		"email":     c.config.Email != "",
		"password":  c.config.Password != "",
		"healthy":   isHealthy,
		"timeout":   c.config.Timeout,
	}).Info("Delcaper client health check")
	
	return isHealthy
} 