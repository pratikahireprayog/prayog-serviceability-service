package dbconn

import (
	"context"
	"database/sql"
)

// Database represents a database connection.
type Database interface {
	// GetDB returns the underlying SQL database.
	GetDB() *sql.DB
	// Close closes the database connection.
	Close() error
	// Ping pings the database to verify connection.
	Ping(ctx context.Context) error
}

// PGDatabase is a PostgreSQL implementation of the Database interface.
type PGDatabase struct {
	db *sql.DB
}

// NewDatabase creates a new database connection.
func NewDatabase() (Database, error) {
	// TODO: Implement actual database connection
	return &PGDatabase{
		db: nil, // Replace with actual DB connection
	}, nil
}

// GetDB returns the underlying SQL database.
func (p *PGDatabase) GetDB() *sql.DB {
	return p.db
}

// Close closes the database connection.
func (p *PGDatabase) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping pings the database to verify connection.
func (p *PGDatabase) Ping(ctx context.Context) error {
	if p.db != nil {
		return p.db.PingContext(ctx)
	}
	return nil
}
