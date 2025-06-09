package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"gorm.io/gorm"
)

type regionTypeRepository struct {
	db *gorm.DB
}

func NewRegionTypeRepository(db *gorm.DB) RegionTypeRepository {
	return &regionTypeRepository{db: db}
}

func (r *regionTypeRepository) GetByCode(ctx context.Context, code string) (*models.RegionType, error) {
	var regionType models.RegionType
	err := r.db.WithContext(ctx).Where("code = ? AND is_active = ?", code, true).First(&regionType).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("region type %s not found", code)
		}
		return nil, fmt.Errorf("failed to get region type: %w", err)
	}
	return &regionType, nil
}

func (r *regionTypeRepository) GetAll(ctx context.Context, offset, limit int) ([]models.RegionType, int64, error) {
	var regionTypes []models.RegionType
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.RegionType{}).Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count region types: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Where("is_active = ?", true).
		Offset(offset).Limit(limit).Find(&regionTypes).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get region types: %w", err)
	}

	return regionTypes, total, nil
}

func (r *regionTypeRepository) Create(ctx context.Context, regionType *models.RegionType) error {
	err := r.db.WithContext(ctx).Create(regionType).Error
	if err != nil {
		return fmt.Errorf("failed to create region type: %w", err)
	}
	return nil
}

func (r *regionTypeRepository) Update(ctx context.Context, code string, regionType *models.RegionType) error {
	result := r.db.WithContext(ctx).Where("code = ?", code).Save(regionType)
	if result.Error != nil {
		return fmt.Errorf("failed to update region type: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("region type not found")
	}
	return nil
}

func (r *regionTypeRepository) Delete(ctx context.Context, code string) error {
	result := r.db.WithContext(ctx).Where("code = ?", code).Delete(&models.RegionType{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete region type: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("region type not found")
	}
	return nil
}
