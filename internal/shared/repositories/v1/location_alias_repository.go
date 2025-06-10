package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// locationAliasRepository implements LocationAliasRepository
type locationAliasRepository struct {
	db *gorm.DB
}

// NewLocationAliasRepository creates a new location alias repository
func NewLocationAliasRepository(db *gorm.DB) LocationAliasRepository {
	return &locationAliasRepository{db: db}
}

// GetByID retrieves a location alias by its ID
func (r *locationAliasRepository) GetByID(ctx context.Context, id string) (*models.LocationAlias, error) {
	aliasID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	var alias models.LocationAlias
	err = r.db.WithContext(ctx).
		Preload("LocationType").
		Where("id = ? AND is_active = ?", aliasID, true).
		First(&alias).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("location alias %s not found", id)
		}
		return nil, fmt.Errorf("failed to get location alias: %w", err)
	}
	return &alias, nil
}

// GetByEntityTypeAndID retrieves all aliases for a specific entity type and ID
func (r *locationAliasRepository) GetByEntityTypeAndID(ctx context.Context, entityType string, entityID string) ([]models.LocationAlias, error) {
	entityUUID, err := uuid.Parse(entityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity UUID format: %w", err)
	}

	var aliases []models.LocationAlias
	err = r.db.WithContext(ctx).
		Preload("LocationType").
		Where("entity_type = ? AND entity_id = ? AND is_active = ?", entityType, entityUUID, true).
		Order("is_primary DESC, alias_name ASC").
		Find(&aliases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get location aliases: %w", err)
	}
	return aliases, nil
}

// GetByAliasName retrieves a location alias by its alias name
func (r *locationAliasRepository) GetByAliasName(ctx context.Context, aliasName string) (*models.LocationAlias, error) {
	var alias models.LocationAlias
	err := r.db.WithContext(ctx).
		Preload("LocationType").
		Where("alias_name = ? AND is_active = ?", aliasName, true).
		First(&alias).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("location alias with name '%s' not found", aliasName)
		}
		return nil, fmt.Errorf("failed to get location alias: %w", err)
	}
	return &alias, nil
}

// GetPrimaryAlias retrieves the primary alias for a specific entity type and ID
func (r *locationAliasRepository) GetPrimaryAlias(ctx context.Context, entityType string, entityID string) (*models.LocationAlias, error) {
	entityUUID, err := uuid.Parse(entityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity UUID format: %w", err)
	}

	var alias models.LocationAlias
	err = r.db.WithContext(ctx).
		Preload("LocationType").
		Where("entity_type = ? AND entity_id = ? AND is_primary = ? AND is_active = ?",
			entityType, entityUUID, true, true).
		First(&alias).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("primary alias for %s:%s not found", entityType, entityID)
		}
		return nil, fmt.Errorf("failed to get primary alias: %w", err)
	}
	return &alias, nil
}

// GetAllByEntityID retrieves all aliases for a specific entity ID regardless of type
func (r *locationAliasRepository) GetAllByEntityID(ctx context.Context, entityID string) ([]models.LocationAlias, error) {
	entityUUID, err := uuid.Parse(entityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity UUID format: %w", err)
	}

	var aliases []models.LocationAlias
	err = r.db.WithContext(ctx).
		Preload("LocationType").
		Where("entity_id = ? AND is_active = ?", entityUUID, true).
		Order("is_primary DESC, alias_name ASC").
		Find(&aliases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get location aliases: %w", err)
	}
	return aliases, nil
}

// GetAll retrieves all location aliases with pagination
func (r *locationAliasRepository) GetAll(ctx context.Context, offset, limit int) ([]models.LocationAlias, int64, error) {
	var aliases []models.LocationAlias
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.LocationAlias{}).
		Where("is_active = ?", true).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count location aliases: %w", err)
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Preload("LocationType").
		Where("is_active = ?", true).
		Order("alias_name ASC").
		Offset(offset).
		Limit(limit).
		Find(&aliases).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get location aliases: %w", err)
	}

	return aliases, total, nil
}

// Create creates a new location alias
func (r *locationAliasRepository) Create(ctx context.Context, alias *models.LocationAlias) error {
	// If setting as primary, ensure no other primary exists for this entity
	if alias.IsPrimary {
		if err := r.ensureNoPrimaryExists(ctx, alias.EntityType, alias.EntityID, nil); err != nil {
			return err
		}
	}

	err := r.db.WithContext(ctx).Create(alias).Error
	if err != nil {
		return fmt.Errorf("failed to create location alias: %w", err)
	}
	return nil
}

// Update updates an existing location alias
func (r *locationAliasRepository) Update(ctx context.Context, alias *models.LocationAlias) error {
	// If setting as primary, ensure no other primary exists for this entity
	if alias.IsPrimary {
		if err := r.ensureNoPrimaryExists(ctx, alias.EntityType, alias.EntityID, &alias.ID); err != nil {
			return err
		}
	}

	result := r.db.WithContext(ctx).Save(alias)
	if result.Error != nil {
		return fmt.Errorf("failed to update location alias: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("location alias not found")
	}
	return nil
}

// Delete soft deletes a location alias by setting is_active to false
func (r *locationAliasRepository) Delete(ctx context.Context, id string) error {
	aliasID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %w", err)
	}

	result := r.db.WithContext(ctx).
		Model(&models.LocationAlias{}).
		Where("id = ?", aliasID).
		Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("failed to delete location alias: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("location alias not found")
	}
	return nil
}

// ensureNoPrimaryExists ensures only one primary alias exists per entity
func (r *locationAliasRepository) ensureNoPrimaryExists(ctx context.Context, entityType *string, entityID uuid.UUID, excludeID *uuid.UUID) error {
	if entityType == nil {
		return nil // Can't check without entity type
	}

	query := r.db.WithContext(ctx).Model(&models.LocationAlias{}).
		Where("entity_type = ? AND entity_id = ? AND is_primary = ? AND is_active = ?",
			*entityType, entityID, true, true)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check for existing primary alias: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("a primary alias already exists for entity %s:%s", *entityType, entityID.String())
	}

	return nil
}
