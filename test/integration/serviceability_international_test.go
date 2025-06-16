package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"prayog-serviceability-service/internal/shared/dtos/v1"
)

// TestGetServiceabilityInternationalSupport tests international postal code support for GET /check/{postal_code}
func TestGetServiceabilityInternationalSupport(t *testing.T) {
	app := setupTestApp(t)

	tests := []struct {
		name           string
		postalCode     string
		expectedStatus int
		description    string
		shouldPass     bool
	}{
		// International Postal Code Format Tests
		{
			name:           "US ZIP Code - 5 digits",
			postalCode:     "90210",
			expectedStatus: 200, // Should pass validation and reach service layer
			description:    "Should accept US 5-digit ZIP codes",
			shouldPass:     true,
		},
		{
			name:           "US ZIP+4 Code",
			postalCode:     "90210-1234",
			expectedStatus: 200,
			description:    "Should accept US ZIP+4 codes with hyphen",
			shouldPass:     true,
		},
		{
			name:           "UK Postal Code",
			postalCode:     "SW1A 1AA",
			expectedStatus: 200,
			description:    "Should accept UK postal codes with space",
			shouldPass:     true,
		},
		{
			name:           "Canadian Postal Code",
			postalCode:     "K1A 0A6",
			expectedStatus: 200,
			description:    "Should accept Canadian postal codes",
			shouldPass:     true,
		},
		{
			name:           "German Postal Code",
			postalCode:     "10115",
			expectedStatus: 200,
			description:    "Should accept German 5-digit postal codes",
			shouldPass:     true,
		},
		{
			name:           "French Postal Code",
			postalCode:     "75001",
			expectedStatus: 200,
			description:    "Should accept French 5-digit postal codes",
			shouldPass:     true,
		},
		{
			name:           "Australian Postal Code",
			postalCode:     "2000",
			expectedStatus: 200,
			description:    "Should accept Australian 4-digit postal codes",
			shouldPass:     true,
		},
		{
			name:           "Indian Postal Code",
			postalCode:     "110001",
			expectedStatus: 200,
			description:    "Should accept Indian 6-digit postal codes",
			shouldPass:     true,
		},
		{
			name:           "Brazilian CEP",
			postalCode:     "01310-100",
			expectedStatus: 200,
			description:    "Should accept Brazilian CEP with hyphen",
			shouldPass:     true,
		},
		{
			name:           "Japanese Postal Code",
			postalCode:     "100-0001",
			expectedStatus: 200,
			description:    "Should accept Japanese postal codes with hyphen",
			shouldPass:     true,
		},

		// Edge Cases That Should Still Pass
		{
			name:           "Mixed case postal code",
			postalCode:     "sw1a 1aa",
			expectedStatus: 200,
			description:    "Should accept mixed case postal codes (normalization)",
			shouldPass:     true,
		},
		{
			name:           "Postal code with extra spaces",
			postalCode:     " 90210 ",
			expectedStatus: 200,
			description:    "Should handle postal codes with extra spaces",
			shouldPass:     true,
		},
		{
			name:           "Alphanumeric postal code",
			postalCode:     "M5V3A8",
			expectedStatus: 200,
			description:    "Should accept alphanumeric postal codes",
			shouldPass:     true,
		},

		// Boundary Cases That Should Still Fail
		{
			name:           "Too short postal code",
			postalCode:     "12",
			expectedStatus: 400,
			description:    "Should reject postal codes shorter than 3 characters",
			shouldPass:     false,
		},
		{
			name:           "Too long postal code",
			postalCode:     "123456789012345678901", // 21 characters
			expectedStatus: 400,
			description:    "Should reject postal codes longer than 20 characters",
			shouldPass:     false,
		},
		{
			name:           "Invalid characters",
			postalCode:     "12345@",
			expectedStatus: 400,
			description:    "Should reject postal codes with invalid characters",
			shouldPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build URL
			url := fmt.Sprintf("/serviceability/v1/check/%s", tt.postalCode)

			// Make request
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			// Check status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode,
				"Test: %s - %s", tt.name, tt.description)

			// For tests that should pass validation
			if tt.shouldPass {
				// Should reach service layer (200 response)
				// The actual serviceability result depends on database content
				assert.Equal(t, 200, resp.StatusCode,
					"Test: %s - Should pass validation and reach service layer", tt.name)
			} else {
				// Should fail validation (400 response)
				assert.Equal(t, 400, resp.StatusCode,
					"Test: %s - Should fail validation", tt.name)

				// Check error response structure
				var response dtos.StandardErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.NotEmpty(t, response.Error.Message)
			}

			resp.Body.Close()
		})
	}
}

// TestGetServiceabilityInternationalWithFilters tests international postal codes with query parameters
func TestGetServiceabilityInternationalWithFilters(t *testing.T) {
	app := setupTestApp(t)

	tests := []struct {
		name           string
		postalCode     string
		queryParams    string
		expectedStatus int
		description    string
	}{
		{
			name:           "US ZIP with parcel category",
			postalCode:     "90210",
			queryParams:    "parcel_category=ecomm",
			expectedStatus: 200,
			description:    "Should accept US postal code with valid parcel category",
		},
		{
			name:           "UK postal code with product type",
			postalCode:     "SW1A 1AA",
			queryParams:    "product_type=express_delivery",
			expectedStatus: 200,
			description:    "Should accept UK postal code with valid product type",
		},
		{
			name:           "Canadian postal code with both filters",
			postalCode:     "K1A 0A6",
			queryParams:    "parcel_category=courier&product_type=standard",
			expectedStatus: 200,
			description:    "Should accept Canadian postal code with multiple filters",
		},
		{
			name:           "International postal code with invalid category",
			postalCode:     "10115",
			queryParams:    "parcel_category=invalid",
			expectedStatus: 400,
			description:    "Should reject invalid parcel category even with valid international postal code",
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

			resp.Body.Close()
		})
	}
}
