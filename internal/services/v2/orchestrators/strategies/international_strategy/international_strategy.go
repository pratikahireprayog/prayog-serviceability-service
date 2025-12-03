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

	services "prayog-serviceability-service/internal/services/v1/data"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/services/v2/partners/factory"
	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// InternationalStrategy orchestrates international flow using HubOps by-pincode → DHL & adapters
type InternationalStrategy struct {
	PartnerFactory     factory.PartnerAdapterFactory
	Logger             *logrus.Logger
	ratesClient        *http.Client
	geolocationService services.GeolocationService
}

func NewInternationalStrategy(pf factory.PartnerAdapterFactory, geolocationService services.GeolocationService) *InternationalStrategy {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	// Log geolocation service availability
	if geolocationService == nil {
		logger.Warn("Geolocation service is nil when creating InternationalStrategy")
	} else {
		logger.Debug("Geolocation service is available for InternationalStrategy")
	}
	
	return &InternationalStrategy{
		PartnerFactory:     pf,
		Logger:             logger,
		geolocationService: geolocationService,
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
				Description  string `json:"description,omitempty"`
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

	srcCC, dstCC := s.resolveCountryCodes(ctx, req, sourcePin)	// 1) HubOps lookup (best-effort)
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

	// Determine partners to call
	// If specific partners are requested, use only those; otherwise use default international partners
	partnerCodes := []string{"dhl", "aramex", "fedex", "shipcube", "india_post_international", "naqel"}
	if len(req.Partners) > 0 {
		partnerCodes = make([]string, 0, len(req.Partners))
		for _, p := range req.Partners {
			// Normalize partner codes (handle variants like "india_post_international" -> "indiapost")
			partnerCode := normalizeInternationalPartnerCode(strings.ToLower(p.Code))
			if partnerCode != "" {
				partnerCodes = append(partnerCodes, partnerCode)
			}
		}
		s.Logger.WithFields(logrus.Fields{
			"component":         "international_strategy",
			"requested_partners": partnerCodes,
		}).Info("Using requested partners for international strategy")
	}

	// 1) Run serviceability checks for each partner (ALL via adapters now)
	serviceablePartners := make([]modelsv1.PartnerV2Response, 0)

	for _, partnerCode := range partnerCodes {
		// Call all partners via adapter (including DHL)
		partnerResp := s.callPartnerViaAdapter(ctx, partnerCode, req, srcPin, dstPin, srcCC, dstCC)

		// Set default values for required fields
		partnerResp = s.ensurePartnerDefaults(partnerResp, partnerCode)

		// Only include serviceable partners in response
		if partnerResp.PartnerCode != "" && partnerResp.IsServiceable {
			serviceablePartners = append(serviceablePartners, partnerResp)
		} else {
			// Log non-serviceable partners for debugging
			s.Logger.WithFields(logrus.Fields{
				"component":     "international_strategy",
				"partner":       partnerCode,
				"is_serviceable": partnerResp.IsServiceable,
				"error":         partnerResp.Error,
			}).Debug("Partner not serviceable, excluding from response")
		}
	}

	// 2) Get rates for serviceable partners
	ratesIncluded := false
	// if len(serviceablePartners) > 0 {
	// 	// rateClient := supplyrates.NewRateClient(s.Logger)
	// 	// rates, err := rateClient.GetRatesForPartners(ctx, req, srcPin, srcCC, dstPin, dstCC, serviceablePartners)
	// 	if err != nil {
	// 		s.Logger.WithError(err).Warn("Failed to get rates for serviceable partners")
	// 	} else if rates != nil {
	// 		s.mergeRatesIntoPartners(serviceablePartners, rates)
	// 		ratesIncluded = true
	// 	}
	// }

	// Calculate response metrics
	responseTimeMs := time.Since(startTime).Milliseconds()
	partnersQueried := len(partnerCodes)
	partnersSucceeded := len(serviceablePartners)
	partnersFailed := partnersQueried - partnersSucceeded

	// Build metadata
	metadata := &modelsv1.V2ResponseMetadata{
		RequestID:                fmt.Sprintf("intl-serviceability-%d", time.Now().Unix()),
		ResponseTimeMs:           responseTimeMs,
		PartnersQueried:          partnersQueried,
		PartnersSucceeded:        partnersSucceeded,
		PartnersFailed:           partnersFailed,
		TotalServiceablePartners: len(serviceablePartners),
		RatesIncluded:            ratesIncluded,
	}

	// Build message
	message := "Serviceability check completed successfully."
	if len(serviceablePartners) == 0 {
		message = "No serviceable partners found for this route."
	}

	// Build response
	serviceabilityResp := &modelsv1.ServiceabilityV2Response{
		Success:  len(serviceablePartners) > 0,
		Message:  message,
		Metadata: metadata,
		Partners: serviceablePartners,
	}
	
	// Add addresses if available
	if len(addresses) > 0 {
		serviceabilityResp.Addresses = addresses
	}
	
	return serviceabilityResp, nil
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

type RateQuote struct {
	RateID       string
	Service      string
	DeliveryDays int
	Description  string
	Price        struct {
		Currency    string
		Amount      float64
		Type        string
		ServiceType string
	}
}

func (s *InternationalStrategy) callPartnerViaAdapter(
	ctx context.Context,
	partnerCode string,
	req *modelsv1.ServiceabilityV2Request,
	srcPin, dstPin, srcCC, dstCC string,
) modelsv1.PartnerV2Response {
	s.Logger.WithFields(logrus.Fields{
		"component": "international_strategy",
		"partner":   partnerCode,
		"src_cc":    srcCC,
		"dst_cc":    dstCC,
		"src_pin":   srcPin,
		"dst_pin":   dstPin,
		"action":    "getting_adapter",
	}).Info("Getting adapter for partner")
	
	adapter, found := s.PartnerFactory.GetAdapter(partnerCode)
	
	if !found || adapter == nil {
		s.Logger.WithFields(logrus.Fields{
			"component": "international_strategy", 
			"partner":   partnerCode,
			"found":     found,
			"adapter_nil": adapter == nil,
		}).Warn("Adapter missing or nil")
		return modelsv1.PartnerV2Response{
		PartnerCode:   partnerCode,
		PartnerName:   getPartnerName(partnerCode),
		Rating:        0,
		Source:        "real_time",
		IsServiceable: false,
		PartnerServices: []modelsv1.ServiceV2{},
			// Error: &modelsv1.PartnerError{
			// 	Code:    modelsv1.ErrorCodeAdapterNotFound,
			// 	Message: "Partner adapter not available",
			// },
			Metadata: map[string]interface{}{
				"destination_country_code": dstCC,
				"source_country_code":      srcCC,
				"flow":                     "international",
			},
		}
	}

	serviceReq := s.buildAdapterRequest(req, srcPin, dstPin, srcCC, dstCC)
	partnerInfo := common.PartnerInfo{PartnerCode: partnerCode}

	s.Logger.WithFields(logrus.Fields{
		"component": "international_strategy",
		"partner":   partnerCode,
		"action":    "calling_adapter",
		"src_cc":    srcCC,
		"dst_cc":    dstCC,
		"src_pin":   srcPin,
		"dst_pin":   dstPin,
	}).Info("Calling adapter CheckServiceability")

	startTime := time.Now()
	result, err := adapter.CheckServiceability(ctx, serviceReq, partnerInfo)
	responseTimeMs := time.Since(startTime).Milliseconds()
	
	s.Logger.WithFields(logrus.Fields{
		"component":      "international_strategy",
		"partner":       partnerCode,
		"action":        "adapter_call_completed",
		"has_error":     err != nil,
		"has_result":    result != nil,
		"services_count": func() int {
			if result != nil {
				return len(result.Services)
			}
			return 0
		}(),
		"response_time_ms": responseTimeMs,
	}).Info("Adapter CheckServiceability completed")

	if err != nil {
		s.Logger.WithError(err).WithFields(logrus.Fields{
			"component": "international_strategy", 
			"partner":   partnerCode,
		}).Warn("Adapter call failed")
		return modelsv1.PartnerV2Response{
			PartnerCode:    partnerCode,
			PartnerName:    getPartnerName(partnerCode),
			Rating:         0,
			Source:         "real_time",
			IsServiceable:  false,
			PartnerServices: []modelsv1.ServiceV2{},
			ResponseTimeMs: responseTimeMs,
			// Error: &modelsv1.PartnerError{
			// 	Code:    modelsv1.ErrorCodeServiceabilityFailed,
			// 	Message: "Failed to check serviceability with partner",
			// 	Details: err.Error(),
			// },
			Metadata: map[string]interface{}{
				"destination_country_code": dstCC,
				"source_country_code":      srcCC,
				"flow":                     "international",
			},
		}
	}

	// Convert based on partner type
	switch partnerCode {
	case "dhl":
		return s.convertDHLResult(result, srcCC, dstCC, responseTimeMs)
	case "aramex":
		return s.convertAramexResult(result, srcCC, dstCC, responseTimeMs)
	case "shipcube":
		return s.convertShipCubeResult(result, srcCC, dstCC, responseTimeMs)
	case "fedex":
		return s.convertFedExResult(result, srcCC, dstCC, responseTimeMs)
	case "india_post_international": 
		return s.convertIndiaPostInternationalResult(result, srcCC, dstCC, responseTimeMs)
	case "naqel":
		return s.convertNaqelResult(result, srcCC, dstCC, responseTimeMs)
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

// resolveCountryCodes resolves country codes from request or geolocation service
func (s *InternationalStrategy) resolveCountryCodes(ctx context.Context, req *modelsv1.ServiceabilityV2Request, sourcePin string) (string, string) {
	src := ""
	dst := ""
	
	if req == nil {
		return src, dst
	}
	
	s.Logger.WithFields(logrus.Fields{
		"component":          "international_strategy",
		"geolocationService": s.geolocationService != nil,
	}).Debug("Resolving country codes")
	
	// Resolve source country code
	if req.SourceCountryCode != nil && *req.SourceCountryCode != "" {
		src = strings.ToUpper(*req.SourceCountryCode)
	} else if sourcePin != "" && s.geolocationService != nil {
		// Fetch from geolocation service using source postal code
		srcCCPtr, err := s.geolocationService.GetCountryCodeByPostalCode(ctx, sourcePin)
		if err != nil {
			s.Logger.WithError(err).WithFields(logrus.Fields{
				"component":  "international_strategy",
				"source_pin": sourcePin,
			}).Warn("Failed to get source country code by postal code")
		} else if srcCCPtr != nil {
			src = strings.ToUpper(*srcCCPtr)
			s.Logger.WithFields(logrus.Fields{
				"component":  "international_strategy",
				"source_pin": sourcePin,
				"country_code": src,
			}).Debug("Resolved source country code from geolocation service")
		}
	} else if sourcePin != "" && s.geolocationService == nil {
		s.Logger.WithFields(logrus.Fields{
			"component":  "international_strategy",
			"source_pin": sourcePin,
		}).Warn("Geolocation service is nil, cannot resolve source country code")
	}

	// Resolve destination country code
	dstPin := resolveDestinationPin(req)
	if req.DestinationCountryCode != nil && *req.DestinationCountryCode != "" {
		dst = strings.ToUpper(*req.DestinationCountryCode)
	} else if dstPin != "" && s.geolocationService != nil {
		// Fetch from geolocation service using destination postal code
		dstCCPtr, err := s.geolocationService.GetCountryCodeByPostalCode(ctx, dstPin)
		if err != nil {
			s.Logger.WithError(err).WithFields(logrus.Fields{
				"component":      "international_strategy",
				"destination_pin": dstPin,
			}).Warn("Failed to get destination country code by postal code")
		} else if dstCCPtr != nil {
			dst = strings.ToUpper(*dstCCPtr)
			s.Logger.WithFields(logrus.Fields{
				"component":      "international_strategy",
				"destination_pin": dstPin,
				"country_code":   dst,
			}).Debug("Resolved destination country code from geolocation service")
		}
	} else if dstPin != "" && s.geolocationService == nil {
		s.Logger.WithFields(logrus.Fields{
			"component":      "international_strategy",
			"destination_pin": dstPin,
		}).Warn("Geolocation service is nil, cannot resolve destination country code")
	}
	
	s.Logger.WithFields(logrus.Fields{
		"component":          "international_strategy",
		"source_country":     src,
		"destination_country": dst,
	}).Info("Resolved country codes")
	
	return src, dst
}


func (s *InternationalStrategy) fetchHubByPincode(ctx context.Context, pin string) (*modelsv1.HubOpsResponse, error) {
	// Read hubOpsURL at runtime to ensure .env is loaded
	hubOpsURL := os.Getenv("SMILE_HUBOPS_BY_SOURCE_PINCODE")
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
		"dhl":       "DHL Express",
		"aramex":    "Aramex",
		"fedex":     "FedEx",
		"shipcube":  "ShipCube",
		"india_post_international": "India Post International",
		"naqel":     "Naqel",
	}
	if n, ok := names[strings.ToLower(code)]; ok {
		return n
	}
	return strings.Title(code)
}

// normalizeInternationalPartnerCode normalizes partner codes to standard international partner codes
func normalizeInternationalPartnerCode(code string) string {
	code = strings.ToLower(code)
	
	// Direct mapping for known codes
	normalizedMap := map[string]string{
		"dhl":                      "dhl",
		"fedex":                    "fedex",
		"aramex":                   "aramex",
		"shipcube":                 "shipcube",
		"india_post_international": "india_post_international",
		"naqel":                    "naqel",
	}
	
	if normalized, exists := normalizedMap[code]; exists {
		return normalized
	}
	
	// Fallback: check if code contains any international partner name
	for key, value := range normalizedMap {
		if strings.Contains(code, key) {
			return value
		}
	}
	
	// Return as-is if no match found (adapter factory will handle it)
	return code
}

func generatePartnerID() string {
	return uuid.New().String()
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


// DEPRECATED: callDHL is no longer used. DHL now uses the adapter pattern like all other partners.
// This function is kept for reference only and will be removed in future versions.
// Use callPartnerViaAdapter("dhl", ...) instead.

// -----------------------------
// Converters (Aramex/ShipCube/FedEx) - unchanged logic but no hardcoded amounts
// -----------------------------
// convertDHLResult converts DHL adapter result to partner response
func (s *InternationalStrategy) convertDHLResult(result *common.PartnerServiceabilityResult, srcCC, dstCC string, responseTimeMs int64) modelsv1.PartnerV2Response {
	if result == nil {
		return modelsv1.PartnerV2Response{
			PartnerID:      "93a2d552-dd7a-4786-aa11-cf44e7b327ab",
			PartnerCode:    "dhl",
			PartnerName:    "DHL Express",
			Rating:         0,
			Source:         "real_time",
			IsServiceable:  false,
			ResponseTimeMs: responseTimeMs,
			Metadata: map[string]interface{}{
				"destination_country_code": dstCC,
				"source_country_code":      srcCC,
				"flow":                     "international",
			},
		}
	}

	isServiceable := len(result.Services) > 0
	partnerID := "93a2d552-dd7a-4786-aa11-cf44e7b327ab"
	if result.PartnerID != nil {
		partnerID = result.PartnerID.String()
	}

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
					"commercial":   true,
					"document":     true,
					"non_document": true,
				},
				DeliveryModes: map[string]bool{
					"express":  true,
					"standard": false,
				},
			}
			services = append(services, service)
		}
	}

	return modelsv1.PartnerV2Response{
		PartnerID:        partnerID,
		PartnerCode:      "dhl",
		PartnerName:      "DHL Express",
		Rating:           0,
		Source:           "real_time",
		IsServiceable:    isServiceable,
		PartnerServices:  services,
		Capabilities:     result.Capabilities,
		ResponseTimeMs:   responseTimeMs,
		Metadata: map[string]interface{}{
			"destination_country_code": dstCC,
			"source_country_code":      srcCC,
			"flow":                     "international",
			"dhl_response":             "success",
		},
	}
}

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
		PartnerID:       partnerID,
		PartnerCode:     "aramex",
		PartnerName:     "Aramex",
		Rating:          0,
		Source:          "real_time",
		IsServiceable:   isServiceable,
		PartnerServices: services,
		ResponseTimeMs:  responseTimeMs,
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
		PartnerID:       "044eef78-97c0-43b2-bb99-5cae2833a63d",
		PartnerCode:     "shipcube",
		PartnerName:     "ShipCube",
		Rating:          0,
		Source:          "real_time",
		IsServiceable:   isServiceable,
		PartnerServices: services,
		ResponseTimeMs:  responseTimeMs,
		Metadata:        result.Metadata,
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
		PartnerID:       "83c5a4ac-b297-466a-9b14-9f2602103737",
		PartnerCode:     "fedex",
		PartnerName:     "FedEx",
		Rating:          0,
		Source:          "real_time",
		IsServiceable:   isServiceable,
		PartnerServices: services,
		ResponseTimeMs:  responseTimeMs,
		Metadata:        result.Metadata,
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
		PartnerID:       partnerID,
		PartnerCode:     code,
		PartnerName:     getPartnerName(code),
		Rating:          0,
		Source:          "real_time",
		IsServiceable:   isServiceable,
		PartnerServices: services,
		Capabilities:    res.Capabilities,
		ResponseTimeMs:  responseTimeMs,
		Metadata:        res.Metadata,
	}
}

// callIndiaPostInternationalViaAdapter - uses India Post International adapter for serviceability check
func (s *InternationalStrategy) callIndiaPostInternationalViaAdapter(
	ctx context.Context,
	req *modelsv1.ServiceabilityV2Request,
	srcPin, dstPin, srcCC, dstCC, srcCity, dstCity string,
) modelsv1.PartnerV2Response {

	adapter, found := s.PartnerFactory.GetAdapter("india_post_international")
	if !found || adapter == nil {
		s.Logger.WithFields(logrus.Fields{
			"component": "international_strategy",
			"partner":   "india_post_international",
		}).Warn("India Post International adapter not available")
		return modelsv1.PartnerV2Response{}
	}

	serviceReq := &modelsv1.ServiceabilityV2Request{
		SourcePostalCode:       &srcPin,
		DestinationPostalCode:  &dstPin,
		SourceCountryCode:      &srcCC,
		DestinationCountryCode: &dstCC,
		Packages:               req.Packages,
	}

	partnerInfo := common.PartnerInfo{
		PartnerCode: "india_post_international",
	}

	s.Logger.WithFields(logrus.Fields{
		"component":           "international_strategy",
		"partner":             "india_post_international",
		"action":              "india_post_serviceability_check",
		"source_pincode":      srcPin,
		"destination_pincode": dstPin,
		"source_country":      srcCC,
		"destination_country": dstCC,
	}).Info("Calling India Post International adapter for serviceability")

	result, err := adapter.CheckServiceability(ctx, serviceReq, partnerInfo)
	if err != nil {
		s.Logger.WithError(err).WithFields(logrus.Fields{
			"component": "international_strategy",
			"partner":   "india_post_international",
		}).Warn("India Post International adapter call failed")
		return modelsv1.PartnerV2Response{}
	}

	s.Logger.WithFields(logrus.Fields{
		"resultReceived": result != nil,
		"hasError":       err != nil,
	}).Info("India Post International adapter CheckServiceability completed")

	resp := s.convertIndiaPostInternationalResult(result, srcCC, dstCC, 0)
	if resp.PartnerCode == "" {
		s.Logger.Warn("India Post International response empty after conversion")
		return modelsv1.PartnerV2Response{}
	}

	s.Logger.WithFields(logrus.Fields{
		"component": "international_strategy",
		"partner":   "india_post_international",
		"services":  len(resp.Services),
	}).Info("India Post International partner response prepared successfully")

	return resp
}

// convertIndiaPostInternationalResult - converts India Post International adapter result to partner response
func (s *InternationalStrategy) convertIndiaPostInternationalResult(
	result *common.PartnerServiceabilityResult,
	srcCC, dstCC string, responseTimeMs int64,
) modelsv1.PartnerV2Response {

	if result == nil {
		return modelsv1.PartnerV2Response{}
	}

	isServiceable := len(result.Services) > 0

	partnerID := ""
	if result.PartnerID != nil {
		partnerID = result.PartnerID.String()
	}

	// If serviceable, map services
	services := []modelsv1.ServiceV2{}
	if isServiceable {
		for _, svc := range result.Services {
			service := modelsv1.ServiceV2{
				ServiceCode:   svc.ServiceCode,
				ServiceName:   svc.ServiceName,
				TATDays:       svc.TATDays,
				IsCOD:         svc.IsCOD,
				Pickup:        svc.Pickup,
				Delivery:      svc.Delivery,
				Insurance:     svc.Insurance,
				ProductTypes:  svc.ProductTypes,
				DeliveryModes: svc.DeliveryModes,
				Rate:          svc.Rate,
			}
			services = append(services, service)
		}
	}
	// Build metadata
	metadata := map[string]interface{}{
		"source_country_code":      srcCC,
		"destination_country_code": dstCC,
		"flow":                     "international",
	}
	if result.Metadata != nil {
		// Merge result metadata
		for k, v := range result.Metadata {
			metadata[k] = v
		}
	}

	partnerResp := modelsv1.PartnerV2Response{
		PartnerID:       partnerID,
		PartnerCode:     result.PartnerCode,
		PartnerName:     "India Post International",
		Rating:          0,
		Source:          "real_time",
		IsServiceable:   isServiceable, // CRITICAL: This was missing!
		PartnerServices: services,
		ResponseTimeMs:  responseTimeMs,
		Metadata:        metadata,
	}

	// Add error if present
	// if result.Error != nil {
	// 	partnerResp.Error = &modelsv1.PartnerError{
	// 		Code:    "INDIA_POST_ERROR",
	// 		Message: result.Error.Error(),
	// 	}
	// } else if result.ErrorMessage != nil {
	// 	partnerResp.Error = &modelsv1.PartnerError{
	// 		Code:    "INDIA_POST_ERROR",
	// 		Message: *result.ErrorMessage,
	// 	}
	// }

	return partnerResp

}

func (s *InternationalStrategy) convertNaqelResult(result *common.PartnerServiceabilityResult, srcCC, dstCC string, responseTimeMs int64) modelsv1.PartnerV2Response {
	if result == nil {
		return modelsv1.PartnerV2Response{
			PartnerID:      "naqel-default-id",
			PartnerCode:    "naqel",
			PartnerName:    "Naqel",
			Rating:         0,
			Source:         "real_time",
			IsServiceable:  false,
			ResponseTimeMs: responseTimeMs,
			Metadata: map[string]interface{}{
				"destination_country_code": dstCC,
				"source_country_code":      srcCC,
				"flow":                     "international",
			},
		}
	}

	isServiceable := len(result.Services) > 0
	partnerID := "naqel-default-id"
	if result.PartnerID != nil {
		partnerID = result.PartnerID.String()
	}

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
					"commercial":   true,
					"document":     true,
					"non_document": true,
				},
				DeliveryModes: map[string]bool{
					"express":  true,
					"standard": false,
				},
			}
			services = append(services, service)
		}
	}

	return modelsv1.PartnerV2Response{
		PartnerID:        partnerID,
		PartnerCode:      "naqel",
		PartnerName:      "Naqel",
		Rating:           0,
		Source:           "real_time",
		IsServiceable:    isServiceable,
		PartnerServices:  services,
		Capabilities:     result.Capabilities,
		ResponseTimeMs:   responseTimeMs,
		Metadata: map[string]interface{}{
			"destination_country_code": dstCC,
			"source_country_code":      srcCC,
			"flow":                     "international",
			"naqel_response":           "success",
		},
	}
}
