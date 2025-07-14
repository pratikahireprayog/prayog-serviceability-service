package repositories

import (
	"context"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"

	"gorm.io/gorm"
)

type NearestHubLocationRepository interface {
	Create(ctx context.Context, loc *models.NearestHubLocation) error
	GetByPostalCode(ctx context.Context, postalCode int) (*models.NearestHubLocation, error)
	GetByFilters(ctx context.Context, filters *dtos.NearestHubLocationFilters) ([]models.NearestHubLocation, error)
	Update(ctx context.Context, postalCode int, update *dtos.UpdateNearestHubLocationRequest) error
	Delete(ctx context.Context, postalCode int) error
}

type nearestHubLocationRepository struct {
	db *gorm.DB
}

func NewNearestHubLocationRepository(db *gorm.DB) NearestHubLocationRepository {
	return &nearestHubLocationRepository{db: db}
}

func (r *nearestHubLocationRepository) Create(ctx context.Context, loc *models.NearestHubLocation) error {
	return r.db.WithContext(ctx).Create(loc).Error
}

func (r *nearestHubLocationRepository) GetByPostalCode(ctx context.Context, postalCode int) (*models.NearestHubLocation, error) {
	var loc models.NearestHubLocation
	err := r.db.WithContext(ctx).Where("postal_code = ?", postalCode).First(&loc).Error
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (r *nearestHubLocationRepository) GetByFilters(ctx context.Context, filters *dtos.NearestHubLocationFilters) ([]models.NearestHubLocation, error) {
	var locs []models.NearestHubLocation
	query := r.db.WithContext(ctx).Model(&models.NearestHubLocation{})
	if len(filters.PostalCodes) > 0 {
		query = query.Where("postal_code IN ?", filters.PostalCodes)
	}
	if len(filters.InternationalHubPostalCodes) > 0 {
		query = query.Where("international_hub_postal_code IN ?", filters.InternationalHubPostalCodes)
	}
	if len(filters.InternationalHubCityCodes) > 0 {
		query = query.Where("international_hub_city_code IN ?", filters.InternationalHubCityCodes)
	}
	err := query.Find(&locs).Error
	return locs, err
}

func (r *nearestHubLocationRepository) Update(ctx context.Context, postalCode int, update *dtos.UpdateNearestHubLocationRequest) error {
	return r.db.WithContext(ctx).Model(&models.NearestHubLocation{}).Where("postal_code = ?", postalCode).Updates(update).Error
}

func (r *nearestHubLocationRepository) Delete(ctx context.Context, postalCode int) error {
	return r.db.WithContext(ctx).Where("postal_code = ?", postalCode).Delete(&models.NearestHubLocation{}).Error
}
