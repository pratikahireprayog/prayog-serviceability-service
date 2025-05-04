package handlers

import (
	"github.com/gofiber/fiber/v2"

	"prayog-serviceability-service/internal/service/usecase"
)

// RegisterHandlers registers all HTTP handlers with the router.
func RegisterHandlers(r fiber.Router, usecaseFactory usecase.Factory) {
	// Health check
	r.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// Register serviceability handlers
	serviceabilityHandler := NewServiceabilityHandler(usecaseFactory)
	serviceabilityHandler.RegisterRoutes(r)
}
