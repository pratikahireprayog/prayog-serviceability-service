package services

import (
	"sync"

	"prayog-serviceability-service/pkg/repository"
)

// ServiceFactory provides a factory for creating services
type ServiceFactory struct {
	repos repository.RepositoryProvider
	mu    sync.RWMutex

	// Services
	countryService             *CountryService
	areaService                *AreaService
	postalCodeService          *PostalCodeService
	serviceAvailabilityService *ServiceAvailabilityService
}

// ServiceProvider defines an interface for service providers
type ServiceProvider interface {
	CountryService() *CountryService
	AreaService() *AreaService
	PostalCodeService() *PostalCodeService
	ServiceAvailabilityService() *ServiceAvailabilityService
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(repos repository.RepositoryProvider) *ServiceFactory {
	return &ServiceFactory{
		repos: repos,
	}
}

// CountryService returns the country service
func (f *ServiceFactory) CountryService() *CountryService {
	f.mu.RLock()
	if f.countryService != nil {
		service := f.countryService
		f.mu.RUnlock()
		return service
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.countryService == nil {
		f.countryService = NewCountryService(f.repos.CountryRepository())
	}
	return f.countryService
}

// AreaService returns the area service
func (f *ServiceFactory) AreaService() *AreaService {
	f.mu.RLock()
	if f.areaService != nil {
		service := f.areaService
		f.mu.RUnlock()
		return service
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.areaService == nil {
		f.areaService = NewAreaService(f.repos.AreaRepository())
	}
	return f.areaService
}

// PostalCodeService returns the postal code service
func (f *ServiceFactory) PostalCodeService() *PostalCodeService {
	f.mu.RLock()
	if f.postalCodeService != nil {
		service := f.postalCodeService
		f.mu.RUnlock()
		return service
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.postalCodeService == nil {
		f.postalCodeService = NewPostalCodeService(f.repos.PostalCodeRepository())
	}
	return f.postalCodeService
}

// ServiceAvailabilityService returns the service availability service
func (f *ServiceFactory) ServiceAvailabilityService() *ServiceAvailabilityService {
	f.mu.RLock()
	if f.serviceAvailabilityService != nil {
		service := f.serviceAvailabilityService
		f.mu.RUnlock()
		return service
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.serviceAvailabilityService == nil {
		f.serviceAvailabilityService = NewServiceAvailabilityService(
			f.repos.ServiceAvailabilityRepository(),
			f.repos.OrderTypeRepository(),
			f.repos.ServiceTypeRepository(),
		)
	}
	return f.serviceAvailabilityService
}

// For demonstration purposes, we'll define stub service constructors.
// In a real implementation, these would be properly defined in separate files.

// NewAreaService creates a new area service
func NewAreaService(repo repository.AreaRepository) *AreaService {
	return &AreaService{repo: repo}
}

// AreaService handles area-related business logic
type AreaService struct {
	repo repository.AreaRepository
}

// NewPostalCodeService creates a new postal code service
func NewPostalCodeService(repo repository.PostalCodeRepository) *PostalCodeService {
	return &PostalCodeService{repo: repo}
}

// PostalCodeService handles postal code-related business logic
type PostalCodeService struct {
	repo repository.PostalCodeRepository
}

// NewServiceAvailabilityService creates a new service availability service
func NewServiceAvailabilityService(
	repoSA repository.ServiceAvailabilityRepository,
	repoOT repository.OrderTypeRepository,
	repoST repository.ServiceTypeRepository,
) *ServiceAvailabilityService {
	return &ServiceAvailabilityService{
		repoSA: repoSA,
		repoOT: repoOT,
		repoST: repoST,
	}
}

// ServiceAvailabilityService handles service availability business logic
type ServiceAvailabilityService struct {
	repoSA repository.ServiceAvailabilityRepository
	repoOT repository.OrderTypeRepository
	repoST repository.ServiceTypeRepository
}
