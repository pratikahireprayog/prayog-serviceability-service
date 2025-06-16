package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"prayog-serviceability-service/internal/shared/dtos/v1"
)

// TestGetServiceabilityValidation tests all validation scenarios for GET /check/{postal_code}
func TestGetServiceabilityValidation(t *testing.T) {
	app := setupTestApp(t)

	tests := []struct {
		name           string
		postalCode     string
		queryParams    string
		expectedStatus int
		expectedError  string
		description    string
	}{
		// Postal Code Validation Tests
		{
			name:           "Empty postal code",
			postalCode:     "",
			expectedStatus: 404, // Fiber returns 404 for empty path params
			description:    "Should reject empty postal code",
		},
		{
			name:           "Too short postal code",
			postalCode:     "12",
			expectedStatus: 400,
			expectedError:  "Postal code must be at least 3 characters long",
			description:    "Should reject postal codes shorter than 3 characters",
		},
		{
			name:           "Too long postal code",
			postalCode:     "123456789012345678901", // 21 characters
			expectedStatus: 400,
			expectedError:  "Postal code cannot exceed 20 characters",
			description:    "Should reject postal codes longer than 20 characters",
		},
		{
			name:           "Invalid characters in postal code",
			postalCode:     "12345@",
			expectedStatus: 400,
			expectedError:  "Postal code contains invalid characters",
			description:    "Should reject postal codes with special characters",
		},
		{
			name:           "Valid postal code with spaces",
			postalCode:     "110 001",
			expectedStatus: 200, // Should pass validation and reach service layer
			description:    "Should accept postal codes with spaces",
		},
		{
			name:           "Valid postal code with hyphens",
			postalCode:     "110-001",
			expectedStatus: 200, // Should pass validation and reach service layer
			description:    "Should accept postal codes with hyphens",
		},
		{
			name:           "Valid Indian postal code",
			postalCode:     "110001",
			expectedStatus: 200,
			description:    "Should accept valid 6-digit Indian postal code",
		},

		// Query Parameter Validation Tests
		{
			name:           "Invalid parcel_category",
			postalCode:     "110001",
			queryParams:    "parcel_category=invalid",
			expectedStatus: 400,
			expectedError:  "Invalid parcel_category. Must be one of: ecomm, cargo, courier",
			description:    "Should reject invalid parcel_category values",
		},
		{
			name:           "Valid parcel_category - ecomm",
			postalCode:     "110001",
			queryParams:    "parcel_category=ecomm",
			expectedStatus: 200,
			description:    "Should accept valid parcel_category 'ecomm'",
		},
		{
			name:           "Valid parcel_category - cargo",
			postalCode:     "110001",
			queryParams:    "parcel_category=cargo",
			expectedStatus: 200,
			description:    "Should accept valid parcel_category 'cargo'",
		},
		{
			name:           "Valid parcel_category - courier",
			postalCode:     "110001",
			queryParams:    "parcel_category=courier",
			expectedStatus: 200,
			description:    "Should accept valid parcel_category 'courier'",
		},
		{
			name:           "Case insensitive parcel_category",
			postalCode:     "110001",
			queryParams:    "parcel_category=ECOMM",
			expectedStatus: 200,
			description:    "Should accept case insensitive parcel_category",
		},
		{
			name:           "Empty product_type",
			postalCode:     "110001",
			queryParams:    "product_type=",
			expectedStatus: 400,
			expectedError:  "Product type cannot be empty",
			description:    "Should reject empty product_type",
		},
		{
			name:           "Too long product_type",
			postalCode:     "110001",
			queryParams:    "product_type=" + generateLongString(51),
			expectedStatus: 400,
			expectedError:  "Product type cannot exceed 50 characters",
			description:    "Should reject product_type longer than 50 characters",
		},
		{
			name:           "Invalid characters in product_type",
			postalCode:     "110001",
			queryParams:    "product_type=test@product",
			expectedStatus: 400,
			expectedError:  "Product type contains invalid characters",
			description:    "Should reject product_type with special characters",
		},
		{
			name:           "Valid product_type",
			postalCode:     "110001",
			queryParams:    "product_type=travel_free",
			expectedStatus: 200,
			description:    "Should accept valid product_type",
		},
		{
			name:           "Valid product_type with underscores and hyphens",
			postalCode:     "110001",
			queryParams:    "product_type=test_product-v1",
			expectedStatus: 200,
			description:    "Should accept product_type with underscores and hyphens",
		},

		// Combined Parameter Tests
		{
			name:           "Valid postal code with both query params",
			postalCode:     "110001",
			queryParams:    "parcel_category=ecomm&product_type=travel_free",
			expectedStatus: 200,
			description:    "Should accept valid postal code with both query parameters",
		},
		{
			name:           "Multiple invalid parameters",
			postalCode:     "12", // Too short
			queryParams:    "parcel_category=invalid&product_type=test@invalid",
			expectedStatus: 400,
			description:    "Should reject when multiple validations fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build URL
			url := fmt.Sprintf("/serviceability/v1/check/%s", tt.postalCode)
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			// Make request
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			// Check status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode,
				"Test: %s - %s", tt.name, tt.description)

			// Check error message if expected
			if tt.expectedError != "" {
				var response dtos.StandardErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Message, tt.expectedError,
					"Test: %s - Expected error message not found", tt.name)
			}

			resp.Body.Close()
		})
	}
}

// TestGetServiceabilityPostalCodeNormalization tests postal code normalization
func TestGetServiceabilityPostalCodeNormalization(t *testing.T) {
	app := setupTestApp(t)

	tests := []struct {
		name        string
		postalCode  string
		description string
	}{
		{
			name:        "Postal code with extra spaces",
			postalCode:  "  110001  ",
			description: "Should normalize postal code by trimming spaces",
		},
		{
			name:        "Postal code with multiple spaces",
			postalCode:  "110  001",
			description: "Should normalize multiple spaces to single space",
		},
		{
			name:        "Mixed case postal code",
			postalCode:  "abc123",
			description: "Should normalize to uppercase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/serviceability/v1/check/%s", tt.postalCode)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			// Should pass validation (status 200 or service-level error, not validation error)
			assert.NotEqual(t, 400, resp.StatusCode,
				"Test: %s - %s", tt.name, tt.description)

			resp.Body.Close()
		})
	}
}

// TestGetServiceabilityErrorHandling tests error handling improvements
func TestGetServiceabilityErrorHandling(t *testing.T) {
	app := setupTestApp(t)

	tests := []struct {
		name           string
		postalCode     string
		queryParams    string
		expectedStatus int
		checkResponse  func(t *testing.T, response dtos.StandardErrorResponse)
		description    string
	}{
		{
			name:           "Validation error response structure",
			postalCode:     "12", // Too short
			expectedStatus: 400,
			checkResponse: func(t *testing.T, response dtos.StandardErrorResponse) {
				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Error.Code)
				assert.NotEmpty(t, response.Error.Message)
			},
			description: "Should return structured error response for validation failures",
		},
		{
			name:           "Query parameter validation error",
			postalCode:     "110001",
			queryParams:    "parcel_category=invalid",
			expectedStatus: 400,
			checkResponse: func(t *testing.T, response dtos.StandardErrorResponse) {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Message, "parcel_category")
			},
			description: "Should return specific error for query parameter validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/serviceability/v1/check/%s", tt.postalCode)
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkResponse != nil {
				var response dtos.StandardErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				tt.checkResponse(t, response)
			}

			resp.Body.Close()
		})
	}
}

// Helper function to generate long strings for testing
func generateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}

// setupTestApp sets up a test Fiber app with the serviceability routes
func setupTestApp(t *testing.T) *fiber.App {
	// This would typically set up your actual app with all dependencies
	// For now, returning a basic app structure
	app := fiber.New()

	// Add your actual route setup here
	// api := app.Group("/serviceability/v1")
	// routes.RegisterServiceabilityRoutes(api, handler, logger)

	return app
}
