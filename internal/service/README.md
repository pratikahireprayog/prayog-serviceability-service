# Usecase Layer

The usecase layer implements application use cases using the domain entities and repositories. Its main responsibilities are:

1. **Business Logic**: Implementing application-specific business logic
2. **Transaction Coordination**: Coordinating operations across multiple repositories
3. **Input Validation**: Validating input data before processing
4. **Authorization**: Enforcing access control rules
5. **Service Composition**: Composing multiple domain services to fulfill a use case

## Structure

```
usecase/
├── admin_usecase.go           # Admin-related use cases
├── factory.go                 # Factory for creating service instances
├── geo_service.go             # Geography-related service implementation
├── interfaces.go              # Service interfaces definitions
├── location_service.go        # Location service implementation
├── order_type_service.go      # Order type service implementation
├── README.md                  # This file
├── service_type_service.go    # Service type service implementation
├── serviceability_service.go  # Serviceability service implementation
├── serviceability_service.go  # Serviceability use case definitions
├── time_rule_service.go       # Time rule service implementation
└── usecase_factory.go         # Service factory implementation
```

## Guidelines

- Use case implementations should be stateless and depend only on interfaces
- Use dependency injection through constructors
- Handle authorization and validation before executing business logic
- Return domain errors rather than technical errors
- Implement a clear separation between query and command operations
- Document complex use cases with comments
- Use concurrency with caution and ensure proper error handling

## Usage Example

```go
// Create a use case factory
factory := usecase.NewUseCaseFactory(
    geoService, 
    orderTypeService, 
    serviceTypeService,
    serviceabilityService,
)

// Get a specific service
serviceSvc := factory.ServiceabilityService()

// Execute a use case
isServiceable, err := serviceSvc.CheckServiceability("123456", "delivery", "standard")
``` 