package orchestrators

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"
)

// HTTPJourneyTemplatesClient is a simple HTTP implementation of JourneyTemplatesClient
type HTTPJourneyTemplatesClient struct {
    baseURL    string
    httpClient *http.Client
}

// NewHTTPJourneyTemplatesClient creates a new HTTP client
func NewHTTPJourneyTemplatesClient(baseURL string, httpClient *http.Client) JourneyTemplatesClient {
    if httpClient == nil {
        httpClient = &http.Client{Timeout: 1 * time.Second}
    }
    return &HTTPJourneyTemplatesClient{
        baseURL:    strings.TrimRight(baseURL, "/"),
        httpClient: httpClient,
    }
}

// GetTemplates fetches templates for a given parcel category
func (c *HTTPJourneyTemplatesClient) GetTemplates(ctx context.Context, parcelCategory string) (*TemplatesResponse, error) {
    endpoint := fmt.Sprintf("%s/api/v1/templates", c.baseURL)
    u, err := url.Parse(endpoint)
    if err != nil {
        return nil, err
    }
    q := u.Query()
    q.Set("parcel_category", parcelCategory)
    u.RawQuery = q.Encode()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Accept", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, fmt.Errorf("journey templates request failed with status %d", resp.StatusCode)
    }

    var parsed TemplatesResponse
    if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
        return nil, err
    }
    return &parsed, nil
}


