package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// attributeCategoryRepository implements AttributeCategoryRepository
type attributeCategoryRepository struct {
	db *gorm.DB
}

// NewAttributeCategoryRepository creates a new attribute category repository
func NewAttributeCategoryRepository(db *gorm.DB) AttributeCategoryRepository {
	return &attributeCategoryRepository{db: db}
}

// GetByID retrieves an attribute category by its ID (excludes soft-deleted records)
func (r *attributeCategoryRepository) GetByID(ctx context.Context, id string) (*models.AttributeCategory, error) {
	var category models.AttributeCategory
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("id = ? AND is_active = ?", categoryID, true).
		Preload("Attributes").First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("attribute category %s not found", id)
		}
		return nil, fmt.Errorf("failed to get attribute category: %w", err)
	}
	return &category, nil
}

// GetByIDWithDeleted retrieves an attribute category by its ID (includes soft-deleted records)
func (r *attributeCategoryRepository) GetByIDWithDeleted(ctx context.Context, id string) (*models.AttributeCategory, error) {
	var category models.AttributeCategory
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Unscoped().Where("id = ?", categoryID).
		Preload("Attributes").First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("attribute category %s not found", id)
		}
		return nil, fmt.Errorf("failed to get attribute category: %w", err)
	}
	return &category, nil
}

// GetByCode retrieves an attribute category by its code (excludes soft-deleted records)
func (r *attributeCategoryRepository) GetByCode(ctx context.Context, code string) (*models.AttributeCategory, error) {
	var category models.AttributeCategory
	err := r.db.WithContext(ctx).Where("code = ? AND is_active = ?", code, true).
		Preload("Attributes").First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("attribute category %s not found", code)
		}
		return nil, fmt.Errorf("failed to get attribute category: %w", err)
	}
	return &category, nil
}

// GetByCodeWithDeleted retrieves an attribute category by its code (includes soft-deleted records)
func (r *attributeCategoryRepository) GetByCodeWithDeleted(ctx context.Context, code string) (*models.AttributeCategory, error) {
	var category models.AttributeCategory
	err := r.db.WithContext(ctx).Unscoped().Where("code = ?", code).
		Preload("Attributes").First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("attribute category %s not found", code)
		}
		return nil, fmt.Errorf("failed to get attribute category: %w", err)
	}
	return &category, nil
}

// GetAll retrieves all attribute categories with pagination (excludes soft-deleted records)
func (r *attributeCategoryRepository) GetAll(ctx context.Context, offset, limit int) ([]models.AttributeCategory, int64, error) {
	var categories []models.AttributeCategory
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Model(&models.AttributeCategory{}).
		Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count attribute categories: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).Where("is_active = ?", true).
		Preload("Attributes").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&categories).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get attribute categories: %w", err)
	}

	return categories, total, nil
}

// GetAllWithDeleted retrieves all attribute categories with pagination (includes soft-deleted records)
func (r *attributeCategoryRepository) GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.AttributeCategory, int64, error) {
	var categories []models.AttributeCategory
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Unscoped().Model(&models.AttributeCategory{}).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count attribute categories: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).Unscoped().
		Preload("Attributes").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&categories).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get attribute categories: %w", err)
	}

	return categories, total, nil
}

// GetOnlyDeleted retrieves only soft-deleted attribute categories with pagination
func (r *attributeCategoryRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.AttributeCategory, int64, error) {
	var categories []models.AttributeCategory
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Unscoped().Model(&models.AttributeCategory{}).
		Where("is_deleted = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted attribute categories: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).Unscoped().Where("is_deleted = ?", true).
		Preload("Attributes").
		Offset(offset).Limit(limit).
		Order("deleted_at DESC").
		Find(&categories).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deleted attribute categories: %w", err)
	}

	return categories, total, nil
}

// Create creates a new attribute category
func (r *attributeCategoryRepository) Create(ctx context.Context, category *models.AttributeCategory) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		return fmt.Errorf("failed to create attribute category: %w", err)
	}
	return nil
}

// Update updates an existing attribute category
func (r *attributeCategoryRepository) Update(ctx context.Context, id string, category *models.AttributeCategory) error {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", categoryID).Save(category)
	if result.Error != nil {
		return fmt.Errorf("failed to update attribute category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("attribute category not found")
	}
	return nil
}

// Delete performs soft delete on an attribute category
func (r *attributeCategoryRepository) Delete(ctx context.Context, id string) error {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	// Get the category first
	var category models.AttributeCategory
	err = r.db.WithContext(ctx).Where("id = ?", categoryID).First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("attribute category not found")
		}
		return fmt.Errorf("failed to find attribute category: %w", err)
	}

	// Check if category can be deleted
	canDelete, err := category.CanBeDeleted(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to check if category can be deleted: %w", err)
	}
	if !canDelete {
		return fmt.Errorf("cannot delete attribute category: it has active attributes")
	}

	// Use the model's soft delete method
	err = category.SoftDelete(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to soft delete attribute category: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted attribute category
func (r *attributeCategoryRepository) Restore(ctx context.Context, id string) error {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	// Get the soft-deleted category
	var category models.AttributeCategory
	err = r.db.WithContext(ctx).Unscoped().Where("id = ? AND is_deleted = ?", categoryID, true).First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("soft-deleted attribute category not found")
		}
		return fmt.Errorf("failed to find soft-deleted attribute category: %w", err)
	}

	// Use the model's restore method
	err = category.Restore(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to restore attribute category: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes an attribute category (hard delete)
func (r *attributeCategoryRepository) ForceDelete(ctx context.Context, id string) error {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Unscoped().Where("id = ?", categoryID).Delete(&models.AttributeCategory{})
	if result.Error != nil {
		return fmt.Errorf("failed to force delete attribute category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("attribute category not found")
	}
	return nil
}

// CanBeDeleted checks if an attribute category can be safely deleted
func (r *attributeCategoryRepository) CanBeDeleted(ctx context.Context, id string) (bool, error) {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return false, fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	var count int64
	err = r.db.WithContext(ctx).Model(&models.Attribute{}).
		Where("category_id = ? AND is_deleted = ?", categoryID, false).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check attribute count: %w", err)
	}

	return count == 0, nil
}

// GetStats retrieves statistics for an attribute category
func (r *attributeCategoryRepository) GetStats(ctx context.Context, id string) (*AttributeCategoryStats, error) {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	// Get category basic info
	var category models.AttributeCategory
	err = r.db.WithContext(ctx).Where("id = ?", categoryID).First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("attribute category not found")
		}
		return nil, fmt.Errorf("failed to get attribute category: %w", err)
	}

	stats := &AttributeCategoryStats{
		ID:   category.ID,
		Code: category.Code,
		Name: category.Name,
	}

	// Count total attributes
	err = r.db.WithContext(ctx).Model(&models.Attribute{}).
		Where("category_id = ? AND is_deleted = ?", categoryID, false).Count(&stats.TotalAttributes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total attributes: %w", err)
	}

	// Count active attributes
	err = r.db.WithContext(ctx).Model(&models.Attribute{}).
		Where("category_id = ? AND is_deleted = ? AND is_active = ?", categoryID, false, true).Count(&stats.ActiveAttributes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count active attributes: %w", err)
	}

	// Count total mappings for this category
	err = r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Joins("JOIN attribute ON partner_attribute_map.attribute_id = attribute.id").
		Where("attribute.category_id = ? AND partner_attribute_map.is_deleted = ?", categoryID, false).
		Count(&stats.TotalMappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total mappings: %w", err)
	}

	// Count active mappings for this category
	err = r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Joins("JOIN attribute ON partner_attribute_map.attribute_id = attribute.id").
		Where("attribute.category_id = ? AND partner_attribute_map.is_deleted = ? AND partner_attribute_map.is_active = ?", categoryID, false, true).
		Count(&stats.ActiveMappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count active mappings: %w", err)
	}

	// Count unique partners for this category
	err = r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Joins("JOIN attribute ON partner_attribute_map.attribute_id = attribute.id").
		Where("attribute.category_id = ? AND partner_attribute_map.is_deleted = ?", categoryID, false).
		Distinct("partner_code").
		Count(&stats.UniquePartners).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count unique partners: %w", err)
	}

	return stats, nil
}
