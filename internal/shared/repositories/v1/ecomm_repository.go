package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// EcommPincode represents a row in ecomm_serviceability_pincodes table
type EcommPincode struct {
	PincodeID     string    `gorm:"column:pincode_id"`
	Pincode       int       `gorm:"column:pincode"`
	City          string    `gorm:"column:city"`
	State         string    `gorm:"column:state"`
	COD           bool      `gorm:"column:cod"`           // Cash on Delivery
	FM            bool      `gorm:"column:fm"`            // Forward Manifest (Pickup)
	LM            bool      `gorm:"column:lm"`            // Last Mile (Delivery)
	IsServiceable bool      `gorm:"column:is_serviceable"` // Serviceable flag
	Serviceability string   `gorm:"column:serviceability"` // Serviceability type (Ecom)
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

// EcommRepository defines methods for querying Ecomm serviceability data
type EcommRepository interface {
	// CheckServiceabilityByPincode checks if a pincode is serviceable
	CheckServiceabilityByPincode(ctx context.Context, pincode string) (*EcommPincode, error)
}

// ecommRepository implements EcommRepository interface
type ecommRepository struct {
	db        *gorm.DB
	tableName string
}

// NewEcommRepository creates a new Ecomm repository
func NewEcommRepository(db *gorm.DB, tableName string) EcommRepository {
	if tableName == "" {
		tableName = "ecomm_serviceability_pincodes"
	}
	return &ecommRepository{
		db:        db,
		tableName: tableName,
	}
}

// CheckServiceabilityByPincode checks if a pincode is serviceable
func (r *ecommRepository) CheckServiceabilityByPincode(ctx context.Context, pincode string) (*EcommPincode, error) {
	var pincodeData EcommPincode
	
	query := r.db.WithContext(ctx).Table(r.tableName).
		Where("pincode = ? AND is_serviceable = ?", pincode, true).
		First(&pincodeData)

	if query.Error != nil {
		if query.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pincode %s not found in Ecomm database or not serviceable", pincode)
		}
		return nil, fmt.Errorf("failed to query Ecomm pincode: %w", query.Error)
	}

	return &pincodeData, nil
}

