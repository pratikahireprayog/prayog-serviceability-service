package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/services/v1/data"
)

// PartnerAttributeHandler handles partner attribute HTTP requests
type PartnerAttributeHandler struct {
	attributeCategoryService   services.AttributeCategoryService
	attributeService           services.AttributeService
	partnerAttributeMapService services.PartnerAttributeMapService
	validator                  *validator.Validate
	logger                     *logrus.Logger
}

// NewPartnerAttributeHandler creates a new partner attribute handler
func NewPartnerAttributeHandler(
	attributeCategoryService services.AttributeCategoryService,
	attributeService services.AttributeService,
	partnerAttributeMapService services.PartnerAttributeMapService,
	validator *validator.Validate,
	logger *logrus.Logger,
) (*PartnerAttributeHandler, error) {
	// Validate all required dependencies are not nil
	if attributeCategoryService == nil {
		return nil, fmt.Errorf("attributeCategoryService cannot be nil")
	}
	if attributeService == nil {
		return nil, fmt.Errorf("attributeService cannot be nil")
	}
	if partnerAttributeMapService == nil {
		return nil, fmt.Errorf("partnerAttributeMapService cannot be nil")
	}
	if validator == nil {
		return nil, fmt.Errorf("validator cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	return &PartnerAttributeHandler{
		attributeCategoryService:   attributeCategoryService,
		attributeService:           attributeService,
		partnerAttributeMapService: partnerAttributeMapService,
		validator:                  validator,
		logger:                     logger,
	}, nil
}

// Helper function to parse pagination parameters
func (h *PartnerAttributeHandler) parsePagination(c *fiber.Ctx) *dtos.PaginationRequest {
	offsetStr := c.Query("offset", "0")
	limitStr := c.Query("limit", "10")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	// Validate and set defaults
	if offset < 0 {
		offset = 0
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	return &dtos.PaginationRequest{
		Offset: offset,
		Limit:  limit,
	}
}

// Helper function to handle errors consistently
func (h *PartnerAttributeHandler) handleError(c *fiber.Ctx, err error, operation string) error {
	errMsg := err.Error()

	// Log the error with context
	h.logger.WithFields(logrus.Fields{
		"operation": operation,
		"error":     errMsg,
		"path":      c.Path(),
		"method":    c.Method(),
	}).Error("Partner attribute handler error")

	// Handle specific error types
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
			Message: "Invalid request",
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

	if strings.Contains(errMsg, "cannot delete") || strings.Contains(errMsg, "has active") {
		return c.Status(fiber.StatusConflict).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Cannot delete resource",
			Error: dtos.ErrorInfo{
				Code:    "DELETION_CONFLICT",
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
func (h *PartnerAttributeHandler) validateRequest(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
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

	if err := h.validator.Struct(req); err != nil {
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

	return nil
}

// ATTRIBUTE CATEGORY HANDLERS

// GetAttributeCategoryByID retrieves an attribute category by ID
// GET /attribute-categories/{id}
func (h *PartnerAttributeHandler) GetAttributeCategoryByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute category ID is required",
			},
		})
	}

	category, err := h.attributeCategoryService.GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "GetAttributeCategoryByID")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute category retrieved successfully",
		Data:    category,
	})
}

// GetAttributeCategoryByCode retrieves an attribute category by code
// GET /attribute-categories/code/{code}
func (h *PartnerAttributeHandler) GetAttributeCategoryByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute category code is required",
			},
		})
	}

	category, err := h.attributeCategoryService.GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err, "GetAttributeCategoryByCode")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute category retrieved successfully",
		Data:    category,
	})
}

// GetAllAttributeCategories retrieves all attribute categories with pagination
// GET /attribute-categories
func (h *PartnerAttributeHandler) GetAllAttributeCategories(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	categories, err := h.attributeCategoryService.GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err, "GetAllAttributeCategories")
	}

	return c.JSON(categories)
}

// CreateAttributeCategory creates a new attribute category
// POST /attribute-categories
func (h *PartnerAttributeHandler) CreateAttributeCategory(c *fiber.Ctx) error {
	var req dtos.CreateAttributeCategoryRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	category, err := h.attributeCategoryService.Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "CreateAttributeCategory")
	}

	return c.Status(fiber.StatusCreated).JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute category created successfully",
		Data:    category,
	})
}

// UpdateAttributeCategory updates an attribute category
// PUT /attribute-categories/{id}
func (h *PartnerAttributeHandler) UpdateAttributeCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute category ID is required",
			},
		})
	}

	var req dtos.UpdateAttributeCategoryRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	category, err := h.attributeCategoryService.Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err, "UpdateAttributeCategory")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute category updated successfully",
		Data:    category,
	})
}

// DeleteAttributeCategory soft deletes an attribute category
// DELETE /attribute-categories/{id}
func (h *PartnerAttributeHandler) DeleteAttributeCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute category ID is required",
			},
		})
	}

	err := h.attributeCategoryService.Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "DeleteAttributeCategory")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute category deleted successfully",
		Data:    fiber.Map{"id": id},
	})
}

// RestoreAttributeCategory restores a soft-deleted attribute category
// PATCH /attribute-categories/{id}/restore
func (h *PartnerAttributeHandler) RestoreAttributeCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute category ID is required",
			},
		})
	}

	err := h.attributeCategoryService.Restore(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "RestoreAttributeCategory")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute category restored successfully",
		Data:    fiber.Map{"id": id},
	})
}

// GetAttributeCategoryStats retrieves statistics for an attribute category
// GET /attribute-categories/{id}/stats
func (h *PartnerAttributeHandler) GetAttributeCategoryStats(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute category ID is required",
			},
		})
	}

	stats, err := h.attributeCategoryService.GetStats(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "GetAttributeCategoryStats")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute category statistics retrieved successfully",
		Data:    stats,
	})
}

// ATTRIBUTE HANDLERS

// GetAttributeByID retrieves an attribute by ID
// GET /attributes/{id}
func (h *PartnerAttributeHandler) GetAttributeByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute ID is required",
			},
		})
	}

	attribute, err := h.attributeService.GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "GetAttributeByID")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute retrieved successfully",
		Data:    attribute,
	})
}

// GetAttributeByCode retrieves an attribute by code
// GET /attributes/code/{code}
func (h *PartnerAttributeHandler) GetAttributeByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute code is required",
			},
		})
	}

	attribute, err := h.attributeService.GetByCode(c.Context(), code)
	if err != nil {
		return h.handleError(c, err, "GetAttributeByCode")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute retrieved successfully",
		Data:    attribute,
	})
}

// GetAllAttributes retrieves all attributes with pagination
// GET /attributes
func (h *PartnerAttributeHandler) GetAllAttributes(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	attributes, err := h.attributeService.GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err, "GetAllAttributes")
	}

	return c.JSON(attributes)
}

// GetAttributesByCategory retrieves attributes by category ID
// GET /attribute-categories/{category_id}/attributes
func (h *PartnerAttributeHandler) GetAttributesByCategory(c *fiber.Ctx) error {
	categoryID := c.Params("category_id")
	if strings.TrimSpace(categoryID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Category ID is required",
			},
		})
	}

	attributes, err := h.attributeService.GetByCategoryID(c.Context(), categoryID)
	if err != nil {
		return h.handleError(c, err, "GetAttributesByCategory")
	}

	return c.JSON(attributes)
}

// GetAttributesByCategoryCode retrieves attributes by category code
// GET /attribute-categories/code/{category_code}/attributes
func (h *PartnerAttributeHandler) GetAttributesByCategoryCode(c *fiber.Ctx) error {
	categoryCode := c.Params("category_code")
	if strings.TrimSpace(categoryCode) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Category code is required",
			},
		})
	}

	attributes, err := h.attributeService.GetByCategoryCode(c.Context(), categoryCode)
	if err != nil {
		return h.handleError(c, err, "GetAttributesByCategoryCode")
	}

	return c.JSON(attributes)
}

// CreateAttribute creates a new attribute
// POST /attributes
func (h *PartnerAttributeHandler) CreateAttribute(c *fiber.Ctx) error {
	var req dtos.CreateAttributeRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	attribute, err := h.attributeService.Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "CreateAttribute")
	}

	return c.Status(fiber.StatusCreated).JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute created successfully",
		Data:    attribute,
	})
}

// UpdateAttribute updates an attribute
// PUT /attributes/{id}
func (h *PartnerAttributeHandler) UpdateAttribute(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute ID is required",
			},
		})
	}

	var req dtos.UpdateAttributeRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	attribute, err := h.attributeService.Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err, "UpdateAttribute")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute updated successfully",
		Data:    attribute,
	})
}

// DeleteAttribute soft deletes an attribute
// DELETE /attributes/{id}
func (h *PartnerAttributeHandler) DeleteAttribute(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute ID is required",
			},
		})
	}

	err := h.attributeService.Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "DeleteAttribute")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute deleted successfully",
		Data:    fiber.Map{"id": id},
	})
}

// RestoreAttribute restores a soft-deleted attribute
// PATCH /attributes/{id}/restore
func (h *PartnerAttributeHandler) RestoreAttribute(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute ID is required",
			},
		})
	}

	err := h.attributeService.Restore(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "RestoreAttribute")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute restored successfully",
		Data:    fiber.Map{"id": id},
	})
}

// GetAttributeStats retrieves statistics for an attribute
// GET /attributes/{id}/stats
func (h *PartnerAttributeHandler) GetAttributeStats(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute ID is required",
			},
		})
	}

	stats, err := h.attributeService.GetStats(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "GetAttributeStats")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Attribute statistics retrieved successfully",
		Data:    stats,
	})
}

// PARTNER ATTRIBUTE MAP HANDLERS

// GetPartnerAttributeMapByID retrieves a partner attribute mapping by ID
// GET /partner-attribute-maps/{id}
func (h *PartnerAttributeHandler) GetPartnerAttributeMapByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Partner attribute mapping ID is required",
			},
		})
	}

	mapping, err := h.partnerAttributeMapService.GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "GetPartnerAttributeMapByID")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Partner attribute mapping retrieved successfully",
		Data:    mapping,
	})
}

// GetPartnerAttributeMapsByPartner retrieves all mappings for a partner
// GET /partners/{partner_code}/attributes
func (h *PartnerAttributeHandler) GetPartnerAttributeMapsByPartner(c *fiber.Ctx) error {
	partnerCode := c.Params("partner_code")
	if strings.TrimSpace(partnerCode) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Partner code is required",
			},
		})
	}

	mappings, err := h.partnerAttributeMapService.GetByPartnerCode(c.Context(), partnerCode)
	if err != nil {
		return h.handleError(c, err, "GetPartnerAttributeMapsByPartner")
	}

	return c.JSON(mappings)
}

// GetPartnerAttributeMapsByPartnerID retrieves all mappings for a partner ID
// GET /partners/id/{partner_id}/attributes
func (h *PartnerAttributeHandler) GetPartnerAttributeMapsByPartnerID(c *fiber.Ctx) error {
	partnerID := c.Params("partner_id")
	if strings.TrimSpace(partnerID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Partner ID is required",
			},
		})
	}

	mappings, err := h.partnerAttributeMapService.GetByPartnerID(c.Context(), partnerID)
	if err != nil {
		return h.handleError(c, err, "GetPartnerAttributeMapsByPartnerID")
	}

	return c.JSON(mappings)
}

// GetPartnerAttributeMapsByAttribute retrieves all mappings for an attribute
// GET /attributes/code/{attribute_code}/partners
func (h *PartnerAttributeHandler) GetPartnerAttributeMapsByAttribute(c *fiber.Ctx) error {
	attributeCode := c.Params("attribute_code")
	if strings.TrimSpace(attributeCode) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute code is required",
			},
		})
	}

	mappings, err := h.partnerAttributeMapService.GetByAttributeCode(c.Context(), attributeCode)
	if err != nil {
		return h.handleError(c, err, "GetPartnerAttributeMapsByAttribute")
	}

	return c.JSON(mappings)
}

// GetAllPartnerAttributeMaps retrieves all partner attribute mappings with pagination
// GET /partner-attribute-maps
func (h *PartnerAttributeHandler) GetAllPartnerAttributeMaps(c *fiber.Ctx) error {
	pagination := h.parsePagination(c)

	mappings, err := h.partnerAttributeMapService.GetAll(c.Context(), pagination)
	if err != nil {
		return h.handleError(c, err, "GetAllPartnerAttributeMaps")
	}

	return c.JSON(mappings)
}

// GetPartnerAttributeMapsWithFilters retrieves mappings with filtering
// POST /partner-attribute-maps/search
func (h *PartnerAttributeHandler) GetPartnerAttributeMapsWithFilters(c *fiber.Ctx) error {
	var filters dtos.PartnerAttributeMapFilterRequest
	if err := h.validateRequest(c, &filters); err != nil {
		return err
	}

	mappings, err := h.partnerAttributeMapService.GetWithFilters(c.Context(), &filters)
	if err != nil {
		return h.handleError(c, err, "GetPartnerAttributeMapsWithFilters")
	}

	return c.JSON(mappings)
}

// GetPartnerCodesByAttribute retrieves partner codes that have a specific attribute
// GET /attributes/code/{attribute_code}/partner-codes
func (h *PartnerAttributeHandler) GetPartnerCodesByAttribute(c *fiber.Ctx) error {
	attributeCode := c.Params("attribute_code")
	if strings.TrimSpace(attributeCode) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Attribute code is required",
			},
		})
	}

	response, err := h.partnerAttributeMapService.GetPartnerCodesByAttribute(c.Context(), attributeCode)
	if err != nil {
		return h.handleError(c, err, "GetPartnerCodesByAttribute")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Partner codes retrieved successfully",
		Data:    response,
	})
}

// GetAttributesByPartner retrieves attributes mapped to a partner
// GET /partners/{partner_code}/attributes/details
func (h *PartnerAttributeHandler) GetAttributesByPartner(c *fiber.Ctx) error {
	partnerCode := c.Params("partner_code")
	if strings.TrimSpace(partnerCode) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Partner code is required",
			},
		})
	}

	attributes, err := h.partnerAttributeMapService.GetAttributesByPartner(c.Context(), partnerCode)
	if err != nil {
		return h.handleError(c, err, "GetAttributesByPartner")
	}

	return c.JSON(attributes)
}

// CreatePartnerAttributeMap creates a new partner attribute mapping
// POST /partner-attribute-maps
func (h *PartnerAttributeHandler) CreatePartnerAttributeMap(c *fiber.Ctx) error {
	var req dtos.CreatePartnerAttributeMapRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	mapping, err := h.partnerAttributeMapService.Create(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "CreatePartnerAttributeMap")
	}

	return c.Status(fiber.StatusCreated).JSON(dtos.StandardResponse{
		Success: true,
		Message: "Partner attribute mapping created successfully",
		Data:    mapping,
	})
}

// UpdatePartnerAttributeMap updates a partner attribute mapping
// PUT /partner-attribute-maps/{id}
func (h *PartnerAttributeHandler) UpdatePartnerAttributeMap(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Partner attribute mapping ID is required",
			},
		})
	}

	var req dtos.UpdatePartnerAttributeMapRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	mapping, err := h.partnerAttributeMapService.Update(c.Context(), id, &req)
	if err != nil {
		return h.handleError(c, err, "UpdatePartnerAttributeMap")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Partner attribute mapping updated successfully",
		Data:    mapping,
	})
}

// DeletePartnerAttributeMap soft deletes a partner attribute mapping
// DELETE /partner-attribute-maps/{id}
func (h *PartnerAttributeHandler) DeletePartnerAttributeMap(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Partner attribute mapping ID is required",
			},
		})
	}

	err := h.partnerAttributeMapService.Delete(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "DeletePartnerAttributeMap")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Partner attribute mapping deleted successfully",
		Data:    fiber.Map{"id": id},
	})
}

// RestorePartnerAttributeMap restores a soft-deleted partner attribute mapping
// PATCH /partner-attribute-maps/{id}/restore
func (h *PartnerAttributeHandler) RestorePartnerAttributeMap(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Partner attribute mapping ID is required",
			},
		})
	}

	err := h.partnerAttributeMapService.Restore(c.Context(), id)
	if err != nil {
		return h.handleError(c, err, "RestorePartnerAttributeMap")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Partner attribute mapping restored successfully",
		Data:    fiber.Map{"id": id},
	})
}

// BulkCreatePartnerAttributeMaps creates multiple partner attribute mappings
// POST /partner-attribute-maps/bulk
func (h *PartnerAttributeHandler) BulkCreatePartnerAttributeMaps(c *fiber.Ctx) error {
	var req dtos.BulkCreatePartnerAttributeMapRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	result, err := h.partnerAttributeMapService.BulkCreate(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "BulkCreatePartnerAttributeMaps")
	}

	// Return 207 Multi-Status if there were any failures
	if len(result.Failed) > 0 {
		return c.Status(fiber.StatusMultiStatus).JSON(dtos.StandardResponse{
			Success: true,
			Message: "Bulk creation completed with some failures",
			Data:    result,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dtos.StandardResponse{
		Success: true,
		Message: "Partner attribute mappings created successfully",
		Data:    result,
	})
}

// BulkDeletePartnerAttributeMaps deletes multiple partner attribute mappings
// DELETE /partner-attribute-maps/bulk
func (h *PartnerAttributeHandler) BulkDeletePartnerAttributeMaps(c *fiber.Ctx) error {
	var req dtos.BulkDeletePartnerAttributeMapRequest
	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	result, err := h.partnerAttributeMapService.BulkDelete(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err, "BulkDeletePartnerAttributeMaps")
	}

	// Return 207 Multi-Status if there were any failures
	if len(result.Failed) > 0 {
		return c.Status(fiber.StatusMultiStatus).JSON(dtos.StandardResponse{
			Success: true,
			Message: "Bulk deletion completed with some failures",
			Data:    result,
		})
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Partner attribute mappings deleted successfully",
		Data:    result,
	})
}

// GetPartnerAttributeMapStats retrieves statistics for a partner's attribute mappings
// GET /partners/{partner_code}/attribute-stats
func (h *PartnerAttributeHandler) GetPartnerAttributeMapStats(c *fiber.Ctx) error {
	partnerCode := c.Params("partner_code")
	if strings.TrimSpace(partnerCode) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.StandardErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error: dtos.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Partner code is required",
			},
		})
	}

	stats, err := h.partnerAttributeMapService.GetStats(c.Context(), partnerCode)
	if err != nil {
		return h.handleError(c, err, "GetPartnerAttributeMapStats")
	}

	return c.JSON(dtos.StandardResponse{
		Success: true,
		Message: "Partner attribute statistics retrieved successfully",
		Data:    stats,
	})
}
