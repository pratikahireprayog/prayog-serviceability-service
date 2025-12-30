package routes

import (
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterLogoRoutes registers routes for logo management
func RegisterLogoRoutes(router fiber.Router, handler *handlers.LogoHandler) {
	group := router.Group("/logos")
	
	group.Post("/", handler.UploadLogo)
	group.Put("/", handler.UpdateLogo)
	group.Get("/", handler.ListLogos)
	group.Delete("/:filename", handler.DeleteLogo)
}
