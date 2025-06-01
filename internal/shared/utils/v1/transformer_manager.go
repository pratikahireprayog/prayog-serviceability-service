package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// PartnerDataTransformer defines the interface for transforming partner service data
type PartnerDataTransformer interface {
	TransformPartnerInfo(partnerInfo *interfaces.PartnerInfo) (*interfaces.PartnerCapability, error)
	TransformPartnerEffectiveDetails(details *interfaces.PartnerEffectiveDetails) (*interfaces.PartnerServiceCapabilities, error)
	TransformSpecificationToCapability(spec *interfaces.EntitySpecification, catalog *interfaces.Catalog) (*interfaces.ServiceCapability, error)
	ValidateAndNormalizePartnerData(capability *interfaces.PartnerCapability) error
	MapServiceTypeToParcelCategory(serviceType string) string
	ExtractOperationTypes(specValue string) []string
	ExtractPaymentModes(specValue string) []string
	ExtractDeliveryModes(specValue string) []string
}

// partnerDataTransformer implements the PartnerDataTransformer interface
type partnerDataTransformer struct {
	// Service type to parcel category mapping
	serviceTypeMapping map[string]string
	// Default values
	defaultRating float64
	timeout       time.Duration
}

// TransformerConfig holds configuration for the transformer
type TransformerConfig struct {
	DefaultRating      float64
	Timeout            time.Duration
	ServiceTypeMapping map[string]string
}

// NewPartnerDataTransformer creates a new instance of the partner data transformer
func NewPartnerDataTransformer(config *TransformerConfig) PartnerDataTransformer {
	if config == nil {
		config = &TransformerConfig{
			DefaultRating: 5.0,
			Timeout:       30 * time.Second,
		}
	}

	// Initialize default service type mappings if not provided
	if config.ServiceTypeMapping == nil {
		config.ServiceTypeMapping = map[string]string{
			"SDD":          "ecom",
			"NDD":          "ecom",
			"Standard":     "courier",
			"Express":      "courier",
			"vayuquick":    "cargo",
			"vayuquickpro": "cargo",
		}
	}

	return &partnerDataTransformer{
		serviceTypeMapping: config.ServiceTypeMapping,
		defaultRating:      config.DefaultRating,
		timeout:            config.Timeout,
	}
}

// TransformPartnerInfo transforms PartnerInfo from Partner Service to internal PartnerCapability
func (t *partnerDataTransformer) TransformPartnerInfo(partnerInfo *interfaces.PartnerInfo) (*interfaces.PartnerCapability, error) {
	if partnerInfo == nil {
		return nil, fmt.Errorf("partner info cannot be nil")
	}

	if partnerInfo.ID == 0 {
		return nil, fmt.Errorf("partner ID cannot be zero")
	}

	if strings.TrimSpace(partnerInfo.Name) == "" {
		return nil, fmt.Errorf("partner name cannot be empty")
	}

	// Create basic partner capability with default values
	capability := &interfaces.PartnerCapability{
		PartnerID:        partnerInfo.ID,
		PartnerName:      strings.TrimSpace(partnerInfo.Name),
		IsActive:         partnerInfo.IsActive,
		ServiceTypes:     []string{}, // Will be populated from specifications
		ParcelCategories: []string{}, // Will be populated from specifications
		OperationTypes:   []string{}, // Will be populated from specifications
		PaymentModes:     []string{}, // Will be populated from specifications
		DeliveryModes:    []string{}, // Will be populated from specifications
	}

	// Set default values based on partner type
	switch strings.ToUpper(partnerInfo.Type) {
	case "ECOM":
		capability.ServiceTypes = []string{"SDD", "NDD"}
		capability.ParcelCategories = []string{"ecom"}
		capability.OperationTypes = []string{"pickup", "delivery"}
	case "COURIER":
		capability.ServiceTypes = []string{"Standard", "Express"}
		capability.ParcelCategories = []string{"courier"}
		capability.OperationTypes = []string{"pickup", "delivery"}
	case "CARGO":
		capability.ServiceTypes = []string{"vayuquick", "vayuquickpro"}
		capability.ParcelCategories = []string{"cargo"}
		capability.OperationTypes = []string{"pickup", "delivery"}
	default:
		// Default to all capabilities if type is unknown
		capability.ServiceTypes = []string{"Standard"}
		capability.ParcelCategories = []string{"courier"}
		capability.OperationTypes = []string{"pickup", "delivery"}
	}

	// Set default payment and delivery modes
	capability.PaymentModes = []string{"ONLINE", "COD"}
	capability.DeliveryModes = []string{"AIR", "SURFACE"}

	return capability, nil
}

// TransformPartnerEffectiveDetails transforms PartnerEffectiveDetails to PartnerServiceCapabilities
func (t *partnerDataTransformer) TransformPartnerEffectiveDetails(details *interfaces.PartnerEffectiveDetails) (*interfaces.PartnerServiceCapabilities, error) {
	if details == nil {
		return nil, fmt.Errorf("partner effective details cannot be nil")
	}

	if details.PartnerID == 0 {
		return nil, fmt.Errorf("partner ID cannot be zero")
	}

	capabilities := &interfaces.PartnerServiceCapabilities{
		PartnerID:    details.PartnerID,
		Capabilities: []interfaces.ServiceCapability{},
		Preferences:  []interfaces.PartnerPreference{},
	}

	// Transform preferences
	for _, pref := range details.Preferences {
		preference := interfaces.PartnerPreference{
			ServiceType:     t.extractServiceType(pref.PreferenceValue),
			ParcelCategory:  t.MapServiceTypeToParcelCategory(t.extractServiceType(pref.PreferenceValue)),
			IsPreferred:     t.isPreferredValue(pref.PreferenceValue),
			Priority:        t.calculatePriority(pref.EffectiveRating),
			EffectiveRating: pref.EffectiveRating,
		}
		capabilities.Preferences = append(capabilities.Preferences, preference)
	}

	// Transform ratings into service capabilities
	for _, rating := range details.Ratings {
		serviceType := t.extractServiceType(rating.EntityID)
		capability := interfaces.ServiceCapability{
			ServiceType:    serviceType,
			ParcelCategory: t.MapServiceTypeToParcelCategory(serviceType),
			OperationTypes: []string{"pickup", "delivery"}, // Default operations
			PaymentModes:   []string{"ONLINE", "COD"},      // Default payment modes
			DeliveryModes:  []string{"AIR", "SURFACE"},     // Default delivery modes
			IsActive:       rating.Rating > 0,              // Active if has rating
			Rating:         rating.Rating,
		}
		capabilities.Capabilities = append(capabilities.Capabilities, capability)
	}

	// If no capabilities derived, create default ones
	if len(capabilities.Capabilities) == 0 {
		defaultCapability := interfaces.ServiceCapability{
			ServiceType:    "Standard",
			ParcelCategory: "courier",
			OperationTypes: []string{"pickup", "delivery"},
			PaymentModes:   []string{"ONLINE", "COD"},
			DeliveryModes:  []string{"AIR", "SURFACE"},
			IsActive:       true,
			Rating:         t.defaultRating,
		}
		capabilities.Capabilities = append(capabilities.Capabilities, defaultCapability)
	}

	return capabilities, nil
}

// TransformSpecificationToCapability transforms entity specification to service capability
func (t *partnerDataTransformer) TransformSpecificationToCapability(spec *interfaces.EntitySpecification, catalog *interfaces.Catalog) (*interfaces.ServiceCapability, error) {
	if spec == nil || catalog == nil {
		return nil, fmt.Errorf("specification and catalog cannot be nil")
	}

	// Find the catalog item that matches the specification
	var catalogItem *interfaces.CatalogItem
	for _, item := range catalog.CatalogItems {
		if item.Code == spec.SpecValue || item.Name == spec.SpecValue {
			catalogItem = &item
			break
		}
	}

	if catalogItem == nil {
		return nil, fmt.Errorf("catalog item not found for spec value: %s", spec.SpecValue)
	}

	// Extract service information from catalog item
	serviceType := t.extractServiceTypeFromCatalog(catalogItem)

	capability := &interfaces.ServiceCapability{
		ServiceType:    serviceType,
		ParcelCategory: t.MapServiceTypeToParcelCategory(serviceType),
		OperationTypes: t.ExtractOperationTypes(catalogItem.Description),
		PaymentModes:   t.ExtractPaymentModes(catalogItem.Description),
		DeliveryModes:  t.ExtractDeliveryModes(catalogItem.Description),
		IsActive:       true, // Default to active for valid specifications
		Rating:         t.calculateRatingFromSpec(spec),
	}

	return capability, nil
}

// ValidateAndNormalizePartnerData validates and normalizes partner capability data
func (t *partnerDataTransformer) ValidateAndNormalizePartnerData(capability *interfaces.PartnerCapability) error {
	if capability == nil {
		return fmt.Errorf("partner capability cannot be nil")
	}

	// Validate partner ID
	if capability.PartnerID == 0 {
		return fmt.Errorf("partner ID cannot be zero")
	}

	// Normalize partner name
	capability.PartnerName = strings.TrimSpace(capability.PartnerName)
	if capability.PartnerName == "" {
		return fmt.Errorf("partner name cannot be empty")
	}

	// Normalize and validate service types
	capability.ServiceTypes = t.normalizeStringSlice(capability.ServiceTypes)
	if len(capability.ServiceTypes) == 0 {
		capability.ServiceTypes = []string{"Standard"} // Default service type
	}

	// Normalize and validate parcel categories
	capability.ParcelCategories = t.normalizeStringSlice(capability.ParcelCategories)
	if len(capability.ParcelCategories) == 0 {
		capability.ParcelCategories = []string{"courier"} // Default parcel category
	}

	// Normalize operation types
	capability.OperationTypes = t.normalizeStringSlice(capability.OperationTypes)
	if len(capability.OperationTypes) == 0 {
		capability.OperationTypes = []string{"pickup", "delivery"} // Default operations
	}

	// Normalize payment modes
	capability.PaymentModes = t.normalizeStringSlice(capability.PaymentModes)
	if len(capability.PaymentModes) == 0 {
		capability.PaymentModes = []string{"ONLINE", "COD"} // Default payment modes
	}

	// Normalize delivery modes
	capability.DeliveryModes = t.normalizeStringSlice(capability.DeliveryModes)
	if len(capability.DeliveryModes) == 0 {
		capability.DeliveryModes = []string{"AIR", "SURFACE"} // Default delivery modes
	}

	return nil
}

// MapServiceTypeToParcelCategory maps service type to parcel category
func (t *partnerDataTransformer) MapServiceTypeToParcelCategory(serviceType string) string {
	if category, exists := t.serviceTypeMapping[serviceType]; exists {
		return category
	}
	return "courier" // Default category
}

// ExtractOperationTypes extracts operation types from specification value
func (t *partnerDataTransformer) ExtractOperationTypes(specValue string) []string {
	operations := []string{}
	lower := strings.ToLower(specValue)

	if strings.Contains(lower, "pickup") {
		operations = append(operations, "pickup")
	}
	if strings.Contains(lower, "delivery") {
		operations = append(operations, "delivery")
	}

	// Default to both if none specified
	if len(operations) == 0 {
		operations = []string{"pickup", "delivery"}
	}

	return operations
}

// ExtractPaymentModes extracts payment modes from specification value
func (t *partnerDataTransformer) ExtractPaymentModes(specValue string) []string {
	modes := []string{}
	upper := strings.ToUpper(specValue)

	if strings.Contains(upper, "ONLINE") {
		modes = append(modes, "ONLINE")
	}
	if strings.Contains(upper, "COD") {
		modes = append(modes, "COD")
	}

	// Default to both if none specified
	if len(modes) == 0 {
		modes = []string{"ONLINE", "COD"}
	}

	return modes
}

// ExtractDeliveryModes extracts delivery modes from specification value
func (t *partnerDataTransformer) ExtractDeliveryModes(specValue string) []string {
	modes := []string{}
	upper := strings.ToUpper(specValue)

	if strings.Contains(upper, "AIR") {
		modes = append(modes, "AIR")
	}
	if strings.Contains(upper, "SURFACE") {
		modes = append(modes, "SURFACE")
	}
	if strings.Contains(upper, "RAIL") {
		modes = append(modes, "RAIL")
	}

	// Default to AIR and SURFACE if none specified
	if len(modes) == 0 {
		modes = []string{"AIR", "SURFACE"}
	}

	return modes
}

// Helper methods

// extractServiceType extracts service type from preference value or entity ID
func (t *partnerDataTransformer) extractServiceType(value string) string {
	// Try to extract service type from common patterns
	upper := strings.ToUpper(value)

	if strings.Contains(upper, "SDD") {
		return "SDD"
	}
	if strings.Contains(upper, "NDD") {
		return "NDD"
	}
	if strings.Contains(upper, "EXPRESS") {
		return "Express"
	}
	if strings.Contains(upper, "VAYUQUICK") {
		if strings.Contains(upper, "PRO") {
			return "vayuquickpro"
		}
		return "vayuquick"
	}

	return "Standard" // Default service type
}

// isPreferredValue determines if a preference value indicates preference
func (t *partnerDataTransformer) isPreferredValue(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "preferred") || strings.Contains(lower, "true") || strings.Contains(lower, "yes")
}

// calculatePriority calculates priority based on effective rating
func (t *partnerDataTransformer) calculatePriority(rating float64) int {
	if rating >= 9.0 {
		return 1 // Highest priority
	} else if rating >= 7.0 {
		return 2 // High priority
	} else if rating >= 5.0 {
		return 3 // Medium priority
	} else if rating >= 3.0 {
		return 4 // Low priority
	}
	return 5 // Lowest priority
}

// extractServiceTypeFromCatalog extracts service type from catalog item
func (t *partnerDataTransformer) extractServiceTypeFromCatalog(item *interfaces.CatalogItem) string {
	// Try to extract from code first, then name, then description
	if serviceType := t.extractServiceType(item.Code); serviceType != "Standard" {
		return serviceType
	}
	if serviceType := t.extractServiceType(item.Name); serviceType != "Standard" {
		return serviceType
	}
	return t.extractServiceType(item.Description)
}

// calculateRatingFromSpec calculates rating from entity specification
func (t *partnerDataTransformer) calculateRatingFromSpec(spec *interfaces.EntitySpecification) float64 {
	// Try to extract numeric rating from spec value
	if rating, err := strconv.ParseFloat(spec.SpecValue, 64); err == nil {
		if rating >= 0 && rating <= 10 {
			return rating
		}
	}

	// Return default rating if can't parse
	return t.defaultRating
}

// normalizeStringSlice normalizes a slice of strings by trimming and removing empty values
func (t *partnerDataTransformer) normalizeStringSlice(slice []string) []string {
	var normalized []string
	seen := make(map[string]bool)

	for _, str := range slice {
		trimmed := strings.TrimSpace(str)
		if trimmed != "" && !seen[trimmed] {
			normalized = append(normalized, trimmed)
			seen[trimmed] = true
		}
	}

	return normalized
}
