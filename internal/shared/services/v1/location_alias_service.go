package services

import (
	"context"
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/shared/dtos/v1"
	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"github.com/sirupsen/logrus"
)

// LocationAliasService defines the business logic interface for location alias operations
type LocationAliasService interface {
	GetByID(ctx context.Context, id string) (*dtos.LocationAliasResponse, error)
	GetByEntityID(ctx context.Context, entityID string) ([]dtos.LocationAliasResponse, error)
	GetByEntityTypeAndID(ctx context.Context, entityType string, entityID string) ([]dtos.LocationAliasResponse, error)
	GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.LocationAliasListResponse, error)
	Create(ctx context.Context, req *dtos.CreateLocationAliasRequest) (*dtos.LocationAliasResponse, error)
	Update(ctx context.Context, id string, req *dtos.UpdateLocationAliasRequest) (*dtos.LocationAliasResponse, error)
	Delete(ctx context.Context, id string) error
	ValidateEntityExists(ctx context.Context, entityType string, entityID string) error
}

// locationAliasService implements LocationAliasService
type locationAliasService struct {
	locationAliasRepo repositories.LocationAliasRepository
	locationRepo      repositories.LocationRepository
	logger            *logrus.Logger
}

// NewLocationAliasService creates a new location alias service
func NewLocationAliasService(
	locationAliasRepo repositories.LocationAliasRepository,
	locationRepo repositories.LocationRepository,
	logger *logrus.Logger,
) LocationAliasService {
	return &locationAliasService{
		locationAliasRepo: locationAliasRepo,
		locationRepo:      locationRepo,
		logger:            logger,
	}
}

// GetByID retrieves a location alias by ID
func (s *locationAliasService) GetByID(ctx context.Context, id string) (*dtos.LocationAliasResponse, error) {
	alias, err := s.locationAliasRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get location alias by ID")
		return nil, err
	}

	return LocationAliasToResponse(alias), nil
}

// GetByEntityID retrieves all aliases for a specific entity ID
func (s *locationAliasService) GetByEntityID(ctx context.Context, entityID string) ([]dtos.LocationAliasResponse, error) {
	aliases, err := s.locationAliasRepo.GetAllByEntityID(ctx, entityID)
	if err != nil {
		s.logger.WithError(err).WithField("entityID", entityID).Error("Failed to get location aliases by entity ID")
		return nil, err
	}

	responses := make([]dtos.LocationAliasResponse, len(aliases))
	for i, alias := range aliases {
		responses[i] = *LocationAliasToResponse(&alias)
	}

	return responses, nil
}

// GetByEntityTypeAndID retrieves all aliases for a specific entity type and ID
func (s *locationAliasService) GetByEntityTypeAndID(ctx context.Context, entityType string, entityID string) ([]dtos.LocationAliasResponse, error) {
	aliases, err := s.locationAliasRepo.GetByEntityTypeAndID(ctx, entityType, entityID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"entityType": entityType,
			"entityID":   entityID,
		}).Error("Failed to get location aliases by entity type and ID")
		return nil, err
	}

	responses := make([]dtos.LocationAliasResponse, len(aliases))
	for i, alias := range aliases {
		responses[i] = *LocationAliasToResponse(&alias)
	}

	return responses, nil
}

// GetAll retrieves all location aliases with pagination
func (s *locationAliasService) GetAll(ctx context.Context, req *dtos.PaginationRequest) (*dtos.LocationAliasListResponse, error) {
	aliases, total, err := s.locationAliasRepo.GetAll(ctx, req.Offset, req.Limit)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get all location aliases")
		return nil, err
	}

	responses := make([]dtos.LocationAliasResponse, len(aliases))
	for i, alias := range aliases {
		responses[i] = *LocationAliasToResponse(&alias)
	}

	hasNext := int64(req.Offset+req.Limit) < total
	hasPrevious := req.Offset > 0

	return &dtos.LocationAliasListResponse{
		Success: true,
		Message: "Location aliases retrieved successfully",
		Data:    responses,
		Pagination: dtos.PaginationResponse{
			Offset:      req.Offset,
			Limit:       req.Limit,
			Total:       total,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	}, nil
}

// Create creates a new location alias
func (s *locationAliasService) Create(ctx context.Context, req *dtos.CreateLocationAliasRequest) (*dtos.LocationAliasResponse, error) {
	// Validate entity exists if entity type is provided
	if req.EntityType != nil && *req.EntityType != "" {
		if err := s.ValidateEntityExists(ctx, *req.EntityType, req.EntityID.String()); err != nil {
			return nil, err
		}
	}

	// Convert DTO to model
	alias := &models.LocationAlias{
		EntityTypeCode: req.EntityTypeCode,
		EntityType:     req.EntityType,
		EntityID:       req.EntityID,
		EntityCode:     req.EntityCode,
		AliasName:      req.AliasName,
		IsPrimary:      req.IsPrimary != nil && *req.IsPrimary,
		IsActive:       req.IsActive == nil || *req.IsActive, // Default to true
	}

	// Create the alias
	if err := s.locationAliasRepo.Create(ctx, alias); err != nil {
		s.logger.WithError(err).WithField("alias_name", req.AliasName).Error("Failed to create location alias")
		return nil, fmt.Errorf("failed to create location alias: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"id":         alias.ID,
		"alias_name": alias.AliasName,
		"entity_id":  alias.EntityID,
	}).Info("Location alias created successfully")

	return LocationAliasToResponse(alias), nil
}

// Update updates an existing location alias
func (s *locationAliasService) Update(ctx context.Context, id string, req *dtos.UpdateLocationAliasRequest) (*dtos.LocationAliasResponse, error) {
	// Get existing alias
	existing, err := s.locationAliasRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.EntityTypeCode != nil {
		existing.EntityTypeCode = req.EntityTypeCode
	}
	if req.EntityType != nil {
		existing.EntityType = req.EntityType
	}
	if req.EntityCode != nil {
		existing.EntityCode = req.EntityCode
	}
	if req.AliasName != nil {
		existing.AliasName = *req.AliasName
	}
	if req.IsPrimary != nil {
		existing.IsPrimary = *req.IsPrimary
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	// Validate entity exists if entity type is provided
	if existing.EntityType != nil && *existing.EntityType != "" {
		if err := s.ValidateEntityExists(ctx, *existing.EntityType, existing.EntityID.String()); err != nil {
			return nil, err
		}
	}

	// Update the alias
	if err := s.locationAliasRepo.Update(ctx, existing); err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to update location alias")
		return nil, fmt.Errorf("failed to update location alias: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"id":         existing.ID,
		"alias_name": existing.AliasName,
	}).Info("Location alias updated successfully")

	return LocationAliasToResponse(existing), nil
}

// Delete deletes a location alias
func (s *locationAliasService) Delete(ctx context.Context, id string) error {
	// Check if alias exists
	_, err := s.locationAliasRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete the alias
	if err := s.locationAliasRepo.Delete(ctx, id); err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to delete location alias")
		return fmt.Errorf("failed to delete location alias: %w", err)
	}

	s.logger.WithField("id", id).Info("Location alias deleted successfully")
	return nil
}

// ValidateEntityExists validates that the referenced entity exists
func (s *locationAliasService) ValidateEntityExists(ctx context.Context, entityType string, entityID string) error {
	switch strings.ToLower(entityType) {
	case "country":
		_, err := s.locationRepo.Countries().GetByID(ctx, entityID)
		if err != nil {
			return fmt.Errorf("country with ID %s not found: %w", entityID, err)
		}
	case "region":
		_, err := s.locationRepo.Regions().GetByID(ctx, entityID)
		if err != nil {
			return fmt.Errorf("region with ID %s not found: %w", entityID, err)
		}
	case "district":
		_, err := s.locationRepo.Districts().GetByID(ctx, entityID)
		if err != nil {
			return fmt.Errorf("district with ID %s not found: %w", entityID, err)
		}
	case "city":
		_, err := s.locationRepo.Cities().GetByID(ctx, entityID)
		if err != nil {
			return fmt.Errorf("city with ID %s not found: %w", entityID, err)
		}
	case "area":
		_, err := s.locationRepo.Areas().GetByID(ctx, entityID)
		if err != nil {
			return fmt.Errorf("area with ID %s not found: %w", entityID, err)
		}
	case "postal_code":
		_, err := s.locationRepo.PostalCodes().GetByID(ctx, entityID)
		if err != nil {
			return fmt.Errorf("postal code with ID %s not found: %w", entityID, err)
		}
	default:
		// For unknown entity types, we'll allow it but log a warning
		s.logger.WithFields(logrus.Fields{
			"entity_type": entityType,
			"entity_id":   entityID,
		}).Warn("Unknown entity type for validation")
	}

	return nil
}
