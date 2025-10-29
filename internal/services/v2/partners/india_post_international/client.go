package india_post_international

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"prayog-serviceability-service/internal/shared/config"

	"github.com/sirupsen/logrus"
)

// IndiaPostClient handles HTTP communication with India Post API
type IndiaPostClient struct {
	httpClient  *http.Client
	config      config.IndiaPostConfig
	logger      *logrus.Logger
	accessToken string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
}

// NewIndiaPostClient creates a new India Post HTTP client
func NewIndiaPostClient(config config.IndiaPostConfig) *IndiaPostClient {
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
		"partner":  "IndiaPostInternational",
		"base_url": config.BaseURL,
		"username": redactCredentialField(config.Username),
		"timeout":  config.Timeout,
	}).Info("Creating India Post International client")

	return &IndiaPostClient{
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

// Authenticate authenticates with India Post API and stores the access token
func (c *IndiaPostClient) Authenticate(ctx context.Context) error {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// Check if token is still valid
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		c.logger.WithFields(logrus.Fields{
			"partner": "IndiaPostInternational",
			"action":  "authenticate",
			"status":  "token_valid",
		}).Info("Using existing valid token")
		return nil
	}

	// Login to get new access token
	loginReq := LoginRequest{
		Username: c.config.Username,
		Password: c.config.Password,
	}

	url := c.config.BaseURL + c.config.LoginURL
	bodyBytes, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.WithFields(logrus.Fields{
		"partner": "IndiaPostInternational",
		"action":  "authenticate",
		"url":     url,
	}).Info("Authenticating with India Post")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &IndiaPostAPIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("India Post login failed with status %d", resp.StatusCode),
			RawBody:    string(respBody),
		}
	}

	// Parse login response - India Post returns nested structure
	var loginResp struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken      string `json:"access_token"`
			ExpiresIn        int    `json:"expires_in"`
			RefreshExpiresIn int    `json:"refresh_expires_in"`
			RefreshToken     string `json:"refresh_token"`
			TokenType        string `json:"token_type"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	// Verify login was successful
	if !loginResp.Success {
		return fmt.Errorf("login failed: success=false in response")
	}

	// Store token and expiry
	c.accessToken = loginResp.Data.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(loginResp.Data.ExpiresIn) * time.Second)

	c.logger.WithFields(logrus.Fields{
		"partner":      "IndiaPostInternational",
		"action":       "authenticate",
		"token_expiry": c.tokenExpiry,
	}).Info("Successfully authenticated with India Post")

	return nil
}

// GetAccessToken returns the current access token, authenticating if necessary
func (c *IndiaPostClient) GetAccessToken(ctx context.Context) (string, error) {
	c.tokenMutex.RLock()
	token := c.accessToken
	expiry := c.tokenExpiry
	c.tokenMutex.RUnlock()

	// Check if token needs refresh
	if token == "" || time.Now().After(expiry.Add(-c.config.TokenExpiryBuffer)) {
		if err := c.Authenticate(ctx); err != nil {
			return "", err
		}
		c.tokenMutex.RLock()
		token = c.accessToken
		c.tokenMutex.RUnlock()
	}

	return token, nil
}

// CalculateTariff calculates the tariff for international shipping
func (c *IndiaPostClient) CalculateTariff(ctx context.Context, request TariffRequest) (*TariffResponse, error) {
	// Get valid access token
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	url := c.config.BaseURL + c.config.TariffURL
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tariff request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create tariff request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	c.logger.WithFields(logrus.Fields{
		"partner": "IndiaPostInternational",
		"action":  "calculate_tariff",
		"url":     url,
		"request": request,
	}).Info("Calling India Post tariff API")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tariff request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tariff response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &IndiaPostAPIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("India Post tariff API failed with status %d", resp.StatusCode),
			RawBody:    string(respBody),
		}
	}

	// Parse tariff response
	var tariffResp TariffResponse
	if err := json.Unmarshal(respBody, &tariffResp); err != nil {
		return nil, fmt.Errorf("failed to parse tariff response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"partner":  "IndiaPostInternational",
		"action":   "calculate_tariff",
		"success":  tariffResp.Success,
		"message":  tariffResp.Message,
		"response": tariffResp,
	}).Info("Received India Post tariff response")

	return &tariffResp, nil
}

