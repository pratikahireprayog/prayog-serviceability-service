package domain

import (
	"testing"
	"time"
)

func TestCountry(t *testing.T) {
	country := Country{
		ID:        1,
		Code:      "IN",
		Name:      "India",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if country.ID != 1 {
		t.Errorf("Expected country ID to be 1, got %d", country.ID)
	}

	if country.Code != "IN" {
		t.Errorf("Expected country code to be IN, got %s", country.Code)
	}

	if country.Name != "India" {
		t.Errorf("Expected country name to be India, got %s", country.Name)
	}
}

func TestAdministrativeRegion(t *testing.T) {
	region := AdministrativeRegion{
		ID:        1,
		CountryID: 1,
		Code:      "KA",
		Name:      "Karnataka",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if region.ID != 1 {
		t.Errorf("Expected region ID to be 1, got %d", region.ID)
	}

	if region.CountryID != 1 {
		t.Errorf("Expected region CountryID to be 1, got %d", region.CountryID)
	}

	if region.Code != "KA" {
		t.Errorf("Expected region code to be KA, got %s", region.Code)
	}

	if region.Name != "Karnataka" {
		t.Errorf("Expected region name to be Karnataka, got %s", region.Name)
	}
}

func TestCity(t *testing.T) {
	city := City{
		ID:                     1,
		AdministrativeRegionID: 1,
		Name:                   "Bangalore",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if city.ID != 1 {
		t.Errorf("Expected city ID to be 1, got %d", city.ID)
	}

	if city.AdministrativeRegionID != 1 {
		t.Errorf("Expected city AdministrativeRegionID to be 1, got %d", city.AdministrativeRegionID)
	}

	if city.Name != "Bangalore" {
		t.Errorf("Expected city name to be Bangalore, got %s", city.Name)
	}
}

func TestArea(t *testing.T) {
	area := Area{
		ID:        1,
		CityID:    1,
		Name:      "Koramangala",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if area.ID != 1 {
		t.Errorf("Expected area ID to be 1, got %d", area.ID)
	}

	if area.CityID != 1 {
		t.Errorf("Expected area CityID to be 1, got %d", area.CityID)
	}

	if area.Name != "Koramangala" {
		t.Errorf("Expected area name to be Koramangala, got %s", area.Name)
	}
}

func TestPostalCode(t *testing.T) {
	postalCode := PostalCode{
		ID:        1,
		Code:      "560034",
		AreaID:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if postalCode.ID != 1 {
		t.Errorf("Expected postal code ID to be 1, got %d", postalCode.ID)
	}

	if postalCode.Code != "560034" {
		t.Errorf("Expected postal code to be 560034, got %s", postalCode.Code)
	}

	if postalCode.AreaID != 1 {
		t.Errorf("Expected postal code AreaID to be 1, got %d", postalCode.AreaID)
	}
}

func TestOrderType(t *testing.T) {
	orderType := OrderType{
		ID:        1,
		Code:      "DEL",
		Name:      "Delivery",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if orderType.ID != 1 {
		t.Errorf("Expected order type ID to be 1, got %d", orderType.ID)
	}

	if orderType.Code != "DEL" {
		t.Errorf("Expected order type code to be DEL, got %s", orderType.Code)
	}

	if orderType.Name != "Delivery" {
		t.Errorf("Expected order type name to be Delivery, got %s", orderType.Name)
	}
}

func TestServiceType(t *testing.T) {
	serviceType := ServiceType{
		ID:          1,
		Code:        "STD",
		Name:        "Standard Delivery",
		Description: "Standard delivery service",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if serviceType.ID != 1 {
		t.Errorf("Expected service type ID to be 1, got %d", serviceType.ID)
	}

	if serviceType.Code != "STD" {
		t.Errorf("Expected service type code to be STD, got %s", serviceType.Code)
	}

	if serviceType.Name != "Standard Delivery" {
		t.Errorf("Expected service type name to be Standard Delivery, got %s", serviceType.Name)
	}

	if serviceType.Description != "Standard delivery service" {
		t.Errorf("Expected service type description to be 'Standard delivery service', got %s", serviceType.Description)
	}
}

func TestServiceAvailability(t *testing.T) {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)

	serviceAvailability := ServiceAvailability{
		ID:             1,
		LocationType:   "city",
		LocationID:     1,
		OrderTypeID:    1,
		ServiceTypeID:  1,
		IsAvailable:    true,
		EffectiveFrom:  now,
		EffectiveTo:    tomorrow,
		AdditionalData: `{"priority": "high"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if serviceAvailability.ID != 1 {
		t.Errorf("Expected service availability ID to be 1, got %d", serviceAvailability.ID)
	}

	if serviceAvailability.LocationType != "city" {
		t.Errorf("Expected location type to be 'city', got %s", serviceAvailability.LocationType)
	}

	if serviceAvailability.LocationID != 1 {
		t.Errorf("Expected location ID to be 1, got %d", serviceAvailability.LocationID)
	}

	if serviceAvailability.OrderTypeID != 1 {
		t.Errorf("Expected order type ID to be 1, got %d", serviceAvailability.OrderTypeID)
	}

	if serviceAvailability.ServiceTypeID != 1 {
		t.Errorf("Expected service type ID to be 1, got %d", serviceAvailability.ServiceTypeID)
	}

	if !serviceAvailability.IsAvailable {
		t.Errorf("Expected service availability to be true, got %t", serviceAvailability.IsAvailable)
	}

	if serviceAvailability.AdditionalData != `{"priority": "high"}` {
		t.Errorf("Expected additional data to be '{\"priority\": \"high\"}', got %s", serviceAvailability.AdditionalData)
	}
}
