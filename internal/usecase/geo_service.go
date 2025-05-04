package usecase

import (
	"context"

	"github.com/prayog/serviceability/internal/domain"
)

// geoService implements the GeoService interface
type geoService struct {
	countryRepo    domain.CountryRepository
	regionRepo     domain.RegionRepository
	cityRepo       domain.CityRepository
	areaRepo       domain.AreaRepository
	postalCodeRepo domain.PostalCodeRepository
	ctx            context.Context
}

// NewGeoService creates a new geography service
func NewGeoService(
	countryRepo domain.CountryRepository,
	regionRepo domain.RegionRepository,
	cityRepo domain.CityRepository,
	areaRepo domain.AreaRepository,
	postalCodeRepo domain.PostalCodeRepository,
) GeoService {
	return &geoService{
		countryRepo:    countryRepo,
		regionRepo:     regionRepo,
		cityRepo:       cityRepo,
		areaRepo:       areaRepo,
		postalCodeRepo: postalCodeRepo,
		ctx:            context.Background(),
	}
}

// Country operations

// GetCountryByID retrieves a country by its ID
func (s *geoService) GetCountryByID(id uint) (*domain.Country, error) {
	return s.countryRepo.GetByID(s.ctx, id)
}

// GetCountryByCode retrieves a country by its code
func (s *geoService) GetCountryByCode(code string) (*domain.Country, error) {
	return s.countryRepo.GetByCode(s.ctx, code)
}

// ListCountries retrieves all countries
func (s *geoService) ListCountries() ([]*domain.Country, error) {
	return s.countryRepo.List(s.ctx)
}

// CreateCountry creates a new country
func (s *geoService) CreateCountry(country *domain.Country) error {
	return s.countryRepo.Create(s.ctx, country)
}

// UpdateCountry updates an existing country
func (s *geoService) UpdateCountry(country *domain.Country) error {
	return s.countryRepo.Update(s.ctx, country)
}

// DeleteCountry deletes a country by its ID
func (s *geoService) DeleteCountry(id uint) error {
	return s.countryRepo.Delete(s.ctx, id)
}

// Region operations

// GetRegionByID retrieves a region by its ID
func (s *geoService) GetRegionByID(id uint) (*domain.AdministrativeRegion, error) {
	return s.regionRepo.GetByID(s.ctx, id)
}

// GetRegionByCode retrieves a region by its code
func (s *geoService) GetRegionByCode(code string) (*domain.AdministrativeRegion, error) {
	return s.regionRepo.GetByCode(s.ctx, code)
}

// GetRegionsByCountryID retrieves all regions for a specific country
func (s *geoService) GetRegionsByCountryID(countryID uint) ([]*domain.AdministrativeRegion, error) {
	return s.regionRepo.GetByCountryID(s.ctx, countryID)
}

// ListRegions retrieves all regions
func (s *geoService) ListRegions() ([]*domain.AdministrativeRegion, error) {
	return s.regionRepo.List(s.ctx)
}

// CreateRegion creates a new region
func (s *geoService) CreateRegion(region *domain.AdministrativeRegion) error {
	return s.regionRepo.Create(s.ctx, region)
}

// UpdateRegion updates an existing region
func (s *geoService) UpdateRegion(region *domain.AdministrativeRegion) error {
	return s.regionRepo.Update(s.ctx, region)
}

// DeleteRegion deletes a region by its ID
func (s *geoService) DeleteRegion(id uint) error {
	return s.regionRepo.Delete(s.ctx, id)
}

// City operations

// GetCityByID retrieves a city by its ID
func (s *geoService) GetCityByID(id uint) (*domain.City, error) {
	return s.cityRepo.GetByID(s.ctx, id)
}

// GetCitiesByRegionID retrieves all cities for a specific region
func (s *geoService) GetCitiesByRegionID(regionID uint) ([]*domain.City, error) {
	return s.cityRepo.GetByRegionID(s.ctx, regionID)
}

// ListCities retrieves all cities
func (s *geoService) ListCities() ([]*domain.City, error) {
	return s.cityRepo.List(s.ctx)
}

// CreateCity creates a new city
func (s *geoService) CreateCity(city *domain.City) error {
	return s.cityRepo.Create(s.ctx, city)
}

// UpdateCity updates an existing city
func (s *geoService) UpdateCity(city *domain.City) error {
	return s.cityRepo.Update(s.ctx, city)
}

// DeleteCity deletes a city by its ID
func (s *geoService) DeleteCity(id uint) error {
	return s.cityRepo.Delete(s.ctx, id)
}

// Area operations

// GetAreaByID retrieves an area by its ID
func (s *geoService) GetAreaByID(id uint) (*domain.Area, error) {
	return s.areaRepo.GetByID(s.ctx, id)
}

// GetAreasByCityID retrieves all areas for a specific city
func (s *geoService) GetAreasByCityID(cityID uint) ([]*domain.Area, error) {
	return s.areaRepo.GetByCityID(s.ctx, cityID)
}

// ListAreas retrieves all areas
func (s *geoService) ListAreas() ([]*domain.Area, error) {
	return s.areaRepo.List(s.ctx)
}

// CreateArea creates a new area
func (s *geoService) CreateArea(area *domain.Area) error {
	return s.areaRepo.Create(s.ctx, area)
}

// UpdateArea updates an existing area
func (s *geoService) UpdateArea(area *domain.Area) error {
	return s.areaRepo.Update(s.ctx, area)
}

// DeleteArea deletes an area by its ID
func (s *geoService) DeleteArea(id uint) error {
	return s.areaRepo.Delete(s.ctx, id)
}

// Postal code operations

// GetPostalCodeByID retrieves a postal code by its ID
func (s *geoService) GetPostalCodeByID(id uint) (*domain.PostalCode, error) {
	return s.postalCodeRepo.GetByID(s.ctx, id)
}

// GetPostalCodeByCode retrieves a postal code by its code
func (s *geoService) GetPostalCodeByCode(code string) (*domain.PostalCode, error) {
	return s.postalCodeRepo.GetByCode(s.ctx, code)
}

// GetPostalCodesByAreaID retrieves all postal codes for a specific area
func (s *geoService) GetPostalCodesByAreaID(areaID uint) ([]*domain.PostalCode, error) {
	return s.postalCodeRepo.GetByAreaID(s.ctx, areaID)
}

// ListPostalCodes retrieves all postal codes
func (s *geoService) ListPostalCodes() ([]*domain.PostalCode, error) {
	return s.postalCodeRepo.List(s.ctx)
}

// CreatePostalCode creates a new postal code
func (s *geoService) CreatePostalCode(postalCode *domain.PostalCode) error {
	return s.postalCodeRepo.Create(s.ctx, postalCode)
}

// UpdatePostalCode updates an existing postal code
func (s *geoService) UpdatePostalCode(postalCode *domain.PostalCode) error {
	return s.postalCodeRepo.Update(s.ctx, postalCode)
}

// DeletePostalCode deletes a postal code by its ID
func (s *geoService) DeletePostalCode(id uint) error {
	return s.postalCodeRepo.Delete(s.ctx, id)
}
