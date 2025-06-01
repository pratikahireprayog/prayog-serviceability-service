package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
)

// ServiceabilityHandler handles serviceability check requests
type ServiceabilityHandler struct {
	orchestrator interfaces.ServiceabilityOrchestrator
	validator    *validator.Validate
	logger       *logrus.Logger
	errorHandler *ErrorHandler
}

// NewServiceabilityHandler creates a new serviceability handler
func NewServiceabilityHandler(
	orchestrator interfaces.ServiceabilityOrchestrator,
	validator *validator.Validate,
	logger *logrus.Logger,
) *ServiceabilityHandler {
	return &ServiceabilityHandler{
		orchestrator: orchestrator,
		validator:    validator,
		logger:       logger,
		errorHandler: NewErrorHandler(logger),
	}
}

// CheckServiceability handles POST /check (single serviceability check)
func (h *ServiceabilityHandler) CheckServiceability(c *fiber.Ctx) error {
	h.logger.Debug("Single serviceability check requested")

	// Parse request body
	var requestDTO dtos.ServiceabilityCheckRequestDTO
	if err := c.BodyParser(&requestDTO); err != nil {
		return h.errorHandler.HandleParsingError(c, err)
	}

	// Validate request
	if err := h.validator.Struct(&requestDTO); err != nil {
		return h.errorHandler.HandleValidationError(c, err)
	}

	// Additional validation for request type
	if err := h.validateRequestType(&requestDTO); err != nil {
		return h.errorHandler.HandleBusinessLogicError(c, ErrorCodeInvalidRequestType, "Invalid request type", err)
	}

	// Convert DTO to model
	requestModel := requestDTO.ToModel()

	// Call orchestrator
	responseModel, err := h.orchestrator.CheckServiceability(c.Context(), requestModel)
	if err != nil {
		return h.errorHandler.HandleServiceError(c, ErrorCodeServiceabilityCheckFailed, "Failed to check serviceability", err)
	}

	// Convert model response to DTO
	responseDTO := dtos.ServiceabilityResponseDTOFromModel(responseModel)

	// Return appropriate status code based on response
	statusCode := fiber.StatusOK
	if !responseDTO.Success {
		if responseDTO.Error != nil && responseDTO.Error.Code == "POSTAL_CODE_INACTIVE" {
			statusCode = fiber.StatusUnprocessableEntity
		} else {
			statusCode = fiber.StatusInternalServerError
		}
	}

	return c.Status(statusCode).JSON(responseDTO)
}

// BulkCheckServiceability handles POST /bulk-check (bulk serviceability check)
func (h *ServiceabilityHandler) BulkCheckServiceability(c *fiber.Ctx) error {
	h.logger.Debug("Bulk serviceability check requested")

	// Parse request body
	var requestDTO dtos.BulkServiceabilityRequestDTO
	if err := c.BodyParser(&requestDTO); err != nil {
		return h.errorHandler.HandleParsingError(c, err)
	}

	// Validate request
	if err := h.validator.Struct(&requestDTO); err != nil {
		return h.errorHandler.HandleValidationError(c, err)
	}

	// Validate each individual request in the bulk
	for i, req := range requestDTO.Requests {
		if err := h.validateRequestType(&req); err != nil {
			h.logger.WithError(err).Errorf("Bulk request item %d validation failed", i)
			return h.errorHandler.HandleBusinessLogicError(c, ErrorCodeInvalidRequestType, "Invalid request type in bulk request", err)
		}
	}

	// Convert DTO to model
	requestModel := requestDTO.ToModel()

	// Call orchestrator
	responseModel, err := h.orchestrator.BulkCheckServiceability(c.Context(), requestModel)
	if err != nil {
		return h.errorHandler.HandleBulkServiceError(c, ErrorCodeBulkServiceabilityCheckFailed, "Failed to process bulk serviceability check", err)
	}

	// Convert model response to DTO
	responseDTO := dtos.BulkServiceabilityResponseDTOFromModel(responseModel)

	// Return appropriate status code
	statusCode := fiber.StatusOK
	if !responseDTO.Success {
		if responseDTO.Error != nil && responseDTO.Error.Code == "PARTIAL_FAILURE" {
			statusCode = fiber.StatusMultiStatus
		} else {
			statusCode = fiber.StatusInternalServerError
		}
	}

	return c.Status(statusCode).JSON(responseDTO)
}

// validateRequestType validates that the request has either postal_code OR both pickup/delivery postal codes
func (h *ServiceabilityHandler) validateRequestType(req *dtos.ServiceabilityCheckRequestDTO) error {
	hasPostalCode := req.PostalCode != nil && *req.PostalCode != ""
	hasPickupCode := req.PickupPostalCode != nil && *req.PickupPostalCode != ""
	hasDeliveryCode := req.DeliveryPostalCode != nil && *req.DeliveryPostalCode != ""

	// Must have either postal_code OR both pickup and delivery postal codes
	if hasPostalCode {
		// Single location query - shouldn't have pickup/delivery codes
		if hasPickupCode || hasDeliveryCode {
			return fiber.NewError(fiber.StatusBadRequest, "Cannot specify both postal_code and pickup/delivery postal codes")
		}
		return nil
	}

	// Origin-destination query - must have both pickup and delivery codes
	if hasPickupCode && hasDeliveryCode {
		return nil
	}

	// Invalid: missing required fields
	if !hasPickupCode && !hasDeliveryCode {
		return fiber.NewError(fiber.StatusBadRequest, "Must specify either postal_code or both pickup_postal_code and delivery_postal_code")
	}

	// Invalid: only one of pickup/delivery specified
	return fiber.NewError(fiber.StatusBadRequest, "Both pickup_postal_code and delivery_postal_code are required for origin-destination queries")
}
