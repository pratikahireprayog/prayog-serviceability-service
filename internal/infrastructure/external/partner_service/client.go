package partner_service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// PartnerServiceClient handles interactions with the partner management service
type PartnerServiceClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

// Response models for the partner service API
type UserPartnerResponse struct {
	Data    []UserPartnerData `json:"data"`
	Success bool              `json:"success"`
}

type UserPartnerData struct {
	ID      string      `json:"id"`
	Partner PartnerInfo `json:"partner"`
}

type PartnerInfo struct {
	ID          string       `json:"id"` // This might be nested or top level, user response example implies partner_id at top level.
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	Capabilities []Capability `json:"capabilities,omitempty"`
}

type Capability struct {
	ID          string                 `json:"id"`
	CapabilityID string                `json:"capability_id"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	IsSupported bool                   `json:"is_supported"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewPartnerServiceClient creates a new client instance
func NewPartnerServiceClient(baseURL string, logger *logrus.Logger) *PartnerServiceClient {
	return &PartnerServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: logger,
	}
}

// GetUserPartners fetches the list of partners assigned to a user/tenant
func (c *PartnerServiceClient) GetUserPartners(ctx context.Context, tenantID, userID string) ([]PartnerInfo, error) {
	url := fmt.Sprintf("%s/partner/v1/user-partners", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("x-tenant-id", tenantID)
	req.Header.Set("x-user-id", userID)
	req.Header.Set("Content-Type", "application/json")

	c.logger.WithFields(logrus.Fields{
		"url":       url,
		"tenant_id": tenantID,
		"user_id":   userID,
	}).Debug("Fetching user partners")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("partner service returned non-200 status: %d", resp.StatusCode)
	}

	var parsedResp UserPartnerResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !parsedResp.Success {
		return nil, fmt.Errorf("partner service returned unsuccessful response")
	}

	partners := make([]PartnerInfo, len(parsedResp.Data))
	for i, data := range parsedResp.Data {
		partners[i] = data.Partner

		// If PartnerInfo ID is empty (maybe not in partner object JSON), fallback to UserPartnerData ID?
		// Usually partner_id is consistent.
		if partners[i].ID == "" {
			partners[i].ID = data.ID
		}
	}

	// Log capabilities for debugging
	for _, p := range partners {
		if len(p.Capabilities) > 0 {
			c.logger.WithFields(logrus.Fields{
				"partner_code": p.Code,
				"capabilities_count": len(p.Capabilities),
			}).Debug("Partner has capabilities from partner service")
		}
	}

	c.logger.WithField("count", len(partners)).Debug("Successfully fetched user partners")
	return partners, nil
}

// TenantPartnerCredentialsResponse represents the response from the tenant credentials API
type TenantPartnerCredentialsResponse struct {
	Success bool                        `json:"success"`
	Data    TenantPartnerCredentialsData `json:"data"`
}

type TenantPartnerCredentialsData struct {
	TenantID    string            `json:"tenant_id"`
	PartnerCode string            `json:"partner_code"`
	Credentials map[string]string `json:"credentials"`
	LogoURL     string            `json:"logo_url,omitempty"`
}

// GetTenantPartnerCredentials fetches tenant-specific credentials for a partner
// Returns the credentials map and optional logo URL if found, or nil/empty string if not found
func (c *PartnerServiceClient) GetTenantPartnerCredentials(ctx context.Context, tenantID, partnerCode string) (map[string]string, string, error) {
	url := fmt.Sprintf("%s/partner/v1/tenant-credentials?tenant_id=%s&partner_code=%s", c.baseURL, tenantID, partnerCode)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("x-tenant-id", tenantID)
	req.Header.Set("Content-Type", "application/json")

	c.logger.WithFields(logrus.Fields{
		"url":          url,
		"tenant_id":    tenantID,
		"partner_code": partnerCode,
	}).Debug("Fetching tenant partner credentials")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Log warning but don't fail - fallback to default credentials
		c.logger.WithFields(logrus.Fields{
			"tenant_id":    tenantID,
			"partner_code": partnerCode,
			"error":        err.Error(),
		}).Debug("Failed to fetch tenant partner credentials, will use default credentials")
		return nil, "", nil // Return nil, not error - expected behavior for missing credentials
	}
	defer resp.Body.Close()

	// If not found (404), return nil - tenant doesn't have negotiated rates with this partner
	if resp.StatusCode == http.StatusNotFound {
		c.logger.WithFields(logrus.Fields{
			"tenant_id":    tenantID,
			"partner_code": partnerCode,
		}).Debug("Tenant partner credentials not found, will use default credentials")
		return nil, "", nil
	}

	// If other error status, log warning and return nil (fallback to defaults)
	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"tenant_id":    tenantID,
			"partner_code": partnerCode,
			"status_code":  resp.StatusCode,
		}).Debug("Partner service returned non-200 status for tenant credentials, will use default credentials")
		return nil, "", nil
	}

	var parsedResp TenantPartnerCredentialsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsedResp); err != nil {
		c.logger.WithFields(logrus.Fields{
			"tenant_id":    tenantID,
			"partner_code": partnerCode,
			"error":        err.Error(),
		}).Debug("Failed to decode tenant credentials response, will use default credentials")
		return nil, "", nil
	}

	if !parsedResp.Success {
		c.logger.WithFields(logrus.Fields{
			"tenant_id":    tenantID,
			"partner_code": partnerCode,
		}).Debug("Partner service returned unsuccessful response for tenant credentials, will use default credentials")
		return nil, "", nil
	}

	// Return credentials if available
	if len(parsedResp.Data.Credentials) > 0 {
		c.logger.WithFields(logrus.Fields{
			"tenant_id":        tenantID,
			"partner_code":     partnerCode,
			"credentials_count": len(parsedResp.Data.Credentials),
			"has_logo_url":     parsedResp.Data.LogoURL != "",
		}).Info("Successfully fetched tenant partner credentials")
		return parsedResp.Data.Credentials, parsedResp.Data.LogoURL, nil
	}

	// No credentials in response, but maybe logo?
	if parsedResp.Data.LogoURL != "" {
		return nil, parsedResp.Data.LogoURL, nil
	}

	return nil, "", nil
}
