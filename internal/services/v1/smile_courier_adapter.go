package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// smileCourierAdapter implements PartnerAdapter interface for Smile Courier API
type smileCourierAdapter struct {
	config     config.SmileCourierConfig
	httpClient *http.Client
}

// NewSmileCourierAdapter creates a new Smile Courier adapter
func NewSmileCourierAdapter(cfg config.SmileCourierConfig, httpClient *http.Client) interfaces.PartnerAdapter {
	return &smileCourierAdapter{
		config:     cfg,
		httpClient: httpClient,
	}
}

// GetPartnerCode returns the partner code for Smile Courier
func (s *smileCourierAdapter) GetPartnerCode() string {
	return "smile_courier"
}

// GetPartnerName returns the partner display name
func (s *smileCourierAdapter) GetPartnerName() string {
	return "Smile Courier"
}

// CheckServiceability checks serviceability through Smile Courier API
func (s *smileCourierAdapter) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request) (*interfaces.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	result := &interfaces.PartnerServiceabilityResult{
		PartnerCode:   s.GetPartnerCode(),
		PartnerName:   s.GetPartnerName(),
		IsServiceable: false,
		Services:      []models.ServiceV2{},
		Capabilities:  make(map[string]interface{}),
		ResponseTime:  0,
		Rating:        s.config.Rating,
	}

	defer func() {
		result.ResponseTime = time.Since(startTime)
	}()

	// Build request payload
	payload := s.buildSmileCourierRequest(req)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		result.Error = err
		errorMsg := fmt.Sprintf("failed to marshal request: %v", err)
		result.ErrorMessage = &errorMsg
		return result, nil
	}

	// Make API request
	apiURL := s.config.BaseURL + s.config.CheckServiceURL
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		result.Error = err
		errorMsg := fmt.Sprintf("failed to create request: %v", err)
		result.ErrorMessage = &errorMsg
		return result, nil
	}

	// Set headers (no authentication required for Smile Courier)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Execute request with retry logic
	resp, err := s.executeWithRetry(ctx, httpReq)
	if err != nil {
		result.Error = err
		errorMsg := fmt.Sprintf("failed to execute request: %v", err)
		result.ErrorMessage = &errorMsg
		return result, nil
	}
	defer resp.Body.Close()

	// Parse response
	var smileResp SmileCourierServiceabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&smileResp); err != nil {
		result.Error = err
		errorMsg := fmt.Sprintf("failed to decode response: %v", err)
		result.ErrorMessage = &errorMsg
		return result, nil
	}

	// Transform response to standard format
	s.transformSmileCourierResponse(&smileResp, result)

	return result, nil
}

// IsHealthy checks if the Smile Courier adapter is healthy
func (s *smileCourierAdapter) IsHealthy(ctx context.Context) bool {
	if !s.config.Enabled {
		return false
	}

	// Simple health check - try to make a request to the API
	healthReq := SmileCourierServiceabilityRequest{
		FromPincode: "110001", // Delhi
		ToPincode:   "400001", // Mumbai
		CountryCode: "IN",
	}

	payloadBytes, err := json.Marshal(healthReq)
	if err != nil {
		return false
	}

	healthCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	apiURL := s.config.BaseURL + s.config.CheckServiceURL
	httpReq, err := http.NewRequestWithContext(healthCtx, "POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Consider it healthy if we get any response (2xx, 4xx, 5xx)
	return resp.StatusCode < 600
}

// GetAdapterType returns the adapter type
func (s *smileCourierAdapter) GetAdapterType() interfaces.AdapterType {
	return interfaces.AdapterTypeHTTP
}

// buildSmileCourierRequest converts V2 request to Smile Courier format
func (s *smileCourierAdapter) buildSmileCourierRequest(req *models.ServiceabilityV2Request) SmileCourierServiceabilityRequest {
	smileReq := SmileCourierServiceabilityRequest{
		CountryCode: req.CountryCode,
	}

	if req.PickupPostalCode != nil {
		smileReq.FromPincode = *req.PickupPostalCode
	}
	if req.DeliveryPostalCode != nil {
		smileReq.ToPincode = *req.DeliveryPostalCode
	}

	// Single postal code check - use as destination
	if req.PostalCode != nil {
		smileReq.ToPincode = *req.PostalCode
		// Use a default source if not specified
		if smileReq.FromPincode == "" {
			smileReq.FromPincode = "110001" // Default to Delhi
		}
	}

	return smileReq
}

// transformSmileCourierResponse converts Smile Courier response to standard format
func (s *smileCourierAdapter) transformSmileCourierResponse(resp *SmileCourierServiceabilityResponse, result *interfaces.PartnerServiceabilityResult) {
	if resp.Success && resp.Data != nil {
		result.IsServiceable = resp.Data.IsServiceable

		// Transform services
		for _, service := range resp.Data.Services {
			v2Service := models.ServiceV2{
				ServiceCode:   service.ServiceCode,
				ServiceName:   service.ServiceName,
				TATDays:       service.DeliveryDays,
				IsCOD:         service.CODSupported,
				Pickup:        service.PickupAvailable,
				Delivery:      service.DeliveryAvailable,
				Insurance:     service.InsuranceAvailable,
				ProductTypes:  map[string]bool{"general": true, "documents": service.DocumentsSupported},
				DeliveryModes: map[string]bool{"standard": true, "express": service.ExpressAvailable},
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
		result.Capabilities["delivery_modes"] = resp.Data.AvailableModes
		result.Capabilities["payment_modes"] = resp.Data.PaymentModes
		result.Capabilities["max_weight"] = resp.Data.MaxWeight
		result.Capabilities["zones"] = resp.Data.ServiceZones
	}
}

// executeWithRetry executes HTTP request with retry logic
func (s *smileCourierAdapter) executeWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(s.config.RetryDelay * time.Duration(attempt)):
			}
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Check if response indicates we should retry
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: server error", resp.StatusCode)
			continue
		}

		// Success or client error (don't retry)
		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %v", s.config.MaxRetries, lastErr)
}

// Smile Courier API request/response types
type SmileCourierServiceabilityRequest struct {
	FromPincode string `json:"from_pincode,omitempty"`
	ToPincode   string `json:"to_pincode"`
	CountryCode string `json:"country_code"`
}

type SmileCourierServiceabilityResponse struct {
	Success bool                            `json:"success"`
	Data    *SmileCourierServiceabilityData `json:"data,omitempty"`
	Error   *SmileCourierError              `json:"error,omitempty"`
}

type SmileCourierServiceabilityData struct {
	IsServiceable  bool                  `json:"is_serviceable"`
	Services       []SmileCourierService `json:"services"`
	AvailableModes []string              `json:"available_modes"`
	PaymentModes   []string              `json:"payment_modes"`
	MaxWeight      float64               `json:"max_weight"`
	ServiceZones   []string              `json:"service_zones"`
}

type SmileCourierService struct {
	ServiceCode        string                      `json:"service_code"`
	ServiceName        string                      `json:"service_name"`
	DeliveryDays       int                         `json:"delivery_days"`
	CODSupported       bool                        `json:"cod_supported"`
	PickupAvailable    bool                        `json:"pickup_available"`
	DeliveryAvailable  bool                        `json:"delivery_available"`
	InsuranceAvailable bool                        `json:"insurance_available"`
	DocumentsSupported bool                        `json:"documents_supported"`
	ExpressAvailable   bool                        `json:"express_available"`
	Pricing            *SmileCourierServicePricing `json:"pricing,omitempty"`
}

type SmileCourierServicePricing struct {
	BaseCost      float64 `json:"base_cost"`
	Currency      string  `json:"currency"`
	CODCharges    float64 `json:"cod_charges"`
	FuelSurcharge float64 `json:"fuel_surcharge"`
}

type SmileCourierError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
