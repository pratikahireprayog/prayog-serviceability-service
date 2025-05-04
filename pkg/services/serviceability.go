package services

import (
	"context"
	"prayog-serviceability-service/pkg/models"
	"prayog-serviceability-service/pkg/repository"
)

// ServiceabilityService defines the interface for serviceability operations
type ServiceabilityService interface {
	CheckServiceability(ctx context.Context, request models.ServiceabilityRequest) (models.ServiceabilityResponse, error)
}

// serviceabilityService implements the ServiceabilityService interface
type serviceabilityService struct {
	repositories *repository.RepositoryFactory
}

// NewServiceabilityService creates a new serviceability service
func NewServiceabilityService(repositories *repository.RepositoryFactory) ServiceabilityService {
	return &serviceabilityService{
		repositories: repositories,
	}
}

// CheckServiceability checks if a service is available for a given location
func (s *serviceabilityService) CheckServiceability(ctx context.Context, request models.ServiceabilityRequest) (models.ServiceabilityResponse, error) {
	// Implementation TBD
	return models.ServiceabilityResponse{
		IsServiceable: false,
		Message:       "Service not available in your area",
	}, nil
}
