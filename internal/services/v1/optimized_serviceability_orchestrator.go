package services

import (
	"context"
	"fmt"
	"sync"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// OptimizedServiceabilityOrchestrator implements an efficient version of serviceability checking
// that minimizes external service calls and optimizes data fetching
type OptimizedServiceabilityOrchestrator struct {
	locationResolver interfaces.LocationResolver
	partnerService   interfaces.PartnerServiceClient
	specService      interfaces.SpecificationServiceClient
	calculator       interfaces.ServiceabilityCalculator
}

// NewOptimizedServiceabilityOrchestrator creates a new optimized orchestrator
func NewOptimizedServiceabilityOrchestrator(
	locationResolver interfaces.LocationResolver,
	partnerService interfaces.PartnerServiceClient,
	specService interfaces.SpecificationServiceClient,
	calculator interfaces.ServiceabilityCalculator,
) interfaces.ServiceabilityOrchestrator {
	return &OptimizedServiceabilityOrchestrator{
		locationResolver: locationResolver,
		partnerService:   partnerService,
		specService:      specService,
		calculator:       calculator,
	}
}

// CheckServiceability implements the optimized single location serviceability check
func (o *OptimizedServiceabilityOrchestrator) CheckServiceability(ctx context.Context, req *models.ServiceabilityCheckRequest) (*models.ServiceabilityResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if err := o.validateServiceabilityRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Step 1: Extract postal codes and validate locations
	postalCodes, locationHierarchies, err := o.extractAndValidateLocations(ctx, req)
	if err != nil {
		return o.buildErrorResponse("POSTAL_CODE_INACTIVE", err.Error()), nil
	}

	// Step 2: Get all partners for the postal codes in a single optimized call
	allPartners, err := o.getPartnersForPostalCodes(ctx, postalCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get partners: %w", err)
	}

	// Step 3: Get all required specifications in a single call
	serviceDefinitions, err := o.getOptimizedServiceDefinitions(ctx, req, allPartners)
	if err != nil {
		return nil, fmt.Errorf("failed to get service definitions: %w", err)
	}

	// Step 4: Build partner capabilities from the collected data
	partnerCapabilities := o.buildPartnerCapabilities(allPartners, serviceDefinitions, req)

	// Step 5: Calculate serviceability
	calculationRequest := &interfaces.ServiceabilityCalculationRequest{
		LocationHierarchy:   locationHierarchies.single,
		PickupHierarchy:     locationHierarchies.pickup,
		DeliveryHierarchy:   locationHierarchies.delivery,
		PartnerCapabilities: partnerCapabilities,
		ServiceDefinitions:  serviceDefinitions,
		Filters:             o.buildServiceFilters(req),
	}

	serviceabilityData, err := o.calculator.CalculateServiceability(ctx, calculationRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate serviceability: %w", err)
	}

	return &models.ServiceabilityResponse{
		Success: true,
		Data:    serviceabilityData,
	}, nil
}

// BulkCheckServiceability implements optimized bulk processing
func (o *OptimizedServiceabilityOrchestrator) BulkCheckServiceability(ctx context.Context, req *models.BulkServiceabilityRequest) (*models.BulkServiceabilityResponse, error) {
	if req == nil || len(req.Requests) == 0 {
		return nil, fmt.Errorf("bulk request cannot be nil or empty")
	}

	// Step 1: Extract all unique postal codes from all requests
	allPostalCodes := o.extractAllPostalCodes(req.Requests)

	// Step 2: Validate all postal codes in parallel
	validPostalCodes, err := o.validatePostalCodesBatch(ctx, allPostalCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to validate postal codes: %w", err)
	}

	// Step 3: Get all partners for all postal codes in one optimized call
	allPartners, err := o.getPartnersForPostalCodesBatch(ctx, validPostalCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get partners: %w", err)
	}

	// Step 4: Get all service definitions once
	serviceDefinitions, err := o.getAllServiceDefinitionsOptimized(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get service definitions: %w", err)
	}

	// Step 5: Process all requests with pre-fetched data
	responses := make([]models.ServiceabilityResponse, 0, len(req.Requests))
	errorCount := 0

	for i, individualReq := range req.Requests {
		response, err := o.processIndividualRequestWithCache(ctx, &individualReq, allPartners, serviceDefinitions, validPostalCodes)
		if err != nil {
			errorCount++
			responses = append(responses, models.ServiceabilityResponse{
				Success: false,
				Data:    &models.ServiceabilityData{},
				Error: &models.ErrorResponse{
					Code:    "REQUEST_FAILED",
					Message: fmt.Sprintf("Request %d failed: %v", i+1, err),
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

// LocationHierarchies holds the different location hierarchies
type LocationHierarchies struct {
	single   *models.LocationHierarchy
	pickup   *models.LocationHierarchy
	delivery *models.LocationHierarchy
}

// extractAndValidateLocations extracts postal codes and validates locations
func (o *OptimizedServiceabilityOrchestrator) extractAndValidateLocations(ctx context.Context, req *models.ServiceabilityCheckRequest) ([]string, *LocationHierarchies, error) {
	var postalCodes []string
	locationHierarchies := &LocationHierarchies{}

	if req.PostalCode != nil && *req.PostalCode != "" {
		// Single location query
		postalCodes = []string{*req.PostalCode}

		// Validate postal code is active
		isActive, err := o.locationResolver.IsPostalCodeActive(ctx, *req.PostalCode, req.CountryCode)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to validate postal code %s: %w", *req.PostalCode, err)
		}
		if !isActive {
			return nil, nil, fmt.Errorf("postal code %s is not active", *req.PostalCode)
		}

		// Get location hierarchy
		locationHierarchies.single, err = o.locationResolver.GetLocationHierarchy(ctx, *req.PostalCode, req.CountryCode)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve location for postal code %s: %w", *req.PostalCode, err)
		}

	} else if req.PickupPostalCode != nil && req.DeliveryPostalCode != nil {
		// Origin-destination query
		postalCodes = []string{*req.PickupPostalCode, *req.DeliveryPostalCode}

		// Validate both postal codes are active in parallel
		var wg sync.WaitGroup
		var pickupErr, deliveryErr error
		var isPickupActive, isDeliveryActive bool

		wg.Add(2)
		go func() {
			defer wg.Done()
			isPickupActive, pickupErr = o.locationResolver.IsPostalCodeActive(ctx, *req.PickupPostalCode, req.CountryCode)
		}()
		go func() {
			defer wg.Done()
			isDeliveryActive, deliveryErr = o.locationResolver.IsPostalCodeActive(ctx, *req.DeliveryPostalCode, req.CountryCode)
		}()
		wg.Wait()

		if pickupErr != nil {
			return nil, nil, fmt.Errorf("failed to validate pickup postal code %s: %w", *req.PickupPostalCode, pickupErr)
		}
		if deliveryErr != nil {
			return nil, nil, fmt.Errorf("failed to validate delivery postal code %s: %w", *req.DeliveryPostalCode, deliveryErr)
		}

		if !isPickupActive || !isDeliveryActive {
			return nil, nil, fmt.Errorf("one or more postal codes are not active")
		}

		// Get location hierarchies in parallel
		wg.Add(2)
		go func() {
			defer wg.Done()
			locationHierarchies.pickup, pickupErr = o.locationResolver.GetLocationHierarchy(ctx, *req.PickupPostalCode, req.CountryCode)
		}()
		go func() {
			defer wg.Done()
			locationHierarchies.delivery, deliveryErr = o.locationResolver.GetLocationHierarchy(ctx, *req.DeliveryPostalCode, req.CountryCode)
		}()
		wg.Wait()

		if pickupErr != nil {
			return nil, nil, fmt.Errorf("failed to resolve pickup location: %w", pickupErr)
		}
		if deliveryErr != nil {
			return nil, nil, fmt.Errorf("failed to resolve delivery location: %w", deliveryErr)
		}
	}

	return postalCodes, locationHierarchies, nil
}

// getPartnersForPostalCodes efficiently fetches partners for multiple postal codes
func (o *OptimizedServiceabilityOrchestrator) getPartnersForPostalCodes(ctx context.Context, postalCodes []string) (map[string][]interfaces.PartnerInfo, error) {
	partnerMap := make(map[string][]interfaces.PartnerInfo)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errorChan = make(chan error, len(postalCodes))

	// Fetch partners for each postal code in parallel
	for _, postalCode := range postalCodes {
		wg.Add(1)
		go func(pc string) {
			defer wg.Done()
			partners, err := o.partnerService.GetPartnersByLocation(ctx, "postal_code", pc)
			if err != nil {
				errorChan <- fmt.Errorf("failed to get partners for postal code %s: %w", pc, err)
				return
			}

			mu.Lock()
			partnerMap[pc] = partners
			mu.Unlock()
		}(postalCode)
	}

	wg.Wait()
	close(errorChan)

	// Check for errors
	if len(errorChan) > 0 {
		return nil, <-errorChan
	}

	return partnerMap, nil
}

// getOptimizedServiceDefinitions fetches service definitions efficiently
func (o *OptimizedServiceabilityOrchestrator) getOptimizedServiceDefinitions(ctx context.Context, req *models.ServiceabilityCheckRequest, partners map[string][]interfaces.PartnerInfo) ([]interfaces.ServiceDefinition, error) {
	// For now, always use the optimized approach that gets all catalogs
	// This could be enhanced in the future to filter based on request parameters
	return o.getAllServiceDefinitionsOptimized(ctx)
}

// getAllServiceDefinitionsOptimized fetches all service definitions efficiently
func (o *OptimizedServiceabilityOrchestrator) getAllServiceDefinitionsOptimized(ctx context.Context) ([]interfaces.ServiceDefinition, error) {
	// Get all catalogs in one call
	catalogs, err := o.specService.GetCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalogs: %w", err)
	}

	// Extract service definitions from catalogs
	var serviceDefinitions []interfaces.ServiceDefinition

	for _, catalog := range catalogs {
		for _, item := range catalog.CatalogItems {
			serviceDef := interfaces.ServiceDefinition{
				ServiceType:       item.Code,
				ParcelCategory:    "standard", // Default, could be enhanced
				DefaultOperations: []string{"pickup", "delivery"},
				DefaultPayments:   []string{"cod", "prepaid"},
				DefaultDelivery:   []string{"standard"},
				Description:       item.Description,
			}
			serviceDefinitions = append(serviceDefinitions, serviceDef)
		}
	}

	return serviceDefinitions, nil
}

// buildPartnerCapabilities efficiently builds partner capabilities from collected data
func (o *OptimizedServiceabilityOrchestrator) buildPartnerCapabilities(
	partners map[string][]interfaces.PartnerInfo,
	serviceDefinitions []interfaces.ServiceDefinition,
	req *models.ServiceabilityCheckRequest,
) []interfaces.PartnerCapability {

	var allCapabilities []interfaces.PartnerCapability
	partnerSet := make(map[uint]bool) // Track unique partners

	// Combine partners from all postal codes
	for _, partnerList := range partners {
		for _, partner := range partnerList {
			if partnerSet[partner.ID] || !partner.IsActive {
				continue // Skip duplicates and inactive partners
			}
			partnerSet[partner.ID] = true

			capability := interfaces.PartnerCapability{
				PartnerID:        partner.ID,
				PartnerName:      partner.Name,
				IsActive:         partner.IsActive,
				ServiceTypes:     o.extractServiceTypes(serviceDefinitions),
				ParcelCategories: o.extractParcelCategories(serviceDefinitions),
				OperationTypes:   []string{"pickup", "delivery"},
				PaymentModes:     []string{"cod", "prepaid"},
				DeliveryModes:    []string{"standard"},
			}

			allCapabilities = append(allCapabilities, capability)
		}
	}

	return allCapabilities
}

// Helper methods
func (o *OptimizedServiceabilityOrchestrator) extractServiceTypes(definitions []interfaces.ServiceDefinition) []string {
	var serviceTypes []string
	seen := make(map[string]bool)

	for _, def := range definitions {
		if !seen[def.ServiceType] {
			serviceTypes = append(serviceTypes, def.ServiceType)
			seen[def.ServiceType] = true
		}
	}

	return serviceTypes
}

func (o *OptimizedServiceabilityOrchestrator) extractParcelCategories(definitions []interfaces.ServiceDefinition) []string {
	var categories []string
	seen := make(map[string]bool)

	for _, def := range definitions {
		if !seen[def.ParcelCategory] {
			categories = append(categories, def.ParcelCategory)
			seen[def.ParcelCategory] = true
		}
	}

	return categories
}

func (o *OptimizedServiceabilityOrchestrator) extractAllPostalCodes(requests []models.ServiceabilityCheckRequest) []string {
	var postalCodes []string
	seen := make(map[string]bool)

	for _, req := range requests {
		if req.PostalCode != nil && *req.PostalCode != "" && !seen[*req.PostalCode] {
			postalCodes = append(postalCodes, *req.PostalCode)
			seen[*req.PostalCode] = true
		}
		if req.PickupPostalCode != nil && *req.PickupPostalCode != "" && !seen[*req.PickupPostalCode] {
			postalCodes = append(postalCodes, *req.PickupPostalCode)
			seen[*req.PickupPostalCode] = true
		}
		if req.DeliveryPostalCode != nil && *req.DeliveryPostalCode != "" && !seen[*req.DeliveryPostalCode] {
			postalCodes = append(postalCodes, *req.DeliveryPostalCode)
			seen[*req.DeliveryPostalCode] = true
		}
	}

	return postalCodes
}

func (o *OptimizedServiceabilityOrchestrator) validatePostalCodesBatch(ctx context.Context, postalCodes []string) (map[string]bool, error) {
	validCodes := make(map[string]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errorChan = make(chan error, len(postalCodes))

	for _, postalCode := range postalCodes {
		wg.Add(1)
		go func(pc string) {
			defer wg.Done()
			isActive, err := o.locationResolver.IsPostalCodeActive(ctx, pc, "IN") // Default country
			if err != nil {
				errorChan <- fmt.Errorf("failed to validate postal code %s: %w", pc, err)
				return
			}

			mu.Lock()
			validCodes[pc] = isActive
			mu.Unlock()
		}(postalCode)
	}

	wg.Wait()
	close(errorChan)

	if len(errorChan) > 0 {
		return nil, <-errorChan
	}

	return validCodes, nil
}

func (o *OptimizedServiceabilityOrchestrator) getPartnersForPostalCodesBatch(ctx context.Context, validPostalCodes map[string]bool) (map[string][]interfaces.PartnerInfo, error) {
	partnerMap := make(map[string][]interfaces.PartnerInfo)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errorChan = make(chan error, len(validPostalCodes))

	for postalCode, isValid := range validPostalCodes {
		if !isValid {
			continue
		}

		wg.Add(1)
		go func(pc string) {
			defer wg.Done()
			partners, err := o.partnerService.GetPartnersByLocation(ctx, "postal_code", pc)
			if err != nil {
				errorChan <- fmt.Errorf("failed to get partners for postal code %s: %w", pc, err)
				return
			}

			mu.Lock()
			partnerMap[pc] = partners
			mu.Unlock()
		}(postalCode)
	}

	wg.Wait()
	close(errorChan)

	if len(errorChan) > 0 {
		return nil, <-errorChan
	}

	return partnerMap, nil
}

func (o *OptimizedServiceabilityOrchestrator) processIndividualRequestWithCache(
	ctx context.Context,
	req *models.ServiceabilityCheckRequest,
	allPartners map[string][]interfaces.PartnerInfo,
	serviceDefinitions []interfaces.ServiceDefinition,
	validPostalCodes map[string]bool,
) (*models.ServiceabilityResponse, error) {

	// Extract postal codes for this request
	var reqPostalCodes []string
	if req.PostalCode != nil && *req.PostalCode != "" {
		reqPostalCodes = []string{*req.PostalCode}
	} else if req.PickupPostalCode != nil && req.DeliveryPostalCode != nil {
		reqPostalCodes = []string{*req.PickupPostalCode, *req.DeliveryPostalCode}
	}

	// Check if postal codes are valid
	for _, pc := range reqPostalCodes {
		if valid, exists := validPostalCodes[pc]; !exists || !valid {
			return &models.ServiceabilityResponse{
				Success: true,
				Data:    &models.ServiceabilityData{},
				Error: &models.ErrorResponse{
					Code:    "POSTAL_CODE_INACTIVE",
					Message: fmt.Sprintf("Postal code %s is not active", pc),
				},
			}, nil
		}
	}

	// Get partners for this request from cached data
	var requestPartners []interfaces.PartnerInfo
	for _, pc := range reqPostalCodes {
		if partners, exists := allPartners[pc]; exists {
			requestPartners = append(requestPartners, partners...)
		}
	}

	// Build partner capabilities
	partnerCapabilities := o.buildPartnerCapabilitiesFromList(requestPartners, serviceDefinitions)

	// Calculate serviceability with cached data
	calculationRequest := &interfaces.ServiceabilityCalculationRequest{
		PartnerCapabilities: partnerCapabilities,
		ServiceDefinitions:  serviceDefinitions,
		Filters:             o.buildServiceFilters(req),
	}

	serviceabilityData, err := o.calculator.CalculateServiceability(ctx, calculationRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate serviceability: %w", err)
	}

	return &models.ServiceabilityResponse{
		Success: true,
		Data:    serviceabilityData,
	}, nil
}

func (o *OptimizedServiceabilityOrchestrator) buildPartnerCapabilitiesFromList(
	partners []interfaces.PartnerInfo,
	serviceDefinitions []interfaces.ServiceDefinition,
) []interfaces.PartnerCapability {

	var capabilities []interfaces.PartnerCapability
	partnerSet := make(map[uint]bool)

	for _, partner := range partners {
		if partnerSet[partner.ID] || !partner.IsActive {
			continue
		}
		partnerSet[partner.ID] = true

		capability := interfaces.PartnerCapability{
			PartnerID:        partner.ID,
			PartnerName:      partner.Name,
			IsActive:         partner.IsActive,
			ServiceTypes:     o.extractServiceTypes(serviceDefinitions),
			ParcelCategories: o.extractParcelCategories(serviceDefinitions),
			OperationTypes:   []string{"pickup", "delivery"},
			PaymentModes:     []string{"cod", "prepaid"},
			DeliveryModes:    []string{"standard"},
		}

		capabilities = append(capabilities, capability)
	}

	return capabilities
}

func (o *OptimizedServiceabilityOrchestrator) validateServiceabilityRequest(req *models.ServiceabilityCheckRequest) error {
	if req.CountryCode == "" {
		return fmt.Errorf("country code is required")
	}

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

func (o *OptimizedServiceabilityOrchestrator) buildServiceFilters(req *models.ServiceabilityCheckRequest) *interfaces.ServiceFilters {
	filters := &interfaces.ServiceFilters{}

	if req.ServiceTypeCode != nil && *req.ServiceTypeCode != "" {
		filters.ServiceTypes = []string{*req.ServiceTypeCode}
	}

	if req.ParcelCategory != nil && *req.ParcelCategory != "" {
		filters.ParcelCategories = []string{*req.ParcelCategory}
	}

	return filters
}

func (o *OptimizedServiceabilityOrchestrator) buildErrorResponse(code, message string) *models.ServiceabilityResponse {
	return &models.ServiceabilityResponse{
		Success: true,
		Data:    &models.ServiceabilityData{},
		Error: &models.ErrorResponse{
			Code:    code,
			Message: message,
		},
	}
}
