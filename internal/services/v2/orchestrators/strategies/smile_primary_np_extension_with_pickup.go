package internationalwithpickupstrategy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/sirupsen/logrus"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/services/v2/partners/factory"
	"prayog-serviceability-service/internal/shared/models/v1"
    "github.com/google/uuid"
)

// InternationalWithPickupStrategy orchestrates the flow:
// 1) Pre-check Journey Templates for product_type=nba; return not serviceable if empty
// 2) Call smile_hubops to get hub pincodes
// 3) Porter: API source → destination = source_hub.pincode
// 4) Shipyaari: source = destination3_pl_hub.pincode (or destination3pl_hub) → API destination
// No hub details/addresses added to the response.
type InternationalWithPickupStrategy struct {
	PartnerFactory factory.PartnerAdapterFactory
	Logger         *logrus.Logger
}

func (s *InternationalWithPickupStrategy) Code() string { return "smile_primary_np_extension_with_pickup" }

func (s *InternationalWithPickupStrategy) Execute(ctx context.Context, req *models.ServiceabilityV2Request) (*models.ServiceabilityV2Response, error) {
	if s.PartnerFactory == nil {
		return &models.ServiceabilityV2Response{Success: false, Partners: []models.PartnerV2Response{}}, nil
	}
	if s.Logger == nil {
		s.Logger = logrus.New()
	}

	// Pre-check: Journey templates by product_type=nba
	if shouldBlock, err := s.isProductTypeBlocked(ctx, "nba"); err == nil {
		if shouldBlock {
			s.Logger.WithFields(logrus.Fields{
				"component": "international_with_pickup_strategy",
				"event":     "journey_templates_empty",
			}).Info("Templates API returned empty data for product_type, returning not serviceable")
			return &models.ServiceabilityV2Response{Success: false, Partners: []models.PartnerV2Response{}}, nil
		}
	} else {
		s.Logger.WithFields(logrus.Fields{
			"component": "international_with_pickup_strategy",
			"error":     err.Error(),
		}).Warn("Templates API check failed; proceeding with strategy by default")
	}

	partners := make([]models.PartnerV2Response, 0)
	var sourceHubPincode string
	var destination3PLHubPincode string
	var topLevelHubDetails interface{}

	// Step 1: Call Smile HubOps with source/destination pincodes to get hubs
	if adapter, ok := s.PartnerFactory.GetAdapter("smile_hubops"); ok && adapter != nil && adapter.IsHealthy(ctx) {
		s.Logger.WithFields(logrus.Fields{
			"component":    "international_with_pickup_strategy",
			"step":         1,
			"partner_code": "smile_hubops",
		}).Info("Calling Smile HubOps for hub resolution")

		res, err := adapter.CheckServiceability(ctx, req, common.PartnerInfo{PartnerCode: "smile_hubops"})
		if err == nil && res != nil && res.Metadata != nil {
			if hd, ok := res.Metadata["hub_details"]; ok {
				topLevelHubDetails = hd
				// Expect snake_case keys in Smile HubOps payload
				if m, ok := hd.(map[string]interface{}); ok {
					if srcHub, ok := m["source_hub"].(map[string]interface{}); ok {
						if pin, ok := srcHub["pincode"]; ok { sourceHubPincode = toString(pin) }
					}
					// destination 3PL hub can appear as destination3_pl_hub (preferred) or destination3pl_hub
					if dest3pl, ok := m["destination3_pl_hub"].(map[string]interface{}); ok {
						if pin, ok := dest3pl["pincode"]; ok { destination3PLHubPincode = toString(pin) }
					} else if dest3pl, ok := m["destination3pl_hub"].(map[string]interface{}); ok {
						if pin, ok := dest3pl["pincode"]; ok { destination3PLHubPincode = toString(pin) }
					}
					topLevelHubDetails = hd
				}
				s.Logger.WithFields(logrus.Fields{
					"component":            "international_with_pickup_strategy",
					"event":                "hubops_hub_extracted",
					"porter_dest_pin":      sourceHubPincode,
					"shipyaari_source_pin": destination3PLHubPincode,
				}).Info("Extracted hubs from Smile HubOps payload")
			}
		} else if err != nil {
			s.Logger.WithFields(logrus.Fields{
				"component": "international_with_pickup_strategy",
				"partner":   "smile_hubops",
				"error":     err.Error(),
			}).Warn("Smile HubOps call failed")
		}
	} else {
		s.Logger.WithFields(logrus.Fields{
			"component": "international_with_pickup_strategy",
			"partner":   "smile_hubops",
		}).Warn("Smile HubOps adapter unavailable or unhealthy")
	}

	// Step 2: Call Porter with source -> sourceHub pincode (only if found)
	if sourceHubPincode != "" {
		if adapter, ok := s.PartnerFactory.GetAdapter("porter"); ok && adapter != nil && adapter.IsHealthy(ctx) {
			// Build modified request for Porter
			cp := *req
			cp.DestinationPostalCode = strPtr(sourceHubPincode)

			// Map Porter partner ID from Journey Templates for partner_code=porter_2w
			porter2WPartnerID := s.getPartnerIDFromTemplates(ctx, "porter_2w")

			s.Logger.WithFields(logrus.Fields{
				"component":           "international_with_pickup_strategy",
				"step":                2,
				"partner_code":        "porter_2w",
				"source_pincode":      valOrEmpty(cp.SourcePostalCode),
				"destination_pincode": sourceHubPincode,
			}).Info("Calling Porter for source → hub pickup feasibility")

			if res, err := adapter.CheckServiceability(ctx, &cp, common.PartnerInfo{PartnerCode: "porter_2w", PartnerID: porter2WPartnerID}); err == nil && res != nil {
				if p, ok := toPartnerV2Response(res, "porter_2w"); ok {
					partners = append(partners, p)
				}
			} else if err != nil {
				s.Logger.WithFields(logrus.Fields{
					"component": "international_with_pickup_strategy",
					"partner":   "porter_2w",
					"error":     err.Error(),
				}).Warn("Porter call failed")
			}
		} else {
			s.Logger.WithFields(logrus.Fields{
				"component": "international_with_pickup_strategy",
				"partner":   "porter_2w",
			}).Warn("Porter adapter unavailable or unhealthy")
		}
	} else {
		s.Logger.WithFields(logrus.Fields{
			"component": "international_with_pickup_strategy",
			"event":     "source_hub_pincode_missing",
		}).Warn("Skipping Porter step as hub pincode was not found")
	}

	// Step 3: Call Shipyaari with destination_3pl_hub pincode -> destination (only if found)
	if destination3PLHubPincode != "" {
		if adapter, ok := s.PartnerFactory.GetAdapter("shipyaari"); ok && adapter != nil && adapter.IsHealthy(ctx) {
			cp := *req
			cp.SourcePostalCode = strPtr(destination3PLHubPincode)

            // Map Shipyaari partner ID from Journey Templates (product_type=nba)
            shipyaariPartnerID := s.getPartnerIDFromTemplates(ctx, "shipyaari")
            if shipyaariPartnerID != nil {
                s.Logger.WithFields(logrus.Fields{
                    "component":    "international_with_pickup_strategy",
                    "partner_code": "shipyaari",
                    "partner_id":   shipyaariPartnerID.String(),
                }).Info("Mapped Shipyaari partner_id from templates")
            } else {
                s.Logger.WithFields(logrus.Fields{
                    "component":    "international_with_pickup_strategy",
                    "partner_code": "shipyaari",
                }).Warn("Could not map Shipyaari partner_id from templates; proceeding without it")
            }

			s.Logger.WithFields(logrus.Fields{
				"component":           "international_with_pickup_strategy",
				"step":                3,
				"partner_code":        "shipyaari",
				"source_pincode":      destination3PLHubPincode,
				"destination_pincode": valOrEmpty(cp.DestinationPostalCode),
			}).Info("Calling Shipyaari for 3PL hub → destination leg")

			if res, err := adapter.CheckServiceability(ctx, &cp, common.PartnerInfo{PartnerCode: "shipyaari", PartnerID: shipyaariPartnerID}); err == nil && res != nil {
				if p, ok := toPartnerV2Response(res, "shipyaari"); ok {
					partners = append(partners, p)
				}
			} else if err != nil {
				s.Logger.WithFields(logrus.Fields{
					"component": "international_with_pickup_strategy",
					"partner":   "shipyaari",
					"error":     err.Error(),
				}).Warn("Shipyaari call failed")
			}
		} else {
			s.Logger.WithFields(logrus.Fields{
				"component": "international_with_pickup_strategy",
				"partner":   "shipyaari",
			}).Warn("Shipyaari adapter unavailable or unhealthy")
		}
	} else {
		s.Logger.WithFields(logrus.Fields{
			"component": "international_with_pickup_strategy",
			"event":     "destination_3pl_hub_pincode_missing",
		}).Warn("Skipping Shipyaari step as 3PL hub pincode was not found")
	}

	// Build response (no hub details or addresses per latest instruction)
	resp := &models.ServiceabilityV2Response{
		Success:  len(partners) > 0,
		Partners: partners,
	}
	if topLevelHubDetails != nil {
		resp.HubDetails = topLevelHubDetails
	}

	return resp, nil
}

// isProductTypeBlocked calls journey templates API with product_type=nba and returns true when data is empty
func (s *InternationalWithPickupStrategy) isProductTypeBlocked(ctx context.Context, productType string) (bool, error) {
	baseURL := os.Getenv("JOURNEY_TEMPLATE_URL")
	if baseURL == "" {
		baseURL = "https://sandbox-apis.prayog.io"
	}
	endpoint, _ := url.Parse(baseURL)
	endpoint.Path = "/journey-service/api/v1/templates"
	q := endpoint.Query()
	// Hardcode product_type=nba per requirement
	q.Set("product_type", "nba")
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil { return false, err }
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil { return false, err }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, nil // treat as non-blocking on non-2xx
	}
	var parsed struct {
		Success bool              `json:"success"`
		Data    []json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return false, err
	}
	return len(parsed.Data) == 0, nil
}

// getPartnerIDFromTemplates fetches journey templates for product_type=nba and extracts the partner_id for the given code
func (s *InternationalWithPickupStrategy) getPartnerIDFromTemplates(ctx context.Context, partnerCode string) *uuid.UUID {
    baseURL := os.Getenv("JOURNEY_TEMPLATE_URL")
    if baseURL == "" {
        baseURL = "https://sandbox-apis.prayog.io"
    }
    endpoint, _ := url.Parse(baseURL)
    endpoint.Path = "/journey-service/api/v1/templates"
    q := endpoint.Query()
    q.Set("product_type", "nba")
    endpoint.RawQuery = q.Encode()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
    if err != nil { return nil }
    req.Header.Set("Accept", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil { return nil }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil
    }
    var parsed struct {
        Success bool `json:"success"`
        Data    []struct {
            Routes []struct {
                Partners []struct {
                    PartnerCode string     `json:"partner_code"`
                    PartnerID   *uuid.UUID `json:"partner_id"`
                } `json:"partners"`
            } `json:"routes"`
        } `json:"data"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
        return nil
    }
    // Scan for partner code match
    for _, t := range parsed.Data {
        for _, r := range t.Routes {
            for _, p := range r.Partners {
                if p.PartnerCode == partnerCode && p.PartnerID != nil {
                    return p.PartnerID
                }
            }
        }
    }
    return nil
}

func toPartnerV2Response(res *common.PartnerServiceabilityResult, code string) (models.PartnerV2Response, bool) {
	if res == nil {
		return models.PartnerV2Response{}, false
	}
	// Exclude non-serviceable (error present)
	if res.ErrorMessage != nil {
		return models.PartnerV2Response{}, false
	}
	// Include if any of services/capabilities/metadata present
	hasServices := len(res.Services) > 0
	hasCaps := len(res.Capabilities) > 0
	hasMeta := len(res.Metadata) > 0
	if !(hasServices || hasCaps || hasMeta) {
		return models.PartnerV2Response{}, false
	}
	partnerID := "unknown"
	if res.PartnerID != nil {
		partnerID = res.PartnerID.String()
	}
	return models.PartnerV2Response{
		PartnerID:       partnerID,
		PartnerCode:     code,
		PartnerName:     "",
		Rating:          0,
		Services:        res.Services,
		PartnerServices: res.PartnerServices,
		Capabilities:    res.Capabilities,
		Error:           res.ErrorMessage,
		ResponseTime:    res.ResponseTime,
		Metadata:        res.Metadata,
	}, true
}

func strPtr(s string) *string { return &s }

func valOrEmpty(p *string) string {
	if p == nil { return "" }
	return *p
}

// toString converts any value to string using a simple %v formatting
func toString(v interface{}) string {
	if v == nil { return "" }
	switch t := v.(type) {
	case string:
		return t
	case int:
		return fmtInt64(int64(t))
	case int64:
		return fmtInt64(t)
	case float64:
		return fmtInt64(int64(t))
	default:
		return fmtInt64(0)
	}
}

func fmtInt64(n int64) string {
	// Minimal int to string without importing strconv
	if n == 0 { return "0" }
	neg := false
	if n < 0 { neg = true; n = -n }
	buf := make([]byte, 0, 20)
	for n > 0 {
		d := n % 10
		buf = append([]byte{byte('0' + d)}, buf...)
		n /= 10
	}
	if neg { buf = append([]byte{'-'}, buf...) }
	return string(buf)
}


