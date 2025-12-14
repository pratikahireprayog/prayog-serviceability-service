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
	ID   string `json:"id"` // This might be nested or top level, user response example implies partner_id at top level.
	Code string `json:"code"`
	Name string `json:"name"`
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

	c.logger.WithField("count", len(partners)).Debug("Successfully fetched user partners")
	return partners, nil
}
