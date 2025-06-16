package services

import (
	"context"
	"fmt"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// serviceabilityOrchestrator implements the ServiceabilityOrchestrator interface
type serviceabilityOrchestrator struct {
	locationResolver            interfaces.LocationResolver
	partnerCapabilityAggregator interfaces.PartnerCapabilityAggregator
	serviceDefinitionResolver   interfaces.ServiceDefinitionResolver
	serviceabilityCalculator    interfaces.ServiceabilityCalculator
}

// NewServiceabilityOrchestrator creates a new instance of ServiceabilityOrchestrator
func NewServiceabilityOrchestrator(
	locationResolver interfaces.LocationResolver,
	partnerCapabilityAggregator interfaces.PartnerCapabilityAggregator,
	serviceDefinitionResolver interfaces.ServiceDefinitionResolver,
	serviceabilityCalculator interfaces.ServiceabilityCalculator,
) interfaces.ServiceabilityOrchestrator {
	return &serviceabilityOrchestrator{
		locationResolver:            locationResolver,
		partnerCapabilityAggregator: partnerCapabilityAggregator,
		serviceDefinitionResolver:   serviceDefinitionResolver,
		serviceabilityCalculator:    serviceabilityCalculator,
	}
}

// CheckServiceability orchestrates the single location serviceability check process
func (s *serviceabilityOrchestrator) CheckServiceability(ctx context.Context, req *models.ServiceabilityCheckRequest) (*models.ServiceabilityResponse, error) {
	// Validate input request
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if err := s.validateServiceabilityRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Step 1: Resolve and validate location(s)
	var locationHierarchy, pickupHierarchy, deliveryHierarchy *models.LocationHierarchy
	var err error

	if req.PostalCode != nil && *req.PostalCode != "" {
		// Single location query
		locationHierarchy, err = s.locationResolver.GetLocationHierarchy(ctx, *req.PostalCode, req.CountryCode)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve location for postal code %s: %w", *req.PostalCode, err)
		}

		// Validate postal code is active
		isActive, err := s.locationResolver.IsPostalCodeActive(ctx, *req.PostalCode, req.CountryCode)
		if err != nil {
			return nil, fmt.Errorf("failed to validate postal code %s: %w", *req.PostalCode, err)
		}
		if !isActive {
			return &models.ServiceabilityResponse{
				Success: true,
				Data:    &models.ServiceabilityData{},
				Error: &models.ErrorResponse{
					Code:    "POSTAL_CODE_INACTIVE",
					Message: fmt.Sprintf("Postal code %s is not active for serviceability", *req.PostalCode),
				},
			}, nil
		}
	} else if req.PickupPostalCode != nil && *req.PickupPostalCode != "" && req.DeliveryPostalCode != nil && *req.DeliveryPostalCode != "" {
		// Origin-destination query
		pickupHierarchy, err = s.locationResolver.GetLocationHierarchy(ctx, *req.PickupPostalCode, req.CountryCode)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve pickup location for postal code %s: %w", *req.PickupPostalCode, err)
		}

		deliveryHierarchy, err = s.locationResolver.GetLocationHierarchy(ctx, *req.DeliveryPostalCode, req.CountryCode)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve delivery location for postal code %s: %w", *req.DeliveryPostalCode, err)
		}

		// Validate both postal codes are active
		isPickupActive, err := s.locationResolver.IsPostalCodeActive(ctx, *req.PickupPostalCode, req.CountryCode)
		if err != nil {
			return nil, fmt.Errorf("failed to validate pickup postal code %s: %w", *req.PickupPostalCode, err)
		}

		isDeliveryActive, err := s.locationResolver.IsPostalCodeActive(ctx, *req.DeliveryPostalCode, req.CountryCode)
		if err != nil {
			return nil, fmt.Errorf("failed to validate delivery postal code %s: %w", *req.DeliveryPostalCode, err)
		}

		if !isPickupActive || !isDeliveryActive {
			return &models.ServiceabilityResponse{
				Success: true,
				Data:    &models.ServiceabilityData{},
				Error: &models.ErrorResponse{
					Code:    "POSTAL_CODE_INACTIVE",
					Message: "One or more postal codes are not active for serviceability",
				},
			}, nil
		}
	}

	// Step 2: Get partner capabilities for the location(s)
	var partnerCapabilities []interfaces.PartnerCapability

	if locationHierarchy != nil {
		partnerCapabilities, err = s.partnerCapabilityAggregator.GetPartnersByLocation(ctx, locationHierarchy)
		if err != nil {
			return nil, fmt.Errorf("failed to get partners for location: %w", err)
		}
	} else if pickupHierarchy != nil && deliveryHierarchy != nil {
		// Get partners for both pickup and delivery locations
		pickupPartners, err := s.partnerCapabilityAggregator.GetPartnersByLocation(ctx, pickupHierarchy)
		if err != nil {
			return nil, fmt.Errorf("failed to get partners for pickup location: %w", err)
		}

		deliveryPartners, err := s.partnerCapabilityAggregator.GetPartnersByLocation(ctx, deliveryHierarchy)
		if err != nil {
			return nil, fmt.Errorf("failed to get partners for delivery location: %w", err)
		}

		// Combine partner capabilities (partners that serve both locations)
		partnerCapabilities = s.combinePartnerCapabilities(pickupPartners, deliveryPartners)
	}

	// Step 3: Get service definitions
	serviceDefinitions, err := s.getRelevantServiceDefinitions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get service definitions: %w", err)
	}

	// Step 4: Calculate serviceability
	calculationRequest := &interfaces.ServiceabilityCalculationRequest{
		LocationHierarchy:   locationHierarchy,
		PickupHierarchy:     pickupHierarchy,
		DeliveryHierarchy:   deliveryHierarchy,
		PartnerCapabilities: partnerCapabilities,
		ServiceDefinitions:  serviceDefinitions,
		Filters:             s.buildServiceFilters(req),
	}

	serviceabilityData, err := s.serviceabilityCalculator.CalculateServiceability(ctx, calculationRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate serviceability: %w", err)
	}

	// Step 5: Build response
	response := &models.ServiceabilityResponse{
		Success: true,
		Data:    serviceabilityData,
	}

	return response, nil
}

// BulkCheckServiceability orchestrates bulk serviceability checks
func (s *serviceabilityOrchestrator) BulkCheckServiceability(ctx context.Context, req *models.BulkServiceabilityRequest) (*models.BulkServiceabilityResponse, error) {
	if req == nil || len(req.Requests) == 0 {
		return nil, fmt.Errorf("bulk request cannot be nil or empty")
	}

	responses := make([]models.ServiceabilityResponse, 0, len(req.Requests))
	errorCount := 0

	// Process each individual request
	for i, individualReq := range req.Requests {
		response, err := s.CheckServiceability(ctx, &individualReq)
		if err != nil {
			// Handle individual request failures
			errorCount++
			errorMsg := fmt.Sprintf("Request %d failed: %v", i+1, err)

			// Add error response for this request
			responses = append(responses, models.ServiceabilityResponse{
				Success: false,
				Data:    &models.ServiceabilityData{},
				Error: &models.ErrorResponse{
					Code:    "REQUEST_FAILED",
					Message: errorMsg,
				},
			})
		} else {
			responses = append(responses, *response)
		}
	}

	bulkResponse := &models.BulkServiceabilityResponse{
		Success: errorCount == 0,
		Data:    responses,
	}

	if errorCount > 0 {
		bulkResponse.Error = &models.ErrorResponse{
			Code:    "PARTIAL_FAILURE",
			Message: fmt.Sprintf("Processed %d requests with %d errors", len(req.Requests), errorCount),
		}
	}

	return bulkResponse, nil
}

// validateServiceabilityRequest validates the input request
func (s *serviceabilityOrchestrator) validateServiceabilityRequest(req *models.ServiceabilityCheckRequest) error {
	if req.CountryCode == "" {
		return fmt.Errorf("country code is required")
	}

	// Must have either single postal code or both pickup and delivery
	hasPostalCode := req.PostalCode != nil && *req.PostalCode != ""
	hasPickupDelivery := req.PickupPostalCode != nil && *req.PickupPostalCode != "" &&
		req.DeliveryPostalCode != nil && *req.DeliveryPostalCode != ""

	if !hasPostalCode && !hasPickupDelivery {
		return fmt.Errorf("either postal_code or both pickup_postal_code and delivery_postal_code must be provided")
	}

	if hasPostalCode && hasPickupDelivery {
		return fmt.Errorf("cannot specify both postal_code and pickup/delivery postal codes")
	}

	return nil
}

// combinePartnerCapabilities finds partners that serve both pickup and delivery locations
func (s *serviceabilityOrchestrator) combinePartnerCapabilities(pickup, delivery []interfaces.PartnerCapability) []interfaces.PartnerCapability {
	partnerMap := make(map[uint]interfaces.PartnerCapability)

	// Index pickup partners
	for _, p := range pickup {
		partnerMap[p.PartnerID] = p
	}

	result := make([]interfaces.PartnerCapability, 0)

	// Find partners that exist in both pickup and delivery
	for _, d := range delivery {
		if p, exists := partnerMap[d.PartnerID]; exists {
			// Combine capabilities (intersection of services)
			combined := interfaces.PartnerCapability{
				PartnerID:   p.PartnerID,
				PartnerName: p.PartnerName,
				IsActive:    p.IsActive && d.IsActive,
			}

			// Find intersection of capabilities
			combined.ServiceTypes = s.intersectStrings(p.ServiceTypes, d.ServiceTypes)
			combined.ParcelCategories = s.intersectStrings(p.ParcelCategories, d.ParcelCategories)
			combined.OperationTypes = s.intersectStrings(p.OperationTypes, d.OperationTypes)
			combined.PaymentModes = s.intersectStrings(p.PaymentModes, d.PaymentModes)
			combined.DeliveryModes = s.intersectStrings(p.DeliveryModes, d.DeliveryModes)

			result = append(result, combined)
		}
	}

	return result
}

// intersectStrings finds common elements between two string slices
func (s *serviceabilityOrchestrator) intersectStrings(a, b []string) []string {
	set := make(map[string]bool)
	for _, item := range a {
		set[item] = true
	}

	result := make([]string, 0)
	for _, item := range b {
		if set[item] {
			result = append(result, item)
		}
	}

	return result
}

// getRelevantServiceDefinitions fetches service definitions based on request filters
func (s *serviceabilityOrchestrator) getRelevantServiceDefinitions(ctx context.Context, req *models.ServiceabilityCheckRequest) ([]interfaces.ServiceDefinition, error) {
	// For now, get all available service definitions
	// In future, this could be optimized to fetch only relevant ones based on filters

	parcelCategories, err := s.serviceDefinitionResolver.GetParcelCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get parcel categories: %w", err)
	}

	definitions := make([]interfaces.ServiceDefinition, 0)

	for _, category := range parcelCategories {
		serviceTypes, err := s.serviceDefinitionResolver.GetServiceTypes(ctx, category)
		if err != nil {
			continue // Skip this category if we can't get service types
		}

		for _, serviceType := range serviceTypes {
			definition, err := s.serviceDefinitionResolver.GetServiceDefinition(ctx, serviceType)
			if err != nil {
				continue // Skip this service type if we can't get definition
			}
			definitions = append(definitions, *definition)
		}
	}

	return definitions, nil
}

// buildServiceFilters creates service filters from the request
func (s *serviceabilityOrchestrator) buildServiceFilters(req *models.ServiceabilityCheckRequest) *interfaces.ServiceFilters {
	filters := &interfaces.ServiceFilters{}

	// Add filters based on request parameters
	if req.ServiceTypeCode != nil && *req.ServiceTypeCode != "" {
		filters.ServiceTypes = []string{*req.ServiceTypeCode}
	}

	if req.ParcelCategory != nil && *req.ParcelCategory != "" {
		filters.ParcelCategories = []string{*req.ParcelCategory}
	}

	// Could add more filters based on request parameters in the future

	return filters
}
