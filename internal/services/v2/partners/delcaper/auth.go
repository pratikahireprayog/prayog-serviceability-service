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
	config       interface{} // Using interface{} to avoid circular dependency
	httpClient   *http.Client
	logger       *logrus.Logger
}

// NewTokenManager creates a new token manager
func NewTokenManager(config interface{}, httpClient *http.Client) *TokenManager {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &TokenManager{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}
}

// GetToken returns a valid access token, refreshing if necessary
func (tm *TokenManager) GetToken(ctx context.Context) (string, error) {
	tm.mu.RLock()
	if tm.isTokenValid() {
		token := tm.accessToken
		tm.mu.RUnlock()
		if tm.logger != nil {
			tm.logger.Debug("Using existing valid token")
		}
		return token, nil
	}
	tm.mu.RUnlock()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after acquiring write lock
	if tm.isTokenValid() {
		if tm.logger != nil {
			tm.logger.Debug("Using existing valid token (double-check)")
		}
		return tm.accessToken, nil
	}

	if tm.logger != nil {
		tm.logger.Info("Token expired or invalid, refreshing token")
	}
	return tm.refreshTokenMethod(ctx)
}

// refreshTokenMethod performs login to get new tokens
func (tm *TokenManager) refreshTokenMethod(ctx context.Context) (string, error) {
	if tm.logger != nil {
		tm.logger.Info("Starting token refresh")
	}
	
	// Cast config to DelcaperConfig
	config, ok := tm.config.(config.DelcaperConfig)
	if !ok {
		if tm.logger != nil {
			tm.logger.Error("Invalid config type for token refresh")
		}
		return "", fmt.Errorf("invalid config type for token refresh")
	}

	// Perform login directly without creating a temporary client
	if tm.logger != nil {
		tm.logger.Info("Performing login for token refresh")
	}
	loginResp, err := tm.performLogin(ctx, config)
	if err != nil {
		if tm.logger != nil {
			tm.logger.WithError(err).Error("Failed to login during token refresh")
		}
		return "", fmt.Errorf("failed to login during token refresh: %w", err)
	}

	if tm.logger != nil {
		tm.logger.Info("Token refresh successful")
	}
	// Return the new access token
	return loginResp.Data.AccessToken, nil
}

// performLogin performs the actual login request without creating a circular dependency
func (tm *TokenManager) performLogin(ctx context.Context, config config.DelcaperConfig) (*LoginResponse, error) {
	loginReq := LoginRequest{
		Email:      config.Email,
		Password:   config.Password,
		VendorType: config.VendorType,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	url := config.BaseURL + config.LoginURL
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Use timeout context for the login request
	loginCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()
	req = req.WithContext(loginCtx)

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute login request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read login response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	// Store tokens in token manager
	tm.SetTokens(loginResp.Data.AccessToken, loginResp.Data.RefreshToken, loginResp.Data.ExpiresIn)

	return &loginResp, nil
}

// isTokenValid checks if the current token is still valid
func (tm *TokenManager) isTokenValid() bool {
	return tm.accessToken != "" && time.Now().Before(tm.expiresAt)
}

// SetTokens sets the tokens and expiry time
func (tm *TokenManager) SetTokens(accessToken, refreshTokenParam, expiresIn string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.accessToken = accessToken
	tm.refreshToken = refreshTokenParam

	// Parse expiresIn (assuming it's in seconds)
	if expiresIn != "" {
		if seconds, err := strconv.Atoi(expiresIn); err == nil {
			tm.expiresAt = time.Now().Add(time.Duration(seconds) * time.Second)
			if tm.logger != nil {
				tm.logger.WithFields(logrus.Fields{
					"expires_in_seconds": seconds,
					"expires_at":         tm.expiresAt,
				}).Info("Tokens set successfully")
			}
		} else {
			if tm.logger != nil {
				tm.logger.WithError(err).Warn("Failed to parse expires_in, using default expiry")
			}
		}
	} else {
		if tm.logger != nil {
			tm.logger.Warn("No expires_in provided, token may expire immediately")
		}
	}
}

// ClearTokens clears stored tokens
func (tm *TokenManager) ClearTokens() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.accessToken = ""
	tm.refreshToken = ""
	tm.expiresAt = time.Time{}
	if tm.logger != nil {
		tm.logger.Info("Tokens cleared")
	}
} 