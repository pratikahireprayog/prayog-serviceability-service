// Package repository contains data storage access layer interfaces
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/prayog/serviceability/internal/domain"
)

// LocationRepository interface for location data operations
type LocationRepository interface {
	GetLocationByPostalCode(ctx context.Context, postalCode string) (*domain.Location, error)
	GetLocationHierarchy(ctx context.Context, postalCode string) (*domain.ServiceabilityLocationHierarchy, error)
	ImportLocations(ctx context.Context, data []byte, format string) (int, error)
	ExportLocations(ctx context.Context, format string) ([]byte, error)
}

// ServiceTypeRepository interface for service type operations
type ServiceTypeRepository interface {
	Create(ctx context.Context, serviceType *domain.ServiceType) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ServiceType, error)
	Update(ctx context.Context, serviceType *domain.ServiceType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// OrderTypeRepository interface for order type operations
type OrderTypeRepository interface {
	Create(ctx context.Context, orderType *domain.OrderType) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderType, error)
	Update(ctx context.Context, orderType *domain.OrderType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ServiceAvailabilityRepository interface for service availability operations
type ServiceAvailabilityRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ServiceAvailability, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*domain.ServiceAvailability, error)
	Create(ctx context.Context, serviceAvailability *domain.ServiceAvailability) error
	Update(ctx context.Context, serviceAvailability *domain.ServiceAvailability) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetServiceableAreas(ctx context.Context, serviceTypeCode string) ([]domain.Area, error)
}

// TimeRuleRepository interface for time rule operations
type TimeRuleRepository interface {
	Create(ctx context.Context, timeRule *domain.TimeRule) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TimeRule, error)
	Update(ctx context.Context, timeRule *domain.TimeRule) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByLocationID(ctx context.Context, locationID uuid.UUID) ([]*domain.TimeRule, error)
}

// Factory interface for creating repositories
type Factory interface {
	LocationRepository() LocationRepository
	ServiceTypeRepository() ServiceTypeRepository
	OrderTypeRepository() OrderTypeRepository
	ServiceAvailabilityRepository() ServiceAvailabilityRepository
	TimeRuleRepository() TimeRuleRepository
}
