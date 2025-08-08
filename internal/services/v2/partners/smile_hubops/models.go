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
	TATDays    int       `json:"tatDays"`
	Mode       int       `json:"mode"`
	Route      []HubInfo `json:"route"`
	CutoffTime *string   `json:"cutoffTime"`
}

// HubOpsResponse represents the response structure from Smile HubOps API
type HubOpsResponse struct {
	SourceHub                 *HubInfo  `json:"sourceHub"`
	SourceInternationalHub    *HubInfo  `json:"sourceInternationalHub"`
	Source3PLHub             *HubInfo  `json:"source3PLHub"`
	DestinationHub            *HubInfo  `json:"destinationHub"`
	DestinationInternationalHub *HubInfo `json:"destinationInternationalHub"`
	Destination3PLHub         *HubInfo  `json:"destination3PLHub"`
	Routes                    []Route   `json:"routes"`
	DeliveryAvailable         bool      `json:"deliveryAvailable"`
}

// HubServiceabilityData represents the hub serviceability data to be included in the response
type HubServiceabilityData struct {
	SourceHub                 *HubInfo  `json:"sourceHub,omitempty"`
	SourceInternationalHub    *HubInfo  `json:"sourceInternationalHub,omitempty"`
	Source3PLHub             *HubInfo  `json:"source3PLHub,omitempty"`
	DestinationHub            *HubInfo  `json:"destinationHub,omitempty"`
	DestinationInternationalHub *HubInfo `json:"destinationInternationalHub,omitempty"`
	Destination3PLHub         *HubInfo  `json:"destination3PLHub,omitempty"`
	Routes                    []Route   `json:"routes,omitempty"`
	DeliveryAvailable         bool      `json:"deliveryAvailable"`
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
