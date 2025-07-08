package dhl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"prayog-serviceability-service/internal/shared/config"
)

// DHLClient handles HTTP communication with DHL API
type DHLClient struct {
	httpClient *http.Client
	config     config.DHLConfig
	auth       DHLAuthenticator
}

// DHLAuthenticator interface for DHL authentication
type DHLAuthenticator interface {
	GetAuthToken(ctx context.Context) (string, error)
	IsTokenValid() bool
	RefreshToken(ctx context.Context) error
}

// DHLBasicAuth implements basic username/password authentication
type DHLBasicAuth struct {
	config config.DHLConfig
	token  string
	expiry time.Time
	client *http.Client
}

// AuthRequest represents DHL authentication request
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	APIKey   string `json:"apiKey,omitempty"`
}

// AuthResponse represents DHL authentication response
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	TokenType string    `json:"tokenType"`
}

// QuoteRequest represents request for DHL quote
type QuoteRequest struct {
	OriginCountryCode      string       `json:"origin_country_code"`
	OriginPostalCode       string       `json:"origin_postal_code"`
	DestinationCountryCode string       `json:"destination_country_code"`
	DestinationPostalCode  string       `json:"destination_postal_code"`
	Packages               []DHLPackage `json:"packages"`
	ServiceType            string       `json:"service_type"`
	ShipmentDate           string       `json:"shipment_date,omitempty"`
}

// DHLPackage represents a package in DHL quote request
type DHLPackage struct {
	Weight        float64 `json:"weight"`         // in KG
	Length        float64 `json:"length"`         // in CM
	Width         float64 `json:"width"`          // in CM
	Height        float64 `json:"height"`         // in CM
	DeclaredValue float64 `json:"declared_value"` // in currency units
	Currency      string  `json:"currency"`
}

// QuoteResponse represents DHL quote response
type QuoteResponse struct {
	QuoteID      string    `json:"quoteId"`
	ServiceType  string    `json:"serviceType"`
	TotalCost    float64   `json:"totalCost"`
	Currency     string    `json:"currency"`
	DeliveryTime string    `json:"deliveryTime"`
	ValidUntil   time.Time `json:"validUntil"`
}

// NewDHLClient creates a new DHL HTTP client
func NewDHLClient(config config.DHLConfig) *DHLClient {
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	auth := &DHLBasicAuth{
		config: config,
		client: httpClient,
	}

	return &DHLClient{
		httpClient: httpClient,
		config:     config,
		auth:       auth,
	}
}

// GetAuthToken gets authentication token for DHL API
func (a *DHLBasicAuth) GetAuthToken(ctx context.Context) (string, error) {
	if a.IsTokenValid() {
		return a.token, nil
	}

	if err := a.RefreshToken(ctx); err != nil {
		return "", err
	}
	return a.token, nil
}

// IsTokenValid checks if current token is valid
func (a *DHLBasicAuth) IsTokenValid() bool {
	return a.token != "" && time.Now().Before(a.expiry.Add(-5*time.Minute))
}

// RefreshToken refreshes the authentication token
func (a *DHLBasicAuth) RefreshToken(ctx context.Context) error {
	authReq := AuthRequest{
		Username: a.config.Username,
		Password: a.config.Password,
		APIKey:   a.config.APIKey,
	}

	reqBody, err := json.Marshal(authReq)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	url := a.config.BaseURL + a.config.AuthURL
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if a.config.APIKey != "" {
		req.Header.Set("X-API-Key", a.config.APIKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	a.token = authResp.Token
	a.expiry = authResp.ExpiresAt

	return nil
}

// CheckServiceability calls DHL serviceability API
func (c *DHLClient) CheckServiceability(ctx context.Context, request ServiceabilityRequest) (*ServiceabilityResponse, error) {
	token, err := c.auth.GetAuthToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.config.BaseURL + c.config.ServiceURL
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
	}

	// Add retry logic
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.config.RetryDelay * time.Duration(attempt))
		}

		resp, lastErr = c.httpClient.Do(req)
		if lastErr == nil && resp.StatusCode < 500 {
			break // Success or client error (don't retry)
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.config.MaxRetries, lastErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var serviceResp ServiceabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&serviceResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &serviceResp, nil
}

// GetQuote calls DHL quote API for international shipping
func (c *DHLClient) GetQuote(ctx context.Context, request QuoteRequest) (*QuoteResponse, error) {
	token, err := c.auth.GetAuthToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal quote request: %w", err)
	}

	url := c.config.BaseURL + "/v1/quote"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create quote request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("quote request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("quote API failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var quoteResp QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteResp); err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %w", err)
	}

	return &quoteResp, nil
}
