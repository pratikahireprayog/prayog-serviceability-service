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
func (c *FedExClient) authenticate(ctx context.Context) error {
    if c.token != "" && time.Now().Before(c.tokenExpiry) {
        return nil
    }

    authURL := c.config.BaseURL + "/oauth/token"
    
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

// CheckServiceability checks serviceability using transit times API
func (c *FedExClient) CheckServiceability(ctx context.Context, request ServiceabilityRequest) (*ServiceabilityResponse, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Build transit times request
	transitReq := c.buildTransitTimeRequest(request)

	c.logger.WithFields(logrus.Fields{
		"partner":           "FedEx",
		"origin":            request.OriginAddress.PostalCode,
		"destination":       request.DestinationAddress.PostalCode,
		"origin_country":    request.OriginAddress.CountryCode,
		"dest_country":      request.DestinationAddress.CountryCode,
		"weight":            request.Weight,
	}).Info("Checking FedEx serviceability via transit times API")

	transitResp, err := c.callTransitTimesAPI(ctx, transitReq)
	if err != nil {
		return nil, fmt.Errorf("serviceability check failed: %w", err)
	}

	return c.convertTransitToServiceability(transitResp), nil
}

// buildTransitTimeRequest builds the transit time request for serviceability check
func (c *FedExClient) buildTransitTimeRequest(request ServiceabilityRequest) TransitTimeRequest {
	return TransitTimeRequest{
		RequestedShipment: TransitTimeShipment{
			Shipper: Shipper{
				Address: Address{
					PostalCode:  request.OriginAddress.PostalCode,
					CountryCode: request.OriginAddress.CountryCode,
				},
			},
			Recipients: []Recipient{
				{
					Address: Address{
						PostalCode:  request.DestinationAddress.PostalCode,
						CountryCode: request.DestinationAddress.CountryCode,
					},
				},
			},
			PackagingType: "YOUR_PACKAGING",
			CustomsClearanceDetail: &CustomsClearanceDetail{
				Commodities: []Commodity{
					{
						Description: "COMMODITIES",
						CustomsValue: CustomsValue{
							Amount:   "100",
							Currency: "USD",
						},
						NumberOfPieces: 1,
					},
				},
			},
			RequestedPackageLineItems: []RequestedPackageLineItem{
				{
					Weight: Weight{
						Units: "LB",
						Value: request.Weight,
					},
				},
			},
		},
		CarrierCodes: []string{"FDXE"},
	}
}

// callTransitTimesAPI calls the FedEx transit times API
func (c *FedExClient) callTransitTimesAPI(ctx context.Context, request TransitTimeRequest) (*TransitTimeResponse, error) {
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transit time request: %w", err)
	}

	url := c.config.BaseURL + "/availability/v1/transittimes"
	c.logger.WithFields(logrus.Fields{
		"partner":   "FedEx",
		"method":    "POST",
		"url":       url,
		"body_size": len(reqBody),
	}).Info("Making FedEx transit times API request")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create transit time request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-locale", "en_US")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transit time request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read transit time response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"partner":     "FedEx",
		"status_code": resp.StatusCode,
		"body_size":   len(body),
	}).Info("Received FedEx transit times API response")

	if resp.StatusCode != http.StatusOK {
		return nil, &FedExAPIError{
			StatusCode: resp.StatusCode,
			RawBody:    string(body),
		}
	}

	var transitResp TransitTimeResponse
	if err := json.Unmarshal(body, &transitResp); err != nil {
		return nil, fmt.Errorf("failed to decode transit time response: %w", err)
	}

	return &transitResp, nil
}

// convertTransitToServiceability converts transit time response to serviceability response
func (c *FedExClient) convertTransitToServiceability(transitResp *TransitTimeResponse) *ServiceabilityResponse {
	serviceability := &ServiceabilityResponse{}

	if transitResp == nil || len(transitResp.Output.TransitTimes) == 0 {
		serviceability.Serviceable = false
		serviceability.Errors = append(serviceability.Errors, Error{
			Code:    "EMPTY_RESPONSE",
			Message: "No transit time data received from FedEx",
		})
		return serviceability
	}

	transitTime := transitResp.Output.TransitTimes[0]
	
	// KEY SERVICEABILITY CHECK: If transitTimeDetails is empty, address is NOT serviceable
	if len(transitTime.TransitTimeDetails) == 0 {
		serviceability.Serviceable = false
		serviceability.Errors = append(serviceability.Errors, Error{
			Code:    "NO_SERVICES_AVAILABLE",
			Message: "No FedEx services available for this route",
		})
	} else {
		// If we have transit time details, address IS serviceable
		serviceability.Serviceable = true
		
		for _, detail := range transitTime.TransitTimeDetails {
			service := Service{
				ServiceType: detail.ServiceType,
				ServiceName: detail.ServiceName,
			}
			serviceability.AvailableServices = append(serviceability.AvailableServices, service)
		}
	}

	// Add alerts for diagnostics
	if len(transitTime.Alerts) > 0 {
		for _, alert := range transitTime.Alerts {
			serviceability.Errors = append(serviceability.Errors, Error{
				Code:    alert.Code,
				Message: alert.Message,
			})
		}
	}

	c.logger.WithFields(logrus.Fields{
		"partner":            "FedEx",
		"serviceable":        serviceability.Serviceable,
		"services_available": len(serviceability.AvailableServices),
		"errors":             len(serviceability.Errors),
	}).Info("FedEx serviceability check completed via transit times API")

	return serviceability
}
