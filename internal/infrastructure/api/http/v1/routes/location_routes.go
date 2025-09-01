package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/infrastructure/api/http/v1/handlers"
	"prayog-serviceability-service/internal/infrastructure/api/http/v1/middleware"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/utils/v1"
)

// RegisterLocationRoutes registers all location-related routes
func RegisterLocationRoutes(router fiber.Router, handler *handlers.LocationHandler, logger *logrus.Logger) {
	logger.Info("📍 Registering location routes...")

	// Create validator with custom validation functions registered
	validatorSetup := utils.NewValidatorSetup()
	validator := validatorSetup.GetValidator()

	// Create validation middleware with properly configured validator
	validationMiddleware := middleware.NewValidationMiddleware(validator, logger)

	// Country routes
	countries := router.Group("/countries")
	{
		countries.Get("/", handler.GetAllCountries)
		countries.Post("/",
			validationMiddleware.ValidateBody(&dtos.CreateCountryRequest{}),
			handler.CreateCountry)
		countries.Get("/by-codes", handler.GetCountriesByCodes)
		countries.Get("/by-code", handler.GetCountryByCodeQuery)
		countries.Get("/by-name", handler.GetCountriesByNameQuery)
		countries.Get("/by-names", handler.GetCountriesByNamesQuery)
		countries.Get("/code/:code", handler.GetCountryByCode)
		countries.Get("/:id", handler.GetCountryByID)
		countries.Put("/:id",
			validationMiddleware.ValidateBody(&dtos.UpdateCountryRequest{}),
			handler.UpdateCountry)
		countries.Delete("/:id", handler.DeleteCountry)
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

	// PostalCode routes
	postalCodes := router.Group("/postal-codes")
	{
		postalCodes.Get("/", handler.GetAllPostalCodes)
		postalCodes.Post("/",
			validationMiddleware.ValidateBody(&dtos.CreatePostalCodeRequest{}),
			handler.CreatePostalCode)
		postalCodes.Get("/:id", handler.GetPostalCodeByID)
		postalCodes.Put("/:id",
			validationMiddleware.ValidateBody(&dtos.UpdatePostalCodeRequest{}),
			handler.UpdatePostalCode)
		postalCodes.Delete("/:id", handler.DeletePostalCode)
		postalCodes.Get("/location", handler.GetPostalCodesByLocation)
	}

	// LocationType routes
	locationTypes := router.Group("/location-types")
	{
		locationTypes.Get("/", handler.GetAllLocationTypes)
		locationTypes.Post("/",
			validationMiddleware.ValidateBody(&dtos.CreateLocationTypeRequest{}),
			handler.CreateLocationType)
		locationTypes.Get("/:code", handler.GetLocationTypeByCode)
		locationTypes.Put("/:code",
			validationMiddleware.ValidateBody(&dtos.UpdateLocationTypeRequest{}),
			handler.UpdateLocationType)
		locationTypes.Delete("/:code", handler.DeleteLocationType)
	}

	logger.Info("📍 Location routes registered successfully")
}
