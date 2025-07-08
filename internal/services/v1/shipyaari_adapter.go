package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// shipyaariAdapter implements PartnerAdapter interface for Shipyaari API
type shipyaariAdapter struct {
	config     config.ShipyaariConfig
	httpClient *http.Client
	tokenMgr   interfaces.TokenManager
}

// NewShipyaariAdapter creates a new Shipyaari adapter
func NewShipyaariAdapter(cfg config.ShipyaariConfig, httpClient *http.Client) interfaces.PartnerAdapter {
	adapter := &shipyaariAdapter{
		config:     cfg,
		httpClient: httpClient,
	}

	// Initialize token manager
	adapter.tokenMgr = &shipyaariTokenManager{
		config:     cfg,
		httpClient: httpClient,
		mutex:      &sync.RWMutex{},
	}

	return adapter
}

// GetPartnerCode returns the partner code for Shipyaari
func (s *shipyaariAdapter) GetPartnerCode() string {
	return "shipyaari"
}

// GetPartnerName returns the partner display name
func (s *shipyaariAdapter) GetPartnerName() string {
	return "Shipyaari"
}

// CheckServiceability checks serviceability through Shipyaari API
func (s *shipyaariAdapter) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request) (*interfaces.PartnerServiceabilityResult, error) {
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

	// Get authentication token
	token, err := s.tokenMgr.GetToken(ctx)
	if err != nil {
		result.Error = err
		errorMsg := fmt.Sprintf("failed to get auth token: %v", err)
		result.ErrorMessage = &errorMsg
		return result, nil // Return result with error, don't fail the entire process
	}

	// Build request payload
	payload := s.buildShipyaariRequest(req)
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

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

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
	var shipyaariResp ShipyaariServiceabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&shipyaariResp); err != nil {
		result.Error = err
		errorMsg := fmt.Sprintf("failed to decode response: %v", err)
		result.ErrorMessage = &errorMsg
		return result, nil
	}

	// Transform response to standard format
	s.transformShipyaariResponse(&shipyaariResp, result)

	return result, nil
}

// IsHealthy checks if the Shipyaari adapter is healthy
func (s *shipyaariAdapter) IsHealthy(ctx context.Context) bool {
	if !s.config.Enabled {
		return false
	}

	// Check if we can get a valid token
	_, err := s.tokenMgr.GetToken(ctx)
	return err == nil
}

// GetAdapterType returns the adapter type
func (s *shipyaariAdapter) GetAdapterType() interfaces.AdapterType {
	return interfaces.AdapterTypeHTTP
}

// buildShipyaariRequest converts V2 request to Shipyaari format
func (s *shipyaariAdapter) buildShipyaariRequest(req *models.ServiceabilityV2Request) ShipyaariServiceabilityRequest {
	shipyaariReq := ShipyaariServiceabilityRequest{
		CountryCode: req.CountryCode,
	}

	if req.PickupPostalCode != nil {
		shipyaariReq.PickupPincode = *req.PickupPostalCode
	}
	if req.DeliveryPostalCode != nil {
		shipyaariReq.DeliveryPincode = *req.DeliveryPostalCode
	}

	// Single postal code check
	if req.PostalCode != nil {
		shipyaariReq.DeliveryPincode = *req.PostalCode
	}

	return shipyaariReq
}

// transformShipyaariResponse converts Shipyaari response to standard format
func (s *shipyaariAdapter) transformShipyaariResponse(resp *ShipyaariServiceabilityResponse, result *interfaces.PartnerServiceabilityResult) {
	if resp.Success && resp.Data != nil {
		result.IsServiceable = resp.Data.IsServiceable

		// Transform services
		for _, service := range resp.Data.Services {
			v2Service := models.ServiceV2{
				ServiceCode:   service.ServiceCode,
				ServiceName:   service.ServiceName,
				TATDays:       service.TATDays,
				IsCOD:         service.CODAvailable,
				Pickup:        service.PickupAvailable,
				Delivery:      service.DeliveryAvailable,
				Insurance:     service.InsuranceAvailable,
				ProductTypes:  map[string]bool{"general": true},
				DeliveryModes: map[string]bool{"standard": true},
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
		result.Capabilities["delivery_modes"] = resp.Data.DeliveryModes
		result.Capabilities["payment_modes"] = resp.Data.PaymentModes
		result.Capabilities["service_types"] = resp.Data.ServiceTypes
	}
}

// executeWithRetry executes HTTP request with retry logic
func (s *shipyaariAdapter) executeWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
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

// Shipyaari API request/response types
type ShipyaariServiceabilityRequest struct {
	PickupPincode   string `json:"pickup_pincode,omitempty"`
	DeliveryPincode string `json:"delivery_pincode"`
	CountryCode     string `json:"country_code"`
}

type ShipyaariServiceabilityResponse struct {
	Success bool                         `json:"success"`
	Data    *ShipyaariServiceabilityData `json:"data,omitempty"`
	Error   *ShipyaariError              `json:"error,omitempty"`
}

type ShipyaariServiceabilityData struct {
	IsServiceable bool               `json:"is_serviceable"`
	Services      []ShipyaariService `json:"services"`
	DeliveryModes []string           `json:"delivery_modes"`
	PaymentModes  []string           `json:"payment_modes"`
	ServiceTypes  []string           `json:"service_types"`
}

type ShipyaariService struct {
	ServiceCode        string                   `json:"service_code"`
	ServiceName        string                   `json:"service_name"`
	TATDays            int                      `json:"tat_days"`
	CODAvailable       bool                     `json:"cod_available"`
	PickupAvailable    bool                     `json:"pickup_available"`
	DeliveryAvailable  bool                     `json:"delivery_available"`
	InsuranceAvailable bool                     `json:"insurance_available"`
	Pricing            *ShipyaariServicePricing `json:"pricing,omitempty"`
}

type ShipyaariServicePricing struct {
	BaseCost      float64 `json:"base_cost"`
	Currency      string  `json:"currency"`
	CODCharges    float64 `json:"cod_charges"`
	FuelSurcharge float64 `json:"fuel_surcharge"`
}

type ShipyaariError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// shipyaariTokenManager handles JWT token management for Shipyaari API
type shipyaariTokenManager struct {
	config       config.ShipyaariConfig
	httpClient   *http.Client
	mutex        *sync.RWMutex
	currentToken string
	tokenExpiry  time.Time
}

// GetToken retrieves a valid token, refreshing if necessary
func (tm *shipyaariTokenManager) GetToken(ctx context.Context) (string, error) {
	tm.mutex.RLock()
	if tm.isCurrentTokenValid() {
		token := tm.currentToken
		tm.mutex.RUnlock()
		return token, nil
	}
	tm.mutex.RUnlock()

	// Token is invalid or expired, refresh it
	return tm.RefreshToken(ctx)
}

// RefreshToken forcefully refreshes the token
func (tm *shipyaariTokenManager) RefreshToken(ctx context.Context) (string, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Check if email and password are configured
	if tm.config.Email == "" || tm.config.Password == "" {
		return "", fmt.Errorf("Shipyaari email and password must be configured")
	}

	// Prepare login request
	loginReq := ShipyaariLoginRequest{
		Email:    tm.config.Email,
		Password: tm.config.Password,
	}

	payloadBytes, err := json.Marshal(loginReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Make token request
	tokenURL := tm.config.BaseURL + tm.config.TokenURL
	httpReq, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := tm.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to execute token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	// Parse token response
	var tokenResp ShipyaariTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if !tokenResp.Success || tokenResp.Data == nil {
		return "", fmt.Errorf("token request failed: %s", tokenResp.Message)
	}

	// Store the new token
	tm.currentToken = tokenResp.Data.Token
	tm.tokenExpiry = time.Now().Add(time.Duration(tokenResp.Data.ExpiresIn)*time.Second - tm.config.TokenExpiryBuffer)

	return tm.currentToken, nil
}

// IsTokenValid checks if the current token is valid
func (tm *shipyaariTokenManager) IsTokenValid() bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.isCurrentTokenValid()
}

// ClearToken clears the stored token
func (tm *shipyaariTokenManager) ClearToken() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.currentToken = ""
	tm.tokenExpiry = time.Time{}
}

// isCurrentTokenValid checks if current token is valid (internal method)
func (tm *shipyaariTokenManager) isCurrentTokenValid() bool {
	return tm.currentToken != "" && time.Now().Before(tm.tokenExpiry)
}

// Shipyaari authentication request/response types
type ShipyaariLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ShipyaariTokenResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    *ShipyaariTokenData `json:"data,omitempty"`
}

type ShipyaariTokenData struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"` // in seconds
}
