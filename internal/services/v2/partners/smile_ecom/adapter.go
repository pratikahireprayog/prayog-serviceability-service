package smile_ecom

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// SmileEcomAdapter implements the common.PartnerAdapter interface for Smile Ecom database operations
type SmileEcomAdapter struct {
	*common.DatabaseBaseAdapter
	dbClient *DatabaseClient
	config   config.SmileEcomConfig
}

// NewSmileEcomAdapter creates a new Smile Ecom adapter
func NewSmileEcomAdapter(cfg config.SmileEcomConfig, db *sql.DB) common.PartnerAdapter {
	// Create partner config
	partnerConfig := common.GetPartnerConfigDefaults("smile_ecom", "Smile Ecom", common.AdapterTypeDatabase)
	partnerConfig.Enabled = cfg.Enabled
	// Rating comes from database, not config

	// Update auth config (database doesn't need authentication)
	partnerConfig.Auth = common.AuthConfig{
		Type: common.AuthTypeNone,
	}

	// Create base adapter
	baseAdapter := common.NewDatabaseBaseAdapter(partnerConfig)

	// Create Smile Ecom specific components
	dbClient := NewDatabaseClient(cfg, db)

	// Set database client
	baseAdapter.SetDatabaseClient(dbClient)

	adapter := &SmileEcomAdapter{
		DatabaseBaseAdapter: baseAdapter,
		dbClient:            dbClient,
		config:              cfg,
	}

	return adapter
}

// CheckServiceability implements the main serviceability check for Smile Ecom
func (s *SmileEcomAdapter) CheckServiceability(ctx context.Context, req *models.ServiceabilityV2Request) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	// Validate request
	if err := common.ValidateServiceabilityRequest(req); err != nil {
		s.RecordRequest(time.Since(startTime), false)
		return nil, err
	}

	// Check cache first if enabled
	if s.config.CacheEnabled {
		if cachedResult := s.getCachedResult(ctx, req); cachedResult != nil {
			cachedResult.ResponseTime = time.Since(startTime)
			s.RecordRequest(time.Since(startTime), true)
			return cachedResult, nil
		}
	}

	// Transform request to database query
	query, err := s.transformRequestToQuery(req)
	if err != nil {
		s.RecordRequest(time.Since(startTime), false)
		return &common.PartnerServiceabilityResult{
			PartnerCode:   s.GetPartnerCode(),
			IsServiceable: false,
			ResponseTime:  time.Since(startTime),
			Error:         err,
		}, nil
	}

	// Execute database query
	dbResult, err := s.dbClient.QueryServiceability(ctx, query)
	if err != nil {
		s.RecordRequest(time.Since(startTime), false)
		return &common.PartnerServiceabilityResult{
			PartnerCode:   s.GetPartnerCode(),
			IsServiceable: false,
			ResponseTime:  time.Since(startTime),
			Error:         err,
		}, nil
	}

	// Transform response to standard format
	result := s.transformDatabaseResult(dbResult)
	result.ResponseTime = time.Since(startTime)

	// Cache result if enabled
	if s.config.CacheEnabled {
		s.cacheResult(ctx, req, result)
	}

	// Record successful request
	s.RecordRequest(time.Since(startTime), true)

	return result, nil
}

// Initialize performs any necessary initialization
func (s *SmileEcomAdapter) Initialize(ctx context.Context) error {
	if err := s.DatabaseBaseAdapter.Initialize(ctx); err != nil {
		return err
	}

	// Test database connectivity
	if err := s.dbClient.Ping(ctx); err != nil {
		return fmt.Errorf("database connectivity check failed: %w", err)
	}

	s.SetHealthStatus("healthy")
	return nil
}

// IsHealthy checks if the Smile Ecom adapter is healthy
func (s *SmileEcomAdapter) IsHealthy(ctx context.Context) bool {
	// Check base health
	if !s.DatabaseBaseAdapter.IsHealthy(ctx) {
		return false
	}

	// Check database connectivity
	if err := s.dbClient.Ping(ctx); err != nil {
		return false
	}

	return true
}

// transformRequestToQuery converts standard request to database query format
func (s *SmileEcomAdapter) transformRequestToQuery(req *models.ServiceabilityV2Request) (*ServiceabilityQuery, error) {
	query := &ServiceabilityQuery{
		FromPincode: getSourcePincode(req),
		ToPincode:   getDestinationPincode(req),
		CountryCode: getCountryCode(req),
		TableName:   s.config.TableName,
	}

	return query, nil
}

// transformDatabaseResult converts database result to standard format
func (s *SmileEcomAdapter) transformDatabaseResult(dbResult *ServiceabilityData) *common.PartnerServiceabilityResult {
	// TODO: Update response structure when Smile Ecom API integration is finalized
	// Current implementation is database-based - may need to match final payload format

	result := &common.PartnerServiceabilityResult{
		PartnerCode:   s.GetPartnerCode(),
		IsServiceable: dbResult.IsServiceable,
		Services:      make([]models.ServiceV2, 0),
		Capabilities:  make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
	}

	// Add services if available
	if dbResult.IsServiceable && len(dbResult.Services) > 0 {
		for _, service := range dbResult.Services {
			v2Service := models.ServiceV2{
				ServiceCode: service.ServiceCode,
				ServiceName: service.ServiceName,
				TATDays:     service.TATDays,
				IsCOD:       service.CODAvailable,
				Pickup:      service.PickupAvailable,
				Delivery:    service.DeliveryAvailable,
				Insurance:   service.InsuranceAvailable,
				ProductTypes: map[string]bool{
					"ecommerce": true,
					"general":   true,
				},
				DeliveryModes: map[string]bool{
					"standard": true,
				},
			}

			// Add pricing if available
			if service.BaseCost > 0 {
				v2Service.Pricing = &models.ServicePricingV2{
					BaseCost:      service.BaseCost,
					Currency:      service.Currency,
					CODCharges:    service.CODCharges,
					FuelSurcharge: service.FuelSurcharge,
				}
			}

			result.Services = append(result.Services, v2Service)
		}
	}

	// Add capabilities from database result
	for key, value := range dbResult.Capabilities {
		result.Capabilities[key] = value
	}

	// Add metadata
	result.Metadata["adapter_type"] = "database"
	result.Metadata["table_name"] = s.config.TableName
	result.Metadata["cache_enabled"] = s.config.CacheEnabled

	return result
}

// getCachedResult retrieves cached result if available
func (s *SmileEcomAdapter) getCachedResult(ctx context.Context, req *models.ServiceabilityV2Request) *common.PartnerServiceabilityResult {
	// TODO: Implement caching logic
	// This could use Redis, in-memory cache, or database-based caching
	// For now, return nil (cache miss)
	return nil
}

// cacheResult caches the result if caching is enabled
func (s *SmileEcomAdapter) cacheResult(ctx context.Context, req *models.ServiceabilityV2Request, result *common.PartnerServiceabilityResult) {
	// TODO: Implement caching logic
	// This could use Redis, in-memory cache, or database-based caching
}

// buildCacheKey builds a cache key for the request
func (s *SmileEcomAdapter) buildCacheKey(req *models.ServiceabilityV2Request) string {
	countryCode := "IN" // default
	if req.CountryCode != nil {
		countryCode = *req.CountryCode
	}

	key := fmt.Sprintf("smile_ecom:serviceability:%s", countryCode)

	if req.SourcePostalCode != nil {
		key += ":" + *req.SourcePostalCode
	}

	if req.DestinationPostalCode != nil {
		key += ":" + *req.DestinationPostalCode
	} else if req.PostalCode != nil {
		key += ":" + *req.PostalCode
	}

	return key
}

// Helper functions to extract data from standard request
func getSourcePincode(req *models.ServiceabilityV2Request) string {
	if req.SourcePostalCode != nil {
		return *req.SourcePostalCode
	}
	if req.PostalCode != nil {
		// For single postal code requests, may not need source
		return ""
	}
	return ""
}

func getDestinationPincode(req *models.ServiceabilityV2Request) string {
	if req.DestinationPostalCode != nil {
		return *req.DestinationPostalCode
	}
	if req.PostalCode != nil {
		return *req.PostalCode
	}
	return ""
}

func getCountryCode(req *models.ServiceabilityV2Request) string {
	if req.CountryCode != nil {
		return *req.CountryCode
	}
	return "IN" // Default to India
}
