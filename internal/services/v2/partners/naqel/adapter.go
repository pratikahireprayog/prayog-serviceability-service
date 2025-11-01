package naqel

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Adapter implements the PartnerAdapter interface for Naqel shipping
type Adapter struct {
	client     *NaqelClient
	config     config.NaqelConfig
	repository repositories.NaqelRepository
	logger     *logrus.Logger
}

// NewAdapter creates a new Naqel adapter instance
func NewAdapter(config config.NaqelConfig, db *sql.DB) *Adapter {
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
	var naqelRepo repositories.NaqelRepository
	if gormDB != nil {
		naqelRepo = repositories.NewNaqelRepository(gormDB, config.TableName)
	} else {
		logger.Warn("No database connection available for Naqel repository")
	}

	logger.WithFields(logrus.Fields{
		"partner":    "Naqel",
		"base_url":   config.BaseURL,
		"client_id":  redactCredentialField(config.ClientID),
		"table_name": config.TableName,
		"enabled":    config.Enabled,
	}).Info("Creating Naqel adapter")

	return &Adapter{
		client:     NewNaqelClient(config),
		config:     config,
		repository: naqelRepo,
		logger:     logger,
	}
}

// redactCredentialField safely redacts sensitive fields for logging
func redactCredentialField(field string) string {
	if field == "" {
		return "[EMPTY]"
	}
	if len(field) <= 3 {
		return "[REDACTED]"
	}
	return field[:3] + "***"
}

// CheckServiceability checks if Naqel can service the given request
func (a *Adapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	partnerID := ""
	if partnerInfo.PartnerID != nil {
		partnerID = partnerInfo.PartnerID.String()
	}

	a.logger.WithFields(logrus.Fields{
		"component":    "naqel_adapter",
		"action":       "check_serviceability_start",
		"partner_code": partnerInfo.PartnerCode,
		"partner_id":   partnerID,
	}).Info("Starting Naqel adapter serviceability check")

	// Validate Naqel-specific requirements
	if err := a.validateNaqelRequirements(request); err != nil {
		a.logger.WithFields(logrus.Fields{
			"component":    "naqel_adapter",
			"event":        "validation_failed",
			"partner_code": partnerInfo.PartnerCode,
			"partner_id":   partnerID,
			"error":        err.Error(),
		}).Warn("Naqel validation failed")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			Services:     make([]models.ServiceV2, 0),
			ResponseTime: time.Since(startTime),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Naqel validation failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "Naqel validation failed",
			},
		}, nil
	}

	// Extract postal codes and country codes
	sourcePostalCode := ""
	if request.SourcePostalCode != nil {
		sourcePostalCode = *request.SourcePostalCode
	}

	destinationPostalCode := ""
	if request.DestinationPostalCode != nil {
		destinationPostalCode = *request.DestinationPostalCode
	}

	sourceCountryCode := ""
	if request.SourceCountryCode != nil && *request.SourceCountryCode != "" {
		sourceCountryCode = *request.SourceCountryCode
	} else if request.CountryCode != nil && *request.CountryCode != "" {
		sourceCountryCode = *request.CountryCode
	}

	destinationCountryCode := ""
	if request.DestinationCountryCode != nil && *request.DestinationCountryCode != "" {
		destinationCountryCode = *request.DestinationCountryCode
	} else if request.CountryCode != nil && *request.CountryCode != "" {
		destinationCountryCode = *request.CountryCode
	}

	a.logger.WithFields(logrus.Fields{
		"component":                "naqel_adapter",
		"partner_code":             partnerInfo.PartnerCode,
		"source_postal_code":       sourcePostalCode,
		"destination_postal_code":  destinationPostalCode,
		"source_country_code":      sourceCountryCode,
		"destination_country_code": destinationCountryCode,
	}).Info("Resolving city codes from postal codes")

	// Check if repository is available
	if a.repository == nil {
		errMsg := "Naqel repository not available"
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

	a.logger.WithFields(logrus.Fields{
		"component":              "naqel_adapter",
		"source_city_code":       sourcePostalCode,
		"destination_city_code":  destinationPostalCode,
	}).Info("City codes resolved, checking serviceability in database")

	// Check serviceability in database
	sourceLocation, destLocation, err := a.repository.CheckServiceabilityByCityCodes(
		ctx, sourcePostalCode, destinationPostalCode, sourceCountryCode, destinationCountryCode,
	)

	
	if err != nil {
		errMsg := fmt.Sprintf("Serviceability check failed: %v", err)
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			ResponseTime: time.Since(startTime),
			ErrorMessage: &errMsg,
			Metadata: map[string]interface{}{
				"reason":             "Database serviceability check failed",
				"source_city_code":   sourcePostalCode,
				"dest_city_code":     destinationPostalCode,
			},
		}, nil
	}

	// Call SOAP API to get transit days
	transitDays, err := a.client.GetTransitDays(ctx, sourcePostalCode, destinationPostalCode, sourceLocation.StationCode, destLocation.StationCode)
	if err != nil {
		a.logger.WithError(err).WithFields(logrus.Fields{
			"component":    "naqel_adapter",
			"partner_code": partnerInfo.PartnerCode,
		}).Warn("Failed to get transit days from Naqel API, but location is serviceable")

		// Location is serviceable, but transit days unavailable - use default
		transitDays = 7 // Default transit days
	}

	// Build service response
	service := models.ServiceV2{
		ServiceCode: "NAQEL_EXPRESS",
		ServiceName: "Naqel Express",
		TATDays:     transitDays,
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

	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		Services:     []models.ServiceV2{service},
		ResponseTime: time.Since(startTime),
		Metadata: map[string]interface{}{
			"flow":                     "international",
			"source_country_code":      sourceCountryCode,
			"destination_country_code": destinationCountryCode,
			"source_city_code":         sourcePostalCode,
			"destination_city_code":    destinationPostalCode,
			"source_location":          sourceLocation.LocationEn,
			"destination_location":     destLocation.LocationEn,
			"transit_days":             transitDays,
		},
	}

	a.logger.WithFields(logrus.Fields{
		"component":                "naqel_adapter",
		"partner_code":             partnerInfo.PartnerCode,
		"source_city_code":         sourcePostalCode,
		"destination_city_code":    destinationPostalCode,
		"transit_days":             transitDays,
		"is_serviceable":           true,
	}).Info("Naqel serviceability check completed successfully")

	return result, nil
}

// validateNaqelRequirements validates Naqel-specific requirements
func (a *Adapter) validateNaqelRequirements(request *models.ServiceabilityV2Request) error {
	if request.SourcePostalCode == nil || *request.SourcePostalCode == "" {
		return fmt.Errorf("source postal code is required for Naqel shipments")
	}

	if request.DestinationPostalCode == nil || *request.DestinationPostalCode == "" {
		return fmt.Errorf("destination postal code is required for Naqel shipments")
	}

	// Validate country codes
	if request.SourceCountryCode == nil || *request.SourceCountryCode == "" {
		if request.CountryCode == nil || *request.CountryCode == "" {
			return fmt.Errorf("country code is required for Naqel shipments")
		}
	}

	return nil
}

// Initialize implements PartnerAdapter interface
func (a *Adapter) Initialize(ctx context.Context) error {
	a.logger.WithFields(logrus.Fields{
		"partner": "Naqel",
	}).Info("Initializing Naqel adapter")
	return nil
}

// IsHealthy implements PartnerAdapter interface
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	if !a.config.Enabled {
		return false
	}
	// Simple health check - verify config is valid
	return a.config.BaseURL != "" && a.config.ClientID != ""
}

// GetMetrics implements PartnerAdapter interface
func (a *Adapter) GetMetrics() *common.PartnerMetrics {
	return &common.PartnerMetrics{
		PartnerCode:         "",
		TotalRequests:       0,
		SuccessfulRequests:  0,
		FailedRequests:      0,
		AverageResponseTime: 0,
		HealthStatus:        "healthy",
		ErrorRate:           0.0,
	}
}

// Shutdown implements PartnerAdapter interface
func (a *Adapter) Shutdown(ctx context.Context) error {
	a.logger.Info("Naqel adapter shutdown completed")
	return nil
}

