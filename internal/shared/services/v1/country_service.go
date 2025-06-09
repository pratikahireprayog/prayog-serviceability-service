package services

import (
	"context"
	"fmt"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"
	"strings"

	"github.com/google/uuid"
)

// countryService implements the CountryService interface
type countryService struct {
	repo repositories.CountryRepository
}

// NewCountryService creates a new country service instance
func NewCountryService(repo repositories.CountryRepository) CountryService {
	return &countryService{
		repo: repo,
	}
}

// GetByID retrieves a country by its ID
func (s *countryService) GetByID(ctx context.Context, id string) (*dtos.CountryResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("country ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid country ID format: %w", err)
	}

	country, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get country by ID: %w", err)
	}

	if country == nil {
		return nil, fmt.Errorf("country not found")
	}

	return CountryToResponse(country), nil
}

// GetByCode retrieves a country by its code
func (s *countryService) GetByCode(ctx context.Context, code string) (*dtos.CountryResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("country code cannot be empty")
	}

	// Validate code format (should be 2-3 characters, uppercase)
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) < 2 || len(code) > 3 {
		return nil, fmt.Errorf("country code must be 2-3 characters long")
	}

	country, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get country by code: %w", err)
	}

	if country == nil {
		return nil, fmt.Errorf("country not found")
	}

	return CountryToResponse(country), nil
}

// GetAll retrieves all countries with pagination
func (s *countryService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.CountryListResponse, error) {
	if req == nil {
		req = &dtos.PaginationRequest{
			Offset: 0,
			Limit:  10,
		}
	}

	// Validate pagination parameters
	if req.Offset < 0 {
		req.Offset = 0
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 10
	}

	countries, total, err := s.repo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get countries: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.CountryResponse, len(countries))
	for i, country := range countries {
		if response := CountryToResponse(&country); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.CountryListResponse{
		Countries: responses,
		Pagination: dtos.PaginationResponse{
			Offset:      req.Offset,
			Limit:       req.Limit,
			Total:       total,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	}, nil
}

// Create creates a new country
func (s *countryService) Create(ctx context.Context, req *dtos.CreateCountryRequest) (*dtos.CountryResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate required fields
	if strings.TrimSpace(req.Code) == "" {
		return nil, fmt.Errorf("country code is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("country name is required")
	}

	// Normalize and validate code
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	if len(req.Code) < 2 || len(req.Code) > 3 {
		return nil, fmt.Errorf("country code must be 2-3 characters long")
	}

	// Check if country with code already exists
	existing, err := s.repo.GetByCode(ctx, req.Code)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check existing country: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("country with code '%s' already exists", req.Code)
	}

	// Create country model
	country := &models.Country{
		ID:       uuid.New(),
		Code:     req.Code,
		Name:     strings.TrimSpace(req.Name),
		IsActive: true, // Default to active
	}

	// Set optional fields
	if req.CurrencyCode != nil {
		currencyCode := strings.ToUpper(strings.TrimSpace(*req.CurrencyCode))
		if currencyCode != "" {
			if len(currencyCode) != 3 {
				return nil, fmt.Errorf("currency code must be 3 characters long")
			}
			country.CurrencyCode = &currencyCode
		}
	}

	if req.PhoneCode != nil {
		phoneCode := strings.TrimSpace(*req.PhoneCode)
		if phoneCode != "" {
			country.PhoneCode = &phoneCode
		}
	}

	if req.IsActive != nil {
		country.IsActive = *req.IsActive
	}

	err = s.repo.Create(ctx, country)
	if err != nil {
		return nil, fmt.Errorf("failed to create country: %w", err)
	}

	return CountryToResponse(country), nil
}

// Update updates an existing country
func (s *countryService) Update(ctx context.Context, id string, req *dtos.UpdateCountryRequest) (*dtos.CountryResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("country ID cannot be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid country ID format: %w", err)
	}

	// Check if country exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing country: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("country not found")
	}

	// Update fields if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("country name cannot be empty")
		}
		existing.Name = name
	}

	if req.CurrencyCode != nil {
		currencyCode := strings.ToUpper(strings.TrimSpace(*req.CurrencyCode))
		if currencyCode != "" && len(currencyCode) != 3 {
			return nil, fmt.Errorf("currency code must be 3 characters long")
		}
		if currencyCode == "" {
			existing.CurrencyCode = nil
		} else {
			existing.CurrencyCode = &currencyCode
		}
	}

	if req.PhoneCode != nil {
		phoneCode := strings.TrimSpace(*req.PhoneCode)
		if phoneCode == "" {
			existing.PhoneCode = nil
		} else {
			existing.PhoneCode = &phoneCode
		}
	}

	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	err = s.repo.Update(ctx, id, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update country: %w", err)
	}

	return CountryToResponse(existing), nil
}

// Delete deletes a country by ID
func (s *countryService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("country ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid country ID format: %w", err)
	}

	// Check if country exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get existing country: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("country not found")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete country: %w", err)
	}

	return nil
}
