package repositories

import (
	"context"
	"errors"
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

// GetByAliasCodeWithDeleted - PostalCodeAlias doesn't support soft deletion, so this is the same as GetByAliasCode
func (r *postalCodeAliasRepository) GetByAliasCodeWithDeleted(ctx context.Context, aliasCode string) (*models.PostalCodeAlias, error) {
	return r.GetByAliasCode(ctx, aliasCode)
}

// GetByPostalCodeIDWithDeleted - PostalCodeAlias doesn't support soft deletion, so this is the same as GetByPostalCodeID
func (r *postalCodeAliasRepository) GetByPostalCodeIDWithDeleted(ctx context.Context, postalCodeID uint) ([]models.PostalCodeAlias, error) {
	return r.GetByPostalCodeID(ctx, postalCodeID)
}

// GetOnlyDeleted - PostalCodeAlias doesn't support soft deletion, so this returns empty
func (r *postalCodeAliasRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.PostalCodeAlias, int64, error) {
	// PostalCodeAlias doesn't support soft deletion, return empty result
	return []models.PostalCodeAlias{}, 0, nil
}

// Restore - PostalCodeAlias doesn't support soft deletion, so this returns an error
func (r *postalCodeAliasRepository) Restore(ctx context.Context, id uint) error {
	return errors.New("PostalCodeAlias does not support soft deletion or restoration")
}

// ForceDelete permanently deletes a postal code alias (same as Delete since there's no soft delete)
func (r *postalCodeAliasRepository) ForceDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.PostalCodeAlias{}, id).Error
}
