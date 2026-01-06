package fedex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"prayog-serviceability-service/internal/shared/config"

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

	return &FedExClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// authenticate gets or refreshes the OAuth token
// credentials is optional - if provided and contains client_id/client_secret, uses them; otherwise uses default config
func (c *FedExClient) authenticate(ctx context.Context, credentials map[string]string) error {
    if c.token != "" && time.Now().Before(c.tokenExpiry) {
        return nil
    }

    // Use tenant credentials if provided, otherwise fallback to default config
    clientID := c.config.ClientID
    clientSecret := c.config.ClientSecret
    
    if credentials != nil {
        if id, ok := credentials["client_id"]; ok && id != "" {
            clientID = id
            c.logger.WithFields(logrus.Fields{
                "partner": "FedEx",
            }).Debug("Using tenant-specific client_id for authentication")
        }
        if secret, ok := credentials["client_secret"]; ok && secret != "" {
            clientSecret = secret
            c.logger.WithFields(logrus.Fields{
                "partner": "FedEx",
            }).Debug("Using tenant-specific client_secret for authentication")
        }
    }

    authURL := c.config.BaseURL + "/oauth/token"
    
    formData := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
        clientID,
        clientSecret,
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
        }).Error("FedEx authentication failed")
        return fmt.Errorf("authentication failed with status %d", resp.StatusCode)
    }

    var tokenResp TokenResponse
    if err := json.Unmarshal(body, &tokenResp); err != nil {
        return fmt.Errorf("failed to decode token response: %w", err)
    }

    c.token = tokenResp.AccessToken
    c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

    c.logger.WithFields(logrus.Fields{
        "partner":    "FedEx",
        "token_type": tokenResp.TokenType,
        "expires_in": tokenResp.ExpiresIn,
    }).Info("FedEx authentication successful")

    return nil
}

// GetRates calls FedEx rates API
// credentials is optional - if provided, uses tenant-specific credentials for authentication
func (c *FedExClient) GetRates(ctx context.Context, request RateRequest, credentials map[string]string) (*RateResponse, error) {
	if err := c.authenticate(ctx, credentials); err != nil {
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

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-locale", "en_US")

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"partner":     "FedEx",
		"status_code": resp.StatusCode,
		"body_size":   len(body),
	}).Info("Received FedEx API response")

	if resp.StatusCode != http.StatusOK {
		var apiErrorOutput Output
		json.Unmarshal(body, &apiErrorOutput)
		
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
// credentials is optional - if provided, uses tenant-specific credentials for authentication
func (c *FedExClient) CheckServiceability(ctx context.Context, request ServiceabilityRequest, credentials map[string]string) (*ServiceabilityResponse, error) {
	// Build minimal rates request matching the exact FedEx payload structure
	ratesReq := RateRequest{
		AccountNumber: AccountNumber{
			Value: c.config.AccountNumber,
		},
		RequestedShipment: RequestedShipment{
			Shipper: Shipper{
				Address: Address{
					PostalCode:  request.OriginAddress.PostalCode,
					CountryCode: request.OriginAddress.CountryCode,
				},
			},
			Recipient: Recipient{
				Address: Address{
					PostalCode:  request.DestinationAddress.PostalCode,
					CountryCode: request.DestinationAddress.CountryCode,
				},
			},
			PickupType:      "USE_SCHEDULED_PICKUP",
			RateRequestType: []string{"LIST"},	
			RequestedPackageLineItems: []RequestedPackageLineItem{
				{
					Weight: Weight{
						Units: "KG",
						Value: 1,
					},
				},
			},
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

	ratesResp, err := c.GetRates(ctx, ratesReq, credentials)
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
	// if len(ratesResp.Output.Errors) > 0 {
	// 	for _, apiErr := range ratesResp.Output.Errors {
	// 		serviceability.Errors = append(serviceability.Errors, Error{
	// 			Code:    apiErr.Code,
	// 			Message: apiErr.Message,
	// 		})
	// 	}
	// }

	// Check if we have any available services
	if len(ratesResp.Output.RateReplyDetails) > 0 {
		serviceability.Serviceable = true
		
		for _, detail := range ratesResp.Output.RateReplyDetails {
			service := Service{
				ServiceType:  detail.ServiceType,
				ServiceName:  detail.ServiceName,
			}

			// Extract cost if available
			if len(detail.RatedShipmentDetails) > 0 {
				ratedDetail := detail.RatedShipmentDetails[0]
				if ratedDetail.TotalNetCharge != 0 {
					service.Cost = ratedDetail.TotalNetCharge
					service.Currency = ratedDetail.Currency
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
		"partner":            "FedEx",
		"serviceable":        serviceability.Serviceable,
		"services_available": len(serviceability.AvailableServices),
		"errors":             len(serviceability.Errors),
	}).Info("FedEx serviceability check completed")

	return serviceability
}