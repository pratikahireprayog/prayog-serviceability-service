package gorm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/prayog/serviceability/internal/domain"
	"github.com/prayog/serviceability/pkg/database"
	"gorm.io/gorm"
)

// ServiceAvailabilityRepository implements domain.ServiceAvailabilityRepository using GORM
type ServiceAvailabilityRepository struct {
	*Repository
}

// NewServiceAvailabilityRepository creates a new service availability repository
func NewServiceAvailabilityRepository(db *database.DB) domain.ServiceAvailabilityRepository {
	return &ServiceAvailabilityRepository{
		Repository: NewRepository(db),
	}
}

// GetByID retrieves a service availability by its ID
func (r *ServiceAvailabilityRepository) GetByID(ctx context.Context, id uint) (*domain.ServiceAvailability, error) {
	var serviceAvailability database.ServiceAvailability
	if err := r.WithContext(ctx).First(&serviceAvailability, id).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabaseServiceAvailabilityToDomain(&serviceAvailability), nil
}

// GetByLocation retrieves service availabilities by location
func (r *ServiceAvailabilityRepository) GetByLocation(ctx context.Context, locationType string, locationID uint) ([]*domain.ServiceAvailability, error) {
	var serviceAvailabilities []database.ServiceAvailability
	if err := r.WithContext(ctx).
		Where("location_type = ? AND location_id = ?", locationType, locationID).
		Find(&serviceAvailabilities).Error; err != nil {
		return nil, HandleError(err)
	}

	domainServiceAvailabilities := make([]*domain.ServiceAvailability, len(serviceAvailabilities))
	for i, serviceAvailability := range serviceAvailabilities {
		domainServiceAvailabilities[i] = mapDatabaseServiceAvailabilityToDomain(&serviceAvailability)
	}
	return domainServiceAvailabilities, nil
}

// GetByService retrieves service availabilities by service type ID
func (r *ServiceAvailabilityRepository) GetByService(ctx context.Context, serviceTypeID uint) ([]*domain.ServiceAvailability, error) {
	var serviceAvailabilities []database.ServiceAvailability
	if err := r.WithContext(ctx).
		Where("service_type_id = ?", serviceTypeID).
		Find(&serviceAvailabilities).Error; err != nil {
		return nil, HandleError(err)
	}

	domainServiceAvailabilities := make([]*domain.ServiceAvailability, len(serviceAvailabilities))
	for i, serviceAvailability := range serviceAvailabilities {
		domainServiceAvailabilities[i] = mapDatabaseServiceAvailabilityToDomain(&serviceAvailability)
	}
	return domainServiceAvailabilities, nil
}

// GetByOrderType retrieves service availabilities by order type ID
func (r *ServiceAvailabilityRepository) GetByOrderType(ctx context.Context, orderTypeID uint) ([]*domain.ServiceAvailability, error) {
	var serviceAvailabilities []database.ServiceAvailability
	if err := r.WithContext(ctx).
		Where("order_type_id = ?", orderTypeID).
		Find(&serviceAvailabilities).Error; err != nil {
		return nil, HandleError(err)
	}

	domainServiceAvailabilities := make([]*domain.ServiceAvailability, len(serviceAvailabilities))
	for i, serviceAvailability := range serviceAvailabilities {
		domainServiceAvailabilities[i] = mapDatabaseServiceAvailabilityToDomain(&serviceAvailability)
	}
	return domainServiceAvailabilities, nil
}

// CheckAvailability checks if a service is available for a postal code and order type
func (r *ServiceAvailabilityRepository) CheckAvailability(ctx context.Context, postalCode string, orderTypeCode string, serviceTypeCode string) (bool, error) {
	// Get postal code entity
	var dbPostalCode database.PostalCode
	if err := r.WithContext(ctx).
		Where("code = ?", postalCode).
		First(&dbPostalCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // Not available if postal code doesn't exist
		}
		return false, HandleError(err)
	}

	// Get area, city, region, and country IDs from the postal code
	var dbArea database.Area
	if err := r.WithContext(ctx).First(&dbArea, dbPostalCode.AreaID).Error; err != nil {
		return false, HandleError(err)
	}

	var dbCity database.City
	if err := r.WithContext(ctx).First(&dbCity, dbArea.CityID).Error; err != nil {
		return false, HandleError(err)
	}

	var dbRegion database.AdministrativeRegion
	if err := r.WithContext(ctx).First(&dbRegion, dbCity.AdministrativeRegionID).Error; err != nil {
		return false, HandleError(err)
	}

	// Get order type ID
	var dbOrderType database.OrderType
	if err := r.WithContext(ctx).Where("code = ?", orderTypeCode).First(&dbOrderType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("order type %s not found", orderTypeCode)
		}
		return false, HandleError(err)
	}

	// Check if a specific service type is requested
	var serviceTypeID *uint
	if serviceTypeCode != "" {
		var dbServiceType database.ServiceType
		if err := r.WithContext(ctx).Where("code = ?", serviceTypeCode).First(&dbServiceType).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, fmt.Errorf("service type %s not found", serviceTypeCode)
			}
			return false, HandleError(err)
		}
		serviceTypeID = &dbServiceType.ID
	}

	// The current time
	now := time.Now()

	// Check availability starting from the most specific location (postal code) to the most general (country)
	// Define location types and IDs to check
	locationsToCheck := []struct {
		locationType string
		locationID   uint
	}{
		{"postal_code", dbPostalCode.ID},
		{"area", dbArea.ID},
		{"city", dbCity.ID},
		{"region", dbRegion.ID},
		{"country", dbRegion.CountryID},
	}

	// Use query builder to check availability
	query := r.WithContext(ctx).
		Table("service_availabilities").
		Where("order_type_id = ?", dbOrderType.ID).
		Where("is_available = ?", true).
		Where("effective_from <= ?", now).
		Where("effective_to >= ?", now)

	// Add service type condition if specified
	if serviceTypeID != nil {
		query = query.Where("service_type_id = ?", *serviceTypeID)
	}

	// Add location conditions
	locationCondition := "("
	for i, loc := range locationsToCheck {
		if i > 0 {
			locationCondition += " OR "
		}
		locationCondition += fmt.Sprintf("(location_type = '%s' AND location_id = %d)", loc.locationType, loc.locationID)
	}
	locationCondition += ")"
	query = query.Where(locationCondition)

	// Execute the query
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, HandleError(err)
	}

	return count > 0, nil
}

// BulkCheckAvailability checks availability for multiple service requests concurrently
func (r *ServiceAvailabilityRepository) BulkCheckAvailability(ctx context.Context, requests []domain.ServiceabilityRequest) ([]domain.ServiceabilityResult, error) {
	results := make([]domain.ServiceabilityResult, len(requests))
	errChan := make(chan error, len(requests))
	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)
		go func(idx int, request domain.ServiceabilityRequest) {
			defer wg.Done()

			// Initialize the result with the request data
			results[idx] = domain.ServiceabilityResult{
				PostalCode:    request.PostalCode,
				CountryCode:   request.CountryCode,
				OrderTypeCode: request.OrderTypeCode,
				Services:      []domain.ServiceAvailable{},
				IsServiceable: false,
			}

			// If a specific service type is requested, check just that one
			if request.ServiceTypeCode != "" {
				available, err := r.CheckAvailability(ctx, request.PostalCode, request.OrderTypeCode, request.ServiceTypeCode)
				if err != nil {
					errChan <- err
					return
				}

				// Get service type name
				var serviceType database.ServiceType
				if err := r.WithContext(ctx).Where("code = ?", request.ServiceTypeCode).First(&serviceType).Error; err != nil {
					errChan <- HandleError(err)
					return
				}

				results[idx].Services = append(results[idx].Services, domain.ServiceAvailable{
					ServiceTypeCode: request.ServiceTypeCode,
					ServiceTypeName: serviceType.Name,
					IsAvailable:     available,
				})

				results[idx].IsServiceable = available
			} else {
				// If no specific service is requested, check all services
				var serviceTypes []database.ServiceType
				if err := r.WithContext(ctx).Find(&serviceTypes).Error; err != nil {
					errChan <- HandleError(err)
					return
				}

				hasAvailableService := false
				for _, serviceType := range serviceTypes {
					available, err := r.CheckAvailability(ctx, request.PostalCode, request.OrderTypeCode, serviceType.Code)
					if err != nil {
						errChan <- err
						return
					}

					results[idx].Services = append(results[idx].Services, domain.ServiceAvailable{
						ServiceTypeCode: serviceType.Code,
						ServiceTypeName: serviceType.Name,
						IsAvailable:     available,
					})

					if available {
						hasAvailableService = true
					}
				}

				results[idx].IsServiceable = hasAvailableService
			}
		}(i, req)
	}

	wg.Wait()
	close(errChan)

	// Check if any errors occurred
	if len(errChan) > 0 {
		return nil, <-errChan
	}

	return results, nil
}

// List retrieves all service availabilities
func (r *ServiceAvailabilityRepository) List(ctx context.Context) ([]*domain.ServiceAvailability, error) {
	var serviceAvailabilities []database.ServiceAvailability
	if err := r.WithContext(ctx).Find(&serviceAvailabilities).Error; err != nil {
		return nil, HandleError(err)
	}

	domainServiceAvailabilities := make([]*domain.ServiceAvailability, len(serviceAvailabilities))
	for i, serviceAvailability := range serviceAvailabilities {
		domainServiceAvailabilities[i] = mapDatabaseServiceAvailabilityToDomain(&serviceAvailability)
	}
	return domainServiceAvailabilities, nil
}

// Create creates a new service availability
func (r *ServiceAvailabilityRepository) Create(ctx context.Context, serviceAvailability *domain.ServiceAvailability) error {
	dbServiceAvailability := mapDomainServiceAvailabilityToDatabase(serviceAvailability)
	if err := r.WithContext(ctx).Create(&dbServiceAvailability).Error; err != nil {
		return HandleError(err)
	}
	// Update the ID after creation
	serviceAvailability.ID = dbServiceAvailability.ID
	return nil
}

// Update updates an existing service availability
func (r *ServiceAvailabilityRepository) Update(ctx context.Context, serviceAvailability *domain.ServiceAvailability) error {
	dbServiceAvailability := mapDomainServiceAvailabilityToDatabase(serviceAvailability)
	result := r.WithContext(ctx).Save(&dbServiceAvailability)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a service availability by its ID
func (r *ServiceAvailabilityRepository) Delete(ctx context.Context, id uint) error {
	result := r.WithContext(ctx).Delete(&database.ServiceAvailability{}, id)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// mapDatabaseServiceAvailabilityToDomain maps a database service availability to a domain service availability
func mapDatabaseServiceAvailabilityToDomain(dbServiceAvailability *database.ServiceAvailability) *domain.ServiceAvailability {
	return &domain.ServiceAvailability{
		ID:             dbServiceAvailability.ID,
		LocationType:   dbServiceAvailability.LocationType,
		LocationID:     dbServiceAvailability.LocationID,
		OrderTypeID:    dbServiceAvailability.OrderTypeID,
		ServiceTypeID:  dbServiceAvailability.ServiceTypeID,
		IsAvailable:    dbServiceAvailability.IsAvailable,
		EffectiveFrom:  dbServiceAvailability.EffectiveFrom,
		EffectiveTo:    dbServiceAvailability.EffectiveTo,
		AdditionalData: dbServiceAvailability.AdditionalData,
		CreatedAt:      dbServiceAvailability.CreatedAt,
		UpdatedAt:      dbServiceAvailability.UpdatedAt,
	}
}

// mapDomainServiceAvailabilityToDatabase maps a domain service availability to a database service availability
func mapDomainServiceAvailabilityToDatabase(domainServiceAvailability *domain.ServiceAvailability) database.ServiceAvailability {
	return database.ServiceAvailability{
		ID:             domainServiceAvailability.ID,
		LocationType:   domainServiceAvailability.LocationType,
		LocationID:     domainServiceAvailability.LocationID,
		OrderTypeID:    domainServiceAvailability.OrderTypeID,
		ServiceTypeID:  domainServiceAvailability.ServiceTypeID,
		IsAvailable:    domainServiceAvailability.IsAvailable,
		EffectiveFrom:  domainServiceAvailability.EffectiveFrom,
		EffectiveTo:    domainServiceAvailability.EffectiveTo,
		AdditionalData: domainServiceAvailability.AdditionalData,
		CreatedAt:      domainServiceAvailability.CreatedAt,
		UpdatedAt:      domainServiceAvailability.UpdatedAt,
	}
}
