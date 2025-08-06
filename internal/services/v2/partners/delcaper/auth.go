package delcaper

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
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
	// For now, we'll return an error indicating token refresh is needed
	// In a real implementation, this would call the login API
	return "", fmt.Errorf("token refresh needed - please login again")
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