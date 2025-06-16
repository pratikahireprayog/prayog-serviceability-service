package dtos

import "time"

// SpecDefinitionsResponse represents the response for fetching specification definitions
type SpecDefinitionsResponse struct {
	SpecDefinitions []SpecDefinitionDTO `json:"spec_definitions"`
	Metadata        ResponseMetadata    `json:"metadata"`
}

// SpecDefinitionDTO represents a specification definition from Specification Service
type SpecDefinitionDTO struct {
	ID          uint      `json:"id"`
	SpecType    string    `json:"spec_type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CatalogsResponse represents the response for fetching catalogs
type CatalogsResponse struct {
	Catalogs []CatalogDTO     `json:"catalogs"`
	Metadata ResponseMetadata `json:"metadata"`
}

// CatalogDTO represents a catalog from Specification Service
type CatalogDTO struct {
	ID           uint             `json:"id"`
	SpecDefID    uint             `json:"spec_def_id"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Version      string           `json:"version"`
	IsActive     bool             `json:"is_active"`
	CatalogItems []CatalogItemDTO `json:"catalog_items"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// CatalogItemDTO represents an item in a catalog
type CatalogItemDTO struct {
	ID          uint              `json:"id"`
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	Priority    int               `json:"priority"`
	IsActive    bool              `json:"is_active"`
	ParentID    *uint             `json:"parent_id,omitempty"`
}

// EntitySpecificationsRequest represents the request for entity specifications
type EntitySpecificationsRequest struct {
	EntityType   string   `json:"entity_type" validate:"required"`
	EntityID     string   `json:"entity_id" validate:"required"`
	SpecTypes    []string `json:"spec_types,omitempty"`
	OnlyActive   bool     `json:"only_active"`
	IncludeItems bool     `json:"include_items"`
}

// EntitySpecificationsResponse represents the response for entity specifications
type EntitySpecificationsResponse struct {
	EntitySpecifications []EntitySpecificationDTO `json:"entity_specifications"`
	Metadata             ResponseMetadata         `json:"metadata"`
}

// EntitySpecificationDTO represents an entity specification from Specification Service
type EntitySpecificationDTO struct {
	ID             uint                        `json:"id"`
	EntityType     string                      `json:"entity_type"`
	EntityID       string                      `json:"entity_id"`
	SpecDefID      uint                        `json:"spec_def_id"`
	SpecType       string                      `json:"spec_type"`
	Specification  map[string]interface{}      `json:"specification"`
	EffectiveRules []EffectiveRuleDTO          `json:"effective_rules,omitempty"`
	Constraints    []ConstraintDTO             `json:"constraints,omitempty"`
	Metadata       EntitySpecificationMetadata `json:"metadata"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}

// EffectiveRuleDTO represents an effective rule for an entity
type EffectiveRuleDTO struct {
	ID          uint   `json:"id"`
	RuleType    string `json:"rule_type"`
	Condition   string `json:"condition"`
	Action      string `json:"action"`
	Priority    int    `json:"priority"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description,omitempty"`
}

// ConstraintDTO represents a constraint for an entity specification
type ConstraintDTO struct {
	ID             uint   `json:"id"`
	ConstraintType string `json:"constraint_type"`
	Field          string `json:"field"`
	Operator       string `json:"operator"`
	Value          string `json:"value"`
	ErrorMessage   string `json:"error_message,omitempty"`
	IsActive       bool   `json:"is_active"`
}

// EntitySpecificationMetadata provides metadata about entity specifications
type EntitySpecificationMetadata struct {
	LastUpdated      time.Time `json:"last_updated"`
	ValidatedAt      time.Time `json:"validated_at"`
	ValidationStatus string    `json:"validation_status"`
	Priority         int       `json:"priority"`
	Inheritance      string    `json:"inheritance,omitempty"`
}

// ResponseMetadata provides common metadata for API responses
type ResponseMetadata struct {
	RequestID     string    `json:"request_id"`
	Timestamp     time.Time `json:"timestamp"`
	TotalCount    int       `json:"total_count"`
	FilteredCount int       `json:"filtered_count,omitempty"`
	Version       string    `json:"version"`
	CacheHit      bool      `json:"cache_hit,omitempty"`
	TTL           int       `json:"ttl_seconds,omitempty"`
}

// ServiceDefinitionQueryRequest represents a query request for service definitions
type ServiceDefinitionQueryRequest struct {
	ServiceTypes     []string `json:"service_types,omitempty"`
	ParcelCategories []string `json:"parcel_categories,omitempty"`
	OnlyActive       bool     `json:"only_active"`
	IncludeDefaults  bool     `json:"include_defaults"`
	Version          string   `json:"version,omitempty"`
}

// ServiceDefinitionQueryResponse represents a query response for service definitions
type ServiceDefinitionQueryResponse struct {
	ServiceDefinitions []ServiceDefinitionDTO `json:"service_definitions"`
	Metadata           ResponseMetadata       `json:"metadata"`
}

// ServiceDefinitionDTO represents a service definition derived from specifications
type ServiceDefinitionDTO struct {
	ServiceType       string   `json:"service_type"`
	ParcelCategory    string   `json:"parcel_category"`
	DefaultOperations []string `json:"default_operations"`
	DefaultPayments   []string `json:"default_payments"`
	DefaultDelivery   []string `json:"default_delivery"`
	Description       string   `json:"description"`
	MaxWeight         float64  `json:"max_weight,omitempty"`
	MaxDimensions     string   `json:"max_dimensions,omitempty"`
	EstimatedDuration string   `json:"estimated_duration,omitempty"`
	Cost              float64  `json:"cost,omitempty"`
	IsActive          bool     `json:"is_active"`
	Priority          int      `json:"priority"`
	ValidationRules   []string `json:"validation_rules,omitempty"`
	BusinessRules     []string `json:"business_rules,omitempty"`
}

// HealthCheckResponse represents the health check response from Specification Service
type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Services  map[string]string `json:"services"`
}

// ErrorResponse represents error response from Specification Service
type SpecificationErrorResponse struct {
	Success   bool              `json:"success"`
	ErrorCode string            `json:"error_code"`
	Message   string            `json:"message"`
	Details   map[string]string `json:"details,omitempty"`
	RequestID string            `json:"request_id"`
	Timestamp time.Time         `json:"timestamp"`
}
