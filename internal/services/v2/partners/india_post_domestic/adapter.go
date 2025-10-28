package india_post_domestic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/sirupsen/logrus"
)

// Adapter implements the PartnerAdapter interface for India Post Domestic
type Adapter struct {
	client *Client
	auth   *Authenticator
	config config.IndiaPostDomesticConfig
	logger *logrus.Logger
}

// NewAdapter creates a new India Post Domestic adapter instance
func NewAdapter(cfg config.IndiaPostDomesticConfig) *Adapter {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.WithFields(logrus.Fields{
		"partner":  "IndiaPostDomestic",
		"base_url": cfg.BaseURL,
		"enabled":  cfg.Enabled,
	}).Info("Creating India Post Domestic adapter")

	// Create authenticator
	auth := NewAuthenticator(cfg)

	// Create client with authenticator
	client := NewClient(cfg, auth)

	return &Adapter{
		client: client,
		auth:   auth,
		config: cfg,
		logger: logger,
	}
}

// GetAdapterType returns the adapter type
func (a *Adapter) GetAdapterType() common.AdapterType {
	return common.AdapterTypeHTTP
}

// IsEnabled returns whether the adapter is enabled
func (a *Adapter) IsEnabled() bool {
	return a.config.Enabled
}

// CheckServiceability checks if India Post Domestic can service the given request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	partnerID := ""
	if partnerInfo.PartnerID != nil {
		partnerID = partnerInfo.PartnerID.String()
	}

	a.logger.WithFields(logrus.Fields{
		"component":    "india_post_domestic_adapter",
		"action":       "check_serviceability_start",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
	}).Info("Starting India Post Domestic adapter serviceability check")

	// Validate India Post Domestic specific requirements
	if err := a.validateRequirements(request); err != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "india_post_domestic_adapter",
			"event":        "validation_failed",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   partnerID,
			"error":        err.Error(),
		}).Warn("India Post Domestic validation failed")

		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("India Post Domestic validation failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "India Post Domestic validation failed",
			},
		}, nil
	}

	// Determine which pincode to check
	// For India Post Domestic, we need the destination pincode
	var pincodeToCheck string
	if request.DestinationPostalCode != nil && *request.DestinationPostalCode != "" {
		pincodeToCheck = *request.DestinationPostalCode
	} else if request.PostalCode != nil && *request.PostalCode != "" {
		pincodeToCheck = *request.PostalCode
	} else {
		err := fmt.Errorf("destination postal code is required for India Post Domestic")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{err.Error()}[0],
		}, nil
	}

	a.logger.WithFields(logrus.Fields{
		"component":    "india_post_domestic_adapter",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
		"pincode":      pincodeToCheck,
	}).Info("Checking serviceability for India Post Domestic")

	// Call India Post Domestic API to search for pincode
	response, err := a.client.SearchPincode(ctx, pincodeToCheck)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "india_post_domestic_adapter",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   partnerID,
			"pincode":      pincodeToCheck,
			"error":        err.Error(),
		}).Error("India Post Domestic API call failed")

		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("India Post Domestic API call failed: %v", err)}[0],
		}, nil
	}

	// Convert response to serviceability result
	result := a.convertToServiceabilityResult(response, partnerInfo, pincodeToCheck)
	result.ResponseTime = time.Since(startTime)

	a.logger.WithFields(logrus.Fields{
		"component":      "india_post_domestic_adapter",
		"partner_code":   partnerInfo.PartnerCode,
		"partner_id":     partnerID,
		"pincode":        pincodeToCheck,
		"is_serviceable": len(result.Services) > 0,
		"offices_found":  response.ReturnedRecordsCount,
		"response_time":  result.ResponseTime,
	}).Info("India Post Domestic serviceability check completed")

	return result, nil
}

// validateRequirements validates India Post Domestic specific requirements
func (a *Adapter) validateRequirements(request *models.ServiceabilityV2Request) error {
	// India Post Domestic requires either destination postal code or generic postal code
	if (request.DestinationPostalCode == nil || *request.DestinationPostalCode == "") &&
		(request.PostalCode == nil || *request.PostalCode == "") {
		return fmt.Errorf("postal code is required for India Post Domestic serviceability")
	}

	// Validate pincode format (6 digits for Indian pincodes)
	var pincodeToCheck string
	if request.DestinationPostalCode != nil && *request.DestinationPostalCode != "" {
		pincodeToCheck = *request.DestinationPostalCode
	} else if request.PostalCode != nil {
		pincodeToCheck = *request.PostalCode
	}

	// Check if pincode is numeric and 6 digits
	if _, err := strconv.Atoi(pincodeToCheck); err != nil {
		return fmt.Errorf("invalid pincode format: must be numeric")
	}

	if len(pincodeToCheck) != 6 {
		return fmt.Errorf("invalid pincode format: must be 6 digits")
	}

	return nil
}

// convertToServiceabilityResult converts India Post Domestic response to common format
func (a *Adapter) convertToServiceabilityResult(response *PincodeSearchResponse, partnerInfo common.PartnerInfo, pincode string) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// If no offices found, pincode is not serviceable
	if response.ReturnedRecordsCount == 0 || len(response.Data) == 0 {
		result.Metadata["reason"] = "Pincode not found in India Post Domestic database"
		result.Metadata["pincode"] = pincode
		result.Metadata["is_serviceable"] = false
		return result
	}

	// India Post Domestic returned data - pincode is serviceable
	// Find delivery offices
	var deliveryOffices []PostalOffice
	var allOffices []PostalOffice
	
	for _, office := range response.Data {
		allOffices = append(allOffices, office)
		if office.DeliveryOfficeFlag && office.IsRolledOut {
			deliveryOffices = append(deliveryOffices, office)
		}
	}

	// Build capabilities from the offices found
	capabilities := map[string]interface{}{
		"pincode":              response.Data[0].Pincode,
		"state_name":           response.Data[0].StateName,
		"city_name":            response.Data[0].CityName,
		"taluk_name":           response.Data[0].TalukName,
		"total_offices":        response.ReturnedRecordsCount,
		"delivery_offices":     len(deliveryOffices),
		"has_delivery_office":  len(deliveryOffices) > 0,
		"is_rolled_out":        response.Data[0].IsRolledOut,
	}

	// Add office details
	if len(deliveryOffices) > 0 {
		capabilities["primary_delivery_office"] = deliveryOffices[0].OfficeName
		capabilities["primary_office_type"] = deliveryOffices[0].OfficeTypeCode
	} else if len(allOffices) > 0 {
		capabilities["primary_office"] = allOffices[0].OfficeName
		capabilities["primary_office_type"] = allOffices[0].OfficeTypeCode
	}

	result.Capabilities = capabilities

	// Create a generic service entry since India Post Domestic is serviceable at this pincode
	// Services array can be populated based on business logic
	// For now, we indicate serviceability through capabilities
	
	// Add metadata
	result.Metadata["pincode"] = pincode
	result.Metadata["is_serviceable"] = true
	result.Metadata["offices_data"] = response.Data
	result.Metadata["search_success"] = response.Success
	result.Metadata["message"] = response.Message

	return result
}

// Initialize implements PartnerAdapter interface
func (a *Adapter) Initialize(ctx context.Context) error {
	// Test authentication
	if err := a.auth.Authenticate(ctx); err != nil {
		return fmt.Errorf("India Post Domestic authentication failed: %w", err)
	}

	a.logger.Info("India Post Domestic adapter initialized successfully")
	return nil
}

// IsHealthy implements PartnerAdapter interface
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	if !a.config.Enabled {
		return false
	}

	// Check if we can authenticate
	if !a.auth.IsAuthenticated() {
		if err := a.auth.Authenticate(ctx); err != nil {
			a.logger.WithError(err).Warn("India Post Domestic health check failed: authentication error")
			return false
		}
	}

	return true
}

// GetMetrics implements PartnerAdapter interface
func (a *Adapter) GetMetrics() *common.PartnerMetrics {
	return &common.PartnerMetrics{
		PartnerCode:         "", // Will be set by orchestrator from database
		TotalRequests:       0,  // TODO: Implement actual metrics
		SuccessfulRequests:  0,
		FailedRequests:      0,
		AverageResponseTime: 0,
		HealthStatus:        "healthy",
		ErrorRate:           0.0,
	}
}

// Shutdown implements PartnerAdapter interface
func (a *Adapter) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down India Post Domestic adapter")
	// No specific cleanup needed for HTTP client
	return nil
}

