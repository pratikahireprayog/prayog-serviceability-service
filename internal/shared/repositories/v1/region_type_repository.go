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

// GetByCode retrieves a region type by its code (excludes soft-deleted records)
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

// GetByCodeWithDeleted retrieves a region type by its code (includes soft-deleted records)
func (r *regionTypeRepository) GetByCodeWithDeleted(ctx context.Context, code string) (*models.RegionType, error) {
	var regionType models.RegionType
	err := r.db.WithContext(ctx).Unscoped().Where("code = ? AND is_active = ?", code, true).First(&regionType).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("region type %s not found", code)
		}
		return nil, fmt.Errorf("failed to get region type: %w", err)
	}
	return &regionType, nil
}

// GetAll retrieves all region types with pagination (excludes soft-deleted records)
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

// GetAllWithDeleted retrieves all region types with pagination (includes soft-deleted records)
func (r *regionTypeRepository) GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.RegionType, int64, error) {
	var regionTypes []models.RegionType
	var total int64

	// Count total records including soft-deleted
	err := r.db.WithContext(ctx).Unscoped().Model(&models.RegionType{}).Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count region types: %w", err)
	}

	// Get paginated records including soft-deleted
	err = r.db.WithContext(ctx).Unscoped().Where("is_active = ?", true).
		Offset(offset).Limit(limit).Find(&regionTypes).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get region types: %w", err)
	}

	return regionTypes, total, nil
}

// GetOnlyDeleted retrieves only soft-deleted region types with pagination
func (r *regionTypeRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.RegionType, int64, error) {
	var regionTypes []models.RegionType
	var total int64

	// Count only soft-deleted records
	err := r.db.WithContext(ctx).Unscoped().Model(&models.RegionType{}).Where("is_deleted = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted region types: %w", err)
	}

	// Get paginated soft-deleted records
	err = r.db.WithContext(ctx).Unscoped().Where("is_deleted = ?", true).
		Offset(offset).Limit(limit).Find(&regionTypes).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deleted region types: %w", err)
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

// Delete performs soft delete on a region type
func (r *regionTypeRepository) Delete(ctx context.Context, code string) error {
	// Get the region type first
	var regionType models.RegionType
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&regionType).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("region type not found")
		}
		return fmt.Errorf("failed to find region type: %w", err)
	}

	// Use the model's soft delete method
	err = regionType.SoftDelete(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to soft delete region type: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted region type
func (r *regionTypeRepository) Restore(ctx context.Context, code string) error {
	// Get the soft-deleted region type
	var regionType models.RegionType
	err := r.db.WithContext(ctx).Unscoped().Where("code = ? AND is_deleted = ?", code, true).First(&regionType).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("soft-deleted region type not found")
		}
		return fmt.Errorf("failed to find soft-deleted region type: %w", err)
	}

	// Use the model's restore method
	err = regionType.Restore(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to restore region type: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes a region type (hard delete)
func (r *regionTypeRepository) ForceDelete(ctx context.Context, code string) error {
	result := r.db.WithContext(ctx).Unscoped().Where("code = ?", code).Delete(&models.RegionType{})
	if result.Error != nil {
		return fmt.Errorf("failed to force delete region type: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("region type not found")
	}
	return nil
}
