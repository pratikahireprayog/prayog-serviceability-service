package services

import (
	"context"
	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// CountryService defines the business logic interface for country operations
type CountryService interface {
	GetByID(ctx context.Context, id string) (*dtos.CountryResponse, error)
	GetByCode(ctx context.Context, code string) (*dtos.CountryResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.CountryListResponse, error)
	Create(ctx context.Context, req *dtos.CreateCountryRequest) (*dtos.CountryResponse, error)
	Update(ctx context.Context, id string, req *dtos.UpdateCountryRequest) (*dtos.CountryResponse, error)
	Delete(ctx context.Context, id string) error
}

// RegionTypeService defines the business logic interface for region type operations
type RegionTypeService interface {
	GetByCode(ctx context.Context, code string) (*dtos.RegionTypeResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.RegionTypeListResponse, error)
	Create(ctx context.Context, req *dtos.CreateRegionTypeRequest) (*dtos.RegionTypeResponse, error)
	Update(ctx context.Context, code string, req *dtos.UpdateRegionTypeRequest) (*dtos.RegionTypeResponse, error)
	Delete(ctx context.Context, code string) error
}

// RegionService defines the business logic interface for region operations
type RegionService interface {
	GetByID(ctx context.Context, id string) (*dtos.RegionResponse, error)
	GetByCode(ctx context.Context, code string) (*dtos.RegionResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.RegionListResponse, error)
	GetByCountryID(ctx context.Context, countryID string) ([]dtos.RegionResponse, error)
	Create(ctx context.Context, req *dtos.CreateRegionRequest) (*dtos.RegionResponse, error)
	Update(ctx context.Context, id string, req *dtos.UpdateRegionRequest) (*dtos.RegionResponse, error)
	Delete(ctx context.Context, id string) error
}

// DistrictService defines the business logic interface for district operations
type DistrictService interface {
	GetByID(ctx context.Context, id string) (*dtos.DistrictResponse, error)
	GetByCode(ctx context.Context, code string) (*dtos.DistrictResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.DistrictListResponse, error)
	GetByRegionID(ctx context.Context, regionID string) ([]dtos.DistrictResponse, error)
	Create(ctx context.Context, req *dtos.CreateDistrictRequest) (*dtos.DistrictResponse, error)
	Update(ctx context.Context, id string, req *dtos.UpdateDistrictRequest) (*dtos.DistrictResponse, error)
	Delete(ctx context.Context, id string) error
}

// CityService defines the business logic interface for city operations
type CityService interface {
	GetByID(ctx context.Context, id string) (*dtos.CityResponse, error)
	GetByCode(ctx context.Context, code string) (*dtos.CityResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.CityListResponse, error)
	GetByRegionID(ctx context.Context, regionID string) ([]dtos.CityResponse, error)
	Create(ctx context.Context, req *dtos.CreateCityRequest) (*dtos.CityResponse, error)
	Update(ctx context.Context, id string, req *dtos.UpdateCityRequest) (*dtos.CityResponse, error)
	Delete(ctx context.Context, id string) error
}

// AreaService defines the business logic interface for area operations
type AreaService interface {
	GetByID(ctx context.Context, id string) (*dtos.AreaResponse, error)
	GetByCode(ctx context.Context, code string) (*dtos.AreaResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.AreaListResponse, error)
	GetByCityID(ctx context.Context, cityID string) ([]dtos.AreaResponse, error)
	Create(ctx context.Context, req *dtos.CreateAreaRequest) (*dtos.AreaResponse, error)
	Update(ctx context.Context, id string, req *dtos.UpdateAreaRequest) (*dtos.AreaResponse, error)
	Delete(ctx context.Context, id string) error
}

// PostalCodeService defines the business logic interface for postal code operations
type PostalCodeService interface {
	GetByID(ctx context.Context, id string) (*dtos.PostalCodeResponse, error)
	GetByCode(ctx context.Context, code string) (*dtos.PostalCodeResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.PostalCodeListResponse, error)
	GetByLocation(ctx context.Context, countryCode, regionCode, cityCode, areaCode string) ([]dtos.PostalCodeResponse, error)
	Create(ctx context.Context, req *dtos.CreatePostalCodeRequest) (*dtos.PostalCodeResponse, error)
	Update(ctx context.Context, id string, req *dtos.UpdatePostalCodeRequest) (*dtos.PostalCodeResponse, error)
	Delete(ctx context.Context, id string) error
	ValidateLocationScope(ctx context.Context, req *dtos.CreatePostalCodeRequest) error
}

// LocationTypeService defines the business logic interface for location type operations
type LocationTypeService interface {
	GetByCode(ctx context.Context, code string) (*dtos.LocationTypeResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.LocationTypeListResponse, error)
	Create(ctx context.Context, req *dtos.CreateLocationTypeRequest) (*dtos.LocationTypeResponse, error)
	Update(ctx context.Context, code string, req *dtos.UpdateLocationTypeRequest) (*dtos.LocationTypeResponse, error)
	Delete(ctx context.Context, code string) error
}

// LocationService provides a unified interface for all location services
type LocationService interface {
	Countries() CountryService
	RegionTypes() RegionTypeService
	Regions() RegionService
	Districts() DistrictService
	Cities() CityService
	Areas() AreaService
	PostalCodes() PostalCodeService
	LocationTypes() LocationTypeService
	LocationAliases() LocationAliasService
}

// Helper functions for model to DTO conversion

// CountryToResponse converts a Country model to CountryResponse DTO
func CountryToResponse(country *models.Country) *dtos.CountryResponse {
	if country == nil {
		return nil
	}
	return &dtos.CountryResponse{
		ID:           country.ID,
		Code:         country.Code,
		Name:         country.Name,
		CurrencyCode: country.CurrencyCode,
		PhoneCode:    country.PhoneCode,
		IsActive:     country.IsActive,
		CreatedAt:    country.CreatedAt,
		UpdatedAt:    country.UpdatedAt,
	}
}

// RegionTypeToResponse converts a RegionType model to RegionTypeResponse DTO
func RegionTypeToResponse(regionType *models.RegionType) *dtos.RegionTypeResponse {
	if regionType == nil {
		return nil
	}
	return &dtos.RegionTypeResponse{
		Code:        regionType.Code,
		Name:        regionType.Name,
		Description: regionType.Description,
		IsActive:    regionType.IsActive,
		CreatedAt:   regionType.CreatedAt,
		UpdatedAt:   regionType.UpdatedAt,
	}
}

// RegionToResponse converts a Region model to RegionResponse DTO
func RegionToResponse(region *models.Region) *dtos.RegionResponse {
	if region == nil {
		return nil
	}
	response := &dtos.RegionResponse{
		ID:             region.ID,
		Code:           region.Code,
		CountryID:      region.CountryID,
		CountryCode:    region.CountryCode,
		RegionTypeCode: region.RegionTypeCode,
		Name:           region.Name,
		IsActive:       region.IsActive,
		CreatedAt:      region.CreatedAt,
		UpdatedAt:      region.UpdatedAt,
	}

	if region.Country != nil {
		response.Country = CountryToResponse(region.Country)
	}
	if region.RegionType != nil {
		response.RegionType = RegionTypeToResponse(region.RegionType)
	}

	return response
}

// DistrictToResponse converts a District model to DistrictResponse DTO
func DistrictToResponse(district *models.District) *dtos.DistrictResponse {
	if district == nil {
		return nil
	}
	response := &dtos.DistrictResponse{
		ID:          district.ID,
		Code:        district.Code,
		RegionID:    district.RegionID,
		RegionCode:  district.RegionCode,
		CountryID:   district.CountryID,
		CountryCode: district.CountryCode,
		Name:        district.Name,
		IsActive:    district.IsActive,
		CreatedAt:   district.CreatedAt,
		UpdatedAt:   district.UpdatedAt,
	}

	if district.Region != nil {
		response.Region = RegionToResponse(district.Region)
	}
	if district.Country != nil {
		response.Country = CountryToResponse(district.Country)
	}

	return response
}

// CityToResponse converts a City model to CityResponse DTO
func CityToResponse(city *models.City) *dtos.CityResponse {
	if city == nil {
		return nil
	}
	response := &dtos.CityResponse{
		ID:           city.ID,
		Code:         city.Code,
		RegionID:     city.RegionID,
		RegionCode:   city.RegionCode,
		CountryID:    city.CountryID,
		CountryCode:  city.CountryCode,
		DistrictID:   city.DistrictID,
		DistrictCode: city.DistrictCode,
		Name:         city.Name,
		IsActive:     city.IsActive,
		CreatedAt:    city.CreatedAt,
		UpdatedAt:    city.UpdatedAt,
	}

	if city.Region != nil {
		response.Region = RegionToResponse(city.Region)
	}
	if city.Country != nil {
		response.Country = CountryToResponse(city.Country)
	}
	if city.District != nil {
		response.District = DistrictToResponse(city.District)
	}

	return response
}

// AreaToResponse converts an Area model to AreaResponse DTO
func AreaToResponse(area *models.Area) *dtos.AreaResponse {
	if area == nil {
		return nil
	}
	response := &dtos.AreaResponse{
		ID:        area.ID,
		Code:      area.Code,
		CityID:    area.CityID,
		CityCode:  area.CityCode,
		Name:      area.Name,
		IsActive:  area.IsActive,
		CreatedAt: area.CreatedAt,
		UpdatedAt: area.UpdatedAt,
	}

	if area.City != nil {
		response.City = CityToResponse(area.City)
	}

	return response
}

// PostalCodeToResponse converts a PostalCode model to PostalCodeResponse DTO
func PostalCodeToResponse(postalCode *models.PostalCode) *dtos.PostalCodeResponse {
	if postalCode == nil {
		return nil
	}
	response := &dtos.PostalCodeResponse{
		ID:            postalCode.ID,
		Code:          postalCode.Code,
		CountryID:     postalCode.CountryID,
		CountryCode:   postalCode.CountryCode,
		RegionID:      postalCode.RegionID,
		RegionCode:    postalCode.RegionCode,
		CityID:        postalCode.CityID,
		CityCode:      postalCode.CityCode,
		AreaID:        postalCode.AreaID,
		AreaCode:      postalCode.AreaCode,
		LocationScope: postalCode.LocationScope,
		IsActive:      postalCode.IsActive,
		CreatedAt:     postalCode.CreatedAt,
		UpdatedAt:     postalCode.UpdatedAt,
	}

	if postalCode.Country != nil {
		response.Country = CountryToResponse(postalCode.Country)
	}
	if postalCode.Region != nil {
		response.Region = RegionToResponse(postalCode.Region)
	}
	if postalCode.City != nil {
		response.City = CityToResponse(postalCode.City)
	}
	if postalCode.Area != nil {
		response.Area = AreaToResponse(postalCode.Area)
	}

	return response
}

// LocationTypeToResponse converts a LocationType model to LocationTypeResponse DTO
func LocationTypeToResponse(locationType *models.LocationType) *dtos.LocationTypeResponse {
	if locationType == nil {
		return nil
	}
	return &dtos.LocationTypeResponse{
		Code:        locationType.Code,
		Description: locationType.Description,
	}
}

// LocationAliasToResponse converts a LocationAlias model to LocationAliasResponse DTO
func LocationAliasToResponse(alias *models.LocationAlias) *dtos.LocationAliasResponse {
	if alias == nil {
		return nil
	}

	response := &dtos.LocationAliasResponse{
		ID:             alias.ID,
		EntityTypeCode: alias.EntityTypeCode,
		EntityType:     alias.EntityType,
		EntityID:       alias.EntityID,
		EntityCode:     alias.EntityCode,
		AliasName:      alias.AliasName,
		IsPrimary:      alias.IsPrimary,
		IsActive:       alias.IsActive,
		CreatedAt:      alias.CreatedAt,
		UpdatedAt:      alias.UpdatedAt,
	}

	if alias.LocationType != nil {
		response.LocationType = LocationTypeToResponse(alias.LocationType)
	}

	return response
}
