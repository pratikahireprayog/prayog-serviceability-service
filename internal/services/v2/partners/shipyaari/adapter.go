package shipyaari

import (
	"context"
	"fmt"
	"strconv"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// ShipyaariAdapter implements the common.PartnerAdapter interface for Shipyaari
type ShipyaariAdapter struct {
	*common.HTTPBaseAdapter
	client *Client
	config config.ShipyaariConfig
}

// NewShipyaariAdapter creates a new Shipyaari adapter
func NewShipyaariAdapter(cfg config.ShipyaariConfig) common.PartnerAdapter {
	// Create partner config
	partnerConfig := common.GetPartnerConfigDefaults("shipyaari", "Shipyaari", common.AdapterTypeHTTP)
	partnerConfig.Timeout = cfg.Timeout
	partnerConfig.Enabled = cfg.Enabled
	// Rating comes from database, not hardcoded

	// Update endpoints
	partnerConfig.Endpoints = map[string]string{
		"base_url":          cfg.BaseURL,
		"token_url":         cfg.TokenURL,
		"check_service_url": cfg.CheckServiceURL,
	}

	// Update auth config
	partnerConfig.Auth = common.AuthConfig{
		Type: common.AuthTypeJWT,
		Credentials: map[string]string{
			"email":    cfg.Email,
			"password": cfg.Password,
		},
		TokenURL: cfg.TokenURL,
	}

	// Create base adapter
	baseAdapter := common.NewHTTPBaseAdapter(partnerConfig)

	// Create Shipyaari specific components
	auth := NewAuthenticator(cfg)
	client := NewClient(cfg, auth)

	// Set HTTP client and authenticator
	baseAdapter.SetHTTPClient(client)
	baseAdapter.SetAuthenticator(auth)

	adapter := &ShipyaariAdapter{
		HTTPBaseAdapter: baseAdapter,
		client:          client,
		config:          cfg,
	}

	return adapter
}

// CheckServiceability checks serviceability for the request
func (s *ShipyaariAdapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	// Basic validation
	// Ensure defaults for missing package details before validating
	cp := *request
	s.ensureDefaultPackage(&cp)

	if err := s.validateShipyaariRequirements(&cp); err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  s.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Shipyaari validation failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "Shipyaari does not support this request type",
			},
		}, nil
	}

	// Convert request to Shipyaari format
	shipyaariRequest, err := s.transformRequest(&cp)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  s.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Shipyaari request transformation failed: %v", err)}[0],
		}, nil
	}

	// Make API call
	response, err := s.client.CheckServiceability(ctx, shipyaariRequest)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  s.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Shipyaari API call failed: %v", err)}[0],
		}, nil
	}

	// Convert response and return
	// The orchestrator will set PartnerCode and PartnerName from database
	return s.transformResponse(response, partnerInfo), nil
}

// Initialize performs any necessary initialization
func (s *ShipyaariAdapter) Initialize(ctx context.Context) error {
	if err := s.HTTPBaseAdapter.Initialize(ctx); err != nil {
		return err
	}

	// Authenticate with Shipyaari
	if err := s.GetAuthenticator().Authenticate(ctx); err != nil {
		s.SetHealthStatus("unhealthy")
		return err
	}

	s.SetHealthStatus("healthy")
	return nil
}

// IsHealthy checks if the Shipyaari adapter is healthy
func (s *ShipyaariAdapter) IsHealthy(ctx context.Context) bool {
	// Check base health
	if !s.HTTPBaseAdapter.IsHealthy(ctx) {
		return false
	}

	// For basic health check, we only require that the adapter is enabled
	// Authentication will be checked during actual API calls
	return true
}

// ensureDefaultPackage fills missing or invalid package/weight/dimensions with sensible defaults
func (s *ShipyaariAdapter) ensureDefaultPackage(req *models.ServiceabilityV2Request) {
	// If no packages provided, create one with defaults
	if len(req.Packages) == 0 {
		req.Packages = []models.Package{{
			Weight:     &models.Weight{Value: 1.0, Unit: "g"},
			Dimensions: &models.Dimensions{Length: 1, Width: 1, Height: 1, Unit: "cm"},
		}}
		return
	}

	// Work with the first package (current API supports single-package pricing)
	first := &req.Packages[0]

	// Default or sanitize weight
	if first.Weight == nil || first.Weight.Value <= 0 {
		first.Weight = &models.Weight{Value: 1.0, Unit: "g"}
	} else if first.Weight.Unit == "" {
		first.Weight.Unit = "g"
	}

	// Default or sanitize dimensions (LBH)
	if first.Dimensions == nil || first.Dimensions.Length <= 0 || first.Dimensions.Width <= 0 || first.Dimensions.Height <= 0 {
		first.Dimensions = &models.Dimensions{Length: 1, Width: 1, Height: 1, Unit: "cm"}
	} else if first.Dimensions.Unit == "" {
		first.Dimensions.Unit = "cm"
	}
}

// defaultWeight returns the first package weight, defaulting if missing
func defaultWeight(req *models.ServiceabilityV2Request) *models.Weight {
    if len(req.Packages) > 0 && req.Packages[0].Weight != nil {
        return req.Packages[0].Weight
    }
    return &models.Weight{Value: 1.0, Unit: "g"}
}

// defaultDimensions returns the first package dimensions, defaulting if missing
func defaultDimensions(req *models.ServiceabilityV2Request) *models.Dimensions {
    if len(req.Packages) > 0 && req.Packages[0].Dimensions != nil {
        return req.Packages[0].Dimensions
    }
    return &models.Dimensions{Length: 1, Width: 1, Height: 1, Unit: "cm"}
}

// validateShipyaariRequirements validates Shipyaari-specific requirements
func (s *ShipyaariAdapter) validateShipyaariRequirements(req *models.ServiceabilityV2Request) error {
	// Shipyaari requires both source and destination postal codes
	if req.SourcePostalCode == nil || *req.SourcePostalCode == "" {
		return fmt.Errorf("source postal code is required for Shipyaari shipments")
	}

	if req.DestinationPostalCode == nil || *req.DestinationPostalCode == "" {
		return fmt.Errorf("destination postal code is required for Shipyaari shipments")
	}

	// Shipyaari requires package information for accurate pricing calculations
	if len(req.Packages) == 0 {
		return fmt.Errorf("at least one package is required for Shipyaari pricing calculations")
	}

	// Validate each package in the array
	for i, pkg := range req.Packages {
		// Validate weight is provided
		if pkg.Weight == nil {
			return fmt.Errorf("weight information is required for package %d in Shipyaari pricing calculations", i+1)
		}

		if pkg.Weight.Value <= 0 {
			return fmt.Errorf("package weight must be greater than 0 for package %d in Shipyaari pricing calculations", i+1)
		}

		// Validate dimensions are provided
		if pkg.Dimensions == nil {
			return fmt.Errorf("dimensions information is required for package %d in Shipyaari pricing calculations", i+1)
		}

		if pkg.Dimensions.Length <= 0 || pkg.Dimensions.Width <= 0 || pkg.Dimensions.Height <= 0 {
			return fmt.Errorf("package dimensions must be greater than 0 for package %d in Shipyaari pricing calculations", i+1)
		}
	}

	return nil
}

// transformRequest converts standard request to Shipyaari format
func (s *ShipyaariAdapter) transformRequest(req *models.ServiceabilityV2Request) (*ServiceabilityRequest, error) {
	// Get pincodes as strings first
	sourcePincode := getSourcePincode(req)
	destPincode := getDestinationPincode(req)

    // Debug: log the raw string pincodes before conversion
    if s.GetMetrics() != nil { // lightweight guard to avoid adding a logger dependency
        // Using fmt.Printf to avoid logger dep; acceptable for debug visibility
        fmt.Printf("[shipyaari] transformRequest pincodes (str) - source: %s, destination: %s\n", sourcePincode, destPincode)
    }

	// Validate pincodes are not empty
	if sourcePincode == "" {
		return nil, fmt.Errorf("source pincode is required for Shipyaari")
	}
	if destPincode == "" {
		return nil, fmt.Errorf("destination pincode is required for Shipyaari")
	}

	// Convert postal codes to integers (Shipyaari expects integers)
	pickupPincode, err := strconv.Atoi(sourcePincode)
	if err != nil {
		return nil, fmt.Errorf("invalid source pincode: %s", sourcePincode)
	}

	deliveryPincode, err := strconv.Atoi(destPincode)
	if err != nil {
		return nil, fmt.Errorf("invalid destination pincode: %s", destPincode)
	}

    // Debug: log the converted integer pincodes
    if s.GetMetrics() != nil {
        fmt.Printf("[shipyaari] transformRequest pincodes (int) - pickup: %d, delivery: %d\n", pickupPincode, deliveryPincode)
    }

    shipyaariReq := &ServiceabilityRequest{
		PickupPincode:   pickupPincode,
		DeliveryPincode: deliveryPincode,
		InvoiceValue:    getOrderValue(req),
		PaymentMode:     getPaymentMode(req),
        Weight:          getWeightInKg(defaultWeight(req)),
		OrderType:       "B2C", // Default to B2C
		Dimension: Dimension{
            Length: getDimensionInCm(defaultDimensions(req).Length, defaultDimensions(req).Unit),
            Width:  getDimensionInCm(defaultDimensions(req).Width, defaultDimensions(req).Unit),
            Height: getDimensionInCm(defaultDimensions(req).Height, defaultDimensions(req).Unit),
		},
	}

	return shipyaariReq, nil
}

// transformResponse converts Shipyaari response to standard format
func (s *ShipyaariAdapter) transformResponse(resp *ServiceabilityResponse, partnerInfo common.PartnerInfo) *common.PartnerServiceabilityResult {
	// TODO: Implement new Shipyaari services structure as per final payload format:
	// services: [
	//   {
	//     "partner_service_id": "e4cfcac0",
	//     "partner_service_name": "XPRESSBEES_SURFACE",
	//     "company_service_id": "a07d01f5",
	//     "company_service_name": "ECONOMY",
	//     "partner_name": "XPRESSBEES",
	//     "service_mode": "SURFACE",
	//     "applied_weight": 1,
	//     "invoice_value": 10,
	//     "collectable_amount": 0,
	//     "insurance": 0,
	//     "base": 10,
	//     "add": 0,
	//     "variables": 0,
	//     "minchargeableamount": 0,
	//     "variable_services": 0,
	//     "cod": 0,
	//     "tax": 1.8,
	//     "total": 11.8,
	//     "zone_name": "ZONE 3",
	//     "edt": 5,
	//     "sort_type": "cheapest",
	//     "priority_rank": null,
	//     "cheapest_rank": 1
	//   }
	// ]

	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		PartnerName:  s.GetPartnerName(),
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// Handle based on success field from Shipyaari response
	if resp.Success {
		// Success response - Shipyaari is serviceable
		// Add services if available
		if len(resp.Services) > 0 {
			for _, service := range resp.Services {
				// Determine delivery mode based on service mode
				deliveryMode := "standard"
				if service.ServiceMode == "AIR" {
					deliveryMode = "express"
				}

				result.Services = append(result.Services, models.ServiceV2{
					ServiceCode:   service.PartnerServiceID,
					ServiceName:   service.PartnerServiceName,
					TATDays:       service.EDT, // Use EDT (Estimated Delivery Time)
					IsCOD:         service.COD > 0, // Check if COD charges exist
					Pickup:        true,
					Delivery:      true,
					Insurance:     service.Insurance > 0,
					ProductTypes:  map[string]bool{"general": true},
					DeliveryModes: map[string]bool{deliveryMode: true},
					Pricing: &models.ServicePricingV2{
						BaseCost:   service.Base,
						Currency:   "INR", // Default to INR
						CODCharges: service.COD,
					},
				})
			}
		}

		// Add capabilities for serviceable response (keeping existing structure)
		result.Capabilities["cod_available"] = getBoolValue(resp.CODAvailable, true)
		result.Capabilities["pickup_available"] = getBoolValue(resp.PickupAvailable, true)
		result.Capabilities["reverse_pickup"] = getBoolValue(resp.ReversePickup, true)
		result.Capabilities["insurance_available"] = getBoolValue(resp.InsuranceAvailable, true)
		result.Capabilities["is_serviceable"] = true

		// Add available services to capabilities
		availableServices := make([]string, 0)
		if resp.Services != nil {
			for _, service := range resp.Services {
				availableServices = append(availableServices, service.PartnerServiceName)
			}
		}
		result.Capabilities["available_services"] = availableServices

		// Add the raw services data from Shipyaari response to partner_services
		result.PartnerServices = resp.Data

	} else {
		// Error response - Shipyaari is not serviceable
		// Set all capabilities to false for non-serviceable response
		result.Capabilities["cod_available"] = false
		result.Capabilities["pickup_available"] = false
		result.Capabilities["reverse_pickup"] = false
		result.Capabilities["insurance_available"] = false
		result.Capabilities["is_serviceable"] = false
		result.Capabilities["available_services"] = []string{}
		result.PartnerServices = []interface{}{} // Empty services array

		// Set error message from Shipyaari response
		if resp.Message != "" {
			result.ErrorMessage = &resp.Message
		}
	}

	// Add metadata
	result.Metadata["api_version"] = "v1"
	if resp.ResponseID != "" {
		result.Metadata["response_id"] = resp.ResponseID
	}
	if resp.Zone != "" {
		result.Metadata["zone"] = resp.Zone
	}

	return result
}

// Helper functions to extract data from standard request
func getSourcePincode(req *models.ServiceabilityV2Request) string {
	if req.SourcePostalCode != nil {
		return *req.SourcePostalCode
	}
	if req.PostalCode != nil {
		return *req.PostalCode
	}
	return ""
}

func getDestinationPincode(req *models.ServiceabilityV2Request) string {
	if req.DestinationPostalCode != nil {
		return *req.DestinationPostalCode
	}
	return ""
}

func getOrderValue(req *models.ServiceabilityV2Request) float64 {
	// Default or extract from request if available
	return 1000.0
}

func getPaymentMode(req *models.ServiceabilityV2Request) string {
	// Default payment mode
	return "COD"
}

func getProductType(req *models.ServiceabilityV2Request) string {
	if req.ProductType != nil {
		return *req.ProductType
	}
	return "general"
}

// getWeightInKg converts weight to kg
func getWeightInKg(weight *models.Weight) float64 {
	if weight == nil {
		return 0.0
	}

	switch weight.Unit {
	case "kg":
		return weight.Value
	case "g":
		return weight.Value / 1000.0
	case "lb":
		return weight.Value * 0.453592
	case "oz":
		return weight.Value * 0.0283495
	default:
		return weight.Value // Assume kg if unit unknown
	}
}

// getDimensionInCm converts dimension to cm
func getDimensionInCm(value float64, unit string) float64 {
	switch unit {
	case "cm":
		return value
	case "mm":
		return value / 10.0
	case "in":
		return value * 2.54
	case "ft":
		return value * 30.48
	default:
		return value // Assume cm if unit unknown
	}
}

// Helper function to safely get boolean value from pointer
func getBoolValue(b *bool, defaultValue bool) bool {
	if b == nil {
		return defaultValue
	}
	return *b
}
