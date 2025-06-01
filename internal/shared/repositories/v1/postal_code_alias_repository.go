package repositories

import (
	"context"
	"prayog-serviceability-service/internal/shared/models/v1"

	"gorm.io/gorm"
)

type postalCodeAliasRepository struct {
	db *gorm.DB
}

func NewPostalCodeAliasRepository(db *gorm.DB) PostalCodeAliasRepository {
	return &postalCodeAliasRepository{db: db}
}

func (r *postalCodeAliasRepository) GetByAliasCode(ctx context.Context, aliasCode string) (*models.PostalCodeAlias, error) {
	var alias models.PostalCodeAlias
	err := r.db.WithContext(ctx).Where("alias_code = ?", aliasCode).First(&alias).Error
	return &alias, err
}

func (r *postalCodeAliasRepository) GetByPostalCodeID(ctx context.Context, postalCodeID uint) ([]models.PostalCodeAlias, error) {
	var aliases []models.PostalCodeAlias
	err := r.db.WithContext(ctx).Where("postal_code_id = ?", postalCodeID).Find(&aliases).Error
	return aliases, err
}

func (r *postalCodeAliasRepository) Create(ctx context.Context, alias *models.PostalCodeAlias) error {
	return r.db.WithContext(ctx).Create(alias).Error
}

func (r *postalCodeAliasRepository) Update(ctx context.Context, alias *models.PostalCodeAlias) error {
	return r.db.WithContext(ctx).Save(alias).Error
}

func (r *postalCodeAliasRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.PostalCodeAlias{}, id).Error
}
