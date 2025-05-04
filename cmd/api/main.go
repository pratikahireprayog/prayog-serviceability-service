package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Define HTTP server
	mux := http.NewServeMux()

	// Add a basic health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Add a basic mock serviceability endpoint
	mux.HandleFunc("/api/v1/serviceability/check/", func(w http.ResponseWriter, r *http.Request) {
		postalCode := r.URL.Path[len("/api/v1/serviceability/check/"):]
		isServiceable := postalCode != "00000" // Just a placeholder check

		response := fmt.Sprintf(`{"postal_code":"%s","is_serviceable":%t}`, postalCode, isServiceable)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	})

	// Configure the HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server in a goroutine
	go func() {
		fmt.Println("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	fmt.Printf("Received signal %s, shutting down...\n", sig.String())

	// Shutdown the server
	if err := server.Close(); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	fmt.Println("Server gracefully stopped")
}
