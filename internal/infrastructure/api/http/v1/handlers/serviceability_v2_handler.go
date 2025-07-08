package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/services/v1"
	modelsv1 "prayog-serviceability-service/internal/shared/models/v1"
)

// ServiceabilityV2Handler handles V2 serviceability check requests
type ServiceabilityV2Handler struct {
	v2Orchestrator services.ServiceabilityV2Orchestrator
	validator      *validator.Validate
	logger         *logrus.Logger
	errorHandler   *ErrorHandler
}

// NewServiceabilityV2Handler creates a new V2 serviceability handler
func NewServiceabilityV2Handler(
	v2Orchestrator services.ServiceabilityV2Orchestrator,
	validator *validator.Validate,
	logger *logrus.Logger,
) *ServiceabilityV2Handler {
	return &ServiceabilityV2Handler{
		v2Orchestrator: v2Orchestrator,
		validator:      validator,
		logger:         logger,
		errorHandler:   NewErrorHandler(logger),
	}
}

// CheckServiceabilityV2 handles POST /check (V2 single serviceability check)
func (h *ServiceabilityV2Handler) CheckServiceabilityV2(c *fiber.Ctx) error {
	h.logger.Debug("V2 single serviceability check requested")

	// Parse request body
	var request modelsv1.ServiceabilityV2Request
	if err := c.BodyParser(&request); err != nil {
		return h.errorHandler.HandleParsingError(c, err)
	}

	// Validate request
	if err := h.validator.Struct(&request); err != nil {
		return h.errorHandler.HandleValidationError(c, err)
	}

	// Additional validation for postal codes and country
	if err := h.validateV2Request(&request); err != nil {
		return h.errorHandler.HandleBusinessLogicError(c, ErrorCodeInvalidRequestType, "Invalid request data", err)
	}

	// Call V2 orchestrator
	response, err := h.v2Orchestrator.CheckServiceability(c.Context(), &request)
	if err != nil {
		return h.errorHandler.HandleServiceError(c, ErrorCodeServiceabilityCheckFailed, "Failed to check V2 serviceability", err)
	}

	// Return response with appropriate status code
	statusCode := fiber.StatusOK
	if !response.Success {
		statusCode = fiber.StatusInternalServerError
	}

	return c.Status(statusCode).JSON(response)
}

// BulkCheckServiceabilityV2 handles POST /bulk-check (V2 bulk serviceability check)
func (h *ServiceabilityV2Handler) BulkCheckServiceabilityV2(c *fiber.Ctx) error {
	h.logger.Debug("V2 bulk serviceability check requested")

	// Parse request body
	var request modelsv1.BulkServiceabilityV2Request
	if err := c.BodyParser(&request); err != nil {
		return h.errorHandler.HandleParsingError(c, err)
	}

	// Validate request
	if err := h.validator.Struct(&request); err != nil {
		return h.errorHandler.HandleValidationError(c, err)
	}

	// Validate each individual request in the bulk
	for i, req := range request.Requests {
		if err := h.validateV2Request(&req); err != nil {
			h.logger.WithError(err).Errorf("V2 bulk request item %d validation failed", i)
			return h.errorHandler.HandleBusinessLogicError(c, ErrorCodeInvalidRequestType, "Invalid request data in bulk request", err)
		}
	}

	// Call V2 orchestrator
	response, err := h.v2Orchestrator.BulkCheckServiceability(c.Context(), &request)
	if err != nil {
		return h.errorHandler.HandleBulkServiceError(c, ErrorCodeBulkServiceabilityCheckFailed, "Failed to process V2 bulk serviceability check", err)
	}

	// Return response with appropriate status code
	statusCode := fiber.StatusOK
	if !response.Success {
		statusCode = fiber.StatusMultiStatus
	}

	return c.Status(statusCode).JSON(response)
}

// GetHealthV2 handles GET /health (V2 health check)
func (h *ServiceabilityV2Handler) GetHealthV2(c *fiber.Ctx) error {
	h.logger.Debug("V2 health check requested")

	// Create health response
	healthResponse := &modelsv1.ServiceabilityV2Response{
		Success: true,
		Partners: []modelsv1.PartnerV2Response{
			{
				PartnerID:     "system",
				PartnerName:   "V2 System Health",
				Rating:        5.0,
				IsServiceable: true,
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
		Metadata: &modelsv1.V2ResponseMetadata{
			TotalPartners:    1,
			ServiceableCount: 1,
			Filters: modelsv1.V2Filters{
				CountryCode: "IN",
			},
			EligiblePartners: []string{"system"},
		},
	}

	return c.Status(fiber.StatusOK).JSON(healthResponse)
}

// validateV2Request validates V2 serviceability request data
func (h *ServiceabilityV2Handler) validateV2Request(request *modelsv1.ServiceabilityV2Request) error {
	// Validate pickup postal code format if provided
	if request.PickupPostalCode != nil && *request.PickupPostalCode != "" {
		// Basic validation - just check if it's not empty
		// TODO: Add proper postal code format validation
	}

	// Validate delivery postal code format if provided
	if request.DeliveryPostalCode != nil && *request.DeliveryPostalCode != "" {
		// Basic validation - just check if it's not empty
		// TODO: Add proper postal code format validation
	}

	// Validate country code
	if request.CountryCode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Country code is required")
	}

	// Validate parcel category if provided
	if request.ParcelCategory != nil && *request.ParcelCategory != "" {
		validCategories := []string{"ecomm", "courier", "cargo", "international", "hyperlocal"}
		isValidCategory := false
		for _, category := range validCategories {
			if *request.ParcelCategory == category {
				isValidCategory = true
				break
			}
		}
		if !isValidCategory {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid parcel category")
		}
	}

	return nil
}
