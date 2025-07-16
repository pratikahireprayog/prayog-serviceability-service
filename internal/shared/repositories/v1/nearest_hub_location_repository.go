package repositories

import (
	"context"
	"fmt"
	"strconv"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"

	"gorm.io/gorm"
)

// NearestHubLocationRepository defines the interface for nearest hub location operations
type NearestHubLocationRepository interface {
	GetByPostalCode(ctx context.Context, postalCode int) (*models.NearestHubLocation, error)
	GetByPostalCodeString(ctx context.Context, postalCode string) (*models.NearestHubLocation, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.NearestHubLocation, int64, error)
	GetByFilters(ctx context.Context, filters *dtos.NearestHubLocationFilters) ([]models.NearestHubLocation, error)
	GetByInternationalHubPostalCode(ctx context.Context, hubPostalCode int) ([]models.NearestHubLocation, error)
	GetByHubCityCode(ctx context.Context, hubCityCode string) ([]models.NearestHubLocation, error)
	GetByInternationalHubCityCode(ctx context.Context, hubCityCode string) ([]models.NearestHubLocation, error)
	GetNearbyHubLocations(ctx context.Context, latitude, longitude float64, radiusKm int, limit int) ([]models.NearestHubLocation, error)
	Create(ctx context.Context, hubLocation *models.NearestHubLocation) error
	Update(ctx context.Context, postalCode int, hubLocation *models.NearestHubLocation) error
	Delete(ctx context.Context, postalCode int) error
	BulkCreate(ctx context.Context, hubLocations []models.NearestHubLocation) error
	BulkUpdate(ctx context.Context, hubLocations []models.NearestHubLocation) error
	Exists(ctx context.Context, postalCode int) (bool, error)
}

// nearestHubLocationRepository implements NearestHubLocationRepository
type nearestHubLocationRepository struct {
	db *gorm.DB
}

// NewNearestHubLocationRepository creates a new nearest hub location repository
func NewNearestHubLocationRepository(db *gorm.DB) NearestHubLocationRepository {
	return &nearestHubLocationRepository{db: db}
}

// GetByPostalCode retrieves a nearest hub location by postal code
func (r *nearestHubLocationRepository) GetByPostalCode(ctx context.Context, postalCode int) (*models.NearestHubLocation, error) {
	var hubLocation models.NearestHubLocation
	err := r.db.WithContext(ctx).Where("postal_code = ?", postalCode).First(&hubLocation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("nearest hub location for postal code %d not found", postalCode)
		}
		return nil, fmt.Errorf("failed to get nearest hub location by postal code: %w", err)
	}
	return &hubLocation, nil
}

// GetByPostalCodeString retrieves a nearest hub location by postal code as string
func (r *nearestHubLocationRepository) GetByPostalCodeString(ctx context.Context, postalCode string) (*models.NearestHubLocation, error) {
	// Convert string to int
	postalCodeInt, err := strconv.Atoi(postalCode)
	if err != nil {
		return nil, fmt.Errorf("invalid postal code format: %s", postalCode)
	}

	return r.GetByPostalCode(ctx, postalCodeInt)
}

// GetByFilters retrieves nearest hub locations based on filters
func (r *nearestHubLocationRepository) GetByFilters(ctx context.Context, filters *dtos.NearestHubLocationFilters) ([]models.NearestHubLocation, error) {
	var hubLocations []models.NearestHubLocation
	query := r.db.WithContext(ctx)

	// Apply postal code filters if provided
	if len(filters.PostalCodes) > 0 {
		query = query.Where("postal_code IN ?", filters.PostalCodes)
	}

	// Apply international hub postal code filters if provided
	if len(filters.InternationalHubPostalCodes) > 0 {
		query = query.Where("international_hub_postal_code IN ?", filters.InternationalHubPostalCodes)
	}

	// Apply international hub city code filters if provided
	if len(filters.InternationalHubCityCodes) > 0 {
		query = query.Where("international_hub_city_code IN ?", filters.InternationalHubCityCodes)
	}

	// Apply hub city filters if provided
	if len(filters.HubCities) > 0 {
		query = query.Where("hub_city IN ?", filters.HubCities)
	}

	// Apply hub state filters if provided
	if len(filters.HubStates) > 0 {
		query = query.Where("hub_state IN ?", filters.HubStates)
	}

	// Apply hub country filters if provided
	if len(filters.HubCountries) > 0 {
		query = query.Where("hub_country IN ?", filters.HubCountries)
	}

	err := query.Find(&hubLocations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get nearest hub locations by filters: %w", err)
	}

	return hubLocations, nil
}

// GetAll retrieves all nearest hub locations with pagination
func (r *nearestHubLocationRepository) GetAll(ctx context.Context, offset, limit int) ([]models.NearestHubLocation, int64, error) {
	var hubLocations []models.NearestHubLocation
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.NearestHubLocation{}).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count nearest hub locations: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Find(&hubLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get nearest hub locations: %w", err)
	}

	return hubLocations, total, nil
}

// GetByInternationalHubPostalCode retrieves all locations served by a specific international hub
func (r *nearestHubLocationRepository) GetByInternationalHubPostalCode(ctx context.Context, hubPostalCode int) ([]models.NearestHubLocation, error) {
	var hubLocations []models.NearestHubLocation
	err := r.db.WithContext(ctx).
		Where("international_hub_postal_code = ?", hubPostalCode).
		Find(&hubLocations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get hub locations by international hub postal code: %w", err)
	}
	return hubLocations, nil
}

// GetByHubCityCode retrieves all locations by hub city code
func (r *nearestHubLocationRepository) GetByHubCityCode(ctx context.Context, hubCityCode string) ([]models.NearestHubLocation, error) {
	var hubLocations []models.NearestHubLocation
	err := r.db.WithContext(ctx).
		Where("hub_city_code = ?", hubCityCode).
		Find(&hubLocations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get hub locations by hub city code: %w", err)
	}
	return hubLocations, nil
}

// GetByInternationalHubCityCode retrieves all locations by international hub city code
func (r *nearestHubLocationRepository) GetByInternationalHubCityCode(ctx context.Context, hubCityCode string) ([]models.NearestHubLocation, error) {
	var hubLocations []models.NearestHubLocation
	err := r.db.WithContext(ctx).
		Where("international_hub_city_code = ?", hubCityCode).
		Find(&hubLocations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get hub locations by international hub city code: %w", err)
	}
	return hubLocations, nil
}

// GetNearbyHubLocations retrieves hub locations within a specified radius
func (r *nearestHubLocationRepository) GetNearbyHubLocations(ctx context.Context, latitude, longitude float64, radiusKm int, limit int) ([]models.NearestHubLocation, error) {
	var hubLocations []models.NearestHubLocation

	// Using Haversine formula for distance calculation
	// This is a simplified approach - in production, you might want to use PostGIS or similar
	query := `
		SELECT *, 
		(6371 * acos(cos(radians(?)) * cos(radians(centroid_lat)) * cos(radians(centroid_lng) - radians(?)) + sin(radians(?)) * sin(radians(centroid_lat)))) AS distance
		FROM nearest_hub_locations
		WHERE centroid_lat IS NOT NULL AND centroid_lng IS NOT NULL
		HAVING distance < ?
		ORDER BY distance
		LIMIT ?
	`

	err := r.db.WithContext(ctx).Raw(query, latitude, longitude, latitude, radiusKm, limit).Scan(&hubLocations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get nearby hub locations: %w", err)
	}

	return hubLocations, nil
}

// Create creates a new nearest hub location
func (r *nearestHubLocationRepository) Create(ctx context.Context, hubLocation *models.NearestHubLocation) error {
	err := r.db.WithContext(ctx).Create(hubLocation).Error
	if err != nil {
		return fmt.Errorf("failed to create nearest hub location: %w", err)
	}
	return nil
}

// Update updates an existing nearest hub location
func (r *nearestHubLocationRepository) Update(ctx context.Context, postalCode int, hubLocation *models.NearestHubLocation) error {
	result := r.db.WithContext(ctx).Where("postal_code = ?", postalCode).Save(hubLocation)
	if result.Error != nil {
		return fmt.Errorf("failed to update nearest hub location: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("nearest hub location with postal code %d not found", postalCode)
	}
	return nil
}

// Delete deletes a nearest hub location
func (r *nearestHubLocationRepository) Delete(ctx context.Context, postalCode int) error {
	result := r.db.WithContext(ctx).Where("postal_code = ?", postalCode).Delete(&models.NearestHubLocation{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete nearest hub location: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("nearest hub location with postal code %d not found", postalCode)
	}
	return nil
}

// BulkCreate creates multiple nearest hub locations
func (r *nearestHubLocationRepository) BulkCreate(ctx context.Context, hubLocations []models.NearestHubLocation) error {
	err := r.db.WithContext(ctx).CreateInBatches(hubLocations, 100).Error
	if err != nil {
		return fmt.Errorf("failed to bulk create nearest hub locations: %w", err)
	}
	return nil
}

// BulkUpdate updates multiple nearest hub locations
func (r *nearestHubLocationRepository) BulkUpdate(ctx context.Context, hubLocations []models.NearestHubLocation) error {
	for _, hubLocation := range hubLocations {
		err := r.Update(ctx, hubLocation.PostalCode, &hubLocation)
		if err != nil {
			return fmt.Errorf("failed to bulk update nearest hub location %d: %w", hubLocation.PostalCode, err)
		}
	}
	return nil
}

// Exists checks if a nearest hub location exists for the given postal code
func (r *nearestHubLocationRepository) Exists(ctx context.Context, postalCode int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.NearestHubLocation{}).
		Where("postal_code = ?", postalCode).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check nearest hub location existence: %w", err)
	}
	return count > 0, nil
}
