package constants

// Query Types
const (
	QueryTypeGenericLocation   = "generic_location"
	QueryTypeOriginDestination = "origin_destination"
)

// Parcel Categories
const (
	ParcelCategoryEcom    = "ecom"
	ParcelCategoryCourier = "courier"
	ParcelCategoryCargo   = "cargo"
)

// Service Types
const (
	ServiceTypeSDD          = "SDD"          // Same Day Delivery
	ServiceTypeNDD          = "NDD"          // Next Day Delivery
	ServiceTypeStandard     = "Standard"     // Standard Delivery
	ServiceTypeExpress      = "Express"      // Express Delivery
	ServiceTypeVayuQuick    = "vayuquick"    // VayuQuick
	ServiceTypeVayuQuickPro = "vayuquickpro" // VayuQuick Pro
)

// Operation Types
const (
	OperationTypePickup   = "pickup"
	OperationTypeDelivery = "delivery"
)

// Payment Modes
const (
	PaymentModeOnline = "ONLINE"
	PaymentModeCOD    = "COD"
)

// Delivery Modes
const (
	DeliveryModeAir     = "AIR"
	DeliveryModeSurface = "SURFACE"
	DeliveryModeRail    = "RAIL"
)

// HTTP Status Codes
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusAccepted            = 202
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusMethodNotAllowed    = 405
	StatusConflict            = 409
	StatusUnprocessableEntity = 422
	StatusTooManyRequests     = 429
	StatusInternalServerError = 500
	StatusBadGateway          = 502
	StatusServiceUnavailable  = 503
	StatusGatewayTimeout      = 504
)

// Error Codes
const (
	ErrorCodeInvalidPostalCode         = "INVALID_POSTAL_CODE"
	ErrorCodePostalCodeNotFound        = "POSTAL_CODE_NOT_FOUND"
	ErrorCodeInvalidCountryCode        = "INVALID_COUNTRY_CODE"
	ErrorCodeLocationNotServiceable    = "LOCATION_NOT_SERVICEABLE"
	ErrorCodeInvalidRequest            = "INVALID_REQUEST"
	ErrorCodeInternalServerError       = "INTERNAL_SERVER_ERROR"
	ErrorCodeServiceUnavailable        = "SERVICE_UNAVAILABLE"
	ErrorCodePartnerServiceError       = "PARTNER_SERVICE_ERROR"
	ErrorCodeSpecificationServiceError = "SPECIFICATION_SERVICE_ERROR"
	ErrorCodeDatabaseError             = "DATABASE_ERROR"
	ErrorCodeValidationError           = "VALIDATION_ERROR"
	ErrorCodeTimeoutError              = "TIMEOUT_ERROR"
	// Authentication Error Codes
	ErrorCodeAuthFailed              = "AUTH_FAILED"
	ErrorCodeInvalidToken            = "INVALID_TOKEN"
	ErrorCodeExpiredToken            = "EXPIRED_TOKEN"
	ErrorCodeMissingCredentials      = "MISSING_CREDENTIALS"
	ErrorCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	ErrorCodeInvalidPartner          = "INVALID_PARTNER"
)

// Error Messages
const (
	MsgInvalidPostalCode         = "Invalid postal code format"
	MsgPostalCodeNotFound        = "Postal code not found"
	MsgInvalidCountryCode        = "Invalid country code"
	MsgLocationNotServiceable    = "Location is not serviceable"
	MsgInvalidRequest            = "Invalid request format"
	MsgInternalServerError       = "Internal server error"
	MsgServiceUnavailable        = "Service temporarily unavailable"
	MsgPartnerServiceError       = "Partner service error"
	MsgSpecificationServiceError = "Specification service error"
	MsgDatabaseError             = "Database error"
	MsgValidationError           = "Validation error"
	MsgTimeoutError              = "Request timeout"
	// Authentication Error Messages
	MsgAuthFailed              = "Authentication failed"
	MsgInvalidToken            = "Invalid authentication token"
	MsgExpiredToken            = "Authentication token has expired"
	MsgMissingCredentials      = "Missing authentication credentials"
	MsgInsufficientPermissions = "Insufficient permissions"
	MsgInvalidPartner          = "Invalid partner"
)

// API Versioning
const (
	APIVersion1 = "v1"
)

// API Endpoints
const (
	EndpointServiceabilityCheck  = "/serviceability/check"
	EndpointServiceabilityRoute  = "/serviceability/route"
	EndpointServiceabilityBulk   = "/serviceability/bulk"
	EndpointServiceabilityLegacy = "/serviceability"
	EndpointHealthCheck          = "/health"
	EndpointReadinessCheck       = "/ready"
)

// HTTP Headers
const (
	HeaderContentType   = "Content-Type"
	HeaderAccept        = "Accept"
	HeaderAuthorization = "Authorization"
	HeaderUserAgent     = "User-Agent"
	HeaderRequestID     = "X-Request-ID"
	HeaderCorrelationID = "X-Correlation-ID"
)

// Content Types
const (
	ContentTypeJSON = "application/json"
	ContentTypeXML  = "application/xml"
)

// Cache Keys
const (
	CacheKeyLocationHierarchy    = "location:hierarchy:%s:%s" // postal_code:country_code
	CacheKeyPartnerCapabilities  = "partner:capabilities:%d"  // partner_id
	CacheKeyServiceDefinitions   = "service:definitions:%s"   // service_type
	CacheKeyServiceabilityResult = "serviceability:%s:%s:%s"  // postal_code:country_code:service_type
)

// Cache TTL (in seconds)
const (
	CacheTTLLocationHierarchy    = 3600 // 1 hour
	CacheTTLPartnerCapabilities  = 1800 // 30 minutes
	CacheTTLServiceDefinitions   = 7200 // 2 hours
	CacheTTLServiceabilityResult = 900  // 15 minutes
)

// External Service URLs
const (
	PartnerServiceBasePath       = "/api/v1"
	SpecificationServiceBasePath = "/api/v1"
)

// Database Table Names
const (
	TableCountries         = "countries"
	TableRegions           = "regions"
	TableCities            = "cities"
	TableAreas             = "areas"
	TablePostalCodes       = "postal_codes"
	TablePostalCodeAliases = "postal_code_aliases"
)

// Validation Constants
const (
	MaxBulkRequestSize  = 100
	MinPostalCodeLength = 3
	MaxPostalCodeLength = 20
	// Country Code Validation - ISO 3166 A-2 standard
	CountryCodeLength    = 2 // Exactly 2 characters for ISO 3166 A-2
	MinCountryCodeLength = 2 // Minimum length for country codes
	MaxCountryCodeLength = 2 // Maximum length for country codes (ISO 3166 A-2)
)

// Timeout Constants (in seconds)
const (
	DefaultRequestTimeout       = 30
	PartnerServiceTimeout       = 10
	SpecificationServiceTimeout = 10
	DatabaseQueryTimeout        = 5
)

// Logging Fields
const (
	LogFieldRequestID   = "request_id"
	LogFieldUserID      = "user_id"
	LogFieldPartnerID   = "partner_id"
	LogFieldPostalCode  = "postal_code"
	LogFieldCountryCode = "country_code"
	LogFieldServiceType = "service_type"
	LogFieldAction      = "action"
	LogFieldDuration    = "duration_ms"
	LogFieldError       = "error"
)

// Metrics Constants
const (
	MetricRequestCount              = "serviceability_requests_total"
	MetricRequestDuration           = "serviceability_request_duration_seconds"
	MetricCacheHits                 = "serviceability_cache_hits_total"
	MetricCacheMisses               = "serviceability_cache_misses_total"
	MetricPartnerServiceCalls       = "partner_service_calls_total"
	MetricSpecificationServiceCalls = "specification_service_calls_total"
	MetricDatabaseQueries           = "database_queries_total"
	MetricErrorCount                = "serviceability_errors_total"
)

// Environment Variables
const (
	EnvServerPort              = "SERVER_PORT"
	EnvLogLevel                = "LOG_LEVEL"
	EnvDatabaseURL             = "DATABASE_URL"
	EnvRedisURL                = "REDIS_URL"
	EnvPartnerServiceURL       = "PARTNER_SERVICE_URL"
	EnvSpecificationServiceURL = "SPECIFICATION_SERVICE_URL"
	EnvPartnerServiceBaseURL   = "PARTNER_SERVICE_BASE_URL"
	EnvAPIKey                  = "API_KEY"
	EnvEnvironment             = "ENVIRONMENT"
)

// Default Values
const (
	DefaultServerPort     = "8080"
	DefaultLogLevel       = "info"
	DefaultEnvironment    = "development"
	DefaultCacheEnabled   = true
	DefaultMetricsEnabled = true
	DefaultTracingEnabled = false
)

// Location Scope Constants for partner coverage
const (
	LocationScopePostalCode = "POSTAL_CODE"
	LocationScopeArea       = "AREA"
	LocationScopeCity       = "CITY"
	LocationScopeRegion     = "REGION"
	LocationScopeCountry    = "COUNTRY"
)

// Zone Type Constants for partner coverage
const (
	ZoneTypePrimary   = "PRIMARY"
	ZoneTypeSecondary = "SECONDARY"
	ZoneTypeBuffer    = "BUFFER"
)

// Valid Location Scopes
var ValidLocationScopes = []string{
	LocationScopePostalCode,
	LocationScopeArea,
	LocationScopeCity,
	LocationScopeRegion,
	LocationScopeCountry,
}

// Valid Zone Types
var ValidZoneTypes = []string{
	ZoneTypePrimary,
	ZoneTypeSecondary,
	ZoneTypeBuffer,
}
