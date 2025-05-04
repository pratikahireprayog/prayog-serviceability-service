package gorm

import (
	"context"

	"prayog-serviceability-service/internal/domain"
	"prayog-serviceability-service/pkg/database"
)

// ServiceTypeRepository implements domain.ServiceTypeRepository using GORM
type ServiceTypeRepository struct {
	*Repository
}

// NewServiceTypeRepository creates a new service type repository
func NewServiceTypeRepository(db *database.DB) domain.ServiceTypeRepository {
	return &ServiceTypeRepository{
		Repository: NewRepository(db),
	}
}

// GetByID retrieves a service type by its ID
func (r *ServiceTypeRepository) GetByID(ctx context.Context, id uint) (*domain.ServiceType, error) {
	var serviceType database.ServiceType
	if err := r.WithContext(ctx).First(&serviceType, id).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabaseServiceTypeToDomain(&serviceType), nil
}

// GetByCode retrieves a service type by its code
func (r *ServiceTypeRepository) GetByCode(ctx context.Context, code string) (*domain.ServiceType, error) {
	var serviceType database.ServiceType
	if err := r.WithContext(ctx).Where("code = ?", code).First(&serviceType).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabaseServiceTypeToDomain(&serviceType), nil
}

// List retrieves all service types
func (r *ServiceTypeRepository) List(ctx context.Context) ([]*domain.ServiceType, error) {
	var serviceTypes []database.ServiceType
	if err := r.WithContext(ctx).Find(&serviceTypes).Error; err != nil {
		return nil, HandleError(err)
	}

	domainServiceTypes := make([]*domain.ServiceType, len(serviceTypes))
	for i, serviceType := range serviceTypes {
		domainServiceTypes[i] = mapDatabaseServiceTypeToDomain(&serviceType)
	}
	return domainServiceTypes, nil
}

// Create creates a new service type
func (r *ServiceTypeRepository) Create(ctx context.Context, serviceType *domain.ServiceType) error {
	dbServiceType := mapDomainServiceTypeToDatabase(serviceType)
	if err := r.WithContext(ctx).Create(&dbServiceType).Error; err != nil {
		return HandleError(err)
	}
	// Update the ID after creation
	serviceType.ID = dbServiceType.ID
	return nil
}

// Update updates an existing service type
func (r *ServiceTypeRepository) Update(ctx context.Context, serviceType *domain.ServiceType) error {
	dbServiceType := mapDomainServiceTypeToDatabase(serviceType)
	result := r.WithContext(ctx).Save(&dbServiceType)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a service type by its ID
func (r *ServiceTypeRepository) Delete(ctx context.Context, id uint) error {
	result := r.WithContext(ctx).Delete(&database.ServiceType{}, id)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// mapDatabaseServiceTypeToDomain maps a database service type to a domain service type
func mapDatabaseServiceTypeToDomain(dbServiceType *database.ServiceType) *domain.ServiceType {
	return &domain.ServiceType{
		ID:          dbServiceType.ID,
		Code:        dbServiceType.Code,
		Name:        dbServiceType.Name,
		Description: dbServiceType.Description,
		CreatedAt:   dbServiceType.CreatedAt,
		UpdatedAt:   dbServiceType.UpdatedAt,
	}
}

// mapDomainServiceTypeToDatabase maps a domain service type to a database service type
func mapDomainServiceTypeToDatabase(domainServiceType *domain.ServiceType) database.ServiceType {
	return database.ServiceType{
		ID:          domainServiceType.ID,
		Code:        domainServiceType.Code,
		Name:        domainServiceType.Name,
		Description: domainServiceType.Description,
		CreatedAt:   domainServiceType.CreatedAt,
		UpdatedAt:   domainServiceType.UpdatedAt,
	}
}
