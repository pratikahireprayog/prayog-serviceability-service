package gorm

import (
	"context"

	"github.com/prayog/serviceability/internal/domain"
	"github.com/prayog/serviceability/pkg/database"
)

// RegionRepository implements domain.RegionRepository using GORM
type RegionRepository struct {
	*Repository
}

// NewRegionRepository creates a new region repository
func NewRegionRepository(db *database.DB) domain.RegionRepository {
	return &RegionRepository{
		Repository: NewRepository(db),
	}
}

// GetByID retrieves a region by its ID
func (r *RegionRepository) GetByID(ctx context.Context, id uint) (*domain.AdministrativeRegion, error) {
	var region database.AdministrativeRegion
	if err := r.WithContext(ctx).First(&region, id).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabaseRegionToDomain(&region), nil
}

// GetByCode retrieves a region by its code
func (r *RegionRepository) GetByCode(ctx context.Context, code string) (*domain.AdministrativeRegion, error) {
	var region database.AdministrativeRegion
	if err := r.WithContext(ctx).Where("code = ?", code).First(&region).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabaseRegionToDomain(&region), nil
}

// GetByCountryID retrieves regions by country ID
func (r *RegionRepository) GetByCountryID(ctx context.Context, countryID uint) ([]*domain.AdministrativeRegion, error) {
	var regions []database.AdministrativeRegion
	if err := r.WithContext(ctx).Where("country_id = ?", countryID).Find(&regions).Error; err != nil {
		return nil, HandleError(err)
	}

	domainRegions := make([]*domain.AdministrativeRegion, len(regions))
	for i, region := range regions {
		domainRegions[i] = mapDatabaseRegionToDomain(&region)
	}
	return domainRegions, nil
}

// List retrieves all regions
func (r *RegionRepository) List(ctx context.Context) ([]*domain.AdministrativeRegion, error) {
	var regions []database.AdministrativeRegion
	if err := r.WithContext(ctx).Find(&regions).Error; err != nil {
		return nil, HandleError(err)
	}

	domainRegions := make([]*domain.AdministrativeRegion, len(regions))
	for i, region := range regions {
		domainRegions[i] = mapDatabaseRegionToDomain(&region)
	}
	return domainRegions, nil
}

// Create creates a new region
func (r *RegionRepository) Create(ctx context.Context, region *domain.AdministrativeRegion) error {
	dbRegion := mapDomainRegionToDatabase(region)
	if err := r.WithContext(ctx).Create(&dbRegion).Error; err != nil {
		return HandleError(err)
	}
	// Update the ID after creation
	region.ID = dbRegion.ID
	return nil
}

// Update updates an existing region
func (r *RegionRepository) Update(ctx context.Context, region *domain.AdministrativeRegion) error {
	dbRegion := mapDomainRegionToDatabase(region)
	result := r.WithContext(ctx).Save(&dbRegion)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a region by its ID
func (r *RegionRepository) Delete(ctx context.Context, id uint) error {
	result := r.WithContext(ctx).Delete(&database.AdministrativeRegion{}, id)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// mapDatabaseRegionToDomain maps a database region to a domain region
func mapDatabaseRegionToDomain(dbRegion *database.AdministrativeRegion) *domain.AdministrativeRegion {
	return &domain.AdministrativeRegion{
		ID:        dbRegion.ID,
		CountryID: dbRegion.CountryID,
		Code:      dbRegion.Code,
		Name:      dbRegion.Name,
		CreatedAt: dbRegion.CreatedAt,
		UpdatedAt: dbRegion.UpdatedAt,
	}
}

// mapDomainRegionToDatabase maps a domain region to a database region
func mapDomainRegionToDatabase(domainRegion *domain.AdministrativeRegion) database.AdministrativeRegion {
	return database.AdministrativeRegion{
		ID:        domainRegion.ID,
		CountryID: domainRegion.CountryID,
		Code:      domainRegion.Code,
		Name:      domainRegion.Name,
		CreatedAt: domainRegion.CreatedAt,
		UpdatedAt: domainRegion.UpdatedAt,
	}
}
