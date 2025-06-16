package services

import (
	"context"
	"fmt"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"
	"strings"
)

// regionTypeService implements the RegionTypeService interface
type regionTypeService struct {
	repo repositories.RegionTypeRepository
}

// NewRegionTypeService creates a new region type service instance
func NewRegionTypeService(repo repositories.RegionTypeRepository) RegionTypeService {
	return &regionTypeService{
		repo: repo,
	}
}

// GetByCode retrieves a region type by its code
func (s *regionTypeService) GetByCode(ctx context.Context, code string) (*dtos.RegionTypeResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("region type code cannot be empty")
	}

	// Normalize code to lowercase snake_case
	code = strings.ToLower(strings.TrimSpace(code))

	regionType, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get region type by code: %w", err)
	}

	if regionType == nil {
		return nil, fmt.Errorf("region type not found")
	}

	return RegionTypeToResponse(regionType), nil
}

// GetAll retrieves all region types with pagination
func (s *regionTypeService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.RegionTypeListResponse, error) {
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

	regionTypes, total, err := s.repo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get region types: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.RegionTypeResponse, len(regionTypes))
	for i, regionType := range regionTypes {
		if response := RegionTypeToResponse(&regionType); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.RegionTypeListResponse{
		Success: true,
		Message: "Region types retrieved successfully",
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

// GetAllWithDeleted retrieves all region types including soft-deleted ones (admin operation)
func (s *regionTypeService) GetAllWithDeleted(ctx context.Context, req *dtos.PaginationRequest) (*dtos.RegionTypeListResponse, error) {
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

	regionTypes, total, err := s.repo.GetAllWithDeleted(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get region types with deleted: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.RegionTypeResponse, len(regionTypes))
	for i, regionType := range regionTypes {
		if response := RegionTypeToResponse(&regionType); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.RegionTypeListResponse{
		Success: true,
		Message: "Region types retrieved successfully (including deleted)",
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

// Create creates a new region type
func (s *regionTypeService) Create(ctx context.Context, req *dtos.CreateRegionTypeRequest) (*dtos.RegionTypeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate required fields
	if strings.TrimSpace(req.Code) == "" {
		return nil, fmt.Errorf("region type code is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("region type name is required")
	}

	// Normalize and validate code (snake_case)
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))
	if !isValidSnakeCase(req.Code) {
		return nil, fmt.Errorf("region type code must be in snake_case format")
	}

	// Check if region type with code already exists
	existing, err := s.repo.GetByCode(ctx, req.Code)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check existing region type: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("region type with code '%s' already exists", req.Code)
	}

	// Create region type model
	regionType := &models.RegionType{
		Code:     req.Code,
		Name:     strings.TrimSpace(req.Name),
		IsActive: true, // Default to active
	}

	// Set optional fields
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if description != "" {
			regionType.Description = &description
		}
	}

	if req.IsActive != nil {
		regionType.IsActive = *req.IsActive
	}

	err = s.repo.Create(ctx, regionType)
	if err != nil {
		return nil, fmt.Errorf("failed to create region type: %w", err)
	}

	return RegionTypeToResponse(regionType), nil
}

// Update updates an existing region type
func (s *regionTypeService) Update(ctx context.Context, code string, req *dtos.UpdateRegionTypeRequest) (*dtos.RegionTypeResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("region type code cannot be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Normalize code
	code = strings.ToLower(strings.TrimSpace(code))

	// Check if region type exists
	existing, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing region type: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("region type not found")
	}

	// Update fields if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("region type name cannot be empty")
		}
		existing.Name = name
	}

	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if description == "" {
			existing.Description = nil
		} else {
			existing.Description = &description
		}
	}

	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	err = s.repo.Update(ctx, code, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update region type: %w", err)
	}

	return RegionTypeToResponse(existing), nil
}

// Delete deletes a region type by code
func (s *regionTypeService) Delete(ctx context.Context, code string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("region type code cannot be empty")
	}

	// Normalize code
	code = strings.ToLower(strings.TrimSpace(code))

	// Check if region type exists
	existing, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to get existing region type: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("region type not found")
	}

	err = s.repo.Delete(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to delete region type: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted region type (admin operation)
func (s *regionTypeService) Restore(ctx context.Context, code string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("region type code cannot be empty")
	}

	// Normalize code
	code = strings.ToLower(strings.TrimSpace(code))

	// Check if region type exists (including soft-deleted)
	existing, err := s.repo.GetByCodeWithDeleted(ctx, code)
	if err != nil {
		return fmt.Errorf("region type not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("region type not found")
	}

	// Check for unique constraint conflicts before restoration
	activeRegionType, err := s.repo.GetByCode(ctx, code)
	if err == nil && activeRegionType != nil {
		return fmt.Errorf("cannot restore region type: another active region type with code '%s' already exists", code)
	}

	err = s.repo.Restore(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to restore region type: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes a region type (admin operation)
func (s *regionTypeService) ForceDelete(ctx context.Context, code string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("region type code cannot be empty")
	}

	// Normalize code
	code = strings.ToLower(strings.TrimSpace(code))

	// Check if region type exists (including soft-deleted)
	existing, err := s.repo.GetByCodeWithDeleted(ctx, code)
	if err != nil {
		return fmt.Errorf("region type not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("region type not found")
	}

	err = s.repo.ForceDelete(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to permanently delete region type: %w", err)
	}

	return nil
}

// isValidSnakeCase checks if a string is in valid snake_case format
func isValidSnakeCase(s string) bool {
	if s == "" {
		return false
	}

	// Must contain only lowercase letters, numbers, and underscores
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}

	// Cannot start or end with underscore
	if strings.HasPrefix(s, "_") || strings.HasSuffix(s, "_") {
		return false
	}

	// Cannot have consecutive underscores
	if strings.Contains(s, "__") {
		return false
	}

	return true
}
