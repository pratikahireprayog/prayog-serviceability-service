package routes

import (
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterHealthRoutes registers health check endpoints
func RegisterHealthRoutes(router fiber.Router, healthHandler *handlers.HealthHandler) {
	// Basic health endpoints
	router.Get("/ping", healthHandler.Ping)
	router.Get("/health", healthHandler.Health)

	// Kubernetes probes
	router.Get("/ready", healthHandler.Ready)
	router.Get("/live", healthHandler.Live)
}
