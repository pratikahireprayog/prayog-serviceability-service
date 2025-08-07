package delcaper

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

// Adapter implements the PartnerAdapter interface for Delcaper
type Adapter struct {
	client *DelcaperClient
	config config.DelcaperConfig
	logger *logrus.Logger
}

// NewAdapter creates a new Delcaper adapter instance
func NewAdapter(config config.DelcaperConfig) *Adapter {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Log configuration with redacted sensitive fields
	logger.WithFields(logrus.Fields{
		"partner":    "Delcaper",
		"base_url":   config.BaseURL,
		"email":      redactField(config.Email),
		"enabled":    config.Enabled,
		"timeout":    config.Timeout,
	}).Info("Creating Delcaper adapter")

	return &Adapter{
		client: NewDelcaperClient(config, &http.Client{
			Timeout: config.Timeout,
		}),
		config: config,
		logger: logger,
	}
}

// redactField safely redacts sensitive fields for logging
func redactField(field string) string {
	if field == "" {
		return "[EMPTY]"
	}
	if len(field) <= 3 {
		return "[REDACTED]"
	}
	return field[:3] + "***"
}

// GetPartnerCode returns the partner code
func (a *Adapter) GetPartnerCode() string {
	return "smile_hyperlocal"
}

// GetPartnerName returns the partner name
func (a *Adapter) GetPartnerName() string {
	return "Smile Hyperlocal"
}

// GetAdapterType returns the adapter type
func (a *Adapter) GetAdapterType() common.AdapterType {
	return common.AdapterTypeHTTP
}

// IsEnabled returns whether the adapter is enabled
func (a *Adapter) IsEnabled() bool {
	return a.config.Enabled
}

// SupportsRequest checks if Delcaper supports the given request
func (a *Adapter) SupportsRequest(ctx context.Context, request *models.ServiceabilityV2Request) bool {
	// Check if we have required postal codes
	if !a.hasValidPincodes(request) {
		return false
	}

	// Check if this is a hyperlocal request
	if request.ParcelCategory != nil && *request.ParcelCategory == "hyperlocal" {
		return true
	}

	return false
}

// CheckServiceability checks serviceability for the request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	a.logger.WithFields(logrus.Fields{
		"partner_code": partnerInfo.PartnerCode,
		"source_pincode": func() string {
			if request.SourcePostalCode != nil {
				return *request.SourcePostalCode
			}
			return ""
		}(),
		"destination_pincode": func() string {
			if request.DestinationPostalCode != nil {
				return *request.DestinationPostalCode
			}
			if request.PostalCode != nil {
				return *request.PostalCode
			}
			return ""
		}(),
		"parcel_category": func() string {
			if request.ParcelCategory != nil {
				return *request.ParcelCategory
			}
			return ""
		}(),
	}).Info("Starting Delcaper serviceability check")

	if !a.SupportsRequest(ctx, request) {
		a.logger.WithFields(logrus.Fields{
			"partner_code": partnerInfo.PartnerCode,
			"reason":       "Request not supported by Delcaper (not hyperlocal or missing pincodes)",
		}).Info("Delcaper request not supported")
		
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: 0,
			Metadata: map[string]interface{}{
				"reason": "Request not supported by Delcaper (not hyperlocal or missing pincodes)",
			},
		}, nil
	}

	// Convert request to Delcaper format
	delcaperRequest := a.convertToCheckFeasibleRequest(request)

	// Use timeout context for the API call
	apiCtx, cancel := context.WithTimeout(ctx, a.config.Timeout)
	defer cancel()

	// Track start time for response time calculation
	startTime := time.Now()

	a.logger.WithFields(logrus.Fields{
		"partner_code": partnerInfo.PartnerCode,
		"timeout":      a.config.Timeout,
	}).Info("Making Delcaper API call")

	// Make API call
	response, err := a.client.CheckFeasible(apiCtx, delcaperRequest)
	
	// Calculate response time
	responseTime := time.Since(startTime)
	
	if err != nil {
		// Check if it's a timeout error
		if ctx.Err() == context.DeadlineExceeded || apiCtx.Err() == context.DeadlineExceeded {
			a.logger.WithFields(logrus.Fields{
				"partner_code":  partnerInfo.PartnerCode,
				"response_time": responseTime,
				"timeout":       a.config.Timeout,
				"error":         err.Error(),
			}).Error("Delcaper API timeout")
			
			return &common.PartnerServiceabilityResult{
				PartnerID:    partnerInfo.PartnerID,
				PartnerCode:  partnerInfo.PartnerCode,
				Services:     make([]models.ServiceV2, 0),
				Error:        err,
				ErrorMessage: &[]string{fmt.Sprintf("Delcaper API timeout after %v: %v", a.config.Timeout, err)}[0],
				ResponseTime: responseTime,
			}, nil
		}
		
		a.logger.WithFields(logrus.Fields{
			"partner_code":  partnerInfo.PartnerCode,
			"response_time": responseTime,
			"error":         err.Error(),
		}).Error("Delcaper API call failed")
		
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Delcaper API call failed: %v", err)}[0],
			ResponseTime: responseTime,
		}, nil
	}

	a.logger.WithFields(logrus.Fields{
		"partner_code":  partnerInfo.PartnerCode,
		"response_time": responseTime,
		"feasible":      response.Data.Feasible,
		"services_count": len(response.Data.Services),
	}).Info("Delcaper API call successful")

	// Convert response and return
	result := a.convertCheckFeasibleResponse(response, partnerInfo)
	result.ResponseTime = responseTime
	return result, nil
}

// IsHealthy checks if the adapter is healthy
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	// Check if configuration is valid
	if !a.config.Enabled || a.config.BaseURL == "" || a.config.Email == "" || a.config.Password == "" {
		a.logger.WithFields(logrus.Fields{
			"enabled":   a.config.Enabled,
			"base_url":  a.config.BaseURL != "",
			"email":     a.config.Email != "",
			"password":  a.config.Password != "",
		}).Warn("Delcaper adapter configuration invalid")
		return false
	}
	
	// Perform a quick health check by testing the connection
	// Use a short timeout for health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	isHealthy := a.client.IsHealthy(healthCtx)
	
	a.logger.WithFields(logrus.Fields{
		"healthy": isHealthy,
	}).Info("Delcaper adapter health check")
	
	return isHealthy
}

// GetMetrics returns adapter metrics
func (a *Adapter) GetMetrics() *common.PartnerMetrics {
	// Return basic metrics for now
	return &common.PartnerMetrics{
		PartnerCode:         a.GetPartnerCode(),
		TotalRequests:       0, // TODO: Implement metrics tracking
		SuccessfulRequests:  0,
		FailedRequests:      0,
		AverageResponseTime: 0,
		LastRequestTime:     time.Time{},
		HealthStatus:        "healthy",
		ErrorRate:           0.0,
		Uptime:             0,
	}
}

// Initialize initializes the adapter
func (a *Adapter) Initialize(ctx context.Context) error {
	a.logger.Info("Initializing Delcaper adapter")
	
	// Perform any initialization tasks
	// For Delcaper, we might want to test the connection or validate credentials
	if !a.config.Enabled {
		err := fmt.Errorf("Delcaper adapter is disabled")
		a.logger.WithError(err).Error("Delcaper adapter initialization failed")
		return err
	}
	
	// Test health check
	if !a.IsHealthy(ctx) {
		err := fmt.Errorf("Delcaper adapter health check failed")
		a.logger.WithError(err).Error("Delcaper adapter initialization failed")
		return err
	}
	
	a.logger.Info("Delcaper adapter initialized successfully")
	return nil
}

// Shutdown gracefully shuts down the adapter
func (a *Adapter) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down Delcaper adapter")
	
	// Perform any cleanup tasks
	// For Delcaper, we might want to clear tokens or close connections
	a.logger.Info("Delcaper adapter shutdown completed")
	return nil
}

// hasValidPincodes checks if the request has valid postal codes
func (a *Adapter) hasValidPincodes(request *models.ServiceabilityV2Request) bool {
	// Check source postal code
	if request.SourcePostalCode == nil || *request.SourcePostalCode == "" {
		return false
	}

	// Check destination postal code (either DestinationPostalCode or PostalCode)
	if request.DestinationPostalCode == nil || *request.DestinationPostalCode == "" {
		if request.PostalCode == nil || *request.PostalCode == "" {
			return false
		}
	}

	return true
}

// convertToCheckFeasibleRequest converts v2 request to Delcaper format
func (a *Adapter) convertToCheckFeasibleRequest(request *models.ServiceabilityV2Request) *CheckFeasibleRequest {
	req := &CheckFeasibleRequest{
		OrderType: "HYPERLOCAL",
	}

	// Set pickup address (source)
	if request.SourcePostalCode != nil {
		req.PickupAddress = AddressInfo{
			Zip: *request.SourcePostalCode,
		}
	}

	// Set shipping address (destination)
	if request.DestinationPostalCode != nil {
		req.ShippingAddress = AddressInfo{
			Zip: *request.DestinationPostalCode,
		}
	} else if request.PostalCode != nil {
		req.ShippingAddress = AddressInfo{
			Zip: *request.PostalCode,
		}
	}

	return req
}

// convertCheckFeasibleResponse converts Delcaper response to common format
func (a *Adapter) convertCheckFeasibleResponse(response *CheckFeasibleResponse, partnerInfo common.PartnerInfo) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// Handle response based on feasibility status
	if response.Data.Feasible {
		// Add services if available
		for _, service := range response.Data.Services {
			v2Service := models.ServiceV2{
				ServiceCode: service.ServiceCode,
				ServiceName: service.ServiceName,
				TATDays:     service.TATDays,
				IsCOD:       service.IsCOD,
				Pickup:      service.Pickup,
				Delivery:    service.Delivery,
				Insurance:   service.Insurance,
				ProductTypes: map[string]bool{
					"hyperlocal": true,
					"local":      true,
				},
				DeliveryModes: map[string]bool{
					"standard": true,
				},
			}

			// Add pricing if available
			if service.Pricing != nil {
				v2Service.Pricing = &models.ServicePricingV2{
					BaseCost:      service.Pricing.BaseCost,
					Currency:      service.Pricing.Currency,
					CODCharges:    service.Pricing.CODCharges,
					FuelSurcharge: service.Pricing.FuelSurcharge,
				}
			}

			result.Services = append(result.Services, v2Service)
		}

		// Add capabilities
		result.Capabilities["hyperlocal"] = true
		result.Capabilities["local_delivery"] = true
		result.Capabilities["same_day"] = response.Data.Feasible
	} else {
		// Service not feasible
		errorMsg := response.Data.Message
		if errorMsg == "" {
			errorMsg = "Service not feasible for hyperlocal delivery"
		}
		result.ErrorMessage = &errorMsg
	}

	// Add metadata
	result.Metadata["api_version"] = "v1"
	result.Metadata["adapter_type"] = "http_with_auth"
	result.Metadata["order_type"] = "HYPERLOCAL"
	result.Metadata["distance_km"] = response.Data.Distance
	result.Metadata["serviceability_issue"] = response.Data.ServiceabilityIssue

	// Handle errors
	if response.Status != 200 {
		errorMsg := fmt.Sprintf("Delcaper API returned status %d", response.Status)
		result.ErrorMessage = &errorMsg
	}

	return result
} 