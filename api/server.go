package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"prayog-serviceability-service/api/handlers"
	v1 "prayog-serviceability-service/api/handlers/v1"
	v1routes "prayog-serviceability-service/api/routes/v1"
	"prayog-serviceability-service/pkg/services"
	"prayog-serviceability-service/pkg/version"
)

// Server represents an HTTP server.
type Server struct {
	app *fiber.App
}

// NewServer creates a new HTTP server.
func NewServer(serviceabilityService services.ServiceabilityService) *Server {
	app := fiber.New(fiber.Config{
		AppName: "Prayog Serviceability Service",
	})

	// Use global middlewares
	app.Use(logger.New())
	app.Use(recover.New())

	// Add version header to all responses
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-API-Version", version.Version())
		return c.Next()
	})

	// Add a health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Add version endpoint
	versionHandler := handlers.NewVersionHandler()
	app.Get("/version", versionHandler.GetVersion)

	// Set up API v1 routes
	apiV1 := app.Group("/api/v1")

	// Initialize handlers with services
	serviceabilityHandler := v1.NewServiceabilityHandler(serviceabilityService)

	// Register routes
	v1routes.RegisterServiceabilityRoutes(apiV1, serviceabilityHandler)

	return &Server{
		app: app,
	}
}

// Listen starts the HTTP server on the specified address.
func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
