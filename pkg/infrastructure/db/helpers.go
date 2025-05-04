package database

import (
	"context"
	"log"

	"prayog-serviceability-service/pkg/config"
)

// SetupDatabase initializes the database connection and runs migrations
func SetupDatabase(cfg *config.Config) (*DB, error) {
	// Connect to database
	db, err := NewDatabase(cfg)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to database")

	// Run migrations
	if err := RunDatabaseMigrations(db); err != nil {
		return nil, err
	}

	// Create foreign key constraints
	if err := db.CreateForeignKeys(); err != nil {
		log.Printf("Warning: Failed to create foreign keys: %v", err)
	}

	// Seed database with initial data if needed
	if err := db.SeedDatabase(); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}

	return db, nil
}

// TestConnection tests the database connection
func TestConnection(db *DB) error {
	// Execute a simple query to test connection
	var result int
	err := db.DB.Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		return err
	}

	if result != 1 {
		log.Println("Warning: Database connection test returned unexpected result")
	}

	return nil
}

// WithTransaction executes a function within a database transaction
func (db *DB) WithTransaction(ctx context.Context, fn func(tx *DB) error) error {
	// Start transaction
	tx := db.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Create a new DB instance with the transaction
	txDB := &DB{tx}

	// Execute the function
	err := fn(txDB)
	if err != nil {
		// Rollback the transaction if there's an error
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}
