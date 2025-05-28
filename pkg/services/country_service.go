package services

import (
	"context"

	"prayog-serviceability-service/pkg/domain"
	"prayog-serviceability-service/pkg/repository"

	"github.com/google/uuid"
)

// CountryService handles country-related business logic
type CountryService struct {
	repo repository.CountryRepository
}

// NewCountryService creates a new country service
func NewCountryService(repo repository.CountryRepository) *CountryService {
	return &CountryService{
		repo: repo,
	}
}

// GetByID retrieves a country by its ID
func (s *CountryService) GetByID(ctx context.Context, id uuid.UUID) (domain.Country, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByCode retrieves a country by its code
func (s *CountryService) GetByCode(ctx context.Context, code string) (domain.Country, error) {
	return s.repo.GetByCode(ctx, code)
}

// ListCountries retrieves all countries
func (s *CountryService) ListCountries(ctx context.Context) ([]domain.Country, error) {
	return s.repo.List(ctx)
}

// CreateCountry creates a new country
func (s *CountryService) CreateCountry(ctx context.Context, country domain.Country) error {
	// Perform business logic validation
	if country.Code == "" {
		return domain.NewValidationError("country code cannot be empty")
	}
	if country.Name == "" {
		return domain.NewValidationError("country name cannot be empty")
	}

	// Generate a new ID if not provided
	if country.ID == uuid.Nil {
		country.ID = uuid.New()
	}

	return s.repo.Create(ctx, country)
}

// UpdateCountry updates an existing country
func (s *CountryService) UpdateCountry(ctx context.Context, country domain.Country) error {
	// Perform business logic validation
	if country.ID == uuid.Nil {
		return domain.NewValidationError("country ID cannot be empty")
	}
	if country.Code == "" {
		return domain.NewValidationError("country code cannot be empty")
	}
	if country.Name == "" {
		return domain.NewValidationError("country name cannot be empty")
	}

	return s.repo.Update(ctx, country)
}

// DeleteCountry deletes a country by its ID
func (s *CountryService) DeleteCountry(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
