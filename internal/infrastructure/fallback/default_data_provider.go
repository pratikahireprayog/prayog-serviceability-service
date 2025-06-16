package fallback

import (
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// DefaultDataProvider provides default fallback data for service definitions
type DefaultDataProvider struct {
	specDefinitions    []interfaces.SpecDefinition
	catalogs           []interfaces.Catalog
	entitySpecs        map[string][]interfaces.EntitySpecification
	serviceDefinitions map[string][]interfaces.ServiceDefinition
}

// NewDefaultDataProvider creates a new default data provider with predefined fallback data
func NewDefaultDataProvider() *DefaultDataProvider {
	provider := &DefaultDataProvider{
		entitySpecs:        make(map[string][]interfaces.EntitySpecification),
		serviceDefinitions: make(map[string][]interfaces.ServiceDefinition),
	}

	provider.initializeDefaultData()
	return provider
}

// GetDefaultSpecDefinitions returns default specification definitions
func (ddp *DefaultDataProvider) GetDefaultSpecDefinitions() []interfaces.SpecDefinition {
	return ddp.specDefinitions
}

// GetDefaultCatalogs returns default catalogs
func (ddp *DefaultDataProvider) GetDefaultCatalogs() []interfaces.Catalog {
	return ddp.catalogs
}

// GetDefaultEntitySpecifications returns default entity specifications for a specific entity
func (ddp *DefaultDataProvider) GetDefaultEntitySpecifications(entityType, entityID string) []interfaces.EntitySpecification {
	key := entityType + ":" + entityID
	if specs, exists := ddp.entitySpecs[key]; exists {
		return specs
	}

	// Return generic default for the entity type
	if specs, exists := ddp.entitySpecs[entityType]; exists {
		return specs
	}

	return []interfaces.EntitySpecification{}
}

// GetDefaultServiceDefinitions returns default service definitions based on criteria
func (ddp *DefaultDataProvider) GetDefaultServiceDefinitions(serviceTypes, parcelCategories []string) []interfaces.ServiceDefinition {
	// For simplicity, return all default service definitions
	// In a real implementation, you might filter based on the criteria
	var allDefaults []interfaces.ServiceDefinition
	for _, defs := range ddp.serviceDefinitions {
		allDefaults = append(allDefaults, defs...)
	}
	return allDefaults
}

// initializeDefaultData sets up the default fallback data
func (ddp *DefaultDataProvider) initializeDefaultData() {
	// Initialize default specification definitions
	ddp.specDefinitions = []interfaces.SpecDefinition{
		{
			ID:          1,
			SpecType:    "service_type",
			Name:        "Standard Delivery",
			Description: "Standard delivery service specification",
		},
		{
			ID:          2,
			SpecType:    "parcel_category",
			Name:        "Documents",
			Description: "Document parcel category specification",
		},
		{
			ID:          3,
			SpecType:    "operation_type",
			Name:        "Pickup and Delivery",
			Description: "Standard pickup and delivery operation",
		},
	}

	// Initialize default catalogs
	ddp.catalogs = []interfaces.Catalog{
		{
			ID:          1,
			SpecDefID:   1,
			Name:        "Service Types Catalog",
			Description: "Catalog of available service types",
			CatalogItems: []interfaces.CatalogItem{
				{
					ID:          1,
					Code:        "STANDARD",
					Name:        "Standard Delivery",
					Description: "Standard delivery service",
				},
				{
					ID:          2,
					Code:        "EXPRESS",
					Name:        "Express Delivery",
					Description: "Express delivery service",
				},
			},
		},
		{
			ID:          2,
			SpecDefID:   2,
			Name:        "Parcel Categories Catalog",
			Description: "Catalog of parcel categories",
			CatalogItems: []interfaces.CatalogItem{
				{
					ID:          3,
					Code:        "DOCUMENTS",
					Name:        "Documents",
					Description: "Document parcels",
				},
				{
					ID:          4,
					Code:        "PACKAGES",
					Name:        "Packages",
					Description: "General packages",
				},
			},
		},
	}

	// Initialize default entity specifications
	ddp.entitySpecs["partner"] = []interfaces.EntitySpecification{
		{
			ID:         1,
			EntityType: "partner",
			EntityID:   "default",
			SpecDefID:  1,
		},
	}

	ddp.entitySpecs["location"] = []interfaces.EntitySpecification{
		{
			ID:         2,
			EntityType: "location",
			EntityID:   "default",
			SpecDefID:  2,
		},
	}

	// Initialize default service definitions
	ddp.serviceDefinitions["standard"] = []interfaces.ServiceDefinition{
		{
			ServiceType:       "STANDARD",
			ParcelCategory:    "DOCUMENTS",
			DefaultOperations: []string{"PICKUP", "DELIVERY"},
			DefaultPayments:   []string{"COD", "PREPAID"},
			DefaultDelivery:   []string{"DOOR_TO_DOOR"},
			Description:       "Standard document delivery service",
		},
		{
			ServiceType:       "STANDARD",
			ParcelCategory:    "PACKAGES",
			DefaultOperations: []string{"PICKUP", "DELIVERY"},
			DefaultPayments:   []string{"COD", "PREPAID"},
			DefaultDelivery:   []string{"DOOR_TO_DOOR"},
			Description:       "Standard package delivery service",
		},
	}

	ddp.serviceDefinitions["express"] = []interfaces.ServiceDefinition{
		{
			ServiceType:       "EXPRESS",
			ParcelCategory:    "DOCUMENTS",
			DefaultOperations: []string{"PICKUP", "DELIVERY", "TRACKING"},
			DefaultPayments:   []string{"PREPAID"},
			DefaultDelivery:   []string{"DOOR_TO_DOOR", "PICKUP_POINT"},
			Description:       "Express document delivery service",
		},
		{
			ServiceType:       "EXPRESS",
			ParcelCategory:    "PACKAGES",
			DefaultOperations: []string{"PICKUP", "DELIVERY", "TRACKING"},
			DefaultPayments:   []string{"PREPAID"},
			DefaultDelivery:   []string{"DOOR_TO_DOOR", "PICKUP_POINT"},
			Description:       "Express package delivery service",
		},
	}
}

// UpdateDefaultData allows updating the default data (useful for configuration changes)
func (ddp *DefaultDataProvider) UpdateDefaultData(specDefs []interfaces.SpecDefinition, catalogs []interfaces.Catalog, serviceDefs map[string][]interfaces.ServiceDefinition) {
	if specDefs != nil {
		ddp.specDefinitions = specDefs
	}
	if catalogs != nil {
		ddp.catalogs = catalogs
	}
	if serviceDefs != nil {
		ddp.serviceDefinitions = serviceDefs
	}
}

// GetAvailableEntityTypes returns the entity types for which default data is available
func (ddp *DefaultDataProvider) GetAvailableEntityTypes() []string {
	types := make([]string, 0, len(ddp.entitySpecs))
	for key := range ddp.entitySpecs {
		types = append(types, key)
	}
	return types
}

// GetAvailableServiceTypes returns the service types for which default data is available
func (ddp *DefaultDataProvider) GetAvailableServiceTypes() []string {
	types := make([]string, 0, len(ddp.serviceDefinitions))
	for key := range ddp.serviceDefinitions {
		types = append(types, key)
	}
	return types
}
