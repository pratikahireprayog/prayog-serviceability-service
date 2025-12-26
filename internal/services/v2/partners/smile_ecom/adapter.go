package smile_ecom

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SmileEcomAdapter implements the common.PartnerAdapter interface for Smile Ecom database operations
type SmileEcomAdapter struct {
	repository repositories.EcommRepository
	config     config.SmileEcomConfig
	logger     *logrus.Logger
}

// NewSmileEcomAdapter creates a new Smile Ecom adapter
func NewSmileEcomAdapter(cfg config.SmileEcomConfig, db *sql.DB) common.PartnerAdapter {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create gorm.DB from sql.DB
	var gormDB *gorm.DB
	var err error
	if db != nil {
		gormDB, err = gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), &gorm.Config{})
		if err != nil {
			logger.WithError(err).Warn("Failed to create gorm.DB from sql.DB, repository operations may fail")
		}
	}

	// Create repository
	var ecommRepo repositories.EcommRepository
	if gormDB != nil {
		ecommRepo = repositories.NewEcommRepository(gormDB, cfg.TableName)
	} else {
		logger.Warn("No database connection available for Ecomm repository")
	}

	logger.WithFields(logrus.Fields{
		"partner":    "Smile Ecom",
		"table_name": cfg.TableName,
		"enabled":    cfg.Enabled,
	}).Info("Creating Smile Ecom adapter")

	return &SmileEcomAdapter{
		repository: ecommRepo,
		config:     cfg,
		logger:     logger,
	}
}

// GetAdapterType returns the adapter type
func (s *SmileEcomAdapter) GetAdapterType() common.AdapterType {
	return common.AdapterTypeDatabase
}

// IsEnabled returns whether the adapter is enabled
func (s *SmileEcomAdapter) IsEnabled() bool {
	return s.config.Enabled
}

// CheckServiceability checks if Smile Ecom can service the given request
func (s *SmileEcomAdapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	partnerID := ""
	if partnerInfo.PartnerID != nil {
		partnerID = partnerInfo.PartnerID.String()
	}

	s.logger.WithFields(logrus.Fields{
		"component":    "smile_ecom_adapter",
		"action":        "check_serviceability_start",
		"partner_code":  partnerInfo.PartnerCode,
		"partner_id":    partnerID,
	}).Info("Starting Smile Ecom adapter serviceability check")

	// Validate Smile Ecom specific requirements
	if err := s.validateRequirements(request); err != nil {
		s.logger.WithFields(logrus.Fields{
			"component":    "smile_ecom_adapter",
			"event":         "validation_failed",
			"partner_code":  partnerInfo.PartnerCode,
			"partner_id":    partnerID,
			"error":         err.Error(),
		}).Warn("Smile Ecom validation failed")

		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Smile Ecom validation failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "Smile Ecom validation failed",
			},
		}, nil
	}

	// Determine which pincode to check
	// For Smile Ecom, we need the destination pincode
	var pincodeToCheck string
	if request.DestinationPostalCode != nil && *request.DestinationPostalCode != "" {
		pincodeToCheck = *request.DestinationPostalCode
	} else if request.PostalCode != nil && *request.PostalCode != "" {
		pincodeToCheck = *request.PostalCode
	} else {
		err := fmt.Errorf("destination postal code is required for Smile Ecom")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{err.Error()}[0],
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"component":    "smile_ecom_adapter",
		"partner_code":  partnerInfo.PartnerCode,
		"partner_id":    partnerID,
		"pincode":       pincodeToCheck,
	}).Info("Checking serviceability for Smile Ecom")

	// Check if repository is available
	if s.repository == nil {
		errMsg := "Smile Ecom repository not available"
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			ErrorMessage: &errMsg,
			Metadata: map[string]interface{}{
				"reason": "Database connection not available",
			},
		}, nil
	}

	// Check serviceability in database
	pincodeData, err := s.repository.CheckServiceabilityByPincode(ctx, pincodeToCheck)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"component":    "smile_ecom_adapter",
			"partner_code":  partnerInfo.PartnerCode,
			"partner_id":    partnerID,
			"pincode":       pincodeToCheck,
			"error":         err.Error(),
		}).Warn("Smile Ecom pincode not found in database")

		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Metadata: map[string]interface{}{
				"reason":  "Pincode not serviceable by Smile Ecom",
				"pincode": pincodeToCheck,
			},
		}, nil
	}

	// Convert database result to serviceability result
	result := s.convertToServiceabilityResult(pincodeData, partnerInfo, pincodeToCheck)
	result.ResponseTime = time.Since(startTime)

	s.logger.WithFields(logrus.Fields{
		"component":      "smile_ecom_adapter",
		"partner_code":   partnerInfo.PartnerCode,
		"partner_id":     partnerID,
		"pincode":        pincodeToCheck,
		"is_serviceable": len(result.Services) > 0,
		"response_time":  result.ResponseTime,
	}).Info("Smile Ecom serviceability check completed")

	return result, nil
}

// validateRequirements validates Smile Ecom specific requirements
func (s *SmileEcomAdapter) validateRequirements(request *models.ServiceabilityV2Request) error {
	// Smile Ecom requires either destination postal code or generic postal code
	if (request.DestinationPostalCode == nil || *request.DestinationPostalCode == "") &&
		(request.PostalCode == nil || *request.PostalCode == "") {
		return fmt.Errorf("postal code is required for Smile Ecom serviceability")
	}

	// Validate pincode format (6 digits for Indian pincodes)
	var pincodeToCheck string
	if request.DestinationPostalCode != nil && *request.DestinationPostalCode != "" {
		pincodeToCheck = *request.DestinationPostalCode
	} else if request.PostalCode != nil {
		pincodeToCheck = *request.PostalCode
	}

	// Check if pincode is numeric and 6 digits
	if _, err := strconv.Atoi(pincodeToCheck); err != nil {
		return fmt.Errorf("invalid pincode format: must be numeric")
	}

	if len(pincodeToCheck) != 6 {
		return fmt.Errorf("invalid pincode format: must be 6 digits")
	}

	return nil
}

// convertToServiceabilityResult converts database result to common format
func (s *SmileEcomAdapter) convertToServiceabilityResult(pincodeData *repositories.EcommPincode, partnerInfo common.PartnerInfo, pincode string) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// Build capabilities from database data
	// FM (First Mile) = Pickup capability
	// LM (Last Mile) = Delivery capability
	capabilities := map[string]interface{}{
		"pincode":          pincode,
		"is_serviceable":   true,
		"parcel_category":  "ecomm",
		"city":             pincodeData.City,
		"state":            pincodeData.State,
		"cod_available":    pincodeData.COD,
		"pickup_available":   pincodeData.FM,  // First Mile = Pickup
		"delivery_available": pincodeData.LM,  // Last Mile = Delivery
		"fm":               pincodeData.FM,    // First Mile (for backward compatibility)
		"lm":               pincodeData.LM,    // Last Mile (for backward compatibility)
	}

	// Create service entry
	// FM (First Mile) maps to Pickup capability
	// LM (Last Mile) maps to Delivery capability
	service := models.ServiceV2{
		ServiceCode: "SMILE_ECOM_STANDARD",
		ServiceName: "Smile Ecom Standard",
		TATDays:     3, // Default TAT, can be enhanced based on zone/distance
		IsCOD:       pincodeData.COD,
		Pickup:      pincodeData.FM, // First Mile = Pickup available
		Delivery:    pincodeData.LM, // Last Mile = Delivery available
		Insurance:   true,            // Insurance available by default
		ProductTypes: map[string]bool{
			"ecommerce": true,
		},
		DeliveryModes: map[string]bool{
			"standard": true,
		},
	}

	result.Services = append(result.Services, service)
	result.Capabilities = capabilities

	// Add metadata
	result.Metadata["pincode"] = pincode
	result.Metadata["is_serviceable"] = true
	result.Metadata["parcel_category"] = "ecomm"
	result.Metadata["serviceability"] = pincodeData.Serviceability

	return result
}

// Initialize implements PartnerAdapter interface
func (s *SmileEcomAdapter) Initialize(ctx context.Context) error {
	s.logger.Info("Smile Ecom adapter initialized successfully")
	return nil
}

// IsHealthy implements PartnerAdapter interface
func (s *SmileEcomAdapter) IsHealthy(ctx context.Context) bool {
	if !s.config.Enabled {
		return false
	}
	// Simple health check - verify config is valid
	return s.config.TableName != ""
}

// GetMetrics implements PartnerAdapter interface
func (s *SmileEcomAdapter) GetMetrics() *common.PartnerMetrics {
	return &common.PartnerMetrics{
		PartnerCode:         "", // Will be set by orchestrator from database
		TotalRequests:       0,  // TODO: Implement actual metrics
		SuccessfulRequests:  0,
		FailedRequests:      0,
		AverageResponseTime: 0,
		HealthStatus:        "healthy",
		ErrorRate:           0.0,
	}
}

// Shutdown implements PartnerAdapter interface
func (s *SmileEcomAdapter) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down Smile Ecom adapter")
	return nil
}

