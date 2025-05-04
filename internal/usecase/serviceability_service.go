package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prayog/serviceability/internal/domain"
)

// serviceabilityService implements the ServiceabilityService interface
type serviceabilityService struct {
	serviceAvailabilityRepo domain.ServiceAvailabilityRepository
	postalCodeRepo          domain.PostalCodeRepository
	serviceTypeRepo         domain.ServiceTypeRepository
	orderTypeRepo           domain.OrderTypeRepository
	countryRepo             domain.CountryRepository
	locationService         LocationService
	timeRuleService         TimeRuleService
	ctx                     context.Context
}

// NewServiceabilityService creates a new serviceability service
func NewServiceabilityService(
	serviceAvailabilityRepo domain.ServiceAvailabilityRepository,
	postalCodeRepo domain.PostalCodeRepository,
	serviceTypeRepo domain.ServiceTypeRepository,
	orderTypeRepo domain.OrderTypeRepository,
	countryRepo domain.CountryRepository,
) ServiceabilityService {
	// Create the location service with all repositories
	locationService := NewLocationService(
		postalCodeRepo,
		&defaultAreaRepository{repo: serviceAvailabilityRepo},
		&defaultCityRepository{repo: serviceAvailabilityRepo},
		&defaultRegionRepository{repo: serviceAvailabilityRepo},
		countryRepo,
	)

	// Create the time rule service
	timeRuleService := NewTimeRuleService()

	return &serviceabilityService{
		serviceAvailabilityRepo: serviceAvailabilityRepo,
		postalCodeRepo:          postalCodeRepo,
		serviceTypeRepo:         serviceTypeRepo,
		orderTypeRepo:           orderTypeRepo,
		countryRepo:             countryRepo,
		locationService:         locationService,
		timeRuleService:         timeRuleService,
		ctx:                     context.Background(),
	}
}

// Default repository adapters to work with the repository interfaces
type defaultAreaRepository struct {
	repo domain.ServiceAvailabilityRepository
}

func (r *defaultAreaRepository) GetByID(ctx context.Context, id uint) (*domain.Area, error) {
	// For now, we'll need to use a real area repository to implement this
	// In a real implementation, this would be wired to the actual repositories
	return nil, fmt.Errorf("area repository not fully implemented: GetByID")
}

func (r *defaultAreaRepository) GetByCityID(ctx context.Context, cityID uint) ([]*domain.Area, error) {
	return nil, fmt.Errorf("area repository not fully implemented: GetByCityID")
}

func (r *defaultAreaRepository) List(ctx context.Context) ([]*domain.Area, error) {
	return nil, fmt.Errorf("area repository not fully implemented: List")
}

func (r *defaultAreaRepository) Create(ctx context.Context, area *domain.Area) error {
	return fmt.Errorf("area repository not fully implemented: Create")
}

func (r *defaultAreaRepository) Update(ctx context.Context, area *domain.Area) error {
	return fmt.Errorf("area repository not fully implemented: Update")
}

func (r *defaultAreaRepository) Delete(ctx context.Context, id uint) error {
	return fmt.Errorf("area repository not fully implemented: Delete")
}

type defaultCityRepository struct {
	repo domain.ServiceAvailabilityRepository
}

func (r *defaultCityRepository) GetByID(ctx context.Context, id uint) (*domain.City, error) {
	return nil, fmt.Errorf("city repository not fully implemented: GetByID")
}

func (r *defaultCityRepository) GetByRegionID(ctx context.Context, regionID uint) ([]*domain.City, error) {
	return nil, fmt.Errorf("city repository not fully implemented: GetByRegionID")
}

func (r *defaultCityRepository) List(ctx context.Context) ([]*domain.City, error) {
	return nil, fmt.Errorf("city repository not fully implemented: List")
}

func (r *defaultCityRepository) Create(ctx context.Context, city *domain.City) error {
	return fmt.Errorf("city repository not fully implemented: Create")
}

func (r *defaultCityRepository) Update(ctx context.Context, city *domain.City) error {
	return fmt.Errorf("city repository not fully implemented: Update")
}

func (r *defaultCityRepository) Delete(ctx context.Context, id uint) error {
	return fmt.Errorf("city repository not fully implemented: Delete")
}

type defaultRegionRepository struct {
	repo domain.ServiceAvailabilityRepository
}

func (r *defaultRegionRepository) GetByID(ctx context.Context, id uint) (*domain.AdministrativeRegion, error) {
	return nil, fmt.Errorf("region repository not fully implemented: GetByID")
}

func (r *defaultRegionRepository) GetByCode(ctx context.Context, code string) (*domain.AdministrativeRegion, error) {
	return nil, fmt.Errorf("region repository not fully implemented: GetByCode")
}

func (r *defaultRegionRepository) GetByCountryID(ctx context.Context, countryID uint) ([]*domain.AdministrativeRegion, error) {
	return nil, fmt.Errorf("region repository not fully implemented: GetByCountryID")
}

func (r *defaultRegionRepository) List(ctx context.Context) ([]*domain.AdministrativeRegion, error) {
	return nil, fmt.Errorf("region repository not fully implemented: List")
}

func (r *defaultRegionRepository) Create(ctx context.Context, region *domain.AdministrativeRegion) error {
	return fmt.Errorf("region repository not fully implemented: Create")
}

func (r *defaultRegionRepository) Update(ctx context.Context, region *domain.AdministrativeRegion) error {
	return fmt.Errorf("region repository not fully implemented: Update")
}

func (r *defaultRegionRepository) Delete(ctx context.Context, id uint) error {
	return fmt.Errorf("region repository not fully implemented: Delete")
}

// CheckServiceability checks if a service is available at a given location
func (s *serviceabilityService) CheckServiceability(request domain.ServiceabilityRequest) (*domain.ServiceabilityResult, error) {
	// Check if a specific service type is requested
	var specificService bool
	var serviceType *domain.ServiceType
	var err error

	if request.ServiceTypeCode != "" {
		specificService = true
		serviceType, err = s.serviceTypeRepo.GetByCode(s.ctx, request.ServiceTypeCode)
		if err != nil {
			return nil, fmt.Errorf("error retrieving service type: %w", err)
		}
	}

	// Get order type
	orderType, err := s.orderTypeRepo.GetByCode(s.ctx, request.OrderTypeCode)
	if err != nil {
		return nil, fmt.Errorf("error retrieving order type: %w", err)
	}

	// Get the location hierarchy for the postal code
	locationEntries, err := s.locationService.GetLocationEntries(request.PostalCode)
	if err != nil {
		return nil, fmt.Errorf("error resolving location hierarchy: %w", err)
	}

	// Initialize result with basic information
	result := &domain.ServiceabilityResult{
		PostalCode:    request.PostalCode,
		CountryCode:   request.CountryCode,
		OrderTypeCode: request.OrderTypeCode,
		Services:      []domain.ServiceAvailable{},
		IsServiceable: false,
	}

	// Current time for time-based checks
	now := time.Now()

	// If checking a specific service
	if specificService {
		// Check if the service is available using our priority-based approach
		isAvailable, err := s.checkAvailabilityWithPriority(
			orderType.ID,
			serviceType.ID,
			locationEntries.Entries,
			now,
		)
		if err != nil {
			return nil, err
		}

		// Add the service to the result
		result.Services = append(result.Services, domain.ServiceAvailable{
			ServiceTypeCode: serviceType.Code,
			ServiceTypeName: serviceType.Name,
			IsAvailable:     isAvailable,
		})

		result.IsServiceable = isAvailable
	} else {
		// Check all services
		serviceTypes, err := s.serviceTypeRepo.List(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("error retrieving service types: %w", err)
		}

		anyAvailable := false

		// Check each service type
		for _, st := range serviceTypes {
			isAvailable, err := s.checkAvailabilityWithPriority(
				orderType.ID,
				st.ID,
				locationEntries.Entries,
				now,
			)
			if err != nil {
				return nil, err
			}

			// Add the service to the result
			result.Services = append(result.Services, domain.ServiceAvailable{
				ServiceTypeCode: st.Code,
				ServiceTypeName: st.Name,
				IsAvailable:     isAvailable,
			})

			if isAvailable {
				anyAvailable = true
			}
		}

		result.IsServiceable = anyAvailable
	}

	return result, nil
}

// checkAvailabilityWithPriority checks if a service is available according to priority rules
// It prioritizes more specific location rules over general ones, and respects time-based rule activation
func (s *serviceabilityService) checkAvailabilityWithPriority(
	orderTypeID uint,
	serviceTypeID uint,
	locationEntries []LocationEntry,
	currentTime time.Time,
) (bool, error) {
	// Query the database for rules at all location levels
	var availabilities []*domain.ServiceAvailability
	for _, entry := range locationEntries {
		avails, err := s.serviceAvailabilityRepo.GetByLocation(s.ctx, entry.LocationType, entry.LocationID)
		if err != nil {
			return false, fmt.Errorf("error querying availability for location %s (ID: %d): %w",
				entry.LocationType, entry.LocationID, err)
		}

		availabilities = append(availabilities, avails...)
	}

	// Filter for matching order type and service type
	var matchingRules []*domain.ServiceAvailability
	for _, avail := range availabilities {
		if avail.OrderTypeID == orderTypeID && avail.ServiceTypeID == serviceTypeID {
			matchingRules = append(matchingRules, avail)
		}
	}

	// Filter to only active rules using the time rule service
	activeRules := s.timeRuleService.FilterActiveRules(matchingRules, currentTime)

	// If no active rules, the service is not available
	if len(activeRules) == 0 {
		return false, nil
	}

	// Sort the matching rules by specificity (location priority)
	// We'll use a simple map to define priority levels
	priorityMap := map[string]int{
		LocationTypePostalCode: 1, // Most specific
		LocationTypeArea:       2,
		LocationTypeCity:       3,
		LocationTypeRegion:     4,
		LocationTypeCountry:    5, // Least specific
	}

	// Find the most specific rule
	var mostSpecificRule *domain.ServiceAvailability
	highestPriority := 999 // Lower number = higher priority

	for _, rule := range activeRules {
		priority := priorityMap[rule.LocationType]
		if priority < highestPriority {
			highestPriority = priority
			mostSpecificRule = rule
		}
	}

	// Return the availability from the most specific rule
	return mostSpecificRule.IsAvailable, nil
}

// BulkCheckServiceability checks multiple serviceability requests at once
func (s *serviceabilityService) BulkCheckServiceability(requests []domain.ServiceabilityRequest) ([]domain.ServiceabilityResult, error) {
	if len(requests) == 0 {
		return []domain.ServiceabilityResult{}, nil
	}

	// For a large number of requests, process concurrently
	results := make([]domain.ServiceabilityResult, len(requests))
	errChan := make(chan error, len(requests))
	resultChan := make(chan struct {
		index  int
		result *domain.ServiceabilityResult
	}, len(requests))

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // limit concurrency to 10 goroutines

	for i, req := range requests {
		wg.Add(1)
		go func(idx int, request domain.ServiceabilityRequest) {
			defer wg.Done()
			sem <- struct{}{}        // acquire semaphore
			defer func() { <-sem }() // release semaphore

			// Process the individual request
			result, err := s.CheckServiceability(request)
			if err != nil {
				errChan <- fmt.Errorf("error processing request %d: %w", idx, err)
				return
			}

			resultChan <- struct {
				index  int
				result *domain.ServiceabilityResult
			}{idx, result}
		}(i, req)
	}

	// Close channels when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// Collect results
	for res := range resultChan {
		if res.result != nil {
			results[res.index] = *res.result
		}
	}

	// Check if any errors occurred
	select {
	case err := <-errChan:
		if err != nil {
			return nil, err
		}
	default:
		// No errors
	}

	return results, nil
}

// Service availability management

// GetServiceAvailabilityByID retrieves a service availability by its ID
func (s *serviceabilityService) GetServiceAvailabilityByID(id uint) (*domain.ServiceAvailability, error) {
	return s.serviceAvailabilityRepo.GetByID(s.ctx, id)
}

// GetServiceAvailabilitiesByLocation retrieves all service availabilities for a specific location
func (s *serviceabilityService) GetServiceAvailabilitiesByLocation(locationType string, locationID uint) ([]*domain.ServiceAvailability, error) {
	return s.serviceAvailabilityRepo.GetByLocation(s.ctx, locationType, locationID)
}

// GetServiceAvailabilitiesByService retrieves all service availabilities for a specific service
func (s *serviceabilityService) GetServiceAvailabilitiesByService(serviceTypeID uint) ([]*domain.ServiceAvailability, error) {
	return s.serviceAvailabilityRepo.GetByService(s.ctx, serviceTypeID)
}

// GetServiceAvailabilitiesByOrderType retrieves all service availabilities for a specific order type
func (s *serviceabilityService) GetServiceAvailabilitiesByOrderType(orderTypeID uint) ([]*domain.ServiceAvailability, error) {
	return s.serviceAvailabilityRepo.GetByOrderType(s.ctx, orderTypeID)
}

// ListServiceAvailabilities retrieves all service availabilities
func (s *serviceabilityService) ListServiceAvailabilities() ([]*domain.ServiceAvailability, error) {
	return s.serviceAvailabilityRepo.List(s.ctx)
}

// CreateServiceAvailability creates a new service availability
func (s *serviceabilityService) CreateServiceAvailability(serviceAvailability *domain.ServiceAvailability) error {
	return s.serviceAvailabilityRepo.Create(s.ctx, serviceAvailability)
}

// UpdateServiceAvailability updates an existing service availability
func (s *serviceabilityService) UpdateServiceAvailability(serviceAvailability *domain.ServiceAvailability) error {
	return s.serviceAvailabilityRepo.Update(s.ctx, serviceAvailability)
}

// DeleteServiceAvailability deletes a service availability by its ID
func (s *serviceabilityService) DeleteServiceAvailability(id uint) error {
	return s.serviceAvailabilityRepo.Delete(s.ctx, id)
}
