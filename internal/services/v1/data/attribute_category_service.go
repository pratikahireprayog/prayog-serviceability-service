package services

import (
	"context"
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"github.com/google/uuid"
)

// AttributeCategoryService defines the business logic interface for attribute category operations
type AttributeCategoryService interface {
	GetByID(ctx context.Context, id string) (*dtos.AttributeCategoryResponse, error)
	GetByCode(ctx context.Context, code string) (*dtos.AttributeCategoryResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.AttributeCategoryListResponse, error)
	GetAllWithDeleted(ctx context.Context, req *dtos.PaginationRequest) (*dtos.AttributeCategoryListResponse, error)
	Create(ctx context.Context, req *dtos.CreateAttributeCategoryRequest) (*dtos.AttributeCategoryResponse, error)
	Update(ctx context.Context, id string, req *dtos.UpdateAttributeCategoryRequest) (*dtos.AttributeCategoryResponse, error)
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
	GetStats(ctx context.Context, id string) (*dtos.AttributeCategoryStatsResponse, error)
}

// AttributeCategoryToResponse converts an AttributeCategory model to AttributeCategoryResponse DTO
func AttributeCategoryToResponse(category *models.AttributeCategory) *dtos.AttributeCategoryResponse {
	if category == nil {
		return nil
	}

	return &dtos.AttributeCategoryResponse{
		ID:        category.ID,
		Code:      category.Code,
		Name:      category.Name,
		IsActive:  category.IsActive,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}

// attributeCategoryService implements the AttributeCategoryService interface
type attributeCategoryService struct {
	repo repositories.AttributeCategoryRepository
}

// NewAttributeCategoryService creates a new attribute category service instance
func NewAttributeCategoryService(repo repositories.AttributeCategoryRepository) AttributeCategoryService {
	return &attributeCategoryService{
		repo: repo,
	}
}

// GetByID retrieves an attribute category by its ID
func (s *attributeCategoryService) GetByID(ctx context.Context, id string) (*dtos.AttributeCategoryResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("attribute category ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute category by ID: %w", err)
	}

	if category == nil {
		return nil, fmt.Errorf("attribute category not found")
	}

	return AttributeCategoryToResponse(category), nil
}

// GetByCode retrieves an attribute category by its code
func (s *attributeCategoryService) GetByCode(ctx context.Context, code string) (*dtos.AttributeCategoryResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("attribute category code cannot be empty")
	}

	// Normalize code to lowercase
	code = strings.ToLower(strings.TrimSpace(code))

	category, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute category by code: %w", err)
	}

	if category == nil {
		return nil, fmt.Errorf("attribute category not found")
	}

	return AttributeCategoryToResponse(category), nil
}

// GetAll retrieves all attribute categories with pagination
func (s *attributeCategoryService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.AttributeCategoryListResponse, error) {
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

	categories, total, err := s.repo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute categories: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.AttributeCategoryResponse, len(categories))
	for i, category := range categories {
		if response := AttributeCategoryToResponse(&category); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.AttributeCategoryListResponse{
		Success: true,
		Message: "Attribute categories retrieved successfully",
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

// GetAllWithDeleted retrieves all attribute categories including soft-deleted ones
func (s *attributeCategoryService) GetAllWithDeleted(ctx context.Context, req *dtos.PaginationRequest) (*dtos.AttributeCategoryListResponse, error) {
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

	categories, total, err := s.repo.GetAllWithDeleted(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute categories with deleted: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.AttributeCategoryResponse, len(categories))
	for i, category := range categories {
		if response := AttributeCategoryToResponse(&category); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.AttributeCategoryListResponse{
		Success: true,
		Message: "Attribute categories with deleted retrieved successfully",
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

// Create creates a new attribute category
func (s *attributeCategoryService) Create(ctx context.Context, req *dtos.CreateAttributeCategoryRequest) (*dtos.AttributeCategoryResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate code format
	code := strings.ToLower(strings.TrimSpace(req.Code))
	if !isValidAttributeSnakeCase(code) {
		return nil, fmt.Errorf("category code must be in snake_case format")
	}

	// Check if code already exists
	existingCategory, err := s.repo.GetByCodeWithDeleted(ctx, code)
	if err == nil && existingCategory != nil {
		return nil, fmt.Errorf("attribute category with code '%s' already exists", code)
	}

	// Create new category model
	category := &models.AttributeCategory{
		Code:     code,
		Name:     strings.TrimSpace(req.Name),
		IsActive: true, // Default to active
	}

	// Override isActive if provided
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	// Create in repository
	err = s.repo.Create(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to create attribute category: %w", err)
	}

	return AttributeCategoryToResponse(category), nil
}

// Update updates an existing attribute category
func (s *attributeCategoryService) Update(ctx context.Context, id string, req *dtos.UpdateAttributeCategoryRequest) (*dtos.AttributeCategoryResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("attribute category ID cannot be empty")
	}

	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	// Get existing category (including inactive ones)
	category, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute category: %w", err)
	}

	if category == nil {
		return nil, fmt.Errorf("attribute category not found")
	}

	// Update fields if provided
	if req.Name != nil {
		category.Name = strings.TrimSpace(*req.Name)
	}

	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	// Update in repository
	err = s.repo.Update(ctx, id, category)
	if err != nil {
		return nil, fmt.Errorf("failed to update attribute category: %w", err)
	}

	// Get updated category to return
	updatedCategory, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated attribute category: %w", err)
	}

	return AttributeCategoryToResponse(updatedCategory), nil
}

// Delete performs soft delete on an attribute category
func (s *attributeCategoryService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("attribute category ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	// Check if category can be deleted
	canDelete, err := s.repo.CanBeDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if attribute category can be deleted: %w", err)
	}

	if !canDelete {
		return fmt.Errorf("cannot delete attribute category: it has active attributes")
	}

	// Perform soft delete
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete attribute category: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted attribute category
func (s *attributeCategoryService) Restore(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("attribute category ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	// Perform restore
	err := s.repo.Restore(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore attribute category: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes an attribute category
func (s *attributeCategoryService) ForceDelete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("attribute category ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	// Perform force delete
	err := s.repo.ForceDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to force delete attribute category: %w", err)
	}

	return nil
}

// GetStats retrieves statistics for an attribute category
func (s *attributeCategoryService) GetStats(ctx context.Context, id string) (*dtos.AttributeCategoryStatsResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("attribute category ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid attribute category ID format: %w", err)
	}

	// Get stats from repository
	stats, err := s.repo.GetStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute category stats: %w", err)
	}

	// Convert to response DTO
	return &dtos.AttributeCategoryStatsResponse{
		ID:               stats.ID,
		Code:             stats.Code,
		Name:             stats.Name,
		TotalAttributes:  int(stats.TotalAttributes),
		ActiveAttributes: int(stats.ActiveAttributes),
		TotalMappings:    int(stats.TotalMappings),
		ActiveMappings:   int(stats.ActiveMappings),
		UniquePartners:   int(stats.UniquePartners),
	}, nil
}

// Helper functions

// isValidAttributeSnakeCase validates that a string is in valid snake_case format
func isValidAttributeSnakeCase(s string) bool {
	if s == "" {
		return false
	}

	// Must start with a letter or number
	if !((s[0] >= 'a' && s[0] <= 'z') || (s[0] >= '0' && s[0] <= '9')) {
		return false
	}

	// Must end with a letter or number
	lastChar := s[len(s)-1]
	if !((lastChar >= 'a' && lastChar <= 'z') || (lastChar >= '0' && lastChar <= '9')) {
		return false
	}

	// Check for valid characters: lowercase letters, numbers, and underscores
	// No consecutive underscores allowed
	for i, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}

		// Check for consecutive underscores
		if r == '_' && i < len(s)-1 && s[i+1] == '_' {
			return false
		}
	}

	return true
}
