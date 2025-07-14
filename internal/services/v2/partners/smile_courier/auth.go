package smile_courier

import (
	"context"

	"prayog-serviceability-service/internal/services/v2/partners/common"
)

// NoAuthenticator implements common.Authenticator for partners that don't require authentication
type NoAuthenticator struct{}

// NewNoAuthenticator creates a new no-authentication authenticator
func NewNoAuthenticator() common.Authenticator {
	return &NoAuthenticator{}
}

// Authenticate is a no-op for partners that don't require authentication
func (n *NoAuthenticator) Authenticate(ctx context.Context) error {
	// No authentication required for Smile Courier
	return nil
}

// GetAuthHeaders returns empty headers since no authentication is needed
func (n *NoAuthenticator) GetAuthHeaders() map[string]string {
	// No authentication headers needed
	return make(map[string]string)
}

// IsAuthenticated always returns true since no authentication is required
func (n *NoAuthenticator) IsAuthenticated() bool {
	// Always consider "authenticated" since no auth is needed
	return true
}

// RefreshAuth is a no-op for partners that don't require authentication
func (n *NoAuthenticator) RefreshAuth(ctx context.Context) error {
	// No authentication refresh needed
	return nil
}

// GetToken returns empty string since no token is used
func (n *NoAuthenticator) GetToken() string {
	// No token required
	return ""
}

// IsTokenExpired always returns false since no token is used
func (n *NoAuthenticator) IsTokenExpired() bool {
	// No token to expire
	return false
}
