package v1

import (
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterServiceabilityV2Routes registers V2 serviceability routes with the provided router.
func RegisterServiceabilityV2Routes(router fiber.Router, handler *handlers.ServiceabilityV2Handler) {
	// Register V2 routes directly on the provided router (which is already /serviceability/api/v2)

	// Main serviceability endpoints
	router.Post("/check", handler.CheckServiceability)          // POST /serviceability/api/v2/check
	router.Post("/bulk-check", handler.BulkCheckServiceability) // POST /serviceability/api/v2/bulk-check

	// Health check endpoint for V2
	router.Get("/health", handler.GetHealthStatus) // GET /serviceability/api/v2/health
}
