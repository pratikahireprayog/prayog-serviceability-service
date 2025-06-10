package repositories

import (
	"context"
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/shared/models/v1"

	"gorm.io/gorm"
)

// postalCodeRepository implements PostalCodeRepository
type postalCodeRepository struct {
	db *gorm.DB
}

// NewPostalCodeRepository creates a new postal code repository
func NewPostalCodeRepository(db *gorm.DB) PostalCodeRepository {
	return &postalCodeRepository{db: db}
}

// GetByCode retrieves a postal code by its code
func (r *postalCodeRepository) GetByCode(ctx context.Context, code string) (*models.PostalCode, error) {
	var postalCode models.PostalCode

	err := r.db.WithContext(ctx).
		Preload("Country").
		Preload("Region").
		Preload("City").
		Preload("Area").
		Where("code = ? AND is_active = ?", strings.ToUpper(code), true).
		First(&postalCode).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("postal code %s not found", code)
		}
		return nil, fmt.Errorf("failed to get postal code: %w", err)
	}

	return &postalCode, nil
}

// GetByID retrieves a postal code by its ID
func (r *postalCodeRepository) GetByID(ctx context.Context, id string) (*models.PostalCode, error) {
	var postalCode models.PostalCode

	err := r.db.WithContext(ctx).
		Preload("Country").
		Preload("Region").
		Preload("City").
		Preload("Area").
		Where("id = ? AND is_active = ?", id, true).
		First(&postalCode).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("postal code with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get postal code: %w", err)
	}

	return &postalCode, nil
}

// GetAll retrieves all postal codes with pagination
func (r *postalCodeRepository) GetAll(ctx context.Context, offset, limit int) ([]models.PostalCode, int64, error) {
	var postalCodes []models.PostalCode
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Model(&models.PostalCode{}).Where("is_active = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count postal codes: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).
		Preload("Country").
		Preload("Region").
		Preload("City").
		Preload("Area").
		Where("is_active = ?", true).
		Offset(offset).
		Limit(limit).
		Find(&postalCodes).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get postal codes: %w", err)
	}

	return postalCodes, total, nil
}

// GetByLocation retrieves postal codes by location hierarchy
func (r *postalCodeRepository) GetByLocation(ctx context.Context, countryCode, regionCode, cityCode, areaCode string) ([]models.PostalCode, error) {
	var postalCodes []models.PostalCode
	query := r.db.WithContext(ctx).
		Preload("Country").
		Preload("Region").
		Preload("City").
		Preload("Area").
		Where("is_active = ?", true)

	// Add filters based on provided location codes
	if countryCode != "" {
		query = query.Where("country_code = ?", strings.ToUpper(countryCode))
	}
	if regionCode != "" {
		query = query.Where("region_code = ?", strings.ToUpper(regionCode))
	}
	if cityCode != "" {
		query = query.Where("city_code = ?", strings.ToUpper(cityCode))
	}
	if areaCode != "" {
		query = query.Where("area_code = ?", strings.ToUpper(areaCode))
	}

	err := query.Find(&postalCodes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get postal codes by location: %w", err)
	}

	return postalCodes, nil
}

// GetByCodeAndCountry retrieves a postal code by code and country
func (r *postalCodeRepository) GetByCodeAndCountry(ctx context.Context, code, countryCode string) (*models.PostalCode, error) {
	var postalCode models.PostalCode

	err := r.db.WithContext(ctx).
		Preload("Area").
		Preload("Area.City").
		Preload("Area.City.Region").
		Preload("Area.City.Region.Country").
		Joins("JOIN areas ON postal_codes.area_id = areas.id").
		Joins("JOIN cities ON areas.city_id = cities.id").
		Joins("JOIN regions ON cities.region_id = regions.id").
		Joins("JOIN countries ON regions.country_id = countries.id").
		Where("postal_codes.code = ? AND countries.code = ? AND postal_codes.is_active = ?",
			strings.ToUpper(code), strings.ToUpper(countryCode), true).
		First(&postalCode).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("postal code %s not found in country %s", code, countryCode)
		}
		return nil, fmt.Errorf("failed to get postal code: %w", err)
	}

	return &postalCode, nil
}

// GetByAreaID retrieves all postal codes for an area
func (r *postalCodeRepository) GetByAreaID(ctx context.Context, areaID string) ([]models.PostalCode, error) {
	var postalCodes []models.PostalCode

	err := r.db.WithContext(ctx).
		Preload("Country").
		Preload("Region").
		Preload("City").
		Preload("Area").
		Where("area_id = ? AND is_active = ?", areaID, true).
		Find(&postalCodes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get postal codes for area %s: %w", areaID, err)
	}

	return postalCodes, nil
}

// GetLocationHierarchy gets the complete location hierarchy for a postal code
func (r *postalCodeRepository) GetLocationHierarchy(ctx context.Context, postalCode, countryCode string) (*models.LocationHierarchy, error) {
	var result struct {
		PostalCode  string   `json:"postal_code"`
		CountryCode string   `json:"country_code"`
		CountryName string   `json:"country_name"`
		RegionCode  string   `json:"region_code"`
		RegionName  string   `json:"region_name"`
		CityCode    string   `json:"city_code"`
		CityName    string   `json:"city_name"`
		AreaCode    string   `json:"area_code"`
		AreaName    string   `json:"area_name"`
		Latitude    *float64 `json:"latitude"`
		Longitude   *float64 `json:"longitude"`
	}

	err := r.db.WithContext(ctx).
		Table("postal_codes pc").
		Select(`
			pc.code as postal_code,
			c.code as country_code,
			c.name as country_name,
			r.code as region_code,
			r.name as region_name,
			ct.code as city_code,
			ct.name as city_name,
			a.code as area_code,
			a.name as area_name,
			pc.latitude,
			pc.longitude
		`).
		Joins("JOIN areas a ON pc.area_id = a.id").
		Joins("JOIN cities ct ON a.city_id = ct.id").
		Joins("JOIN regions r ON ct.region_id = r.id").
		Joins("JOIN countries c ON r.country_id = c.id").
		Where("pc.code = ? AND c.code = ? AND pc.is_active = ?",
			strings.ToUpper(postalCode), strings.ToUpper(countryCode), true).
		First(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("postal code %s not found in country %s", postalCode, countryCode)
		}
		return nil, fmt.Errorf("failed to get location hierarchy: %w", err)
	}

	hierarchy := &models.LocationHierarchy{
		PostalCode:  result.PostalCode,
		CountryCode: result.CountryCode,
		CountryName: result.CountryName,
		RegionCode:  result.RegionCode,
		RegionName:  result.RegionName,
		CityCode:    result.CityCode,
		CityName:    result.CityName,
		AreaCode:    result.AreaCode,
		AreaName:    result.AreaName,
		Latitude:    result.Latitude,
		Longitude:   result.Longitude,
	}

	return hierarchy, nil
}

// ValidatePostalCode validates if a postal code exists and is active
func (r *postalCodeRepository) ValidatePostalCode(ctx context.Context, postalCode, countryCode string) (*models.LocationValidationResult, error) {
	hierarchy, err := r.GetLocationHierarchy(ctx, postalCode, countryCode)
	if err != nil {
		return &models.LocationValidationResult{
			IsValid: false,
			Error:   err.Error(),
		}, nil
	}

	return &models.LocationValidationResult{
		IsValid:   true,
		Hierarchy: hierarchy,
	}, nil
}

// IsActive checks if a postal code is active
func (r *postalCodeRepository) IsActive(ctx context.Context, postalCode, countryCode string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Table("postal_codes pc").
		Joins("JOIN areas a ON pc.area_id = a.id").
		Joins("JOIN cities ct ON a.city_id = ct.id").
		Joins("JOIN regions r ON ct.region_id = r.id").
		Joins("JOIN countries c ON r.country_id = c.id").
		Where("pc.code = ? AND c.code = ? AND pc.is_active = ?",
			strings.ToUpper(postalCode), strings.ToUpper(countryCode), true).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check postal code status: %w", err)
	}

	return count > 0, nil
}

// Create creates a new postal code
func (r *postalCodeRepository) Create(ctx context.Context, postalCode *models.PostalCode) error {
	err := r.db.WithContext(ctx).Create(postalCode).Error
	if err != nil {
		return fmt.Errorf("failed to create postal code: %w", err)
	}
	return nil
}

// Update updates an existing postal code
func (r *postalCodeRepository) Update(ctx context.Context, postalCode *models.PostalCode) error {
	result := r.db.WithContext(ctx).Save(postalCode)
	if result.Error != nil {
		return fmt.Errorf("failed to update postal code: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("postal code not found")
	}
	return nil
}

// Delete deletes a postal code
func (r *postalCodeRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.PostalCode{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete postal code: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("postal code with ID %s not found", id)
	}
	return nil
}

// BulkCreate creates multiple postal codes in a transaction
func (r *postalCodeRepository) BulkCreate(ctx context.Context, postalCodes []models.PostalCode) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(postalCodes, 100).Error; err != nil {
			return fmt.Errorf("failed to bulk create postal codes: %w", err)
		}
		return nil
	})
}
