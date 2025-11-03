package handlers

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	services "prayog-serviceability-service/internal/services/v1/data"
	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/internal/shared/errors"
	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"
)

// ServiceabilityV3Handler handles V3 serviceability check requests
type ServiceabilityV3Handler struct {
	v2Orchestrator     orchestrators.ServiceabilityOrchestrator
	geolocationService services.GeolocationService
	validator          *validator.Validate
	logger             *logrus.Logger
	errorHandler       *ErrorHandler
}

// NewServiceabilityV3Handler creates a new V3 serviceability handler
func NewServiceabilityV3Handler(
	v2Orchestrator orchestrators.ServiceabilityOrchestrator,
	geolocationService services.GeolocationService,
	validator *validator.Validate,
	logger *logrus.Logger,
) *ServiceabilityV3Handler {
	return &ServiceabilityV3Handler{
		v2Orchestrator:     v2Orchestrator,
		geolocationService: geolocationService,
		validator:          validator,
		logger:             logger,
		errorHandler:       NewErrorHandler(logger),
	}
}

// CheckServiceability handles POST /check (V3 single serviceability check)
func (h *ServiceabilityV3Handler) CheckServiceability(c *fiber.Ctx) error {
	h.logger.Debug("V3 single serviceability check requested")

	// Parse request body
	var request modelsv1.ServiceabilityV3Request
	if err := c.BodyParser(&request); err != nil {
		h.logger.WithError(err).Error("Failed to parse request body")
		return h.errorHandler.HandleParsingError(c, err)
	}

	h.logger.Debug("Request parsed successfully")

	// Validate request
	if err := h.validator.Struct(&request); err != nil {
		h.logger.WithError(err).Error("Request validation failed")
		return h.errorHandler.HandleValidationError(c, err)
	}

	h.logger.Debug("Request struct validation passed")

	// Convert V3 request to V2 request format
	v2Request := request.ToV2Request()
	if v2Request == nil {
		h.logger.Error("Failed to convert V3 request to V2 format")
		return h.errorHandler.HandleBusinessLogicError(c, ErrorCodeInvalidRequestType, "Invalid request format", fmt.Errorf("addresses object is required"))
	}

	// Force all parcel categories to use international strategy behavior for V3
	// V3 API always uses international strategy regardless of parcel_category
	international := "international"
	v2Request.ParcelCategory = &international
	h.logger.WithFields(logrus.Fields{
		"component":      "serviceability_v3_handler",
		"parcel_category": "international",
		"original_category": request.ParcelCategory,
	}).Info("Forcing parcel_category to international for V3 request (all categories use international strategy)")

	h.logger.WithFields(logrus.Fields{
		"component":             "serviceability_v3_handler",
		"source_postal_code":    v2Request.SourcePostalCode,
		"destination_postal_code": v2Request.DestinationPostalCode,
		"parcel_category":       v2Request.ParcelCategory,
	}).Info("Converted V3 request to V2 format")

	// Call V2 orchestrator (which will route to appropriate strategy, likely international)
	h.logger.Debug("Calling V2 orchestrator")
	v2Response, err := h.v2Orchestrator.CheckServiceability(c.Context(), v2Request)
	if err != nil {
		h.logger.WithError(err).Error("V2 orchestrator failed")

		// Handle structured errors with appropriate HTTP status codes
		if serviceErr, ok := err.(*errors.ServiceError); ok {
			return c.Status(serviceErr.HTTPStatus).JSON(modelsv1.ServiceabilityV3Response{
				Success:  false,
				Partners: []modelsv1.PartnerV3Response{},
				Error: &modelsv1.ErrorResponse{
					Code:    serviceErr.Code,
					Message: serviceErr.Message,
				},
			})
		}

		// Fallback to generic error handler for non-structured errors
		return h.errorHandler.HandleServiceError(c, ErrorCodeServiceabilityCheckFailed, "Failed to check V3 serviceability", err)
	}

	h.logger.Debug("V2 orchestrator completed successfully")

	// Convert V2 response to V3 format
	v3Response := modelsv1.ServiceabilityV3ResponseFromV2(v2Response)
	if v3Response == nil {
		h.logger.Error("Failed to convert V2 response to V3 format")
		return h.errorHandler.HandleServiceError(c, ErrorCodeServiceabilityCheckFailed, "Failed to convert response", fmt.Errorf("response conversion failed"))
	}

	h.logger.Debug("V3 response conversion completed")

	// Return response with 200 OK status
	// Note: Success=false in response doesn't mean HTTP error - it means no serviceable partners found
	return c.Status(fiber.StatusOK).JSON(v3Response)
}

// GetHealthStatus handles GET /health (V3 health check)
func (h *ServiceabilityV3Handler) GetHealthStatus(c *fiber.Ctx) error {
	h.logger.Debug("V3 health check requested")

	// Create health response
	healthResponse := &modelsv1.ServiceabilityV3Response{
		Success: true,
		Partners: []modelsv1.PartnerV3Response{
			{
				PartnerID:   "system",
				PartnerName: "V3 System Health",
				Rating:      5.0,
				Services: []modelsv1.ServiceV2{
					{
						ServiceName: "health-check",
						TATDays:     1,
						IsCOD:       false,
						Pickup:      true,
						Delivery:    true,
					},
				},
			},
		},
		Metadata: &modelsv1.V3ResponseMetadata{
			TotalPartners:    1,
			ServiceableCount: 1,
			Filters: modelsv1.V2Filters{
				CountryCode: func() *string { s := "IN"; return &s }(),
			},
		},
	}

	return c.Status(fiber.StatusOK).JSON(healthResponse)
}

