package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prayog/serviceability/internal/domain"
	"github.com/prayog/serviceability/internal/transport/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockServiceabilityService is a mock implementation of the ServiceabilityService
type MockServiceabilityService struct {
	mock.Mock
}

// CheckServiceability is a mock implementation
func (m *MockServiceabilityService) CheckServiceability(postalCode, serviceType, orderType string) (bool, error) {
	args := m.Called(postalCode, serviceType, orderType)
	return args.Bool(0), args.Error(1)
}

// Required to implement the full interface
func (m *MockServiceabilityService) GetServiceableAreas(serviceType string) ([]domain.Area, error) {
	args := m.Called(serviceType)
	return args.Get(0).([]domain.Area), args.Error(1)
}

func (m *MockServiceabilityService) GetServiceabilityRule(id uint) (domain.ServiceabilityRule, error) {
	args := m.Called(id)
	return args.Get(0).(domain.ServiceabilityRule), args.Error(1)
}

func (m *MockServiceabilityService) ListServiceabilityRules(filters map[string]string) ([]domain.ServiceabilityRule, error) {
	args := m.Called(filters)
	return args.Get(0).([]domain.ServiceabilityRule), args.Error(1)
}

func (m *MockServiceabilityService) CreateServiceabilityRule(rule *domain.ServiceabilityRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockServiceabilityService) UpdateServiceabilityRule(rule *domain.ServiceabilityRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockServiceabilityService) DeleteServiceabilityRule(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCheckServiceability(t *testing.T) {
	// Create a new mock service
	mockService := new(MockServiceabilityService)

	// Set up expectations
	mockService.On("CheckServiceability", "123456", "delivery", "").Return(true, nil)

	// Create handler with the mock
	handler := NewServiceabilityHandler(mockService)

	// Create a new router for testing
	r := chi.NewRouter()
	r.Use(middleware.Response()) // Add the middleware to standardize responses
	handler.RegisterRoutes(r)

	// Create a test request
	req := httptest.NewRequest("GET", "/check?location=123456&service_type=delivery", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response body
	var response middleware.ResponseData
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	// Assert response structure
	assert.Equal(t, "success", response.Status)

	// Assert response data
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, responseData["isServiceable"])

	// Verify our expectations were met
	mockService.AssertExpectations(t)
}
