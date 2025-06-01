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

	// Step 1: Validate we have partners
	if len(request.PartnerCapabilities) == 0 {
		return &models.ServiceabilityData{}, nil
	}

	// Step 2: Create service options from partner capabilities and service definitions
	serviceOptions := sc.buildServiceOptions(request)

	// Step 3: Apply filters if provided
	if request.Filters != nil {
		serviceOptions = sc.applyFilters(serviceOptions, request.Filters)
	}

	// Step 4: Apply preferences for ranking (basic implementation)
	serviceOptions = sc.applyBasicPreferences(serviceOptions)

	// Step 5: Build final response - we'll populate the location and services later
	// For now, just return basic structure
	result := &models.ServiceabilityData{}

	return result, nil
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

// buildServiceOptions creates service options from partner capabilities and definitions
func (sc *serviceabilityCalculator) buildServiceOptions(request *interfaces.ServiceabilityCalculationRequest) []interfaces.ServiceOption {
	var serviceOptions []interfaces.ServiceOption

	// Create a map to avoid duplicates
	serviceMap := make(map[string]*interfaces.ServiceOption)

	// Iterate through each partner capability
	for _, partner := range request.PartnerCapabilities {
		// Create basic service options from partner capabilities
		serviceOption := interfaces.ServiceOption{
			ServiceType:       "standard", // Default service type for catalog-based approach
			ParcelCategory:    "ecom",     // Default category
			OperationTypes:    []string{"pickup", "delivery"},
			PaymentModes:      []string{"cod", "prepaid"},
			DeliveryModes:     []string{"standard"},
			AvailablePartners: []uint{partner.PartnerID},
			Rating:            4.0, // Default rating
		}

		// Check if we already have this service type
		key := fmt.Sprintf("%s_%s", serviceOption.ServiceType, serviceOption.ParcelCategory)
		if existing, exists := serviceMap[key]; exists {
			// Add partner to existing service
			existing.AvailablePartners = append(existing.AvailablePartners, partner.PartnerID)
			// Update rating if this partner has better coverage
			if len(existing.AvailablePartners) > 1 {
				existing.Rating = 4.5 // Improve rating with more partners
			}
		} else {
			serviceMap[key] = &serviceOption
		}
	}

	// Convert map to slice
	for _, service := range serviceMap {
		serviceOptions = append(serviceOptions, *service)
	}

	return serviceOptions
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
