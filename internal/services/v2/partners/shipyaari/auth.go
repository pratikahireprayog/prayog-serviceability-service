package shipyaari

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
)

// Authenticator implements common.Authenticator for Shipyaari JWT authentication
type Authenticator struct {
	config       config.ShipyaariConfig
	httpClient   *http.Client
	token        string
	tokenExpiry  time.Time
	refreshToken string
	mu           sync.RWMutex
}

// NewAuthenticator creates a new Shipyaari authenticator
func NewAuthenticator(cfg config.ShipyaariConfig) common.Authenticator {
	return &Authenticator{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Authenticate authenticates with Shipyaari and obtains a JWT token
func (a *Authenticator) Authenticate(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Prepare authentication request
	authReq := AuthRequest{
		Email:    a.config.Email,
		Password: a.config.Password,
	}

	reqBody, err := json.Marshal(authReq)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.config.TokenURL, bytes.NewBuffer(reqBody))
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
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	// Check if authentication was successful
	if !authResp.Success {
		return fmt.Errorf("authentication failed: %s", authResp.Message)
	}

	// Check if we have data
	if len(authResp.Data) == 0 {
		return fmt.Errorf("no authentication data received")
	}

	// Get the first auth data (usually there's only one)
	authData := authResp.Data[0]

	// Store token information (use JWT token)
	a.token = authData.JWT
	// Parse JWT to get expiry (simplified - you might want to use a JWT library)
	// For now, set expiry to 24 hours from now
	a.tokenExpiry = time.Now().Add(24 * time.Hour)
	a.refreshToken = authData.Token // Use the regular token as refresh token

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

	// Check if token is expired (with 5-minute buffer)
	return time.Now().Add(5 * time.Minute).Before(a.tokenExpiry)
}

// RefreshAuth refreshes the authentication token
func (a *Authenticator) RefreshAuth(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.refreshToken == "" {
		// If no refresh token, do full authentication
		a.mu.Unlock() // Unlock temporarily to avoid deadlock
		return a.Authenticate(ctx)
	}

	// Implement refresh token logic here if Shipyaari supports it
	// For now, just do full authentication
	a.mu.Unlock() // Unlock temporarily to avoid deadlock
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
