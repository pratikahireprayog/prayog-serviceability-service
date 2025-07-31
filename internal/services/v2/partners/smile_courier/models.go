package smile_courier

// ServiceabilityRequest represents a request to Smile Courier's serviceability API
type ServiceabilityRequest struct {
	FromPincode int `json:"fromPincode"`
	ToPincode   int `json:"toPincode"`
}

// ServiceabilityResponse represents a response from Smile Courier's serviceability API
type ServiceabilityResponse struct {
	Status  int                `json:"status"`
	Message string             `json:"message,omitempty"`
	Data    *ServiceabilityData `json:"data,omitempty"`
}

// ServiceabilityData contains the actual serviceability information
type ServiceabilityData struct {
	Serviceable      bool                    `json:"serviceable"`
	PincodeData      *PincodeData           `json:"pincodeData"`
	AvailableServices []AvailableService     `json:"availableServices"`
}

// PincodeData contains pincode-specific information
type PincodeData struct {
	Pincode         int                    `json:"pincode"`
	StateName       string                 `json:"stateName"`
	StateCode       string                 `json:"stateCode"`
	City            string                 `json:"city"`
	Zone            string                 `json:"zone"`
	Serviceability  ServiceabilityInfo     `json:"serviceability"`
	PincodeType     PincodeTypeInfo        `json:"pincodeType"`
	NewCity         string                 `json:"newCity"`
	DistrictIP      string                 `json:"districtIp"`
	IsCOD           bool                   `json:"isCod"`
}

// ServiceabilityInfo contains serviceability details
type ServiceabilityInfo struct {
	Serviceability string `json:"serviceability"`
}

// PincodeTypeInfo contains pincode type details
type PincodeTypeInfo struct {
	PincodeType string `json:"pincodeType"`
}

// AvailableService represents an available service
type AvailableService struct {
	ServiceName string `json:"serviceName"`
	Serviceable bool   `json:"serviceable"`
}




