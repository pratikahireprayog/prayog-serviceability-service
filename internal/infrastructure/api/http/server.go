package http

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/sirupsen/logrus"

	handlers "prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
	v1middleware "prayog-serviceability-service/internal/infrastructure/api/http/v1/middleware"
	routes "prayog-serviceability-service/internal/infrastructure/api/http/v1/routes"
	"prayog-serviceability-service/internal/infrastructure/db"
	"prayog-serviceability-service/internal/infrastructure/external/partner_service"
	dataServices "prayog-serviceability-service/internal/services/v1/data"
	integrationServices "prayog-serviceability-service/internal/services/v1/integration"
	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	repositories "prayog-serviceability-service/internal/shared/repositories/v1"
	"prayog-serviceability-service/internal/shared/utils/v1"

	// Add imports for geo location routes and handlers
	v1handlers "prayog-serviceability-service/api/handlers/v1"
	v1routes "prayog-serviceability-service/api/routes/v1"
	v3handlers "prayog-serviceability-service/internal/infrastructure/api/http/v3/handlers"
	v3routes "prayog-serviceability-service/internal/infrastructure/api/http/v3/routes"
	v3 "prayog-serviceability-service/internal/services/v3"
)

// Server represents the HTTP server with all dependencies
type Server struct {
	app                *fiber.App
	config             *config.ConfigManager
	dbManager          *db.DatabaseManager
	integrationFactory *integrationServices.IntegrationFactory
	orchestrator       interfaces.ServiceabilityOrchestrator
	v2Orchestrator     orchestrators.ServiceabilityOrchestrator
	logger             *logrus.Logger
}

// ServerDependencies holds all the dependencies needed to create a server
type ServerDependencies struct {
	Config             *config.ConfigManager
	DBManager          *db.DatabaseManager
	IntegrationFactory *integrationServices.IntegrationFactory
	Orchestrator       interfaces.ServiceabilityOrchestrator
	V2Orchestrator     orchestrators.ServiceabilityOrchestrator
	Logger             *logrus.Logger
}

// NewServer creates a new HTTP server with all dependencies injected
func NewServer(deps *ServerDependencies) (*Server, error) {
	if deps == nil {
		return nil, fmt.Errorf("server dependencies cannot be nil")
	}

	// Initialize Fiber app with configuration optimized for MAXIMUM performance
	app := fiber.New(fiber.Config{
		AppName:                   "prayog-serviceability-service",
		DisableStartupMessage:     false,            // Enable startup message for better UX
		ReadTimeout:               5 * time.Second,  // Aggressive timeout for faster resource recycling
		WriteTimeout:              5 * time.Second,  // Aggressive timeout
		IdleTimeout:               30 * time.Second, // Quick connection recycling
		BodyLimit:                 50 * 1024 * 1024, // 50MB body limit
		Concurrency:               1024 * 1024,      // 1 MILLION concurrent connections
		ReadBufferSize:            16384,            // 16KB read buffer (increased)
		WriteBufferSize:           16384,            // 16KB write buffer (increased)
		CompressedFileSuffix:      ".fiber.gz",      // Enable compression
		ProxyHeader:               fiber.HeaderXForwardedFor,
		DisableKeepalive:          false, // Keep connections alive
		DisableDefaultDate:        true,  // Disable date header for performance
		DisableDefaultContentType: true,  // Disable default content type for performance
		DisableHeaderNormalizing:  true,  // Disable header normalization for performance
		ReduceMemoryUsage:         false, // Don't reduce memory for maximum performance
		ErrorHandler:              customErrorHandler(deps.Logger),
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
		v2Orchestrator:     deps.V2Orchestrator,
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

	// CORS middleware using Fiber's built-in cors.Config - optimized for Flutter Web
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS,HEAD,PATCH",
		AllowHeaders: strings.Join([]string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
			"X-Requested-With",
			"X-Tenant-ID", // COMMENTED FOR TESTING - Your API requires this
			"tenantid",    // COMMENTED FOR TESTING - Your API requires this (lowercase variant)
			"User-Agent",
			"Referer",
			"sec-ch-ua", // Chrome security headers
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
		}, ","),
		AllowCredentials: false,
		ExposeHeaders: strings.Join([]string{
			"Content-Length",
			"X-API-Version",
			"X-Service-Name",
			"X-Request-ID",
		}, ","),
		MaxAge: 86400, // 24 hours preflight cache
	}))

	// CORS debugging middleware (remove in production)
	app.Use(func(c *fiber.Ctx) error {
		if c.Method() == "OPTIONS" {
			logger.WithFields(logrus.Fields{
				"method":  c.Method(),
				"path":    c.Path(),
				"origin":  c.Get("Origin"),
				"headers": c.Get("Access-Control-Request-Headers"),
			}).Debug("CORS preflight request received")
		}
		return c.Next()
	})

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

	// Tenant middleware to extract tenant_id and user_id from headers
	app.Use(v1middleware.TenantMiddleware())
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

	// Create API v2 group under serviceability
	v2 := serviceabilityGroup.Group("/v2")

	// Create V2 serviceability handler if v2Orchestrator is available
	if s.v2Orchestrator != nil {
		v2ServiceabilityHandler, err := s.createServiceabilityV2Handler()
		if err != nil {
			s.logger.WithError(err).Warn("V2 serviceability features are disabled")
			// Create placeholder routes that return service unavailable
			v2.All("/*", func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
					"error": fiber.Map{
						"code":    "SERVICE_UNAVAILABLE",
						"message": "V2 serviceability features are temporarily unavailable",
					},
				})
			})
		} else {
			// Register V2 serviceability routes under /serviceability/v2/
			routes.RegisterServiceabilityV2Routes(v2, v2ServiceabilityHandler)
		}

		// Add a status route for the V2 serviceability service
		v2.Get("/status", func(c *fiber.Ctx) error {
			status := fiber.Map{
				"service": "serviceability-v2",
				"version": "2.0.0",
				"status":  "available",
				"message": "Serviceability V2 API is ready",
			}

			// Add partner adapter status information
			if s.v2Orchestrator != nil {
				status["partner_adapters"] = "available"
				status["features"] = fiber.Map{
					"multi_partner_orchestration": "available",
					"concurrent_partner_calls":    "available",
					"partner_filtering":           "available",
					"attribute_based_selection":   "available",
				}
			} else {
				status["partner_adapters"] = "unavailable"
				status["features"] = fiber.Map{
					"multi_partner_orchestration": "unavailable",
					"concurrent_partner_calls":    "unavailable",
					"partner_filtering":           "unavailable",
					"attribute_based_selection":   "unavailable",
				}
			}

			return c.JSON(status)
		})
	} else {
		s.logger.Warn("V2 orchestrator is not available - V2 serviceability features will be disabled")
		// Create placeholder routes that return service unavailable
		v2.All("/*", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "V2 serviceability features are not enabled",
				},
			})
		})
	}

	// Create API v3 group under serviceability
	v3 := serviceabilityGroup.Group("/v3")

	// Create V3 serviceability handler if v2Orchestrator is available (V3 uses V2 orchestrator)
	if s.v2Orchestrator != nil {
		v3ServiceabilityHandler, err := s.createServiceabilityV3Handler()
		if err != nil {
			s.logger.WithError(err).Warn("V3 serviceability features are disabled")
			// Create placeholder routes that return service unavailable
			v3.All("/*", func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
					"error": fiber.Map{
						"code": "SERVICE_UNAVAILABLE",
					},
				})
			})
		} else {
			// Register V3 serviceability routes under /serviceability/v3/
			v3routes.RegisterServiceabilityV3Routes(v3, v3ServiceabilityHandler)
		}

		// Add a status route for the V3 serviceability service
		v3.Get("/status", func(c *fiber.Ctx) error {
			status := fiber.Map{
				"service": "serviceability-v3",
				"version": "3.0.0",
				"status":  "available",
				"message": "Serviceability V3 API is ready",
			}

			// Add partner adapter status information
			if s.v2Orchestrator != nil {
				status["partner_adapters"] = "available"
				status["features"] = fiber.Map{
					"addresses_object_format":     "available",
					"services_response_format":    "available",
					"unified_response_structure":  "available",
					"multi_partner_orchestration": "available",
				}
			} else {
				status["partner_adapters"] = "unavailable"
				status["features"] = fiber.Map{
					"addresses_object_format":     "unavailable",
					"services_response_format":    "unavailable",
					"unified_response_structure":  "unavailable",
					"multi_partner_orchestration": "unavailable",
				}
			}

			return c.JSON(status)
		})
	} else {
		s.logger.Warn("V2 orchestrator is not available - V3 serviceability features will be disabled")
		// Create placeholder routes that return service unavailable
		v3.All("/*", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "V3 serviceability features are not enabled",
				},
			})
		})
	}

	// Add a status route for the serviceability service
	v1.Get("/status", func(c *fiber.Ctx) error {
		status := fiber.Map{
			"service": "serviceability",
			"status":  "available",
			"message": "Serviceability API is ready",
		}

		// Add database status information
		if s.dbManager == nil {
			status["database"] = "unavailable"
			status["features"] = fiber.Map{
				"basic_serviceability":       "available",
				"postal_code_serviceability": "unavailable - database required",
				"location_management":        "unavailable - database required",
				"partner_location_coverage":  "unavailable - database required",
			}
		} else {
			status["database"] = "available"
			status["features"] = fiber.Map{
				"basic_serviceability":       "available",
				"postal_code_serviceability": "available",
				"location_management":        "available",
				"partner_location_coverage":  "available",
			}
		}

		return c.JSON(status)
	})

	// Try to create location handler and register location routes if database is available
	locationHandler, err := s.createLocationHandler()
	if err != nil {
		s.logger.WithError(err).Warn("Location management features are disabled")
		// Create a placeholder route that returns service unavailable
		v1.All("/locations/*", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "Location management features are temporarily unavailable - database connection required",
				},
			})
		})
	} else {
		// Register location routes under /serviceability/v1/
		routes.RegisterLocationRoutes(v1, locationHandler, s.logger)
	}

	// Try to create partner location coverage handler and register routes if database is available
	partnerLocationCoverageHandler, err := s.createPartnerLocationCoverageHandler()
	if err != nil {
		s.logger.WithError(err).Warn("Partner location coverage features are disabled")
		// Create a placeholder route that returns service unavailable
		v1.All("/partner-location-coverage/*", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "Partner location coverage features are temporarily unavailable - database connection required",
				},
			})
		})
	} else {
		// Register partner location coverage routes under /serviceability/v1/
		routes.RegisterPartnerLocationCoverageRoutes(v1, partnerLocationCoverageHandler, s.logger)
	}

	// Try to create partner attribute handler and register routes if database is available
	partnerAttributeHandler, err := s.createPartnerAttributeHandler()
	if err != nil {
		s.logger.WithError(err).Warn("Partner attribute features are disabled")
		// Create placeholder routes that return service unavailable
		v1.All("/attribute-categories/*", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "Partner attribute features are temporarily unavailable - database connection required",
				},
			})
		})
		v1.All("/attributes/*", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "Partner attribute features are temporarily unavailable - database connection required",
				},
			})
		})
		v1.All("/partner-attribute-maps/*", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "Partner attribute features are temporarily unavailable - database connection required",
				},
			})
		})
		v1.All("/partners/*", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "Partner attribute features are temporarily unavailable - database connection required",
				},
			})
		})
	} else {
		// Register partner attribute routes under /serviceability/v1/
		routes.RegisterPartnerAttributeRoutes(v1, partnerAttributeHandler, s.logger)
	}

	// Try to create geo location handler and register geo location routes if database is available
	geoLocationHandler, err := s.createGeoLocationHandler()
	if err != nil {
		s.logger.WithError(err).Warn("Geo location features are disabled")
		// Create a placeholder route that returns service unavailable
		v1.All("/geo-locations/*", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "Geo location features are temporarily unavailable - database connection required",
				},
			})
		})
	} else {
		// Register geo location routes under /serviceability/v1/
		v1routes.RegisterGeoLocationRoutes(v1, geoLocationHandler)
	}

	// --- Add NearestHubLocation routes registration ---
	if s.dbManager != nil {
		db := s.dbManager.GetDB()
		if db != nil {
			repo := repositories.NewNearestHubLocationRepository(db)
			handler := v1handlers.NewNearestHubLocationHandler(repo)
			v1routes.RegisterNearestHubLocationRoutes(v1, handler)
		}
	}

	s.logger.Info("✅ Routes configured successfully - some features may be disabled due to database unavailability")

	return nil
}

// createServiceabilityHandler creates a serviceability handler with all dependencies
func (s *Server) createServiceabilityHandler() (*handlers.ServiceabilityHandler, error) {
	// Create validator instance with all custom validations registered
	validatorSetup := utils.NewValidatorSetup()
	validator := validatorSetup.GetValidator()

	// Check if database manager is available for postal code serviceability
	if s.dbManager == nil {
		s.logger.Warn("Database manager is not available - creating serviceability handler without postal code serviceability features")

		// Create serviceability handler with only orchestrator (no postal code service)
		// This allows the service to run with limited functionality
		serviceabilityHandler := handlers.NewServiceabilityHandler(
			s.orchestrator,
			nil, // No postal code serviceability service
			validator,
			s.logger,
		)

		return serviceabilityHandler, nil
	}

	// Create repository factory from database connection
	db := s.dbManager.GetDB()
	if db == nil {
		s.logger.Warn("Database connection is not available - creating serviceability handler without postal code serviceability features")

		// Create serviceability handler with only orchestrator (no postal code service)
		serviceabilityHandler := handlers.NewServiceabilityHandler(
			s.orchestrator,
			nil, // No postal code serviceability service
			validator,
			s.logger,
		)

		return serviceabilityHandler, nil
	}

	repoFactory := repositories.NewRepositoryFactory(db)

	// Get partner location coverage repository for postal code serviceability
	partnerLocationCoverageRepo := repoFactory.GetPartnerLocationCoverageRepository()

	// Create postal code serviceability service
	postalCodeServiceabilityService := dataServices.NewPostalCodeServiceabilityService(
		partnerLocationCoverageRepo,
	)

	// Create serviceability handler with both orchestrator and postal code service
	serviceabilityHandler := handlers.NewServiceabilityHandler(
		s.orchestrator,
		postalCodeServiceabilityService,
		validator,
		s.logger,
	)

	return serviceabilityHandler, nil
}

// createServiceabilityV2Handler creates a V2 serviceability handler with all dependencies
func (s *Server) createServiceabilityV2Handler() (*handlers.ServiceabilityV2Handler, error) {
	// Check if V2 orchestrator is available
	if s.v2Orchestrator == nil {
		return nil, fmt.Errorf("V2 orchestrator is required")
	}

	// Create validator instance with all custom validations registered
	validatorSetup := utils.NewValidatorSetup()
	validator := validatorSetup.GetValidator()

	// Create geolocation service for country code resolution
	var geolocationService dataServices.GeolocationService
	if s.dbManager != nil {
		// Create repository factory from database connection
		db := s.dbManager.GetDB()
		if db != nil {
			repoFactory := repositories.NewRepositoryFactory(db)
			geoLocationRepo := repoFactory.GetGeoLocationRepository()
			geolocationService = dataServices.NewGeolocationService(geoLocationRepo)
		}
	}

	// Create V2 serviceability handler
	v2ServiceabilityHandler := handlers.NewServiceabilityV2Handler(
		s.v2Orchestrator,
		geolocationService,
		validator,
		s.logger,
	)

	return v2ServiceabilityHandler, nil
}

// createServiceabilityV3Handler creates a V3 serviceability handler with all dependencies
func (s *Server) createServiceabilityV3Handler() (*v3handlers.ServiceabilityHandler, error) {
	// Check if V2 orchestrator is available (V3 uses V2 orchestrator)
	if s.v2Orchestrator == nil {
		return nil, fmt.Errorf("V2 orchestrator is required for V3 API")
	}

	// Create validator instance with all custom validations registered
	validatorSetup := utils.NewValidatorSetup()
	validator := validatorSetup.GetValidator()

	// Create Partner Service Client
	// Read partner service URL from environment variable (required)
	partnerServiceURL := os.Getenv("PARTNER_SERVICE_URL")
	if partnerServiceURL == "" {
		return nil, fmt.Errorf("PARTNER_SERVICE_URL environment variable is required for V3 API")
	}
	// Make sure to import "prayog-serviceability-service/internal/infrastructure/external/partner_service"
	partnerClient := partner_service.NewPartnerServiceClient(partnerServiceURL, s.logger)

	// Create V3 Service
	v3Service := v3.NewServiceabilityService(s.v2Orchestrator, partnerClient, s.logger)

	// Create V3 serviceability handler
	v3ServiceabilityHandler := v3handlers.NewServiceabilityHandler(
		v3Service,
		validator,
		s.logger,
	)

	return v3ServiceabilityHandler, nil
}

// createLocationHandler creates a location handler with all dependencies
func (s *Server) createLocationHandler() (*handlers.LocationHandler, error) {
	// Check if database manager is available
	if s.dbManager == nil {
		s.logger.Warn("Database manager is not available - location management features will be disabled")
		return nil, fmt.Errorf("location management features require database connection - currently unavailable")
	}

	// Create validator instance with all custom validations registered
	validatorSetup := utils.NewValidatorSetup()
	validator := validatorSetup.GetValidator()

	// Create repository factory from database connection
	db := s.dbManager.GetDB()
	if db == nil {
		s.logger.Warn("Database connection is not available - location management features will be disabled")
		return nil, fmt.Errorf("location management features require database connection - currently unavailable")
	}

	repoFactory := repositories.NewRepositoryFactory(db)

	// Create location service from repository factory
	locationService := dataServices.NewLocationService(repoFactory.GetLocationRepository(), s.logger)

	// Create location handler
	locationHandler := handlers.NewLocationHandler(
		locationService,
		validator,
		s.logger,
	)

	return locationHandler, nil
}

// createPartnerAttributeHandler creates a partner attribute handler with all dependencies
func (s *Server) createPartnerAttributeHandler() (*handlers.PartnerAttributeHandler, error) {
	// Check if database manager is available
	if s.dbManager == nil {
		return nil, fmt.Errorf("database manager is required for partner attribute handler")
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

	// Get repositories and create services
	attributeCategoryRepo := repoFactory.GetAttributeCategoryRepository()
	attributeRepo := repoFactory.GetAttributeRepository()
	partnerAttributeMapRepo := repoFactory.GetPartnerAttributeMapRepository()

	// Create services
	attributeCategoryService := dataServices.NewAttributeCategoryService(attributeCategoryRepo)
	attributeService := dataServices.NewAttributeService(attributeRepo)
	partnerAttributeMapService := dataServices.NewPartnerAttributeMapService(
		partnerAttributeMapRepo,
		attributeRepo,
		attributeCategoryRepo,
	)

	// Create partner attribute handler with all services
	partnerAttributeHandler, err := handlers.NewPartnerAttributeHandler(
		attributeCategoryService,
		attributeService,
		partnerAttributeMapService,
		validator,
		s.logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create partner attribute handler: %w", err)
	}

	return partnerAttributeHandler, nil
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
	partnerValidationConfig := integrationServices.LoadPartnerValidationConfig()

	// Create standard log.Logger for partner services
	stdLogger := log.New(s.logger.WithField("component", "partner").WriterLevel(logrus.InfoLevel), "[partner] ", log.LstdFlags)

	partnerHTTPClient := integrationServices.NewPartnerHTTPClient(partnerValidationConfig, stdLogger)
	partnerValidationService := integrationServices.NewPartnerValidationService(
		partnerValidationConfig,
		partnerHTTPClient,
		stdLogger,
	)

	// Create partner location coverage service
	partnerLocationCoverageService := dataServices.NewPartnerLocationCoverageService(
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

// createGeoLocationHandler creates a geo location handler with all dependencies
func (s *Server) createGeoLocationHandler() (*v1handlers.GeoLocationHandler, error) {
	// Check if database manager is available
	if s.dbManager == nil {
		s.logger.Warn("Database manager is not available - geo location features will be disabled")
		return nil, fmt.Errorf("geo location features require database connection - currently unavailable")
	}

	// Create validator instance with all custom validations registered
	validatorSetup := utils.NewValidatorSetup()
	validator := validatorSetup.GetValidator()

	// Create repository factory from database connection
	db := s.dbManager.GetDB()
	if db == nil {
		s.logger.Warn("Database connection is not available - geo location features will be disabled")
		return nil, fmt.Errorf("geo location features require database connection - currently unavailable")
	}

	repoFactory := repositories.NewRepositoryFactory(db)

	// Create geo location service from repository factory
	geoLocationService := dataServices.NewGeoLocationService(repoFactory.GetGeoLocationRepository(), validator)

	// Create geo location handler
	geoLocationHandler := v1handlers.NewGeoLocationHandler(geoLocationService)

	return geoLocationHandler, nil
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
