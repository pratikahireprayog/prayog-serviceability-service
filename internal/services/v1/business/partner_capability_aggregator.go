package services

import (
	"context"
	"fmt"
	"strconv"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// PartnerCapabilityAggregator implements the PartnerCapabilityAggregator interface
// with simplified, direct API calls for catalog-based serviceability
type PartnerCapabilityAggregator struct {
	partnerService interfaces.PartnerServiceClient
	specService    interfaces.SpecificationServiceClient
}

// NewPartnerCapabilityAggregator creates a new simplified partner capability aggregator
func NewPartnerCapabilityAggregator(
	partnerService interfaces.PartnerServiceClient,
	specService interfaces.SpecificationServiceClient,
) interfaces.PartnerCapabilityAggregator {
	return &PartnerCapabilityAggregator{
		partnerService: partnerService,
		specService:    specService,
	}
}

// GetPartnersByLocation retrieves partners serving a specific location
func (pca *PartnerCapabilityAggregator) GetPartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy) ([]interfaces.PartnerCapability, error) {
	// Validate inputs
	if locationHierarchy == nil {
		return nil, fmt.Errorf("location hierarchy cannot be nil")
	}

	if locationHierarchy.PostalCode == "" {
		return nil, fmt.Errorf("postal code is required")
	}

	// Call Partner Service directly to get partners for location
	locationType := "postal_code"
	locationID := locationHierarchy.PostalCode

	partnerInfos, err := pca.partnerService.GetPartnersByLocation(ctx, locationType, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch partners from partner service: %w", err)
	}

	// Transform partner info to partner capability
	var capabilities []interfaces.PartnerCapability
	for _, partnerInfo := range partnerInfos {
		capability := interfaces.PartnerCapability{
			PartnerID:   partnerInfo.ID,
			PartnerName: partnerInfo.Name,
			IsActive:    partnerInfo.IsActive,
		}

		// Skip inactive partners
		if !capability.IsActive {
			continue
		}

		capabilities = append(capabilities, capability)
	}

	return capabilities, nil
}

// GetPartnerCapabilities retrieves detailed capabilities for a specific partner
func (pca *PartnerCapabilityAggregator) GetPartnerCapabilities(ctx context.Context, partnerID uint, serviceFilters *interfaces.ServiceFilters) (*interfaces.PartnerServiceCapabilities, error) {
	// Validate inputs
	if partnerID == 0 {
		return nil, fmt.Errorf("partner ID cannot be zero")
	}

	// Get partner effective details from Partner Service
	entityType := "partner"
	entityID := strconv.FormatUint(uint64(partnerID), 10)

	effectiveDetails, err := pca.partnerService.GetPartnerEffectiveDetails(ctx, partnerID, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch partner effective details: %w", err)
	}

	// Transform effective details to service capabilities
	capabilities := &interfaces.PartnerServiceCapabilities{
		PartnerID:    partnerID,
		Capabilities: []interfaces.ServiceCapability{},
		Preferences:  []interfaces.PartnerPreference{},
	}

	// Convert effective preferences to partner preferences
	for _, effectivePref := range effectiveDetails.Preferences {
		preference := interfaces.PartnerPreference{
			ServiceType:     effectivePref.EntityType,
			ParcelCategory:  "standard", // default
			IsPreferred:     effectivePref.EffectiveRating > 5.0,
			Priority:        1,
			EffectiveRating: effectivePref.EffectiveRating,
		}
		capabilities.Preferences = append(capabilities.Preferences, preference)
	}

	// Create basic service capabilities - for catalog-based only
	basicCapability := interfaces.ServiceCapability{
		ServiceType:    "standard",
		ParcelCategory: "ecom",
		OperationTypes: []string{"pickup", "delivery"},
		PaymentModes:   []string{"cod", "prepaid"},
		DeliveryModes:  []string{"standard"},
		IsActive:       true,
		Rating:         4.0,
	}

	// Apply service filters if provided
	if serviceFilters == nil || pca.matchesServiceFilters(&basicCapability, serviceFilters) {
		capabilities.Capabilities = append(capabilities.Capabilities, basicCapability)
	}

	return capabilities, nil
}

// buildServiceCapabilityFromSpec creates a service capability - simplified for catalog-based approach
func (pca *PartnerCapabilityAggregator) buildServiceCapabilityFromSpec(name string) interfaces.ServiceCapability {
	capability := interfaces.ServiceCapability{
		ServiceType:    name,
		ParcelCategory: "standard", // default
		OperationTypes: []string{"pickup", "delivery"},
		PaymentModes:   []string{"cod", "prepaid"},
		DeliveryModes:  []string{"standard"},
		IsActive:       true,
		Rating:         4.0,
	}

	return capability
}

// matchesServiceFilters checks if a service capability matches the provided filters
func (pca *PartnerCapabilityAggregator) matchesServiceFilters(capability *interfaces.ServiceCapability, filters *interfaces.ServiceFilters) bool {
	// Check service types
	if len(filters.ServiceTypes) > 0 {
		matched := false
		for _, filterType := range filters.ServiceTypes {
			if capability.ServiceType == filterType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check parcel categories
	if len(filters.ParcelCategories) > 0 {
		matched := false
		for _, filterCategory := range filters.ParcelCategories {
			if capability.ParcelCategory == filterCategory {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check operation types
	if len(filters.OperationTypes) > 0 {
		matched := false
		for _, filterOp := range filters.OperationTypes {
			for _, capOp := range capability.OperationTypes {
				if capOp == filterOp {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}
