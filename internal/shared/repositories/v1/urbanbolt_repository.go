package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// UrbanboltPincode represents a row in urbanbolt_serviceability_pincodes table (or configured table)
// Based on user provided schema: id, pincode, inbound, outbound, rtn, isActive, serviceCenter, city, state, region, zone, routeCode, serviceType
type UrbanboltPincode struct {
	ID            int       `gorm:"column:id;primaryKey"`
	Pincode       string    `gorm:"column:pincode"`
	Inbound       bool      `gorm:"column:inbound"`
	Outbound      bool      `gorm:"column:outbound"`
	RTN           bool      `gorm:"column:rtn"`
	IsActive      bool      `gorm:"column:is_active"`
	ServiceCenter string    `gorm:"column:service_center"`
	City          string    `gorm:"column:city"`
	State         string    `gorm:"column:state"`
	Region        string    `gorm:"column:region"`
	Zone          string    `gorm:"column:zone"`
	RouteCode     string    `gorm:"column:route_code"`
	ServiceType   string    `gorm:"column:service_type"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// UrbanboltRepository defines methods for querying Urbanbolt serviceability data
type UrbanboltRepository interface {
	// CheckServiceabilityByPincode checks if a pincode is serviceable
	CheckServiceabilityByPincode(ctx context.Context, pincode string) (*UrbanboltPincode, error)
}

// urbanboltRepository implements UrbanboltRepository interface
type urbanboltRepository struct {
	db        *gorm.DB
	tableName string
}

// NewUrbanboltRepository creates a new Urbanbolt repository
func NewUrbanboltRepository(db *gorm.DB, tableName string) UrbanboltRepository {
	if tableName == "" {
		tableName = "urbanbolt_serviceability_pincodes" // Default fallback
	}
	return &urbanboltRepository{
		db:        db,
		tableName: tableName,
	}
}

// CheckServiceabilityByPincode checks if a pincode is serviceable
func (r *urbanboltRepository) CheckServiceabilityByPincode(ctx context.Context, pincode string) (*UrbanboltPincode, error) {
	var pincodeData UrbanboltPincode
	
	// Assuming pincode in DB is stored as number or string. The struct uses string.
	// If the DB column is INT, gorm usually handles string->int conversion if simple.
	// However, to be safe, we'll pass it as is.
	
	query := r.db.WithContext(ctx).Table(r.tableName).
		Where("pincode = ? AND is_active = ?", pincode, true).
		First(&pincodeData)

	if query.Error != nil {
		if query.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pincode %s not found in Urbanbolt database or not serviceable", pincode)
		}
		return nil, fmt.Errorf("failed to query Urbanbolt pincode: %w", query.Error)
	}

	return &pincodeData, nil
}
