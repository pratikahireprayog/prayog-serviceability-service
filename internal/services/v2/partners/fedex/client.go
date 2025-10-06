package fedex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"prayog-serviceability-service/internal/shared/config"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// FedExClient handles HTTP communication with FedEx API
type FedExClient struct {
	httpClient *http.Client
	config     config.FedExConfig
	logger     *logrus.Logger
	token      string
	tokenExpiry time.Time
}

// FedExAPIError represents a structured FedEx API error
type FedExAPIError struct {
	StatusCode int
	RawBody    string
	Errors     []APIError
}

func (e *FedExAPIError) Error() string {
	if e == nil {
		return ""
	}
	if len(e.Errors) > 0 {
		return fmt.Sprintf("FedEx API error [%d]: %s", e.StatusCode, e.Errors[0].Message)
	}
	return fmt.Sprintf("FedEx API error [%d]: %s", e.StatusCode, e.RawBody)
}

// NewFedExClient creates a new FedEx HTTP client
func NewFedExClient(config config.FedExConfig) *FedExClient {
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
		"partner":        "FedEx",
		"base_url":       config.BaseURL,
		"client_id":      redactCredentialField(config.ClientID),
		"account_number": redactCredentialField(config.AccountNumber),
		"timeout":        config.Timeout,
	}).Info("Creating FedEx client")

	return &FedExClient{
		httpClient: httpClient,
		config:     config,
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

// authenticate gets or refreshes the OAuth token
func (c *FedExClient) authenticate(ctx context.Context) error {
    // Check if token is still valid
    if c.token != "" && time.Now().Before(c.tokenExpiry) {
        return nil
    }

    authURL := c.config.BaseURL + "/oauth/token"
    
    // Manually create form data without url.Values
    formData := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
        c.config.ClientID,
        c.config.ClientSecret,
    )

    req, err := http.NewRequestWithContext(ctx, "POST", authURL, strings.NewReader(formData))
    if err != nil {
        return fmt.Errorf("failed to create auth request: %w", err)
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    c.logger.WithFields(logrus.Fields{
        "partner": "FedEx",
        "method":  "POST",
        "url":     authURL,
    }).Info("Requesting FedEx OAuth token")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("authentication request failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to read auth response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        c.logger.WithFields(logrus.Fields{
            "partner":     "FedEx",
            "status_code": resp.StatusCode,
            "response":    truncateForLog(string(body), 512),
        }).Error("FedEx authentication failed")
        return fmt.Errorf("authentication failed with status %d", resp.StatusCode)
    }

    var tokenResp TokenResponse
    if err := json.Unmarshal(body, &tokenResp); err != nil {
        return fmt.Errorf("failed to decode token response: %w", err)
    }

    c.token = tokenResp.AccessToken
    // Set expiry with 5-minute buffer
    c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

    c.logger.WithFields(logrus.Fields{
        "partner":    "FedEx",
        "token_type": tokenResp.TokenType,
        "expires_in": tokenResp.ExpiresIn,
    }).Info("FedEx authentication successful")

    return nil
}

// GetRates calls FedEx rates API
func (c *FedExClient) GetRates(ctx context.Context, request RateRequest) (*RateResponse, error) {
	// Authenticate first
	if err := c.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.config.BaseURL + "/rate/v1/rates/quotes"
	c.logger.WithFields(logrus.Fields{
		"partner":   "FedEx",
		"method":    "POST",
		"url":       url,
		"body_size": len(reqBody),
	}).Info("Making FedEx rates API request")

	// Log request body (truncated)
	c.logger.WithFields(logrus.Fields{
		"partner":      "FedEx",
		"event":        "request_body",
		"body_preview": truncateForLog(string(reqBody), 2048),
		"body_size":    len(reqBody),
	}).Debug("FedEx rates request body")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-locale", "en_US")

	// Add retry logic
	var resp *http.Response
	var lastErr error
	maxRetries := 3
	retryDelay := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay * time.Duration(attempt))
			c.logger.WithFields(logrus.Fields{
				"partner": "FedEx",
				"attempt": attempt,
			}).Warn("Retrying FedEx API call")
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
		return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response details safely
	c.logger.WithFields(logrus.Fields{
		"partner":     "FedEx",
		"status_code": resp.StatusCode,
		"body_size":   len(body),
	}).Info("Received FedEx API response")

	// Log response body (truncated)
	c.logger.WithFields(logrus.Fields{
		"partner":      "FedEx",
		"event":        "response_body",
		"status_code":  resp.StatusCode,
		"body_preview": truncateForLog(string(body), 4096),
		"body_size":    len(body),
	}).Debug("FedEx rates response body")

	if resp.StatusCode != http.StatusOK {
		var apiErrorOutput Output
		json.Unmarshal(body, &apiErrorOutput) // Try to parse error details
		
		return nil, &FedExAPIError{
			StatusCode: resp.StatusCode,
			RawBody:    string(body),
			Errors:     apiErrorOutput.Errors,
		}
	}

	var ratesResp RateResponse
	if err := json.Unmarshal(body, &ratesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ratesResp, nil
}

// CheckServiceability checks serviceability using rates API
func (c *FedExClient) CheckServiceability(ctx context.Context, request ServiceabilityRequest) (*ServiceabilityResponse, error) {
	// Build rates request for serviceability check
	ratesReq := RateRequest{
		AccountNumber: AccountNumber{
			Value: c.config.AccountNumber,
		},
		RequestedShipment: RequestedShipment{
			Shipper: Shipper{
				Address: request.OriginAddress,
			},
			Recipient: Recipient{
				Address: request.DestinationAddress,
			},
			PickupType:      "USE_SCHEDULED_PICKUP",
			RateRequestType: []string{"LIST"},
			RequestedPackageLineItems: []RequestedPackageLineItem{
				{
					Weight: Weight{
						Units: "KG",
						Value: request.Weight,
					},
					Dimensions: Dimensions{
						Length: 10,
						Width:  10,
						Height: 10,
						Units:  "CM",
					},
				},
			},
			ShipDate:     request.ShipDate,
			TotalWeight:  &Weight{Units: "KG", Value: request.Weight},
			PackageCount: 1,
		},
	}

	c.logger.WithFields(logrus.Fields{
		"partner":           "FedEx",
		"origin":            request.OriginAddress.PostalCode,
		"destination":       request.DestinationAddress.PostalCode,
		"origin_country":    request.OriginAddress.CountryCode,
		"dest_country":      request.DestinationAddress.CountryCode,
		"weight":            request.Weight,
	}).Info("Checking FedEx serviceability")

	ratesResp, err := c.GetRates(ctx, ratesReq)
	if err != nil {
		return nil, fmt.Errorf("serviceability check failed: %w", err)
	}

	return c.ConvertRatesToServiceability(ratesResp), nil
}

// ConvertRatesToServiceability converts rates response to serviceability response
func (c *FedExClient) ConvertRatesToServiceability(ratesResp *RateResponse) *ServiceabilityResponse {
	serviceability := &ServiceabilityResponse{}

	if ratesResp == nil {
		serviceability.Serviceable = false
		serviceability.Errors = append(serviceability.Errors, Error{
			Code:    "EMPTY_RESPONSE",
			Message: "No response received from FedEx",
		})
		return serviceability
	}

	// Check for API errors
	if len(ratesResp.Output.Errors) > 0 {
		for _, apiErr := range ratesResp.Output.Errors {
			serviceability.Errors = append(serviceability.Errors, Error{
				Code:    apiErr.Code,
				Message: apiErr.Message,
			})
		}
	}

	// Check if we have any available services
	if len(ratesResp.Output.RateReplyDetails) > 0 {
		serviceability.Serviceable = true
		
		for _, detail := range ratesResp.Output.RateReplyDetails {
			// Skip restricted services
			if c.isServiceRestricted(detail) {
				serviceability.Restrictions = append(serviceability.Restrictions, 
					fmt.Sprintf("%s: Service restricted", detail.ServiceType))
				continue
			}

			service := Service{
				ServiceType:  detail.ServiceType,
				ServiceName:  detail.ServiceName,
				TransitDays:  c.extractTransitDays(detail),
				DeliveryDate: c.extractDeliveryDate(detail),
			}

			// Extract cost if available
			if len(detail.RatedShipmentDetails) > 0 {
				ratedDetail := detail.RatedShipmentDetails[0]
				if ratedDetail.TotalNetCharge != nil {
					service.Cost = ratedDetail.TotalNetCharge.Amount
					service.Currency = ratedDetail.TotalNetCharge.Currency
				}
			}

			serviceability.AvailableServices = append(serviceability.AvailableServices, service)
		}
	}

	// If no services found but no specific errors, add generic message
	if !serviceability.Serviceable && len(serviceability.Errors) == 0 {
		serviceability.Errors = append(serviceability.Errors, Error{
			Code:    "NO_SERVICES_AVAILABLE",
			Message: "No FedEx services available for this route",
		})
	}

	c.logger.WithFields(logrus.Fields{
		"partner":           "FedEx",
		"serviceable":       serviceability.Serviceable,
		"services_available": len(serviceability.AvailableServices),
		"errors":            len(serviceability.Errors),
	}).Info("FedEx serviceability check completed")

	return serviceability
}

// Helper methods
func (c *FedExClient) isServiceRestricted(detail RateReplyDetail) bool {
	// Check for restriction indicators
	if len(detail.CustomerMessages) > 0 {
		for _, msg := range detail.CustomerMessages {
			restrictionIndicators := []string{
				"not available", "restricted", "unavailable", 
				"not serviceable", "cannot ship",
			}
			for _, indicator := range restrictionIndicators {
				if strings.Contains(strings.ToLower(msg.Message), indicator) {
					return true
				}
			}
		}
	}
	return false
}

func (c *FedExClient) extractTransitDays(detail RateReplyDetail) int {
	if detail.CommitedDetail != nil {
		return detail.CommitedDetail.DayOfWeek
	}
	if detail.OperationalDetail != nil && detail.OperationalDetail.TransitTime != "" {
		// Parse transit time string like "THREE_DAYS" or "FOUR_DAYS"
		transitMap := map[string]int{
			"ONE_DAY": 1, "TWO_DAYS": 2, "THREE_DAYS": 3, "FOUR_DAYS": 4,
			"FIVE_DAYS": 5, "SIX_DAYS": 6, "SEVEN_DAYS": 7,
		}
		if days, exists := transitMap[detail.OperationalDetail.TransitTime]; exists {
			return days
		}
	}
	return 0
}

func (c *FedExClient) extractDeliveryDate(detail RateReplyDetail) string {
	if detail.CommitedDetail != nil && detail.CommitedDetail.Date != "" {
		return detail.CommitedDetail.Date
	}
	if detail.OperationalDetail != nil && detail.OperationalDetail.DeliveryDay != "" {
		return detail.OperationalDetail.DeliveryDay
	}
	return ""
}