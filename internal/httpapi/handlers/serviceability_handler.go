package handlers

import (
	"github.com/gofiber/fiber/v2"

	"prayog-serviceability-service/internal/httpapi/middleware"
	"prayog-serviceability-service/internal/service/usecase"
)

// ServiceabilityServiceInterface defines the minimal interface required for serviceability checks
type ServiceabilityServiceInterface interface {
	CheckServiceability(location, serviceType, orderType string) (bool, error)
}

// ServiceabilityHandler handles serviceability-related HTTP requests.
type ServiceabilityHandler struct {
	usecaseFactory usecase.Factory
}

// NewServiceabilityHandler creates a new serviceability handler.
func NewServiceabilityHandler(usecaseFactory usecase.Factory) *ServiceabilityHandler {
	return &ServiceabilityHandler{
		usecaseFactory: usecaseFactory,
	}
}

// RegisterRoutes registers the serviceability routes.
func (h *ServiceabilityHandler) RegisterRoutes(r fiber.Router) {
	r.Get("/check/:postalCode", h.CheckServiceability)
	r.Post("/bulk-check", middleware.RateLimiter(), h.BulkCheckServiceability)
}

// CheckServiceability checks if a location is serviceable.
func (h *ServiceabilityHandler) CheckServiceability(c *fiber.Ctx) error {
	postalCode := c.Params("postalCode")
	if postalCode == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Postal code is required")
	}

	// TODO: Implement actual serviceability check
	isServiceable := postalCode != "00000" // Just a placeholder check

	response := map[string]interface{}{
		"postal_code":    postalCode,
		"is_serviceable": isServiceable,
	}

	return c.JSON(response)
}

// BulkCheckServiceability checks if multiple locations are serviceable.
func (h *ServiceabilityHandler) BulkCheckServiceability(c *fiber.Ctx) error {
	var request struct {
		PostalCodes []string `json:"postal_codes"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	results := make(map[string]bool)
	for _, code := range request.PostalCodes {
		// TODO: Implement actual serviceability check
		results[code] = code != "00000" // Just a placeholder check
	}

	response := map[string]interface{}{
		"results": results,
	}

	return c.JSON(response)
}
