package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SunilBaralPincode represents a row in sunil_baral_pincodes table
type SunilBaralPincode struct {
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

// SunilBaralRepository defines methods for querying Sunil Baral serviceability data
type SunilBaralRepository interface {
	// CheckServiceabilityByPincode checks if a pincode is serviceable
	CheckServiceabilityByPincode(ctx context.Context, pincode string) (*SunilBaralPincode, error)
}

// sunilBaralRepository implements SunilBaralRepository interface
type sunilBaralRepository struct {
	db        *gorm.DB
	tableName string
}

// NewSunilBaralRepository creates a new Sunil Baral repository
func NewSunilBaralRepository(db *gorm.DB, tableName string) SunilBaralRepository {
	if tableName == "" {
		tableName = "sunil_baral_pincodes"
	}
	return &sunilBaralRepository{
		db:        db,
		tableName: tableName,
	}
}

// CheckServiceabilityByPincode checks if a pincode is serviceable
func (r *sunilBaralRepository) CheckServiceabilityByPincode(ctx context.Context, pincode string) (*SunilBaralPincode, error) {
	var pincodeData SunilBaralPincode
	
	query := r.db.WithContext(ctx).Table(r.tableName).
		Where("pincode = ?", pincode).
		First(&pincodeData)

	if query.Error != nil {
		if query.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pincode %s not found in Sunil Baral database", pincode)
		}
		return nil, fmt.Errorf("failed to query Sunil Baral pincode: %w", query.Error)
	}

	return &pincodeData, nil
}

