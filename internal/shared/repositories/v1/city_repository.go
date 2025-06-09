package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cityRepository struct {
	db *gorm.DB
}

func NewCityRepository(db *gorm.DB) CityRepository {
	return &cityRepository{db: db}
}

func (r *cityRepository) GetByID(ctx context.Context, id string) (*models.City, error) {
	var city models.City
	cityID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid city ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("id = ? AND is_active = ?", cityID, true).
		Preload("Region").Preload("Country").Preload("District").First(&city).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("city %s not found", id)
		}
		return nil, fmt.Errorf("failed to get city: %w", err)
	}
	return &city, nil
}

func (r *cityRepository) GetByCode(ctx context.Context, code string) (*models.City, error) {
	var city models.City
	err := r.db.WithContext(ctx).Where("code = ? AND is_active = ?", code, true).
		Preload("Region").Preload("Country").Preload("District").First(&city).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("city %s not found", code)
		}
		return nil, fmt.Errorf("failed to get city: %w", err)
	}
	return &city, nil
}

func (r *cityRepository) GetAll(ctx context.Context, offset, limit int) ([]models.City, int64, error) {
	var cities []models.City
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.City{}).Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count cities: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Where("is_active = ?", true).
		Preload("Region").Preload("Country").Preload("District").
		Offset(offset).Limit(limit).Find(&cities).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get cities: %w", err)
	}

	return cities, total, nil
}

func (r *cityRepository) GetByRegionID(ctx context.Context, regionID string) ([]models.City, error) {
	var cities []models.City
	regionUUID, err := uuid.Parse(regionID)
	if err != nil {
		return nil, fmt.Errorf("invalid region ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("region_id = ? AND is_active = ?", regionUUID, true).
		Preload("Region").Preload("Country").Preload("District").Find(&cities).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get cities by region: %w", err)
	}
	return cities, nil
}

func (r *cityRepository) Create(ctx context.Context, city *models.City) error {
	err := r.db.WithContext(ctx).Create(city).Error
	if err != nil {
		return fmt.Errorf("failed to create city: %w", err)
	}
	return nil
}

func (r *cityRepository) Update(ctx context.Context, id string, city *models.City) error {
	cityID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid city ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", cityID).Save(city)
	if result.Error != nil {
		return fmt.Errorf("failed to update city: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("city not found")
	}
	return nil
}

func (r *cityRepository) Delete(ctx context.Context, id string) error {
	cityID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid city ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", cityID).Delete(&models.City{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete city: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("city not found")
	}
	return nil
}
