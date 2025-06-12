package http

import (
	"context"
	"fmt"
	"log"
	"time"

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
	servicesv1 "prayog-serviceability-service/internal/services/v1"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	repositories "prayog-serviceability-service/internal/shared/repositories/v1"
	sharedServices "prayog-serviceability-service/internal/shared/services/v1"
	"prayog-serviceability-service/internal/shared/utils/v1"
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

	// Create global serviceability group
	serviceabilityGroup := s.app.Group("/serviceability")

	// Setup health check routes under serviceability prefix
	routes.RegisterHealthRoutes(serviceabilityGroup, healthHandler)

	// Create API version group under serviceability
	v1 := serviceabilityGroup.Group("/v1")

	// Create serviceability handler with all dependencies
	serviceabilityHandler, err := s.createServiceabilityHandler()
	if err != nil {
		return fmt.Errorf("failed to create serviceability handler: %w", err)
	}

	// Register serviceability routes directly under /serviceability/v1/
	routes.RegisterServiceabilityRoutes(v1, serviceabilityHandler, s.logger)

	// Add a status route for the serviceability service
	v1.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "serviceability",
			"status":  "available",
			"message": "Serviceability API is ready",
		})
	})

	// Create location handler and register location routes
	locationHandler, err := s.createLocationHandler()
	if err != nil {
		return fmt.Errorf("failed to create location handler: %w", err)
	}

	// Register location routes under /serviceability/v1/
	routes.RegisterLocationRoutes(v1, locationHandler, s.logger)

	// Create partner location coverage handler and register routes
	partnerLocationCoverageHandler, err := s.createPartnerLocationCoverageHandler()
	if err != nil {
		return fmt.Errorf("failed to create partner location coverage handler: %w", err)
	}

	// Register partner location coverage routes under /serviceability/v1/
	routes.RegisterPartnerLocationCoverageRoutes(v1, partnerLocationCoverageHandler, s.logger)

	s.logger.Info("All routes configured successfully")
	return nil
}

// createServiceabilityHandler creates a serviceability handler with all dependencies
func (s *Server) createServiceabilityHandler() (*handlers.ServiceabilityHandler, error) {
	// Create validator instance with all custom validations registered
	validatorSetup := utils.NewValidatorSetup()
	validator := validatorSetup.GetValidator()

	// Create serviceability handler
	serviceabilityHandler := handlers.NewServiceabilityHandler(
		s.orchestrator,
		validator,
		s.logger,
	)

	return serviceabilityHandler, nil
}

// createLocationHandler creates a location handler with all dependencies
func (s *Server) createLocationHandler() (*handlers.LocationHandler, error) {
	// Check if database manager is available
	if s.dbManager == nil {
		return nil, fmt.Errorf("database manager is required for location handler")
	}

	// Create validator instance with all custom validations registered
	validatorSetup := utils.NewValidatorSetup()
	validator := validatorSetup.GetValidator()

	// Create repository factory from database connection
	db := s.dbManager.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	repoFactory := repositories.NewRepositoryFactory(db)

	// Create location service from repository factory
	locationService := sharedServices.NewLocationService(repoFactory.GetLocationRepository(), s.logger)

	// Create location handler
	locationHandler := handlers.NewLocationHandler(
		locationService,
		validator,
		s.logger,
	)

	return locationHandler, nil
}

// createPartnerLocationCoverageHandler creates a partner location coverage handler with all dependencies
func (s *Server) createPartnerLocationCoverageHandler() (*handlers.PartnerLocationCoverageHandler, error) {
	// Check if database manager is available
	if s.dbManager == nil {
		return nil, fmt.Errorf("database manager is required for partner location coverage handler")
	}

	// Check if integration factory is available
	if s.integrationFactory == nil {
		return nil, fmt.Errorf("integration factory is required for partner location coverage handler")
	}

	// Create validator instance with all custom validations registered
	validatorSetup := utils.NewValidatorSetup()
	validator := validatorSetup.GetValidator()

	// Create repository factory from database connection
	db := s.dbManager.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	repoFactory := repositories.NewRepositoryFactory(db)

	// Get partner location coverage repository
	partnerLocationCoverageRepo := repoFactory.GetPartnerLocationCoverageRepository()

	// Get location repository
	locationRepo := repoFactory.GetLocationRepository()

	// Create partner validation service
	partnerValidationConfig := servicesv1.LoadPartnerValidationConfig()

	// Create standard log.Logger for partner services
	stdLogger := log.New(s.logger.WithField("component", "partner").WriterLevel(logrus.InfoLevel), "[partner] ", log.LstdFlags)

	partnerHTTPClient := servicesv1.NewPartnerHTTPClient(partnerValidationConfig, stdLogger)
	partnerValidationService := servicesv1.NewPartnerValidationService(
		partnerValidationConfig,
		partnerHTTPClient,
		stdLogger,
	)

	// Create partner location coverage service
	partnerLocationCoverageService := sharedServices.NewPartnerLocationCoverageService(
		partnerLocationCoverageRepo,
		locationRepo,
		partnerValidationService,
	)

	// Create partner location coverage handler
	partnerLocationCoverageHandler := handlers.NewPartnerLocationCoverageHandler(
		partnerLocationCoverageService,
		validator,
		s.logger,
	)

	return partnerLocationCoverageHandler, nil
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
