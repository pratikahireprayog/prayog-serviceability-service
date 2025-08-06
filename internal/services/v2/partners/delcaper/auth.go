package delcaper

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"prayog-serviceability-service/internal/shared/config"
)

// TokenManager manages JWT tokens for Delcaper API
type TokenManager struct {
	config      config.DelcaperConfig
	httpClient  *http.Client
	accessToken string
	refreshToken string
	expiresAt   time.Time
}

// NewTokenManager creates a new token manager for Delcaper API
func NewTokenManager(cfg config.DelcaperConfig, httpClient *http.Client) *TokenManager {
	return &TokenManager{
		config:     cfg,
		httpClient: httpClient,
	}
}

// GetToken retrieves a valid token, refreshing if necessary
func (tm *TokenManager) GetToken(ctx context.Context) (string, error) {
	// Check if current token is still valid
	if tm.IsTokenValid() {
		return tm.accessToken, nil
	}

	// Token is expired or doesn't exist, need to login
	return tm.RefreshToken(ctx)
}

// RefreshToken forcefully refreshes the token by logging in again
func (tm *TokenManager) RefreshToken(ctx context.Context) (string, error) {
	client := NewDelcaperClient(tm.config, tm.httpClient)
	
	loginResp, err := client.Login(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	// Parse expiry time from expiresIn string (e.g., "1d")
	expiryDuration, err := parseExpiryDuration(loginResp.Data.ExpiresIn)
	if err != nil {
		return "", fmt.Errorf("failed to parse expiry duration: %w", err)
	}

	// Set expiry time with buffer
	tm.expiresAt = time.Now().Add(expiryDuration).Add(-tm.config.TokenExpiryBuffer)
	tm.accessToken = loginResp.Data.AccessToken
	tm.refreshToken = loginResp.Data.RefreshToken

	return tm.accessToken, nil
}

// IsTokenValid checks if the current token is valid
func (tm *TokenManager) IsTokenValid() bool {
	return tm.accessToken != "" && time.Now().Before(tm.expiresAt)
}

// ClearToken clears the stored token
func (tm *TokenManager) ClearToken() {
	tm.accessToken = ""
	tm.refreshToken = ""
	tm.expiresAt = time.Time{}
}

// SetTokens sets the tokens manually (used by client after login)
func (tm *TokenManager) SetTokens(accessToken, refreshToken, expiresIn string) {
	expiryDuration, err := parseExpiryDuration(expiresIn)
	if err != nil {
		// If parsing fails, set a default expiry
		expiryDuration = 24 * time.Hour
	}

	tm.accessToken = accessToken
	tm.refreshToken = refreshToken
	tm.expiresAt = time.Now().Add(expiryDuration).Add(-tm.config.TokenExpiryBuffer)
}

// parseExpiryDuration parses expiry duration string (e.g., "1d", "30d")
func parseExpiryDuration(expiresIn string) (time.Duration, error) {
	expiresIn = strings.TrimSpace(expiresIn)
	
	// Handle common formats
	switch expiresIn {
	case "1d":
		return 24 * time.Hour, nil
	case "30d":
		return 30 * 24 * time.Hour, nil
	case "7d":
		return 7 * 24 * time.Hour, nil
	case "1h":
		return time.Hour, nil
	case "1m":
		return time.Minute, nil
	}

	// Try to parse as number of seconds
	if seconds, err := strconv.Atoi(expiresIn); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}

	// Try to parse as number of minutes
	if strings.HasSuffix(expiresIn, "m") {
		if minutes, err := strconv.Atoi(strings.TrimSuffix(expiresIn, "m")); err == nil {
			return time.Duration(minutes) * time.Minute, nil
		}
	}

	// Try to parse as number of hours
	if strings.HasSuffix(expiresIn, "h") {
		if hours, err := strconv.Atoi(strings.TrimSuffix(expiresIn, "h")); err == nil {
			return time.Duration(hours) * time.Hour, nil
		}
	}

	// Try to parse as number of days
	if strings.HasSuffix(expiresIn, "d") {
		if days, err := strconv.Atoi(strings.TrimSuffix(expiresIn, "d")); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}

	return 0, fmt.Errorf("unable to parse expiry duration: %s", expiresIn)
} 