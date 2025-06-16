package interfaces

import (
	"context"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// ServiceabilityOrchestrator defines the main business logic interface for serviceability operations
type ServiceabilityOrchestrator interface {
	CheckServiceability(ctx context.Context, req *models.ServiceabilityCheckRequest) (*models.ServiceabilityResponse, error)
	BulkCheckServiceability(ctx context.Context, req *models.BulkServiceabilityRequest) (*models.BulkServiceabilityResponse, error)
}

// LocationResolver defines the interface for location validation and resolution
type LocationResolver interface {
	ValidatePostalCode(ctx context.Context, postalCode, countryCode string) (*models.LocationValidationResult, error)
	GetLocationHierarchy(ctx context.Context, postalCode, countryCode string) (*models.LocationHierarchy, error)
	IsPostalCodeActive(ctx context.Context, postalCode, countryCode string) (bool, error)
}

// PartnerCapabilityAggregator defines the interface for fetching partner capabilities
type PartnerCapabilityAggregator interface {
	GetPartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy) ([]PartnerCapability, error)
	GetPartnerCapabilities(ctx context.Context, partnerID uint, serviceFilters *ServiceFilters) (*PartnerServiceCapabilities, error)
}

// ServiceDefinitionResolver defines the interface for fetching service definitions
type ServiceDefinitionResolver interface {
	GetParcelCategories(ctx context.Context) ([]string, error)
	GetServiceTypes(ctx context.Context, parcelCategory string) ([]string, error)
	GetOperationTypes(ctx context.Context) ([]string, error)
	GetPaymentModes(ctx context.Context) ([]string, error)
	GetDeliveryModes(ctx context.Context) ([]string, error)
	GetServiceDefinition(ctx context.Context, serviceType string) (*ServiceDefinition, error)
}

// ServiceabilityCalculator defines the interface for business logic calculations
type ServiceabilityCalculator interface {
	CalculateServiceability(ctx context.Context, request *ServiceabilityCalculationRequest) (*models.ServiceabilityData, error)
	FilterServicesByPreferences(ctx context.Context, services []ServiceOption, preferences []PartnerPreference) ([]ServiceOption, error)
	CombineRouteCapabilities(ctx context.Context, pickupServices, deliveryServices []ServiceOption) ([]ServiceOption, error)
}

// External Service Client Interfaces

// PartnerServiceClient defines the interface for Partner Service integration
type PartnerServiceClient interface {
	GetPartnersByLocation(ctx context.Context, locationType, locationID string) ([]PartnerInfo, error)
	GetPartnerEffectiveDetails(ctx context.Context, partnerID uint, entityType, entityID string) (*PartnerEffectiveDetails, error)
}

// SpecificationServiceClient defines the interface for Specification Service integration
type SpecificationServiceClient interface {
	GetSpecDefinitions(ctx context.Context) ([]SpecDefinition, error)
	GetCatalogs(ctx context.Context) ([]Catalog, error)
	GetEntitySpecifications(ctx context.Context, entityType, entityID string) ([]EntitySpecification, error)
}

// Data Transfer Objects for Interfaces

// PartnerCapability represents a partner's service capability
type PartnerCapability struct {
	PartnerID        uint     `json:"partner_id"`
	PartnerName      string   `json:"partner_name"`
	IsActive         bool     `json:"is_active"`
	ServiceTypes     []string `json:"service_types"`
	ParcelCategories []string `json:"parcel_categories"`
	OperationTypes   []string `json:"operation_types"`
	PaymentModes     []string `json:"payment_modes"`
	DeliveryModes    []string `json:"delivery_modes"`
}

// PartnerServiceCapabilities represents detailed service capabilities for a partner
type PartnerServiceCapabilities struct {
	PartnerID    uint                `json:"partner_id"`
	Capabilities []ServiceCapability `json:"capabilities"`
	Preferences  []PartnerPreference `json:"preferences"`
}

// ServiceCapability represents a specific service capability
type ServiceCapability struct {
	ServiceType    string   `json:"service_type"`
	ParcelCategory string   `json:"parcel_category"`
	OperationTypes []string `json:"operation_types"`
	PaymentModes   []string `json:"payment_modes"`
	DeliveryModes  []string `json:"delivery_modes"`
	IsActive       bool     `json:"is_active"`
	Rating         float64  `json:"rating"`
}

// PartnerPreference represents partner preferences for specific services
type PartnerPreference struct {
	ServiceType     string  `json:"service_type"`
	ParcelCategory  string  `json:"parcel_category"`
	IsPreferred     bool    `json:"is_preferred"`
	Priority        int     `json:"priority"`
	EffectiveRating float64 `json:"effective_rating"`
}

// ServiceFilters represents filters for service queries
type ServiceFilters struct {
	ServiceTypes     []string `json:"service_types,omitempty"`
	ParcelCategories []string `json:"parcel_categories,omitempty"`
	OperationTypes   []string `json:"operation_types,omitempty"`
}

// ServiceDefinition represents a service type definition from Specification Service
type ServiceDefinition struct {
	ServiceType       string   `json:"service_type"`
	ParcelCategory    string   `json:"parcel_category"`
	DefaultOperations []string `json:"default_operations"`
	DefaultPayments   []string `json:"default_payments"`
	DefaultDelivery   []string `json:"default_delivery"`
	Description       string   `json:"description"`
}

// ServiceabilityCalculationRequest represents input for serviceability calculation
type ServiceabilityCalculationRequest struct {
	LocationHierarchy   *models.LocationHierarchy `json:"location_hierarchy,omitempty"`
	PickupHierarchy     *models.LocationHierarchy `json:"pickup_hierarchy,omitempty"`
	DeliveryHierarchy   *models.LocationHierarchy `json:"delivery_hierarchy,omitempty"`
	PartnerCapabilities []PartnerCapability       `json:"partner_capabilities"`
	ServiceDefinitions  []ServiceDefinition       `json:"service_definitions"`
	Filters             *ServiceFilters           `json:"filters,omitempty"`
}

// ServiceOption represents a calculated service option
type ServiceOption struct {
	ServiceType       string   `json:"service_type"`
	ParcelCategory    string   `json:"parcel_category"`
	OperationTypes    []string `json:"operation_types"`
	PaymentModes      []string `json:"payment_modes"`
	DeliveryModes     []string `json:"delivery_modes"`
	AvailablePartners []uint   `json:"available_partners"`
	Rating            float64  `json:"rating"`
}

// External Service Data Types

// PartnerInfo represents basic partner information from Partner Service
type PartnerInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	Type     string `json:"type"`
}

// PartnerEffectiveDetails represents partner effective details with preferences
type PartnerEffectiveDetails struct {
	PartnerID   uint                  `json:"partner_id"`
	Preferences []EffectivePreference `json:"preferences"`
	Ratings     []EffectiveRating     `json:"ratings"`
}

// EffectivePreference represents an effective preference from Partner Service
type EffectivePreference struct {
	EntityType      string  `json:"entity_type"`
	EntityID        string  `json:"entity_id"`
	PreferenceValue string  `json:"preference_value"`
	EffectiveRating float64 `json:"effective_rating"`
}

// EffectiveRating represents an effective rating from Partner Service
type EffectiveRating struct {
	EntityType string  `json:"entity_type"`
	EntityID   string  `json:"entity_id"`
	Rating     float64 `json:"rating"`
}

// SpecDefinition represents a specification definition from Specification Service
type SpecDefinition struct {
	ID          uint   `json:"id"`
	SpecType    string `json:"spec_type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Catalog represents a catalog from Specification Service
type Catalog struct {
	ID           uint          `json:"id"`
	SpecDefID    uint          `json:"spec_def_id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	CatalogItems []CatalogItem `json:"catalog_items"`
}

// CatalogItem represents an item in a catalog
type CatalogItem struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// EntitySpecification represents an entity specification from Specification Service
type EntitySpecification struct {
	ID          uint   `json:"id"`
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	SpecDefID   uint   `json:"spec_def_id"`
	SpecValue   string `json:"spec_value"`
	Description string `json:"description"`
}
