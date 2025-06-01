package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	httpServer "prayog-serviceability-service/internal/infrastructure/api/http"
	"prayog-serviceability-service/internal/infrastructure/db"
	"prayog-serviceability-service/internal/services/v1"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

func main() {
	// Initialize logger
	logger := initLogger()
	logger.Info("Starting Prayog Serviceability Service...")

	// Load configuration
	appConfig, err := initConfig(logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize configuration")
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

	// Create HTTP server with all dependencies
	server, err := initHTTPServer(appConfig, dbManager, integrationFactory, orchestrator, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize HTTP server")
	}

	// Setup graceful shutdown
	setupGracefulShutdown(server, logger)

	// Start server
	port := getPort(appConfig)
	logger.WithField("port", port).Info("Starting HTTP server")

	if err := server.Start(fmt.Sprintf(":%d", port)); err != nil {
		logger.WithError(err).Fatal("Failed to start HTTP server")
	}
}

// initLogger initializes the application logger
func initLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
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
	logger.Info("Initializing configuration...")

	appConfig, err := config.LoadAppConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	logger.Info("Configuration initialized successfully")
	return appConfig, nil
}

// initDatabase initializes the database manager
func initDatabase(appConfig *config.AppConfig, logger *logrus.Logger) (*db.DatabaseManager, error) {
	logger.Info("Initializing database connection...")

	dbManager, err := db.NewDatabaseManager(appConfig, logger)
	if err != nil {
		logger.WithError(err).Warn("Failed to create database manager - continuing without database")
		// Return nil for database manager but don't fail startup
		return nil, nil
	}

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbManager.Ping(ctx); err != nil {
		logger.WithError(err).Warn("Database connection test failed - continuing with service startup")
		// Don't fail the startup for database connectivity issues
	} else {
		logger.Info("Database connection successful")
	}

	return dbManager, nil
}

// initIntegrationFactory initializes the integration factory for external services
func initIntegrationFactory(appConfig *config.AppConfig, logger *logrus.Logger) (*services.IntegrationFactory, error) {
	logger.Info("Initializing integration factory...")

	factory, err := services.NewIntegrationFactory(appConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create integration factory: %w", err)
	}

	logger.Info("Integration factory initialized successfully")
	return factory, nil
}

// initOrchestrator initializes the serviceability orchestrator with all dependencies
func initOrchestrator(integrationFactory *services.IntegrationFactory, logger *logrus.Logger) (interfaces.ServiceabilityOrchestrator, error) {
	logger.Info("Initializing serviceability orchestrator with real service integrations...")

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
	locationResolver := services.NewLocationResolver(partnerService)

	// Create service definition resolver
	serviceDefinitionResolver := services.NewServiceDefinitionResolver(specService)

	// Create partner capability aggregator
	partnerCapabilityAggregator := services.NewPartnerCapabilityAggregator(partnerService, specService)

	// Create serviceability calculator
	serviceabilityCalculator := services.NewServiceabilityCalculator()

	// Create and return the real orchestrator
	orchestrator := services.NewServiceabilityOrchestrator(
		locationResolver,
		partnerCapabilityAggregator,
		serviceDefinitionResolver,
		serviceabilityCalculator,
	)

	logger.Info("Successfully initialized real serviceability orchestrator")
	return orchestrator, nil
}

// initHTTPServer initializes the HTTP server with dependency injection
func initHTTPServer(
	appConfig *config.AppConfig,
	dbManager *db.DatabaseManager,
	integrationFactory *services.IntegrationFactory,
	orchestrator interfaces.ServiceabilityOrchestrator,
	logger *logrus.Logger,
) (*httpServer.Server, error) {
	logger.Info("Initializing HTTP server...")

	// Create a proper config manager
	configManager, err := config.NewConfigManager()
	if err != nil {
		// If we can't create a proper config manager, create a minimal one
		logger.WithError(err).Warn("Failed to create full config manager, using minimal version")
		configManager = &config.ConfigManager{
			App:         appConfig,
			Integration: config.LoadIntegrationConfig(),
		}
	}

	// Create server dependencies
	deps := &httpServer.ServerDependencies{
		Config:             configManager,
		DBManager:          dbManager,
		IntegrationFactory: integrationFactory,
		Orchestrator:       orchestrator,
		Logger:             logger,
	}

	server, err := httpServer.NewServer(deps)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP server: %w", err)
	}

	logger.Info("HTTP server initialized successfully")
	return server, nil
}

// setupGracefulShutdown sets up graceful shutdown handling
func setupGracefulShutdown(server *httpServer.Server, logger *logrus.Logger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Received shutdown signal, starting graceful shutdown...")

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown server
		if err := server.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("Error during server shutdown")
		}

		logger.Info("Server shutdown complete")
		os.Exit(0)
	}()
}

// getPort returns the port to listen on
func getPort(appConfig *config.AppConfig) int {
	return appConfig.Server.Port
}
