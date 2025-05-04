package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// AuthConfig holds configuration for the authentication middleware.
type AuthConfig struct {
	// API keys for authorization
	APIKeys []string
}

// Auth creates a middleware that checks for a valid API key.
func Auth(apiKeys []string) fiber.Handler {
	authConfig := AuthConfig{
		APIKeys: apiKeys,
	}

	return func(c *fiber.Ctx) error {
		// Skip auth for some endpoints
		path := c.Path()
		if path == "/health" || path == "/version" {
			return c.Next()
		}

		// Check for API key in header
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).SendString("API key is required")
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
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid API key")
		}

		return c.Next()
	}
}
