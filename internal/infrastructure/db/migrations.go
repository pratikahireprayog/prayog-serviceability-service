package database

import (
	"fmt"
	"log"
)

// RunDatabaseMigrations initializes and runs database migrations
func RunDatabaseMigrations(db *DB) error {
	log.Println("Running database migrations...")

	if err := db.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// CreateForeignKeys adds foreign key constraints after tables are created
func (db *DB) CreateForeignKeys() error {
	// Add foreign keys for AdministrativeRegion
	if err := db.DB.Exec("ALTER TABLE administrative_regions ADD CONSTRAINT fk_administrative_regions_country " +
		"FOREIGN KEY (country_id) REFERENCES countries(id) ON DELETE CASCADE").Error; err != nil {
		return err
	}

	// Add foreign keys for City
	if err := db.DB.Exec("ALTER TABLE cities ADD CONSTRAINT fk_cities_administrative_region " +
		"FOREIGN KEY (administrative_region_id) REFERENCES administrative_regions(id) ON DELETE CASCADE").Error; err != nil {
		return err
	}

	// Add foreign keys for Area
	if err := db.DB.Exec("ALTER TABLE areas ADD CONSTRAINT fk_areas_city " +
		"FOREIGN KEY (city_id) REFERENCES cities(id) ON DELETE CASCADE").Error; err != nil {
		return err
	}

	// Add foreign keys for PostalCode
	if err := db.DB.Exec("ALTER TABLE postal_codes ADD CONSTRAINT fk_postal_codes_area " +
		"FOREIGN KEY (area_id) REFERENCES areas(id) ON DELETE CASCADE").Error; err != nil {
		return err
	}

	// Add foreign keys for ServiceAvailability
	if err := db.DB.Exec("ALTER TABLE service_availabilities ADD CONSTRAINT fk_service_availabilities_order_type " +
		"FOREIGN KEY (order_type_id) REFERENCES order_types(id) ON DELETE CASCADE").Error; err != nil {
		return err
	}

	if err := db.DB.Exec("ALTER TABLE service_availabilities ADD CONSTRAINT fk_service_availabilities_service_type " +
		"FOREIGN KEY (service_type_id) REFERENCES service_types(id) ON DELETE CASCADE").Error; err != nil {
		return err
	}

	return nil
}

// SeedDatabase populates the database with initial data if needed
func (db *DB) SeedDatabase() error {
	// Check if countries table is empty
	var count int64
	if err := db.DB.Model(&Country{}).Count(&count).Error; err != nil {
		return err
	}

	// If no countries exist, add some sample data
	if count == 0 {
		// Sample countries
		countries := []Country{
			{Code: "IN", Name: "India"},
			{Code: "US", Name: "United States"},
			{Code: "CA", Name: "Canada"},
		}

		// Create sample countries
		if err := db.DB.Create(&countries).Error; err != nil {
			return err
		}

		// Sample order types
		orderTypes := []OrderType{
			{Code: "DELIVERY", Name: "Home Delivery"},
			{Code: "PICKUP", Name: "Store Pickup"},
		}

		// Create sample order types
		if err := db.DB.Create(&orderTypes).Error; err != nil {
			return err
		}

		// Sample service types
		serviceTypes := []ServiceType{
			{Code: "STANDARD", Name: "Standard Shipping", Description: "Regular delivery service with standard timeframes"},
			{Code: "EXPRESS", Name: "Express Shipping", Description: "Fast delivery service with premium rates"},
			{Code: "SAME_DAY", Name: "Same Day Delivery", Description: "Delivery on the same day of order"},
		}

		// Create sample service types
		if err := db.DB.Create(&serviceTypes).Error; err != nil {
			return err
		}
	}

	return nil
}
