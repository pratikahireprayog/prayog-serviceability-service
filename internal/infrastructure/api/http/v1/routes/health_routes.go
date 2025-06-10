package routes

import (
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterHealthRoutes registers health check endpoints
func RegisterHealthRoutes(router fiber.Router, healthHandler *handlers.HealthHandler) {
	// Primary health endpoint under serviceability prefix
	router.Get("/ping", healthHandler.Ping)
}
