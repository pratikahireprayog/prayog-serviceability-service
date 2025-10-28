package internationalstrategy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/services/v2/partners/dhl"
	"prayog-serviceability-service/internal/services/v2/partners/factory"
	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// InternationalStrategy orchestrates international flow using HubOps by-pincode → DHL & adapters
type InternationalStrategy struct {
	PartnerFactory factory.PartnerAdapterFactory
	Logger         *logrus.Logger
	ratesClient    *http.Client
}

var hubOpsURL = os.Getenv("SMILE_HUBOPS_BY_SOURCE_PINCODE")

func NewInternationalStrategy(pf factory.PartnerAdapterFactory) *InternationalStrategy {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	return &InternationalStrategy{
		PartnerFactory: pf,
		Logger:         logger,
		ratesClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *InternationalStrategy) Code() string { return "international" }

// Rate API response structures
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

// Rate request structures
type RatePartnerEntry struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type RateLocation struct {
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
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

type RateRequestPayload struct {
	SourceLocation      RateLocation      `json:"source_location"`
	DestinationLocation RateLocation      `json:"destination_location"`
	Packages            []RatePackage     `json:"packages"`
	Partners            []RatePartnerEntry `json:"partners"`
	Metadata            map[string]string `json:"metadata"`
}

func (s *InternationalStrategy) Execute(ctx context.Context, req *modelsv1.ServiceabilityV2Request) (*modelsv1.ServiceabilityV2Response, error) {
	startTime := time.Now()
	if s.Logger == nil {
		s.Logger = logrus.New()
	}
	if s.PartnerFactory == nil {
		return s.buildErrorResponse("Partner factory not available", startTime), nil
	}

	sourcePin := resolveSourcePin(req)
	if sourcePin == "" {
		s.Logger.WithField("component", "international_strategy").Warn("source postal code missing; returning not serviceable")
		return s.buildErrorResponse("Source postal code missing", startTime), nil
	}

	srcCC, dstCC := resolveCountryCodes(req)

	// 1) HubOps lookup (best-effort)
	hubResp, err := s.fetchHubByPincode(ctx, sourcePin)
	if err != nil {
		s.Logger.WithError(err).WithFields(logrus.Fields{
			"component":  "international_strategy",
			"source_pin": sourcePin,
		}).Warn("HubOps by-pincode call failed; proceeding without addresses")
	}

	// Build addresses (if available)
	var addresses []modelsv1.DetailedAddress
	if hubResp != nil && hubResp.NearestInternationalHub != nil {
		addresses = []modelsv1.DetailedAddress{toDetailedAddress(hubResp.NearestInternationalHub)}
	}

	// Resolve source/destination pins used for partner calls
	srcPin := resolveHubSourcePin(hubResp, sourcePin)
	dstPin := resolveDestinationPin(req)

	s.Logger.WithFields(logrus.Fields{
		"component":                "international_strategy",
		"original_source_pin":      sourcePin,
		"hub_source_pin":           srcPin,
		"destination_pin":          dstPin,
		"source_country_code":      srcCC,
		"destination_country_code": dstCC,
	}).Info("Resolved source and destination pincodes for international carriers")

	// Extract city info from hub
	shipperCity := getHubCity(hubResp)
	receiverCity := "Unknown City"

	// Determine partners to call
	partnerCodes := s.getPartnerList(req.Partners)

	// 1) Run serviceability checks for each partner
	partners := make([]modelsv1.PartnerV2Response, 0, len(partnerCodes))
	serviceablePartners := make([]modelsv1.PartnerV2Response, 0)

	for _, partnerCode := range partnerCodes {
		var partnerResp modelsv1.PartnerV2Response
		
		if partnerCode == "dhl" {
			partnerResp = s.callDHL(ctx, req, srcPin, dstPin, srcCC, dstCC, shipperCity, receiverCity)
		} else {
			partnerResp = s.callPartnerViaAdapter(ctx, partnerCode, req, srcPin, dstPin, srcCC, dstCC)
		}

		// Set default values for required fields
		partnerResp = s.ensurePartnerDefaults(partnerResp, partnerCode)

		if partnerResp.PartnerCode != "" {
			partners = append(partners, partnerResp)
			if partnerResp.IsServiceable {
				serviceablePartners = append(serviceablePartners, partnerResp)
			}
		}
	}

	// 2) Get rates for serviceable partners
	ratesIncluded := false
	if len(serviceablePartners) > 0 {
		rates, err := s.getRatesForPartners(ctx, req, srcPin, srcCC, dstPin, dstCC, serviceablePartners)
		if err != nil {
			s.Logger.WithError(err).Warn("Failed to get rates for serviceable partners")
		} else if rates != nil {
			s.mergeRatesIntoPartners(partners, rates)
			ratesIncluded = true
		}
	}

	// Build final response
	return s.buildSuccessResponse(partners, addresses, startTime, len(partnerCodes), len(serviceablePartners), ratesIncluded), nil
}

func (s *InternationalStrategy) getPartnerList(partners []modelsv1.PartnerFilter) []string {
	codes := make([]string, 0, len(partners))
	for _, p := range partners {
		// normalize code to lowercase (or uppercase if needed)
		if p.Code != "" {
			codes = append(codes, strings.ToLower(p.Code))
		}
	}
	return codes
}

func (s *InternationalStrategy) ensurePartnerDefaults(partner modelsv1.PartnerV2Response, code string) modelsv1.PartnerV2Response {
	if partner.PartnerCode == "" {
		partner.PartnerCode = code
	}
	if partner.PartnerName == "" {
		partner.PartnerName = getPartnerName(code)
	}
	if partner.Source == "" {
		partner.Source = "real_time"
	}
	if partner.PartnerID == "" {
		partner.PartnerID = "unknown"
	}
	if partner.Rating == 0 {
		partner.Rating = 0
	}
	if partner.Metadata == nil {
		partner.Metadata = make(map[string]interface{})
	}
	// Ensure metadata has required fields
	if partner.Metadata["destination_country_code"] == nil {
		partner.Metadata["destination_country_code"] = "CN" // Default as per your example
	}
	if partner.Metadata["source_country_code"] == nil {
		partner.Metadata["source_country_code"] = "IN" // Default as per your example
	}
	if partner.Metadata["flow"] == nil {
		partner.Metadata["flow"] = "international"
	}
	return partner
}

func (s *InternationalStrategy) getRatesForPartners(
	ctx context.Context, 
	req *modelsv1.ServiceabilityV2Request, 
	srcPin, srcCC, dstPin, dstCC string, 
	serviceable []modelsv1.PartnerV2Response,
) (*RateQuoteResponse, error) {
	ratesURL := os.Getenv("SUPPLY_RATE_URL")
	payload := s.buildRatesRequestPayload(req, srcPin, srcCC, dstPin, dstCC, serviceable)
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rates payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, ratesURL, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create rates request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Prayog-Serviceability-Service/1.0")

	s.Logger.WithFields(logrus.Fields{
		"component":   "international_strategy",
		"rates_url":   ratesURL,
		"partners":    len(payload.Partners),
		"packages":    len(payload.Packages),
	}).Info("Calling rates API")

	// resp, err := s.ratesClient.Do(httpReq)
	if s.ratesClient == nil {
		s.ratesClient = &http.Client{Timeout: 30 * time.Second}
	}

	resp, err := s.ratesClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("rates API call failed: %w", err)
	}
	defer resp.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("rates API call failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read rates response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		s.Logger.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(bodyBytes),
		}).Warn("Rates API returned non-200 status")
		return nil, fmt.Errorf("rates API returned status %d", resp.StatusCode)
	}

	var rateResp RateQuoteResponse
	if err := json.Unmarshal(bodyBytes, &rateResp); err != nil {
		return nil, fmt.Errorf("failed to parse rates response: %w", err)
	}

	if !rateResp.Success {
		s.Logger.WithField("message", rateResp.Message).Warn("Rates API returned unsuccessful response")
		return nil, fmt.Errorf("rates API returned error: %s", rateResp.Message)
	}

	s.Logger.WithFields(logrus.Fields{
		"component":          "international_strategy",
		"rates_found":        rateResp.Metadata.TotalRatesFound,
		"partners_succeeded": rateResp.Metadata.PartnersSucceeded,
		"response_time_ms":   rateResp.Metadata.ResponseTimeMs,
	}).Info("Successfully retrieved rates from API")

	return &rateResp, nil
}

func (s *InternationalStrategy) buildRatesRequestPayload(
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

type RateQuote struct {
	RateID       string
	Service      string
	DeliveryDays int
	Price        struct {
		Currency    string
		Amount      float64
		Type        string
		ServiceType string
	}
}

func (s *InternationalStrategy) mergeRatesIntoPartners(partners []modelsv1.PartnerV2Response, rates *RateQuoteResponse) {
	// Create a map for quick lookup: partnerCode -> available rates
	ratesMap := make(map[string][]RateQuote)

	// Populate the map from API response
	for _, successResp := range rates.Data.SuccessfulResponses {
		partnerCode := strings.ToLower(successResp.Partner.Code)

		// Convert anonymous struct to our named struct type
		partnerRates := make([]RateQuote, 0, len(successResp.AvailableRates))
		for _, r := range successResp.AvailableRates {
			partnerRates = append(partnerRates, RateQuote{
				RateID:       r.RateID,
				Service:      r.Service,
				DeliveryDays: r.DeliveryDays,
				Price: struct {
					Currency    string
					Amount      float64
					Type        string
					ServiceType string
				}{
					Currency:    r.Price.Currency,
					Amount:      r.Price.Amount,
					Type:        r.Price.Type,
					ServiceType: r.Price.ServiceType,
				},
			})
		}

		ratesMap[partnerCode] = partnerRates
	}

	// Merge rates into partners
	for i := range partners {
		partnerCode := strings.ToLower(partners[i].PartnerCode)
		if availableRates, exists := ratesMap[partnerCode]; exists && len(availableRates) > 0 {
			services := make([]modelsv1.ServiceV2, 0, len(availableRates))

			for _, rate := range availableRates {
				service := modelsv1.ServiceV2{
					ServiceCode: rate.Price.ServiceType,
					ServiceName: rate.Service,
					TATDays:     rate.DeliveryDays,
					IsCOD:       false,
					Pickup:      true,
					Delivery:    true,
					Insurance:   true,
					ProductTypes: map[string]bool{
						"commercial":   true,
						"document":     true,
						"non_document": true,
					},
					DeliveryModes: map[string]bool{
						"express":  strings.Contains(strings.ToLower(rate.Service), "express"),
						"standard": !strings.Contains(strings.ToLower(rate.Service), "express"),
					},
					Rate: &modelsv1.Rate{
						RateID: rate.RateID,
						Price: modelsv1.Price{
							Currency: rate.Price.Currency,
							Amount:   rate.Price.Amount,
							Type:     rate.Price.Type,
						},
					},
				}
				services = append(services, service)
			}

			partners[i].Services = services
			if partners[i].Metadata == nil {
				partners[i].Metadata = make(map[string]interface{})
			}
			partners[i].Metadata["rates_available"] = true
			partners[i].Metadata["rates_count"] = len(availableRates)
		}
	}
}

func (s *InternationalStrategy) callPartnerViaAdapter(
	ctx context.Context,
	partnerCode string,
	req *modelsv1.ServiceabilityV2Request,
	srcPin, dstPin, srcCC, dstCC string,
) modelsv1.PartnerV2Response {
	adapter, found := s.PartnerFactory.GetAdapter(partnerCode)
	
	if !found || adapter == nil {
		s.Logger.WithFields(logrus.Fields{"component": "international_strategy", "partner": partnerCode}).Warn("Adapter missing or nil")
		return modelsv1.PartnerV2Response{
			PartnerCode:   partnerCode,
			PartnerName:   getPartnerName(partnerCode),
			Rating:        0,
			Source:        "real_time",
			IsServiceable: false,
			Services:      []modelsv1.ServiceV2{},
			Error: &modelsv1.PartnerError{
				Code:    modelsv1.ErrorCodeAdapterNotFound,
				Message: "Partner adapter not available",
			},
			Metadata: map[string]interface{}{
				"destination_country_code": dstCC,
				"source_country_code":      srcCC,
				"flow":                     "international",
			},
		}
	}

	serviceReq := s.buildAdapterRequest(req, srcPin, dstPin, srcCC, dstCC)
	partnerInfo := common.PartnerInfo{PartnerCode: partnerCode}

	startTime := time.Now()
	result, err := adapter.CheckServiceability(ctx, serviceReq, partnerInfo)
	responseTimeMs := time.Since(startTime).Milliseconds()

	if err != nil {
		s.Logger.WithError(err).WithFields(logrus.Fields{"component": "international_strategy", "partner": partnerCode}).Warn("Adapter call failed")
		return modelsv1.PartnerV2Response{
			PartnerCode:    partnerCode,
			PartnerName:    getPartnerName(partnerCode),
			Rating:         0,
			Source:         "real_time",
			IsServiceable:  false,
			Services:       []modelsv1.ServiceV2{},
			ResponseTimeMs: responseTimeMs,
			Error: &modelsv1.PartnerError{
				Code:    modelsv1.ErrorCodeServiceabilityFailed,
				Message: "Failed to check serviceability with partner",
				Details: err.Error(),
			},
			Metadata: map[string]interface{}{
				"destination_country_code": dstCC,
				"source_country_code":      srcCC,
				"flow":                     "international",
			},
		}
	}

	// Convert based on partner type
	switch partnerCode {
	case "aramex":
		return s.convertAramexResult(result, srcCC, dstCC, responseTimeMs)
	case "shipcube":
		return s.convertShipCubeResult(result, srcCC, dstCC, responseTimeMs)
	case "fedex":
		return s.convertFedExResult(result, srcCC, dstCC, responseTimeMs)
	default:
		return s.convertGenericResult(result, partnerCode, srcCC, dstCC, responseTimeMs)
	}
}

func (s *InternationalStrategy) buildAdapterRequest(req *modelsv1.ServiceabilityV2Request, srcPin, dstPin, srcCC, dstCC string) *modelsv1.ServiceabilityV2Request {
	return &modelsv1.ServiceabilityV2Request{
		SourcePostalCode:       &srcPin,
		DestinationPostalCode:  &dstPin,
		SourceCountryCode:      &srcCC,
		DestinationCountryCode: &dstCC,
		Packages:               req.Packages,
		PostalCode:             req.PostalCode,
		CountryCode:            req.CountryCode,
	}
}

// Partner-specific conversions (implementation details remain the same as before)
// [Include all the conversion methods: convertAramexResult, convertShipCubeResult, convertFedExResult, convertGenericResult]

// DHL implementation (implementation details remain the same as before)
// [Include callDHL method]

// Response building
func (s *InternationalStrategy) buildSuccessResponse(
	partners []modelsv1.PartnerV2Response, 
	addresses []modelsv1.DetailedAddress, 
	startTime time.Time,
	partnersQueried int,
	serviceablePartners int,
	ratesIncluded bool,
) *modelsv1.ServiceabilityV2Response {
	responseTimeMs := time.Since(startTime).Milliseconds()

	// Calculate statistics
	partnersSucceeded := 0
	partnersFailed := 0
	for _, p := range partners {
		if p.IsServiceable {
			partnersSucceeded++
		} else if p.Error != nil {
			partnersFailed++
		}
	}

	metadata := &modelsv1.V2ResponseMetadata{
		RequestID:                fmt.Sprintf("test-serviceability-%d", time.Now().Unix()),
		ResponseTimeMs:           responseTimeMs,
		PartnersQueried:          partnersQueried,
		PartnersSucceeded:        partnersSucceeded,
		PartnersFailed:           partnersFailed,
		TotalServiceablePartners: serviceablePartners,
		RatesIncluded:            ratesIncluded,
	}

	return &modelsv1.ServiceabilityV2Response{
		Success:   true,
		Message:   "Serviceability check completed successfully.",
		Metadata:  metadata,
		Addresses: addresses,
		Partners:  partners,
		// Timestamp: time.Now().Format(time.RFC3339Nano),
	}
}

func (s *InternationalStrategy) buildErrorResponse(message string, startTime time.Time) *modelsv1.ServiceabilityV2Response {
	responseTimeMs := time.Since(startTime).Milliseconds()

	metadata := &modelsv1.V2ResponseMetadata{
		RequestID:                fmt.Sprintf("test-serviceability-%d", time.Now().Unix()),
		ResponseTimeMs:           responseTimeMs,
		PartnersQueried:          0,
		PartnersSucceeded:        0,
		PartnersFailed:           0,
		TotalServiceablePartners: 0,
		RatesIncluded:            false,
	}

	return &modelsv1.ServiceabilityV2Response{
		Success:  false,
		Message:  message,
		Metadata: metadata,
		Partners: []modelsv1.PartnerV2Response{},
		// Timestamp: time.Now().Format(time.RFC3339Nano),
	}
}

// -----------------------------
// Pin & country resolution
// -----------------------------
func resolveSourcePin(req *modelsv1.ServiceabilityV2Request) string {
	if req == nil {
		return ""
	}
	if req.SourcePostalCode != nil && *req.SourcePostalCode != "" {
		return *req.SourcePostalCode
	}
	if req.PostalCode != nil && *req.PostalCode != "" {
		return *req.PostalCode
	}
	return ""
}

// default dimension/weight helpers (no hardcoded "magic" other than safe fallbacks)
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

// -----------------------------
// Pin & country resolution
// -----------------------------

func resolveDestinationPin(req *modelsv1.ServiceabilityV2Request) string {
	if req == nil {
		return ""
	}
	if req.DestinationPostalCode != nil && *req.DestinationPostalCode != "" {
		return *req.DestinationPostalCode
	}
	if req.PostalCode != nil && *req.PostalCode != "" {
		return *req.PostalCode
	}
	return ""
}

func resolveHubSourcePin(hubResp *modelsv1.HubOpsResponse, fallback string) string {
	if hubResp != nil && hubResp.NearestInternationalHub != nil && hubResp.NearestInternationalHub.Pincode != nil {
		return fmt.Sprintf("%v", *hubResp.NearestInternationalHub.Pincode)
	}
	return fallback
}

func resolveCountryCodes(req *modelsv1.ServiceabilityV2Request) (string, string) {
	src := "IN"
	dst := "US"
	if req == nil {
		return src, dst
	}
	if req.SourceCountryCode != nil && *req.SourceCountryCode != "" {
		src = strings.ToUpper(*req.SourceCountryCode)
	} else if req.CountryCode != nil && *req.CountryCode != "" {
		src = strings.ToUpper(*req.CountryCode)
	}
	if req.DestinationCountryCode != nil && *req.DestinationCountryCode != "" {
		dst = strings.ToUpper(*req.DestinationCountryCode)
	} else if req.CountryCode != nil && *req.CountryCode != "" {
		dst = strings.ToUpper(*req.CountryCode)
	}
	return src, dst
}


func (s *InternationalStrategy) fetchHubByPincode(ctx context.Context, pin string) (*modelsv1.HubOpsResponse, error) {
	if hubOpsURL == "" {
		return nil, errors.New("hubops url not configured")
	}
	body, _ := json.Marshal(modelsv1.HubOpsRequest{PostalCode: pin})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hubOpsURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if cookie := os.Getenv("HUBOPS_COOKIE"); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hubops status %d: %s", resp.StatusCode, string(b))
	}
	var parsed modelsv1.HubOpsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}

func toDetailedAddress(h *modelsv1.HubInfoData) modelsv1.DetailedAddress {
	addr := modelsv1.DetailedAddress{Type: "INTERNATIONAL_HUB_ADDRESS", AddressName: "WAREHOUSE"}
	if h == nil {
		return addr
	}
	if h.Pincode != nil {
		addr.Zip = fmt.Sprintf("%v", *h.Pincode)
	}
	if h.PremiseName != nil {
		addr.Name = *h.PremiseName
	}
	if s := anyToString(h.OfficialNumber); s != "" {
		addr.Phone = s
	} else {
		addr.Phone = anyToString(h.PersonalNumber)
	}
	if h.OfficialEmailId != nil && *h.OfficialEmailId != "" {
		addr.Email = *h.OfficialEmailId
	} else if h.PersonalEmailId != nil {
		addr.Email = *h.PersonalEmailId
	}
	// Build street
	parts := []string{}
	if h.AddressLine1 != nil && *h.AddressLine1 != "" {
		parts = append(parts, *h.AddressLine1)
	}
	if h.AddressLine2 != nil && *h.AddressLine2 != "" {
		parts = append(parts, *h.AddressLine2)
	}
	if h.Address != nil && *h.Address != "" {
		parts = append(parts, *h.Address)
	}
	if len(parts) > 0 {
		addr.Street = strings.Join(parts, ", ")
	}
	if h.City != nil {
		addr.City = *h.City
	}
	if h.State != nil {
		addr.State = *h.State
	}
	if h.Latitude != nil {
		if v, ok := toFloat(*h.Latitude); ok {
			addr.Latitude = &v
		}
	}
	if h.Longitude != nil {
		if v, ok := toFloat(*h.Longitude); ok {
			addr.Longitude = &v
		}
	}
	return addr
}

func getPartnerName(code string) string {
	names := map[string]string{
		"dhl":      "DHL Express",
		"aramex":   "Aramex",
		"fedex":    "FedEx",
		"shipcube": "ShipCube",
	}
	if n, ok := names[strings.ToLower(code)]; ok {
		return n
	}
	return strings.Title(code)
}

func generatePartnerID() string {
	return uuid.New().String()
}

func getHubCity(h *modelsv1.HubOpsResponse) string {
	if h == nil || h.NearestInternationalHub == nil || h.NearestInternationalHub.City == nil {
		return "Unknown City"
	}
	if *h.NearestInternationalHub.City == "" {
		return "Unknown City"
	}
	return *h.NearestInternationalHub.City
}

func toFloat(s string) (float64, bool) {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return 0, false
	}
	return f, true
}

func anyToString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		return fmt.Sprintf("%.0f", t)
	case int:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	case *string:
		if t == nil {
			return ""
		}
		return *t
	case *int:
		if t == nil {
			return ""
		}
		return strconv.Itoa(*t)
	case *int64:
		if t == nil {
			return ""
		}
		return strconv.FormatInt(*t, 10)
	case *float64:
		if t == nil {
			return ""
		}
		return strconv.FormatFloat(*t, 'f', -1, 64)
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}


func (s *InternationalStrategy) callDHL(ctx context.Context, req *modelsv1.ServiceabilityV2Request, srcPin, dstPin, srcCC, dstCC, shipperCity, receiverCity string) modelsv1.PartnerV2Response {
	baseURL := os.Getenv("DHL_BASE_URL")
	if baseURL == "" {
		baseURL = "https://express.api.dhl.com/mydhlapi"
	}

	dhlReq := dhl.RatesRequest{
		CustomerDetails: dhl.CustomerDetails{
			ShipperDetails:  dhl.ShipperDetails{PostalCode: srcPin, CityName: shipperCity, CountryCode: srcCC},
			ReceiverDetails: dhl.ReceiverDetails{PostalCode: dstPin, CityName: receiverCity, CountryCode: dstCC},
		},
		Accounts:                    []dhl.Account{{TypeCode: "shipper", Number: "533748932"}},
		ProductsAndServices:         []dhl.ProductAndService{{ProductCode: "P", LocalProductCode: "P"}},
		PayerCountryCode:            "IN",
		PlannedShippingDateAndTime:  nextBusinessDayOnePMIST(),
		UnitOfMeasurement:           "metric",
		IsCustomsDeclarable:        true,
		EstimatedDeliveryDate:       dhl.EstimatedDeliveryDate{IsRequested: true, TypeCode: "QDDC"},
		ReturnStandardProductsOnly:  true,
		Packages: []dhl.Package{{
			Weight: defaultWeight(req),
			Dimensions: dhl.Dimensions{
				Length: defaultLen(req),
				Width:  defaultWid(req),
				Height: defaultHei(req),
			},
		}},
	}

	url := fmt.Sprintf("%s/rates?strictValidation=false", baseURL)
	body, _ := json.Marshal(dhlReq)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Message-Reference", "d0e7832e-5c98-11ea-bc55-0242ac13")
	httpReq.Header.Set("Message-Reference-Date", "Wed, 21 Oct 2015 07:28:00 GMT")
	httpReq.Header.Set("X-Version", "2.12.0")

	if basicAuth := os.Getenv("DHL_BASIC_AUTH"); basicAuth != "" {
		httpReq.Header.Set("Authorization", "Basic "+basicAuth)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	startTime := time.Now()
	dhlResp, doErr := client.Do(httpReq)
	responseTimeMs := time.Since(startTime).Milliseconds()

	if doErr != nil {
		s.Logger.WithError(doErr).WithFields(logrus.Fields{"component": "international_strategy", "partner": "dhl"}).Warn("DHL HTTP call failed")
		return modelsv1.PartnerV2Response{
			PartnerID:     "93a2d552-dd7a-4786-aa11-cf44e7b327ab",
			PartnerCode:   "dhl",
			PartnerName:   "DHL Express",
			Rating:        0,
			Source:        "real_time",
			IsServiceable: false,
			ResponseTimeMs: responseTimeMs,
			Error: &modelsv1.PartnerError{
				Code:    "SERVICEABILITY_CHECK_FAILED",
				Message: "DHL API call failed",
				Details: doErr.Error(),
			},
		}
	}
	defer dhlResp.Body.Close()

	var rates dhl.RatesResponse
	if dhlResp.StatusCode == http.StatusOK {
		if decErr := json.NewDecoder(dhlResp.Body).Decode(&rates); decErr == nil && len(rates.Products) > 0 {
			prod := rates.Products[0]

			capabilities := map[string]interface{}{
				"estimated_delivery_date_and_time": prod.DeliveryCapabilities.EstimatedDeliveryDateAndTime,
				"total_transit_days":               prod.DeliveryCapabilities.TotalTransitDays,
				"pickup_available":                 true,
				"delivery_available":               true,
			}

			services := []modelsv1.ServiceV2{
				{
					ServiceCode: "EXPRESS",
					ServiceName: "DHL Express",
					TATDays:     prod.DeliveryCapabilities.TotalTransitDays,
					IsCOD:       false,
					Pickup:      true,
					Delivery:    true,
					Insurance:   true,
					ProductTypes: map[string]bool{
						"commercial":  true,
						"document":    true,
						"non_document": true,
					},
					DeliveryModes: map[string]bool{
						"express":  true,
						"standard": false,
					},
					// DHL rates will be populated from rates API
				},
			}

			return modelsv1.PartnerV2Response{
				PartnerID:      "93a2d552-dd7a-4786-aa11-cf44e7b327ab",
				PartnerCode:    "dhl",
				PartnerName:    "DHL Express",
				Rating:         0,
				Source:         "real_time",
				IsServiceable:  true,
				Capabilities:   capabilities,
				Services:       services,
				ResponseTimeMs: responseTimeMs,
				Metadata: map[string]interface{}{
					"destination_country_code": dstCC,
					"source_country_code":      srcCC,
					"flow":                     "international",
					"dhl_response":             "success",
				},
			}
		}
	}

	// If we reach here, DHL call was not successful
	bodyBytes, _ := io.ReadAll(dhlResp.Body)
	s.Logger.WithFields(logrus.Fields{
		"component": "international_strategy",
		"partner":   "dhl",
		"status":    dhlResp.StatusCode,
	}).Warn("DHL API call unsuccessful")

	return modelsv1.PartnerV2Response{
		PartnerID:     "93a2d552-dd7a-4786-aa11-cf44e7b327ab",
		PartnerCode:   "dhl",
		PartnerName:   "DHL Express",
		Rating:        0,
		Source:        "real_time",
		IsServiceable: false,
		ResponseTimeMs: responseTimeMs,
		Error: &modelsv1.PartnerError{
			Code:    "SERVICEABILITY_CHECK_FAILED",
			Message: "DHL serviceability check failed",
			Details: fmt.Sprintf("Status: %d, Response: %s", dhlResp.StatusCode, string(bodyBytes)),
		},
	}
}

func nextBusinessDayOnePMIST() string {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("GMT+05:30", 5*60*60+30*60)
	}
	now := time.Now().In(loc)
	next := now.Add(24 * time.Hour)
	for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
		next = next.Add(24 * time.Hour)
	}
	t := time.Date(next.Year(), next.Month(), next.Day(), 13, 0, 0, 0, loc)
	return t.Format("2006-01-02T15:04:05") + "GMT+05:30"
}

// -----------------------------
// Converters (Aramex/ShipCube/FedEx) - unchanged logic but no hardcoded amounts
// -----------------------------
func (s *InternationalStrategy) convertAramexResult(res *common.PartnerServiceabilityResult, srcCC, dstCC string, responseTimeMs int64) modelsv1.PartnerV2Response {
	if res == nil {
		return modelsv1.PartnerV2Response{
			PartnerCode:    "aramex",
			IsServiceable:  false,
			ResponseTimeMs: responseTimeMs,
		}
	}

	isServiceable := false
	if res.Metadata != nil {
		if serviceable, ok := res.Metadata["is_serviceable"].(bool); ok {
			isServiceable = serviceable
		}
	}

	partnerID := "5b0795d4-ef0b-40ae-8ed4-c2cabbbecc4b"
	if res.PartnerID != nil {
		partnerID = res.PartnerID.String()
	}

	// Base services (will be enriched with rates later)
	var services []modelsv1.ServiceV2
	if isServiceable {
		services = []modelsv1.ServiceV2{
			{
				ServiceCode: "INTL_EXPRESS",
				ServiceName: "Aramex International Express",
				TATDays:     3,
				IsCOD:       false,
				Pickup:      true,
				Delivery:    true,
				Insurance:   true,
				ProductTypes: map[string]bool{
					"commercial":  true,
					"document":    true,
					"non_document": true,
				},
				DeliveryModes: map[string]bool{
					"express":  true,
					"standard": false,
				},
			},
		}
	}

	return modelsv1.PartnerV2Response{
		PartnerID:      partnerID,
		PartnerCode:    "aramex",
		PartnerName:    "Aramex",
		Rating:         0,
		Source:         "real_time",
		IsServiceable:  isServiceable,
		Services:       services,
		ResponseTimeMs: responseTimeMs,
		Metadata: map[string]interface{}{
			"destination_country_code": dstCC,
			"source_country_code":      srcCC,
			"flow":                     "international",
			"aramex_metadata":          res.Metadata,
		},
	}
}

func (s *InternationalStrategy) convertShipCubeResult(result *common.PartnerServiceabilityResult, srcCC, dstCC string, responseTimeMs int64) modelsv1.PartnerV2Response {
	if result == nil {
		return modelsv1.PartnerV2Response{
			PartnerCode:    "shipcube",
			IsServiceable:  false,
			ResponseTimeMs: responseTimeMs,
		}
	}

	isServiceable := len(result.Services) > 0
	services := make([]modelsv1.ServiceV2, 0)

	if isServiceable {
		for _, svc := range result.Services {
			service := modelsv1.ServiceV2{
				ServiceCode: svc.ServiceCode,
				ServiceName: svc.ServiceName,
				TATDays:     svc.TATDays,
				Pickup:      svc.Pickup,
				Delivery:    svc.Delivery,
				Insurance:   svc.Insurance,
				ProductTypes: map[string]bool{
					"document":     true,
					"non_document": true,
					"commercial":   true,
				},
				DeliveryModes: map[string]bool{
					"express":  true,
					"standard": true,
				},
			}
			services = append(services, service)
		}
	}

	return modelsv1.PartnerV2Response{
		PartnerID:      "044eef78-97c0-43b2-bb99-5cae2833a63d",
		PartnerCode:    "shipcube",
		PartnerName:    "ShipCube",
		Rating:         0,
		Source:         "real_time",
		IsServiceable:  isServiceable,
		Services:       services,
		ResponseTimeMs: responseTimeMs,
		Metadata:       result.Metadata,
	}
}

func (s *InternationalStrategy) convertFedExResult(result *common.PartnerServiceabilityResult, srcCC, dstCC string, responseTimeMs int64) modelsv1.PartnerV2Response {
	if result == nil {
		return modelsv1.PartnerV2Response{
			PartnerID:     "83c5a4ac-b297-466a-9b14-9f2602103737",
			PartnerCode:   "fedex",
			PartnerName:   "FedEx",
			Rating:        0,
			Source:        "real_time",
			IsServiceable: false,
			ResponseTimeMs: responseTimeMs,
			Metadata: map[string]interface{}{
				"destination_country_code": dstCC,
				"source_country_code":      srcCC,
				"flow":                     "international",
			},
		}
	}

	isServiceable := len(result.Services) > 0
	var services []modelsv1.ServiceV2

	if isServiceable {
		for _, svc := range result.Services {
			service := modelsv1.ServiceV2{
				ServiceCode: svc.ServiceCode,
				ServiceName: svc.ServiceName,
				TATDays:     svc.TATDays,
				IsCOD:       false,
				Pickup:      true,
				Delivery:    true,
				Insurance:   true,
				ProductTypes: map[string]bool{
					"document":     true,
					"non_document": true,
				},
				DeliveryModes: map[string]bool{
					"express":  true,
					"standard": true,
				},
			}
			services = append(services, service)
		}
	}

	return modelsv1.PartnerV2Response{
		PartnerID:      "83c5a4ac-b297-466a-9b14-9f2602103737",
		PartnerCode:    "fedex",
		PartnerName:    "FedEx",
		Rating:         0,
		Source:         "real_time",
		IsServiceable:  isServiceable,
		Services:       services,
		ResponseTimeMs: responseTimeMs,
		Metadata:       result.Metadata,
	}
}

func (s *InternationalStrategy) convertGenericResult(res *common.PartnerServiceabilityResult, code, srcCC, dstCC string, responseTimeMs int64) modelsv1.PartnerV2Response {
	if res == nil {
		return modelsv1.PartnerV2Response{
			PartnerCode:    code,
			IsServiceable:  false,
			ResponseTimeMs: responseTimeMs,
		}
	}

	isServiceable := len(res.Services) > 0
	if res.Metadata != nil {
		if serviceable, ok := res.Metadata["is_serviceable"].(bool); ok {
			isServiceable = serviceable
		}
	}

	partnerID := generatePartnerID()
	if res.PartnerID != nil {
		partnerID = res.PartnerID.String()
	}

	services := make([]modelsv1.ServiceV2, len(res.Services))
	for i, svc := range res.Services {
		services[i] = modelsv1.ServiceV2{
			ServiceCode: svc.ServiceCode,
			ServiceName: svc.ServiceName,
			TATDays:     svc.TATDays,
			Pickup:      svc.Pickup,
			Delivery:    svc.Delivery,
			Insurance:   svc.Insurance,
			ProductTypes: svc.ProductTypes,
			DeliveryModes: svc.DeliveryModes,
			Rate:        svc.Rate,
		}
	}

	return modelsv1.PartnerV2Response{
		PartnerID:      partnerID,
		PartnerCode:    code,
		PartnerName:    getPartnerName(code),
		Rating:         0,
		Source:         "real_time",
		IsServiceable:  isServiceable,
		Services:       services,
		Capabilities:   res.Capabilities,
		ResponseTimeMs: responseTimeMs,
		Metadata:       res.Metadata,
	}
}

