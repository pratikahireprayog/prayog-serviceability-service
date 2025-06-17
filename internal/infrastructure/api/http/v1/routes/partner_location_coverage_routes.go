package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
)

// RegisterPartnerLocationCoverageRoutes registers all partner location coverage routes
func RegisterPartnerLocationCoverageRoutes(router fiber.Router, handler *handlers.PartnerLocationCoverageHandler, logger *logrus.Logger) {
	logger.Info("🤝 Registering partner location coverage routes...")

	// Partner location coverage routes - nested under /partners
	partners := router.Group("/partners")
	{
		// Partner ID based routes - /partners/{partner_id}/location-coverages
		partnerCoverages := partners.Group("/:partner_id/location-coverages")
		{
			// GET /partners/{partner_id}/location-coverages - List all location coverages for a partner
			partnerCoverages.Get("/",
				handler.GetPartnerLocationCoverages)

			// POST /partners/{partner_id}/location-coverages - Create a new location coverage for a partner
			partnerCoverages.Post("/",
				handler.CreatePartnerLocationCoverage)

			// POST /partners/{partner_id}/location-coverages/bulk - Bulk create location coverages for a partner
			partnerCoverages.Post("/bulk",
				handler.BulkCreatePartnerLocationCoverages)

			// GET /partners/{partner_id}/location-coverages/postal-code/{postal_code} - Get coverage by postal code
			partnerCoverages.Get("/postal-code/:postal_code",
				handler.GetPartnerLocationCoverageByPostalCode)

			// PUT /partners/{partner_id}/location-coverages/postal-code/{postal_code} - Update coverage by postal code
			partnerCoverages.Put("/postal-code/:postal_code",
				handler.UpdatePartnerLocationCoverageByPostalCode)

			// DELETE /partners/{partner_id}/location-coverages/postal-code/{postal_code} - Delete coverage by postal code
			partnerCoverages.Delete("/postal-code/:postal_code",
				handler.DeletePartnerLocationCoverageByPostalCode)

			// GET /partners/{partner_id}/location-coverages/postal-code/{postal_code}/check - Check coverage availability
			partnerCoverages.Get("/postal-code/:postal_code/check",
				handler.CheckPartnerLocationCoverage)

			// Legacy routes for backward compatibility (using coverage_id)
			// GET /partners/{partner_id}/location-coverages/{coverage_id} - Get a specific location coverage
			partnerCoverages.Get("/:coverage_id",
				handler.GetPartnerLocationCoverage)

			// PUT /partners/{partner_id}/location-coverages/{coverage_id} - Update a specific location coverage
			partnerCoverages.Put("/:coverage_id",
				handler.UpdatePartnerLocationCoverage)

			// DELETE /partners/{partner_id}/location-coverages/{coverage_id} - Delete a specific location coverage
			partnerCoverages.Delete("/:coverage_id",
				handler.DeletePartnerLocationCoverage)
		}

		// Partner Code based routes - /partners/code/{partner_code}/location-coverages
		partnerCodeCoverages := partners.Group("/code/:partner_code/location-coverages")
		{
			// GET /partners/code/{partner_code}/location-coverages - List all location coverages for a partner
			partnerCodeCoverages.Get("/",
				handler.GetPartnerLocationCoveragesByCode)

			// POST /partners/code/{partner_code}/location-coverages - Create a new location coverage for a partner
			partnerCodeCoverages.Post("/",
				handler.CreatePartnerLocationCoverageByCode)

			// POST /partners/code/{partner_code}/location-coverages/bulk - Bulk create location coverages for a partner
			partnerCodeCoverages.Post("/bulk",
				handler.BulkCreatePartnerLocationCoveragesByCode)

			// GET /partners/code/{partner_code}/location-coverages/postal-code/{postal_code} - Get coverage by postal code
			partnerCodeCoverages.Get("/postal-code/:postal_code",
				handler.GetPartnerLocationCoverageByCodeAndPostalCode)

			// PUT /partners/code/{partner_code}/location-coverages/postal-code/{postal_code} - Update coverage by postal code
			partnerCodeCoverages.Put("/postal-code/:postal_code",
				handler.UpdatePartnerLocationCoverageByCodeAndPostalCode)

			// DELETE /partners/code/{partner_code}/location-coverages/postal-code/{postal_code} - Delete coverage by postal code
			partnerCodeCoverages.Delete("/postal-code/:postal_code",
				handler.DeletePartnerLocationCoverageByCodeAndPostalCode)

			// GET /partners/code/{partner_code}/location-coverages/postal-code/{postal_code}/check - Check coverage availability
			partnerCodeCoverages.Get("/postal-code/:postal_code/check",
				handler.CheckPartnerLocationCoverageByCode)
		}
	}

	logger.Info("🤝 Partner location coverage routes registered successfully")
}
