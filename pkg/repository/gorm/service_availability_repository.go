package gorm

import (
	"context"
	"time"

	"prayog-serviceability-service/pkg/domain"
	database "prayog-serviceability-service/pkg/infrastructure/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceAvailabilityRepository implements the ServiceAvailabilityRepository interface
type ServiceAvailabilityRepository struct {
	*BaseRepository[domain.ServiceAvailability, database.ServiceAvailability]
}

// NewServiceAvailabilityRepository creates a new service availability repository
func NewServiceAvailabilityRepository(db *database.DB) *ServiceAvailabilityRepository {
	return &ServiceAvailabilityRepository{
		BaseRepository: NewBaseRepository[domain.ServiceAvailability, database.ServiceAvailability](db),
	}
}

// GetByID retrieves a service availability by its ID
func (r *ServiceAvailabilityRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.ServiceAvailability, error) {
	return r.BaseRepository.GetByID(ctx, id, mapDatabaseServiceAvailabilityToDomain)
}

// GetByLocationAndTypes retrieves service availabilities by location and service types
func (r *ServiceAvailabilityRepository) GetByLocationAndTypes(
	ctx context.Context,
	locationType string,
	locationID uuid.UUID,
	orderTypeID, serviceTypeID uuid.UUID,
) ([]domain.ServiceAvailability, error) {
	return r.BaseRepository.Query(ctx, func(db *gorm.DB) *gorm.DB {
		query := db.Where("location_type = ?", locationType).
			Where("location_id = ?", locationID)

		if orderTypeID != uuid.Nil {
			query = query.Where("order_type_id = ?", orderTypeID)
		}

		if serviceTypeID != uuid.Nil {
			query = query.Where("service_type_id = ?", serviceTypeID)
		}

		return query
	}, mapDatabaseServiceAvailabilityToDomain)
}

// List retrieves all service availabilities
func (r *ServiceAvailabilityRepository) List(ctx context.Context) ([]domain.ServiceAvailability, error) {
	return r.BaseRepository.List(ctx, mapDatabaseServiceAvailabilityToDomain)
}

// Create creates a new service availability
func (r *ServiceAvailabilityRepository) Create(ctx context.Context, serviceAvailability domain.ServiceAvailability) error {
	return r.BaseRepository.Create(ctx, serviceAvailability, mapDomainServiceAvailabilityToDatabase)
}

// Update updates an existing service availability
func (r *ServiceAvailabilityRepository) Update(ctx context.Context, serviceAvailability domain.ServiceAvailability) error {
	return r.BaseRepository.Update(ctx, serviceAvailability, mapDomainServiceAvailabilityToDatabase)
}

// Delete deletes a service availability by its ID
func (r *ServiceAvailabilityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id, database.ServiceAvailability{})
}

// GetActiveServiceAvailabilities retrieves active service availabilities for the current time
// This shows how to implement a complex custom query using the BaseRepository
func (r *ServiceAvailabilityRepository) GetActiveServiceAvailabilities(
	ctx context.Context,
	locationType string,
	locationID uuid.UUID,
) ([]domain.ServiceAvailability, error) {
	return r.BaseRepository.Query(ctx, func(db *gorm.DB) *gorm.DB {
		now := time.Now()
		return db.Where("location_type = ?", locationType).
			Where("location_id = ?", locationID).
			Where("effective_from <= ?", now).
			Where("effective_to >= ?", now).
			Where("is_available = ?", true)
	}, mapDatabaseServiceAvailabilityToDomain)
}

// Helper functions to map between domain and database models

func mapDatabaseServiceAvailabilityToDomain(dbSA *database.ServiceAvailability) domain.ServiceAvailability {
	return domain.ServiceAvailability{
		ID:            dbSA.ID,
		LocationType:  dbSA.LocationType,
		LocationID:    dbSA.LocationID,
		OrderTypeID:   dbSA.OrderTypeID,
		ServiceTypeID: dbSA.ServiceTypeID,
		IsAvailable:   dbSA.IsAvailable,
		Constraints:   domain.Constraints{}, // Simplified mapping
		EffectiveFrom: dbSA.EffectiveFrom,
		EffectiveTo:   dbSA.EffectiveTo,
		IsActive:      true, // Default value
		CreatedAt:     dbSA.CreatedAt,
		UpdatedAt:     dbSA.UpdatedAt,
	}
}

func mapDomainServiceAvailabilityToDatabase(domainSA domain.ServiceAvailability) database.ServiceAvailability {
	constraints, _ := domainSA.Constraints.Value() // Simplified - should handle errors
	additionalData, _ := constraints.(string)      // Simplified - should do proper conversion

	return database.ServiceAvailability{
		ID:             domainSA.ID,
		LocationType:   domainSA.LocationType,
		LocationID:     domainSA.LocationID,
		OrderTypeID:    domainSA.OrderTypeID,
		ServiceTypeID:  domainSA.ServiceTypeID,
		IsAvailable:    domainSA.IsAvailable,
		EffectiveFrom:  domainSA.EffectiveFrom,
		EffectiveTo:    domainSA.EffectiveTo,
		AdditionalData: additionalData,
		CreatedAt:      domainSA.CreatedAt,
		UpdatedAt:      domainSA.UpdatedAt,
	}
}
