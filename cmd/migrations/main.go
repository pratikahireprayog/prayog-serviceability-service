package main

import (
	"flag"
	"log"
	"os"

	"prayog-serviceability-service/pkg/config"
	database "prayog-serviceability-service/pkg/infrastructure/db"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	database, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		if err := database.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Test connection
	if err := database.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Handle drop tables if requested
	if *drop {
		log.Println("WARNING: Dropping all tables...")
		if err := dropAllTables(database); err != nil {
			log.Fatalf("Failed to drop tables: %v", err)
		}
		log.Println("All tables dropped successfully")
	}

	// Run migrations if requested
	if *migrate {
		log.Println("Running database migrations...")
		if err := database.RunMigrations(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Migrations completed successfully")
	}

	// Seed database if requested
	if *seed {
		log.Println("Seeding database...")
		if err := database.SeedDatabase(); err != nil {
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

// dropAllTables drops all tables in the database
func dropAllTables(database *database.DB) error {
	// Drop all tables in the correct order to avoid foreign key constraints
	tables := []string{
		"service_availabilities",
		"service_types",
		"order_types",
		"postal_codes",
		"areas",
		"cities",
		"administrative_regions",
		"countries",
	}

	for _, table := range tables {
		if err := database.DB.Exec("DROP TABLE IF EXISTS " + table + " CASCADE").Error; err != nil {
			return err
		}
		log.Printf("Dropped table %s", table)
	}

	return nil
}
