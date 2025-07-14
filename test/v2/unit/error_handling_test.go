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

// TestReturnOnlyServiceableErrorStructure tests error response structure for returnOnlyServiceable=true
func TestReturnOnlyServiceableErrorStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		returnOnlyServiceable bool
		setupMocks            func(*mocks.MockPartnerAdapterFactory, *mocks.MockPartnerAttributeMapRepository)
		expectedErrorCode     string
		expectedErrorMessage  string
		expectedSuccessValue  bool
		expectedPartnersCount int
		description           string
	}{
		{
			name:                  "POSTAL_CODE_NOT_FOUND_structure_validation",
			returnOnlyServiceable: true,
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})

				// DHL adapter with postal code not found error
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityError(errors.New("Failed to determine destination country: country code not found for postal code 266001: POSTAL_CODE_NOT_FOUND: Postal code not found (Postal code '266001' does not exist in our database)"))
				factory.SetAdapter("dhl", dhlAdapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedErrorCode:     "POSTAL_CODE_NOT_FOUND",
			expectedErrorMessage:  "Failed to determine destination country: country code not found for postal code 266001: POSTAL_CODE_NOT_FOUND: Postal code not found (Postal code '266001' does not exist in our database)",
			expectedSuccessValue:  false,
			expectedPartnersCount: 0,
			description:           "POSTAL_CODE_NOT_FOUND should have correct structure with specific error message",
		},
		{
			name:                  "NO_SERVICEABLE_PARTNERS_structure_validation",
			returnOnlyServiceable: true,
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl", "shipyaari"})

				// Both partners non-serviceable (no errors)
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityResult(dhlAdapter.CreateCleanNonServiceableResult())
				factory.SetAdapter("dhl", dhlAdapter)

				shipyaariAdapter := mocks.NewMockPartnerAdapter("shipyaari")
				shipyaariAdapter.SetServiceabilityResult(shipyaariAdapter.CreateCleanNonServiceableResult())
				factory.SetAdapter("shipyaari", shipyaariAdapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedErrorCode:     "NO_SERVICEABLE_PARTNERS",
			expectedErrorMessage:  "No serviceable partners found for the given request",
			expectedSuccessValue:  false,
			expectedPartnersCount: 0,
			description:           "NO_SERVICEABLE_PARTNERS should have correct structure with generic message",
		},
		{
			name:                  "Error_structure_consistency_check",
			returnOnlyServiceable: true,
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})

				// DHL adapter with embedded error message
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				result := dhlAdapter.CreateNonServiceableResult() // This has embedded error message
				dhlAdapter.SetServiceabilityResult(result)
				factory.SetAdapter("dhl", dhlAdapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedErrorCode:     "PARTNER_ERROR",
			expectedErrorMessage:  "Not serviceable in this location",
			expectedSuccessValue:  false,
			expectedPartnersCount: 0,
			description:           "Embedded error messages should also trigger PARTNER_ERROR structure",
		},
		{
			name:                  "VALIDATION_ERROR_structure_validation",
			returnOnlyServiceable: true,
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})

				// DHL adapter with validation error
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityError(errors.New("VALIDATION_ERROR: Invalid request format"))
				factory.SetAdapter("dhl", dhlAdapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedErrorCode:     "VALIDATION_ERROR",
			expectedErrorMessage:  "VALIDATION_ERROR: Invalid request format",
			expectedSuccessValue:  false,
			expectedPartnersCount: 0,
			description:           "VALIDATION_ERROR should be properly classified",
		},
		{
			name:                  "DATABASE_ERROR_structure_validation",
			returnOnlyServiceable: true,
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})

				// DHL adapter with database error
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityError(errors.New("DATABASE_ERROR: Connection failed"))
				factory.SetAdapter("dhl", dhlAdapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedErrorCode:     "DATABASE_ERROR",
			expectedErrorMessage:  "DATABASE_ERROR: Connection failed",
			expectedSuccessValue:  false,
			expectedPartnersCount: 0,
			description:           "DATABASE_ERROR should be properly classified",
		},
		{
			name:                  "TIMEOUT_ERROR_structure_validation",
			returnOnlyServiceable: true,
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})

				// DHL adapter with timeout error
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityError(errors.New("Request timeout after 30 seconds"))
				factory.SetAdapter("dhl", dhlAdapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedErrorCode:     "TIMEOUT_ERROR",
			expectedErrorMessage:  "Request timeout after 30 seconds",
			expectedSuccessValue:  false,
			expectedPartnersCount: 0,
			description:           "TIMEOUT_ERROR should be properly classified",
		},
		{
			name:                  "ACTUAL_PARTNER_ERROR_structure_validation",
			returnOnlyServiceable: true,
			setupMocks: func(factory *mocks.MockPartnerAdapterFactory, repo *mocks.MockPartnerAttributeMapRepository) {
				factory.SetSupportedPartners([]string{"dhl"})

				// DHL adapter with actual partner API error
				dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
				dhlAdapter.SetServiceabilityError(errors.New("DHL API returned HTTP 503: Service temporarily unavailable"))
				factory.SetAdapter("dhl", dhlAdapter)

				repo.SetGetPartnerInfoByAttributeError(errors.New("not filtered"))
			},
			expectedErrorCode:     "SERVICE_UNAVAILABLE",
			expectedErrorMessage:  "DHL API returned HTTP 503: Service temporarily unavailable",
			expectedSuccessValue:  false,
			expectedPartnersCount: 0,
			description:           "SERVICE_UNAVAILABLE should be properly classified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockFactory := mocks.NewMockPartnerAdapterFactory()
			mockRepo := mocks.NewMockPartnerAttributeMapRepository()
			tt.setupMocks(mockFactory, mockRepo)

			// Create orchestrator with returnOnlyServiceable setting
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				5*time.Second,
				tt.returnOnlyServiceable,
			)

			// Create test request
			request := &models.ServiceabilityV2Request{
				PostalCode:     stringPtr("266001"),
				CountryCode:    stringPtr("IN"),
				ParcelCategory: stringPtr("international"),
			}

			// Execute
			response, err := orchestrator.CheckServiceability(context.Background(), request)

			// Verify no system error
			require.NoError(t, err, "Should not return system error")
			require.NotNil(t, response, "Response should not be nil")

			// Verify response structure consistency
			assert.Equal(t, tt.expectedSuccessValue, response.Success, "Success value should match expected")
			assert.Len(t, response.Partners, tt.expectedPartnersCount, "Partners count should match expected")

			// Verify error structure exists and is correct
			require.NotNil(t, response.Error, "Error should be present")
			assert.Equal(t, tt.expectedErrorCode, response.Error.Code, "Error code should match expected")
			assert.Equal(t, tt.expectedErrorMessage, response.Error.Message, "Error message should match expected")

			// Verify metadata is always present
			require.NotNil(t, response.Metadata, "Metadata should always be present")
			assert.GreaterOrEqual(t, response.Metadata.TotalPartners, 1, "Total partners should be reported")
			assert.Equal(t, 0, response.Metadata.ServiceableCount, "Serviceable count should be 0 for error scenarios")

			// Verify filters are preserved
			assert.Equal(t, "international", *response.Metadata.Filters.ParcelCategory, "Parcel category should be preserved")
			assert.Equal(t, "IN", *response.Metadata.Filters.CountryCode, "Country code should be preserved")

			// Verify JSON serialization would work (all fields are properly typed)
			assert.IsType(t, false, response.Success, "Success should be boolean")
			assert.IsType(t, []models.PartnerV2Response{}, response.Partners, "Partners should be array")
			assert.IsType(t, &models.ErrorResponse{}, response.Error, "Error should be pointer to ErrorResponse")
			assert.IsType(t, &models.V2ResponseMetadata{}, response.Metadata, "Metadata should be pointer to V2ResponseMetadata")

			t.Logf("✅ %s: Error structure validated", tt.description)
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
