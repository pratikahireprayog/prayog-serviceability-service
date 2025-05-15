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

	// Debug point 1: Configuration loaded
	log.Println("Configuration loaded successfully") // Set a breakpoint on this line

	// Initialize database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Debug point 2: Database initialized
	log.Println("Database connection established") // Set a breakpoint on this line

	// Initialize repositories
	repositories := repository.New(db)

	// Initialize services
	serviceabilityService := services.NewServiceabilityService(repositories.(*repository.RepositoryFactory))

	// Create HTTP server
	server := api.NewServer(serviceabilityService)

	// Start the server in a goroutine
	go func() {
		// Debug point 3: Server starting
		log.Println("About to start server on port 8080") // Set a breakpoint on this line

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

	// Debug point 4: Server shutting down
	log.Println("Server shutdown initiated") // Set a breakpoint on this line

	// Shutdown the server
	if err := server.Shutdown(); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	fmt.Println("Server gracefully stopped")
}
