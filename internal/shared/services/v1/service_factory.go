package services

import (
	"prayog-serviceability-service/internal/shared/repositories/v1"
)

// locationService implements the LocationService interface
type locationService struct {
	countryService    CountryService
	regionTypeService RegionTypeService
	regionService     RegionService
	districtService   DistrictService
	cityService       CityService
	areaService       AreaService
}

// NewLocationService creates a new location service with all sub-services
func NewLocationService(repoFactory repositories.LocationRepository) LocationService {
	return &locationService{
		countryService:    NewCountryService(repoFactory.Countries()),
		regionTypeService: NewRegionTypeService(repoFactory.RegionTypes()),
		regionService:     NewRegionService(repoFactory.Regions()),
		districtService:   NewDistrictService(repoFactory.Districts()),
		cityService:       NewCityService(repoFactory.Cities()),
		areaService:       NewAreaService(repoFactory.Areas()),
	}
}

// Countries returns the country service
func (s *locationService) Countries() CountryService {
	return s.countryService
}

// RegionTypes returns the region type service
func (s *locationService) RegionTypes() RegionTypeService {
	return s.regionTypeService
}

// Regions returns the region service
func (s *locationService) Regions() RegionService {
	return s.regionService
}

// Districts returns the district service
func (s *locationService) Districts() DistrictService {
	return s.districtService
}

// Cities returns the city service
func (s *locationService) Cities() CityService {
	return s.cityService
}

// Areas returns the area service
func (s *locationService) Areas() AreaService {
	return s.areaService
}
