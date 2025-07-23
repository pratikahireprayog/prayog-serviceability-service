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

	if err := dm.migratePartnerAttributeModels(); err != nil {
		return fmt.Errorf("failed to migrate partner attribute models: %w", err)
	}

	// Add soft delete columns to all tables
	if err := dm.addSoftDeleteColumns(); err != nil {
		return fmt.Errorf("failed to add soft delete columns: %w", err)
	}

	// Create indexes for performance optimization
	if err := dm.createPerformanceIndexes(); err != nil {
		return fmt.Errorf("failed to create performance indexes: %w", err)
	}

	// Add is_active column to geo_locations table
	if err := dm.addIsActiveColumnToGeoLocations(); err != nil {
		return fmt.Errorf("failed to add is_active column to geo_locations: %w", err)
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

	// Migrate in dependency order: Country -> RegionType -> Region -> District -> City -> Area -> PostalCode -> GeoLocation
	models := []interface{}{
		&models.Country{},
		&models.RegionType{},
		&models.Region{},
		&models.District{},
		&models.City{},
		&models.Area{},
		&models.PostalCode{},
		&models.GeoLocation{},
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
		&models.NearestHubLocation{},
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

// migratePartnerAttributeModels migrates partner attribute entities
func (dm *DatabaseManager) migratePartnerAttributeModels() error {
	dm.logger.Info("Migrating partner attribute models...")

	models := []interface{}{
		&models.AttributeCategory{},
		&models.Attribute{},
		&models.PartnerAttributeMap{},
	}

	for _, model := range models {
		if err := dm.db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
		dm.logger.Infof("Successfully migrated model: %T", model)
	}

	dm.logger.Info("Partner attribute models migrated successfully")
	return nil
}

// addSoftDeleteColumns adds soft delete columns to all existing tables
func (dm *DatabaseManager) addSoftDeleteColumns() error {
	dm.logger.Info("Adding soft delete columns to all tables...")

	// List of all tables that need soft delete columns
	tables := []string{
		"country",
		"region_type",
		"region",
		"district",
		"city",
		"area",
		"postal_code",
		"hub",
		"hub_location_coverage",
		"hub_specification",
		"location_type",
		"location_alias",
		"partner_location_coverage",
		"attribute_category",
		"attribute",
		"partner_attribute_map",
	}

	for _, table := range tables {
		if err := dm.addSoftDeleteColumnsToTable(table); err != nil {
			return fmt.Errorf("failed to add soft delete columns to table %s: %w", table, err)
		}
		dm.logger.Infof("Added soft delete columns to table: %s", table)
	}

	// Update unique constraints to include is_deleted field
	if err := dm.updateUniqueConstraintsForSoftDelete(); err != nil {
		return fmt.Errorf("failed to update unique constraints: %w", err)
	}

	dm.logger.Info("Soft delete columns added successfully to all tables")
	return nil
}

// addSoftDeleteColumnsToTable adds deleted_at and is_deleted columns to a specific table
func (dm *DatabaseManager) addSoftDeleteColumnsToTable(tableName string) error {
	// Check if columns already exist
	var deletedAtExists, isDeletedExists bool

	// Check for deleted_at column
	deletedAtQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = ? AND column_name = 'deleted_at'
		)
	`
	if err := dm.db.Raw(deletedAtQuery, tableName).Scan(&deletedAtExists).Error; err != nil {
		return fmt.Errorf("failed to check deleted_at column existence: %w", err)
	}

	// Check for is_deleted column
	isDeletedQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = ? AND column_name = 'is_deleted'
		)
	`
	if err := dm.db.Raw(isDeletedQuery, tableName).Scan(&isDeletedExists).Error; err != nil {
		return fmt.Errorf("failed to check is_deleted column existence: %w", err)
	}

	// Add deleted_at column if it doesn't exist
	if !deletedAtExists {
		addDeletedAtSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN deleted_at TIMESTAMP NULL", tableName)
		if err := dm.db.Exec(addDeletedAtSQL).Error; err != nil {
			return fmt.Errorf("failed to add deleted_at column: %w", err)
		}
		dm.logger.Debugf("Added deleted_at column to %s", tableName)
	}

	// Add is_deleted column if it doesn't exist
	if !isDeletedExists {
		addIsDeletedSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE", tableName)
		if err := dm.db.Exec(addIsDeletedSQL).Error; err != nil {
			return fmt.Errorf("failed to add is_deleted column: %w", err)
		}
		dm.logger.Debugf("Added is_deleted column to %s", tableName)
	}

	// Create indexes on soft delete columns
	if err := dm.createSoftDeleteIndexes(tableName); err != nil {
		dm.logger.Warnf("Failed to create soft delete indexes for %s: %v", tableName, err)
		// Continue execution as indexes are not critical for functionality
	}

	return nil
}

// createSoftDeleteIndexes creates indexes on deleted_at and is_deleted columns
func (dm *DatabaseManager) createSoftDeleteIndexes(tableName string) error {
	// Create index on deleted_at column
	deletedAtIndexName := fmt.Sprintf("idx_%s_deleted_at", tableName)
	if err := dm.createIndex(tableName, []string{"deleted_at"}, deletedAtIndexName); err != nil {
		return fmt.Errorf("failed to create deleted_at index: %w", err)
	}

	// Create index on is_deleted column
	isDeletedIndexName := fmt.Sprintf("idx_%s_is_deleted", tableName)
	if err := dm.createIndex(tableName, []string{"is_deleted"}, isDeletedIndexName); err != nil {
		return fmt.Errorf("failed to create is_deleted index: %w", err)
	}

	return nil
}

// updateUniqueConstraintsForSoftDelete updates existing unique constraints to include is_deleted field
func (dm *DatabaseManager) updateUniqueConstraintsForSoftDelete() error {
	dm.logger.Info("Updating unique constraints to include is_deleted field...")

	// Define tables and their unique constraints that need updating
	constraintUpdates := []struct {
		table      string
		oldColumns []string
		newColumns []string
		constraint string
	}{
		{
			table:      "country",
			oldColumns: []string{"code"},
			newColumns: []string{"code", "is_deleted"},
			constraint: "country_code_key",
		},
		{
			table:      "region",
			oldColumns: []string{"code"},
			newColumns: []string{"code", "is_deleted"},
			constraint: "region_code_key",
		},
		{
			table:      "district",
			oldColumns: []string{"code"},
			newColumns: []string{"code", "is_deleted"},
			constraint: "district_code_key",
		},
		{
			table:      "city",
			oldColumns: []string{"code"},
			newColumns: []string{"code", "is_deleted"},
			constraint: "city_code_key",
		},
		{
			table:      "area",
			oldColumns: []string{"code"},
			newColumns: []string{"code", "is_deleted"},
			constraint: "area_code_key",
		},
		{
			table:      "postal_code",
			oldColumns: []string{"code"},
			newColumns: []string{"code", "is_deleted"},
			constraint: "postal_code_code_key",
		},
	}

	for _, update := range constraintUpdates {
		if err := dm.updateUniqueConstraint(update.table, update.constraint, update.oldColumns, update.newColumns); err != nil {
			dm.logger.Warnf("Failed to update unique constraint %s on table %s: %v", update.constraint, update.table, err)
			// Continue with other constraints even if one fails
		} else {
			dm.logger.Infof("Updated unique constraint %s on table %s", update.constraint, update.table)
		}
	}

	dm.logger.Info("Unique constraints update completed")
	return nil
}

// updateUniqueConstraint drops and recreates a unique constraint with additional columns
func (dm *DatabaseManager) updateUniqueConstraint(tableName, constraintName string, oldColumns, newColumns []string) error {
	// Check if constraint exists
	var exists bool
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.table_constraints 
			WHERE table_name = ? AND constraint_name = ? AND constraint_type = 'UNIQUE'
		)
	`
	if err := dm.db.Raw(checkQuery, tableName, constraintName).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check constraint existence: %w", err)
	}

	if !exists {
		dm.logger.Debugf("Constraint %s does not exist on table %s, skipping", constraintName, tableName)
		return nil
	}

	// Drop existing constraint
	dropSQL := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", tableName, constraintName)
	if err := dm.db.Exec(dropSQL).Error; err != nil {
		return fmt.Errorf("failed to drop constraint %s: %w", constraintName, err)
	}

	// Create new constraint with additional columns
	columnList := ""
	for i, col := range newColumns {
		if i > 0 {
			columnList += ", "
		}
		columnList += col
	}

	createSQL := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)", tableName, constraintName, columnList)
	if err := dm.db.Exec(createSQL).Error; err != nil {
		return fmt.Errorf("failed to create new constraint %s: %w", constraintName, err)
	}

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

		// Geo locations indexes for performance optimization
		{"geo_locations", []string{"postal_code"}, "idx_geo_locations_postal_code"},
		{"geo_locations", []string{"country_code"}, "idx_geo_locations_country_code"},
		{"geo_locations", []string{"postal_code", "country_code"}, "idx_geo_locations_postal_country"},
		{"geo_locations", []string{"deleted_at"}, "idx_geo_locations_deleted_at"},

		// Hub indexes
		{"hub", []string{"code"}, "idx_hub_code"},
		{"hub_location_coverage", []string{"hub_code", "location_scope"}, "idx_hub_coverage"},
		{"hub_specification", []string{"hub_code", "spec_code"}, "idx_hub_spec"},

		// Location management indexes
		{"location_alias", []string{"entity_type_code", "entity_id"}, "idx_alias_entity"},
		{"partner_location_coverage", []string{"partner_code", "location_scope"}, "idx_partner_coverage"},

		// Partner attribute indexes
		{"attribute_category", []string{"code"}, "idx_attribute_category_code"},
		{"attribute", []string{"category_id", "code"}, "idx_attribute_category_code"},
		{"attribute", []string{"code"}, "idx_attribute_code"},
		{"partner_attribute_map", []string{"partner_code", "attribute_id"}, "idx_partner_attr_map"},

		{"partner_attribute_map", []string{"partner_code"}, "idx_partner_attr_partner"},
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

// addIsActiveColumnToGeoLocations adds is_active column to geo_locations table
func (dm *DatabaseManager) addIsActiveColumnToGeoLocations() error {
	dm.logger.Info("Adding is_active column to geo_locations table...")

	// Check if is_active column already exists
	var isActiveExists bool
	isActiveQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'geo_locations' AND column_name = 'is_active'
		)
	`
	if err := dm.db.Raw(isActiveQuery).Scan(&isActiveExists).Error; err != nil {
		return fmt.Errorf("failed to check is_active column existence: %w", err)
	}

	// Add is_active column if it doesn't exist
	if !isActiveExists {
		addIsActiveSQL := "ALTER TABLE geo_locations ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE"
		if err := dm.db.Exec(addIsActiveSQL).Error; err != nil {
			return fmt.Errorf("failed to add is_active column: %w", err)
		}
		dm.logger.Info("Added is_active column to geo_locations table")

		// Create index on is_active column
		createIndexSQL := "CREATE INDEX idx_geo_locations_is_active ON geo_locations(is_active)"
		if err := dm.db.Exec(createIndexSQL).Error; err != nil {
			dm.logger.Warnf("Failed to create index on is_active column: %v", err)
			// Continue execution as index is not critical for functionality
		} else {
			dm.logger.Info("Created index on is_active column")
		}
	} else {
		dm.logger.Info("is_active column already exists in geo_locations table")
	}

	dm.logger.Info("Successfully added is_active column to geo_locations table")
	return nil
}

// DropAllTables drops all tables in the correct order to handle foreign key constraints
func (dm *DatabaseManager) DropAllTables() error {
	dm.logger.Warn("Dropping all tables...")
	return nil
}

// fixSchemaIssues fixes any database schema inconsistencies
func (dm *DatabaseManager) fixSchemaIssues() error {
	dm.logger.Info("Fixing database schema issues...")

	// Ensure attribute_code column exists in partner_attribute_map table
	if err := dm.ensureAttributeCodeColumn(); err != nil {
		return fmt.Errorf("failed to ensure attribute_code column: %w", err)
	}

	dm.logger.Info("Database schema issues fixed successfully")
	return nil
}

// ensureAttributeCodeColumn ensures the attribute_code column exists in partner_attribute_map table
func (dm *DatabaseManager) ensureAttributeCodeColumn() error {
	dm.logger.Info("Checking for attribute_code column in partner_attribute_map table...")

	// Check if column exists
	var columnExists bool
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'partner_attribute_map' AND column_name = 'attribute_code'
		)
	`
	if err := dm.db.Raw(checkQuery).Scan(&columnExists).Error; err != nil {
		return fmt.Errorf("failed to check for attribute_code column: %w", err)
	}

	if !columnExists {
		dm.logger.Info("attribute_code column missing, adding it...")

		// Add the column
		addQuery := "ALTER TABLE partner_attribute_map ADD COLUMN IF NOT EXISTS attribute_code VARCHAR(50) NOT NULL DEFAULT ''"
		if err := dm.db.Exec(addQuery).Error; err != nil {
			return fmt.Errorf("failed to add attribute_code column: %w", err)
		}

		// Create index for the new column
		indexQuery := "CREATE INDEX IF NOT EXISTS idx_partner_attr_code ON partner_attribute_map (attribute_code)"
		if err := dm.db.Exec(indexQuery).Error; err != nil {
			dm.logger.Warnf("Failed to create index for attribute_code: %v", err)
		}

		dm.logger.Info("Successfully added attribute_code column")
	} else {
		dm.logger.Info("attribute_code column already exists")
	}

	// Populate attribute_code for existing records where it's empty
	if err := dm.populateAttributeCodeForExistingRecords(); err != nil {
		return fmt.Errorf("failed to populate attribute_code for existing records: %w", err)
	}

	// Note: attribute_category_code column has been dropped manually via SQL

	return nil
}

// populateAttributeCodeForExistingRecords populates the attribute_code field for records where it's empty
func (dm *DatabaseManager) populateAttributeCodeForExistingRecords() error {
	dm.logger.Info("Populating attribute_code for existing partner_attribute_map records...")

	// Update records where attribute_code is empty by joining with the attribute table
	updateQuery := `
		UPDATE partner_attribute_map 
		SET attribute_code = attribute.code 
		FROM attribute 
		WHERE partner_attribute_map.attribute_id = attribute.id 
		AND (partner_attribute_map.attribute_code = '' OR partner_attribute_map.attribute_code IS NULL)
	`

	result := dm.db.Exec(updateQuery)
	if result.Error != nil {
		return fmt.Errorf("failed to populate attribute_code: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		dm.logger.Infof("Successfully populated attribute_code for %d records", result.RowsAffected)
	} else {
		dm.logger.Info("No records needed attribute_code population")
	}

	return nil
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
