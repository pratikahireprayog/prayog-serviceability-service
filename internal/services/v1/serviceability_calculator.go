package services

import (
	"context"
	"fmt"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// serviceabilityCalculator implements the ServiceabilityCalculator interface
type serviceabilityCalculator struct{}

// NewServiceabilityCalculator creates a new ServiceabilityCalculator instance
func NewServiceabilityCalculator() interfaces.ServiceabilityCalculator {
	return &serviceabilityCalculator{}
}

// CalculateServiceability implements the core business logic for serviceability calculation
func (sc *serviceabilityCalculator) CalculateServiceability(ctx context.Context, request *interfaces.ServiceabilityCalculationRequest) (*models.ServiceabilityData, error) {
	if request == nil {
		return nil, fmt.Errorf("calculation request cannot be nil")
	}

	// Step 1: Determine query type and build basic response structure
	result := &models.ServiceabilityData{}

	// Determine query type based on what locations we have
	if request.LocationHierarchy != nil {
		// Single location query
		result.QueryType = "generic_location"
		result.Location = sc.buildLocationData(request.LocationHierarchy, request.PartnerCapabilities, request.ServiceDefinitions, request.Filters)
	} else if request.PickupHierarchy != nil && request.DeliveryHierarchy != nil {
		// Origin-destination query
		result.QueryType = "point_to_point"
		result.PickupLocation = sc.buildLocationData(request.PickupHierarchy, request.PartnerCapabilities, request.ServiceDefinitions, request.Filters)
		result.DeliveryLocation = sc.buildLocationData(request.DeliveryHierarchy, request.PartnerCapabilities, request.ServiceDefinitions, request.Filters)
	} else {
		return nil, fmt.Errorf("invalid request: missing location hierarchies")
	}

	return result, nil
}

// buildLocationData creates location data with serviceability information
func (sc *serviceabilityCalculator) buildLocationData(
	locationHierarchy *models.LocationHierarchy,
	partnerCapabilities []interfaces.PartnerCapability,
	serviceDefinitions []interfaces.ServiceDefinition,
	filters *interfaces.ServiceFilters,
) *models.LocationData {
	if locationHierarchy == nil {
		return nil
	}

	// Step 1: Create service options from partner capabilities
	serviceOptions := sc.buildServiceOptionsForLocation(partnerCapabilities)

	// Step 2: Apply filters if provided
	if filters != nil {
		serviceOptions = sc.applyFilters(serviceOptions, filters)
	}

	// Step 3: Apply basic preferences for ranking
	serviceOptions = sc.applyBasicPreferences(serviceOptions)

	// Step 4: Group services by parcel category
	parcelCategoryServices := sc.groupServicesByParcelCategory(serviceOptions)

	// Step 5: Build location data
	locationData := &models.LocationData{
		PostalCode:     locationHierarchy.PostalCode,
		CountryCode:    locationHierarchy.CountryCode,
		Serviceability: parcelCategoryServices,
	}

	return locationData
}

// buildServiceOptionsForLocation creates service options for a specific location
func (sc *serviceabilityCalculator) buildServiceOptionsForLocation(partnerCapabilities []interfaces.PartnerCapability) []interfaces.ServiceOption {
	if len(partnerCapabilities) == 0 {
		return []interfaces.ServiceOption{}
	}

	// Create basic service options - for now we'll create standard services
	// In a real implementation, this would be based on actual partner capabilities
	serviceOptions := []interfaces.ServiceOption{
		{
			ServiceType:       "Standard",
			ParcelCategory:    "ecom",
			OperationTypes:    []string{"pickup", "delivery"},
			PaymentModes:      []string{"COD", "ONLINE"},
			DeliveryModes:     []string{"SURFACE"},
			AvailablePartners: sc.extractPartnerIDs(partnerCapabilities),
			Rating:            4.0,
		},
		{
			ServiceType:       "Express",
			ParcelCategory:    "ecom",
			OperationTypes:    []string{"pickup", "delivery"},
			PaymentModes:      []string{"COD", "ONLINE"},
			DeliveryModes:     []string{"AIR"},
			AvailablePartners: sc.extractPartnerIDs(partnerCapabilities),
			Rating:            4.2,
		},
	}

	// Add courier services if we have partners
	if len(partnerCapabilities) > 0 {
		serviceOptions = append(serviceOptions, interfaces.ServiceOption{
			ServiceType:       "SDD",
			ParcelCategory:    "courier",
			OperationTypes:    []string{"pickup", "delivery"},
			PaymentModes:      []string{"COD", "ONLINE"},
			DeliveryModes:     []string{"SURFACE"},
			AvailablePartners: sc.extractPartnerIDs(partnerCapabilities),
			Rating:            4.5,
		})
	}

	return serviceOptions
}

// groupServicesByParcelCategory groups service options by parcel category
func (sc *serviceabilityCalculator) groupServicesByParcelCategory(serviceOptions []interfaces.ServiceOption) []models.ParcelCategoryService {
	categoryMap := make(map[string][]models.Service)

	// Group services by category
	for _, option := range serviceOptions {
		service := models.Service{
			ServiceType:    option.ServiceType,
			OperationTypes: option.OperationTypes,
			PaymentModes:   option.PaymentModes,
			DeliveryModes:  option.DeliveryModes,
		}

		categoryMap[option.ParcelCategory] = append(categoryMap[option.ParcelCategory], service)
	}

	// Convert map to slice
	var parcelCategoryServices []models.ParcelCategoryService
	for category, services := range categoryMap {
		parcelCategoryServices = append(parcelCategoryServices, models.ParcelCategoryService{
			ParcelCategory: category,
			Services:       services,
		})
	}

	return parcelCategoryServices
}

// extractPartnerIDs extracts partner IDs from capabilities
func (sc *serviceabilityCalculator) extractPartnerIDs(capabilities []interfaces.PartnerCapability) []uint {
	var partnerIDs []uint
	for _, capability := range capabilities {
		partnerIDs = append(partnerIDs, capability.PartnerID)
	}
	return partnerIDs
}

// FilterServicesByPreferences applies partner preferences to service options
func (sc *serviceabilityCalculator) FilterServicesByPreferences(ctx context.Context, services []interfaces.ServiceOption, preferences []interfaces.PartnerPreference) ([]interfaces.ServiceOption, error) {
	if len(preferences) == 0 {
		return services, nil
	}

	var filteredServices []interfaces.ServiceOption

	for _, service := range services {
		// Check if this service matches any preferences
		for _, pref := range preferences {
			if sc.serviceMatchesPreference(service, pref) {
				// Apply preference rating if preferred
				if pref.IsPreferred {
					service.Rating = pref.EffectiveRating
				}
				filteredServices = append(filteredServices, service)
				break
			}
		}
	}

	return filteredServices, nil
}

// CombineRouteCapabilities combines pickup and delivery services for route-based queries
func (sc *serviceabilityCalculator) CombineRouteCapabilities(ctx context.Context, pickupServices, deliveryServices []interfaces.ServiceOption) ([]interfaces.ServiceOption, error) {
	if len(pickupServices) == 0 || len(deliveryServices) == 0 {
		return []interfaces.ServiceOption{}, nil
	}

	var combinedServices []interfaces.ServiceOption

	// Find common services available at both pickup and delivery locations
	for _, pickupService := range pickupServices {
		for _, deliveryService := range deliveryServices {
			if sc.servicesAreCompatible(pickupService, deliveryService) {
				// Create combined service with intersection of capabilities
				combinedService := sc.combineServices(pickupService, deliveryService)
				combinedServices = append(combinedServices, combinedService)
			}
		}
	}

	return combinedServices, nil
}

// applyFilters applies service filters to the options
func (sc *serviceabilityCalculator) applyFilters(services []interfaces.ServiceOption, filters *interfaces.ServiceFilters) []interfaces.ServiceOption {
	var filteredServices []interfaces.ServiceOption

	for _, service := range services {
		if sc.serviceMatchesFilters(service, filters) {
			filteredServices = append(filteredServices, service)
		}
	}

	return filteredServices
}

// serviceMatchesFilters checks if a service matches the provided filters
func (sc *serviceabilityCalculator) serviceMatchesFilters(service interfaces.ServiceOption, filters *interfaces.ServiceFilters) bool {
	// Check service types
	if len(filters.ServiceTypes) > 0 {
		found := false
		for _, filterType := range filters.ServiceTypes {
			if service.ServiceType == filterType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check parcel categories
	if len(filters.ParcelCategories) > 0 {
		found := false
		for _, filterCategory := range filters.ParcelCategories {
			if service.ParcelCategory == filterCategory {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check operation types
	if len(filters.OperationTypes) > 0 {
		found := false
		for _, filterOp := range filters.OperationTypes {
			for _, serviceOp := range service.OperationTypes {
				if serviceOp == filterOp {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// applyBasicPreferences applies basic preference logic (simplified for catalog-based approach)
func (sc *serviceabilityCalculator) applyBasicPreferences(services []interfaces.ServiceOption) []interfaces.ServiceOption {
	// For catalog-based approach, we'll just ensure services with more partners get higher ratings
	for i := range services {
		partnerCount := len(services[i].AvailablePartners)
		if partnerCount > 2 {
			services[i].Rating = 4.8
		} else if partnerCount > 1 {
			services[i].Rating = 4.5
		}
	}

	return services
}

// serviceMatchesPreference checks if a service option matches a partner preference
func (sc *serviceabilityCalculator) serviceMatchesPreference(service interfaces.ServiceOption, pref interfaces.PartnerPreference) bool {
	return service.ServiceType == pref.ServiceType && service.ParcelCategory == pref.ParcelCategory
}

// servicesAreCompatible checks if two services can be combined for route queries
func (sc *serviceabilityCalculator) servicesAreCompatible(pickupService, deliveryService interfaces.ServiceOption) bool {
	// Services are compatible if they have the same service type and parcel category
	return pickupService.ServiceType == deliveryService.ServiceType &&
		pickupService.ParcelCategory == deliveryService.ParcelCategory
}

// combineServices creates a combined service from pickup and delivery services
func (sc *serviceabilityCalculator) combineServices(pickupService, deliveryService interfaces.ServiceOption) interfaces.ServiceOption {
	// Combine available partners (intersection)
	partnerMap := make(map[uint]bool)
	for _, partner := range pickupService.AvailablePartners {
		partnerMap[partner] = true
	}

	var combinedPartners []uint
	for _, partner := range deliveryService.AvailablePartners {
		if partnerMap[partner] {
			combinedPartners = append(combinedPartners, partner)
		}
	}

	// Create combined service
	return interfaces.ServiceOption{
		ServiceType:       pickupService.ServiceType,
		ParcelCategory:    pickupService.ParcelCategory,
		OperationTypes:    sc.intersectOperationTypes(pickupService.OperationTypes, deliveryService.OperationTypes),
		PaymentModes:      sc.intersectPaymentModes(pickupService.PaymentModes, deliveryService.PaymentModes),
		DeliveryModes:     sc.intersectDeliveryModes(pickupService.DeliveryModes, deliveryService.DeliveryModes),
		AvailablePartners: combinedPartners,
		Rating:            (pickupService.Rating + deliveryService.Rating) / 2,
	}
}

// intersectOperationTypes finds common operation types
func (sc *serviceabilityCalculator) intersectOperationTypes(a, b []string) []string {
	return sc.intersectStringSlices(a, b)
}

// intersectPaymentModes finds common payment modes
func (sc *serviceabilityCalculator) intersectPaymentModes(a, b []string) []string {
	return sc.intersectStringSlices(a, b)
}

// intersectDeliveryModes finds common delivery modes
func (sc *serviceabilityCalculator) intersectDeliveryModes(a, b []string) []string {
	return sc.intersectStringSlices(a, b)
}

// intersectStringSlices finds intersection of two string slices
func (sc *serviceabilityCalculator) intersectStringSlices(a, b []string) []string {
	elementMap := make(map[string]bool)
	for _, item := range a {
		elementMap[item] = true
	}

	var result []string
	for _, item := range b {
		if elementMap[item] {
			result = append(result, item)
		}
	}

	return result
}
