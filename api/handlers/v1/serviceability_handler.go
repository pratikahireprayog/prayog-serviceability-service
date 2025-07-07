package v1

import (
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/gofiber/fiber/v2"
)

// ServiceabilityHandler handles serviceability-related HTTP requests.
type ServiceabilityHandler struct {
	orchestrator interfaces.ServiceabilityOrchestrator
}

// NewServiceabilityHandler creates a new serviceability handler.
func NewServiceabilityHandler(orchestrator interfaces.ServiceabilityOrchestrator) *ServiceabilityHandler {
	return &ServiceabilityHandler{
		orchestrator: orchestrator,
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
	request := &models.ServiceabilityCheckRequest{
		PostalCode: &postalCode,
	}

	result, err := h.orchestrator.CheckServiceability(c.Context(), request)
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

	// Convert to bulk request format
	requests := make([]models.ServiceabilityCheckRequest, len(request.PostalCodes))
	for i, pc := range request.PostalCodes {
		requests[i] = models.ServiceabilityCheckRequest{
			PostalCode: &pc,
		}
	}

	bulkRequest := &models.BulkServiceabilityRequest{
		Requests: requests,
	}

	results, err := h.orchestrator.BulkCheckServiceability(c.Context(), bulkRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(results)
}
