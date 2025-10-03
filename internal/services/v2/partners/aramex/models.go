package aramex

import "time"

// ServiceabilityRequest represents a request to Aramex's serviceability API
type ServiceabilityRequest struct {
	ClientInfo     ClientInfo     `json:"ClientInfo"`
	Address        Address        `json:"Address"`
	ServiceDetails ServiceDetails `json:"ServiceDetails"`
	Transaction    Transaction    `json:"Transaction"`
}

// ClientInfo represents client information for Aramex API
type ClientInfo struct {
	UserName           string `json:"UserName"`
	Password           string `json:"Password"`
	Version            string `json:"Version"`
	AccountNumber      string `json:"AccountNumber"`
	AccountPin         string `json:"AccountPin"`
	AccountEntity      string `json:"AccountEntity"`
	AccountCountryCode string `json:"AccountCountryCode"`
	Source             int    `json:"Source"`
}

// Address represents address information for serviceability check
type Address struct {
	Line1               string  `json:"Line1"`
	Line2               string  `json:"Line2"`
	Line3               string  `json:"Line3"`
	City                string  `json:"City"`
	StateOrProvinceCode string  `json:"StateOrProvinceCode"`
	PostCode            string  `json:"PostCode"`
	CountryCode         string  `json:"CountryCode"`
	Longitude           float64 `json:"Longitude"`
	Latitude            float64 `json:"Latitude"`
	BuildingNumber      *string `json:"BuildingNumber"`
	BuildingName        *string `json:"BuildingName"`
	Floor               *string `json:"Floor"`
	Apartment           *string `json:"Apartment"`
	POBox               *string `json:"POBox"`
	Description         *string `json:"Description"`
}

// ServiceDetails represents service details for Aramex
type ServiceDetails struct {
	ProductGroup string `json:"ProductGroup"`
	ProductType  string `json:"ProductType"`
	ServiceMode  int    `json:"ServiceMode"`
}

// Transaction represents transaction information
type Transaction struct {
	Reference1 string `json:"Reference1"`
	Reference2 string `json:"Reference2"`
	Reference3 string `json:"Reference3"`
	Reference4 string `json:"Reference4"`
	Reference5 string `json:"Reference5"`
}

// ServiceabilityResponse represents a response from Aramex's serviceability API
type ServiceabilityResponse struct {
	IsAddressServiced bool        `json:"IsAddressServiced"`
	HasErrors         bool        `json:"HasErrors"`
	Errors            []string    `json:"Errors"`
	Transaction       Transaction `json:"Transaction"`
	// Note: Aramex response structure might include additional fields
}

// AramexConfig represents the configuration for Aramex adapter
type AramexConfig struct {
	Enabled            bool          `json:"enabled"`
	BaseURL            string        `json:"base_url"`
	Username           string        `json:"username"`
	Password           string        `json:"password"`
	AccountNumber      string        `json:"account_number"`
	AccountPin         string        `json:"account_pin"`
	AccountEntity      string        `json:"account_entity"`
	AccountCountryCode string        `json:"account_country_code"`
	Source             int           `json:"source"`
	Timeout            time.Duration `json:"timeout"`
}

// DefaultAramexConfig returns default configuration for Aramex
func DefaultAramexConfig() AramexConfig {
	return AramexConfig{
		Enabled:            true,
		BaseURL:            "https://ws.dev.aramex.net",
		Username:           "test.api@aramex.com",
		Password:           "Aramex@12345",
		AccountNumber:      "60531487",
		AccountPin:         "654654",
		AccountEntity:      "BOM",
		AccountCountryCode: "IN",
		Source:             24,
		Timeout:            30 * time.Second,
	}
}
