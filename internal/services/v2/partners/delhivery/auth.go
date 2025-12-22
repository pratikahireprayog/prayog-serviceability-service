package delhivery

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
)

// Authenticator implements common.Authenticator for Delhivery token authentication
type Authenticator struct {
	config config.DelhiveryConfig
	token  string
}

// NewAuthenticator creates a new Delhivery authenticator
func NewAuthenticator(cfg config.DelhiveryConfig) common.Authenticator {
	return &Authenticator{
		config: cfg,
		token:  cfg.AccessToken,
	}
}

// Authenticate authenticates with Delhivery (token is pre-configured, no API call needed)
func (a *Authenticator) Authenticate(ctx context.Context) error {
	// Delhivery uses a static token, no authentication API call required
	if a.token == "" {
		return fmt.Errorf("Delhivery access token is not configured")
	}
	return nil
}

// GetAuthHeaders returns the headers needed for authentication
func (a *Authenticator) GetAuthHeaders() map[string]string {
	headers := make(map[string]string)
	if a.token != "" {
		headers["Authorization"] = fmt.Sprintf("Token %s", a.token)
	}
	return headers
}

// IsAuthenticated checks if we have a valid token
func (a *Authenticator) IsAuthenticated() bool {
	return a.token != ""
}

// RefreshAuth refreshes the authentication token (not applicable for static tokens)
func (a *Authenticator) RefreshAuth(ctx context.Context) error {
	// For static tokens, just verify token is still configured
	return a.Authenticate(ctx)
}

// GetToken returns the current token
func (a *Authenticator) GetToken() string {
	return a.token
}

// IsTokenExpired checks if the current token is expired (static tokens don't expire)
func (a *Authenticator) IsTokenExpired() bool {
	return false // Static tokens don't expire
}

