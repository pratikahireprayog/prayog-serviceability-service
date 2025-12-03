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
	httpClient   *http.Client
	config       config.IndiaPostConfig
	logger       *logrus.Logger
	accessToken  string
	refreshToken string
	tokenExpiry  time.Time
	refreshExpiry time.Time
	tokenMutex   sync.RWMutex
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
	c.refreshToken = loginResp.Data.RefreshToken
	c.tokenExpiry = time.Now().Add(time.Duration(loginResp.Data.ExpiresIn) * time.Second)
	if loginResp.Data.RefreshExpiresIn > 0 {
		c.refreshExpiry = time.Now().Add(time.Duration(loginResp.Data.RefreshExpiresIn) * time.Second)
	}

	c.logger.WithFields(logrus.Fields{
		"partner":        "IndiaPostInternational",
		"action":         "authenticate",
		"token_expiry":   c.tokenExpiry,
		"refresh_expiry": c.refreshExpiry,
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
		// Need to acquire write lock to update token
		c.tokenMutex.Lock()
		// Double-check token is still invalid (another goroutine might have refreshed it)
		if c.accessToken == "" || time.Now().After(c.tokenExpiry.Add(-c.config.TokenExpiryBuffer)) {
			// Try to refresh first if we have a refresh token
			if c.refreshToken != "" && (c.refreshExpiry.IsZero() || time.Now().Before(c.refreshExpiry)) {
				// Call internal refresh method (without locking, since we already have the lock)
				refreshErr := c.refreshTokenUnsafe(ctx)
				if refreshErr == nil {
					token = c.accessToken
					c.tokenMutex.Unlock()
					return token, nil
				}
				// If refresh failed, fall through to full authentication
			}
			
			// Fall back to full authentication (unlock first since Authenticate will lock)
			c.tokenMutex.Unlock()
			if err := c.Authenticate(ctx); err != nil {
				return "", err
			}
			c.tokenMutex.RLock()
			token = c.accessToken
			c.tokenMutex.RUnlock()
			return token, nil
		} else {
			// Token was refreshed by another goroutine
			token = c.accessToken
			c.tokenMutex.Unlock()
			return token, nil
		}
	}

	return token, nil
}

// RefreshToken refreshes the access token using the refresh token
// This method handles its own locking and should be called when the mutex is not already held
func (c *IndiaPostClient) RefreshToken(ctx context.Context) error {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	return c.refreshTokenUnsafe(ctx)
}

// refreshTokenUnsafe refreshes the access token without locking
// Caller must hold the write lock (tokenMutex.Lock())
func (c *IndiaPostClient) refreshTokenUnsafe(ctx context.Context) error {
	if c.refreshToken == "" {
		return fmt.Errorf("no refresh token available, need to re-authenticate")
	}

	// Check if refresh token is still valid
	if !c.refreshExpiry.IsZero() && time.Now().After(c.refreshExpiry) {
		c.logger.WithFields(logrus.Fields{
			"partner": "IndiaPostInternational",
			"action":  "refresh_token",
			"status":  "refresh_token_expired",
		}).Warn("Refresh token expired, need to re-authenticate")
		// Clear tokens to force re-authentication
		c.accessToken = ""
		c.refreshToken = ""
		return fmt.Errorf("refresh token expired, need to re-authenticate")
	}

	// Build refresh token URL - according to docs: /beextcustomer/v1/access/TokenWithRtoken
	refreshURL := c.config.BaseURL + "/beextcustomer/v1/access/TokenWithRtoken"
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, refreshURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create refresh token request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.refreshToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c.logger.WithFields(logrus.Fields{
		"partner": "IndiaPostInternational",
		"action":  "refresh_token",
		"url":      refreshURL,
	}).Info("Refreshing India Post access token")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh token request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read refresh token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// If refresh fails, clear tokens to force re-authentication
		c.accessToken = ""
		c.refreshToken = ""
		return &IndiaPostAPIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("India Post token refresh failed with status %d", resp.StatusCode),
			RawBody:    string(respBody),
		}
	}

	// Parse refresh response
	var refreshResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(respBody, &refreshResp); err != nil {
		return fmt.Errorf("failed to parse refresh token response: %w", err)
	}

	// Update access token
	c.accessToken = refreshResp.AccessToken
	if refreshResp.ExpiresIn > 0 {
		c.tokenExpiry = time.Now().Add(time.Duration(refreshResp.ExpiresIn) * time.Second)
	}

	c.logger.WithFields(logrus.Fields{
		"partner":      "IndiaPostInternational",
		"action":       "refresh_token",
		"token_expiry": c.tokenExpiry,
	}).Info("Successfully refreshed India Post access token")

	return nil
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
		c.logger.WithError(err).WithFields(logrus.Fields{
			"partner": "IndiaPostInternational",
			"action":  "calculate_tariff",
			"url":     url,
		}).Error("HTTP request failed")
		return nil, fmt.Errorf("tariff request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"partner":     "IndiaPostInternational",
			"action":      "calculate_tariff",
			"status_code": resp.StatusCode,
		}).Error("Failed to read response body")
		return nil, fmt.Errorf("failed to read tariff response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"partner":     "IndiaPostInternational",
			"action":      "calculate_tariff",
			"status_code": resp.StatusCode,
			"response_body": string(respBody),
		}).Error("India Post tariff API returned non-OK status")
		return nil, &IndiaPostAPIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("India Post tariff API failed with status %d", resp.StatusCode),
			RawBody:    string(respBody),
		}
	}

	// Log raw response for debugging (especially important since API only works on production)
	c.logger.WithFields(logrus.Fields{
		"partner":      "IndiaPostInternational",
		"action":       "calculate_tariff",
		"status_code":  resp.StatusCode,
		"response_body": string(respBody),
	}).Info("Received India Post tariff API response")

	// Parse tariff response - handle both direct format and wrapped format
	var tariffResp TariffResponse
	
	// First try to parse as direct format (from API docs example)
	if err := json.Unmarshal(respBody, &tariffResp); err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"partner":      "IndiaPostInternational",
			"action":       "calculate_tariff",
			"response_body": string(respBody),
		}).Error("Failed to parse tariff response JSON")
		return nil, fmt.Errorf("failed to parse tariff response: %w", err)
	}

	// If response is wrapped in a success/data structure, extract it
	if tariffResp.Status == "" && tariffResp.TariffAmount == 0 && tariffResp.Data != nil {
		// Response might be wrapped, try to extract from data (Data is already map[string]interface{})
		dataMap := tariffResp.Data
		if tariffAmount, ok := dataMap["tariffAmount"].(float64); ok {
			tariffResp.TariffAmount = tariffAmount
		}
		if currency, ok := dataMap["currency"].(string); ok {
			tariffResp.Currency = currency
		}
		if deliveryTime, ok := dataMap["deliveryTime"].(string); ok {
			tariffResp.DeliveryTime = deliveryTime
		}
		if status, ok := dataMap["status"].(string); ok {
			tariffResp.Status = status
		}
	}

	c.logger.WithFields(logrus.Fields{
		"partner":       "IndiaPostInternational",
		"action":        "calculate_tariff",
		"tariff_amount": tariffResp.TariffAmount,
		"currency":      tariffResp.Currency,
		"delivery_time": tariffResp.DeliveryTime,
		"status":        tariffResp.Status,
		"success":       tariffResp.Success,
		"message":       tariffResp.Message,
	}).Info("Parsed India Post tariff response")

	return &tariffResp, nil
}

