package delcaper

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"prayog-serviceability-service/internal/shared/config"
)

// TokenManager handles JWT token management for Delcaper API
type TokenManager struct {
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	mu           sync.RWMutex
	config       interface{} // Using interface{} to avoid circular dependency
	httpClient   *http.Client
}

// NewTokenManager creates a new token manager
func NewTokenManager(config interface{}, httpClient *http.Client) *TokenManager {
	return &TokenManager{
		config:     config,
		httpClient: httpClient,
	}
}

// GetToken returns a valid access token, refreshing if necessary
func (tm *TokenManager) GetToken(ctx context.Context) (string, error) {
	tm.mu.RLock()
	if tm.isTokenValid() {
		token := tm.accessToken
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after acquiring write lock
	if tm.isTokenValid() {
		return tm.accessToken, nil
	}

	return tm.refreshTokenMethod(ctx)
}

// refreshTokenMethod performs login to get new tokens
func (tm *TokenManager) refreshTokenMethod(ctx context.Context) (string, error) {
	// Cast config to DelcaperConfig
	config, ok := tm.config.(config.DelcaperConfig)
	if !ok {
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
	
	loginResp, err := tempClient.Login(loginCtx)
	if err != nil {
		return "", fmt.Errorf("failed to login during token refresh: %w", err)
	}

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
} 