package v1

import (
	v1 "prayog-serviceability-service/api/handlers/v1"

	"github.com/gofiber/fiber/v2"
)

func RegisterNearestHubLocationRoutes(router fiber.Router, handler *v1.NearestHubLocationHandler) {
	group := router.Group("/nearest-hub-locations")
	group.Post("/", handler.Create)
	group.Get("/:postal_code", handler.GetByPostalCode)
	group.Get("/", handler.GetByFilters)
	group.Put("/:postal_code", handler.Update)
	group.Delete("/:postal_code", handler.Delete)
}
