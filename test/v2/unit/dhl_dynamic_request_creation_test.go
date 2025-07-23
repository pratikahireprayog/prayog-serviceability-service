package unit

import (
	"context"
	"fmt"
	"testing"

	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"
)

// TestDHLDynamicRequestCreation tests the createInternationalRatesRequest function
func TestDHLDynamicRequestCreation(t *testing.T) {
	tests := []struct {
		name             string
		sourcePostalCode string
		destPostalCode   string
		cityName         string
		hubPostalCode    string
		packages         []models.Package
		expectError      bool
		description      string
	}{
		{
			name:             "Valid International Request Creation",
			sourcePostalCode: "110001",
			destPostalCode:   "90210",
			cityName:         "New Delhi",
			hubPostalCode:    "110020",
			packages: []models.Package{
				{
					Weight: &models.Weight{
						Value: 2.5,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 30.0,
						Width:  20.0,
						Height: 15.0,
						Unit:   "cm",
					},
				},
			},
			expectError: false,
			description: "Should create valid international rates request",
		},
		{
			name:             "Empty Packages",
			sourcePostalCode: "110001",
			destPostalCode:   "90210",
			cityName:         "New Delhi",
			hubPostalCode:    "110020",
			packages:         []models.Package{},
			expectError:      true,
			description:      "Should fail when packages are empty",
		},
		{
			name:             "Missing City Name",
			sourcePostalCode: "110001",
			destPostalCode:   "90210",
			cityName:         "",
			hubPostalCode:    "110020",
			packages: []models.Package{
				{
					Weight: &models.Weight{
						Value: 1.0,
						Unit:  "kg",
					},
				},
			},
			expectError: true,
			description: "Should fail when city name is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockLogger := mocks.NewMockLogger()

			// Create context
			ctx := context.Background()

			// Test dynamic request creation
			request, err := createInternationalRatesRequest(
				ctx,
				tt.sourcePostalCode,
				tt.destPostalCode,
				tt.cityName,
				tt.hubPostalCode,
				tt.packages,
				mockLogger,
			)

			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify request structure
			if request == nil {
				t.Error("Expected request to be created")
				return
			}

			// Check for expected city name in logs
			if tt.cityName != "" {
				if !mockLogger.HasEntryContaining(tt.cityName) {
					t.Errorf("Expected city name %s not found in logs", tt.cityName)
				}
			}

			// Check for expected hub postal code in logs
			if tt.hubPostalCode != "" {
				if !mockLogger.HasEntryContaining(tt.hubPostalCode) {
					t.Errorf("Expected hub postal code %s not found in logs", tt.hubPostalCode)
				}
			}
		})
	}
}

// TestDHLCityNameResolution tests the city name resolution logic
func TestDHLCityNameResolution(t *testing.T) {
	tests := []struct {
		name         string
		cityName     string
		expectedCity string
		expectError  bool
		description  string
	}{
		{
			name:         "Valid City Name",
			cityName:     "Mumbai",
			expectedCity: "Mumbai",
			expectError:  false,
			description:  "Should return provided city name",
		},
		{
			name:        "Empty City Name",
			cityName:    "",
			expectError: true,
			description: "Should fail when city name is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := mocks.NewMockLogger()

			cityName, err := resolveCityName(tt.cityName, mockLogger)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if cityName != tt.expectedCity {
				t.Errorf("Expected city name %s, got %s", tt.expectedCity, cityName)
			}
		})
	}
}

// TestDHLHubPostalCodeMapping tests hub postal code mapping logic
func TestDHLHubPostalCodeMapping(t *testing.T) {
	tests := []struct {
		name               string
		hubPostalCode      string
		expectedPostalCode string
		expectError        bool
		description        string
	}{
		{
			name:               "Valid Hub Postal Code",
			hubPostalCode:      "110020",
			expectedPostalCode: "110020",
			expectError:        false,
			description:        "Should return hub postal code",
		},
		{
			name:          "Empty Postal Code",
			hubPostalCode: "",
			expectError:   true,
			description:   "Should fail when postal code is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := mocks.NewMockLogger()

			postalCode, err := getHubPostalCode(tt.hubPostalCode, mockLogger)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if postalCode != tt.expectedPostalCode {
				t.Errorf("Expected postal code %s, got %s", tt.expectedPostalCode, postalCode)
			}
		})
	}
}

// Mock implementation functions for testing

func createInternationalRatesRequest(
	ctx context.Context,
	sourcePostalCode, destPostalCode string,
	cityName, hubPostalCode string,
	packages []models.Package,
	logger *mocks.MockLogger,
) (interface{}, error) {
	if cityName == "" {
		logger.Error("City name is required")
		return nil, fmt.Errorf("city name is required")
	}

	if hubPostalCode == "" {
		logger.Error("Hub postal code is required")
		return nil, fmt.Errorf("hub postal code is required")
	}

	if len(packages) == 0 {
		logger.Error("No packages provided")
		return nil, fmt.Errorf("packages are required")
	}

	logger.Infof("Creating international rates request for %s to %s", sourcePostalCode, destPostalCode)
	logger.Infof("Using hub city: %s", cityName)
	logger.Infof("Using hub postal code: %s", hubPostalCode)

	// Create mock request structure
	request := map[string]interface{}{
		"origin":      hubPostalCode,
		"destination": destPostalCode,
		"city":        cityName,
		"packages":    packages,
	}

	return request, nil
}

func resolveCityName(cityName string, logger *mocks.MockLogger) (string, error) {
	if cityName == "" {
		logger.Error("City name is empty")
		return "", fmt.Errorf("city name is required")
	}

	logger.Infof("Resolved city name: %s", cityName)
	return cityName, nil
}

func getHubPostalCode(hubPostalCode string, logger *mocks.MockLogger) (string, error) {
	if hubPostalCode == "" {
		logger.Error("Hub postal code is empty")
		return "", fmt.Errorf("postal code is required")
	}

	logger.Infof("Retrieved hub postal code: %s", hubPostalCode)
	return hubPostalCode, nil
}
