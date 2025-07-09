package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/shared/services/v1"
)

// GeolocationHandler handles geo-location related HTTP requests
type GeolocationHandler struct {
	geolocationService services.GeolocationService
	logger             *logrus.Logger
}

// NewGeolocationHandler creates a new GeolocationHandler instance
func NewGeolocationHandler(geolocationService services.GeolocationService, logger *logrus.Logger) *GeolocationHandler {
	return &GeolocationHandler{
		geolocationService: geolocationService,
		logger:             logger,
	}
}

// GetCountryCodeByPostalCode retrieves the country code for a given postal code
// GET /geo-locations/postal-code/{postal_code}
func (h *GeolocationHandler) GetCountryCodeByPostalCode(c *fiber.Ctx) error {
	postalCode := c.Params("postal_code")
	if strings.TrimSpace(postalCode) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    400,
				"message": "Postal code is required",
			},
			"timestamp": c.Context().Value("timestamp"),
		})
	}

	// Get country code from geolocation service
	countryCode, err := h.geolocationService.GetCountryCodeByPostalCode(c.Context(), postalCode)
	if err != nil {
		h.logger.WithError(err).WithField("postal_code", postalCode).Error("Failed to get country code")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    500,
				"message": "Failed to get country code for postal code",
			},
			"timestamp": c.Context().Value("timestamp"),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"postal_code":  postalCode,
			"country_code": countryCode,
		},
		"timestamp": c.Context().Value("timestamp"),
	})
}

// GetLocationHierarchy retrieves the complete location hierarchy for a postal code
// GET /geo-locations/postal-code/{postal_code}/hierarchy
func (h *GeolocationHandler) GetLocationHierarchy(c *fiber.Ctx) error {
	postalCode := c.Params("postal_code")
	if strings.TrimSpace(postalCode) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    400,
				"message": "Postal code is required",
			},
			"timestamp": c.Context().Value("timestamp"),
		})
	}

	// Get location hierarchy from geolocation service
	hierarchy, err := h.geolocationService.GetLocationHierarchy(c.Context(), postalCode)
	if err != nil {
		h.logger.WithError(err).WithField("postal_code", postalCode).Error("Failed to get location hierarchy")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    500,
				"message": "Failed to get location hierarchy for postal code",
			},
			"timestamp": c.Context().Value("timestamp"),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"data":      hierarchy,
		"timestamp": c.Context().Value("timestamp"),
	})
}

// ValidateInternationalRequest validates if a request is international and returns country code
// GET /geo-locations/postal-code/{postal_code}/validate-international
func (h *GeolocationHandler) ValidateInternationalRequest(c *fiber.Ctx) error {
	postalCode := c.Params("postal_code")
	if strings.TrimSpace(postalCode) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    400,
				"message": "Postal code is required",
			},
			"timestamp": c.Context().Value("timestamp"),
		})
	}

	// Validate international request from geolocation service
	isInternational, countryCode, err := h.geolocationService.ValidateInternationalRequest(c.Context(), postalCode)
	if err != nil {
		h.logger.WithError(err).WithField("postal_code", postalCode).Error("Failed to validate international request")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    500,
				"message": "Failed to validate international request for postal code",
			},
			"timestamp": c.Context().Value("timestamp"),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"postal_code":      postalCode,
			"is_international": isInternational,
			"country_code":     countryCode,
		},
		"timestamp": c.Context().Value("timestamp"),
	})
}

// GetGeolocationHealth returns the health status of the geolocation service
// GET /geo-locations/health
func (h *GeolocationHandler) GetGeolocationHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"service": "geolocation",
			"status":  "healthy",
			"message": "Geolocation service is operational",
		},
		"timestamp": c.Context().Value("timestamp"),
	})
}
