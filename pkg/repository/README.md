# Repository Pattern Implementation

This directory contains the implementation of the Repository Pattern for the Prayog Serviceability Service.

## Overview

The repository pattern is an abstraction that separates the domain model from the data access logic. This improves testability, maintainability, and flexibility of the codebase.

## Structure

- `interfaces.go`: Defines repository interfaces using Go's generics for common CRUD operations
- `factory.go`: Provides a factory for creating and accessing repositories
- `gorm/`: Directory containing GORM-specific implementations of the repository interfaces

## Key Components

### Generic Repository Interface

The base `Repository` interface uses Go generics to provide a common contract for CRUD operations:

```go
type Repository[T any] interface {
    GetByID(ctx context.Context, id uuid.UUID) (T, error)
    List(ctx context.Context) ([]T, error)
    Create(ctx context.Context, entity T) error
    Update(ctx context.Context, entity T) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

Entity-specific repository interfaces extend this generic interface with additional methods:

```go
type CountryRepository interface {
    Repository[domain.Country]
    GetByCode(ctx context.Context, code string) (domain.Country, error)
}
```

### BaseRepository Implementation

The `BaseRepository` is implemented in `gorm/base.go` and provides the concrete implementation of common operations:

```go
type BaseRepository[T any, M any] struct {
    db *database.DB
}
```

Where:

- `T` is the domain model type
- `M` is the database model type

The BaseRepository offers:

- Standard CRUD operations (GetByID, List, Create, Update, Delete)
- Query and FindOne methods for custom queries
- Transaction support
- Consistent error handling

### Repository Factory

The `RepositoryFactory` provides a clean way to create and access repository instances:

```go
type RepositoryFactory struct {
    db *database.DB

    // Repositories instances (lazy initialization)
    countryRepo CountryRepository
    regionRepo  RegionRepository
    // Other repositories...
}
```

It implements lazy initialization to create repositories only when they're first requested.

### RepositoryProvider Interface

The `RepositoryProvider` interface defines the contract for repository providers, making it easier to inject repositories into services:

```go
type RepositoryProvider interface {
    CountryRepository() CountryRepository
    RegionRepository() RegionRepository
    // Other repository accessors...
}
```

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

## Best Practices

1. **Use the generic Repository interface** when defining new repository interfaces
2. **Add only essential methods** to entity-specific repositories
3. **Keep mapping logic** in the repository implementation
4. **Use dependency injection** via the RepositoryProvider interface
5. **Handle all database errors** in the repository layer
6. **Prefer composition over inheritance** when implementing repositories
