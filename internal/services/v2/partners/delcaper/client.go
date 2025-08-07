package delcaper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

	// Create token manager without logger to avoid circular dependency
	tokenManager := NewTokenManager(cfg, httpClient)

	return &DelcaperClient{
		config:      cfg,
		httpClient:  httpClient,
		tokenManager: tokenManager,
		logger:      logger,
	}
}

// Login authenticates with Delcaper API and returns access token
func (c *DelcaperClient) Login(ctx context.Context) (*LoginResponse, error) {
	c.logger.Info("Starting Delcaper login")
	
	loginReq := LoginRequest{
		Email:      c.config.Email,
		Password:   c.config.Password,
		VendorType: c.config.VendorType,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal login request")
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	url := c.config.BaseURL + c.config.LoginURL
	c.logger.WithFields(logrus.Fields{
		"url": url,
	}).Info("Making Delcaper login request")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.WithError(err).Error("Failed to create login request")
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Use timeout context for the login request
	loginCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()
	req = req.WithContext(loginCtx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to execute login request")
		return nil, fmt.Errorf("failed to execute login request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read login response body")
		return nil, fmt.Errorf("failed to read login response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Login failed with non-OK status")
		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		c.logger.WithError(err).Error("Failed to unmarshal login response")
		return nil, fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	// Store tokens in token manager
	c.tokenManager.SetTokens(loginResp.Data.AccessToken, loginResp.Data.RefreshToken, loginResp.Data.ExpiresIn)

	c.logger.WithFields(logrus.Fields{
		"user_id": loginResp.Data.UserDto.ID,
		"email":   loginResp.Data.UserDto.Email,
	}).Info("Delcaper login successful")

	return &loginResp, nil
}

// CheckFeasible checks serviceability for the given request
func (c *DelcaperClient) CheckFeasible(ctx context.Context, req *CheckFeasibleRequest) (*CheckFeasibleResponse, error) {
	c.logger.Info("Starting Delcaper feasibility check")
	
	// Get valid token with timeout
	tokenCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()
	
	c.logger.Info("Getting valid token for feasibility check")
	token, err := c.tokenManager.GetToken(tokenCtx)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get valid token")
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal feasibility request")
		return nil, fmt.Errorf("failed to marshal feasibility request: %w", err)
	}

	url := c.config.BaseURL + c.config.CheckServiceURL
	c.logger.WithFields(logrus.Fields{
		"url": url,
		"order_type": req.OrderType,
	}).Info("Making Delcaper feasibility request")

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.WithError(err).Error("Failed to create feasibility request")
		return nil, fmt.Errorf("failed to create feasibility request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	// Use a timeout context for the HTTP request
	requestCtx, requestCancel := context.WithTimeout(ctx, c.config.Timeout)
	defer requestCancel()
	httpReq = httpReq.WithContext(requestCtx)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.WithError(err).Error("Failed to execute feasibility request")
		return nil, fmt.Errorf("failed to execute feasibility request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read feasibility response body")
		return nil, fmt.Errorf("failed to read feasibility response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Feasibility check failed with non-OK status")
		return nil, fmt.Errorf("feasibility check failed with status %d: %s", resp.StatusCode, string(body))
	}

	var feasibleResp CheckFeasibleResponse
	if err := json.Unmarshal(body, &feasibleResp); err != nil {
		c.logger.WithError(err).Error("Failed to unmarshal feasibility response")
		return nil, fmt.Errorf("failed to unmarshal feasibility response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"feasible": feasibleResp.Data.Feasible,
		"distance": feasibleResp.Data.Distance,
		"services_count": len(feasibleResp.Data.Services),
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
	}).Info("Delcaper client health check")
	
	return isHealthy
} 