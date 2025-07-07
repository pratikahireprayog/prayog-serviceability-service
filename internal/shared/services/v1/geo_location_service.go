package services

import (
	"context"
	"time"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"github.com/go-playground/validator/v10"
)

// GeoLocationService defines the interface for geo location business logic
type GeoLocationService interface {
	// Only 5 methods for the 5 requested APIs
	GetAll(ctx context.Context, pagination *dtos.PaginationRequest, filters *dtos.GeoLocationFilters) (*dtos.GeoLocationListResponse, error)
	GetByPostalCode(ctx context.Context, postalCode string) (*dtos.GeoLocationSingleResponse, error)
	Create(ctx context.Context, req *dtos.CreateGeoLocationRequest) (*dtos.GeoLocationSingleResponse, error)
	Update(ctx context.Context, postalCode string, req *dtos.UpdateGeoLocationRequest) (*dtos.GeoLocationSingleResponse, error)
	Delete(ctx context.Context, postalCode string) (*dtos.StandardResponse, error)
}

type geoLocationService struct {
	repo      repositories.GeoLocationRepository
	validator *validator.Validate
}

// NewGeoLocationService creates a new geo location service
func NewGeoLocationService(repo repositories.GeoLocationRepository, validator *validator.Validate) GeoLocationService {
	return &geoLocationService{
		repo:      repo,
		validator: validator,
	}
}

// GetAll retrieves all geo locations with pagination
func (s *geoLocationService) GetAll(ctx context.Context, pagination *dtos.PaginationRequest, filters *dtos.GeoLocationFilters) (*dtos.GeoLocationListResponse, error) {
	geoLocations, total, err := s.repo.GetAll(ctx, pagination.Offset, pagination.Limit, filters)
	if err != nil {
		return nil, err
	}

	responses := make([]dtos.GeoLocationResponse, len(geoLocations))
	for i, location := range geoLocations {
		responses[i] = *GeoLocationToResponseDTO(&location)
	}

	paginationResponse := dtos.PaginationResponse{
		Offset:      pagination.Offset,
		Limit:       pagination.Limit,
		Total:       total,
		HasNext:     int64(pagination.Offset+pagination.Limit) < total,
		HasPrevious: pagination.Offset > 0,
	}

	return &dtos.GeoLocationListResponse{
		Success:    true,
		Message:    "Geo locations retrieved successfully",
		Data:       responses,
		Pagination: paginationResponse,
	}, nil
}

// GetByPostalCode retrieves a single geo location by postal code
func (s *geoLocationService) GetByPostalCode(ctx context.Context, postalCode string) (*dtos.GeoLocationSingleResponse, error) {
	geoLocation, err := s.repo.GetByID(ctx, postalCode)
	if err != nil {
		return nil, err
	}

	return &dtos.GeoLocationSingleResponse{
		Success: true,
		Message: "Geo location retrieved successfully",
		Data:    *GeoLocationToResponseDTO(geoLocation),
	}, nil
}

// Create creates a new geo location
func (s *geoLocationService) Create(ctx context.Context, req *dtos.CreateGeoLocationRequest) (*dtos.GeoLocationSingleResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, err
	}

	// Convert DTO to model
	geoLocation := &models.GeoLocation{
		PostalCode:     req.PostalCode,
		Name:           req.Name,
		AsciiName:      req.AsciiName,
		AlternateNames: req.AlternateNames,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		FeatureClass:   req.FeatureClass,
		FeatureCode:    req.FeatureCode,
		CountryCode:    req.CountryCode,
		CC2:            req.CC2,
		Admin1Code:     req.Admin1Code,
		Admin2Code:     req.Admin2Code,
		Admin3Code:     req.Admin3Code,
		Admin4Code:     req.Admin4Code,
		Population:     req.Population,
		Elevation:      req.Elevation,
		DEM:            req.DEM,
		Timezone:       req.Timezone,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if req.ModificationDate != nil {
		geoLocation.ModificationDate = *req.ModificationDate
	} else {
		geoLocation.ModificationDate = time.Now()
	}

	// Create in repository
	err := s.repo.Create(ctx, geoLocation)
	if err != nil {
		return nil, err
	}

	return &dtos.GeoLocationSingleResponse{
		Success: true,
		Message: "Geo location created successfully",
		Data:    *GeoLocationToResponseDTO(geoLocation),
	}, nil
}

// Update updates an existing geo location
func (s *geoLocationService) Update(ctx context.Context, postalCode string, req *dtos.UpdateGeoLocationRequest) (*dtos.GeoLocationSingleResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, err
	}

	// Get existing geo location
	existingGeoLocation, err := s.repo.GetByID(ctx, postalCode)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		existingGeoLocation.Name = *req.Name
	}
	if req.AsciiName != nil {
		existingGeoLocation.AsciiName = *req.AsciiName
	}
	if req.AlternateNames != nil {
		existingGeoLocation.AlternateNames = req.AlternateNames
	}
	if req.Latitude != nil {
		existingGeoLocation.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		existingGeoLocation.Longitude = *req.Longitude
	}
	if req.FeatureClass != nil {
		existingGeoLocation.FeatureClass = *req.FeatureClass
	}
	if req.FeatureCode != nil {
		existingGeoLocation.FeatureCode = *req.FeatureCode
	}
	if req.CountryCode != nil {
		existingGeoLocation.CountryCode = *req.CountryCode
	}
	if req.CC2 != nil {
		existingGeoLocation.CC2 = req.CC2
	}
	if req.Admin1Code != nil {
		existingGeoLocation.Admin1Code = req.Admin1Code
	}
	if req.Admin2Code != nil {
		existingGeoLocation.Admin2Code = req.Admin2Code
	}
	if req.Admin3Code != nil {
		existingGeoLocation.Admin3Code = req.Admin3Code
	}
	if req.Admin4Code != nil {
		existingGeoLocation.Admin4Code = req.Admin4Code
	}
	if req.Population != nil {
		existingGeoLocation.Population = req.Population
	}
	if req.Elevation != nil {
		existingGeoLocation.Elevation = req.Elevation
	}
	if req.DEM != nil {
		existingGeoLocation.DEM = req.DEM
	}
	if req.Timezone != nil {
		existingGeoLocation.Timezone = req.Timezone
	}

	existingGeoLocation.UpdatedAt = time.Now()

	// Update in repository
	err = s.repo.Update(ctx, postalCode, existingGeoLocation)
	if err != nil {
		return nil, err
	}

	return &dtos.GeoLocationSingleResponse{
		Success: true,
		Message: "Geo location updated successfully",
		Data:    *GeoLocationToResponseDTO(existingGeoLocation),
	}, nil
}

// Delete soft deletes a geo location
func (s *geoLocationService) Delete(ctx context.Context, postalCode string) (*dtos.StandardResponse, error) {
	err := s.repo.Delete(ctx, postalCode)
	if err != nil {
		return nil, err
	}

	return &dtos.StandardResponse{
		Success: true,
		Message: "Geo location deleted successfully",
	}, nil
}

// GeoLocationToResponseDTO converts a GeoLocation model to response DTO
func GeoLocationToResponseDTO(geoLocation *models.GeoLocation) *dtos.GeoLocationResponse {
	return &dtos.GeoLocationResponse{
		PostalCode:       geoLocation.PostalCode,
		Name:             geoLocation.Name,
		AsciiName:        geoLocation.AsciiName,
		AlternateNames:   geoLocation.AlternateNames,
		Latitude:         geoLocation.Latitude,
		Longitude:        geoLocation.Longitude,
		FeatureClass:     geoLocation.FeatureClass,
		FeatureCode:      geoLocation.FeatureCode,
		CountryCode:      geoLocation.CountryCode,
		CC2:              geoLocation.CC2,
		Admin1Code:       geoLocation.Admin1Code,
		Admin2Code:       geoLocation.Admin2Code,
		Admin3Code:       geoLocation.Admin3Code,
		Admin4Code:       geoLocation.Admin4Code,
		Population:       geoLocation.Population,
		Elevation:        geoLocation.Elevation,
		DEM:              geoLocation.DEM,
		Timezone:         geoLocation.Timezone,
		ModificationDate: geoLocation.ModificationDate,
		CreatedAt:        geoLocation.CreatedAt,
		UpdatedAt:        geoLocation.UpdatedAt,
	}
}

// GeoLocationToResponse converts a GeoLocation model to single response format (kept for backward compatibility)
func GeoLocationToResponse(geoLocation *models.GeoLocation) *dtos.GeoLocationSingleResponse {
	return &dtos.GeoLocationSingleResponse{
		Success: true,
		Message: "Geo location retrieved successfully",
		Data:    *GeoLocationToResponseDTO(geoLocation),
	}
}
