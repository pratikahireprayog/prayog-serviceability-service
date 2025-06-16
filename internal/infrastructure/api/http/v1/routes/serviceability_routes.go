package routes

import (
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// RegisterServiceabilityRoutes registers serviceability check endpoints with middleware
func RegisterServiceabilityRoutes(router fiber.Router, serviceabilityHandler *handlers.ServiceabilityHandler, logger *logrus.Logger) {
	// Create middleware instance
	serviceabilityMiddleware := middleware.NewServiceabilityMiddleware(logger)

	// Apply middleware to all serviceability routes
	router.Use(serviceabilityMiddleware.SecurityHeaders())
	router.Use(serviceabilityMiddleware.ContentTypeValidation())
	router.Use(serviceabilityMiddleware.RequestSizeLimit())
	router.Use(serviceabilityMiddleware.RequestLogging())

	// NEW POSTAL CODE BASED SERVICEABILITY ROUTES

	// Check endpoint group with standard rate limiting
	checkRoute := router.Group("/check")
	checkRoute.Use(serviceabilityMiddleware.RateLimiter())

	// Single postal code check endpoint (GET /check/{postal_code})
	checkRoute.Get("/:postal_code", serviceabilityHandler.CheckPostalCodeServiceability)

	// Source and destination postal code check endpoint (POST /check)
	// This replaces the existing enhanced serviceability endpoint
	checkRoute.Post("/", serviceabilityHandler.CheckPostalCodeServiceabilityPost)

	// Bulk check endpoint with stricter rate limiting (keeping existing functionality)
	bulkCheckRoute := router.Group("/bulk-check")
	bulkCheckRoute.Use(serviceabilityMiddleware.BulkRequestLimiter())
	bulkCheckRoute.Post("/", serviceabilityHandler.BulkCheckServiceability)
}
