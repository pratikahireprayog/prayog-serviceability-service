package gorm

import (
	"context"

	"github.com/prayog/serviceability/internal/domain"
	"github.com/prayog/serviceability/pkg/database"
)

// OrderTypeRepository implements domain.OrderTypeRepository using GORM
type OrderTypeRepository struct {
	*Repository
}

// NewOrderTypeRepository creates a new order type repository
func NewOrderTypeRepository(db *database.DB) domain.OrderTypeRepository {
	return &OrderTypeRepository{
		Repository: NewRepository(db),
	}
}

// GetByID retrieves an order type by its ID
func (r *OrderTypeRepository) GetByID(ctx context.Context, id uint) (*domain.OrderType, error) {
	var orderType database.OrderType
	if err := r.WithContext(ctx).First(&orderType, id).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabaseOrderTypeToDomain(&orderType), nil
}

// GetByCode retrieves an order type by its code
func (r *OrderTypeRepository) GetByCode(ctx context.Context, code string) (*domain.OrderType, error) {
	var orderType database.OrderType
	if err := r.WithContext(ctx).Where("code = ?", code).First(&orderType).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabaseOrderTypeToDomain(&orderType), nil
}

// List retrieves all order types
func (r *OrderTypeRepository) List(ctx context.Context) ([]*domain.OrderType, error) {
	var orderTypes []database.OrderType
	if err := r.WithContext(ctx).Find(&orderTypes).Error; err != nil {
		return nil, HandleError(err)
	}

	domainOrderTypes := make([]*domain.OrderType, len(orderTypes))
	for i, orderType := range orderTypes {
		domainOrderTypes[i] = mapDatabaseOrderTypeToDomain(&orderType)
	}
	return domainOrderTypes, nil
}

// Create creates a new order type
func (r *OrderTypeRepository) Create(ctx context.Context, orderType *domain.OrderType) error {
	dbOrderType := mapDomainOrderTypeToDatabase(orderType)
	if err := r.WithContext(ctx).Create(&dbOrderType).Error; err != nil {
		return HandleError(err)
	}
	// Update the ID after creation
	orderType.ID = dbOrderType.ID
	return nil
}

// Update updates an existing order type
func (r *OrderTypeRepository) Update(ctx context.Context, orderType *domain.OrderType) error {
	dbOrderType := mapDomainOrderTypeToDatabase(orderType)
	result := r.WithContext(ctx).Save(&dbOrderType)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes an order type by its ID
func (r *OrderTypeRepository) Delete(ctx context.Context, id uint) error {
	result := r.WithContext(ctx).Delete(&database.OrderType{}, id)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// mapDatabaseOrderTypeToDomain maps a database order type to a domain order type
func mapDatabaseOrderTypeToDomain(dbOrderType *database.OrderType) *domain.OrderType {
	return &domain.OrderType{
		ID:        dbOrderType.ID,
		Code:      dbOrderType.Code,
		Name:      dbOrderType.Name,
		CreatedAt: dbOrderType.CreatedAt,
		UpdatedAt: dbOrderType.UpdatedAt,
	}
}

// mapDomainOrderTypeToDatabase maps a domain order type to a database order type
func mapDomainOrderTypeToDatabase(domainOrderType *domain.OrderType) database.OrderType {
	return database.OrderType{
		ID:        domainOrderType.ID,
		Code:      domainOrderType.Code,
		Name:      domainOrderType.Name,
		CreatedAt: domainOrderType.CreatedAt,
		UpdatedAt: domainOrderType.UpdatedAt,
	}
}
