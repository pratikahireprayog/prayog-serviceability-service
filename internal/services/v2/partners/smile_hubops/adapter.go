package smile_hubops

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/sirupsen/logrus"
)

// Adapter implements the PartnerAdapter interface for Smile HubOps
type Adapter struct {
	config *config.SmileHubOpsConfig
	client *HubOpsClient
	logger *logrus.Logger
}

// NewAdapter creates a new Smile HubOps adapter
func NewAdapter(config config.SmileHubOpsConfig) *Adapter {
	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create HubOps client
	client := NewHubOpsClient(config.BaseURL, httpClient)

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Adapter{
		config: &config,
		client: client,
		logger: logger,
	}
}

// CheckServiceability checks serviceability using Smile HubOps API
func (a *Adapter) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	a.logger.WithFields(logrus.Fields{
		"component":        "smile_hubops_adapter",
		"action":           "check_serviceability",
		"partner_code":     partnerInfo.PartnerCode,
		"partner_id":       partnerInfo.PartnerID,
		"source_postal":    req.SourcePostalCode,
		"dest_postal":      req.DestinationPostalCode,
		"parcel_category":  req.ParcelCategory,
		"country_code":     req.CountryCode,
		"product_type":     req.ProductType,
	}).Info("Starting Smile HubOps serviceability check")

	// Validate request
	if err := a.validateRequest(req); err != nil {
		a.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_adapter",
			"error":     err.Error(),
		}).Error("Request validation failed")
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Determine source and destination postal codes
	sourcePostalCode, destPostalCode, err := a.getPostalCodes(req)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_adapter",
			"error":     err.Error(),
		}).Error("Failed to get postal codes")
		return nil, fmt.Errorf("failed to get postal codes: %w", err)
	}

	// Call HubOps API
	hubOpsResponse, err := a.client.CheckServiceability(ctx, sourcePostalCode, destPostalCode)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_adapter",
			"error":     err.Error(),
		}).Error("HubOps API call failed")
		return nil, fmt.Errorf("HubOps API call failed: %w", err)
	}

	// Transform response to common format
	result := a.transformResponse(hubOpsResponse, partnerInfo, time.Since(startTime))

	a.logger.WithFields(logrus.Fields{
		"component":           "smile_hubops_adapter",
		"delivery_available":  hubOpsResponse.DeliveryAvailable,
		"route_count":         len(hubOpsResponse.Routes),
		"response_time":       time.Since(startTime),
		"has_services":        len(result.Services) > 0,
		"has_capabilities":    len(result.Capabilities) > 0,
	}).Info("Smile HubOps serviceability check completed")

	return result, nil
}

// IsHealthy checks if the adapter is healthy
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	// Check configuration
	if a.config == nil || !a.config.Enabled {
		a.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_adapter",
		}).Warn("Smile HubOps adapter not enabled in configuration")
		return false
	}

	// Check if base URL is configured
	if a.config.BaseURL == "" {
		a.logger.WithFields(logrus.Fields{
			"component": "smile_hubops_adapter",
		}).Error("Smile HubOps base URL not configured")
		return false
	}

	// Perform health check
	isHealthy := a.client.IsHealthy(ctx)

	a.logger.WithFields(logrus.Fields{
		"component":  "smile_hubops_adapter",
		"is_healthy": isHealthy,
	}).Info("Smile HubOps adapter health check completed")

	return isHealthy
}

// Initialize initializes the adapter
func (a *Adapter) Initialize(ctx context.Context) error {
	a.logger.WithFields(logrus.Fields{
		"component": "smile_hubops_adapter",
		"action":    "initialize",
		"base_url":  a.config.BaseURL,
		"enabled":   a.config.Enabled,
	}).Info("Initializing Smile HubOps adapter")

	// Perform health check during initialization
	if !a.IsHealthy(ctx) {
		return fmt.Errorf("Smile HubOps adapter health check failed")
	}

	a.logger.WithFields(logrus.Fields{
		"component": "smile_hubops_adapter",
	}).Info("Smile HubOps adapter initialized successfully")

	return nil
}

// Shutdown shuts down the adapter
func (a *Adapter) Shutdown(ctx context.Context) error {
	a.logger.WithFields(logrus.Fields{
		"component": "smile_hubops_adapter",
		"action":    "shutdown",
	}).Info("Shutting down Smile HubOps adapter")

	// No cleanup needed for HTTP client
	return nil
}

// validateRequest validates the serviceability request
func (a *Adapter) validateRequest(req *models.ServiceabilityV2Request) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Check if we have postal codes
	hasSourceDest := req.SourcePostalCode != nil && *req.SourcePostalCode != "" &&
		req.DestinationPostalCode != nil && *req.DestinationPostalCode != ""

	if !hasSourceDest {
		return fmt.Errorf("both source and destination postal codes are required")
	}

	return nil
}

// getPostalCodes extracts source and destination postal codes from the request
func (a *Adapter) getPostalCodes(req *models.ServiceabilityV2Request) (string, string, error) {
	if req.SourcePostalCode == nil || *req.SourcePostalCode == "" {
		return "", "", fmt.Errorf("source postal code is required")
	}

	if req.DestinationPostalCode == nil || *req.DestinationPostalCode == "" {
		return "", "", fmt.Errorf("destination postal code is required")
	}

	return *req.SourcePostalCode, *req.DestinationPostalCode, nil
}

// transformResponse transforms HubOps response to common format
func (a *Adapter) transformResponse(hubOpsResponse *HubOpsResponse, partnerInfo common.PartnerInfo, responseTime time.Duration) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		ResponseTime: responseTime,
		Services:     []models.ServiceV2{},
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// Return the raw HubOps response under hub_details key
	result.Metadata["hub_details"] = hubOpsResponse

	return result
}

// transformHubInfo transforms HubInfo to use snake_case field names
func (a *Adapter) transformHubInfo(hubInfo *HubInfo) map[string]interface{} {
	if hubInfo == nil {
		return nil
	}

	return map[string]interface{}{
		"premise_id":              hubInfo.PremiseID,
		"premise_name":            hubInfo.PremiseName,
		"parent_premise_name":     hubInfo.ParentPremiseName,
		"personal_number":         hubInfo.PersonalNumber,
		"official_number":         hubInfo.OfficialNumber,
		"city":                    hubInfo.City,
		"address":                 hubInfo.Address,
		"address_line1":           hubInfo.AddressLine1,
		"address_line2":           hubInfo.AddressLine2,
		"billing_cycle":           hubInfo.BillingCycle,
		"pincode":                 hubInfo.Pincode,
		"state":                   hubInfo.State,
		"zone":                    hubInfo.Zone,
		"type":                    hubInfo.Type,
		"parent_id":               hubInfo.ParentID,
		"parent_id_air":           hubInfo.ParentIDAir,
		"gst":                     hubInfo.GST,
		"state_code":              hubInfo.StateCode,
		"cutoff_time":             hubInfo.CutoffTime,
		"is_metro":                hubInfo.IsMetro,
		"fov":                     hubInfo.FOV,
		"cod":                     hubInfo.COD,
		"premium":                 hubInfo.Premium,
		"latitude":                hubInfo.Latitude,
		"longitude":               hubInfo.Longitude,
		"personal_email_id":       hubInfo.PersonalEmailID,
		"official_email_id":       hubInfo.OfficialEmailID,
		"pan":                     hubInfo.PAN,
		"cp_type":                 hubInfo.CPType,
		"rate_card_type":          hubInfo.RateCardType,
		"force_update_allowed":    hubInfo.ForceUpdateAllowed,
		"is_terminal_hub":         hubInfo.IsTerminalHub,
		"created_date":            hubInfo.CreatedDate,
		"status":                  hubInfo.Status,
		"areas":                   hubInfo.Areas,
		"hub_type":                hubInfo.HubType,
		"hub_mode":                hubInfo.HubMode,
		"rate_card_id":            hubInfo.RateCardID,
		"ho_id":                   hubInfo.HOID,
		"hub_pincode_map":         hubInfo.HubPincodeMap,
		"wallet_mapping_cust_id":  hubInfo.WalletMappingCustID,
		"miscellaneous_details":    hubInfo.MiscellaneousDetails,
		"center_map":              hubInfo.CenterMap,
		"sp_id":                   hubInfo.SPID,
	}
}

// transformRoutes transforms routes to use snake_case field names
func (a *Adapter) transformRoutes(routes []Route) []map[string]interface{} {
	var transformedRoutes []map[string]interface{}
	
	for _, route := range routes {
		transformedRoute := map[string]interface{}{
			"tat_days":     route.TATDays,
			"mode":         route.Mode,
			"route":        a.transformRouteHubs(route.Route),
			"cutoff_time":  route.CutoffTime,
		}
		transformedRoutes = append(transformedRoutes, transformedRoute)
	}
	
	return transformedRoutes
}

// transformRouteHubs transforms route hubs to use snake_case field names
func (a *Adapter) transformRouteHubs(hubs []HubInfo) []map[string]interface{} {
	var transformedHubs []map[string]interface{}
	
	for _, hub := range hubs {
		transformedHubs = append(transformedHubs, a.transformHubInfo(&hub))
	}
	
	return transformedHubs
}

// hasAirMode checks if any route has air mode
func (a *Adapter) hasAirMode(routes []Route) bool {
	for _, route := range routes {
		if route.Mode == 1 { // Assuming 1 is air mode
			return true
		}
	}
	return false
}

// hasSurfaceMode checks if any route has surface mode
func (a *Adapter) hasSurfaceMode(routes []Route) bool {
	for _, route := range routes {
		if route.Mode == 0 { // Assuming 0 is surface mode
			return true
		}
	}
	return false
}

// createServices creates services from routes
func (a *Adapter) createServices(routes []Route) []models.ServiceV2 {
	var services []models.ServiceV2

	for i, route := range routes {
		service := models.ServiceV2{
			ServiceCode:   fmt.Sprintf("hubops_route_%d", i+1),
			ServiceName:   fmt.Sprintf("HubOps Route %d", i+1),
			TATDays:       route.TATDays,
			IsCOD:         true, // HubOps supports COD
			Pickup:        true, // HubOps supports pickup
			Delivery:      true, // HubOps supports delivery
			Insurance:     true, // HubOps supports insurance
			ProductTypes:  map[string]bool{"general": true},
			DeliveryModes: map[string]bool{
				"air":    route.Mode == 1,
				"surface": route.Mode == 0,
			},
		}

		services = append(services, service)
	}

	return services
}

// GetMetrics returns metrics for the adapter
func (a *Adapter) GetMetrics() *common.PartnerMetrics {
	// For now, return basic metrics
	// In a real implementation, you would track actual metrics
	return &common.PartnerMetrics{
		PartnerCode:         "smile_hubops",
		TotalRequests:       0, // TODO: Implement actual tracking
		SuccessfulRequests:  0, // TODO: Implement actual tracking
		FailedRequests:      0, // TODO: Implement actual tracking
		AverageResponseTime: 0, // TODO: Implement actual tracking
		LastRequestTime:     time.Time{}, // TODO: Implement actual tracking
		HealthStatus:        "unknown", // TODO: Implement actual tracking
		ErrorRate:           0.0, // TODO: Implement actual tracking
		Uptime:             0, // TODO: Implement actual tracking
	}
}
