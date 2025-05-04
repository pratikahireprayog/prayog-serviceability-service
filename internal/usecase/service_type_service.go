package usecase

import (
	"context"

	"github.com/prayog/serviceability/internal/domain"
)

// serviceTypeService implements the ServiceTypeService interface
type serviceTypeService struct {
	repo domain.ServiceTypeRepository
	ctx  context.Context
}

// NewServiceTypeService creates a new service type service
func NewServiceTypeService(repo domain.ServiceTypeRepository) ServiceTypeService {
	return &serviceTypeService{
		repo: repo,
		ctx:  context.Background(),
	}
}

// GetByID retrieves a service type by its ID
func (s *serviceTypeService) GetByID(id uint) (*domain.ServiceType, error) {
	return s.repo.GetByID(s.ctx, id)
}

// GetByCode retrieves a service type by its code
func (s *serviceTypeService) GetByCode(code string) (*domain.ServiceType, error) {
	return s.repo.GetByCode(s.ctx, code)
}

// List retrieves all service types
func (s *serviceTypeService) List() ([]*domain.ServiceType, error) {
	return s.repo.List(s.ctx)
}

// Create creates a new service type
func (s *serviceTypeService) Create(serviceType *domain.ServiceType) error {
	return s.repo.Create(s.ctx, serviceType)
}

// Update updates an existing service type
func (s *serviceTypeService) Update(serviceType *domain.ServiceType) error {
	return s.repo.Update(s.ctx, serviceType)
}

// Delete deletes a service type by its ID
func (s *serviceTypeService) Delete(id uint) error {
	return s.repo.Delete(s.ctx, id)
}
