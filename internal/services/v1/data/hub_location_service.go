package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"prayog-serviceability-service/internal/shared/errors"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"github.com/sirupsen/logrus"
)

// HubLocationService defines the interface for hub location operations
type HubLocationService interface {
	GetNearestHubByPostalCode(ctx context.Context, postalCode string) (*models.HubLocationInfo, error)
	GetNearestHubByPostalCodeInt(ctx context.Context, postalCode int) (*models.HubLocationInfo, error)
	ValidateHubAvailability(ctx context.Context, postalCode string) (bool, error)
	GetHubInfo(ctx context.Context, postalCode string) (*models.HubInfo, error)
	GetHubLocationsByCity(ctx context.Context, hubCityCode string) ([]models.HubLocationInfo, error)
	GetNearbyHubLocations(ctx context.Context, latitude, longitude float64, radiusKm int) ([]models.HubLocationInfo, error)
}

// hubLocationService implements the HubLocationService interface
type hubLocationService struct {
	hubLocationRepo repositories.NearestHubLocationRepository
	logger          *logrus.Logger
}

// NewHubLocationService creates a new hub location service instance
func NewHubLocationService(
	hubLocationRepo repositories.NearestHubLocationRepository,
	logger *logrus.Logger,
) HubLocationService {
	return &hubLocationService{
		hubLocationRepo: hubLocationRepo,
		logger:          logger,
	}
}

// GetNearestHubByPostalCode retrieves the nearest hub for a given postal code
func (h *hubLocationService) GetNearestHubByPostalCode(ctx context.Context, postalCode string) (*models.HubLocationInfo, error) {
	if strings.TrimSpace(postalCode) == "" {
		return nil, errors.ErrValidationFailed("postal_code", "cannot be empty")
	}

	h.logger.WithFields(logrus.Fields{
		"service":     "HubLocationService",
		"method":      "GetNearestHubByPostalCode",
		"postal_code": postalCode,
	}).Debug("Looking up nearest hub for postal code")

	// Check if hub location repository is available
	if h.hubLocationRepo == nil {
		return nil, errors.ErrServiceUnavailable("hub location repository")
	}

	// Get hub location data by postal code
	hubLocation, err := h.hubLocationRepo.GetByPostalCodeString(ctx, postalCode)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"service":     "HubLocationService",
			"postal_code": postalCode,
			"error":       err.Error(),
		}).Error("Failed to get nearest hub location")

		// Return a structured error for not found case
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.ErrNotFound("nearest hub location", fmt.Sprintf("postal code %s", postalCode))
		}
		return nil, errors.ErrInternalError("failed to get nearest hub location", err)
	}

	h.logger.WithFields(logrus.Fields{
		"service":              "HubLocationService",
		"postal_code":          postalCode,
		"city_code":            hubLocation.CityCode,
		"hub_postal_code":      hubLocation.HubPostalCode,
		"hub_city_code":        hubLocation.HubCityCode,
		"is_international_hub": hubLocation.IsInternationalHub,
	}).Info("Successfully found nearest hub location")

	return hubLocation.ToHubLocationInfo(), nil
}

// GetNearestHubByPostalCodeInt retrieves the nearest hub for a given postal code as integer
func (h *hubLocationService) GetNearestHubByPostalCodeInt(ctx context.Context, postalCode int) (*models.HubLocationInfo, error) {
	if postalCode <= 0 {
		return nil, errors.ErrValidationFailed("postal_code", "must be positive")
	}

	h.logger.WithFields(logrus.Fields{
		"service":     "HubLocationService",
		"method":      "GetNearestHubByPostalCodeInt",
		"postal_code": postalCode,
	}).Debug("Looking up nearest hub for postal code")

	// Check if hub location repository is available
	if h.hubLocationRepo == nil {
		return nil, errors.ErrServiceUnavailable("hub location repository")
	}

	// Get hub location data by postal code
	hubLocation, err := h.hubLocationRepo.GetByPostalCode(ctx, postalCode)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"service":     "HubLocationService",
			"postal_code": postalCode,
			"error":       err.Error(),
		}).Error("Failed to get nearest hub location")

		// Return a structured error for not found case
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.ErrNotFound("nearest hub location", fmt.Sprintf("postal code %d", postalCode))
		}
		return nil, errors.ErrInternalError("failed to get nearest hub location", err)
	}

	h.logger.WithFields(logrus.Fields{
		"service":              "HubLocationService",
		"postal_code":          postalCode,
		"city_code":            hubLocation.CityCode,
		"hub_postal_code":      hubLocation.HubPostalCode,
		"hub_city_code":        hubLocation.HubCityCode,
		"is_international_hub": hubLocation.IsInternationalHub,
	}).Info("Successfully found nearest hub location")

	return hubLocation.ToHubLocationInfo(), nil
}

// ValidateHubAvailability checks if a hub is available for the given postal code
func (h *hubLocationService) ValidateHubAvailability(ctx context.Context, postalCode string) (bool, error) {
	if strings.TrimSpace(postalCode) == "" {
		return false, errors.ErrValidationFailed("postal_code", "cannot be empty")
	}

	h.logger.WithFields(logrus.Fields{
		"service":     "HubLocationService",
		"method":      "ValidateHubAvailability",
		"postal_code": postalCode,
	}).Debug("Validating hub availability for postal code")

	// Check if hub location repository is available
	if h.hubLocationRepo == nil {
		return false, errors.ErrServiceUnavailable("hub location repository")
	}

	// Convert postal code to int for existence check
	postalCodeInt, err := strconv.Atoi(postalCode)
	if err != nil {
		return false, errors.ErrValidationFailed("postal_code", "invalid format")
	}

	// Check if hub exists
	exists, err := h.hubLocationRepo.Exists(ctx, postalCodeInt)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"service":     "HubLocationService",
			"postal_code": postalCode,
			"error":       err.Error(),
		}).Error("Failed to check hub availability")
		return false, errors.ErrInternalError("failed to check hub availability", err)
	}

	h.logger.WithFields(logrus.Fields{
		"service":     "HubLocationService",
		"postal_code": postalCode,
		"available":   exists,
	}).Debug("Hub availability check completed")

	return exists, nil
}

// GetHubInfo retrieves hub information for a given postal code
func (h *hubLocationService) GetHubInfo(ctx context.Context, postalCode string) (*models.HubInfo, error) {
	if strings.TrimSpace(postalCode) == "" {
		return nil, errors.ErrValidationFailed("postal_code", "cannot be empty")
	}

	h.logger.WithFields(logrus.Fields{
		"service":     "HubLocationService",
		"method":      "GetHubInfo",
		"postal_code": postalCode,
	}).Debug("Getting hub info for postal code")

	// Get hub location info
	hubLocationInfo, err := h.GetNearestHubByPostalCode(ctx, postalCode)
	if err != nil {
		return nil, err
	}

	// Return hub info if available
	if hubLocationInfo.HubInfo != nil {
		return hubLocationInfo.HubInfo, nil
	}

	// If no hub is found, return an error
	return nil, errors.ErrNotFound("hub", fmt.Sprintf("postal code %s", postalCode))
}

// GetHubLocationsByCity retrieves all hub locations for a given city code
func (h *hubLocationService) GetHubLocationsByCity(ctx context.Context, hubCityCode string) ([]models.HubLocationInfo, error) {
	if strings.TrimSpace(hubCityCode) == "" {
		return nil, errors.ErrValidationFailed("hub_city_code", "cannot be empty")
	}

	h.logger.WithFields(logrus.Fields{
		"service":       "HubLocationService",
		"method":        "GetHubLocationsByCity",
		"hub_city_code": hubCityCode,
	}).Debug("Getting hub locations by city code")

	// Check if hub location repository is available
	if h.hubLocationRepo == nil {
		return nil, errors.ErrServiceUnavailable("hub location repository")
	}

	// Get hub locations by city code
	hubLocations, err := h.hubLocationRepo.GetByHubCityCode(ctx, hubCityCode)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"service":       "HubLocationService",
			"hub_city_code": hubCityCode,
			"error":         err.Error(),
		}).Error("Failed to get hub locations by city code")
		return nil, errors.ErrInternalError("failed to get hub locations by city code", err)
	}

	// Convert to response format
	hubInfos := make([]models.HubLocationInfo, len(hubLocations))
	for i, hubLocation := range hubLocations {
		hubInfos[i] = *hubLocation.ToHubLocationInfo()
	}

	h.logger.WithFields(logrus.Fields{
		"service":       "HubLocationService",
		"hub_city_code": hubCityCode,
		"count":         len(hubInfos),
	}).Info("Successfully retrieved hub locations by city code")

	return hubInfos, nil
}

// GetNearbyHubLocations retrieves hub locations within a specified radius
func (h *hubLocationService) GetNearbyHubLocations(ctx context.Context, latitude, longitude float64, radiusKm int) ([]models.HubLocationInfo, error) {
	if radiusKm <= 0 {
		return nil, errors.ErrValidationFailed("radius", "must be positive")
	}

	h.logger.WithFields(logrus.Fields{
		"service":   "HubLocationService",
		"method":    "GetNearbyHubLocations",
		"latitude":  latitude,
		"longitude": longitude,
		"radius":    radiusKm,
	}).Debug("Getting nearby hub locations")

	// Check if hub location repository is available
	if h.hubLocationRepo == nil {
		return nil, errors.ErrServiceUnavailable("hub location repository")
	}

	// Get nearby hub locations (limit to 50 for performance)
	hubLocations, err := h.hubLocationRepo.GetNearbyHubLocations(ctx, latitude, longitude, radiusKm, 50)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"service":   "HubLocationService",
			"latitude":  latitude,
			"longitude": longitude,
			"radius":    radiusKm,
			"error":     err.Error(),
		}).Error("Failed to get nearby hub locations")
		return nil, errors.ErrInternalError("failed to get nearby hub locations", err)
	}

	// Convert to response format
	hubInfos := make([]models.HubLocationInfo, len(hubLocations))
	for i, hubLocation := range hubLocations {
		hubInfos[i] = *hubLocation.ToHubLocationInfo()
	}

	h.logger.WithFields(logrus.Fields{
		"service":   "HubLocationService",
		"latitude":  latitude,
		"longitude": longitude,
		"radius":    radiusKm,
		"count":     len(hubInfos),
	}).Info("Successfully retrieved nearby hub locations")

	return hubInfos, nil
}
