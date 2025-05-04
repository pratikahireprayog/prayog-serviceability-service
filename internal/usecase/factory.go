// Package usecase contains application business rules and use cases
package usecase

import (
	"context"

	"github.com/prayog/serviceability/internal/domain"
	"github.com/prayog/serviceability/internal/repository"
)

// RepositoryFactory defines an interface for creating repositories
type RepositoryFactory interface {
	Country() domain.CountryRepository
	Region() domain.RegionRepository
	City() domain.CityRepository
	Area() domain.AreaRepository
	PostalCode() domain.PostalCodeRepository
	OrderType() domain.OrderTypeRepository
	ServiceType() domain.ServiceTypeRepository
	ServiceAvailability() domain.ServiceAvailabilityRepository
}

// Factory creates all the use case services
type Factory interface {
	NewServiceabilityUseCase(ctx context.Context) ServiceabilityUseCase
}

// factoryImpl implements the Factory interface
type factoryImpl struct {
	repoFactory repository.Factory
}

// NewFactory creates a new usecase factory
func NewFactory(repoFactory repository.Factory) Factory {
	return &factoryImpl{
		repoFactory: repoFactory,
	}
}

// NewServiceabilityUseCase creates a new ServiceabilityUseCase
func (f *factoryImpl) NewServiceabilityUseCase(ctx context.Context) ServiceabilityUseCase {
	return NewServiceabilityUseCase(
		ctx,
		f.repoFactory.LocationRepository(),
		f.repoFactory.ServiceTypeRepository(),
		f.repoFactory.OrderTypeRepository(),
		f.repoFactory.ServiceAvailabilityRepository(),
		f.repoFactory.TimeRuleRepository(),
	)
}
