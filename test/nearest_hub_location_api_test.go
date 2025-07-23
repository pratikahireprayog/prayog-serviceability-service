package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "prayog-serviceability-service/api/handlers/v1"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNearestHubLocationRepository is a mock implementation of the repository
type MockNearestHubLocationRepository struct {
	mock.Mock
}

func (m *MockNearestHubLocationRepository) GetByPostalCode(ctx context.Context, postalCode int) (*models.NearestHubLocation, error) {
	args := m.Called(ctx, postalCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NearestHubLocation), args.Error(1)
}

func (m *MockNearestHubLocationRepository) GetByPostalCodeString(ctx context.Context, postalCode string) (*models.NearestHubLocation, error) {
	args := m.Called(ctx, postalCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NearestHubLocation), args.Error(1)
}

func (m *MockNearestHubLocationRepository) GetAll(ctx context.Context, offset, limit int) ([]models.NearestHubLocation, int64, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]models.NearestHubLocation), args.Get(1).(int64), args.Error(2)
}

func (m *MockNearestHubLocationRepository) GetByFilters(ctx context.Context, filters *dtos.NearestHubLocationFilters) ([]models.NearestHubLocation, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]models.NearestHubLocation), args.Error(1)
}

func (m *MockNearestHubLocationRepository) GetByHubPostalCode(ctx context.Context, hubPostalCode int) ([]models.NearestHubLocation, error) {
	args := m.Called(ctx, hubPostalCode)
	return args.Get(0).([]models.NearestHubLocation), args.Error(1)
}

func (m *MockNearestHubLocationRepository) GetByHubCityCode(ctx context.Context, hubCityCode string) ([]models.NearestHubLocation, error) {
	args := m.Called(ctx, hubCityCode)
	return args.Get(0).([]models.NearestHubLocation), args.Error(1)
}

func (m *MockNearestHubLocationRepository) GetNearbyHubLocations(ctx context.Context, latitude, longitude float64, radiusKm int, limit int) ([]models.NearestHubLocation, error) {
	args := m.Called(ctx, latitude, longitude, radiusKm, limit)
	return args.Get(0).([]models.NearestHubLocation), args.Error(1)
}

func (m *MockNearestHubLocationRepository) Create(ctx context.Context, hubLocation *models.NearestHubLocation) error {
	args := m.Called(ctx, hubLocation)
	return args.Error(0)
}

func (m *MockNearestHubLocationRepository) Update(ctx context.Context, postalCode int, hubLocation *models.NearestHubLocation) error {
	args := m.Called(ctx, postalCode, hubLocation)
	return args.Error(0)
}

func (m *MockNearestHubLocationRepository) Delete(ctx context.Context, postalCode int) error {
	args := m.Called(ctx, postalCode)
	return args.Error(0)
}

func (m *MockNearestHubLocationRepository) BulkCreate(ctx context.Context, hubLocations []models.NearestHubLocation) error {
	args := m.Called(ctx, hubLocations)
	return args.Error(0)
}

func (m *MockNearestHubLocationRepository) BulkUpdate(ctx context.Context, hubLocations []models.NearestHubLocation) error {
	args := m.Called(ctx, hubLocations)
	return args.Error(0)
}

func (m *MockNearestHubLocationRepository) Exists(ctx context.Context, postalCode int) (bool, error) {
	args := m.Called(ctx, postalCode)
	return args.Bool(0), args.Error(1)
}

func createTestNearestHubLocation() *models.NearestHubLocation {
	address := "Test Address"
	lat := 19.1197
	lng := 72.8696
	hubPostalCode := 400093
	hubLat := 19.1197
	hubLng := 72.8696
	cityCode := "BOM"    // Original city code
	hubCityCode := "MUM" // Hub city code
	isInternationalHub := true
	hubContactName := "Test Contact"
	hubContactPhone := "+91-9876543210"
	hubContactEmail := "test@example.com"
	hubStreet := "Test Street"
	hubLandmark := "Test Landmark"
	hubCity := "Mumbai"
	hubState := "Maharashtra"
	hubCountry := "India"
	hubLocationLat := 19.1197
	hubLocationLng := 72.8696

	return &models.NearestHubLocation{
		PostalCode:  400001,
		Address:     &address,
		CentroidLat: &lat,
		CentroidLng: &lng,

		// Original city code field (renamed from hub_city_code)
		CityCode: &cityCode,

		// Hub location fields (renamed from international_hub_*)
		HubPostalCode:  &hubPostalCode,
		HubCentroidLat: &hubLat,
		HubCentroidLng: &hubLng,
		HubCityCode:    &hubCityCode,

		// International hub flag
		IsInternationalHub: &isInternationalHub,

		// Hub contact and location fields
		HubContactPersonName:  &hubContactName,
		HubContactPersonPhone: &hubContactPhone,
		HubContactPersonEmail: &hubContactEmail,
		HubStreet:             &hubStreet,
		HubLandmark:           &hubLandmark,
		HubCity:               &hubCity,
		HubState:              &hubState,
		HubCountry:            &hubCountry,
		HubLat:                &hubLocationLat,
		HubLng:                &hubLocationLng,
	}
}

func TestCreateNearestHubLocation(t *testing.T) {
	mockRepo := new(MockNearestHubLocationRepository)
	handler := v1.NewNearestHubLocationHandler(mockRepo)

	app := fiber.New()
	app.Post("/nearest-hub-locations", handler.Create)

	t.Run("successful creation", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.NearestHubLocation")).Return(nil)

		reqBody := dtos.CreateNearestHubLocationRequest{
			PostalCode:            400001,
			HubContactPersonName:  stringPtr("Test Contact"),
			HubContactPersonPhone: stringPtr("+91-9876543210"),
			HubContactPersonEmail: stringPtr("test@example.com"),
			HubStreet:             stringPtr("Test Street"),
			HubLandmark:           stringPtr("Test Landmark"),
			HubCity:               stringPtr("Mumbai"),
			HubState:              stringPtr("Maharashtra"),
			HubCountry:            stringPtr("India"),
			HubLat:                float64Ptr(19.1197),
			HubLng:                float64Ptr(72.8696),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/nearest-hub-locations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/nearest-hub-locations", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestGetNearestHubLocationByPostalCode(t *testing.T) {
	mockRepo := new(MockNearestHubLocationRepository)
	handler := v1.NewNearestHubLocationHandler(mockRepo)

	app := fiber.New()
	app.Get("/nearest-hub-locations/:postal_code", handler.GetByPostalCode)

	t.Run("successful retrieval", func(t *testing.T) {
		testLocation := createTestNearestHubLocation()
		mockRepo.On("GetByPostalCode", mock.Anything, 400001).Return(testLocation, nil)

		req := httptest.NewRequest(http.MethodGet, "/nearest-hub-locations/400001", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify response contains new hub fields
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Test Contact", data["hub_contact_person_name"])
		assert.Equal(t, "Mumbai", data["hub_city"])
		assert.Equal(t, "Maharashtra", data["hub_state"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("postal code not found", func(t *testing.T) {
		mockRepo.On("GetByPostalCode", mock.Anything, 999999).Return(nil, fmt.Errorf("not found"))

		req := httptest.NewRequest(http.MethodGet, "/nearest-hub-locations/999999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid postal code format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/nearest-hub-locations/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestGetNearestHubLocationsByFilters(t *testing.T) {
	mockRepo := new(MockNearestHubLocationRepository)
	handler := v1.NewNearestHubLocationHandler(mockRepo)

	app := fiber.New()
	app.Get("/nearest-hub-locations", handler.GetByFilters)

	t.Run("filter by hub city", func(t *testing.T) {
		testLocations := []models.NearestHubLocation{*createTestNearestHubLocation()}
		mockRepo.On("GetByFilters", mock.Anything, mock.MatchedBy(func(filters *dtos.NearestHubLocationFilters) bool {
			return len(filters.HubCities) == 1 && filters.HubCities[0] == "Mumbai"
		})).Return(testLocations, nil)

		req := httptest.NewRequest(http.MethodGet, "/nearest-hub-locations?hub_city=Mumbai", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		data := response["data"].([]interface{})
		assert.Len(t, data, 1)

		mockRepo.AssertExpectations(t)
	})

	t.Run("filter by hub state", func(t *testing.T) {
		testLocations := []models.NearestHubLocation{*createTestNearestHubLocation()}
		mockRepo.On("GetByFilters", mock.Anything, mock.MatchedBy(func(filters *dtos.NearestHubLocationFilters) bool {
			return len(filters.HubStates) == 1 && filters.HubStates[0] == "Maharashtra"
		})).Return(testLocations, nil)

		req := httptest.NewRequest(http.MethodGet, "/nearest-hub-locations?hub_state=Maharashtra", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("filter by hub country", func(t *testing.T) {
		testLocations := []models.NearestHubLocation{*createTestNearestHubLocation()}
		mockRepo.On("GetByFilters", mock.Anything, mock.MatchedBy(func(filters *dtos.NearestHubLocationFilters) bool {
			return len(filters.HubCountries) == 1 && filters.HubCountries[0] == "India"
		})).Return(testLocations, nil)

		req := httptest.NewRequest(http.MethodGet, "/nearest-hub-locations?hub_country=India", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateNearestHubLocation(t *testing.T) {
	mockRepo := new(MockNearestHubLocationRepository)
	handler := v1.NewNearestHubLocationHandler(mockRepo)

	app := fiber.New()
	app.Put("/nearest-hub-locations/:postal_code", handler.Update)

	t.Run("successful update with new hub fields", func(t *testing.T) {
		existingLocation := createTestNearestHubLocation()
		mockRepo.On("GetByPostalCode", mock.Anything, 400001).Return(existingLocation, nil)
		mockRepo.On("Update", mock.Anything, 400001, mock.AnythingOfType("*models.NearestHubLocation")).Return(nil)

		updatedLocation := createTestNearestHubLocation()
		*updatedLocation.HubContactPersonName = "Updated Contact"
		*updatedLocation.HubCity = "Pune"
		mockRepo.On("GetByPostalCode", mock.Anything, 400001).Return(updatedLocation, nil)

		reqBody := dtos.UpdateNearestHubLocationRequest{
			HubContactPersonName: stringPtr("Updated Contact"),
			HubCity:              stringPtr("Pune"),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/nearest-hub-locations/400001", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("postal code not found", func(t *testing.T) {
		mockRepo.On("GetByPostalCode", mock.Anything, 999999).Return(nil, fmt.Errorf("not found"))

		reqBody := dtos.UpdateNearestHubLocationRequest{
			HubCity: stringPtr("TestCity"),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/nearest-hub-locations/999999", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteNearestHubLocation(t *testing.T) {
	mockRepo := new(MockNearestHubLocationRepository)
	handler := v1.NewNearestHubLocationHandler(mockRepo)

	app := fiber.New()
	app.Delete("/nearest-hub-locations/:postal_code", handler.Delete)

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo.On("Delete", mock.Anything, 400001).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/nearest-hub-locations/400001", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("postal code not found", func(t *testing.T) {
		mockRepo.On("Delete", mock.Anything, 999999).Return(fmt.Errorf("not found"))

		req := httptest.NewRequest(http.MethodDelete, "/nearest-hub-locations/999999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})
}

func TestNearestHubLocationModelValidation(t *testing.T) {
	t.Run("valid model", func(t *testing.T) {
		model := createTestNearestHubLocation()
		err := model.ValidateBusinessRules()
		assert.NoError(t, err)
	})

	t.Run("invalid postal code", func(t *testing.T) {
		model := createTestNearestHubLocation()
		model.PostalCode = -1
		err := model.ValidateBusinessRules()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "postal code must be positive")
	})

	t.Run("invalid hub latitude", func(t *testing.T) {
		model := createTestNearestHubLocation()
		invalidLat := 91.0
		model.HubLat = &invalidLat
		err := model.ValidateBusinessRules()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hub latitude must be between -90 and 90")
	})

	t.Run("invalid hub longitude", func(t *testing.T) {
		model := createTestNearestHubLocation()
		invalidLng := 181.0
		model.HubLng = &invalidLng
		err := model.ValidateBusinessRules()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hub longitude must be between -180 and 180")
	})

	t.Run("invalid email", func(t *testing.T) {
		model := createTestNearestHubLocation()
		invalidEmail := "invalid-email"
		model.HubContactPersonEmail = &invalidEmail
		err := model.ValidateBusinessRules()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hub contact person email must be a valid email address")
	})
}

func TestHubLocationInfoConversion(t *testing.T) {
	t.Run("conversion includes hub info and contact info", func(t *testing.T) {
		model := createTestNearestHubLocation()
		hubLocationInfo := model.ToHubLocationInfo()

		// Test HubInfo (basic hub information)
		assert.NotNil(t, hubLocationInfo.HubInfo)
		assert.Equal(t, 400093, *hubLocationInfo.HubInfo.PostalCode)
		assert.Equal(t, "MUM", *hubLocationInfo.HubInfo.CityCode)
		assert.Equal(t, true, *hubLocationInfo.HubInfo.IsInternationalHub)

		// Test HubContactInfo (contact and location details)
		assert.NotNil(t, hubLocationInfo.HubContactInfo)
		assert.Equal(t, "Test Contact", *hubLocationInfo.HubContactInfo.ContactPersonName)
		assert.Equal(t, "Mumbai", *hubLocationInfo.HubContactInfo.City)
		assert.Equal(t, "Maharashtra", *hubLocationInfo.HubContactInfo.State)
		assert.Equal(t, "India", *hubLocationInfo.HubContactInfo.Country)
		assert.Equal(t, 19.1197, *hubLocationInfo.HubContactInfo.Lat)
		assert.Equal(t, 72.8696, *hubLocationInfo.HubContactInfo.Lng)
	})

	t.Run("conversion without hub info", func(t *testing.T) {
		model := &models.NearestHubLocation{
			PostalCode: 400001,
		}
		hubLocationInfo := model.ToHubLocationInfo()

		assert.Nil(t, hubLocationInfo.HubInfo)
		assert.Nil(t, hubLocationInfo.HubContactInfo)
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
