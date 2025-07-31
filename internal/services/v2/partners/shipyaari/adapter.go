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
	if err := s.validateShipyaariRequirements(request); err != nil {
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
	shipyaariRequest, err := s.transformRequest(request)
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
	// Use postal codes as strings (Shipyaari expects strings, not integers)
	pickupPincode := getSourcePincode(req)
	deliveryPincode := getDestinationPincode(req)

	shipyaariReq := &ServiceabilityRequest{
		PickupPincode:   pickupPincode,
		DeliveryPincode: deliveryPincode,
		InvoiceValue:    getOrderValue(req),
		PaymentMode:     getPaymentMode(req),
		Weight:          getWeightInKg(req.Packages[0].Weight),
		OrderType:       "B2C", // Default to B2C
		Dimension: Dimension{
			Length: getDimensionInCm(req.Packages[0].Dimensions.Length, req.Packages[0].Dimensions.Unit),
			Width:  getDimensionInCm(req.Packages[0].Dimensions.Width, req.Packages[0].Dimensions.Unit),
			Height: getDimensionInCm(req.Packages[0].Dimensions.Height, req.Packages[0].Dimensions.Unit),
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

	// Add services if available
	if resp.IsServiceable && len(resp.Services) > 0 {
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

	// Add capabilities
	result.Capabilities["cod_available"] = resp.CODAvailable
	result.Capabilities["pickup_available"] = resp.PickupAvailable
	result.Capabilities["reverse_pickup"] = resp.ReversePickup
	result.Capabilities["insurance_available"] = resp.InsuranceAvailable

	// Add metadata
	result.Metadata["api_version"] = "v1"
	result.Metadata["response_id"] = resp.ResponseID
	result.Metadata["zone"] = resp.Zone

	// Handle errors
	if !resp.Success && resp.Message != "" {
		result.ErrorMessage = &resp.Message
	} else if resp.ErrorMessage != "" {
		result.ErrorMessage = &resp.ErrorMessage
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
