package gorm

import (
	"context"

	"prayog-serviceability-service/internal/domain"
	"prayog-serviceability-service/pkg/database"
)

// CityRepository implements domain.CityRepository using GORM
type CityRepository struct {
	*Repository
}

// NewCityRepository creates a new city repository
func NewCityRepository(db *database.DB) domain.CityRepository {
	return &CityRepository{
		Repository: NewRepository(db),
	}
}

// GetByID retrieves a city by its ID
func (r *CityRepository) GetByID(ctx context.Context, id uint) (*domain.City, error) {
	var city database.City
	if err := r.WithContext(ctx).First(&city, id).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabaseCityToDomain(&city), nil
}

// GetByRegionID retrieves cities by region ID
func (r *CityRepository) GetByRegionID(ctx context.Context, regionID uint) ([]*domain.City, error) {
	var cities []database.City
	if err := r.WithContext(ctx).Where("administrative_region_id = ?", regionID).Find(&cities).Error; err != nil {
		return nil, HandleError(err)
	}

	domainCities := make([]*domain.City, len(cities))
	for i, city := range cities {
		domainCities[i] = mapDatabaseCityToDomain(&city)
	}
	return domainCities, nil
}

// List retrieves all cities
func (r *CityRepository) List(ctx context.Context) ([]*domain.City, error) {
	var cities []database.City
	if err := r.WithContext(ctx).Find(&cities).Error; err != nil {
		return nil, HandleError(err)
	}

	domainCities := make([]*domain.City, len(cities))
	for i, city := range cities {
		domainCities[i] = mapDatabaseCityToDomain(&city)
	}
	return domainCities, nil
}

// Create creates a new city
func (r *CityRepository) Create(ctx context.Context, city *domain.City) error {
	dbCity := mapDomainCityToDatabase(city)
	if err := r.WithContext(ctx).Create(&dbCity).Error; err != nil {
		return HandleError(err)
	}
	// Update the ID after creation
	city.ID = dbCity.ID
	return nil
}

// Update updates an existing city
func (r *CityRepository) Update(ctx context.Context, city *domain.City) error {
	dbCity := mapDomainCityToDatabase(city)
	result := r.WithContext(ctx).Save(&dbCity)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a city by its ID
func (r *CityRepository) Delete(ctx context.Context, id uint) error {
	result := r.WithContext(ctx).Delete(&database.City{}, id)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// mapDatabaseCityToDomain maps a database city to a domain city
func mapDatabaseCityToDomain(dbCity *database.City) *domain.City {
	return &domain.City{
		ID:                     dbCity.ID,
		AdministrativeRegionID: dbCity.AdministrativeRegionID,
		Name:                   dbCity.Name,
		CreatedAt:              dbCity.CreatedAt,
		UpdatedAt:              dbCity.UpdatedAt,
	}
}

// mapDomainCityToDatabase maps a domain city to a database city
func mapDomainCityToDatabase(domainCity *domain.City) database.City {
	return database.City{
		ID:                     domainCity.ID,
		AdministrativeRegionID: domainCity.AdministrativeRegionID,
		Name:                   domainCity.Name,
		CreatedAt:              domainCity.CreatedAt,
		UpdatedAt:              domainCity.UpdatedAt,
	}
}
