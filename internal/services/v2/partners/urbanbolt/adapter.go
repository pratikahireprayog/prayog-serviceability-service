package urbanbolt

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// UrbanBoltAdapter implements the common.PartnerAdapter interface for UrbanBolt
type UrbanBoltAdapter struct {
	repository repositories.UrbanboltRepository
	config     config.UrbanBoltConfig
	logger     *logrus.Logger
}

// NewUrbanBoltAdapter creates a new UrbanBolt adapter
func NewUrbanBoltAdapter(cfg config.UrbanBoltConfig, db *sql.DB) common.PartnerAdapter {
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
	var repo repositories.UrbanboltRepository
	if gormDB != nil {
		repo = repositories.NewUrbanboltRepository(gormDB, cfg.TableName)
	} else {
		logger.Warn("No database connection available for UrbanBolt repository")
	}

	logger.WithFields(logrus.Fields{
		"partner":    "UrbanBolt",
		"table_name": cfg.TableName,
		"enabled":    cfg.Enabled,
	}).Info("Creating UrbanBolt adapter")

	return &UrbanBoltAdapter{
		repository: repo,
		config:     cfg,
		logger:     logger,
	}
}

// GetAdapterType returns the adapter type
func (u *UrbanBoltAdapter) GetAdapterType() common.AdapterType {
	return common.AdapterTypeDatabase
}

// IsEnabled returns whether the adapter is enabled
func (u *UrbanBoltAdapter) IsEnabled() bool {
	return u.config.Enabled
}

// CheckServiceability checks if UrbanBolt can service the given request
func (u *UrbanBoltAdapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	partnerID := ""
	if partnerInfo.PartnerID != nil {
		partnerID = partnerInfo.PartnerID.String()
	}

	u.logger.WithFields(logrus.Fields{
		"component":    "urbanbolt_adapter",
		"action":        "check_serviceability_start",
		"partner_code":  partnerInfo.PartnerCode,
		"partner_id":    partnerID,
	}).Info("Starting UrbanBolt adapter serviceability check")

	// Validate requirements
	if err := u.validateRequirements(request); err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("UrbanBolt validation failed: %v", err)}[0],
		}, nil
	}

	// Check if repository is available
	if u.repository == nil {
		errMsg := "UrbanBolt repository not available"
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			ErrorMessage: &errMsg,
		}, nil
	}

	// Get source/dest pincodes
	sourcePincode := getSourcePincode(request)
	destPincode := getDestinationPincode(request)

	// Check Source Pincode
	sourceData, err := u.repository.CheckServiceabilityByPincode(ctx, sourcePincode)
	if err != nil {
		return u.createNonServiceableResult(partnerInfo, startTime, fmt.Sprintf("Source pincode %s not serviceable active", sourcePincode)), nil
	}

	// Check Destination Pincode
	destData, err := u.repository.CheckServiceabilityByPincode(ctx, destPincode)
	if err != nil {
		return u.createNonServiceableResult(partnerInfo, startTime, fmt.Sprintf("Destination pincode %s not serviceable active", destPincode)), nil
	}

	// Check Logic: Source Outbound && Dest Inbound
	if !sourceData.Outbound {
		return u.createNonServiceableResult(partnerInfo, startTime, fmt.Sprintf("Source pincode %s does not support outbound", sourcePincode)), nil
	}
	if !destData.Inbound {
		return u.createNonServiceableResult(partnerInfo, startTime, fmt.Sprintf("Destination pincode %s does not support inbound", destPincode)), nil
	}

	// Make sure Service Types match? Or just take the intersection?
	// User data has service_type column "SDD,NDD".
	// We can compute common service types or just list all potential services and let the system filter.
	// For now, let's use the Destination services as the main driver, checked against Source support?
	// Actually, usually the Route or Destination defines the service type (e.g. NDD to that dest).
	// We will follow the logic of converting the data to result.

	result := u.convertToServiceabilityResult(sourceData, destData, partnerInfo)
	result.ResponseTime = time.Since(startTime)
	
	u.logger.WithFields(logrus.Fields{
		"component":      "urbanbolt_adapter",
		"partner_code":   partnerInfo.PartnerCode,
		"pincode_src":    sourcePincode,
		"pincode_dst":    destPincode,
		"is_serviceable": len(result.Services) > 0,
	}).Info("UrbanBolt serviceability check completed")

	return result, nil
}

func (u *UrbanBoltAdapter) createNonServiceableResult(partnerInfo common.PartnerInfo, startTime time.Time, reason string) *common.PartnerServiceabilityResult {
	return &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		ResponseTime: time.Since(startTime),
		Metadata: map[string]interface{}{
			"reason": reason,
		},
	}
}

// validateRequirements validates specific requirements
func (u *UrbanBoltAdapter) validateRequirements(request *models.ServiceabilityV2Request) error {
	src := getSourcePincode(request)
	dst := getDestinationPincode(request)
	
	if src == "" || dst == "" {
		return fmt.Errorf("both source and destination pincodes are required")
	}

	if len(src) != 6 || len(dst) != 6 {
		return fmt.Errorf("pincodes must be 6 digits")
	}

	return nil
}

// convertToServiceabilityResult converts database result to common format
func (u *UrbanBoltAdapter) convertToServiceabilityResult(sourceData, destData *repositories.UrbanboltPincode, partnerInfo common.PartnerInfo) *common.PartnerServiceabilityResult {
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}

	// Capabilities
	result.Capabilities["is_serviceable"] = true
	result.Capabilities["source_pincode"] = sourceData.Pincode
	result.Capabilities["dest_pincode"] = destData.Pincode
	result.Capabilities["pickup_available"] = sourceData.Outbound
	result.Capabilities["delivery_available"] = destData.Inbound
	result.Capabilities["rtn_available"] = sourceData.RTN || destData.RTN // Logical OR or AND? Usually if one supports it? Let's say dest supports RTN pickup?
    // Actually RTN usually means Return to Origin is supported.
	
	// Parse Service Types from Destination (and maybe Source intersection)
	// Example: "SDD,NDD"
	// We'll simplisticly take Destination's service types as the services offered to that destination.
	serviceTypes := strings.Split(destData.ServiceType, ",")
	for _, st := range serviceTypes {
		st = strings.TrimSpace(st)
		if st == "" { continue }
		
		// Map 'SDD' -> Standard? 'NDD' -> Next Day?
		// We can just create services with these codes.
		
		deliveryMode := "standard"
		if st == "SDD" || st == "NDD" || st == "2HR" { // NDD is fast, SDD is Same Day (fast)
			deliveryMode = "express"
		}

		svc := models.ServiceV2{
			ServiceCode:   "URBANBOLT_" + st,
			ServiceName:   "UrbanBolt " + st,
			Pickup:        sourceData.Outbound,
			Delivery:      destData.Inbound,
			DeliveryModes: map[string]bool{deliveryMode: true},
			ProductTypes:  map[string]bool{"general": true},
			TATDays: 2, // Default
		}
		
		if st == "SDD" { svc.TATDays = 0 }
		if st == "NDD" { svc.TATDays = 1 }
		
		result.Services = append(result.Services, svc)
	}
	
	result.Metadata["route_code"] = destData.RouteCode
	result.Metadata["zone"] = destData.Zone
	
	return result
}

// Initialize
func (u *UrbanBoltAdapter) Initialize(ctx context.Context) error {
	return nil
}

// IsHealthy
func (u *UrbanBoltAdapter) IsHealthy(ctx context.Context) bool {
	return u.config.Enabled && u.repository != nil
}

// GetMetrics
func (u *UrbanBoltAdapter) GetMetrics() *common.PartnerMetrics {
	return &common.PartnerMetrics{
		PartnerCode:  "urbanbolt",
		HealthStatus: "healthy",
	}
}

// Shutdown
func (u *UrbanBoltAdapter) Shutdown(ctx context.Context) error {
	return nil
}


// Helpers (copied/adapted from previous file)
func getSourcePincode(req *models.ServiceabilityV2Request) string {
	if req.SourcePostalCode != nil {
		return *req.SourcePostalCode
	}
	if req.PostalCode != nil {
		return *req.PostalCode // Fallback if applicable
	}
	return ""
}

func getDestinationPincode(req *models.ServiceabilityV2Request) string {
	if req.DestinationPostalCode != nil {
		return *req.DestinationPostalCode
	}
	return ""
}
