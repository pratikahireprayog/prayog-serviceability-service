package np_extension_with_pickup_and_delivery_strategy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/google/uuid"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/services/v2/partners/factory"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// NPExtensionWithPickupAndDeliveryStrategy orchestrates the flow:
// 1) Call smile_hubops to get hub pincodes (source and destination)
// 2) Porter Segment 1: source → sourceHub pincode (pickup)
// 3) Porter Segment 2: sourceHub → destination3PLHub (transfer through operational hub)
// 4) Porter Segment 3: destination3PLHub → destination (final delivery on event trigger)
type NPExtensionWithPickupAndDeliveryStrategy struct {
	PartnerFactory factory.PartnerAdapterFactory
	Logger         *logrus.Logger
}

func (s *NPExtensionWithPickupAndDeliveryStrategy) Code() string { 
	return "np_extension_with_pickup_and_delivery" 
}

func (s *NPExtensionWithPickupAndDeliveryStrategy) Execute(ctx context.Context, req *models.ServiceabilityV2Request) (*models.ServiceabilityV2Response, error) {
	if s.PartnerFactory == nil {
		return &models.ServiceabilityV2Response{Success: false, Partners: []models.PartnerV2Response{}}, nil
	}
	if s.Logger == nil {
		s.Logger = logrus.New()
	}

	partners := make([]models.PartnerV2Response, 0)
	var sourceHubPincode string
	var destination3PLHubPincode string
	var topLevelHubDetails interface{}

	// Step 1: Call Smile HubOps with source/destination pincodes to get hubs
	if adapter, ok := s.PartnerFactory.GetAdapter("smile_hubops"); ok && adapter != nil && adapter.IsHealthy(ctx) {
		s.Logger.WithFields(logrus.Fields{
			"component":    "np_extension_with_pickup_and_delivery_strategy",
			"step":         1,
			"partner_code": "smile_hubops",
		}).Info("Calling Smile HubOps for hub resolution")

		// Use partner ID from Journey Templates API for smile_hubops
		smileHubOpsPartnerID := s.getPartnerIDFromTemplates(ctx, "smile_hubops")
		
		res, err := adapter.CheckServiceability(ctx, req, common.PartnerInfo{PartnerCode: "smile_hubops", PartnerID: smileHubOpsPartnerID})
		if err == nil && res != nil {
					// Note: Smile HubOps is not added to partners array - only used for hub resolution
			
			if res.Metadata != nil {
				if hd, ok := res.Metadata["hub_details"]; ok {
					topLevelHubDetails = hd
					// Extract pincodes from Smile HubOps response
					if m, ok := hd.(map[string]interface{}); ok {
						// Log the actual structure for debugging
						s.Logger.WithFields(logrus.Fields{
							"component": "np_extension_with_pickup_and_delivery_strategy",
							"event":     "hub_details_structure",
							"keys":      getMapKeys(m),
						}).Info("Hub details structure received from Smile HubOps")
						
						// Try multiple possible key formats for source hub
						if srcHub, ok := m["sourceHub"].(map[string]interface{}); ok {
							if pin, ok := srcHub["pincode"]; ok { 
								sourceHubPincode = toString(pin) 
								s.Logger.WithFields(logrus.Fields{
									"component": "np_extension_with_pickup_and_delivery_strategy",
									"event":     "source_hub_pincode_found",
									"key":       "sourceHub",
									"pincode":   sourceHubPincode,
								}).Info("Extracted source hub pincode from sourceHub key")
							}
						} else if srcHub, ok := m["source_hub"].(map[string]interface{}); ok {
							if pin, ok := srcHub["pincode"]; ok { 
								sourceHubPincode = toString(pin) 
								s.Logger.WithFields(logrus.Fields{
									"component": "np_extension_with_pickup_and_delivery_strategy",
									"event":     "source_hub_pincode_found",
									"key":       "source_hub",
									"pincode":   sourceHubPincode,
								}).Info("Extracted source hub pincode from source_hub key")
							}
						}
						
						// Try multiple possible key formats for destination 3PL hub
						if dest3pl, ok := m["destination3PLHub"].(map[string]interface{}); ok {
							if pin, ok := dest3pl["pincode"]; ok { 
								destination3PLHubPincode = toString(pin) 
								s.Logger.WithFields(logrus.Fields{
									"component": "np_extension_with_pickup_and_delivery_strategy",
									"event":     "dest_3pl_hub_pincode_found",
									"key":       "destination3PLHub",
									"pincode":   destination3PLHubPincode,
								}).Info("Extracted destination 3PL hub pincode from destination3PLHub key")
							}
						} else if dest3pl, ok := m["destination3_pl_hub"].(map[string]interface{}); ok {
							if pin, ok := dest3pl["pincode"]; ok { 
								destination3PLHubPincode = toString(pin) 
								s.Logger.WithFields(logrus.Fields{
									"component": "np_extension_with_pickup_and_delivery_strategy",
									"event":     "dest_3pl_hub_pincode_found",
									"key":       "destination3_pl_hub",
									"pincode":   destination3PLHubPincode,
								}).Info("Extracted destination 3PL hub pincode from destination3_pl_hub key")
							}
						} else if dest3pl, ok := m["destination3pl_hub"].(map[string]interface{}); ok {
							if pin, ok := dest3pl["pincode"]; ok { 
								destination3PLHubPincode = toString(pin) 
								s.Logger.WithFields(logrus.Fields{
									"component": "np_extension_with_pickup_and_delivery_strategy",
									"event":     "dest_3pl_hub_pincode_found",
									"key":       "destination3pl_hub",
									"pincode":   destination3PLHubPincode,
								}).Info("Extracted destination 3PL hub pincode from destination3pl_hub key")
							}
						}
						
						// Inject partner_id for hub_details from Journey Templates
						if smileHubOpsPartnerID != nil {
							m["partner_id"] = smileHubOpsPartnerID.String()
						}
						topLevelHubDetails = m
					}
					s.Logger.WithFields(logrus.Fields{
						"component":            "np_extension_with_pickup_and_delivery_strategy",
						"event":                "hubops_hub_extracted",
						"source_hub_pincode":   sourceHubPincode,
						"dest_3pl_hub_pincode": destination3PLHubPincode,
					}).Info("Extracted hubs from Smile HubOps payload")
				}
			}
		} else if err != nil {
			s.Logger.WithFields(logrus.Fields{
				"component": "np_extension_with_pickup_and_delivery_strategy",
				"partner":   "smile_hubops",
				"error":     err.Error(),
			}).Warn("Smile HubOps call failed")
		}
	} else {
		s.Logger.WithFields(logrus.Fields{
			"component": "np_extension_with_pickup_and_delivery_strategy",
			"partner":   "smile_hubops",
		}).Warn("Smile HubOps adapter unavailable or unhealthy")
	}

	// If Smile HubOps didn't return hub details, we cannot proceed
	// Return empty partners array (no serviceability)
	if topLevelHubDetails == nil {
		s.Logger.WithFields(logrus.Fields{
			"component": "np_extension_with_pickup_and_delivery_strategy",
			"event":     "no_hub_details",
		}).Warn("Smile HubOps returned no hub details - cannot proceed with pickup and delivery")
		return &models.ServiceabilityV2Response{
			Success:  false,
			Partners: []models.PartnerV2Response{},
		}, nil
	}

	// Step 2: Porter Segment 1 - source → sourceHub pincode (pickup)
	if sourceHubPincode != "" {
		if adapter, ok := s.PartnerFactory.GetAdapter("porter"); ok && adapter != nil && adapter.IsHealthy(ctx) {
					// Build modified request for Porter Segment 1
		cp := *req
		cp.DestinationPostalCode = strPtr(sourceHubPincode)

		// Use partner ID from Journey Templates API for porter_2w
		porter2WPartnerID := s.getPartnerIDFromTemplates(ctx, "porter_2w")

		s.Logger.WithFields(logrus.Fields{
				"component":           "np_extension_with_pickup_and_delivery_strategy",
				"step":                2,
				"segment":             "segment_1",
				"partner_code":        "porter_2w",
				"source_pincode":      valOrEmpty(cp.SourcePostalCode),
				"destination_pincode": sourceHubPincode,
			}).Info("Calling Porter Segment 1 for source → hub pickup")

			if res, err := adapter.CheckServiceability(ctx, &cp, common.PartnerInfo{PartnerCode: "porter_2w", PartnerID: porter2WPartnerID}); err == nil && res != nil {
				if p, ok := toPartnerV2Response(res, "porter_2w"); ok {
					// Add segment metadata
					p.Metadata["segment_id"] = "seg_1_dea9825ef6174fbabc0cbe2c69b6b40c"
					p.Metadata["segment_code"] = "segment_1"
					p.Metadata["segment_type"] = "transport"
					p.Metadata["sequence"] = 1
					partners = append(partners, p)
				}
			} else if err != nil {
				s.Logger.WithFields(logrus.Fields{
					"component": "np_extension_with_pickup_and_delivery_strategy",
					"partner":   "porter_2w",
					"segment":   "segment_1",
					"error":     err.Error(),
				}).Warn("Porter Segment 1 call failed")
			}
		} else {
			s.Logger.WithFields(logrus.Fields{
				"component": "np_extension_with_pickup_and_delivery_strategy",
				"partner":   "porter_2w",
				"segment":   "segment_1",
			}).Warn("Porter adapter unavailable or unhealthy")
		}
	} else {
		s.Logger.WithFields(logrus.Fields{
			"component": "np_extension_with_pickup_and_delivery_strategy",
			"event":     "source_hub_pincode_missing",
		}).Warn("Skipping Porter Segment 1 as hub pincode was not found")
	}

	// Step 3: Smile HubOps Segment 2 - Hub details are preserved in topLevelHubDetails
	// but not added to partners array (only used for hub resolution)
	if topLevelHubDetails != nil {
		s.Logger.WithFields(logrus.Fields{
			"component": "np_extension_with_pickup_and_delivery_strategy",
			"step":      3,
			"event":     "hub_details_preserved",
		}).Info("Hub details preserved for response (not added to partners array)")
	}

	// Step 4: Porter Segment 3 - destination3PLHub → destination (final delivery on event trigger)
	if destination3PLHubPincode != "" {
		if adapter, ok := s.PartnerFactory.GetAdapter("porter"); ok && adapter != nil && adapter.IsHealthy(ctx) {
					// Build modified request for Porter Segment 3
		cp := *req
		cp.SourcePostalCode = strPtr(destination3PLHubPincode)

		// Use partner ID from Journey Templates API for porter_2w
		porter2WPartnerID := s.getPartnerIDFromTemplates(ctx, "porter_2w")

		s.Logger.WithFields(logrus.Fields{
				"component":           "np_extension_with_pickup_and_delivery_strategy",
				"step":                4,
				"segment":             "segment_3",
				"partner_code":        "porter_2w",
				"source_pincode":      destination3PLHubPincode,
				"destination_pincode": valOrEmpty(cp.DestinationPostalCode),
			}).Info("Calling Porter Segment 3 for 3PL hub → destination final delivery")

			if res, err := adapter.CheckServiceability(ctx, &cp, common.PartnerInfo{PartnerCode: "porter_2w", PartnerID: porter2WPartnerID}); err == nil && res != nil {
				if p, ok := toPartnerV2Response(res, "porter_2w"); ok {
					// Add segment metadata
					p.Metadata["segment_id"] = "seg_3_7d047c0f229144ceabd1d08860ea1f0a"
					p.Metadata["segment_code"] = "segment_3"
					p.Metadata["segment_type"] = "transport"
					p.Metadata["sequence"] = 3
					p.Metadata["trigger_type"] = "on_event"
					partners = append(partners, p)
				}
			} else if err != nil {
				s.Logger.WithFields(logrus.Fields{
					"component": "np_extension_with_pickup_and_delivery_strategy",
					"partner":   "porter_2w",
					"segment":   "segment_3",
					"error":     err.Error(),
				}).Warn("Porter Segment 3 call failed")
			}
		} else {
			s.Logger.WithFields(logrus.Fields{
				"component": "np_extension_with_pickup_and_delivery_strategy",
				"partner":   "porter_2w",
				"segment":   "segment_3",
			}).Warn("Porter adapter unavailable or unhealthy")
		}
	} else {
		s.Logger.WithFields(logrus.Fields{
			"component": "np_extension_with_pickup_and_delivery_strategy",
			"event":     "destination_3pl_hub_pincode_missing",
		}).Warn("Skipping Porter Segment 3 as 3PL hub pincode was not found")
	}

	// Build response
	resp := &models.ServiceabilityV2Response{
		Success:  len(partners) > 0,
		Partners: partners,
	}
	if topLevelHubDetails != nil {
		resp.HubDetails = topLevelHubDetails
	}

	return resp, nil
}

// getPartnerIDFromTemplates fetches journey templates for product_type=pickup_and_delivery and extracts the partner_id for the given code
func (s *NPExtensionWithPickupAndDeliveryStrategy) getPartnerIDFromTemplates(ctx context.Context, partnerCode string) *uuid.UUID {
    baseURL := os.Getenv("JOURNEY_TEMPLATE_URL")
    if baseURL == "" {
        baseURL = "https://sandbox-apis.prayog.io"
    }
    endpoint, _ := url.Parse(baseURL)
    endpoint.Path = "/journey-service/api/v1/templates"
    q := endpoint.Query()
    q.Set("product_type", "pickup_and_delivery")
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

// getMapKeys extracts all keys from a map for debugging purposes
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
