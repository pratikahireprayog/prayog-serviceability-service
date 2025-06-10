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

// GetByIDWithDeleted retrieves a city by its ID (includes soft-deleted records)
func (r *cityRepository) GetByIDWithDeleted(ctx context.Context, id string) (*models.City, error) {
	var city models.City
	cityID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid city ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Unscoped().Where("id = ? AND is_active = ?", cityID, true).
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

// GetByCodeWithDeleted retrieves a city by its code (includes soft-deleted records)
func (r *cityRepository) GetByCodeWithDeleted(ctx context.Context, code string) (*models.City, error) {
	var city models.City
	err := r.db.WithContext(ctx).Unscoped().Where("code = ? AND is_active = ?", code, true).
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

// GetAllWithDeleted retrieves all cities with pagination (includes soft-deleted records)
func (r *cityRepository) GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.City, int64, error) {
	var cities []models.City
	var total int64

	// Count total records including soft-deleted
	err := r.db.WithContext(ctx).Unscoped().Model(&models.City{}).Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count cities: %w", err)
	}

	// Get paginated records including soft-deleted
	err = r.db.WithContext(ctx).Unscoped().Where("is_active = ?", true).
		Preload("Region").Preload("Country").Preload("District").
		Offset(offset).Limit(limit).Find(&cities).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get cities: %w", err)
	}

	return cities, total, nil
}

// GetOnlyDeleted retrieves only soft-deleted cities with pagination
func (r *cityRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.City, int64, error) {
	var cities []models.City
	var total int64

	// Count only soft-deleted records
	err := r.db.WithContext(ctx).Unscoped().Model(&models.City{}).Where("is_deleted = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted cities: %w", err)
	}

	// Get paginated soft-deleted records
	err = r.db.WithContext(ctx).Unscoped().Where("is_deleted = ?", true).
		Preload("Region").Preload("Country").Preload("District").
		Offset(offset).Limit(limit).Find(&cities).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deleted cities: %w", err)
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

// GetByRegionIDWithDeleted retrieves cities by region ID (includes soft-deleted records)
func (r *cityRepository) GetByRegionIDWithDeleted(ctx context.Context, regionID string) ([]models.City, error) {
	var cities []models.City
	regionUUID, err := uuid.Parse(regionID)
	if err != nil {
		return nil, fmt.Errorf("invalid region ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Unscoped().Where("region_id = ? AND is_active = ?", regionUUID, true).
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

// Delete performs soft delete on a city
func (r *cityRepository) Delete(ctx context.Context, id string) error {
	cityID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid city ID format: %w", err)
	}

	// Get the city first
	var city models.City
	err = r.db.WithContext(ctx).Where("id = ?", cityID).First(&city).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("city not found")
		}
		return fmt.Errorf("failed to find city: %w", err)
	}

	// Use the model's soft delete method
	err = city.SoftDelete(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to soft delete city: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted city
func (r *cityRepository) Restore(ctx context.Context, id string) error {
	cityID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid city ID format: %w", err)
	}

	// Get the soft-deleted city
	var city models.City
	err = r.db.WithContext(ctx).Unscoped().Where("id = ? AND is_deleted = ?", cityID, true).First(&city).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("soft-deleted city not found")
		}
		return fmt.Errorf("failed to find soft-deleted city: %w", err)
	}

	// Use the model's restore method
	err = city.Restore(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to restore city: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes a city (hard delete)
func (r *cityRepository) ForceDelete(ctx context.Context, id string) error {
	cityID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid city ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Unscoped().Where("id = ?", cityID).Delete(&models.City{})
	if result.Error != nil {
		return fmt.Errorf("failed to force delete city: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("city not found")
	}
	return nil
}
