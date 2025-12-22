package delhivery

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// DelhiveryAdapter implements the common.PartnerAdapter interface for Delhivery
type DelhiveryAdapter struct {
	*common.HTTPBaseAdapter
	client *Client
	config config.DelhiveryConfig
	logger *logrus.Logger
}

// NewDelhiveryAdapter creates a new Delhivery adapter
func NewDelhiveryAdapter(cfg config.DelhiveryConfig) common.PartnerAdapter {
	// Create partner config
	partnerConfig := common.GetPartnerConfigDefaults("delhivery", "Delhivery", common.AdapterTypeHTTP)
	partnerConfig.Timeout = cfg.Timeout
	partnerConfig.Enabled = cfg.Enabled

	// Update endpoints
	partnerConfig.Endpoints = map[string]string{
		"base_url":          cfg.BaseURL,
		"serviceability_url": cfg.ServiceabilityURL,
	}

	// Update auth config
	partnerConfig.Auth = common.AuthConfig{
		Type: common.AuthTypeAPIKey,
		Credentials: map[string]string{
			"access_token": cfg.AccessToken,
		},
	}

	// Create base adapter
	baseAdapter := common.NewHTTPBaseAdapter(partnerConfig)

	// Create Delhivery specific components
	auth := NewAuthenticator(cfg)
	client := NewClient(cfg, auth)

	// Set HTTP client and authenticator
	baseAdapter.SetHTTPClient(client)
	baseAdapter.SetAuthenticator(auth)

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	adapter := &DelhiveryAdapter{
		HTTPBaseAdapter: baseAdapter,
		client:          client,
		config:          cfg,
		logger:          logger,
	}

	return adapter
}

// CheckServiceability checks serviceability for the request
func (d *DelhiveryAdapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	d.logger.WithFields(logrus.Fields{
		"component":      "delhivery_adapter",
		"action":         "check_serviceability",
		"partner_code":   partnerInfo.PartnerCode,
		"partner_id":     partnerInfo.PartnerID,
		"source_postal":  request.SourcePostalCode,
		"dest_postal":    request.DestinationPostalCode,
	}).Info("Starting Delhivery serviceability check")

	// Validate request
	if err := common.ValidateServiceabilityRequest(request); err != nil {
		d.RecordRequest(time.Since(startTime), false)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        err,
		}, nil
	}

	// Get pincodes from request
	sourcePincode := getSourcePincode(request)
	destPincode := getDestinationPincode(request)

	if sourcePincode == "" || destPincode == "" {
		errorMsg := "source and destination pincodes are required for Delhivery"
		d.logger.WithFields(logrus.Fields{
			"component": "delhivery_adapter",
			"error":     errorMsg,
		}).Warn("Missing pincodes - returning non-serviceable")
		d.RecordRequest(time.Since(startTime), false)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        fmt.Errorf(errorMsg),
		}, nil
	}

	// Default to surface mode (S) - can be made configurable later
	mot := "S"

	// Make API call
	response, err := d.client.CheckServiceability(ctx, sourcePincode, destPincode, mot)
	if err != nil {
		d.logger.WithFields(logrus.Fields{
			"component": "delhivery_adapter",
			"error":     err.Error(),
		}).Error("API call failed - returning non-serviceable")
		d.RecordRequest(time.Since(startTime), false)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        err,
		}, nil
	}

	// Transform response
	result := d.transformResponse(response, partnerInfo)
	result.ResponseTime = time.Since(startTime)

	// Record successful request
	d.RecordRequest(time.Since(startTime), true)

	return result, nil
}

// Initialize performs any necessary initialization
func (d *DelhiveryAdapter) Initialize(ctx context.Context) error {
	if err := d.HTTPBaseAdapter.Initialize(ctx); err != nil {
		return err
	}

	// Authenticate with Delhivery (just verify token is configured)
	if err := d.GetAuthenticator().Authenticate(ctx); err != nil {
		d.SetHealthStatus("unhealthy")
		return err
	}

	d.SetHealthStatus("healthy")
	return nil
}

// IsHealthy checks if the Delhivery adapter is healthy
func (d *DelhiveryAdapter) IsHealthy(ctx context.Context) bool {
	// Check base health
	if !d.HTTPBaseAdapter.IsHealthy(ctx) {
		return false
	}

	// Check if authenticated
	return d.GetAuthenticator().IsAuthenticated()
}

// transformResponse converts Delhivery response to standard format
func (d *DelhiveryAdapter) transformResponse(resp *ServiceabilityResponse, partnerInfo common.PartnerInfo) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// Check if response indicates success
	if !resp.Success {
		errorMsg := resp.Msg
		if errorMsg == "" {
			errorMsg = "Delhivery API returned unsuccessful response"
		}
		d.logger.WithFields(logrus.Fields{
			"component": "delhivery_adapter",
			"success":   resp.Success,
			"message":   resp.Msg,
		}).Info("Non-serviceable response")
		result.ErrorMessage = &errorMsg
		return result
	}

	// If TAT is 0 or negative, consider it non-serviceable
	if resp.Data.TAT <= 0 {
		errorMsg := "Route is not serviceable (TAT is 0 or negative)"
		result.ErrorMessage = &errorMsg
		return result
	}

	// Route is serviceable - create a service entry
	result.Services = append(result.Services, models.ServiceV2{
		ServiceCode:   "delhivery_standard",
		ServiceName:   "Delhivery Standard",
		TATDays:       resp.Data.TAT,
		Pickup:        true,
		Delivery:      true,
		ProductTypes:  map[string]bool{"general": true},
		DeliveryModes: map[string]bool{"standard": true},
	})

	// Set capabilities
	result.Capabilities["is_serviceable"] = true
	result.Capabilities["pickup_available"] = true
	result.Capabilities["delivery_available"] = true
	result.Capabilities["tat_days"] = resp.Data.TAT

	// Set metadata
	result.Metadata["api_version"] = "v1"
	result.Metadata["adapter_type"] = "http_token_auth"
	result.Metadata["response_success"] = resp.Success
	result.Metadata["tat"] = resp.Data.TAT

	// Store raw partner services
	result.PartnerServices = resp.Data

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

