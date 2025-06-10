package services

import (
	"context"
	"fmt"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"
	"regexp"
	"strings"
)

// locationTypeService implements the LocationTypeService interface
type locationTypeService struct {
	repo repositories.LocationTypeRepository
}

// NewLocationTypeService creates a new location type service instance
func NewLocationTypeService(repo repositories.LocationTypeRepository) LocationTypeService {
	return &locationTypeService{
		repo: repo,
	}
}

// GetByCode retrieves a location type by its code
func (s *locationTypeService) GetByCode(ctx context.Context, code string) (*dtos.LocationTypeResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("location type code cannot be empty")
	}

	// Normalize code
	code = strings.ToLower(strings.TrimSpace(code))

	locationType, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get location type by code: %w", err)
	}

	if locationType == nil {
		return nil, fmt.Errorf("location type not found")
	}

	return LocationTypeToResponse(locationType), nil
}

// GetAll retrieves all location types with pagination
func (s *locationTypeService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.LocationTypeListResponse, error) {
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

	locationTypes, total, err := s.repo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get location types: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.LocationTypeResponse, len(locationTypes))
	for i, locationType := range locationTypes {
		if response := LocationTypeToResponse(&locationType); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.LocationTypeListResponse{
		LocationTypes: responses,
		Pagination: dtos.PaginationResponse{
			Offset:      req.Offset,
			Limit:       req.Limit,
			Total:       total,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	}, nil
}

// validateSnakeCase validates if a string is in snake_case format
func (s *locationTypeService) validateSnakeCase(code string) error {
	// Snake case: lowercase letters, numbers, and underscores only
	// Must start with letter, no consecutive underscores, no trailing underscore
	pattern := `^[a-z][a-z0-9_]*[a-z0-9]$|^[a-z]$`
	matched, _ := regexp.MatchString(pattern, code)
	if !matched {
		return fmt.Errorf("code must be in snake_case format (lowercase letters, numbers, and underscores only)")
	}
	return nil
}

// Create creates a new location type
func (s *locationTypeService) Create(ctx context.Context, req *dtos.CreateLocationTypeRequest) (*dtos.LocationTypeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate required fields
	if strings.TrimSpace(req.Code) == "" {
		return nil, fmt.Errorf("location type code is required")
	}

	// Normalize and validate code
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))
	if err := s.validateSnakeCase(req.Code); err != nil {
		return nil, fmt.Errorf("invalid code format: %w", err)
	}

	// Check if location type with code already exists
	existing, err := s.repo.GetByCode(ctx, req.Code)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check existing location type: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("location type with code '%s' already exists", req.Code)
	}

	// Create location type model
	locationType := &models.LocationType{
		Code: req.Code,
	}

	// Set optional fields
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if description != "" {
			locationType.Description = &description
		}
	}

	err = s.repo.Create(ctx, locationType)
	if err != nil {
		return nil, fmt.Errorf("failed to create location type: %w", err)
	}

	return LocationTypeToResponse(locationType), nil
}

// Update updates an existing location type
func (s *locationTypeService) Update(ctx context.Context, code string, req *dtos.UpdateLocationTypeRequest) (*dtos.LocationTypeResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("location type code cannot be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Normalize code
	code = strings.ToLower(strings.TrimSpace(code))

	// Check if location type exists
	existing, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get location type: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("location type not found")
	}

	// Update fields
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if description != "" {
			existing.Description = &description
		} else {
			existing.Description = nil
		}
	}

	err = s.repo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update location type: %w", err)
	}

	return LocationTypeToResponse(existing), nil
}

// Delete deletes a location type
func (s *locationTypeService) Delete(ctx context.Context, code string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("location type code cannot be empty")
	}

	// Normalize code
	code = strings.ToLower(strings.TrimSpace(code))

	// Check if location type exists
	existing, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to get location type: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("location type not found")
	}

	err = s.repo.Delete(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to delete location type: %w", err)
	}

	return nil
}
