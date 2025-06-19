package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"prayog-serviceability-service/api/handlers"
	v1 "prayog-serviceability-service/api/handlers/v1"
	v1routes "prayog-serviceability-service/api/routes/v1"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// Server represents an HTTP server.
type Server struct {
	app *fiber.App
}

// NewServer creates a new HTTP server.
func NewServer(serviceabilityService interfaces.ServiceabilityOrchestrator) *Server {
	app := fiber.New(fiber.Config{
		AppName: "Prayog Serviceability Service",
	})

	// Use global middlewares
	app.Use(logger.New())
	app.Use(recover.New())

	// Configure CORS with wildcard origin support
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowCredentials: false, // Set to false when using wildcard origin
		ExposeHeaders:    "Content-Length, X-API-Version",
		MaxAge:           86400, // 24 hours
	}))

	// Add version header to all responses (temporarily hardcoded)
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-API-Version", "1.0.0")
		return c.Next()
	})

	// Create serviceability group with global prefix
	serviceabilityGroup := app.Group("/serviceability")

	// Add health check endpoint under serviceability prefix
	serviceabilityGroup.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "pong"})
	})

	// Add version endpoint under serviceability prefix
	versionHandler := handlers.NewVersionHandler()
	serviceabilityGroup.Get("/version", versionHandler.GetVersion)

	// Set up API v1 routes under serviceability prefix
	apiV1 := serviceabilityGroup.Group("/api/v1")

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
