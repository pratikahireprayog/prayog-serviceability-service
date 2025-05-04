// Package usecase contains application business rules and use cases
package usecase

import (
	"context"

	"prayog-serviceability-service/internal/repository"
)

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
