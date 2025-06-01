package gorm

import (
	"context"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"gorm.io/gorm"
)

type areaRepository struct {
	db *gorm.DB
}

func NewAreaRepository(db *gorm.DB) repositories.AreaRepository {
	return &areaRepository{db: db}
}

func (r *areaRepository) GetByID(ctx context.Context, id uint) (*models.Area, error) {
	var area models.Area
	err := r.db.WithContext(ctx).First(&area, id).Error
	return &area, err
}

func (r *areaRepository) GetByCityID(ctx context.Context, cityID uint) ([]models.Area, error) {
	var areas []models.Area
	err := r.db.WithContext(ctx).Where("city_id = ?", cityID).Find(&areas).Error
	return areas, err
}

func (r *areaRepository) GetByCode(ctx context.Context, code string, cityID uint) (*models.Area, error) {
	var area models.Area
	err := r.db.WithContext(ctx).Where("code = ? AND city_id = ?", code, cityID).First(&area).Error
	return &area, err
}

func (r *areaRepository) Create(ctx context.Context, area *models.Area) error {
	return r.db.WithContext(ctx).Create(area).Error
}

func (r *areaRepository) Update(ctx context.Context, area *models.Area) error {
	return r.db.WithContext(ctx).Save(area).Error
}

func (r *areaRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Area{}, id).Error
}
