package repository

import (
	database "prayog-serviceability-service/pkg/infrastructure/db"
	"prayog-serviceability-service/pkg/repository/gorm"
	"sync"
)

// RepositoryFactory provides a factory for creating repositories
type RepositoryFactory struct {
	db *database.DB
	mu sync.RWMutex

	// Repositories cache
	repositories map[string]interface{}
}

// RepositoryProvider defines an interface for repository providers
// This makes it easier to inject repositories into services
type RepositoryProvider interface {
	CountryRepository() CountryRepository
	RegionRepository() RegionRepository
	CityRepository() CityRepository
	AreaRepository() AreaRepository
	PostalCodeRepository() PostalCodeRepository
	OrderTypeRepository() OrderTypeRepository
	ServiceTypeRepository() ServiceTypeRepository
	ServiceAvailabilityRepository() ServiceAvailabilityRepository
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(db *database.DB) *RepositoryFactory {
	return &RepositoryFactory{
		db:           db,
		repositories: make(map[string]interface{}),
	}
}

// New creates a new repository factory - convenience function
func New(db *database.DB) RepositoryProvider {
	return NewRepositoryFactory(db)
}

// getOrCreateRepository returns an existing repository or creates a new one using the factory function
func (f *RepositoryFactory) getOrCreateRepository(key string, factory func() interface{}) interface{} {
	f.mu.RLock()
	if repo, exists := f.repositories[key]; exists {
		f.mu.RUnlock()
		return repo
	}
	f.mu.RUnlock()

	// Need to create the repository
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check again in case another goroutine created it while we were waiting for the lock
	if repo, exists := f.repositories[key]; exists {
		return repo
	}

	// Create the repository
	repo := factory()
	f.repositories[key] = repo
	return repo
}

// Repository creation functions
// These functions could be moved to a separate file if the factory gets too large

func createCountryRepository(db *database.DB) CountryRepository {
	return gorm.NewCountryRepository(db)
}

func createRegionRepository(db *database.DB) RegionRepository {
	return gorm.NewRegionRepository(db)
}

func createCityRepository(db *database.DB) CityRepository {
	return gorm.NewCityRepository(db)
}

func createAreaRepository(db *database.DB) AreaRepository {
	return gorm.NewAreaRepository(db)
}

func createPostalCodeRepository(db *database.DB) PostalCodeRepository {
	return gorm.NewPostalCodeRepository(db)
}

func createOrderTypeRepository(db *database.DB) OrderTypeRepository {
	return gorm.NewOrderTypeRepository(db)
}

func createServiceTypeRepository(db *database.DB) ServiceTypeRepository {
	return gorm.NewServiceTypeRepository(db)
}

func createServiceAvailabilityRepository(db *database.DB) ServiceAvailabilityRepository {
	return gorm.NewServiceAvailabilityRepository(db)
}

// CountryRepository returns the country repository
func (f *RepositoryFactory) CountryRepository() CountryRepository {
	repo := f.getOrCreateRepository("country", func() interface{} {
		// Import implementation dynamically based on configuration
		// This helps with testing and flexibility
		return createCountryRepository(f.db)
	})
	return repo.(CountryRepository)
}

// RegionRepository returns the region repository
func (f *RepositoryFactory) RegionRepository() RegionRepository {
	repo := f.getOrCreateRepository("region", func() interface{} {
		return createRegionRepository(f.db)
	})
	return repo.(RegionRepository)
}

// CityRepository returns the city repository
func (f *RepositoryFactory) CityRepository() CityRepository {
	repo := f.getOrCreateRepository("city", func() interface{} {
		return createCityRepository(f.db)
	})
	return repo.(CityRepository)
}

// AreaRepository returns the area repository
func (f *RepositoryFactory) AreaRepository() AreaRepository {
	repo := f.getOrCreateRepository("area", func() interface{} {
		return createAreaRepository(f.db)
	})
	return repo.(AreaRepository)
}

// PostalCodeRepository returns the postal code repository
func (f *RepositoryFactory) PostalCodeRepository() PostalCodeRepository {
	repo := f.getOrCreateRepository("postalCode", func() interface{} {
		return createPostalCodeRepository(f.db)
	})
	return repo.(PostalCodeRepository)
}

// OrderTypeRepository returns the order type repository
func (f *RepositoryFactory) OrderTypeRepository() OrderTypeRepository {
	repo := f.getOrCreateRepository("orderType", func() interface{} {
		return createOrderTypeRepository(f.db)
	})
	return repo.(OrderTypeRepository)
}

// ServiceTypeRepository returns the service type repository
func (f *RepositoryFactory) ServiceTypeRepository() ServiceTypeRepository {
	repo := f.getOrCreateRepository("serviceType", func() interface{} {
		return createServiceTypeRepository(f.db)
	})
	return repo.(ServiceTypeRepository)
}

// ServiceAvailabilityRepository returns the service availability repository
func (f *RepositoryFactory) ServiceAvailabilityRepository() ServiceAvailabilityRepository {
	repo := f.getOrCreateRepository("serviceAvailability", func() interface{} {
		return createServiceAvailabilityRepository(f.db)
	})
	return repo.(ServiceAvailabilityRepository)
}
