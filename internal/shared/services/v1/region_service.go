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

// regionService implements the RegionService interface
type regionService struct {
	repo repositories.RegionRepository
}

// NewRegionService creates a new region service instance
func NewRegionService(repo repositories.RegionRepository) RegionService {
	return &regionService{
		repo: repo,
	}
}

// validateSnakeCase validates that a string is in snake_case format
func validateSnakeCase(s string) bool {
	pattern := `^[a-z0-9]+(_[a-z0-9]+)*$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// GetByID retrieves a region by its ID
func (s *regionService) GetByID(ctx context.Context, id string) (*dtos.RegionResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("region ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid region ID format: %w", err)
	}

	region, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get region by ID: %w", err)
	}

	if region == nil {
		return nil, fmt.Errorf("region not found")
	}

	return RegionToResponse(region), nil
}

// GetByCode retrieves a region by its code
func (s *regionService) GetByCode(ctx context.Context, code string) (*dtos.RegionResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("region code cannot be empty")
	}

	// Validate and normalize code format (should be snake_case)
	code = strings.ToLower(strings.TrimSpace(code))
	if !validateSnakeCase(code) {
		return nil, fmt.Errorf("region code must be in snake_case format")
	}

	region, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get region by code: %w", err)
	}

	if region == nil {
		return nil, fmt.Errorf("region not found")
	}

	return RegionToResponse(region), nil
}

// GetAll retrieves all regions with pagination
func (s *regionService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.RegionListResponse, error) {
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

	regions, total, err := s.repo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.RegionResponse, len(regions))
	for i, region := range regions {
		if response := RegionToResponse(&region); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.RegionListResponse{
		Success: true,
		Message: "Regions retrieved successfully",
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

// GetAllWithDeleted retrieves all regions including soft-deleted ones (admin operation)
func (s *regionService) GetAllWithDeleted(ctx context.Context, req *dtos.PaginationRequest) (*dtos.RegionListResponse, error) {
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

	regions, total, err := s.repo.GetAllWithDeleted(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions with deleted: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.RegionResponse, len(regions))
	for i, region := range regions {
		if response := RegionToResponse(&region); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.RegionListResponse{
		Success: true,
		Message: "Regions retrieved successfully (including deleted)",
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

// GetByCountryID retrieves all regions for a specific country
func (s *regionService) GetByCountryID(ctx context.Context, countryID string) ([]dtos.RegionResponse, error) {
	if strings.TrimSpace(countryID) == "" {
		return nil, fmt.Errorf("country ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(countryID); err != nil {
		return nil, fmt.Errorf("invalid country ID format: %w", err)
	}

	regions, err := s.repo.GetByCountryID(ctx, countryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions by country ID: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.RegionResponse, len(regions))
	for i, region := range regions {
		if response := RegionToResponse(&region); response != nil {
			responses[i] = *response
		}
	}

	return responses, nil
}

// Create creates a new region
func (s *regionService) Create(ctx context.Context, req *dtos.CreateRegionRequest) (*dtos.RegionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate required fields
	if strings.TrimSpace(req.Code) == "" {
		return nil, fmt.Errorf("region code is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("region name is required")
	}

	// Normalize and validate code
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))
	if !validateSnakeCase(req.Code) {
		return nil, fmt.Errorf("region code must be in snake_case format")
	}

	// Check if region with code already exists
	existing, err := s.repo.GetByCode(ctx, req.Code)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check existing region: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("region with code '%s' already exists", req.Code)
	}

	// Create region model
	region := &models.Region{
		ID:       uuid.New(),
		Code:     req.Code,
		Name:     strings.TrimSpace(req.Name),
		IsActive: true, // Default to active
	}

	// Set optional fields
	if req.CountryID != nil {
		region.CountryID = req.CountryID
	}

	if req.CountryCode != nil {
		countryCode := strings.ToUpper(strings.TrimSpace(*req.CountryCode))
		if countryCode != "" {
			region.CountryCode = &countryCode
		}
	}

	if req.RegionTypeCode != nil {
		regionTypeCode := strings.ToLower(strings.TrimSpace(*req.RegionTypeCode))
		if regionTypeCode != "" {
			if !validateSnakeCase(regionTypeCode) {
				return nil, fmt.Errorf("region type code must be in snake_case format")
			}
			region.RegionTypeCode = &regionTypeCode
		}
	}

	if req.IsActive != nil {
		region.IsActive = *req.IsActive
	}

	err = s.repo.Create(ctx, region)
	if err != nil {
		return nil, fmt.Errorf("failed to create region: %w", err)
	}

	return RegionToResponse(region), nil
}

// Update updates an existing region
func (s *regionService) Update(ctx context.Context, id string, req *dtos.UpdateRegionRequest) (*dtos.RegionResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("region ID cannot be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid region ID format: %w", err)
	}

	// Check if region exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("region not found")
	}

	// Create updated region model
	updated := *existing

	// Update only provided fields
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

	if req.RegionTypeCode != nil {
		regionTypeCode := strings.ToLower(strings.TrimSpace(*req.RegionTypeCode))
		if regionTypeCode != "" {
			if !validateSnakeCase(regionTypeCode) {
				return nil, fmt.Errorf("region type code must be in snake_case format")
			}
			updated.RegionTypeCode = &regionTypeCode
		} else {
			updated.RegionTypeCode = nil
		}
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("region name cannot be empty")
		}
		updated.Name = name
	}

	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}

	err = s.repo.Update(ctx, id, &updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update region: %w", err)
	}

	return RegionToResponse(&updated), nil
}

// Delete deletes a region by ID
func (s *regionService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("region ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid region ID format: %w", err)
	}

	// Check if region exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get region: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("region not found")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete region: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted region (admin operation)
func (s *regionService) Restore(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("region ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid region ID format: %w", err)
	}

	// Check if region exists (including soft-deleted)
	existing, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("region not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("region not found")
	}

	// Check for unique constraint conflicts before restoration
	if existing.Code != "" {
		activeRegion, err := s.repo.GetByCode(ctx, existing.Code)
		if err == nil && activeRegion != nil {
			return fmt.Errorf("cannot restore region: another active region with code '%s' already exists", existing.Code)
		}
	}

	err = s.repo.Restore(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore region: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes a region (admin operation)
func (s *regionService) ForceDelete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("region ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid region ID format: %w", err)
	}

	// Check if region exists (including soft-deleted)
	existing, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("region not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("region not found")
	}

	err = s.repo.ForceDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete region: %w", err)
	}

	return nil
}
