package services

import (
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"github.com/sirupsen/logrus"
)

// locationService implements the LocationService interface
type locationService struct {
	countryService       CountryService
	regionTypeService    RegionTypeService
	regionService        RegionService
	districtService      DistrictService
	cityService          CityService
	areaService          AreaService
	postalCodeService    PostalCodeService
	locationTypeService  LocationTypeService
	locationAliasService LocationAliasService
}

// NewLocationService creates a new location service with all sub-services
func NewLocationService(repoFactory repositories.LocationRepository, logger *logrus.Logger) LocationService {
	return &locationService{
		countryService:       NewCountryService(repoFactory.Countries()),
		regionTypeService:    NewRegionTypeService(repoFactory.RegionTypes()),
		regionService:        NewRegionService(repoFactory.Regions()),
		districtService:      NewDistrictService(repoFactory.Districts()),
		cityService:          NewCityService(repoFactory.Cities()),
		areaService:          NewAreaService(repoFactory.Areas()),
		postalCodeService:    NewPostalCodeService(repoFactory.PostalCodes(), repoFactory.Countries(), repoFactory.Regions(), repoFactory.Cities(), repoFactory.Areas()),
		locationTypeService:  NewLocationTypeService(repoFactory.LocationTypes()),
		locationAliasService: NewLocationAliasService(repoFactory.LocationAliases(), repoFactory, logger),
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

// PostalCodes returns the postal code service
func (s *locationService) PostalCodes() PostalCodeService {
	return s.postalCodeService
}

// LocationTypes returns the location type service
func (s *locationService) LocationTypes() LocationTypeService {
	return s.locationTypeService
}

// LocationAliases returns the location alias service
func (s *locationService) LocationAliases() LocationAliasService {
	return s.locationAliasService
}
