package middleware

import (
	"net/http"
)

// AuthConfig holds configuration for the authentication middleware.
type AuthConfig struct {
	// API keys for authorization
	APIKeys []string
}

// Auth creates a middleware that checks for a valid API key.
func Auth(apiKeys []string) func(http.Handler) http.Handler {
	authConfig := AuthConfig{
		APIKeys: apiKeys,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for some endpoints
			if r.URL.Path == "/health" || r.URL.Path == "/version" {
				next.ServeHTTP(w, r)
				return
			}

			// Check for API key in header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("API key is required"))
				return
			}

			// Check if API key is valid
			valid := false
			for _, key := range authConfig.APIKeys {
				if apiKey == key {
					valid = true
					break
				}
			}

			if !valid {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid API key"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
