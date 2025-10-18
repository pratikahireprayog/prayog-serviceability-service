package factory

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	services "prayog-serviceability-service/internal/services/v1/data"
	"prayog-serviceability-service/internal/services/v2/partners/aramex"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/services/v2/partners/delcaper"
	"prayog-serviceability-service/internal/services/v2/partners/dhl"
	"prayog-serviceability-service/internal/services/v2/partners/fedex"
	"prayog-serviceability-service/internal/services/v2/partners/porter"
	"prayog-serviceability-service/internal/services/v2/partners/shipcube"
	"prayog-serviceability-service/internal/services/v2/partners/shipyaari"
	"prayog-serviceability-service/internal/services/v2/partners/smile_cargo"
	"prayog-serviceability-service/internal/services/v2/partners/smile_courier"
	"prayog-serviceability-service/internal/services/v2/partners/smile_ecom"
	"prayog-serviceability-service/internal/services/v2/partners/smile_hubops"
	"prayog-serviceability-service/internal/shared/config"

	"github.com/sirupsen/logrus"
)

// PartnerAdapterFactory defines the interface for creating partner adapters in v2
type PartnerAdapterFactory interface {
	CreateAdapter(partnerCode string) (common.PartnerAdapter, error)
	IsPartnerSupported(partnerCode string) bool
	GetSupportedPartners() []string
	GetAdapter(partnerCode string) (common.PartnerAdapter, bool)
	GetAllAdapters() map[string]common.PartnerAdapter
	RefreshAdapters()
	GetAdapterConfig(partnerCode string) (interface{}, error)
	ValidateConfigurations() error
}

// adapterImplementationMap maps database partner codes to their implementation codes
// This is the ONLY place where partner code mapping should be defined
var adapterImplementationMap = map[string]string{
	"smile_ecomm":       "smile_ecom",        // Database uses smile_ecomm, implementation is smile_ecom
	"smile_ecom":        "smile_ecom",        // Direct mapping
	"shipyaari":         "shipyaari",         // Direct mapping
	"smile_courier":     "smile_courier",     // Direct mapping
	"dhl":               "dhl",               // Direct mapping
	"smile_cargo":       "smile_cargo",       // Direct mapping
	"smile_hyperlocal":  "delcaper",          // Database uses smile_hyperlocal, implementation is delcaper
	"smile_hubops":      "smile_hubops",      // Direct mapping
	"porter":            "porter",            // Direct mapping
	"aramex":            "aramex",  
    "fedex":             "fedex",   
	"shipcube": 		 "shipcube",
	// Any partner code not in this map will get a generic adapter
}

// getImplementationCode returns the implementation code for a given database partner code
func getImplementationCode(dbPartnerCode string) string {
	if implCode, exists := adapterImplementationMap[dbPartnerCode]; exists {
		return implCode
	}
	return dbPartnerCode // Return original code if no mapping exists
}

// getPartnerDisplayName converts partner code to a human-readable display name
func getPartnerDisplayName(code string) string {
	// Simple name mapping for known partners
	nameMap := map[string]string{
		"dhl":           "DHL",
		"smile_cargo":   "Smile Cargo",
		"smile_ecomm":   "Smile Ecommerce",
		"smile_ecom":    "Smile Ecommerce",
		"shipyaari":     "Shipyaari",
		"smile_courier": "Smile Courier",
		"smile_hubops":  "Smile HubOps",
		"porter":        "Porter",
	}

	if name, exists := nameMap[code]; exists {
		return name
	}

	// Convert snake_case to Title Case for unknown codes
	parts := strings.Split(code, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

// partnerAdapterFactory implements PartnerAdapterFactory interface
type partnerAdapterFactory struct {
	config             config.PartnerAdaptersConfig
	httpClient         *http.Client
	db                 *sql.DB
	geolocationService services.GeolocationService
	hubLocationService services.HubLocationService
	implementations    map[string]common.PartnerAdapter // Implementation code -> adapter instance
	logger             *logrus.Logger
}

// NewPartnerAdapterFactory creates a new partner adapter factory
func NewPartnerAdapterFactory(cfg config.PartnerAdaptersConfig, httpClient *http.Client, db *sql.DB, geolocationService services.GeolocationService, hubLocationService services.HubLocationService) PartnerAdapterFactory {
	factory := &partnerAdapterFactory{
		config:             cfg,
		httpClient:         httpClient,
		db:                 db,
		geolocationService: geolocationService,
		hubLocationService: hubLocationService,
		implementations:    make(map[string]common.PartnerAdapter),
		logger:             logrus.New(),
	}

	// Initialize all available adapter implementations
	factory.initializeImplementations()

	return factory
}

// CreateAdapter creates a partner adapter for the given partner code (from database)
func (f *partnerAdapterFactory) CreateAdapter(dbPartnerCode string) (common.PartnerAdapter, error) {
	// Get the implementation code for this database partner code
	implCode := getImplementationCode(dbPartnerCode)

	// Check if we have a specific implementation for this implementation code
	if adapter, exists := f.implementations[implCode]; exists {
		// For mapped codes (like smile_ecomm -> smile_ecom), we need to return an adapter
		// that reports the database partner code, not the implementation code
		if dbPartnerCode != implCode {
			return common.NewPartnerCodeAdapter(common.PartnerInfo{PartnerCode: dbPartnerCode}, adapter), nil
		}
		return adapter, nil
	}

	// For unknown partner codes, create a generic adapter
	return common.NewGenericAdapter(dbPartnerCode, getPartnerDisplayName(dbPartnerCode)), nil
}

// GetSupportedPartners returns ALL partner codes (this should be called with database codes)
// The orchestrator will provide the database partner codes, and this factory handles the mapping
func (f *partnerAdapterFactory) GetSupportedPartners() []string {
	var partners []string

	// Add all implementation codes that have enabled configurations
	for implCode, adapter := range f.implementations {
		if f.isImplementationEnabled(implCode, adapter) {
			partners = append(partners, implCode)
		}
	}

	// Debug logging to see what's happening
	f.logger.WithFields(logrus.Fields{
		"component":           "partner_adapter_factory",
		"implementations":     f.implementations,
		"supported_partners":  partners,
	}).Info("Factory supported partners retrieved")

	// NOTE: This method should ideally be called with the actual database partner codes
	// For now, we return implementation codes. The orchestrator should filter based on
	// actual database codes and check if they can be mapped to implementations.

	return partners
}

// IsPartnerSupported checks if a database partner code can be supported
func (f *partnerAdapterFactory) IsPartnerSupported(dbPartnerCode string) bool {
	implCode := getImplementationCode(dbPartnerCode)

	if adapter, exists := f.implementations[implCode]; exists {
		return f.isImplementationEnabled(implCode, adapter)
	}

	// Generic adapters are always "supported" but not functional
	return true
}

// GetAdapter returns an adapter for the given partner code
func (f *partnerAdapterFactory) GetAdapter(dbPartnerCode string) (common.PartnerAdapter, bool) {
	adapter, err := f.CreateAdapter(dbPartnerCode)
	if err != nil {
		return nil, false
	}
	return adapter, true
}

// GetAllAdapters returns all available adapters
func (f *partnerAdapterFactory) GetAllAdapters() map[string]common.PartnerAdapter {
	result := make(map[string]common.PartnerAdapter)

	// Include all implementation-based adapters
	for implCode, adapter := range f.implementations {
		result[implCode] = adapter
	}

	// Include any mapped database codes
	for dbCode, implCode := range adapterImplementationMap {
		if adapter, exists := f.implementations[implCode]; exists && dbCode != implCode {
			result[dbCode] = common.NewPartnerCodeAdapter(common.PartnerInfo{PartnerCode: dbCode}, adapter)
		}
	}

	return result
}

// RefreshAdapters refreshes all adapter configurations
func (f *partnerAdapterFactory) RefreshAdapters() {
	f.implementations = make(map[string]common.PartnerAdapter)
	f.initializeImplementations()
}

// GetAdapterConfig returns configuration for a specific adapter
func (f *partnerAdapterFactory) GetAdapterConfig(dbPartnerCode string) (interface{}, error) {
	implCode := getImplementationCode(dbPartnerCode)

	switch implCode {
	case "shipyaari":
		return f.config.Shipyaari, nil
	case "smile_courier":
		return f.config.SmileCourier, nil
	case "smile_ecom":
		return f.config.SmileEcom, nil
	case "dhl":
		return f.config.DHL, nil
	case "smile_cargo":
		return f.config.SmileCargo, nil
	case "delcaper":
		return f.config.Delcaper, nil
	case "smile_hubops":
		return f.config.SmileHubOps, nil
	case "porter":
		return f.config.Porter, nil
	case "aramex":                           
        return f.config.Aramex, nil
    case "fedex":                              
        return f.config.FedEx, nil
	case "shipcube":                              
        return f.config.ShipCube, nil
	default:
		return nil, fmt.Errorf("no configuration found for partner: %s", dbPartnerCode)
	}
}

// ValidateConfigurations validates all adapter configurations
func (f *partnerAdapterFactory) ValidateConfigurations() error {
	if err := f.validateShipyaariConfig(); err != nil {
		return fmt.Errorf("shipyaari config validation failed: %w", err)
	}

	if err := f.validateSmileCourierConfig(); err != nil {
		return fmt.Errorf("smile_courier config validation failed: %w", err)
	}

	if err := f.validateSmileEcomConfig(); err != nil {
		return fmt.Errorf("smile_ecom config validation failed: %w", err)
	}

	return nil
}

// initializeImplementations initializes all available adapter implementations
func (f *partnerAdapterFactory) initializeImplementations() {
	f.logger.WithFields(logrus.Fields{
		"component": "partner_adapter_factory",
		"action":    "initialize_implementations",
	}).Info("Starting adapter implementations initialization")

	f.logger.WithFields(logrus.Fields{
		"component": "partner_adapter_factory",
		"delcaper_enabled": f.config.Delcaper.Enabled,
	}).Info("Delcaper configuration status")
	
	// Initialize Shipyaari adapter with DB precheck implementation
	if f.config.Shipyaari.Enabled {
		f.implementations["shipyaari"] = shipyaari.NewShipyaariAdapter(f.config.Shipyaari, f.db)
		f.logger.WithFields(logrus.Fields{
			"component": "partner_adapter_factory",
			"adapter":   "shipyaari",
		}).Info("Initialized shipyaari adapter")
	}

	// Initialize Smile Courier adapter with real implementation
	if f.config.SmileCourier.Enabled {
		f.implementations["smile_courier"] = smile_courier.NewSmileCourierAdapter(f.config.SmileCourier)
		f.logger.WithFields(logrus.Fields{
			"component": "partner_adapter_factory",
			"adapter":   "smile_courier",
		}).Info("Initialized smile_courier adapter")
	}

	// Initialize Smile Ecom adapter with real implementation
	if f.config.SmileEcom.Enabled {
		f.implementations["smile_ecom"] = smile_ecom.NewSmileEcomAdapter(f.config.SmileEcom, f.db)
		f.logger.WithFields(logrus.Fields{
			"component": "partner_adapter_factory",
			"adapter":   "smile_ecom",
		}).Info("Initialized smile_ecom adapter")
	}

	// Initialize DHL adapter with real implementation and geolocation service
	if f.config.DHL.Enabled {
		f.implementations["dhl"] = dhl.NewAdapter(f.config.DHL, f.geolocationService, f.hubLocationService)
		f.logger.WithFields(logrus.Fields{
			"component": "partner_adapter_factory",
			"adapter":   "dhl",
		}).Info("Initialized dhl adapter")
	}

	// Enable Smile Cargo adapter when configuration is available
	if f.config.SmileCargo.Enabled {
		f.implementations["smile_cargo"] = smile_cargo.NewAdapter(f.config.SmileCargo)
		f.logger.WithFields(logrus.Fields{
			"component": "partner_adapter_factory",
			"adapter":   "smile_cargo",
		}).Info("Initialized smile_cargo adapter")
	}

	// Initialize Delcaper adapter for hyperlocal serviceability
	if f.config.Delcaper.Enabled {
		f.implementations["delcaper"] = delcaper.NewAdapter(f.config.Delcaper)
		f.logger.WithFields(logrus.Fields{
			"component": "partner_adapter_factory",
			"adapter":   "delcaper",
		}).Info("Initialized delcaper adapter")
	} else {
		f.logger.WithFields(logrus.Fields{
			"component": "partner_adapter_factory",
			"adapter":   "delcaper",
		}).Warn("Delcaper adapter not enabled in config")
	}

	// Initialize Smile HubOps adapter for hub operations
	if f.config.SmileHubOps.Enabled {
		f.implementations["smile_hubops"] = smile_hubops.NewAdapter(f.config.SmileHubOps)
		f.logger.WithFields(logrus.Fields{
			"component": "partner_adapter_factory",
			"adapter":   "smile_hubops",
		}).Info("Initialized smile_hubops adapter")
	} else {
		f.logger.WithFields(logrus.Fields{
			"component": "partner_adapter_factory",
			"adapter":   "smile_hubops",
		}).Warn("Smile HubOps adapter not enabled in config")
	}

	// Initialize Porter adapter for database-based serviceability
	if f.config.Porter.Enabled {
		f.implementations["porter"] = porter.NewPorterAdapter(f.config.Porter, f.db)
		f.logger.WithFields(logrus.Fields{
			"component": "partner_adapter_factory",
			"adapter":   "porter",
		}).Info("Initialized porter adapter")
	} else {
		f.logger.WithFields(logrus.Fields{
			"component": "partner_adapter_factory",
			"adapter":   "porter",
		}).Warn("Porter adapter not enabled in config")
	}

	// Initialize Aramex adapter
    if f.config.Aramex.Enabled {
        f.implementations["aramex"] = aramex.NewAdapter(
			f.config.Aramex, 
			f.geolocationService, 
			f.hubLocationService) 
        f.logger.WithFields(logrus.Fields{
            "component": "partner_adapter_factory",
            "adapter":   "aramex",
        }).Info("Initialized aramex adapter")
    } else {
        f.logger.WithFields(logrus.Fields{
            "component": "partner_adapter_factory",
            "adapter":   "aramex",
        }).Warn("Aramex adapter not enabled in config")
    }

    // Initialize FedEx adapter
    if f.config.FedEx.Enabled {
        f.implementations["fedex"] = fedex.NewAdapter(
			f.config.FedEx,
			f.geolocationService,
			f.hubLocationService)

 
        f.logger.WithFields(logrus.Fields{
            "component": "partner_adapter_factory",
            "adapter":   "fedex",
        }).Info("Initialized fedex adapter")
    } else {
        f.logger.WithFields(logrus.Fields{
            "component": "partner_adapter_factory",
            "adapter":   "fedex",
        }).Warn("FedEx adapter not enabled in config")
    }

	if f.config.ShipCube.Enabled {
        f.implementations["shipcube"] = shipcube.NewAdapter(
			f.config.ShipCube,
		)

        f.logger.WithFields(logrus.Fields{
            "component": "partner_adapter_factory",
            "adapter":   "shipcube",
        }).Info("Initialized fedex adapter")
    } else {
        f.logger.WithFields(logrus.Fields{
            "component": "partner_adapter_factory",
            "adapter":   "shipcube",
        }).Warn("shipcube adapter not enabled in config")
    }
	
	f.logger.WithFields(logrus.Fields{
		"component":      "partner_adapter_factory",
		"implementations": f.implementations,
	}).Info("Adapter implementations initialization completed")
}

// isImplementationEnabled checks if an implementation is enabled
func (f *partnerAdapterFactory) isImplementationEnabled(implCode string, adapter common.PartnerAdapter) bool {
	// Check if the adapter is healthy and enabled
	ctx := context.Background()
	isHealthy := adapter.IsHealthy(ctx)
	f.logger.WithFields(logrus.Fields{
		"component": "partner_adapter_factory",
		"adapter":   implCode,
		"is_healthy": isHealthy,
	}).Info("Checking adapter health status")
	return isHealthy
}

// Helper validation methods (simplified for now)
func (f *partnerAdapterFactory) validateShipyaariConfig() error {
	if !f.config.Shipyaari.Enabled {
		return nil
	}
	// Add specific validation logic for Shipyaari
	return nil
}

func (f *partnerAdapterFactory) validateSmileCourierConfig() error {
	if !f.config.SmileCourier.Enabled {
		return nil
	}
	// Add specific validation logic for Smile Courier
	return nil
}

func (f *partnerAdapterFactory) validateSmileEcomConfig() error {
	if !f.config.SmileEcom.Enabled {
		return nil
	}
	// Add specific validation logic for Smile Ecom
	return nil
}
