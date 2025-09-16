package routes

import (
    "github.com/gofiber/fiber/v2"

    "prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
    "prayog-serviceability-service/internal/shared/config"
)

// RegisterUniUniRoutes registers UniUni specific routes under /uniuni
func RegisterUniUniRoutes(router fiber.Router, cfg config.UniUniConfig) {
    uniuniHandler := handlers.NewUniUniHandler(cfg)
    group := router.Group("/uniuni")
    group.Post("/check-postal-code", uniuniHandler.CheckPostalCode)
}


