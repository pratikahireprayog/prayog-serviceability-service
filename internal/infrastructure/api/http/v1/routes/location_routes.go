package routes

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/middleware"
	"prayog-serviceability-service/internal/shared/dtos/v1"
)

// RegisterLocationRoutes registers all location-related routes
func RegisterLocationRoutes(router fiber.Router, handler *handlers.LocationHandler, logger *logrus.Logger) {
	logger.Info("Registering location routes...")

	// Create validation middleware
	validationMiddleware := middleware.NewValidationMiddleware(validator.New(), logger)
	paramValidators := middleware.GetCommonParamValidators()

	// Country routes
	countries := router.Group("/countries")
	{
		countries.Get("/", handler.GetAllCountries)
		countries.Post("/",
			validationMiddleware.ValidateBody(&dtos.CreateCountryRequest{}),
			handler.CreateCountry)
		countries.Get("/:id",
			validationMiddleware.ValidateParams(map[string]func(string) error{"id": paramValidators["id"]}),
			handler.GetCountryByID)
		countries.Put("/:id",
			validationMiddleware.ValidateParams(map[string]func(string) error{"id": paramValidators["id"]}),
			validationMiddleware.ValidateBody(&dtos.UpdateCountryRequest{}),
			handler.UpdateCountry)
		countries.Delete("/:id",
			validationMiddleware.ValidateParams(map[string]func(string) error{"id": paramValidators["id"]}),
			handler.DeleteCountry)
		countries.Get("/code/:code",
			validationMiddleware.ValidateParams(map[string]func(string) error{"code": paramValidators["code"]}),
			handler.GetCountryByCode)
	}

	// Region Type routes
	regionTypes := router.Group("/region-types")
	{
		regionTypes.Get("/", handler.GetAllRegionTypes)
		regionTypes.Post("/", handler.CreateRegionType)
		regionTypes.Get("/:code", handler.GetRegionTypeByCode)
		regionTypes.Put("/:code", handler.UpdateRegionType)
		regionTypes.Delete("/:code", handler.DeleteRegionType)
	}

	// Region routes
	regions := router.Group("/regions")
	{
		regions.Get("/", handler.GetAllRegions)
		regions.Post("/", handler.CreateRegion)
		regions.Get("/:id", handler.GetRegionByID)
		regions.Put("/:id", handler.UpdateRegion)
		regions.Delete("/:id", handler.DeleteRegion)
		regions.Get("/code/:code", handler.GetRegionByCode)
		regions.Get("/country/:countryId", handler.GetRegionsByCountryID)
	}

	// District routes
	districts := router.Group("/districts")
	{
		districts.Get("/", handler.GetAllDistricts)
		districts.Post("/", handler.CreateDistrict)
		districts.Get("/:id", handler.GetDistrictByID)
		districts.Put("/:id", handler.UpdateDistrict)
		districts.Delete("/:id", handler.DeleteDistrict)
		districts.Get("/code/:code", handler.GetDistrictByCode)
		districts.Get("/region/:regionId", handler.GetDistrictsByRegionID)
	}

	// City routes
	cities := router.Group("/cities")
	{
		cities.Get("/", handler.GetAllCities)
		cities.Post("/", handler.CreateCity)
		cities.Get("/:id", handler.GetCityByID)
		cities.Put("/:id", handler.UpdateCity)
		cities.Delete("/:id", handler.DeleteCity)
		cities.Get("/code/:code", handler.GetCityByCode)
		cities.Get("/region/:regionId", handler.GetCitiesByRegionID)
	}

	// Area routes
	areas := router.Group("/areas")
	{
		areas.Get("/", handler.GetAllAreas)
		areas.Post("/", handler.CreateArea)
		areas.Get("/:id", handler.GetAreaByID)
		areas.Put("/:id", handler.UpdateArea)
		areas.Delete("/:id", handler.DeleteArea)
		areas.Get("/code/:code", handler.GetAreaByCode)
		areas.Get("/city/:cityId", handler.GetAreasByCityID)
	}

	logger.Info("Location routes registered successfully")
}
