package naqel

// GetTransitDaysRequest represents the SOAP request structure for GetTransitDays
type GetTransitDaysRequest struct {
	ClientInfo ClientInfo
	Origin     string
	Destination string
	LoadTypeID int
}

// ClientInfo represents client information for Naqel SOAP API
type ClientInfo struct {
	ClientAddress ClientAddress
	ClientContact ClientContact
	ClientID      string
	Password      string
	Version       string
}

// ClientAddress represents client address information
type ClientAddress struct {
	PhoneNumber     string
	NationalAddress string
	POBox           string
	ZipCode         string
	Fax             string
	Latitude        string
	Longitude       string
	ShipperName     string
	FirstAddress    string
	Location        string
	CountryCode     string
	CityCode        string
}

// ClientContact represents client contact information
type ClientContact struct {
	Name        string
	Email       string
	PhoneNumber string
	MobileNo    string
}

// GetTransitDaysResponse represents the SOAP response structure for GetTransitDays
type GetTransitDaysResponse struct {
	Days int `xml:"Body>GetTransitDaysResponse>GetTransitDaysResult>Days"`
}

// ServiceabilityRequest represents a Naqel serviceability check request
type ServiceabilityRequest struct {
	OriginCityCode      string
	DestinationCityCode string
	OriginCountryCode   string
	DestinationCountryCode string
}

// ServiceabilityResponse represents a Naqel serviceability check response
type ServiceabilityResponse struct {
	IsServiceable bool
	TransitDays   int
}

