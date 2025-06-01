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

	// Single check endpoint with standard rate limiting
	checkRoute := router.Group("/check")
	checkRoute.Use(serviceabilityMiddleware.RateLimiter())
	checkRoute.Post("/", serviceabilityHandler.CheckServiceability)

	// Bulk check endpoint with stricter rate limiting
	bulkCheckRoute := router.Group("/bulk-check")
	bulkCheckRoute.Use(serviceabilityMiddleware.BulkRequestLimiter())
	bulkCheckRoute.Post("/", serviceabilityHandler.BulkCheckServiceability)
}
