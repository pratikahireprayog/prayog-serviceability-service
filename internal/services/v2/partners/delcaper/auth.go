package delcaper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"prayog-serviceability-service/internal/shared/config"

	"github.com/sirupsen/logrus"
)

// TokenManager handles JWT token management for Delcaper API
type TokenManager struct {
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	mu           sync.RWMutex
	config       config.DelcaperConfig
	httpClient   *http.Client
	logger       *logrus.Logger
}

// NewTokenManager creates a new token manager
func NewTokenManager(cfg config.DelcaperConfig, httpClient *http.Client) *TokenManager {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &TokenManager{
		config:     cfg,
		httpClient: httpClient,
		logger:     logger,
	}
}

// Login performs login and returns access token
func (tm *TokenManager) Login(ctx context.Context) (string, error) {
	tm.logger.WithFields(logrus.Fields{
		"email":      tm.config.Email,
		"vendor_type": tm.config.VendorType,
		"base_url":   tm.config.BaseURL,
		"login_url":  tm.config.LoginURL,
	}).Info("Starting Delcaper login")
	
	loginReq := LoginRequest{
		Email:      tm.config.Email,
		Password:   tm.config.Password,
		VendorType: tm.config.VendorType,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		tm.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to marshal login request")
		return "", fmt.Errorf("failed to marshal login request: %w", err)
	}

	url := tm.config.BaseURL + tm.config.LoginURL
	tm.logger.WithFields(logrus.Fields{
		"url":           url,
		"request_size":  len(jsonData),
		"content_type":  "application/json",
	}).Info("Making Delcaper login request")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		tm.logger.WithFields(logrus.Fields{
			"url":   url,
			"error": err.Error(),
		}).Error("Failed to create login request")
		return "", fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Use timeout context for the login request
	loginCtx, cancel := context.WithTimeout(ctx, tm.config.Timeout)
	defer cancel()
	req = req.WithContext(loginCtx)

	startTime := time.Now()
	resp, err := tm.httpClient.Do(req)
	responseTime := time.Since(startTime)
	
	if err != nil {
		tm.logger.WithFields(logrus.Fields{
			"url":           url,
			"response_time": responseTime,
			"error":         err.Error(),
		}).Error("Failed to execute login request")
		return "", fmt.Errorf("failed to execute login request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		tm.logger.WithFields(logrus.Fields{
			"status_code":   resp.StatusCode,
			"response_time": responseTime,
			"error":         err.Error(),
		}).Error("Failed to read login response body")
		return "", fmt.Errorf("failed to read login response body: %w", err)
	}

	tm.logger.WithFields(logrus.Fields{
		"status_code":   resp.StatusCode,
		"response_size": len(body),
		"response_time": responseTime,
	}).Info("Received login response")

	if resp.StatusCode != http.StatusOK {
		tm.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
			"url":        url,
		}).Error("Login failed with non-OK status")
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	tm.logger.Info("Starting JSON unmarshalling of login response")
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		tm.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
			"error":       err.Error(),
		}).Error("Failed to unmarshal login response")
		return "", fmt.Errorf("failed to unmarshal login response: %w", err)
	}
	tm.logger.Info("JSON unmarshalling completed successfully")

	tm.logger.WithFields(logrus.Fields{
		"has_access_token": loginResp.Data.AccessToken != "",
		"has_refresh_token": loginResp.Data.RefreshToken != "",
		"expires_in": loginResp.Data.ExpiresIn,
		"user_id": loginResp.Data.UserDto.ID,
	}).Info("Login response parsed, storing tokens")

	// Store tokens
	tm.SetTokens(loginResp.Data.AccessToken, loginResp.Data.RefreshToken, loginResp.Data.ExpiresIn)

	tm.logger.WithFields(logrus.Fields{
		"user_id":       loginResp.Data.UserDto.ID,
		"email":         loginResp.Data.UserDto.Email,
		"token_length":  len(loginResp.Data.AccessToken),
		"response_time": responseTime,
		"status":        "success",
	}).Info("Delcaper login successful")

	return loginResp.Data.AccessToken, nil
}

// GetToken returns a valid access token, performing login if necessary
func (tm *TokenManager) GetToken(ctx context.Context) (string, error) {
	tm.mu.RLock()
	if tm.isTokenValid() {
		token := tm.accessToken
		tm.mu.RUnlock()
		tm.logger.WithFields(logrus.Fields{
			"token_length": len(token),
			"expires_at":   tm.expiresAt,
			"time_until_expiry": time.Until(tm.expiresAt),
		}).Debug("Using existing valid token")
		return token, nil
	}
	tm.mu.RUnlock()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after acquiring write lock
	if tm.isTokenValid() {
		tm.logger.WithFields(logrus.Fields{
			"token_length": len(tm.accessToken),
			"expires_at":   tm.expiresAt,
		}).Debug("Using existing valid token (double-check)")
		return tm.accessToken, nil
	}

	tm.logger.WithFields(logrus.Fields{
		"has_token":     tm.accessToken != "",
		"expires_at":    tm.expiresAt,
		"current_time":  time.Now(),
		"time_until_expiry": time.Until(tm.expiresAt),
	}).Info("Token expired or invalid, performing login")
	return tm.Login(ctx)
}

// isTokenValid checks if the current token is still valid
func (tm *TokenManager) isTokenValid() bool {
	return tm.accessToken != "" && time.Now().Before(tm.expiresAt)
}

// SetTokens sets the tokens and expiry time
func (tm *TokenManager) SetTokens(accessToken, refreshTokenParam, expiresIn string) {
	tm.logger.Info("Starting SetTokens method")
	
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.logger.Info("Acquired lock, setting tokens")
	tm.accessToken = accessToken
	tm.refreshToken = refreshTokenParam

	tm.logger.WithFields(logrus.Fields{
		"access_token_length": len(accessToken),
		"refresh_token_length": len(refreshTokenParam),
		"expires_in": expiresIn,
	}).Info("Tokens set, parsing expiry")

	// Parse expiresIn (assuming it's in seconds)
	if expiresIn != "" {
		if seconds, err := strconv.Atoi(expiresIn); err == nil {
			tm.expiresAt = time.Now().Add(time.Duration(seconds) * time.Second)
			tm.logger.WithFields(logrus.Fields{
				"expires_in_seconds": seconds,
				"expires_at":         tm.expiresAt,
			}).Info("Tokens set successfully")
		} else {
			tm.logger.WithError(err).Warn("Failed to parse expires_in, using default expiry")
		}
	} else {
		tm.logger.Warn("No expires_in provided, token may expire immediately")
	}
	
	tm.logger.Info("SetTokens method completed")
}

// ClearTokens clears stored tokens
func (tm *TokenManager) ClearTokens() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.accessToken = ""
	tm.refreshToken = ""
	tm.expiresAt = time.Time{}
	tm.logger.Info("Tokens cleared")
} 