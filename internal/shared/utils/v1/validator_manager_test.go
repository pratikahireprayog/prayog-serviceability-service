package utils

import (
	"context"
	"testing"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

func TestNewPartnerDataValidator(t *testing.T) {
	tests := []struct {
		name   string
		config *ValidationConfig
	}{
		{
			name:   "nil config should use defaults",
			config: nil,
		},
		{
			name: "custom config should be used",
			config: &ValidationConfig{
				AllowedServiceTypes: []string{"Custom"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewPartnerDataValidator(tt.config)
			if validator == nil {
				t.Error("NewPartnerDataValidator() returned nil")
			}
		})
	}
}

func TestValidatePartnerCapability(t *testing.T) {
	validator := NewPartnerDataValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name       string
		capability *interfaces.PartnerCapability
		wantErr    bool
	}{
		{
			name:       "nil capability should return error",
			capability: nil,
			wantErr:    true,
		},
		{
			name: "zero partner ID should return error",
			capability: &interfaces.PartnerCapability{
				PartnerID: 0,
			},
			wantErr: true,
		},
		{
			name: "empty partner name should return error",
			capability: &interfaces.PartnerCapability{
				PartnerID:   1,
				PartnerName: "",
			},
			wantErr: true,
		},
		{
			name: "invalid partner name characters should return error",
			capability: &interfaces.PartnerCapability{
				PartnerID:   1,
				PartnerName: "Partner@#$%",
			},
			wantErr: true,
		},
		{
			name: "empty service types should return error",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "Valid Partner",
				ServiceTypes:     []string{},
				ParcelCategories: []string{"ecom"},
			},
			wantErr: true,
		},
		{
			name: "invalid service type should return error",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "Valid Partner",
				ServiceTypes:     []string{"INVALID"},
				ParcelCategories: []string{"ecom"},
			},
			wantErr: true,
		},
		{
			name: "empty parcel categories should return error",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "Valid Partner",
				ServiceTypes:     []string{"SDD"},
				ParcelCategories: []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid parcel category should return error",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "Valid Partner",
				ServiceTypes:     []string{"SDD"},
				ParcelCategories: []string{"invalid"},
			},
			wantErr: true,
		},
		{
			name: "invalid operation type should return error",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "Valid Partner",
				ServiceTypes:     []string{"SDD"},
				ParcelCategories: []string{"ecom"},
				OperationTypes:   []string{"invalid"},
			},
			wantErr: true,
		},
		{
			name: "invalid payment mode should return error",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "Valid Partner",
				ServiceTypes:     []string{"SDD"},
				ParcelCategories: []string{"ecom"},
				OperationTypes:   []string{"pickup"},
				PaymentModes:     []string{"INVALID"},
			},
			wantErr: true,
		},
		{
			name: "invalid delivery mode should return error",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "Valid Partner",
				ServiceTypes:     []string{"SDD"},
				ParcelCategories: []string{"ecom"},
				OperationTypes:   []string{"pickup"},
				PaymentModes:     []string{"ONLINE"},
				DeliveryModes:    []string{"INVALID"},
			},
			wantErr: true,
		},
		{
			name: "valid capability should pass",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "Valid Partner",
				IsActive:         true,
				ServiceTypes:     []string{"SDD", "NDD"},
				ParcelCategories: []string{"ecom"},
				OperationTypes:   []string{"pickup", "delivery"},
				PaymentModes:     []string{"ONLINE", "COD"},
				DeliveryModes:    []string{"AIR", "SURFACE"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePartnerCapability(ctx, tt.capability)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidatePartnerCapability() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePartnerCapability() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidatePartnerServiceCapabilities(t *testing.T) {
	validator := NewPartnerDataValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name         string
		capabilities *interfaces.PartnerServiceCapabilities
		wantErr      bool
	}{
		{
			name:         "nil capabilities should return error",
			capabilities: nil,
			wantErr:      true,
		},
		{
			name: "zero partner ID should return error",
			capabilities: &interfaces.PartnerServiceCapabilities{
				PartnerID: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid service capability should return error",
			capabilities: &interfaces.PartnerServiceCapabilities{
				PartnerID: 1,
				Capabilities: []interfaces.ServiceCapability{
					{
						ServiceType:    "INVALID",
						ParcelCategory: "ecom",
						Rating:         5.0,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid preference should return error",
			capabilities: &interfaces.PartnerServiceCapabilities{
				PartnerID: 1,
				Capabilities: []interfaces.ServiceCapability{
					{
						ServiceType:    "SDD",
						ParcelCategory: "ecom",
						Rating:         5.0,
					},
				},
				Preferences: []interfaces.PartnerPreference{
					{
						ServiceType:     "INVALID",
						ParcelCategory:  "ecom",
						Priority:        1,
						EffectiveRating: 8.0,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid capabilities should pass",
			capabilities: &interfaces.PartnerServiceCapabilities{
				PartnerID: 1,
				Capabilities: []interfaces.ServiceCapability{
					{
						ServiceType:    "SDD",
						ParcelCategory: "ecom",
						OperationTypes: []string{"pickup", "delivery"},
						PaymentModes:   []string{"ONLINE", "COD"},
						DeliveryModes:  []string{"AIR", "SURFACE"},
						IsActive:       true,
						Rating:         8.5,
					},
				},
				Preferences: []interfaces.PartnerPreference{
					{
						ServiceType:     "SDD",
						ParcelCategory:  "ecom",
						IsPreferred:     true,
						Priority:        1,
						EffectiveRating: 9.0,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePartnerServiceCapabilities(ctx, tt.capabilities)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidatePartnerServiceCapabilities() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePartnerServiceCapabilities() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateServiceFilters(t *testing.T) {
	validator := NewPartnerDataValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		filters *interfaces.ServiceFilters
		wantErr bool
	}{
		{
			name:    "nil filters should pass (optional)",
			filters: nil,
			wantErr: false,
		},
		{
			name: "invalid service type should return error",
			filters: &interfaces.ServiceFilters{
				ServiceTypes: []string{"INVALID"},
			},
			wantErr: true,
		},
		{
			name: "invalid parcel category should return error",
			filters: &interfaces.ServiceFilters{
				ParcelCategories: []string{"invalid"},
			},
			wantErr: true,
		},
		{
			name: "invalid operation type should return error",
			filters: &interfaces.ServiceFilters{
				OperationTypes: []string{"invalid"},
			},
			wantErr: true,
		},
		{
			name: "valid filters should pass",
			filters: &interfaces.ServiceFilters{
				ServiceTypes:     []string{"SDD", "NDD"},
				ParcelCategories: []string{"ecom", "courier"},
				OperationTypes:   []string{"pickup", "delivery"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateServiceFilters(ctx, tt.filters)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateServiceFilters() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateServiceFilters() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidatePartnerInfo(t *testing.T) {
	validator := NewPartnerDataValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name        string
		partnerInfo *interfaces.PartnerInfo
		wantErr     bool
	}{
		{
			name:        "nil partner info should return error",
			partnerInfo: nil,
			wantErr:     true,
		},
		{
			name: "zero ID should return error",
			partnerInfo: &interfaces.PartnerInfo{
				ID:   0,
				Name: "Test Partner",
				Type: "ECOM",
			},
			wantErr: true,
		},
		{
			name: "empty name should return error",
			partnerInfo: &interfaces.PartnerInfo{
				ID:   1,
				Name: "",
				Type: "ECOM",
			},
			wantErr: true,
		},
		{
			name: "empty type should return error",
			partnerInfo: &interfaces.PartnerInfo{
				ID:   1,
				Name: "Test Partner",
				Type: "",
			},
			wantErr: true,
		},
		{
			name: "valid partner info should pass",
			partnerInfo: &interfaces.PartnerInfo{
				ID:       1,
				Name:     "Test Partner",
				IsActive: true,
				Type:     "ECOM",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePartnerInfo(ctx, tt.partnerInfo)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidatePartnerInfo() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePartnerInfo() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateLocationHierarchy(t *testing.T) {
	validator := NewPartnerDataValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		hierarchy *models.LocationHierarchy
		wantErr   bool
	}{
		{
			name:      "nil hierarchy should return error",
			hierarchy: nil,
			wantErr:   true,
		},
		{
			name: "invalid postal code should return error",
			hierarchy: &models.LocationHierarchy{
				PostalCode:  "invalid@postal",
				CountryCode: "IN",
			},
			wantErr: true,
		},
		{
			name: "invalid country code should return error",
			hierarchy: &models.LocationHierarchy{
				PostalCode:  "110001",
				CountryCode: "INVALID",
			},
			wantErr: true,
		},
		{
			name: "valid hierarchy should pass",
			hierarchy: &models.LocationHierarchy{
				PostalCode:  "110001",
				CountryCode: "IN",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateLocationHierarchy(ctx, tt.hierarchy)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateLocationHierarchy() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateLocationHierarchy() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateServiceabilityRequest(t *testing.T) {
	validator := NewPartnerDataValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *models.ServiceabilityCheckRequest
		wantErr bool
	}{
		{
			name:    "nil request should return error",
			req:     nil,
			wantErr: true,
		},
		{
			name: "mixed query types should return error",
			req: &models.ServiceabilityCheckRequest{
				PostalCode:         stringPtr("110001"),
				PickupPostalCode:   stringPtr("110002"),
				DeliveryPostalCode: stringPtr("110003"),
				CountryCode:        "IN",
			},
			wantErr: true,
		},
		{
			name: "incomplete origin-destination should return error",
			req: &models.ServiceabilityCheckRequest{
				PickupPostalCode: stringPtr("110001"),
				CountryCode:      "IN",
			},
			wantErr: true,
		},
		{
			name: "valid generic location request should pass",
			req: &models.ServiceabilityCheckRequest{
				PostalCode:  stringPtr("110001"),
				CountryCode: "IN",
			},
			wantErr: false,
		},
		{
			name: "valid origin-destination request should pass",
			req: &models.ServiceabilityCheckRequest{
				PickupPostalCode:   stringPtr("110001"),
				DeliveryPostalCode: stringPtr("110002"),
				CountryCode:        "IN",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateServiceabilityRequest(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateServiceabilityRequest() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateServiceabilityRequest() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateBulkServiceabilityRequest(t *testing.T) {
	validator := NewPartnerDataValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *models.BulkServiceabilityRequest
		wantErr bool
	}{
		{
			name:    "nil request should return error",
			req:     nil,
			wantErr: true,
		},
		{
			name: "invalid individual request should return error",
			req: &models.BulkServiceabilityRequest{
				Requests: []models.ServiceabilityCheckRequest{
					{
						CountryCode: "IN", // Missing postal codes
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid bulk request should pass",
			req: &models.BulkServiceabilityRequest{
				Requests: []models.ServiceabilityCheckRequest{
					{
						PostalCode:  stringPtr("110001"),
						CountryCode: "IN",
					},
					{
						PickupPostalCode:   stringPtr("110002"),
						DeliveryPostalCode: stringPtr("110003"),
						CountryCode:        "IN",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBulkServiceabilityRequest(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateBulkServiceabilityRequest() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateBulkServiceabilityRequest() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidatePartnerRating(t *testing.T) {
	validator := NewPartnerDataValidator(nil)

	tests := []struct {
		name    string
		rating  float64
		wantErr bool
	}{
		{"negative rating should error", -1.0, true},
		{"rating above 10 should error", 11.0, true},
		{"zero rating should pass", 0.0, false},
		{"valid rating should pass", 5.5, false},
		{"max rating should pass", 10.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePartnerRating(tt.rating)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidatePartnerRating() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePartnerRating() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateServiceType(t *testing.T) {
	validator := NewPartnerDataValidator(nil)

	tests := []struct {
		name        string
		serviceType string
		wantErr     bool
	}{
		{"empty service type should error", "", true},
		{"whitespace only should error", "   ", true},
		{"invalid service type should error", "INVALID", true},
		{"valid SDD should pass", "SDD", false},
		{"valid NDD should pass", "NDD", false},
		{"valid Standard should pass", "Standard", false},
		{"valid Express should pass", "Express", false},
		{"valid vayuquick should pass", "vayuquick", false},
		{"valid vayuquickpro should pass", "vayuquickpro", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateServiceType(tt.serviceType)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateServiceType() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateServiceType() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateParcelCategory(t *testing.T) {
	validator := NewPartnerDataValidator(nil)

	tests := []struct {
		name     string
		category string
		wantErr  bool
	}{
		{"empty category should error", "", true},
		{"whitespace only should error", "   ", true},
		{"invalid category should error", "invalid", true},
		{"valid ecom should pass", "ecom", false},
		{"valid courier should pass", "courier", false},
		{"valid cargo should pass", "cargo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateParcelCategory(tt.category)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateParcelCategory() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateParcelCategory() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateOperationType(t *testing.T) {
	validator := NewPartnerDataValidator(nil)

	tests := []struct {
		name          string
		operationType string
		wantErr       bool
	}{
		{"empty operation type should error", "", true},
		{"whitespace only should error", "   ", true},
		{"invalid operation type should error", "invalid", true},
		{"valid pickup should pass", "pickup", false},
		{"valid delivery should pass", "delivery", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOperationType(tt.operationType)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateOperationType() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateOperationType() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidatePaymentMode(t *testing.T) {
	validator := NewPartnerDataValidator(nil)

	tests := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{"empty payment mode should error", "", true},
		{"whitespace only should error", "   ", true},
		{"invalid payment mode should error", "INVALID", true},
		{"valid ONLINE should pass", "ONLINE", false},
		{"valid COD should pass", "COD", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePaymentMode(tt.mode)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidatePaymentMode() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePaymentMode() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateDeliveryMode(t *testing.T) {
	validator := NewPartnerDataValidator(nil)

	tests := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{"empty delivery mode should error", "", true},
		{"whitespace only should error", "   ", true},
		{"invalid delivery mode should error", "INVALID", true},
		{"valid AIR should pass", "AIR", false},
		{"valid SURFACE should pass", "SURFACE", false},
		{"valid RAIL should pass", "RAIL", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDeliveryMode(tt.mode)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateDeliveryMode() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateDeliveryMode() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	errors := ValidationErrors{
		Errors: []ValidationError{
			{
				Field:   "field1",
				Message: "error message 1",
				Value:   "value1",
			},
			{
				Field:   "field2",
				Message: "error message 2",
			},
		},
	}

	result := errors.Error()
	expected := "validation failed: field1: error message 1 (value: value1), field2: error message 2"

	if result != expected {
		t.Errorf("ValidationErrors.Error() = %s, want %s", result, expected)
	}
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
