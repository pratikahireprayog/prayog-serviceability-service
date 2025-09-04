package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
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
func Auth(apiKeys []string) fiber.Handler {
	authConfig := AuthConfig{
		APIKeys: apiKeys,
		SkipPaths: []string{
			"/health",
			"/ready",
			"/version",
		},
	}

	return func(c *fiber.Ctx) error {
		// Skip auth for configured paths
		path := c.Path()
		for _, skipPath := range authConfig.SkipPaths {
			if path == skipPath {
				return c.Next()
			}
		}

		// Extract API key from Authorization header
		apiKey, err := extractAPIKey(c)
		if err != nil {
			return c.Status(constants.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    constants.ErrorCodeMissingCredentials,
					"message": constants.MsgMissingCredentials,
				},
			})
		}

		// Validate API key
		if !isValidAPIKey(apiKey, authConfig.APIKeys) {
			return c.Status(constants.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    constants.ErrorCodeInvalidToken,
					"message": constants.MsgInvalidToken,
				},
			})
		}

		// Continue to next handler
		return c.Next()
	}
}

// extractAPIKey extracts the API key from the Authorization header.
// Supports both "Bearer <token>" and "ApiKey <token>" formats.
func extractAPIKey(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")
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
