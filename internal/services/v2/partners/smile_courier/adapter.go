package smile_courier

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// SmileCourierAdapter implements the common.PartnerAdapter interface for Smile Courier
type SmileCourierAdapter struct {
	*common.HTTPBaseAdapter
	client *Client
	config config.SmileCourierConfig
}

// NewSmileCourierAdapter creates a new Smile Courier adapter
func NewSmileCourierAdapter(cfg config.SmileCourierConfig) common.PartnerAdapter {
	// Create partner config
	partnerConfig := common.GetPartnerConfigDefaults("smile_courier", "Smile Courier", common.AdapterTypeHTTP)
	partnerConfig.Timeout = cfg.Timeout
	partnerConfig.Enabled = cfg.Enabled
	// Rating comes from database, not config

	// Update endpoints
	partnerConfig.Endpoints = map[string]string{
		"base_url":          cfg.BaseURL,
		"check_service_url": cfg.CheckServiceURL,
	}

	// Update auth config (no authentication needed)
	partnerConfig.Auth = common.AuthConfig{
		Type: common.AuthTypeNone,
	}

	// Create base adapter
	baseAdapter := common.NewHTTPBaseAdapter(partnerConfig)

	// Create Smile Courier specific components
	auth := NewNoAuthenticator()
	client := NewClient(cfg)

	// Set HTTP client and authenticator
	baseAdapter.SetHTTPClient(client)
	baseAdapter.SetAuthenticator(auth)

	adapter := &SmileCourierAdapter{
		HTTPBaseAdapter: baseAdapter,
		client:          client,
		config:          cfg,
	}

	return adapter
}

// CheckServiceability implements the main serviceability check for Smile Courier
func (s *SmileCourierAdapter) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	// Validate request
	if err := common.ValidateServiceabilityRequest(req); err != nil {
		s.RecordRequest(time.Since(startTime), false)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        err,
		}, nil
	}

	// Transform request to Smile Courier format
	smileCourierReq, err := s.transformRequest(req)
	if err != nil {
		s.RecordRequest(time.Since(startTime), false)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        err,
		}, nil
	}

	// Make API call to Smile Courier
	smileCourierResp, err := s.client.CheckServiceability(ctx, smileCourierReq)
	if err != nil {
		s.RecordRequest(time.Since(startTime), false)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			Error:        err,
		}, nil
	}

	// Transform response to standard format
	result := s.transformResponse(smileCourierResp, partnerInfo)
	result.ResponseTime = time.Since(startTime)

	// Record successful request
	s.RecordRequest(time.Since(startTime), true)

	return result, nil
}

// Initialize performs any necessary initialization
func (s *SmileCourierAdapter) Initialize(ctx context.Context) error {
	if err := s.HTTPBaseAdapter.Initialize(ctx); err != nil {
		return err
	}

	// No authentication needed for Smile Courier
	s.SetHealthStatus("healthy")
	return nil
}

// IsHealthy checks if the Smile Courier adapter is healthy
func (s *SmileCourierAdapter) IsHealthy(ctx context.Context) bool {
	// Check base health
	if !s.HTTPBaseAdapter.IsHealthy(ctx) {
		return false
	}

	// Since no authentication is required, we can do a simple connectivity check
	// This could be enhanced to actually ping the Smile Courier API
	return true
}

// transformRequest converts standard request to Smile Courier format
func (s *SmileCourierAdapter) transformRequest(req *models.ServiceabilityV2Request) (*ServiceabilityRequest, error) {
	// Convert postal codes to integers
	fromPincode, err := strconv.Atoi(getSourcePincode(req))
	if err != nil {
		return nil, fmt.Errorf("invalid source pincode: %s", getSourcePincode(req))
	}

	toPincode, err := strconv.Atoi(getDestinationPincode(req))
	if err != nil {
		return nil, fmt.Errorf("invalid destination pincode: %s", getDestinationPincode(req))
	}

	smileCourierReq := &ServiceabilityRequest{
		FromPincode: fromPincode,
		ToPincode:   toPincode,
	}

	return smileCourierReq, nil
}

// transformResponse converts Smile Courier response to standard format
func (s *SmileCourierAdapter) transformResponse(resp *ServiceabilityResponse, partnerInfo common.PartnerInfo) *common.PartnerServiceabilityResult {
	// TODO: Update response structure when Smile Courier API is integrated
	// Current implementation is placeholder - may need to match final payload format

	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// Handle response based on serviceability status
	if resp.Data != nil && resp.Data.Serviceable {
		// Add services based on available services
		for _, service := range resp.Data.AvailableServices {
			if service.Serviceable {
				v2Service := models.ServiceV2{
					ServiceCode: service.ServiceName,
					ServiceName: service.ServiceName,
					TATDays:     3, // Default TAT for courier services
					IsCOD:       resp.Data.PincodeData.IsCOD,
					Pickup:      true,
					Delivery:    true,
					Insurance:   false, // Default to false
					ProductTypes: map[string]bool{
						"general":   true,
						"documents": true,
					},
					DeliveryModes: map[string]bool{
						"standard": true,
						"express":  service.ServiceName == "vayuquick" || service.ServiceName == "vayuquick_pro",
					},
				}

				result.Services = append(result.Services, v2Service)
			}
		}

		// Add capabilities based on pincode data
		result.Capabilities["cod_available"] = resp.Data.PincodeData.IsCOD
		result.Capabilities["pickup_available"] = true
		result.Capabilities["delivery_available"] = true
		result.Capabilities["zone"] = resp.Data.PincodeData.Zone
		result.Capabilities["state"] = resp.Data.PincodeData.StateName
		result.Capabilities["city"] = resp.Data.PincodeData.City
		result.Capabilities["serviceability_type"] = resp.Data.PincodeData.Serviceability.Serviceability
		result.Capabilities["pincode_type"] = resp.Data.PincodeData.PincodeType.PincodeType
	}

	// Add metadata
	result.Metadata["api_version"] = "v1"
	result.Metadata["adapter_type"] = "http_no_auth"

	// Handle errors
	if resp.Status != 200 {
		errorMsg := resp.Message
		if errorMsg == "" {
			errorMsg = fmt.Sprintf("Smile Courier API returned status %d", resp.Status)
		}
		result.ErrorMessage = &errorMsg
	}

	return result
}

// Helper functions to extract data from standard request
func getSourcePincode(req *models.ServiceabilityV2Request) string {
	if req.SourcePostalCode != nil {
		return *req.SourcePostalCode
	}
	if req.PostalCode != nil {
		// For single postal code requests, use default source or the same code
		return "110001" // Default to Delhi if no source specified
	}
	return ""
}

func getDestinationPincode(req *models.ServiceabilityV2Request) string {
	if req.DestinationPostalCode != nil {
		return *req.DestinationPostalCode
	}
	if req.PostalCode != nil {
		return *req.PostalCode
	}
	return ""
}

func getCountryCode(req *models.ServiceabilityV2Request) string {
	if req.CountryCode != nil {
		return *req.CountryCode
	}
	return "IN" // Default to India
}
