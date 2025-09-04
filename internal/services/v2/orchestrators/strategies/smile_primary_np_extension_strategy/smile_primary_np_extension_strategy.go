package smileprimarynpextensionstrategy

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"

    "prayog-serviceability-service/internal/services/v2/partners/common"
    "prayog-serviceability-service/internal/services/v2/partners/factory"
    "prayog-serviceability-service/internal/shared/models/v1"
    "github.com/sirupsen/logrus"
)

// SmilePrimaryNPExtensionStrategy runs a two-segment flow:
// 1) segment_1: call smile_hubops and smile_courier concurrently
// 2) segment_2: call shipyaari using pickup pincode from hubops hub_details
type SmilePrimaryNPExtensionStrategy struct{
    PartnerFactory factory.PartnerAdapterFactory
    Logger         *logrus.Logger
}

func (s *SmilePrimaryNPExtensionStrategy) Code() string { return "smile_primary_np_extension" }

func (s *SmilePrimaryNPExtensionStrategy) Execute(ctx context.Context, req *models.ServiceabilityV2Request) (*models.ServiceabilityV2Response, error) {
    if s.PartnerFactory == nil {
        return &models.ServiceabilityV2Response{Success: false, Partners: []models.PartnerV2Response{}}, nil
    }
    if s.Logger == nil {
        s.Logger = logrus.New()
    }

    partners := make([]models.PartnerV2Response, 0)
    var topLevelHubDetails interface{}

    s.Logger.WithFields(logrus.Fields{
        "component":       "smile_primary_np_extension_strategy",
        "action":          "execute_start",
        "source_postal":   req.SourcePostalCode,
        "dest_postal":     req.DestinationPostalCode,
        "parcel_category": req.ParcelCategory,
        "product_type":    req.ProductType,
    }).Info("Starting smile_primary_np_extension orchestration")

    // segment_1: smile_hubops + smile_courier
    type seg1Result struct {
        code   string
        result *common.PartnerServiceabilityResult
        err    error
    }
    seg1Codes := []string{"smile_hubops", "smile_courier"}
    seg1Results := make([]seg1Result, len(seg1Codes))

    var wg sync.WaitGroup
    s.Logger.WithFields(logrus.Fields{
        "component": "smile_primary_np_extension_strategy",
        "segment":   1,
        "partners":  seg1Codes,
    }).Info("Launching segment 1 partner calls in parallel")

    for i, code := range seg1Codes {
        wg.Add(1)
        go func(idx int, partnerCode string) {
            defer wg.Done()
            s.Logger.WithFields(logrus.Fields{
                "component":   "smile_primary_np_extension_strategy",
                "segment":     1,
                "partner_code": partnerCode,
                "step":        "get_adapter",
            }).Info("Resolving adapter")
            adapter, ok := s.PartnerFactory.GetAdapter(partnerCode)
            if !ok || adapter == nil || !adapter.IsHealthy(ctx) {
                s.Logger.WithFields(logrus.Fields{
                    "component":   "smile_primary_np_extension_strategy",
                    "segment":     1,
                    "partner_code": partnerCode,
                }).Warn("Adapter unavailable or unhealthy")
                seg1Results[idx] = seg1Result{code: partnerCode, result: nil, err: fmt.Errorf("adapter unavailable")}
                return
            }
            s.Logger.WithFields(logrus.Fields{
                "component":   "smile_primary_np_extension_strategy",
                "segment":     1,
                "partner_code": partnerCode,
                "step":        "check_serviceability",
            }).Info("Calling CheckServiceability")
            res, err := adapter.CheckServiceability(ctx, req, common.PartnerInfo{PartnerCode: partnerCode})
            seg1Results[idx] = seg1Result{code: partnerCode, result: res, err: err}
            s.Logger.WithFields(logrus.Fields{
                "component":    "smile_primary_np_extension_strategy",
                "segment":      1,
                "partner_code": partnerCode,
                "has_error":    err != nil,
            }).Info("Partner call completed")
        }(i, code)
    }
    wg.Wait()
    s.Logger.WithFields(logrus.Fields{
        "component": "smile_primary_np_extension_strategy",
        "segment":   1,
        "completed": len(seg1Results),
    }).Info("Segment 1 partner calls completed")

    // Extract hub details and build courier response (do not include hubops in partners)
    var pickupPincodeFromHubops string
    for _, r := range seg1Results {
        if r.err != nil || r.result == nil {
            continue
        }
        if r.code == "smile_hubops" {
            if r.result.Metadata != nil {
                if hd, ok := r.result.Metadata["hub_details"]; ok {
                        // Log raw hub_details to debug structure
                        if raw, err := json.Marshal(hd); err == nil {
                            s.Logger.WithFields(logrus.Fields{
                                "component": "smile_primary_np_extension_strategy",
                                "segment":   1,
                                "event":     "hub_details_raw",
                                "raw":       string(raw),
                            }).Info("HubOps hub_details payload")
                        }
                    topLevelHubDetails = hd
                        // Try to read destination3_pl_hub.pincode (preferred), or destination3pl_hub / destination3PLHub
                        // Fallback to hub_details.pincode
                    if m, ok := hd.(map[string]interface{}); ok {
                            // Log top-level keys to aid debugging
                            keys := make([]string, 0, len(m))
                            for k := range m { keys = append(keys, k) }
                            s.Logger.WithFields(logrus.Fields{
                                "component": "smile_primary_np_extension_strategy",
                                "segment":   1,
                                "event":     "hub_details_keys",
                                "keys":      keys,
                            }).Info("HubOps hub_details keys")
                            // Support multiple possible keys from HubOps response
                            for _, hubKey := range []string{"destination3_pl_hub", "destination3pl_hub", "destination3PLHub"} {
                                if d3pl, ok := m[hubKey]; ok {
                                    if mm, ok := d3pl.(map[string]interface{}); ok {
                                        // Log nested hub keys
                                        dkeys := make([]string, 0, len(mm))
                                        for k := range mm { dkeys = append(dkeys, k) }
                                        s.Logger.WithFields(logrus.Fields{
                                            "component": "smile_primary_np_extension_strategy",
                                            "segment":   1,
                                            "event":     "destination3pl_hub_variant_keys",
                                            "variant":   hubKey,
                                            "keys":      dkeys,
                                        }).Info("HubOps destination3pl_hub variant keys")
                                        if p, ok := mm["pincode"]; ok {
                                            pickupPincodeFromHubops = fmt.Sprintf("%v", p)
                                            s.Logger.WithFields(logrus.Fields{
                                                "component": "smile_primary_np_extension_strategy",
                                                "segment":   1,
                                                "event":     "pickup_pincode_found",
                                                "variant":   hubKey,
                                                "pincode":   pickupPincodeFromHubops,
                                            }).Info("Pickup pincode extracted from hub variant")
                                            break
                                        }
                                    }
                                }
                            }
                        if pickupPincodeFromHubops == "" { // fallback
                            if p, ok := m["pincode"]; ok {
                                pickupPincodeFromHubops = fmt.Sprintf("%v", p)
                            }
                        }
                        s.Logger.WithFields(logrus.Fields{
                            "component": "smile_primary_np_extension_strategy",
                            "segment":   1,
                            "event":     "hub_details_extracted",
                            "pincode":   pickupPincodeFromHubops,
                        }).Info("Extracted hub details from Smile HubOps")
                            if pickupPincodeFromHubops == "" {
                                s.Logger.WithFields(logrus.Fields{
                                    "component": "smile_primary_np_extension_strategy",
                                    "segment":   1,
                                    "event":     "hub_details_pincode_missing",
                                }).Warn("Pincode not found in hub_details")
                            }
                    }
                }
            }
            continue // do not include hubops in partners array
        }

        // Include smile_courier only if serviceable (has data) and no error
        if r.result != nil {
            hasServices := len(r.result.Services) > 0
            hasCapabilities := len(r.result.Capabilities) > 0
            hasMetadata := len(r.result.Metadata) > 0
            hasError := r.result.ErrorMessage != nil
            isServiceable := (hasServices || hasCapabilities || hasMetadata) && !hasError

            if isServiceable {
                partners = append(partners, toPartnerV2Response(r.result, r.code))
            } else {
                s.Logger.WithFields(logrus.Fields{
                    "component":    "smile_primary_np_extension_strategy",
                    "segment":      1,
                    "partner_code": r.code,
                    "has_services": hasServices,
                    "has_caps":     hasCapabilities,
                    "has_meta":     hasMetadata,
                    "has_error":    hasError,
                }).Info("Skipping non-serviceable partner from segment 1")
            }
        }
    }

    // segment_2: shipyaari using pickup pincode from hubops
    if pickupPincodeFromHubops != "" {
        // Build a shallow copy of request with SourcePostalCode set to hubops pincode
        cp := *req
        cp.SourcePostalCode = strPtr(pickupPincodeFromHubops)

        if adapter, ok := s.PartnerFactory.GetAdapter("shipyaari"); ok && adapter != nil && adapter.IsHealthy(ctx) {
            destPin := ""
            if cp.DestinationPostalCode != nil { destPin = *cp.DestinationPostalCode }
            s.Logger.WithFields(logrus.Fields{
                "component": "smile_primary_np_extension_strategy",
                "segment":   2,
                "partner_code": "shipyaari",
                "pickup_pincode": pickupPincodeFromHubops,
                "source_pincode": pickupPincodeFromHubops,
                "destination_pincode": destPin,
            }).Info("Calling Shipyaari with modified SourcePostalCode")
            if res, err := adapter.CheckServiceability(ctx, &cp, common.PartnerInfo{PartnerCode: "shipyaari"}); err == nil && res != nil {
                partners = append(partners, toPartnerV2Response(res, "shipyaari"))
                s.Logger.WithFields(logrus.Fields{
                    "component": "smile_primary_np_extension_strategy",
                    "segment":   2,
                    "partner_code": "shipyaari",
                }).Info("Shipyaari call completed")
            }
        }
    }

    // Final safety filter: remove smile_courier if it has an error (non-serviceable)
    if len(partners) > 0 {
        filtered := make([]models.PartnerV2Response, 0, len(partners))
        for _, p := range partners {
            if p.PartnerCode == "smile_courier" && p.Error != nil {
                s.Logger.WithFields(logrus.Fields{
                    "component":    "smile_primary_np_extension_strategy",
                    "partner_code": p.PartnerCode,
                    "event":        "filtered_non_serviceable",
                    "error":        *p.Error,
                }).Info("Excluding non-serviceable partner from final partners array")
                continue
            }
            filtered = append(filtered, p)
        }
        partners = filtered
    }

    resp := &models.ServiceabilityV2Response{
        Success:  len(partners) > 0,
        Partners: partners,
    }
    if topLevelHubDetails != nil {
        resp.HubDetails = topLevelHubDetails
    }
    s.Logger.WithFields(logrus.Fields{
        "component":        "smile_primary_np_extension_strategy",
        "action":           "execute_complete",
        "partners_count":   len(partners),
        "has_hub_details":  topLevelHubDetails != nil,
        "success":          resp.Success,
    }).Info("Orchestration completed")
    return resp, nil
}

func toPartnerV2Response(res *common.PartnerServiceabilityResult, code string) models.PartnerV2Response {
    partnerID := "unknown"
    if res != nil && res.PartnerID != nil {
        partnerID = res.PartnerID.String()
    }
    // Fallback: set known partner IDs from journey templates if not provided by adapter
    if partnerID == "unknown" {
        // These IDs originate from the journey templates API for parcel_category=courier
        // https://sandbox-apis.prayog.io/journey-service/api/v1/templates?parcel_category=courier
        switch code {
        case "smile_courier":
            partnerID = "ed8d5144-cc53-453e-b9d6-a69d59e1620c"
        case "shipyaari":
            partnerID = "be9fdb7c-3767-4a3f-854a-037fb745916a"
        }
    }
    md := map[string]interface{}{}
    if res != nil && res.Metadata != nil {
        md = res.Metadata
    }
    var errMsg *string
    if res != nil && res.ErrorMessage != nil {
        errMsg = res.ErrorMessage
    }
    services := []models.ServiceV2{}
    if res != nil && res.Services != nil {
        services = res.Services
    }
    return models.PartnerV2Response{
        PartnerID:       partnerID,
        PartnerCode:     code,
        PartnerName:     "",
        Rating:          0,
        Services:        services,
        PartnerServices: res.PartnerServices,
        Capabilities:    res.Capabilities,
        Error:           errMsg,
        ResponseTime:    res.ResponseTime,
        Metadata:        md,
    }
}

func strPtr(s string) *string { return &s }

// No explicit interface assertion to avoid import cycles; methods satisfy the interface.


