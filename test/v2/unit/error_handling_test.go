package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/services/v2/partners/common"
	serviceErrors "prayog-serviceability-service/internal/shared/errors"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStructuredErrorHandling tests how the orchestrator handles structured errors
func TestStructuredErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		request        *models.ServiceabilityV2Request
		setupMocks     func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectedError  string
		expectedType   string
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "Nil request error",
			request:        nil,
			setupMocks:     func(_ *mocks.MockPartnerAdapterFactory, _ *mocks.MockPartnerAttributeMapRepository) {},
			expectedError:  "request cannot be nil",
			expectedType:   "INVALID_REQUEST",
			expectedCode:   "INVALID_REQUEST",
			expectedStatus: 400,
		},
		{
			name: "Missing required field error",
			request: &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
				// Missing all postal code fields
			},
			setupMocks:     func(_ *mocks.MockPartnerAdapterFactory, _ *mocks.MockPartnerAttributeMapRepository) {},
			expectedError:  "postal_code or source_postal_code/destination_postal_code",
			expectedType:   "MISSING_REQUIRED_FIELD",
			expectedCode:   "MISSING_REQUIRED_FIELD",
			expectedStatus: 400,
		},
		{
			name: "Invalid request - conflicting postal codes",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				PostalCode:            stringPtr("110001"),
				SourcePostalCode:      stringPtr("110001"),
				DestinationPostalCode: stringPtr("110002"),
			},
			setupMocks:     func(_ *mocks.MockPartnerAdapterFactory, _ *mocks.MockPartnerAttributeMapRepository) {},
			expectedError:  "provide either postal_code only or source/destination postal codes, not both",
			expectedType:   "INVALID_REQUEST",
			expectedCode:   "INVALID_REQUEST",
			expectedStatus: 400,
		},
		{
			name: "Invalid request - conflicting postal code as destination",
			request: &models.ServiceabilityV2Request{
				CountryCode:           stringPtr("IN"),
				PostalCode:            stringPtr("110001"),
				SourcePostalCode:      stringPtr("110002"),
				DestinationPostalCode: stringPtr("110003"),
			},
			setupMocks:     func(_ *mocks.MockPartnerAdapterFactory, _ *mocks.MockPartnerAttributeMapRepository) {},
			expectedError:  "when using postal_code as destination with source_postal_code, do not provide destination_postal_code",
			expectedType:   "INVALID_REQUEST",
			expectedCode:   "INVALID_REQUEST",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()
			tt.setupMocks(mockFactory, mockRepo)

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				5*time.Second,
				false,
			)

			// Execute
			response, err := orchestrator.CheckServiceability(context.Background(), tt.request)

			// Verify error
			require.Error(t, err)
			require.Nil(t, response)
			assert.Contains(t, err.Error(), tt.expectedError)

			// Check if it's a structured error
			if serviceErr, ok := err.(*serviceErrors.ServiceError); ok {
				assert.Equal(t, tt.expectedCode, serviceErr.Code)
				assert.Equal(t, tt.expectedStatus, serviceErr.HTTPStatus)
			} else {
				// For wrapped errors, check the underlying error
				if unwrapped := errors.Unwrap(err); unwrapped != nil {
					if serviceErr, ok := unwrapped.(*serviceErrors.ServiceError); ok {
						assert.Equal(t, tt.expectedCode, serviceErr.Code)
						assert.Equal(t, tt.expectedStatus, serviceErr.HTTPStatus)
					}
				}
			}
		})
	}
}

// TestPartnerSpecificErrorHandling tests how the orchestrator handles partner-specific errors
func TestPartnerSpecificErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		request        *models.ServiceabilityV2Request
		setupMocks     func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectedErrors []PartnerErrorExpectation
	}{
		{
			name: "Partner not found error",
			request: &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr("110001"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"unknown_partner"})
				// Don't set adapter, so GetAdapter will return false
			},
			expectedErrors: []PartnerErrorExpectation{
				{
					PartnerCode:    "unknown_partner",
					ExpectedError:  "Partner not found",
					ExpectedCode:   "PARTNER_NOT_FOUND",
					ExpectedStatus: 404,
				},
			},
		},
		{
			name: "Partner unavailable error",
			request: &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr("110001"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"shipyaari"})
				adapter := mocks.NewMockPartnerAdapter("shipyaari")
				adapter.SetHealthy(false)
				adapter.SetServiceabilityError(serviceErrors.ErrPartnerUnavailable("shipyaari"))
				factory.SetAdapter("shipyaari", adapter)
			},
			expectedErrors: []PartnerErrorExpectation{
				{
					PartnerCode:    "shipyaari",
					ExpectedError:  "Partner service unavailable",
					ExpectedCode:   "PARTNER_UNAVAILABLE",
					ExpectedStatus: 503,
				},
			},
		},
		{
			name: "Multiple partner errors",
			request: &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr("110001"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"shipyaari", "smile_courier"})

				// Partner unavailable
				adapter1 := mocks.NewMockPartnerAdapter("shipyaari")
				adapter1.SetHealthy(false)
				adapter1.SetServiceabilityError(serviceErrors.ErrPartnerUnavailable("shipyaari"))
				factory.SetAdapter("shipyaari", adapter1)

				// Partner service error
				adapter2 := mocks.NewMockPartnerAdapter("smile_courier")
				adapter2.SetHealthy(true)
				adapter2.SetServiceabilityError(serviceErrors.ErrTimeout("check_serviceability"))
				factory.SetAdapter("smile_courier", adapter2)
			},
			expectedErrors: []PartnerErrorExpectation{
				{
					PartnerCode:    "shipyaari",
					ExpectedError:  "Partner service unavailable",
					ExpectedCode:   "PARTNER_UNAVAILABLE",
					ExpectedStatus: 503,
				},
				{
					PartnerCode:    "smile_courier",
					ExpectedError:  "Operation timed out",
					ExpectedCode:   "TIMEOUT",
					ExpectedStatus: 504,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()
			tt.setupMocks(mockFactory, mockRepo)

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				5*time.Second,
				false, // Return all partners when success=false
			)

			// Execute
			response, err := orchestrator.CheckServiceability(context.Background(), tt.request)

			// Verify response
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.False(t, response.Success) // Should be false when all partners have errors

			// Verify partner errors
			assert.Len(t, response.Partners, len(tt.expectedErrors))

			for i, expectedErr := range tt.expectedErrors {
				partner := response.Partners[i]
				assert.Equal(t, expectedErr.PartnerCode, partner.PartnerCode)
				assert.False(t, partner.IsServiceable)
				assert.NotNil(t, partner.Error)
				assert.Contains(t, *partner.Error, expectedErr.ExpectedError)
			}
		})
	}
}

// TestErrorMetadataHandling tests error metadata in responses
func TestErrorMetadataHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		request    *models.ServiceabilityV2Request
		setupMocks func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		verify     func(t *testing.T, response *models.ServiceabilityV2Response)
	}{
		{
			name: "Error metadata with no partners",
			request: &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr("110001"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{}) // No partners
			},
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.False(t, response.Success)
				assert.Empty(t, response.Partners)
				assert.NotNil(t, response.Metadata)
				assert.Equal(t, 0, response.Metadata.TotalPartners)
				assert.Equal(t, 0, response.Metadata.ServiceableCount)
			},
		},
		{
			name: "Error metadata with mixed results",
			request: &models.ServiceabilityV2Request{
				CountryCode: stringPtr("IN"),
				PostalCode:  stringPtr("110001"),
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"shipyaari", "smile_courier"})

				// One serviceable partner
				adapter1 := mocks.NewMockPartnerAdapter("shipyaari")
				adapter1.SetHealthy(true)
				adapter1.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerCode:   "shipyaari",
					IsServiceable: true,
					Services:      []models.ServiceV2{{ServiceCode: "SDD", ServiceName: "Same Day Delivery"}},
				})
				factory.SetAdapter("shipyaari", adapter1)

				// One partner with error
				adapter2 := mocks.NewMockPartnerAdapter("smile_courier")
				adapter2.SetHealthy(false)
				adapter2.SetServiceabilityError(serviceErrors.ErrPartnerUnavailable("smile_courier"))
				factory.SetAdapter("smile_courier", adapter2)
			},
			verify: func(t *testing.T, response *models.ServiceabilityV2Response) {
				assert.True(t, response.Success)    // Should be true when at least one partner is serviceable
				assert.Len(t, response.Partners, 1) // Only serviceable partner returned when success=true
				assert.NotNil(t, response.Metadata)
				assert.Equal(t, 2, response.Metadata.TotalPartners)
				assert.Equal(t, 1, response.Metadata.ServiceableCount)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()
			tt.setupMocks(mockFactory, mockRepo)

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				5*time.Second,
				false,
			)

			// Execute
			response, err := orchestrator.CheckServiceability(context.Background(), tt.request)

			// Verify
			require.NoError(t, err)
			require.NotNil(t, response)
			tt.verify(t, response)
		})
	}
}

// TestBulkErrorHandling tests error handling in bulk operations
func TestBulkErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		request       *models.BulkServiceabilityV2Request
		setupMocks    func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectedError string
	}{
		{
			name:          "Nil bulk request",
			request:       nil,
			setupMocks:    func(_ *mocks.MockPartnerAdapterFactory, _ *mocks.MockPartnerAttributeMapRepository) {},
			expectedError: "bulk request cannot be nil or empty",
		},
		{
			name: "Empty bulk request",
			request: &models.BulkServiceabilityV2Request{
				Requests: []models.ServiceabilityV2Request{},
			},
			setupMocks:    func(_ *mocks.MockPartnerAdapterFactory, _ *mocks.MockPartnerAttributeMapRepository) {},
			expectedError: "bulk request cannot be nil or empty",
		},
		{
			name: "Bulk request with individual errors",
			request: &models.BulkServiceabilityV2Request{
				Requests: []models.ServiceabilityV2Request{
					{
						CountryCode: stringPtr("IN"),
						PostalCode:  stringPtr("110001"),
					},
					{
						CountryCode: stringPtr("IN"),
						// Missing postal code - should cause error
					},
				},
			},
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"shipyaari"})
				adapter := mocks.NewMockPartnerAdapter("shipyaari")
				adapter.SetHealthy(true)
				adapter.SetServiceabilityResult(&common.PartnerServiceabilityResult{
					PartnerCode:   "shipyaari",
					IsServiceable: true,
					Services:      []models.ServiceV2{{ServiceCode: "SDD", ServiceName: "Same Day Delivery"}},
				})
				factory.SetAdapter("shipyaari", adapter)
			},
			expectedError: "", // No error at bulk level, but individual errors
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()
			tt.setupMocks(mockFactory, mockRepo)

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				5*time.Second,
				false,
			)

			// Execute
			response, err := orchestrator.BulkCheckServiceability(context.Background(), tt.request)

			if tt.expectedError != "" {
				// Verify error
				require.Error(t, err)
				require.Nil(t, response)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				// Verify response for mixed results
				require.NoError(t, err)
				require.NotNil(t, response)

				if tt.name == "Bulk request with individual errors" {
					assert.False(t, response.Success) // Should be false when there are errors
					assert.Len(t, response.Data, 2)

					// First request should succeed
					assert.True(t, response.Data[0].Success)
					assert.Len(t, response.Data[0].Partners, 1)

					// Second request should fail
					assert.False(t, response.Data[1].Success)
					assert.NotNil(t, response.Data[1].Error)
					assert.Contains(t, response.Data[1].Error.Message, "Request 2 failed")
				}
			}
		})
	}
}

// TestErrorTypeChecking tests error type checking functions
func TestErrorTypeChecking(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		error          error
		isNotFound     bool
		isValidation   bool
		isInternal     bool
		isUnavailable  bool
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "PostalCodeNotFound error",
			error:          serviceErrors.ErrPostalCodeNotFound("110001"),
			isNotFound:     true,
			isValidation:   false,
			isInternal:     false,
			isUnavailable:  false,
			expectedStatus: 404,
			expectedCode:   "POSTAL_CODE_NOT_FOUND",
		},
		{
			name:           "PartnerNotFound error",
			error:          serviceErrors.ErrPartnerNotFound("unknown_partner"),
			isNotFound:     true,
			isValidation:   false,
			isInternal:     false,
			isUnavailable:  false,
			expectedStatus: 404,
			expectedCode:   "PARTNER_NOT_FOUND",
		},
		{
			name:           "InvalidRequest error",
			error:          serviceErrors.ErrInvalidRequest("test message"),
			isNotFound:     false,
			isValidation:   true,
			isInternal:     false,
			isUnavailable:  false,
			expectedStatus: 400,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name:           "MissingRequiredField error",
			error:          serviceErrors.ErrMissingRequiredField("postal_code"),
			isNotFound:     false,
			isValidation:   true,
			isInternal:     false,
			isUnavailable:  false,
			expectedStatus: 400,
			expectedCode:   "MISSING_REQUIRED_FIELD",
		},
		{
			name:           "PartnerUnavailable error",
			error:          serviceErrors.ErrPartnerUnavailable("shipyaari"),
			isNotFound:     false,
			isValidation:   false,
			isInternal:     false,
			isUnavailable:  true,
			expectedStatus: 503,
			expectedCode:   "PARTNER_UNAVAILABLE",
		},
		{
			name:           "InternalError error",
			error:          serviceErrors.ErrInternalError("system failure", nil),
			isNotFound:     false,
			isValidation:   false,
			isInternal:     true,
			isUnavailable:  false,
			expectedStatus: 500,
			expectedCode:   "INTERNAL_ERROR",
		},
		{
			name:           "Timeout error",
			error:          serviceErrors.ErrTimeout("operation"),
			isNotFound:     false,
			isValidation:   false,
			isInternal:     false,
			isUnavailable:  false,
			expectedStatus: 504,
			expectedCode:   "TIMEOUT",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test type checking functions
			assert.Equal(t, tt.isNotFound, serviceErrors.IsNotFound(tt.error))
			assert.Equal(t, tt.isValidation, serviceErrors.IsValidationError(tt.error))
			assert.Equal(t, tt.isInternal, serviceErrors.IsInternalError(tt.error))
			assert.Equal(t, tt.isUnavailable, serviceErrors.IsServiceUnavailable(tt.error))

			// Test HTTP status
			assert.Equal(t, tt.expectedStatus, serviceErrors.GetHTTPStatus(tt.error))

			// Test error code
			assert.Equal(t, tt.expectedCode, serviceErrors.GetErrorCode(tt.error))
		})
	}
}

// Helper types and functions

type PartnerErrorExpectation struct {
	PartnerCode    string
	ExpectedError  string
	ExpectedCode   string
	ExpectedStatus int
}

// stringPtr helper removed - using shared function from other test files
