package smile_hubops

import (
)

// HubOpsRequest represents the request structure for Smile HubOps API
type HubOpsRequest struct {
	SourcePostalCode      int `json:"sourcePostalCode"`
	DestinationPostalCode int `json:"destinationPostalCode"`
}

// HubInfo represents hub information in the response
type HubInfo struct {
	PremiseID            int     `json:"premiseId"`
	PremiseName          string  `json:"premiseName"`
	ParentPremiseName    *string `json:"parentPremiseName"`
	PersonalNumber       int64   `json:"personalNumber"`
	OfficialNumber       int64   `json:"officialNumber"`
	City                 string  `json:"city"`
	Address              *string `json:"address"`
	AddressLine1         *string `json:"addressLine1"`
	AddressLine2         *string `json:"addressLine2"`
	BillingCycle         *string `json:"billingCycle"`
	Pincode              int     `json:"pincode"`
	State                string  `json:"state"`
	Zone                 string  `json:"zone"`
	Type                 string  `json:"type"`
	ParentID             *int    `json:"parentId"`
	ParentIDAir          *int    `json:"parentIdAir"`
	GST                  *string `json:"gst"`
	StateCode            *string `json:"stateCode"`
	CutoffTime           *string `json:"cutoffTime"`
	IsMetro              *bool   `json:"isMetro"`
	FOV                  *float64 `json:"fov"`
	COD                  *float64 `json:"cod"`
	Premium              *float64 `json:"premium"`
	Latitude             *string `json:"latitude"`
	Longitude            *string `json:"longitude"`
	PersonalEmailID      *string `json:"personalEmailId"`
	OfficialEmailID      *string `json:"officialEmailId"`
	PAN                  *string `json:"pan"`
	CPType               *string `json:"cpType"`
	RateCardType         *string `json:"rateCardType"`
	ForceUpdateAllowed   bool    `json:"forceUpdateAllowed"`
	IsTerminalHub        bool    `json:"isTerminalHub"`
	CreatedDate          *string `json:"createdDate"`
	Status               string  `json:"status"`
	Areas                []interface{} `json:"areas"`
	HubType              string  `json:"hubType"`
	HubMode              string  `json:"hubMode"`
	RateCardID           *int    `json:"rateCardId"`
	HOID                 *int    `json:"hoId"`
	HubPincodeMap        []interface{} `json:"hubPincodeMap"`
	WalletMappingCustID  *int    `json:"walletMappingCustId"`
	MiscellaneousDetails *interface{} `json:"miscellaneousDetails"`
	CenterMap            *interface{} `json:"centerMap"`
	SPID                 *int    `json:"spId"`
}

// Route represents a delivery route
type Route struct {
	TATDays    int       `json:"tat_days"`
	Mode       int       `json:"mode"`
	Route      []HubInfo `json:"route"`
	CutoffTime *string   `json:"cutoff_time"`
}

// HubOpsResponse represents the response structure from Smile HubOps API
type HubOpsResponse struct {
	SourceHub                 *HubInfo  `json:"source_hub"`
	SourceInternationalHub    *HubInfo  `json:"source_international_hub"`
	Source3PLHub             *HubInfo  `json:"source_3pl_hub"`
	DestinationHub            *HubInfo  `json:"destination_hub"`
	DestinationInternationalHub *HubInfo `json:"destination_international_hub"`
	Destination3PLHub         *HubInfo  `json:"destination_3pl_hub"`
	Routes                    []Route   `json:"routes"`
	DeliveryAvailable         bool      `json:"delivery_available"`
}

// HubServiceabilityData represents the hub serviceability data to be included in the response
type HubServiceabilityData struct {
	SourceHub                 *HubInfo  `json:"source_hub,omitempty"`
	SourceInternationalHub    *HubInfo  `json:"source_international_hub,omitempty"`
	Source3PLHub             *HubInfo  `json:"source_3pl_hub,omitempty"`
	DestinationHub            *HubInfo  `json:"destination_hub,omitempty"`
	DestinationInternationalHub *HubInfo `json:"destination_international_hub,omitempty"`
	Destination3PLHub         *HubInfo  `json:"destination_3pl_hub,omitempty"`
	Routes                    []Route   `json:"routes,omitempty"`
	DeliveryAvailable         bool      `json:"delivery_available"`
}

// Service represents a service offered by Smile HubOps
type Service struct {
	ServiceCode   string            `json:"service_code"`
	ServiceName   string            `json:"service_name"`
	TATDays       int               `json:"tat_days"`
	IsCOD         bool              `json:"is_cod"`
	Pickup        bool              `json:"pickup"`
	Delivery      bool              `json:"delivery"`
	Insurance     bool              `json:"insurance"`
	ProductTypes  map[string]bool   `json:"product_types"`
	DeliveryModes map[string]bool   `json:"delivery_modes"`
	Pricing       *ServicePricing   `json:"pricing,omitempty"`
}

// ServicePricing represents pricing information for services
type ServicePricing struct {
	BaseCost      float64 `json:"base_cost"`
	Currency      string  `json:"currency"`
	CODCharges    float64 `json:"cod_charges,omitempty"`
	FuelSurcharge float64 `json:"fuel_surcharge,omitempty"`
}

// Capabilities represents the capabilities structure for Smile HubOps
type Capabilities struct {
	HubServiceability *HubServiceabilityData `json:"hub_serviceability,omitempty"`
	DeliveryAvailable bool                   `json:"delivery_available"`
	RouteCount        int                    `json:"route_count"`
	AirModeAvailable  bool                   `json:"air_mode_available"`
	SurfaceModeAvailable bool                `json:"surface_mode_available"`
}
