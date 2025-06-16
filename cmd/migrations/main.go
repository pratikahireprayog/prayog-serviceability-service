package main

import (
	"context"
	"flag"
	"log"
	"os"

	"prayog-serviceability-service/internal/infrastructure/db"
	"prayog-serviceability-service/internal/shared/config"

	"github.com/sirupsen/logrus"
)

func main() {
	log.SetPrefix("[MIGRATION] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Parse command line flags
	migrate := flag.Bool("migrate", false, "Run database migrations")
	seed := flag.Bool("seed", false, "Seed the database with initial data")
	drop := flag.Bool("drop", false, "Drop all tables before running migrations (WARNING: destructive operation)")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadAppConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Connect to database
	dbManager, err := db.NewDatabaseManager(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		if err := dbManager.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Test connection
	ctx := context.Background()
	if err := dbManager.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Handle drop tables if requested
	if *drop {
		log.Println("WARNING: Dropping all tables...")
		if err := dbManager.DropAllTables(); err != nil {
			log.Fatalf("Failed to drop tables: %v", err)
		}
		log.Println("All tables dropped successfully")
	}

	// Run migrations if requested
	if *migrate {
		log.Println("Running database migrations...")
		if err := dbManager.RunMigrations(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Migrations completed successfully")
	}

	// Seed database if requested
	if *seed {
		log.Println("Seeding database...")
		if err := dbManager.SeedDatabase(); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
		log.Println("Database seeded successfully")
	}

	// If no flag is provided, show usage
	if !*migrate && !*seed && !*drop {
		log.Println("No action specified. Use one or more of the following flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Println("All operations completed successfully")
}
