package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"gorm.io/gorm"
)

type regionRepository struct {
	db *gorm.DB
}

func NewRegionRepository(db *gorm.DB) RegionRepository {
	return &regionRepository{db: db}
}

func (r *regionRepository) GetByID(ctx context.Context, id uint) (*models.Region, error) {
	var region models.Region
	err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&region).Error
	if err != nil {
		return nil, fmt.Errorf("region not found: %w", err)
	}
	return &region, nil
}

func (r *regionRepository) GetByCountryID(ctx context.Context, countryID uint) ([]models.Region, error) {
	var regions []models.Region
	err := r.db.WithContext(ctx).Where("country_id = ? AND is_active = ?", countryID, true).Find(&regions).Error
	return regions, err
}

func (r *regionRepository) GetByCode(ctx context.Context, code string, countryID uint) (*models.Region, error) {
	var region models.Region
	err := r.db.WithContext(ctx).Where("code = ? AND country_id = ? AND is_active = ?", code, countryID, true).First(&region).Error
	if err != nil {
		return nil, fmt.Errorf("region not found: %w", err)
	}
	return &region, nil
}

func (r *regionRepository) Create(ctx context.Context, region *models.Region) error {
	return r.db.WithContext(ctx).Create(region).Error
}

func (r *regionRepository) Update(ctx context.Context, region *models.Region) error {
	return r.db.WithContext(ctx).Save(region).Error
}

func (r *regionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Region{}, id).Error
}
