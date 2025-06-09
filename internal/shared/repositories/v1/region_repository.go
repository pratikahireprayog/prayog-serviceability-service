package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type regionRepository struct {
	db *gorm.DB
}

func NewRegionRepository(db *gorm.DB) RegionRepository {
	return &regionRepository{db: db}
}

func (r *regionRepository) GetByID(ctx context.Context, id string) (*models.Region, error) {
	var region models.Region
	regionID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid region ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("id = ? AND is_active = ?", regionID, true).
		Preload("Country").Preload("RegionType").First(&region).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("region %s not found", id)
		}
		return nil, fmt.Errorf("failed to get region: %w", err)
	}
	return &region, nil
}

func (r *regionRepository) GetByCode(ctx context.Context, code string) (*models.Region, error) {
	var region models.Region
	err := r.db.WithContext(ctx).Where("code = ? AND is_active = ?", code, true).
		Preload("Country").Preload("RegionType").First(&region).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("region %s not found", code)
		}
		return nil, fmt.Errorf("failed to get region: %w", err)
	}
	return &region, nil
}

func (r *regionRepository) GetAll(ctx context.Context, offset, limit int) ([]models.Region, int64, error) {
	var regions []models.Region
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.Region{}).Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count regions: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Where("is_active = ?", true).
		Preload("Country").Preload("RegionType").
		Offset(offset).Limit(limit).Find(&regions).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get regions: %w", err)
	}

	return regions, total, nil
}

func (r *regionRepository) GetByCountryID(ctx context.Context, countryID string) ([]models.Region, error) {
	var regions []models.Region
	countryUUID, err := uuid.Parse(countryID)
	if err != nil {
		return nil, fmt.Errorf("invalid country ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("country_id = ? AND is_active = ?", countryUUID, true).
		Preload("Country").Preload("RegionType").Find(&regions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get regions by country: %w", err)
	}
	return regions, nil
}

func (r *regionRepository) Create(ctx context.Context, region *models.Region) error {
	return r.db.WithContext(ctx).Create(region).Error
}

func (r *regionRepository) Update(ctx context.Context, id string, region *models.Region) error {
	regionID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid region ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", regionID).Save(region)
	if result.Error != nil {
		return fmt.Errorf("failed to update region: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("region not found")
	}
	return nil
}

func (r *regionRepository) Delete(ctx context.Context, id string) error {
	regionID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid region ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", regionID).Delete(&models.Region{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete region: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("region not found")
	}
	return nil
}
