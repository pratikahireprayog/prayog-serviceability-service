package dharmendra

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

// Adapter implements the PartnerAdapter interface for Dharmendra
type Adapter struct {
	repository repositories.DharmendraRepository
	config     config.DharmendraConfig
	logger     *logrus.Logger
}

// NewAdapter creates a new Dharmendra adapter instance
func NewAdapter(cfg config.DharmendraConfig, db *sql.DB) *Adapter {
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
	var dharmendraRepo repositories.DharmendraRepository
	if gormDB != nil {
		dharmendraRepo = repositories.NewDharmendraRepository(gormDB, cfg.TableName)
	} else {
		logger.Warn("No database connection available for Dharmendra repository")
	}

	logger.WithFields(logrus.Fields{
		"partner":    "Dharmendra",
		"table_name": cfg.TableName,
		"enabled":    cfg.Enabled,
	}).Info("Creating Dharmendra adapter")

	return &Adapter{
		repository: dharmendraRepo,
		config:     cfg,
		logger:     logger,
	}
}

// GetAdapterType returns the adapter type
func (a *Adapter) GetAdapterType() common.AdapterType {
	return common.AdapterTypeDatabase
}

// IsEnabled returns whether the adapter is enabled
func (a *Adapter) IsEnabled() bool {
	return a.config.Enabled
}

// CheckServiceability checks if Dharmendra can service the given request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	partnerID := ""
	if partnerInfo.PartnerID != nil {
		partnerID = partnerInfo.PartnerID.String()
	}

	a.logger.WithFields(logrus.Fields{
		"component":    "dharmendra_adapter",
		"action":       "check_serviceability_start",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
	}).Info("Starting Dharmendra adapter serviceability check")

	// Validate Dharmendra specific requirements
	if err := a.validateRequirements(request); err != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "dharmendra_adapter",
			"event":        "validation_failed",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   partnerID,
			"error":        err.Error(),
		}).Warn("Dharmendra validation failed")

		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Dharmendra validation failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "Dharmendra validation failed",
			},
		}, nil
	}

	// Determine which pincode to check
	// For Dharmendra, we need the destination pincode
	var pincodeToCheck string
	if request.DestinationPostalCode != nil && *request.DestinationPostalCode != "" {
		pincodeToCheck = *request.DestinationPostalCode
	} else if request.PostalCode != nil && *request.PostalCode != "" {
		pincodeToCheck = *request.PostalCode
	} else {
		err := fmt.Errorf("destination postal code is required for Dharmendra")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{err.Error()}[0],
		}, nil
	}

	a.logger.WithFields(logrus.Fields{
		"component":    "dharmendra_adapter",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
		"pincode":      pincodeToCheck,
	}).Info("Checking serviceability for Dharmendra")

	// Check if repository is available
	if a.repository == nil {
		errMsg := "Dharmendra repository not available"
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
	pincodeData, err := a.repository.CheckServiceabilityByPincode(ctx, pincodeToCheck)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "dharmendra_adapter",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   partnerID,
			"pincode":      pincodeToCheck,
			"error":        err.Error(),
		}).Warn("Dharmendra pincode not found in database")

		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Metadata: map[string]interface{}{
				"reason":  "Pincode not serviceable by Dharmendra",
				"pincode": pincodeToCheck,
			},
		}, nil
	}

	// Convert database result to serviceability result
	result := a.convertToServiceabilityResult(pincodeData, partnerInfo, pincodeToCheck)
	result.ResponseTime = time.Since(startTime)

	a.logger.WithFields(logrus.Fields{
		"component":      "dharmendra_adapter",
		"partner_code":   partnerInfo.PartnerCode,
		"partner_id":     partnerID,
		"pincode":        pincodeToCheck,
		"is_serviceable": len(result.Services) > 0,
		"response_time":  result.ResponseTime,
	}).Info("Dharmendra serviceability check completed")

	return result, nil
}

// validateRequirements validates Dharmendra specific requirements
func (a *Adapter) validateRequirements(request *models.ServiceabilityV2Request) error {
	// Dharmendra requires either destination postal code or generic postal code
	if (request.DestinationPostalCode == nil || *request.DestinationPostalCode == "") &&
		(request.PostalCode == nil || *request.PostalCode == "") {
		return fmt.Errorf("postal code is required for Dharmendra serviceability")
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
func (a *Adapter) convertToServiceabilityResult(pincodeData *repositories.DharmendraPincode, partnerInfo common.PartnerInfo, pincode string) *common.PartnerServiceabilityResult {
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
		"pincode":        pincode,
		"is_serviceable": true,
		"parcel_category": "ecomm",
		"city":           pincodeData.City,
		"state":          pincodeData.State,
		"hub_code":       pincodeData.HubCode,
		"hub_name":       pincodeData.HubName,
		"zone":           pincodeData.Zone,
		"cod_available":  pincodeData.COD,
		"pickup_available": pincodeData.FM,  // First Mile = Pickup
		"delivery_available": pincodeData.LM, // Last Mile = Delivery
		"fm":             pincodeData.FM,     // First Mile (for backward compatibility)
		"lm":             pincodeData.LM,     // Last Mile (for backward compatibility)
	}

	// Build delivery modes based on available options
	deliveryModes := make(map[string]bool)
	if pincodeData.Surface {
		deliveryModes["surface"] = true
	}
	if pincodeData.Air {
		deliveryModes["air"] = true
	}
	if pincodeData.Rail {
		deliveryModes["rail"] = true
	}
	if len(deliveryModes) == 0 {
		deliveryModes["standard"] = true
	}

	// Create service entry
	// FM (First Mile) maps to Pickup capability
	// LM (Last Mile) maps to Delivery capability
	service := models.ServiceV2{
		ServiceCode: "DHARMENDRA_STANDARD",
		ServiceName: "Dharmendra Standard",
		TATDays:     3, // Default TAT, can be enhanced based on zone/distance
		IsCOD:       pincodeData.COD,
		Pickup:      pincodeData.FM, // First Mile = Pickup available
		Delivery:    pincodeData.LM, // Last Mile = Delivery available
		Insurance:   true,            // Insurance available by default
		ProductTypes: map[string]bool{
			"ecommerce": true,
		},
		DeliveryModes: deliveryModes,
	}

	result.Services = append(result.Services, service)
	result.Capabilities = capabilities

	// Add metadata
	result.Metadata["pincode"] = pincode
	result.Metadata["is_serviceable"] = true
	result.Metadata["parcel_category"] = "ecomm"
	result.Metadata["district"] = pincodeData.DistrictName
	result.Metadata["to_pay"] = pincodeData.ToPay

	return result
}

// Initialize implements PartnerAdapter interface
func (a *Adapter) Initialize(ctx context.Context) error {
	a.logger.Info("Dharmendra adapter initialized successfully")
	return nil
}

// IsHealthy implements PartnerAdapter interface
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	if !a.config.Enabled {
		return false
	}
	// Simple health check - verify config is valid
	return a.config.TableName != ""
}

// GetMetrics implements PartnerAdapter interface
func (a *Adapter) GetMetrics() *common.PartnerMetrics {
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
func (a *Adapter) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down Dharmendra adapter")
	return nil
}

