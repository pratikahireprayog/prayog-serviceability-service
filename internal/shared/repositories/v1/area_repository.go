package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type areaRepository struct {
	db *gorm.DB
}

func NewAreaRepository(db *gorm.DB) AreaRepository {
	return &areaRepository{db: db}
}

func (r *areaRepository) GetByID(ctx context.Context, id string) (*models.Area, error) {
	var area models.Area
	areaID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid area ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("id = ? AND is_active = ?", areaID, true).
		Preload("City").First(&area).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("area %s not found", id)
		}
		return nil, fmt.Errorf("failed to get area: %w", err)
	}
	return &area, nil
}

func (r *areaRepository) GetByCode(ctx context.Context, code string) (*models.Area, error) {
	var area models.Area
	err := r.db.WithContext(ctx).Where("code = ? AND is_active = ?", code, true).
		Preload("City").First(&area).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("area %s not found", code)
		}
		return nil, fmt.Errorf("failed to get area: %w", err)
	}
	return &area, nil
}

func (r *areaRepository) GetAll(ctx context.Context, offset, limit int) ([]models.Area, int64, error) {
	var areas []models.Area
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.Area{}).Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count areas: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Where("is_active = ?", true).
		Preload("City").
		Offset(offset).Limit(limit).Find(&areas).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get areas: %w", err)
	}

	return areas, total, nil
}

func (r *areaRepository) GetByCityID(ctx context.Context, cityID string) ([]models.Area, error) {
	var areas []models.Area
	cityUUID, err := uuid.Parse(cityID)
	if err != nil {
		return nil, fmt.Errorf("invalid city ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("city_id = ? AND is_active = ?", cityUUID, true).
		Preload("City").Find(&areas).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get areas by city: %w", err)
	}
	return areas, nil
}

func (r *areaRepository) Create(ctx context.Context, area *models.Area) error {
	err := r.db.WithContext(ctx).Create(area).Error
	if err != nil {
		return fmt.Errorf("failed to create area: %w", err)
	}
	return nil
}

func (r *areaRepository) Update(ctx context.Context, id string, area *models.Area) error {
	areaID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid area ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", areaID).Save(area)
	if result.Error != nil {
		return fmt.Errorf("failed to update area: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("area not found")
	}
	return nil
}

func (r *areaRepository) Delete(ctx context.Context, id string) error {
	areaID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid area ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", areaID).Delete(&models.Area{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete area: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("area not found")
	}
	return nil
}
