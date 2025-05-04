package domain

import "time"

// Area represents a geographic area
type Area struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	CityID uint   `json:"cityId"`
	Active bool   `json:"active"`
}

// City represents a city
type City struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	RegionID uint   `json:"regionId"`
	Active   bool   `json:"active"`
}

// Region represents an administrative region
type Region struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	CountryID uint   `json:"countryId"`
	Active    bool   `json:"active"`
}

// Country represents a country
type Country struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Active bool   `json:"active"`
}

// PostalCode represents a postal/ZIP code
type PostalCode struct {
	ID     uint   `json:"id"`
	Code   string `json:"code"`
	AreaID uint   `json:"areaId"`
	Active bool   `json:"active"`
}

// ServiceType represents a type of service offered
type ServiceType struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Code        string `json:"code"`
	Active      bool   `json:"active"`
}

// OrderType represents a type of order
type OrderType struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Code        string `json:"code"`
	Active      bool   `json:"active"`
}

// ServiceabilityRule represents a rule defining whether a service is available for a specific location and order type
type ServiceabilityRule struct {
	ID            uint      `json:"id"`
	PostalCodeID  uint      `json:"postalCodeId"`
	ServiceTypeID uint      `json:"serviceTypeId"`
	OrderTypeID   uint      `json:"orderTypeId"`
	IsServiceable bool      `json:"isServiceable"`
	EffectiveFrom time.Time `json:"effectiveFrom"`
	EffectiveTo   time.Time `json:"effectiveTo"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
