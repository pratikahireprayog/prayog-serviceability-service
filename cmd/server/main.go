package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	httpServer "prayog-serviceability-service/internal/infrastructure/api/http"
	"prayog-serviceability-service/internal/infrastructure/db"
	businessServices "prayog-serviceability-service/internal/services/v1/business"
	integrationServices "prayog-serviceability-service/internal/services/v1/integration"
	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/services/v2/partners/factory"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"

	// NOTE: gRPC server imports are commented out for now
	// grpcServer "prayog-serviceability-service/internal/infrastructure/api/grpc"
	dataServices "prayog-serviceability-service/internal/services/v1/data"
	repositories "prayog-serviceability-service/internal/shared/repositories/v1"
)

func main() {
	// Initialize logger
	logger := initLogger()
	logger.Info("🚀 Starting Prayog Serviceability Service...")
	logger.Info("📝 Note: Starting with HTTP server only - gRPC server is disabled for now")

	// Load configuration
	appConfig, err := initConfig(logger)
	if err != nil {
		logger.WithError(err).Fatal("❌ Failed to initialize configuration")
	}

	// Initialize database manager
	dbManager, err := initDatabase(appConfig, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}

	// Initialize integration factory (external service clients)
	integrationFactory, err := initIntegrationFactory(appConfig, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize integration factory")
	}

	// Initialize services and orchestrator
	orchestrator, err := initOrchestrator(integrationFactory, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize orchestrator")
	}

	// Initialize V2 orchestrator
	v2Orchestrator, err := initV2Orchestrator(appConfig, dbManager, integrationFactory, logger)
	if err != nil {
		logger.WithError(err).Warn("⚠️ Failed to initialize V2 orchestrator - V2 features will be disabled")
		v2Orchestrator = nil
	}

	// Create HTTP server with all dependencies
	// NOTE: Only HTTP server is initialized - gRPC server setup is commented out
	server, err := initHTTPServer(appConfig, dbManager, integrationFactory, orchestrator, v2Orchestrator, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize HTTP server")
	}

	// TODO: gRPC server initialization will be added here when needed
	// grpcServer, err := initGRPCServer(appConfig, dbManager, integrationFactory, orchestrator, logger)
	// if err != nil {
	//     logger.WithError(err).Fatal("Failed to initialize gRPC server")
	// }

	// Setup graceful shutdown
	setupGracefulShutdown(server, logger)

	// Start server
	port := getPort(appConfig)
	logger.WithField("port", port).Info("🚀 Starting HTTP server")
	logger.Info("📝 gRPC server is disabled - only HTTP endpoints are available")

	// Add some startup info like the partner service
	logger.Infof("📍 Health check available at: http://localhost:%d/serviceability/ping", port)
	logger.Infof("🌐 API endpoints available at: http://localhost:%d/serviceability/v1/", port)
	logger.Info("🔥 High-concurrency mode: 1048576 max connections")

	if err := server.Start(fmt.Sprintf(":%d", port)); err != nil {
		logger.WithError(err).Fatal("❌ Failed to start HTTP server")
	}
}

// initLogger initializes the application logger
func initLogger() *logrus.Logger {
	logger := logrus.New()

	// Use text formatter for human-readable logs with colors and timestamps
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006/01/02 15:04:05",
		FullTimestamp:   true,
		ForceColors:     true,
		DisableQuote:    true,
	})

	// Set log level from environment or default to Info
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}

// initConfig initializes the configuration
func initConfig(logger *logrus.Logger) (*config.AppConfig, error) {
	logger.Info("🔧 Initializing configuration...")

	appConfig, err := config.LoadAppConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	logger.Info("✅ Configuration initialized successfully")
	return appConfig, nil
}

// initDatabase initializes the database manager
func initDatabase(appConfig *config.AppConfig, logger *logrus.Logger) (*db.DatabaseManager, error) {
	logger.Info("🗄️ Initializing database connection...")

	dbManager, err := db.NewDatabaseManager(appConfig, logger)
	if err != nil {
		logger.WithError(err).Warn("⚠️ Failed to create database manager - continuing without database")
		// Return nil for database manager but don't fail startup
		return nil, nil
	}

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbManager.Ping(ctx); err != nil {
		logger.WithError(err).Warn("⚠️ Database connection test failed - continuing with service startup")
		// Don't fail the startup for database connectivity issues
	} else {
		logger.Info("✅ Database connection successful")
	}

	return dbManager, nil
}

// initIntegrationFactory initializes the integration factory for external services
func initIntegrationFactory(appConfig *config.AppConfig, logger *logrus.Logger) (*integrationServices.IntegrationFactory, error) {
	logger.Info("🔌 Initializing integration factory...")

	factory, err := integrationServices.NewIntegrationFactory(appConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create integration factory: %w", err)
	}

	logger.Info("✅ Integration factory initialized successfully")
	return factory, nil
}

// initOrchestrator initializes the serviceability orchestrator with all dependencies
func initOrchestrator(integrationFactory *integrationServices.IntegrationFactory, logger *logrus.Logger) (interfaces.ServiceabilityOrchestrator, error) {
	logger.Info("⚙️ Initializing serviceability orchestrator with real service integrations...")

	// Create external service clients
	partnerService, err := integrationFactory.CreatePartnerServiceClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create partner service client: %w", err)
	}

	specService, err := integrationFactory.CreateSpecificationServiceClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create specification service client: %w", err)
	}

	// Create core services with real implementations
	locationResolver := businessServices.NewLocationResolver(partnerService)

	// Create serviceability calculator
	serviceabilityCalculator := businessServices.NewServiceabilityCalculator()

	// Create and return the real orchestrator
	orchestrator := businessServices.NewOptimizedServiceabilityOrchestrator(
		locationResolver,
		partnerService,
		specService,
		serviceabilityCalculator,
	)

	logger.Info("✅ Successfully initialized real serviceability orchestrator")
	return orchestrator, nil
}

// initV2Orchestrator initializes the V2 serviceability orchestrator with partner adapters
func initV2Orchestrator(
	appConfig *config.AppConfig,
	dbManager *db.DatabaseManager,
	integrationFactory *integrationServices.IntegrationFactory,
	logger *logrus.Logger,
) (orchestrators.ServiceabilityOrchestrator, error) {
	logger.Info("⚙️ Initializing V2 serviceability orchestrator with partner adapters...")

	// Create a proper config manager for integration config
	configManager, err := config.NewConfigManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager for V2 orchestrator: %w", err)
	}

	// Create HTTP client for partner adapters
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Get database connections
	var gormDB *gorm.DB
	var sqlDB *sql.DB
	if dbManager != nil {
		gormDB = dbManager.GetDB()
		if gormDB != nil {
			// Get underlying *sql.DB from GORM
			var err error
			sqlDB, err = gormDB.DB()
			if err != nil {
				logger.WithError(err).Warn("⚠️ Failed to get underlying SQL DB from GORM")
				sqlDB = nil
			}
		}
	}

	// Create geolocation service for partner adapters
	var geolocationService dataServices.GeolocationService
	if gormDB != nil {
		// Create repository factory from database connection
		repoFactory := repositories.NewRepositoryFactory(gormDB)
		postalCodeRepo := repoFactory.GetPostalCodeRepository()
		geolocationService = dataServices.NewGeolocationService(postalCodeRepo)
	} else {
		// Create a dummy geolocation service for graceful degradation
		geolocationService = dataServices.NewGeolocationService(nil)
	}

	// Create partner adapter factory using v2 factory
	partnerAdapterFactory := factory.NewPartnerAdapterFactory(
		configManager.Integration.PartnerAdapters,
		httpClient,
		sqlDB,
		geolocationService,
	)

	// Get partner attribute mapping repository for filtering
	var partnerAttributeRepo repositories.PartnerAttributeMapRepository
	if gormDB != nil {
		// Create repository factory to get partner attribute mapping repo
		repoFactory := repositories.NewRepositoryFactory(gormDB)
		partnerAttributeRepo = repoFactory.GetPartnerAttributeMapRepository()
	} else {
		// For now, we'll pass nil and the orchestrator should handle it gracefully
		partnerAttributeRepo = nil
	}

	// Create V2 orchestrator using v2 orchestrator
	v2Orchestrator := orchestrators.NewServiceabilityOrchestrator(
		partnerAdapterFactory,
		partnerAttributeRepo,
		60*time.Second, // timeout for partner requests - increased for database queries
		configManager.App.Serviceability.ReturnOnlyServiceablePartners,
	)

	logger.Info("✅ Successfully initialized V2 serviceability orchestrator")
	return v2Orchestrator, nil
}

// initHTTPServer initializes the HTTP server with dependency injection
func initHTTPServer(
	appConfig *config.AppConfig,
	dbManager *db.DatabaseManager,
	integrationFactory *integrationServices.IntegrationFactory,
	orchestrator interfaces.ServiceabilityOrchestrator,
	v2Orchestrator orchestrators.ServiceabilityOrchestrator,
	logger *logrus.Logger,
) (*httpServer.Server, error) {
	logger.Info("🌐 Initializing HTTP server...")

	// Create a proper config manager
	configManager, err := config.NewConfigManager()
	if err != nil {
		// If we can't create a proper config manager, create a minimal one
		logger.WithError(err).Warn("⚠️ Failed to create full config manager, using minimal version")
		configManager = &config.ConfigManager{
			App:         appConfig,
			Integration: config.LoadIntegrationConfig(),
		}
	}

	// Create server with all dependencies
	server, err := httpServer.NewServer(&httpServer.ServerDependencies{
		Config:             configManager,
		DBManager:          dbManager,
		IntegrationFactory: integrationFactory,
		Orchestrator:       orchestrator,
		V2Orchestrator:     v2Orchestrator,
		Logger:             logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP server: %w", err)
	}

	logger.Info("✅ HTTP server initialized successfully")
	return server, nil
}

// setupGracefulShutdown sets up graceful shutdown handling
func setupGracefulShutdown(server *httpServer.Server, logger *logrus.Logger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("🛑 Received shutdown signal, starting graceful shutdown...")

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown server
		if err := server.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("❌ Error during server shutdown")
		}

		logger.Info("✅ Server shutdown complete")
		os.Exit(0)
	}()
}

// getPort returns the port to listen on
func getPort(appConfig *config.AppConfig) int {
	return appConfig.Server.Port
}
