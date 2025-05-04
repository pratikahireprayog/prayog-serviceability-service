package usecase

import (
	"context"

	"github.com/prayog/serviceability/internal/domain"
)

// orderTypeService implements the OrderTypeService interface
type orderTypeService struct {
	repo domain.OrderTypeRepository
	ctx  context.Context
}

// NewOrderTypeService creates a new order type service
func NewOrderTypeService(repo domain.OrderTypeRepository) OrderTypeService {
	return &orderTypeService{
		repo: repo,
		ctx:  context.Background(),
	}
}

// GetByID retrieves an order type by its ID
func (s *orderTypeService) GetByID(id uint) (*domain.OrderType, error) {
	return s.repo.GetByID(s.ctx, id)
}

// GetByCode retrieves an order type by its code
func (s *orderTypeService) GetByCode(code string) (*domain.OrderType, error) {
	return s.repo.GetByCode(s.ctx, code)
}

// List retrieves all order types
func (s *orderTypeService) List() ([]*domain.OrderType, error) {
	return s.repo.List(s.ctx)
}

// Create creates a new order type
func (s *orderTypeService) Create(orderType *domain.OrderType) error {
	return s.repo.Create(s.ctx, orderType)
}

// Update updates an existing order type
func (s *orderTypeService) Update(orderType *domain.OrderType) error {
	return s.repo.Update(s.ctx, orderType)
}

// Delete deletes an order type by its ID
func (s *orderTypeService) Delete(id uint) error {
	return s.repo.Delete(s.ctx, id)
}
