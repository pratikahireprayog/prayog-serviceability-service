package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prayog/serviceability/internal/domain"
	"github.com/prayog/serviceability/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing

type MockServiceabilityService struct {
	mock.Mock
}

func (m *MockServiceabilityService) CheckServiceability(request domain.ServiceabilityRequest) (*domain.ServiceabilityResult, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ServiceabilityResult), args.Error(1)
}

func (m *MockServiceabilityService) BulkCheckServiceability(requests []domain.ServiceabilityRequest) ([]domain.ServiceabilityResult, error) {
	args := m.Called(requests)
	return args.Get(0).([]domain.ServiceabilityResult), args.Error(1)
}

func (m *MockServiceabilityService) GetServiceAvailabilityByID(id uint) (*domain.ServiceAvailability, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ServiceAvailability), args.Error(1)
}

func (m *MockServiceabilityService) GetServiceAvailabilitiesByLocation(locationType string, locationID uint) ([]*domain.ServiceAvailability, error) {
	args := m.Called(locationType, locationID)
	return args.Get(0).([]*domain.ServiceAvailability), args.Error(1)
}

func (m *MockServiceabilityService) GetServiceAvailabilitiesByService(serviceTypeID uint) ([]*domain.ServiceAvailability, error) {
	args := m.Called(serviceTypeID)
	return args.Get(0).([]*domain.ServiceAvailability), args.Error(1)
}

func (m *MockServiceabilityService) GetServiceAvailabilitiesByOrderType(orderTypeID uint) ([]*domain.ServiceAvailability, error) {
	args := m.Called(orderTypeID)
	return args.Get(0).([]*domain.ServiceAvailability), args.Error(1)
}

func (m *MockServiceabilityService) ListServiceAvailabilities() ([]*domain.ServiceAvailability, error) {
	args := m.Called()
	return args.Get(0).([]*domain.ServiceAvailability), args.Error(1)
}

func (m *MockServiceabilityService) CreateServiceAvailability(serviceAvailability *domain.ServiceAvailability) error {
	args := m.Called(serviceAvailability)
	return args.Error(0)
}

func (m *MockServiceabilityService) UpdateServiceAvailability(serviceAvailability *domain.ServiceAvailability) error {
	args := m.Called(serviceAvailability)
	return args.Error(0)
}

func (m *MockServiceabilityService) DeleteServiceAvailability(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// Simplified mocks that stub the minimal implementations needed for the test

// MockGeoService is a simplified mock for GeoService
type MockGeoService struct{}

// Methods required by the GeoService interface - all return errors as they aren't expected to be called during test
func (m *MockGeoService) GetCountryByID(id uint) (*domain.Country, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) GetCountryByCode(code string) (*domain.Country, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) ListCountries() ([]*domain.Country, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) CreateCountry(country *domain.Country) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) UpdateCountry(country *domain.Country) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) DeleteCountry(id uint) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) GetRegionByID(id uint) (*domain.AdministrativeRegion, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) GetRegionByCode(code string) (*domain.AdministrativeRegion, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) GetRegionsByCountryID(countryID uint) ([]*domain.AdministrativeRegion, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) ListRegions() ([]*domain.AdministrativeRegion, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) CreateRegion(region *domain.AdministrativeRegion) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) UpdateRegion(region *domain.AdministrativeRegion) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) DeleteRegion(id uint) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) GetCityByID(id uint) (*domain.City, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) GetCitiesByRegionID(regionID uint) ([]*domain.City, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) ListCities() ([]*domain.City, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) CreateCity(city *domain.City) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) UpdateCity(city *domain.City) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) DeleteCity(id uint) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) GetAreaByID(id uint) (*domain.Area, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) GetAreasByCityID(cityID uint) ([]*domain.Area, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) ListAreas() ([]*domain.Area, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) CreateArea(area *domain.Area) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) UpdateArea(area *domain.Area) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) DeleteArea(id uint) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) GetPostalCodeByID(id uint) (*domain.PostalCode, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) GetPostalCodeByCode(code string) (*domain.PostalCode, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) GetPostalCodesByAreaID(areaID uint) ([]*domain.PostalCode, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) ListPostalCodes() ([]*domain.PostalCode, error) {
	return nil, errors.New("not implemented")
}

func (m *MockGeoService) CreatePostalCode(postalCode *domain.PostalCode) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) UpdatePostalCode(postalCode *domain.PostalCode) error {
	return errors.New("not implemented")
}

func (m *MockGeoService) DeletePostalCode(id uint) error {
	return errors.New("not implemented")
}

// MockOrderTypeService is a simplified mock for OrderTypeService
type MockOrderTypeService struct{}

// Methods required by the OrderTypeService interface
func (m *MockOrderTypeService) GetByID(id uint) (*domain.OrderType, error) {
	return nil, errors.New("not implemented")
}

func (m *MockOrderTypeService) GetByCode(code string) (*domain.OrderType, error) {
	return nil, errors.New("not implemented")
}

func (m *MockOrderTypeService) List() ([]*domain.OrderType, error) {
	return nil, errors.New("not implemented")
}

func (m *MockOrderTypeService) Create(orderType *domain.OrderType) error {
	return errors.New("not implemented")
}

func (m *MockOrderTypeService) Update(orderType *domain.OrderType) error {
	return errors.New("not implemented")
}

func (m *MockOrderTypeService) Delete(id uint) error {
	return errors.New("not implemented")
}

// MockServiceTypeService is a simplified mock for ServiceTypeService
type MockServiceTypeService struct{}

// Methods required by the ServiceTypeService interface
func (m *MockServiceTypeService) GetByID(id uint) (*domain.ServiceType, error) {
	return nil, errors.New("not implemented")
}

func (m *MockServiceTypeService) GetByCode(code string) (*domain.ServiceType, error) {
	return nil, errors.New("not implemented")
}

func (m *MockServiceTypeService) List() ([]*domain.ServiceType, error) {
	return nil, errors.New("not implemented")
}

func (m *MockServiceTypeService) Create(serviceType *domain.ServiceType) error {
	return errors.New("not implemented")
}

func (m *MockServiceTypeService) Update(serviceType *domain.ServiceType) error {
	return errors.New("not implemented")
}

func (m *MockServiceTypeService) Delete(id uint) error {
	return errors.New("not implemented")
}

// MockUseCaseFactory is a test implementation of UseProvider interface
type MockUseCaseFactory struct {
	serviceabilityService *MockServiceabilityService
	geoService            *MockGeoService
	orderTypeService      *MockOrderTypeService
	serviceTypeService    *MockServiceTypeService
}

// NewMockUseCaseFactory creates a new mock factory for testing
func NewMockUseCaseFactory() *MockUseCaseFactory {
	return &MockUseCaseFactory{
		serviceabilityService: new(MockServiceabilityService),
		geoService:            new(MockGeoService),
		orderTypeService:      new(MockOrderTypeService),
		serviceTypeService:    new(MockServiceTypeService),
	}
}

// ServiceabilityService returns the mock serviceability service
func (m *MockUseCaseFactory) ServiceabilityService() usecase.ServiceabilityService {
	return m.serviceabilityService
}

// GeoService returns the mock geo service
func (m *MockUseCaseFactory) GeoService() usecase.GeoService {
	return m.geoService
}

// OrderTypeService returns the mock order type service
func (m *MockUseCaseFactory) OrderTypeService() usecase.OrderTypeService {
	return m.orderTypeService
}

// ServiceTypeService returns the mock service type service
func (m *MockUseCaseFactory) ServiceTypeService() usecase.ServiceTypeService {
	return m.serviceTypeService
}

// Tests

func TestCheckServiceability(t *testing.T) {
	// Create mock factory
	mockFactory := NewMockUseCaseFactory()

	// Create handler
	handler := NewHandler(mockFactory, "/api/v1")

	// Create test request
	requestData := domain.ServiceabilityRequest{
		PostalCode:      "12345",
		CountryCode:     "US",
		OrderTypeCode:   "DELIVERY",
		ServiceTypeCode: "STANDARD",
	}

	// Create expected response
	expectedResult := &domain.ServiceabilityResult{
		PostalCode:    "12345",
		CountryCode:   "US",
		OrderTypeCode: "DELIVERY",
		IsServiceable: true,
		Services: []domain.ServiceAvailable{
			{
				ServiceTypeCode: "STANDARD",
				ServiceTypeName: "Standard Delivery",
				IsAvailable:     true,
			},
		},
	}

	// Setup mock expectation
	mockFactory.serviceabilityService.On("CheckServiceability", requestData).Return(expectedResult, nil)

	// Create request
	reqBody, _ := json.Marshal(requestData)
	req, err := http.NewRequest("POST", "/check/serviceability", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Create router for testing
	r := chi.NewRouter()
	r.Post("/check/serviceability", handler.CheckServiceability)

	// Serve the request
	r.ServeHTTP(rr, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response SuccessResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)

	// Verify that result matches what we expect
	resultBytes, _ := json.Marshal(response.Data)
	var result domain.ServiceabilityResult
	json.Unmarshal(resultBytes, &result)

	assert.Equal(t, expectedResult.PostalCode, result.PostalCode)
	assert.Equal(t, expectedResult.CountryCode, result.CountryCode)
	assert.Equal(t, expectedResult.OrderTypeCode, result.OrderTypeCode)
	assert.Equal(t, expectedResult.IsServiceable, result.IsServiceable)
	assert.Equal(t, len(expectedResult.Services), len(result.Services))

	// Verify mock was called
	mockFactory.serviceabilityService.AssertExpectations(t)
}

func TestCheckServiceabilityWithInvalidRequest(t *testing.T) {
	// Create mock factory
	mockFactory := NewMockUseCaseFactory()

	// Create handler
	handler := NewHandler(mockFactory, "/api/v1")

	// Create invalid test request (missing postal code)
	requestData := domain.ServiceabilityRequest{
		CountryCode:     "US",
		OrderTypeCode:   "DELIVERY",
		ServiceTypeCode: "STANDARD",
	}

	// Create request
	reqBody, _ := json.Marshal(requestData)
	req, err := http.NewRequest("POST", "/check/serviceability", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Create router for testing
	r := chi.NewRouter()
	r.Post("/check/serviceability", handler.CheckServiceability)

	// Serve the request
	r.ServeHTTP(rr, req)

	// Assert the response is a validation error
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "validation_error", response.Code)
}

func TestBulkCheckServiceability(t *testing.T) {
	// Create mock factory
	mockFactory := NewMockUseCaseFactory()

	// Create handler
	handler := NewHandler(mockFactory, "/api/v1")

	// Create test requests
	requestsData := []domain.ServiceabilityRequest{
		{
			PostalCode:      "12345",
			CountryCode:     "US",
			OrderTypeCode:   "DELIVERY",
			ServiceTypeCode: "STANDARD",
		},
		{
			PostalCode:      "67890",
			CountryCode:     "US",
			OrderTypeCode:   "PICKUP",
			ServiceTypeCode: "EXPRESS",
		},
	}

	// Create expected responses
	expectedResults := []domain.ServiceabilityResult{
		{
			PostalCode:    "12345",
			CountryCode:   "US",
			OrderTypeCode: "DELIVERY",
			IsServiceable: true,
			Services: []domain.ServiceAvailable{
				{
					ServiceTypeCode: "STANDARD",
					ServiceTypeName: "Standard Delivery",
					IsAvailable:     true,
				},
			},
		},
		{
			PostalCode:    "67890",
			CountryCode:   "US",
			OrderTypeCode: "PICKUP",
			IsServiceable: false,
			Services: []domain.ServiceAvailable{
				{
					ServiceTypeCode: "EXPRESS",
					ServiceTypeName: "Express Delivery",
					IsAvailable:     false,
				},
			},
		},
	}

	// Setup mock expectation
	mockFactory.serviceabilityService.On("BulkCheckServiceability", requestsData).Return(expectedResults, nil)

	// Create request
	reqBody, _ := json.Marshal(requestsData)
	req, err := http.NewRequest("POST", "/check/serviceability/bulk", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Create router for testing
	r := chi.NewRouter()
	r.Post("/check/serviceability/bulk", handler.BulkCheckServiceability)

	// Serve the request
	r.ServeHTTP(rr, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response SuccessResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)

	// Verify that mock was called
	mockFactory.serviceabilityService.AssertExpectations(t)
}
