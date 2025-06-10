package routes

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/middleware"
	"prayog-serviceability-service/internal/shared/dtos/v1"
)

// RegisterPartnerLocationCoverageRoutes registers all partner location coverage routes
func RegisterPartnerLocationCoverageRoutes(router fiber.Router, handler *handlers.PartnerLocationCoverageHandler, logger *logrus.Logger) {
	logger.Info("Registering partner location coverage routes...")

	// Create validation middleware
	validationMiddleware := middleware.NewValidationMiddleware(validator.New(), logger)
	paramValidators := middleware.GetCommonParamValidators()

	// Partner location coverage routes - nested under /partners/{partner_id}/location-coverages
	partners := router.Group("/partners")
	{
		// Partner location coverage routes
		partnerCoverages := partners.Group("/:partner_id/location-coverages")
		{
			// GET /partners/{partner_id}/location-coverages - List all location coverages for a partner
			partnerCoverages.Get("/",
				validationMiddleware.ValidateParams(map[string]func(string) error{"partner_id": paramValidators["uuid"]}),
				handler.GetPartnerLocationCoverages)

			// POST /partners/{partner_id}/location-coverages - Create a new location coverage for a partner
			partnerCoverages.Post("/",
				validationMiddleware.ValidateParams(map[string]func(string) error{"partner_id": paramValidators["uuid"]}),
				validationMiddleware.ValidateBody(&dtos.CreatePartnerLocationCoverageRequest{}),
				handler.CreatePartnerLocationCoverage)

			// POST /partners/{partner_id}/location-coverages/bulk - Bulk create location coverages for a partner
			partnerCoverages.Post("/bulk",
				validationMiddleware.ValidateParams(map[string]func(string) error{"partner_id": paramValidators["uuid"]}),
				validationMiddleware.ValidateBody(&dtos.BulkCreatePartnerLocationCoverageRequest{}),
				handler.BulkCreatePartnerLocationCoverages)

			// GET /partners/{partner_id}/location-coverages/{coverage_id} - Get a specific location coverage
			partnerCoverages.Get("/:coverage_id",
				validationMiddleware.ValidateParams(map[string]func(string) error{
					"partner_id":  paramValidators["uuid"],
					"coverage_id": paramValidators["uuid"],
				}),
				handler.GetPartnerLocationCoverage)

			// PUT /partners/{partner_id}/location-coverages/{coverage_id} - Update a specific location coverage
			partnerCoverages.Put("/:coverage_id",
				validationMiddleware.ValidateParams(map[string]func(string) error{
					"partner_id":  paramValidators["uuid"],
					"coverage_id": paramValidators["uuid"],
				}),
				validationMiddleware.ValidateBody(&dtos.UpdatePartnerLocationCoverageRequest{}),
				handler.UpdatePartnerLocationCoverage)

			// DELETE /partners/{partner_id}/location-coverages/{coverage_id} - Delete a specific location coverage
			partnerCoverages.Delete("/:coverage_id",
				validationMiddleware.ValidateParams(map[string]func(string) error{
					"partner_id":  paramValidators["uuid"],
					"coverage_id": paramValidators["uuid"],
				}),
				handler.DeletePartnerLocationCoverage)
		}
	}

	logger.Info("Partner location coverage routes registered successfully")
}
