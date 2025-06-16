package db

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"prayog-serviceability-service/internal/shared/config"
)

// DatabaseManager handles database connections and operations for serviceability service
type DatabaseManager struct {
	db     *gorm.DB
	config *config.AppConfig
	logger *logrus.Logger
}

// ConnectionPoolConfig holds database connection pool configuration
type ConnectionPoolConfig struct {
	MaxIdleConns    int           // Maximum number of idle connections
	MaxOpenConns    int           // Maximum number of open connections
	ConnMaxLifetime time.Duration // Maximum amount of time a connection may be reused
	ConnMaxIdleTime time.Duration // Maximum amount of time a connection may be idle
}

// NewDatabaseManager creates a new database manager instance
func NewDatabaseManager(appConfig *config.AppConfig, logger *logrus.Logger) (*DatabaseManager, error) {
	dbManager := &DatabaseManager{
		config: appConfig,
		logger: logger,
	}

	if err := dbManager.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return dbManager, nil
}

// Connect establishes database connection with optimized settings
func (dm *DatabaseManager) Connect() error {
	// Configure GORM logger based on log level
	var gormLogLevel logger.LogLevel
	switch dm.config.Log.Level {
	case "debug":
		gormLogLevel = logger.Info
	case "info":
		gormLogLevel = logger.Warn
	default:
		gormLogLevel = logger.Error
	}

	// Configure GORM with custom logger
	gormConfig := &gorm.Config{
		Logger: logger.New(
			dm.logger,
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  gormLogLevel,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Log database connection details before connecting
	dsn := dm.config.DB.DSN()
	connectionInfo := dm.config.DB.GetConnectionInfo()
	dm.logger.WithFields(logrus.Fields(connectionInfo)).Info("Attempting to connect to database")

	// Log SSL enforcement details
	if dm.config.DB.SSLMode == "require" || dm.config.DB.SSLMode == "verify-ca" || dm.config.DB.SSLMode == "verify-full" {
		dm.logger.WithField("ssl_mode", dm.config.DB.SSLMode).Info("SSL is enforced for database connection")
	}

	// Connect to PostgreSQL with UUID support
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	if err := dm.configureConnectionPool(db); err != nil {
		return fmt.Errorf("failed to configure connection pool: %w", err)
	}

	dm.db = db
	dm.logger.Info("Successfully connected to database")

	return nil
}

// configureConnectionPool sets up database connection pool settings
func (dm *DatabaseManager) configureConnectionPool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool with values from config
	sqlDB.SetMaxIdleConns(dm.config.DB.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dm.config.DB.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(dm.config.DB.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute) // Default idle time

	dm.logger.WithFields(logrus.Fields{
		"max_idle_conns":     dm.config.DB.MaxIdleConns,
		"max_open_conns":     dm.config.DB.MaxOpenConns,
		"conn_max_lifetime":  dm.config.DB.ConnMaxLifetime,
		"conn_max_idle_time": 5 * time.Minute,
	}).Info("Database connection pool configured")

	return nil
}

// GetDB returns the GORM database instance
func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.db
}

// Ping checks database connectivity
func (dm *DatabaseManager) Ping(ctx context.Context) error {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.PingContext(ctx)
}

// Close gracefully closes the database connection
func (dm *DatabaseManager) Close() error {
	if dm.db == nil {
		return nil
	}

	sqlDB, err := dm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	dm.logger.Info("Database connection closed successfully")
	return nil
}

// GetConnectionStats returns current database connection statistics
func (dm *DatabaseManager) GetConnectionStats(ctx context.Context) (map[string]interface{}, error) {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// HealthCheck performs a comprehensive database health check
func (dm *DatabaseManager) HealthCheck(ctx context.Context) error {
	// Check basic connectivity
	if err := dm.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check if we can execute a simple query
	var result int
	if err := dm.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database query returned unexpected result: %d", result)
	}

	dm.logger.Debug("Database health check passed")
	return nil
}

// AutoMigrate runs database migrations for provided models
func (dm *DatabaseManager) AutoMigrate(models ...interface{}) error {
	if err := dm.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run auto migrations: %w", err)
	}

	dm.logger.WithField("models_count", len(models)).Info("Database auto-migration completed")
	return nil
}

// WithTransaction executes the given function within a database transaction
func (dm *DatabaseManager) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return dm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
