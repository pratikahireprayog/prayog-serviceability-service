package porter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
	services "prayog-serviceability-service/internal/services/v1/data"

	"github.com/sirupsen/logrus"
)

// PorterAdapter implements the common.PartnerAdapter interface for Porter
type PorterAdapter struct {
	*common.DatabaseBaseAdapter
	geolocationService services.GeolocationService
	config             config.PorterConfig
	logger             *logrus.Logger
}

// NewPorterAdapter creates a new Porter adapter
func NewPorterAdapter(cfg config.PorterConfig, db *sql.DB, geolocationService services.GeolocationService) common.PartnerAdapter {
	// Create partner config
	partnerConfig := common.GetPartnerConfigDefaults("porter", "Porter", common.AdapterTypeDatabase)
	partnerConfig.Timeout = cfg.Timeout
	partnerConfig.Enabled = cfg.Enabled

	// Create base adapter
	baseAdapter := common.NewDatabaseBaseAdapter(partnerConfig, db)

	adapter := &PorterAdapter{
		DatabaseBaseAdapter: baseAdapter,
		geolocationService:  geolocationService,
		config:              cfg,
		logger:              logrus.New(),
	}

	return adapter
}

// SupportsRequest checks if Porter can handle this request
func (p *PorterAdapter) SupportsRequest(ctx context.Context, request *models.ServiceabilityV2Request) bool {
	// Porter only supports hyperlocal parcel category
	if request.ParcelCategory == nil || *request.ParcelCategory != "hyperlocal" {
		return false
	}

	// Porter requires either coordinates OR postal codes for both source and destination
	hasSourceCoords := request.SourceLatitude != nil && request.SourceLongitude != nil
	hasDestCoords := request.DestinationLatitude != nil && request.DestinationLongitude != nil
	hasSourcePostal := request.SourcePostalCode != nil
	hasDestPostal := request.DestinationPostalCode != nil

	// Must have either coordinates OR postal codes for both locations
	if !((hasSourceCoords || hasSourcePostal) && (hasDestCoords || hasDestPostal)) {
		return false
	}

	return true
}

// CheckServiceability checks serviceability for the request using Porter's database query
func (p *PorterAdapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	// Check if Porter supports this request
	if !p.SupportsRequest(ctx, request) {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  p.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Metadata: map[string]interface{}{
				"reason": "Porter does not support this request type (requires hyperlocal parcel category with coordinates OR postal codes)",
			},
		}, nil
	}

	// Validate request requirements
	if err := p.validatePorterRequirements(request); err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  p.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Porter validation failed: %v", err)}[0],
			Metadata: map[string]interface{}{
				"reason": "Porter validation failed",
			},
		}, nil
	}

	// Get coordinates directly from the request
	sourceLat, sourceLng, err := p.getCoordinatesFromRequest(request, "source")
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  p.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Failed to get source coordinates: %v", err)}[0],
		}, nil
	}

	destLat, destLng, err := p.getCoordinatesFromRequest(request, "destination")
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  p.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Failed to get destination coordinates: %v", err)}[0],
		}, nil
	}

	// Check serviceability for both source and destination using the same query
	// Both locations must return "INSIDE" for the route to be serviceable
	sourceServiceable, err := p.checkLocationServiceability(ctx, sourceLat, sourceLng)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  p.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Failed to check source serviceability: %v", err)}[0],
		}, nil
	}

	destServiceable, err := p.checkLocationServiceability(ctx, destLat, destLng)
	if err != nil {
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  p.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Failed to check destination serviceability: %v", err)}[0],
		}, nil
	}

	// Both source and destination must be serviceable (return "INSIDE") for Porter to be available
	isServiceable := sourceServiceable && destServiceable

	// Build response
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		PartnerName:  p.GetPartnerName(),
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata: map[string]interface{}{
			"source_serviceable":      sourceServiceable,
			"destination_serviceable": destServiceable,
			"source_coordinates":      fmt.Sprintf("%f,%f", sourceLat, sourceLng),
			"destination_coordinates": fmt.Sprintf("%f,%f", destLat, destLng),
		},
		ResponseTime: time.Since(startTime),
	}

	if isServiceable {
		// Add Porter hyperlocal service
		result.Services = append(result.Services, models.ServiceV2{
			ServiceCode:   "porter_hyperlocal",
			ServiceName:   "Porter Hyperlocal Delivery",
			TATDays:       1, // Same day or next day for hyperlocal
			IsCOD:         true,
			Pickup:        true,
			Delivery:      true,
			Insurance:     false,
			ProductTypes:  map[string]bool{"hyperlocal": true},
			DeliveryModes: map[string]bool{"hyperlocal": true},
			Pricing: &models.ServicePricingV2{
				BaseCost: 0, // Pricing to be determined by Porter
				Currency: "INR",
			},
		})

		// Set capabilities
		result.Capabilities["is_serviceable"] = true
		result.Capabilities["pickup_available"] = true
		result.Capabilities["delivery_available"] = true
		result.Capabilities["cod_available"] = true
		result.Capabilities["insurance_available"] = false
	} else {
		// Set error message for non-serviceable areas
		// Both source and destination must be "INSIDE" Porter's pickup boundaries
		errorMsg := "Location not serviceable by Porter"
		if !sourceServiceable && !destServiceable {
			errorMsg = "Both source and destination locations are outside Porter's pickup boundaries"
		} else if !sourceServiceable {
			errorMsg = "Source location is outside Porter's pickup boundaries"
		} else {
			errorMsg = "Destination location is outside Porter's pickup boundaries"
		}
		result.ErrorMessage = &errorMsg

		// Set capabilities to false
		result.Capabilities["is_serviceable"] = false
		result.Capabilities["pickup_available"] = false
		result.Capabilities["delivery_available"] = false
		result.Capabilities["cod_available"] = false
		result.Capabilities["insurance_available"] = false
	}

	return result, nil
}

// Initialize performs any necessary initialization
func (p *PorterAdapter) Initialize(ctx context.Context) error {
	if err := p.DatabaseBaseAdapter.Initialize(ctx); err != nil {
		return err
	}

	// Test database connection
	if err := p.GetDatabaseClient().Ping(ctx); err != nil {
		p.SetHealthStatus("unhealthy")
		return fmt.Errorf("failed to ping Porter database: %w", err)
	}

	p.SetHealthStatus("healthy")
	return nil
}

// IsHealthy checks if the Porter adapter is healthy
func (p *PorterAdapter) IsHealthy(ctx context.Context) bool {
	// Check base health
	if !p.DatabaseBaseAdapter.IsHealthy(ctx) {
		return false
	}

	// Test database connection
	if err := p.GetDatabaseClient().Ping(ctx); err != nil {
		return false
	}

	return true
}

// validatePorterRequirements validates Porter-specific requirements
func (p *PorterAdapter) validatePorterRequirements(req *models.ServiceabilityV2Request) error {
	// Porter only supports hyperlocal parcel category
	if req.ParcelCategory == nil || *req.ParcelCategory != "hyperlocal" {
		return fmt.Errorf("Porter only supports hyperlocal parcel category")
	}

	// Porter requires either coordinates OR postal codes for both source and destination
	hasSourceCoords := req.SourceLatitude != nil && req.SourceLongitude != nil
	hasDestCoords := req.DestinationLatitude != nil && req.DestinationLongitude != nil
	hasSourcePostal := req.SourcePostalCode != nil
	hasDestPostal := req.DestinationPostalCode != nil

	// Must have either coordinates OR postal codes for both locations
	if !((hasSourceCoords || hasSourcePostal) && (hasDestCoords || hasDestPostal)) {
		return fmt.Errorf("Porter requires either coordinates OR postal codes for both source and destination locations")
	}

	return nil
}

// getCoordinatesFromRequest retrieves latitude and longitude from the request
// It can get coordinates directly from the request OR convert postal codes to coordinates
func (p *PorterAdapter) getCoordinatesFromRequest(req *models.ServiceabilityV2Request, locationType string) (float64, float64, error) {
	var lat, lng *float64
	var postalCode *string

	switch locationType {
	case "source":
		lat = req.SourceLatitude
		lng = req.SourceLongitude
		postalCode = req.SourcePostalCode
	case "destination":
		lat = req.DestinationLatitude
		lng = req.DestinationLongitude
		postalCode = req.DestinationPostalCode
	default:
		return 0, 0, fmt.Errorf("invalid location type: %s", locationType)
	}

	// If coordinates are provided directly, use them
	if lat != nil && lng != nil {
		return *lat, *lng, nil
	}

	// If postal code is provided, convert it to coordinates using geolocation service
	if postalCode != nil {
		if p.geolocationService == nil {
			return 0, 0, fmt.Errorf("geolocation service is required to convert postal code to coordinates")
		}

		// Get coordinates from postal code
		geoLocation, err := p.geolocationService.GetLocationByPostalCode(context.Background(), *postalCode, "IN")
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get coordinates for %s postal code %s: %w", locationType, *postalCode, err)
		}

		if geoLocation.Latitude == nil || geoLocation.Longitude == nil {
			return 0, 0, fmt.Errorf("no coordinates found for %s postal code %s", locationType, *postalCode)
		}

		return *geoLocation.Latitude, *geoLocation.Longitude, nil
	}

	return 0, 0, fmt.Errorf("neither coordinates nor postal code provided for %s location", locationType)
}

// checkLocationServiceability checks if a location is serviceable using Porter's database query
// This function is called for both source and destination coordinates
// Both locations must return "INSIDE" for the route to be serviceable
func (p *PorterAdapter) checkLocationServiceability(ctx context.Context, latitude, longitude float64) (bool, error) {
	query := `
		SELECT 
			CASE 
				WHEN EXISTS (
					SELECT 1
					FROM pickup_boundaries
					WHERE ST_Contains(
						boundary,
						ST_GeomFromText(?, 4326)
					)
				) 
				THEN 'INSIDE'
				ELSE 'OUTSIDE'
			END AS location_status;
	`

	// Create the POINT geometry string
	pointGeometry := fmt.Sprintf("POINT(%f %f)", longitude, latitude)

	// Execute query
	var locationStatus string
	err := p.GetDatabaseClient().QueryRow(ctx, query, pointGeometry).Scan(&locationStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // No boundaries found, consider as not serviceable
		}
		return false, fmt.Errorf("failed to execute Porter serviceability query: %w", err)
	}

	return locationStatus == "INSIDE", nil
}
