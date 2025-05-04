package database

import (
	"context"
	"fmt"
	"time"

	"prayog-serviceability-service/pkg/config"
	"prayog-serviceability-service/pkg/domain"

	"github.com/google/uuid"
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

// WithContext returns a GORM DB with the given context
func (db *DB) WithContext(ctx context.Context) *gorm.DB {
	return db.DB.WithContext(ctx)
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
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Code      string    `gorm:"size:10;uniqueIndex;not null"`
	Name      string    `gorm:"size:100;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// AdministrativeRegion represents a state/province/region within a country for GORM
type AdministrativeRegion struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CountryID uuid.UUID `gorm:"type:uuid;index;not null"`
	Code      string    `gorm:"size:20;index;not null"`
	Name      string    `gorm:"size:100;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// City represents a city entity for GORM
type City struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdministrativeRegionID uuid.UUID `gorm:"type:uuid;index;not null"`
	Name                   string    `gorm:"size:100;not null"`
	CreatedAt              time.Time `gorm:"autoCreateTime"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime"`
}

// Area represents a specific area or locality within a city for GORM
type Area struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CityID    uuid.UUID `gorm:"type:uuid;index;not null"`
	Name      string    `gorm:"size:100;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// PostalCode represents a postal/zip code for GORM
type PostalCode struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Code      string    `gorm:"size:20;uniqueIndex;not null"`
	AreaID    uuid.UUID `gorm:"type:uuid;index;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// OrderType represents different types of orders for GORM
type OrderType struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Code      string    `gorm:"size:20;uniqueIndex;not null"`
	Name      string    `gorm:"size:100;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// ServiceType represents different types of services offered for GORM
type ServiceType struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Code        string    `gorm:"size:20;uniqueIndex;not null"`
	Name        string    `gorm:"size:100;not null"`
	Description string    `gorm:"size:500"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// ServiceAvailability represents the availability of a service in a specific location for GORM
type ServiceAvailability struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	LocationType   string    `gorm:"size:20;not null;index:idx_location"`
	LocationID     uuid.UUID `gorm:"type:uuid;not null;index:idx_location"`
	OrderTypeID    uuid.UUID `gorm:"type:uuid;index;not null"`
	ServiceTypeID  uuid.UUID `gorm:"type:uuid;index;not null"`
	IsAvailable    bool      `gorm:"default:false;not null"`
	EffectiveFrom  time.Time `gorm:"not null"`
	EffectiveTo    time.Time `gorm:"not null"`
	AdditionalData string    `gorm:"type:jsonb"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}
