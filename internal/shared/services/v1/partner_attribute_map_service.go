package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"github.com/google/uuid"
)

// PartnerAttributeMapService defines the business logic interface for partner attribute mapping operations
type PartnerAttributeMapService interface {
	GetByID(ctx context.Context, id string) (*dtos.PartnerAttributeMapResponse, error)
	GetByPartnerID(ctx context.Context, partnerID string) (*dtos.PartnerAttributeMapListResponse, error)
	GetByPartnerCode(ctx context.Context, partnerCode string) (*dtos.PartnerAttributeMapListResponse, error)
	GetByAttributeCode(ctx context.Context, attributeCode string) (*dtos.PartnerAttributeMapListResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.PartnerAttributeMapListResponse, error)
	GetAllWithDeleted(ctx context.Context, req *dtos.PaginationRequest) (*dtos.PartnerAttributeMapListResponse, error)
	GetWithFilters(ctx context.Context, filters *dtos.PartnerAttributeMapFilterRequest) (*dtos.PartnerAttributeMapListResponse, error)
	GetPartnerCodesByAttribute(ctx context.Context, attributeCode string) (*dtos.PartnerCodesByAttributeResponse, error)
	GetAttributesByPartner(ctx context.Context, partnerCode string) (*dtos.AttributeListResponse, error)
	Create(ctx context.Context, req *dtos.CreatePartnerAttributeMapRequest) (*dtos.PartnerAttributeMapResponse, error)
	Update(ctx context.Context, id string, req *dtos.UpdatePartnerAttributeMapRequest) (*dtos.PartnerAttributeMapResponse, error)
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, id string) error
	BulkCreate(ctx context.Context, req *dtos.BulkCreatePartnerAttributeMapRequest) (*dtos.BulkCreatePartnerAttributeMapResponse, error)
	BulkDelete(ctx context.Context, req *dtos.BulkDeletePartnerAttributeMapRequest) (*dtos.BulkDeletePartnerAttributeMapResponse, error)
	GetStats(ctx context.Context, partnerCode string) (*dtos.PartnerAttributeStatsResponse, error)
}

// PartnerAttributeMapToResponse converts a PartnerAttributeMap model to PartnerAttributeMapResponse DTO
func PartnerAttributeMapToResponse(mapping *models.PartnerAttributeMap) *dtos.PartnerAttributeMapResponse {
	if mapping == nil {
		return nil
	}

	response := &dtos.PartnerAttributeMapResponse{
		ID:            mapping.ID,
		PartnerID:     mapping.PartnerID,
		PartnerCode:   mapping.PartnerCode,
		AttributeID:   mapping.AttributeID,
		AttributeCode: mapping.AttributeCode,
		IsActive:      mapping.IsActive,
		CreatedAt:     mapping.CreatedAt,
		UpdatedAt:     mapping.UpdatedAt,
	}

	// Convert attribute if loaded
	if mapping.Attribute != nil {
		response.Attribute = AttributeToResponse(mapping.Attribute)
	}

	return response
}

// partnerAttributeMapService implements the PartnerAttributeMapService interface
type partnerAttributeMapService struct {
	repo                  repositories.PartnerAttributeMapRepository
	attributeRepo         repositories.AttributeRepository
	attributeCategoryRepo repositories.AttributeCategoryRepository
}

// NewPartnerAttributeMapService creates a new partner attribute map service instance
func NewPartnerAttributeMapService(
	repo repositories.PartnerAttributeMapRepository,
	attributeRepo repositories.AttributeRepository,
	attributeCategoryRepo repositories.AttributeCategoryRepository,
) PartnerAttributeMapService {
	return &partnerAttributeMapService{
		repo:                  repo,
		attributeRepo:         attributeRepo,
		attributeCategoryRepo: attributeCategoryRepo,
	}
}

// GetByID retrieves a partner attribute mapping by its ID
func (s *partnerAttributeMapService) GetByID(ctx context.Context, id string) (*dtos.PartnerAttributeMapResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("partner attribute mapping ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid partner attribute mapping ID format: %w", err)
	}

	mapping, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mapping by ID: %w", err)
	}

	if mapping == nil {
		return nil, fmt.Errorf("partner attribute mapping not found")
	}

	return PartnerAttributeMapToResponse(mapping), nil
}

// GetByPartnerID retrieves all mappings for a partner ID
func (s *partnerAttributeMapService) GetByPartnerID(ctx context.Context, partnerID string) (*dtos.PartnerAttributeMapListResponse, error) {
	if strings.TrimSpace(partnerID) == "" {
		return nil, fmt.Errorf("partner ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(partnerID); err != nil {
		return nil, fmt.Errorf("invalid partner ID format: %w", err)
	}

	mappings, err := s.repo.GetByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings by partner ID: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.PartnerAttributeMapResponse, len(mappings))
	for i, mapping := range mappings {
		if response := PartnerAttributeMapToResponse(&mapping); response != nil {
			responses[i] = *response
		}
	}

	return &dtos.PartnerAttributeMapListResponse{
		Success: true,
		Message: "Partner attribute mappings retrieved successfully",
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

// GetByPartnerCode retrieves all mappings for a partner code
func (s *partnerAttributeMapService) GetByPartnerCode(ctx context.Context, partnerCode string) (*dtos.PartnerAttributeMapListResponse, error) {
	if strings.TrimSpace(partnerCode) == "" {
		return nil, fmt.Errorf("partner code cannot be empty")
	}

	// Normalize partner code
	partnerCode = strings.TrimSpace(partnerCode)

	mappings, err := s.repo.GetByPartnerCode(ctx, partnerCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings by partner code: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.PartnerAttributeMapResponse, len(mappings))
	for i, mapping := range mappings {
		if response := PartnerAttributeMapToResponse(&mapping); response != nil {
			responses[i] = *response
		}
	}

	return &dtos.PartnerAttributeMapListResponse{
		Success: true,
		Message: "Partner attribute mappings retrieved successfully",
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

// GetByAttributeCode retrieves all mappings for an attribute code
func (s *partnerAttributeMapService) GetByAttributeCode(ctx context.Context, attributeCode string) (*dtos.PartnerAttributeMapListResponse, error) {
	if strings.TrimSpace(attributeCode) == "" {
		return nil, fmt.Errorf("attribute code cannot be empty")
	}

	// Normalize attribute code
	attributeCode = strings.ToLower(strings.TrimSpace(attributeCode))

	mappings, err := s.repo.GetByAttributeCode(ctx, attributeCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings by attribute code: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.PartnerAttributeMapResponse, len(mappings))
	for i, mapping := range mappings {
		if response := PartnerAttributeMapToResponse(&mapping); response != nil {
			responses[i] = *response
		}
	}

	return &dtos.PartnerAttributeMapListResponse{
		Success: true,
		Message: "Partner attribute mappings retrieved successfully",
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

// GetAll retrieves all partner attribute mappings with pagination
func (s *partnerAttributeMapService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.PartnerAttributeMapListResponse, error) {
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

	mappings, total, err := s.repo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.PartnerAttributeMapResponse, len(mappings))
	for i, mapping := range mappings {
		if response := PartnerAttributeMapToResponse(&mapping); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.PartnerAttributeMapListResponse{
		Success: true,
		Message: "Partner attribute mappings retrieved successfully",
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

// GetAllWithDeleted retrieves all partner attribute mappings including soft-deleted ones
func (s *partnerAttributeMapService) GetAllWithDeleted(ctx context.Context, req *dtos.PaginationRequest) (*dtos.PartnerAttributeMapListResponse, error) {
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

	mappings, total, err := s.repo.GetAllWithDeleted(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings with deleted: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.PartnerAttributeMapResponse, len(mappings))
	for i, mapping := range mappings {
		if response := PartnerAttributeMapToResponse(&mapping); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.PartnerAttributeMapListResponse{
		Success: true,
		Message: "Partner attribute mappings with deleted retrieved successfully",
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

// GetWithFilters retrieves partner attribute mappings with filters
func (s *partnerAttributeMapService) GetWithFilters(ctx context.Context, filters *dtos.PartnerAttributeMapFilterRequest) (*dtos.PartnerAttributeMapListResponse, error) {
	if filters == nil {
		return s.GetAll(ctx, &dtos.PaginationRequest{Offset: 0, Limit: 10})
	}

	// Set default pagination if not provided
	limit := 10
	offset := 0
	if filters.Limit != nil {
		if *filters.Limit > 0 && *filters.Limit <= 100 {
			limit = *filters.Limit
		}
	}
	if filters.Offset != nil && *filters.Offset >= 0 {
		offset = *filters.Offset
	}

	// Build repository filters
	repoFilters := &repositories.PartnerAttributeMapFilters{
		Limit:  limit,
		Offset: offset,
	}

	if filters.PartnerID != nil {
		repoFilters.PartnerID = filters.PartnerID
	}

	if filters.PartnerCode != nil && strings.TrimSpace(*filters.PartnerCode) != "" {
		partnerCode := strings.TrimSpace(*filters.PartnerCode)
		repoFilters.PartnerCode = &partnerCode
	}
	if filters.AttributeCode != nil && strings.TrimSpace(*filters.AttributeCode) != "" {
		attributeCode := strings.ToLower(strings.TrimSpace(*filters.AttributeCode))
		repoFilters.AttributeCode = &attributeCode
	}
	if filters.AttributeID != nil {
		repoFilters.AttributeID = filters.AttributeID
	}
	if filters.CategoryID != nil {
		repoFilters.CategoryID = filters.CategoryID
	}
	if filters.IsActive != nil {
		repoFilters.IsActive = filters.IsActive
	}

	mappings, total, err := s.repo.GetWithFilters(ctx, repoFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute mappings with filters: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.PartnerAttributeMapResponse, len(mappings))
	for i, mapping := range mappings {
		if response := PartnerAttributeMapToResponse(&mapping); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(offset+limit) < total
	hasPrevious := offset > 0

	return &dtos.PartnerAttributeMapListResponse{
		Success: true,
		Message: "Partner attribute mappings retrieved successfully",
		Data:    responses,
		Pagination: dtos.PaginationResponse{
			Offset:      offset,
			Limit:       limit,
			Total:       total,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	}, nil
}

// GetPartnerCodesByAttribute retrieves partner codes by attribute code (KEY FUNCTIONALITY)
func (s *partnerAttributeMapService) GetPartnerCodesByAttribute(ctx context.Context, attributeCode string) (*dtos.PartnerCodesByAttributeResponse, error) {
	if strings.TrimSpace(attributeCode) == "" {
		return nil, fmt.Errorf("attribute code cannot be empty")
	}

	// Normalize attribute code
	attributeCode = strings.ToLower(strings.TrimSpace(attributeCode))

	// Get attribute details
	attribute, err := s.attributeRepo.GetByCode(ctx, attributeCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}

	if attribute == nil {
		return nil, fmt.Errorf("attribute not found")
	}

	// Get partner codes from repository
	partnerCodes, err := s.repo.GetPartnerCodesByAttribute(ctx, attributeCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner codes by attribute: %w", err)
	}

	// Get mappings to calculate statistics
	mappings, err := s.repo.GetByAttributeCode(ctx, attributeCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get mappings for statistics: %w", err)
	}

	// Calculate statistics
	totalPartners := len(mappings)
	activePartners := 0
	inactivePartners := 0
	for _, mapping := range mappings {
		if mapping.IsActive {
			activePartners++
		} else {
			inactivePartners++
		}
	}

	return &dtos.PartnerCodesByAttributeResponse{
		AttributeCode:    attributeCode,
		AttributeID:      attribute.ID,
		AttributeName:    attribute.Name,
		PartnerCodes:     partnerCodes,
		TotalPartners:    totalPartners,
		ActivePartners:   activePartners,
		InactivePartners: inactivePartners,
	}, nil
}

// GetAttributesByPartner retrieves attributes mapped to a partner
func (s *partnerAttributeMapService) GetAttributesByPartner(ctx context.Context, partnerCode string) (*dtos.AttributeListResponse, error) {
	if strings.TrimSpace(partnerCode) == "" {
		return nil, fmt.Errorf("partner code cannot be empty")
	}

	// Normalize partner code
	partnerCode = strings.TrimSpace(partnerCode)

	// Get attributes from repository
	attributes, err := s.repo.GetAttributesByPartner(ctx, partnerCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes by partner: %w", err)
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

// Create creates a new partner attribute mapping
func (s *partnerAttributeMapService) Create(ctx context.Context, req *dtos.CreatePartnerAttributeMapRequest) (*dtos.PartnerAttributeMapResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate inputs
	partnerCode := strings.TrimSpace(req.PartnerCode)
	if partnerCode == "" {
		return nil, fmt.Errorf("partner code cannot be empty")
	}

	// Validate attribute exists
	attribute, err := s.attributeRepo.GetByID(ctx, req.AttributeID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to validate attribute: %w", err)
	}

	if attribute == nil {
		return nil, fmt.Errorf("attribute not found")
	}

	// Check if mapping already exists
	existing, err := s.repo.GetByPartnerAndAttribute(ctx, partnerCode, req.AttributeID.String())
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check existing mapping: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("partner attribute mapping already exists for partner '%s' and attribute '%s'", partnerCode, attribute.Code)
	}

	// Create new mapping model
	mapping := &models.PartnerAttributeMap{
		PartnerID:     req.PartnerID,
		PartnerCode:   partnerCode,
		AttributeID:   req.AttributeID,
		AttributeCode: attribute.Code,
		IsActive:      true, // Default to active
	}

	// Override isActive if provided
	if req.IsActive != nil {
		mapping.IsActive = *req.IsActive
	}

	// Create in repository
	err = s.repo.Create(ctx, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create partner attribute mapping: %w", err)
	}

	return PartnerAttributeMapToResponse(mapping), nil
}

// Update updates an existing partner attribute mapping
func (s *partnerAttributeMapService) Update(ctx context.Context, id string, req *dtos.UpdatePartnerAttributeMapRequest) (*dtos.PartnerAttributeMapResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("partner attribute mapping ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid partner attribute mapping ID format: %w", err)
	}

	// Get existing mapping (including inactive ones)
	existing, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing mapping: %w", err)
	}

	if existing == nil {
		return nil, fmt.Errorf("partner attribute mapping not found")
	}

	// Create updated mapping from existing
	updated := &models.PartnerAttributeMap{
		ID:            existing.ID,
		PartnerID:     existing.PartnerID,
		PartnerCode:   existing.PartnerCode,
		AttributeID:   existing.AttributeID,
		AttributeCode: existing.AttributeCode,
		IsActive:      existing.IsActive,
		CreatedAt:     existing.CreatedAt,
		UpdatedAt:     existing.UpdatedAt,
	}

	// Apply updates
	if req.PartnerID != nil {
		updated.PartnerID = req.PartnerID
	}

	if req.PartnerCode != nil {
		partnerCode := strings.TrimSpace(*req.PartnerCode)
		if partnerCode == "" {
			return nil, fmt.Errorf("partner code cannot be empty")
		}
		updated.PartnerCode = partnerCode
	}

	if req.AttributeID != nil {
		// Validate new attribute exists
		attribute, err := s.attributeRepo.GetByID(ctx, req.AttributeID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to validate attribute: %w", err)
		}

		if attribute == nil {
			return nil, fmt.Errorf("attribute not found")
		}

		updated.AttributeID = *req.AttributeID
		updated.AttributeCode = attribute.Code
	}

	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}

	// Update in repository
	err = s.repo.Update(ctx, id, updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update partner attribute mapping: %w", err)
	}

	return PartnerAttributeMapToResponse(updated), nil
}

// Delete soft deletes a partner attribute mapping
func (s *partnerAttributeMapService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("partner attribute mapping ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid partner attribute mapping ID format: %w", err)
	}

	// Verify mapping exists (including inactive ones)
	existing, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get mapping: %w", err)
	}

	if existing == nil {
		return fmt.Errorf("partner attribute mapping not found")
	}

	// Delete from repository
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete partner attribute mapping: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted partner attribute mapping
func (s *partnerAttributeMapService) Restore(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("partner attribute mapping ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid partner attribute mapping ID format: %w", err)
	}

	// Restore from repository
	err := s.repo.Restore(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore partner attribute mapping: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes a partner attribute mapping
func (s *partnerAttributeMapService) ForceDelete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("partner attribute mapping ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid partner attribute mapping ID format: %w", err)
	}

	// Force delete from repository
	err := s.repo.ForceDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to force delete partner attribute mapping: %w", err)
	}

	return nil
}

// BulkCreate creates multiple partner attribute mappings
func (s *partnerAttributeMapService) BulkCreate(ctx context.Context, req *dtos.BulkCreatePartnerAttributeMapRequest) (*dtos.BulkCreatePartnerAttributeMapResponse, error) {
	if req == nil || len(req.Mappings) == 0 {
		return nil, fmt.Errorf("bulk create request cannot be empty")
	}

	if len(req.Mappings) > 100 {
		return nil, fmt.Errorf("bulk create limited to 100 mappings at a time")
	}

	var created []dtos.PartnerAttributeMapResponse
	var failed []dtos.BulkCreatePartnerAttributeError

	// Process each mapping
	for i, mapping := range req.Mappings {
		result, err := s.Create(ctx, &mapping)
		if err != nil {
			failed = append(failed, dtos.BulkCreatePartnerAttributeError{
				Index:   i,
				Error:   err.Error(),
				Request: mapping,
			})
		} else {
			created = append(created, *result)
		}
	}

	return &dtos.BulkCreatePartnerAttributeMapResponse{
		Created: created,
		Failed:  failed,
	}, nil
}

// BulkDelete deletes multiple partner attribute mappings
func (s *partnerAttributeMapService) BulkDelete(ctx context.Context, req *dtos.BulkDeletePartnerAttributeMapRequest) (*dtos.BulkDeletePartnerAttributeMapResponse, error) {
	if req == nil || len(req.IDs) == 0 {
		return nil, fmt.Errorf("bulk delete request cannot be empty")
	}

	if len(req.IDs) > 100 {
		return nil, fmt.Errorf("bulk delete limited to 100 mappings at a time")
	}

	var deleted []uuid.UUID
	var failed []dtos.BulkDeletePartnerAttributeError

	// Process each ID
	for _, id := range req.IDs {
		err := s.Delete(ctx, id.String())
		if err != nil {
			failed = append(failed, dtos.BulkDeletePartnerAttributeError{
				ID:    id,
				Error: err.Error(),
			})
		} else {
			deleted = append(deleted, id)
		}
	}

	return &dtos.BulkDeletePartnerAttributeMapResponse{
		Deleted: deleted,
		Failed:  failed,
	}, nil
}

// GetStats retrieves statistics for a partner's attribute mappings
func (s *partnerAttributeMapService) GetStats(ctx context.Context, partnerCode string) (*dtos.PartnerAttributeStatsResponse, error) {
	if strings.TrimSpace(partnerCode) == "" {
		return nil, fmt.Errorf("partner code cannot be empty")
	}

	// Normalize partner code
	partnerCode = strings.TrimSpace(partnerCode)

	// Get stats from repository
	stats, err := s.repo.GetStats(ctx, partnerCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner attribute stats: %w", err)
	}

	// Convert timestamps
	var firstMappingAt *time.Time
	var lastMappingAt *time.Time
	if stats.FirstMappingAt != nil {
		if t, err := time.Parse(time.RFC3339, *stats.FirstMappingAt); err == nil {
			firstMappingAt = &t
		}
	}
	if stats.LastMappingAt != nil {
		if t, err := time.Parse(time.RFC3339, *stats.LastMappingAt); err == nil {
			lastMappingAt = &t
		}
	}

	// Convert to response DTO
	return &dtos.PartnerAttributeStatsResponse{
		PartnerCode:       stats.PartnerCode,
		TotalMappings:     int(stats.TotalMappings),
		ActiveMappings:    int(stats.ActiveMappings),
		UniqueCategories:  int(stats.UniqueCategories),
		FirstMappingAt:    firstMappingAt,
		LastMappingAt:     lastMappingAt,
		MostUsedCategory:  stats.MostUsedCategory,
		MostUsedAttribute: stats.MostUsedAttribute,
	}, nil
}
