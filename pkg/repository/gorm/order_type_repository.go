package gorm

import (
	"context"

	"prayog-serviceability-service/pkg/domain"
	database "prayog-serviceability-service/pkg/infrastructure/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrderTypeRepository implements the OrderTypeRepository interface
type OrderTypeRepository struct {
	*BaseRepository[domain.OrderType, database.OrderType]
}

// NewOrderTypeRepository creates a new order type repository
func NewOrderTypeRepository(db *database.DB) *OrderTypeRepository {
	return &OrderTypeRepository{
		BaseRepository: NewBaseRepository[domain.OrderType, database.OrderType](db),
	}
}

// GetByID retrieves an order type by its ID
func (r *OrderTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.OrderType, error) {
	return r.BaseRepository.GetByID(ctx, id, mapDatabaseOrderTypeToDomain)
}

// GetByCode retrieves an order type by its code
func (r *OrderTypeRepository) GetByCode(ctx context.Context, code string) (domain.OrderType, error) {
	return r.BaseRepository.FindOne(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", code)
	}, mapDatabaseOrderTypeToDomain)
}

// List retrieves all order types
func (r *OrderTypeRepository) List(ctx context.Context) ([]domain.OrderType, error) {
	return r.BaseRepository.List(ctx, mapDatabaseOrderTypeToDomain)
}

// Create creates a new order type
func (r *OrderTypeRepository) Create(ctx context.Context, orderType domain.OrderType) error {
	return r.BaseRepository.Create(ctx, orderType, mapDomainOrderTypeToDatabase)
}

// Update updates an existing order type
func (r *OrderTypeRepository) Update(ctx context.Context, orderType domain.OrderType) error {
	return r.BaseRepository.Update(ctx, orderType, mapDomainOrderTypeToDatabase)
}

// Delete deletes an order type by its ID
func (r *OrderTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id, database.OrderType{})
}

// Helper functions to map between domain and database models

func mapDatabaseOrderTypeToDomain(dbOrderType *database.OrderType) domain.OrderType {
	return domain.OrderType{
		ID:          dbOrderType.ID,
		Name:        dbOrderType.Name,
		Code:        dbOrderType.Code,
		Description: "", // Simplified mapping, should include proper fields
		IsActive:    true,
		CreatedAt:   dbOrderType.CreatedAt,
		UpdatedAt:   dbOrderType.UpdatedAt,
	}
}

func mapDomainOrderTypeToDatabase(domainOrderType domain.OrderType) database.OrderType {
	return database.OrderType{
		ID:        domainOrderType.ID,
		Name:      domainOrderType.Name,
		Code:      domainOrderType.Code,
		CreatedAt: domainOrderType.CreatedAt,
		UpdatedAt: domainOrderType.UpdatedAt,
	}
}
