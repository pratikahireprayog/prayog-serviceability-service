package middleware

import (
	"errors"
	"net/http"
	"strings"
)

// AuthConfig holds configuration for the auth middleware
type AuthConfig struct {
	APIKeys []string
}

// Auth creates middleware that performs API key authentication
func Auth(config AuthConfig) Middleware {
	// Create a map for faster lookup
	apiKeys := make(map[string]bool)
	for _, key := range config.APIKeys {
		apiKeys[key] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract API key from Authorization header
			apiKey, err := extractAPIKey(r)
			if err != nil {
				http.Error(w, "Unauthorized - Invalid Authorization Header", http.StatusUnauthorized)
				return
			}

			// Check if the API key is valid
			if !apiKeys[apiKey] {
				http.Error(w, "Unauthorized - Invalid API Key", http.StatusUnauthorized)
				return
			}

			// API key is valid, proceed with the request
			next.ServeHTTP(w, r)
		})
	}
}

// extractAPIKey extracts the API key from the Authorization header
// Expected format: "Bearer YOUR_API_KEY" or "ApiKey YOUR_API_KEY"
func extractAPIKey(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return "", errors.New("invalid authorization header format")
	}

	// Accept "Bearer" or "ApiKey" prefix
	scheme := strings.ToLower(parts[0])
	if scheme != "bearer" && scheme != "apikey" {
		return "", errors.New("unsupported authorization scheme")
	}

	return parts[1], nil
}
