package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	httpserver "github.com/prayog/serviceability/internal/transport/http"
	"github.com/prayog/serviceability/internal/usecase"
	"github.com/prayog/serviceability/pkg/config"
	"github.com/prayog/serviceability/pkg/database"
	"github.com/prayog/serviceability/pkg/logger"
)

const (
	version = "1.0.0"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	log, err := logger.NewLogger(&cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Prayog Serviceability Service...",
		zap.String("version", version),
		zap.String("environment", os.Getenv("APP_ENV")),
	)

	// Setup database
	db, err := database.SetupDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to setup database", zap.Error(err))
	}
	defer func() {
		log.Info("Closing database connection...")
		if err := db.Close(); err != nil {
			log.Error("Error closing database connection", zap.Error(err))
		}
	}()

	log.Info("Database setup completed successfully")

	// Initialize use cases
	usecases := usecase.NewUseCaseFactory()

	// Create HTTP server with the new config structure
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	serverConfig := httpserver.ServerConfig{
		Addr:         addr,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		Logger:       log.AsHTTPLogger(),
		RouterConfig: httpserver.RouterConfig{
			UseCases:     usecases,
			Version:      version,
			Logger:       log.AsHTTPLogger(),
			EnableCORS:   true,
			AuthEnabled:  false,
			APIKeys:      []string{}, // TODO: Load from config
			RateLimiting: true,
		},
	}

	server := httpserver.NewServer(serverConfig)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	log.Info("Server started successfully on " + addr)
	log.Info("API endpoints available at http://localhost" + addr + "/api/v1/serviceability/...")

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info("Received signal, shutting down...", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown failed", zap.Error(err))
	}

	log.Info("Server gracefully stopped")
}
