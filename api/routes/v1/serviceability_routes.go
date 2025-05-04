package v1

import (
	v1 "prayog-serviceability-service/api/handlers/v1"

	"github.com/gofiber/fiber/v2"
)

// RegisterServiceabilityRoutes registers serviceability routes with the provided router.
func RegisterServiceabilityRoutes(router fiber.Router, handler *v1.ServiceabilityHandler) {
	serviceability := router.Group("/serviceability")

	// Register routes
	serviceability.Get("/check/:postalCode", handler.CheckServiceability)
	serviceability.Post("/bulk-check", handler.BulkCheckServiceability)
}
