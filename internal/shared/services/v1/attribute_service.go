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

// AttributeService defines the business logic interface for attribute operations
type AttributeService interface {
	GetByID(ctx context.Context, id string) (*dtos.AttributeResponse, error)
	GetByCode(ctx context.Context, code string) (*dtos.AttributeResponse, error)
	GetByCategoryID(ctx context.Context, categoryID string) (*dtos.AttributeListResponse, error)
	GetByCategoryCode(ctx context.Context, categoryCode string) (*dtos.AttributeListResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.AttributeListResponse, error)
	GetAllWithDeleted(ctx context.Context, req *dtos.PaginationRequest) (*dtos.AttributeListResponse, error)
	Create(ctx context.Context, req *dtos.CreateAttributeRequest) (*dtos.AttributeResponse, error)
	Update(ctx context.Context, id string, req *dtos.UpdateAttributeRequest) (*dtos.AttributeResponse, error)
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
	GetStats(ctx context.Context, id string) (*dtos.AttributeStatsResponse, error)
}

// AttributeToResponse converts an Attribute model to AttributeResponse DTO
func AttributeToResponse(attribute *models.Attribute) *dtos.AttributeResponse {
	if attribute == nil {
		return nil
	}

	response := &dtos.AttributeResponse{
		ID:            attribute.ID,
		CategoryID:    attribute.CategoryID,
		Code:          attribute.Code,
		AttributeCode: attribute.AttributeCode,
		Name:          attribute.Name,
		IsActive:      attribute.IsActive,
		CreatedAt:     attribute.CreatedAt,
		UpdatedAt:     attribute.UpdatedAt,
	}

	// Convert category if loaded
	if attribute.Category != nil {
		response.Category = AttributeCategoryToResponse(attribute.Category)
	}

	return response
}

// attributeService implements the AttributeService interface
type attributeService struct {
	repo repositories.AttributeRepository
}

// NewAttributeService creates a new attribute service instance
func NewAttributeService(repo repositories.AttributeRepository) AttributeService {
	return &attributeService{
		repo: repo,
	}
}

// GetByID retrieves an attribute by its ID
func (s *attributeService) GetByID(ctx context.Context, id string) (*dtos.AttributeResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("attribute ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid attribute ID format: %w", err)
	}

	attribute, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute by ID: %w", err)
	}

	if attribute == nil {
		return nil, fmt.Errorf("attribute not found")
	}

	return AttributeToResponse(attribute), nil
}

// GetByCode retrieves an attribute by its code
func (s *attributeService) GetByCode(ctx context.Context, code string) (*dtos.AttributeResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("attribute code cannot be empty")
	}

	// Normalize code to lowercase
	code = strings.ToLower(strings.TrimSpace(code))

	attribute, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute by code: %w", err)
	}

	if attribute == nil {
		return nil, fmt.Errorf("attribute not found")
	}

	return AttributeToResponse(attribute), nil
}

// GetByCategoryID retrieves attributes by category ID
func (s *attributeService) GetByCategoryID(ctx context.Context, categoryID string) (*dtos.AttributeListResponse, error) {
	if strings.TrimSpace(categoryID) == "" {
		return nil, fmt.Errorf("category ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(categoryID); err != nil {
		return nil, fmt.Errorf("invalid category ID format: %w", err)
	}

	attributes, err := s.repo.GetByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes by category ID: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.AttributeResponse, len(attributes))
	for i, attribute := range attributes {
		if response := AttributeToResponse(&attribute); response != nil {
			responses[i] = *response
		}
	}

	return &dtos.AttributeListResponse{
		Success: true,
		Message: "Attributes retrieved successfully",
		Data:    responses,
		Pagination: dtos.PaginationResponse{
			Offset:      0,
			Limit:       len(responses),
			Total:       int64(len(responses)),
			HasNext:     false,
			HasPrevious: false,
		},
	}, nil
}

// GetByCategoryCode retrieves attributes by category code
func (s *attributeService) GetByCategoryCode(ctx context.Context, categoryCode string) (*dtos.AttributeListResponse, error) {
	if strings.TrimSpace(categoryCode) == "" {
		return nil, fmt.Errorf("category code cannot be empty")
	}

	// Normalize code to lowercase
	categoryCode = strings.ToLower(strings.TrimSpace(categoryCode))

	attributes, err := s.repo.GetByCategoryCode(ctx, categoryCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes by category code: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.AttributeResponse, len(attributes))
	for i, attribute := range attributes {
		if response := AttributeToResponse(&attribute); response != nil {
			responses[i] = *response
		}
	}

	return &dtos.AttributeListResponse{
		Success: true,
		Message: "Attributes retrieved successfully",
		Data:    responses,
		Pagination: dtos.PaginationResponse{
			Offset:      0,
			Limit:       len(responses),
			Total:       int64(len(responses)),
			HasNext:     false,
			HasPrevious: false,
		},
	}, nil
}

// GetAll retrieves all attributes with pagination
func (s *attributeService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.AttributeListResponse, error) {
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

	attributes, total, err := s.repo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.AttributeResponse, len(attributes))
	for i, attribute := range attributes {
		if response := AttributeToResponse(&attribute); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.AttributeListResponse{
		Success: true,
		Message: "Attributes retrieved successfully",
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

// GetAllWithDeleted retrieves all attributes including soft-deleted ones
func (s *attributeService) GetAllWithDeleted(ctx context.Context, req *dtos.PaginationRequest) (*dtos.AttributeListResponse, error) {
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

	attributes, total, err := s.repo.GetAllWithDeleted(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes with deleted: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.AttributeResponse, len(attributes))
	for i, attribute := range attributes {
		if response := AttributeToResponse(&attribute); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.AttributeListResponse{
		Success: true,
		Message: "Attributes with deleted retrieved successfully",
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

// Create creates a new attribute
func (s *attributeService) Create(ctx context.Context, req *dtos.CreateAttributeRequest) (*dtos.AttributeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate code format
	code := strings.ToLower(strings.TrimSpace(req.Code))
	if !isValidAttributeSnakeCase(code) {
		return nil, fmt.Errorf("attribute code must be in snake_case format")
	}

	// Check if code already exists within the category
	existingAttribute, err := s.repo.GetByCodeWithDeleted(ctx, code)
	if err == nil && existingAttribute != nil && existingAttribute.CategoryID == req.CategoryID {
		return nil, fmt.Errorf("attribute with code '%s' already exists in this category", code)
	}

	// Create new attribute model
	attribute := &models.Attribute{
		CategoryID:    req.CategoryID,
		Code:          code,
		AttributeCode: code, // Default to same as code
		Name:          strings.TrimSpace(req.Name),
		IsActive:      true, // Default to active
	}

	// Override attributeCode if provided
	if req.AttributeCode != nil && strings.TrimSpace(*req.AttributeCode) != "" {
		attributeCode := strings.ToLower(strings.TrimSpace(*req.AttributeCode))
		if !isValidAttributeSnakeCase(attributeCode) {
			return nil, fmt.Errorf("attribute code must be in snake_case format")
		}
		attribute.AttributeCode = attributeCode
	}

	// Override isActive if provided
	if req.IsActive != nil {
		attribute.IsActive = *req.IsActive
	}

	// Create in repository
	err = s.repo.Create(ctx, attribute)
	if err != nil {
		return nil, fmt.Errorf("failed to create attribute: %w", err)
	}

	return AttributeToResponse(attribute), nil
}

// Update updates an existing attribute
func (s *attributeService) Update(ctx context.Context, id string, req *dtos.UpdateAttributeRequest) (*dtos.AttributeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("attribute ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid attribute ID format: %w", err)
	}

	// Get existing attribute (including inactive ones)
	existing, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing attribute: %w", err)
	}

	if existing == nil {
		return nil, fmt.Errorf("attribute not found")
	}

	// Create updated attribute from existing
	updated := &models.Attribute{
		ID:            existing.ID,
		CategoryID:    existing.CategoryID,
		Code:          existing.Code,
		AttributeCode: existing.AttributeCode,
		Name:          existing.Name,
		IsActive:      existing.IsActive,
		CreatedAt:     existing.CreatedAt,
		UpdatedAt:     existing.UpdatedAt,
	}

	// Apply updates
	if req.CategoryID != nil {
		updated.CategoryID = *req.CategoryID
	}
	if req.Code != nil {
		code := strings.ToLower(strings.TrimSpace(*req.Code))
		if !isValidAttributeSnakeCase(code) {
			return nil, fmt.Errorf("attribute code must be in snake_case format")
		}
		updated.Code = code
	}
	if req.AttributeCode != nil {
		attributeCode := strings.ToLower(strings.TrimSpace(*req.AttributeCode))
		if !isValidAttributeSnakeCase(attributeCode) {
			return nil, fmt.Errorf("attribute code must be in snake_case format")
		}
		updated.AttributeCode = attributeCode
	}
	if req.Name != nil {
		updated.Name = strings.TrimSpace(*req.Name)
	}
	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}

	// Update in repository
	err = s.repo.Update(ctx, id, updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update attribute: %w", err)
	}

	return AttributeToResponse(updated), nil
}

// Delete soft deletes an attribute
func (s *attributeService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("attribute ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid attribute ID format: %w", err)
	}

	// Check if attribute can be deleted (no active partner mappings)
	canDelete, err := s.repo.CanBeDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if attribute can be deleted: %w", err)
	}

	if !canDelete {
		return fmt.Errorf("cannot delete attribute: it has active partner mappings")
	}

	// Delete from repository
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete attribute: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted attribute
func (s *attributeService) Restore(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("attribute ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid attribute ID format: %w", err)
	}

	// Restore from repository
	err := s.repo.Restore(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore attribute: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes an attribute
func (s *attributeService) ForceDelete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("attribute ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid attribute ID format: %w", err)
	}

	// Force delete from repository
	err := s.repo.ForceDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to force delete attribute: %w", err)
	}

	return nil
}

// GetStats retrieves statistics for an attribute
func (s *attributeService) GetStats(ctx context.Context, id string) (*dtos.AttributeStatsResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("attribute ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid attribute ID format: %w", err)
	}

	// Get attribute details for created date
	attribute, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}

	if attribute == nil {
		return nil, fmt.Errorf("attribute not found")
	}

	// Get stats from repository
	stats, err := s.repo.GetStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute stats: %w", err)
	}

	// Convert to response DTO
	return &dtos.AttributeStatsResponse{
		ID:             stats.ID,
		Code:           stats.Code,
		Name:           stats.Name,
		CategoryCode:   stats.CategoryCode,
		CategoryName:   stats.CategoryName,
		TotalMappings:  int(stats.TotalMappings),
		ActiveMappings: int(stats.ActiveMappings),
		UniquePartners: int(stats.UniquePartners),
		CreatedAt:      attribute.CreatedAt,
	}, nil
}
