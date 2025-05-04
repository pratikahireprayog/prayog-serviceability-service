package repository

import (
	"context"
	"errors"
	database "prayog-serviceability-service/pkg/infrastructure/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Common repository errors
var (
	ErrNotFound = errors.New("entity not found")
	ErrDatabase = errors.New("database error")
	ErrInvalid  = errors.New("invalid entity")
	ErrConflict = errors.New("entity already exists")
)

// BaseRepository provides the common functionality for all repositories
type BaseRepository[T any, M any] struct {
	db *database.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T any, M any](db *database.DB) *BaseRepository[T, M] {
	return &BaseRepository[T, M]{db: db}
}

// DB returns the underlying database instance
func (r *BaseRepository[T, M]) DB() *database.DB {
	return r.db
}

// WithContext returns a GORM DB instance with the given context
func (r *BaseRepository[T, M]) WithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// GetByID retrieves an entity by its ID
func (r *BaseRepository[T, M]) GetByID(ctx context.Context, id uuid.UUID, mapFn func(*M) T) (T, error) {
	var result T
	var dbModel M

	if err := r.WithContext(ctx).First(&dbModel, id).Error; err != nil {
		return result, HandleError(err)
	}

	return mapFn(&dbModel), nil
}

// List retrieves all entities
func (r *BaseRepository[T, M]) List(ctx context.Context, mapFn func(*M) T) ([]T, error) {
	var dbModels []M
	if err := r.WithContext(ctx).Find(&dbModels).Error; err != nil {
		return nil, HandleError(err)
	}

	result := make([]T, len(dbModels))
	for i, model := range dbModels {
		result[i] = mapFn(&model)
	}
	return result, nil
}

// Query executes a custom query and returns the results
func (r *BaseRepository[T, M]) Query(ctx context.Context, query func(*gorm.DB) *gorm.DB, mapFn func(*M) T) ([]T, error) {
	var dbModels []M
	if err := query(r.WithContext(ctx)).Find(&dbModels).Error; err != nil {
		return nil, HandleError(err)
	}

	result := make([]T, len(dbModels))
	for i, model := range dbModels {
		result[i] = mapFn(&model)
	}
	return result, nil
}

// FindOne executes a custom query and returns a single result
func (r *BaseRepository[T, M]) FindOne(ctx context.Context, query func(*gorm.DB) *gorm.DB, mapFn func(*M) T) (T, error) {
	var result T
	var dbModel M

	if err := query(r.WithContext(ctx)).First(&dbModel).Error; err != nil {
		return result, HandleError(err)
	}

	return mapFn(&dbModel), nil
}

// Create creates a new entity
func (r *BaseRepository[T, M]) Create(ctx context.Context, entity T, mapToDbFn func(T) M) error {
	dbModel := mapToDbFn(entity)
	if err := r.WithContext(ctx).Create(&dbModel).Error; err != nil {
		return HandleError(err)
	}
	return nil
}

// Update updates an existing entity
func (r *BaseRepository[T, M]) Update(ctx context.Context, entity T, mapToDbFn func(T) M) error {
	dbModel := mapToDbFn(entity)
	result := r.WithContext(ctx).Save(&dbModel)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes an entity by its ID
func (r *BaseRepository[T, M]) Delete(ctx context.Context, id uuid.UUID, modelType M) error {
	result := r.WithContext(ctx).Delete(&modelType, id)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Transaction executes operations within a transaction
func (r *BaseRepository[T, M]) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := fn(tx); err != nil {
			return err
		}
		return nil
	})
}

// HandleError converts GORM errors to repository errors
func HandleError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, gorm.ErrInvalidTransaction) || errors.Is(err, gorm.ErrInvalidData) {
		return ErrInvalid
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrConflict
	}
	return ErrDatabase
}
