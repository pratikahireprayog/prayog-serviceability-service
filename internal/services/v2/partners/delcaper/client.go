package delcaper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"prayog-serviceability-service/internal/shared/config"
)

// DelcaperClient handles HTTP communication with Delcaper API
type DelcaperClient struct {
	config      config.DelcaperConfig
	httpClient  *http.Client
	tokenManager *TokenManager
}

// NewDelcaperClient creates a new Delcaper client
func NewDelcaperClient(cfg config.DelcaperConfig, httpClient *http.Client) *DelcaperClient {
	return &DelcaperClient{
		config:      cfg,
		httpClient:  httpClient,
		tokenManager: NewTokenManager(cfg, httpClient),
	}
}

// Login authenticates with Delcaper API and returns access token
func (c *DelcaperClient) Login(ctx context.Context) (*LoginResponse, error) {
	loginReq := LoginRequest{
		Email:      c.config.Email,
		Password:   c.config.Password,
		VendorType: c.config.VendorType,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	url := c.config.BaseURL + c.config.LoginURL
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute login request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read login response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	// Store tokens in token manager
	c.tokenManager.SetTokens(loginResp.Data.AccessToken, loginResp.Data.RefreshToken, loginResp.Data.ExpiresIn)

	return &loginResp, nil
}

// CheckFeasible checks serviceability for the given request
func (c *DelcaperClient) CheckFeasible(ctx context.Context, req *CheckFeasibleRequest) (*CheckFeasibleResponse, error) {
	// Get valid token
	token, err := c.tokenManager.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal feasibility request: %w", err)
	}

	url := c.config.BaseURL + c.config.CheckServiceURL
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create feasibility request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute feasibility request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read feasibility response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feasibility check failed with status %d: %s", resp.StatusCode, string(body))
	}

	var feasibleResp CheckFeasibleResponse
	if err := json.Unmarshal(body, &feasibleResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feasibility response: %w", err)
	}

	return &feasibleResp, nil
}

// IsHealthy checks if the client can successfully authenticate
func (c *DelcaperClient) IsHealthy(ctx context.Context) bool {
	// For now, just check if the configuration is valid
	// The actual authentication will happen when making API calls
	return c.config.Enabled && c.config.BaseURL != "" && c.config.Email != "" && c.config.Password != ""
} 