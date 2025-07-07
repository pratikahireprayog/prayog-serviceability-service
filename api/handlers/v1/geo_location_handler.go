package v1

import (
	"strconv"
	"strings"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/services/v1"

	"github.com/gofiber/fiber/v2"
)

// GeoLocationHandler handles geo location-related HTTP requests
type GeoLocationHandler struct {
	geoLocationService services.GeoLocationService
}

// NewGeoLocationHandler creates a new geo location handler
func NewGeoLocationHandler(geoLocationService services.GeoLocationService) *GeoLocationHandler {
	return &GeoLocationHandler{
		geoLocationService: geoLocationService,
	}
}

// Helper function to parse pagination parameters
func (h *GeoLocationHandler) parsePagination(c *fiber.Ctx) *dtos.PaginationRequest {
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
func (h *GeoLocationHandler) handleError(c *fiber.Ctx, err error) error {
	errMsg := err.Error()

	// Handle specific error types
	if strings.Contains(errMsg, "not found") {
		errorResponse := dtos.NewNotFoundErrorResponse(errMsg)
		return c.Status(fiber.StatusNotFound).JSON(errorResponse)
	}

	if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "cannot be empty") {
		errorResponse := dtos.NewValidationErrorResponse(errMsg, nil)
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	if strings.Contains(errMsg, "already exists") {
		errorResponse := dtos.NewConflictErrorResponse(errMsg)
		return c.Status(fiber.StatusConflict).JSON(errorResponse)
	}

	// Default to internal server error
	errorResponse := dtos.NewInternalErrorResponse(errMsg)
	return c.Status(fiber.StatusInternalServerError).JSON(errorResponse)
}

// 1. Create API
func (h *GeoLocationHandler) CreateGeoLocation(c *fiber.Ctx) error {
	var req dtos.CreateGeoLocationRequest
	if err := c.BodyParser(&req); err != nil {
		errorResponse := dtos.NewValidationErrorResponse("Invalid request body", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	result, err := h.geoLocationService.Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// 2. Get list API with filters
func (h *GeoLocationHandler) GetAllGeoLocations(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	// Parse optional filters from query parameters
	filters := &dtos.GeoLocationFilters{}

	if countryCode := c.Query("country_code"); countryCode != "" {
		filters.CountryCode = &countryCode
	}

	if postalCode := c.Query("postal_code"); postalCode != "" {
		filters.PostalCode = &postalCode
	}

	if name := c.Query("name"); name != "" {
		filters.Name = &name
	}

	if featureCode := c.Query("feature_code"); featureCode != "" {
		filters.FeatureCode = &featureCode
	}

	result, err := h.geoLocationService.GetAll(c.Context(), pagination, filters)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// 3. Get by Postal code API
func (h *GeoLocationHandler) GetGeoLocationByPostalCode(c *fiber.Ctx) error {
	postalCode := c.Params("postalCode")
	if strings.TrimSpace(postalCode) == "" {
		errorResponse := dtos.NewValidationErrorResponse("Postal code is required", nil)
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	result, err := h.geoLocationService.GetByPostalCode(c.Context(), postalCode)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// 4. Update API
func (h *GeoLocationHandler) UpdateGeoLocation(c *fiber.Ctx) error {
	postalCode := c.Params("postal_code")
	if strings.TrimSpace(postalCode) == "" {
		errorResponse := dtos.NewValidationErrorResponse("Postal code is required", nil)
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	var req dtos.UpdateGeoLocationRequest
	if err := c.BodyParser(&req); err != nil {
		errorResponse := dtos.NewValidationErrorResponse("Invalid request body", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	result, err := h.geoLocationService.Update(c.Context(), postalCode, &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// 5. Delete API
func (h *GeoLocationHandler) DeleteGeoLocation(c *fiber.Ctx) error {
	postalCode := c.Params("postal_code")
	if strings.TrimSpace(postalCode) == "" {
		errorResponse := dtos.NewValidationErrorResponse("Postal code is required", nil)
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	result, err := h.geoLocationService.Delete(c.Context(), postalCode)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
