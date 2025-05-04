# Domain Layer

This directory contains the domain layer components of the application. The domain layer represents the core business entities and rules, independent of any external frameworks or technologies.

## Components

* Business entities (models)
* Value objects
* Repository interfaces
* Domain services and interfaces
* Custom errors and types

## Guidelines

* This layer should be framework-independent
* No import of external packages except standard library
* No database or external service concerns
* Define repository interfaces here, implement in the repository layer
* Focus on business rules and validation
* Use immutable data structures where possible
* Document domain concepts thoroughly

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