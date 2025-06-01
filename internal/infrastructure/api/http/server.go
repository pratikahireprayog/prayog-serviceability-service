package http

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/routes"
	"prayog-serviceability-service/internal/infrastructure/db"
	"prayog-serviceability-service/internal/services/v1"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// Server represents the HTTP server with all dependencies
type Server struct {
	app                *fiber.App
	config             *config.ConfigManager
	dbManager          *db.DatabaseManager
	integrationFactory *services.IntegrationFactory
	orchestrator       interfaces.ServiceabilityOrchestrator
	logger             *logrus.Logger
}

// ServerDependencies holds all the dependencies needed to create a server
type ServerDependencies struct {
	Config             *config.ConfigManager
	DBManager          *db.DatabaseManager
	IntegrationFactory *services.IntegrationFactory
	Orchestrator       interfaces.ServiceabilityOrchestrator
	Logger             *logrus.Logger
}

// NewServer creates a new HTTP server with all dependencies injected
func NewServer(deps *ServerDependencies) (*Server, error) {
	if deps == nil {
		return nil, fmt.Errorf("server dependencies cannot be nil")
	}

	// Initialize Fiber app with configuration
	app := fiber.New(fiber.Config{
		AppName:               "Prayog Serviceability Service",
		DisableStartupMessage: false,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           120 * time.Second,
		ErrorHandler:          customErrorHandler(deps.Logger),
	})

	// Setup global middleware
	setupMiddleware(app, deps.Logger)

	// Create server instance
	server := &Server{
		app:                app,
		config:             deps.Config,
		dbManager:          deps.DBManager,
		integrationFactory: deps.IntegrationFactory,
		orchestrator:       deps.Orchestrator,
		logger:             deps.Logger,
	}

	// Setup routes
	if err := server.setupRoutes(); err != nil {
		return nil, fmt.Errorf("failed to setup routes: %w", err)
	}

	return server, nil
}

// setupMiddleware configures global middleware for the Fiber app
func setupMiddleware(app *fiber.App, logger *logrus.Logger) {
	// Request ID middleware
	app.Use(requestid.New())

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID",
	}))

	// Logger middleware
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format:     "${time} ${status} - ${method} ${path} ${latency}\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "UTC",
	}))

	// Recovery middleware
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// Add version and service headers to all responses
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Service-Name", "prayog-serviceability-service")
		c.Set("X-API-Version", "v1")
		return c.Next()
	})
}

// setupRoutes configures all application routes
func (s *Server) setupRoutes() error {
	// Create health handler
	healthHandler := handlers.NewHealthHandler(
		s.dbManager,
		s.integrationFactory,
		s.config,
		s.logger,
	)

	// Setup health check routes at root level
	routes.RegisterHealthRoutes(s.app, healthHandler)

	// Create API group
	api := s.app.Group("/api")
	v1 := api.Group("/v1")

	// Create serviceability group
	serviceabilityGroup := v1.Group("/serviceability")

	// Create serviceability handler with all dependencies
	serviceabilityHandler, err := s.createServiceabilityHandler()
	if err != nil {
		return fmt.Errorf("failed to create serviceability handler: %w", err)
	}

	// Register serviceability routes
	routes.RegisterServiceabilityRoutes(serviceabilityGroup, serviceabilityHandler, s.logger)

	// Add a status route for the serviceability service
	serviceabilityGroup.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "serviceability",
			"status":  "available",
			"message": "Serviceability API is ready",
		})
	})

	s.logger.Info("All routes configured successfully")
	return nil
}

// createServiceabilityHandler creates a serviceability handler with all dependencies
func (s *Server) createServiceabilityHandler() (*handlers.ServiceabilityHandler, error) {
	// Create validator instance
	validator := validator.New()

	// Create serviceability handler
	serviceabilityHandler := handlers.NewServiceabilityHandler(
		s.orchestrator,
		validator,
		s.logger,
	)

	return serviceabilityHandler, nil
}

// customErrorHandler creates a custom error handler for Fiber
func customErrorHandler(logger *logrus.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default to 500 server error
		code := fiber.StatusInternalServerError

		// Check if it's a fiber error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		// Log the error
		logger.WithFields(logrus.Fields{
			"status": code,
			"method": c.Method(),
			"path":   c.Path(),
			"error":  err.Error(),
		}).Error("HTTP request error")

		// Return error response
		return c.Status(code).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    code,
				"message": err.Error(),
			},
			"timestamp": time.Now().UTC(),
		})
	}
}

// Start starts the HTTP server
func (s *Server) Start(address string) error {
	s.logger.WithField("address", address).Info("Starting HTTP server")

	if err := s.app.Listen(address); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.logger.Info("HTTP server shutdown complete")
	return nil
}

// GetApp returns the underlying Fiber app (useful for testing)
func (s *Server) GetApp() *fiber.App {
	return s.app
}
