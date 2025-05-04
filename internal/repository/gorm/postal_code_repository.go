package gorm

import (
	"context"

	"github.com/prayog/serviceability/internal/domain"
	"github.com/prayog/serviceability/pkg/database"
)

// PostalCodeRepository implements domain.PostalCodeRepository using GORM
type PostalCodeRepository struct {
	*Repository
}

// NewPostalCodeRepository creates a new postal code repository
func NewPostalCodeRepository(db *database.DB) domain.PostalCodeRepository {
	return &PostalCodeRepository{
		Repository: NewRepository(db),
	}
}

// GetByID retrieves a postal code by its ID
func (r *PostalCodeRepository) GetByID(ctx context.Context, id uint) (*domain.PostalCode, error) {
	var postalCode database.PostalCode
	if err := r.WithContext(ctx).First(&postalCode, id).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabasePostalCodeToDomain(&postalCode), nil
}

// GetByCode retrieves a postal code by its code
func (r *PostalCodeRepository) GetByCode(ctx context.Context, code string) (*domain.PostalCode, error) {
	var postalCode database.PostalCode
	if err := r.WithContext(ctx).Where("code = ?", code).First(&postalCode).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabasePostalCodeToDomain(&postalCode), nil
}

// GetByAreaID retrieves postal codes by area ID
func (r *PostalCodeRepository) GetByAreaID(ctx context.Context, areaID uint) ([]*domain.PostalCode, error) {
	var postalCodes []database.PostalCode
	if err := r.WithContext(ctx).Where("area_id = ?", areaID).Find(&postalCodes).Error; err != nil {
		return nil, HandleError(err)
	}

	domainPostalCodes := make([]*domain.PostalCode, len(postalCodes))
	for i, postalCode := range postalCodes {
		domainPostalCodes[i] = mapDatabasePostalCodeToDomain(&postalCode)
	}
	return domainPostalCodes, nil
}

// List retrieves all postal codes
func (r *PostalCodeRepository) List(ctx context.Context) ([]*domain.PostalCode, error) {
	var postalCodes []database.PostalCode
	if err := r.WithContext(ctx).Find(&postalCodes).Error; err != nil {
		return nil, HandleError(err)
	}

	domainPostalCodes := make([]*domain.PostalCode, len(postalCodes))
	for i, postalCode := range postalCodes {
		domainPostalCodes[i] = mapDatabasePostalCodeToDomain(&postalCode)
	}
	return domainPostalCodes, nil
}

// Create creates a new postal code
func (r *PostalCodeRepository) Create(ctx context.Context, postalCode *domain.PostalCode) error {
	dbPostalCode := mapDomainPostalCodeToDatabase(postalCode)
	if err := r.WithContext(ctx).Create(&dbPostalCode).Error; err != nil {
		return HandleError(err)
	}
	// Update the ID after creation
	postalCode.ID = dbPostalCode.ID
	return nil
}

// Update updates an existing postal code
func (r *PostalCodeRepository) Update(ctx context.Context, postalCode *domain.PostalCode) error {
	dbPostalCode := mapDomainPostalCodeToDatabase(postalCode)
	result := r.WithContext(ctx).Save(&dbPostalCode)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a postal code by its ID
func (r *PostalCodeRepository) Delete(ctx context.Context, id uint) error {
	result := r.WithContext(ctx).Delete(&database.PostalCode{}, id)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// mapDatabasePostalCodeToDomain maps a database postal code to a domain postal code
func mapDatabasePostalCodeToDomain(dbPostalCode *database.PostalCode) *domain.PostalCode {
	return &domain.PostalCode{
		ID:        dbPostalCode.ID,
		Code:      dbPostalCode.Code,
		AreaID:    dbPostalCode.AreaID,
		CreatedAt: dbPostalCode.CreatedAt,
		UpdatedAt: dbPostalCode.UpdatedAt,
	}
}

// mapDomainPostalCodeToDatabase maps a domain postal code to a database postal code
func mapDomainPostalCodeToDatabase(domainPostalCode *domain.PostalCode) database.PostalCode {
	return database.PostalCode{
		ID:        domainPostalCode.ID,
		Code:      domainPostalCode.Code,
		AreaID:    domainPostalCode.AreaID,
		CreatedAt: domainPostalCode.CreatedAt,
		UpdatedAt: domainPostalCode.UpdatedAt,
	}
}
