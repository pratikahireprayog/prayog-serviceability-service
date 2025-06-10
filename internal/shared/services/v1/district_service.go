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

// districtService implements the DistrictService interface
type districtService struct {
	repo repositories.DistrictRepository
}

// NewDistrictService creates a new district service instance
func NewDistrictService(repo repositories.DistrictRepository) DistrictService {
	return &districtService{
		repo: repo,
	}
}

// validateSnakeCase validates that a string is in snake_case format
func validateDistrictSnakeCase(s string) bool {
	pattern := `^[a-z0-9]+(_[a-z0-9]+)*$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// GetByID retrieves a district by its ID
func (s *districtService) GetByID(ctx context.Context, id string) (*dtos.DistrictResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("district ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid district ID format: %w", err)
	}

	district, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get district by ID: %w", err)
	}

	if district == nil {
		return nil, fmt.Errorf("district not found")
	}

	return DistrictToResponse(district), nil
}

// GetByCode retrieves a district by its code
func (s *districtService) GetByCode(ctx context.Context, code string) (*dtos.DistrictResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("district code cannot be empty")
	}

	// Validate and normalize code format (should be snake_case)
	code = strings.ToLower(strings.TrimSpace(code))
	if !validateDistrictSnakeCase(code) {
		return nil, fmt.Errorf("district code must be in snake_case format")
	}

	district, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get district by code: %w", err)
	}

	if district == nil {
		return nil, fmt.Errorf("district not found")
	}

	return DistrictToResponse(district), nil
}

// GetAll retrieves all districts with pagination
func (s *districtService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.DistrictListResponse, error) {
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

	districts, total, err := s.repo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get districts: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.DistrictResponse, len(districts))
	for i, district := range districts {
		if response := DistrictToResponse(&district); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.DistrictListResponse{
		Success: true,
		Message: "Districts retrieved successfully",
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

// GetAllWithDeleted retrieves all districts including soft-deleted ones (admin operation)
func (s *districtService) GetAllWithDeleted(ctx context.Context, req *dtos.PaginationRequest) (*dtos.DistrictListResponse, error) {
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

	districts, total, err := s.repo.GetAllWithDeleted(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get districts with deleted: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.DistrictResponse, len(districts))
	for i, district := range districts {
		if response := DistrictToResponse(&district); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.DistrictListResponse{
		Success: true,
		Message: "Districts retrieved successfully (including deleted)",
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

// GetByRegionID retrieves all districts for a specific region
func (s *districtService) GetByRegionID(ctx context.Context, regionID string) ([]dtos.DistrictResponse, error) {
	if strings.TrimSpace(regionID) == "" {
		return nil, fmt.Errorf("region ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(regionID); err != nil {
		return nil, fmt.Errorf("invalid region ID format: %w", err)
	}

	districts, err := s.repo.GetByRegionID(ctx, regionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get districts by region ID: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.DistrictResponse, len(districts))
	for i, district := range districts {
		if response := DistrictToResponse(&district); response != nil {
			responses[i] = *response
		}
	}

	return responses, nil
}

// Create creates a new district
func (s *districtService) Create(ctx context.Context, req *dtos.CreateDistrictRequest) (*dtos.DistrictResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate required fields
	if strings.TrimSpace(req.Code) == "" {
		return nil, fmt.Errorf("district code is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("district name is required")
	}

	// Normalize and validate code
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))
	if !validateDistrictSnakeCase(req.Code) {
		return nil, fmt.Errorf("district code must be in snake_case format")
	}

	// Check if district with code already exists
	existing, err := s.repo.GetByCode(ctx, req.Code)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check existing district: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("district with code '%s' already exists", req.Code)
	}

	// Create district model
	district := &models.District{
		ID:       uuid.New(),
		Code:     req.Code,
		Name:     strings.TrimSpace(req.Name),
		IsActive: true, // Default to active
	}

	// Set optional fields
	if req.RegionID != nil {
		district.RegionID = req.RegionID
	}

	if req.RegionCode != nil {
		regionCode := strings.ToLower(strings.TrimSpace(*req.RegionCode))
		if regionCode != "" {
			if !validateDistrictSnakeCase(regionCode) {
				return nil, fmt.Errorf("region code must be in snake_case format")
			}
			district.RegionCode = &regionCode
		}
	}

	if req.CountryID != nil {
		district.CountryID = req.CountryID
	}

	if req.CountryCode != nil {
		countryCode := strings.ToUpper(strings.TrimSpace(*req.CountryCode))
		if countryCode != "" {
			district.CountryCode = &countryCode
		}
	}

	if req.IsActive != nil {
		district.IsActive = *req.IsActive
	}

	err = s.repo.Create(ctx, district)
	if err != nil {
		return nil, fmt.Errorf("failed to create district: %w", err)
	}

	return DistrictToResponse(district), nil
}

// Update updates an existing district
func (s *districtService) Update(ctx context.Context, id string, req *dtos.UpdateDistrictRequest) (*dtos.DistrictResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("district ID cannot be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid district ID format: %w", err)
	}

	// Check if district exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get district: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("district not found")
	}

	// Create updated district model
	updated := *existing

	// Update only provided fields
	if req.RegionID != nil {
		updated.RegionID = req.RegionID
	}

	if req.RegionCode != nil {
		regionCode := strings.ToLower(strings.TrimSpace(*req.RegionCode))
		if regionCode != "" {
			if !validateDistrictSnakeCase(regionCode) {
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

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("district name cannot be empty")
		}
		updated.Name = name
	}

	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}

	err = s.repo.Update(ctx, id, &updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update district: %w", err)
	}

	return DistrictToResponse(&updated), nil
}

// Delete deletes a district by ID
func (s *districtService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("district ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid district ID format: %w", err)
	}

	// Check if district exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get district: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("district not found")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete district: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted district (admin operation)
func (s *districtService) Restore(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("district ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid district ID format: %w", err)
	}

	// Check if district exists (including soft-deleted)
	existing, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("district not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("district not found")
	}

	// Check for unique constraint conflicts before restoration
	if existing.Code != "" {
		activeDistrict, err := s.repo.GetByCode(ctx, existing.Code)
		if err == nil && activeDistrict != nil {
			return fmt.Errorf("cannot restore district: another active district with code '%s' already exists", existing.Code)
		}
	}

	err = s.repo.Restore(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore district: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes a district (admin operation)
func (s *districtService) ForceDelete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("district ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid district ID format: %w", err)
	}

	// Check if district exists (including soft-deleted)
	existing, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("district not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("district not found")
	}

	err = s.repo.ForceDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete district: %w", err)
	}

	return nil
}
