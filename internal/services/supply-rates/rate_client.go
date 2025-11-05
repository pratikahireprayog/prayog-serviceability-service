package supplyrates

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"

	"github.com/sirupsen/logrus"
)

// RateClient handles all Supply Rate API interactions.
type RateClient struct {
	Logger  *logrus.Logger
	Client  *http.Client
	BaseURL string

}

// RateQuoteResponse defines the structure of the Supply Rate API response.
type RateQuoteResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Metadata struct {
		RequestID         string `json:"request_id"`
		ResponseTimeMs    int64  `json:"response_time_ms"`
		PartnersQueried   int    `json:"partners_queried"`
		PartnersSucceeded int    `json:"partners_succeeded"`
		PartnersFailed    int    `json:"partners_failed"`
		TotalRatesFound   int    `json:"total_rates_found"`
	} `json:"metadata"`
	Data struct {
		SuccessfulResponses []struct {
			Partner struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"partner"`
			Source         string `json:"source"`
			AvailableRates []struct {
				RateID       string `json:"rate_id"`
				Service      string `json:"service"`
				DeliveryDays int    `json:"delivery_days"`
				Price        struct {
					Currency    string  `json:"currency"`
					Amount      float64 `json:"amount"`
					Type        string  `json:"type"`
					ServiceType string  `json:"ServiceType"`
				} `json:"price"`
			} `json:"available_rates"`
			ResponseTimeMs int64 `json:"response_time_ms"`
		} `json:"successful_responses"`
		FailedResponses interface{} `json:"failed_responses"`
	} `json:"data"`
	Timestamp string `json:"timestamp"`
}

// RateRequestPayload defines the structure of the Supply Rate API request.
type RateRequestPayload struct {
	SourceLocation      RateLocation       `json:"source_location"`
	DestinationLocation RateLocation       `json:"destination_location"`
	Packages            []RatePackage      `json:"packages"`
	Partners            []RatePartnerEntry `json:"partners"`
	Metadata            map[string]string  `json:"metadata"`
}

type RateLocation struct {
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

type RatePartnerEntry struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type RatePackage struct {
	Weight struct {
		Value float64 `json:"value"`
		Unit  string  `json:"unit"`
	} `json:"weight"`
	Dimensions struct {
		Length float64 `json:"length"`
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
		Unit   string  `json:"unit"`
	} `json:"dimensions"`
}

// NewRateClient creates a reusable client.
func NewRateClient(logger *logrus.Logger) *RateClient {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}
	logger.Info("aaaaaaa", os.Getenv("SUPPLY_RATE_URL"));
	return &RateClient{
		Logger:  logger,
		Client:  &http.Client{Timeout: 30 * time.Second},
		BaseURL: os.Getenv("SUPPLY_RATE_URL"),
	}
}

// GetRatesForPartners builds payload + calls Supply Rate API.
func (rc *RateClient) GetRatesForPartners(
	ctx context.Context,
	req *modelsv1.ServiceabilityV2Request,
	srcPin, srcCC, dstPin, dstCC string,
	serviceable []modelsv1.PartnerV2Response,
) (*RateQuoteResponse, error) {
	ratesURL := os.Getenv("SUPPLY_RATE_URL")
	fmt.Println("ratesURL", ratesURL);
	if rc.BaseURL == "" {
		// rc.BaseURL = "https://sandbox-apis.prayog.io/supply-rate/v1/quotes"
		return nil, fmt.Errorf("SUPPLY_RATE_URL not configured")
	}

	// 1️⃣ Build request payload
	payload := rc.buildRatesRequestPayload(req, srcPin, srcCC, dstPin, dstCC, serviceable)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// 2️⃣ Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, rc.BaseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Prayog-Serviceability-Service/1.0")

	rc.Logger.WithFields(logrus.Fields{
		"component": "supply_rates_client",
		"url":       rc.BaseURL,
		"partners":  len(payload.Partners),
		"packages":  len(payload.Packages),
	}).Info("Calling Supply Rate API")

	// 3️⃣ Execute HTTP request
	resp, err := rc.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call rate API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read rate API response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		rc.Logger.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(bodyBytes),
		}).Warn("Supply Rate API returned non-200 status")
		return nil, fmt.Errorf("rate API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 4️⃣ Parse response
	var rateResp RateQuoteResponse
	if err := json.Unmarshal(bodyBytes, &rateResp); err != nil {
		return nil, fmt.Errorf("failed to parse rate response: %w", err)
	}

	if !rateResp.Success {
		rc.Logger.WithField("message", rateResp.Message).Warn("Rate API returned unsuccessful response")
		return nil, fmt.Errorf("rate API returned error: %s", rateResp.Message)
	}

	rc.Logger.WithFields(logrus.Fields{
		"component":          "supply_rates_client",
		"rates_found":        rateResp.Metadata.TotalRatesFound,
		"partners_succeeded": rateResp.Metadata.PartnersSucceeded,
		"response_time_ms":   rateResp.Metadata.ResponseTimeMs,
	}).Info("Successfully retrieved rates from API")

	return &rateResp, nil
}

// buildRatesRequestPayload creates the payload for the rates API.
// func (rc *RateClient) buildRatesRequestPayload(
// 	req *modelsv1.ServiceabilityV2Request,
// 	srcPin, srcCC, dstPin, dstCC string,
// 	serviceable []modelsv1.PartnerV2Response,
// ) RateRequestPayload {

// 	// --- Build Packages ---
// 	var pkgs []RatePackage
// 	if req != nil && len(req.Packages) > 0 {
// 		for _, p := range req.Packages {
// 			var pkg RatePackage
// 			// weight
// 			if p.Weight != nil && p.Weight.Value > 0 {
// 				pkg.Weight.Value = p.Weight.Value
// 			} else {
// 				pkg.Weight.Value = 1.0
// 			}
// 			pkg.Weight.Unit = "kg"

// 			// dimensions
// 			if p.Dimensions != nil {
// 				if p.Dimensions.Length > 0 {
// 					pkg.Dimensions.Length = p.Dimensions.Length
// 				} else {
// 					pkg.Dimensions.Length = 10
// 				}
// 				if p.Dimensions.Width > 0 {
// 					pkg.Dimensions.Width = p.Dimensions.Width
// 				} else {
// 					pkg.Dimensions.Width = 10
// 				}
// 				if p.Dimensions.Height > 0 {
// 					pkg.Dimensions.Height = p.Dimensions.Height
// 				} else {
// 					pkg.Dimensions.Height = 10
// 				}
// 			} else {
// 				pkg.Dimensions.Length = 10
// 				pkg.Dimensions.Width = 10
// 				pkg.Dimensions.Height = 10
// 			}
// 			pkg.Dimensions.Unit = "cm"
// 			pkgs = append(pkgs, pkg)
// 		}
// 	} else {
// 		pkgs = []RatePackage{{
// 			Weight: struct {
// 				Value float64 `json:"value"`
// 				Unit  string  `json:"unit"`
// 			}{1.0, "kg"},
// 			Dimensions: struct {
// 				Length float64 `json:"length"`
// 				Width  float64 `json:"width"`
// 				Height float64 `json:"height"`
// 				Unit   string  `json:"unit"`
// 			}{10, 10, 10, "cm"},
// 		}}
// 	}

// 	// --- Build Partners ---
// 	var partners []RatePartnerEntry
// 	for _, p := range serviceable {
// 		partners = append(partners, RatePartnerEntry{
// 			ID:   p.PartnerID,
// 			Code: strings.ToUpper(p.PartnerCode),
// 		})
// 	}

// 	// --- Metadata ---
// 	meta := map[string]string{
// 		"currency":     getEnvOrDefault("RATE_CURRENCY", "INR"),
// 		"service_type": getEnvOrDefault("RATE_SERVICE_TYPE", "express"),
// 		"flow":         "international",
// 	}

// 	return RateRequestPayload{
// 		SourceLocation: RateLocation{
// 			PostalCode:  srcPin,
// 			CountryCode: srcCC,
// 		},
// 		DestinationLocation: RateLocation{
// 			PostalCode:  dstPin,
// 			CountryCode: dstCC,
// 		},
// 		Packages: pkgs,
// 		Partners: partners,
// 		Metadata: meta,
// 	}
// }

func (rc *RateClient) buildRatesRequestPayload(
	req *modelsv1.ServiceabilityV2Request, 
	srcPin, srcCC, dstPin, dstCC string, 
	serviceable []modelsv1.PartnerV2Response,
) RateRequestPayload {
	// Build packages
	pkgs := make([]RatePackage, 0)
	if req != nil && len(req.Packages) > 0 {
		for _, p := range req.Packages {
			var rp RatePackage
			
			// Weight
			w := defaultWeight(req)
			if p.Weight != nil && p.Weight.Value > 0 {
				w = p.Weight.Value
			}
			rp.Weight.Value = w
			rp.Weight.Unit = "kg"
			
			// Dimensions
			l := defaultLen(req)
			wid := defaultWid(req)
			h := defaultHei(req)
			
			if p.Dimensions != nil {
				if p.Dimensions.Length > 0 {
					l = p.Dimensions.Length
				}
				if p.Dimensions.Width > 0 {
					wid = p.Dimensions.Width
				}
				if p.Dimensions.Height > 0 {
					h = p.Dimensions.Height
				}
			}
			
			rp.Dimensions.Length = l
			rp.Dimensions.Width = wid
			rp.Dimensions.Height = h
			rp.Dimensions.Unit = "cm"
			
			pkgs = append(pkgs, rp)
		}
	} else {
		// Default package
		var rp RatePackage
		rp.Weight.Value = 1
		rp.Weight.Unit = "kg"
		rp.Dimensions.Length = 10
		rp.Dimensions.Width = 10
		rp.Dimensions.Height = 10
		rp.Dimensions.Unit = "cm"
		pkgs = append(pkgs, rp)
	}

	// Build partners list
	partners := make([]RatePartnerEntry, 0, len(serviceable))
	for _, p := range serviceable {
		partners = append(partners, RatePartnerEntry{
			ID:   p.PartnerID,
			Code: strings.ToUpper(p.PartnerCode),
		})
	}

	// Build metadata
	meta := map[string]string{
		"currency":      "INR",
		"service_type":  "express",
		"flow":          "international",
	}

	// Allow overrides via env
	if cur := os.Getenv("RATE_CURRENCY"); cur != "" {
		meta["currency"] = cur
	}
	if st := os.Getenv("RATE_SERVICE_TYPE"); st != "" {
		meta["service_type"] = st
	}

	return RateRequestPayload{
		SourceLocation: RateLocation{
			PostalCode:  srcPin,
			CountryCode: srcCC,
		},
		DestinationLocation: RateLocation{
			PostalCode:  dstPin,
			CountryCode: dstCC,
		},
		Packages: pkgs,
		Partners: partners,
		Metadata: meta,
	}
}

func getEnvOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func defaultWeight(req *modelsv1.ServiceabilityV2Request) float64 {
	if req != nil && len(req.Packages) > 0 && req.Packages[0].Weight != nil && req.Packages[0].Weight.Value > 0 {
		return req.Packages[0].Weight.Value
	}
	// fallback 1 kg
	return 1.0
}

func defaultLen(req *modelsv1.ServiceabilityV2Request) float64 {
	if req != nil && len(req.Packages) > 0 && req.Packages[0].Dimensions != nil && req.Packages[0].Dimensions.Length > 0 {
		return req.Packages[0].Dimensions.Length
	}
	return 10.0
}
func defaultWid(req *modelsv1.ServiceabilityV2Request) float64 {
	if req != nil && len(req.Packages) > 0 && req.Packages[0].Dimensions != nil && req.Packages[0].Dimensions.Width > 0 {
		return req.Packages[0].Dimensions.Width
	}
	return 10.0
}

func defaultHei(req *modelsv1.ServiceabilityV2Request) float64 {
	if req != nil && len(req.Packages) > 0 && req.Packages[0].Dimensions != nil && req.Packages[0].Dimensions.Height > 0 {
		return req.Packages[0].Dimensions.Height
	}
	return 10.0
}