package database

import (
	"fmt"
	"time"

	"github.com/prayog/serviceability/internal/domain"
	"github.com/prayog/serviceability/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is a wrapper for gorm.DB
type DB struct {
	*gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config) (*DB, error) {
	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Open connection
	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	return &DB{db}, nil
}

// Ping checks the database connection
func (db *DB) Ping() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// RunMigrations runs all migrations
func (db *DB) RunMigrations() error {
	return db.DB.AutoMigrate(
		&domain.Country{},
		&domain.CountryAlias{},
		&domain.AdministrativeRegion{},
		&domain.AdministrativeRegionAlias{},
		&domain.City{},
		&domain.CityAlias{},
		&domain.Area{},
		&domain.AreaAlias{},
		&domain.PostalCode{},
		&domain.OrderType{},
		&domain.ServiceType{},
		&domain.ServiceAvailability{},
	)
}

// Country represents a country entity for GORM
type Country struct {
	ID        uint      `gorm:"primaryKey"`
	Code      string    `gorm:"size:10;uniqueIndex;not null"`
	Name      string    `gorm:"size:100;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// AdministrativeRegion represents a state/province/region within a country for GORM
type AdministrativeRegion struct {
	ID        uint      `gorm:"primaryKey"`
	CountryID uint      `gorm:"index;not null"`
	Code      string    `gorm:"size:20;index;not null"`
	Name      string    `gorm:"size:100;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// City represents a city entity for GORM
type City struct {
	ID                     uint      `gorm:"primaryKey"`
	AdministrativeRegionID uint      `gorm:"index;not null"`
	Name                   string    `gorm:"size:100;not null"`
	CreatedAt              time.Time `gorm:"autoCreateTime"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime"`
}

// Area represents a specific area or locality within a city for GORM
type Area struct {
	ID        uint      `gorm:"primaryKey"`
	CityID    uint      `gorm:"index;not null"`
	Name      string    `gorm:"size:100;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// PostalCode represents a postal/zip code for GORM
type PostalCode struct {
	ID        uint      `gorm:"primaryKey"`
	Code      string    `gorm:"size:20;uniqueIndex;not null"`
	AreaID    uint      `gorm:"index;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// OrderType represents different types of orders for GORM
type OrderType struct {
	ID        uint      `gorm:"primaryKey"`
	Code      string    `gorm:"size:20;uniqueIndex;not null"`
	Name      string    `gorm:"size:100;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// ServiceType represents different types of services offered for GORM
type ServiceType struct {
	ID          uint      `gorm:"primaryKey"`
	Code        string    `gorm:"size:20;uniqueIndex;not null"`
	Name        string    `gorm:"size:100;not null"`
	Description string    `gorm:"size:500"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// ServiceAvailability represents the availability of a service in a specific location for GORM
type ServiceAvailability struct {
	ID             uint      `gorm:"primaryKey"`
	LocationType   string    `gorm:"size:20;not null;index:idx_location"`
	LocationID     uint      `gorm:"not null;index:idx_location"`
	OrderTypeID    uint      `gorm:"index;not null"`
	ServiceTypeID  uint      `gorm:"index;not null"`
	IsAvailable    bool      `gorm:"default:false;not null"`
	EffectiveFrom  time.Time `gorm:"not null"`
	EffectiveTo    time.Time `gorm:"not null"`
	AdditionalData string    `gorm:"type:jsonb"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}
