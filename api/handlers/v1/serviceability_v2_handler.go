package v1

import (
	"prayog-serviceability-service/internal/services/v1"
	"prayog-serviceability-service/internal/shared/models/v1"

	"github.com/gofiber/fiber/v2"
)

// ServiceabilityV2Handler handles V2 serviceability-related HTTP requests
type ServiceabilityV2Handler struct {
	orchestrator services.ServiceabilityV2Orchestrator
}

// NewServiceabilityV2Handler creates a new V2 serviceability handler
func NewServiceabilityV2Handler(orchestrator services.ServiceabilityV2Orchestrator) *ServiceabilityV2Handler {
	return &ServiceabilityV2Handler{
		orchestrator: orchestrator,
	}
}

// CheckServiceability handles POST /serviceability/api/v2/check
func (h *ServiceabilityV2Handler) CheckServiceability(c *fiber.Ctx) error {
	var request models.ServiceabilityV2Request

	// Parse request body
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ServiceabilityV2Response{
			Success: false,
			Error: &models.ErrorResponse{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Failed to parse request body: " + err.Error(),
			},
		})
	}

	// Execute serviceability check through orchestrator
	response, err := h.orchestrator.CheckServiceability(c.Context(), &request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ServiceabilityV2Response{
			Success: false,
			Error: &models.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error: " + err.Error(),
			},
		})
	}

	// Determine appropriate HTTP status code
	statusCode := fiber.StatusOK
	if !response.Success {
		// Check if it's a validation error or other client error
		if response.Error != nil {
			switch response.Error.Code {
			case "INVALID_REQUEST", "VALIDATION_ERROR":
				statusCode = fiber.StatusBadRequest
			case "PARTNER_FILTERING_ERROR":
				statusCode = fiber.StatusUnprocessableEntity
			default:
				statusCode = fiber.StatusInternalServerError
			}
		}
	}

	return c.Status(statusCode).JSON(response)
}

// BulkCheckServiceability handles POST /serviceability/api/v2/bulk-check
func (h *ServiceabilityV2Handler) BulkCheckServiceability(c *fiber.Ctx) error {
	var request models.BulkServiceabilityV2Request

	// Parse request body
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.BulkServiceabilityV2Response{
			Success: false,
			Error: &models.ErrorResponse{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Failed to parse request body: " + err.Error(),
			},
		})
	}

	// Validate request
	if len(request.Requests) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.BulkServiceabilityV2Response{
			Success: false,
			Error: &models.ErrorResponse{
				Code:    "EMPTY_REQUEST",
				Message: "At least one serviceability request is required",
			},
		})
	}

	// Limit bulk requests to prevent abuse
	if len(request.Requests) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(models.BulkServiceabilityV2Response{
			Success: false,
			Error: &models.ErrorResponse{
				Code:    "REQUEST_TOO_LARGE",
				Message: "Maximum 100 requests allowed per bulk operation",
			},
		})
	}

	// Execute bulk serviceability check through orchestrator
	response, err := h.orchestrator.BulkCheckServiceability(c.Context(), &request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.BulkServiceabilityV2Response{
			Success: false,
			Error: &models.ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error: " + err.Error(),
			},
		})
	}

	// Determine appropriate HTTP status code
	statusCode := fiber.StatusOK
	if !response.Success {
		// Check if it's a validation error or partial failure
		if response.Error != nil {
			switch response.Error.Code {
			case "INVALID_REQUEST", "VALIDATION_ERROR":
				statusCode = fiber.StatusBadRequest
			case "PARTIAL_FAILURE":
				statusCode = fiber.StatusMultiStatus // 207 Multi-Status for partial success
			default:
				statusCode = fiber.StatusInternalServerError
			}
		}
	}

	return c.Status(statusCode).JSON(response)
}

// GetHealthStatus handles GET /serviceability/api/v2/health (optional health check endpoint)
func (h *ServiceabilityV2Handler) GetHealthStatus(c *fiber.Ctx) error {
	// This is a simple health check to verify the V2 endpoint is working
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"version": "2.0.0",
		"message": "Serviceability V2 API is operational",
		"timestamp": fiber.Map{
			"utc": c.Context().Time().UTC().Format("2006-01-02T15:04:05Z"),
		},
	})
}
