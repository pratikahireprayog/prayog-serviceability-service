package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// partnerAttributeMapRepository implements PartnerAttributeMapRepository
type partnerAttributeMapRepository struct {
	db *gorm.DB
}

// NewPartnerAttributeMapRepository creates a new partner attribute map repository
func NewPartnerAttributeMapRepository(db *gorm.DB) PartnerAttributeMapRepository {
	return &partnerAttributeMapRepository{db: db}
}

// GetByID retrieves a partner attribute mapping by its ID (excludes soft-deleted records)
func (r *partnerAttributeMapRepository) GetByID(ctx context.Context, id string) (*models.PartnerAttributeMap, error) {
	var mapping models.PartnerAttributeMap
	mappingID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid partner attribute mapping ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("id = ? AND is_active = ?", mappingID, true).
		Preload("Attribute").Preload("Attribute.Category").First(&mapping).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("partner attribute mapping %s not found", id)
		}
		return nil, fmt.Errorf("failed to get partner attribute mapping: %w", err)
	}
	return &mapping, nil
}

// GetByIDWithDeleted retrieves a partner attribute mapping by its ID (includes soft-deleted records)
func (r *partnerAttributeMapRepository) GetByIDWithDeleted(ctx context.Context, id string) (*models.PartnerAttributeMap, error) {
	var mapping models.PartnerAttributeMap
	mappingID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid partner attribute mapping ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Unscoped().Where("id = ?", mappingID).
		Preload("Attribute").Preload("Attribute.Category").First(&mapping).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("partner attribute mapping %s not found", id)
		}
		return nil, fmt.Errorf("failed to get partner attribute mapping: %w", err)
	}
	return &mapping, nil
}

// GetByPartnerID retrieves all mappings for a partner ID (excludes soft-deleted records)
func (r *partnerAttributeMapRepository) GetByPartnerID(ctx context.Context, partnerID string) ([]models.PartnerAttributeMap, error) {
	var mappings []models.PartnerAttributeMap
	partnerUUID, err := uuid.Parse(partnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid partner ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("partner_id = ? AND is_active = ?", partnerUUID, true).
		Preload("Attribute").Preload("Attribute.Category").
		Order("created_at DESC").Find(&mappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings by partner ID: %w", err)
	}
	return mappings, nil
}

// GetByPartnerIDWithDeleted retrieves all mappings for a partner ID (includes soft-deleted records)
func (r *partnerAttributeMapRepository) GetByPartnerIDWithDeleted(ctx context.Context, partnerID string) ([]models.PartnerAttributeMap, error) {
	var mappings []models.PartnerAttributeMap
	partnerUUID, err := uuid.Parse(partnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid partner ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Unscoped().Where("partner_id = ?", partnerUUID).
		Preload("Attribute").Preload("Attribute.Category").
		Order("created_at DESC").Find(&mappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings by partner ID: %w", err)
	}
	return mappings, nil
}

// GetByPartnerCode retrieves all mappings for a partner (excludes soft-deleted records)
func (r *partnerAttributeMapRepository) GetByPartnerCode(ctx context.Context, partnerCode string) ([]models.PartnerAttributeMap, error) {
	var mappings []models.PartnerAttributeMap
	err := r.db.WithContext(ctx).Where("partner_code = ? AND is_active = ?", partnerCode, true).
		Preload("Attribute").Preload("Attribute.Category").
		Order("created_at DESC").Find(&mappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings by partner code: %w", err)
	}
	return mappings, nil
}

// GetByPartnerCodeWithDeleted retrieves all mappings for a partner (includes soft-deleted records)
func (r *partnerAttributeMapRepository) GetByPartnerCodeWithDeleted(ctx context.Context, partnerCode string) ([]models.PartnerAttributeMap, error) {
	var mappings []models.PartnerAttributeMap
	err := r.db.WithContext(ctx).Unscoped().Where("partner_code = ?", partnerCode).
		Preload("Attribute").Preload("Attribute.Category").
		Order("created_at DESC").Find(&mappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings by partner code: %w", err)
	}
	return mappings, nil
}

// GetByAttributeID retrieves all mappings for an attribute (excludes soft-deleted records)
func (r *partnerAttributeMapRepository) GetByAttributeID(ctx context.Context, attributeID string) ([]models.PartnerAttributeMap, error) {
	var mappings []models.PartnerAttributeMap
	attributeUUID, err := uuid.Parse(attributeID)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("attribute_id = ? AND is_active = ?", attributeUUID, true).
		Preload("Attribute").Preload("Attribute.Category").
		Order("created_at DESC").Find(&mappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings by attribute ID: %w", err)
	}
	return mappings, nil
}

// GetByAttributeIDWithDeleted retrieves all mappings for an attribute (includes soft-deleted records)
func (r *partnerAttributeMapRepository) GetByAttributeIDWithDeleted(ctx context.Context, attributeID string) ([]models.PartnerAttributeMap, error) {
	var mappings []models.PartnerAttributeMap
	attributeUUID, err := uuid.Parse(attributeID)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Unscoped().Where("attribute_id = ?", attributeUUID).
		Preload("Attribute").Preload("Attribute.Category").
		Order("created_at DESC").Find(&mappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings by attribute ID: %w", err)
	}
	return mappings, nil
}

// GetByAttributeCode retrieves all mappings for an attribute by code (excludes soft-deleted records)
func (r *partnerAttributeMapRepository) GetByAttributeCode(ctx context.Context, attributeCode string) ([]models.PartnerAttributeMap, error) {
	var mappings []models.PartnerAttributeMap
	err := r.db.WithContext(ctx).
		Where("attribute_code = ? AND is_active = ? AND is_deleted = ?", attributeCode, true, false).
		Preload("Attribute").Preload("Attribute.Category").
		Order("created_at DESC").Find(&mappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings by attribute code: %w", err)
	}
	return mappings, nil
}

// GetByAttributeCodeWithDeleted retrieves all mappings for an attribute by code (includes soft-deleted records)
func (r *partnerAttributeMapRepository) GetByAttributeCodeWithDeleted(ctx context.Context, attributeCode string) ([]models.PartnerAttributeMap, error) {
	var mappings []models.PartnerAttributeMap
	err := r.db.WithContext(ctx).Unscoped().
		Joins("JOIN attribute ON partner_attribute_map.attribute_id = attribute.id").
		Where("attribute.code = ?", attributeCode).
		Preload("Attribute").Preload("Attribute.Category").
		Order("partner_attribute_map.created_at DESC").Find(&mappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings by attribute code: %w", err)
	}
	return mappings, nil
}

// GetByPartnerAndAttribute retrieves a specific mapping by partner and attribute
func (r *partnerAttributeMapRepository) GetByPartnerAndAttribute(ctx context.Context, partnerCode, attributeID string) (*models.PartnerAttributeMap, error) {
	var mapping models.PartnerAttributeMap
	attributeUUID, err := uuid.Parse(attributeID)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("partner_code = ? AND attribute_id = ?", partnerCode, attributeUUID).
		Preload("Attribute").Preload("Attribute.Category").First(&mapping).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("partner attribute mapping not found")
		}
		return nil, fmt.Errorf("failed to get partner attribute mapping: %w", err)
	}
	return &mapping, nil
}

// GetAll retrieves all mappings with pagination (excludes soft-deleted records)
func (r *partnerAttributeMapRepository) GetAll(ctx context.Context, offset, limit int) ([]models.PartnerAttributeMap, int64, error) {
	var mappings []models.PartnerAttributeMap
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count partner attribute mappings: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).Where("is_active = ?", true).
		Preload("Attribute").Preload("Attribute.Category").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&mappings).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get partner attribute mappings: %w", err)
	}

	return mappings, total, nil
}

// GetAllWithDeleted retrieves all mappings with pagination (includes soft-deleted records)
func (r *partnerAttributeMapRepository) GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.PartnerAttributeMap, int64, error) {
	var mappings []models.PartnerAttributeMap
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Unscoped().Model(&models.PartnerAttributeMap{}).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count partner attribute mappings: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).Unscoped().
		Preload("Attribute").Preload("Attribute.Category").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&mappings).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get partner attribute mappings: %w", err)
	}

	return mappings, total, nil
}

// GetOnlyDeleted retrieves only soft-deleted mappings with pagination
func (r *partnerAttributeMapRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.PartnerAttributeMap, int64, error) {
	var mappings []models.PartnerAttributeMap
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Unscoped().Model(&models.PartnerAttributeMap{}).
		Where("is_deleted = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted partner attribute mappings: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).Unscoped().Where("is_deleted = ?", true).
		Preload("Attribute").Preload("Attribute.Category").
		Offset(offset).Limit(limit).
		Order("deleted_at DESC").
		Find(&mappings).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deleted partner attribute mappings: %w", err)
	}

	return mappings, total, nil
}

// GetWithFilters retrieves mappings with complex filtering
func (r *partnerAttributeMapRepository) GetWithFilters(ctx context.Context, filters *PartnerAttributeMapFilters) ([]models.PartnerAttributeMap, int64, error) {
	var mappings []models.PartnerAttributeMap
	var total int64

	// Build base query
	query := r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{})
	countQuery := r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{})

	// Apply filters
	if filters.PartnerID != nil {
		query = query.Where("partner_id = ?", *filters.PartnerID)
		countQuery = countQuery.Where("partner_id = ?", *filters.PartnerID)
	}

	if filters.PartnerCode != nil {
		query = query.Where("partner_code = ?", *filters.PartnerCode)
		countQuery = countQuery.Where("partner_code = ?", *filters.PartnerCode)
	}

	if filters.AttributeID != nil {
		query = query.Where("attribute_id = ?", *filters.AttributeID)
		countQuery = countQuery.Where("attribute_id = ?", *filters.AttributeID)
	}

	if filters.CategoryID != nil {
		query = query.Joins("JOIN attribute ON partner_attribute_map.attribute_id = attribute.id").
			Where("attribute.category_id = ?", *filters.CategoryID)
		countQuery = countQuery.Joins("JOIN attribute ON partner_attribute_map.attribute_id = attribute.id").
			Where("attribute.category_id = ?", *filters.CategoryID)
	}

	if filters.IsActive != nil {
		query = query.Where("partner_attribute_map.is_active = ?", *filters.IsActive)
		countQuery = countQuery.Where("partner_attribute_map.is_active = ?", *filters.IsActive)
	}

	// Always exclude soft-deleted records
	query = query.Where("partner_attribute_map.is_deleted = ?", false)
	countQuery = countQuery.Where("partner_attribute_map.is_deleted = ?", false)

	// Get total count
	err := countQuery.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count filtered partner attribute mappings: %w", err)
	}

	// Get paginated results
	err = query.Preload("Attribute").Preload("Attribute.Category").
		Offset(filters.Offset).Limit(filters.Limit).
		Order("partner_attribute_map.created_at DESC").
		Find(&mappings).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get filtered partner attribute mappings: %w", err)
	}

	return mappings, total, nil
}

// GetPartnerCodesByAttribute retrieves unique partner codes that have a specific attribute
func (r *partnerAttributeMapRepository) GetPartnerCodesByAttribute(ctx context.Context, attributeCode string) ([]string, error) {
	var partnerCodes []string
	err := r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Joins("JOIN attribute ON partner_attribute_map.attribute_id = attribute.id").
		Where("attribute.code = ? AND partner_attribute_map.is_active = ? AND partner_attribute_map.is_deleted = ?", attributeCode, true, false).
		Distinct("partner_code").
		Pluck("partner_code", &partnerCodes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get partner codes by attribute: %w", err)
	}
	return partnerCodes, nil
}

// GetAttributesByPartner retrieves all attributes mapped to a partner
func (r *partnerAttributeMapRepository) GetAttributesByPartner(ctx context.Context, partnerCode string) ([]models.Attribute, error) {
	var attributes []models.Attribute
	err := r.db.WithContext(ctx).Model(&models.Attribute{}).
		Joins("JOIN partner_attribute_map ON attribute.id = partner_attribute_map.attribute_id").
		Where("partner_attribute_map.partner_code = ? AND partner_attribute_map.is_active = ? AND partner_attribute_map.is_deleted = ?", partnerCode, true, false).
		Preload("Category").
		Order("attribute.created_at DESC").
		Find(&attributes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes by partner: %w", err)
	}
	return attributes, nil
}

// Create creates a new partner attribute mapping
func (r *partnerAttributeMapRepository) Create(ctx context.Context, mapping *models.PartnerAttributeMap) error {
	err := r.db.WithContext(ctx).Create(mapping).Error
	if err != nil {
		return fmt.Errorf("failed to create partner attribute mapping: %w", err)
	}
	return nil
}

// Update updates an existing partner attribute mapping
func (r *partnerAttributeMapRepository) Update(ctx context.Context, id string, mapping *models.PartnerAttributeMap) error {
	mappingID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid partner attribute mapping ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", mappingID).Save(mapping)
	if result.Error != nil {
		return fmt.Errorf("failed to update partner attribute mapping: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("partner attribute mapping not found")
	}
	return nil
}

// Delete performs soft delete on a partner attribute mapping
func (r *partnerAttributeMapRepository) Delete(ctx context.Context, id string) error {
	mappingID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid partner attribute mapping ID format: %w", err)
	}

	// Get the mapping first
	var mapping models.PartnerAttributeMap
	err = r.db.WithContext(ctx).Where("id = ?", mappingID).First(&mapping).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("partner attribute mapping not found")
		}
		return fmt.Errorf("failed to find partner attribute mapping: %w", err)
	}

	// Use the model's soft delete method
	err = mapping.SoftDelete(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to soft delete partner attribute mapping: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted partner attribute mapping
func (r *partnerAttributeMapRepository) Restore(ctx context.Context, id string) error {
	mappingID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid partner attribute mapping ID format: %w", err)
	}

	// Get the soft-deleted mapping
	var mapping models.PartnerAttributeMap
	err = r.db.WithContext(ctx).Unscoped().Where("id = ? AND is_deleted = ?", mappingID, true).First(&mapping).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("soft-deleted partner attribute mapping not found")
		}
		return fmt.Errorf("failed to find soft-deleted partner attribute mapping: %w", err)
	}

	// Use the model's restore method
	err = mapping.Restore(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to restore partner attribute mapping: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes a partner attribute mapping (hard delete)
func (r *partnerAttributeMapRepository) ForceDelete(ctx context.Context, id string) error {
	mappingID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid partner attribute mapping ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Unscoped().Where("id = ?", mappingID).Delete(&models.PartnerAttributeMap{})
	if result.Error != nil {
		return fmt.Errorf("failed to force delete partner attribute mapping: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("partner attribute mapping not found")
	}
	return nil
}

// BulkCreate creates multiple partner attribute mappings in a transaction
func (r *partnerAttributeMapRepository) BulkCreate(ctx context.Context, mappings []models.PartnerAttributeMap) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, mapping := range mappings {
			if err := tx.Create(&mapping).Error; err != nil {
				return fmt.Errorf("failed to create partner attribute mapping in bulk: %w", err)
			}
		}
		return nil
	})
}

// BulkDelete performs soft delete on multiple partner attribute mappings in a transaction
func (r *partnerAttributeMapRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			// Get the mapping first
			var mapping models.PartnerAttributeMap
			err := tx.Where("id = ?", id).First(&mapping).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					continue // Skip if not found
				}
				return fmt.Errorf("failed to find partner attribute mapping %s: %w", id, err)
			}

			// Use the model's soft delete method
			err = mapping.SoftDelete(tx)
			if err != nil {
				return fmt.Errorf("failed to soft delete partner attribute mapping %s: %w", id, err)
			}
		}
		return nil
	})
}

// GetStats retrieves statistics for a partner's attribute mappings
func (r *partnerAttributeMapRepository) GetStats(ctx context.Context, partnerCode string) (*PartnerAttributeStats, error) {
	stats := &PartnerAttributeStats{
		PartnerCode: partnerCode,
	}

	// Count total mappings
	err := r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Where("partner_code = ? AND is_deleted = ?", partnerCode, false).Count(&stats.TotalMappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total mappings: %w", err)
	}

	// Count active mappings
	err = r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Where("partner_code = ? AND is_deleted = ? AND is_active = ?", partnerCode, false, true).Count(&stats.ActiveMappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count active mappings: %w", err)
	}

	// Count unique categories
	err = r.db.WithContext(ctx).Model(&models.PartnerAttributeMap{}).
		Joins("JOIN attribute ON partner_attribute_map.attribute_id = attribute.id").
		Where("partner_attribute_map.partner_code = ? AND partner_attribute_map.is_deleted = ?", partnerCode, false).
		Distinct("attribute.category_id").
		Count(&stats.UniqueCategories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count unique categories: %w", err)
	}

	// Get first mapping date
	var firstMapping models.PartnerAttributeMap
	err = r.db.WithContext(ctx).Where("partner_code = ? AND is_deleted = ?", partnerCode, false).
		Order("created_at ASC").First(&firstMapping).Error
	if err == nil {
		firstMappingTime := firstMapping.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		stats.FirstMappingAt = &firstMappingTime
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get first mapping date: %w", err)
	}

	// Get last mapping date
	var lastMapping models.PartnerAttributeMap
	err = r.db.WithContext(ctx).Where("partner_code = ? AND is_deleted = ?", partnerCode, false).
		Order("created_at DESC").First(&lastMapping).Error
	if err == nil {
		lastMappingTime := lastMapping.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		stats.LastMappingAt = &lastMappingTime
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get last mapping date: %w", err)
	}

	// Get most used category
	var mostUsedCategoryResult struct {
		CategoryCode string
		Count        int64
	}
	err = r.db.WithContext(ctx).Raw(`
		SELECT attribute_category.code as category_code, COUNT(*) as count
		FROM partner_attribute_map 
		JOIN attribute ON partner_attribute_map.attribute_id = attribute.id
		JOIN attribute_category ON attribute.category_id = attribute_category.id
		WHERE partner_attribute_map.partner_code = ? AND partner_attribute_map.is_deleted = ?
		GROUP BY attribute_category.code
		ORDER BY count DESC
		LIMIT 1
	`, partnerCode, false).Scan(&mostUsedCategoryResult).Error

	if err == nil && mostUsedCategoryResult.CategoryCode != "" {
		stats.MostUsedCategory = &mostUsedCategoryResult.CategoryCode
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get most used category: %w", err)
	}

	// Get most used attribute
	var mostUsedAttributeResult struct {
		AttributeCode string
		Count         int64
	}
	err = r.db.WithContext(ctx).Raw(`
		SELECT attribute.code as attribute_code, COUNT(*) as count
		FROM partner_attribute_map 
		JOIN attribute ON partner_attribute_map.attribute_id = attribute.id
		WHERE partner_attribute_map.partner_code = ? AND partner_attribute_map.is_deleted = ?
		GROUP BY attribute.code
		ORDER BY count DESC
		LIMIT 1
	`, partnerCode, false).Scan(&mostUsedAttributeResult).Error

	if err == nil && mostUsedAttributeResult.AttributeCode != "" {
		stats.MostUsedAttribute = &mostUsedAttributeResult.AttributeCode
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get most used attribute: %w", err)
	}

	return stats, nil
}
