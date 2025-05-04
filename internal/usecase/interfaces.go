// Package usecase contains application business rules and use cases
package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/prayog/serviceability/internal/domain"
)

// ServiceabilityUseCase defines the interface for serviceability operations
type ServiceabilityUseCase interface {
	// Core API functionality
	CheckServiceability(ctx context.Context, request domain.ServiceabilityRequest) (*domain.ServiceabilityResult, error)
	BulkCheckServiceability(ctx context.Context, requests []domain.ServiceabilityRequest) ([]domain.ServiceabilityResult, error)
	GetServiceableAreas(ctx context.Context, serviceTypeCode string) ([]domain.Area, error)

	// Rule management
	GetServiceAvailabilityByID(ctx context.Context, id uuid.UUID) (*domain.ServiceAvailability, error)
	ListServiceAvailabilities(ctx context.Context, filters map[string]interface{}) ([]*domain.ServiceAvailability, error)
	CreateServiceAvailability(ctx context.Context, serviceAvailability *domain.ServiceAvailability) error
	UpdateServiceAvailability(ctx context.Context, serviceAvailability *domain.ServiceAvailability) error
	DeleteServiceAvailability(ctx context.Context, id uuid.UUID) error

	// Location data management
	ImportLocations(ctx context.Context, data []byte, format string) (int, error)
	ExportLocations(ctx context.Context, format string) ([]byte, error)
	GetLocationHierarchy(ctx context.Context, postalCode string) (*domain.ServiceabilityLocationHierarchy, error)

	// Service and order type management
	CreateOrderType(ctx context.Context, orderType *domain.OrderType) error
	GetOrderTypeByID(ctx context.Context, id uuid.UUID) (*domain.OrderType, error)
	UpdateOrderType(ctx context.Context, orderType *domain.OrderType) error
	DeleteOrderType(ctx context.Context, id uuid.UUID) error

	CreateServiceType(ctx context.Context, serviceType *domain.ServiceType) error
	GetServiceTypeByID(ctx context.Context, id uuid.UUID) (*domain.ServiceType, error)
	UpdateServiceType(ctx context.Context, serviceType *domain.ServiceType) error
	DeleteServiceType(ctx context.Context, id uuid.UUID) error

	// Time rule management
	CreateTimeRule(ctx context.Context, timeRule *domain.TimeRule) error
	GetTimeRuleByID(ctx context.Context, id uuid.UUID) (*domain.TimeRule, error)
	UpdateTimeRule(ctx context.Context, timeRule *domain.TimeRule) error
	DeleteTimeRule(ctx context.Context, id uuid.UUID) error
	GetTimeRulesByLocationID(ctx context.Context, locationID uuid.UUID) ([]*domain.TimeRule, error)
}

// Factory is a factory for creating use cases
type Factory interface {
	NewServiceabilityUseCase(ctx context.Context) ServiceabilityUseCase
}
