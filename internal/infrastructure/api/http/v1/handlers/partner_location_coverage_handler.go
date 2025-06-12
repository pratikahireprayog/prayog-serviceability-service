package handlers

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/services/v1"
)

// PartnerLocationCoverageHandler handles partner location coverage HTTP requests
type PartnerLocationCoverageHandler struct {
	service   services.PartnerLocationCoverageService
	validator *validator.Validate
	logger    *logrus.Logger
}

// NewPartnerLocationCoverageHandler creates a new partner location coverage handler
func NewPartnerLocationCoverageHandler(
	service services.PartnerLocationCoverageService,
	validator *validator.Validate,
	logger *logrus.Logger,
) *PartnerLocationCoverageHandler {
	return &PartnerLocationCoverageHandler{
		service:   service,
		validator: validator,
		logger:    logger,
	}
}

// Helper function to parse partner ID from URL parameters
func (h *PartnerLocationCoverageHandler) parsePartnerID(c *fiber.Ctx) (uuid.UUID, error) {
	partnerIDStr := c.Params("partner_id")
	if strings.TrimSpace(partnerIDStr) == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "Partner ID is required")
	}

	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "Invalid partner ID format")
	}

	return partnerID, nil
}

// Helper function to parse coverage ID from URL parameters
func (h *PartnerLocationCoverageHandler) parseCoverageID(c *fiber.Ctx) (uuid.UUID, error) {
	coverageIDStr := c.Params("coverage_id")
	if strings.TrimSpace(coverageIDStr) == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "Coverage ID is required")
	}

	coverageID, err := uuid.Parse(coverageIDStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "Invalid coverage ID format")
	}

	return coverageID, nil
}

// Helper function to parse filters from query parameters
func (h *PartnerLocationCoverageHandler) parseFilters(c *fiber.Ctx) *dtos.PartnerLocationCoverageFiltersRequest {
	filters := &dtos.PartnerLocationCoverageFiltersRequest{}

	// Parse location_scope
	if locationScope := c.Query("location_scope"); locationScope != "" {
		filters.LocationScope = &locationScope
	}

	// Parse location_id
	if locationIDStr := c.Query("location_id"); locationIDStr != "" {
		if locationID, err := uuid.Parse(locationIDStr); err == nil {
			filters.LocationID = &locationID
		}
	}

	// Parse zone_type
	if zoneType := c.Query("zone_type"); zoneType != "" {
		filters.ZoneType = &zoneType
	}

	// Parse source_postal_code
	if sourcePostalCode := c.Query("source_postal_code"); sourcePostalCode != "" {
		filters.SourcePostalCode = &sourcePostalCode
	}

	// Parse destination_postal_code
	if destinationPostalCode := c.Query("destination_postal_code"); destinationPostalCode != "" {
		filters.DestinationPostalCode = &destinationPostalCode
	}

	// Parse source_postal_code_id
	if sourcePostalCodeIDStr := c.Query("source_postal_code_id"); sourcePostalCodeIDStr != "" {
		if sourcePostalCodeID, err := uuid.Parse(sourcePostalCodeIDStr); err == nil {
			filters.SourcePostalCodeID = &sourcePostalCodeID
		}
	}

	// Parse destination_postal_code_id
	if destinationPostalCodeIDStr := c.Query("destination_postal_code_id"); destinationPostalCodeIDStr != "" {
		if destinationPostalCodeID, err := uuid.Parse(destinationPostalCodeIDStr); err == nil {
			filters.DestinationPostalCodeID = &destinationPostalCodeID
		}
	}

	// Parse is_active
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActiveStr == "true" {
			isActive := true
			filters.IsActive = &isActive
		} else if isActiveStr == "false" {
			isActive := false
			filters.IsActive = &isActive
		}
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit := parseInt(limitStr, 10); limit > 0 && limit <= 100 {
			filters.Limit = &limit
		}
	}
	if filters.Limit == nil {
		limit := 10
		filters.Limit = &limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset := parseInt(offsetStr, 0); offset >= 0 {
			filters.Offset = &offset
		}
	}
	if filters.Offset == nil {
		offset := 0
		filters.Offset = &offset
	}

	return filters
}

// Helper function to handle errors consistently
func (h *PartnerLocationCoverageHandler) handleError(c *fiber.Ctx, err error, operation string) error {
	errMsg := err.Error()

	// Log the error with context
	h.logger.WithFields(logrus.Fields{
		"operation": operation,
		"error":     errMsg,
		"path":      c.Path(),
		"method":    c.Method(),
	}).Error("Partner location coverage handler error")

	// Handle specific error types
	if strings.Contains(errMsg, "partner validation failed") {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid partner",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_PARTNER",
				Message: "Partner not found or inactive",
				Details: errMsg,
			},
		})
	}

	if strings.Contains(errMsg, "not found") {
		return c.Status(fiber.StatusNotFound).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Resource not found",
			Error: dtos.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: errMsg,
			},
		})
	}

	if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "cannot be empty") || strings.Contains(errMsg, "validation failed") {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: errMsg,
			},
		})
	}

	if strings.Contains(errMsg, "already exists") {
		return c.Status(fiber.StatusConflict).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Resource conflict",
			Error: dtos.ErrorInfo{
				Code:    "CONFLICT",
				Message: errMsg,
			},
		})
	}

	// Default to internal server error
	return c.Status(fiber.StatusInternalServerError).JSON(dtos.StandardErrorResponse{
		Success: false,
		Message: "Internal server error",
		Error: dtos.ErrorInfo{
			Code:    "INTERNAL_ERROR",
			Message: errMsg,
		},
	})
}

// Helper function to validate request body
func (h *PartnerLocationCoverageHandler) validateRequest(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Failed to parse request body",
				Details: err.Error(),
			},
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Request validation failed",
				Details: err.Error(),
			},
		})
	}

	return nil
}

// GetPartnerLocationCoverages retrieves all location coverages for a partner
// GET /partners/{partner_id}/location-coverages
func (h *PartnerLocationCoverageHandler) GetPartnerLocationCoverages(c *fiber.Ctx) error {
	partnerID, err := h.parsePartnerID(c)
	if err != nil {
		return err
	}

	filters := h.parseFilters(c)

	coverages, err := h.service.GetByPartnerID(c.Context(), partnerID, filters)
	if err != nil {
		return h.handleError(c, err, "GetPartnerLocationCoverages")
	}

	return c.JSON(coverages)
}

// GetPartnerLocationCoverage retrieves a specific location coverage for a partner
// GET /partners/{partner_id}/location-coverages/{coverage_id}
func (h *PartnerLocationCoverageHandler) GetPartnerLocationCoverage(c *fiber.Ctx) error {
	partnerID, err := h.parsePartnerID(c)
	if err != nil {
		return err
	}

	coverageID, err := h.parseCoverageID(c)
	if err != nil {
		return err
	}

	coverage, err := h.service.GetByID(c.Context(), partnerID, coverageID)
	if err != nil {
		return h.handleError(c, err, "GetPartnerLocationCoverage")
	}

	return c.JSON(coverage)
}

// CreatePartnerLocationCoverage creates a new location coverage for a partner
// POST /partners/{partner_id}/location-coverages
func (h *PartnerLocationCoverageHandler) CreatePartnerLocationCoverage(c *fiber.Ctx) error {
	partnerID, err := h.parsePartnerID(c)
	if err != nil {
		return err
	}

	var req dtos.CreatePartnerLocationCoverageRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	coverage, err := h.service.Create(c.Context(), partnerID, &req)
	if err != nil {
		return h.handleError(c, err, "CreatePartnerLocationCoverage")
	}

	return c.Status(fiber.StatusCreated).JSON(coverage)
}

// UpdatePartnerLocationCoverage updates an existing location coverage for a partner
// PUT /partners/{partner_id}/location-coverages/{coverage_id}
func (h *PartnerLocationCoverageHandler) UpdatePartnerLocationCoverage(c *fiber.Ctx) error {
	partnerID, err := h.parsePartnerID(c)
	if err != nil {
		return err
	}

	coverageID, err := h.parseCoverageID(c)
	if err != nil {
		return err
	}

	var req dtos.UpdatePartnerLocationCoverageRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	coverage, err := h.service.Update(c.Context(), partnerID, coverageID, &req)
	if err != nil {
		return h.handleError(c, err, "UpdatePartnerLocationCoverage")
	}

	return c.JSON(coverage)
}

// DeletePartnerLocationCoverage deletes a location coverage for a partner
// DELETE /partners/{partner_id}/location-coverages/{coverage_id}
func (h *PartnerLocationCoverageHandler) DeletePartnerLocationCoverage(c *fiber.Ctx) error {
	partnerID, err := h.parsePartnerID(c)
	if err != nil {
		return err
	}

	coverageID, err := h.parseCoverageID(c)
	if err != nil {
		return err
	}

	err = h.service.Delete(c.Context(), partnerID, coverageID)
	if err != nil {
		return h.handleError(c, err, "DeletePartnerLocationCoverage")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// BulkCreatePartnerLocationCoverages creates multiple location coverages for a partner
// POST /partners/{partner_id}/location-coverages/bulk
func (h *PartnerLocationCoverageHandler) BulkCreatePartnerLocationCoverages(c *fiber.Ctx) error {
	partnerID, err := h.parsePartnerID(c)
	if err != nil {
		return err
	}

	var req dtos.BulkCreatePartnerLocationCoverageRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	result, err := h.service.BulkCreate(c.Context(), partnerID, &req)
	if err != nil {
		return h.handleError(c, err, "BulkCreatePartnerLocationCoverages")
	}

	// Return 207 Multi-Status if there were any failures
	if len(result.Failed) > 0 {
		return c.Status(fiber.StatusMultiStatus).JSON(result)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// Helper function to parse integer with default value
func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}

	// Simple integer parsing - in production, you'd want proper error handling
	result := defaultValue
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			return defaultValue
		}
	}
	return result
}
