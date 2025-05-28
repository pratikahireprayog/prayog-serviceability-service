package gorm

import (
	"context"

	"prayog-serviceability-service/pkg/domain"
	database "prayog-serviceability-service/pkg/infrastructure/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceTypeRepository implements the ServiceTypeRepository interface
type ServiceTypeRepository struct {
	*BaseRepository[domain.ServiceType, database.ServiceType]
}

// NewServiceTypeRepository creates a new service type repository
func NewServiceTypeRepository(db *database.DB) *ServiceTypeRepository {
	return &ServiceTypeRepository{
		BaseRepository: NewBaseRepository[domain.ServiceType, database.ServiceType](db),
	}
}

// GetByID retrieves a service type by its ID
func (r *ServiceTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.ServiceType, error) {
	return r.BaseRepository.GetByID(ctx, id, mapDatabaseServiceTypeToDomain)
}

// GetByCode retrieves a service type by its code
func (r *ServiceTypeRepository) GetByCode(ctx context.Context, code string) (domain.ServiceType, error) {
	return r.BaseRepository.FindOne(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", code)
	}, mapDatabaseServiceTypeToDomain)
}

// List retrieves all service types
func (r *ServiceTypeRepository) List(ctx context.Context) ([]domain.ServiceType, error) {
	return r.BaseRepository.List(ctx, mapDatabaseServiceTypeToDomain)
}

// Create creates a new service type
func (r *ServiceTypeRepository) Create(ctx context.Context, serviceType domain.ServiceType) error {
	return r.BaseRepository.Create(ctx, serviceType, mapDomainServiceTypeToDatabase)
}

// Update updates an existing service type
func (r *ServiceTypeRepository) Update(ctx context.Context, serviceType domain.ServiceType) error {
	return r.BaseRepository.Update(ctx, serviceType, mapDomainServiceTypeToDatabase)
}

// Delete deletes a service type by its ID
func (r *ServiceTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id, database.ServiceType{})
}

// Helper functions to map between domain and database models

func mapDatabaseServiceTypeToDomain(dbServiceType *database.ServiceType) domain.ServiceType {
	return domain.ServiceType{
		ID:          dbServiceType.ID,
		Name:        dbServiceType.Name,
		Code:        dbServiceType.Code,
		SLAHours:    0, // Simplified mapping, should include proper fields
		Description: dbServiceType.Description,
		IsActive:    true,
		CreatedAt:   dbServiceType.CreatedAt,
		UpdatedAt:   dbServiceType.UpdatedAt,
	}
}

func mapDomainServiceTypeToDatabase(domainServiceType domain.ServiceType) database.ServiceType {
	return database.ServiceType{
		ID:          domainServiceType.ID,
		Name:        domainServiceType.Name,
		Code:        domainServiceType.Code,
		Description: domainServiceType.Description,
		CreatedAt:   domainServiceType.CreatedAt,
		UpdatedAt:   domainServiceType.UpdatedAt,
	}
}
