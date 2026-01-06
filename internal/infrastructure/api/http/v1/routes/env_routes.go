package routes

import (
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterEnvRoutes registers environment variable routes
func RegisterEnvRoutes(router fiber.Router, envHandler *handlers.EnvHandler) {
	// Debug endpoint for environment variables
	debug := router.Group("/debug")
	{
		debug.Get("/env", envHandler.GetEnvVars) // GET /serviceability/debug/env
	}
}

