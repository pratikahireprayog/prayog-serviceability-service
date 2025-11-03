package v1

import (
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterServiceabilityV3Routes registers V3 serviceability routes with the provided router.
func RegisterServiceabilityV3Routes(router fiber.Router, handler *handlers.ServiceabilityV3Handler) {
	// Register V3 routes directly on the provided router (which is already /serviceability/api/v3)

	// Main serviceability endpoints
	router.Post("/check", handler.CheckServiceability) // POST /serviceability/api/v3/check

	// Health check endpoint for V3
	router.Get("/health", handler.GetHealthStatus) // GET /serviceability/api/v3/health
}

