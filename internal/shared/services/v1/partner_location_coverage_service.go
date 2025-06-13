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
	// Legacy methods (using coverage ID)
	GetByPartnerID(ctx context.Context, partnerID uuid.UUID, filters *dtos.PartnerLocationCoverageFiltersRequest) (*dtos.PartnerLocationCoverageListResponse, error)
	GetByID(ctx context.Context, partnerID uuid.UUID, coverageID uuid.UUID) (*dtos.PartnerLocationCoverageResponse, error)
	Create(ctx context.Context, partnerID uuid.UUID, req *dtos.CreatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error)
	Update(ctx context.Context, partnerID uuid.UUID, coverageID uuid.UUID, req *dtos.UpdatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error)
	Delete(ctx context.Context, partnerID uuid.UUID, coverageID uuid.UUID) error
	BulkCreate(ctx context.Context, partnerID uuid.UUID, req *dtos.BulkCreatePartnerLocationCoverageRequest) (*dtos.BulkCreatePartnerLocationCoverageResponse, error)

	// New postal code based methods (using partner ID)
	GetByPartnerIDAndPostalCode(ctx context.Context, partnerID uuid.UUID, postalCode string) (*dtos.PartnerLocationCoverageResponse, error)
	UpdateByPartnerIDAndPostalCode(ctx context.Context, partnerID uuid.UUID, postalCode string, req *dtos.UpdatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error)
	DeleteByPartnerIDAndPostalCode(ctx context.Context, partnerID uuid.UUID, postalCode string) error
	CheckCoverage(ctx context.Context, partnerID uuid.UUID, postalCode string, filters *dtos.PartnerLocationCoverageFiltersRequest) (*dtos.CoverageCheckResponse, error)

	// New partner code based methods
	GetByPartnerCode(ctx context.Context, partnerCode string, filters *dtos.PartnerLocationCoverageFiltersRequest) (*dtos.PartnerLocationCoverageListResponse, error)
	CreateByPartnerCode(ctx context.Context, partnerCode string, req *dtos.CreatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error)
	BulkCreateByPartnerCode(ctx context.Context, partnerCode string, req *dtos.BulkCreatePartnerLocationCoverageRequest) (*dtos.BulkCreatePartnerLocationCoverageResponse, error)
	GetByPartnerCodeAndPostalCode(ctx context.Context, partnerCode string, postalCode string) (*dtos.PartnerLocationCoverageResponse, error)
	UpdateByPartnerCodeAndPostalCode(ctx context.Context, partnerCode string, postalCode string, req *dtos.UpdatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error)
	DeleteByPartnerCodeAndPostalCode(ctx context.Context, partnerCode string, postalCode string) error
	CheckCoverageByPartnerCode(ctx context.Context, partnerCode string, postalCode string, filters *dtos.PartnerLocationCoverageFiltersRequest) (*dtos.CoverageCheckResponse, error)
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
	if filters.PostalCode != nil {
		repoFilters.PostalCode = filters.PostalCode
	}
	if filters.PostalCodeID != nil {
		repoFilters.PostalCodeID = filters.PostalCodeID
	}
	if filters.ZoneType != nil {
		repoFilters.ZoneType = strings.ToUpper(*filters.ZoneType)
	}

	// Add serviceability filters
	if filters.CountryCode != nil {
		repoFilters.CountryCode = filters.CountryCode
	}
	if filters.ProductType != nil {
		repoFilters.ProductType = filters.ProductType
	}
	if filters.ParcelCategory != nil {
		repoFilters.ParcelCategory = filters.ParcelCategory
	}
	if filters.ServiceType != nil {
		repoFilters.ServiceType = filters.ServiceType
	}
	if filters.TATDays != nil {
		repoFilters.TATDays = filters.TATDays
	}
	if filters.Pickup != nil {
		repoFilters.Pickup = filters.Pickup
	}
	if filters.Delivery != nil {
		repoFilters.Delivery = filters.Delivery
	}
	if filters.DeliveryMode != nil {
		repoFilters.DeliveryMode = filters.DeliveryMode
	}
	if filters.CODAvailable != nil {
		repoFilters.CODAvailable = filters.CODAvailable
	}
	if filters.Insurance != nil {
		repoFilters.Insurance = filters.Insurance
	}
	if filters.MinWeightKG != nil {
		repoFilters.MinWeightKG = filters.MinWeightKG
	}
	if filters.MaxWeightKG != nil {
		repoFilters.MaxWeightKG = filters.MaxWeightKG
	}

	if filters.IsActive != nil {
		repoFilters.IsActive = filters.IsActive
	}

	// Get coverages from repository
	// Note: Using simplified approach, would need to extend repository for proper pagination with filters
	allCoverages, err := s.repo.GetByPartnerID(ctx, partnerID.String())
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
	coverages, err := s.repo.GetByPartnerID(ctx, partnerID.String())
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

// Create creates a new partner location coverage
func (s *partnerLocationCoverageService) Create(ctx context.Context, partnerID uuid.UUID, req *dtos.CreatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate partner exists
	if err := s.partnerValidationSvc.ValidatePartner(ctx, partnerID.String()); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Note: Removed postal code ID validation - API now works with postal code only

	// Create coverage model
	coverage := &models.PartnerLocationCoverage{
		ID:        uuid.New(),
		PartnerID: &partnerID,
		IsActive:  true, // Default to active
	}

	// Set optional fields
	if req.PartnerCode != nil {
		code := strings.TrimSpace(*req.PartnerCode)
		if code != "" {
			coverage.PartnerCode = &code
		}
	}
	if req.PostalCode != nil {
		code := strings.TrimSpace(*req.PostalCode)
		if code != "" {
			coverage.PostalCode = &code
		}
	}
	if req.PostalCodeID != nil {
		coverage.PostalCodeID = req.PostalCodeID
	}
	if req.ZoneType != nil {
		zoneType := strings.ToUpper(*req.ZoneType)
		coverage.ZoneType = &zoneType
	}

	// Set geographic context
	if req.CountryCode != nil {
		countryCode := strings.ToUpper(strings.TrimSpace(*req.CountryCode))
		if countryCode != "" {
			coverage.CountryCode = &countryCode
		}
	}

	// Set core serviceability attributes
	if req.ProductType != nil {
		productType := strings.TrimSpace(*req.ProductType)
		if productType != "" {
			coverage.ProductType = &productType
		}
	}
	if req.ParcelCategory != nil {
		parcelCategory := strings.TrimSpace(*req.ParcelCategory)
		if parcelCategory != "" {
			coverage.ParcelCategory = &parcelCategory
		}
	}
	if req.ServiceType != nil {
		serviceType := strings.TrimSpace(*req.ServiceType)
		if serviceType != "" {
			coverage.ServiceType = &serviceType
		}
	}
	if req.TATDays != nil {
		coverage.TATDays = req.TATDays
	}

	// Set service capabilities
	if req.Pickup != nil {
		coverage.Pickup = *req.Pickup
	}
	if req.Delivery != nil {
		coverage.Delivery = *req.Delivery
	}
	if req.DeliveryMode != nil {
		deliveryMode := strings.ToLower(strings.TrimSpace(*req.DeliveryMode))
		if deliveryMode != "" {
			coverage.DeliveryMode = &deliveryMode
		}
	}
	if req.CODAvailable != nil {
		coverage.CODAvailable = *req.CODAvailable
	}
	if req.Insurance != nil {
		coverage.Insurance = *req.Insurance
	}

	// Set weight constraints
	if req.MinWeightKG != nil {
		coverage.MinWeightKG = req.MinWeightKG
	}
	if req.MaxWeightKG != nil {
		coverage.MaxWeightKG = req.MaxWeightKG
	}

	if req.IsActive != nil {
		coverage.IsActive = *req.IsActive
	}

	// Validate business rules
	if err := coverage.ValidateBusinessRules(); err != nil {
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(ctx, coverage); err != nil {
		// Check for unique constraint violation (duplicate key)
		if strings.Contains(err.Error(), "unique_partner_postal_coverage") ||
			strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			postalCode := "unknown"
			if coverage.PostalCode != nil {
				postalCode = *coverage.PostalCode
			}
			return nil, fmt.Errorf("partner location coverage already exists for postal code %s", postalCode)
		}
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
		return nil, err
	}

	// Note: Removed postal code ID validation - API now works with postal code only

	// Create updated coverage model from existing response
	coverage := &models.PartnerLocationCoverage{
		ID:             existing.ID,
		PartnerID:      existing.PartnerID,
		PartnerCode:    existing.PartnerCode,
		PostalCode:     existing.PostalCode,
		PostalCodeID:   existing.PostalCodeID,
		ZoneType:       existing.ZoneType,
		CountryCode:    existing.CountryCode,
		ProductType:    existing.ProductType,
		ParcelCategory: existing.ParcelCategory,
		ServiceType:    existing.ServiceType,
		TATDays:        existing.TATDays,
		Pickup:         existing.Pickup,
		Delivery:       existing.Delivery,
		DeliveryMode:   existing.DeliveryMode,
		CODAvailable:   existing.CODAvailable,
		Insurance:      existing.Insurance,
		MinWeightKG:    existing.MinWeightKG,
		MaxWeightKG:    existing.MaxWeightKG,
		IsActive:       existing.IsActive,
		CreatedAt:      existing.CreatedAt,
		UpdatedAt:      existing.UpdatedAt,
	}

	// Apply updates
	if req.PartnerCode != nil {
		code := strings.TrimSpace(*req.PartnerCode)
		if code != "" {
			coverage.PartnerCode = &code
		} else {
			coverage.PartnerCode = nil
		}
	}
	if req.ZoneType != nil {
		zoneType := strings.ToUpper(*req.ZoneType)
		coverage.ZoneType = &zoneType
	}
	if req.PostalCode != nil {
		code := strings.TrimSpace(*req.PostalCode)
		if code != "" {
			coverage.PostalCode = &code
		} else {
			coverage.PostalCode = nil
		}
	}
	if req.PostalCodeID != nil {
		coverage.PostalCodeID = req.PostalCodeID
	}

	// Update geographic context
	if req.CountryCode != nil {
		countryCode := strings.ToUpper(strings.TrimSpace(*req.CountryCode))
		if countryCode != "" {
			coverage.CountryCode = &countryCode
		} else {
			coverage.CountryCode = nil
		}
	}

	// Update core serviceability attributes
	if req.ProductType != nil {
		productType := strings.TrimSpace(*req.ProductType)
		if productType != "" {
			coverage.ProductType = &productType
		} else {
			coverage.ProductType = nil
		}
	}
	if req.ParcelCategory != nil {
		parcelCategory := strings.TrimSpace(*req.ParcelCategory)
		if parcelCategory != "" {
			coverage.ParcelCategory = &parcelCategory
		} else {
			coverage.ParcelCategory = nil
		}
	}
	if req.ServiceType != nil {
		serviceType := strings.TrimSpace(*req.ServiceType)
		if serviceType != "" {
			coverage.ServiceType = &serviceType
		} else {
			coverage.ServiceType = nil
		}
	}
	if req.TATDays != nil {
		coverage.TATDays = req.TATDays
	}

	// Update service capabilities
	if req.Pickup != nil {
		coverage.Pickup = *req.Pickup
	}
	if req.Delivery != nil {
		coverage.Delivery = *req.Delivery
	}
	if req.DeliveryMode != nil {
		deliveryMode := strings.ToLower(strings.TrimSpace(*req.DeliveryMode))
		if deliveryMode != "" {
			coverage.DeliveryMode = &deliveryMode
		} else {
			coverage.DeliveryMode = nil
		}
	}
	if req.CODAvailable != nil {
		coverage.CODAvailable = *req.CODAvailable
	}
	if req.Insurance != nil {
		coverage.Insurance = *req.Insurance
	}

	// Update weight constraints
	if req.MinWeightKG != nil {
		coverage.MinWeightKG = req.MinWeightKG
	}
	if req.MaxWeightKG != nil {
		coverage.MaxWeightKG = req.MaxWeightKG
	}

	if req.IsActive != nil {
		coverage.IsActive = *req.IsActive
	}

	// Validate business rules
	if err := coverage.ValidateBusinessRules(); err != nil {
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	// Update in repository
	if err := s.repo.Update(ctx, coverage); err != nil {
		// Check for unique constraint violation (duplicate key)
		if strings.Contains(err.Error(), "unique_partner_postal_coverage") ||
			strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			postalCode := "unknown"
			if coverage.PostalCode != nil {
				postalCode = *coverage.PostalCode
			}
			return nil, fmt.Errorf("partner location coverage already exists for postal code %s", postalCode)
		}
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
	err = s.repo.Delete(ctx, coverageID.String())
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

// validatePostalCodeFields validates postal code field consistency
// DEPRECATED: This validation has been removed to allow API to work with postal code only
// func (s *partnerLocationCoverageService) validatePostalCodeFields(ctx context.Context, postalCode *string, postalCodeID *uuid.UUID) error {
// 	// Both postal code and postal code ID should be provided together for consistency
// 	if postalCode != nil && strings.TrimSpace(*postalCode) != "" && postalCodeID == nil {
// 		return fmt.Errorf("postal code provided but postal code ID is missing - both should be provided for consistency")
// 	}
// 	if postalCodeID != nil && (postalCode == nil || strings.TrimSpace(*postalCode) == "") {
// 		return fmt.Errorf("postal code ID provided but postal code is missing - both should be provided for consistency")
// 	}
//
// 	return nil
// }

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
		// Basic filters
		if filters.PostalCode != nil && (coverage.PostalCode == nil || *coverage.PostalCode != *filters.PostalCode) {
			continue
		}
		if filters.PostalCodeID != nil && (coverage.PostalCodeID == nil || *coverage.PostalCodeID != *filters.PostalCodeID) {
			continue
		}
		if filters.ZoneType != "" && (coverage.ZoneType == nil || *coverage.ZoneType != filters.ZoneType) {
			continue
		}

		// Serviceability filters
		if filters.CountryCode != nil && (coverage.CountryCode == nil || strings.ToUpper(*coverage.CountryCode) != strings.ToUpper(*filters.CountryCode)) {
			continue
		}
		if filters.ProductType != nil && (coverage.ProductType == nil || *coverage.ProductType != *filters.ProductType) {
			continue
		}
		if filters.ParcelCategory != nil && (coverage.ParcelCategory == nil || *coverage.ParcelCategory != *filters.ParcelCategory) {
			continue
		}
		if filters.ServiceType != nil && (coverage.ServiceType == nil || *coverage.ServiceType != *filters.ServiceType) {
			continue
		}
		if filters.TATDays != nil && (coverage.TATDays == nil || *coverage.TATDays != *filters.TATDays) {
			continue
		}
		if filters.Pickup != nil && coverage.Pickup != *filters.Pickup {
			continue
		}
		if filters.Delivery != nil && coverage.Delivery != *filters.Delivery {
			continue
		}
		if filters.DeliveryMode != nil && (coverage.DeliveryMode == nil || strings.ToLower(*coverage.DeliveryMode) != strings.ToLower(*filters.DeliveryMode)) {
			continue
		}
		if filters.CODAvailable != nil && coverage.CODAvailable != *filters.CODAvailable {
			continue
		}
		if filters.Insurance != nil && coverage.Insurance != *filters.Insurance {
			continue
		}
		if filters.MinWeightKG != nil && (coverage.MinWeightKG == nil || *coverage.MinWeightKG < *filters.MinWeightKG) {
			continue
		}
		if filters.MaxWeightKG != nil && (coverage.MaxWeightKG == nil || *coverage.MaxWeightKG > *filters.MaxWeightKG) {
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
		ID:             coverage.ID,
		PartnerID:      coverage.PartnerID,
		PartnerCode:    coverage.PartnerCode,
		PostalCode:     coverage.PostalCode,
		PostalCodeID:   coverage.PostalCodeID,
		ZoneType:       coverage.ZoneType,
		CountryCode:    coverage.CountryCode,
		ProductType:    coverage.ProductType,
		ParcelCategory: coverage.ParcelCategory,
		ServiceType:    coverage.ServiceType,
		TATDays:        coverage.TATDays,
		Pickup:         coverage.Pickup,
		Delivery:       coverage.Delivery,
		DeliveryMode:   coverage.DeliveryMode,
		CODAvailable:   coverage.CODAvailable,
		Insurance:      coverage.Insurance,
		MinWeightKG:    coverage.MinWeightKG,
		MaxWeightKG:    coverage.MaxWeightKG,
		IsActive:       coverage.IsActive,
		CreatedAt:      coverage.CreatedAt,
		UpdatedAt:      coverage.UpdatedAt,
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

// NEW POSTAL CODE BASED METHODS (using partner ID)

// GetByPartnerIDAndPostalCode retrieves a specific location coverage by partner ID and postal code
func (s *partnerLocationCoverageService) GetByPartnerIDAndPostalCode(ctx context.Context, partnerID uuid.UUID, postalCode string) (*dtos.PartnerLocationCoverageResponse, error) {
	// Validate partner exists
	if err := s.partnerValidationSvc.ValidatePartner(ctx, partnerID.String()); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Get all coverages for the partner
	coverages, err := s.repo.GetByPartnerID(ctx, partnerID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get partner location coverages: %w", err)
	}

	// Find the specific coverage by postal code
	for _, coverage := range coverages {
		if coverage.PostalCode != nil && *coverage.PostalCode == postalCode {
			return PartnerLocationCoverageToResponse(&coverage), nil
		}
	}

	return nil, fmt.Errorf("partner location coverage not found for postal code: %s", postalCode)
}

// UpdateByPartnerIDAndPostalCode updates a location coverage by partner ID and postal code
func (s *partnerLocationCoverageService) UpdateByPartnerIDAndPostalCode(ctx context.Context, partnerID uuid.UUID, postalCode string, req *dtos.UpdatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error) {
	// Validate partner exists
	if err := s.partnerValidationSvc.ValidatePartner(ctx, partnerID.String()); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Get all coverages for the partner
	coverages, err := s.repo.GetByPartnerID(ctx, partnerID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get partner location coverages: %w", err)
	}

	// Find the specific coverage by postal code
	var targetCoverage *models.PartnerLocationCoverage
	for i, coverage := range coverages {
		if coverage.PostalCode != nil && *coverage.PostalCode == postalCode {
			targetCoverage = &coverages[i]
			break
		}
	}

	if targetCoverage == nil {
		return nil, fmt.Errorf("partner location coverage not found for postal code: %s", postalCode)
	}

	// Use the existing Update method with the found coverage ID
	return s.Update(ctx, partnerID, targetCoverage.ID, req)
}

// DeleteByPartnerIDAndPostalCode deletes a location coverage by partner ID and postal code
func (s *partnerLocationCoverageService) DeleteByPartnerIDAndPostalCode(ctx context.Context, partnerID uuid.UUID, postalCode string) error {
	// Validate partner exists
	if err := s.partnerValidationSvc.ValidatePartner(ctx, partnerID.String()); err != nil {
		return fmt.Errorf("partner validation failed: %w", err)
	}

	// Get all coverages for the partner
	coverages, err := s.repo.GetByPartnerID(ctx, partnerID.String())
	if err != nil {
		return fmt.Errorf("failed to get partner location coverages: %w", err)
	}

	// Find the specific coverage by postal code
	for _, coverage := range coverages {
		if coverage.PostalCode != nil && *coverage.PostalCode == postalCode {
			return s.Delete(ctx, partnerID, coverage.ID)
		}
	}

	return fmt.Errorf("partner location coverage not found for postal code: %s", postalCode)
}

// CheckCoverage checks if a partner covers a specific postal code with optional service requirements
func (s *partnerLocationCoverageService) CheckCoverage(ctx context.Context, partnerID uuid.UUID, postalCode string, filters *dtos.PartnerLocationCoverageFiltersRequest) (*dtos.CoverageCheckResponse, error) {
	// Validate partner exists
	if err := s.partnerValidationSvc.ValidatePartner(ctx, partnerID.String()); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Get all coverages for the partner
	coverages, err := s.repo.GetByPartnerID(ctx, partnerID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get partner location coverages: %w", err)
	}

	// Find coverage for the postal code
	var targetCoverage *models.PartnerLocationCoverage
	for i, coverage := range coverages {
		if coverage.PostalCode != nil && *coverage.PostalCode == postalCode && coverage.IsActive {
			targetCoverage = &coverages[i]
			break
		}
	}

	response := &dtos.CoverageCheckResponse{
		PostalCode: postalCode,
		Covered:    false,
	}

	if targetCoverage == nil {
		message := "No coverage available for this postal code"
		response.Message = &message
		return response, nil
	}

	// Check if coverage meets the filter requirements
	if filters != nil {
		if !s.coverageMeetsRequirements(targetCoverage, filters) {
			message := "Coverage available but does not meet service requirements"
			response.Message = &message
			return response, nil
		}
	}

	// Coverage found and meets requirements
	response.Covered = true
	response.PartnerCode = targetCoverage.PartnerCode
	response.ServiceCapabilities = &dtos.CoverageServiceCapabilities{
		Pickup:       targetCoverage.Pickup,
		Delivery:     targetCoverage.Delivery,
		CODAvailable: targetCoverage.CODAvailable,
		Insurance:    targetCoverage.Insurance,
		ServiceType:  targetCoverage.ServiceType,
		TATDays:      targetCoverage.TATDays,
		DeliveryMode: targetCoverage.DeliveryMode,
	}

	if targetCoverage.MinWeightKG != nil || targetCoverage.MaxWeightKG != nil {
		response.ServiceCapabilities.WeightRange = &dtos.CoverageWeightRange{
			MinKG: targetCoverage.MinWeightKG,
			MaxKG: targetCoverage.MaxWeightKG,
		}
	}

	message := "Coverage available with all requested services"
	response.Message = &message

	return response, nil
}

// NEW PARTNER CODE BASED METHODS

// GetByPartnerCode retrieves all location coverages for a partner by partner code
func (s *partnerLocationCoverageService) GetByPartnerCode(ctx context.Context, partnerCode string, filters *dtos.PartnerLocationCoverageFiltersRequest) (*dtos.PartnerLocationCoverageListResponse, error) {
	// Validate partner exists by code
	if err := s.partnerValidationSvc.ValidatePartnerByCode(ctx, partnerCode); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Get partner ID from code (simplified approach - in real implementation, you'd have a partner service)
	// For now, we'll get all coverages and filter by partner code
	// This is not optimal but works for the current structure

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

	// Get coverages by partner code from repository
	coverages, err := s.repo.GetByPartnerCode(ctx, partnerCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner location coverages: %w", err)
	}

	// Build repository filters
	repoFilters := &models.PartnerLocationCoverageFilters{
		PartnerCode: &partnerCode,
	}
	if filters.PostalCode != nil {
		repoFilters.PostalCode = filters.PostalCode
	}
	if filters.PostalCodeID != nil {
		repoFilters.PostalCodeID = filters.PostalCodeID
	}
	if filters.ZoneType != nil {
		repoFilters.ZoneType = strings.ToUpper(*filters.ZoneType)
	}
	// Add all other serviceability filters...
	s.applyServiceabilityFilters(repoFilters, filters)

	// Apply client-side filtering
	filteredCoverages := s.filterCoverages(coverages, repoFilters)

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

// CreateByPartnerCode creates a new location coverage for a partner by partner code
func (s *partnerLocationCoverageService) CreateByPartnerCode(ctx context.Context, partnerCode string, req *dtos.CreatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	// Validate partner exists by code
	if err := s.partnerValidationSvc.ValidatePartnerByCode(ctx, partnerCode); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Set partner code in request if not provided
	if req.PartnerCode == nil {
		req.PartnerCode = &partnerCode
	}

	// Get partner ID from code (simplified approach)
	partnerID, err := s.getPartnerIDByCode(ctx, partnerCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner ID: %w", err)
	}

	return s.Create(ctx, partnerID, req)
}

// BulkCreateByPartnerCode creates multiple location coverages for a partner by partner code
func (s *partnerLocationCoverageService) BulkCreateByPartnerCode(ctx context.Context, partnerCode string, req *dtos.BulkCreatePartnerLocationCoverageRequest) (*dtos.BulkCreatePartnerLocationCoverageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("bulk create request cannot be nil")
	}

	// Validate partner exists by code
	if err := s.partnerValidationSvc.ValidatePartnerByCode(ctx, partnerCode); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Set partner code in all requests if not provided
	for i := range req.Coverages {
		if req.Coverages[i].PartnerCode == nil {
			req.Coverages[i].PartnerCode = &partnerCode
		}
	}

	// Get partner ID from code
	partnerID, err := s.getPartnerIDByCode(ctx, partnerCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner ID: %w", err)
	}

	return s.BulkCreate(ctx, partnerID, req)
}

// GetByPartnerCodeAndPostalCode retrieves a specific location coverage by partner code and postal code
func (s *partnerLocationCoverageService) GetByPartnerCodeAndPostalCode(ctx context.Context, partnerCode string, postalCode string) (*dtos.PartnerLocationCoverageResponse, error) {
	// Validate partner exists by code
	if err := s.partnerValidationSvc.ValidatePartnerByCode(ctx, partnerCode); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Get coverages by partner code
	coverages, err := s.repo.GetByPartnerCode(ctx, partnerCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner location coverages: %w", err)
	}

	// Find the specific coverage by postal code
	for _, coverage := range coverages {
		if coverage.PostalCode != nil && *coverage.PostalCode == postalCode {
			return PartnerLocationCoverageToResponse(&coverage), nil
		}
	}

	return nil, fmt.Errorf("partner location coverage not found for postal code: %s", postalCode)
}

// UpdateByPartnerCodeAndPostalCode updates a location coverage by partner code and postal code
func (s *partnerLocationCoverageService) UpdateByPartnerCodeAndPostalCode(ctx context.Context, partnerCode string, postalCode string, req *dtos.UpdatePartnerLocationCoverageRequest) (*dtos.PartnerLocationCoverageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	// Validate partner exists by code
	if err := s.partnerValidationSvc.ValidatePartnerByCode(ctx, partnerCode); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Get partner ID from code
	partnerID, err := s.getPartnerIDByCode(ctx, partnerCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner ID: %w", err)
	}

	return s.UpdateByPartnerIDAndPostalCode(ctx, partnerID, postalCode, req)
}

// DeleteByPartnerCodeAndPostalCode deletes a location coverage by partner code and postal code
func (s *partnerLocationCoverageService) DeleteByPartnerCodeAndPostalCode(ctx context.Context, partnerCode string, postalCode string) error {
	// Validate partner exists by code
	if err := s.partnerValidationSvc.ValidatePartnerByCode(ctx, partnerCode); err != nil {
		return fmt.Errorf("partner validation failed: %w", err)
	}

	// Get partner ID from code
	partnerID, err := s.getPartnerIDByCode(ctx, partnerCode)
	if err != nil {
		return fmt.Errorf("failed to get partner ID: %w", err)
	}

	return s.DeleteByPartnerIDAndPostalCode(ctx, partnerID, postalCode)
}

// CheckCoverageByPartnerCode checks if a partner covers a specific postal code with optional service requirements
func (s *partnerLocationCoverageService) CheckCoverageByPartnerCode(ctx context.Context, partnerCode string, postalCode string, filters *dtos.PartnerLocationCoverageFiltersRequest) (*dtos.CoverageCheckResponse, error) {
	// Validate partner exists by code
	if err := s.partnerValidationSvc.ValidatePartnerByCode(ctx, partnerCode); err != nil {
		return nil, fmt.Errorf("partner validation failed: %w", err)
	}

	// Get partner ID from code
	partnerID, err := s.getPartnerIDByCode(ctx, partnerCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner ID: %w", err)
	}

	return s.CheckCoverage(ctx, partnerID, postalCode, filters)
}

// HELPER METHODS

// coverageMeetsRequirements checks if a coverage meets the filter requirements
func (s *partnerLocationCoverageService) coverageMeetsRequirements(coverage *models.PartnerLocationCoverage, filters *dtos.PartnerLocationCoverageFiltersRequest) bool {
	if filters.ServiceType != nil && (coverage.ServiceType == nil || *coverage.ServiceType != *filters.ServiceType) {
		return false
	}
	if filters.TATDays != nil && (coverage.TATDays == nil || *coverage.TATDays != *filters.TATDays) {
		return false
	}
	if filters.Pickup != nil && coverage.Pickup != *filters.Pickup {
		return false
	}
	if filters.Delivery != nil && coverage.Delivery != *filters.Delivery {
		return false
	}
	if filters.CODAvailable != nil && coverage.CODAvailable != *filters.CODAvailable {
		return false
	}
	if filters.Insurance != nil && coverage.Insurance != *filters.Insurance {
		return false
	}
	if filters.DeliveryMode != nil && (coverage.DeliveryMode == nil || *coverage.DeliveryMode != *filters.DeliveryMode) {
		return false
	}
	if filters.MinWeightKG != nil && (coverage.MinWeightKG == nil || *coverage.MinWeightKG > *filters.MinWeightKG) {
		return false
	}
	if filters.MaxWeightKG != nil && (coverage.MaxWeightKG == nil || *coverage.MaxWeightKG < *filters.MaxWeightKG) {
		return false
	}
	return true
}

// applyServiceabilityFilters applies serviceability filters to repository filters
func (s *partnerLocationCoverageService) applyServiceabilityFilters(repoFilters *models.PartnerLocationCoverageFilters, filters *dtos.PartnerLocationCoverageFiltersRequest) {
	if filters.CountryCode != nil {
		repoFilters.CountryCode = filters.CountryCode
	}
	if filters.ProductType != nil {
		repoFilters.ProductType = filters.ProductType
	}
	if filters.ParcelCategory != nil {
		repoFilters.ParcelCategory = filters.ParcelCategory
	}
	if filters.ServiceType != nil {
		repoFilters.ServiceType = filters.ServiceType
	}
	if filters.TATDays != nil {
		repoFilters.TATDays = filters.TATDays
	}
	if filters.Pickup != nil {
		repoFilters.Pickup = filters.Pickup
	}
	if filters.Delivery != nil {
		repoFilters.Delivery = filters.Delivery
	}
	if filters.DeliveryMode != nil {
		repoFilters.DeliveryMode = filters.DeliveryMode
	}
	if filters.CODAvailable != nil {
		repoFilters.CODAvailable = filters.CODAvailable
	}
	if filters.Insurance != nil {
		repoFilters.Insurance = filters.Insurance
	}
	if filters.MinWeightKG != nil {
		repoFilters.MinWeightKG = filters.MinWeightKG
	}
	if filters.MaxWeightKG != nil {
		repoFilters.MaxWeightKG = filters.MaxWeightKG
	}
	if filters.IsActive != nil {
		repoFilters.IsActive = filters.IsActive
	}
}

// getPartnerIDByCode gets partner ID from partner code (simplified implementation)
func (s *partnerLocationCoverageService) getPartnerIDByCode(ctx context.Context, partnerCode string) (uuid.UUID, error) {
	// This is a simplified implementation
	// In a real system, you would have a partner service to get partner details
	// For now, we'll get coverages by partner code and extract the partner ID
	coverages, err := s.repo.GetByPartnerCode(ctx, partnerCode)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get coverages by partner code: %w", err)
	}

	if len(coverages) == 0 {
		return uuid.Nil, fmt.Errorf("no coverages found for partner code: %s", partnerCode)
	}

	if coverages[0].PartnerID == nil {
		return uuid.Nil, fmt.Errorf("partner ID not found for partner code: %s", partnerCode)
	}

	return *coverages[0].PartnerID, nil
}
