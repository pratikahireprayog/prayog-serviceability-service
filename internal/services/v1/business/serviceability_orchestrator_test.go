package services

import (
	"context"
	"testing"

	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// Mock implementations for testing

type mockLocationResolver struct{}

func (m *mockLocationResolver) ValidatePostalCode(ctx context.Context, postalCode, countryCode string) (*models.LocationValidationResult, error) {
	return &models.LocationValidationResult{
		IsValid: true,
		Hierarchy: &models.LocationHierarchy{
			PostalCode:  postalCode,
			CountryCode: countryCode,
			CountryName: "India",
			RegionCode:  "KA",
			RegionName:  "Karnataka",
			CityCode:    "BLR",
			CityName:    "Bangalore",
			AreaCode:    "HSR",
			AreaName:    "HSR Layout",
		},
	}, nil
}

func (m *mockLocationResolver) GetLocationHierarchy(ctx context.Context, postalCode, countryCode string) (*models.LocationHierarchy, error) {
	return &models.LocationHierarchy{
		PostalCode:  postalCode,
		CountryCode: countryCode,
		CountryName: "India",
		RegionCode:  "KA",
		RegionName:  "Karnataka",
		CityCode:    "BLR",
		CityName:    "Bangalore",
		AreaCode:    "HSR",
		AreaName:    "HSR Layout",
	}, nil
}

func (m *mockLocationResolver) IsPostalCodeActive(ctx context.Context, postalCode, countryCode string) (bool, error) {
	// Active for specific test postal codes
	return postalCode == "110001" || postalCode == "400001", nil
}

type mockPartnerCapabilityAggregator struct{}

func (m *mockPartnerCapabilityAggregator) GetPartnersByLocation(ctx context.Context, locationHierarchy *models.LocationHierarchy) ([]interfaces.PartnerCapability, error) {
	return []interfaces.PartnerCapability{
		{
			PartnerID:        1,
			PartnerName:      "Test Partner 1",
			IsActive:         true,
			ServiceTypes:     []string{"SDD", "NDD"},
			ParcelCategories: []string{"ecom"},
			OperationTypes:   []string{"pickup", "delivery"},
			PaymentModes:     []string{"ONLINE", "COD"},
			DeliveryModes:    []string{"AIR", "SURFACE"},
		},
	}, nil
}

func (m *mockPartnerCapabilityAggregator) GetPartnerCapabilities(ctx context.Context, partnerID uint, serviceFilters *interfaces.ServiceFilters) (*interfaces.PartnerServiceCapabilities, error) {
	return &interfaces.PartnerServiceCapabilities{
		PartnerID: partnerID,
		Capabilities: []interfaces.ServiceCapability{
			{
				ServiceType:    "SDD",
				ParcelCategory: "ecom",
				OperationTypes: []string{"pickup", "delivery"},
				PaymentModes:   []string{"ONLINE", "COD"},
				DeliveryModes:  []string{"AIR"},
				IsActive:       true,
				Rating:         4.5,
			},
		},
	}, nil
}

type mockServiceDefinitionResolver struct{}

func (m *mockServiceDefinitionResolver) GetParcelCategories(ctx context.Context) ([]string, error) {
	return []string{"ecom", "courier", "cargo"}, nil
}

func (m *mockServiceDefinitionResolver) GetServiceTypes(ctx context.Context, parcelCategory string) ([]string, error) {
	switch parcelCategory {
	case "ecom":
		return []string{"SDD", "NDD", "Standard"}, nil
	case "courier":
		return []string{"Express", "Standard"}, nil
	default:
		return []string{}, nil
	}
}

func (m *mockServiceDefinitionResolver) GetOperationTypes(ctx context.Context) ([]string, error) {
	return []string{"pickup", "delivery"}, nil
}

func (m *mockServiceDefinitionResolver) GetPaymentModes(ctx context.Context) ([]string, error) {
	return []string{"ONLINE", "COD"}, nil
}

func (m *mockServiceDefinitionResolver) GetDeliveryModes(ctx context.Context) ([]string, error) {
	return []string{"AIR", "SURFACE", "RAIL"}, nil
}

func (m *mockServiceDefinitionResolver) GetServiceDefinition(ctx context.Context, serviceType string) (*interfaces.ServiceDefinition, error) {
	return &interfaces.ServiceDefinition{
		ServiceType:       serviceType,
		ParcelCategory:    "ecom",
		DefaultOperations: []string{"pickup", "delivery"},
		DefaultPayments:   []string{"ONLINE", "COD"},
		DefaultDelivery:   []string{"AIR"},
		Description:       "Test service definition",
	}, nil
}

type mockServiceabilityCalculator struct{}

func (m *mockServiceabilityCalculator) CalculateServiceability(ctx context.Context, request *interfaces.ServiceabilityCalculationRequest) (*models.ServiceabilityData, error) {
	// Create mock serviceability data
	data := &models.ServiceabilityData{
		QueryType: "generic_location",
	}

	if request.LocationHierarchy != nil {
		data.Location = &models.LocationData{
			PostalCode:  request.LocationHierarchy.PostalCode,
			CountryCode: request.LocationHierarchy.CountryCode,
			Serviceability: []models.ParcelCategoryService{
				{
					ParcelCategory: "ecom",
					Services: []models.Service{
						{
							ServiceType:    "SDD",
							OperationTypes: []string{"pickup", "delivery"},
							PaymentModes:   []string{"ONLINE", "COD"},
							DeliveryModes:  []string{"AIR"},
						},
					},
				},
			},
		}
	}

	return data, nil
}

func (m *mockServiceabilityCalculator) FilterServicesByPreferences(ctx context.Context, services []interfaces.ServiceOption, preferences []interfaces.PartnerPreference) ([]interfaces.ServiceOption, error) {
	return services, nil
}

func (m *mockServiceabilityCalculator) CombineRouteCapabilities(ctx context.Context, pickupServices, deliveryServices []interfaces.ServiceOption) ([]interfaces.ServiceOption, error) {
	return pickupServices, nil
}

// Test functions

func TestNewServiceabilityOrchestrator(t *testing.T) {
	locationResolver := &mockLocationResolver{}
	partnerAggregator := &mockPartnerCapabilityAggregator{}
	serviceDefinitionResolver := &mockServiceDefinitionResolver{}
	serviceabilityCalculator := &mockServiceabilityCalculator{}

	orchestrator := NewServiceabilityOrchestrator(
		locationResolver,
		partnerAggregator,
		serviceDefinitionResolver,
		serviceabilityCalculator,
	)

	if orchestrator == nil {
		t.Fatal("Expected orchestrator to be created, got nil")
	}
}

func TestCheckServiceability_SingleLocation_Success(t *testing.T) {
	// Setup
	orchestrator := setupTestOrchestrator()
	ctx := context.Background()

	postalCode := "110001"
	req := &models.ServiceabilityCheckRequest{
		PostalCode:  &postalCode,
		CountryCode: "IN",
	}

	// Execute
	response, err := orchestrator.CheckServiceability(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if !response.Success {
		t.Errorf("Expected success=true, got: %v", response.Success)
	}

	if response.Data == nil {
		t.Fatal("Expected data to be present")
	}

	if response.Data.QueryType != "generic_location" {
		t.Errorf("Expected query_type='generic_location', got: %v", response.Data.QueryType)
	}
}

func TestCheckServiceability_SingleLocation_InactivePostalCode(t *testing.T) {
	// Setup
	orchestrator := setupTestOrchestrator()
	ctx := context.Background()

	postalCode := "999999" // Inactive postal code
	req := &models.ServiceabilityCheckRequest{
		PostalCode:  &postalCode,
		CountryCode: "IN",
	}

	// Execute
	response, err := orchestrator.CheckServiceability(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if !response.Success {
		t.Errorf("Expected success=true, got: %v", response.Success)
	}

	if response.Error == nil {
		t.Fatal("Expected error to be present for inactive postal code")
	}

	if response.Error.Code != "POSTAL_CODE_INACTIVE" {
		t.Errorf("Expected error code='POSTAL_CODE_INACTIVE', got: %v", response.Error.Code)
	}
}

func TestCheckServiceability_OriginDestination_Success(t *testing.T) {
	// Setup
	orchestrator := setupTestOrchestrator()
	ctx := context.Background()

	pickupPostalCode := "110001"
	deliveryPostalCode := "400001"
	req := &models.ServiceabilityCheckRequest{
		PickupPostalCode:   &pickupPostalCode,
		DeliveryPostalCode: &deliveryPostalCode,
		CountryCode:        "IN",
	}

	// Execute
	response, err := orchestrator.CheckServiceability(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if !response.Success {
		t.Errorf("Expected success=true, got: %v", response.Success)
	}

	if response.Data == nil {
		t.Fatal("Expected data to be present")
	}
}

func TestCheckServiceability_ValidationErrors(t *testing.T) {
	orchestrator := setupTestOrchestrator()
	ctx := context.Background()

	testCases := []struct {
		name    string
		request *models.ServiceabilityCheckRequest
	}{
		{
			name:    "nil request",
			request: nil,
		},
		{
			name: "missing country code",
			request: &models.ServiceabilityCheckRequest{
				PostalCode: stringPtr("110001"),
			},
		},
		{
			name: "no postal codes provided",
			request: &models.ServiceabilityCheckRequest{
				CountryCode: "IN",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := orchestrator.CheckServiceability(ctx, tc.request)
			if err == nil {
				t.Errorf("Expected error for test case: %s", tc.name)
			}
		})
	}
}

func TestBulkCheckServiceability_Success(t *testing.T) {
	// Setup
	orchestrator := setupTestOrchestrator()
	ctx := context.Background()

	postalCode1 := "110001"
	postalCode2 := "400001"
	req := &models.BulkServiceabilityRequest{
		Requests: []models.ServiceabilityCheckRequest{
			{
				PostalCode:  &postalCode1,
				CountryCode: "IN",
			},
			{
				PostalCode:  &postalCode2,
				CountryCode: "IN",
			},
		},
	}

	// Execute
	response, err := orchestrator.BulkCheckServiceability(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if !response.Success {
		t.Errorf("Expected success=true, got: %v", response.Success)
	}

	if len(response.Data) != 2 {
		t.Errorf("Expected 2 responses, got: %d", len(response.Data))
	}
}

// Helper functions

func setupTestOrchestrator() interfaces.ServiceabilityOrchestrator {
	return NewServiceabilityOrchestrator(
		&mockLocationResolver{},
		&mockPartnerCapabilityAggregator{},
		&mockServiceDefinitionResolver{},
		&mockServiceabilityCalculator{},
	)
}

func stringPtr(s string) *string {
	return &s
}
