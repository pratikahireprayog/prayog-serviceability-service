package v1

import (
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type NearestHubLocationHandler struct {
	repo repositories.NearestHubLocationRepository
}

func NewNearestHubLocationHandler(repo repositories.NearestHubLocationRepository) *NearestHubLocationHandler {
	return &NearestHubLocationHandler{repo: repo}
}

// POST /nearest-hub-locations
func (h *NearestHubLocationHandler) Create(c *fiber.Ctx) error {
	var req dtos.CreateNearestHubLocationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": 400, "message": "invalid request body"},
		})
	}
	model := createNearestHubLocationModelFromDTO(&req)
	if err := h.repo.Create(c.Context(), model); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": 500, "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Nearest hub location created successfully",
		"data":    modelToNearestHubLocationResponse(model),
	})
}

// GET /nearest-hub-locations/:postal_code
func (h *NearestHubLocationHandler) GetByPostalCode(c *fiber.Ctx) error {
	postalCode, err := strconv.Atoi(c.Params("postal_code"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": 400, "message": "invalid postal_code"},
		})
	}
	loc, err := h.repo.GetByPostalCode(c.Context(), postalCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": 404, "message": "not found"},
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Nearest hub location fetched successfully",
		"data":    modelToNearestHubLocationResponse(loc),
	})
}

// GET /nearest-hub-locations (with filters)
func (h *NearestHubLocationHandler) GetByFilters(c *fiber.Ctx) error {
	filters := parseNearestHubLocationFiltersFromQuery(c)
	locs, err := h.repo.GetByFilters(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": 500, "message": err.Error()},
		})
	}
	resp := make([]dtos.NearestHubLocationResponse, len(locs))
	for i, loc := range locs {
		resp[i] = *modelToNearestHubLocationResponse(&loc)
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Nearest hub locations fetched successfully",
		"data":    resp,
	})
}

// PUT /nearest-hub-locations/:postal_code
func (h *NearestHubLocationHandler) Update(c *fiber.Ctx) error {
	postalCode, err := strconv.Atoi(c.Params("postal_code"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": 400, "message": "invalid postal_code"},
		})
	}
	var req dtos.UpdateNearestHubLocationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": 400, "message": "invalid request body"},
		})
	}
	if err := h.repo.Update(c.Context(), postalCode, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": 500, "message": err.Error()},
		})
	}
	loc, err := h.repo.GetByPostalCode(c.Context(), postalCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": 404, "message": "not found after update"},
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Nearest hub location updated successfully",
		"data":    modelToNearestHubLocationResponse(loc),
	})
}

// DELETE /nearest-hub-locations/:postal_code
func (h *NearestHubLocationHandler) Delete(c *fiber.Ctx) error {
	postalCode, err := strconv.Atoi(c.Params("postal_code"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": 400, "message": "invalid postal_code"},
		})
	}
	if err := h.repo.Delete(c.Context(), postalCode); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": 500, "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"success": true,
		"message": "Nearest hub location deleted successfully",
		"data":    nil,
	})
}

// --- Helpers ---
func createNearestHubLocationModelFromDTO(req *dtos.CreateNearestHubLocationRequest) *models.NearestHubLocation {
	return &models.NearestHubLocation{
		PostalCode:                  req.PostalCode,
		Address:                     req.Address,
		CentroidLat:                 req.CentroidLat,
		CentroidLng:                 req.CentroidLng,
		InternationalHubPostalCode:  req.InternationalHubPostalCode,
		InternationalHubAddress:     req.InternationalHubAddress,
		InternationalHubCentroidLat: req.InternationalHubCentroidLat,
		InternationalHubCentroidLng: req.InternationalHubCentroidLng,
		InternationalHubCityCode:    req.InternationalHubCityCode,
		HubCityCode:                 req.HubCityCode,
	}
}

func modelToNearestHubLocationResponse(m *models.NearestHubLocation) *dtos.NearestHubLocationResponse {
	return &dtos.NearestHubLocationResponse{
		PostalCode:                  m.PostalCode,
		Address:                     m.Address,
		CentroidLat:                 m.CentroidLat,
		CentroidLng:                 m.CentroidLng,
		InternationalHubPostalCode:  m.InternationalHubPostalCode,
		InternationalHubAddress:     m.InternationalHubAddress,
		InternationalHubCentroidLat: m.InternationalHubCentroidLat,
		InternationalHubCentroidLng: m.InternationalHubCentroidLng,
		InternationalHubCityCode:    m.InternationalHubCityCode,
		HubCityCode:                 m.HubCityCode,
	}
}

func parseNearestHubLocationFiltersFromQuery(c *fiber.Ctx) *dtos.NearestHubLocationFilters {
	var filters dtos.NearestHubLocationFilters
	queryArgs := c.Request().URI().QueryArgs()
	// Parse multiple postal_code
	postalCodes := queryArgs.PeekMulti("postal_code")
	for _, v := range postalCodes {
		if code, err := strconv.Atoi(string(v)); err == nil {
			filters.PostalCodes = append(filters.PostalCodes, code)
		}
	}
	// Parse multiple international_hub_postal_code
	ihpCodes := queryArgs.PeekMulti("international_hub_postal_code")
	for _, v := range ihpCodes {
		if code, err := strconv.Atoi(string(v)); err == nil {
			filters.InternationalHubPostalCodes = append(filters.InternationalHubPostalCodes, code)
		}
	}
	// Parse multiple international_hub_city_code
	ihcCodes := queryArgs.PeekMulti("international_hub_city_code")
	for _, v := range ihcCodes {
		filters.InternationalHubCityCodes = append(filters.InternationalHubCityCodes, string(v))
	}
	return &filters
}
