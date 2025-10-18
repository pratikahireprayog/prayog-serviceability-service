package shipcube

type ShipCubeRateRequest struct {
	OriginCountry      string  `json:"origin_country"`
	OriginPostalCode   string  `json:"origin_postal_code"`
	DestinationCountry string  `json:"destination_country"`
	DestinationPostal  string  `json:"destination_postal"`
	Weight             float64 `json:"weight"`
}

type ShipCubeRateResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Rates   []ShipCubeRate `json:"rates"`
	Errors  []string      `json:"errors,omitempty"`
}

type ShipCubeRate struct {
	ServiceType string  `json:"service_type"`
	Price       float64 `json:"price"`
	TransitDays int     `json:"transit_days"`
}