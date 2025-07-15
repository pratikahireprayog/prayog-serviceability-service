package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"

	"gorm.io/gorm"
)

// GeoLocationRepository defines the interface for geo location operations
type GeoLocationRepository interface {
	GetByID(ctx context.Context, postalCode string) (*models.GeoLocation, error)
	GetByIDWithDeleted(ctx context.Context, postalCode string) (*models.GeoLocation, error)
	GetAll(ctx context.Context, offset, limit int, filters *dtos.GeoLocationFilters) ([]models.GeoLocation, int64, error)
	GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.GeoLocation, int64, error)
	GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.GeoLocation, int64, error)
	GetByCountryCode(ctx context.Context, countryCode string, offset, limit int) ([]models.GeoLocation, int64, error)
	GetByCountryCodeWithDeleted(ctx context.Context, countryCode string, offset, limit int) ([]models.GeoLocation, int64, error)
	Search(ctx context.Context, filters *dtos.GeoLocationSearchFilters) ([]models.GeoLocation, int64, error)
	SearchWithDeleted(ctx context.Context, filters *dtos.GeoLocationSearchFilters) ([]models.GeoLocation, int64, error)
	GetStats(ctx context.Context) (*dtos.GeoLocationStatsResponse, error)
	GetByFeatureClass(ctx context.Context, featureClass string, offset, limit int) ([]models.GeoLocation, int64, error)
	GetByFeatureCode(ctx context.Context, featureCode string, offset, limit int) ([]models.GeoLocation, int64, error)
	GetByTimezone(ctx context.Context, timezone string, offset, limit int) ([]models.GeoLocation, int64, error)
	GetInBoundingBox(ctx context.Context, minLat, maxLat, minLng, maxLng float64, offset, limit int) ([]models.GeoLocation, int64, error)
	Create(ctx context.Context, geoLocation *models.GeoLocation) error
	Update(ctx context.Context, postalCode string, geoLocation *models.GeoLocation) error
	Delete(ctx context.Context, postalCode string) error
	Restore(ctx context.Context, postalCode string) error
	ForceDelete(ctx context.Context, postalCode string) error
	BulkCreate(ctx context.Context, geoLocations []models.GeoLocation) error
	BulkUpdate(ctx context.Context, geoLocations []models.GeoLocation) error
}

type geoLocationRepository struct {
	db *gorm.DB
}

// NewGeoLocationRepository creates a new geo location repository
func NewGeoLocationRepository(db *gorm.DB) GeoLocationRepository {
	return &geoLocationRepository{db: db}
}

// GetByID retrieves a geo location by ID
func (r *geoLocationRepository) GetByID(ctx context.Context, postalCode string) (*models.GeoLocation, error) {
	var geoLocation models.GeoLocation
	
	// Use optimized query with limit and specific columns to improve performance
	err := r.db.WithContext(ctx).
		Select("postal_code, country_code, name, latitude, longitude, created_at, updated_at").
		Where("postal_code = ? AND deleted_at IS NULL", postalCode).
		Order("created_at DESC").
		Limit(1).
		First(&geoLocation).Error
		
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("geo location with postal code %s not found", postalCode)
		}
		return nil, fmt.Errorf("failed to get geo location by postal code: %w", err)
	}
	return &geoLocation, nil
}

// GetByIDWithDeleted retrieves a geo location by ID including soft-deleted records
func (r *geoLocationRepository) GetByIDWithDeleted(ctx context.Context, postalCode string) (*models.GeoLocation, error) {
	var geoLocation models.GeoLocation
	err := r.db.WithContext(ctx).Unscoped().Where("postal_code = ?", postalCode).First(&geoLocation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("geo location with postal code %s not found", postalCode)
		}
		return nil, fmt.Errorf("failed to get geo location by postal code: %w", err)
	}
	return &geoLocation, nil
}

// GetAll retrieves all geo locations with pagination (excludes soft-deleted records)
func (r *geoLocationRepository) GetAll(ctx context.Context, offset, limit int, filters *dtos.GeoLocationFilters) ([]models.GeoLocation, int64, error) {
	var geoLocations []models.GeoLocation
	var total int64

	query := r.db.WithContext(ctx).Model(&models.GeoLocation{})

	// Apply filters if provided
	if filters != nil {
		if filters.CountryCode != nil {
			query = query.Where("country_code = ?", strings.ToUpper(*filters.CountryCode))
		}
		if len(filters.PostalCodes) > 0 {
			query = query.Where("postal_code IN ?", filters.PostalCodes)
		}
		if filters.Name != nil {
			query = query.Where("name ILIKE ?", "%"+*filters.Name+"%")
		}
		if filters.FeatureCode != nil {
			query = query.Where("feature_code = ?", *filters.FeatureCode)
		}
	}

	// Count total records
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count geo locations: %w", err)
	}

	// Get paginated records
	err = query.Offset(offset).Limit(limit).Find(&geoLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get geo locations: %w", err)
	}

	return geoLocations, total, nil
}

// GetAllWithDeleted retrieves all geo locations with pagination (includes soft-deleted records)
func (r *geoLocationRepository) GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.GeoLocation, int64, error) {
	var geoLocations []models.GeoLocation
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Unscoped().Model(&models.GeoLocation{}).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count geo locations: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Unscoped().Offset(offset).Limit(limit).Find(&geoLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get geo locations: %w", err)
	}

	return geoLocations, total, nil
}

// GetOnlyDeleted retrieves only soft-deleted geo locations with pagination
func (r *geoLocationRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.GeoLocation, int64, error) {
	var geoLocations []models.GeoLocation
	var total int64

	// Count total soft-deleted records
	err := r.db.WithContext(ctx).Unscoped().Model(&models.GeoLocation{}).Where("is_deleted = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count soft-deleted geo locations: %w", err)
	}

	// Get paginated soft-deleted records
	err = r.db.WithContext(ctx).Unscoped().Where("is_deleted = ?", true).Offset(offset).Limit(limit).Find(&geoLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get soft-deleted geo locations: %w", err)
	}

	return geoLocations, total, nil
}

// GetByCountryCode retrieves geo locations by country code with pagination
func (r *geoLocationRepository) GetByCountryCode(ctx context.Context, countryCode string, offset, limit int) ([]models.GeoLocation, int64, error) {
	var geoLocations []models.GeoLocation
	var total int64

	countryCode = strings.ToUpper(countryCode)

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.GeoLocation{}).Where("country_code = ?", countryCode).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count geo locations by country: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Where("country_code = ?", countryCode).Offset(offset).Limit(limit).Find(&geoLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get geo locations by country: %w", err)
	}

	return geoLocations, total, nil
}

// GetByCountryCodeWithDeleted retrieves geo locations by country code with pagination (includes soft-deleted)
func (r *geoLocationRepository) GetByCountryCodeWithDeleted(ctx context.Context, countryCode string, offset, limit int) ([]models.GeoLocation, int64, error) {
	var geoLocations []models.GeoLocation
	var total int64

	countryCode = strings.ToUpper(countryCode)

	// Count total records
	err := r.db.WithContext(ctx).Unscoped().Model(&models.GeoLocation{}).Where("country_code = ?", countryCode).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count geo locations by country: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Unscoped().Where("country_code = ?", countryCode).Offset(offset).Limit(limit).Find(&geoLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get geo locations by country: %w", err)
	}

	return geoLocations, total, nil
}

// Search performs advanced search with multiple filters
func (r *geoLocationRepository) Search(ctx context.Context, filters *dtos.GeoLocationSearchFilters) ([]models.GeoLocation, int64, error) {
	var geoLocations []models.GeoLocation
	var total int64

	query := r.db.WithContext(ctx).Model(&models.GeoLocation{})

	// Apply filters
	query = r.applySearchFilters(query, filters)

	// Count total records
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count geo locations: %w", err)
	}

	// Get paginated records
	err = query.Offset(filters.Offset).Limit(filters.Limit).Find(&geoLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search geo locations: %w", err)
	}

	return geoLocations, total, nil
}

// SearchWithDeleted performs advanced search with multiple filters (includes soft-deleted)
func (r *geoLocationRepository) SearchWithDeleted(ctx context.Context, filters *dtos.GeoLocationSearchFilters) ([]models.GeoLocation, int64, error) {
	var geoLocations []models.GeoLocation
	var total int64

	query := r.db.WithContext(ctx).Unscoped().Model(&models.GeoLocation{})

	// Apply filters
	query = r.applySearchFilters(query, filters)

	// Count total records
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count geo locations: %w", err)
	}

	// Get paginated records
	err = query.Offset(filters.Offset).Limit(filters.Limit).Find(&geoLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search geo locations: %w", err)
	}

	return geoLocations, total, nil
}

// applySearchFilters applies search filters to the query
func (r *geoLocationRepository) applySearchFilters(query *gorm.DB, filters *dtos.GeoLocationSearchFilters) *gorm.DB {
	if filters.CountryCode != nil {
		query = query.Where("country_code = ?", strings.ToUpper(*filters.CountryCode))
	}

	if filters.FeatureClass != nil {
		query = query.Where("feature_class = ?", *filters.FeatureClass)
	}

	if filters.FeatureCode != nil {
		query = query.Where("feature_code = ?", *filters.FeatureCode)
	}

	if filters.Admin1Code != nil {
		query = query.Where("admin1_code = ?", *filters.Admin1Code)
	}

	if filters.Admin2Code != nil {
		query = query.Where("admin2_code = ?", *filters.Admin2Code)
	}

	if filters.Admin3Code != nil {
		query = query.Where("admin3_code = ?", *filters.Admin3Code)
	}

	if filters.Admin4Code != nil {
		query = query.Where("admin4_code = ?", *filters.Admin4Code)
	}

	if filters.MinLatitude != nil {
		query = query.Where("latitude >= ?", *filters.MinLatitude)
	}

	if filters.MaxLatitude != nil {
		query = query.Where("latitude <= ?", *filters.MaxLatitude)
	}

	if filters.MinLongitude != nil {
		query = query.Where("longitude >= ?", *filters.MinLongitude)
	}

	if filters.MaxLongitude != nil {
		query = query.Where("longitude <= ?", *filters.MaxLongitude)
	}

	if filters.MinPopulation != nil {
		query = query.Where("population >= ?", *filters.MinPopulation)
	}

	if filters.MaxPopulation != nil {
		query = query.Where("population <= ?", *filters.MaxPopulation)
	}

	if filters.MinElevation != nil {
		query = query.Where("elevation >= ?", *filters.MinElevation)
	}

	if filters.MaxElevation != nil {
		query = query.Where("elevation <= ?", *filters.MaxElevation)
	}

	if filters.Timezone != nil {
		query = query.Where("timezone = ?", *filters.Timezone)
	}

	if filters.NameSearch != nil {
		searchTerm := "%" + *filters.NameSearch + "%"
		query = query.Where("name ILIKE ? OR asciiname ILIKE ? OR alternatenames ILIKE ?", searchTerm, searchTerm, searchTerm)
	}

	return query
}

// GetStats retrieves statistics for geo locations
func (r *geoLocationRepository) GetStats(ctx context.Context) (*dtos.GeoLocationStatsResponse, error) {
	var stats dtos.GeoLocationStatsResponse
	var err error

	// Get total count
	err = r.db.WithContext(ctx).Model(&models.GeoLocation{}).Count(&stats.TotalCount).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get country stats
	var countryStats []struct {
		CountryCode string `json:"country_code"`
		Count       int64  `json:"count"`
	}
	err = r.db.WithContext(ctx).Model(&models.GeoLocation{}).
		Select("country_code, COUNT(*) as count").
		Group("country_code").
		Find(&countryStats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get country stats: %w", err)
	}

	stats.CountryStats = make(map[string]int64)
	for _, stat := range countryStats {
		stats.CountryStats[stat.CountryCode] = stat.Count
	}

	// Get feature class stats
	var featureClassStats []struct {
		FeatureClass string `json:"feature_class"`
		Count        int64  `json:"count"`
	}
	err = r.db.WithContext(ctx).Model(&models.GeoLocation{}).
		Select("feature_class, COUNT(*) as count").
		Group("feature_class").
		Find(&featureClassStats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get feature class stats: %w", err)
	}

	stats.FeatureClassStats = make(map[string]int64)
	for _, stat := range featureClassStats {
		stats.FeatureClassStats[stat.FeatureClass] = stat.Count
	}

	// Get feature code stats
	var featureCodeStats []struct {
		FeatureCode string `json:"feature_code"`
		Count       int64  `json:"count"`
	}
	err = r.db.WithContext(ctx).Model(&models.GeoLocation{}).
		Select("feature_code, COUNT(*) as count").
		Group("feature_code").
		Find(&featureCodeStats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get feature code stats: %w", err)
	}

	stats.FeatureCodeStats = make(map[string]int64)
	for _, stat := range featureCodeStats {
		stats.FeatureCodeStats[stat.FeatureCode] = stat.Count
	}

	// Get timezone stats
	var timezoneStats []struct {
		Timezone string `json:"timezone"`
		Count    int64  `json:"count"`
	}
	err = r.db.WithContext(ctx).Model(&models.GeoLocation{}).
		Select("timezone, COUNT(*) as count").
		Where("timezone IS NOT NULL").
		Group("timezone").
		Find(&timezoneStats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get timezone stats: %w", err)
	}

	stats.TimezoneStats = make(map[string]int64)
	for _, stat := range timezoneStats {
		stats.TimezoneStats[stat.Timezone] = stat.Count
	}

	// Get average coordinates
	var avgCoords struct {
		AvgLat float64 `json:"avg_lat"`
		AvgLng float64 `json:"avg_lng"`
	}
	err = r.db.WithContext(ctx).Model(&models.GeoLocation{}).
		Select("AVG(latitude) as avg_lat, AVG(longitude) as avg_lng").
		First(&avgCoords).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get average coordinates: %w", err)
	}

	stats.AverageLatitude = avgCoords.AvgLat
	stats.AverageLongitude = avgCoords.AvgLng

	// Get population stats
	var popStats struct {
		Total   int64   `json:"total"`
		Average float64 `json:"average"`
		Min     int64   `json:"min"`
		Max     int64   `json:"max"`
		Count   int64   `json:"count"`
	}
	err = r.db.WithContext(ctx).Model(&models.GeoLocation{}).
		Select("SUM(population) as total, AVG(population) as average, MIN(population) as min, MAX(population) as max, COUNT(*) as count").
		Where("population IS NOT NULL").
		First(&popStats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get population stats: %w", err)
	}

	if popStats.Count > 0 {
		stats.PopulationStats = &dtos.PopulationStatsResponse{
			Total:    popStats.Total,
			Average:  popStats.Average,
			Min:      popStats.Min,
			Max:      popStats.Max,
			WithData: popStats.Count,
		}
	}

	// Get elevation stats
	var elevStats struct {
		Average float64 `json:"average"`
		Min     int     `json:"min"`
		Max     int     `json:"max"`
		Count   int64   `json:"count"`
	}
	err = r.db.WithContext(ctx).Model(&models.GeoLocation{}).
		Select("AVG(elevation) as average, MIN(elevation) as min, MAX(elevation) as max, COUNT(*) as count").
		Where("elevation IS NOT NULL").
		First(&elevStats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get elevation stats: %w", err)
	}

	if elevStats.Count > 0 {
		stats.ElevationStats = &dtos.ElevationStatsResponse{
			Average:  elevStats.Average,
			Min:      elevStats.Min,
			Max:      elevStats.Max,
			WithData: elevStats.Count,
		}
	}

	return &stats, nil
}

// GetByFeatureClass retrieves geo locations by feature class
func (r *geoLocationRepository) GetByFeatureClass(ctx context.Context, featureClass string, offset, limit int) ([]models.GeoLocation, int64, error) {
	var geoLocations []models.GeoLocation
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.GeoLocation{}).Where("feature_class = ?", featureClass).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count geo locations by feature class: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Where("feature_class = ?", featureClass).Offset(offset).Limit(limit).Find(&geoLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get geo locations by feature class: %w", err)
	}

	return geoLocations, total, nil
}

// GetByFeatureCode retrieves geo locations by feature code
func (r *geoLocationRepository) GetByFeatureCode(ctx context.Context, featureCode string, offset, limit int) ([]models.GeoLocation, int64, error) {
	var geoLocations []models.GeoLocation
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.GeoLocation{}).Where("feature_code = ?", featureCode).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count geo locations by feature code: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Where("feature_code = ?", featureCode).Offset(offset).Limit(limit).Find(&geoLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get geo locations by feature code: %w", err)
	}

	return geoLocations, total, nil
}

// GetByTimezone retrieves geo locations by timezone
func (r *geoLocationRepository) GetByTimezone(ctx context.Context, timezone string, offset, limit int) ([]models.GeoLocation, int64, error) {
	var geoLocations []models.GeoLocation
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.GeoLocation{}).Where("timezone = ?", timezone).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count geo locations by timezone: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).Where("timezone = ?", timezone).Offset(offset).Limit(limit).Find(&geoLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get geo locations by timezone: %w", err)
	}

	return geoLocations, total, nil
}

// GetInBoundingBox retrieves geo locations within a bounding box
func (r *geoLocationRepository) GetInBoundingBox(ctx context.Context, minLat, maxLat, minLng, maxLng float64, offset, limit int) ([]models.GeoLocation, int64, error) {
	var geoLocations []models.GeoLocation
	var total int64

	// Count total records
	err := r.db.WithContext(ctx).Model(&models.GeoLocation{}).
		Where("latitude >= ? AND latitude <= ? AND longitude >= ? AND longitude <= ?", minLat, maxLat, minLng, maxLng).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count geo locations in bounding box: %w", err)
	}

	// Get paginated records
	err = r.db.WithContext(ctx).
		Where("latitude >= ? AND latitude <= ? AND longitude >= ? AND longitude <= ?", minLat, maxLat, minLng, maxLng).
		Offset(offset).Limit(limit).Find(&geoLocations).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get geo locations in bounding box: %w", err)
	}

	return geoLocations, total, nil
}

// Create creates a new geo location
func (r *geoLocationRepository) Create(ctx context.Context, geoLocation *models.GeoLocation) error {
	err := r.db.WithContext(ctx).Create(geoLocation).Error
	if err != nil {
		return fmt.Errorf("failed to create geo location: %w", err)
	}
	return nil
}

// Update updates a geo location
func (r *geoLocationRepository) Update(ctx context.Context, postalCode string, geoLocation *models.GeoLocation) error {
	result := r.db.WithContext(ctx).Where("postal_code = ?", postalCode).Save(geoLocation)
	if result.Error != nil {
		return fmt.Errorf("failed to update geo location: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("geo location not found")
	}
	return nil
}

// Delete performs soft delete on a geo location
func (r *geoLocationRepository) Delete(ctx context.Context, postalCode string) error {
	geoLocation, err := r.GetByID(ctx, postalCode)
	if err != nil {
		return err
	}

	err = geoLocation.SoftDelete(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to soft delete geo location: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted geo location
func (r *geoLocationRepository) Restore(ctx context.Context, postalCode string) error {
	geoLocation, err := r.GetByIDWithDeleted(ctx, postalCode)
	if err != nil {
		return err
	}

	if !geoLocation.IsDeletedRecord() {
		return fmt.Errorf("geo location is not deleted")
	}

	err = geoLocation.Restore(r.db.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to restore geo location: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes a geo location
func (r *geoLocationRepository) ForceDelete(ctx context.Context, postalCode string) error {
	result := r.db.WithContext(ctx).Unscoped().Where("postal_code = ?", postalCode).Delete(&models.GeoLocation{})
	if result.Error != nil {
		return fmt.Errorf("failed to force delete geo location: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("geo location not found")
	}
	return nil
}

// BulkCreate creates multiple geo locations
func (r *geoLocationRepository) BulkCreate(ctx context.Context, geoLocations []models.GeoLocation) error {
	err := r.db.WithContext(ctx).CreateInBatches(geoLocations, 100).Error
	if err != nil {
		return fmt.Errorf("failed to bulk create geo locations: %w", err)
	}
	return nil
}

// BulkUpdate updates multiple geo locations
func (r *geoLocationRepository) BulkUpdate(ctx context.Context, geoLocations []models.GeoLocation) error {
	for _, geoLocation := range geoLocations {
		err := r.Update(ctx, geoLocation.PostalCode, &geoLocation)
		if err != nil {
			return fmt.Errorf("failed to bulk update geo location %s: %w", geoLocation.PostalCode, err)
		}
	}
	return nil
}
