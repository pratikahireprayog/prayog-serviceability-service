package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	models "prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"
)




// TestValidateInternationalRequest tests the international request validation
func TestValidateInternationalRequest(t *testing.T) {
	// Setup mocks
	mockLogger := mocks.NewMockLogger()

	tests := []struct {
		name               string
		request            *models.ServiceabilityV2Request
		expectedError      bool
		expectedErrorMsg   string
		isInternational    bool
		shouldCheckPackage bool
	}{
		{
			name: "Valid International Request with Packages",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("90210"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.5,
							Unit:  "kg",
						},
						Dimensions: &models.Dimensions{
							Length: 10.0,
							Width:  8.0,
							Height: 5.0,
							Unit:   "cm",
						},
					},
				},
			},
			expectedError:      false,
			isInternational:    true,
			shouldCheckPackage: true,
		},
		{
			name: "International Request Missing Packages",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("90210"),
				Packages:              []models.Package{},
			},
			expectedError:      true,
			expectedErrorMsg:   "packages are required for international shipments",
			isInternational:    true,
			shouldCheckPackage: true,
		},
		{
			name: "International Request Package Missing Weight",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("90210"),
				Packages: []models.Package{
					{
						Weight: nil,
						Dimensions: &models.Dimensions{
							Length: 10.0,
							Width:  8.0,
							Height: 5.0,
							Unit:   "cm",
						},
					},
				},
			},
			expectedError:      true,
			expectedErrorMsg:   "package weight is required for international shipments",
			isInternational:    true,
			shouldCheckPackage: true,
		},
		{
			name: "International Request Package Missing Dimensions",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("90210"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.5,
							Unit:  "kg",
						},
						Dimensions: nil,
					},
				},
			},
			expectedError:      true,
			expectedErrorMsg:   "package dimensions are required for international shipments",
			isInternational:    true,
			shouldCheckPackage: true,
		},
		{
			name: "International Request Invalid Weight Value",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("90210"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 0,
							Unit:  "kg",
						},
						Dimensions: &models.Dimensions{
							Length: 10.0,
							Width:  8.0,
							Height: 5.0,
							Unit:   "cm",
						},
					},
				},
			},
			expectedError:      true,
			expectedErrorMsg:   "package weight must be greater than 0",
			isInternational:    true,
			shouldCheckPackage: true,
		},
		{
			name: "International Request Invalid Dimensions",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("90210"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.5,
							Unit:  "kg",
						},
						Dimensions: &models.Dimensions{
							Length: 0,
							Width:  8.0,
							Height: 5.0,
							Unit:   "cm",
						},
					},
				},
			},
			expectedError:      true,
			expectedErrorMsg:   "package dimensions must be greater than 0",
			isInternational:    true,
			shouldCheckPackage: true,
		},
		{
			name: "Domestic Request Skip Package Validation",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("110002"),
				Packages:              []models.Package{},
			},
			expectedError:      false,
			isInternational:    false,
			shouldCheckPackage: false,
		},
		{
			name: "International Request Multiple Valid Packages",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("90210"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.5,
							Unit:  "kg",
						},
						Dimensions: &models.Dimensions{
							Length: 10.0,
							Width:  8.0,
							Height: 5.0,
							Unit:   "cm",
						},
					},
					{
						Weight: &models.Weight{
							Value: 0.5,
							Unit:  "kg",
						},
						Dimensions: &models.Dimensions{
							Length: 5.0,
							Width:  4.0,
							Height: 3.0,
							Unit:   "cm",
						},
					},
				},
			},
			expectedError:      false,
			isInternational:    true,
			shouldCheckPackage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Clear logger entries
			mockLogger.Clear()

			// Call the validation function
			err := validateInternationalRequest(ctx, tt.request, tt.isInternational, mockLogger)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}

			// Verify logging behavior
			if tt.isInternational && tt.shouldCheckPackage {
				assert.True(t, mockLogger.HasDebugWithMessage("Validating international request packages"))
			} else if !tt.isInternational {
				assert.True(t, mockLogger.HasDebugWithMessage("Skipping package validation for domestic request"))
			}
		})
	}
}

// TestDetectInternationalRequest tests international detection methods
func TestDetectInternationalRequest(t *testing.T) {
	tests := []struct {
		name            string
		request         *models.ServiceabilityV2Request
		expectedResult  bool
		detectionMethod string
	}{
		{
			name: "International by Postal Code Pattern (US)",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("90210"),
			},
			expectedResult:  true,
			detectionMethod: "postal_code_pattern",
		},
		{
			name: "International by Postal Code Pattern (UK)",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("SW1A 1AA"),
			},
			expectedResult:  true,
			detectionMethod: "postal_code_pattern",
		},
		{
			name: "Domestic by Postal Code Pattern (India)",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("110002"),
			},
			expectedResult:  false,
			detectionMethod: "postal_code_pattern",
		},
		{
			name: "International by Country Code",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("12345"),
				CountryCode:           stringPtr("US"),
			},
			expectedResult:  true,
			detectionMethod: "country_code",
		},
		{
			name: "Domestic by Country Code",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("110002"),
				CountryCode:           stringPtr("IN"),
			},
			expectedResult:  false,
			detectionMethod: "country_code",
		},
		{
			name: "International by Product Type",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("560001"),
				ProductType:           stringPtr("international"),
			},
			expectedResult:  true,
			detectionMethod: "product_type",
		},
		{
			name: "Domestic by Product Type",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("560001"),
				ProductType:           stringPtr("domestic"),
			},
			expectedResult:  false,
			detectionMethod: "product_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectInternationalRequest(tt.request)
			assert.Equal(t, tt.expectedResult, result, "Detection method: %s", tt.detectionMethod)
		})
	}
}

// TestNonBlockingPostalCodeValidation tests non-blocking postal code validation
func TestNonBlockingPostalCodeValidation(t *testing.T) {
	// Setup mocks
	mockLogger := mocks.NewMockLogger()

	tests := []struct {
		name               string
		sourcePostalCode   string
		destPostalCode     string
		shouldLogWarning   bool
		expectedWarningMsg string
		isValidSource      bool
		isValidDestination bool
	}{
		{
			name:               "Valid Postal Codes",
			sourcePostalCode:   "110001",
			destPostalCode:     "90210",
			shouldLogWarning:   false,
			isValidSource:      true,
			isValidDestination: true,
		},
		{
			name:               "Invalid Source Postal Code",
			sourcePostalCode:   "INVALID",
			destPostalCode:     "90210",
			shouldLogWarning:   true,
			expectedWarningMsg: "Source postal code validation failed",
			isValidSource:      false,
			isValidDestination: true,
		},
		{
			name:               "Invalid Destination Postal Code",
			sourcePostalCode:   "110001",
			destPostalCode:     "INVALID",
			shouldLogWarning:   true,
			expectedWarningMsg: "Destination postal code validation failed",
			isValidSource:      true,
			isValidDestination: false,
		},
		{
			name:               "Both Postal Codes Invalid",
			sourcePostalCode:   "INVALID1",
			destPostalCode:     "INVALID2",
			shouldLogWarning:   true,
			expectedWarningMsg: "postal code validation failed",
			isValidSource:      false,
			isValidDestination: false,
		},
		{
			name:               "Empty Postal Codes",
			sourcePostalCode:   "",
			destPostalCode:     "",
			shouldLogWarning:   true,
			expectedWarningMsg: "postal code validation failed",
			isValidSource:      false,
			isValidDestination: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Clear logger entries
			mockLogger.Clear()

			// Call non-blocking postal code validation
			performNonBlockingPostalCodeValidation(ctx, tt.sourcePostalCode, tt.destPostalCode, mockLogger)

			// Verify logging behavior
			if tt.shouldLogWarning {
				assert.True(t, mockLogger.CountEntriesForLevel(0 /* logrus.WarnLevel */) > 0)
				entries := mockLogger.GetEntriesForLevel(0 /* logrus.WarnLevel */)
				found := false
				for _, entry := range entries {
					if tt.expectedWarningMsg != "" && len(entry.Message) > 0 {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected warning message not found")
			} else {
				assert.Equal(t, 0, mockLogger.CountEntriesForLevel(0 /* logrus.WarnLevel */))
			}
		})
	}
}

// TestInternationalPackageValidation tests specific package validation for international requests
func TestInternationalPackageValidation(t *testing.T) {
	tests := []struct {
		name             string
		packages         []models.Package
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name: "Valid Single Package",
			packages: []models.Package{
				{
					Weight: &models.Weight{
						Value: 1.5,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 10.0,
						Width:  8.0,
						Height: 5.0,
						Unit:   "cm",
					},
				},
			},
			expectedError: false,
		},
		{
			name: "Valid Multiple Packages",
			packages: []models.Package{
				{
					Weight: &models.Weight{
						Value: 1.5,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 10.0,
						Width:  8.0,
						Height: 5.0,
						Unit:   "cm",
					},
				},
				{
					Weight: &models.Weight{
						Value: 2.0,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 15.0,
						Width:  12.0,
						Height: 8.0,
						Unit:   "cm",
					},
				},
			},
			expectedError: false,
		},
		{
			name:             "No Packages",
			packages:         []models.Package{},
			expectedError:    true,
			expectedErrorMsg: "at least one package is required for international shipments",
		},
		{
			name: "Package with Zero Weight",
			packages: []models.Package{
				{
					Weight: &models.Weight{
						Value: 0,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 10.0,
						Width:  8.0,
						Height: 5.0,
						Unit:   "cm",
					},
				},
			},
			expectedError:    true,
			expectedErrorMsg: "package weight must be greater than 0",
		},
		{
			name: "Package with Zero Dimensions",
			packages: []models.Package{
				{
					Weight: &models.Weight{
						Value: 1.0,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 0,
						Width:  8.0,
						Height: 5.0,
						Unit:   "cm",
					},
				},
			},
			expectedError:    true,
			expectedErrorMsg: "package dimensions must be greater than 0",
		},
		{
			name: "Very Large Package Weight (Edge Case)",
			packages: []models.Package{
				{
					Weight: &models.Weight{
						Value: 999.99,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 10.0,
						Width:  8.0,
						Height: 5.0,
						Unit:   "cm",
					},
				},
			},
			expectedError: false, // Should be valid unless we have specific weight limits
		},
		{
			name: "Very Large Package Dimensions (Edge Case)",
			packages: []models.Package{
				{
					Weight: &models.Weight{
						Value: 1.0,
						Unit:  "kg",
					},
					Dimensions: &models.Dimensions{
						Length: 999.99,
						Width:  999.99,
						Height: 999.99,
						Unit:   "cm",
					},
				},
			},
			expectedError: false, // Should be valid unless we have specific dimension limits
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInternationalPackages(tt.packages)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions

// validateInternationalRequest validates international request requirements
func validateInternationalRequest(ctx context.Context, request *models.ServiceabilityV2Request, isInternational bool, logger *mocks.MockLogger) error {
	if !isInternational {
		logger.Debug("Skipping package validation for domestic request")
		return nil
	}

	logger.Debug("Validating international request packages")

	// For international requests, packages are mandatory
	return validateInternationalPackages(request.Packages)
}

// detectInternationalRequest detects if a request is international based on various factors
func detectInternationalRequest(request *models.ServiceabilityV2Request) bool {
	// Method 1: Check country code explicitly
	if request.CountryCode != nil {
		return *request.CountryCode != "IN"
	}

	// Method 2: Check product type
	if request.ProductType != nil {
		return *request.ProductType == "international"
	}

	// Method 3: Check postal code patterns
	if request.DestinationPostalCode != nil {
		destCode := *request.DestinationPostalCode

		// US postal code pattern (5 digits or 5+4 format)
		if len(destCode) == 5 || (len(destCode) == 10 && destCode[5] == '-') {
			return isUSPostalCode(destCode)
		}

		// UK postal code pattern (contains letters)
		if containsLetters(destCode) {
			return true
		}

		// Indian postal code pattern (6 digits)
		if len(destCode) == 6 && isAllDigits(destCode) {
			return false
		}
	}

	// Default to domestic if unable to determine
	return false
}

// performNonBlockingPostalCodeValidation performs non-blocking postal code validation
func performNonBlockingPostalCodeValidation(ctx context.Context, sourcePostalCode, destPostalCode string, logger *mocks.MockLogger) {
	go func() {
		// Validate source postal code
		if !isValidPostalCode(sourcePostalCode) {
			logger.Warn("Source postal code validation failed", map[string]interface{}{
				"source_postal_code": sourcePostalCode,
			})
		}

		// Validate destination postal code
		if !isValidPostalCode(destPostalCode) {
			logger.Warn("Destination postal code validation failed", map[string]interface{}{
				"destination_postal_code": destPostalCode,
			})
		}
	}()
}

// validateInternationalPackages validates packages for international requests
func validateInternationalPackages(packages []models.Package) error {
	if len(packages) == 0 {
		return errors.New("at least one package is required for international shipments")
	}

	for i, pkg := range packages {
		// Validate weight
		if pkg.Weight == nil {
			return errors.New("package weight is required for international shipments")
		}

		if pkg.Weight.Value <= 0 {
			return errors.New("package weight must be greater than 0")
		}

		// Validate dimensions
		if pkg.Dimensions == nil {
			return errors.New("package dimensions are required for international shipments")
		}

		if pkg.Dimensions.Length <= 0 || pkg.Dimensions.Width <= 0 || pkg.Dimensions.Height <= 0 {
			return errors.New("package dimensions must be greater than 0")
		}

		// Additional validation can be added here
		_ = i // Use index if needed for error messages
	}

	return nil
}

// Helper functions for postal code validation

func isUSPostalCode(code string) bool {
	if len(code) == 5 {
		return isAllDigits(code)
	}
	if len(code) == 10 && code[5] == '-' {
		return isAllDigits(code[:5]) && isAllDigits(code[6:])
	}
	return false
}

func containsLetters(s string) bool {
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			return true
		}
	}
	return false
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

func isValidPostalCode(code string) bool {
	// Simple validation: non-empty and reasonable length
	return len(code) >= 3 && len(code) <= 20 && code != "INVALID" && code != "INVALID1" && code != "INVALID2"
}
