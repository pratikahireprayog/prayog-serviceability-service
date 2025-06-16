package v1

import (
	"strconv"
	"strings"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/services/v1"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// LocationHandler handles location-related HTTP requests
type LocationHandler struct {
	locationService services.LocationService
}

// NewLocationHandler creates a new location handler
func NewLocationHandler(locationService services.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
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
func (h *LocationHandler) handleError(c *fiber.Ctx, err error) error {
	errMsg := err.Error()

	// Handle specific error types
	if strings.Contains(errMsg, "not found") {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": errMsg,
		})
	}

	if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "cannot be empty") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": errMsg,
		})
	}

	if strings.Contains(errMsg, "already exists") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": errMsg,
		})
	}

	// Default to internal server error
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": errMsg,
	})
}

// Country Handlers

// GetCountryByID retrieves a country by ID
func (h *LocationHandler) GetCountryByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Country ID is required",
		})
	}

	country, err := h.locationService.Countries().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(country)
}

// GetCountryByCode retrieves a country by code
func (h *LocationHandler) GetCountryByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Country code is required",
		})
	}

	country, err := h.locationService.Countries().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(country)
}

// GetAllCountries retrieves all countries with pagination
func (h *LocationHandler) GetAllCountries(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	countries, err := h.locationService.Countries().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(countries)
}

// CreateCountry creates a new country
func (h *LocationHandler) CreateCountry(c *fiber.Ctx) error {
	var req dtos.CreateCountryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	country, err := h.locationService.Countries().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(country)
}

// UpdateCountry updates a country
func (h *LocationHandler) UpdateCountry(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Country ID is required",
		})
	}

	var req dtos.UpdateCountryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	country, err := h.locationService.Countries().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(country)
}

// DeleteCountry deletes a country
func (h *LocationHandler) DeleteCountry(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Country ID is required",
		})
	}

	err := h.locationService.Countries().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// RegionType Handlers

// GetRegionTypeByCode retrieves a region type by code
func (h *LocationHandler) GetRegionTypeByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Region type code is required",
		})
	}

	regionType, err := h.locationService.RegionTypes().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(regionType)
}

// GetAllRegionTypes retrieves all region types with pagination
func (h *LocationHandler) GetAllRegionTypes(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	regionTypes, err := h.locationService.RegionTypes().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(regionTypes)
}

// CreateRegionType creates a new region type
func (h *LocationHandler) CreateRegionType(c *fiber.Ctx) error {
	var req dtos.CreateRegionTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	regionType, err := h.locationService.RegionTypes().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(regionType)
}

// UpdateRegionType updates a region type
func (h *LocationHandler) UpdateRegionType(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Region type code is required",
		})
	}

	var req dtos.UpdateRegionTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	regionType, err := h.locationService.RegionTypes().Update(c.Context(), code, &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(regionType)
}

// DeleteRegionType deletes a region type
func (h *LocationHandler) DeleteRegionType(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Region type code is required",
		})
	}

	err := h.locationService.RegionTypes().Delete(c.Context(), code)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// Region Handlers

// GetRegionByID retrieves a region by ID
func (h *LocationHandler) GetRegionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Region ID is required",
		})
	}

	region, err := h.locationService.Regions().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(region)
}

// GetRegionByCode retrieves a region by code
func (h *LocationHandler) GetRegionByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Region code is required",
		})
	}

	region, err := h.locationService.Regions().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(region)
}

// GetAllRegions retrieves all regions with pagination
func (h *LocationHandler) GetAllRegions(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	regions, err := h.locationService.Regions().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(regions)
}

// GetRegionsByCountryID retrieves all regions for a specific country
func (h *LocationHandler) GetRegionsByCountryID(c *fiber.Ctx) error {
	countryID := c.Params("countryID")
	if strings.TrimSpace(countryID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Country ID is required",
		})
	}

	regions, err := h.locationService.Regions().GetByCountryID(c.Context(), countryID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Regions retrieved successfully",
		Data:    regions,
	})
}

// CreateRegion creates a new region
func (h *LocationHandler) CreateRegion(c *fiber.Ctx) error {
	var req dtos.CreateRegionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	region, err := h.locationService.Regions().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(region)
}

// UpdateRegion updates a region
func (h *LocationHandler) UpdateRegion(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Region ID is required",
		})
	}

	var req dtos.UpdateRegionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	region, err := h.locationService.Regions().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(region)
}

// DeleteRegion deletes a region
func (h *LocationHandler) DeleteRegion(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Region ID is required",
		})
	}

	err := h.locationService.Regions().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// District Handlers

// GetDistrictByID retrieves a district by ID
func (h *LocationHandler) GetDistrictByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "District ID is required",
		})
	}

	district, err := h.locationService.Districts().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(district)
}

// GetDistrictByCode retrieves a district by code
func (h *LocationHandler) GetDistrictByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "District code is required",
		})
	}

	district, err := h.locationService.Districts().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(district)
}

// GetAllDistricts retrieves all districts with pagination
func (h *LocationHandler) GetAllDistricts(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	districts, err := h.locationService.Districts().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(districts)
}

// GetDistrictsByRegionID retrieves all districts for a specific region
func (h *LocationHandler) GetDistrictsByRegionID(c *fiber.Ctx) error {
	regionID := c.Params("regionID")
	if strings.TrimSpace(regionID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Region ID is required",
		})
	}

	districts, err := h.locationService.Districts().GetByRegionID(c.Context(), regionID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Districts retrieved successfully",
		Data:    districts,
	})
}

// CreateDistrict creates a new district
func (h *LocationHandler) CreateDistrict(c *fiber.Ctx) error {
	var req dtos.CreateDistrictRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	district, err := h.locationService.Districts().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(district)
}

// UpdateDistrict updates a district
func (h *LocationHandler) UpdateDistrict(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "District ID is required",
		})
	}

	var req dtos.UpdateDistrictRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	district, err := h.locationService.Districts().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(district)
}

// DeleteDistrict deletes a district
func (h *LocationHandler) DeleteDistrict(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "District ID is required",
		})
	}

	err := h.locationService.Districts().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// City Handlers

// GetCityByID retrieves a city by ID
func (h *LocationHandler) GetCityByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "City ID is required",
		})
	}

	city, err := h.locationService.Cities().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(city)
}

// GetCityByCode retrieves a city by code
func (h *LocationHandler) GetCityByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "City code is required",
		})
	}

	city, err := h.locationService.Cities().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(city)
}

// GetAllCities retrieves all cities with pagination
func (h *LocationHandler) GetAllCities(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	cities, err := h.locationService.Cities().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(cities)
}

// GetCitiesByRegionID retrieves all cities for a specific region
func (h *LocationHandler) GetCitiesByRegionID(c *fiber.Ctx) error {
	regionID := c.Params("regionID")
	if strings.TrimSpace(regionID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Region ID is required",
		})
	}

	cities, err := h.locationService.Cities().GetByRegionID(c.Context(), regionID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Cities retrieved successfully",
		Data:    cities,
	})
}

// CreateCity creates a new city
func (h *LocationHandler) CreateCity(c *fiber.Ctx) error {
	var req dtos.CreateCityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	city, err := h.locationService.Cities().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(city)
}

// UpdateCity updates a city
func (h *LocationHandler) UpdateCity(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "City ID is required",
		})
	}

	var req dtos.UpdateCityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	city, err := h.locationService.Cities().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(city)
}

// DeleteCity deletes a city
func (h *LocationHandler) DeleteCity(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "City ID is required",
		})
	}

	err := h.locationService.Cities().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// Area Handlers

// GetAreaByID retrieves an area by ID
func (h *LocationHandler) GetAreaByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Area ID is required",
		})
	}

	area, err := h.locationService.Areas().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(area)
}

// GetAreaByCode retrieves an area by code
func (h *LocationHandler) GetAreaByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Area code is required",
		})
	}

	area, err := h.locationService.Areas().GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(area)
}

// GetAllAreas retrieves all areas with pagination
func (h *LocationHandler) GetAllAreas(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	areas, err := h.locationService.Areas().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(areas)
}

// GetAreasByCityID retrieves all areas for a specific city
func (h *LocationHandler) GetAreasByCityID(c *fiber.Ctx) error {
	cityID := c.Params("cityID")
	if strings.TrimSpace(cityID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "City ID is required",
		})
	}

	areas, err := h.locationService.Areas().GetByCityID(c.Context(), cityID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Areas retrieved successfully",
		Data:    areas,
	})
}

// CreateArea creates a new area
func (h *LocationHandler) CreateArea(c *fiber.Ctx) error {
	var req dtos.CreateAreaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	area, err := h.locationService.Areas().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(area)
}

// UpdateArea updates an area
func (h *LocationHandler) UpdateArea(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Area ID is required",
		})
	}

	var req dtos.UpdateAreaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	area, err := h.locationService.Areas().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(area)
}

// DeleteArea deletes an area
func (h *LocationHandler) DeleteArea(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Area ID is required",
		})
	}

	err := h.locationService.Areas().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// Location Alias Handlers

// CreateLocationAliasByEntityID creates a new location alias for a specific entity (nested POST)
func (h *LocationHandler) CreateLocationAliasByEntityID(c *fiber.Ctx) error {
	entityID := c.Params("entityID")
	if strings.TrimSpace(entityID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Entity ID is required",
		})
	}

	var req dtos.CreateLocationAliasRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set the entity ID from the URL parameter
	req.EntityID = uuid.MustParse(entityID)

	alias, err := h.locationService.LocationAliases().Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(alias)
}

// GetLocationAliasesByEntityID retrieves all aliases for a specific entity (nested GET)
func (h *LocationHandler) GetLocationAliasesByEntityID(c *fiber.Ctx) error {
	entityID := c.Params("entityID")
	if strings.TrimSpace(entityID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Entity ID is required",
		})
	}

	aliases, err := h.locationService.LocationAliases().GetByEntityID(c.Context(), entityID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"location_aliases": aliases,
	})
}

// GetLocationAliasesByEntityTypeAndID retrieves all aliases for a specific entity type and ID (nested GET with entity type)
func (h *LocationHandler) GetLocationAliasesByEntityTypeAndID(c *fiber.Ctx) error {
	entityType := c.Params("entityType")
	entityID := c.Params("entityID")

	if strings.TrimSpace(entityType) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Entity type is required",
		})
	}

	if strings.TrimSpace(entityID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Entity ID is required",
		})
	}

	aliases, err := h.locationService.LocationAliases().GetByEntityTypeAndID(c.Context(), entityType, entityID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"location_aliases": aliases,
	})
}

// GetLocationAliasByID retrieves a location alias by ID (standalone GET)
func (h *LocationHandler) GetLocationAliasByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Location alias ID is required",
		})
	}

	alias, err := h.locationService.LocationAliases().GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(alias)
}

// GetAllLocationAliases retrieves all location aliases with pagination (standalone GET)
func (h *LocationHandler) GetAllLocationAliases(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	aliases, err := h.locationService.LocationAliases().GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(aliases)
}

// UpdateLocationAlias updates an existing location alias (standalone PUT)
func (h *LocationHandler) UpdateLocationAlias(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Location alias ID is required",
		})
	}

	var req dtos.UpdateLocationAliasRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	alias, err := h.locationService.LocationAliases().Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(alias)
}

// DeleteLocationAlias deletes a location alias (standalone DELETE)
func (h *LocationHandler) DeleteLocationAlias(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Location alias ID is required",
		})
	}

	err := h.locationService.LocationAliases().Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
