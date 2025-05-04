# Domain Layer

The domain layer is the core of the application and contains:

1. **Business Entities**: The core data structures and models that represent the business concepts
2. **Repository Interfaces**: Abstractions for data access operations
3. **Domain Services**: Core business logic that is independent of application use cases
4. **Value Objects**: Immutable objects representing domain concepts without identity
5. **DTOs (Data Transfer Objects)**: Objects used for data exchange between systems

## Structure

```
domain/
├── entity.go          # Core business entities 
├── repository.go      # Repository interfaces for data access
├── dto.go             # Data Transfer Objects for API communication
└── README.md          # This file
```

## Guidelines

- Domain entities should be independent of any infrastructure concerns
- Repository interfaces define contracts for data access operations
- DTOs should be used for external communication and not contain business logic
- All domain entities should follow DDD (Domain-Driven Design) principles
- The domain layer should be the most stable layer in the application

## Key Concepts

### Entities

Core business objects with identity that persist over time.

### Repositories

Abstractions over data storage mechanisms, allowing domain objects to be persisted and retrieved without knowledge of database details.

### DTOs

Objects used to transfer data between processes, particularly useful for API requests and responses.

## Example

```go
// Entity
type Location struct {
    ID          string
    Name        string
    Type        LocationType
    Coordinates GeoCoordinates
    Active      bool
}

// Value object
type GeoCoordinates struct {
    Latitude  float64
    Longitude float64
}

// Repository interface
type LocationRepository interface {
    FindByID(ctx context.Context, id string) (*Location, error)
    FindAll(ctx context.Context, filters LocationFilters) ([]*Location, error)
    Create(ctx context.Context, location *Location) error
    Update(ctx context.Context, location *Location) error
    Delete(ctx context.Context, id string) error
}
``` 