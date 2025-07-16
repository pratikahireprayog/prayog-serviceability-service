package models

import (
	"fmt"

	"gorm.io/gorm"
)

// NearestHubLocation represents the mapping between postal codes and their nearest hub
type NearestHubLocation struct {
	PostalCode  int      `json:"postal_code" gorm:"primaryKey;column:postal_code"`
	Address     *string  `json:"address,omitempty" gorm:"type:text;column:address"`
	CentroidLat *float64 `json:"centroid_lat,omitempty" gorm:"type:decimal(9,6);column:centroid_lat"`
	CentroidLng *float64 `json:"centroid_lng,omitempty" gorm:"type:decimal(9,6);column:centroid_lng"`

	// Original city code field (renamed from hub_city_code)
	CityCode *string `json:"city_code,omitempty" gorm:"type:text;column:city_code"`

	// Hub location fields (renamed from international_hub_*)
	HubPostalCode  *int     `json:"hub_postal_code,omitempty" gorm:"column:hub_postal_code"`
	HubCentroidLat *float64 `json:"hub_centroid_lat,omitempty" gorm:"type:decimal(9,6);column:hub_centroid_lat"`
	HubCentroidLng *float64 `json:"hub_centroid_lng,omitempty" gorm:"type:decimal(9,6);column:hub_centroid_lng"`
	HubCityCode    *string  `json:"hub_city_code,omitempty" gorm:"type:text;column:hub_city_code"`

	// International hub flag
	IsInternationalHub *bool `json:"is_international_hub,omitempty" gorm:"column:is_international_hub;default:false"`

	// Hub contact and location fields
	HubContactPersonName  *string  `json:"hub_contact_person_name,omitempty" gorm:"type:text;column:hub_contact_person_name"`
	HubContactPersonPhone *string  `json:"hub_contact_person_phone,omitempty" gorm:"type:text;column:hub_contact_person_phone"`
	HubContactPersonEmail *string  `json:"hub_contact_person_email,omitempty" gorm:"type:text;column:hub_contact_person_email"`
	HubStreet             *string  `json:"hub_street,omitempty" gorm:"type:text;column:hub_street"`
	HubLandmark           *string  `json:"hub_landmark,omitempty" gorm:"type:text;column:hub_landmark"`
	HubCity               *string  `json:"hub_city,omitempty" gorm:"type:text;column:hub_city"`
	HubState              *string  `json:"hub_state,omitempty" gorm:"type:text;column:hub_state"`
	HubCountry            *string  `json:"hub_country,omitempty" gorm:"type:text;column:hub_country"`
	HubLat                *float64 `json:"hub_lat,omitempty" gorm:"type:decimal(9,6);column:hub_lat"`
	HubLng                *float64 `json:"hub_lng,omitempty" gorm:"type:decimal(9,6);column:hub_lng"`
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

	// Validate hub centroid latitude range if provided
	if n.HubCentroidLat != nil && (*n.HubCentroidLat < -90 || *n.HubCentroidLat > 90) {
		return fmt.Errorf("hub centroid latitude must be between -90 and 90")
	}

	// Validate hub centroid longitude range if provided
	if n.HubCentroidLng != nil && (*n.HubCentroidLng < -180 || *n.HubCentroidLng > 180) {
		return fmt.Errorf("hub centroid longitude must be between -180 and 180")
	}

	// Validate hub latitude range if provided
	if n.HubLat != nil && (*n.HubLat < -90 || *n.HubLat > 90) {
		return fmt.Errorf("hub latitude must be between -90 and 90")
	}

	// Validate hub longitude range if provided
	if n.HubLng != nil && (*n.HubLng < -180 || *n.HubLng > 180) {
		return fmt.Errorf("hub longitude must be between -180 and 180")
	}

	// Validate hub postal code is positive if provided
	if n.HubPostalCode != nil && *n.HubPostalCode <= 0 {
		return fmt.Errorf("hub postal code must be positive")
	}

	// Validate email format if provided
	if n.HubContactPersonEmail != nil && *n.HubContactPersonEmail != "" {
		// Simple email validation
		email := *n.HubContactPersonEmail
		if len(email) < 3 || !contains(email, "@") || !contains(email, ".") {
			return fmt.Errorf("hub contact person email must be a valid email address")
		}
	}

	return nil
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
	CityCode    *string  `json:"city_code,omitempty"`

	// Hub information
	HubInfo *HubInfo `json:"hub_info,omitempty"`

	// Hub contact and location information
	HubContactInfo *HubContactInfo `json:"hub_contact_info,omitempty"`
}

// HubInfo represents hub information
type HubInfo struct {
	PostalCode         *int     `json:"postal_code,omitempty"`
	CentroidLat        *float64 `json:"centroid_lat,omitempty"`
	CentroidLng        *float64 `json:"centroid_lng,omitempty"`
	CityCode           *string  `json:"city_code,omitempty"`
	IsInternationalHub *bool    `json:"is_international_hub,omitempty"`
}

// HubContactInfo represents hub contact and location information
type HubContactInfo struct {
	ContactPersonName  *string  `json:"contact_person_name,omitempty"`
	ContactPersonPhone *string  `json:"contact_person_phone,omitempty"`
	ContactPersonEmail *string  `json:"contact_person_email,omitempty"`
	Street             *string  `json:"street,omitempty"`
	Landmark           *string  `json:"landmark,omitempty"`
	City               *string  `json:"city,omitempty"`
	State              *string  `json:"state,omitempty"`
	Country            *string  `json:"country,omitempty"`
	Lat                *float64 `json:"lat,omitempty"`
	Lng                *float64 `json:"lng,omitempty"`
}

// ToHubLocationInfo converts NearestHubLocation to HubLocationInfo
func (n *NearestHubLocation) ToHubLocationInfo() *HubLocationInfo {
	hubLocationInfo := &HubLocationInfo{
		PostalCode:  n.PostalCode,
		Address:     n.Address,
		CentroidLat: n.CentroidLat,
		CentroidLng: n.CentroidLng,
		CityCode:    n.CityCode,
	}

	// Add hub info if available
	if n.HubPostalCode != nil {
		hubLocationInfo.HubInfo = &HubInfo{
			PostalCode:         n.HubPostalCode,
			CentroidLat:        n.HubCentroidLat,
			CentroidLng:        n.HubCentroidLng,
			CityCode:           n.HubCityCode,
			IsInternationalHub: n.IsInternationalHub,
		}
	}

	// Add hub contact info if available
	if n.HubContactPersonName != nil || n.HubCity != nil || n.HubLat != nil {
		hubLocationInfo.HubContactInfo = &HubContactInfo{
			ContactPersonName:  n.HubContactPersonName,
			ContactPersonPhone: n.HubContactPersonPhone,
			ContactPersonEmail: n.HubContactPersonEmail,
			Street:             n.HubStreet,
			Landmark:           n.HubLandmark,
			City:               n.HubCity,
			State:              n.HubState,
			Country:            n.HubCountry,
			Lat:                n.HubLat,
			Lng:                n.HubLng,
		}
	}

	return hubLocationInfo
}
