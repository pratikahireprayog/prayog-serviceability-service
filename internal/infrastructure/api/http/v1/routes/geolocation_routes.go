package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
)

// RegisterGeolocationRoutes registers all geo-location related routes
func RegisterGeolocationRoutes(router fiber.Router, handler *handlers.GeolocationHandler, logger *logrus.Logger) {
	logger.Info("🗺️  Registering geo-location routes...")

	// Create geo-location routes group
	geoGroup := router.Group("/geo-locations")
	{
		// Health check endpoint
		geoGroup.Get("/health", handler.GetGeolocationHealth)

		// Postal code specific routes
		postalCodeGroup := geoGroup.Group("/postal-code")
		{
			// GET /geo-locations/postal-code/{postal_code} - Get country code for postal code
			postalCodeGroup.Get("/:postal_code", handler.GetCountryCodeByPostalCode)

			// GET /geo-locations/postal-code/{postal_code}/hierarchy - Get location hierarchy for postal code
			postalCodeGroup.Get("/:postal_code/hierarchy", handler.GetLocationHierarchy)

			// GET /geo-locations/postal-code/{postal_code}/validate-international - Validate if postal code is international
			postalCodeGroup.Get("/:postal_code/validate-international", handler.ValidateInternationalRequest)
		}
	}

	logger.Info("🗺️  Geo-location routes registered successfully")
}
