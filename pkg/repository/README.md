# Repository Pattern Implementation

This directory contains the implementation of the Repository Pattern for the Prayog Serviceability Service.

## Overview

The repository pattern is an abstraction that separates the domain model from the data access logic. This improves testability, maintainability, and flexibility of the codebase.

## Structure

- `interfaces.go`: Defines the repository interfaces that each implementation must adhere to
- `repository.go`: Contains the generic base repository implementation with common CRUD operations
- `factory.go`: Provides a factory for creating and accessing repositories
- `gorm/`: Directory containing GORM-specific implementations of the repository interfaces

## Key Components

### BaseRepository

The `BaseRepository` provides a generic implementation of common CRUD operations for all entity types:

```go
type BaseRepository[T any, M any] struct {
    db *database.DB
}
```

where:

- `T` is the domain model type
- `M` is the database model type

The BaseRepository offers:

- Standard CRUD operations (GetByID, List, Create, Update, Delete)
- Query and FindOne methods for custom queries
- Transaction support
- Consistent error handling

### RepositoryProvider Interface

The `RepositoryProvider` interface defines the contract for repository providers, making it easier to inject repositories into services:

```go
type RepositoryProvider interface {
    CountryRepository() CountryRepository
    RegionRepository() RegionRepository
    CityRepository() CityRepository
    // ...other repositories
}
```

### Repository Factory

The `RepositoryFactory` provides a thread-safe way to create and access repository instances:

```go
type RepositoryFactory struct {
    db *database.DB
    mu sync.RWMutex
    repositories map[string]interface{}
}
```

It implements the Service Locator pattern and uses lazy initialization with double-checked locking to create repositories only when needed.

### Repository Interfaces

Each domain entity has a corresponding repository interface that defines the operations that can be performed on that entity.

Example:

```go
type CountryRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (domain.Country, error)
    GetByCode(ctx context.Context, code string) (domain.Country, error)
    List(ctx context.Context) ([]domain.Country, error)
    Create(ctx context.Context, country domain.Country) error
    Update(ctx context.Context, country domain.Country) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### Implementation

Each repository interface has a GORM-specific implementation in the `gorm/` directory that extends the `BaseRepository`.

Example:

```go
type CountryRepository struct {
    *repository.BaseRepository[domain.Country, database.Country]
}
```

Repositories can add custom methods for entity-specific operations beyond the standard CRUD operations provided by the BaseRepository.

## Usage Examples

### Creating and Using the Repository Factory

```go
// Create the repository provider
db, _ := database.NewDatabase(config)
repos := repository.New(db)

// Use the provider to access repositories
countryRepo := repos.CountryRepository()
```

### Using a Repository in a Service

```go
type CountryService struct {
    repo repository.CountryRepository
}

func NewCountryService(repo repository.CountryRepository) *CountryService {
    return &CountryService{repo: repo}
}

func (s *CountryService) GetCountryByCode(ctx context.Context, code string) (domain.Country, error) {
    return s.repo.GetByCode(ctx, code)
}
```

### Custom Repository Methods

Repositories can implement custom methods for entity-specific operations:

```go
func (r *AreaRepository) FindNearbyAreas(ctx context.Context, point domain.Point, radiusKm float64) ([]domain.Area, error) {
    return r.BaseRepository.Query(ctx, func(db *gorm.DB) *gorm.DB {
        // Complex query implementation
        return db.Where("is_active = ?", true).Limit(10)
    }, mapDatabaseAreaToDomain)
}
```

## Error Handling

Repositories return domain-specific errors instead of database-specific errors:

- `ErrNotFound`: When the requested entity does not exist
- `ErrDatabase`: When a database error occurs
- `ErrInvalid`: When the entity is invalid
- `ErrConflict`: When there's a conflict (e.g., duplicate key)

## Testing

The interface-based design makes testing easier by allowing mock implementations:

```go
type MockCountryRepository struct {
    // Mock implementation
}

func TestCountryService(t *testing.T) {
    mockRepo := &MockCountryRepository{}
    service := NewCountryService(mockRepo)
    // Test service methods
}
```
