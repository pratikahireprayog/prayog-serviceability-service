package urbanbolt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// UrbanBoltAdapter implements the common.PartnerAdapter interface for UrbanBolt
type UrbanBoltAdapter struct {
	*common.HTTPBaseAdapter
	client *Client
	config config.UrbanBoltConfig
	logger *logrus.Logger
}

// NewUrbanBoltAdapter creates a new UrbanBolt adapter
func NewUrbanBoltAdapter(cfg config.UrbanBoltConfig) common.PartnerAdapter {
	// Create partner config
	partnerConfig := common.GetPartnerConfigDefaults("urbanbolt", "UrbanBolt", common.AdapterTypeHTTP)
	partnerConfig.Timeout = cfg.Timeout
	partnerConfig.Enabled = cfg.Enabled

	// Update endpoints
	partnerConfig.Endpoints = map[string]string{
		"base_url":           cfg.BaseURL,
		"auth_token_path":     cfg.AuthTokenPath,
		"serviceability_url":  cfg.ServiceabilityURL,
	}

	// Update auth config
	partnerConfig.Auth = common.AuthConfig{
		Type: common.AuthTypeJWT,
		Credentials: map[string]string{
			"username": cfg.Username,
			"password": cfg.Password,
		},
		TokenURL: cfg.AuthTokenPath,
	}

	// Create base adapter
	baseAdapter := common.NewHTTPBaseAdapter(partnerConfig)

	// Create UrbanBolt specific components
	auth := NewAuthenticator(cfg)
	client := NewClient(cfg, auth)

	// Set HTTP client and authenticator
	baseAdapter.SetHTTPClient(client)
	baseAdapter.SetAuthenticator(auth)

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	adapter := &UrbanBoltAdapter{
		HTTPBaseAdapter: baseAdapter,
		client:          client,
		config:          cfg,
		logger:          logger,
	}

	return adapter
}

// CheckServiceability checks serviceability for the request
func (u *UrbanBoltAdapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	u.logger.WithFields(logrus.Fields{
		"component":       "urbanbolt_adapter",
		"action":          "check_serviceability",
		"partner_code":    partnerInfo.PartnerCode,
		"partner_id":      partnerInfo.PartnerID,
		"source_postal":  request.SourcePostalCode,
		"dest_postal":     request.DestinationPostalCode,
	}).Info("Starting UrbanBolt serviceability check")

	// Validate request
	if err := common.ValidateServiceabilityRequest(request); err != nil {
		u.RecordRequest(time.Since(startTime), false)
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
		errorMsg := "source and destination pincodes are required for UrbanBolt"
		u.logger.WithFields(logrus.Fields{
			"component": "urbanbolt_adapter",
			"error":     errorMsg,
		}).Warn("Missing pincodes - returning non-serviceable")
		u.RecordRequest(time.Since(startTime), false)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        fmt.Errorf(errorMsg),
		}, nil
	}

	// Prepare pincodes list (UrbanBolt accepts multiple pincodes)
	pincodes := []string{sourcePincode, destPincode}

	// Make API call
	response, err := u.client.CheckServiceability(ctx, pincodes)
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"component": "urbanbolt_adapter",
			"error":     err.Error(),
		}).Error("API call failed - returning non-serviceable")
		u.RecordRequest(time.Since(startTime), false)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        err,
		}, nil
	}

	// Transform response
	result := u.transformResponse(response, partnerInfo, sourcePincode, destPincode)
	result.ResponseTime = time.Since(startTime)

	// Record successful request
	u.RecordRequest(time.Since(startTime), true)

	return result, nil
}

// Initialize performs any necessary initialization
func (u *UrbanBoltAdapter) Initialize(ctx context.Context) error {
	if err := u.HTTPBaseAdapter.Initialize(ctx); err != nil {
		return err
	}

	// Authenticate with UrbanBolt
	if err := u.GetAuthenticator().Authenticate(ctx); err != nil {
		u.SetHealthStatus("unhealthy")
		return err
	}

	u.SetHealthStatus("healthy")
	return nil
}

// IsHealthy checks if the UrbanBolt adapter is healthy
func (u *UrbanBoltAdapter) IsHealthy(ctx context.Context) bool {
	// Check base health
	if !u.HTTPBaseAdapter.IsHealthy(ctx) {
		return false
	}

	// Check if configuration is valid (don't require authentication for health check)
	// Authentication will happen when needed or during Initialize
	if !u.config.Enabled {
		return false
	}
	if u.config.BaseURL == "" || u.config.Username == "" || u.config.Password == "" {
		return false
	}

	return true
}

// transformResponse converts UrbanBolt response to standard format
func (u *UrbanBoltAdapter) transformResponse(resp *ServiceabilityResponse, partnerInfo common.PartnerInfo, sourcePincode, destPincode string) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// Check if response indicates success
	if resp.Status != "Success" {
		errorMsg := resp.Message
		if errorMsg == "" {
			errorMsg = fmt.Sprintf("UrbanBolt API returned status: %s", resp.Status)
		}
		u.logger.WithFields(logrus.Fields{
			"component": "urbanbolt_adapter",
			"status":    resp.Status,
			"message":   resp.Message,
		}).Info("Non-serviceable response")
		result.ErrorMessage = &errorMsg
		return result
	}

	// Check if we have serviceability data for the requested pincodes
	sourcePincodeInt, _ := strconv.Atoi(sourcePincode)
	destPincodeInt, _ := strconv.Atoi(destPincode)

	var sourceData, destData *PincodeServiceability
	for _, data := range resp.Data {
		if data.Pincode == sourcePincodeInt {
			sourceData = &data
		}
		if data.Pincode == destPincodeInt {
			destData = &data
		}
	}

	// If either pincode is not serviceable, return non-serviceable
	if sourceData == nil || destData == nil {
		u.logger.WithFields(logrus.Fields{
			"component":      "urbanbolt_adapter",
			"source_found":    sourceData != nil,
			"dest_found":      destData != nil,
			"error_pincodes":  resp.ErrorPincodes,
		}).Info("Pincode not found in serviceable data")
		
		errorMsg := "One or more pincodes are not serviceable"
		if len(resp.ErrorPincodes) > 0 {
			errorMsg = fmt.Sprintf("Invalid pincodes: %s", strings.Join(resp.ErrorPincodes, ", "))
		}
		result.ErrorMessage = &errorMsg
		return result
	}

	// Both pincodes are serviceable - build capabilities
	isServiceable := (sourceData.IsActive && destData.IsActive) && 
		((sourceData.Outbound && destData.Inbound) || (sourceData.Inbound && destData.Outbound))

	if !isServiceable {
		errorMsg := "Route is not serviceable"
		result.ErrorMessage = &errorMsg
		return result
	}

	// Parse service types from comma-separated string
	serviceTypes := strings.Split(sourceData.ServiceType, ",")
	for i, st := range serviceTypes {
		serviceTypes[i] = strings.TrimSpace(st)
	}

	// Build services from available service types
	for _, serviceType := range serviceTypes {
		if serviceType == "" {
			continue
		}

		// Map UrbanBolt service types to standard service codes
		serviceCode := strings.ToLower(strings.ReplaceAll(serviceType, ",", "_"))
		serviceName := serviceType

		// Determine delivery mode based on service type
		deliveryMode := "standard"
		if strings.Contains(serviceType, "2HR") || strings.Contains(serviceType, "PTP") {
			deliveryMode = "express"
		}

		result.Services = append(result.Services, models.ServiceV2{
			ServiceCode:   serviceCode,
			ServiceName:   serviceName,
			Pickup:        sourceData.Outbound,
			Delivery:      destData.Inbound,
			ProductTypes:  map[string]bool{"general": true},
			DeliveryModes: map[string]bool{deliveryMode: true},
		})
	}

	// Set capabilities
	result.Capabilities["inbound"] = destData.Inbound
	result.Capabilities["outbound"] = sourceData.Outbound
	result.Capabilities["rtn"] = sourceData.RTN || destData.RTN
	result.Capabilities["pickup_available"] = sourceData.Outbound
	result.Capabilities["delivery_available"] = destData.Inbound
	result.Capabilities["is_serviceable"] = true
	result.Capabilities["service_center"] = sourceData.ServiceCenter
	result.Capabilities["city"] = sourceData.City
	result.Capabilities["state"] = sourceData.State
	result.Capabilities["region"] = sourceData.Region
	result.Capabilities["zone"] = sourceData.Zone
	result.Capabilities["route_code"] = sourceData.RouteCode
	result.Capabilities["available_service_types"] = serviceTypes

	// Set metadata
	result.Metadata["api_version"] = "v1"
	result.Metadata["adapter_type"] = "http_jwt_auth"
	result.Metadata["response_status"] = resp.Status
	result.Metadata["source_pincode_data"] = sourceData
	result.Metadata["dest_pincode_data"] = destData

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

