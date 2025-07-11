package services

import (
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// ServiceDefinitionValidator handles validation of service definition data
type ServiceDefinitionValidator struct {
	config config.ServiceDefinitionResolverConfig

	// Validation rules
	allowedParcelCategories map[string]bool
	allowedServiceTypes     map[string]bool
	allowedOperationTypes   map[string]bool
	allowedPaymentModes     map[string]bool
	allowedDeliveryModes    map[string]bool
}

// NewServiceDefinitionValidator creates a new validator instance
func NewServiceDefinitionValidator(config config.ServiceDefinitionResolverConfig) *ServiceDefinitionValidator {
	validator := &ServiceDefinitionValidator{
		config:                  config,
		allowedParcelCategories: make(map[string]bool),
		allowedServiceTypes:     make(map[string]bool),
		allowedOperationTypes:   make(map[string]bool),
		allowedPaymentModes:     make(map[string]bool),
		allowedDeliveryModes:    make(map[string]bool),
	}

	// Initialize default allowed values
	validator.initializeDefaultValidationRules()

	return validator
}

// initializeDefaultValidationRules sets up default validation rules
func (v *ServiceDefinitionValidator) initializeDefaultValidationRules() {
	// Default parcel categories
	defaultParcelCategories := []string{
		"ecom", "courier", "cargo", "documents", "packages", "freight", "pharmaceuticals", "food", "automotive",
	}
	for _, category := range defaultParcelCategories {
		v.allowedParcelCategories[strings.ToLower(category)] = true
	}

	// Default service types
	defaultServiceTypes := []string{
		"SDD", "NDD", "Standard", "Express", "Premium", "Economy", "Rush", "Scheduled", "COD", "Prepaid",
	}
	for _, serviceType := range defaultServiceTypes {
		v.allowedServiceTypes[strings.ToUpper(serviceType)] = true
	}

	// Default operation types
	defaultOperationTypes := []string{
		"pickup", "delivery", "tracking", "insurance", "packaging", "labeling", "sorting", "customs", "warehousing",
	}
	for _, operationType := range defaultOperationTypes {
		v.allowedOperationTypes[strings.ToLower(operationType)] = true
	}

	// Default payment modes
	defaultPaymentModes := []string{
		"COD", "ONLINE", "PREPAID", "CREDIT", "DEBIT", "UPI", "WALLET", "NET_BANKING", "CASH",
	}
	for _, paymentMode := range defaultPaymentModes {
		v.allowedPaymentModes[strings.ToUpper(paymentMode)] = true
	}

	// Default delivery modes
	defaultDeliveryModes := []string{
		"AIR", "SURFACE", "RAIL", "SEA", "ROAD", "MULTIMODAL", "EXPRESS", "STANDARD",
	}
	for _, deliveryMode := range defaultDeliveryModes {
		v.allowedDeliveryModes[strings.ToUpper(deliveryMode)] = true
	}
}

// ValidateParcelCategories validates a list of parcel categories
func (v *ServiceDefinitionValidator) ValidateParcelCategories(categories []string) error {
	if len(categories) == 0 {
		return fmt.Errorf("parcel categories list is empty")
	}

	for i, category := range categories {
		if err := v.ValidateParcelCategory(category); err != nil {
			return fmt.Errorf("invalid parcel category at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidateParcelCategory validates a single parcel category
func (v *ServiceDefinitionValidator) ValidateParcelCategory(category string) error {
	if strings.TrimSpace(category) == "" {
		return fmt.Errorf("parcel category cannot be empty")
	}

	normalizedCategory := strings.ToLower(strings.TrimSpace(category))

	if v.config.StrictValidation {
		if !v.allowedParcelCategories[normalizedCategory] {
			return fmt.Errorf("parcel category '%s' is not in allowed list", category)
		}
	}

	// Basic format validation
	if len(normalizedCategory) > 50 {
		return fmt.Errorf("parcel category '%s' is too long (max 50 characters)", category)
	}

	// Check for invalid characters
	if strings.ContainsAny(normalizedCategory, "!@#$%^&*()+={}[]|\\:;\"'<>?,/~`") {
		return fmt.Errorf("parcel category '%s' contains invalid characters", category)
	}

	return nil
}

// ValidateServiceTypes validates a list of service types
func (v *ServiceDefinitionValidator) ValidateServiceTypes(serviceTypes []string) error {
	if len(serviceTypes) == 0 {
		return fmt.Errorf("service types list is empty")
	}

	for i, serviceType := range serviceTypes {
		if err := v.ValidateServiceType(serviceType); err != nil {
			return fmt.Errorf("invalid service type at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidateServiceType validates a single service type
func (v *ServiceDefinitionValidator) ValidateServiceType(serviceType string) error {
	if strings.TrimSpace(serviceType) == "" {
		return fmt.Errorf("service type cannot be empty")
	}

	normalizedServiceType := strings.ToUpper(strings.TrimSpace(serviceType))

	if v.config.StrictValidation {
		if !v.allowedServiceTypes[normalizedServiceType] {
			return fmt.Errorf("service type '%s' is not in allowed list", serviceType)
		}
	}

	// Basic format validation
	if len(normalizedServiceType) > 20 {
		return fmt.Errorf("service type '%s' is too long (max 20 characters)", serviceType)
	}

	return nil
}

// ValidateOperationTypes validates a list of operation types
func (v *ServiceDefinitionValidator) ValidateOperationTypes(operationTypes []string) error {
	if len(operationTypes) == 0 {
		return fmt.Errorf("operation types list is empty")
	}

	for i, operationType := range operationTypes {
		if err := v.ValidateOperationType(operationType); err != nil {
			return fmt.Errorf("invalid operation type at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidateOperationType validates a single operation type
func (v *ServiceDefinitionValidator) ValidateOperationType(operationType string) error {
	if strings.TrimSpace(operationType) == "" {
		return fmt.Errorf("operation type cannot be empty")
	}

	normalizedOperationType := strings.ToLower(strings.TrimSpace(operationType))

	if v.config.StrictValidation {
		if !v.allowedOperationTypes[normalizedOperationType] {
			return fmt.Errorf("operation type '%s' is not in allowed list", operationType)
		}
	}

	// Basic format validation
	if len(normalizedOperationType) > 30 {
		return fmt.Errorf("operation type '%s' is too long (max 30 characters)", operationType)
	}

	return nil
}

// ValidatePaymentModes validates a list of payment modes
func (v *ServiceDefinitionValidator) ValidatePaymentModes(paymentModes []string) error {
	if len(paymentModes) == 0 {
		return fmt.Errorf("payment modes list is empty")
	}

	for i, paymentMode := range paymentModes {
		if err := v.ValidatePaymentMode(paymentMode); err != nil {
			return fmt.Errorf("invalid payment mode at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidatePaymentMode validates a single payment mode
func (v *ServiceDefinitionValidator) ValidatePaymentMode(paymentMode string) error {
	if strings.TrimSpace(paymentMode) == "" {
		return fmt.Errorf("payment mode cannot be empty")
	}

	normalizedPaymentMode := strings.ToUpper(strings.TrimSpace(paymentMode))

	if v.config.StrictValidation {
		if !v.allowedPaymentModes[normalizedPaymentMode] {
			return fmt.Errorf("payment mode '%s' is not in allowed list", paymentMode)
		}
	}

	// Basic format validation
	if len(normalizedPaymentMode) > 20 {
		return fmt.Errorf("payment mode '%s' is too long (max 20 characters)", paymentMode)
	}

	return nil
}

// ValidateDeliveryModes validates a list of delivery modes
func (v *ServiceDefinitionValidator) ValidateDeliveryModes(deliveryModes []string) error {
	if len(deliveryModes) == 0 {
		return fmt.Errorf("delivery modes list is empty")
	}

	for i, deliveryMode := range deliveryModes {
		if err := v.ValidateDeliveryMode(deliveryMode); err != nil {
			return fmt.Errorf("invalid delivery mode at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidateDeliveryMode validates a single delivery mode
func (v *ServiceDefinitionValidator) ValidateDeliveryMode(deliveryMode string) error {
	if strings.TrimSpace(deliveryMode) == "" {
		return fmt.Errorf("delivery mode cannot be empty")
	}

	normalizedDeliveryMode := strings.ToUpper(strings.TrimSpace(deliveryMode))

	if v.config.StrictValidation {
		if !v.allowedDeliveryModes[normalizedDeliveryMode] {
			return fmt.Errorf("delivery mode '%s' is not in allowed list", deliveryMode)
		}
	}

	// Basic format validation
	if len(normalizedDeliveryMode) > 20 {
		return fmt.Errorf("delivery mode '%s' is too long (max 20 characters)", deliveryMode)
	}

	return nil
}

// ValidateServiceDefinition validates a complete service definition
func (v *ServiceDefinitionValidator) ValidateServiceDefinition(serviceDef *interfaces.ServiceDefinition) error {
	if serviceDef == nil {
		return fmt.Errorf("service definition cannot be nil")
	}

	// Validate service type
	if err := v.ValidateServiceType(serviceDef.ServiceType); err != nil {
		return fmt.Errorf("invalid service type: %w", err)
	}

	// Validate parcel category
	if err := v.ValidateParcelCategory(serviceDef.ParcelCategory); err != nil {
		return fmt.Errorf("invalid parcel category: %w", err)
	}

	// Validate default operations
	if err := v.ValidateOperationTypes(serviceDef.DefaultOperations); err != nil {
		return fmt.Errorf("invalid default operations: %w", err)
	}

	// Validate default payments
	if err := v.ValidatePaymentModes(serviceDef.DefaultPayments); err != nil {
		return fmt.Errorf("invalid default payments: %w", err)
	}

	// Validate default delivery modes
	if err := v.ValidateDeliveryModes(serviceDef.DefaultDelivery); err != nil {
		return fmt.Errorf("invalid default delivery modes: %w", err)
	}

	// Validate description
	if strings.TrimSpace(serviceDef.Description) == "" {
		return fmt.Errorf("service definition description cannot be empty")
	}

	if len(serviceDef.Description) > 500 {
		return fmt.Errorf("service definition description is too long (max 500 characters)")
	}

	return nil
}

// UpdateAllowedValues allows updating the validation rules at runtime
func (v *ServiceDefinitionValidator) UpdateAllowedValues(
	parcelCategories []string,
	serviceTypes []string,
	operationTypes []string,
	paymentModes []string,
	deliveryModes []string,
) {
	// Update parcel categories
	if parcelCategories != nil {
		v.allowedParcelCategories = make(map[string]bool)
		for _, category := range parcelCategories {
			v.allowedParcelCategories[strings.ToLower(category)] = true
		}
	}

	// Update service types
	if serviceTypes != nil {
		v.allowedServiceTypes = make(map[string]bool)
		for _, serviceType := range serviceTypes {
			v.allowedServiceTypes[strings.ToUpper(serviceType)] = true
		}
	}

	// Update operation types
	if operationTypes != nil {
		v.allowedOperationTypes = make(map[string]bool)
		for _, operationType := range operationTypes {
			v.allowedOperationTypes[strings.ToLower(operationType)] = true
		}
	}

	// Update payment modes
	if paymentModes != nil {
		v.allowedPaymentModes = make(map[string]bool)
		for _, paymentMode := range paymentModes {
			v.allowedPaymentModes[strings.ToUpper(paymentMode)] = true
		}
	}

	// Update delivery modes
	if deliveryModes != nil {
		v.allowedDeliveryModes = make(map[string]bool)
		for _, deliveryMode := range deliveryModes {
			v.allowedDeliveryModes[strings.ToUpper(deliveryMode)] = true
		}
	}
}

// GetAllowedValues returns the currently allowed values for validation
func (v *ServiceDefinitionValidator) GetAllowedValues() map[string][]string {
	result := make(map[string][]string)

	// Extract parcel categories
	var parcelCategories []string
	for category := range v.allowedParcelCategories {
		parcelCategories = append(parcelCategories, category)
	}
	result["parcel_categories"] = parcelCategories

	// Extract service types
	var serviceTypes []string
	for serviceType := range v.allowedServiceTypes {
		serviceTypes = append(serviceTypes, serviceType)
	}
	result["service_types"] = serviceTypes

	// Extract operation types
	var operationTypes []string
	for operationType := range v.allowedOperationTypes {
		operationTypes = append(operationTypes, operationType)
	}
	result["operation_types"] = operationTypes

	// Extract payment modes
	var paymentModes []string
	for paymentMode := range v.allowedPaymentModes {
		paymentModes = append(paymentModes, paymentMode)
	}
	result["payment_modes"] = paymentModes

	// Extract delivery modes
	var deliveryModes []string
	for deliveryMode := range v.allowedDeliveryModes {
		deliveryModes = append(deliveryModes, deliveryMode)
	}
	result["delivery_modes"] = deliveryModes

	return result
}

// IsStrictValidationEnabled returns whether strict validation is enabled
func (v *ServiceDefinitionValidator) IsStrictValidationEnabled() bool {
	return v.config.StrictValidation
}

// SetStrictValidation enables or disables strict validation
func (v *ServiceDefinitionValidator) SetStrictValidation(enabled bool) {
	v.config.StrictValidation = enabled
}
