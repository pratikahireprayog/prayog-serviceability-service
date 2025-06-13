package services

import (
	"context"
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"
)

// PostalCodeServiceabilityService defines the interface for postal code based serviceability operations
type PostalCodeServiceabilityService interface {
	// GetServiceabilityByPostalCode retrieves serviceability data for a single postal code
	GetServiceabilityByPostalCode(ctx context.Context, postalCode string, filters *dtos.PostalCodeServiceabilityRequest) (*dtos.PostalCodeServiceabilityResponse, error)

	// CheckServiceability checks serviceability for source and destination postal codes
	CheckServiceability(ctx context.Context, req *dtos.PostalCodeServiceabilityRequest) (*dtos.PostalCodeServiceabilityResponse, error)
}

// postalCodeServiceabilityService implements the PostalCodeServiceabilityService interface
type postalCodeServiceabilityService struct {
	partnerCoverageRepo repositories.PartnerLocationCoverageRepository
}

// NewPostalCodeServiceabilityService creates a new postal code serviceability service instance
func NewPostalCodeServiceabilityService(
	partnerCoverageRepo repositories.PartnerLocationCoverageRepository,
) PostalCodeServiceabilityService {
	return &postalCodeServiceabilityService{
		partnerCoverageRepo: partnerCoverageRepo,
	}
}

// GetServiceabilityByPostalCode retrieves serviceability information for a specific postal code
func (s *postalCodeServiceabilityService) GetServiceabilityByPostalCode(ctx context.Context, postalCode string, filters *dtos.PostalCodeServiceabilityRequest) (*dtos.PostalCodeServiceabilityResponse, error) {
	// Fetch partner coverages for the postal code
	coverages, err := s.getPartnerCoveragesByPostalCode(ctx, postalCode)
	if err != nil {
		return s.createErrorResponse("DATA_FETCH_ERROR", fmt.Sprintf("Failed to fetch coverage data: %v", err)), nil
	}

	// If no coverage found, return not serviceable
	if len(coverages) == 0 {
		isServiceable := false
		return &dtos.PostalCodeServiceabilityResponse{
			Success:       true,
			IsServiceable: &isServiceable,
			Data: map[string]interface{}{
				"code":    "POSTAL_CODE_NOT_SERVICEABLE",
				"message": "Postal code is not serviceable",
				"details": fmt.Sprintf("Postal code %s is not available in our service area", postalCode),
			},
		}, nil
	}

	// Apply filters if provided
	if filters != nil {
		coverages = s.applyCoverageFilters(coverages, filters)
	}

	// If product_type is specified and no coverages remain after filtering, return error response
	if filters != nil && filters.ProductType != nil && *filters.ProductType != "" && len(coverages) == 0 {
		// Return error response when product_type filtering results in no coverages
		isServiceable := false
		return &dtos.PostalCodeServiceabilityResponse{
			Success:       true,
			IsServiceable: &isServiceable,
			Data: map[string]interface{}{
				"code":    "POSTAL_CODE_NOT_SERVICEABLE",
				"message": "Postal code is not serviceable",
				"details": fmt.Sprintf("Postal code %s is not available in our service area", postalCode),
			},
		}, nil
	}

	// Transform coverage data to serviceability format
	serviceabilityData := s.transformCoverageToServiceability(coverages, filters)

	// Normal case: return serviceability data
	isServiceable := true
	return &dtos.PostalCodeServiceabilityResponse{
		Success:       true,
		IsServiceable: &isServiceable,
		Data:          serviceabilityData,
	}, nil
}

// CheckServiceability checks serviceability for source and destination postal codes
func (s *postalCodeServiceabilityService) CheckServiceability(ctx context.Context, req *dtos.PostalCodeServiceabilityRequest) (*dtos.PostalCodeServiceabilityResponse, error) {
	if req == nil {
		return s.createErrorResponse("INVALID_REQUEST", "Request cannot be nil"), nil
	}

	// Step 1: Check if source postal code is provided and is serviceable
	if req.SourcePostalCode != nil && *req.SourcePostalCode != "" {
		sourceCoverages, err := s.getPartnerCoveragesByPostalCode(ctx, *req.SourcePostalCode)
		if err != nil {
			return s.createErrorResponse("DATA_FETCH_ERROR", fmt.Sprintf("Failed to fetch source coverage data: %v", err)), nil
		}

		// If source postal code has no coverage (not in DB), return not serviceable
		if len(sourceCoverages) == 0 {
			isServiceable := false
			return &dtos.PostalCodeServiceabilityResponse{
				Success:       true,
				IsServiceable: &isServiceable,
				Data: map[string]interface{}{
					"code":    "SOURCE_POSTAL_CODE_NOT_SERVICEABLE",
					"message": "Source postal code is not serviceable",
					"details": fmt.Sprintf("Source postal code %s is not available in our service area", *req.SourcePostalCode),
				},
			}, nil
		}

		// Source postal code validation complete - no filtering needed
		// We only check if source postal code exists in DB, not apply filters
	}

	// Step 2: If source is serviceable (or not provided), check destination postal code
	destinationCoverages, err := s.getPartnerCoveragesByPostalCode(ctx, req.DestinationPostalCode)
	if err != nil {
		return s.createErrorResponse("DATA_FETCH_ERROR", fmt.Sprintf("Failed to fetch destination coverage data: %v", err)), nil
	}

	// If destination postal code has no coverage, return not serviceable
	if len(destinationCoverages) == 0 {
		isServiceable := false
		return &dtos.PostalCodeServiceabilityResponse{
			Success:       true,
			IsServiceable: &isServiceable,
			Data: map[string]interface{}{
				"code":    "DESTINATION_POSTAL_CODE_NOT_SERVICEABLE",
				"message": "Destination postal code is not serviceable",
				"details": fmt.Sprintf("Destination postal code %s is not available in our service area", req.DestinationPostalCode),
			},
		}, nil
	}

	// Step 3: Apply filters to destination coverage
	filteredDestinationCoverages := s.applyCoverageFilters(destinationCoverages, req)

	// If product_type is specified and no coverages remain after filtering, return error response
	if req.ProductType != nil && *req.ProductType != "" && len(filteredDestinationCoverages) == 0 {
		// Return error response when product_type filtering results in no coverages
		isServiceable := false
		return &dtos.PostalCodeServiceabilityResponse{
			Success:       true,
			IsServiceable: &isServiceable,
			Data: map[string]interface{}{
				"code":    "DESTINATION_POSTAL_CODE_NOT_SERVICEABLE",
				"message": "Destination postal code is not serviceable",
				"details": fmt.Sprintf("Destination postal code %s is not available in our service area", req.DestinationPostalCode),
			},
		}, nil
	}

	// Step 4: Transform coverage data based on destination
	serviceabilityData := s.transformCoverageToServiceability(filteredDestinationCoverages, req)

	// Normal case: return serviceability data
	isServiceable := true
	return &dtos.PostalCodeServiceabilityResponse{
		Success:       true,
		IsServiceable: &isServiceable,
		Data:          serviceabilityData,
	}, nil
}

// getPartnerCoveragesByPostalCode retrieves all partner coverage data for a postal code
func (s *postalCodeServiceabilityService) getPartnerCoveragesByPostalCode(ctx context.Context, postalCode string) ([]models.PartnerLocationCoverage, error) {
	// Use the repository method to get all partner location coverage data for the postal code
	coverages, err := s.partnerCoverageRepo.GetByPostalCode(ctx, postalCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner location coverages by postal code: %w", err)
	}

	return coverages, nil
}

// applyCoverageFilters applies request filters to the coverage data
func (s *postalCodeServiceabilityService) applyCoverageFilters(coverages []models.PartnerLocationCoverage, req *dtos.PostalCodeServiceabilityRequest) []models.PartnerLocationCoverage {
	filtered := make([]models.PartnerLocationCoverage, 0)

	for _, coverage := range coverages {
		// Only include active coverages
		if !coverage.IsActive {
			continue
		}

		// Filter by parcel category if specifically requested
		// If no filter is provided, include ALL parcel categories found in database
		if req.ParcelCategory != nil && *req.ParcelCategory != "" {
			if coverage.ParcelCategory == nil || !s.normalizeParcelCategory(*coverage.ParcelCategory, *req.ParcelCategory) {
				continue
			}
		}

		// Filter by product type if specified
		if req.ProductType != nil && *req.ProductType != "" {
			if coverage.ProductType == nil || *coverage.ProductType != *req.ProductType {
				continue
			}
		}

		filtered = append(filtered, coverage)
	}

	return filtered
}

// normalizeParcelCategory handles the mapping between "ecomm" and "ecom"
func (s *postalCodeServiceabilityService) normalizeParcelCategory(dbCategory, requestCategory string) bool {
	dbCat := strings.ToLower(strings.TrimSpace(dbCategory))
	reqCat := strings.ToLower(strings.TrimSpace(requestCategory))

	// Handle ecomm/ecom mapping
	if (dbCat == "ecom" && reqCat == "ecomm") || (dbCat == "ecomm" && reqCat == "ecom") {
		return true
	}

	return dbCat == reqCat
}

// transformCoverageToServiceability transforms partner coverage data to serviceability response
func (s *postalCodeServiceabilityService) transformCoverageToServiceability(coverages []models.PartnerLocationCoverage, req *dtos.PostalCodeServiceabilityRequest) *dtos.PostalCodeServiceabilityData {
	// Group coverages by parcel category
	categoryMap := make(map[string]map[string]*serviceAggregator)
	// Track all categories found, even if they have no valid services
	allCategories := make(map[string]bool)

	// If a specific parcel category is requested, ensure it's included even if no services are found
	if req != nil && req.ParcelCategory != nil && *req.ParcelCategory != "" {
		requestedCategory := s.getCategoryCode(req.ParcelCategory)
		allCategories[requestedCategory] = true
	}

	for _, coverage := range coverages {
		categoryCode := s.getCategoryCode(coverage.ParcelCategory)
		serviceCode := s.getServiceCode(coverage.ServiceType)

		// Track this category as existing
		allCategories[categoryCode] = true

		// Skip unknown services but still track the category
		if serviceCode == "unknown" {
			continue
		}

		// Initialize category if not exists
		if categoryMap[categoryCode] == nil {
			categoryMap[categoryCode] = make(map[string]*serviceAggregator)
		}

		// Initialize service if not exists
		if categoryMap[categoryCode][serviceCode] == nil {
			categoryMap[categoryCode][serviceCode] = &serviceAggregator{
				ServiceCode:  serviceCode,
				TATDays:      coverage.TATDays,
				IsCOD:        coverage.CODAvailable,
				Pickup:       &coverage.Pickup,
				Delivery:     &coverage.Delivery,
				Insurance:    &coverage.Insurance,
				AirMode:      false,
				SurfaceMode:  false,
				ProductTypes: make(map[string]bool),
			}
		}

		service := categoryMap[categoryCode][serviceCode]

		// Track product types for this service
		if coverage.ProductType != nil && *coverage.ProductType != "" {
			service.ProductTypes[*coverage.ProductType] = true
		}

		// Update delivery modes based on coverage delivery mode
		if coverage.DeliveryMode != nil {
			deliveryMode := strings.ToLower(*coverage.DeliveryMode)
			if deliveryMode == "air" {
				service.AirMode = true
			} else if deliveryMode == "surface" {
				service.SurfaceMode = true
			}
		}

		// Update service capabilities with best values
		if coverage.TATDays != nil && (service.TATDays == nil || *coverage.TATDays < *service.TATDays) {
			service.TATDays = coverage.TATDays
		}

		service.IsCOD = service.IsCOD || coverage.CODAvailable
		if service.Pickup != nil {
			pickup := *service.Pickup || coverage.Pickup
			service.Pickup = &pickup
		}
		if service.Delivery != nil {
			delivery := *service.Delivery || coverage.Delivery
			service.Delivery = &delivery
		}
		if service.Insurance != nil {
			insurance := *service.Insurance || coverage.Insurance
			service.Insurance = &insurance
		}
	}

	// Convert to response format - include all categories found, even if they have no services
	serviceability := make([]dtos.PostalCodeParcelCategoryService, 0, len(allCategories))

	for categoryCode := range allCategories {
		serviceList := make([]dtos.PostalCodeServiceInfo, 0)

		// Add services if they exist for this category
		if services, exists := categoryMap[categoryCode]; exists {
			serviceList = make([]dtos.PostalCodeServiceInfo, 0, len(services))
			for _, service := range services {
				serviceInfo := dtos.PostalCodeServiceInfo{
					ServiceCode: service.ServiceCode,
					TATDays:     service.TATDays,
					IsCOD:       service.IsCOD,
					Pickup:      service.Pickup,
					Delivery:    service.Delivery,
					Insurance:   service.Insurance,
					DeliveryModes: dtos.PostalCodeServiceDeliveryModes{
						Air:     service.AirMode,
						Surface: service.SurfaceMode,
					},
				}

				// Only include product types if the service has any
				if len(service.ProductTypes) > 0 {
					serviceInfo.ProductTypes = service.ProductTypes
				}

				serviceList = append(serviceList, serviceInfo)
			}
		}

		serviceability = append(serviceability, dtos.PostalCodeParcelCategoryService{
			ParcelCategoryCode: categoryCode,
			IsServiceable:      true, // Category is serviceable since postal code exists
			Services:           serviceList,
		})
	}

	// For POST API - include both source and destination postal codes
	var sourcePostalCode *string
	if req.SourcePostalCode != nil && *req.SourcePostalCode != "" {
		sourcePostalCode = req.SourcePostalCode
	}

	return &dtos.PostalCodeServiceabilityData{
		SourcePostalCode:      sourcePostalCode,
		DestinationPostalCode: req.DestinationPostalCode,
		Serviceability:        serviceability,
	}
}

// serviceAggregator is a helper struct for aggregating service data
type serviceAggregator struct {
	ServiceCode  string
	TATDays      *int
	IsCOD        bool
	Pickup       *bool
	Delivery     *bool
	Insurance    *bool
	AirMode      bool
	SurfaceMode  bool
	ProductTypes map[string]bool
}

// getCategoryCode uses the actual parcel category from database with minimal normalization
func (s *postalCodeServiceabilityService) getCategoryCode(category *string) string {
	if category == nil || strings.TrimSpace(*category) == "" {
		return "unknown" // For null/empty categories, use unknown instead of defaulting
	}

	// Use the actual category from database with just basic normalization
	cat := strings.ToLower(strings.TrimSpace(*category))

	// Only normalize common variations, otherwise use the exact database value
	switch cat {
	case "ecom", "ecommerce":
		return "ecomm"
	default:
		// Return the actual database value as-is (lowercased and trimmed)
		return cat
	}
}

// getServiceCode uses the actual service type from database with minimal normalization
func (s *postalCodeServiceabilityService) getServiceCode(serviceType *string) string {
	if serviceType == nil || strings.TrimSpace(*serviceType) == "" {
		return "unknown" // For null/empty service types, use unknown instead of defaulting
	}

	// Use the actual service type from database with just basic normalization
	service := strings.ToLower(strings.TrimSpace(*serviceType))

	// Only normalize critical business mappings, otherwise use the exact database value
	switch service {
	case "express":
		return "sdd" // Business requirement: express maps to same day delivery
	case "standard":
		return "ndd" // Business requirement: standard maps to next day delivery
	default:
		// Return the actual database value as-is (lowercased and trimmed)
		return service
	}
}

// createErrorResponse creates an error response
func (s *postalCodeServiceabilityService) createErrorResponse(code, message string) *dtos.PostalCodeServiceabilityResponse {
	isServiceable := false
	return &dtos.PostalCodeServiceabilityResponse{
		Success:       false,
		IsServiceable: &isServiceable,
		Error: &dtos.PostalCodeServiceabilityError{
			Code:    code,
			Message: message,
		},
	}
}

// createEmptyServiceabilityResponse creates a response with no serviceability for POST API
func (s *postalCodeServiceabilityService) createEmptyServiceabilityResponse(req *dtos.PostalCodeServiceabilityRequest) *dtos.PostalCodeServiceabilityResponse {
	// For POST API - include source postal code if provided
	var sourcePostalCode *string
	if req.SourcePostalCode != nil && *req.SourcePostalCode != "" {
		sourcePostalCode = req.SourcePostalCode
	}

	isServiceable := false
	return &dtos.PostalCodeServiceabilityResponse{
		Success:       true,
		IsServiceable: &isServiceable,
		Data: &dtos.PostalCodeServiceabilityData{
			SourcePostalCode:      sourcePostalCode,
			DestinationPostalCode: req.DestinationPostalCode,
			Serviceability:        []dtos.PostalCodeParcelCategoryService{},
		},
	}
}

// createEmptyServiceabilityResponseSingle creates a response with no serviceability for GET API
func (s *postalCodeServiceabilityService) createEmptyServiceabilityResponseSingle(postalCode string, filters *dtos.PostalCodeServiceabilityRequest) *dtos.PostalCodeServiceabilityResponse {
	isServiceable := false
	return &dtos.PostalCodeServiceabilityResponse{
		Success:       true,
		IsServiceable: &isServiceable,
		Data: &dtos.PostalCodeServiceabilityData{
			SourcePostalCode:      nil, // No source postal code for GET API
			DestinationPostalCode: postalCode,
			Serviceability:        []dtos.PostalCodeParcelCategoryService{},
		},
	}
}
