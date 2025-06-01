package gorm

import (
	"context"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"gorm.io/gorm"
)

type cityRepository struct {
	db *gorm.DB
}

func NewCityRepository(db *gorm.DB) repositories.CityRepository {
	return &cityRepository{db: db}
}

func (r *cityRepository) GetByID(ctx context.Context, id uint) (*models.City, error) {
	var city models.City
	err := r.db.WithContext(ctx).First(&city, id).Error
	return &city, err
}

func (r *cityRepository) GetByRegionID(ctx context.Context, regionID uint) ([]models.City, error) {
	var cities []models.City
	err := r.db.WithContext(ctx).Where("region_id = ?", regionID).Find(&cities).Error
	return cities, err
}

func (r *cityRepository) GetByCode(ctx context.Context, code string, regionID uint) (*models.City, error) {
	var city models.City
	err := r.db.WithContext(ctx).Where("code = ? AND region_id = ?", code, regionID).First(&city).Error
	return &city, err
}

func (r *cityRepository) Create(ctx context.Context, city *models.City) error {
	return r.db.WithContext(ctx).Create(city).Error
}

func (r *cityRepository) Update(ctx context.Context, city *models.City) error {
	return r.db.WithContext(ctx).Save(city).Error
}

func (r *cityRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.City{}, id).Error
}
