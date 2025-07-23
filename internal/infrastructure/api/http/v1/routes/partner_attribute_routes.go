package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/middleware"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/utils/v1"
)

// RegisterPartnerAttributeRoutes registers all partner attribute routes
func RegisterPartnerAttributeRoutes(router fiber.Router, handler *handlers.PartnerAttributeHandler, logger *logrus.Logger) {
	logger.Info("🏷️ Registering partner attribute routes...")

	// Create validator with custom validation functions registered
	validatorSetup := utils.NewValidatorSetup()
	validator := validatorSetup.GetValidator()

	// Create validation middleware with properly configured validator
	validationMiddleware := middleware.NewValidationMiddleware(validator, logger)

	// ATTRIBUTE CATEGORY ROUTES
	attributeCategories := router.Group("/attribute-categories")
	{
		// GET /attribute-categories - List all attribute categories
		attributeCategories.Get("/", handler.GetAllAttributeCategories)

		// POST /attribute-categories - Create a new attribute category
		attributeCategories.Post("/",
			validationMiddleware.ValidateBody(&dtos.CreateAttributeCategoryRequest{}),
			handler.CreateAttributeCategory)

		// GET /attribute-categories/{id} - Get attribute category by ID
		attributeCategories.Get("/:id", handler.GetAttributeCategoryByID)

		// PUT /attribute-categories/{id} - Update attribute category
		attributeCategories.Put("/:id",
			validationMiddleware.ValidateBody(&dtos.UpdateAttributeCategoryRequest{}),
			handler.UpdateAttributeCategory)

		// DELETE /attribute-categories/{id} - Delete attribute category
		attributeCategories.Delete("/:id", handler.DeleteAttributeCategory)

		// PATCH /attribute-categories/{id}/restore - Restore attribute category
		attributeCategories.Patch("/:id/restore", handler.RestoreAttributeCategory)

		// GET /attribute-categories/code/{code} - Get attribute category by code
		attributeCategories.Get("/code/:code", handler.GetAttributeCategoryByCode)

		// GET /attribute-categories/{id}/stats - Get attribute category statistics
		attributeCategories.Get("/:id/stats", handler.GetAttributeCategoryStats)

		// GET /attribute-categories/{category_id}/attributes - Get attributes by category ID
		attributeCategories.Get("/:category_id/attributes", handler.GetAttributesByCategory)

		// GET /attribute-categories/code/{category_code}/attributes - Get attributes by category code
		attributeCategories.Get("/code/:category_code/attributes", handler.GetAttributesByCategoryCode)
	}

	// ATTRIBUTE ROUTES
	attributes := router.Group("/attributes")
	{
		// GET /attributes - List all attributes
		attributes.Get("/", handler.GetAllAttributes)

		// POST /attributes - Create a new attribute
		attributes.Post("/",
			validationMiddleware.ValidateBody(&dtos.CreateAttributeRequest{}),
			handler.CreateAttribute)

		// GET /attributes/{id} - Get attribute by ID
		attributes.Get("/:id", handler.GetAttributeByID)

		// PUT /attributes/{id} - Update attribute
		attributes.Put("/:id",
			validationMiddleware.ValidateBody(&dtos.UpdateAttributeRequest{}),
			handler.UpdateAttribute)

		// DELETE /attributes/{id} - Delete attribute
		attributes.Delete("/:id", handler.DeleteAttribute)

		// PATCH /attributes/{id}/restore - Restore attribute
		attributes.Patch("/:id/restore", handler.RestoreAttribute)

		// GET /attributes/code/{code} - Get attribute by code
		attributes.Get("/code/:code", handler.GetAttributeByCode)

		// GET /attributes/{id}/stats - Get attribute statistics
		attributes.Get("/:id/stats", handler.GetAttributeStats)

		// GET /attributes/code/{attribute_code}/partners - Get partner mappings by attribute code
		attributes.Get("/code/:attribute_code/partners", handler.GetPartnerAttributeMapsByAttribute)

		// GET /attributes/code/{attribute_code}/partner-codes - Get partner codes by attribute
		attributes.Get("/code/:attribute_code/partner-codes", handler.GetPartnerCodesByAttribute)
	}

	// PARTNER ATTRIBUTE MAP ROUTES
	partnerAttributeMaps := router.Group("/partner-attribute-maps")
	{
		// GET /partner-attribute-maps - List all partner attribute mappings
		partnerAttributeMaps.Get("/", handler.GetAllPartnerAttributeMaps)

		// POST /partner-attribute-maps - Create a new partner attribute mapping
		partnerAttributeMaps.Post("/",
			validationMiddleware.ValidateBody(&dtos.CreatePartnerAttributeMapRequest{}),
			handler.CreatePartnerAttributeMap)

		// GET /partner-attribute-maps/{id} - Get partner attribute mapping by ID
		partnerAttributeMaps.Get("/:id", handler.GetPartnerAttributeMapByID)

		// PUT /partner-attribute-maps/{id} - Update partner attribute mapping
		partnerAttributeMaps.Put("/:id",
			validationMiddleware.ValidateBody(&dtos.UpdatePartnerAttributeMapRequest{}),
			handler.UpdatePartnerAttributeMap)

		// DELETE /partner-attribute-maps/{id} - Delete partner attribute mapping
		partnerAttributeMaps.Delete("/:id", handler.DeletePartnerAttributeMap)

		// PATCH /partner-attribute-maps/{id}/restore - Restore partner attribute mapping
		partnerAttributeMaps.Patch("/:id/restore", handler.RestorePartnerAttributeMap)

		// POST /partner-attribute-maps/search - Search with filters
		partnerAttributeMaps.Post("/search",
			validationMiddleware.ValidateBody(&dtos.PartnerAttributeMapFilterRequest{}),
			handler.GetPartnerAttributeMapsWithFilters)

		// POST /partner-attribute-maps/bulk - Bulk create partner attribute mappings
		partnerAttributeMaps.Post("/bulk",
			validationMiddleware.ValidateBody(&dtos.BulkCreatePartnerAttributeMapRequest{}),
			handler.BulkCreatePartnerAttributeMaps)

		// DELETE /partner-attribute-maps/bulk - Bulk delete partner attribute mappings
		partnerAttributeMaps.Delete("/bulk",
			validationMiddleware.ValidateBody(&dtos.BulkDeletePartnerAttributeMapRequest{}),
			handler.BulkDeletePartnerAttributeMaps)
	}

	// PARTNER-CENTRIC ROUTES
	partners := router.Group("/partners")
	{
		// GET /partners/{partner_code}/attributes - Get all attributes mapped to a partner
		partners.Get("/:partner_code/attributes", handler.GetPartnerAttributeMapsByPartner)

		// GET /partners/id/{partner_id}/attributes - Get all attributes mapped to a partner by ID
		partners.Get("/id/:partner_id/attributes", handler.GetPartnerAttributeMapsByPartnerID)

		// GET /partners/{partner_code}/attributes/details - Get detailed attributes for a partner
		partners.Get("/:partner_code/attributes/details", handler.GetAttributesByPartner)

		// GET /partners/{partner_code}/attribute-stats - Get partner attribute statistics
		partners.Get("/:partner_code/attribute-stats", handler.GetPartnerAttributeMapStats)
	}

	logger.Info("🏷️ Partner attribute routes registered successfully")
}
