package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DharmendraPincode represents a row in dharmendra_pincodes table
type DharmendraPincode struct {
	PincodeID    string    `gorm:"column:pincode_id"`
	Pincode      int       `gorm:"column:pincode"`
	City         string    `gorm:"column:city"`
	HubCode      int       `gorm:"column:hubcode"`
	HubName      string    `gorm:"column:hubname"`
	DistrictName *string   `gorm:"column:districtname"`
	State        string    `gorm:"column:state"`
	Zone         string    `gorm:"column:zone"`
	FM           bool      `gorm:"column:fm"`      // Forward Manifest
	LM           bool      `gorm:"column:lm"`      // Last Mile
	ToPay        bool      `gorm:"column:topay"`   // To Pay
	COD          bool      `gorm:"column:cod"`     // Cash on Delivery
	Surface      bool      `gorm:"column:surface"` // Surface delivery
	Air          bool      `gorm:"column:air"`     // Air delivery
	Rail         bool      `gorm:"column:rail"`    // Rail delivery
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

// DharmendraRepository defines methods for querying Dharmendra serviceability data
type DharmendraRepository interface {
	// CheckServiceabilityByPincode checks if a pincode is serviceable
	CheckServiceabilityByPincode(ctx context.Context, pincode string) (*DharmendraPincode, error)
}

// dharmendraRepository implements DharmendraRepository interface
type dharmendraRepository struct {
	db        *gorm.DB
	tableName string
}

// NewDharmendraRepository creates a new Dharmendra repository
func NewDharmendraRepository(db *gorm.DB, tableName string) DharmendraRepository {
	if tableName == "" {
		tableName = "dharmendra_pincodes"
	}
	return &dharmendraRepository{
		db:        db,
		tableName: tableName,
	}
}

// CheckServiceabilityByPincode checks if a pincode is serviceable
func (r *dharmendraRepository) CheckServiceabilityByPincode(ctx context.Context, pincode string) (*DharmendraPincode, error) {
	var pincodeData DharmendraPincode
	
	query := r.db.WithContext(ctx).Table(r.tableName).
		Where("pincode = ?", pincode).
		First(&pincodeData)

	if query.Error != nil {
		if query.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pincode %s not found in Dharmendra database", pincode)
		}
		return nil, fmt.Errorf("failed to query Dharmendra pincode: %w", query.Error)
	}

	return &pincodeData, nil
}

