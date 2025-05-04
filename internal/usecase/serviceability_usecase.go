// Package usecase contains application business rules and use cases
package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/prayog/serviceability/internal/domain"
	"github.com/prayog/serviceability/internal/repository"
)

// serviceabilityUseCase implements the ServiceabilityUseCase interface
type serviceabilityUseCase struct {
	ctx                     context.Context
	locationRepo            repository.LocationRepository
	serviceTypeRepo         repository.ServiceTypeRepository
	orderTypeRepo           repository.OrderTypeRepository
	serviceAvailabilityRepo repository.ServiceAvailabilityRepository
	timeRuleRepo            repository.TimeRuleRepository
}

// NewServiceabilityUseCase creates a new serviceability use case
func NewServiceabilityUseCase(
	ctx context.Context,
	locationRepo repository.LocationRepository,
	serviceTypeRepo repository.ServiceTypeRepository,
	orderTypeRepo repository.OrderTypeRepository,
	serviceAvailabilityRepo repository.ServiceAvailabilityRepository,
	timeRuleRepo repository.TimeRuleRepository,
) ServiceabilityUseCase {
	return &serviceabilityUseCase{
		ctx:                     ctx,
		locationRepo:            locationRepo,
		serviceTypeRepo:         serviceTypeRepo,
		orderTypeRepo:           orderTypeRepo,
		serviceAvailabilityRepo: serviceAvailabilityRepo,
		timeRuleRepo:            timeRuleRepo,
	}
}

// CheckServiceability checks if a location is serviceable for a given service type and order type
func (s *serviceabilityUseCase) CheckServiceability(ctx context.Context, request domain.ServiceabilityRequest) (*domain.ServiceabilityResult, error) {
	// Get location by postal code
	location, err := s.locationRepo.GetLocationByPostalCode(ctx, request.PostalCode)
	if err != nil {
		return &domain.ServiceabilityResult{
			PostalCode:    request.PostalCode,
			CountryCode:   request.CountryCode,
			OrderTypeCode: request.OrderTypeCode,
			Services:      []domain.ServiceAvailable{},
			IsServiceable: false,
		}, nil
	}

	// Check if service availability exists for the location
	availabilities, err := s.serviceAvailabilityRepo.List(ctx, map[string]interface{}{
		"locationId":      location.ID,
		"serviceTypeCode": request.ServiceTypeCode,
		"orderTypeCode":   request.OrderTypeCode,
	})
	if err != nil {
		return nil, err
	}

	if len(availabilities) == 0 {
		return &domain.ServiceabilityResult{
			PostalCode:    request.PostalCode,
			CountryCode:   request.CountryCode,
			OrderTypeCode: request.OrderTypeCode,
			Services:      []domain.ServiceAvailable{},
			IsServiceable: false,
		}, nil
	}

	// Check time rules if applicable
	timeRules, err := s.timeRuleRepo.GetByLocationID(ctx, location.ID)
	if err != nil {
		return nil, err
	}

	// Check time rules for availability
	isAvailable := true

	if len(timeRules) > 0 {
		// Default to unavailable when time rules exist (must match a rule to be available)
		isAvailable = false
		now := time.Now()
		for _, rule := range timeRules {
			// Check if the current time is within any time rule's valid period
			if now.After(rule.StartTime) && now.Before(rule.EndTime) {
				isAvailable = rule.Available
				break
			}
		}
	}

	// Get service type details for the response
	var serviceType *domain.ServiceType
	if request.ServiceTypeCode != "" {
		serviceType, err = s.serviceTypeRepo.GetByID(ctx, availabilities[0].ServiceTypeID)
		if err != nil {
			return nil, err
		}
	}

	// Create response with service availability
	services := []domain.ServiceAvailable{
		{
			ServiceTypeCode: serviceType.Code,
			ServiceTypeName: serviceType.Name,
			IsAvailable:     isAvailable,
		},
	}

	// If no time rules applicable or no time rules exist
	return &domain.ServiceabilityResult{
		PostalCode:    request.PostalCode,
		CountryCode:   request.CountryCode,
		OrderTypeCode: request.OrderTypeCode,
		Services:      services,
		IsServiceable: isAvailable,
	}, nil
}

// BulkCheckServiceability checks serviceability for multiple requests at once
func (s *serviceabilityUseCase) BulkCheckServiceability(ctx context.Context, requests []domain.ServiceabilityRequest) ([]domain.ServiceabilityResult, error) {
	results := make([]domain.ServiceabilityResult, len(requests))
	for i, request := range requests {
		result, err := s.CheckServiceability(ctx, request)
		if err != nil {
			return nil, err
		}
		results[i] = *result
	}
	return results, nil
}

// GetServiceableAreas returns all areas where a specific service type is available
func (s *serviceabilityUseCase) GetServiceableAreas(ctx context.Context, serviceTypeCode string) ([]domain.Area, error) {
	return s.serviceAvailabilityRepo.GetServiceableAreas(ctx, serviceTypeCode)
}

// GetServiceAvailabilityByID finds a service availability by ID
func (s *serviceabilityUseCase) GetServiceAvailabilityByID(ctx context.Context, id uuid.UUID) (*domain.ServiceAvailability, error) {
	return s.serviceAvailabilityRepo.GetByID(ctx, id)
}

// ListServiceAvailabilities lists service availabilities with optional filters
func (s *serviceabilityUseCase) ListServiceAvailabilities(ctx context.Context, filters map[string]interface{}) ([]*domain.ServiceAvailability, error) {
	return s.serviceAvailabilityRepo.List(ctx, filters)
}

// CreateServiceAvailability creates a new service availability
func (s *serviceabilityUseCase) CreateServiceAvailability(ctx context.Context, serviceAvailability *domain.ServiceAvailability) error {
	return s.serviceAvailabilityRepo.Create(ctx, serviceAvailability)
}

// UpdateServiceAvailability updates an existing service availability
func (s *serviceabilityUseCase) UpdateServiceAvailability(ctx context.Context, serviceAvailability *domain.ServiceAvailability) error {
	return s.serviceAvailabilityRepo.Update(ctx, serviceAvailability)
}

// DeleteServiceAvailability deletes a service availability
func (s *serviceabilityUseCase) DeleteServiceAvailability(ctx context.Context, id uuid.UUID) error {
	return s.serviceAvailabilityRepo.Delete(ctx, id)
}

// ImportLocations imports location data from a file
func (s *serviceabilityUseCase) ImportLocations(ctx context.Context, data []byte, format string) (int, error) {
	return s.locationRepo.ImportLocations(ctx, data, format)
}

// ExportLocations exports location data to a file
func (s *serviceabilityUseCase) ExportLocations(ctx context.Context, format string) ([]byte, error) {
	return s.locationRepo.ExportLocations(ctx, format)
}

// GetLocationHierarchy gets the complete location hierarchy for a postal code
func (s *serviceabilityUseCase) GetLocationHierarchy(ctx context.Context, postalCode string) (*domain.ServiceabilityLocationHierarchy, error) {
	return s.locationRepo.GetLocationHierarchy(ctx, postalCode)
}

// OrderType management
func (s *serviceabilityUseCase) CreateOrderType(ctx context.Context, orderType *domain.OrderType) error {
	return s.orderTypeRepo.Create(ctx, orderType)
}

func (s *serviceabilityUseCase) GetOrderTypeByID(ctx context.Context, id uuid.UUID) (*domain.OrderType, error) {
	return s.orderTypeRepo.GetByID(ctx, id)
}

func (s *serviceabilityUseCase) UpdateOrderType(ctx context.Context, orderType *domain.OrderType) error {
	return s.orderTypeRepo.Update(ctx, orderType)
}

func (s *serviceabilityUseCase) DeleteOrderType(ctx context.Context, id uuid.UUID) error {
	return s.orderTypeRepo.Delete(ctx, id)
}

// ServiceType management
func (s *serviceabilityUseCase) CreateServiceType(ctx context.Context, serviceType *domain.ServiceType) error {
	return s.serviceTypeRepo.Create(ctx, serviceType)
}

func (s *serviceabilityUseCase) GetServiceTypeByID(ctx context.Context, id uuid.UUID) (*domain.ServiceType, error) {
	return s.serviceTypeRepo.GetByID(ctx, id)
}

func (s *serviceabilityUseCase) UpdateServiceType(ctx context.Context, serviceType *domain.ServiceType) error {
	return s.serviceTypeRepo.Update(ctx, serviceType)
}

func (s *serviceabilityUseCase) DeleteServiceType(ctx context.Context, id uuid.UUID) error {
	return s.serviceTypeRepo.Delete(ctx, id)
}

// TimeRule management
func (s *serviceabilityUseCase) CreateTimeRule(ctx context.Context, timeRule *domain.TimeRule) error {
	return s.timeRuleRepo.Create(ctx, timeRule)
}

func (s *serviceabilityUseCase) GetTimeRuleByID(ctx context.Context, id uuid.UUID) (*domain.TimeRule, error) {
	return s.timeRuleRepo.GetByID(ctx, id)
}

func (s *serviceabilityUseCase) UpdateTimeRule(ctx context.Context, timeRule *domain.TimeRule) error {
	return s.timeRuleRepo.Update(ctx, timeRule)
}

func (s *serviceabilityUseCase) DeleteTimeRule(ctx context.Context, id uuid.UUID) error {
	return s.timeRuleRepo.Delete(ctx, id)
}

func (s *serviceabilityUseCase) GetTimeRulesByLocationID(ctx context.Context, locationID uuid.UUID) ([]*domain.TimeRule, error) {
	return s.timeRuleRepo.GetByLocationID(ctx, locationID)
}
