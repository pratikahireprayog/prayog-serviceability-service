package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type districtRepository struct {
	db *gorm.DB
}

func NewDistrictRepository(db *gorm.DB) DistrictRepository {
	return &districtRepository{db: db}
}

func (r *districtRepository) GetByID(ctx context.Context, id string) (*models.District, error) {
	var district models.District
	districtID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid district ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("id = ? AND is_active = ?", districtID, true).
		Preload("Region").Preload("Country").First(&district).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("district %s not found", id)
		}
		return nil, fmt.Errorf("failed to get district: %w", err)
	}
	return &district, nil
}

func (r *districtRepository) GetByCode(ctx context.Context, code string) (*models.District, error) {
	var district models.District
	err := r.db.WithContext(ctx).Where("code = ? AND is_active = ?", code, true).
		Preload("Region").Preload("Country").First(&district).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("district %s not found", code)
		}
		return nil, fmt.Errorf("failed to get district: %w", err)
	}
	return &district, nil
}

func (r *districtRepository) GetAll(ctx context.Context, offset, limit int) ([]models.District, int64, error) {
	var districts []models.District
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.District{}).Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count districts: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Where("is_active = ?", true).
		Preload("Region").Preload("Country").
		Offset(offset).Limit(limit).Find(&districts).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get districts: %w", err)
	}

	return districts, total, nil
}

func (r *districtRepository) GetByRegionID(ctx context.Context, regionID string) ([]models.District, error) {
	var districts []models.District
	regionUUID, err := uuid.Parse(regionID)
	if err != nil {
		return nil, fmt.Errorf("invalid region ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("region_id = ? AND is_active = ?", regionUUID, true).
		Preload("Region").Preload("Country").Find(&districts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get districts by region: %w", err)
	}
	return districts, nil
}

func (r *districtRepository) Create(ctx context.Context, district *models.District) error {
	err := r.db.WithContext(ctx).Create(district).Error
	if err != nil {
		return fmt.Errorf("failed to create district: %w", err)
	}
	return nil
}

func (r *districtRepository) Update(ctx context.Context, id string, district *models.District) error {
	districtID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid district ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", districtID).Save(district)
	if result.Error != nil {
		return fmt.Errorf("failed to update district: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("district not found")
	}
	return nil
}

func (r *districtRepository) Delete(ctx context.Context, id string) error {
	districtID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid district ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", districtID).Delete(&models.District{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete district: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("district not found")
	}
	return nil
}
