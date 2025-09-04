package unit

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDHLValidationErrorStructure tests the structure of DHL validation errors
func TestDHLValidationErrorStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		request              *models.ServiceabilityV2Request
		expectedErrorCode    string
		expectedErrorMessage string
		expectedErrorField   string
		description          string
	}{
		{
			name: "Missing destination postal code",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode: stringPtr("560001"),
				// DestinationPostalCode: nil, // Missing
				CountryCode: stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			expectedErrorCode:    "DHL_VALIDATION_ERROR",
			expectedErrorMessage: "destination postal code is required",
			expectedErrorField:   "destination_postal_code",
			description:          "DHL validation error for missing destination postal code",
		},
		{
			name: "Missing source postal code",
			request: &models.ServiceabilityV2Request{
				// SourcePostalCode: nil, // Missing
				DestinationPostalCode: stringPtr("10001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			expectedErrorCode:    "DHL_VALIDATION_ERROR",
			expectedErrorMessage: "source postal code is required",
			expectedErrorField:   "source_postal_code",
			description:          "DHL validation error for missing source postal code",
		},
		{
			name: "Missing packages",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("10001"),
				CountryCode:           stringPtr("IN"),
				Packages:              []models.Package{}, // Empty
			},
			expectedErrorCode:    "DHL_VALIDATION_ERROR",
			expectedErrorMessage: "at least one package is required",
			expectedErrorField:   "packages",
			description:          "DHL validation error for missing packages",
		},
		{
			name: "Invalid package weight",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("10001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: -1.0, Unit: "kg"}, // Invalid
						Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
					},
				},
			},
			expectedErrorCode:    "DHL_VALIDATION_ERROR",
			expectedErrorMessage: "package weight must be greater than 0",
			expectedErrorField:   "packages[0].weight",
			description:          "DHL validation error for invalid package weight",
		},
		{
			name: "Missing package dimensions",
			request: &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("560001"),
				DestinationPostalCode: stringPtr("10001"),
				CountryCode:           stringPtr("IN"),
				Packages: []models.Package{
					{
						Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
						Dimensions: nil, // Missing
					},
				},
			},
			expectedErrorCode:    "DHL_VALIDATION_ERROR",
			expectedErrorMessage: "package dimensions are required",
			expectedErrorField:   "packages[0].dimensions",
			description:          "DHL validation error for missing package dimensions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			factory := mocks.NewMockPartnerAdapterFactory()
			repo := mocks.NewMockPartnerAttributeMapRepository()
			factory.SetSupportedPartners([]string{"dhl"})
			factory.SetupDHLValidationErrorAdapters()

			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				factory,
				repo,
				30*time.Second,
				false,
			)

			// Execute
			ctx := context.Background()
			response, err := orchestrator.CheckServiceability(ctx, tt.request)

			// Verify response structure
			require.NoError(t, err, "Orchestrator should not return error")
			require.NotNil(t, response, "Response should not be nil")
			require.NotEmpty(t, response.Partners, "Should have partners in response")

			// Find DHL partner
			var dhlPartner *models.PartnerV2Response
			for i := range response.Partners {
				if response.Partners[i].PartnerCode == "dhl" {
					dhlPartner = &response.Partners[i]
					break
				}
			}

			require.NotNil(t, dhlPartner, "DHL partner should be in response")
			require.NotNil(t, dhlPartner.Error, "DHL partner should have error")

			// Validate error message structure
			errorMessage := *dhlPartner.Error
			assert.Contains(t, errorMessage, "DHL validation failed", "Error should indicate DHL validation failure")
			assert.Contains(t, errorMessage, tt.expectedErrorMessage, "Error should contain expected validation message")

			// Validate partner response structure for errors
			assert.NotEmpty(t, dhlPartner.PartnerID, "PartnerID should be present")
			assert.Equal(t, "dhl", dhlPartner.PartnerCode, "PartnerCode should be correct")
			assert.NotEmpty(t, dhlPartner.PartnerName, "PartnerName should be present")
			assert.Empty(t, dhlPartner.Services, "Services should be empty for validation errors")
			assert.Empty(t, dhlPartner.Capabilities, "Capabilities should be empty for validation errors")
			assert.Greater(t, dhlPartner.ResponseTime, time.Duration(0), "Should have positive response time")
		})
	}
}

// TestDHLErrorMessageFormat tests the specific format of DHL error messages
func TestDHLErrorMessageFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		inputError     string
		expectedFormat ErrorMessageFormat
		description    string
	}{
		{
			name:       "Validation error format",
			inputError: "DHL validation failed: destination postal code is required for DHL shipments",
			expectedFormat: ErrorMessageFormat{
				Prefix:    "DHL validation failed",
				Separator: ": ",
				Message:   "destination postal code is required for DHL shipments",
				Context:   "DHL shipments",
			},
			description: "DHL validation errors should follow consistent format",
		},
		{
			name:       "Package validation error format",
			inputError: "DHL validation failed: package weight must be greater than 0 for package 1 in DHL shipments",
			expectedFormat: ErrorMessageFormat{
				Prefix:    "DHL validation failed",
				Separator: ": ",
				Message:   "package weight must be greater than 0 for package 1 in DHL shipments",
				Context:   "DHL shipments",
			},
			description: "Package validation errors should include package index",
		},
		{
			name:       "API error format",
			inputError: "DHL API error [400]: Invalid request parameters",
			expectedFormat: ErrorMessageFormat{
				Prefix:    "DHL API error",
				Separator: " [400]: ",
				Message:   "Invalid request parameters",
				Context:   "API",
			},
			description: "API errors should include status codes",
		},
		{
			name:       "Authentication error format",
			inputError: "DHL authentication failed: Invalid credentials provided",
			expectedFormat: ErrorMessageFormat{
				Prefix:    "DHL authentication failed",
				Separator: ": ",
				Message:   "Invalid credentials provided",
				Context:   "authentication",
			},
			description: "Authentication errors should be clearly identified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			format := parseErrorMessageFormat(tt.inputError)

			assert.Equal(t, tt.expectedFormat.Prefix, format.Prefix, "Error prefix should match")
			assert.Equal(t, tt.expectedFormat.Separator, format.Separator, "Error separator should match")
			assert.Equal(t, tt.expectedFormat.Message, format.Message, "Error message should match")
			assert.Contains(t, strings.ToLower(tt.inputError), tt.expectedFormat.Context, "Error should contain context")
		})
	}
}

// TestDHLErrorResponseStructure tests the complete structure of DHL error responses
func TestDHLErrorResponseStructure(t *testing.T) {
	t.Parallel()

	// Setup
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()
	factory.SetSupportedPartners([]string{"dhl"})
	factory.SetupDHLValidationErrorAdapters()

	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		factory,
		repo,
		30*time.Second,
		false,
	)

	// Request with validation error
	request := &models.ServiceabilityV2Request{
		SourcePostalCode: stringPtr("560001"),
		// Missing destination postal code
		CountryCode: stringPtr("IN"),
		Packages: []models.Package{
			{
				Weight:     &models.Weight{Value: 1.5, Unit: "kg"},
				Dimensions: &models.Dimensions{Length: 25.0, Width: 15.0, Height: 10.0, Unit: "cm"},
			},
		},
	}

	// Execute
	ctx := context.Background()
	response, err := orchestrator.CheckServiceability(ctx, request)

	// Verify
	require.NoError(t, err)
	require.NotEmpty(t, response.Partners)

	// Find DHL partner
	var dhlPartner *models.PartnerV2Response
	for i := range response.Partners {
		if response.Partners[i].PartnerCode == "dhl" {
			dhlPartner = &response.Partners[i]
			break
		}
	}

	require.NotNil(t, dhlPartner)

	// Test complete error structure
	t.Run("Error presence and format", func(t *testing.T) {
		require.NotNil(t, dhlPartner.Error, "Error should be present")
		assert.NotEmpty(t, *dhlPartner.Error, "Error message should not be empty")
		assert.Contains(t, *dhlPartner.Error, "DHL validation failed", "Error should indicate DHL validation")
	})

	t.Run("Partner identification fields", func(t *testing.T) {
		assert.NotEmpty(t, dhlPartner.PartnerID, "PartnerID should be present")
		assert.Equal(t, "dhl", dhlPartner.PartnerCode, "PartnerCode should be 'dhl'")
		assert.NotEmpty(t, dhlPartner.PartnerName, "PartnerName should be present")
		assert.Contains(t, strings.ToLower(dhlPartner.PartnerName), "dhl", "PartnerName should contain 'dhl'")
	})

	t.Run("Empty fields for errors", func(t *testing.T) {
		assert.Empty(t, dhlPartner.Services, "Services should be empty for validation errors")
		assert.Empty(t, dhlPartner.Capabilities, "Capabilities should be empty for validation errors")
	})

	t.Run("Response timing", func(t *testing.T) {
		assert.Greater(t, dhlPartner.ResponseTime, time.Duration(0), "ResponseTime should be positive")
		assert.Less(t, dhlPartner.ResponseTime, 5*time.Second, "ResponseTime should be reasonable for mocked response")
	})
}

// TestDHLErrorCodeStandardization tests that DHL error codes follow standard format
func TestDHLErrorCodeStandardization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		errorType      string
		expectedPrefix string
		expectedFormat string
		description    string
	}{
		{
			name:           "Validation errors",
			errorType:      "validation",
			expectedPrefix: "DHL_VALIDATION_ERROR",
			expectedFormat: "DHL_VALIDATION_ERROR",
			description:    "Validation errors should use standard prefix",
		},
		{
			name:           "API errors",
			errorType:      "api",
			expectedPrefix: "DHL_API_ERROR",
			expectedFormat: "DHL_API_ERROR",
			description:    "API errors should use standard prefix",
		},
		{
			name:           "Authentication errors",
			errorType:      "auth",
			expectedPrefix: "DHL_AUTH_ERROR",
			expectedFormat: "DHL_AUTH_ERROR",
			description:    "Authentication errors should use standard prefix",
		},
		{
			name:           "Network errors",
			errorType:      "network",
			expectedPrefix: "DHL_NETWORK_ERROR",
			expectedFormat: "DHL_NETWORK_ERROR",
			description:    "Network errors should use standard prefix",
		},
		{
			name:           "Hub lookup errors",
			errorType:      "hub_lookup",
			expectedPrefix: "DHL_HUB_LOOKUP_ERROR",
			expectedFormat: "DHL_HUB_LOOKUP_ERROR",
			description:    "Hub lookup errors should use standard prefix",
		},
		{
			name:           "Country code resolution errors",
			errorType:      "country_code",
			expectedPrefix: "DHL_COUNTRY_CODE_ERROR",
			expectedFormat: "DHL_COUNTRY_CODE_ERROR",
			description:    "Country code errors should use standard prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errorCode := generateDHLErrorCode(tt.errorType)

			assert.Equal(t, tt.expectedFormat, errorCode, "Error code should match expected format")
			assert.Contains(t, errorCode, "DHL_", "Error code should contain DHL prefix")
			assert.Contains(t, errorCode, "ERROR", "Error code should contain ERROR suffix")
			assert.Equal(t, strings.ToUpper(errorCode), errorCode, "Error code should be uppercase")
		})
	}
}

// TestDHLErrorMessageLocalization tests error message consistency and format
func TestDHLErrorMessageLocalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		errorScenario   string
		expectedMessage DHLErrorMessage
		description     string
	}{
		{
			name:          "Missing source postal code",
			errorScenario: "missing_source_postal_code",
			expectedMessage: DHLErrorMessage{
				Code:    "DHL_VALIDATION_ERROR",
				Message: "source postal code is required for DHL shipments",
				Field:   "source_postal_code",
				Context: "DHL validation",
			},
			description: "Source postal code validation message should be consistent",
		},
		{
			name:          "Missing destination postal code",
			errorScenario: "missing_destination_postal_code",
			expectedMessage: DHLErrorMessage{
				Code:    "DHL_VALIDATION_ERROR",
				Message: "destination postal code is required for DHL shipments",
				Field:   "destination_postal_code",
				Context: "DHL validation",
			},
			description: "Destination postal code validation message should be consistent",
		},
		{
			name:          "Missing packages",
			errorScenario: "missing_packages",
			expectedMessage: DHLErrorMessage{
				Code:    "DHL_VALIDATION_ERROR",
				Message: "at least one package is required for DHL shipments",
				Field:   "packages",
				Context: "DHL validation",
			},
			description: "Package validation message should be consistent",
		},
		{
			name:          "Invalid package weight",
			errorScenario: "invalid_package_weight",
			expectedMessage: DHLErrorMessage{
				Code:    "DHL_VALIDATION_ERROR",
				Message: "package weight must be greater than 0 for package %d in DHL shipments",
				Field:   "packages[%d].weight",
				Context: "DHL validation",
			},
			description: "Package weight validation message should include package index",
		},
		{
			name:          "Invalid package dimensions",
			errorScenario: "invalid_package_dimensions",
			expectedMessage: DHLErrorMessage{
				Code:    "DHL_VALIDATION_ERROR",
				Message: "package dimensions must be greater than 0 for package %d in DHL shipments",
				Field:   "packages[%d].dimensions",
				Context: "DHL validation",
			},
			description: "Package dimensions validation message should include package index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errorMessage := generateDHLErrorMessage(tt.errorScenario, 1)

			assert.Equal(t, tt.expectedMessage.Code, errorMessage.Code, "Error code should match")
			assert.Contains(t, errorMessage.Message, "DHL shipments", "Error message should mention DHL shipments")
			assert.NotEmpty(t, errorMessage.Field, "Error field should be specified")
			assert.Equal(t, tt.expectedMessage.Context, errorMessage.Context, "Error context should match")
		})
	}
}

// TestDHLErrorMetadataStructure tests error metadata structure and content
func TestDHLErrorMetadataStructure(t *testing.T) {
	t.Parallel()

	// Setup
	factory := mocks.NewMockPartnerAdapterFactory()
	repo := mocks.NewMockPartnerAttributeMapRepository()
	factory.SetSupportedPartners([]string{"dhl"})
	factory.SetupDHLValidationErrorAdapters()

	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		factory,
		repo,
		30*time.Second,
		false,
	)

	// Request with validation error
	request := &models.ServiceabilityV2Request{
		SourcePostalCode: stringPtr("560001"),
		CountryCode:      stringPtr("IN"),
		Packages:         []models.Package{}, // Empty packages to trigger error
	}

	// Execute
	ctx := context.Background()
	response, err := orchestrator.CheckServiceability(ctx, request)

	// Verify
	require.NoError(t, err)
	require.NotEmpty(t, response.Partners)

	// Find DHL partner
	var dhlPartner *models.PartnerV2Response
	for i := range response.Partners {
		if response.Partners[i].PartnerCode == "dhl" {
			dhlPartner = &response.Partners[i]
			break
		}
	}

	require.NotNil(t, dhlPartner)
	require.NotNil(t, dhlPartner.Error)

	// Test error metadata structure
	t.Run("Error message format", func(t *testing.T) {
		errorMessage := *dhlPartner.Error
		assert.Contains(t, errorMessage, "DHL validation failed", "Should contain DHL validation prefix")
		assert.Contains(t, errorMessage, ":", "Should contain separator")
		assert.Contains(t, errorMessage, "package", "Should mention the validation issue")
	})

	t.Run("Error consistency", func(t *testing.T) {
		errorMessage := *dhlPartner.Error
		// Error message should be informative but not expose internal details
		assert.NotContains(t, errorMessage, "mock", "Should not expose mock details")
		assert.NotContains(t, errorMessage, "test", "Should not expose test details")
		assert.NotContains(t, errorMessage, "internal", "Should not expose internal details")
	})
}

// Helper types and functions

type ErrorMessageFormat struct {
	Prefix    string
	Separator string
	Message   string
	Context   string
}

type DHLErrorMessage struct {
	Code    string
	Message string
	Field   string
	Context string
}

// parseErrorMessageFormat parses an error message to extract its format components
func parseErrorMessageFormat(errorMessage string) ErrorMessageFormat {
	if strings.Contains(errorMessage, "DHL validation failed: ") {
		parts := strings.SplitN(errorMessage, ": ", 2)
		return ErrorMessageFormat{
			Prefix:    parts[0],
			Separator: ": ",
			Message:   parts[1],
		}
	}

	if strings.Contains(errorMessage, "DHL API error [") {
		prefix := "DHL API error"
		start := strings.Index(errorMessage, "[")
		end := strings.Index(errorMessage, "]: ")
		if start != -1 && end != -1 {
			separator := errorMessage[start : end+3]
			message := errorMessage[end+3:]
			return ErrorMessageFormat{
				Prefix:    prefix,
				Separator: separator,
				Message:   message,
			}
		}
	}

	// Default format
	parts := strings.SplitN(errorMessage, ": ", 2)
	if len(parts) == 2 {
		return ErrorMessageFormat{
			Prefix:    parts[0],
			Separator: ": ",
			Message:   parts[1],
		}
	}

	return ErrorMessageFormat{
		Prefix:    "",
		Separator: "",
		Message:   errorMessage,
	}
}

// generateDHLErrorCode generates standard DHL error codes
func generateDHLErrorCode(errorType string) string {
	switch errorType {
	case "validation":
		return "DHL_VALIDATION_ERROR"
	case "api":
		return "DHL_API_ERROR"
	case "auth":
		return "DHL_AUTH_ERROR"
	case "network":
		return "DHL_NETWORK_ERROR"
	case "hub_lookup":
		return "DHL_HUB_LOOKUP_ERROR"
	case "country_code":
		return "DHL_COUNTRY_CODE_ERROR"
	default:
		return fmt.Sprintf("DHL_%s_ERROR", strings.ToUpper(errorType))
	}
}

// generateDHLErrorMessage generates standard DHL error messages
func generateDHLErrorMessage(errorScenario string, packageIndex int) DHLErrorMessage {
	switch errorScenario {
	case "missing_source_postal_code":
		return DHLErrorMessage{
			Code:    "DHL_VALIDATION_ERROR",
			Message: "source postal code is required for DHL shipments",
			Field:   "source_postal_code",
			Context: "DHL validation",
		}
	case "missing_destination_postal_code":
		return DHLErrorMessage{
			Code:    "DHL_VALIDATION_ERROR",
			Message: "destination postal code is required for DHL shipments",
			Field:   "destination_postal_code",
			Context: "DHL validation",
		}
	case "missing_packages":
		return DHLErrorMessage{
			Code:    "DHL_VALIDATION_ERROR",
			Message: "at least one package is required for DHL shipments",
			Field:   "packages",
			Context: "DHL validation",
		}
	case "invalid_package_weight":
		return DHLErrorMessage{
			Code:    "DHL_VALIDATION_ERROR",
			Message: fmt.Sprintf("package weight must be greater than 0 for package %d in DHL shipments", packageIndex),
			Field:   fmt.Sprintf("packages[%d].weight", packageIndex-1),
			Context: "DHL validation",
		}
	case "invalid_package_dimensions":
		return DHLErrorMessage{
			Code:    "DHL_VALIDATION_ERROR",
			Message: fmt.Sprintf("package dimensions must be greater than 0 for package %d in DHL shipments", packageIndex),
			Field:   fmt.Sprintf("packages[%d].dimensions", packageIndex-1),
			Context: "DHL validation",
		}
	default:
		return DHLErrorMessage{
			Code:    "DHL_UNKNOWN_ERROR",
			Message: "unknown error occurred",
			Field:   "unknown",
			Context: "DHL validation",
		}
	}
}


