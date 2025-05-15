package v1

import (
	"prayog-serviceability-service/pkg/models"
	"prayog-serviceability-service/pkg/services"

	"github.com/gofiber/fiber/v2"
)

// ServiceabilityHandler handles serviceability-related HTTP requests.
type ServiceabilityHandler struct {
	service services.ServiceabilityService
}

// NewServiceabilityHandler creates a new serviceability handler.
func NewServiceabilityHandler(service services.ServiceabilityService) *ServiceabilityHandler {
	return &ServiceabilityHandler{
		service: service,
	}
}

// CheckServiceability checks if a location is serviceable.
func (h *ServiceabilityHandler) CheckServiceability(c *fiber.Ctx) error {
	postalCode := c.Params("postalCode")
	if postalCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Postal code is required",
		})
	}

	// Create a serviceability request with the postal code
	request := models.ServiceabilityRequest{
		PostalCode: postalCode,
	}

	result, err := h.service.CheckServiceability(c.Context(), request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// BulkCheckServiceability checks if multiple locations are serviceable.
func (h *ServiceabilityHandler) BulkCheckServiceability(c *fiber.Ctx) error {
	var request struct {
		PostalCodes []string `json:"postal_codes"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(request.PostalCodes) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one postal code is required",
		})
	}

	results, err := h.service.BulkCheckServiceability(c.Context(), request.PostalCodes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"results": results,
	})
}
