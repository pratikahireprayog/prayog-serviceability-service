package gorm

import (
	"context"

	"prayog-serviceability-service/pkg/domain"
	database "prayog-serviceability-service/pkg/infrastructure/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CountryRepository implements CountryRepository interface
type CountryRepository struct {
	*BaseRepository[domain.Country, database.Country]
}

// NewCountryRepository creates a new country repository
func NewCountryRepository(db *database.DB) *CountryRepository {
	return &CountryRepository{
		BaseRepository: NewBaseRepository[domain.Country, database.Country](db),
	}
}

// GetByID retrieves a country by its ID
func (r *CountryRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Country, error) {
	return r.BaseRepository.GetByID(ctx, id, mapDatabaseCountryToDomain)
}

// GetByCode retrieves a country by its code
func (r *CountryRepository) GetByCode(ctx context.Context, code string) (domain.Country, error) {
	return r.BaseRepository.FindOne(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", code)
	}, mapDatabaseCountryToDomain)
}

// List retrieves all countries
func (r *CountryRepository) List(ctx context.Context) ([]domain.Country, error) {
	return r.BaseRepository.List(ctx, mapDatabaseCountryToDomain)
}

// Create creates a new country
func (r *CountryRepository) Create(ctx context.Context, country domain.Country) error {
	return r.BaseRepository.Create(ctx, country, mapDomainCountryToDatabase)
}

// Update updates an existing country
func (r *CountryRepository) Update(ctx context.Context, country domain.Country) error {
	return r.BaseRepository.Update(ctx, country, mapDomainCountryToDatabase)
}

// Delete deletes a country by its ID
func (r *CountryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id, database.Country{})
}

// Helper functions to map between domain and database models

func mapDatabaseCountryToDomain(dbCountry *database.Country) domain.Country {
	return domain.Country{
		ID:        dbCountry.ID,
		Name:      dbCountry.Name,
		Code:      dbCountry.Code,
		IsActive:  true, // Default value, could be stored in DB in the future
		CreatedAt: dbCountry.CreatedAt,
		UpdatedAt: dbCountry.UpdatedAt,
	}
}

func mapDomainCountryToDatabase(domainCountry domain.Country) database.Country {
	return database.Country{
		ID:        domainCountry.ID,
		Name:      domainCountry.Name,
		Code:      domainCountry.Code,
		CreatedAt: domainCountry.CreatedAt,
		UpdatedAt: domainCountry.UpdatedAt,
	}
}
