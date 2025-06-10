package services

import (
	"context"
	"fmt"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// cityService implements the CityService interface
type cityService struct {
	repo repositories.CityRepository
}

// NewCityService creates a new city service instance
func NewCityService(repo repositories.CityRepository) CityService {
	return &cityService{
		repo: repo,
	}
}

// validateCitySnakeCase validates that a string is in snake_case format
func validateCitySnakeCase(s string) bool {
	pattern := `^[a-z0-9]+(_[a-z0-9]+)*$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// GetByID retrieves a city by its ID
func (s *cityService) GetByID(ctx context.Context, id string) (*dtos.CityResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("city ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid city ID format: %w", err)
	}

	city, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get city by ID: %w", err)
	}

	if city == nil {
		return nil, fmt.Errorf("city not found")
	}

	return CityToResponse(city), nil
}

// GetByCode retrieves a city by its code
func (s *cityService) GetByCode(ctx context.Context, code string) (*dtos.CityResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("city code cannot be empty")
	}

	// Normalize code (cities can use various formats, not strictly snake_case)
	code = strings.ToLower(strings.TrimSpace(code))

	city, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get city by code: %w", err)
	}

	if city == nil {
		return nil, fmt.Errorf("city not found")
	}

	return CityToResponse(city), nil
}

// GetAll retrieves all cities with pagination
func (s *cityService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.CityListResponse, error) {
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

	cities, total, err := s.repo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get cities: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.CityResponse, len(cities))
	for i, city := range cities {
		if response := CityToResponse(&city); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.CityListResponse{
		Success: true,
		Message: "Cities retrieved successfully",
		Data:    responses,
		Pagination: dtos.PaginationResponse{
			Offset:      req.Offset,
			Limit:       req.Limit,
			Total:       total,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	}, nil
}

// GetAllWithDeleted retrieves all cities including soft-deleted ones (admin operation)
func (s *cityService) GetAllWithDeleted(ctx context.Context, req *dtos.PaginationRequest) (*dtos.CityListResponse, error) {
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

	cities, total, err := s.repo.GetAllWithDeleted(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get cities with deleted: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.CityResponse, len(cities))
	for i, city := range cities {
		if response := CityToResponse(&city); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.CityListResponse{
		Success: true,
		Message: "Cities retrieved successfully (including deleted)",
		Data:    responses,
		Pagination: dtos.PaginationResponse{
			Offset:      req.Offset,
			Limit:       req.Limit,
			Total:       total,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	}, nil
}

// GetByRegionID retrieves all cities for a specific region
func (s *cityService) GetByRegionID(ctx context.Context, regionID string) ([]dtos.CityResponse, error) {
	if strings.TrimSpace(regionID) == "" {
		return nil, fmt.Errorf("region ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(regionID); err != nil {
		return nil, fmt.Errorf("invalid region ID format: %w", err)
	}

	cities, err := s.repo.GetByRegionID(ctx, regionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cities by region ID: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.CityResponse, len(cities))
	for i, city := range cities {
		if response := CityToResponse(&city); response != nil {
			responses[i] = *response
		}
	}

	return responses, nil
}

// Create creates a new city
func (s *cityService) Create(ctx context.Context, req *dtos.CreateCityRequest) (*dtos.CityResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate required fields
	if strings.TrimSpace(req.Code) == "" {
		return nil, fmt.Errorf("city code is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("city name is required")
	}

	// Normalize code
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))

	// Check if city with code already exists
	existing, err := s.repo.GetByCode(ctx, req.Code)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check existing city: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("city with code '%s' already exists", req.Code)
	}

	// Create city model
	city := &models.City{
		ID:       uuid.New(),
		Code:     req.Code,
		Name:     strings.TrimSpace(req.Name),
		IsActive: true, // Default to active
	}

	// Set optional fields
	if req.RegionID != nil {
		city.RegionID = req.RegionID
	}

	if req.RegionCode != nil {
		regionCode := strings.ToLower(strings.TrimSpace(*req.RegionCode))
		if regionCode != "" {
			if !validateCitySnakeCase(regionCode) {
				return nil, fmt.Errorf("region code must be in snake_case format")
			}
			city.RegionCode = &regionCode
		}
	}

	if req.CountryID != nil {
		city.CountryID = req.CountryID
	}

	if req.CountryCode != nil {
		countryCode := strings.ToUpper(strings.TrimSpace(*req.CountryCode))
		if countryCode != "" {
			city.CountryCode = &countryCode
		}
	}

	if req.DistrictID != nil {
		city.DistrictID = req.DistrictID
	}

	if req.DistrictCode != nil {
		districtCode := strings.ToLower(strings.TrimSpace(*req.DistrictCode))
		if districtCode != "" {
			if !validateCitySnakeCase(districtCode) {
				return nil, fmt.Errorf("district code must be in snake_case format")
			}
			city.DistrictCode = &districtCode
		}
	}

	if req.IsActive != nil {
		city.IsActive = *req.IsActive
	}

	err = s.repo.Create(ctx, city)
	if err != nil {
		return nil, fmt.Errorf("failed to create city: %w", err)
	}

	return CityToResponse(city), nil
}

// Update updates an existing city
func (s *cityService) Update(ctx context.Context, id string, req *dtos.UpdateCityRequest) (*dtos.CityResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("city ID cannot be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid city ID format: %w", err)
	}

	// Check if city exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get city: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("city not found")
	}

	// Create updated city model
	updated := *existing

	// Update only provided fields
	if req.RegionID != nil {
		updated.RegionID = req.RegionID
	}

	if req.RegionCode != nil {
		regionCode := strings.ToLower(strings.TrimSpace(*req.RegionCode))
		if regionCode != "" {
			if !validateCitySnakeCase(regionCode) {
				return nil, fmt.Errorf("region code must be in snake_case format")
			}
			updated.RegionCode = &regionCode
		} else {
			updated.RegionCode = nil
		}
	}

	if req.CountryID != nil {
		updated.CountryID = req.CountryID
	}

	if req.CountryCode != nil {
		countryCode := strings.ToUpper(strings.TrimSpace(*req.CountryCode))
		if countryCode != "" {
			updated.CountryCode = &countryCode
		} else {
			updated.CountryCode = nil
		}
	}

	if req.DistrictID != nil {
		updated.DistrictID = req.DistrictID
	}

	if req.DistrictCode != nil {
		districtCode := strings.ToLower(strings.TrimSpace(*req.DistrictCode))
		if districtCode != "" {
			if !validateCitySnakeCase(districtCode) {
				return nil, fmt.Errorf("district code must be in snake_case format")
			}
			updated.DistrictCode = &districtCode
		} else {
			updated.DistrictCode = nil
		}
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("city name cannot be empty")
		}
		updated.Name = name
	}

	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}

	err = s.repo.Update(ctx, id, &updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update city: %w", err)
	}

	return CityToResponse(&updated), nil
}

// Delete deletes a city by ID
func (s *cityService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("city ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid city ID format: %w", err)
	}

	// Check if city exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get city: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("city not found")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete city: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted city (admin operation)
func (s *cityService) Restore(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("city ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid city ID format: %w", err)
	}

	// Check if city exists (including soft-deleted)
	existing, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("city not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("city not found")
	}

	// Check for unique constraint conflicts before restoration
	if existing.Code != "" {
		activeCity, err := s.repo.GetByCode(ctx, existing.Code)
		if err == nil && activeCity != nil {
			return fmt.Errorf("cannot restore city: another active city with code '%s' already exists", existing.Code)
		}
	}

	err = s.repo.Restore(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore city: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes a city (admin operation)
func (s *cityService) ForceDelete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("city ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid city ID format: %w", err)
	}

	// Check if city exists (including soft-deleted)
	existing, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("city not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("city not found")
	}

	err = s.repo.ForceDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete city: %w", err)
	}

	return nil
}
