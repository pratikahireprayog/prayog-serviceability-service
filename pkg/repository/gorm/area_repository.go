package gorm

import (
	"context"
	"fmt"

	"prayog-serviceability-service/pkg/domain"
	database "prayog-serviceability-service/pkg/infrastructure/db"
	"prayog-serviceability-service/pkg/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AreaRepository implements repository.AreaRepository using GORM
type AreaRepository struct {
	*repository.BaseRepository[domain.Area, database.Area]
}

// NewAreaRepository creates a new area repository
func NewAreaRepository(db *database.DB) *AreaRepository {
	return &AreaRepository{
		BaseRepository: repository.NewBaseRepository[domain.Area, database.Area](db),
	}
}

// GetByID retrieves an area by its ID
func (r *AreaRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Area, error) {
	return r.BaseRepository.GetByID(ctx, id, mapDatabaseAreaToDomain)
}

// GetByCityID retrieves all areas by city ID
func (r *AreaRepository) GetByCityID(ctx context.Context, cityID uuid.UUID) ([]domain.Area, error) {
	return r.BaseRepository.Query(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("city_id = ?", cityID)
	}, mapDatabaseAreaToDomain)
}

// GetActiveByName retrieves active areas by name (case insensitive)
// This is an example of a custom method specific to this repository
func (r *AreaRepository) GetActiveByName(ctx context.Context, name string) ([]domain.Area, error) {
	return r.BaseRepository.Query(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("LOWER(name) LIKE LOWER(?)", fmt.Sprintf("%%%s%%", name)).
			Where("is_active = ?", true)
	}, mapDatabaseAreaToDomain)
}

// List retrieves all areas
func (r *AreaRepository) List(ctx context.Context) ([]domain.Area, error) {
	return r.BaseRepository.List(ctx, mapDatabaseAreaToDomain)
}

// Create creates a new area
func (r *AreaRepository) Create(ctx context.Context, area domain.Area) error {
	return r.BaseRepository.Create(ctx, area, mapDomainAreaToDatabase)
}

// Update updates an existing area
func (r *AreaRepository) Update(ctx context.Context, area domain.Area) error {
	return r.BaseRepository.Update(ctx, area, mapDomainAreaToDatabase)
}

// Delete deletes an area by its ID
func (r *AreaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id, database.Area{})
}

// FindNearbyAreas finds areas within a certain distance of a given point
// This shows how to implement complex custom logic in a repository
func (r *AreaRepository) FindNearbyAreas(ctx context.Context, point domain.Point, radiusKm float64) ([]domain.Area, error) {
	// Example of a complex query that would be specific to this repository
	// In a real implementation, this might use PostGIS functions or other spatial queries

	// For now, we'll use a simple query as an example
	return r.BaseRepository.Query(ctx, func(db *gorm.DB) *gorm.DB {
		// This is a simplified example
		// A real implementation would use proper spatial queries
		return db.Where("is_active = ?", true).Limit(10)
	}, mapDatabaseAreaToDomain)
}

// Helper functions to map between domain and database models

func mapDatabaseAreaToDomain(dbArea *database.Area) domain.Area {
	return domain.Area{
		ID:          dbArea.ID,
		CityID:      dbArea.CityID,
		Name:        dbArea.Name,
		AreaType:    domain.AreaTypeLocality, // Default value, should be stored in DB
		GeoLocation: domain.Point{},          // Simplified, should map properly
		IsActive:    true,                    // Default value, should be stored in DB
		CreatedAt:   dbArea.CreatedAt,
		UpdatedAt:   dbArea.UpdatedAt,
	}
}

func mapDomainAreaToDatabase(domainArea domain.Area) database.Area {
	return database.Area{
		ID:        domainArea.ID,
		CityID:    domainArea.CityID,
		Name:      domainArea.Name,
		CreatedAt: domainArea.CreatedAt,
		UpdatedAt: domainArea.UpdatedAt,
	}
}
