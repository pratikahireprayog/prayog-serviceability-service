package repositories

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// countryRepository implements CountryRepository
type countryRepository struct {
	db *gorm.DB
}

// NewCountryRepository creates a new country repository
func NewCountryRepository(db *gorm.DB) CountryRepository {
	return &countryRepository{db: db}
}

// GetByID retrieves a country by its ID (excludes soft-deleted records)
func (r *countryRepository) GetByID(ctx context.Context, id string) (*models.Country, error) {
	var country models.Country
	countryID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid country ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Where("id = ? AND is_active = ?", countryID, true).First(&country).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("country %s not found", id)
		}
		return nil, fmt.Errorf("failed to get country: %w", err)
	}
	return &country, nil
}

// GetByIDWithDeleted retrieves a country by its ID (includes soft-deleted records)
func (r *countryRepository) GetByIDWithDeleted(ctx context.Context, id string) (*models.Country, error) {
	var country models.Country
	countryID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid country ID format: %w", err)
	}

	err = r.db.WithContext(ctx).Unscoped().Where("id = ? AND is_active = ?", countryID, true).First(&country).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("country %s not found", id)
		}
		return nil, fmt.Errorf("failed to get country: %w", err)
	}
	return &country, nil
}

// GetByCode retrieves a country by its code (excludes soft-deleted records)
func (r *countryRepository) GetByCode(ctx context.Context, code string) (*models.Country, error) {
	var country models.Country
	err := r.db.WithContext(ctx).Where("code = ? AND is_active = ?", code, true).First(&country).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("country %s not found", code)
		}
		return nil, fmt.Errorf("failed to get country: %w", err)
	}
	return &country, nil
}

// GetByCodeWithDeleted retrieves a country by its code (includes soft-deleted records)
func (r *countryRepository) GetByCodeWithDeleted(ctx context.Context, code string) (*models.Country, error) {
	var country models.Country
	err := r.db.WithContext(ctx).Unscoped().Where("code = ? AND is_active = ?", code, true).First(&country).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("country %s not found", code)
		}
		return nil, fmt.Errorf("failed to get country: %w", err)
	}
	return &country, nil
}

// GetAll retrieves all countries with pagination (excludes soft-deleted records)
func (r *countryRepository) GetAll(ctx context.Context, offset, limit int) ([]models.Country, int64, error) {
	var countries []models.Country
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.Country{}).Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count countries: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Where("is_active = ?", true).
		Offset(offset).Limit(limit).Find(&countries).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get countries: %w", err)
	}

	return countries, total, nil
}

// GetAllWithDeleted retrieves all countries with pagination (includes soft-deleted records)
func (r *countryRepository) GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.Country, int64, error) {
	var countries []models.Country
	var total int64

	// Count total records including soft-deleted
	err := r.db.WithContext(ctx).Unscoped().Model(&models.Country{}).Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count countries: %w", err)
	}

	// Get paginated records including soft-deleted
	err = r.db.WithContext(ctx).Unscoped().Where("is_active = ?", true).
		Offset(offset).Limit(limit).Find(&countries).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get countries: %w", err)
	}

	return countries, total, nil
}

// GetOnlyDeleted retrieves only soft-deleted countries with pagination
func (r *countryRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.Country, int64, error) {
	var countries []models.Country
	var total int64

	// Count only soft-deleted records
	err := r.db.WithContext(ctx).Unscoped().Model(&models.Country{}).Where("is_deleted = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted countries: %w", err)
	}

	// Get paginated soft-deleted records
	err = r.db.WithContext(ctx).Unscoped().Where("is_deleted = ?", true).
		Offset(offset).Limit(limit).Find(&countries).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deleted countries: %w", err)
	}

	return countries, total, nil
}

// Create creates a new country
func (r *countryRepository) Create(ctx context.Context, country *models.Country) error {
	err := r.db.WithContext(ctx).Create(country).Error
	if err != nil {
		return fmt.Errorf("failed to create country: %w", err)
	}
	return nil
}

// Update updates an existing country
func (r *countryRepository) Update(ctx context.Context, id string, country *models.Country) error {
	countryID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid country ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", countryID).Save(country)
	if result.Error != nil {
		return fmt.Errorf("failed to update country: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("country not found")
	}
	return nil
}

// Delete performs soft delete on a country
func (r *countryRepository) Delete(ctx context.Context, id string) error {
	countryID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid country ID format: %w", err)
	}

	// Get the country first
	var country models.Country
	err = r.db.WithContext(ctx).Where("id = ?", countryID).First(&country).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("country not found")
		}
		return fmt.Errorf("failed to find country: %w", err)
	}

	// Use the model's soft delete method
	err = country.SoftDelete(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to soft delete country: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted country
func (r *countryRepository) Restore(ctx context.Context, id string) error {
	countryID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid country ID format: %w", err)
	}

	// Get the soft-deleted country
	var country models.Country
	err = r.db.WithContext(ctx).Unscoped().Where("id = ? AND is_deleted = ?", countryID, true).First(&country).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("soft-deleted country not found")
		}
		return fmt.Errorf("failed to find soft-deleted country: %w", err)
	}

	// Use the model's restore method
	err = country.Restore(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to restore country: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes a country (hard delete)
func (r *countryRepository) ForceDelete(ctx context.Context, id string) error {
	countryID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid country ID format: %w", err)
	}

	result := r.db.WithContext(ctx).Unscoped().Where("id = ?", countryID).Delete(&models.Country{})
	if result.Error != nil {
		return fmt.Errorf("failed to force delete country: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("country not found")
	}
	return nil
}
