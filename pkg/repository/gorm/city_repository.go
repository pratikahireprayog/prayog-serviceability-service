package gorm

import (
	"context"
	"prayog-serviceability-service/pkg/domain"
	database "prayog-serviceability-service/pkg/infrastructure/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CityRepository implements city repository interface
type CityRepository struct {
	*BaseRepository[domain.City, database.City]
}

// NewCityRepository creates a new city repository
func NewCityRepository(db *database.DB) *CityRepository {
	return &CityRepository{
		BaseRepository: NewBaseRepository[domain.City, database.City](db),
	}
}

// GetByID retrieves a city by its ID
func (r *CityRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.City, error) {
	return r.BaseRepository.GetByID(ctx, id, mapDatabaseCityToDomain)
}

// GetByRegionID retrieves all cities for a region
func (r *CityRepository) GetByRegionID(ctx context.Context, regionID uuid.UUID) ([]domain.City, error) {
	return r.BaseRepository.Query(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("administrative_region_id = ?", regionID)
	}, mapDatabaseCityToDomain)
}

// List retrieves all cities
func (r *CityRepository) List(ctx context.Context) ([]domain.City, error) {
	return r.BaseRepository.List(ctx, mapDatabaseCityToDomain)
}

// Create creates a new city
func (r *CityRepository) Create(ctx context.Context, city domain.City) error {
	return r.BaseRepository.Create(ctx, city, mapDomainCityToDatabase)
}

// Update updates an existing city
func (r *CityRepository) Update(ctx context.Context, city domain.City) error {
	return r.BaseRepository.Update(ctx, city, mapDomainCityToDatabase)
}

// Delete deletes a city by its ID
func (r *CityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id, database.City{})
}

// Helper functions to map between domain and database models

func mapDatabaseCityToDomain(dbCity *database.City) domain.City {
	return domain.City{
		ID:        dbCity.ID,
		RegionID:  dbCity.AdministrativeRegionID,
		Name:      dbCity.Name,
		IsActive:  true,
		CreatedAt: dbCity.CreatedAt,
		UpdatedAt: dbCity.UpdatedAt,
	}
}

func mapDomainCityToDatabase(domainCity domain.City) database.City {
	return database.City{
		ID:                     domainCity.ID,
		AdministrativeRegionID: domainCity.RegionID,
		Name:                   domainCity.Name,
		CreatedAt:              domainCity.CreatedAt,
		UpdatedAt:              domainCity.UpdatedAt,
	}
}
