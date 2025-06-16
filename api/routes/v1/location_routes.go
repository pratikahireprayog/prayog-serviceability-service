package v1

import (
	v1 "prayog-serviceability-service/api/handlers/v1"

	"github.com/gofiber/fiber/v2"
)

// RegisterLocationRoutes registers all location routes with the provided router
func RegisterLocationRoutes(router fiber.Router, handler *v1.LocationHandler) {
	// Country routes
	countries := router.Group("/countries")
	{
		countries.Get("/", handler.GetAllCountries)            // GET /countries
		countries.Post("/", handler.CreateCountry)             // POST /countries
		countries.Get("/:id", handler.GetCountryByID)          // GET /countries/{id}
		countries.Put("/:id", handler.UpdateCountry)           // PUT /countries/{id}
		countries.Delete("/:id", handler.DeleteCountry)        // DELETE /countries/{id}
		countries.Get("/code/:code", handler.GetCountryByCode) // GET /countries/code/{code}
	}

	// RegionType routes
	regionTypes := router.Group("/region-types")
	{
		regionTypes.Get("/", handler.GetAllRegionTypes)        // GET /region-types
		regionTypes.Post("/", handler.CreateRegionType)        // POST /region-types
		regionTypes.Get("/:code", handler.GetRegionTypeByCode) // GET /region-types/{code}
		regionTypes.Put("/:code", handler.UpdateRegionType)    // PUT /region-types/{code}
		regionTypes.Delete("/:code", handler.DeleteRegionType) // DELETE /region-types/{code}
	}

	// Region routes
	regions := router.Group("/regions")
	{
		regions.Get("/", handler.GetAllRegions)                           // GET /regions
		regions.Post("/", handler.CreateRegion)                           // POST /regions
		regions.Get("/:id", handler.GetRegionByID)                        // GET /regions/{id}
		regions.Put("/:id", handler.UpdateRegion)                         // PUT /regions/{id}
		regions.Delete("/:id", handler.DeleteRegion)                      // DELETE /regions/{id}
		regions.Get("/code/:code", handler.GetRegionByCode)               // GET /regions/code/{code}
		regions.Get("/country/:countryID", handler.GetRegionsByCountryID) // GET /regions/country/{countryID}
	}

	// District routes
	districts := router.Group("/districts")
	{
		districts.Get("/", handler.GetAllDistricts)                        // GET /districts
		districts.Post("/", handler.CreateDistrict)                        // POST /districts
		districts.Get("/:id", handler.GetDistrictByID)                     // GET /districts/{id}
		districts.Put("/:id", handler.UpdateDistrict)                      // PUT /districts/{id}
		districts.Delete("/:id", handler.DeleteDistrict)                   // DELETE /districts/{id}
		districts.Get("/code/:code", handler.GetDistrictByCode)            // GET /districts/code/{code}
		districts.Get("/region/:regionID", handler.GetDistrictsByRegionID) // GET /districts/region/{regionID}
	}

	// City routes
	cities := router.Group("/cities")
	{
		cities.Get("/", handler.GetAllCities)                        // GET /cities
		cities.Post("/", handler.CreateCity)                         // POST /cities
		cities.Get("/:id", handler.GetCityByID)                      // GET /cities/{id}
		cities.Put("/:id", handler.UpdateCity)                       // PUT /cities/{id}
		cities.Delete("/:id", handler.DeleteCity)                    // DELETE /cities/{id}
		cities.Get("/code/:code", handler.GetCityByCode)             // GET /cities/code/{code}
		cities.Get("/region/:regionID", handler.GetCitiesByRegionID) // GET /cities/region/{regionID}
	}

	// Area routes
	areas := router.Group("/areas")
	{
		areas.Get("/", handler.GetAllAreas)                  // GET /areas
		areas.Post("/", handler.CreateArea)                  // POST /areas
		areas.Get("/:id", handler.GetAreaByID)               // GET /areas/{id}
		areas.Put("/:id", handler.UpdateArea)                // PUT /areas/{id}
		areas.Delete("/:id", handler.DeleteArea)             // DELETE /areas/{id}
		areas.Get("/code/:code", handler.GetAreaByCode)      // GET /areas/code/{code}
		areas.Get("/city/:cityID", handler.GetAreasByCityID) // GET /areas/city/{cityID}
	}

	// Location Alias standalone routes
	locationAliases := router.Group("/location-aliases")
	{
		locationAliases.Get("/", handler.GetAllLocationAliases)     // GET /location-aliases
		locationAliases.Get("/:id", handler.GetLocationAliasByID)   // GET /location-aliases/{id}
		locationAliases.Put("/:id", handler.UpdateLocationAlias)    // PUT /location-aliases/{id}
		locationAliases.Delete("/:id", handler.DeleteLocationAlias) // DELETE /location-aliases/{id}
	}

	// Nested Location Alias routes for different entity types

	// Generic entity routes (entity_id parameter based)
	locations := router.Group("/locations")
	{
		locations.Post("/:entityID/aliases", handler.CreateLocationAliasByEntityID) // POST /locations/{entityID}/aliases
		locations.Get("/:entityID/aliases", handler.GetLocationAliasesByEntityID)   // GET /locations/{entityID}/aliases
	}

	// Specific entity type routes (for more explicit API usage)

	// Country aliases
	countries.Post("/:entityID/aliases", handler.CreateLocationAliasByEntityID) // POST /countries/{entityID}/aliases
	countries.Get("/:entityID/aliases", handler.GetLocationAliasesByEntityID)   // GET /countries/{entityID}/aliases

	// Region aliases
	regions.Post("/:entityID/aliases", handler.CreateLocationAliasByEntityID) // POST /regions/{entityID}/aliases
	regions.Get("/:entityID/aliases", handler.GetLocationAliasesByEntityID)   // GET /regions/{entityID}/aliases

	// District aliases
	districts.Post("/:entityID/aliases", handler.CreateLocationAliasByEntityID) // POST /districts/{entityID}/aliases
	districts.Get("/:entityID/aliases", handler.GetLocationAliasesByEntityID)   // GET /districts/{entityID}/aliases

	// City aliases
	cities.Post("/:entityID/aliases", handler.CreateLocationAliasByEntityID) // POST /cities/{entityID}/aliases
	cities.Get("/:entityID/aliases", handler.GetLocationAliasesByEntityID)   // GET /cities/{entityID}/aliases

	// Area aliases
	areas.Post("/:entityID/aliases", handler.CreateLocationAliasByEntityID) // POST /areas/{entityID}/aliases
	areas.Get("/:entityID/aliases", handler.GetLocationAliasesByEntityID)   // GET /areas/{entityID}/aliases
}
