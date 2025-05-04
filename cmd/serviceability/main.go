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

	httpserver "github.com/prayog/serviceability/internal/delivery/http"
	"github.com/prayog/serviceability/internal/delivery/http/handlers"
	"github.com/prayog/serviceability/internal/delivery/http/middleware"
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

	// Setup router and handlers
	mux := http.NewServeMux()

	// Register handlers
	mux.Handle("/health", handlers.HealthHandler(version))

	// Setup middleware
	// TODO: Load API keys from configuration
	allowedOrigins := []string{"http://localhost:3000"}
	middlewareChain := middleware.Chain(
		mux,
		log.RecoveryMiddleware,
		log.HTTPMiddleware,
		middleware.CORS(allowedOrigins),
	)

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := httpserver.NewServer(
		addr,
		middlewareChain,
		httpserver.WithLogger(log.AsHTTPLogger()),
		httpserver.WithReadTimeout(cfg.Server.ReadTimeout),
		httpserver.WithWriteTimeout(cfg.Server.WriteTimeout),
		httpserver.WithIdleTimeout(cfg.Server.IdleTimeout),
	)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

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
