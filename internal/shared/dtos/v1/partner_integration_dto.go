package dtos

import "time"

// PartnerCapabilitiesResponse represents Partner Service capabilities response
type PartnerCapabilitiesResponse struct {
	PartnerID    string               `json:"partner_id"`
	Capabilities []CapabilityResponse `json:"capabilities"`
	Metadata     CapabilityMetadata   `json:"metadata"`
}

// CapabilityResponse represents a single capability
type CapabilityResponse struct {
	ServiceType     string   `json:"service_type"`
	ParcelCategory  string   `json:"parcel_category"`
	OperationTypes  []string `json:"operation_types"`
	PaymentModes    []string `json:"payment_modes"`
	DeliveryModes   []string `json:"delivery_modes"`
	MaxWeight       float64  `json:"max_weight"`
	MaxDimensions   string   `json:"max_dimensions"`
	IsActive        bool     `json:"is_active"`
	Priority        int      `json:"priority"`
	EffectiveRating float64  `json:"effective_rating"`
}

// CapabilityMetadata provides metadata about capabilities
type CapabilityMetadata struct {
	LastUpdated   time.Time `json:"last_updated"`
	TotalCount    int       `json:"total_count"`
	ActiveCount   int       `json:"active_count"`
	InactiveCount int       `json:"inactive_count"`
}

// PartnerAvailabilityRequest represents request for partner availability check
type PartnerAvailabilityRequest struct {
	LocationType   string  `json:"location_type" validate:"required,oneof=postal_code city state country"`
	LocationValue  string  `json:"location_value" validate:"required"`
	ServiceType    string  `json:"service_type,omitempty"`
	ParcelCategory string  `json:"parcel_category,omitempty"`
	RequestedDate  *string `json:"requested_date,omitempty"` // ISO 8601 format
}

// PartnerAvailabilityResponse represents partner availability response
type PartnerAvailabilityResponse struct {
	PartnerID       string                         `json:"partner_id"`
	IsAvailable     bool                           `json:"is_available"`
	Reason          string                         `json:"reason,omitempty"`
	AvailabilityMap map[string]ServiceAvailability `json:"availability_map,omitempty"`
	NextAvailable   *string                        `json:"next_available,omitempty"` // ISO 8601 format
}

// ServiceAvailability represents availability for a specific service
type ServiceAvailability struct {
	IsAvailable    bool    `json:"is_available"`
	CapacityUsed   float64 `json:"capacity_used"`             // 0.0 to 1.0
	EstimatedDelay *string `json:"estimated_delay,omitempty"` // Duration format like "2h30m"
}

// PartnersLocationRequest represents request for partners by location
type PartnersLocationRequest struct {
	LocationType     string   `json:"location_type" validate:"required,oneof=postal_code city state country"`
	LocationValue    string   `json:"location_value" validate:"required"`
	ServiceTypes     []string `json:"service_types,omitempty"`
	ParcelCategories []string `json:"parcel_categories,omitempty"`
	OnlyActive       bool     `json:"only_active"`
	Limit            int      `json:"limit,omitempty"`
	Offset           int      `json:"offset,omitempty"`
}

// PartnersLocationResponse represents response with partners by location
type PartnersLocationResponse struct {
	Partners   []PartnerSummary `json:"partners"`
	Total      int              `json:"total"`
	Pagination PaginationInfo   `json:"pagination"`
}

// PartnerSummary represents summarized partner information
type PartnerSummary struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	Status           string    `json:"status"`
	Rating           float64   `json:"rating"`
	ServiceTypes     []string  `json:"service_types"`
	ParcelCategories []string  `json:"parcel_categories"`
	MaxRadius        float64   `json:"max_radius_km"`
	LastActive       time.Time `json:"last_active"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Total   int  `json:"total"`
	HasNext bool `json:"has_next"`
}

// PartnerDetailsRequest represents request for detailed partner information
type PartnerDetailsRequest struct {
	PartnerID         string   `json:"partner_id" validate:"required"`
	IncludeInactive   bool     `json:"include_inactive"`
	ServiceTypeFilter []string `json:"service_type_filter,omitempty"`
}

// PartnerDetailsResponse represents detailed partner information
type PartnerDetailsResponse struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Type              string                 `json:"type"`
	Status            string                 `json:"status"`
	ContactInfo       ContactInfoResponse    `json:"contact_info"`
	ServiceAreas      []ServiceAreaResponse  `json:"service_areas"`
	Capabilities      []CapabilityResponse   `json:"capabilities"`
	OperatingHours    OperatingHoursResponse `json:"operating_hours"`
	DeliveryTimeSlots []TimeSlotResponse     `json:"delivery_time_slots"`
	MaxRadius         float64                `json:"max_radius_km"`
	Metrics           PartnerMetrics         `json:"metrics"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// ContactInfoResponse represents partner contact information
type ContactInfoResponse struct {
	PrimaryPhone   string `json:"primary_phone"`
	SecondaryPhone string `json:"secondary_phone,omitempty"`
	Email          string `json:"email"`
	Address        string `json:"address"`
	City           string `json:"city"`
	State          string `json:"state"`
	PostalCode     string `json:"postal_code"`
	Country        string `json:"country"`
}

// ServiceAreaResponse represents a service area
type ServiceAreaResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"` // postal_code, city, state, etc.
	Code      string  `json:"code"`
	Priority  int     `json:"priority"`
	IsActive  bool    `json:"is_active"`
	MaxRadius float64 `json:"max_radius_km"`
}

// OperatingHoursResponse represents operating hours
type OperatingHoursResponse struct {
	Monday    DayHoursResponse `json:"monday"`
	Tuesday   DayHoursResponse `json:"tuesday"`
	Wednesday DayHoursResponse `json:"wednesday"`
	Thursday  DayHoursResponse `json:"thursday"`
	Friday    DayHoursResponse `json:"friday"`
	Saturday  DayHoursResponse `json:"saturday"`
	Sunday    DayHoursResponse `json:"sunday"`
	Timezone  string           `json:"timezone"`
}

// DayHoursResponse represents hours for a single day
type DayHoursResponse struct {
	IsOpen    bool   `json:"is_open"`
	OpenTime  string `json:"open_time,omitempty"`  // "09:00"
	CloseTime string `json:"close_time,omitempty"` // "18:00"
}

// TimeSlotResponse represents a delivery time slot
type TimeSlotResponse struct {
	ID        string `json:"id"`
	StartTime string `json:"start_time"` // "09:00"
	EndTime   string `json:"end_time"`   // "12:00"
	SlotType  string `json:"slot_type"`  // morning, afternoon, evening
	IsActive  bool   `json:"is_active"`
	Capacity  int    `json:"capacity"`
}

// PartnerMetrics represents partner performance metrics
type PartnerMetrics struct {
	OverallRating         float64   `json:"overall_rating"`
	DeliveryRating        float64   `json:"delivery_rating"`
	TimelinessRating      float64   `json:"timeliness_rating"`
	QualityRating         float64   `json:"quality_rating"`
	TotalDeliveries       int       `json:"total_deliveries"`
	SuccessfulDeliveries  int       `json:"successful_deliveries"`
	SuccessRate           float64   `json:"success_rate"`
	AverageDeliveryTime   string    `json:"average_delivery_time"`
	LastPerformanceUpdate time.Time `json:"last_performance_update"`
}
