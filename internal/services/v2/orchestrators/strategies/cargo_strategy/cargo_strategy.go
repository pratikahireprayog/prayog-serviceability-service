package cargostrategy

import (
	"context"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/services/v2/partners/factory"
	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"
)

// CargoStrategy handles cargo/freight serviceability requests
// This strategy directly calls the Smile Cargo adapter for cargo-related requests
type CargoStrategy struct {
	PartnerFactory factory.PartnerAdapterFactory
	Logger         *logrus.Logger
}

func (s *CargoStrategy) Code() string { return "cargo" }

func (s *CargoStrategy) Execute(ctx context.Context, req *modelsv1.ServiceabilityV2Request) (*modelsv1.ServiceabilityV2Response, error) {
	if s.Logger == nil {
		s.Logger = logrus.New()
	}
	if s.PartnerFactory == nil {
		return &modelsv1.ServiceabilityV2Response{Success: false, Partners: []modelsv1.PartnerV2Response{}}, nil
	}

	startTime := time.Now()
	
	s.Logger.WithFields(logrus.Fields{
		"component": "cargo_strategy",
		"method":    "Execute",
		"source_pincode": req.SourcePostalCode,
		"destination_pincode": req.DestinationPostalCode,
		"postal_code": req.PostalCode,
		"parcel_category": req.ParcelCategory,
		"product_type": req.ProductType,
	}).Info("🚛 Starting Cargo Strategy execution")

	// Validate cargo request
	if !s.isCargoRequest(req) {
		s.Logger.WithFields(logrus.Fields{
			"component": "cargo_strategy",
			"method":    "Execute",
			"parcel_category": req.ParcelCategory,
			"product_type": req.ProductType,
		}).Warn("Request is not a cargo request - strategy not applicable")
		
		return &modelsv1.ServiceabilityV2Response{
			Success:  false,
			Partners: []modelsv1.PartnerV2Response{},
		}, nil
	}

	// Validate required pincodes
	if !s.hasValidPincodes(req) {
		s.Logger.WithFields(logrus.Fields{
			"component": "cargo_strategy",
			"method":    "Execute",
			"source_pincode": req.SourcePostalCode,
			"destination_pincode": req.DestinationPostalCode,
			"postal_code": req.PostalCode,
		}).Warn("Missing required pincodes for cargo request")
		
		return &modelsv1.ServiceabilityV2Response{
			Success:  false,
			Partners: []modelsv1.PartnerV2Response{},
		}, nil
	}

	// Get Smile Cargo adapter
	cargoAdapter, exists := s.PartnerFactory.GetAdapter("smile_cargo")
	if !exists {
		s.Logger.WithFields(logrus.Fields{
			"component": "cargo_strategy",
			"method":    "Execute",
			"partner":   "smile_cargo",
		}).Error("Smile Cargo adapter not found")
		
		return &modelsv1.ServiceabilityV2Response{
			Success:  false,
			Partners: []modelsv1.PartnerV2Response{},
		}, nil
	}

	// Get partner info for Smile Cargo
	partnerInfo := common.PartnerInfo{
		PartnerCode: "smile_cargo",
	}

	// Call Smile Cargo adapter
	s.Logger.WithFields(logrus.Fields{
		"component": "cargo_strategy",
		"method":    "Execute",
		"adapter":   "smile_cargo",
	}).Info("📞 Calling Smile Cargo adapter for serviceability check")
	
	result, err := cargoAdapter.CheckServiceability(ctx, req, partnerInfo)
	if err != nil {
		s.Logger.WithFields(logrus.Fields{
			"component": "cargo_strategy",
			"method":    "Execute",
			"adapter":   "smile_cargo",
			"error":     err.Error(),
		}).Error("Smile Cargo adapter call failed")
		
		return &modelsv1.ServiceabilityV2Response{
			Success:  false,
			Partners: []modelsv1.PartnerV2Response{},
		}, nil
	}

	// Convert result to V2 response format
	partners := s.convertToPartnerV2Response(result)
	
	executionTime := time.Since(startTime)
	s.Logger.WithFields(logrus.Fields{
		"component": "cargo_strategy",
		"method":    "Execute",
		"total_time_ms": executionTime.Milliseconds(),
		"partners_count": len(partners),
		"success": len(partners) > 0,
	}).Info("🚛 Cargo Strategy execution completed")

	return &modelsv1.ServiceabilityV2Response{
		Success:  len(partners) > 0,
		Partners: partners,
	}, nil
}

// isCargoRequest determines if this is a cargo/freight request
func (s *CargoStrategy) isCargoRequest(req *modelsv1.ServiceabilityV2Request) bool {
	// Check parcel category
	if req.ParcelCategory != nil {
		category := strings.ToLower(*req.ParcelCategory)
		if category == "cargo" || category == "freight" {
			return true
		}
	}

	// Check product type
	if req.ProductType != nil {
		productType := strings.ToLower(*req.ProductType)
		if productType == "cargo" || productType == "freight" {
			return true
		}
	}

	return false
}

// hasValidPincodes checks if request has valid pincodes for cargo
func (s *CargoStrategy) hasValidPincodes(req *modelsv1.ServiceabilityV2Request) bool {
	// Need source pincode
	if req.SourcePostalCode == nil || *req.SourcePostalCode == "" {
		return false
	}

	// Need destination pincode
	hasDestination := (req.DestinationPostalCode != nil && *req.DestinationPostalCode != "") ||
		(req.PostalCode != nil && *req.PostalCode != "")

	return hasDestination
}

// convertToPartnerV2Response converts adapter result to V2 response format
func (s *CargoStrategy) convertToPartnerV2Response(result *common.PartnerServiceabilityResult) []modelsv1.PartnerV2Response {
	partners := make([]modelsv1.PartnerV2Response, 0)

	if result == nil || len(result.Services) == 0 {
		return partners
	}

	// Create partner response
	partner := modelsv1.PartnerV2Response{
		PartnerID:   "smile_cargo_uuid", // Use a fixed UUID for Smile Cargo
		PartnerCode: result.PartnerCode,
		PartnerName: "Smile Cargo",
		Services:    result.Services,
		Capabilities: result.Capabilities,
		Metadata: map[string]interface{}{
			"strategy": "cargo",
			"adapter":  "smile_cargo",
			"reason":   result.Metadata["reason"],
		},
	}

	partners = append(partners, partner)
	return partners
}
