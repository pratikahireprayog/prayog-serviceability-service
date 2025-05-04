package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Create new Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Prayog Serviceability Service",
	})

	// Add a basic health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Add a basic mock serviceability endpoint
	app.Get("/api/v1/serviceability/check/:postalCode", func(c *fiber.Ctx) error {
		postalCode := c.Params("postalCode")
		isServiceable := postalCode != "00000" // Just a placeholder check

		return c.JSON(fiber.Map{
			"postal_code":    postalCode,
			"is_serviceable": isServiceable,
		})
	})

	// Start the server in a goroutine
	go func() {
		fmt.Println("Starting server on :8080")
		if err := app.Listen(":8080"); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	fmt.Printf("Received signal %s, shutting down...\n", sig.String())

	// Shutdown the server
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	fmt.Println("Server gracefully stopped")
}
