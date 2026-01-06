package urbanbolt

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
)

// Authenticator implements common.Authenticator for UrbanBolt JWT authentication
type Authenticator struct {
	config      config.UrbanBoltConfig
	httpClient  *http.Client
	token       string
	tokenExpiry time.Time
	mu          sync.RWMutex
}

// NewAuthenticator creates a new UrbanBolt authenticator
func NewAuthenticator(cfg config.UrbanBoltConfig) common.Authenticator {
	// Create HTTP client with SSL verification disabled for UAT environment
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Bypass SSL certificate verification for UAT
		},
	}

	return &Authenticator{
		config: cfg,
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: tr,
		},
	}
}

// Authenticate authenticates with UrbanBolt and obtains an access token
func (a *Authenticator) Authenticate(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if we have a valid token that hasn't expired
	if a.token != "" && time.Now().Add(a.config.TokenExpiryBuffer).Before(a.tokenExpiry) {
		return nil // Token is still valid
	}

	// Prepare authentication request
	authReq := AuthRequest{
		Username: a.config.Username,
		Password: a.config.Password,
	}

	reqBody, err := json.Marshal(authReq)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	// Create HTTP request with full URL
	authURL := a.config.BaseURL + a.config.AuthTokenPath
	
	// Log authentication attempt (without sensitive data)
	fmt.Printf("[urbanbolt] Attempting authentication: URL=%s\n", authURL)
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Make the request
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Provide more helpful error message for 404
		if resp.StatusCode == 404 {
			return fmt.Errorf("authentication endpoint not found (404): %s. Please check URBANBOLT_AUTH_TOKEN_PATH environment variable. Current path: %s", authURL, a.config.AuthTokenPath)
		}
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	// Check status if present (optional, but good for debugging)
	if authResp.Status != "" && authResp.Status != "Success" {
		return fmt.Errorf("authentication failed with status: %s", authResp.Status)
	}

	// Validate response
	if authResp.AccessToken == "" {
		return fmt.Errorf("invalid authentication response: access_token is empty")
	}

	// Store token information
	a.token = authResp.AccessToken
	
	// Calculate token expiry (subtract buffer time for safety)
	expiresIn := authResp.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 3600 // Default to 1 hour if not provided
	}
	// Subtract 5 minutes (300 seconds) for safety
	if expiresIn > 300 {
		expiresIn -= 300
	}
	a.tokenExpiry = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return nil
}

// GetAuthHeaders returns the headers needed for authentication
func (a *Authenticator) GetAuthHeaders() map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	headers := make(map[string]string)
	if a.token != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", a.token)
	}
	return headers
}

// IsAuthenticated checks if we have a valid token
func (a *Authenticator) IsAuthenticated() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.token == "" {
		return false
	}

	// Check if token is expired (with buffer)
	return time.Now().Add(a.config.TokenExpiryBuffer).Before(a.tokenExpiry)
}

// RefreshAuth refreshes the authentication token
func (a *Authenticator) RefreshAuth(ctx context.Context) error {
	// For UrbanBolt, we just re-authenticate
	return a.Authenticate(ctx)
}

// GetToken returns the current token
func (a *Authenticator) GetToken() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.token
}

// IsTokenExpired checks if the current token is expired
func (a *Authenticator) IsTokenExpired() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.token == "" {
		return true
	}

	return time.Now().After(a.tokenExpiry)
}

