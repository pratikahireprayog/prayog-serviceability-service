package utils

import (
	"testing"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

func TestNewPartnerDataTransformer(t *testing.T) {
	tests := []struct {
		name   string
		config *TransformerConfig
		want   PartnerDataTransformer
	}{
		{
			name:   "nil config should use defaults",
			config: nil,
		},
		{
			name: "custom config should be used",
			config: &TransformerConfig{
				DefaultRating: 8.0,
				ServiceTypeMapping: map[string]string{
					"Custom": "custom",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewPartnerDataTransformer(tt.config)
			if transformer == nil {
				t.Error("NewPartnerDataTransformer() returned nil")
			}
		})
	}
}

func TestTransformPartnerInfo(t *testing.T) {
	transformer := NewPartnerDataTransformer(nil)

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
			name: "zero partner ID should return error",
			partnerInfo: &interfaces.PartnerInfo{
				ID:   0,
				Name: "Test Partner",
			},
			wantErr: true,
		},
		{
			name: "empty partner name should return error",
			partnerInfo: &interfaces.PartnerInfo{
				ID:   1,
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "valid ECOM partner should succeed",
			partnerInfo: &interfaces.PartnerInfo{
				ID:       1,
				Name:     "ECOM Partner",
				IsActive: true,
				Type:     "ECOM",
			},
			wantErr: false,
		},
		{
			name: "valid COURIER partner should succeed",
			partnerInfo: &interfaces.PartnerInfo{
				ID:       2,
				Name:     "Courier Partner",
				IsActive: true,
				Type:     "COURIER",
			},
			wantErr: false,
		},
		{
			name: "valid CARGO partner should succeed",
			partnerInfo: &interfaces.PartnerInfo{
				ID:       3,
				Name:     "Cargo Partner",
				IsActive: true,
				Type:     "CARGO",
			},
			wantErr: false,
		},
		{
			name: "unknown partner type should use defaults",
			partnerInfo: &interfaces.PartnerInfo{
				ID:       4,
				Name:     "Unknown Partner",
				IsActive: true,
				Type:     "UNKNOWN",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformPartnerInfo(tt.partnerInfo)

			if tt.wantErr {
				if err == nil {
					t.Error("TransformPartnerInfo() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("TransformPartnerInfo() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("TransformPartnerInfo() returned nil result")
				return
			}

			// Validate basic fields
			if result.PartnerID != tt.partnerInfo.ID {
				t.Errorf("PartnerID = %d, want %d", result.PartnerID, tt.partnerInfo.ID)
			}

			if result.PartnerName != tt.partnerInfo.Name {
				t.Errorf("PartnerName = %s, want %s", result.PartnerName, tt.partnerInfo.Name)
			}

			if result.IsActive != tt.partnerInfo.IsActive {
				t.Errorf("IsActive = %t, want %t", result.IsActive, tt.partnerInfo.IsActive)
			}

			// Validate type-specific mappings
			switch tt.partnerInfo.Type {
			case "ECOM":
				if len(result.ServiceTypes) != 2 || result.ServiceTypes[0] != "SDD" || result.ServiceTypes[1] != "NDD" {
					t.Errorf("ECOM partner should have SDD and NDD service types, got %v", result.ServiceTypes)
				}
				if len(result.ParcelCategories) != 1 || result.ParcelCategories[0] != "ecom" {
					t.Errorf("ECOM partner should have ecom parcel category, got %v", result.ParcelCategories)
				}
			case "COURIER":
				if len(result.ServiceTypes) != 2 || result.ServiceTypes[0] != "Standard" || result.ServiceTypes[1] != "Express" {
					t.Errorf("COURIER partner should have Standard and Express service types, got %v", result.ServiceTypes)
				}
			case "CARGO":
				if len(result.ServiceTypes) != 2 || result.ServiceTypes[0] != "vayuquick" || result.ServiceTypes[1] != "vayuquickpro" {
					t.Errorf("CARGO partner should have vayuquick and vayuquickpro service types, got %v", result.ServiceTypes)
				}
			}

			// Validate default fields
			if len(result.PaymentModes) != 2 || result.PaymentModes[0] != "ONLINE" || result.PaymentModes[1] != "COD" {
				t.Errorf("Expected default payment modes [ONLINE, COD], got %v", result.PaymentModes)
			}

			if len(result.DeliveryModes) != 2 || result.DeliveryModes[0] != "AIR" || result.DeliveryModes[1] != "SURFACE" {
				t.Errorf("Expected default delivery modes [AIR, SURFACE], got %v", result.DeliveryModes)
			}
		})
	}
}

func TestTransformPartnerEffectiveDetails(t *testing.T) {
	transformer := NewPartnerDataTransformer(nil)

	tests := []struct {
		name    string
		details *interfaces.PartnerEffectiveDetails
		wantErr bool
	}{
		{
			name:    "nil details should return error",
			details: nil,
			wantErr: true,
		},
		{
			name: "zero partner ID should return error",
			details: &interfaces.PartnerEffectiveDetails{
				PartnerID: 0,
			},
			wantErr: true,
		},
		{
			name: "valid details with preferences and ratings",
			details: &interfaces.PartnerEffectiveDetails{
				PartnerID: 1,
				Preferences: []interfaces.EffectivePreference{
					{
						EntityType:      "SERVICE",
						EntityID:        "SDD",
						PreferenceValue: "preferred",
						EffectiveRating: 9.0,
					},
				},
				Ratings: []interfaces.EffectiveRating{
					{
						EntityType: "SERVICE",
						EntityID:   "Express",
						Rating:     8.5,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty preferences and ratings should create defaults",
			details: &interfaces.PartnerEffectiveDetails{
				PartnerID:   2,
				Preferences: []interfaces.EffectivePreference{},
				Ratings:     []interfaces.EffectiveRating{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformPartnerEffectiveDetails(tt.details)

			if tt.wantErr {
				if err == nil {
					t.Error("TransformPartnerEffectiveDetails() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("TransformPartnerEffectiveDetails() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("TransformPartnerEffectiveDetails() returned nil result")
				return
			}

			if result.PartnerID != tt.details.PartnerID {
				t.Errorf("PartnerID = %d, want %d", result.PartnerID, tt.details.PartnerID)
			}

			// Check that we have either derived capabilities or defaults
			if len(result.Capabilities) == 0 {
				t.Error("Expected at least one capability (default if none derived)")
			}

			// If no ratings provided, should have default capability
			if len(tt.details.Ratings) == 0 {
				if len(result.Capabilities) != 1 {
					t.Errorf("Expected 1 default capability, got %d", len(result.Capabilities))
				}
				defaultCap := result.Capabilities[0]
				if defaultCap.ServiceType != "Standard" || defaultCap.ParcelCategory != "courier" {
					t.Errorf("Default capability should be Standard/courier, got %s/%s", defaultCap.ServiceType, defaultCap.ParcelCategory)
				}
			}
		})
	}
}

func TestMapServiceTypeToParcelCategory(t *testing.T) {
	transformer := NewPartnerDataTransformer(nil)

	tests := []struct {
		serviceType string
		want        string
	}{
		{"SDD", "ecom"},
		{"NDD", "ecom"},
		{"Standard", "courier"},
		{"Express", "courier"},
		{"vayuquick", "cargo"},
		{"vayuquickpro", "cargo"},
		{"unknown", "courier"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.serviceType, func(t *testing.T) {
			result := transformer.MapServiceTypeToParcelCategory(tt.serviceType)
			if result != tt.want {
				t.Errorf("MapServiceTypeToParcelCategory(%s) = %s, want %s", tt.serviceType, result, tt.want)
			}
		})
	}
}

func TestExtractOperationTypes(t *testing.T) {
	transformer := NewPartnerDataTransformer(nil)

	tests := []struct {
		name      string
		specValue string
		want      []string
	}{
		{
			name:      "pickup only",
			specValue: "supports pickup operations",
			want:      []string{"pickup"},
		},
		{
			name:      "delivery only",
			specValue: "supports delivery operations",
			want:      []string{"delivery"},
		},
		{
			name:      "both pickup and delivery",
			specValue: "supports pickup and delivery operations",
			want:      []string{"pickup", "delivery"},
		},
		{
			name:      "no operations mentioned",
			specValue: "general service",
			want:      []string{"pickup", "delivery"}, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformer.ExtractOperationTypes(tt.specValue)
			if len(result) != len(tt.want) {
				t.Errorf("ExtractOperationTypes() length = %d, want %d", len(result), len(tt.want))
				return
			}
			for i, op := range result {
				if op != tt.want[i] {
					t.Errorf("ExtractOperationTypes()[%d] = %s, want %s", i, op, tt.want[i])
				}
			}
		})
	}
}

func TestExtractPaymentModes(t *testing.T) {
	transformer := NewPartnerDataTransformer(nil)

	tests := []struct {
		name      string
		specValue string
		want      []string
	}{
		{
			name:      "online only",
			specValue: "ONLINE payments accepted",
			want:      []string{"ONLINE"},
		},
		{
			name:      "cod only",
			specValue: "COD payments accepted",
			want:      []string{"COD"},
		},
		{
			name:      "both online and cod",
			specValue: "ONLINE and COD payments accepted",
			want:      []string{"ONLINE", "COD"},
		},
		{
			name:      "no payment modes mentioned",
			specValue: "general service",
			want:      []string{"ONLINE", "COD"}, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformer.ExtractPaymentModes(tt.specValue)
			if len(result) != len(tt.want) {
				t.Errorf("ExtractPaymentModes() length = %d, want %d", len(result), len(tt.want))
				return
			}
			for i, mode := range result {
				if mode != tt.want[i] {
					t.Errorf("ExtractPaymentModes()[%d] = %s, want %s", i, mode, tt.want[i])
				}
			}
		})
	}
}

func TestExtractDeliveryModes(t *testing.T) {
	transformer := NewPartnerDataTransformer(nil)

	tests := []struct {
		name      string
		specValue string
		want      []string
	}{
		{
			name:      "air only",
			specValue: "AIR delivery mode",
			want:      []string{"AIR"},
		},
		{
			name:      "surface only",
			specValue: "SURFACE delivery mode",
			want:      []string{"SURFACE"},
		},
		{
			name:      "rail only",
			specValue: "RAIL delivery mode",
			want:      []string{"RAIL"},
		},
		{
			name:      "multiple modes",
			specValue: "AIR and SURFACE delivery modes",
			want:      []string{"AIR", "SURFACE"},
		},
		{
			name:      "no delivery modes mentioned",
			specValue: "general service",
			want:      []string{"AIR", "SURFACE"}, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformer.ExtractDeliveryModes(tt.specValue)
			if len(result) != len(tt.want) {
				t.Errorf("ExtractDeliveryModes() length = %d, want %d", len(result), len(tt.want))
				return
			}
			for i, mode := range result {
				if mode != tt.want[i] {
					t.Errorf("ExtractDeliveryModes()[%d] = %s, want %s", i, mode, tt.want[i])
				}
			}
		})
	}
}

func TestValidateAndNormalizePartnerData(t *testing.T) {
	transformer := NewPartnerDataTransformer(nil)

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
			name: "valid capability should be normalized",
			capability: &interfaces.PartnerCapability{
				PartnerID:        1,
				PartnerName:      "  Test Partner  ", // will be trimmed
				ServiceTypes:     []string{""},       // will be defaulted
				ParcelCategories: []string{},         // will be defaulted
				OperationTypes:   []string{},         // will be defaulted
				PaymentModes:     []string{},         // will be defaulted
				DeliveryModes:    []string{},         // will be defaulted
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transformer.ValidateAndNormalizePartnerData(tt.capability)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateAndNormalizePartnerData() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateAndNormalizePartnerData() unexpected error: %v", err)
				return
			}

			// Check normalization
			if tt.capability.PartnerName != "Test Partner" {
				t.Errorf("PartnerName not trimmed correctly: got %s", tt.capability.PartnerName)
			}

			// Check defaults were applied
			if len(tt.capability.ServiceTypes) == 0 {
				t.Error("ServiceTypes should have defaults")
			}
			if len(tt.capability.ParcelCategories) == 0 {
				t.Error("ParcelCategories should have defaults")
			}
			if len(tt.capability.OperationTypes) == 0 {
				t.Error("OperationTypes should have defaults")
			}
			if len(tt.capability.PaymentModes) == 0 {
				t.Error("PaymentModes should have defaults")
			}
			if len(tt.capability.DeliveryModes) == 0 {
				t.Error("DeliveryModes should have defaults")
			}
		})
	}
}

func TestTransformSpecificationToCapability(t *testing.T) {
	transformer := NewPartnerDataTransformer(nil)

	tests := []struct {
		name    string
		spec    *interfaces.EntitySpecification
		catalog *interfaces.Catalog
		wantErr bool
	}{
		{
			name:    "nil spec should return error",
			spec:    nil,
			catalog: &interfaces.Catalog{},
			wantErr: true,
		},
		{
			name:    "nil catalog should return error",
			spec:    &interfaces.EntitySpecification{},
			catalog: nil,
			wantErr: true,
		},
		{
			name: "spec value not found in catalog should return error",
			spec: &interfaces.EntitySpecification{
				SpecValue: "NOT_FOUND",
			},
			catalog: &interfaces.Catalog{
				CatalogItems: []interfaces.CatalogItem{
					{Code: "FOUND", Name: "Found Item"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid spec and catalog should succeed",
			spec: &interfaces.EntitySpecification{
				SpecValue: "SDD",
			},
			catalog: &interfaces.Catalog{
				CatalogItems: []interfaces.CatalogItem{
					{
						Code:        "SDD",
						Name:        "Same Day Delivery",
						Description: "Fast pickup and delivery service with ONLINE and COD payments via AIR transport",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformSpecificationToCapability(tt.spec, tt.catalog)

			if tt.wantErr {
				if err == nil {
					t.Error("TransformSpecificationToCapability() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("TransformSpecificationToCapability() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("TransformSpecificationToCapability() returned nil result")
				return
			}

			// Validate result structure
			if result.ServiceType == "" {
				t.Error("ServiceType should not be empty")
			}
			if result.ParcelCategory == "" {
				t.Error("ParcelCategory should not be empty")
			}
			if len(result.OperationTypes) == 0 {
				t.Error("OperationTypes should not be empty")
			}
			if len(result.PaymentModes) == 0 {
				t.Error("PaymentModes should not be empty")
			}
			if len(result.DeliveryModes) == 0 {
				t.Error("DeliveryModes should not be empty")
			}
			if !result.IsActive {
				t.Error("IsActive should be true by default")
			}
		})
	}
}
