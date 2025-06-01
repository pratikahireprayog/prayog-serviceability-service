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

// GetAll retrieves all location types
func (r *locationTypeRepository) GetAll(ctx context.Context) ([]models.LocationType, error) {
	var locationTypes []models.LocationType
	err := r.db.WithContext(ctx).Find(&locationTypes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get location types: %w", err)
	}
	return locationTypes, nil
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

// Delete deletes a location type
func (r *locationTypeRepository) Delete(ctx context.Context, code string) error {
	result := r.db.WithContext(ctx).Where("code = ?", code).Delete(&models.LocationType{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete location type: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("location type not found")
	}
	return nil
}
