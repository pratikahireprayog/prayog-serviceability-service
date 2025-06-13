package services

import (
	"context"
	"fmt"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"
	"strings"

	"github.com/google/uuid"
)

// postalCodeService implements the PostalCodeService interface
type postalCodeService struct {
	repo        repositories.PostalCodeRepository
	countryRepo repositories.CountryRepository
	regionRepo  repositories.RegionRepository
	cityRepo    repositories.CityRepository
	areaRepo    repositories.AreaRepository
}

// NewPostalCodeService creates a new postal code service instance
func NewPostalCodeService(
	repo repositories.PostalCodeRepository,
	countryRepo repositories.CountryRepository,
	regionRepo repositories.RegionRepository,
	cityRepo repositories.CityRepository,
	areaRepo repositories.AreaRepository,
) PostalCodeService {
	return &postalCodeService{
		repo:        repo,
		countryRepo: countryRepo,
		regionRepo:  regionRepo,
		cityRepo:    cityRepo,
		areaRepo:    areaRepo,
	}
}

// GetByID retrieves a postal code by its ID
func (s *postalCodeService) GetByID(ctx context.Context, id string) (*dtos.PostalCodeResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("postal code ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid postal code ID format: %w", err)
	}

	postalCode, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get postal code by ID: %w", err)
	}

	if postalCode == nil {
		return nil, fmt.Errorf("postal code not found")
	}

	return PostalCodeToResponse(postalCode), nil
}

// GetByCode retrieves a postal code by its code
func (s *postalCodeService) GetByCode(ctx context.Context, code string) (*dtos.PostalCodeResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("postal code cannot be empty")
	}

	// Normalize code
	code = strings.TrimSpace(code)

	postalCode, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get postal code by code: %w", err)
	}

	if postalCode == nil {
		return nil, fmt.Errorf("postal code not found")
	}

	return PostalCodeToResponse(postalCode), nil
}

// GetAll retrieves all postal codes with pagination
func (s *postalCodeService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.PostalCodeListResponse, error) {
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

	postalCodes, total, err := s.repo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get postal codes: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.PostalCodeResponse, len(postalCodes))
	for i, postalCode := range postalCodes {
		if response := PostalCodeToResponse(&postalCode); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.PostalCodeListResponse{
		Success: true,
		Message: "Postal codes retrieved successfully",
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

// GetAllWithDeleted retrieves all postal codes including soft-deleted ones (admin operation)
func (s *postalCodeService) GetAllWithDeleted(ctx context.Context, req *dtos.PaginationRequest) (*dtos.PostalCodeListResponse, error) {
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

	postalCodes, total, err := s.repo.GetAllWithDeleted(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get postal codes with deleted: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.PostalCodeResponse, len(postalCodes))
	for i, postalCode := range postalCodes {
		if response := PostalCodeToResponse(&postalCode); response != nil {
			responses[i] = *response
		}
	}

	// Calculate pagination metadata
	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.PostalCodeListResponse{
		Success: true,
		Message: "Postal codes retrieved successfully (including deleted)",
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

// GetByLocation retrieves postal codes by location hierarchy
func (s *postalCodeService) GetByLocation(ctx context.Context, countryCode, regionCode, cityCode, areaCode string) ([]dtos.PostalCodeResponse, error) {
	// At least country code is required
	if strings.TrimSpace(countryCode) == "" {
		return nil, fmt.Errorf("country code is required")
	}

	postalCodes, err := s.repo.GetByLocation(ctx, countryCode, regionCode, cityCode, areaCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get postal codes by location: %w", err)
	}

	// Convert models to response DTOs
	responses := make([]dtos.PostalCodeResponse, len(postalCodes))
	for i, postalCode := range postalCodes {
		if response := PostalCodeToResponse(&postalCode); response != nil {
			responses[i] = *response
		}
	}

	return responses, nil
}

// ValidateLocationScope validates the location scope for a postal code request
func (s *postalCodeService) ValidateLocationScope(ctx context.Context, req *dtos.CreatePostalCodeRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Location scope validation is now optional - only validate if provided
	if req.LocationScope != nil {
		validScopes := []string{"country", "region", "city", "area"}
		scopeValid := false
		for _, validScope := range validScopes {
			if *req.LocationScope == validScope {
				scopeValid = true
				break
			}
		}
		if !scopeValid {
			return fmt.Errorf("invalid location scope: %s, must be one of: %v", *req.LocationScope, validScopes)
		}
	}

	// Removed entity existence validation - entities can be created independently
	// This allows postal codes to be created without requiring pre-existing country/region/city/area records

	return nil
}

// Create creates a new postal code
func (s *postalCodeService) Create(ctx context.Context, req *dtos.CreatePostalCodeRequest) (*dtos.PostalCodeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate required fields
	if strings.TrimSpace(req.Code) == "" {
		return nil, fmt.Errorf("postal code is required")
	}

	// Validate location scope
	if err := s.ValidateLocationScope(ctx, req); err != nil {
		return nil, fmt.Errorf("location scope validation failed: %w", err)
	}

	// Normalize code
	req.Code = strings.TrimSpace(req.Code)

	// Check if postal code with code already exists
	existing, err := s.repo.GetByCode(ctx, req.Code)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check existing postal code: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("postal code with code '%s' already exists", req.Code)
	}

	// Create postal code model
	postalCode := &models.PostalCode{
		ID:       uuid.New(),
		Code:     req.Code,
		IsActive: true, // Default to active
	}

	// Set optional fields
	if req.CountryID != nil {
		postalCode.CountryID = req.CountryID
	}
	if req.CountryCode != nil {
		postalCode.CountryCode = req.CountryCode
	}
	if req.RegionID != nil {
		postalCode.RegionID = req.RegionID
	}
	if req.RegionCode != nil {
		postalCode.RegionCode = req.RegionCode
	}
	if req.CityID != nil {
		postalCode.CityID = req.CityID
	}
	if req.CityCode != nil {
		postalCode.CityCode = req.CityCode
	}
	if req.AreaID != nil {
		postalCode.AreaID = req.AreaID
	}
	if req.AreaCode != nil {
		postalCode.AreaCode = req.AreaCode
	}
	if req.LocationScope != nil {
		postalCode.LocationScope = req.LocationScope
	}
	if req.IsActive != nil {
		postalCode.IsActive = *req.IsActive
	}

	err = s.repo.Create(ctx, postalCode)
	if err != nil {
		return nil, fmt.Errorf("failed to create postal code: %w", err)
	}

	return PostalCodeToResponse(postalCode), nil
}

// Update updates an existing postal code
func (s *postalCodeService) Update(ctx context.Context, id string, req *dtos.UpdatePostalCodeRequest) (*dtos.PostalCodeResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("postal code ID cannot be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid postal code ID format: %w", err)
	}

	// Check if postal code exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get postal code: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("postal code not found")
	}

	// Create validation request for location scope validation
	validationReq := &dtos.CreatePostalCodeRequest{
		Code:          existing.Code,
		CountryID:     req.CountryID,
		CountryCode:   req.CountryCode,
		RegionID:      req.RegionID,
		RegionCode:    req.RegionCode,
		CityID:        req.CityID,
		CityCode:      req.CityCode,
		AreaID:        req.AreaID,
		AreaCode:      req.AreaCode,
		LocationScope: req.LocationScope,
	}

	// Use existing values if not provided in update
	if validationReq.CountryID == nil {
		validationReq.CountryID = existing.CountryID
	}
	if validationReq.CountryCode == nil {
		validationReq.CountryCode = existing.CountryCode
	}
	if validationReq.RegionID == nil {
		validationReq.RegionID = existing.RegionID
	}
	if validationReq.RegionCode == nil {
		validationReq.RegionCode = existing.RegionCode
	}
	if validationReq.CityID == nil {
		validationReq.CityID = existing.CityID
	}
	if validationReq.CityCode == nil {
		validationReq.CityCode = existing.CityCode
	}
	if validationReq.AreaID == nil {
		validationReq.AreaID = existing.AreaID
	}
	if validationReq.AreaCode == nil {
		validationReq.AreaCode = existing.AreaCode
	}
	if validationReq.LocationScope == nil {
		validationReq.LocationScope = existing.LocationScope
	}

	// Validate location scope
	if err := s.ValidateLocationScope(ctx, validationReq); err != nil {
		return nil, fmt.Errorf("location scope validation failed: %w", err)
	}

	// Update fields
	if req.CountryID != nil {
		existing.CountryID = req.CountryID
	}
	if req.CountryCode != nil {
		existing.CountryCode = req.CountryCode
	}
	if req.RegionID != nil {
		existing.RegionID = req.RegionID
	}
	if req.RegionCode != nil {
		existing.RegionCode = req.RegionCode
	}
	if req.CityID != nil {
		existing.CityID = req.CityID
	}
	if req.CityCode != nil {
		existing.CityCode = req.CityCode
	}
	if req.AreaID != nil {
		existing.AreaID = req.AreaID
	}
	if req.AreaCode != nil {
		existing.AreaCode = req.AreaCode
	}
	if req.LocationScope != nil {
		existing.LocationScope = req.LocationScope
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	err = s.repo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update postal code: %w", err)
	}

	return PostalCodeToResponse(existing), nil
}

// Delete deletes a postal code
func (s *postalCodeService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("postal code ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid postal code ID format: %w", err)
	}

	// Check if postal code exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get postal code: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("postal code not found")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete postal code: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted postal code (admin operation)
func (s *postalCodeService) Restore(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("postal code ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid postal code ID format: %w", err)
	}

	// Check if postal code exists (including soft-deleted)
	existing, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("postal code not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("postal code not found")
	}

	// Check for unique constraint conflicts before restoration
	if existing.Code != "" {
		activePostalCode, err := s.repo.GetByCode(ctx, existing.Code)
		if err == nil && activePostalCode != nil {
			return fmt.Errorf("cannot restore postal code: another active postal code with code '%s' already exists", existing.Code)
		}
	}

	err = s.repo.Restore(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore postal code: %w", err)
	}

	return nil
}

// ForceDelete permanently deletes a postal code (admin operation)
func (s *postalCodeService) ForceDelete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("postal code ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid postal code ID format: %w", err)
	}

	// Check if postal code exists (including soft-deleted)
	existing, err := s.repo.GetByIDWithDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("postal code not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("postal code not found")
	}

	err = s.repo.ForceDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete postal code: %w", err)
	}

	return nil
}
