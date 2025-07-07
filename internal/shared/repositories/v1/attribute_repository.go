package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// attributeRepository implements AttributeRepository
type attributeRepository struct {
	db *gorm.DB
}

// NewAttributeRepository creates a new attribute repository
func NewAttributeRepository(db *gorm.DB) AttributeRepository {
	return &attributeRepository{db: db}
}

// GetByID retrieves an attribute by its ID (excludes soft-deleted records)
func (r *attributeRepository) GetByID(ctx context.Context, id string) (*models.Attribute, error) {
	var attribute models.Attribute
	attributeID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("id = ? AND is_active = ?", attributeID, true).
		Preload("Category").First(&attribute).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("attribute %s not found", id)
		}
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}
	return &attribute, nil
}

// GetByIDWithDeleted retrieves an attribute by its ID (includes soft-deleted records)
func (r *attributeRepository) GetByIDWithDeleted(ctx context.Context, id string) (*models.Attribute, error) {
	var attribute models.Attribute
	attributeID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Unscoped().Where("id = ?", attributeID).
		Preload("Category").First(&attribute).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("attribute %s not found", id)
		}
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}
	return &attribute, nil
}

// GetByCode retrieves an attribute by its code (excludes soft-deleted records)
func (r *attributeRepository) GetByCode(ctx context.Context, code string) (*models.Attribute, error) {
	var attribute models.Attribute
	err := r.db.WithContext(ctx).Where("code = ? AND is_active = ?", code, true).
		Preload("Category").First(&attribute).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("attribute %s not found", code)
		}
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}
	return &attribute, nil
}

// GetByCodeWithDeleted retrieves an attribute by its code (includes soft-deleted records)
func (r *attributeRepository) GetByCodeWithDeleted(ctx context.Context, code string) (*models.Attribute, error) {
	var attribute models.Attribute
	err := r.db.WithContext(ctx).Unscoped().Where("code = ?", code).
		Preload("Category").First(&attribute).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("attribute %s not found", code)
		}
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}
	return &attribute, nil
}

// GetByCategoryID retrieves all attributes by category ID (excludes soft-deleted records)
func (r *attributeRepository) GetByCategoryID(ctx context.Context, categoryID string) ([]models.Attribute, error) {
	var attributes []models.Attribute
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("category_id = ? AND is_active = ?", categoryUUID, true).
		Preload("Category").Order("created_at DESC").Find(&attributes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes by category ID: %w", err)
	}
	return attributes, nil
}

// GetByCategoryIDWithDeleted retrieves all attributes by category ID (includes soft-deleted records)
func (r *attributeRepository) GetByCategoryIDWithDeleted(ctx context.Context, categoryID string) ([]models.Attribute, error) {
	var attributes []models.Attribute
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Unscoped().Where("category_id = ?", categoryUUID).
		Preload("Category").Order("created_at DESC").Find(&attributes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes by category ID: %w", err)
	}
	return attributes, nil
}

// GetByCategoryCode retrieves all attributes by category code (excludes soft-deleted records)
func (r *attributeRepository) GetByCategoryCode(ctx context.Context, categoryCode string) ([]models.Attribute, error) {
	var attributes []models.Attribute
	err := r.db.WithContext(ctx).
		Joins("JOIN attribute_category ON attribute.category_id = attribute_category.id").
		Where("attribute_category.code = ? AND attribute.is_active = ? AND attribute_category.is_active = ?", categoryCode, true, true).
		Preload("Category").Order("attribute.created_at DESC").Find(&attributes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes by category code: %w", err)
	}
	return attributes, nil
}

// GetAll retrieves all attributes with pagination (excludes soft-deleted records)
func (r *attributeRepository) GetAll(ctx context.Context, offset, limit int) ([]models.Attribute, int64, error) {
	var attributes []models.Attribute
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Model(&models.Attribute{}).
		Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count attributes: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).Where("is_active = ?", true).
		Preload("Category").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&attributes).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get attributes: %w", err)
	}

	return attributes, total, nil
}

// GetAllWithDeleted retrieves all attributes with pagination (includes soft-deleted records)
func (r *attributeRepository) GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.Attribute, int64, error) {
	var attributes []models.Attribute
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Unscoped().Model(&models.Attribute{}).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count attributes: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).Unscoped().
		Preload("Category").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&attributes).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get attributes: %w", err)
	}

	return attributes, total, nil
}

// GetOnlyDeleted retrieves only soft-deleted attributes with pagination
func (r *attributeRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.Attribute, int64, error) {
	var attributes []models.Attribute
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Unscoped().Model(&models.Attribute{}).
		Where("is_deleted = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted attributes: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).Unscoped().Where("is_deleted = ?", true).
		Preload("Category").
		Offset(offset).Limit(limit).
		Order("deleted_at DESC").
		Find(&attributes).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deleted attributes: %w", err)
	}

	return attributes, total, nil
}

// Create creates a new attribute
func (r *attributeRepository) Create(ctx context.Context, attribute *models.Attribute) error {
	err := r.db.WithContext(ctx).Create(attribute).Error
	if err != nil {
		return fmt.Errorf("failed to create attribute: %w", err)
	}
	return nil
}

// Update updates an existing attribute
func (r *attributeRepository) Update(ctx context.Context, id string, attribute *models.Attribute) error {
	attributeID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid attribute ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", attributeID).Save(attribute)
	if result.Error != nil {
		return fmt.Errorf("failed to update attribute: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("attribute not found")
	}
	return nil
}

// Delete performs soft delete on an attribute
func (r *attributeRepository) Delete(ctx context.Context, id string) error {
	attributeID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid attribute ID format: %w", err)
	}

	// Get the attribute first
	var attribute models.Attribute
	err = r.db.WithContext(ctx).Where("id = ?", attributeID).First(&attribute).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("attribute not found")
		}
		return fmt.Errorf("failed to find attribute: %w", err)
	}

	// Check if attribute can be deleted
	canDelete, err := attribute.CanBeDeleted(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to check if attribute can be deleted: %w", err)
	}
	if !canDelete {
		return fmt.Errorf("cannot delete attribute: it has active partner mappings")
	}

	// Use the model's soft delete method
	err = attribute.SoftDelete(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to soft delete attribute: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted attribute
func (r *attributeRepository) Restore(ctx context.Context, id string) error {
	attributeID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid attribute ID format: %w", err)
	}

	// Get the soft-deleted attribute
	var attribute models.Attribute
	err = r.db.WithContext(ctx).Unscoped().Where("id = ? AND is_deleted = ?", attributeID, true).First(&attribute).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("soft-deleted attribute not found")
		}
		return fmt.Errorf("failed to find soft-deleted attribute: %w", err)
	}

	// Use the model's restore method
	err = attribute.Restore(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to restore attribute: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes an attribute (hard delete)
func (r *attributeRepository) ForceDelete(ctx context.Context, id string) error {
	attributeID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid attribute ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Unscoped().Where("id = ?", attributeID).Delete(&models.Attribute{})
	if result.Error != nil {
		return fmt.Errorf("failed to force delete attribute: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("attribute not found")
	}
	return nil
}

// CanBeDeleted checks if an attribute can be safely deleted
func (r *attributeRepository) CanBeDeleted(ctx context.Context, id string) (bool, error) {
	attributeID, err := uuid.Parse(id)
	if err != nil {
		return false, fmt.Errorf("invalid attribute ID format: %w", err)
	}

	var count int64
	err = r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Where("attribute_id = ? AND is_deleted = ?", attributeID, false).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check partner mapping count: %w", err)
	}

	return count == 0, nil
}

// GetStats retrieves statistics for an attribute
func (r *attributeRepository) GetStats(ctx context.Context, id string) (*AttributeStats, error) {
	attributeID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute ID format: %w", err)
	}

	// Get attribute basic info with category
	var attribute models.Attribute
	err = r.db.WithContext(ctx).Where("id = ?", attributeID).
		Preload("Category").First(&attribute).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("attribute not found")
		}
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}

	stats := &AttributeStats{
		ID:           attribute.ID,
		Code:         attribute.Code,
		Name:         attribute.Name,
		CategoryCode: attribute.Category.Code,
		CategoryName: attribute.Category.Name,
	}

	// Count total mappings
	err = r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Where("attribute_id = ? AND is_deleted = ?", attributeID, false).Count(&stats.TotalMappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total mappings: %w", err)
	}

	// Count active mappings
	err = r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Where("attribute_id = ? AND is_deleted = ? AND is_active = ?", attributeID, false, true).Count(&stats.ActiveMappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count active mappings: %w", err)
	}

	// Count unique partners
	err = r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Where("attribute_id = ? AND is_deleted = ?", attributeID, false).
		Distinct("partner_code").
		Count(&stats.UniquePartners).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count unique partners: %w", err)
	}

	// Get last mapping date
	var lastMapping models.PartnerAttributeMap
	err = r.db.WithContext(ctx).Where("attribute_id = ? AND is_deleted = ?", attributeID, false).
		Order("created_at DESC").First(&lastMapping).Error
	if err == nil {
		lastMappingTime := lastMapping.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		stats.LastMappingAt = &lastMappingTime
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get last mapping date: %w", err)
	}

	return stats, nil
}
