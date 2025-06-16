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

// Helper function to parse partner code from URL parameters
func (h *PartnerLocationCoverageHandler) parsePartnerCode(c *fiber.Ctx) (string, error) {
	partnerCode := strings.TrimSpace(c.Params("partner_code"))
	if partnerCode == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, "Partner code is required")
	}
	return partnerCode, nil
}

// Helper function to parse postal code from URL parameters
func (h *PartnerLocationCoverageHandler) parsePostalCode(c *fiber.Ctx) (string, error) {
	postalCode := strings.TrimSpace(c.Params("postal_code"))
	if postalCode == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, "Postal code is required")
	}
	return postalCode, nil
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

	// Parse postal_code
	if postalCode := c.Query("postal_code"); postalCode != "" {
		filters.PostalCode = &postalCode
	}

	// Parse postal_code_id
	if postalCodeIDStr := c.Query("postal_code_id"); postalCodeIDStr != "" {
		if postalCodeID, err := uuid.Parse(postalCodeIDStr); err == nil {
			filters.PostalCodeID = &postalCodeID
		}
	}

	// Parse zone_type
	if zoneType := c.Query("zone_type"); zoneType != "" {
		filters.ZoneType = &zoneType
	}

	// Parse country_code
	if countryCode := c.Query("country_code"); countryCode != "" {
		filters.CountryCode = &countryCode
	}

	// Parse product_type
	if productType := c.Query("product_type"); productType != "" {
		filters.ProductType = &productType
	}

	// Parse parcel_category
	if parcelCategory := c.Query("parcel_category"); parcelCategory != "" {
		filters.ParcelCategory = &parcelCategory
	}

	// Parse service_type
	if serviceType := c.Query("service_type"); serviceType != "" {
		filters.ServiceType = &serviceType
	}

	// Parse tat_days
	if tatDaysStr := c.Query("tat_days"); tatDaysStr != "" {
		if tatDays := parseInt(tatDaysStr, -1); tatDays >= 0 {
			filters.TATDays = &tatDays
		}
	}

	// Parse pickup
	if pickupStr := c.Query("pickup"); pickupStr != "" {
		if pickupStr == "true" {
			pickup := true
			filters.Pickup = &pickup
		} else if pickupStr == "false" {
			pickup := false
			filters.Pickup = &pickup
		}
	}

	// Parse delivery
	if deliveryStr := c.Query("delivery"); deliveryStr != "" {
		if deliveryStr == "true" {
			delivery := true
			filters.Delivery = &delivery
		} else if deliveryStr == "false" {
			delivery := false
			filters.Delivery = &delivery
		}
	}

	// Parse delivery_mode
	if deliveryMode := c.Query("delivery_mode"); deliveryMode != "" {
		filters.DeliveryMode = &deliveryMode
	}

	// Parse cod_available
	if codAvailableStr := c.Query("cod_available"); codAvailableStr != "" {
		if codAvailableStr == "true" {
			codAvailable := true
			filters.CODAvailable = &codAvailable
		} else if codAvailableStr == "false" {
			codAvailable := false
			filters.CODAvailable = &codAvailable
		}
	}

	// Parse insurance
	if insuranceStr := c.Query("insurance"); insuranceStr != "" {
		if insuranceStr == "true" {
			insurance := true
			filters.Insurance = &insurance
		} else if insuranceStr == "false" {
			insurance := false
			filters.Insurance = &insurance
		}
	}

	// Parse min_weight_kg
	if minWeightStr := c.Query("min_weight_kg"); minWeightStr != "" {
		if minWeight := parseFloat(minWeightStr, -1); minWeight >= 0 {
			filters.MinWeightKG = &minWeight
		}
	}

	// Parse max_weight_kg
	if maxWeightStr := c.Query("max_weight_kg"); maxWeightStr != "" {
		if maxWeight := parseFloat(maxWeightStr, -1); maxWeight >= 0 {
			filters.MaxWeightKG = &maxWeight
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
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Failed to parse request body")

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

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Request validation failed")

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
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Failed to parse request body")

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

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Request validation failed")

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
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Failed to parse request body")

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

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Request validation failed")

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

// NEW POSTAL CODE BASED METHODS FOR PARTNER ID

// GetPartnerLocationCoverageByPostalCode retrieves a specific location coverage by partner ID and postal code
// GET /partners/{partner_id}/location-coverages/postal-code/{postal_code}
func (h *PartnerLocationCoverageHandler) GetPartnerLocationCoverageByPostalCode(c *fiber.Ctx) error {
	partnerID, err := h.parsePartnerID(c)
	if err != nil {
		return err
	}

	postalCode, err := h.parsePostalCode(c)
	if err != nil {
		return err
	}

	coverages, err := h.service.GetByPartnerIDAndPostalCode(c.Context(), partnerID, postalCode)
	if err != nil {
		return h.handleError(c, err, "GetPartnerLocationCoverageByPostalCode")
	}

	return c.JSON(coverages)
}

// UpdatePartnerLocationCoverageByPostalCode updates a location coverage by partner ID and postal code
// PUT /partners/{partner_id}/location-coverages/postal-code/{postal_code}
func (h *PartnerLocationCoverageHandler) UpdatePartnerLocationCoverageByPostalCode(c *fiber.Ctx) error {
	partnerID, err := h.parsePartnerID(c)
	if err != nil {
		return err
	}

	postalCode, err := h.parsePostalCode(c)
	if err != nil {
		return err
	}

	var req dtos.UpdatePartnerLocationCoverageRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Failed to parse request body")

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

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Request validation failed")

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

	coverage, err := h.service.UpdateByPartnerIDAndPostalCode(c.Context(), partnerID, postalCode, &req)
	if err != nil {
		return h.handleError(c, err, "UpdatePartnerLocationCoverageByPostalCode")
	}

	return c.JSON(coverage)
}

// DeletePartnerLocationCoverageByPostalCode deletes a location coverage by partner ID and postal code
// DELETE /partners/{partner_id}/location-coverages/postal-code/{postal_code}
func (h *PartnerLocationCoverageHandler) DeletePartnerLocationCoverageByPostalCode(c *fiber.Ctx) error {
	partnerID, err := h.parsePartnerID(c)
	if err != nil {
		return err
	}

	postalCode, err := h.parsePostalCode(c)
	if err != nil {
		return err
	}

	err = h.service.DeleteByPartnerIDAndPostalCode(c.Context(), partnerID, postalCode)
	if err != nil {
		return h.handleError(c, err, "DeletePartnerLocationCoverageByPostalCode")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// CheckPartnerLocationCoverage checks if a partner covers a specific postal code with optional service requirements
// GET /partners/{partner_id}/location-coverages/postal-code/{postal_code}/check
func (h *PartnerLocationCoverageHandler) CheckPartnerLocationCoverage(c *fiber.Ctx) error {
	partnerID, err := h.parsePartnerID(c)
	if err != nil {
		return err
	}

	postalCode, err := h.parsePostalCode(c)
	if err != nil {
		return err
	}

	// Parse service requirements from query parameters
	filters := h.parseFilters(c)

	result, err := h.service.CheckCoverage(c.Context(), partnerID, postalCode, filters)
	if err != nil {
		return h.handleError(c, err, "CheckPartnerLocationCoverage")
	}

	return c.JSON(result)
}

// NEW PARTNER CODE BASED METHODS

// GetPartnerLocationCoveragesByCode retrieves all location coverages for a partner by partner code
// GET /partners/code/{partner_code}/location-coverages
func (h *PartnerLocationCoverageHandler) GetPartnerLocationCoveragesByCode(c *fiber.Ctx) error {
	partnerCode, err := h.parsePartnerCode(c)
	if err != nil {
		return err
	}

	filters := h.parseFilters(c)

	coverages, err := h.service.GetByPartnerCode(c.Context(), partnerCode, filters)
	if err != nil {
		return h.handleError(c, err, "GetPartnerLocationCoveragesByCode")
	}

	return c.JSON(coverages)
}

// CreatePartnerLocationCoverageByCode creates a new location coverage for a partner by partner code
// POST /partners/code/{partner_code}/location-coverages
func (h *PartnerLocationCoverageHandler) CreatePartnerLocationCoverageByCode(c *fiber.Ctx) error {
	partnerCode, err := h.parsePartnerCode(c)
	if err != nil {
		return err
	}

	var req dtos.CreatePartnerLocationCoverageRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Failed to parse request body")

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

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Request validation failed")

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

	coverage, err := h.service.CreateByPartnerCode(c.Context(), partnerCode, &req)
	if err != nil {
		return h.handleError(c, err, "CreatePartnerLocationCoverageByCode")
	}

	return c.Status(fiber.StatusCreated).JSON(coverage)
}

// BulkCreatePartnerLocationCoveragesByCode creates multiple location coverages for a partner by partner code
// POST /partners/code/{partner_code}/location-coverages/bulk
func (h *PartnerLocationCoverageHandler) BulkCreatePartnerLocationCoveragesByCode(c *fiber.Ctx) error {
	partnerCode, err := h.parsePartnerCode(c)
	if err != nil {
		return err
	}

	var req dtos.BulkCreatePartnerLocationCoverageRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Failed to parse request body")

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

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Request validation failed")

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

	result, err := h.service.BulkCreateByPartnerCode(c.Context(), partnerCode, &req)
	if err != nil {
		return h.handleError(c, err, "BulkCreatePartnerLocationCoveragesByCode")
	}

	// Return 207 Multi-Status if there were any failures
	if len(result.Failed) > 0 {
		return c.Status(fiber.StatusMultiStatus).JSON(result)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetPartnerLocationCoverageByCodeAndPostalCode retrieves a specific location coverage by partner code and postal code
// GET /partners/code/{partner_code}/location-coverages/postal-code/{postal_code}
func (h *PartnerLocationCoverageHandler) GetPartnerLocationCoverageByCodeAndPostalCode(c *fiber.Ctx) error {
	partnerCode, err := h.parsePartnerCode(c)
	if err != nil {
		return err
	}

	postalCode, err := h.parsePostalCode(c)
	if err != nil {
		return err
	}

	coverages, err := h.service.GetByPartnerCodeAndPostalCode(c.Context(), partnerCode, postalCode)
	if err != nil {
		return h.handleError(c, err, "GetPartnerLocationCoverageByCodeAndPostalCode")
	}

	return c.JSON(coverages)
}

// UpdatePartnerLocationCoverageByCodeAndPostalCode updates a location coverage by partner code and postal code
// PUT /partners/code/{partner_code}/location-coverages/postal-code/{postal_code}
func (h *PartnerLocationCoverageHandler) UpdatePartnerLocationCoverageByCodeAndPostalCode(c *fiber.Ctx) error {
	partnerCode, err := h.parsePartnerCode(c)
	if err != nil {
		return err
	}

	postalCode, err := h.parsePostalCode(c)
	if err != nil {
		return err
	}

	var req dtos.UpdatePartnerLocationCoverageRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Failed to parse request body")

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

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		}).Error("Request validation failed")

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

	coverage, err := h.service.UpdateByPartnerCodeAndPostalCode(c.Context(), partnerCode, postalCode, &req)
	if err != nil {
		return h.handleError(c, err, "UpdatePartnerLocationCoverageByCodeAndPostalCode")
	}

	return c.JSON(coverage)
}

// DeletePartnerLocationCoverageByCodeAndPostalCode deletes a location coverage by partner code and postal code
// DELETE /partners/code/{partner_code}/location-coverages/postal-code/{postal_code}
func (h *PartnerLocationCoverageHandler) DeletePartnerLocationCoverageByCodeAndPostalCode(c *fiber.Ctx) error {
	partnerCode, err := h.parsePartnerCode(c)
	if err != nil {
		return err
	}

	postalCode, err := h.parsePostalCode(c)
	if err != nil {
		return err
	}

	err = h.service.DeleteByPartnerCodeAndPostalCode(c.Context(), partnerCode, postalCode)
	if err != nil {
		return h.handleError(c, err, "DeletePartnerLocationCoverageByCodeAndPostalCode")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// CheckPartnerLocationCoverageByCode checks if a partner covers a specific postal code with optional service requirements
// GET /partners/code/{partner_code}/location-coverages/postal-code/{postal_code}/check
func (h *PartnerLocationCoverageHandler) CheckPartnerLocationCoverageByCode(c *fiber.Ctx) error {
	partnerCode, err := h.parsePartnerCode(c)
	if err != nil {
		return err
	}

	postalCode, err := h.parsePostalCode(c)
	if err != nil {
		return err
	}

	// Parse service requirements from query parameters
	filters := h.parseFilters(c)

	result, err := h.service.CheckCoverageByPartnerCode(c.Context(), partnerCode, postalCode, filters)
	if err != nil {
		return h.handleError(c, err, "CheckPartnerLocationCoverageByCode")
	}

	return c.JSON(result)
}

// Helper function to parse float with default value
func parseFloat(s string, defaultValue float64) float64 {
	if s == "" {
		return defaultValue
	}

	// Simple float parsing - in production, you'd want proper error handling
	result := 0.0
	decimal := false
	decimalPlace := 1.0

	for _, char := range s {
		if char >= '0' && char <= '9' {
			if decimal {
				decimalPlace *= 10
				result += float64(char-'0') / decimalPlace
			} else {
				result = result*10 + float64(char-'0')
			}
		} else if char == '.' && !decimal {
			decimal = true
		} else {
			return defaultValue
		}
	}
	return result
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
