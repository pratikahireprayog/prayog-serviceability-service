# Usecase Layer

This directory contains the usecase (or application service) layer. This layer implements the business logic of the application by orchestrating domain entities and services.

## Responsibilities

* Implementing business logic and application use cases
* Orchestrating domain entities and services
* Transaction coordination
* Input/output transformation
* Business validation
* Authorization checks

## Structure

* Each file typically represents a logical group of related use cases
* Use cases are organized around business capabilities
* Each use case should be a single self-contained function

## Guidelines

* Use cases should depend on domain interfaces, not concrete implementations
* Keep use case functions focused on a single responsibility
* Use dependency injection to provide required services and repositories
* Handle all business validation and authorization checks
* Use contexts for cancellation, timeout, and value propagation
* Document each use case thoroughly with its business purpose
* Write comprehensive tests with mocked dependencies

## Example

```go
type LocationUseCase struct {
    repo domain.LocationRepository
    logger logger.Logger
}

func NewLocationUseCase(repo domain.LocationRepository, logger logger.Logger) *LocationUseCase {
    return &LocationUseCase{
        repo: repo,
        logger: logger,
    }
}

func (uc *LocationUseCase) GetLocationById(ctx context.Context, id string) (*domain.Location, error) {
    // Authorization check
    if err := auth.CanViewLocation(ctx, id); err != nil {
        return nil, err
    }
    
    // Get location from repository
    location, err := uc.repo.FindByID(ctx, id)
    if err != nil {
        uc.logger.Error("Failed to get location", logger.Field("id", id), logger.Field("error", err))
        return nil, err
    }
    
    return location, nil
}
``` 