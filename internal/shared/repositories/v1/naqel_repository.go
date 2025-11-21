package repositories

import (
	"context"
	"fmt"
	"prayog-serviceability-service/internal/shared/models/v1"
	"time"

	"gorm.io/gorm"
)

type NaqelCity struct {
	CountryName   string    `gorm:"column:country_name"`
	CountryCode   string    `gorm:"column:country_code"`
	AdminArea     string    `gorm:"column:admin_area"`
	LocationEn    string    `gorm:"column:location_en"`
	LocationAr    string    `gorm:"column:location_ar"`
	CityCode      string    `gorm:"column:city_code"`
	StationCode   string    `gorm:"column:station_code"`
	FacilityCode  string    `gorm:"column:facility_code"`
	PoE           string    `gorm:"column:poe"`
	ODA           string    `gorm:"column:oda"`
	IsServiceable bool      `gorm:"column:is_serviceable"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

// NaqelRepository defines methods for querying Naqel serviceability data
type NaqelRepository interface {
	// CheckServiceabilityByCityCodes checks if both source and destination city codes are serviceable
	// Returns locations with country codes from the database
	CheckServiceabilityByCityCodes(ctx context.Context, sourceCityCode, destCityCode string) (*models.NaqelLocation, *models.NaqelLocation, error)
	// GetCityCodeByPostalCode retrieves city code for a given postal code and country code
	GetCityCodeByPostalCode(ctx context.Context, postalCode, countryCode string) (string, error)

}

// naqelRepository implements NaqelRepository interface
type naqelRepository struct {
	db     *gorm.DB
	tableName string
}

// NewNaqelRepository creates a new Naqel repository
func NewNaqelRepository(db *gorm.DB, tableName string) NaqelRepository {
	if tableName == "" {
		tableName = "naqel_cities"
	}
	return &naqelRepository{
		db:        db,
		tableName: tableName,
	}
}

// CheckServiceabilityByCityCodes checks if both source and destination city codes exist and are serviceable
// Uses country codes directly from the request parameters (no geolocation lookup needed)
func (r *naqelRepository) CheckServiceabilityByCityCodes(ctx context.Context, sourceCityCode, destCityCode string) (*models.NaqelLocation, *models.NaqelLocation, error) {
	
	var sourceLocation models.NaqelLocation
	sourceQuery := r.db.WithContext(ctx).Table(r.tableName).
		Where("city_code = ? AND is_serviceable = ?", sourceCityCode, true).
		First(&sourceLocation)

	if sourceQuery.Error != nil {
		if sourceQuery.Error == gorm.ErrRecordNotFound {
			return nil, nil, fmt.Errorf("source city code %s in country %s not found or not serviceable", sourceCityCode, "sourceCountryCode")
		}
		return nil, nil, fmt.Errorf("failed to query source location: %w", sourceQuery.Error)
	}

	// Query destination location
	var destLocation models.NaqelLocation
	destQuery := r.db.WithContext(ctx).Table(r.tableName).
		Where("city_code = ? AND is_serviceable = ?", destCityCode, true).
		First(&destLocation)

	if destQuery.Error != nil {
		if destQuery.Error == gorm.ErrRecordNotFound {
			return nil, nil, fmt.Errorf("destination city code %s in country %s not found or not serviceable", destCityCode, "destCountryCode")
		}
		return nil, nil, fmt.Errorf("failed to query destination location: %w", destQuery.Error)
	}

	return &sourceLocation, &destLocation, nil
}

// GetCityCodeByPostalCode retrieves city code from postal code
// Since naqel_cities table doesn't have postal_code column, we lookup via postal_codes table
func (r *naqelRepository) GetCityCodeByPostalCode(ctx context.Context, postalCode, countryCode string) (string, error) {
	// Query postal_codes table to get city code, then verify it exists in naqel_cities
	var postalCodeRow struct {
		CityCode string `gorm:"column:city_code"`
	}
	
	postalQuery := r.db.WithContext(ctx).
		Table("postal_codes").
		Joins("JOIN cities ON postal_codes.city_id = cities.id").
		Select("cities.code as city_code").
		Where("postal_codes.code = ? AND postal_codes.country_code = ?", postalCode, countryCode).
		First(&postalCodeRow)

	if postalQuery.Error != nil {
		if postalQuery.Error == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("postal code %s in country %s not found", postalCode, countryCode)
		}
		return "", fmt.Errorf("failed to get city code for postal code: %w", postalQuery.Error)
	}

	if postalCodeRow.CityCode == "" {
		return "", fmt.Errorf("city code not found for postal code %s in country %s", postalCode, countryCode)
	}

	// Verify that the city code exists in naqel_cities table
	var count int64
	verifyQuery := r.db.WithContext(ctx).
		Table(r.tableName).
		Where("city_code = ? AND country_code = ?", postalCodeRow.CityCode, countryCode).
		Count(&count)

	if verifyQuery.Error != nil {
		return "", fmt.Errorf("failed to verify city code in naqel_cities: %w", verifyQuery.Error)
	}

	if count == 0 {
		return "", fmt.Errorf("city code %s found for postal code %s but not present in naqel_cities table", postalCodeRow.CityCode, postalCode)
	}

	return postalCodeRow.CityCode, nil
}



// Utility: safe map read
func safeGet(m map[string]string, key string) string {
	if val, ok := m[key]; ok {
		return val
	}
	return ""
}

// Utility: pick first non-empty string
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}