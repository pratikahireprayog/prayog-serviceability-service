package gorm

import (
	"context"

	"prayog-serviceability-service/internal/domain"
	"prayog-serviceability-service/pkg/database"
)

// CountryRepository implements domain.CountryRepository using GORM
type CountryRepository struct {
	*Repository
}

// NewCountryRepository creates a new country repository
func NewCountryRepository(db *database.DB) domain.CountryRepository {
	return &CountryRepository{
		Repository: NewRepository(db),
	}
}

// GetByID retrieves a country by its ID
func (r *CountryRepository) GetByID(ctx context.Context, id uint) (*domain.Country, error) {
	var country database.Country
	if err := r.WithContext(ctx).First(&country, id).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabaseCountryToDomain(&country), nil
}

// GetByCode retrieves a country by its code
func (r *CountryRepository) GetByCode(ctx context.Context, code string) (*domain.Country, error) {
	var country database.Country
	if err := r.WithContext(ctx).Where("code = ?", code).First(&country).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabaseCountryToDomain(&country), nil
}

// List retrieves all countries
func (r *CountryRepository) List(ctx context.Context) ([]*domain.Country, error) {
	var countries []database.Country
	if err := r.WithContext(ctx).Find(&countries).Error; err != nil {
		return nil, HandleError(err)
	}

	domainCountries := make([]*domain.Country, len(countries))
	for i, country := range countries {
		domainCountries[i] = mapDatabaseCountryToDomain(&country)
	}
	return domainCountries, nil
}

// Create creates a new country
func (r *CountryRepository) Create(ctx context.Context, country *domain.Country) error {
	dbCountry := mapDomainCountryToDatabase(country)
	if err := r.WithContext(ctx).Create(&dbCountry).Error; err != nil {
		return HandleError(err)
	}
	// Update the ID after creation
	country.ID = dbCountry.ID
	return nil
}

// Update updates an existing country
func (r *CountryRepository) Update(ctx context.Context, country *domain.Country) error {
	dbCountry := mapDomainCountryToDatabase(country)
	result := r.WithContext(ctx).Save(&dbCountry)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a country by its ID
func (r *CountryRepository) Delete(ctx context.Context, id uint) error {
	result := r.WithContext(ctx).Delete(&database.Country{}, id)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// mapDatabaseCountryToDomain maps a database country to a domain country
func mapDatabaseCountryToDomain(dbCountry *database.Country) *domain.Country {
	return &domain.Country{
		ID:        dbCountry.ID,
		Code:      dbCountry.Code,
		Name:      dbCountry.Name,
		CreatedAt: dbCountry.CreatedAt,
		UpdatedAt: dbCountry.UpdatedAt,
	}
}

// mapDomainCountryToDatabase maps a domain country to a database country
func mapDomainCountryToDatabase(domainCountry *domain.Country) database.Country {
	return database.Country{
		ID:        domainCountry.ID,
		Code:      domainCountry.Code,
		Name:      domainCountry.Name,
		CreatedAt: domainCountry.CreatedAt,
		UpdatedAt: domainCountry.UpdatedAt,
	}
}
