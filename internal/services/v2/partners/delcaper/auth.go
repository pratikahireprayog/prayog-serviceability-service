package delcaper

import (
	"context"
	"fmt"
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
		tm.logger.Debug("Using existing valid token")
		return token, nil
	}
	tm.mu.RUnlock()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after acquiring write lock
	if tm.isTokenValid() {
		tm.logger.Debug("Using existing valid token (double-check)")
		return tm.accessToken, nil
	}

	tm.logger.Info("Token expired or invalid, refreshing token")
	return tm.refreshTokenMethod(ctx)
}

// refreshTokenMethod performs login to get new tokens
func (tm *TokenManager) refreshTokenMethod(ctx context.Context) (string, error) {
	tm.logger.Info("Starting token refresh")
	
	// Cast config to DelcaperConfig
	config, ok := tm.config.(config.DelcaperConfig)
	if !ok {
		tm.logger.Error("Invalid config type for token refresh")
		return "", fmt.Errorf("invalid config type for token refresh")
	}

	// Create a temporary client to perform login with timeout
	tempClient := &DelcaperClient{
		config:      config,
		httpClient:  tm.httpClient,
		tokenManager: tm,
	}

	// Perform login with timeout context
	loginCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()
	
	tm.logger.Info("Performing login for token refresh")
	loginResp, err := tempClient.Login(loginCtx)
	if err != nil {
		tm.logger.WithError(err).Error("Failed to login during token refresh")
		return "", fmt.Errorf("failed to login during token refresh: %w", err)
	}

	tm.logger.Info("Token refresh successful")
	// Return the new access token
	return loginResp.Data.AccessToken, nil
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