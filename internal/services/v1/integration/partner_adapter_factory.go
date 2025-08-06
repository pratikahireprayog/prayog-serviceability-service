package services

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// adapterImplementationMap maps database partner codes to their implementation codes
// This is the ONLY place where partner code mapping should be defined
var adapterImplementationMap = map[string]string{
	"smile_ecomm":       "smile_ecom",        // Database uses smile_ecomm, implementation is smile_ecom
	"smile_ecom":        "smile_ecom",        // Direct mapping
	"shipyaari":         "shipyaari",         // Direct mapping
	"smile_courier":     "smile_courier",     // Direct mapping
	"smile_hyperlocal":  "delcaper",          // Database uses smile_hyperlocal, implementation is delcaper
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

// genericPartnerAdapter is a fallback adapter for partners without specific implementations
type genericPartnerAdapter struct {
	partnerCode string
	partnerName string
}

func (g *genericPartnerAdapter) GetPartnerCode() string {
	return g.partnerCode
}

func (g *genericPartnerAdapter) GetPartnerName() string {
	return g.partnerName
}

func (g *genericPartnerAdapter) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request) (*interfaces.PartnerServiceabilityResult, error) {
	errorMsg := fmt.Sprintf("%s adapter implementation pending", g.partnerName)
	return &interfaces.PartnerServiceabilityResult{
		PartnerCode:   g.partnerCode,
		PartnerName:   g.partnerName,
		IsServiceable: false,
		Services:      []models.ServiceV2{},
		Capabilities: map[string]interface{}{
			"status":  "implementation_pending",
			"message": errorMsg,
		},
		ResponseTime: time.Millisecond * 10,
		Rating:       0.0,
		ErrorMessage: &errorMsg,
		Error:        fmt.Errorf(errorMsg),
	}, nil
}

func (g *genericPartnerAdapter) IsHealthy(ctx context.Context) bool {
	return false // Generic adapters are not healthy by design
}

func (g *genericPartnerAdapter) GetAdapterType() interfaces.AdapterType {
	return interfaces.AdapterTypeHTTP
}

// partnerAdapterFactory implements PartnerAdapterFactory interface
type partnerAdapterFactory struct {
	config          config.PartnerAdaptersConfig
	httpClient      *http.Client
	db              *sql.DB
	implementations map[string]interfaces.PartnerAdapter // Implementation code -> adapter instance
}

// NewPartnerAdapterFactory creates a new partner adapter factory
func NewPartnerAdapterFactory(cfg config.PartnerAdaptersConfig, httpClient *http.Client, db *sql.DB) interfaces.PartnerAdapterFactory {
	factory := &partnerAdapterFactory{
		config:          cfg,
		httpClient:      httpClient,
		db:              db,
		implementations: make(map[string]interfaces.PartnerAdapter),
	}

	// Initialize all available adapter implementations
	factory.initializeImplementations()

	return factory
}

// CreateAdapter creates a partner adapter for the given partner code (from database)
func (f *partnerAdapterFactory) CreateAdapter(dbPartnerCode string) (interfaces.PartnerAdapter, error) {
	// Get the implementation code for this database partner code
	implCode := getImplementationCode(dbPartnerCode)

	// Check if we have a specific implementation for this implementation code
	if adapter, exists := f.implementations[implCode]; exists {
		// For mapped codes (like smile_ecomm -> smile_ecom), we need to return an adapter
		// that reports the database partner code, not the implementation code
		if dbPartnerCode != implCode {
			return &partnerCodeAdapter{
				baseAdapter: adapter,
				partnerCode: dbPartnerCode,
				partnerName: getPartnerDisplayName(dbPartnerCode),
			}, nil
		}
		return adapter, nil
	}

	// For unknown partner codes, create a generic adapter
	return &genericPartnerAdapter{
		partnerCode: dbPartnerCode,
		partnerName: getPartnerDisplayName(dbPartnerCode),
	}, nil
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

// initializeImplementations initializes all available adapter implementations
func (f *partnerAdapterFactory) initializeImplementations() {
	// Initialize Shipyaari adapter
	if f.config.Shipyaari.Enabled {
		shipyaariAdapter := NewShipyaariAdapter(f.config.Shipyaari, f.httpClient)
		f.implementations["shipyaari"] = shipyaariAdapter
	}

	// Initialize Smile Courier adapter
	if f.config.SmileCourier.Enabled {
		smileCourierAdapter := NewSmileCourierAdapter(f.config.SmileCourier, f.httpClient)
		f.implementations["smile_courier"] = smileCourierAdapter
	}

	// Initialize Smile Ecom adapter
	if f.config.SmileEcom.Enabled {
		smileEcomAdapter := NewSmileEcomAdapter(f.config.SmileEcom, f.db)
		f.implementations["smile_ecom"] = smileEcomAdapter
	}
}

// isImplementationEnabled checks if a specific implementation is enabled
func (f *partnerAdapterFactory) isImplementationEnabled(implCode string, adapter interfaces.PartnerAdapter) bool {
	switch implCode {
	case "shipyaari":
		return f.config.Shipyaari.Enabled
	case "smile_courier":
		return f.config.SmileCourier.Enabled
	case "smile_ecom":
		return f.config.SmileEcom.Enabled
	default:
		// Generic adapters are always "enabled" but return errors
		return true
	}
}

// partnerCodeAdapter wraps an implementation adapter to return the database partner code
type partnerCodeAdapter struct {
	baseAdapter interfaces.PartnerAdapter
	partnerCode string
	partnerName string
}

func (p *partnerCodeAdapter) GetPartnerCode() string {
	return p.partnerCode // Return database partner code
}

func (p *partnerCodeAdapter) GetPartnerName() string {
	return p.partnerName // Return database partner display name
}

func (p *partnerCodeAdapter) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request) (*interfaces.PartnerServiceabilityResult, error) {
	// Delegate to base adapter but ensure the response has the correct partner code
	result, err := p.baseAdapter.CheckServiceability(ctx, req)
	if err != nil {
		return nil, err
	}

	// Override the partner code and name in the result to match database values
	result.PartnerCode = p.partnerCode
	result.PartnerName = p.partnerName

	return result, nil
}

func (p *partnerCodeAdapter) IsHealthy(ctx context.Context) bool {
	return p.baseAdapter.IsHealthy(ctx)
}

func (p *partnerCodeAdapter) GetAdapterType() interfaces.AdapterType {
	return p.baseAdapter.GetAdapterType()
}

// GetAdapter returns an existing adapter instance (for reuse)
func (f *partnerAdapterFactory) GetAdapter(dbPartnerCode string) (interfaces.PartnerAdapter, bool) {
	implCode := getImplementationCode(dbPartnerCode)

	if adapter, exists := f.implementations[implCode]; exists {
		isEnabled := f.isImplementationEnabled(implCode, adapter)
		if !isEnabled {
			return nil, false
		}

		// Wrap with code adapter if needed
		if dbPartnerCode != implCode {
			return &partnerCodeAdapter{
				baseAdapter: adapter,
				partnerCode: dbPartnerCode,
				partnerName: getPartnerDisplayName(dbPartnerCode),
			}, true
		}

		return adapter, true
	}

	return nil, false
}

// GetAllAdapters returns all enabled adapters mapped by database partner codes
func (f *partnerAdapterFactory) GetAllAdapters() map[string]interfaces.PartnerAdapter {
	enabledAdapters := make(map[string]interfaces.PartnerAdapter)

	// Add implementation adapters with their direct codes
	for implCode, adapter := range f.implementations {
		if f.isImplementationEnabled(implCode, adapter) {
			enabledAdapters[implCode] = adapter
		}
	}

	// Add mapped codes (like smile_ecomm that maps to smile_ecom implementation)
	for dbCode, implCode := range adapterImplementationMap {
		if dbCode != implCode { // Only for mapped codes
			if adapter, exists := f.implementations[implCode]; exists && f.isImplementationEnabled(implCode, adapter) {
				enabledAdapters[dbCode] = &partnerCodeAdapter{
					baseAdapter: adapter,
					partnerCode: dbCode,
					partnerName: getPartnerDisplayName(dbCode),
				}
			}
		}
	}

	return enabledAdapters
}

// RefreshAdapters reinitializes adapters (useful for config updates)
func (f *partnerAdapterFactory) RefreshAdapters() {
	f.implementations = make(map[string]interfaces.PartnerAdapter)
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
	case "delcaper":
		return f.config.Delcaper, nil
	default:
		return map[string]interface{}{
			"status":              "generic_adapter",
			"partner_code":        dbPartnerCode,
			"implementation_code": implCode,
		}, nil
	}
}

// ValidateConfigurations validates all adapter configurations
func (f *partnerAdapterFactory) ValidateConfigurations() error {
	var errors []string

	// Validate only enabled configurations
	if f.config.Shipyaari.Enabled {
		if err := f.validateShipyaariConfig(); err != nil {
			errors = append(errors, fmt.Sprintf("Shipyaari: %v", err))
		}
	}

	if f.config.SmileCourier.Enabled {
		if err := f.validateSmileCourierConfig(); err != nil {
			errors = append(errors, fmt.Sprintf("Smile Courier: %v", err))
		}
	}

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
