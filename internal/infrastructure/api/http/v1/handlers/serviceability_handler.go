package handlers

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/shared/constants/v1"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/interfaces/v1"
	"prayog-serviceability-service/internal/shared/services/v1"
	"prayog-serviceability-service/internal/shared/utils/v1"
)

// ServiceabilityHandler handles serviceability check requests
type ServiceabilityHandler struct {
	orchestrator                    interfaces.ServiceabilityOrchestrator
	postalCodeServiceabilityService services.PostalCodeServiceabilityService
	validator                       *validator.Validate
	logger                          *logrus.Logger
	errorHandler                    *ErrorHandler
	postalCodeValidator             *utils.PostalCodeValidator
}

// NewServiceabilityHandler creates a new serviceability handler
func NewServiceabilityHandler(
	orchestrator interfaces.ServiceabilityOrchestrator,
	postalCodeServiceabilityService services.PostalCodeServiceabilityService,
	validator *validator.Validate,
	logger *logrus.Logger,
) *ServiceabilityHandler {
	return &ServiceabilityHandler{
		orchestrator:                    orchestrator,
		postalCodeServiceabilityService: postalCodeServiceabilityService,
		validator:                       validator,
		logger:                          logger,
		errorHandler:                    NewErrorHandler(logger),
		postalCodeValidator:             utils.NewPostalCodeValidator(),
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

// NEW POSTAL CODE BASED SERVICEABILITY ENDPOINTS

// CheckPostalCodeServiceability handles GET /check/{postal_code} (single postal code serviceability check)
func (h *ServiceabilityHandler) CheckPostalCodeServiceability(c *fiber.Ctx) error {
	h.logger.Debug("Single postal code serviceability check requested")

	// 1. Validate and sanitize postal code from URL parameter
	postalCode := c.Params("postal_code")
	if err := h.validatePostalCodeParam(postalCode); err != nil {
		return h.errorHandler.HandleValidationError(c, err)
	}

	// Normalize postal code
	postalCode = h.normalizePostalCode(postalCode)

	// 2. Parse and validate query parameters
	filters, err := h.parseAndValidateQueryParams(c)
	if err != nil {
		return h.errorHandler.HandleValidationError(c, err)
	}

	// 3. Validate postal code format (using default country IN for now)
	if err := h.postalCodeValidator.ValidatePostalCode(postalCode, "IN"); err != nil {
		return h.errorHandler.HandleBusinessLogicError(c, ErrorCodeInvalidPostalCode, "Invalid postal code format", err)
	}

	// 4. Apply struct validation to filters
	if err := h.validator.Struct(filters); err != nil {
		return h.errorHandler.HandleValidationError(c, err)
	}

	// 5. Call serviceability service
	response, err := h.postalCodeServiceabilityService.GetServiceabilityByPostalCode(c.Context(), postalCode, filters)
	if err != nil {
		return h.errorHandler.HandleServiceError(c, ErrorCodeServiceabilityCheckFailed, "Failed to check postal code serviceability", err)
	}

	// 6. Return appropriate status code based on response
	statusCode := fiber.StatusOK
	if !response.Success {
		if response.Error != nil {
			switch response.Error.Code {
			case "POSTAL_CODE_NOT_FOUND":
				statusCode = fiber.StatusNotFound
			case "POSTAL_CODE_NOT_SERVICEABLE":
				statusCode = fiber.StatusOK // This is a valid business response
			default:
				statusCode = fiber.StatusInternalServerError
			}
		} else {
			statusCode = fiber.StatusInternalServerError
		}
	}

	return c.Status(statusCode).JSON(response)
}

// validatePostalCodeParam validates the postal code URL parameter
func (h *ServiceabilityHandler) validatePostalCodeParam(postalCode string) error {
	// Check if postal code is empty
	if postalCode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Postal code is required")
	}

	// Check length constraints
	if len(postalCode) < constants.MinPostalCodeLength {
		return fiber.NewError(fiber.StatusBadRequest, "Postal code must be at least 3 characters long")
	}

	if len(postalCode) > constants.MaxPostalCodeLength {
		return fiber.NewError(fiber.StatusBadRequest, "Postal code cannot exceed 20 characters")
	}

	// Check for valid characters (alphanumeric, spaces, hyphens only)
	if !h.isValidPostalCodeChars(postalCode) {
		return fiber.NewError(fiber.StatusBadRequest, "Postal code contains invalid characters")
	}

	return nil
}

// parseAndValidateQueryParams parses and validates query parameters
func (h *ServiceabilityHandler) parseAndValidateQueryParams(c *fiber.Ctx) (*dtos.PostalCodeServiceabilityRequest, error) {
	filters := &dtos.PostalCodeServiceabilityRequest{}

	// Parse and validate parcel_category
	if parcelCategory := c.Query("parcel_category"); parcelCategory != "" {
		// Sanitize input
		parcelCategory = strings.TrimSpace(strings.ToLower(parcelCategory))

		// Validate enum values
		if !h.isValidParcelCategory(parcelCategory) {
			return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid parcel_category. Must be one of: ecomm, cargo, courier")
		}
		filters.ParcelCategory = &parcelCategory
	}

	// Parse and validate product_type
	if productType := c.Query("product_type"); productType != "" {
		// Sanitize input
		productType = strings.TrimSpace(productType)

		// Basic validation - not empty after trimming
		if productType == "" {
			return nil, fiber.NewError(fiber.StatusBadRequest, "Product type cannot be empty")
		}

		// Length validation
		if len(productType) > 50 {
			return nil, fiber.NewError(fiber.StatusBadRequest, "Product type cannot exceed 50 characters")
		}

		// Character validation - alphanumeric, underscore, hyphen only
		if !h.isValidProductType(productType) {
			return nil, fiber.NewError(fiber.StatusBadRequest, "Product type contains invalid characters")
		}

		filters.ProductType = &productType
	}

	return filters, nil
}

// normalizePostalCode normalizes postal code format
func (h *ServiceabilityHandler) normalizePostalCode(postalCode string) string {
	stringUtils := utils.StringUtils{}
	return stringUtils.NormalizePostalCode(postalCode)
}

// isValidPostalCodeChars checks if postal code contains only valid characters
func (h *ServiceabilityHandler) isValidPostalCodeChars(postalCode string) bool {
	for _, char := range postalCode {
		if !((char >= '0' && char <= '9') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			char == ' ' || char == '-') {
			return false
		}
	}
	return true
}

// isValidParcelCategory validates parcel category enum values
func (h *ServiceabilityHandler) isValidParcelCategory(category string) bool {
	validCategories := map[string]bool{
		constants.ParcelCategoryEcom:    true, // "ecom"
		constants.ParcelCategoryCourier: true, // "courier"
		constants.ParcelCategoryCargo:   true, // "cargo"
		"ecomm":                         true, // Alternative spelling
	}
	return validCategories[category]
}

// isValidProductType validates product type format
func (h *ServiceabilityHandler) isValidProductType(productType string) bool {
	for _, char := range productType {
		if !((char >= '0' && char <= '9') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			char == '_' || char == '-') {
			return false
		}
	}
	return true
}

// CheckPostalCodeServiceabilityPost handles POST /check (source and destination postal code serviceability check)
func (h *ServiceabilityHandler) CheckPostalCodeServiceabilityPost(c *fiber.Ctx) error {
	h.logger.Debug("Postal code serviceability check with source/destination requested")

	// Parse request body
	var requestDTO dtos.PostalCodeServiceabilityRequest
	if err := c.BodyParser(&requestDTO); err != nil {
		return h.errorHandler.HandleParsingError(c, err)
	}

	// Validate request
	if err := h.validator.Struct(&requestDTO); err != nil {
		return h.errorHandler.HandleValidationError(c, err)
	}

	// Call serviceability service
	response, err := h.postalCodeServiceabilityService.CheckServiceability(c.Context(), &requestDTO)
	if err != nil {
		return h.errorHandler.HandleServiceError(c, ErrorCodeServiceabilityCheckFailed, "Failed to check serviceability", err)
	}

	// Return appropriate status code based on response
	statusCode := fiber.StatusOK
	if !response.Success {
		if response.Error != nil && response.Error.Code == "POSTAL_CODE_NOT_FOUND" {
			statusCode = fiber.StatusNotFound
		} else {
			statusCode = fiber.StatusInternalServerError
		}
	}

	return c.Status(statusCode).JSON(response)
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
