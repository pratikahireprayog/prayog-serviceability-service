package gorm

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"gorm.io/gorm"
)

// countryRepository implements repositories.CountryRepository
type countryRepository struct {
	db *gorm.DB
}

// NewCountryRepository creates a new country repository
func NewCountryRepository(db *gorm.DB) repositories.CountryRepository {
	return &countryRepository{db: db}
}

// GetByCode retrieves a country by its code
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

// GetAll retrieves all countries
func (r *countryRepository) GetAll(ctx context.Context) ([]models.Country, error) {
	var countries []models.Country
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&countries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get countries: %w", err)
	}
	return countries, nil
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
func (r *countryRepository) Update(ctx context.Context, country *models.Country) error {
	result := r.db.WithContext(ctx).Save(country)
	if result.Error != nil {
		return fmt.Errorf("failed to update country: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("country not found")
	}
	return nil
}

// Delete deletes a country
func (r *countryRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Country{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete country: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("country not found")
	}
	return nil
}
