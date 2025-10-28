package india_post_domestic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"

	"github.com/sirupsen/logrus"
)

// Authenticator handles authentication for India Post Domestic API
type Authenticator struct {
	config     config.IndiaPostDomesticConfig
	httpClient *http.Client
	tokenInfo  *TokenInfo
	mutex      sync.RWMutex
	logger     *logrus.Logger
}

// NewAuthenticator creates a new India Post Domestic authenticator
func NewAuthenticator(cfg config.IndiaPostDomesticConfig) *Authenticator {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Authenticator{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

// GetAuthHeaders returns the authentication headers
func (a *Authenticator) GetAuthHeaders() map[string]string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if a.tokenInfo == nil || a.tokenInfo.AccessToken == "" {
		a.logger.Warn("No access token available")
		return make(map[string]string)
	}

	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", a.tokenInfo.AccessToken),
	}
}

// IsAuthenticated checks if we have a valid token
func (a *Authenticator) IsAuthenticated() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if a.tokenInfo == nil {
		return false
	}

	return !a.tokenInfo.IsExpired()
}

// Authenticate performs authentication with India Post Domestic API
func (a *Authenticator) Authenticate(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.logger.WithFields(logrus.Fields{
		"partner": "IndiaPostDomestic",
		"action":  "authenticate",
	}).Info("Starting authentication")

	// Check if we can refresh token instead of full login
	if a.tokenInfo != nil && a.tokenInfo.RefreshToken != "" {
		refreshErr := a.refreshTokenInternal(ctx)
		if refreshErr == nil {
			a.logger.Info("Successfully refreshed India Post Domestic token")
			return nil
		}
		a.logger.WithError(refreshErr).Warn("Token refresh failed, attempting full login")
	}

	// Perform full login
	return a.loginInternal(ctx)
}

// loginInternal performs the login operation (must be called with mutex locked)
func (a *Authenticator) loginInternal(ctx context.Context) error {
	// Create JSON request body for India Post Domestic login API
	loginReq := map[string]string{
		"username": a.config.Username,
		"password": a.config.Password,
	}

	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Create request
	loginURL := a.config.AuthBaseURL + a.config.LoginURL
	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	a.logger.WithFields(logrus.Fields{
		"partner":  "IndiaPostDomestic",
		"url":      loginURL,
		"username": redactForLog(a.config.Username),
	}).Info("Sending authentication request")

	// Make request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		a.logger.WithFields(logrus.Fields{
			"partner":     "IndiaPostDomestic",
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Authentication failed")
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - India Post Domestic wraps data in {"success": true, "data": {...}}
	var apiResp struct {
		Success bool         `json:"success"`
		Data    AuthResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("authentication failed: success=false")
	}

	// Validate response
	if apiResp.Data.AccessToken == "" {
		return fmt.Errorf("no access token in response")
	}

	// Store token info
	expiresAt := time.Now().Add(time.Duration(apiResp.Data.ExpiresIn) * time.Second)
	a.tokenInfo = &TokenInfo{
		AccessToken:  apiResp.Data.AccessToken,
		RefreshToken: apiResp.Data.RefreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    apiResp.Data.TokenType,
	}

	a.logger.WithFields(logrus.Fields{
		"partner":    "IndiaPostDomestic",
		"expires_at": expiresAt.Format(time.RFC3339),
		"expires_in": apiResp.Data.ExpiresIn,
	}).Info("Successfully authenticated with India Post Domestic")

	return nil
}

// refreshTokenInternal refreshes the access token (must be called with mutex locked)
func (a *Authenticator) refreshTokenInternal(ctx context.Context) error {
	if a.tokenInfo == nil || a.tokenInfo.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// Create request - India Post Domestic refresh API doesn't need a body
	refreshURL := a.config.AuthBaseURL + a.config.RefreshTokenURL
	req, err := http.NewRequestWithContext(ctx, "POST", refreshURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %w", err)
	}

	// India Post Domestic passes refresh token as Bearer in Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.tokenInfo.RefreshToken))

	a.logger.WithFields(logrus.Fields{
		"partner": "IndiaPostDomestic",
		"url":     refreshURL,
	}).Debug("Sending token refresh request")

	// Make request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		a.logger.WithFields(logrus.Fields{
			"partner":     "IndiaPostDomestic",
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Warn("Token refresh failed")
		return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - India Post Domestic wraps data in {"success": true, "data": {...}}
	var apiResp struct {
		Success bool         `json:"success"`
		Data    AuthResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("token refresh failed: success=false")
	}

	// Validate response
	if apiResp.Data.AccessToken == "" {
		return fmt.Errorf("no access token in refresh response")
	}

	// Update token info
	expiresAt := time.Now().Add(time.Duration(apiResp.Data.ExpiresIn) * time.Second)
	a.tokenInfo.AccessToken = apiResp.Data.AccessToken
	a.tokenInfo.ExpiresAt = expiresAt
	
	// Update refresh token if provided (India Post Domestic provides new refresh token)
	if apiResp.Data.RefreshToken != "" {
		a.tokenInfo.RefreshToken = apiResp.Data.RefreshToken
	}

	a.logger.WithFields(logrus.Fields{
		"partner":    "IndiaPostDomestic",
		"expires_at": expiresAt.Format(time.RFC3339),
	}).Debug("Successfully refreshed India Post Domestic token")

	return nil
}

// RefreshToken refreshes the access token
func (a *Authenticator) RefreshToken(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.refreshTokenInternal(ctx)
}

// RefreshAuth implements common.Authenticator interface (alias for RefreshToken)
func (a *Authenticator) RefreshAuth(ctx context.Context) error {
	return a.RefreshToken(ctx)
}

// GetToken returns the current access token
func (a *Authenticator) GetToken() string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if a.tokenInfo == nil {
		return ""
	}
	return a.tokenInfo.AccessToken
}

// IsTokenExpired checks if the token is expired
func (a *Authenticator) IsTokenExpired() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if a.tokenInfo == nil {
		return true
	}
	return a.tokenInfo.IsExpired()
}

// Implement common.Authenticator interface
var _ common.Authenticator = (*Authenticator)(nil)

// redactForLog redacts sensitive information for logging
func redactForLog(s string) string {
	if s == "" {
		return "[EMPTY]"
	}
	if len(s) <= 3 {
		return "[REDACTED]"
	}
	return s[:3] + "***"
}

