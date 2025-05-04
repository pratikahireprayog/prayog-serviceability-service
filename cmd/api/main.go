package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"prayog-serviceability-service/api"
	"prayog-serviceability-service/pkg/config"
	database "prayog-serviceability-service/pkg/infrastructure/db"
	"prayog-serviceability-service/pkg/repository"
	"prayog-serviceability-service/pkg/services"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	repositories := repository.New(db)

	// Initialize services
	serviceabilityService := services.NewServiceabilityService(repositories)

	// Create HTTP server
	server := api.NewServer(serviceabilityService)

	// Start the server in a goroutine
	go func() {
		fmt.Println("Starting server on :8080")
		if err := server.Listen(":8080"); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	fmt.Printf("Received signal %s, shutting down...\n", sig.String())

	// Shutdown the server
	if err := server.Shutdown(); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	fmt.Println("Server gracefully stopped")
}
