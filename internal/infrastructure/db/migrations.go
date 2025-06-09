package db

import (
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"
)

// RunMigrations executes all database migrations in the correct order
func (dm *DatabaseManager) RunMigrations() error {
	dm.logger.Info("Starting database migrations...")

	// Enable UUID extension for PostgreSQL
	if err := dm.enableUUIDExtension(); err != nil {
		return fmt.Errorf("failed to enable UUID extension: %w", err)
	}

	// Run migrations in dependency order
	if err := dm.migrateGeographicalModels(); err != nil {
		return fmt.Errorf("failed to migrate geographical models: %w", err)
	}

	if err := dm.migrateHubModels(); err != nil {
		return fmt.Errorf("failed to migrate hub models: %w", err)
	}

	if err := dm.migrateLocationManagementModels(); err != nil {
		return fmt.Errorf("failed to migrate location management models: %w", err)
	}

	// Create indexes for performance optimization
	if err := dm.createPerformanceIndexes(); err != nil {
		return fmt.Errorf("failed to create performance indexes: %w", err)
	}

	dm.logger.Info("Database migrations completed successfully")
	return nil
}

// enableUUIDExtension enables the uuid-ossp extension for PostgreSQL
func (dm *DatabaseManager) enableUUIDExtension() error {
	dm.logger.Info("Enabling UUID extension...")

	if err := dm.db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	dm.logger.Info("UUID extension enabled successfully")
	return nil
}

// migrateGeographicalModels migrates core geographical entities in dependency order
func (dm *DatabaseManager) migrateGeographicalModels() error {
	dm.logger.Info("Migrating geographical models...")

	// Migrate in dependency order: Country -> RegionType -> Region -> District -> City -> Area -> PostalCode
	models := []interface{}{
		&models.Country{},
		&models.RegionType{},
		&models.Region{},
		&models.District{},
		&models.City{},
		&models.Area{},
		&models.PostalCode{},
	}

	for _, model := range models {
		if err := dm.db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
		dm.logger.Infof("Successfully migrated model: %T", model)
	}

	dm.logger.Info("Geographical models migrated successfully")
	return nil
}

// migrateHubModels migrates hub-related entities
func (dm *DatabaseManager) migrateHubModels() error {
	dm.logger.Info("Migrating hub models...")

	models := []interface{}{
		&models.Hub{},
		&models.HubLocationCoverage{},
		&models.HubSpecification{},
	}

	for _, model := range models {
		if err := dm.db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
		dm.logger.Infof("Successfully migrated model: %T", model)
	}

	dm.logger.Info("Hub models migrated successfully")
	return nil
}

// migrateLocationManagementModels migrates location management entities
func (dm *DatabaseManager) migrateLocationManagementModels() error {
	dm.logger.Info("Migrating location management models...")

	models := []interface{}{
		&models.LocationType{},
		&models.LocationAlias{},
		&models.PartnerLocationCoverage{},
	}

	for _, model := range models {
		if err := dm.db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
		dm.logger.Infof("Successfully migrated model: %T", model)
	}

	dm.logger.Info("Location management models migrated successfully")
	return nil
}

// createPerformanceIndexes creates additional indexes for query optimization
func (dm *DatabaseManager) createPerformanceIndexes() error {
	dm.logger.Info("Creating performance indexes...")

	indexes := []struct {
		table   string
		columns []string
		name    string
	}{
		// Geographical hierarchy indexes
		{"country", []string{"code"}, "idx_country_code"},
		{"region_type", []string{"code"}, "idx_region_type_code"},
		{"region", []string{"country_code", "region_type_code"}, "idx_region_country_type"},
		{"district", []string{"region_code", "country_code"}, "idx_district_region_country"},
		{"city", []string{"region_code", "country_code", "district_code"}, "idx_city_hierarchy"},
		{"area", []string{"city_code"}, "idx_area_city"},
		{"postal_code", []string{"country_code", "region_code", "city_code", "area_code"}, "idx_postal_hierarchy"},
		{"postal_code", []string{"location_scope"}, "idx_postal_scope"},

		// Hub indexes
		{"hub", []string{"code"}, "idx_hub_code"},
		{"hub_location_coverage", []string{"hub_code", "location_scope"}, "idx_hub_coverage"},
		{"hub_specification", []string{"hub_code", "spec_code"}, "idx_hub_spec"},

		// Location management indexes
		{"location_alias", []string{"entity_type_code", "entity_id"}, "idx_alias_entity"},
		{"partner_location_coverage", []string{"partner_code", "location_scope"}, "idx_partner_coverage"},
	}

	for _, idx := range indexes {
		if err := dm.createIndex(idx.table, idx.columns, idx.name); err != nil {
			dm.logger.Warnf("Failed to create index %s: %v", idx.name, err)
			// Continue with other indexes even if one fails
		} else {
			dm.logger.Infof("Created index: %s", idx.name)
		}
	}

	dm.logger.Info("Performance indexes creation completed")
	return nil
}

// createIndex creates a database index if it doesn't exist
func (dm *DatabaseManager) createIndex(table string, columns []string, indexName string) error {
	// Check if index already exists
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE tablename = ? AND indexname = ?
		)
	`
	if err := dm.db.Raw(query, table, indexName).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	if exists {
		dm.logger.Debugf("Index %s already exists, skipping", indexName)
		return nil
	}

	// Create the index
	columnList := ""
	for i, col := range columns {
		if i > 0 {
			columnList += ", "
		}
		columnList += col
	}

	createIndexSQL := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", indexName, table, columnList)
	if err := dm.db.Exec(createIndexSQL).Error; err != nil {
		return fmt.Errorf("failed to create index %s: %w", indexName, err)
	}

	return nil
}

// SeedDatabase populates the database with initial reference data
func (dm *DatabaseManager) SeedDatabase() error {
	dm.logger.Info("Starting database seeding...")

	// Seed reference data in dependency order
	if err := dm.seedRegionTypes(); err != nil {
		return fmt.Errorf("failed to seed region types: %w", err)
	}

	if err := dm.seedLocationTypes(); err != nil {
		return fmt.Errorf("failed to seed location types: %w", err)
	}

	dm.logger.Info("Database seeding completed successfully")
	return nil
}

// seedRegionTypes seeds the region_type table with common region types
func (dm *DatabaseManager) seedRegionTypes() error {
	dm.logger.Info("Seeding region types...")

	regionTypes := []models.RegionType{
		{Code: "state", Name: "State", Description: stringPtr("Administrative state or province"), IsActive: true},
		{Code: "province", Name: "Province", Description: stringPtr("Administrative province"), IsActive: true},
		{Code: "territory", Name: "Territory", Description: stringPtr("Administrative territory"), IsActive: true},
		{Code: "region", Name: "Region", Description: stringPtr("Administrative region"), IsActive: true},
		{Code: "division", Name: "Division", Description: stringPtr("Administrative division"), IsActive: true},
	}

	for _, rt := range regionTypes {
		// Use FirstOrCreate to avoid duplicates
		var existing models.RegionType
		result := dm.db.Where("code = ?", rt.Code).FirstOrCreate(&existing, rt)
		if result.Error != nil {
			return fmt.Errorf("failed to seed region type %s: %w", rt.Code, result.Error)
		}

		if result.RowsAffected > 0 {
			dm.logger.Infof("Seeded region type: %s", rt.Code)
		} else {
			dm.logger.Debugf("Region type already exists: %s", rt.Code)
		}
	}

	dm.logger.Info("Region types seeded successfully")
	return nil
}

// seedLocationTypes seeds the location_type table with common location types
func (dm *DatabaseManager) seedLocationTypes() error {
	dm.logger.Info("Seeding location types...")

	locationTypes := []models.LocationType{
		{Code: "country", Description: stringPtr("Country level location")},
		{Code: "region", Description: stringPtr("Region/State level location")},
		{Code: "district", Description: stringPtr("District level location")},
		{Code: "city", Description: stringPtr("City level location")},
		{Code: "area", Description: stringPtr("Area level location")},
		{Code: "postal_code", Description: stringPtr("Postal code level location")},
	}

	for _, lt := range locationTypes {
		// Use FirstOrCreate to avoid duplicates
		var existing models.LocationType
		result := dm.db.Where("code = ?", lt.Code).FirstOrCreate(&existing, lt)
		if result.Error != nil {
			return fmt.Errorf("failed to seed location type %s: %w", lt.Code, result.Error)
		}

		if result.RowsAffected > 0 {
			dm.logger.Infof("Seeded location type: %s", lt.Code)
		} else {
			dm.logger.Debugf("Location type already exists: %s", lt.Code)
		}
	}

	dm.logger.Info("Location types seeded successfully")
	return nil
}

// DropAllTables drops all tables in the correct order to handle foreign key constraints
func (dm *DatabaseManager) DropAllTables() error {
	dm.logger.Warn("Dropping all tables...")

	// Drop tables in reverse dependency order
	tables := []string{
		"partner_location_coverage",
		"location_alias",
		"location_type",
		"hub_specification",
		"hub_location_coverage",
		"hub",
		"postal_code",
		"area",
		"city",
		"district",
		"region",
		"region_type",
		"country",
	}

	for _, table := range tables {
		if err := dm.db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE").Error; err != nil {
			dm.logger.Errorf("Failed to drop table %s: %v", table, err)
			// Continue with other tables
		} else {
			dm.logger.Infof("Dropped table: %s", table)
		}
	}

	dm.logger.Info("All tables dropped successfully")
	return nil
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
