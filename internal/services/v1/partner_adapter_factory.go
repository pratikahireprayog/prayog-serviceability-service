package services

import (
	"database/sql"
	"fmt"
	"net/http"

	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// partnerAdapterFactory implements PartnerAdapterFactory interface
type partnerAdapterFactory struct {
	config     config.PartnerAdaptersConfig
	httpClient *http.Client
	db         *sql.DB
	adapters   map[string]interfaces.PartnerAdapter
}

// NewPartnerAdapterFactory creates a new partner adapter factory
func NewPartnerAdapterFactory(cfg config.PartnerAdaptersConfig, httpClient *http.Client, db *sql.DB) interfaces.PartnerAdapterFactory {
	factory := &partnerAdapterFactory{
		config:     cfg,
		httpClient: httpClient,
		db:         db,
		adapters:   make(map[string]interfaces.PartnerAdapter),
	}

	// Initialize all supported adapters
	factory.initializeAdapters()

	return factory
}

// CreateAdapter creates a partner adapter for the given partner code
func (f *partnerAdapterFactory) CreateAdapter(partnerCode string) (interfaces.PartnerAdapter, error) {
	adapter, exists := f.adapters[partnerCode]
	if !exists {
		return nil, fmt.Errorf("unsupported partner code: %s", partnerCode)
	}

	return adapter, nil
}

// GetSupportedPartners returns a list of supported partner codes
func (f *partnerAdapterFactory) GetSupportedPartners() []string {
	var partners []string
	for partnerCode, adapter := range f.adapters {
		// Only include enabled adapters
		if f.isAdapterEnabled(partnerCode, adapter) {
			partners = append(partners, partnerCode)
		}
	}
	return partners
}

// IsPartnerSupported checks if a partner code is supported
func (f *partnerAdapterFactory) IsPartnerSupported(partnerCode string) bool {
	adapter, exists := f.adapters[partnerCode]
	if !exists {
		return false
	}

	return f.isAdapterEnabled(partnerCode, adapter)
}

// initializeAdapters initializes all partner adapters
func (f *partnerAdapterFactory) initializeAdapters() {
	// Initialize Shipyaari adapter
	if f.config.Shipyaari.Enabled {
		shipyaariAdapter := NewShipyaariAdapter(f.config.Shipyaari, f.httpClient)
		f.adapters["shipyaari"] = shipyaariAdapter
	}

	// Initialize Smile Courier adapter
	if f.config.SmileCourier.Enabled {
		smileCourierAdapter := NewSmileCourierAdapter(f.config.SmileCourier, f.httpClient)
		f.adapters["smile_courier"] = smileCourierAdapter
	}

	// Initialize Smile Ecom adapter
	if f.config.SmileEcom.Enabled {
		smileEcomAdapter := NewSmileEcomAdapter(f.config.SmileEcom, f.db)
		f.adapters["smile_ecom"] = smileEcomAdapter
	}
}

// isAdapterEnabled checks if a specific adapter is enabled and healthy
func (f *partnerAdapterFactory) isAdapterEnabled(partnerCode string, adapter interfaces.PartnerAdapter) bool {
	switch partnerCode {
	case "shipyaari":
		return f.config.Shipyaari.Enabled
	case "smile_courier":
		return f.config.SmileCourier.Enabled
	case "smile_ecom":
		return f.config.SmileEcom.Enabled
	default:
		return false
	}
}

// GetAdapter returns an existing adapter instance (for reuse)
func (f *partnerAdapterFactory) GetAdapter(partnerCode string) (interfaces.PartnerAdapter, bool) {
	adapter, exists := f.adapters[partnerCode]
	if !exists {
		return nil, false
	}

	return adapter, f.isAdapterEnabled(partnerCode, adapter)
}

// GetAllAdapters returns all enabled adapters
func (f *partnerAdapterFactory) GetAllAdapters() map[string]interfaces.PartnerAdapter {
	enabledAdapters := make(map[string]interfaces.PartnerAdapter)

	for partnerCode, adapter := range f.adapters {
		if f.isAdapterEnabled(partnerCode, adapter) {
			enabledAdapters[partnerCode] = adapter
		}
	}

	return enabledAdapters
}

// RefreshAdapters reinitializes adapters (useful for config updates)
func (f *partnerAdapterFactory) RefreshAdapters() {
	f.adapters = make(map[string]interfaces.PartnerAdapter)
	f.initializeAdapters()
}

// GetAdapterConfig returns configuration for a specific adapter
func (f *partnerAdapterFactory) GetAdapterConfig(partnerCode string) (interface{}, error) {
	switch partnerCode {
	case "shipyaari":
		return f.config.Shipyaari, nil
	case "smile_courier":
		return f.config.SmileCourier, nil
	case "smile_ecom":
		return f.config.SmileEcom, nil
	default:
		return nil, fmt.Errorf("unknown partner code: %s", partnerCode)
	}
}

// ValidateConfigurations validates all adapter configurations
func (f *partnerAdapterFactory) ValidateConfigurations() error {
	var errors []string

	// Validate Shipyaari config
	if f.config.Shipyaari.Enabled {
		if err := f.validateShipyaariConfig(); err != nil {
			errors = append(errors, fmt.Sprintf("Shipyaari: %v", err))
		}
	}

	// Validate Smile Courier config
	if f.config.SmileCourier.Enabled {
		if err := f.validateSmileCourierConfig(); err != nil {
			errors = append(errors, fmt.Sprintf("Smile Courier: %v", err))
		}
	}

	// Validate Smile Ecom config
	if f.config.SmileEcom.Enabled {
		if err := f.validateSmileEcomConfig(); err != nil {
			errors = append(errors, fmt.Sprintf("Smile Ecom: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %v", errors)
	}

	return nil
}

// validateShipyaariConfig validates Shipyaari adapter configuration
func (f *partnerAdapterFactory) validateShipyaariConfig() error {
	cfg := f.config.Shipyaari

	if cfg.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if cfg.Email == "" {
		return fmt.Errorf("email is required for authentication")
	}

	if cfg.Password == "" {
		return fmt.Errorf("password is required for authentication")
	}

	if cfg.TokenURL == "" {
		return fmt.Errorf("token URL is required")
	}

	if cfg.CheckServiceURL == "" {
		return fmt.Errorf("check service URL is required")
	}

	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// validateSmileCourierConfig validates Smile Courier adapter configuration
func (f *partnerAdapterFactory) validateSmileCourierConfig() error {
	cfg := f.config.SmileCourier

	if cfg.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if cfg.CheckServiceURL == "" {
		return fmt.Errorf("check service URL is required")
	}

	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// validateSmileEcomConfig validates Smile Ecom adapter configuration
func (f *partnerAdapterFactory) validateSmileEcomConfig() error {
	cfg := f.config.SmileEcom

	if cfg.TableName == "" {
		return fmt.Errorf("table name is required")
	}

	if cfg.CacheEnabled && cfg.CacheTTL <= 0 {
		return fmt.Errorf("cache TTL must be positive when cache is enabled")
	}

	return nil
}
