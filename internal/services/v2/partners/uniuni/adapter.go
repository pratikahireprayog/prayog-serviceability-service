package uniuni

import (
    "context"
    "fmt"
    "time"

    outbound "prayog-serviceability-service/internal/infrastructure/api/http/outbound"
    "prayog-serviceability-service/internal/infrastructure/resilience"
    "prayog-serviceability-service/internal/services/v2/partners/common"
    "prayog-serviceability-service/internal/shared/config"
    models "prayog-serviceability-service/internal/shared/models/v1"
)

// Adapter implements UniUni partner integration for simple postal code availability
type Adapter struct {
    *common.BaseAdapter
    tokenClient *outbound.HTTPTransport
    apiClient   *outbound.HTTPTransport
    cfg         config.UniUniConfig
}

// NewAdapter creates a new UniUni adapter with provided config
func NewAdapter(uCfg config.UniUniConfig) common.PartnerAdapter {
    cfg := common.GetPartnerConfigDefaults("uniuni", "UniUni", common.AdapterTypeHTTP)

    tokenClient := outbound.NewHTTPTransport(outbound.Config{
        ServiceName:    "uniuni-token",
        BaseURL:        uCfg.TokenBaseURL,
        Timeout:        uCfg.Timeout,
        DefaultHeaders: map[string]string{"Accept": "application/json"},
        EnableLogging:  false,
        CircuitBreaker: resilience.Config{Name: "uniuni-token", MaxRequests: 10, FailureThreshold: 5, Timeout: 60 * time.Second, Interval: 5 * time.Second},
        RetryPolicy:    resilience.RetryPolicy{MaxRetries: uCfg.MaxRetries, InitialDelay: 500 * time.Millisecond, MaxDelay: 2 * time.Second, BackoffMultiplier: 2.0, Jitter: true},
        TimeoutConfig:  resilience.TimeoutConfig{RequestTimeout: uCfg.Timeout, ConnectionTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second},
    })

    apiClient := outbound.NewHTTPTransport(outbound.Config{
        ServiceName:    "uniuni-api",
        BaseURL:        uCfg.APIBaseURL,
        Timeout:        uCfg.Timeout,
        DefaultHeaders: map[string]string{"Accept": "application/json"},
        EnableLogging:  false,
        CircuitBreaker: resilience.Config{Name: "uniuni-api", MaxRequests: 10, FailureThreshold: 5, Timeout: 60 * time.Second, Interval: 5 * time.Second},
        RetryPolicy:    resilience.RetryPolicy{MaxRetries: uCfg.MaxRetries, InitialDelay: 500 * time.Millisecond, MaxDelay: 2 * time.Second, BackoffMultiplier: 2.0, Jitter: true},
        TimeoutConfig:  resilience.TimeoutConfig{RequestTimeout: uCfg.Timeout, ConnectionTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second},
    })

    return &Adapter{
        BaseAdapter: common.NewBaseAdapter(cfg),
        tokenClient: tokenClient,
        apiClient:   apiClient,
        cfg:         uCfg,
    }
}

// uniuniTokenResponse models the token API response
type uniuniTokenResponse struct {
    Status  string `json:"status"`
    RetMsg  string `json:"ret_msg"`
    ErrCode int    `json:"err_code"`
    Data    interface{} `json:"data"`
}

// uniuniAvailabilityRequest models the availability request
type uniuniAvailabilityRequest struct {
    PostalCode string `json:"postal_code"`
}

// CheckServiceability implements the partner check by calling UniUni token then availability
func (a *Adapter) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
    start := time.Now()

    // Determine which postal code to check (prefer destination or single postal_code)
    var postalCode string
    if req.PostalCode != nil && *req.PostalCode != "" {
        postalCode = *req.PostalCode
    } else if req.DestinationPostalCode != nil && *req.DestinationPostalCode != "" {
        postalCode = *req.DestinationPostalCode
    }
    if postalCode == "" {
        a.RecordRequest(time.Since(start), false)
        return &common.PartnerServiceabilityResult{
            PartnerID:    partnerInfo.PartnerID,
            PartnerCode:  partnerInfo.PartnerCode,
            PartnerName:  a.GetPartnerName(),
            Services:     []models.ServiceV2{},
            Capabilities: map[string]interface{}{"is_serviceable": false},
            ErrorMessage: strPtr("postal_code is required for UniUni"),
            ResponseTime: time.Since(start),
        }, nil
    }

    // 1) Fetch access token
    var tokenRes uniuniTokenResponse
    tokenPayload := map[string]string{
        "grant_type":    a.getGrantType(),
        "client_id":     a.getClientID(),
        "client_secret": a.getClientSecret(),
    }
    if err := a.tokenClient.Execute(ctx, outbound.Request{
        Method:  "POST",
        Path:    "/storeauth/customertoken",
        Body:    tokenPayload,
        Headers: map[string]string{"Content-Type": "application/json"},
    }, &tokenRes); err != nil {
        a.RecordRequest(time.Since(start), false)
        return &common.PartnerServiceabilityResult{
            PartnerID:    partnerInfo.PartnerID,
            PartnerCode:  partnerInfo.PartnerCode,
            PartnerName:  a.GetPartnerName(),
            Services:     []models.ServiceV2{},
            Capabilities: map[string]interface{}{"is_serviceable": false},
            ErrorMessage: strPtr(fmt.Sprintf("UniUni token error: %v", err)),
            Metadata:     map[string]interface{}{"token_status": tokenRes.Status, "token_error": tokenRes.RetMsg},
            ResponseTime: time.Since(start),
        }, nil
    }
    accessToken := ""
    switch t := tokenRes.Data.(type) {
    case map[string]interface{}:
        if v, ok := t["access_token"].(string); ok {
            accessToken = v
        }
    }
    if accessToken == "" {
        a.RecordRequest(time.Since(start), false)
        return &common.PartnerServiceabilityResult{
            PartnerID:    partnerInfo.PartnerID,
            PartnerCode:  partnerInfo.PartnerCode,
            PartnerName:  a.GetPartnerName(),
            Services:     []models.ServiceV2{},
            Capabilities: map[string]interface{}{"is_serviceable": false},
            ErrorMessage: strPtr("UniUni token missing access_token"),
            Metadata:     map[string]interface{}{"token_status": tokenRes.Status, "token_error": tokenRes.RetMsg},
            ResponseTime: time.Since(start),
        }, nil
    }

    // 2) Call availability with bearer token
    reqBody := uniuniAvailabilityRequest{PostalCode: postalCode}
    rawResp := make(map[string]interface{})
    if err := a.apiClient.Execute(ctx, outbound.Request{
        Method: "POST",
        Path:   "/orders/checkserviceavailability",
        Body:   reqBody,
        Headers: map[string]string{
            "Authorization": "Bearer " + accessToken,
            "Content-Type":  "application/json",
        },
    }, &rawResp); err != nil {
        a.RecordRequest(time.Since(start), false)
        return &common.PartnerServiceabilityResult{
            PartnerID:    partnerInfo.PartnerID,
            PartnerCode:  partnerInfo.PartnerCode,
            PartnerName:  a.GetPartnerName(),
            Services:     []models.ServiceV2{},
            Capabilities: map[string]interface{}{"is_serviceable": false},
            ErrorMessage: strPtr(fmt.Sprintf("UniUni availability error: %v", err)),
            ResponseTime: time.Since(start),
        }, nil
    }

    // We don't transform UniUni response; attach as partner_services for transparency
    result := &common.PartnerServiceabilityResult{
        PartnerID:       partnerInfo.PartnerID,
        PartnerCode:     partnerInfo.PartnerCode,
        PartnerName:     a.GetPartnerName(),
        Services:        []models.ServiceV2{},
        Capabilities:    map[string]interface{}{},
        PartnerServices: []interface{}{rawResp},
        Metadata:        map[string]interface{}{"provider": "uniuni"},
        ResponseTime:    time.Since(start),
    }

    a.RecordRequest(time.Since(start), true)
    return result, nil
}

func (a *Adapter) Initialize(ctx context.Context) error {
    return a.BaseAdapter.Initialize(ctx)
}

func (a *Adapter) Shutdown(ctx context.Context) error {
    return a.BaseAdapter.Shutdown(ctx)
}

func strPtr(s string) *string { return &s }

func (a *Adapter) getGrantType() string { if a.cfg.GrantType != "" { return a.cfg.GrantType }; return "client_credentials" }
func (a *Adapter) getClientID() string { return a.cfg.ClientID }
func (a *Adapter) getClientSecret() string { return a.cfg.ClientSecret }


