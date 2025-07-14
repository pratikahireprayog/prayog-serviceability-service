package smile_ecom

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/config"
)

// DatabaseClient implements the common.DatabaseClient interface for Smile Ecom
type DatabaseClient struct {
	db     *sql.DB
	config config.SmileEcomConfig
}

// NewDatabaseClient creates a new Smile Ecom database client
func NewDatabaseClient(cfg config.SmileEcomConfig, db *sql.DB) *DatabaseClient {
	return &DatabaseClient{
		db:     db,
		config: cfg,
	}
}

// QueryServiceability queries the database for serviceability information
func (d *DatabaseClient) QueryServiceability(ctx context.Context, query *ServiceabilityQuery) (*ServiceabilityData, error) {
	// Build the SQL query
	sqlQuery := d.buildServiceabilityQuery(query)

	// Execute the query
	rows, err := d.db.QueryContext(ctx, sqlQuery, query.ToPincode, query.CountryCode)
	if err != nil {
		return nil, fmt.Errorf("failed to execute serviceability query: %w", err)
	}
	defer rows.Close()

	var services []SmileEcomService

	// Process results
	for rows.Next() {
		var result DatabaseQueryResult
		err := rows.Scan(
			&result.ServiceCode,
			&result.ServiceName,
			&result.TATDays,
			&result.CODAvailable,
			&result.PickupAvailable,
			&result.DeliveryAvailable,
			&result.InsuranceAvailable,
			&result.BaseCost,
			&result.Currency,
			&result.CODCharges,
			&result.FuelSurcharge,
			&result.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert to service model
		service := SmileEcomService{
			ServiceCode:        result.ServiceCode,
			ServiceName:        result.ServiceName,
			TATDays:            result.TATDays,
			CODAvailable:       result.CODAvailable,
			PickupAvailable:    result.PickupAvailable,
			DeliveryAvailable:  result.DeliveryAvailable,
			InsuranceAvailable: result.InsuranceAvailable,
			BaseCost:           result.BaseCost,
			Currency:           result.Currency,
			CODCharges:         result.CODCharges,
			FuelSurcharge:      result.FuelSurcharge,
			IsActive:           result.IsActive,
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Build result
	result := &ServiceabilityData{
		IsServiceable: len(services) > 0,
		Services:      services,
		Capabilities:  make(map[string]interface{}),
	}

	// Add capabilities
	if len(services) > 0 {
		result.Capabilities["table_source"] = query.TableName
		result.Capabilities["query_type"] = "database"
		result.Capabilities["service_count"] = len(services)
	}

	return result, nil
}

// buildServiceabilityQuery builds the SQL query for serviceability check
func (d *DatabaseClient) buildServiceabilityQuery(query *ServiceabilityQuery) string {
	tableName := query.TableName
	if tableName == "" {
		tableName = d.config.TableName // Use default table name
	}

	sqlQuery := fmt.Sprintf(`
		SELECT 
			service_code, service_name, tat_days, cod_available,
			pickup_available, delivery_available, insurance_available,
			base_cost, currency, cod_charges, fuel_surcharge, is_active
		FROM %s
		WHERE 
			(from_pincode = $1 OR from_pincode IS NULL)
			AND to_pincode = $1
			AND country_code = $2
			AND is_active = true
		ORDER BY service_code`, tableName)

	return sqlQuery
}

// Query implements common.DatabaseClient.Query
func (d *DatabaseClient) Query(ctx context.Context, query string, args ...interface{}) (*common.DatabaseResult, error) {
	startTime := time.Now()

	rows, err := d.db.QueryContext(ctx, query, args...)
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
func (d *DatabaseClient) QueryRow(ctx context.Context, query string, args ...interface{}) (*common.DatabaseRow, error) {
	startTime := time.Now()

	// Since we can't easily scan without knowing the structure,
	// we'll use Query and return the first row
	result, err := d.Query(ctx, query, args...)
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
func (d *DatabaseClient) BeginTx(ctx context.Context) (common.DatabaseTransaction, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &DatabaseTransaction{tx: tx}, nil
}

// Ping implements common.DatabaseClient.Ping
func (d *DatabaseClient) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// DatabaseTransaction implements common.DatabaseTransaction
type DatabaseTransaction struct {
	tx *sql.Tx
}

// Commit implements common.DatabaseTransaction.Commit
func (dt *DatabaseTransaction) Commit() error {
	return dt.tx.Commit()
}

// Rollback implements common.DatabaseTransaction.Rollback
func (dt *DatabaseTransaction) Rollback() error {
	return dt.tx.Rollback()
}

// Query implements common.DatabaseTransaction.Query
func (dt *DatabaseTransaction) Query(ctx context.Context, query string, args ...interface{}) (*common.DatabaseResult, error) {
	startTime := time.Now()

	rows, err := dt.tx.QueryContext(ctx, query, args...)
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
func (dt *DatabaseTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) (*common.DatabaseRow, error) {
	startTime := time.Now()

	// Since we can't easily scan without knowing the structure,
	// we'll use Query and return the first row
	result, err := dt.Query(ctx, query, args...)
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
