package models

import (
	"fmt"

	"gorm.io/gorm"
)

// NearestHubLocation represents the mapping between postal codes and their nearest international hub
type NearestHubLocation struct {
	PostalCode                  int      `json:"postal_code" gorm:"primaryKey;column:postal_code"`
	Address                     *string  `json:"address,omitempty" gorm:"type:text;column:address"`
	CentroidLat                 *float64 `json:"centroid_lat,omitempty" gorm:"type:decimal(9,6);column:centroid_lat"`
	CentroidLng                 *float64 `json:"centroid_lng,omitempty" gorm:"type:decimal(9,6);column:centroid_lng"`
	InternationalHubPostalCode  *int     `json:"international_hub_postal_code,omitempty" gorm:"column:international_hub_postal_code"`
	InternationalHubAddress     *string  `json:"international_hub_address,omitempty" gorm:"type:text;column:international_hub_address"`
	InternationalHubCentroidLat *float64 `json:"international_hub_centroid_lat,omitempty" gorm:"type:decimal(9,6);column:international_hub_centroid_lat"`
	InternationalHubCentroidLng *float64 `json:"international_hub_centroid_lng,omitempty" gorm:"type:decimal(9,6);column:international_hub_centroid_lng"`
	InternationalHubCityCode    *string  `json:"international_hub_city_code,omitempty" gorm:"type:text;column:international_hub_city_code"`
	HubCityCode                 *string  `json:"hub_city_code,omitempty" gorm:"type:text;column:hub_city_code"`
}

// TableName returns the table name for NearestHubLocation
func (NearestHubLocation) TableName() string {
	return "nearest_hub_locations"
}

// ValidateBusinessRules performs custom business rule validation for NearestHubLocation
func (n *NearestHubLocation) ValidateBusinessRules() error {
	// Validate postal code is positive
	if n.PostalCode <= 0 {
		return fmt.Errorf("postal code must be positive")
	}

	// Validate latitude range if provided
	if n.CentroidLat != nil && (*n.CentroidLat < -90 || *n.CentroidLat > 90) {
		return fmt.Errorf("centroid latitude must be between -90 and 90")
	}

	// Validate longitude range if provided
	if n.CentroidLng != nil && (*n.CentroidLng < -180 || *n.CentroidLng > 180) {
		return fmt.Errorf("centroid longitude must be between -180 and 180")
	}

	// Validate international hub latitude range if provided
	if n.InternationalHubCentroidLat != nil && (*n.InternationalHubCentroidLat < -90 || *n.InternationalHubCentroidLat > 90) {
		return fmt.Errorf("international hub centroid latitude must be between -90 and 90")
	}

	// Validate international hub longitude range if provided
	if n.InternationalHubCentroidLng != nil && (*n.InternationalHubCentroidLng < -180 || *n.InternationalHubCentroidLng > 180) {
		return fmt.Errorf("international hub centroid longitude must be between -180 and 180")
	}

	// Validate international hub postal code is positive if provided
	if n.InternationalHubPostalCode != nil && *n.InternationalHubPostalCode <= 0 {
		return fmt.Errorf("international hub postal code must be positive")
	}

	return nil
}

// BeforeCreate hook to validate before creating
func (n *NearestHubLocation) BeforeCreate(tx *gorm.DB) error {
	return n.ValidateBusinessRules()
}

// BeforeUpdate hook to validate before updating
func (n *NearestHubLocation) BeforeUpdate(tx *gorm.DB) error {
	return n.ValidateBusinessRules()
}

// HubLocationInfo represents hub location information for response
type HubLocationInfo struct {
	PostalCode  int      `json:"postal_code"`
	Address     *string  `json:"address,omitempty"`
	CentroidLat *float64 `json:"centroid_lat,omitempty"`
	CentroidLng *float64 `json:"centroid_lng,omitempty"`
	HubCityCode *string  `json:"hub_city_code,omitempty"`

	// International hub information
	InternationalHub *InternationalHubInfo `json:"international_hub,omitempty"`
}

// InternationalHubInfo represents international hub information
type InternationalHubInfo struct {
	PostalCode  *int     `json:"postal_code,omitempty"`
	Address     *string  `json:"address,omitempty"`
	CentroidLat *float64 `json:"centroid_lat,omitempty"`
	CentroidLng *float64 `json:"centroid_lng,omitempty"`
	CityCode    *string  `json:"city_code,omitempty"`
}

// ToHubLocationInfo converts NearestHubLocation to HubLocationInfo
func (n *NearestHubLocation) ToHubLocationInfo() *HubLocationInfo {
	hubInfo := &HubLocationInfo{
		PostalCode:  n.PostalCode,
		Address:     n.Address,
		CentroidLat: n.CentroidLat,
		CentroidLng: n.CentroidLng,
		HubCityCode: n.HubCityCode,
	}

	// Add international hub info if available
	if n.InternationalHubPostalCode != nil {
		hubInfo.InternationalHub = &InternationalHubInfo{
			PostalCode:  n.InternationalHubPostalCode,
			Address:     n.InternationalHubAddress,
			CentroidLat: n.InternationalHubCentroidLat,
			CentroidLng: n.InternationalHubCentroidLng,
			CityCode:    n.InternationalHubCityCode,
		}
	}

	return hubInfo
}
