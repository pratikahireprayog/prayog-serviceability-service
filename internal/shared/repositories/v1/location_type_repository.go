package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"gorm.io/gorm"
)

// locationTypeRepository implements LocationTypeRepository
type locationTypeRepository struct {
	db *gorm.DB
}

// NewLocationTypeRepository creates a new location type repository
func NewLocationTypeRepository(db *gorm.DB) LocationTypeRepository {
	return &locationTypeRepository{db: db}
}

// GetByCode retrieves a location type by its code
func (r *locationTypeRepository) GetByCode(ctx context.Context, code string) (*models.LocationType, error) {
	var locationType models.LocationType
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&locationType).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("location type %s not found", code)
		}
		return nil, fmt.Errorf("failed to get location type: %w", err)
	}
	return &locationType, nil
}

// GetAll retrieves all location types with pagination
func (r *locationTypeRepository) GetAll(ctx context.Context, offset, limit int) ([]models.LocationType, int64, error) {
	var locationTypes []models.LocationType
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Model(&models.LocationType{}).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count location types: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Find(&locationTypes).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get location types: %w", err)
	}

	return locationTypes, total, nil
}

// Create creates a new location type
func (r *locationTypeRepository) Create(ctx context.Context, locationType *models.LocationType) error {
	err := r.db.WithContext(ctx).Create(locationType).Error
	if err != nil {
		return fmt.Errorf("failed to create location type: %w", err)
	}
	return nil
}

// Update updates an existing location type
func (r *locationTypeRepository) Update(ctx context.Context, locationType *models.LocationType) error {
	result := r.db.WithContext(ctx).Save(locationType)
	if result.Error != nil {
		return fmt.Errorf("failed to update location type: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("location type not found")
	}
	return nil
}

// Delete soft deletes a location type by marking it as deleted
func (r *locationTypeRepository) Delete(ctx context.Context, code string) error {
	// Get the location type
	var locationType models.LocationType
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&locationType).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("location type not found")
		}
		return fmt.Errorf("failed to find location type: %w", err)
	}

	// Use the model's soft delete method
	return locationType.SoftDelete(r.db.WithContext(ctx))
}

// GetByCodeWithDeleted retrieves a location type by its code (includes soft-deleted records)
func (r *locationTypeRepository) GetByCodeWithDeleted(ctx context.Context, code string) (*models.LocationType, error) {
	var locationType models.LocationType
	err := r.db.WithContext(ctx).Unscoped().Where("code = ?", code).First(&locationType).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("location type %s not found", code)
		}
		return nil, fmt.Errorf("failed to get location type: %w", err)
	}
	return &locationType, nil
}

// GetAllWithDeleted retrieves all location types with pagination (includes soft-deleted records)
func (r *locationTypeRepository) GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.LocationType, int64, error) {
	var locationTypes []models.LocationType
	var total int64

	// Get total count (including soft-deleted)
	err := r.db.WithContext(ctx).Unscoped().Model(&models.LocationType{}).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count location types: %w", err)
	}

	// Get paginated results (including soft-deleted)
	err = r.db.WithContext(ctx).Unscoped().
		Offset(offset).
		Limit(limit).
		Find(&locationTypes).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get location types: %w", err)
	}

	return locationTypes, total, nil
}

// GetOnlyDeleted retrieves only soft-deleted location types with pagination
func (r *locationTypeRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.LocationType, int64, error) {
	var locationTypes []models.LocationType
	var total int64

	// Count only soft-deleted records
	err := r.db.WithContext(ctx).Unscoped().Model(&models.LocationType{}).Where("is_deleted = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted location types: %w", err)
	}

	// Get paginated soft-deleted records
	err = r.db.WithContext(ctx).Unscoped().
		Where("is_deleted = ?", true).
		Offset(offset).Limit(limit).
		Find(&locationTypes).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deleted location types: %w", err)
	}

	return locationTypes, total, nil
}

// Restore restores a soft-deleted location type
func (r *locationTypeRepository) Restore(ctx context.Context, code string) error {
	// Get the soft-deleted location type
	var locationType models.LocationType
	err := r.db.WithContext(ctx).Unscoped().Where("code = ? AND is_deleted = ?", code, true).First(&locationType).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("soft-deleted location type not found")
		}
		return fmt.Errorf("failed to find soft-deleted location type: %w", err)
	}

	// Use the model's restore method
	return locationType.Restore(r.db.WithContext(ctx))
}

// ForceDelete permanently deletes a location type (hard delete)
func (r *locationTypeRepository) ForceDelete(ctx context.Context, code string) error {
	result := r.db.WithContext(ctx).Unscoped().Where("code = ?", code).Delete(&models.LocationType{})
	if result.Error != nil {
		return fmt.Errorf("failed to force delete location type: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("location type not found")
	}
	return nil
}
