package gorm

import (
	"context"
	"prayog-serviceability-service/pkg/domain"
	database "prayog-serviceability-service/pkg/infrastructure/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegionRepository implements the region repository interface
type RegionRepository struct {
	*BaseRepository[domain.AdministrativeRegion, database.AdministrativeRegion]
}

// NewRegionRepository creates a new region repository
func NewRegionRepository(db *database.DB) *RegionRepository {
	return &RegionRepository{
		BaseRepository: NewBaseRepository[domain.AdministrativeRegion, database.AdministrativeRegion](db),
	}
}

// GetByID retrieves a region by its ID
func (r *RegionRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.AdministrativeRegion, error) {
	return r.BaseRepository.GetByID(ctx, id, mapDatabaseRegionToDomain)
}

// GetByCode retrieves a region by its code
func (r *RegionRepository) GetByCode(ctx context.Context, code string) (domain.AdministrativeRegion, error) {
	return r.BaseRepository.FindOne(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", code)
	}, mapDatabaseRegionToDomain)
}

// GetByCountryID retrieves all regions for a country
func (r *RegionRepository) GetByCountryID(ctx context.Context, countryID uuid.UUID) ([]domain.AdministrativeRegion, error) {
	return r.BaseRepository.Query(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("country_id = ?", countryID)
	}, mapDatabaseRegionToDomain)
}

// List retrieves all regions
func (r *RegionRepository) List(ctx context.Context) ([]domain.AdministrativeRegion, error) {
	return r.BaseRepository.List(ctx, mapDatabaseRegionToDomain)
}

// Create creates a new region
func (r *RegionRepository) Create(ctx context.Context, region domain.AdministrativeRegion) error {
	return r.BaseRepository.Create(ctx, region, mapDomainRegionToDatabase)
}

// Update updates an existing region
func (r *RegionRepository) Update(ctx context.Context, region domain.AdministrativeRegion) error {
	return r.BaseRepository.Update(ctx, region, mapDomainRegionToDatabase)
}

// Delete deletes a region by its ID
func (r *RegionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id, database.AdministrativeRegion{})
}

// Helper functions to map between domain and database models

func mapDatabaseRegionToDomain(dbRegion *database.AdministrativeRegion) domain.AdministrativeRegion {
	return domain.AdministrativeRegion{
		ID:        dbRegion.ID,
		CountryID: dbRegion.CountryID,
		Name:      dbRegion.Name,
		Code:      dbRegion.Code,
		IsActive:  true, // Default value, could be stored in DB in future
		CreatedAt: dbRegion.CreatedAt,
		UpdatedAt: dbRegion.UpdatedAt,
	}
}

func mapDomainRegionToDatabase(domainRegion domain.AdministrativeRegion) database.AdministrativeRegion {
	return database.AdministrativeRegion{
		ID:        domainRegion.ID,
		CountryID: domainRegion.CountryID,
		Name:      domainRegion.Name,
		Code:      domainRegion.Code,
		CreatedAt: domainRegion.CreatedAt,
		UpdatedAt: domainRegion.UpdatedAt,
	}
}
