package services

import (
	"context"
	"fmt"
	servicesv1 "prayog-serviceability-service/internal/services/v1"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"
	"strings"

	"github.com/google/uuid"
)

// PartnerLocationCoverageService defines the interface for partner location coverage operations
type PartnerLocationCoverageService interface {
	GetByPartnerID(ctx context.Context, partnerID uuid.UUID, filters *dtos.PartnerLocationCoverageFiltersRequest) (*dtos.PartnerLocationCoverageListResponse, error)
	GetByID(ctx context.Context, partnerID uuid.UUID, coverageID uuid.UUID) (*dtos.PartnerLocationCoverageResponse, error)
	Create(ctx context.Context, partnerID uuid.UUID, req *dtos.CreatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error)
	Update(ctx context.Context, partnerID uuid.UUID, coverageID uuid.UUID, req *dtos.UpdatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error)
	Delete(ctx context.Context, partnerID uuid.UUID, coverageID uuid.UUID) error
	BulkCreate(ctx context.Context, partnerID uuid.UUID, req *dtos.BulkCreatePartnerLocationCoverageRequest) (*dtos.BulkCreatePartnerLocationCoverageResponse, error)
}

// partnerLocationCoverageService implements the PartnerLocationCoverageService interface
type partnerLocationCoverageService struct {
	repo                 repositories.PartnerLocationCoverageRepository
	locationRepo         repositories.LocationRepository
	partnerValidationSvc servicesv1.PartnerValidationService
}

// NewPartnerLocationCoverageService creates a new partner location coverage service instance
func NewPartnerLocationCoverageService(
	repo repositories.PartnerLocationCoverageRepository,
	locationRepo repositories.LocationRepository,
	partnerValidationSvc servicesv1.PartnerValidationService,
) PartnerLocationCoverageService {
	return &partnerLocationCoverageService{
		repo:                 repo,
		locationRepo:         locationRepo,
		partnerValidationSvc: partnerValidationSvc,
	}
}

// GetByPartnerID retrieves all location coverages for a specific partner with optional filters
func (s *partnerLocationCoverageService) GetByPartnerID(ctx context.Context, partnerID uuid.UUID, filters *dtos.PartnerLocationCoverageFiltersRequest) (*dtos.PartnerLocationCoverageListResponse, error) {
	// Validate partner exists
	if err := s.partnerValidationSvc.ValidatePartner(ctx, partnerID.String()); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Set default pagination if not provided
	if filters == nil {
		filters = &dtos.PartnerLocationCoverageFiltersRequest{
			Limit:  intPtr(10),
			Offset: intPtr(0),
		}
	}
	if filters.Limit == nil || *filters.Limit < 1 || *filters.Limit > 100 {
		filters.Limit = intPtr(10)
	}
	if filters.Offset == nil || *filters.Offset < 0 {
		filters.Offset = intPtr(0)
	}

	// Build repository filters
	repoFilters := &models.PartnerLocationCoverageFilters{
		PartnerID: &partnerID,
	}
	if filters.LocationScope != nil {
		repoFilters.LocationScope = strings.ToUpper(*filters.LocationScope)
	}
	if filters.LocationID != nil {
		repoFilters.LocationID = filters.LocationID
	}
	if filters.ZoneType != nil {
		repoFilters.ZoneType = strings.ToUpper(*filters.ZoneType)
	}
	if filters.SourcePostalCode != nil {
		repoFilters.SourcePostalCode = filters.SourcePostalCode
	}
	if filters.DestinationPostalCode != nil {
		repoFilters.DestinationPostalCode = filters.DestinationPostalCode
	}
	if filters.SourcePostalCodeID != nil {
		repoFilters.SourcePostalCodeID = filters.SourcePostalCodeID
	}
	if filters.DestinationPostalCodeID != nil {
		repoFilters.DestinationPostalCodeID = filters.DestinationPostalCodeID
	}
	if filters.IsActive != nil {
		repoFilters.IsActive = filters.IsActive
	}

	// Get coverages from repository
	// Note: Using simplified approach, would need to extend repository for proper pagination with filters
	allCoverages, err := s.repo.GetByPartnerID(ctx, convertUUIDToUint(partnerID))
	if err != nil {
		return nil, fmt.Errorf("failed to get partner location coverages: %w", err)
	}

	// Apply client-side filtering and pagination (for now)
	filteredCoverages := s.filterCoverages(allCoverages, repoFilters)

	// Apply pagination
	total := int64(len(filteredCoverages))
	start := *filters.Offset
	end := start + *filters.Limit
	if start > len(filteredCoverages) {
		start = len(filteredCoverages)
	}
	if end > len(filteredCoverages) {
		end = len(filteredCoverages)
	}

	paginatedCoverages := filteredCoverages[start:end]

	// Convert to response DTOs
	responses := make([]dtos.PartnerLocationCoverageResponse, len(paginatedCoverages))
	for i, coverage := range paginatedCoverages {
		responses[i] = *PartnerLocationCoverageToResponse(&coverage)
	}

	// Calculate pagination metadata
	hasNext := int64(*filters.Offset+*filters.Limit) < total
	hasPrevious := *filters.Offset > 0

	return &dtos.PartnerLocationCoverageListResponse{
		Data: responses,
		Pagination: dtos.PaginationResponse{
			Offset:      *filters.Offset,
			Limit:       *filters.Limit,
			Total:       total,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	}, nil
}

// GetByID retrieves a specific location coverage for a partner
func (s *partnerLocationCoverageService) GetByID(ctx context.Context, partnerID uuid.UUID, coverageID uuid.UUID) (*dtos.PartnerLocationCoverageResponse, error) {
	// Validate partner exists
	if err := s.partnerValidationSvc.ValidatePartner(ctx, partnerID.String()); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Get all coverages for the partner (simplified approach)
	coverages, err := s.repo.GetByPartnerID(ctx, convertUUIDToUint(partnerID))
	if err != nil {
		return nil, fmt.Errorf("failed to get partner location coverages: %w", err)
	}

	// Find the specific coverage
	for _, coverage := range coverages {
		if coverage.ID == coverageID {
			return PartnerLocationCoverageToResponse(&coverage), nil
		}
	}

	return nil, fmt.Errorf("partner location coverage not found")
}

// Create creates a new partner location coverage with validation
func (s *partnerLocationCoverageService) Create(ctx context.Context, partnerID uuid.UUID, req *dtos.CreatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate partner exists
	if err := s.partnerValidationSvc.ValidatePartner(ctx, partnerID.String()); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Validate location scope and location_id/location_code
	if err := s.validateLocationScopeAndCode(ctx, req.LocationScope, req.LocationID, req.LocationCode); err != nil {
		return nil, fmt.Errorf("location validation failed: %w", err)
	}

	// Validate postal code fields consistency
	if err := s.validatePostalCodeFields(ctx, req.SourcePostalCode, req.DestinationPostalCode, req.SourcePostalCodeID, req.DestinationPostalCodeID); err != nil {
		return nil, fmt.Errorf("postal code validation failed: %w", err)
	}

	// Create coverage model
	coverage := &models.PartnerLocationCoverage{
		ID:        uuid.New(),
		PartnerID: &partnerID,
		IsActive:  true, // Default to active
	}

	// Set optional fields
	if req.LocationScope != nil {
		scope := strings.ToUpper(*req.LocationScope)
		coverage.LocationScope = &scope
	}
	if req.LocationID != nil {
		coverage.LocationID = req.LocationID
	}
	if req.LocationCode != nil {
		code := strings.TrimSpace(*req.LocationCode)
		if code != "" {
			coverage.LocationCode = &code
		}
	}
	if req.ZoneType != nil {
		zoneType := strings.ToUpper(*req.ZoneType)
		coverage.ZoneType = &zoneType
	}
	if req.SourcePostalCode != nil {
		code := strings.TrimSpace(*req.SourcePostalCode)
		if code != "" {
			coverage.SourcePostalCode = &code
		}
	}
	if req.DestinationPostalCode != nil {
		code := strings.TrimSpace(*req.DestinationPostalCode)
		if code != "" {
			coverage.DestinationPostalCode = &code
		}
	}
	if req.SourcePostalCodeID != nil {
		coverage.SourcePostalCodeID = req.SourcePostalCodeID
	}
	if req.DestinationPostalCodeID != nil {
		coverage.DestinationPostalCodeID = req.DestinationPostalCodeID
	}
	if req.IsActive != nil {
		coverage.IsActive = *req.IsActive
	}

	// Create in repository
	err := s.repo.Create(ctx, coverage)
	if err != nil {
		return nil, fmt.Errorf("failed to create partner location coverage: %w", err)
	}

	return PartnerLocationCoverageToResponse(coverage), nil
}

// Update updates an existing partner location coverage
func (s *partnerLocationCoverageService) Update(ctx context.Context, partnerID uuid.UUID, coverageID uuid.UUID, req *dtos.UpdatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Validate partner exists
	if err := s.partnerValidationSvc.ValidatePartner(ctx, partnerID.String()); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Get existing coverage
	existing, err := s.GetByID(ctx, partnerID, coverageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing coverage: %w", err)
	}

	// Validate location scope and location_id/location_code if being updated
	if req.LocationScope != nil || req.LocationID != nil || req.LocationCode != nil {
		var locationScope *string
		var locationID *uuid.UUID
		var locationCode *string

		if req.LocationScope != nil {
			locationScope = req.LocationScope
		} else {
			locationScope = existing.LocationScope
		}
		if req.LocationID != nil {
			locationID = req.LocationID
		} else {
			locationID = existing.LocationID
		}
		if req.LocationCode != nil {
			locationCode = req.LocationCode
		} else {
			locationCode = existing.LocationCode
		}

		if err := s.validateLocationScopeAndCode(ctx, locationScope, locationID, locationCode); err != nil {
			return nil, fmt.Errorf("location validation failed: %w", err)
		}
	}

	// Validate postal code fields if being updated
	if req.SourcePostalCode != nil || req.DestinationPostalCode != nil || req.SourcePostalCodeID != nil || req.DestinationPostalCodeID != nil {
		var sourcePostalCode *string
		var destinationPostalCode *string
		var sourcePostalCodeID *uuid.UUID
		var destinationPostalCodeID *uuid.UUID

		if req.SourcePostalCode != nil {
			sourcePostalCode = req.SourcePostalCode
		} else {
			sourcePostalCode = existing.SourcePostalCode
		}
		if req.DestinationPostalCode != nil {
			destinationPostalCode = req.DestinationPostalCode
		} else {
			destinationPostalCode = existing.DestinationPostalCode
		}
		if req.SourcePostalCodeID != nil {
			sourcePostalCodeID = req.SourcePostalCodeID
		} else {
			sourcePostalCodeID = existing.SourcePostalCodeID
		}
		if req.DestinationPostalCodeID != nil {
			destinationPostalCodeID = req.DestinationPostalCodeID
		} else {
			destinationPostalCodeID = existing.DestinationPostalCodeID
		}

		if err := s.validatePostalCodeFields(ctx, sourcePostalCode, destinationPostalCode, sourcePostalCodeID, destinationPostalCodeID); err != nil {
			return nil, fmt.Errorf("postal code validation failed: %w", err)
		}
	}

	// Convert back to model for update
	coverage := &models.PartnerLocationCoverage{
		ID:                      coverageID,
		PartnerID:               &partnerID,
		LocationScope:           existing.LocationScope,
		LocationID:              existing.LocationID,
		LocationCode:            existing.LocationCode,
		ZoneType:                existing.ZoneType,
		SourcePostalCode:        existing.SourcePostalCode,
		DestinationPostalCode:   existing.DestinationPostalCode,
		SourcePostalCodeID:      existing.SourcePostalCodeID,
		DestinationPostalCodeID: existing.DestinationPostalCodeID,
		IsActive:                existing.IsActive,
		CreatedAt:               existing.CreatedAt,
		UpdatedAt:               existing.UpdatedAt,
	}

	// Update fields
	if req.LocationScope != nil {
		scope := strings.ToUpper(*req.LocationScope)
		coverage.LocationScope = &scope
	}
	if req.LocationID != nil {
		coverage.LocationID = req.LocationID
	}
	if req.LocationCode != nil {
		code := strings.TrimSpace(*req.LocationCode)
		if code != "" {
			coverage.LocationCode = &code
		} else {
			coverage.LocationCode = nil
		}
	}
	if req.ZoneType != nil {
		zoneType := strings.ToUpper(*req.ZoneType)
		coverage.ZoneType = &zoneType
	}
	if req.SourcePostalCode != nil {
		code := strings.TrimSpace(*req.SourcePostalCode)
		if code != "" {
			coverage.SourcePostalCode = &code
		} else {
			coverage.SourcePostalCode = nil
		}
	}
	if req.DestinationPostalCode != nil {
		code := strings.TrimSpace(*req.DestinationPostalCode)
		if code != "" {
			coverage.DestinationPostalCode = &code
		} else {
			coverage.DestinationPostalCode = nil
		}
	}
	if req.SourcePostalCodeID != nil {
		coverage.SourcePostalCodeID = req.SourcePostalCodeID
	}
	if req.DestinationPostalCodeID != nil {
		coverage.DestinationPostalCodeID = req.DestinationPostalCodeID
	}
	if req.IsActive != nil {
		coverage.IsActive = *req.IsActive
	}

	// Update in repository
	err = s.repo.Update(ctx, coverage)
	if err != nil {
		return nil, fmt.Errorf("failed to update partner location coverage: %w", err)
	}

	return PartnerLocationCoverageToResponse(coverage), nil
}

// Delete deletes a partner location coverage
func (s *partnerLocationCoverageService) Delete(ctx context.Context, partnerID uuid.UUID, coverageID uuid.UUID) error {
	// Validate partner exists
	if err := s.partnerValidationSvc.ValidatePartner(ctx, partnerID.String()); err != nil {
		return fmt.Errorf("partner validation failed: %w", err)
	}

	// Verify the coverage exists and belongs to the partner
	_, err := s.GetByID(ctx, partnerID, coverageID)
	if err != nil {
		return fmt.Errorf("coverage not found or doesn't belong to partner: %w", err)
	}

	// Delete from repository
	err = s.repo.Delete(ctx, convertUUIDToUint(coverageID))
	if err != nil {
		return fmt.Errorf("failed to delete partner location coverage: %w", err)
	}

	return nil
}

// BulkCreate creates multiple partner location coverages in a transaction
func (s *partnerLocationCoverageService) BulkCreate(ctx context.Context, partnerID uuid.UUID, req *dtos.BulkCreatePartnerLocationCoverageRequest) (*dtos.BulkCreatePartnerLocationCoverageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("bulk create request cannot be nil")
	}

	if len(req.Coverages) == 0 {
		return nil, fmt.Errorf("at least one coverage is required")
	}

	if len(req.Coverages) > 100 {
		return nil, fmt.Errorf("maximum 100 coverages allowed per bulk operation")
	}

	// Validate partner exists
	if err := s.partnerValidationSvc.ValidatePartner(ctx, partnerID.String()); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	var created []dtos.PartnerLocationCoverageResponse
	var failed []dtos.BulkCreateError

	// Process each coverage request
	for i, coverageReq := range req.Coverages {
		result, err := s.Create(ctx, partnerID, &coverageReq)
		if err != nil {
			failed = append(failed, dtos.BulkCreateError{
				Index:   i,
				Error:   err.Error(),
				Request: coverageReq,
			})
		} else {
			created = append(created, *result)
		}
	}

	return &dtos.BulkCreatePartnerLocationCoverageResponse{
		Created: created,
		Failed:  failed,
	}, nil
}

// Helper methods

// validatePostalCodeFields validates postal code fields consistency
func (s *partnerLocationCoverageService) validatePostalCodeFields(ctx context.Context, sourcePostalCode, destinationPostalCode *string, sourcePostalCodeID, destinationPostalCodeID *uuid.UUID) error {
	// If postal codes are provided, both source and destination should be provided for route coverage
	if (sourcePostalCode != nil && *sourcePostalCode != "") || (destinationPostalCode != nil && *destinationPostalCode != "") {
		if sourcePostalCode == nil || *sourcePostalCode == "" {
			return fmt.Errorf("destination postal code provided but source postal code is missing - both are required for route coverage")
		}
		if destinationPostalCode == nil || *destinationPostalCode == "" {
			return fmt.Errorf("source postal code provided but destination postal code is missing - both are required for route coverage")
		}
	}

	// If postal code IDs are provided, both source and destination should be provided for route coverage
	if sourcePostalCodeID != nil || destinationPostalCodeID != nil {
		if sourcePostalCodeID == nil {
			return fmt.Errorf("destination postal code ID provided but source postal code ID is missing - both are required for route coverage")
		}
		if destinationPostalCodeID == nil {
			return fmt.Errorf("source postal code ID provided but destination postal code ID is missing - both are required for route coverage")
		}
	}

	// Ensure consistency between postal codes and postal code IDs
	if (sourcePostalCode != nil && *sourcePostalCode != "") && sourcePostalCodeID == nil {
		return fmt.Errorf("source postal code provided but source postal code ID is missing - both should be provided for consistency")
	}
	if (destinationPostalCode != nil && *destinationPostalCode != "") && destinationPostalCodeID == nil {
		return fmt.Errorf("destination postal code provided but destination postal code ID is missing - both should be provided for consistency")
	}
	if sourcePostalCodeID != nil && (sourcePostalCode == nil || *sourcePostalCode == "") {
		return fmt.Errorf("source postal code ID provided but source postal code is missing - both should be provided for consistency")
	}
	if destinationPostalCodeID != nil && (destinationPostalCode == nil || *destinationPostalCode == "") {
		return fmt.Errorf("destination postal code ID provided but destination postal code is missing - both should be provided for consistency")
	}

	return nil
}

// validateLocationScopeAndCode validates location scope and ensures location_id or location_code is provided
func (s *partnerLocationCoverageService) validateLocationScopeAndCode(ctx context.Context, locationScope *string, locationID *uuid.UUID, locationCode *string) error {
	if locationScope == nil {
		return fmt.Errorf("location_scope is required")
	}

	scope := strings.ToUpper(*locationScope)
	validScopes := []string{"POSTAL_CODE", "AREA", "CITY", "REGION", "COUNTRY"}
	if !contains(validScopes, scope) {
		return fmt.Errorf("invalid location_scope: %s, must be one of: %v", scope, validScopes)
	}

	// Either location_id or location_code must be provided
	if locationID == nil && (locationCode == nil || strings.TrimSpace(*locationCode) == "") {
		return fmt.Errorf("either location_id or location_code must be provided")
	}

	// TODO: Add actual location validation by calling appropriate repository methods
	// This would involve checking if the location exists in the respective tables

	return nil
}

// filterCoverages applies filters to coverage slice (client-side filtering)
func (s *partnerLocationCoverageService) filterCoverages(coverages []models.PartnerLocationCoverage, filters *models.PartnerLocationCoverageFilters) []models.PartnerLocationCoverage {
	if filters == nil {
		return coverages
	}

	var filtered []models.PartnerLocationCoverage
	for _, coverage := range coverages {
		if filters.LocationScope != "" && (coverage.LocationScope == nil || *coverage.LocationScope != filters.LocationScope) {
			continue
		}
		if filters.LocationID != nil && (coverage.LocationID == nil || *coverage.LocationID != *filters.LocationID) {
			continue
		}
		if filters.ZoneType != "" && (coverage.ZoneType == nil || *coverage.ZoneType != filters.ZoneType) {
			continue
		}
		if filters.IsActive != nil && coverage.IsActive != *filters.IsActive {
			continue
		}
		filtered = append(filtered, coverage)
	}
	return filtered
}

// PartnerLocationCoverageToResponse converts a model to response DTO
func PartnerLocationCoverageToResponse(coverage *models.PartnerLocationCoverage) *dtos.PartnerLocationCoverageResponse {
	if coverage == nil {
		return nil
	}

	return &dtos.PartnerLocationCoverageResponse{
		ID:                      coverage.ID,
		PartnerID:               coverage.PartnerID,
		PartnerCode:             coverage.PartnerCode,
		LocationScope:           coverage.LocationScope,
		LocationID:              coverage.LocationID,
		LocationCode:            coverage.LocationCode,
		ZoneType:                coverage.ZoneType,
		SourcePostalCode:        coverage.SourcePostalCode,
		DestinationPostalCode:   coverage.DestinationPostalCode,
		SourcePostalCodeID:      coverage.SourcePostalCodeID,
		DestinationPostalCodeID: coverage.DestinationPostalCodeID,
		IsActive:                coverage.IsActive,
		CreatedAt:               coverage.CreatedAt,
		UpdatedAt:               coverage.UpdatedAt,
	}
}

// Helper utility functions
func intPtr(i int) *int {
	return &i
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// convertUUIDToUint is a temporary helper to convert UUID to uint for repository compatibility
// TODO: Update repository interface to use UUID instead of uint
func convertUUIDToUint(id uuid.UUID) uint {
	// This is a simplified conversion - in production, you'd want a proper mapping
	// For now, we'll use a hash of the UUID
	bytes := id[12:16] // Use last 4 bytes
	return uint(bytes[0])<<24 + uint(bytes[1])<<16 + uint(bytes[2])<<8 + uint(bytes[3])
}
