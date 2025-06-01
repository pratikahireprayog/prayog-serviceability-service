package gorm

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"gorm.io/gorm"
)

// locationAliasRepository implements repositories.LocationAliasRepository
type locationAliasRepository struct {
	db *gorm.DB
}

// NewLocationAliasRepository creates a new location alias repository
func NewLocationAliasRepository(db *gorm.DB) repositories.LocationAliasRepository {
	return &locationAliasRepository{db: db}
}

// GetByID retrieves a location alias by its ID
func (r *locationAliasRepository) GetByID(ctx context.Context, id uint) (*models.LocationAlias, error) {
	var alias models.LocationAlias
	err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&alias).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("location alias %d not found", id)
		}
		return nil, fmt.Errorf("failed to get location alias: %w", err)
	}
	return &alias, nil
}

// GetByEntityTypeAndID retrieves all aliases for a specific entity type and ID
func (r *locationAliasRepository) GetByEntityTypeAndID(ctx context.Context, entityType string, entityID uint) ([]models.LocationAlias, error) {
	var aliases []models.LocationAlias
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ? AND is_active = ?", entityType, entityID, true).
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
func (r *locationAliasRepository) GetPrimaryAlias(ctx context.Context, entityType string, entityID uint) (*models.LocationAlias, error) {
	var alias models.LocationAlias
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ? AND is_primary = ? AND is_active = ?",
			entityType, entityID, true, true).
		First(&alias).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("primary alias for %s:%d not found", entityType, entityID)
		}
		return nil, fmt.Errorf("failed to get primary alias: %w", err)
	}
	return &alias, nil
}

// Create creates a new location alias
func (r *locationAliasRepository) Create(ctx context.Context, alias *models.LocationAlias) error {
	err := r.db.WithContext(ctx).Create(alias).Error
	if err != nil {
		return fmt.Errorf("failed to create location alias: %w", err)
	}
	return nil
}

// Update updates an existing location alias
func (r *locationAliasRepository) Update(ctx context.Context, alias *models.LocationAlias) error {
	result := r.db.WithContext(ctx).Save(alias)
	if result.Error != nil {
		return fmt.Errorf("failed to update location alias: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("location alias not found")
	}
	return nil
}

// Delete deletes a location alias
func (r *locationAliasRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.LocationAlias{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete location alias: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("location alias not found")
	}
	return nil
}
