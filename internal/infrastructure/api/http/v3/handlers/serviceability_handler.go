package handlers

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	handlers "prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
	v3service "prayog-serviceability-service/internal/services/v3"
	"prayog-serviceability-service/internal/shared/errors"
	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"
	modelsv3 "prayog-serviceability-service/internal/shared/models/v3"
)

// ServiceabilityHandler handles V3 serviceability check requests
type ServiceabilityHandler struct {
	service      *v3service.ServiceabilityService
	validator    *validator.Validate
	logger       *logrus.Logger
	errorHandler *handlers.ErrorHandler
}

// NewServiceabilityHandler creates a new V3 serviceability handler
func NewServiceabilityHandler(
	service *v3service.ServiceabilityService,
	validator *validator.Validate,
	logger *logrus.Logger,
) *ServiceabilityHandler {
	return &ServiceabilityHandler{
		service:      service,
		validator:    validator,
		logger:       logger,
		errorHandler: handlers.NewErrorHandler(logger),
	}
}

// CheckServiceability handles POST /check (V3 single serviceability check)
func (h *ServiceabilityHandler) CheckServiceability(c *fiber.Ctx) error {
	h.logger.Debug("V3 single serviceability check requested")

	// Parse request body
	var request modelsv3.ServiceabilityV3Request
	if err := c.BodyParser(&request); err != nil {
		h.logger.WithError(err).Error("Failed to parse request body")
		return h.errorHandler.HandleParsingError(c, err)
	}

	// Validate request locally (fields, etc)
	if err := h.validator.Struct(&request); err != nil {
		h.logger.WithError(err).Error("Request validation failed")
		return h.errorHandler.HandleValidationError(c, err)
	}

	// Fetch user partners headers
	tenantID := c.Get("x-tenant-id")
	userID := c.Get("x-user-id")
	if tenantID == "" || userID == "" {
		h.logger.Error("Missing tenant or user ID headers")
		return h.errorHandler.HandleBusinessLogicError(c, "MISSING_HEADERS", "Missing required headers", fmt.Errorf("x-tenant-id and x-user-id are required"))
	}

	v3Response, err := h.service.CheckServiceability(c.Context(), &request, tenantID, userID)
	if err != nil {
		if serviceErr, ok := err.(*errors.ServiceError); ok {
			status := serviceErr.HTTPStatus
			if status == 0 {
				status = fiber.StatusInternalServerError
			}
			return c.Status(status).JSON(modelsv3.ServiceabilityV3Response{
				Success:  false,
				Partners: []modelsv3.PartnerV3Response{},
				Error: &modelsv1.ErrorResponse{
					Code:    serviceErr.Code,
					Message: serviceErr.Message,
				},
			})
		}
		// Generic error during fetch or orchestration
		return h.errorHandler.HandleServiceError(c, "SERVICEABILITY_CHECK_FAILED", "Failed to check serviceability", err)
	}

	return c.Status(fiber.StatusOK).JSON(v3Response)
}
