package porter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/sirupsen/logrus"
)

// PorterAdapter implements the common.PartnerAdapter interface for Porter
type PorterAdapter struct {
	*common.DatabaseBaseAdapter
	config config.PorterConfig
	logger *logrus.Logger
}

// NewPorterAdapter creates a new Porter adapter
func NewPorterAdapter(cfg config.PorterConfig, db *sql.DB) common.PartnerAdapter {
	// Create partner config
	partnerConfig := common.GetPartnerConfigDefaults("porter", "Porter", common.AdapterTypeDatabase)
	partnerConfig.Timeout = cfg.Timeout
	partnerConfig.Enabled = cfg.Enabled

	// Create base adapter
	baseAdapter := common.NewDatabaseBaseAdapter(partnerConfig)

	// Create database client and set it
	dbClient := &PorterDatabaseClient{DB: db}
	baseAdapter.SetDatabaseClient(dbClient)

	adapter := &PorterAdapter{
		DatabaseBaseAdapter: baseAdapter,
		config:              cfg,
		logger:              logrus.New(),
	}

	return adapter
}

// SupportsRequest checks if Porter can handle this request using only start/destination pincodes
func (p *PorterAdapter) SupportsRequest(ctx context.Context, request *models.ServiceabilityV2Request) bool {
	// Require postal codes for both source and destination
	hasSourcePostal := request.SourcePostalCode != nil && *request.SourcePostalCode != ""
	hasDestPostal := request.DestinationPostalCode != nil && *request.DestinationPostalCode != ""
	return hasSourcePostal && hasDestPostal
}

// CheckServiceability checks serviceability for the request using Porter's database query
func (p *PorterAdapter) CheckServiceability(ctx context.Context, request *models.ServiceabilityV2Request, partnerInfo common.PartnerInfo) (*common.PartnerServiceabilityResult, error) {
	startTime := time.Now()

	// Log start with pincodes
	sourcePin := ""
	if request.SourcePostalCode != nil {
		sourcePin = *request.SourcePostalCode
	}
	destPin := ""
	if request.DestinationPostalCode != nil {
		destPin = *request.DestinationPostalCode
	}
	p.logger.WithFields(logrus.Fields{
		"component":           "porter_adapter",
		"method":              "CheckServiceability",
		"source_pincode":      sourcePin,
		"destination_pincode": destPin,
	}).Info("Starting Porter serviceability check")

	// Check if Porter supports this request using only pincodes
	if !p.SupportsRequest(ctx, request) {
		hasSourcePostal := request.SourcePostalCode != nil && *request.SourcePostalCode != ""
		hasDestPostal := request.DestinationPostalCode != nil && *request.DestinationPostalCode != ""
		p.logger.WithFields(logrus.Fields{
			"component":              "porter_adapter",
			"method":                 "CheckServiceability",
			"has_source_postal":      hasSourcePostal,
			"has_destination_postal": hasDestPostal,
		}).Warn("Unsupported request for Porter - missing required pincodes")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  p.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Metadata: map[string]interface{}{
				"reason": "Porter requires both source and destination postal codes",
			},
		}, nil
	}

	// Validate request requirements
	if err := p.validatePorterRequirements(request); err != nil {
		p.logger.WithFields(logrus.Fields{
			"component": "porter_adapter",
			"method":    "CheckServiceability",
			"error":     err.Error(),
		}).Warn("Porter validation failed")
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

	// Get coordinates from database using postal codes
	sourceLat, sourceLng, err := p.getCoordinatesFromDatabase(ctx, request, "source")
	if err != nil {
		p.logger.WithFields(logrus.Fields{
			"component": "porter_adapter",
			"method":    "CheckServiceability",
			"stage":     "get_source_coordinates",
			"error":     err.Error(),
		}).Error("Failed to get source coordinates from DB")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  p.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Failed to get source coordinates: %v", err)}[0],
		}, nil
	}

	destLat, destLng, err := p.getCoordinatesFromDatabase(ctx, request, "destination")
	if err != nil {
		p.logger.WithFields(logrus.Fields{
			"component": "porter_adapter",
			"method":    "CheckServiceability",
			"stage":     "get_destination_coordinates",
			"error":     err.Error(),
		}).Error("Failed to get destination coordinates from DB")
		return &common.PartnerServiceabilityResult{
			PartnerID:    partnerInfo.PartnerID,
			PartnerCode:  partnerInfo.PartnerCode,
			PartnerName:  p.GetPartnerName(),
			Services:     make([]models.ServiceV2, 0),
			Error:        err,
			ErrorMessage: &[]string{fmt.Sprintf("Failed to get destination coordinates: %v", err)}[0],
		}, nil
	}

	// Check serviceability for both source and destination and get boundary IDs
	sourceBoundaryID, sourceServiceable, err := p.checkLocationServiceabilityWithID(ctx, sourceLat, sourceLng)
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

	destBoundaryID, destServiceable, err := p.checkLocationServiceabilityWithID(ctx, destLat, destLng)
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

	// Both source and destination must be serviceable AND in the same boundary for Porter to be available
	isServiceable := sourceServiceable && destServiceable && sourceBoundaryID == destBoundaryID

	// If not serviceable, return nil to exclude Porter from partners array
	if !isServiceable {
		p.logger.WithFields(logrus.Fields{
			"component":              "porter_adapter",
			"method":                 "CheckServiceability",
			"source_serviceable":     sourceServiceable,
			"destination_serviceable": destServiceable,
			"source_boundary_id":     sourceBoundaryID,
			"destination_boundary_id": destBoundaryID,
			"same_boundary":          sourceBoundaryID == destBoundaryID,
			"source_coordinates":     fmt.Sprintf("%f,%f", sourceLat, sourceLng),
			"destination_coordinates": fmt.Sprintf("%f,%f", destLat, destLng),
		}).Info("Porter not serviceable - excluding from partners array")
		return nil, nil
	}

	// Build response for serviceable case
	result := &common.PartnerServiceabilityResult{
		PartnerID:    partnerInfo.PartnerID,
		PartnerCode:  partnerInfo.PartnerCode,
		PartnerName:  p.GetPartnerName(),
		Services:     make([]models.ServiceV2, 0),
		Capabilities: make(map[string]interface{}),
		Metadata: map[string]interface{}{
			"source_serviceable":      sourceServiceable,
			"destination_serviceable": destServiceable,
			"source_boundary_id":      sourceBoundaryID,
			"destination_boundary_id": destBoundaryID,
			"same_boundary":           sourceBoundaryID == destBoundaryID,
			"source_coordinates":      fmt.Sprintf("%f,%f", sourceLat, sourceLng),
			"destination_coordinates": fmt.Sprintf("%f,%f", destLat, destLng),
		},
		ResponseTime: time.Since(startTime),
	}

	p.logger.WithFields(logrus.Fields{
		"component":              "porter_adapter",
		"method":                 "CheckServiceability",
		"source_serviceable":     sourceServiceable,
		"destination_serviceable": destServiceable,
		"source_boundary_id":     sourceBoundaryID,
		"destination_boundary_id": destBoundaryID,
	}).Info("Porter serviceable - adding service to result")

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
	// Porter requires postal codes for both source and destination to fetch coordinates from database
	hasSourcePostal := req.SourcePostalCode != nil && *req.SourcePostalCode != ""
	hasDestPostal := req.DestinationPostalCode != nil && *req.DestinationPostalCode != ""

	// Must have postal codes for both locations
	if !(hasSourcePostal && hasDestPostal) {
		return fmt.Errorf("Porter requires postal codes for both source and destination locations")
	}

	return nil
}

// getCoordinatesFromRequest retrieves latitude and longitude directly from the request
// Porter only uses coordinates from the API request, no postal code conversion
func (p *PorterAdapter) getCoordinatesFromRequest(req *models.ServiceabilityV2Request, locationType string) (float64, float64, error) {
	var lat, lng *float64

	switch locationType {
	case "source":
		lat = req.SourceLatitude
		lng = req.SourceLongitude
	case "destination":
		lat = req.DestinationLatitude
		lng = req.DestinationLongitude
	default:
		return 0, 0, fmt.Errorf("invalid location type: %s", locationType)
	}

	// Porter only uses coordinates directly from the API request
	if lat == nil || lng == nil {
		return 0, 0, fmt.Errorf("%s coordinates are required for Porter serviceability check", locationType)
	}

	return *lat, *lng, nil
}

// getCoordinatesFromDatabase retrieves latitude and longitude from the database using postal codes
// This is the new method for Porter to fetch coordinates from geo_locations table
func (p *PorterAdapter) getCoordinatesFromDatabase(ctx context.Context, req *models.ServiceabilityV2Request, locationType string) (float64, float64, error) {
	var postalCode string
	var countryCode string

	switch locationType {
	case "source":
		if req.SourcePostalCode == nil || *req.SourcePostalCode == "" {
			return 0, 0, fmt.Errorf("source postal code is required for Porter serviceability check")
		}
		postalCode = *req.SourcePostalCode
		countryCode = "IN" // Default to India for Porter
	case "destination":
		if req.DestinationPostalCode == nil || *req.DestinationPostalCode == "" {
			return 0, 0, fmt.Errorf("destination postal code is required for Porter serviceability check")
		}
		postalCode = *req.DestinationPostalCode
		countryCode = "IN" // Default to India for Porter
	default:
		return 0, 0, fmt.Errorf("invalid location type: %s", locationType)
	}

	// Query to fetch coordinates from geo_locations table
	query := `
		SELECT latitude, longitude
		FROM public.geo_locations 
		WHERE postal_code = $1 AND country_code = $2;
	`

	// Log the query being executed
	p.logger.WithFields(logrus.Fields{
		"component":   "porter_adapter",
		"method":      "getCoordinatesFromDatabase",
		"location_type": locationType,
		"postal_code": postalCode,
		"country_code": countryCode,
		"query":       query,
	}).Info("Fetching coordinates from database")

	// Execute query
	dbRow, err := p.GetDatabaseClient().QueryRow(ctx, query, postalCode, countryCode)
	if err != nil {
		if err == sql.ErrNoRows {
			p.logger.WithFields(logrus.Fields{
				"component":   "porter_adapter",
				"method":      "getCoordinatesFromDatabase",
				"location_type": locationType,
				"postal_code": postalCode,
				"country_code": countryCode,
				"error":       "No coordinates found for postal code",
			}).Warn("No coordinates found in database for postal code")
			return 0, 0, fmt.Errorf("no coordinates found for postal code %s in country %s", postalCode, countryCode)
		}
		p.logger.WithFields(logrus.Fields{
			"component":   "porter_adapter",
			"method":      "getCoordinatesFromDatabase",
			"location_type": locationType,
			"postal_code": postalCode,
			"country_code": countryCode,
			"error":       err.Error(),
		}).Error("Failed to fetch coordinates from database")
		return 0, 0, fmt.Errorf("failed to fetch coordinates from database: %w", err)
	}

	// Extract latitude and longitude from the database row
	latitude, ok := dbRow.Data["latitude"]
	if !ok {
		p.logger.WithFields(logrus.Fields{
			"component":   "porter_adapter",
			"method":      "getCoordinatesFromDatabase",
			"location_type": locationType,
			"postal_code": postalCode,
			"row_data":    dbRow.Data,
		}).Error("latitude column not found in query result")
		return 0, 0, fmt.Errorf("latitude column not found in query result")
	}

	longitude, ok := dbRow.Data["longitude"]
	if !ok {
		p.logger.WithFields(logrus.Fields{
			"component":   "porter_adapter",
			"method":      "getCoordinatesFromDatabase",
			"location_type": locationType,
			"postal_code": postalCode,
			"row_data":    dbRow.Data,
		}).Error("longitude column not found in query result")
		return 0, 0, fmt.Errorf("longitude column not found in query result")
	}

	// Convert to float64
	lat, ok := latitude.(float64)
	if !ok {
		p.logger.WithFields(logrus.Fields{
			"component":   "porter_adapter",
			"method":      "getCoordinatesFromDatabase",
			"location_type": locationType,
			"postal_code": postalCode,
			"latitude":    latitude,
		}).Error("latitude is not a valid float64")
		return 0, 0, fmt.Errorf("latitude is not a valid float64: %v", latitude)
	}

	lng, ok := longitude.(float64)
	if !ok {
		p.logger.WithFields(logrus.Fields{
			"component":   "porter_adapter",
			"method":      "getCoordinatesFromDatabase",
			"location_type": locationType,
			"postal_code": postalCode,
			"longitude":   longitude,
		}).Error("longitude is not a valid float64")
		return 0, 0, fmt.Errorf("longitude is not a valid float64: %v", longitude)
	}

	p.logger.WithFields(logrus.Fields{
		"component":   "porter_adapter",
		"method":      "getCoordinatesFromDatabase",
		"location_type": locationType,
		"postal_code": postalCode,
		"latitude":    lat,
		"longitude":   lng,
	}).Info("Successfully fetched coordinates from database")

	return lat, lng, nil
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
						ST_GeomFromText($1, 4326)
					)
				) 
				THEN 'INSIDE'
				ELSE 'OUTSIDE'
			END AS location_status;
	`

	// Create the POINT geometry string
	pointGeometry := fmt.Sprintf("POINT(%f %f)", longitude, latitude)

	// Log the query being executed with actual coordinate values
	p.logger.WithFields(logrus.Fields{
		"component": "porter_adapter",
		"method":    "checkLocationServiceability",
		"latitude":  latitude,
		"longitude": longitude,
		"query":     query,
		"params":    pointGeometry,
		"actual_query": fmt.Sprintf(`
			SELECT 
				CASE 
					WHEN EXISTS (
						SELECT 1
						FROM pickup_boundaries
						WHERE ST_Contains(
							boundary,
							ST_GeomFromText('POINT(%f %f)', 4326)
						)
					) 
					THEN 'INSIDE'
					ELSE 'OUTSIDE'
				END AS location_status;`, longitude, latitude),
	}).Info("Executing Porter serviceability query")

	// Execute query
	dbRow, err := p.GetDatabaseClient().QueryRow(ctx, query, pointGeometry)
	if err != nil {
		if err == sql.ErrNoRows {
			p.logger.WithFields(logrus.Fields{
				"component": "porter_adapter",
				"method":    "checkLocationServiceability",
				"latitude":  latitude,
				"longitude": longitude,
				"error":     "No rows returned",
			}).Warn("Porter serviceability query returned no rows")
			return false, nil // No boundaries found, consider as not serviceable
		}
		p.logger.WithFields(logrus.Fields{
			"component": "porter_adapter",
			"method":    "checkLocationServiceability",
			"latitude":  latitude,
			"longitude": longitude,
			"error":     err.Error(),
		}).Error("Failed to execute Porter serviceability query")
		return false, fmt.Errorf("failed to execute Porter serviceability query: %w", err)
	}

	// Extract location_status from the database row
	locationStatus, ok := dbRow.Data["location_status"]
	if !ok {
		p.logger.WithFields(logrus.Fields{
			"component": "porter_adapter",
			"method":    "checkLocationServiceability",
			"latitude":  latitude,
			"longitude": longitude,
			"row_data":  dbRow.Data,
		}).Error("location_status column not found in query result")
		return false, fmt.Errorf("location_status column not found in query result")
	}

	// Convert to string and check if it's "INSIDE"
	statusStr, ok := locationStatus.(string)
	if !ok {
		p.logger.WithFields(logrus.Fields{
			"component":      "porter_adapter",
			"method":         "checkLocationServiceability",
			"latitude":       latitude,
			"longitude":      longitude,
			"locationStatus": locationStatus,
			"type":           fmt.Sprintf("%T", locationStatus),
		}).Error("location_status is not a string")
		return false, fmt.Errorf("location_status is not a string: %v", locationStatus)
	}

	// Log the query result
	p.logger.WithFields(logrus.Fields{
		"component":      "porter_adapter",
		"method":         "checkLocationServiceability",
		"latitude":       latitude,
		"longitude":      longitude,
		"locationStatus": statusStr,
		"isServiceable":  statusStr == "INSIDE",
		"queryDuration":  dbRow.Duration.String(),
	}).Info("Porter serviceability query result")

	return statusStr == "INSIDE", nil
}

// checkLocationServiceabilityWithID checks if a location is serviceable and returns the boundary ID
// This function is called for both source and destination coordinates
// Returns boundary ID, serviceability status, and error
func (p *PorterAdapter) checkLocationServiceabilityWithID(ctx context.Context, latitude, longitude float64) (int, bool, error) {
	query := `
		SELECT 
			id,
			'INSIDE' AS location_status
		FROM pickup_boundaries
		WHERE ST_Contains(
			boundary,
			ST_GeomFromText($1, 4326)
		);
	`

	// Create the POINT geometry string
	pointGeometry := fmt.Sprintf("POINT(%f %f)", longitude, latitude)

	// Log the query being executed with actual coordinate values
	p.logger.WithFields(logrus.Fields{
		"component": "porter_adapter",
		"method":    "checkLocationServiceabilityWithID",
		"latitude":  latitude,
		"longitude": longitude,
		"query":     query,
		"params":    pointGeometry,
		"actual_query": fmt.Sprintf(`
			SELECT 
				id,
				'INSIDE' AS location_status
			FROM pickup_boundaries
			WHERE ST_Contains(
				boundary,
				ST_GeomFromText('POINT(%f %f)', 4326)
			);`, longitude, latitude),
	}).Info("Executing Porter serviceability query with boundary ID")

	// Execute query
	dbRow, err := p.GetDatabaseClient().QueryRow(ctx, query, pointGeometry)
	if err != nil {
		if err == sql.ErrNoRows {
			p.logger.WithFields(logrus.Fields{
				"component": "porter_adapter",
				"method":    "checkLocationServiceabilityWithID",
				"latitude":  latitude,
				"longitude": longitude,
				"error":     "No rows returned - location outside all boundaries",
			}).Warn("Porter serviceability query returned no rows")
			return 0, false, nil // No boundaries found, consider as not serviceable
		}
		p.logger.WithFields(logrus.Fields{
			"component": "porter_adapter",
			"method":    "checkLocationServiceabilityWithID",
			"latitude":  latitude,
			"longitude": longitude,
			"error":     err.Error(),
		}).Error("Failed to execute Porter serviceability query")
		return 0, false, fmt.Errorf("failed to execute Porter serviceability query: %w", err)
	}

	// Extract boundary ID from the database row
	boundaryID, ok := dbRow.Data["id"]
	if !ok {
		p.logger.WithFields(logrus.Fields{
			"component": "porter_adapter",
			"method":    "checkLocationServiceabilityWithID",
			"latitude":  latitude,
			"longitude": longitude,
			"row_data":  dbRow.Data,
		}).Error("id column not found in query result")
		return 0, false, fmt.Errorf("id column not found in query result")
	}

	// Convert boundary ID to int
	id, ok := boundaryID.(int64)
	if !ok {
		p.logger.WithFields(logrus.Fields{
			"component": "porter_adapter",
			"method":    "checkLocationServiceabilityWithID",
			"latitude":  latitude,
			"longitude": longitude,
			"boundary_id": boundaryID,
		}).Error("boundary ID is not a valid int64")
		return 0, false, fmt.Errorf("boundary ID is not a valid int64: %v", boundaryID)
	}

	// Log the query result
	p.logger.WithFields(logrus.Fields{
		"component":      "porter_adapter",
		"method":         "checkLocationServiceabilityWithID",
		"latitude":       latitude,
		"longitude":      longitude,
		"boundary_id":    id,
		"isServiceable":  true,
		"queryDuration":  dbRow.Duration.String(),
	}).Info("Porter serviceability query result with boundary ID")

	return int(id), true, nil
}

// PorterDatabaseClient implements the common.DatabaseClient interface for Porter
type PorterDatabaseClient struct {
	DB *sql.DB
}

// Query implements common.DatabaseClient.Query
func (p *PorterDatabaseClient) Query(ctx context.Context, query string, args ...interface{}) (*common.DatabaseResult, error) {
	startTime := time.Now()

	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Create map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &common.DatabaseResult{
		Rows:     results,
		RowCount: int64(len(results)),
		Duration: time.Since(startTime),
	}, nil
}

// QueryRow implements common.DatabaseClient.QueryRow
func (p *PorterDatabaseClient) QueryRow(ctx context.Context, query string, args ...interface{}) (*common.DatabaseRow, error) {
	startTime := time.Now()

	// Use Query and return the first row
	result, err := p.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if len(result.Rows) == 0 {
		return nil, sql.ErrNoRows
	}

	return &common.DatabaseRow{
		Data:     result.Rows[0],
		Duration: time.Since(startTime),
	}, nil
}

// BeginTx implements common.DatabaseClient.BeginTx
func (p *PorterDatabaseClient) BeginTx(ctx context.Context) (common.DatabaseTransaction, error) {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &PorterDatabaseTransaction{tx: tx}, nil
}

// Ping implements common.DatabaseClient.Ping
func (p *PorterDatabaseClient) Ping(ctx context.Context) error {
	return p.DB.PingContext(ctx)
}

// PorterDatabaseTransaction implements common.DatabaseTransaction
type PorterDatabaseTransaction struct {
	tx *sql.Tx
}

// Commit implements common.DatabaseTransaction.Commit
func (pt *PorterDatabaseTransaction) Commit() error {
	return pt.tx.Commit()
}

// Rollback implements common.DatabaseTransaction.Rollback
func (pt *PorterDatabaseTransaction) Rollback() error {
	return pt.tx.Rollback()
}

// Query implements common.DatabaseTransaction.Query
func (pt *PorterDatabaseTransaction) Query(ctx context.Context, query string, args ...interface{}) (*common.DatabaseResult, error) {
	startTime := time.Now()

	rows, err := pt.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Create map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &common.DatabaseResult{
		Rows:     results,
		RowCount: int64(len(results)),
		Duration: time.Since(startTime),
	}, nil
}

// QueryRow implements common.DatabaseTransaction.QueryRow
func (pt *PorterDatabaseTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) (*common.DatabaseRow, error) {
	startTime := time.Now()

	// Use Query and return the first row
	result, err := pt.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if len(result.Rows) == 0 {
		return nil, sql.ErrNoRows
	}

	return &common.DatabaseRow{
		Data:     result.Rows[0],
		Duration: time.Since(startTime),
	}, nil
}
