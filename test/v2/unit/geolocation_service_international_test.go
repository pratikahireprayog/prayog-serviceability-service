package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"prayog-serviceability-service/test/v2/mocks"
)

// TestGeolocationValidateInternationalRequest tests the geolocation service international request validation
func TestGeolocationValidateInternationalRequest(t *testing.T) {
	// Setup mocks
	mockLogger := mocks.NewMockLogger()

	tests := []struct {
		name                    string
		sourcePostalCode        string
		destinationPostalCode   string
		expectedIsInternational bool
		expectedSourceCountry   string
		expectedDestCountry     string
		expectedError           bool
		expectedErrorMsg        string
		setupCountryResolution  func() map[string]string
	}{
		{
			name:                    "International Request India to US",
			sourcePostalCode:        "110001",
			destinationPostalCode:   "90210",
			expectedIsInternational: true,
			expectedSourceCountry:   "IN",
			expectedDestCountry:     "US",
			expectedError:           false,
			setupCountryResolution: func() map[string]string {
				return map[string]string{
					"110001": "IN",
					"90210":  "US",
				}
			},
		},
		{
			name:                    "International Request India to UK",
			sourcePostalCode:        "400001",
			destinationPostalCode:   "SW1A 1AA",
			expectedIsInternational: true,
			expectedSourceCountry:   "IN",
			expectedDestCountry:     "GB",
			expectedError:           false,
			setupCountryResolution: func() map[string]string {
				return map[string]string{
					"400001":   "IN",
					"SW1A 1AA": "GB",
				}
			},
		},
		{
			name:                    "Domestic Request India to India",
			sourcePostalCode:        "110001",
			destinationPostalCode:   "110002",
			expectedIsInternational: false,
			expectedSourceCountry:   "IN",
			expectedDestCountry:     "IN",
			expectedError:           false,
			setupCountryResolution: func() map[string]string {
				return map[string]string{
					"110001": "IN",
					"110002": "IN",
				}
			},
		},
		{
			name:                    "Source Country Resolution Failed",
			sourcePostalCode:        "INVALID_SOURCE",
			destinationPostalCode:   "90210",
			expectedIsInternational: false,
			expectedError:           true,
			expectedErrorMsg:        "failed to resolve source country code",
			setupCountryResolution: func() map[string]string {
				return map[string]string{
					"90210": "US",
				}
			},
		},
		{
			name:                    "Destination Country Resolution Failed",
			sourcePostalCode:        "110001",
			destinationPostalCode:   "INVALID_DEST",
			expectedIsInternational: false,
			expectedError:           true,
			expectedErrorMsg:        "failed to resolve destination country code",
			setupCountryResolution: func() map[string]string {
				return map[string]string{
					"110001": "IN",
				}
			},
		},
		{
			name:                    "Both Country Resolutions Failed",
			sourcePostalCode:        "INVALID_SOURCE",
			destinationPostalCode:   "INVALID_DEST",
			expectedIsInternational: false,
			expectedError:           true,
			expectedErrorMsg:        "failed to resolve country codes",
			setupCountryResolution: func() map[string]string {
				return map[string]string{}
			},
		},
		{
			name:                    "International Request with Uncommon Countries",
			sourcePostalCode:        "110001",
			destinationPostalCode:   "12345",
			expectedIsInternational: true,
			expectedSourceCountry:   "IN",
			expectedDestCountry:     "DE",
			expectedError:           false,
			setupCountryResolution: func() map[string]string {
				return map[string]string{
					"110001": "IN",
					"12345":  "DE",
				}
			},
		},
		{
			name:                    "Same Country Different Format",
			sourcePostalCode:        "110001",
			destinationPostalCode:   "560001",
			expectedIsInternational: false,
			expectedSourceCountry:   "IN",
			expectedDestCountry:     "IN",
			expectedError:           false,
			setupCountryResolution: func() map[string]string {
				return map[string]string{
					"110001": "IN",
					"560001": "IN",
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Clear logger entries
			mockLogger.Clear()

			// Setup country resolution mapping
			countryMap := tt.setupCountryResolution()

			// Create mock geolocation service
			mockService := &MockGeolocationService{
				countryResolution: countryMap,
				logger:            mockLogger,
			}

			// Call the validation function
			result, err := mockService.ValidateInternationalRequest(ctx, tt.sourcePostalCode, tt.destinationPostalCode)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedIsInternational, result.IsInternational)
				assert.Equal(t, tt.expectedSourceCountry, result.SourceCountryCode)
				assert.Equal(t, tt.expectedDestCountry, result.DestinationCountryCode)
			}

			// Verify logging behavior
			if tt.expectedError {
				assert.True(t, mockLogger.CountEntriesForLevel(3 /* ErrorLevel */) > 0)
			} else {
				assert.True(t, mockLogger.HasInfoWithMessage("Country code resolution completed"))
			}
		})
	}
}

// TestCountryCodeResolution tests specific country code resolution scenarios
func TestCountryCodeResolution(t *testing.T) {
	mockLogger := mocks.NewMockLogger()

	tests := []struct {
		name                string
		postalCode          string
		expectedCountryCode string
		expectedError       bool
		expectedErrorMsg    string
		setupResolution     func() map[string]string
		testType            string
	}{
		{
			name:                "Valid India Postal Code",
			postalCode:          "110001",
			expectedCountryCode: "IN",
			expectedError:       false,
			setupResolution: func() map[string]string {
				return map[string]string{"110001": "IN"}
			},
			testType: "valid",
		},
		{
			name:                "Valid US Postal Code",
			postalCode:          "90210",
			expectedCountryCode: "US",
			expectedError:       false,
			setupResolution: func() map[string]string {
				return map[string]string{"90210": "US"}
			},
			testType: "valid",
		},
		{
			name:                "Valid UK Postal Code",
			postalCode:          "SW1A 1AA",
			expectedCountryCode: "GB",
			expectedError:       false,
			setupResolution: func() map[string]string {
				return map[string]string{"SW1A 1AA": "GB"}
			},
			testType: "valid",
		},
		{
			name:             "Invalid Postal Code",
			postalCode:       "INVALID",
			expectedError:    true,
			expectedErrorMsg: "country code not found",
			setupResolution: func() map[string]string {
				return map[string]string{}
			},
			testType: "invalid",
		},
		{
			name:             "Empty Postal Code",
			postalCode:       "",
			expectedError:    true,
			expectedErrorMsg: "postal code cannot be empty",
			setupResolution: func() map[string]string {
				return map[string]string{}
			},
			testType: "empty",
		},
		{
			name:                "Postal Code with Special Characters",
			postalCode:          "M5V 3L9",
			expectedCountryCode: "CA",
			expectedError:       false,
			setupResolution: func() map[string]string {
				return map[string]string{"M5V 3L9": "CA"}
			},
			testType: "special",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Clear logger entries
			mockLogger.Clear()

			// Setup country resolution mapping
			countryMap := tt.setupResolution()

			// Create mock geolocation service
			mockService := &MockGeolocationService{
				countryResolution: countryMap,
				logger:            mockLogger,
			}

			// Call the resolution function
			countryCode, err := mockService.ResolveCountryCode(ctx, tt.postalCode)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				assert.Empty(t, countryCode)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCountryCode, countryCode)
			}

			// Verify logging behavior based on test type
			switch tt.testType {
			case "valid":
				assert.True(t, mockLogger.HasDebugWithMessage("Resolving country code for postal code"))
			case "invalid":
				assert.True(t, mockLogger.CountEntriesForLevel(4 /* WarnLevel */) > 0)
			case "empty":
				assert.True(t, mockLogger.CountEntriesForLevel(3 /* ErrorLevel */) > 0)
			}
		})
	}
}

// TestInternationalDetectionLogic tests the international detection logic
func TestInternationalDetectionLogic(t *testing.T) {
	tests := []struct {
		name                    string
		sourceCountryCode       string
		destinationCountryCode  string
		expectedIsInternational bool
		description             string
	}{
		{
			name:                    "India to US - International",
			sourceCountryCode:       "IN",
			destinationCountryCode:  "US",
			expectedIsInternational: true,
			description:             "Different countries should be international",
		},
		{
			name:                    "India to India - Domestic",
			sourceCountryCode:       "IN",
			destinationCountryCode:  "IN",
			expectedIsInternational: false,
			description:             "Same country should be domestic",
		},
		{
			name:                    "US to Canada - International",
			sourceCountryCode:       "US",
			destinationCountryCode:  "CA",
			expectedIsInternational: true,
			description:             "US to Canada should be international",
		},
		{
			name:                    "UK to UK - Domestic",
			sourceCountryCode:       "GB",
			destinationCountryCode:  "GB",
			expectedIsInternational: false,
			description:             "UK to UK should be domestic",
		},
		{
			name:                    "Germany to France - International",
			sourceCountryCode:       "DE",
			destinationCountryCode:  "FR",
			expectedIsInternational: true,
			description:             "EU countries are still international",
		},
		{
			name:                    "Case Insensitive Same Country",
			sourceCountryCode:       "in",
			destinationCountryCode:  "IN",
			expectedIsInternational: false,
			description:             "Country codes should be case insensitive",
		},
		{
			name:                    "Case Insensitive Different Countries",
			sourceCountryCode:       "in",
			destinationCountryCode:  "us",
			expectedIsInternational: true,
			description:             "Different country codes should be international regardless of case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInternationalShipment(tt.sourceCountryCode, tt.destinationCountryCode)
			assert.Equal(t, tt.expectedIsInternational, result, tt.description)
		})
	}
}

// TestGeolocationServiceEdgeCases tests edge cases for geolocation service
func TestGeolocationServiceEdgeCases(t *testing.T) {
	mockLogger := mocks.NewMockLogger()

	tests := []struct {
		name                  string
		sourcePostalCode      string
		destinationPostalCode string
		countryResolutionMap  map[string]string
		expectServiceCall     bool
		expectedBehavior      string
	}{
		{
			name:                  "Both Empty Postal Codes",
			sourcePostalCode:      "",
			destinationPostalCode: "",
			countryResolutionMap:  map[string]string{},
			expectServiceCall:     false,
			expectedBehavior:      "Should fail early validation",
		},
		{
			name:                  "Very Long Postal Codes",
			sourcePostalCode:      "VERYLONGPOSTALCODE12345678901234567890",
			destinationPostalCode: "ANOTHERLONGPOSTALCODE09876543210987654321",
			countryResolutionMap: map[string]string{
				"VERYLONGPOSTALCODE12345678901234567890":    "IN",
				"ANOTHERLONGPOSTALCODE09876543210987654321": "US",
			},
			expectServiceCall: true,
			expectedBehavior:  "Should handle long postal codes",
		},
		{
			name:                  "Postal Codes with Special Characters",
			sourcePostalCode:      "110-001",
			destinationPostalCode: "SW1A-1AA",
			countryResolutionMap: map[string]string{
				"110-001":  "IN",
				"SW1A-1AA": "GB",
			},
			expectServiceCall: true,
			expectedBehavior:  "Should handle special characters",
		},
		{
			name:                  "Partial Country Resolution Success",
			sourcePostalCode:      "110001",
			destinationPostalCode: "UNKNOWN",
			countryResolutionMap: map[string]string{
				"110001": "IN",
			},
			expectServiceCall: true,
			expectedBehavior:  "Should fail when destination resolution fails",
		},
		{
			name:                  "Country Resolution Network Timeout",
			sourcePostalCode:      "TIMEOUT_SOURCE",
			destinationPostalCode: "TIMEOUT_DEST",
			countryResolutionMap:  map[string]string{},
			expectServiceCall:     true,
			expectedBehavior:      "Should handle network timeouts gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Clear logger entries
			mockLogger.Clear()

			// Create mock geolocation service
			mockService := &MockGeolocationService{
				countryResolution: tt.countryResolutionMap,
				logger:            mockLogger,
			}

			// Call the validation function
			result, err := mockService.ValidateInternationalRequest(ctx, tt.sourcePostalCode, tt.destinationPostalCode)

			// Basic assertions based on expected behavior
			switch tt.expectedBehavior {
			case "Should fail early validation":
				assert.Error(t, err)
				assert.Nil(t, result)
			case "Should handle long postal codes":
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsInternational)
			case "Should handle special characters":
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsInternational)
			case "Should fail when destination resolution fails":
				assert.Error(t, err)
			case "Should handle network timeouts gracefully":
				assert.Error(t, err)
			}

			// Verify appropriate logging occurred
			assert.True(t, mockLogger.CountEntries() > 0, "Expected some logging to occur")
		})
	}
}

// Mock structures and implementations

// InternationalValidationResult represents the result of international validation
type InternationalValidationResult struct {
	IsInternational        bool   `json:"is_international"`
	SourceCountryCode      string `json:"source_country_code"`
	DestinationCountryCode string `json:"destination_country_code"`
	RequiresPackageInfo    bool   `json:"requires_package_info"`
	ValidationType         string `json:"validation_type"`
}

// MockGeolocationService mocks the geolocation service for testing
type MockGeolocationService struct {
	countryResolution map[string]string
	logger            *mocks.MockLogger
}

// ValidateInternationalRequest validates an international request
func (m *MockGeolocationService) ValidateInternationalRequest(ctx context.Context, sourcePostalCode, destinationPostalCode string) (*InternationalValidationResult, error) {
	m.logger.Info("Validating international request", map[string]interface{}{
		"source_postal_code":      sourcePostalCode,
		"destination_postal_code": destinationPostalCode,
	})

	// Early validation
	if sourcePostalCode == "" || destinationPostalCode == "" {
		m.logger.Error("Postal codes cannot be empty", map[string]interface{}{
			"source_empty":      sourcePostalCode == "",
			"destination_empty": destinationPostalCode == "",
		})
		return nil, errors.New("postal codes cannot be empty")
	}

	// Resolve country codes
	sourceCountry, err := m.ResolveCountryCode(ctx, sourcePostalCode)
	if err != nil {
		m.logger.Error("Failed to resolve source country code", map[string]interface{}{
			"source_postal_code": sourcePostalCode,
			"error":              err.Error(),
		})
		return nil, errors.New("failed to resolve source country code: " + err.Error())
	}

	destCountry, err := m.ResolveCountryCode(ctx, destinationPostalCode)
	if err != nil {
		m.logger.Error("Failed to resolve destination country code", map[string]interface{}{
			"destination_postal_code": destinationPostalCode,
			"error":                   err.Error(),
		})
		return nil, errors.New("failed to resolve destination country code: " + err.Error())
	}

	// Determine if international
	isInternational := isInternationalShipment(sourceCountry, destCountry)

	m.logger.Info("Country code resolution completed", map[string]interface{}{
		"source_country":      sourceCountry,
		"destination_country": destCountry,
		"is_international":    isInternational,
	})

	return &InternationalValidationResult{
		IsInternational:        isInternational,
		SourceCountryCode:      sourceCountry,
		DestinationCountryCode: destCountry,
		RequiresPackageInfo:    isInternational,
		ValidationType:         "geolocation_service",
	}, nil
}

// ResolveCountryCode resolves a postal code to a country code
func (m *MockGeolocationService) ResolveCountryCode(ctx context.Context, postalCode string) (string, error) {
	m.logger.Debug("Resolving country code for postal code", map[string]interface{}{
		"postal_code": postalCode,
	})

	if postalCode == "" {
		return "", errors.New("postal code cannot be empty")
	}

	// Handle timeout scenarios
	if postalCode == "TIMEOUT_SOURCE" || postalCode == "TIMEOUT_DEST" {
		m.logger.Warn("Country resolution timed out", map[string]interface{}{
			"postal_code": postalCode,
		})
		return "", errors.New("country resolution timed out")
	}

	// Lookup in resolution map
	if countryCode, exists := m.countryResolution[postalCode]; exists {
		m.logger.Debug("Country code resolved successfully", map[string]interface{}{
			"postal_code":  postalCode,
			"country_code": countryCode,
		})
		return countryCode, nil
	}

	m.logger.Warn("Country code not found for postal code", map[string]interface{}{
		"postal_code": postalCode,
	})
	return "", errors.New("country code not found for postal code: " + postalCode)
}

// Helper functions

// isInternationalShipment determines if a shipment is international based on country codes
func isInternationalShipment(sourceCountry, destCountry string) bool {
	// Normalize country codes to uppercase for comparison
	source := normalizeCountryCode(sourceCountry)
	dest := normalizeCountryCode(destCountry)

	return source != dest
}

// normalizeCountryCode normalizes country codes to uppercase
func normalizeCountryCode(countryCode string) string {
	result := ""
	for _, r := range countryCode {
		if r >= 'a' && r <= 'z' {
			result += string(r - 32) // Convert to uppercase
		} else {
			result += string(r)
		}
	}
	return result
}
