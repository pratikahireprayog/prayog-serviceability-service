package handlers

import (
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/services/v1"
)

// LocationHandler handles location-related HTTP requests in the infrastructure layer
type LocationHandler struct {
	locationService services.LocationService
	validator       *validator.Validate
	logger          *logrus.Logger
}

// NewLocationHandler creates a new location handler with dependencies
func NewLocationHandler(
	locationService services.LocationService,
	validator *validator.Validate,
	logger *logrus.Logger,
) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
		validator:       validator,
		logger:          logger,
	}
}

// Helper function to parse pagination parameters
func (h *LocationHandler) parsePagination(c *fiber.Ctx) *dtos.PaginationRequest {
	offsetStr := c.Query("offset", "0")
	limitStr := c.Query("limit", "10")

	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	// Validate and set defaults
	if offset < 0 {
		offset = 0
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	return &dtos.PaginationRequest{
		Offset: offset,
		Limit:  limit,
	}
}

// Helper function to handle errors consistently
func (h *LocationHandler) handleError(c *fiber.Ctx, err error, operation string) error {
	errMsg := err.Error()

	// Log the error with context
	h.logger.WithFields(logrus.Fields{
		"operation": operation,
		"error":     errMsg,
		"path":      c.Path(),
		"method":    c.Method(),
	}).Error("Location handler error")

	// Handle specific error types
	if strings.Contains(errMsg, "not found") {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusNotFound,
				"message": errMsg,
			},
		})
	}

	if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "cannot be empty") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": errMsg,
			},
		})
	}

	if strings.Contains(errMsg, "already exists") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusConflict,
				"message": errMsg,
			},
		})
	}

	// Default to internal server error
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"message": errMsg,
		},
	})
}

// Helper function to validate request body
func (h *LocationHandler) validateRequest(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Invalid request body",
			},
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Validation failed",
				"details": err.Error(),
			},
		})
	}

	return nil
}

// Country Handlers

// GetCountryByID retrieves a country by ID
func (h *LocationHandler) GetCountryByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Country ID is required",
			},
		})
	}

	country, err := h.locationService.Countries().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "GetCountryByID")
	}

	return c.JSON(country)
}

// GetCountryByCode retrieves a country by code
func (h *LocationHandler) GetCountryByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Country code is required",
			},
		})
	}

	country, err := h.locationService.Countries().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err, "GetCountryByCode")
	}

	return c.JSON(country)
}

// GetAllCountries retrieves all countries with pagination
func (h *LocationHandler) GetAllCountries(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	countries, err := h.locationService.Countries().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err, "GetAllCountries")
	}

	return c.JSON(countries)
}

// CreateCountry creates a new country
func (h *LocationHandler) CreateCountry(c *fiber.Ctx) error {
	var req dtos.CreateCountryRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	country, err := h.locationService.Countries().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "CreateCountry")
	}

	return c.Status(fiber.StatusCreated).JSON(country)
}

// UpdateCountry updates a country
func (h *LocationHandler) UpdateCountry(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Country ID is required",
			},
		})
	}

	var req dtos.UpdateCountryRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	country, err := h.locationService.Countries().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err, "UpdateCountry")
	}

	return c.JSON(country)
}

// DeleteCountry deletes a country
func (h *LocationHandler) DeleteCountry(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Country ID is required",
			},
		})
	}

	err := h.locationService.Countries().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "DeleteCountry")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// RegionType Handlers

// GetRegionTypeByCode retrieves a region type by code
func (h *LocationHandler) GetRegionTypeByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Region type code is required",
			},
		})
	}

	regionType, err := h.locationService.RegionTypes().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err, "GetRegionTypeByCode")
	}

	return c.JSON(regionType)
}

// GetAllRegionTypes retrieves all region types with pagination
func (h *LocationHandler) GetAllRegionTypes(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	regionTypes, err := h.locationService.RegionTypes().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err, "GetAllRegionTypes")
	}

	return c.JSON(regionTypes)
}

// CreateRegionType creates a new region type
func (h *LocationHandler) CreateRegionType(c *fiber.Ctx) error {
	var req dtos.CreateRegionTypeRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	regionType, err := h.locationService.RegionTypes().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "CreateRegionType")
	}

	return c.Status(fiber.StatusCreated).JSON(regionType)
}

// UpdateRegionType updates a region type
func (h *LocationHandler) UpdateRegionType(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Region type code is required",
			},
		})
	}

	var req dtos.UpdateRegionTypeRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	regionType, err := h.locationService.RegionTypes().Update(c.Context(), code, &req)
	if err != nil {
		return h.handleError(c, err, "UpdateRegionType")
	}

	return c.JSON(regionType)
}

// DeleteRegionType deletes a region type
func (h *LocationHandler) DeleteRegionType(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Region type code is required",
			},
		})
	}

	err := h.locationService.RegionTypes().Delete(c.Context(), code)
	if err != nil {
		return h.handleError(c, err, "DeleteRegionType")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// Region Handlers

// GetRegionByID retrieves a region by ID
func (h *LocationHandler) GetRegionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Region ID is required",
			},
		})
	}

	region, err := h.locationService.Regions().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "GetRegionByID")
	}

	return c.JSON(region)
}

// GetRegionByCode retrieves a region by code
func (h *LocationHandler) GetRegionByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Region code is required",
			},
		})
	}

	region, err := h.locationService.Regions().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err, "GetRegionByCode")
	}

	return c.JSON(region)
}

// GetAllRegions retrieves all regions with pagination
func (h *LocationHandler) GetAllRegions(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	regions, err := h.locationService.Regions().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err, "GetAllRegions")
	}

	return c.JSON(regions)
}

// GetRegionsByCountryID retrieves all regions for a specific country
func (h *LocationHandler) GetRegionsByCountryID(c *fiber.Ctx) error {
	countryID := c.Params("countryID")
	if strings.TrimSpace(countryID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Country ID is required",
			},
		})
	}

	regions, err := h.locationService.Regions().GetByCountryID(c.Context(), countryID)
	if err != nil {
		return h.handleError(c, err, "GetRegionsByCountryID")
	}

	return c.JSON(fiber.Map{
		"regions": regions,
	})
}

// CreateRegion creates a new region
func (h *LocationHandler) CreateRegion(c *fiber.Ctx) error {
	var req dtos.CreateRegionRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	region, err := h.locationService.Regions().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "CreateRegion")
	}

	return c.Status(fiber.StatusCreated).JSON(region)
}

// UpdateRegion updates a region
func (h *LocationHandler) UpdateRegion(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Region ID is required",
			},
		})
	}

	var req dtos.UpdateRegionRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	region, err := h.locationService.Regions().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err, "UpdateRegion")
	}

	return c.JSON(region)
}

// DeleteRegion deletes a region
func (h *LocationHandler) DeleteRegion(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Region ID is required",
			},
		})
	}

	err := h.locationService.Regions().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "DeleteRegion")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// District Handlers

// GetDistrictByID retrieves a district by ID
func (h *LocationHandler) GetDistrictByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "District ID is required",
			},
		})
	}

	district, err := h.locationService.Districts().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "GetDistrictByID")
	}

	return c.JSON(district)
}

// GetDistrictByCode retrieves a district by code
func (h *LocationHandler) GetDistrictByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "District code is required",
			},
		})
	}

	district, err := h.locationService.Districts().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err, "GetDistrictByCode")
	}

	return c.JSON(district)
}

// GetAllDistricts retrieves all districts with pagination
func (h *LocationHandler) GetAllDistricts(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	districts, err := h.locationService.Districts().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err, "GetAllDistricts")
	}

	return c.JSON(districts)
}

// GetDistrictsByRegionID retrieves all districts for a specific region
func (h *LocationHandler) GetDistrictsByRegionID(c *fiber.Ctx) error {
	regionID := c.Params("regionID")
	if strings.TrimSpace(regionID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Region ID is required",
			},
		})
	}

	districts, err := h.locationService.Districts().GetByRegionID(c.Context(), regionID)
	if err != nil {
		return h.handleError(c, err, "GetDistrictsByRegionID")
	}

	return c.JSON(fiber.Map{
		"districts": districts,
	})
}

// CreateDistrict creates a new district
func (h *LocationHandler) CreateDistrict(c *fiber.Ctx) error {
	var req dtos.CreateDistrictRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	district, err := h.locationService.Districts().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "CreateDistrict")
	}

	return c.Status(fiber.StatusCreated).JSON(district)
}

// UpdateDistrict updates a district
func (h *LocationHandler) UpdateDistrict(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "District ID is required",
			},
		})
	}

	var req dtos.UpdateDistrictRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	district, err := h.locationService.Districts().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err, "UpdateDistrict")
	}

	return c.JSON(district)
}

// DeleteDistrict deletes a district
func (h *LocationHandler) DeleteDistrict(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "District ID is required",
			},
		})
	}

	err := h.locationService.Districts().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "DeleteDistrict")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// City Handlers

// GetCityByID retrieves a city by ID
func (h *LocationHandler) GetCityByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "City ID is required",
			},
		})
	}

	city, err := h.locationService.Cities().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "GetCityByID")
	}

	return c.JSON(city)
}

// GetCityByCode retrieves a city by code
func (h *LocationHandler) GetCityByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "City code is required",
			},
		})
	}

	city, err := h.locationService.Cities().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err, "GetCityByCode")
	}

	return c.JSON(city)
}

// GetAllCities retrieves all cities with pagination
func (h *LocationHandler) GetAllCities(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	cities, err := h.locationService.Cities().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err, "GetAllCities")
	}

	return c.JSON(cities)
}

// GetCitiesByRegionID retrieves all cities for a specific region
func (h *LocationHandler) GetCitiesByRegionID(c *fiber.Ctx) error {
	regionID := c.Params("regionID")
	if strings.TrimSpace(regionID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Region ID is required",
			},
		})
	}

	cities, err := h.locationService.Cities().GetByRegionID(c.Context(), regionID)
	if err != nil {
		return h.handleError(c, err, "GetCitiesByRegionID")
	}

	return c.JSON(fiber.Map{
		"cities": cities,
	})
}

// CreateCity creates a new city
func (h *LocationHandler) CreateCity(c *fiber.Ctx) error {
	var req dtos.CreateCityRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	city, err := h.locationService.Cities().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "CreateCity")
	}

	return c.Status(fiber.StatusCreated).JSON(city)
}

// UpdateCity updates a city
func (h *LocationHandler) UpdateCity(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "City ID is required",
			},
		})
	}

	var req dtos.UpdateCityRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	city, err := h.locationService.Cities().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err, "UpdateCity")
	}

	return c.JSON(city)
}

// DeleteCity deletes a city
func (h *LocationHandler) DeleteCity(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "City ID is required",
			},
		})
	}

	err := h.locationService.Cities().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "DeleteCity")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// Area Handlers

// GetAreaByID retrieves an area by ID
func (h *LocationHandler) GetAreaByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Area ID is required",
			},
		})
	}

	area, err := h.locationService.Areas().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "GetAreaByID")
	}

	return c.JSON(area)
}

// GetAreaByCode retrieves an area by code
func (h *LocationHandler) GetAreaByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Area code is required",
			},
		})
	}

	area, err := h.locationService.Areas().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err, "GetAreaByCode")
	}

	return c.JSON(area)
}

// GetAllAreas retrieves all areas with pagination
func (h *LocationHandler) GetAllAreas(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	areas, err := h.locationService.Areas().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err, "GetAllAreas")
	}

	return c.JSON(areas)
}

// GetAreasByCityID retrieves all areas for a specific city
func (h *LocationHandler) GetAreasByCityID(c *fiber.Ctx) error {
	cityID := c.Params("cityID")
	if strings.TrimSpace(cityID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "City ID is required",
			},
		})
	}

	areas, err := h.locationService.Areas().GetByCityID(c.Context(), cityID)
	if err != nil {
		return h.handleError(c, err, "GetAreasByCityID")
	}

	return c.JSON(fiber.Map{
		"areas": areas,
	})
}

// CreateArea creates a new area
func (h *LocationHandler) CreateArea(c *fiber.Ctx) error {
	var req dtos.CreateAreaRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	area, err := h.locationService.Areas().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "CreateArea")
	}

	return c.Status(fiber.StatusCreated).JSON(area)
}

// UpdateArea updates an area
func (h *LocationHandler) UpdateArea(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Area ID is required",
			},
		})
	}

	var req dtos.UpdateAreaRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	area, err := h.locationService.Areas().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err, "UpdateArea")
	}

	return c.JSON(area)
}

// DeleteArea deletes an area
func (h *LocationHandler) DeleteArea(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Area ID is required",
			},
		})
	}

	err := h.locationService.Areas().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "DeleteArea")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// PostalCode Handlers

// GetPostalCodeByID retrieves a postal code by ID
func (h *LocationHandler) GetPostalCodeByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Postal code ID is required",
			},
		})
	}

	postalCode, err := h.locationService.PostalCodes().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "GetPostalCodeByID")
	}

	return c.JSON(postalCode)
}

// GetAllPostalCodes retrieves all postal codes with pagination
func (h *LocationHandler) GetAllPostalCodes(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	postalCodes, err := h.locationService.PostalCodes().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err, "GetAllPostalCodes")
	}

	return c.JSON(postalCodes)
}

// GetPostalCodesByLocation retrieves postal codes by location parameters
func (h *LocationHandler) GetPostalCodesByLocation(c *fiber.Ctx) error {
	// Extract location parameters from query
	countryCode := c.Query("country_code")
	regionCode := c.Query("region_code")
	cityCode := c.Query("city_code")
	areaCode := c.Query("area_code")

	postalCodes, err := h.locationService.PostalCodes().GetByLocation(c.Context(), countryCode, regionCode, cityCode, areaCode)
	if err != nil {
		return h.handleError(c, err, "GetPostalCodesByLocation")
	}

	return c.JSON(fiber.Map{
		"postal_codes": postalCodes,
	})
}

// CreatePostalCode creates a new postal code
func (h *LocationHandler) CreatePostalCode(c *fiber.Ctx) error {
	var req dtos.CreatePostalCodeRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	postalCode, err := h.locationService.PostalCodes().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "CreatePostalCode")
	}

	return c.Status(fiber.StatusCreated).JSON(postalCode)
}

// UpdatePostalCode updates a postal code
func (h *LocationHandler) UpdatePostalCode(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Postal code ID is required",
			},
		})
	}

	var req dtos.UpdatePostalCodeRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	postalCode, err := h.locationService.PostalCodes().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err, "UpdatePostalCode")
	}

	return c.JSON(postalCode)
}

// DeletePostalCode deletes a postal code
func (h *LocationHandler) DeletePostalCode(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Postal code ID is required",
			},
		})
	}

	err := h.locationService.PostalCodes().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "DeletePostalCode")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// LocationType Handlers

// GetLocationTypeByCode retrieves a location type by code
func (h *LocationHandler) GetLocationTypeByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Location type code is required",
			},
		})
	}

	locationType, err := h.locationService.LocationTypes().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err, "GetLocationTypeByCode")
	}

	return c.JSON(locationType)
}

// GetAllLocationTypes retrieves all location types with pagination
func (h *LocationHandler) GetAllLocationTypes(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	locationTypes, err := h.locationService.LocationTypes().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err, "GetAllLocationTypes")
	}

	return c.JSON(locationTypes)
}

// CreateLocationType creates a new location type
func (h *LocationHandler) CreateLocationType(c *fiber.Ctx) error {
	var req dtos.CreateLocationTypeRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	locationType, err := h.locationService.LocationTypes().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "CreateLocationType")
	}

	return c.Status(fiber.StatusCreated).JSON(locationType)
}

// UpdateLocationType updates a location type
func (h *LocationHandler) UpdateLocationType(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Location type code is required",
			},
		})
	}

	var req dtos.UpdateLocationTypeRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	locationType, err := h.locationService.LocationTypes().Update(c.Context(), code, &req)
	if err != nil {
		return h.handleError(c, err, "UpdateLocationType")
	}

	return c.JSON(locationType)
}

// DeleteLocationType deletes a location type
func (h *LocationHandler) DeleteLocationType(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Location type code is required",
			},
		})
	}

	err := h.locationService.LocationTypes().Delete(c.Context(), code)
	if err != nil {
		return h.handleError(c, err, "DeleteLocationType")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
