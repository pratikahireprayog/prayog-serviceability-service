package services

import (
	"sort"
	"strings"

	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// ServiceDefinitionTransformer handles transformation of service definition data
type ServiceDefinitionTransformer struct {
	config config.ServiceDefinitionResolverConfig

	// Transformation mappings
	categoryMappings    map[string]string
	serviceTypeMappings map[string]string
	operationMappings   map[string]string
	paymentMappings     map[string]string
	deliveryMappings    map[string]string
}

// NewServiceDefinitionTransformer creates a new transformer instance
func NewServiceDefinitionTransformer(config config.ServiceDefinitionResolverConfig) *ServiceDefinitionTransformer {
	transformer := &ServiceDefinitionTransformer{
		config:              config,
		categoryMappings:    make(map[string]string),
		serviceTypeMappings: make(map[string]string),
		operationMappings:   make(map[string]string),
		paymentMappings:     make(map[string]string),
		deliveryMappings:    make(map[string]string),
	}

	// Initialize default mappings
	transformer.initializeDefaultMappings()

	return transformer
}

// initializeDefaultMappings sets up default transformation mappings
func (t *ServiceDefinitionTransformer) initializeDefaultMappings() {
	// Category mappings (catalog name -> standard name)
	t.categoryMappings = map[string]string{
		"PARCEL_CATEGORY": "parcel_category",
		"ECOMMERCE":       "ecom",
		"COURIER":         "courier",
		"CARGO":           "cargo",
		"DOCUMENTS":       "documents",
		"PACKAGES":        "packages",
		"FREIGHT":         "freight",
		"PHARMA":          "pharmaceuticals",
		"FOOD":            "food",
		"AUTO":            "automotive",
	}

	// Service type mappings
	t.serviceTypeMappings = map[string]string{
		"SAME_DAY_DELIVERY": "SDD",
		"NEXT_DAY_DELIVERY": "NDD",
		"STANDARD_DELIVERY": "Standard",
		"EXPRESS_DELIVERY":  "Express",
		"PREMIUM_DELIVERY":  "Premium",
		"ECONOMY_DELIVERY":  "Economy",
	}

	// Operation mappings
	t.operationMappings = map[string]string{
		"PICKUP_SERVICE":    "pickup",
		"DELIVERY_SERVICE":  "delivery",
		"TRACKING_SERVICE":  "tracking",
		"INSURANCE_SERVICE": "insurance",
		"PACKAGING_SERVICE": "packaging",
		"LABELING_SERVICE":  "labeling",
		"SORTING_SERVICE":   "sorting",
		"CUSTOMS_SERVICE":   "customs",
		"WAREHOUSE_SERVICE": "warehousing",
	}

	// Payment mappings
	t.paymentMappings = map[string]string{
		"CASH_ON_DELIVERY": "COD",
		"ONLINE_PAYMENT":   "ONLINE",
		"PREPAID_PAYMENT":  "PREPAID",
		"CREDIT_CARD":      "CREDIT",
		"DEBIT_CARD":       "DEBIT",
		"UPI_PAYMENT":      "UPI",
		"WALLET_PAYMENT":   "WALLET",
		"NET_BANKING":      "NET_BANKING",
		"CASH_PAYMENT":     "CASH",
	}

	// Delivery mappings
	t.deliveryMappings = map[string]string{
		"AIR_TRANSPORT":     "AIR",
		"SURFACE_TRANSPORT": "SURFACE",
		"RAIL_TRANSPORT":    "RAIL",
		"SEA_TRANSPORT":     "SEA",
		"ROAD_TRANSPORT":    "ROAD",
		"MULTI_MODAL":       "MULTIMODAL",
		"EXPRESS_MODE":      "EXPRESS",
		"STANDARD_MODE":     "STANDARD",
	}
}

// TransformParcelCategories transforms and normalizes parcel categories
func (t *ServiceDefinitionTransformer) TransformParcelCategories(categories []string) []string {
	if !t.config.EnableTransformation {
		return categories
	}

	transformed := make([]string, 0, len(categories))
	seen := make(map[string]bool)

	for _, category := range categories {
		normalized := t.normalizeParcelCategory(category)
		if normalized != "" && !seen[normalized] {
			transformed = append(transformed, normalized)
			seen[normalized] = true
		}
	}

	// Sort for consistency
	sort.Strings(transformed)
	return transformed
}

// TransformServiceTypes transforms and normalizes service types
func (t *ServiceDefinitionTransformer) TransformServiceTypes(serviceTypes []string) []string {
	if !t.config.EnableTransformation {
		return serviceTypes
	}

	transformed := make([]string, 0, len(serviceTypes))
	seen := make(map[string]bool)

	for _, serviceType := range serviceTypes {
		normalized := t.normalizeServiceType(serviceType)
		if normalized != "" && !seen[normalized] {
			transformed = append(transformed, normalized)
			seen[normalized] = true
		}
	}

	// Sort for consistency
	sort.Strings(transformed)
	return transformed
}

// TransformOperationTypes transforms and normalizes operation types
func (t *ServiceDefinitionTransformer) TransformOperationTypes(operationTypes []string) []string {
	if !t.config.EnableTransformation {
		return operationTypes
	}

	transformed := make([]string, 0, len(operationTypes))
	seen := make(map[string]bool)

	for _, operationType := range operationTypes {
		normalized := t.normalizeOperationType(operationType)
		if normalized != "" && !seen[normalized] {
			transformed = append(transformed, normalized)
			seen[normalized] = true
		}
	}

	// Sort for consistency
	sort.Strings(transformed)
	return transformed
}

// TransformPaymentModes transforms and normalizes payment modes
func (t *ServiceDefinitionTransformer) TransformPaymentModes(paymentModes []string) []string {
	if !t.config.EnableTransformation {
		return paymentModes
	}

	transformed := make([]string, 0, len(paymentModes))
	seen := make(map[string]bool)

	for _, paymentMode := range paymentModes {
		normalized := t.normalizePaymentMode(paymentMode)
		if normalized != "" && !seen[normalized] {
			transformed = append(transformed, normalized)
			seen[normalized] = true
		}
	}

	// Sort for consistency
	sort.Strings(transformed)
	return transformed
}

// TransformDeliveryModes transforms and normalizes delivery modes
func (t *ServiceDefinitionTransformer) TransformDeliveryModes(deliveryModes []string) []string {
	if !t.config.EnableTransformation {
		return deliveryModes
	}

	transformed := make([]string, 0, len(deliveryModes))
	seen := make(map[string]bool)

	for _, deliveryMode := range deliveryModes {
		normalized := t.normalizeDeliveryMode(deliveryMode)
		if normalized != "" && !seen[normalized] {
			transformed = append(transformed, normalized)
			seen[normalized] = true
		}
	}

	// Sort for consistency
	sort.Strings(transformed)
	return transformed
}

// TransformServiceDefinition transforms a complete service definition
func (t *ServiceDefinitionTransformer) TransformServiceDefinition(serviceDef *interfaces.ServiceDefinition) *interfaces.ServiceDefinition {
	if !t.config.EnableTransformation || serviceDef == nil {
		return serviceDef
	}

	transformed := &interfaces.ServiceDefinition{
		ServiceType:       t.normalizeServiceType(serviceDef.ServiceType),
		ParcelCategory:    t.normalizeParcelCategory(serviceDef.ParcelCategory),
		DefaultOperations: t.TransformOperationTypes(serviceDef.DefaultOperations),
		DefaultPayments:   t.TransformPaymentModes(serviceDef.DefaultPayments),
		DefaultDelivery:   t.TransformDeliveryModes(serviceDef.DefaultDelivery),
		Description:       strings.TrimSpace(serviceDef.Description),
	}

	return transformed
}

// Catalog extraction methods (used by ServiceDefinitionResolver)

// ExtractParcelCategoriesFromCatalogs extracts parcel categories from catalogs
func (t *ServiceDefinitionTransformer) ExtractParcelCategoriesFromCatalogs(catalogs []interfaces.Catalog) []string {
	categories := make([]string, 0)
	seen := make(map[string]bool)

	for _, catalog := range catalogs {
		// Look for catalogs that represent parcel categories
		if t.isParcelCategoryCatalog(catalog) {
			for _, item := range catalog.CatalogItems {
				normalized := t.normalizeParcelCategory(item.Code)
				if normalized != "" && !seen[normalized] {
					categories = append(categories, normalized)
					seen[normalized] = true
				}
			}
		}
	}

	// Sort for consistency
	sort.Strings(categories)
	return categories
}

// ExtractServiceTypesFromCatalogs extracts service types for a specific parcel category
func (t *ServiceDefinitionTransformer) ExtractServiceTypesFromCatalogs(catalogs []interfaces.Catalog, parcelCategory string) []string {
	serviceTypes := make([]string, 0)
	seen := make(map[string]bool)

	for _, catalog := range catalogs {
		// Look for catalogs that represent service types
		if t.isServiceTypeCatalog(catalog) {
			for _, item := range catalog.CatalogItems {
				// Filter by parcel category if specified
				if parcelCategory != "" && !t.isServiceTypeForCategory(item, parcelCategory) {
					continue
				}

				normalized := t.normalizeServiceType(item.Code)
				if normalized != "" && !seen[normalized] {
					serviceTypes = append(serviceTypes, normalized)
					seen[normalized] = true
				}
			}
		}
	}

	// Sort for consistency
	sort.Strings(serviceTypes)
	return serviceTypes
}

// ExtractOperationTypesFromCatalogs extracts operation types from catalogs
func (t *ServiceDefinitionTransformer) ExtractOperationTypesFromCatalogs(catalogs []interfaces.Catalog) []string {
	operationTypes := make([]string, 0)
	seen := make(map[string]bool)

	for _, catalog := range catalogs {
		if t.isOperationTypeCatalog(catalog) {
			for _, item := range catalog.CatalogItems {
				normalized := t.normalizeOperationType(item.Code)
				if normalized != "" && !seen[normalized] {
					operationTypes = append(operationTypes, normalized)
					seen[normalized] = true
				}
			}
		}
	}

	// Sort for consistency
	sort.Strings(operationTypes)
	return operationTypes
}

// ExtractPaymentModesFromCatalogs extracts payment modes from catalogs
func (t *ServiceDefinitionTransformer) ExtractPaymentModesFromCatalogs(catalogs []interfaces.Catalog) []string {
	paymentModes := make([]string, 0)
	seen := make(map[string]bool)

	for _, catalog := range catalogs {
		if t.isPaymentModeCatalog(catalog) {
			for _, item := range catalog.CatalogItems {
				normalized := t.normalizePaymentMode(item.Code)
				if normalized != "" && !seen[normalized] {
					paymentModes = append(paymentModes, normalized)
					seen[normalized] = true
				}
			}
		}
	}

	// Sort for consistency
	sort.Strings(paymentModes)
	return paymentModes
}

// ExtractDeliveryModesFromCatalogs extracts delivery modes from catalogs
func (t *ServiceDefinitionTransformer) ExtractDeliveryModesFromCatalogs(catalogs []interfaces.Catalog) []string {
	deliveryModes := make([]string, 0)
	seen := make(map[string]bool)

	for _, catalog := range catalogs {
		if t.isDeliveryModeCatalog(catalog) {
			for _, item := range catalog.CatalogItems {
				normalized := t.normalizeDeliveryMode(item.Code)
				if normalized != "" && !seen[normalized] {
					deliveryModes = append(deliveryModes, normalized)
					seen[normalized] = true
				}
			}
		}
	}

	// Sort for consistency
	sort.Strings(deliveryModes)
	return deliveryModes
}

// BuildServiceDefinitionFromCatalogs builds a service definition from catalogs
func (t *ServiceDefinitionTransformer) BuildServiceDefinitionFromCatalogs(catalogs []interfaces.Catalog, serviceType string) *interfaces.ServiceDefinition {
	// This is a simplified implementation
	// In a real implementation, you would need business logic to determine
	// which operations, payments, and delivery modes are associated with each service type

	return &interfaces.ServiceDefinition{
		ServiceType:       t.normalizeServiceType(serviceType),
		ParcelCategory:    "ecom", // Default category
		DefaultOperations: []string{"pickup", "delivery"},
		DefaultPayments:   []string{"COD", "ONLINE"},
		DefaultDelivery:   []string{"SURFACE"},
		Description:       "Service definition for " + serviceType,
	}
}

// Private normalization methods

func (t *ServiceDefinitionTransformer) normalizeParcelCategory(category string) string {
	if category == "" {
		return ""
	}

	upperCategory := strings.ToUpper(strings.TrimSpace(category))

	// Check for direct mapping
	if mapped, exists := t.categoryMappings[upperCategory]; exists {
		return mapped
	}

	// Return lowercase version as fallback
	return strings.ToLower(strings.TrimSpace(category))
}

func (t *ServiceDefinitionTransformer) normalizeServiceType(serviceType string) string {
	if serviceType == "" {
		return ""
	}

	upperServiceType := strings.ToUpper(strings.TrimSpace(serviceType))

	// Check for direct mapping
	if mapped, exists := t.serviceTypeMappings[upperServiceType]; exists {
		return mapped
	}

	// Return uppercase version as fallback
	return strings.ToUpper(strings.TrimSpace(serviceType))
}

func (t *ServiceDefinitionTransformer) normalizeOperationType(operationType string) string {
	if operationType == "" {
		return ""
	}

	upperOperationType := strings.ToUpper(strings.TrimSpace(operationType))

	// Check for direct mapping
	if mapped, exists := t.operationMappings[upperOperationType]; exists {
		return mapped
	}

	// Return lowercase version as fallback
	return strings.ToLower(strings.TrimSpace(operationType))
}

func (t *ServiceDefinitionTransformer) normalizePaymentMode(paymentMode string) string {
	if paymentMode == "" {
		return ""
	}

	upperPaymentMode := strings.ToUpper(strings.TrimSpace(paymentMode))

	// Check for direct mapping
	if mapped, exists := t.paymentMappings[upperPaymentMode]; exists {
		return mapped
	}

	// Return uppercase version as fallback
	return strings.ToUpper(strings.TrimSpace(paymentMode))
}

func (t *ServiceDefinitionTransformer) normalizeDeliveryMode(deliveryMode string) string {
	if deliveryMode == "" {
		return ""
	}

	upperDeliveryMode := strings.ToUpper(strings.TrimSpace(deliveryMode))

	// Check for direct mapping
	if mapped, exists := t.deliveryMappings[upperDeliveryMode]; exists {
		return mapped
	}

	// Return uppercase version as fallback
	return strings.ToUpper(strings.TrimSpace(deliveryMode))
}

// Catalog type detection methods

func (t *ServiceDefinitionTransformer) isParcelCategoryCatalog(catalog interfaces.Catalog) bool {
	catalogName := strings.ToUpper(catalog.Name)
	return strings.Contains(catalogName, "PARCEL") && strings.Contains(catalogName, "CATEGORY")
}

func (t *ServiceDefinitionTransformer) isServiceTypeCatalog(catalog interfaces.Catalog) bool {
	catalogName := strings.ToUpper(catalog.Name)
	return strings.Contains(catalogName, "SERVICE") && strings.Contains(catalogName, "TYPE")
}

func (t *ServiceDefinitionTransformer) isOperationTypeCatalog(catalog interfaces.Catalog) bool {
	catalogName := strings.ToUpper(catalog.Name)
	return strings.Contains(catalogName, "OPERATION") && strings.Contains(catalogName, "TYPE")
}

func (t *ServiceDefinitionTransformer) isPaymentModeCatalog(catalog interfaces.Catalog) bool {
	catalogName := strings.ToUpper(catalog.Name)
	return strings.Contains(catalogName, "PAYMENT") && strings.Contains(catalogName, "MODE")
}

func (t *ServiceDefinitionTransformer) isDeliveryModeCatalog(catalog interfaces.Catalog) bool {
	catalogName := strings.ToUpper(catalog.Name)
	return strings.Contains(catalogName, "DELIVERY") && strings.Contains(catalogName, "MODE")
}

func (t *ServiceDefinitionTransformer) isServiceTypeForCategory(item interfaces.CatalogItem, parcelCategory string) bool {
	// This is a simplified implementation
	// In a real implementation, you would have business logic to determine
	// which service types are applicable to which parcel categories
	return true
}

// Update mappings methods

// UpdateCategoryMappings allows updating category mappings at runtime
func (t *ServiceDefinitionTransformer) UpdateCategoryMappings(mappings map[string]string) {
	if mappings != nil {
		t.categoryMappings = mappings
	}
}

// UpdateServiceTypeMappings allows updating service type mappings at runtime
func (t *ServiceDefinitionTransformer) UpdateServiceTypeMappings(mappings map[string]string) {
	if mappings != nil {
		t.serviceTypeMappings = mappings
	}
}

// UpdateOperationMappings allows updating operation mappings at runtime
func (t *ServiceDefinitionTransformer) UpdateOperationMappings(mappings map[string]string) {
	if mappings != nil {
		t.operationMappings = mappings
	}
}

// UpdatePaymentMappings allows updating payment mappings at runtime
func (t *ServiceDefinitionTransformer) UpdatePaymentMappings(mappings map[string]string) {
	if mappings != nil {
		t.paymentMappings = mappings
	}
}

// UpdateDeliveryMappings allows updating delivery mappings at runtime
func (t *ServiceDefinitionTransformer) UpdateDeliveryMappings(mappings map[string]string) {
	if mappings != nil {
		t.deliveryMappings = mappings
	}
}

// GetMappings returns current transformation mappings
func (t *ServiceDefinitionTransformer) GetMappings() map[string]map[string]string {
	return map[string]map[string]string{
		"category":  t.categoryMappings,
		"service":   t.serviceTypeMappings,
		"operation": t.operationMappings,
		"payment":   t.paymentMappings,
		"delivery":  t.deliveryMappings,
	}
}

// IsTransformationEnabled returns whether transformation is enabled
func (t *ServiceDefinitionTransformer) IsTransformationEnabled() bool {
	return t.config.EnableTransformation
}

// SetTransformationEnabled enables or disables transformation
func (t *ServiceDefinitionTransformer) SetTransformationEnabled(enabled bool) {
	t.config.EnableTransformation = enabled
}
