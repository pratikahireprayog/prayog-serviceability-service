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
type SmilePrimaryNPExtensionStrategy struct {
    PartnerFactory factory.PartnerAdapterFactory
    Logger         *logrus.Logger
}

func (s *SmilePrimaryNPExtensionStrategy) Code() string {
    return "smile_primary_np_extension"
}

func (s *SmilePrimaryNPExtensionStrategy) Execute(
    ctx context.Context,
    req *models.ServiceabilityV2Request,
) (*models.ServiceabilityV2Response, error) {

    if s.PartnerFactory == nil {
        return &models.ServiceabilityV2Response{
            Success:  false,
            Partners: []models.PartnerV2Response{},
        }, nil
    }

    if s.Logger == nil {
        s.Logger = logrus.New()
    }

    partners := make([]models.PartnerV2Response, 0)
    var topLevelHubDetails interface{}

    s.Logger.WithFields(logrus.Fields{
        "component": "smile_primary_np_extension_strategy",
        "action":    "execute_start",
    }).Info("Starting orchestration")

    /* ---------------- SEGMENT 1 ---------------- */

    type seg1Result struct {
        code   string
        result *common.PartnerServiceabilityResult
        err    error
    }

    seg1Codes := []string{"smile_hubops", "smile_courier"}
    seg1Results := make([]seg1Result, len(seg1Codes))

    var wg sync.WaitGroup

    for i, code := range seg1Codes {
        wg.Add(1)
        go func(idx int, partnerCode string) {
            defer wg.Done()

            adapter, ok := s.PartnerFactory.GetAdapter(partnerCode)
            if !ok || adapter == nil || !adapter.IsHealthy(ctx) {
                seg1Results[idx] = seg1Result{
                    code: partnerCode,
                    err:  fmt.Errorf("adapter unavailable"),
                }
                return
            }

            res, err := adapter.CheckServiceability(
                ctx,
                req,
                common.PartnerInfo{PartnerCode: partnerCode},
            )

            seg1Results[idx] = seg1Result{
                code:   partnerCode,
                result: res,
                err:    err,
            }
        }(i, code)
    }

    wg.Wait()

    /* ----------- PROCESS SEGMENT 1 RESULTS ----------- */

    var pickupPincodeFromHubops string

    for _, r := range seg1Results {

        if r.err != nil || r.result == nil {
            continue
        }

        /* ---------- HUBOPS ---------- */
        if r.code == "smile_hubops" {

            if r.result.Metadata != nil {
                if hd, ok := r.result.Metadata["hub_details"]; ok {

                    if raw, err := json.Marshal(hd); err == nil {
                        s.Logger.WithField("hub_details_raw", string(raw)).
                            Info("HubOps hub_details payload")
                    }

                    topLevelHubDetails = hd

                    if m, ok := hd.(map[string]interface{}); ok {
                        for _, key := range []string{
                            "destination3_pl_hub",
                            "destination3pl_hub",
                            "destination3PLHub",
                        } {
                            if v, ok := m[key]; ok {
                                if mm, ok := v.(map[string]interface{}); ok {
                                    if p, ok := mm["pincode"]; ok {
                                        pickupPincodeFromHubops = fmt.Sprintf("%v", p)
                                        break
                                    }
                                }
                            }
                        }

                        if pickupPincodeFromHubops == "" {
                            if p, ok := m["pincode"]; ok {
                                pickupPincodeFromHubops = fmt.Sprintf("%v", p)
                            }
                        }
                    }
                }
            }
            continue // HubOps never added to partners
        }

        /* ---------- SMILE COURIER (STRICT RULE) ---------- */
        if r.code == "smile_courier" {

            if isSmileCourierActuallyServiceable(r.result) {
                partners = append(partners, toPartnerV2Response(r.result, r.code))
            } else {
                s.Logger.WithFields(logrus.Fields{
                    "component":    "smile_primary_np_extension_strategy",
                    "partner_code": "smile_courier",
                    "reason":       "raw API marks serviceable=false",
                }).Info("Smile Courier skipped due to non-serviceable pincode")
            }
            continue
        }

        /* ---------- GENERIC PARTNERS ---------- */
        hasServices := len(r.result.Services) > 0
        hasCapabilities := len(r.result.Capabilities) > 0
        hasMetadata := len(r.result.Metadata) > 0
        hasError := r.result.ErrorMessage != nil

        isServiceable := (hasServices || hasCapabilities || hasMetadata) && !hasError

        if isServiceable {
            partners = append(partners, toPartnerV2Response(r.result, r.code))
        }
    }

    /* ---------------- SEGMENT 2 ---------------- */

    if pickupPincodeFromHubops != "" {

        cp := *req
        cp.SourcePostalCode = strPtr(pickupPincodeFromHubops)

        if adapter, ok := s.PartnerFactory.GetAdapter("shipyaari"); ok &&
            adapter != nil && adapter.IsHealthy(ctx) {

            if res, err := adapter.CheckServiceability(
                ctx,
                &cp,
                common.PartnerInfo{PartnerCode: "shipyaari"},
            ); err == nil && res != nil {

                partners = append(partners, toPartnerV2Response(res, "shipyaari"))
            }
        }
    }

    /* ---------------- RESPONSE ---------------- */

    resp := &models.ServiceabilityV2Response{
        Success:  len(partners) > 0,
        Partners: partners,
    }

    if topLevelHubDetails != nil {
        resp.HubDetails = topLevelHubDetails
    }

    return resp, nil
}

/* ================= HELPERS ================= */

func isSmileCourierActuallyServiceable(res *common.PartnerServiceabilityResult) bool {
    if res == nil || res.Metadata == nil {
        return false
    }

    // Use raw API response if available
    raw, ok := res.Metadata["data"].(map[string]interface{})
    if !ok {
        return false
    }

    svc, ok := raw["serviceable"].(bool)
    return ok && svc
}

func toPartnerV2Response(
    res *common.PartnerServiceabilityResult,
    code string,
) models.PartnerV2Response {

    partnerID := "unknown"
    if res != nil && res.PartnerID != nil {
        partnerID = res.PartnerID.String()
    }

    if partnerID == "unknown" {
        switch code {
        case "smile_courier":
            partnerID = "ed8d5144-cc53-453e-b9d6-a69d59e1620c"
        case "shipyaari":
            partnerID = "be9fdb7c-3767-4a3f-854a-037fb745916a"
        }
    }

    var errMsg *string
    if res != nil && res.ErrorMessage != nil {
        errMsg = res.ErrorMessage
    }

    return models.PartnerV2Response{
        PartnerID:       partnerID,
        PartnerCode:     code,
        Services:        res.Services,
        PartnerServices: res.PartnerServices,
        Capabilities:    res.Capabilities,
        Error:           errMsg,
        ResponseTime:    res.ResponseTime,
        Metadata:        res.Metadata,
    }
}

func strPtr(s string) *string {
    return &s
}
