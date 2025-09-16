package routes

import (
	"github.com/gofiber/fiber/v2"

	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
)

// RegisterServiceabilityV2Routes registers all V2 serviceability routes
func RegisterServiceabilityV2Routes(router fiber.Router, handler *handlers.ServiceabilityV2Handler) {
	// POST /check - Single V2 serviceability check
	router.Post("/check", handler.CheckServiceabilityV2)

	// POST /bulk-check - Bulk V2 serviceability check
	router.Post("/bulk-check", handler.BulkCheckServiceabilityV2)

	// GET /health - V2 health check
	router.Get("/health", handler.GetHealthV2)

	// Public simplified endpoint group
	publicGroup := router.Group("/public")
	publicGroup.Post("/check", handler.CheckServiceabilityPublic)
}
