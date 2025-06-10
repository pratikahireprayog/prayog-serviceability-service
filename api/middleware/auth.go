package middleware

import (
	"errors"
	"net/http"
	"strings"

	"prayog-serviceability-service/internal/shared/constants/v1"
)

// AuthConfig holds configuration for the authentication middleware.
type AuthConfig struct {
	// API keys for authorization
	APIKeys []string
	// Skip authentication for these paths
	SkipPaths []string
}

// Auth creates a middleware that checks for a valid API key in Authorization header.
// Supports both "Bearer <token>" and "ApiKey <token>" formats.
func Auth(apiKeys []string) Middleware {
	authConfig := AuthConfig{
		APIKeys: apiKeys,
		SkipPaths: []string{
			"/health",
			"/ready",
			"/version",
		},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for configured paths
			path := r.URL.Path
			for _, skipPath := range authConfig.SkipPaths {
				if path == skipPath {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract API key from Authorization header
			apiKey, err := extractAPIKey(r)
			if err != nil {
				RespondWithError(w, constants.StatusUnauthorized, constants.MsgMissingCredentials, constants.ErrorCodeMissingCredentials)
				return
			}

			// Validate API key
			if !isValidAPIKey(apiKey, authConfig.APIKeys) {
				RespondWithError(w, constants.StatusUnauthorized, constants.MsgInvalidToken, constants.ErrorCodeInvalidToken)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// extractAPIKey extracts the API key from the Authorization header.
// Supports both "Bearer <token>" and "ApiKey <token>" formats.
func extractAPIKey(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	// Split the header into scheme and token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", errors.New("invalid authorization header format")
	}

	scheme := strings.ToLower(parts[0])
	token := parts[1]

	// Support both Bearer and ApiKey schemes
	if scheme != "bearer" && scheme != "apikey" {
		return "", errors.New("unsupported authorization scheme")
	}

	if token == "" {
		return "", errors.New("empty token")
	}

	return token, nil
}

// isValidAPIKey checks if the provided API key is valid.
func isValidAPIKey(apiKey string, validKeys []string) bool {
	for _, key := range validKeys {
		if apiKey == key {
			return true
		}
	}
	return false
}
