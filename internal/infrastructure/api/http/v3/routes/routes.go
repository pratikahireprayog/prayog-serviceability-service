package routes

import (
	"github.com/gofiber/fiber/v2"

	"prayog-serviceability-service/internal/infrastructure/api/http/v3/handlers"
)

// RegisterServiceabilityV3Routes registers V3 serviceability routes
func RegisterServiceabilityV3Routes(router fiber.Router, handler *handlers.ServiceabilityHandler) {
	// POST /serviceability/v3/check
	router.Post("/check", handler.CheckServiceability)
}
