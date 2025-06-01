package services

import (
	"context"
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// serviceDefinitionResolver implements the ServiceDefinitionResolver interface
type serviceDefinitionResolver struct {
	specService interfaces.SpecificationServiceClient
}

// NewServiceDefinitionResolver creates a new ServiceDefinitionResolver instance
func NewServiceDefinitionResolver(specService interfaces.SpecificationServiceClient) interfaces.ServiceDefinitionResolver {
	return &serviceDefinitionResolver{
		specService: specService,
	}
}

// GetParcelCategories retrieves all available parcel categories
func (sdr *serviceDefinitionResolver) GetParcelCategories(ctx context.Context) ([]string, error) {
	// Get catalogs from specification service
	catalogs, err := sdr.specService.GetCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalogs from specification service: %w", err)
	}

	// Extract parcel categories from catalogs
	parcelCategories := make([]string, 0)
	for _, catalog := range catalogs {
		if strings.Contains(strings.ToLower(catalog.Name), "parcel") ||
			strings.Contains(strings.ToLower(catalog.Name), "category") {
			for _, item := range catalog.CatalogItems {
				parcelCategories = append(parcelCategories, item.Code)
			}
		}
	}

	// If no parcel categories found in catalogs, return defaults
	if len(parcelCategories) == 0 {
		parcelCategories = []string{"ecom", "courier", "cargo", "express"}
	}

	return parcelCategories, nil
}

// GetServiceTypes retrieves service types for a specific parcel category
func (sdr *serviceDefinitionResolver) GetServiceTypes(ctx context.Context, parcelCategory string) ([]string, error) {
	// Get entity specifications for the parcel category
	entitySpecs, err := sdr.specService.GetEntitySpecifications(ctx, "parcel_category", parcelCategory)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity specifications for parcel category %s: %w", parcelCategory, err)
	}

	// Extract service types from specifications
	serviceTypes := make([]string, 0)
	for _, spec := range entitySpecs {
		if strings.Contains(strings.ToLower(spec.Description), "service") ||
			strings.Contains(strings.ToLower(spec.Description), "type") {
			serviceTypes = append(serviceTypes, spec.SpecValue)
		}
	}

	// If no service types found, return defaults based on parcel category
	if len(serviceTypes) == 0 {
		serviceTypes = sdr.getDefaultServiceTypes(parcelCategory)
	}

	return serviceTypes, nil
}

// GetOperationTypes retrieves all available operation types
func (sdr *serviceDefinitionResolver) GetOperationTypes(ctx context.Context) ([]string, error) {
	// Get catalogs from specification service
	catalogs, err := sdr.specService.GetCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalogs from specification service: %w", err)
	}

	// Extract operation types from catalogs
	operationTypes := make([]string, 0)
	for _, catalog := range catalogs {
		if strings.Contains(strings.ToLower(catalog.Name), "operation") {
			for _, item := range catalog.CatalogItems {
				operationTypes = append(operationTypes, item.Code)
			}
		}
	}

	// If no operation types found, return defaults
	if len(operationTypes) == 0 {
		operationTypes = []string{"pickup", "delivery", "transit", "hub"}
	}

	return operationTypes, nil
}

// GetPaymentModes retrieves all available payment modes
func (sdr *serviceDefinitionResolver) GetPaymentModes(ctx context.Context) ([]string, error) {
	// Get catalogs from specification service
	catalogs, err := sdr.specService.GetCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalogs from specification service: %w", err)
	}

	// Extract payment modes from catalogs
	paymentModes := make([]string, 0)
	for _, catalog := range catalogs {
		if strings.Contains(strings.ToLower(catalog.Name), "payment") {
			for _, item := range catalog.CatalogItems {
				paymentModes = append(paymentModes, item.Code)
			}
		}
	}

	// If no payment modes found, return defaults
	if len(paymentModes) == 0 {
		paymentModes = []string{"COD", "ONLINE", "CARD", "UPI", "WALLET"}
	}

	return paymentModes, nil
}

// GetDeliveryModes retrieves all available delivery modes
func (sdr *serviceDefinitionResolver) GetDeliveryModes(ctx context.Context) ([]string, error) {
	// Get catalogs from specification service
	catalogs, err := sdr.specService.GetCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalogs from specification service: %w", err)
	}

	// Extract delivery modes from catalogs
	deliveryModes := make([]string, 0)
	for _, catalog := range catalogs {
		if strings.Contains(strings.ToLower(catalog.Name), "delivery") ||
			strings.Contains(strings.ToLower(catalog.Name), "transport") {
			for _, item := range catalog.CatalogItems {
				deliveryModes = append(deliveryModes, item.Code)
			}
		}
	}

	// If no delivery modes found, return defaults
	if len(deliveryModes) == 0 {
		deliveryModes = []string{"SURFACE", "AIR", "EXPRESS", "RAIL", "SEA"}
	}

	return deliveryModes, nil
}

// GetServiceDefinition retrieves detailed service definition for a specific service type
func (sdr *serviceDefinitionResolver) GetServiceDefinition(ctx context.Context, serviceType string) (*interfaces.ServiceDefinition, error) {
	// Get entity specifications for the service type
	entitySpecs, err := sdr.specService.GetEntitySpecifications(ctx, "service_type", serviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity specifications for service type %s: %w", serviceType, err)
	}

	// Build service definition from specifications
	serviceDef := &interfaces.ServiceDefinition{
		ServiceType: serviceType,
	}

	// Extract details from specifications
	for _, spec := range entitySpecs {
		switch strings.ToLower(spec.Description) {
		case "parcel_category":
			serviceDef.ParcelCategory = spec.SpecValue
		case "default_operations":
			serviceDef.DefaultOperations = strings.Split(spec.SpecValue, ",")
		case "default_payments":
			serviceDef.DefaultPayments = strings.Split(spec.SpecValue, ",")
		case "default_delivery":
			serviceDef.DefaultDelivery = strings.Split(spec.SpecValue, ",")
		case "description":
			serviceDef.Description = spec.SpecValue
		}
	}

	// If no specifications found, provide defaults
	if serviceDef.ParcelCategory == "" {
		serviceDef = sdr.getDefaultServiceDefinition(serviceType)
	}

	return serviceDef, nil
}

// getDefaultServiceTypes returns default service types for a parcel category
func (sdr *serviceDefinitionResolver) getDefaultServiceTypes(parcelCategory string) []string {
	switch strings.ToLower(parcelCategory) {
	case "ecom":
		return []string{"Standard", "Express", "Premium"}
	case "courier":
		return []string{"SDD", "NDD", "Express"}
	case "cargo":
		return []string{"LTL", "FTL", "Heavy"}
	case "express":
		return []string{"Same Day", "Next Day", "2-Day"}
	default:
		return []string{"Standard", "Express"}
	}
}

// getDefaultServiceDefinition returns a default service definition
func (sdr *serviceDefinitionResolver) getDefaultServiceDefinition(serviceType string) *interfaces.ServiceDefinition {
	serviceDef := &interfaces.ServiceDefinition{
		ServiceType:       serviceType,
		DefaultOperations: []string{"pickup", "delivery"},
		DefaultPayments:   []string{"COD", "ONLINE"},
		DefaultDelivery:   []string{"SURFACE"},
		Description:       fmt.Sprintf("Default service definition for %s", serviceType),
	}

	switch strings.ToLower(serviceType) {
	case "standard":
		serviceDef.ParcelCategory = "ecom"
		serviceDef.DefaultDelivery = []string{"SURFACE"}
	case "express":
		serviceDef.ParcelCategory = "ecom"
		serviceDef.DefaultDelivery = []string{"AIR"}
	case "sdd":
		serviceDef.ParcelCategory = "courier"
		serviceDef.DefaultDelivery = []string{"SURFACE"}
	case "premium":
		serviceDef.ParcelCategory = "ecom"
		serviceDef.DefaultDelivery = []string{"AIR"}
		serviceDef.DefaultPayments = []string{"ONLINE"}
	default:
		serviceDef.ParcelCategory = "ecom"
	}

	return serviceDef
}
