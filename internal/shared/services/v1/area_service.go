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

// areaService implements the AreaService interface
type areaService struct {
	repo repositories.AreaRepository
}

// NewAreaService creates a new area service instance
func NewAreaService(repo repositories.AreaRepository) AreaService {
	return &areaService{
		repo: repo,
	}
}

// validateAreaSnakeCase validates that a string is in snake_case format
func validateAreaSnakeCase(s string) bool {
	pattern := `^[a-z0-9]+(_[a-z0-9]+)*$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// GetByID retrieves an area by its ID
func (s *areaService) GetByID(ctx context.Context, id string) (*dtos.AreaResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("area ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid area ID format: %w", err)
	}

	area, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get area by ID: %w", err)
	}

	if area == nil {
		return nil, fmt.Errorf("area not found")
	}

	return AreaToResponse(area), nil
}

// GetByCode retrieves an area by its code
func (s *areaService) GetByCode(ctx context.Context, code string) (*dtos.AreaResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("area code cannot be empty")
	}

	// Normalize code (areas can use various formats, not strictly snake_case)
	code = strings.ToLower(strings.TrimSpace(code))

	area, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get area by code: %w", err)
	}

	if area == nil {
		return nil, fmt.Errorf("area not found")
	}

	return AreaToResponse(area), nil
}

// GetAll retrieves all areas with pagination
func (s *areaService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.AreaListResponse, error) {
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

	areas, total, err := s.repo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get areas: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.AreaResponse, len(areas))
	for i, area := range areas {
		if response := AreaToResponse(&area); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.AreaListResponse{
		Success: true,
		Message: "Areas retrieved successfully",
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

// GetByCityID retrieves all areas for a specific city
func (s *areaService) GetByCityID(ctx context.Context, cityID string) ([]dtos.AreaResponse, error) {
	if strings.TrimSpace(cityID) == "" {
		return nil, fmt.Errorf("city ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(cityID); err != nil {
		return nil, fmt.Errorf("invalid city ID format: %w", err)
	}

	areas, err := s.repo.GetByCityID(ctx, cityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get areas by city ID: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.AreaResponse, len(areas))
	for i, area := range areas {
		if response := AreaToResponse(&area); response != nil {
			responses[i] = *response
		}
	}

	return responses, nil
}

// Create creates a new area
func (s *areaService) Create(ctx context.Context, req *dtos.CreateAreaRequest) (*dtos.AreaResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate required fields
	if strings.TrimSpace(req.Code) == "" {
		return nil, fmt.Errorf("area code is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("area name is required")
	}

	// Normalize code
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))

	// Check if area with code already exists
	existing, err := s.repo.GetByCode(ctx, req.Code)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check existing area: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("area with code '%s' already exists", req.Code)
	}

	// Create area model
	area := &models.Area{
		ID:       uuid.New(),
		Code:     req.Code,
		Name:     strings.TrimSpace(req.Name),
		IsActive: true, // Default to active
	}

	// Set optional fields
	if req.CityID != nil {
		area.CityID = req.CityID
	}

	if req.CityCode != nil {
		cityCode := strings.ToLower(strings.TrimSpace(*req.CityCode))
		if cityCode != "" {
			area.CityCode = &cityCode
		}
	}

	if req.IsActive != nil {
		area.IsActive = *req.IsActive
	}

	err = s.repo.Create(ctx, area)
	if err != nil {
		return nil, fmt.Errorf("failed to create area: %w", err)
	}

	return AreaToResponse(area), nil
}

// Update updates an existing area
func (s *areaService) Update(ctx context.Context, id string, req *dtos.UpdateAreaRequest) (*dtos.AreaResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("area ID cannot be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid area ID format: %w", err)
	}

	// Check if area exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get area: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("area not found")
	}

	// Create updated area model
	updated := *existing

	// Update only provided fields
	if req.CityID != nil {
		updated.CityID = req.CityID
	}

	if req.CityCode != nil {
		cityCode := strings.ToLower(strings.TrimSpace(*req.CityCode))
		if cityCode != "" {
			updated.CityCode = &cityCode
		} else {
			updated.CityCode = nil
		}
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("area name cannot be empty")
		}
		updated.Name = name
	}

	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}

	err = s.repo.Update(ctx, id, &updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update area: %w", err)
	}

	return AreaToResponse(&updated), nil
}

// Delete deletes an area by ID
func (s *areaService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("area ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid area ID format: %w", err)
	}

	// Check if area exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get area: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("area not found")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete area: %w", err)
	}

	return nil
}
