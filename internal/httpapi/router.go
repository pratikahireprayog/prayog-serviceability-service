package httpapi

import (
	"github.com/gofiber/fiber/v2"
	fiberLimiter "github.com/gofiber/fiber/v2/middleware/limiter"

	"prayog-serviceability-service/internal/httpapi/handlers"
	apimiddleware "prayog-serviceability-service/internal/httpapi/middleware"
)

// Router handles HTTP routing for the application.
type Router struct {
	config RouterConfig
	app    *fiber.App
}

// NewRouter creates a new HTTP router.
func NewRouter(app *fiber.App, config RouterConfig) *Router {
	// Rate limiting if enabled
	if config.RateLimiting {
		app.Use(fiberLimiter.New())
	}

	// Auth middleware if enabled
	if config.AuthEnabled {
		app.Use(apimiddleware.Auth(config.APIKeys))
	}

	// Health check endpoints
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString(config.Version)
	})

	// API routes
	apiGroup := app.Group("/api/v1")
	serviceabilityGroup := apiGroup.Group("/serviceability")

	// Register all handlers
	handlers.RegisterHandlers(serviceabilityGroup, config.UsecaseFactory)

	return &Router{
		config: config,
		app:    app,
	}
}
