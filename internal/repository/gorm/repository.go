package gorm

import (
	"context"
	"errors"

	"github.com/prayog/serviceability/pkg/database"
	"gorm.io/gorm"
)

// Repository is a base repository for GORM
type Repository struct {
	db *database.DB
}

// NewRepository creates a new repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// WithContext returns a GORM DB with the given context
func (r *Repository) WithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// Transaction executes a function in a transaction
func (r *Repository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// ErrNotFound is returned when a record is not found
var ErrNotFound = errors.New("record not found")

// ErrInvalidInput is returned when invalid input is provided
var ErrInvalidInput = errors.New("invalid input")

// ErrConflict is returned when there is a conflict (e.g., unique constraint violation)
var ErrConflict = errors.New("conflict error")

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}

// HandleError handles database errors
func HandleError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}

	// TODO: Add more specific error handling for database errors

	return err
}
