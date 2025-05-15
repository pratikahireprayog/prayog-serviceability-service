package services

import (
	"context"
	"prayog-serviceability-service/pkg/models"
	"prayog-serviceability-service/pkg/repository"
)

// ServiceabilityService defines the interface for serviceability operations
type ServiceabilityService interface {
	CheckServiceability(ctx context.Context, request models.ServiceabilityRequest) (models.ServiceabilityResponse, error)
	BulkCheckServiceability(ctx context.Context, postalCodes []string) ([]models.ServiceabilityResponse, error)
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
		PostalCode:    request.PostalCode,
		IsServiceable: false,
		Message:       "Service not available in your area",
	}, nil
}

// BulkCheckServiceability checks if services are available for multiple postal codes
func (s *serviceabilityService) BulkCheckServiceability(ctx context.Context, postalCodes []string) ([]models.ServiceabilityResponse, error) {
	results := make([]models.ServiceabilityResponse, 0, len(postalCodes))

	for _, code := range postalCodes {
		// Create a request for each postal code
		request := models.ServiceabilityRequest{
			PostalCode: code,
		}

		// Check serviceability for each postal code
		result, _ := s.CheckServiceability(ctx, request)
		results = append(results, result)
	}

	return results, nil
}
