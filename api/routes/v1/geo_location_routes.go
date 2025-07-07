package v1

import (
	handlers "prayog-serviceability-service/api/handlers/v1"

	"github.com/gofiber/fiber/v2"
)

// RegisterGeoLocationRoutes registers geo location routes
func RegisterGeoLocationRoutes(router fiber.Router, handler *handlers.GeoLocationHandler) {
	geoLocations := router.Group("/geo-locations")

	// Basic CRUD operations - Only 5 APIs as requested
	geoLocations.Post("/", handler.CreateGeoLocation)                                // 1. Create API
	geoLocations.Get("/", handler.GetAllGeoLocations)                                // 2. Get list API with filters
	geoLocations.Get("/postal-code/:postalCode", handler.GetGeoLocationByPostalCode) // 3. Get by Postal code API
	geoLocations.Put("/:postal_code", handler.UpdateGeoLocation)                     // 4. Update API
	geoLocations.Delete("/:postal_code", handler.DeleteGeoLocation)                  // 5. Delete API
}
