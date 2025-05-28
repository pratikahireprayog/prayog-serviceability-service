package v1

import (
	v1 "prayog-serviceability-service/api/handlers/v1"

	"github.com/gofiber/fiber/v2"
)

// RegisterServiceabilityRoutes registers serviceability routes with the provided router.
func RegisterServiceabilityRoutes(router fiber.Router, handler *v1.ServiceabilityHandler) {
	// Register routes directly on the provided router (which is already /serviceability/api/v1)
	router.Get("/check/:postalCode", handler.CheckServiceability)
	router.Post("/bulk-check", handler.BulkCheckServiceability)
}
