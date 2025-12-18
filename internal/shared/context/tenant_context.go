package context

import (
	"context"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey contextKey = "tenant_id"
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// PartnerCredentialsKey is the context key for partner credentials map
	// Structure: map[string]map[string]string where first key is partner_code
	PartnerCredentialsKey contextKey = "partner_credentials"
)

// WithTenantID adds tenant_id to the context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID retrieves tenant_id from the context
func GetTenantID(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	return tenantID, ok
}

// WithUserID adds user_id to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID retrieves user_id from the context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// WithPartnerCredentials adds partner credentials to the context
// credentials is a map[string]string containing partner-specific credential fields
func WithPartnerCredentials(ctx context.Context, partnerCode string, credentials map[string]string) context.Context {
	// Get existing credentials map or create new one
	credsMap, ok := ctx.Value(PartnerCredentialsKey).(map[string]map[string]string)
	if !ok || credsMap == nil {
		credsMap = make(map[string]map[string]string)
	}

	// Add or update credentials for this partner
	credsMap[partnerCode] = credentials

	return context.WithValue(ctx, PartnerCredentialsKey, credsMap)
}

// GetPartnerCredentials retrieves partner credentials from the context
// Returns the credentials map for the given partner code, or nil if not found
func GetPartnerCredentials(ctx context.Context, partnerCode string) (map[string]string, bool) {
	credsMap, ok := ctx.Value(PartnerCredentialsKey).(map[string]map[string]string)
	if !ok || credsMap == nil {
		return nil, false
	}

	credentials, exists := credsMap[partnerCode]
	return credentials, exists
}


