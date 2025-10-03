package internationalstrategy

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

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/services/v2/partners/dhl"
	"prayog-serviceability-service/internal/services/v2/partners/factory"
	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"

	"github.com/sirupsen/logrus"
)

// InternationalStrategy orchestrates international flow using HubOps by-pincode → DHL & Aramex
type InternationalStrategy struct {
	PartnerFactory factory.PartnerAdapterFactory
	Logger         *logrus.Logger
}

func (s *InternationalStrategy) Code() string { return "international" }

func (s *InternationalStrategy) Execute(ctx context.Context, req *modelsv1.ServiceabilityV2Request) (*modelsv1.ServiceabilityV2Response, error) {
	if s.Logger == nil {
		s.Logger = logrus.New()
	}
	if s.PartnerFactory == nil {
		return &modelsv1.ServiceabilityV2Response{Success: false, Partners: []modelsv1.PartnerV2Response{}}, nil
	}

	// Determine source postal code
	sourcePin := ""
	if req.SourcePostalCode != nil && *req.SourcePostalCode != "" {
		sourcePin = *req.SourcePostalCode
	} else if req.PostalCode != nil && *req.PostalCode != "" {
		sourcePin = *req.PostalCode
	}
	if sourcePin == "" {
		s.Logger.WithField("component", "international_strategy").Warn("source postal code missing; returning not serviceable")
		return &modelsv1.ServiceabilityV2Response{Success: false, Partners: []modelsv1.PartnerV2Response{}}, nil
	}

	// Set country codes from request
	sourceCountryCode := "IN"  // Default to India
	destinationCountryCode := "US"  // Default to USA

	if req.SourceCountryCode != nil && *req.SourceCountryCode != "" {
		sourceCountryCode = strings.ToUpper(*req.SourceCountryCode)
	} else if req.CountryCode != nil && *req.CountryCode != "" {
		sourceCountryCode = strings.ToUpper(*req.CountryCode)
	}

	if req.DestinationCountryCode != nil && *req.DestinationCountryCode != "" {
		destinationCountryCode = strings.ToUpper(*req.DestinationCountryCode)
	} else if req.CountryCode != nil && *req.CountryCode != "" {
		destinationCountryCode = strings.ToUpper(*req.CountryCode)
	}

	// 1) Call HubOps by-pincode API
	hubResp, err := s.fetchHubByPincode(ctx, sourcePin)
	if err != nil {
		s.Logger.WithError(err).WithFields(logrus.Fields{
			"component":  "international_strategy",
			"source_pin": sourcePin,
		}).Warn("HubOps by-pincode call failed; proceeding without addresses")
	}

	// Build addresses from nearestInternationalHub if available
	var addresses []modelsv1.DetailedAddress
	if hubResp != nil && hubResp.NearestInternationalHub != nil {
		addr := toDetailedAddress(hubResp.NearestInternationalHub)
		addresses = []modelsv1.DetailedAddress{addr}
	}

	// 3) Call DHL & Aramex with adjusted source pincode when available
	partners := make([]modelsv1.PartnerV2Response, 0)
	
	// Determine source and destination pincodes for carrier calls
	srcPin := sourcePin
	dstPin := sourcePin  // Default to source pincode
		
	// Extract pincode and city from nearestInternationalHub
	if hubResp != nil && hubResp.NearestInternationalHub != nil && hubResp.NearestInternationalHub.Pincode != nil {
		// Use nearestInternationalHub pincode for SOURCE only in carriers
		hubPincode := fmt.Sprintf("%v", *hubResp.NearestInternationalHub.Pincode)
		srcPin = hubPincode  // ← HubOps pincode as SOURCE
		
		// Use request destination_postal_code as DESTINATION
		if req.DestinationPostalCode != nil && *req.DestinationPostalCode != "" {
			dstPin = *req.DestinationPostalCode  // ← Request destination as DESTINATION
		} else if req.PostalCode != nil && *req.PostalCode != "" {
			dstPin = *req.PostalCode  // ← Fallback
		}
		
		s.Logger.WithFields(logrus.Fields{
			"component": "international_strategy",
			"original_source_pin": sourcePin,
			"hub_pincode": hubPincode,
			"carrier_source_pin": srcPin,        // ← HubOps pincode
			"carrier_destination_pin": dstPin,   // ← Request destination_postal_code
		}).Info("Using HubOps pincode as source, request destination as destination for carriers")
	}
	
	// Use the resolved country codes
	dstCC := destinationCountryCode
	shipperCity := "Unknown City"
	if hubResp != nil && hubResp.NearestInternationalHub != nil && hubResp.NearestInternationalHub.City != nil && *hubResp.NearestInternationalHub.City != "" { 
		shipperCity = *hubResp.NearestInternationalHub.City 
	}
	receiverCity := "Unknown City"

	// Call DHL (existing code)
	dhlPartner := s.callDHL(ctx, req, srcPin, dstPin, sourceCountryCode, dstCC, shipperCity, receiverCity)
	if dhlPartner.PartnerCode != "" {
		partners = append(partners, dhlPartner)
	}

	// Call Aramex using the existing adapter (new code)
	aramexPartner := s.callAramexViaAdapter(ctx, req, srcPin, dstPin, sourceCountryCode, dstCC, shipperCity, receiverCity)
	if aramexPartner.PartnerCode != "" {
		partners = append(partners, aramexPartner)
	}

	// Call FedEx using the adapter
	fedexPartner := s.callFedExViaAdapter(ctx, req, srcPin, dstPin, sourceCountryCode, dstCC, shipperCity, receiverCity)
	if fedexPartner.PartnerCode != "" {
		partners = append(partners, fedexPartner)
	}

	serviceabilityResp := &modelsv1.ServiceabilityV2Response{
		Success:  len(partners) > 0,
		Partners: partners,
	}
	if len(addresses) > 0 {
		serviceabilityResp.Addresses = addresses
	}
	return serviceabilityResp, nil
}

// callDHL - extracted existing DHL call into separate function
func (s *InternationalStrategy) callDHL(ctx context.Context, req *modelsv1.ServiceabilityV2Request, srcPin, dstPin, srcCC, dstCC, shipperCity, receiverCity string) modelsv1.PartnerV2Response {
	// DHL base URL from .env with fallback to hardcoded
	baseURL := os.Getenv("DHL_BASE_URL")
	if baseURL == "" {
		baseURL = ""
	}
	
	dhlReq := dhl.RatesRequest{
		CustomerDetails: dhl.CustomerDetails{
			ShipperDetails: dhl.ShipperDetails{ PostalCode: srcPin, CityName: shipperCity, CountryCode: srcCC },
			ReceiverDetails: dhl.ReceiverDetails{ PostalCode: dstPin, CityName: receiverCity, CountryCode: dstCC },
		},
		Accounts: []dhl.Account{{ TypeCode: "shipper", Number: "533748932" }},
		ProductsAndServices: []dhl.ProductAndService{{ ProductCode: "P", LocalProductCode: "P" }},
		PayerCountryCode: "IN",
		PlannedShippingDateAndTime: nextBusinessDayOnePMIST(),
		UnitOfMeasurement: "metric",
		IsCustomsDeclarable: true,
		EstimatedDeliveryDate: dhl.EstimatedDeliveryDate{ IsRequested: true, TypeCode: "QDDC" },
		ReturnStandardProductsOnly: true,
		Packages: []dhl.Package{{ Weight: defaultWeight(req), Dimensions: dhl.Dimensions{ Length: defaultLen(req), Width: defaultWid(req), Height: defaultHei(req) }}},
	}

	// Log complete DHL request details
	requestBody, _ := json.MarshalIndent(dhlReq, "", "  ")
	s.Logger.WithFields(logrus.Fields{
		"component": "international_strategy",
		"partner": "dhl",
		"action": "dhl_rates_request",
		"request_body": string(requestBody),
		"source_pincode": srcPin,
		"destination_pincode": dstPin,
		"source_city": shipperCity,
		"destination_city": receiverCity,
		"source_country": srcCC,
		"destination_country": dstCC,
		"weight": defaultWeight(req),
		"dimensions": fmt.Sprintf("%.2fx%.2fx%.2f", defaultLen(req), defaultWid(req), defaultHei(req)),
		"shipping_date": nextBusinessDayOnePMIST(),
		"dhl_base_url": baseURL,
		"dhl_enabled": true,
	}).Info("Complete DHL Rates API Request")

	url := fmt.Sprintf("%s/rates?strictValidation=false", baseURL)
	body, _ := json.Marshal(dhlReq)
	reqHTTP, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Accept", "application/json")
	reqHTTP.Header.Set("Message-Reference", "d0e7832e-5c98-11ea-bc55-0242ac13")
	reqHTTP.Header.Set("Message-Reference-Date", "Wed, 21 Oct 2015 07:28:00 GMT")
	reqHTTP.Header.Set("Plugin-Name", "")
	reqHTTP.Header.Set("Plugin-Version", "")
	reqHTTP.Header.Set("Shipping-System-Platform-Name", "")
	reqHTTP.Header.Set("Shipping-System-Platform-Version", "")
	reqHTTP.Header.Set("Webstore-Platform-Name", "")
	reqHTTP.Header.Set("Webstore-Platform-Version", "")
	reqHTTP.Header.Set("X-Version", "2.12.0")
	// Authorization header from .env
	if basicAuth := os.Getenv("DHL_BASIC_AUTH"); basicAuth != "" {
		reqHTTP.Header.Set("Authorization", "Basic "+basicAuth)
	} 
	client := &http.Client{ Timeout: 15 * time.Second }
	dhlResp, doErr := client.Do(reqHTTP)
	if doErr != nil {
		s.Logger.WithError(doErr).WithFields(logrus.Fields{"component":"international_strategy","partner":"dhl"}).Warn("DHL HTTP call failed")
		return modelsv1.PartnerV2Response{}
	}
	defer dhlResp.Body.Close()
	var rates dhl.RatesResponse
	if dhlResp.StatusCode == http.StatusOK {
		if decErr := json.NewDecoder(dhlResp.Body).Decode(&rates); decErr == nil && len(rates.Products) > 0 {
			cap := map[string]interface{}{}
			prod := rates.Products[0]
			cap["total_transit_days"] = prod.DeliveryCapabilities.TotalTransitDays
			cap["estimated_delivery_date_and_time"] = prod.DeliveryCapabilities.EstimatedDeliveryDateAndTime
			return modelsv1.PartnerV2Response{
				PartnerID:   "93a2d552-dd7a-4786-aa11-cf44e7b327ab",
				PartnerCode: "dhl",
				Rating:      0,
				Capabilities: cap,
				Metadata:    map[string]interface{}{"source_country_code":srcCC,"destination_country_code":dstCC},
			}
		} else if decErr != nil {
			s.Logger.WithError(decErr).WithFields(logrus.Fields{"component":"international_strategy","partner":"dhl"}).Warn("Failed to decode DHL response")
		}
	} else {
		b, _ := io.ReadAll(dhlResp.Body)
		s.Logger.WithFields(logrus.Fields{"component":"international_strategy","partner":"dhl","status":dhlResp.StatusCode,"body":string(b)}).Warn("DHL returned non-200")
	}
	
	return modelsv1.PartnerV2Response{}
}

// callAramexViaAdapter - uses existing Aramex adapter for serviceability check
func (s *InternationalStrategy) callAramexViaAdapter(ctx context.Context, req *modelsv1.ServiceabilityV2Request, srcPin, dstPin, srcCC, dstCC, shipperCity, receiverCity string) modelsv1.PartnerV2Response {
    // Get Aramex adapter from factory - returns (adapter, found)
    aramexAdapter, adapterFound := s.PartnerFactory.GetAdapter("aramex")
    
    s.Logger.WithFields(logrus.Fields{
        "adapterFound": adapterFound,
        "adapterNil":   aramexAdapter == nil,
    }).Info("Aramex adapter lookup result")
    
    if !adapterFound {
        s.Logger.WithFields(logrus.Fields{
            "component": "international_strategy",
            "partner":   "aramex",
        }).Warn("Aramex adapter not found in factory")
        return modelsv1.PartnerV2Response{}
    }
    
    if aramexAdapter == nil {
        s.Logger.WithFields(logrus.Fields{"component":"international_strategy","partner":"aramex"}).Warn("Aramex adapter is nil")
        return modelsv1.PartnerV2Response{}
    }

    // Prepare request for Aramex adapter
    aramexReq := &modelsv1.ServiceabilityV2Request{
        SourcePostalCode:       &srcPin,
        DestinationPostalCode:  &dstPin,
        SourceCountryCode:      &srcCC,
        DestinationCountryCode: &dstCC,
        Packages:               req.Packages,
        PostalCode:             req.PostalCode,
        CountryCode:            req.CountryCode,
    }

    partnerInfo := common.PartnerInfo{
        PartnerCode: "aramex",
    }

    s.Logger.WithFields(logrus.Fields{
        "component":           "international_strategy",
        "partner":             "aramex", 
        "action":              "aramex_serviceability_check",
        "source_pincode":      srcPin,
        "destination_pincode": dstPin,
        "source_country":      srcCC,
        "destination_country": dstCC,
        "request":             aramexReq, // Log the actual request
    }).Info("Calling Aramex adapter for serviceability")

    // Call Aramex adapter
    result, serviceabilityErr := aramexAdapter.CheckServiceability(ctx, aramexReq, partnerInfo)
    s.Logger.WithFields(logrus.Fields{
        "resultReceived": result != nil,
        "hasError":       serviceabilityErr != nil,
    }).Info("Aramex adapter CheckServiceability completed")

    if serviceabilityErr != nil {
        s.Logger.WithError(serviceabilityErr).WithFields(logrus.Fields{
            "component": "international_strategy",
            "partner":   "aramex",
        }).Warn("Aramex adapter call failed")
        return modelsv1.PartnerV2Response{}
    }

    // Use the special Aramex converter
    partnerResp := s.convertAramexResult(result, srcCC, dstCC)
    
    if partnerResp.PartnerCode != "" {
        s.Logger.WithFields(logrus.Fields{
            "servicesCount": len(partnerResp.Services),
            "serviceable":   len(partnerResp.Services) > 0,
        }).Info("Successfully created Aramex partner response")
        return partnerResp
    }

    s.Logger.Warn("Failed to create Aramex partner response")
    return modelsv1.PartnerV2Response{}
}


// callFedExViaAdapter - uses existing FedEx adapter for serviceability check
func (s *InternationalStrategy) callFedExViaAdapter(ctx context.Context, req *modelsv1.ServiceabilityV2Request, srcPin, dstPin, srcCC, dstCC, shipperCity, receiverCity string) modelsv1.PartnerV2Response {
    // Get FedEx adapter from factory - returns (adapter, found)
    fedexAdapter, adapterFound := s.PartnerFactory.GetAdapter("fedex")
    
    s.Logger.WithFields(logrus.Fields{
        "adapterFound": adapterFound,
        "adapterNil":   fedexAdapter == nil,
    }).Info("FedEx adapter lookup result")
    
    if !adapterFound {
        s.Logger.WithFields(logrus.Fields{
            "component": "international_strategy",
            "partner":   "fedex",
        }).Warn("FedEx adapter not found in factory")
        return modelsv1.PartnerV2Response{}
    }
    
    if fedexAdapter == nil {
        s.Logger.WithFields(logrus.Fields{"component":"international_strategy","partner":"fedex"}).Warn("FedEx adapter is nil")
        return modelsv1.PartnerV2Response{}
    }

    // Prepare request for FedEx adapter
    fedexReq := &modelsv1.ServiceabilityV2Request{
        SourcePostalCode:       &srcPin,
        DestinationPostalCode:  &dstPin,
        SourceCountryCode:      &srcCC,
        DestinationCountryCode: &dstCC,
        Packages:               req.Packages,
        PostalCode:             req.PostalCode,
        CountryCode:            req.CountryCode,
    }

    partnerInfo := common.PartnerInfo{
        PartnerCode: "fedex",
    }

    s.Logger.WithFields(logrus.Fields{
        "component":           "international_strategy",
        "partner":             "fedex", 
        "action":              "fedex_serviceability_check",
        "source_pincode":      srcPin,
        "destination_pincode": dstPin,
        "source_country":      srcCC,
        "destination_country": dstCC,
        "request":             fedexReq, // Log the actual request
    }).Info("Calling FedEx adapter for serviceability")

    // Call FedEx adapter
    result, serviceabilityErr := fedexAdapter.CheckServiceability(ctx, fedexReq, partnerInfo)
    s.Logger.WithFields(logrus.Fields{
        "resultReceived": result != nil,
        "hasError":       serviceabilityErr != nil,
    }).Info("FedEx adapter CheckServiceability completed")

    if serviceabilityErr != nil {
        s.Logger.WithError(serviceabilityErr).WithFields(logrus.Fields{
            "component": "international_strategy",
            "partner":   "fedex",
        }).Warn("FedEx adapter call failed")
        return modelsv1.PartnerV2Response{}
    }

    // Use the special FedEx converter
    partnerResp := s.convertFedExResult(result, srcCC, dstCC)
    
    if partnerResp.PartnerCode != "" {
        s.Logger.WithFields(logrus.Fields{
            "servicesCount": len(partnerResp.Services),
            "serviceable":   len(partnerResp.Services) > 0,
        }).Info("Successfully created FedEx partner response")
        return partnerResp
    }


    s.Logger.Warn("Failed to create FedEx partner response")
    return modelsv1.PartnerV2Response{}
}

// convertFedExResult - special handler for FedEx responses
func (s *InternationalStrategy) convertFedExResult(res *common.PartnerServiceabilityResult, srcCC, dstCC string) modelsv1.PartnerV2Response {
    if res == nil {
        return modelsv1.PartnerV2Response{}
    }

    // Check if serviceable from metadata or services
    isServiceable := len(res.Services) > 0
    if res.Metadata != nil {
        if serviceable, ok := res.Metadata["is_serviceable"].(bool); ok {
            isServiceable = serviceable
        }
    }

    // Create capabilities based on metadata
    capabilities := make(map[string]interface{})
    if res.Metadata != nil {
        // Copy relevant metadata to capabilities
        if transitDays, ok := res.Metadata["transit_days"]; ok {
            capabilities["total_transit_days"] = transitDays
        }
        if deliveryDate, ok := res.Metadata["estimated_delivery_date"]; ok {
            capabilities["estimated_delivery_date_and_time"] = deliveryDate
        }
        if availableServices, ok := res.Metadata["available_services_count"]; ok {
            capabilities["available_services"] = availableServices
        }
    }

    // Use existing services or create default ones if serviceable
    var services []modelsv1.ServiceV2
    if isServiceable {
        if len(res.Services) > 0 {
            // Use the services from the result
            services = res.Services
        } else {
            // Create default service for FedEx
            service := modelsv1.ServiceV2{
                ServiceName: "FedEx International Express",
                TATDays:     3, // Default for international
                Pickup:      true,
                Delivery:    true,
                Insurance:   true,
                ProductTypes: map[string]bool{
                    "document":     true,
                    "non_document": true,
                    "commercial":   true,
                },
                DeliveryModes: map[string]bool{
                    "express":  true,
                    "standard": false,
                },
            }
            services = []modelsv1.ServiceV2{service}
        }
    }

    partnerID := "unknown"
    if res.PartnerID != nil {
        partnerID = res.PartnerID.String()
    }

    return modelsv1.PartnerV2Response{
        PartnerID:       partnerID,
        PartnerCode:     "fedex",
        PartnerName:     "FedEx",
        Rating:          0,
        Services:        services,
        PartnerServices: res.PartnerServices,
        Capabilities:    capabilities,
        Error:           res.ErrorMessage,
        ResponseTime:    res.ResponseTime,
        Metadata: map[string]interface{}{
            "source_country_code":      srcCC,
            "destination_country_code": dstCC,
            "flow":                     "international",
            "fedex_metadata":           res.Metadata,
        },
    }
}
// [Rest of the file remains exactly the same - all existing HubOps, helper functions, etc.]
// HubOps API integration
var hubOpsURL = func() string {
	if url := os.Getenv("SMILE_HUBOPS_BY_SOURCE_PINCODE"); url != "" {
		return url
	}
	// No fallback - must be set in .env
	return ""
}()

type hubOpsRequest struct {
	PostalCode string `json:"postalCode"`
}

type hubInfo struct {
	PremiseID           *int64  `json:"premiseId"`
	PremiseName         *string `json:"premiseName"`
	City                *string `json:"city"`
	Address             *string `json:"address"`
	AddressLine1        *string `json:"addressLine1"`
	AddressLine2        *string `json:"addressLine2"`
	Pincode             *int64  `json:"pincode"`
	State               *string `json:"state"`
	Latitude            *string `json:"latitude"`
	Longitude           *string `json:"longitude"`
	PersonalEmailId     *string     `json:"personalEmailId"`
	OfficialEmailId     *string     `json:"officialEmailId"`
	PersonalNumber      interface{} `json:"personalNumber"`
	OfficialNumber      interface{} `json:"officialNumber"`
}

type hubOpsResponse struct {
	NearestHub              *hubInfo `json:"nearestHub"`
	NearestInternationalHub *hubInfo `json:"nearestInternationalHub"`
	Nearest3PLHub           *hubInfo `json:"nearest3PLHub"`
}

func (s *InternationalStrategy) fetchHubByPincode(ctx context.Context, pin string) (*hubOpsResponse, error) {
	body, _ := json.Marshal(hubOpsRequest{PostalCode: pin})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hubOpsURL, bytes.NewBuffer(body))
	if err != nil { return nil, err }
	req.Header.Set("Content-Type", "application/json")
    if cookie := os.Getenv("HUBOPS_COOKIE"); cookie != "" { req.Header.Set("Cookie", cookie) }

	client := &http.Client{ Timeout: 5 * time.Second }
	resp, err := client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("hubops API status %d", resp.StatusCode)
	}
	var parsed hubOpsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}

func toDetailedAddress(h *hubInfo) modelsv1.DetailedAddress {
	addr := modelsv1.DetailedAddress{ Type: "INTERNATIONAL_HUB_ADDRESS", AddressName: "WAREHOUSE" }
	if h == nil { return addr }
	if h.Pincode != nil { addr.Zip = fmt.Sprintf("%v", *h.Pincode) }
	if h.PremiseName != nil { addr.Name = *h.PremiseName }
	// Prefer official contact details, fallback to personal
	if s := anyToString(h.OfficialNumber); s != "" { addr.Phone = s } else { addr.Phone = anyToString(h.PersonalNumber) }
	if h.OfficialEmailId != nil && *h.OfficialEmailId != "" { addr.Email = *h.OfficialEmailId } else if h.PersonalEmailId != nil { addr.Email = *h.PersonalEmailId }
	if h.AddressLine1 != nil { addr.Street = *h.AddressLine1 } else if h.Address != nil { addr.Street = *h.Address }
	// No explicit landmark field in payload; leave empty
	if h.City != nil { addr.City = *h.City }
	if h.State != nil { addr.State = *h.State }
	// Country not provided; leave empty to avoid incorrect data
	if h.Latitude != nil { if v, ok := toFloat(*h.Latitude); ok { addr.Latitude = &v } }
	if h.Longitude != nil { if v, ok := toFloat(*h.Longitude); ok { addr.Longitude = &v } }
	return addr
}

func toFloat(s string) (float64, bool) {
	var f float64
	// simple parse without importing strconv to keep deps minimal in this file
	// however, to ensure correctness, we will use fmt.Sscanf
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil { return 0, false }
	return f, true
}

// anyToString converts number or string to string
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
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func toPartnerV2Response(res *common.PartnerServiceabilityResult, code string) (modelsv1.PartnerV2Response, bool) {
	if res == nil { return modelsv1.PartnerV2Response{}, false }
	if res.ErrorMessage != nil { return modelsv1.PartnerV2Response{}, false }
	hasServices := len(res.Services) > 0
	hasCaps := len(res.Capabilities) > 0
	hasMeta := len(res.Metadata) > 0
	if !(hasServices || hasCaps || hasMeta) { return modelsv1.PartnerV2Response{}, false }
	partnerID := "unknown"
	if res.PartnerID != nil { partnerID = res.PartnerID.String() }
	return modelsv1.PartnerV2Response{
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

// Helpers for DHL request defaults
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

func defaultWeight(req *modelsv1.ServiceabilityV2Request) float64 {
    if req != nil && len(req.Packages) > 0 && req.Packages[0].Weight != nil && req.Packages[0].Weight.Value > 0 {
        return req.Packages[0].Weight.Value
    }
    return 1
}

func defaultLen(req *modelsv1.ServiceabilityV2Request) float64 {
    if req != nil && len(req.Packages) > 0 && req.Packages[0].Dimensions != nil && req.Packages[0].Dimensions.Length > 0 {
        return req.Packages[0].Dimensions.Length
    }
    return 10
}

func defaultWid(req *modelsv1.ServiceabilityV2Request) float64 {
    if req != nil && len(req.Packages) > 0 && req.Packages[0].Dimensions != nil && req.Packages[0].Dimensions.Width > 0 {
        return req.Packages[0].Dimensions.Width
    }
    return 10
}

func defaultHei(req *modelsv1.ServiceabilityV2Request) float64 {
    if req != nil && len(req.Packages) > 0 && req.Packages[0].Dimensions != nil && req.Packages[0].Dimensions.Height > 0 {
        return req.Packages[0].Dimensions.Height
    }
    return 10
}

// convertAramexResult - special handler for Aramex responses
func (s *InternationalStrategy) convertAramexResult(res *common.PartnerServiceabilityResult, srcCC, dstCC string) modelsv1.PartnerV2Response {
	if res == nil {
		return modelsv1.PartnerV2Response{}
	}

	// Check if serviceable from metadata
	isServiceable := false
	if res.Metadata != nil {
		if serviceable, ok := res.Metadata["is_serviceable"].(bool); ok {
			isServiceable = serviceable
		}
	}

	// Create capabilities based on metadata
	capabilities := make(map[string]interface{})
	if res.Metadata != nil {
		// Copy relevant metadata to capabilities
		if transitDays, ok := res.Metadata["total_transit_days"]; ok {
			capabilities["total_transit_days"] = transitDays
		}
		if deliveryDate, ok := res.Metadata["estimated_delivery_date_and_time"]; ok {
			capabilities["estimated_delivery_date_and_time"] = deliveryDate
		}
	}

	// Create services if serviceable
	var services []modelsv1.ServiceV2
	if isServiceable {
		service := modelsv1.ServiceV2{
			ServiceName: "Aramex International Express",
			TATDays:     3, // Default for international
			Pickup:      true,
			Delivery:    true,
			Insurance:   true,
			ProductTypes: map[string]bool{
				"document":     true,
				"non_document": true,
				"commercial":   true,
			},
			DeliveryModes: map[string]bool{
				"express":  true,
				"standard": false,
			},
		}
		services = []modelsv1.ServiceV2{service}
	}

	partnerID := "unknown"
	if res.PartnerID != nil {
		partnerID = res.PartnerID.String()
	}

	return modelsv1.PartnerV2Response{
		PartnerID:       partnerID,
		PartnerCode:     "aramex",
		PartnerName:     "Aramex",
		Rating:          0,
		Services:        services,
		PartnerServices: res.PartnerServices,
		Capabilities:    capabilities,
		Error:           res.ErrorMessage,
		ResponseTime:    res.ResponseTime,
		Metadata: map[string]interface{}{
			"source_country_code":      srcCC,
			"destination_country_code": dstCC,
			"flow":                     "international",
			"aramex_metadata":          res.Metadata,
		},
	}
}