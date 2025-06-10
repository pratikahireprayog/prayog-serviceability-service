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

// GetByIDWithDeleted retrieves an area by its ID (includes soft-deleted records)
func (r *areaRepository) GetByIDWithDeleted(ctx context.Context, id string) (*models.Area, error) {
	var area models.Area
	areaID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid area ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Unscoped().Where("id = ? AND is_active = ?", areaID, true).
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

// GetByCodeWithDeleted retrieves an area by its code (includes soft-deleted records)
func (r *areaRepository) GetByCodeWithDeleted(ctx context.Context, code string) (*models.Area, error) {
	var area models.Area
	err := r.db.WithContext(ctx).Unscoped().Where("code = ? AND is_active = ?", code, true).
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

// GetAllWithDeleted retrieves all areas with pagination (includes soft-deleted records)
func (r *areaRepository) GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.Area, int64, error) {
	var areas []models.Area
	var total int64

	// Count total records including soft-deleted
	err := r.db.WithContext(ctx).Unscoped().Model(&models.Area{}).Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count areas: %w", err)
	}

	// Get paginated records including soft-deleted
	err = r.db.WithContext(ctx).Unscoped().Where("is_active = ?", true).
		Preload("City").
		Offset(offset).Limit(limit).Find(&areas).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get areas: %w", err)
	}

	return areas, total, nil
}

// GetOnlyDeleted retrieves only soft-deleted areas with pagination
func (r *areaRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.Area, int64, error) {
	var areas []models.Area
	var total int64

	// Count only soft-deleted records
	err := r.db.WithContext(ctx).Unscoped().Model(&models.Area{}).Where("is_deleted = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted areas: %w", err)
	}

	// Get paginated soft-deleted records
	err = r.db.WithContext(ctx).Unscoped().Where("is_deleted = ?", true).
		Preload("City").
		Offset(offset).Limit(limit).Find(&areas).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deleted areas: %w", err)
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

// GetByCityIDWithDeleted retrieves areas by city ID (includes soft-deleted records)
func (r *areaRepository) GetByCityIDWithDeleted(ctx context.Context, cityID string) ([]models.Area, error) {
	var areas []models.Area
	cityUUID, err := uuid.Parse(cityID)
	if err != nil {
		return nil, fmt.Errorf("invalid city ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Unscoped().Where("city_id = ? AND is_active = ?", cityUUID, true).
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

// Delete performs soft delete on an area
func (r *areaRepository) Delete(ctx context.Context, id string) error {
	areaID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid area ID format: %w", err)
	}

	// Get the area first
	var area models.Area
	err = r.db.WithContext(ctx).Where("id = ?", areaID).First(&area).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("area not found")
		}
		return fmt.Errorf("failed to find area: %w", err)
	}

	// Use the model's soft delete method
	err = area.SoftDelete(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to soft delete area: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted area
func (r *areaRepository) Restore(ctx context.Context, id string) error {
	areaID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid area ID format: %w", err)
	}

	// Get the soft-deleted area
	var area models.Area
	err = r.db.WithContext(ctx).Unscoped().Where("id = ? AND is_deleted = ?", areaID, true).First(&area).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("soft-deleted area not found")
		}
		return fmt.Errorf("failed to find soft-deleted area: %w", err)
	}

	// Use the model's restore method
	err = area.Restore(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to restore area: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes an area (hard delete)
func (r *areaRepository) ForceDelete(ctx context.Context, id string) error {
	areaID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid area ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Unscoped().Where("id = ?", areaID).Delete(&models.Area{})
	if result.Error != nil {
		return fmt.Errorf("failed to force delete area: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("area not found")
	}
	return nil
}
