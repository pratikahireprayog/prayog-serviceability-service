package dhl

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"prayog-serviceability-service/internal/shared/config"

	"github.com/sirupsen/logrus"
)

// DHLClient handles HTTP communication with DHL API
type DHLClient struct {
	httpClient *http.Client
	config     config.DHLConfig
	auth       DHLAuthenticator
	logger     *logrus.Logger
}

// DHLAuthenticator interface for DHL authentication
type DHLAuthenticator interface {
	GetAuthHeaders() map[string]string
	IsAuthenticated() bool
	Authenticate(ctx context.Context) error
}

// DHLBasicAuth implements basic authentication for DHL API
type DHLBasicAuth struct {
	config    config.DHLConfig
	username  string
	password  string
	basicAuth string
}

// NewDHLClient creates a new DHL HTTP client
func NewDHLClient(config config.DHLConfig) *DHLClient {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	// Log configuration with redacted sensitive fields
	logger.WithFields(logrus.Fields{
		"partner":        "DHL",
		"base_url":       config.BaseURL,
		"username":       redactCredentialField(config.Username),
		"timeout":        config.Timeout,
		"max_retries":    config.MaxRetries,
		"has_basic_auth": config.BasicAuth != "",
		"has_env_auth":   os.Getenv("DHL_BASIC_AUTH") != "",
	}).Info("Creating DHL client")

	auth := &DHLBasicAuth{
		config:    config,
		username:  config.Username,
		password:  config.Password,
		basicAuth: config.BasicAuth,
	}

	return &DHLClient{
		httpClient: httpClient,
		config:     config,
		auth:       auth,
		logger:     logger,
	}
}

// redactCredentialField safely redacts sensitive credential fields for logging
func redactCredentialField(field string) string {
	if field == "" {
		return "[EMPTY]"
	}
	if len(field) <= 3 {
		return "[REDACTED]"
	}
	return field[:3] + "***"
}

// redactHeaderValue redacts sensitive header values for logging
func redactHeaderValue(key, value string) string {
	sensitiveHeaders := []string{"authorization", "basic", "token", "api-key", "x-api-key"}
	keyLower := strings.ToLower(key)

	for _, sensitive := range sensitiveHeaders {
		if strings.Contains(keyLower, sensitive) {
			return "[REDACTED]"
		}
	}
	return value
}

// truncateForLog safely truncates large strings for logging
func truncateForLog(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	if limit < 0 {
		return ""
	}
	return s[:limit] + "...(truncated)"
}

// getEnvWithPrefix tries to get environment variable with DHL prefix
func getEnvWithPrefix(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

// GetAuthHeaders returns basic auth headers
func (a *DHLBasicAuth) GetAuthHeaders() map[string]string {
	// If basic auth token is provided directly, use it
	if a.basicAuth != "" {
		return map[string]string{
			"Authorization": "Basic " + a.basicAuth,
		}
	}

	// Fall back to username/password approach
	if a.username == "" || a.password == "" {
		return make(map[string]string)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(a.username + ":" + a.password))
	return map[string]string{
		"Authorization": "Basic " + auth,
	}
}

// IsAuthenticated checks if credentials are available
func (a *DHLBasicAuth) IsAuthenticated() bool {
	// Check if basic auth token is provided directly
	if a.basicAuth != "" {
		return true
	}

	// Fall back to username/password check
	return a.username != "" && a.password != ""
}

// Authenticate is a no-op for basic auth
func (a *DHLBasicAuth) Authenticate(ctx context.Context) error {
	if !a.IsAuthenticated() {
		if a.basicAuth != "" {
			return fmt.Errorf("DHL basic auth token is invalid or empty")
		}
		return fmt.Errorf("DHL credentials not available - username: '%s', password: '%s'",
			redactCredentialField(a.username), "[REDACTED]")
	}
	return nil
}

// DHLAPIError represents a structured DHL API error with raw body for diagnostics
type DHLAPIError struct {
	StatusCode   int
	Title        string
	Detail       string
	Message      string
	Status       string
	Instance     string
	RawBody      string
}

func (e *DHLAPIError) Error() string {
	if e == nil {
		return ""
	}
	title := e.Title
	if title == "" {
		title = "DHL API error"
	}
	if e.Detail != "" {
		return fmt.Sprintf("%s [%d]: %s", title, e.StatusCode, e.Detail)
	}
	return fmt.Sprintf("%s [%d]", title, e.StatusCode)
}

// CheckRates calls DHL rates API for serviceability and pricing
func (c *DHLClient) CheckRates(ctx context.Context, request RatesRequest) (*RatesResponse, error) {
	if err := c.auth.Authenticate(ctx); err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Log request with appropriate level of detail
	url := fmt.Sprintf("%s/rates?strictValidation=false", c.config.BaseURL)
	c.logger.WithFields(logrus.Fields{
		"partner":   "DHL",
		"method":    "POST",
		"url":       url,
		"url_path":  "/rates?strictValidation=false",
		"body_size": len(reqBody),
	}).Info("Making DHL API request")
	// Log request body (truncated)
	c.logger.WithFields(logrus.Fields{
		"partner":      "DHL",
		"event":        "request_body",
		"body_preview": truncateForLog(string(reqBody), 2048),
		"body_size":    len(reqBody),
	}).Info("DHL rates request body")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers to match working curl command
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Message-Reference", "d0e7832e-5c98-11ea-bc55-0242ac13")   // Static as per working curl
	req.Header.Set("Message-Reference-Date", "Wed, 21 Oct 2015 07:28:00 GMT") // Static as per working curl
	req.Header.Set("Plugin-Name", "")                                         // Empty as per working curl
	req.Header.Set("Plugin-Version", "")                                      // Empty as per working curl
	req.Header.Set("Shipping-System-Platform-Name", "")                       // Empty as per working curl
	req.Header.Set("Shipping-System-Platform-Version", "")                    // Empty as per working curl
	req.Header.Set("Webstore-Platform-Name", "")                              // Empty as per working curl
	req.Header.Set("Webstore-Platform-Version", "")                           // Empty as per working curl
	req.Header.Set("X-Version", "2.12.0")                                     // Use X-Version instead of x-version

	// Add authentication headers
	authHeaders := c.auth.GetAuthHeaders()
	for key, value := range authHeaders {
		req.Header.Set(key, value)
	}

	// Cookie header intentionally not set

	// Log headers safely (with sensitive headers redacted)
	{
		logFields := logrus.Fields{
			"partner": "DHL",
			"headers": make(map[string]string),
		}
		for key, values := range req.Header {
			for _, value := range values {
				logFields["headers"].(map[string]string)[key] = redactHeaderValue(key, value)
			}
		}
		c.logger.WithFields(logFields).Info("Request headers prepared")
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

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response details safely
	c.logger.WithFields(logrus.Fields{
		"partner":     "DHL",
		"status_code": resp.StatusCode,
		"body_size":   len(body),
	}).Info("Received DHL API response")
	// Log response body (truncated)
	c.logger.WithFields(logrus.Fields{
		"partner":      "DHL",
		"event":        "response_body",
		"status_code":  resp.StatusCode,
		"body_preview": truncateForLog(string(body), 4096),
		"body_size":    len(body),
	}).Info("DHL rates response body")

	if resp.StatusCode != http.StatusOK {
		// Try to parse DHL error response for better error messages
		var dhlError struct {
			Instance string `json:"instance"`
			Detail   string `json:"detail"`
			Title    string `json:"title"`
			Message  string `json:"message"`
			Status   string `json:"status"`
		}

		_ = json.Unmarshal(body, &dhlError)
		return nil, &DHLAPIError{
			StatusCode: resp.StatusCode,
			Title:      dhlError.Title,
			Detail:     dhlError.Detail,
			Message:    dhlError.Message,
			Status:     dhlError.Status,
			Instance:   dhlError.Instance,
			RawBody:    string(body),
		}
	}

	var ratesResp RatesResponse
	if err := json.Unmarshal(body, &ratesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ratesResp, nil
}

// Legacy method for backward compatibility
func (c *DHLClient) CheckServiceability(ctx context.Context, request ServiceabilityRequest) (*ServiceabilityResponse, error) {
	// Convert legacy request to new rates request
	ratesRequest := c.convertLegacyRequest(request)

	// Call new rates API
	ratesResponse, err := c.CheckRates(ctx, ratesRequest)
	if err != nil {
		return &ServiceabilityResponse{
			Success: false,
			Error: &DHLError{
				Code:    "API_ERROR",
				Message: err.Error(),
			},
		}, nil
	}

	// Convert response to legacy format
	return c.convertToLegacyResponse(ratesResponse), nil
}

// convertLegacyRequest converts legacy serviceability request to rates request
func (c *DHLClient) convertLegacyRequest(request ServiceabilityRequest) RatesRequest {
	return RatesRequest{
		CustomerDetails: CustomerDetails{
			ShipperDetails: ShipperDetails{
				PostalCode:  request.OriginPostalCode,
				CityName:    "Origin City", // Default or could be looked up
				CountryCode: request.OriginCountryCode,
			},
			ReceiverDetails: ReceiverDetails{
				PostalCode:  request.DestinationPostalCode,
				CityName:    "Destination City", // Default or could be looked up
				CountryCode: request.DestinationCountryCode,
			},
		},
		Accounts: []Account{
			{
				TypeCode: "shipper",
				Number:   c.config.AccountNumber,
			},
		},
		ProductsAndServices: []ProductAndService{
			{
				ProductCode:      request.ProductType,
				LocalProductCode: request.ProductType,
			},
		},
		PayerCountryCode:           request.OriginCountryCode,
		PlannedShippingDateAndTime: c.getPlannedShippingDateTime(request.ShipmentDate),
		UnitOfMeasurement:          "metric",
		IsCustomsDeclarable:        true,
		EstimatedDeliveryDate: EstimatedDeliveryDate{
			IsRequested: true,
			TypeCode:    "QDDC",
		},
		ReturnStandardProductsOnly: true,
		Packages: []Package{
			{
				Weight: request.Weight,
				Dimensions: Dimensions{
					Length: 30, // Default dimensions
					Width:  20,
					Height: 15,
				},
			},
		},
	}
}

// convertToLegacyResponse converts rates response to legacy serviceability response
func (c *DHLClient) convertToLegacyResponse(ratesResponse *RatesResponse) *ServiceabilityResponse {
	if len(ratesResponse.Products) == 0 {
		return &ServiceabilityResponse{
			Success: false,
			Error: &DHLError{
				Code:    "NO_PRODUCTS",
				Message: "No products available for the requested route",
			},
		}
	}

	// Convert products to services
	var services []DHLService
	for _, product := range ratesResponse.Products {
		// Get pricing - use the first available currency
		var pricing *DHLServicePricing
		if len(product.TotalPrice) > 0 {
			pricing = &DHLServicePricing{
				BaseCost:          product.TotalPrice[0].Price,
				Currency:          product.TotalPrice[0].PriceCurrency,
				FuelSurcharge:     0, // Could be extracted from detailed breakdown
				SecuritySurcharge: 0,
				TotalCost:         product.TotalPrice[0].Price,
			}
		}

		services = append(services, DHLService{
			ProductCode:             product.ProductCode,
			ProductName:             product.ProductName,
			ServiceType:             product.NetworkTypeCode,
			EstimatedDelivery:       product.DeliveryCapabilities.EstimatedDeliveryDateAndTime,
			TransitDays:             product.DeliveryCapabilities.TotalTransitDays,
			IsDocumentsSupported:    true,
			IsNonDocumentsSupported: true,
			TrackingSupported:       true,
			SignatureRequired:       false,
			Pricing:                 pricing,
		})
	}

	return &ServiceabilityResponse{
		Success: true,
		Data: &DHLServiceabilityData{
			IsServiceable: len(services) > 0,
			Services:      services,
			ServiceCapabilities: map[string]bool{
				"international": true,
				"tracking":      true,
				"express":       true,
				"pickup":        true,
				"delivery":      true,
			},
		},
	}
}

// getPlannedShippingDateTime gets the planned shipping date and time
func (c *DHLClient) getPlannedShippingDateTime(shipmentDate string) string {
	if shipmentDate != "" {
		return shipmentDate
	}

	// Default to next business day at 1 PM IST
	now := time.Now()
	// Add 1 day to get next business day (simplified)
	nextDay := now.Add(24 * time.Hour)
	return nextDay.Format("2006-01-02T15:04:05GMT+05:30")
}

// GetQuote calls DHL rates API for quote information (alias for CheckRates)
func (c *DHLClient) GetQuote(ctx context.Context, request QuoteRequest) (*QuoteResponse, error) {
	// Convert QuoteRequest to RatesRequest
	ratesRequest := RatesRequest{
		CustomerDetails: CustomerDetails{
			ShipperDetails: ShipperDetails{
				PostalCode:  request.OriginPostalCode,
				CityName:    "Origin City",
				CountryCode: request.OriginCountryCode,
			},
			ReceiverDetails: ReceiverDetails{
				PostalCode:  request.DestinationPostalCode,
				CityName:    "Destination City",
				CountryCode: request.DestinationCountryCode,
			},
		},
		Accounts: []Account{
			{
				TypeCode: "shipper",
				Number:   c.config.AccountNumber,
			},
		},
		ProductsAndServices: []ProductAndService{
			{
				ProductCode:      request.ServiceType,
				LocalProductCode: request.ServiceType,
			},
		},
		PayerCountryCode:           request.OriginCountryCode,
		PlannedShippingDateAndTime: c.getPlannedShippingDateTime(request.ShipmentDate),
		UnitOfMeasurement:          "metric",
		IsCustomsDeclarable:        true,
		EstimatedDeliveryDate: EstimatedDeliveryDate{
			IsRequested: true,
			TypeCode:    "QDDC",
		},
		ReturnStandardProductsOnly: true,
		Packages:                   c.convertPackages(request.Packages),
	}

	ratesResponse, err := c.CheckRates(ctx, ratesRequest)
	if err != nil {
		return nil, err
	}

	// Convert to QuoteResponse
	if len(ratesResponse.Products) == 0 {
		return nil, fmt.Errorf("no products available for quote")
	}

	product := ratesResponse.Products[0]
	var totalCost float64
	var currency string

	if len(product.TotalPrice) > 0 {
		totalCost = product.TotalPrice[0].Price
		currency = product.TotalPrice[0].PriceCurrency
	}

	return &QuoteResponse{
		QuoteID:      fmt.Sprintf("DHL-%d", time.Now().UnixNano()),
		ServiceType:  product.ProductName,
		TotalCost:    totalCost,
		Currency:     currency,
		DeliveryTime: product.DeliveryCapabilities.EstimatedDeliveryDateAndTime,
		ValidUntil:   time.Now().Add(24 * time.Hour),
	}, nil
}

// convertPackages converts QuoteRequest packages to RatesRequest packages
func (c *DHLClient) convertPackages(packages []DHLPackage) []Package {
	var result []Package
	for _, pkg := range packages {
		result = append(result, Package{
			Weight: pkg.Weight,
			Dimensions: Dimensions{
				Length: pkg.Length,
				Width:  pkg.Width,
				Height: pkg.Height,
			},
		})
	}
	return result
}

// Legacy models for backward compatibility
type QuoteRequest struct {
	OriginCountryCode      string       `json:"origin_country_code"`
	OriginPostalCode       string       `json:"origin_postal_code"`
	DestinationCountryCode string       `json:"destination_country_code"`
	DestinationPostalCode  string       `json:"destination_postal_code"`
	Packages               []DHLPackage `json:"packages"`
	ServiceType            string       `json:"service_type"`
	ShipmentDate           string       `json:"shipment_date,omitempty"`
}

type DHLPackage struct {
	Weight        float64 `json:"weight"`
	Length        float64 `json:"length"`
	Width         float64 `json:"width"`
	Height        float64 `json:"height"`
	DeclaredValue float64 `json:"declared_value"`
	Currency      string  `json:"currency"`
}

type QuoteResponse struct {
	QuoteID      string    `json:"quoteId"`
	ServiceType  string    `json:"serviceType"`
	TotalCost    float64   `json:"totalCost"`
	Currency     string    `json:"currency"`
	DeliveryTime string    `json:"deliveryTime"`
	ValidUntil   time.Time `json:"validUntil"`
}

// Legacy authentication structures
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	APIKey   string `json:"apiKey,omitempty"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	TokenType string    `json:"tokenType"`
}
