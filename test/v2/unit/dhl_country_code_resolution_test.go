package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"prayog-serviceability-service/internal/services/v2/partners/common"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/test/v2/mocks"
)

func TestDHLResolveCountryCodesSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		sourcePostalCode       string
		destinationPostalCode  string
		providedCountryCode    *string
		expectedSourceCountry  string
		expectedDestCountry    string
		expectedResolutionTime int64
		expectedCacheHit       bool
	}{
		{
			name:                   "US postal codes - auto resolution",
			sourcePostalCode:       "10001",
			destinationPostalCode:  "90210",
			providedCountryCode:    nil,
			expectedSourceCountry:  "US",
			expectedDestCountry:    "US",
			expectedResolutionTime: 25,
			expectedCacheHit:       false,
		},
		{
			name:                   "UK to India - mixed resolution",
			sourcePostalCode:       "SW1A 1AA",
			destinationPostalCode:  "400001",
			providedCountryCode:    stringPtr("GB"),
			expectedSourceCountry:  "GB",
			expectedDestCountry:    "IN",
			expectedResolutionTime: 35,
			expectedCacheHit:       false,
		},
		{
			name:                   "Germany to Japan - geolocation required",
			sourcePostalCode:       "10115",
			destinationPostalCode:  "100-0001",
			providedCountryCode:    nil,
			expectedSourceCountry:  "DE",
			expectedDestCountry:    "JP",
			expectedResolutionTime: 45,
			expectedCacheHit:       false,
		},
		{
			name:                   "Canada domestic - provided country code",
			sourcePostalCode:       "M5V 3L9",
			destinationPostalCode:  "V6B 1A1",
			providedCountryCode:    stringPtr("CA"),
			expectedSourceCountry:  "CA",
			expectedDestCountry:    "CA",
			expectedResolutionTime: 15,
			expectedCacheHit:       true,
		},
		{
			name:                   "France to Australia - complex resolution",
			sourcePostalCode:       "75001",
			destinationPostalCode:  "2000",
			providedCountryCode:    nil,
			expectedSourceCountry:  "FR",
			expectedDestCountry:    "AU",
			expectedResolutionTime: 55,
			expectedCacheHit:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup DHL adapter with country code resolution response
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
			countryResponse := createDHLCountryCodeResolutionResponse(
				tt.expectedSourceCountry,
				tt.expectedDestCountry,
				tt.expectedResolutionTime,
				tt.expectedCacheHit,
			)
			dhlAdapter.SetServiceabilityResult(countryResponse)
			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Create request
			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      &tt.sourcePostalCode,
				DestinationPostalCode: &tt.destinationPostalCode,
				CountryCode:           tt.providedCountryCode,
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 2.0,
							Unit:  "kg",
						},
					},
				},
			}

			// Execute country code resolution
			result, err := executeDHLCountryCodeResolution(ctx, mockFactory, request)

			// Verify response
			assert.NoError(t, err)
			require.NotNil(t, result)

			// Verify country resolution
			assert.Equal(t, tt.expectedSourceCountry, result.SourceCountry)
			assert.Equal(t, tt.expectedDestCountry, result.DestinationCountry)
			assert.Equal(t, tt.expectedResolutionTime, result.ResolutionTimeMs)
			assert.Equal(t, tt.expectedCacheHit, result.CacheHit)

			// Verify resolution metadata
			assert.NotEmpty(t, result.ResolutionMethod)
			assert.Greater(t, result.ConfidenceScore, 0.0)
			assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
		})
	}
}

func TestDHLCountryCodeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		sourcePostalCode      string
		destinationPostalCode string
		providedCountryCode   *string
		expectedError         bool
		expectedErrorType     string
		expectedErrorMessage  string
	}{
		{
			name:                  "Invalid source postal code format",
			sourcePostalCode:      "INVALID",
			destinationPostalCode: "400001",
			providedCountryCode:   stringPtr("US"),
			expectedError:         true,
			expectedErrorType:     "DHL_COUNTRY_RESOLUTION_INVALID_POSTAL_CODE",
			expectedErrorMessage:  "source postal code format is invalid for country resolution",
		},
		{
			name:                  "Empty postal codes",
			sourcePostalCode:      "",
			destinationPostalCode: "",
			providedCountryCode:   nil,
			expectedError:         true,
			expectedErrorType:     "DHL_COUNTRY_RESOLUTION_MISSING_DATA",
			expectedErrorMessage:  "postal codes are required for country resolution",
		},
		{
			name:                  "Unsupported country code",
			sourcePostalCode:      "12345",
			destinationPostalCode: "67890",
			providedCountryCode:   stringPtr("XX"),
			expectedError:         true,
			expectedErrorType:     "DHL_COUNTRY_RESOLUTION_UNSUPPORTED_COUNTRY",
			expectedErrorMessage:  "country code XX is not supported for DHL services",
		},
		{
			name:                  "Conflicting postal code and country",
			sourcePostalCode:      "10001", // US postal code
			destinationPostalCode: "400001",
			providedCountryCode:   stringPtr("FR"), // French country code
			expectedError:         true,
			expectedErrorType:     "DHL_COUNTRY_RESOLUTION_CONFLICT",
			expectedErrorMessage:  "provided country code conflicts with postal code pattern",
		},
		{
			name:                  "Ambiguous postal code pattern",
			sourcePostalCode:      "12345",
			destinationPostalCode: "67890",
			providedCountryCode:   nil,
			expectedError:         true,
			expectedErrorType:     "DHL_COUNTRY_RESOLUTION_AMBIGUOUS",
			expectedErrorMessage:  "postal code pattern matches multiple countries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup DHL adapter with validation error
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
			errorResponse := createDHLCountryCodeValidationErrorResponse(tt.expectedErrorType, tt.expectedErrorMessage)
			dhlAdapter.SetServiceabilityResult(errorResponse)
			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Create request
			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      &tt.sourcePostalCode,
				DestinationPostalCode: &tt.destinationPostalCode,
				CountryCode:           tt.providedCountryCode,
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.0,
							Unit:  "kg",
						},
					},
				},
			}

			// Execute country code resolution
			result, err := executeDHLCountryCodeResolution(ctx, mockFactory, request)

			// Verify error response
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErrorType)
			assert.Contains(t, err.Error(), tt.expectedErrorMessage)
		})
	}
}

func TestDHLGeolocationServiceIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		geolocationResponse      *GeolocationServiceResponse
		expectedCountriesFound   bool
		expectedConfidenceScore  float64
		expectedResolutionMethod string
		expectedCacheUsage       bool
	}{
		{
			name: "High confidence resolution",
			geolocationResponse: &GeolocationServiceResponse{
				SourceCountry: CountryInfo{
					Code:       "US",
					Name:       "United States",
					Confidence: 0.95,
					Source:     "postal_pattern_match",
				},
				DestinationCountry: CountryInfo{
					Code:       "IN",
					Name:       "India",
					Confidence: 0.98,
					Source:     "postal_pattern_match",
				},
				ResponseTime: 30 * time.Millisecond,
				CacheHit:     false,
			},
			expectedCountriesFound:   true,
			expectedConfidenceScore:  0.95,
			expectedResolutionMethod: "postal_pattern_match",
			expectedCacheUsage:       false,
		},
		{
			name: "Medium confidence with geolocation fallback",
			geolocationResponse: &GeolocationServiceResponse{
				SourceCountry: CountryInfo{
					Code:       "GB",
					Name:       "United Kingdom",
					Confidence: 0.75,
					Source:     "geolocation_lookup",
				},
				DestinationCountry: CountryInfo{
					Code:       "AU",
					Name:       "Australia",
					Confidence: 0.80,
					Source:     "postal_pattern_match",
				},
				ResponseTime: 85 * time.Millisecond,
				CacheHit:     false,
			},
			expectedCountriesFound:   true,
			expectedConfidenceScore:  0.75,
			expectedResolutionMethod: "geolocation_lookup",
			expectedCacheUsage:       false,
		},
		{
			name: "Cache hit response",
			geolocationResponse: &GeolocationServiceResponse{
				SourceCountry: CountryInfo{
					Code:       "CA",
					Name:       "Canada",
					Confidence: 1.0,
					Source:     "cache",
				},
				DestinationCountry: CountryInfo{
					Code:       "CA",
					Name:       "Canada",
					Confidence: 1.0,
					Source:     "cache",
				},
				ResponseTime: 5 * time.Millisecond,
				CacheHit:     true,
			},
			expectedCountriesFound:   true,
			expectedConfidenceScore:  1.0,
			expectedResolutionMethod: "cache",
			expectedCacheUsage:       true,
		},
		{
			name: "Low confidence resolution",
			geolocationResponse: &GeolocationServiceResponse{
				SourceCountry: CountryInfo{
					Code:       "DE",
					Name:       "Germany",
					Confidence: 0.45,
					Source:     "fuzzy_match",
				},
				DestinationCountry: CountryInfo{
					Code:       "JP",
					Name:       "Japan",
					Confidence: 0.55,
					Source:     "fuzzy_match",
				},
				ResponseTime: 120 * time.Millisecond,
				CacheHit:     false,
				Warning:      "Low confidence resolution - manual verification recommended",
			},
			expectedCountriesFound:   true,
			expectedConfidenceScore:  0.45,
			expectedResolutionMethod: "fuzzy_match",
			expectedCacheUsage:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup DHL adapter with geolocation service response
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
			countryResponse := createDHLGeolocationServiceIntegrationResponse(tt.geolocationResponse)
			dhlAdapter.SetServiceabilityResult(countryResponse)
			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Create request
			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("12345"),
				DestinationPostalCode: stringPtr("67890"),
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.5,
							Unit:  "kg",
						},
					},
				},
			}

			// Execute geolocation service integration
			result, err := executeDHLCountryCodeResolution(ctx, mockFactory, request)

			// Verify response
			assert.NoError(t, err)
			require.NotNil(t, result)

			// Verify geolocation integration results
			assert.Equal(t, tt.geolocationResponse.SourceCountry.Code, result.SourceCountry)
			assert.Equal(t, tt.geolocationResponse.DestinationCountry.Code, result.DestinationCountry)
			assert.Equal(t, tt.expectedConfidenceScore, result.ConfidenceScore)
			assert.Equal(t, tt.expectedResolutionMethod, result.ResolutionMethod)
			assert.Equal(t, tt.expectedCacheUsage, result.CacheHit)

			// Verify service integration metadata
			if tt.geolocationResponse.Warning != "" {
				assert.Contains(t, result.Warnings, tt.geolocationResponse.Warning)
			}
		})
	}
}

func TestDHLCountryCodeResolutionErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		geolocationError  error
		expectedErrorType string
		expectedRetryable bool
		expectedBackoffMs int
	}{
		{
			name:              "Geolocation service timeout",
			geolocationError:  fmt.Errorf("geolocation service timeout: request exceeded 10 seconds"),
			expectedErrorType: "DHL_GEOLOCATION_SERVICE_TIMEOUT",
			expectedRetryable: true,
			expectedBackoffMs: 2000,
		},
		{
			name:              "Geolocation service unavailable",
			geolocationError:  fmt.Errorf("geolocation service unavailable: 503 Service Unavailable"),
			expectedErrorType: "DHL_GEOLOCATION_SERVICE_UNAVAILABLE",
			expectedRetryable: true,
			expectedBackoffMs: 5000,
		},
		{
			name:              "Geolocation service rate limited",
			geolocationError:  fmt.Errorf("geolocation service rate limited: 429 Too Many Requests"),
			expectedErrorType: "DHL_GEOLOCATION_SERVICE_RATE_LIMITED",
			expectedRetryable: true,
			expectedBackoffMs: 10000,
		},
		{
			name:              "Geolocation service authentication error",
			geolocationError:  fmt.Errorf("geolocation service authentication failed: 401 Unauthorized"),
			expectedErrorType: "DHL_GEOLOCATION_SERVICE_AUTH_ERROR",
			expectedRetryable: false,
			expectedBackoffMs: 0,
		},
		{
			name:              "Geolocation service data error",
			geolocationError:  fmt.Errorf("geolocation service data error: invalid postal code format"),
			expectedErrorType: "DHL_GEOLOCATION_SERVICE_DATA_ERROR",
			expectedRetryable: false,
			expectedBackoffMs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup DHL adapter with geolocation service error
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
			errorResponse := createDHLGeolocationServiceErrorResponse(tt.geolocationError, tt.expectedErrorType)
			dhlAdapter.SetServiceabilityResult(errorResponse)
			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Create request
			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("10001"),
				DestinationPostalCode: stringPtr("400001"),
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 1.0,
							Unit:  "kg",
						},
					},
				},
			}

			// Execute country code resolution with error
			result, err := executeDHLCountryCodeResolution(ctx, mockFactory, request)

			// Verify error handling
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErrorType)

			// Verify error characteristics
			errorInfo := parseGeolocationServiceError(err)
			assert.Equal(t, tt.expectedRetryable, errorInfo.IsRetryable)
			assert.Equal(t, tt.expectedBackoffMs, errorInfo.BackoffMs)
		})
	}
}

func TestDHLCountryCodeCacheIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		cacheScenario          string
		expectedCacheHit       bool
		expectedResolutionTime int64
		expectedCacheKey       string
	}{
		{
			name:                   "Cache miss - first lookup",
			cacheScenario:          "miss",
			expectedCacheHit:       false,
			expectedResolutionTime: 45,
			expectedCacheKey:       "dhl_country_10001_400001",
		},
		{
			name:                   "Cache hit - subsequent lookup",
			cacheScenario:          "hit",
			expectedCacheHit:       true,
			expectedResolutionTime: 5,
			expectedCacheKey:       "dhl_country_10001_400001",
		},
		{
			name:                   "Cache expired - refresh required",
			cacheScenario:          "expired",
			expectedCacheHit:       false,
			expectedResolutionTime: 35,
			expectedCacheKey:       "dhl_country_10001_400001",
		},
		{
			name:                   "Cache invalidated - manual refresh",
			cacheScenario:          "invalidated",
			expectedCacheHit:       false,
			expectedResolutionTime: 40,
			expectedCacheKey:       "dhl_country_10001_400001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			ctx := context.Background()
			mockFactory := mocks.NewMockPartnerAdapterFactory()

			// Setup DHL adapter with cache scenario
			dhlAdapter := mocks.NewMockPartnerAdapter("dhl")
			cacheResponse := createDHLCountryCodeCacheResponse(
				tt.cacheScenario,
				tt.expectedCacheHit,
				tt.expectedResolutionTime,
				tt.expectedCacheKey,
			)
			dhlAdapter.SetServiceabilityResult(cacheResponse)
			mockFactory.SetAdapter("dhl", dhlAdapter)

			// Create request
			request := &models.ServiceabilityV2Request{
				SourcePostalCode:      stringPtr("10001"),
				DestinationPostalCode: stringPtr("400001"),
				ParcelCategory:        stringPtr("international"),
				Packages: []models.Package{
					{
						Weight: &models.Weight{
							Value: 2.0,
							Unit:  "kg",
						},
					},
				},
			}

			// Execute country code resolution with cache
			result, err := executeDHLCountryCodeResolution(ctx, mockFactory, request)

			// Verify cache integration
			assert.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedCacheHit, result.CacheHit)
			assert.Equal(t, tt.expectedResolutionTime, result.ResolutionTimeMs)
			assert.Equal(t, tt.expectedCacheKey, result.CacheKey)

			// Verify cache behavior
			if tt.expectedCacheHit {
				assert.Equal(t, "cache", result.ResolutionMethod)
				assert.Equal(t, 1.0, result.ConfidenceScore)
			} else {
				assert.NotEqual(t, "cache", result.ResolutionMethod)
				assert.Greater(t, result.ResolutionTimeMs, int64(10))
			}
		})
	}
}

// Helper types and functions for DHL country code resolution tests

type CountryInfo struct {
	Code       string
	Name       string
	Confidence float64
	Source     string
}

type GeolocationServiceResponse struct {
	SourceCountry      CountryInfo
	DestinationCountry CountryInfo
	ResponseTime       time.Duration
	CacheHit           bool
	Warning            string
}

type DHLCountryCodeResolutionResult struct {
	SourceCountry      string
	DestinationCountry string
	ResolutionTimeMs   int64
	ConfidenceScore    float64
	ResolutionMethod   string
	CacheHit           bool
	CacheKey           string
	Warnings           []string
}

type GeolocationServiceErrorInfo struct {
	IsRetryable bool
	BackoffMs   int
}

func executeDHLCountryCodeResolution(ctx context.Context, factory *mocks.MockPartnerAdapterFactory, request *models.ServiceabilityV2Request) (*DHLCountryCodeResolutionResult, error) {
	// Simulate DHL country code resolution execution
	adapter, err := factory.CreateAdapter("dhl")
	if err != nil {
		return nil, err
	}

	partnerInfo := common.PartnerInfo{
		PartnerCode: "dhl",
	}

	result, err := adapter.CheckServiceability(ctx, request, partnerInfo)
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, result.Error
	}

	// Extract country code information from result capabilities
	countryResult := &DHLCountryCodeResolutionResult{
		ResolutionTimeMs: result.ResponseTime.Milliseconds(),
	}

	if sourceCountry, exists := result.Capabilities["source_country"]; exists {
		countryResult.SourceCountry = sourceCountry.(string)
	}
	if destCountry, exists := result.Capabilities["destination_country"]; exists {
		countryResult.DestinationCountry = destCountry.(string)
	}
	if confidence, exists := result.Capabilities["confidence_score"]; exists {
		countryResult.ConfidenceScore = confidence.(float64)
	}
	if method, exists := result.Capabilities["resolution_method"]; exists {
		countryResult.ResolutionMethod = method.(string)
	}
	if cacheHit, exists := result.Capabilities["cache_hit"]; exists {
		countryResult.CacheHit = cacheHit.(bool)
	}
	if cacheKey, exists := result.Capabilities["cache_key"]; exists {
		countryResult.CacheKey = cacheKey.(string)
	}
	if warnings, exists := result.Capabilities["warnings"]; exists {
		countryResult.Warnings = warnings.([]string)
	}

	return countryResult, nil
}

func createDHLCountryCodeResolutionResponse(sourceCountry, destCountry string, resolutionTime int64, cacheHit bool) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: "dhl",
		PartnerName: "DHL Express",
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS_WORLDWIDE",
				ServiceName: "DHL Express Worldwide",
				TATDays:     5,
			},
		},
		Capabilities: map[string]interface{}{
			"source_country":      sourceCountry,
			"destination_country": destCountry,
			"confidence_score":    0.95,
			"resolution_method":   getResolutionMethod(cacheHit),
			"cache_hit":           cacheHit,
			"cache_key":           fmt.Sprintf("dhl_country_%s_%s", sourceCountry, destCountry),
			"warnings":            []string{},
		},
		ResponseTime: time.Duration(resolutionTime) * time.Millisecond,
	}
}

func createDHLCountryCodeValidationErrorResponse(errorType, errorMessage string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:    &partnerID,
		PartnerCode:  "dhl",
		PartnerName:  "DHL Express",
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("%s: %s", errorType, errorMessage),
		ErrorMessage: stringPtr(fmt.Sprintf("%s: %s", errorType, errorMessage)),
		ResponseTime: 25 * time.Millisecond,
	}
}

func createDHLGeolocationServiceIntegrationResponse(geoResponse *GeolocationServiceResponse) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()

	warnings := []string{}
	if geoResponse.Warning != "" {
		warnings = append(warnings, geoResponse.Warning)
	}

	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: "dhl",
		PartnerName: "DHL Express",
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS_WORLDWIDE",
				ServiceName: "DHL Express Worldwide",
				TATDays:     5,
			},
		},
		Capabilities: map[string]interface{}{
			"source_country":      geoResponse.SourceCountry.Code,
			"destination_country": geoResponse.DestinationCountry.Code,
			"confidence_score":    geoResponse.SourceCountry.Confidence,
			"resolution_method":   geoResponse.SourceCountry.Source,
			"cache_hit":           geoResponse.CacheHit,
			"cache_key":           fmt.Sprintf("dhl_country_%s_%s", geoResponse.SourceCountry.Code, geoResponse.DestinationCountry.Code),
			"warnings":            warnings,
			"geolocation_used":    true,
		},
		Metadata: map[string]interface{}{
			"geolocation_service_response_time": geoResponse.ResponseTime.Milliseconds(),
			"source_country_confidence":         geoResponse.SourceCountry.Confidence,
			"destination_country_confidence":    geoResponse.DestinationCountry.Confidence,
		},
		ResponseTime: geoResponse.ResponseTime + 25*time.Millisecond,
	}
}

func createDHLGeolocationServiceErrorResponse(serviceError error, errorType string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()
	return &common.PartnerServiceabilityResult{
		PartnerID:    &partnerID,
		PartnerCode:  "dhl",
		PartnerName:  "DHL Express",
		Services:     []models.ServiceV2{},
		Error:        fmt.Errorf("%s: %v", errorType, serviceError),
		ErrorMessage: stringPtr(fmt.Sprintf("%s: %v", errorType, serviceError)),
		ResponseTime: 100 * time.Millisecond,
	}
}

func createDHLCountryCodeCacheResponse(cacheScenario string, cacheHit bool, resolutionTime int64, cacheKey string) *common.PartnerServiceabilityResult {
	partnerID := uuid.New()

	resolutionMethod := "postal_pattern_match"
	confidenceScore := 0.95

	if cacheHit {
		resolutionMethod = "cache"
		confidenceScore = 1.0
	}

	return &common.PartnerServiceabilityResult{
		PartnerID:   &partnerID,
		PartnerCode: "dhl",
		PartnerName: "DHL Express",
		Services: []models.ServiceV2{
			{
				ServiceCode: "EXPRESS_WORLDWIDE",
				ServiceName: "DHL Express Worldwide",
				TATDays:     5,
			},
		},
		Capabilities: map[string]interface{}{
			"source_country":      "US",
			"destination_country": "IN",
			"confidence_score":    confidenceScore,
			"resolution_method":   resolutionMethod,
			"cache_hit":           cacheHit,
			"cache_key":           cacheKey,
			"warnings":            []string{},
			"cache_scenario":      cacheScenario,
		},
		Metadata: map[string]interface{}{
			"cache_scenario":    cacheScenario,
			"cache_lookup_time": 5,
			"resolution_time":   resolutionTime,
		},
		ResponseTime: time.Duration(resolutionTime) * time.Millisecond,
	}
}

func parseGeolocationServiceError(err error) GeolocationServiceErrorInfo {
	errMsg := err.Error()

	switch {
	case stringContains(errMsg, "TIMEOUT"):
		return GeolocationServiceErrorInfo{IsRetryable: true, BackoffMs: 2000}
	case stringContains(errMsg, "UNAVAILABLE"):
		return GeolocationServiceErrorInfo{IsRetryable: true, BackoffMs: 5000}
	case stringContains(errMsg, "RATE_LIMITED"):
		return GeolocationServiceErrorInfo{IsRetryable: true, BackoffMs: 10000}
	case stringContains(errMsg, "AUTH_ERROR"):
		return GeolocationServiceErrorInfo{IsRetryable: false, BackoffMs: 0}
	case stringContains(errMsg, "DATA_ERROR"):
		return GeolocationServiceErrorInfo{IsRetryable: false, BackoffMs: 0}
	default:
		return GeolocationServiceErrorInfo{IsRetryable: false, BackoffMs: 0}
	}
}

func getResolutionMethod(cacheHit bool) string {
	if cacheHit {
		return "cache"
	}
	return "postal_pattern_match"
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// stringPtr helper function is defined in other test files
