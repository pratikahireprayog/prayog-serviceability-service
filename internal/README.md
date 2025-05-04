# Internal Package

This directory contains the private application code that is specific to this service.

## Directory Structure

```
internal/
  ├── api/                  # HTTP API handlers and routing
  │   ├── handlers/         # HTTP Request handlers
  │   ├── middleware/       # HTTP middlewares
  │   ├── router.go         # Router configuration
  │   └── server.go         # HTTP Server setup
  │
  ├── di/                   # Dependency Injection
  │   ├── container.go      # DI container definition
  │   ├── providers.go      # Component providers
  │   └── wire.go           # Wire DI setup (if using Wire)
  │
  ├── domain/               # Domain models and core business rules
  │   ├── models/           # Data structures/entities
  │   └── interfaces/       # Service interfaces (ports)
  │
  ├── infrastructure/       # Infrastructure concerns
  │   ├── db/               # Database connections, migrations
  │   ├── logger/           # Logging utilities
  │   └── errors/           # Error handling
  │
  ├── repository/           # Data access layer
  │   └── interfaces/       # Repository interfaces
  │
  └── service/              # Business logic implementation
      └── [feature]/        # Feature-specific services
```

## Package Descriptions

### api

The `api` package contains all HTTP-related code, including handlers, middleware, the router, and the server.

### di

The `di` package contains the dependency injection container and component providers.

### domain

The `domain` package contains the core domain models and business logic interfaces.

### infrastructure

The `infrastructure` package contains infrastructure concerns such as database connections, logging, and error handling.

### repository

The `repository` package contains the data access layer.

### service

The `service` package contains the business logic implementation.

## Architectural Layers

### Domain Layer

The core layer containing business entities, value objects, repository interfaces, and domain services.

- **Location**: `./domain/`
- **Dependencies**: None (should be independent of all other layers)
- **Purpose**: Defines the core business concepts and rules

See [Domain README](./domain/README.md) for more details.

### Repository Layer

The data access layer that implements domain repository interfaces.

- **Location**: `./repository/`
- **Dependencies**: Domain layer
- **Purpose**: Data persistence and retrieval

See [Repository README](./repository/README.md) for more details.

### Use Case Layer

The application service layer that implements business logic and use cases.

- **Location**: `./usecase/`
- **Dependencies**: Domain layer
- **Purpose**: Orchestration of domain objects and services

See [Use Case README](./usecase/README.md) for more details.

### Transport Layer

The interface layer that handles external communication (HTTP, gRPC, etc.).

- **Location**: `./transport/`
- **Dependencies**: Use Case layer
- **Purpose**: Handling external requests and responses

See [Transport README](./transport/README.md) for more details.

## Dependency Flow

The dependency flow should always point inward:

```
Transport → Use Cases → Domain ← Repositories
```

This ensures that the domain layer remains independent and stable, with adapters in the outer layers managing the details of external interfaces and persistence.

## Design Principles

1. **Separation of Concerns**: Each layer has a specific responsibility
2. **Dependency Inversion**: High-level modules don't depend on low-level modules
3. **Single Responsibility**: Each component has one reason to change
4. **Interface Segregation**: Clients shouldn't depend on methods they don't use
5. **Dependency Injection**: Dependencies are provided externally 

## Refactoring Status

The codebase is currently undergoing architectural refactoring to improve consistency and maintainability. The following changes have been implemented:

1. **Domain Layer**:
   - Consolidated entity definitions in `domain/entity.go`
   - Standardized on `uuid.UUID` for all entity IDs
   - Created `domain/dto.go` for Data Transfer Objects
   - Added clear naming conventions (prefixing DTOs with "API" to avoid naming conflicts)
   - Consolidated repository interfaces in `domain/repository.go`
   - Added ServiceabilityLocationHierarchy and related types to domain package

2. **Interface Declarations**:
   - Consolidated all service interfaces in `usecase/interfaces.go`
   - Removed duplicate interface declarations from individual files
   - Ensured consistent method signatures across implementations
   - Fixed parameter and return types to use UUID consistently

3. **Factory Pattern**:
   - Implemented proper factory pattern with `UseCaseFactory` interface
   - Created `UseCaseFactoryImpl` concrete implementation
   - Ensured factory provides all required services

Outstanding tasks that need to be completed:

1. **Implementation Updates**: 
   - Update the service implementations to match the updated interfaces
   - Change ID types from `uint` to `uuid.UUID` consistently
   - Adapt the service implementations to use the domain package types

2. **Repository Layer**:
   - Remove redundant repository interfaces and align implementations
   - Update repository methods to use domain entities

3. **Testing**:
   - Update and expand test coverage for the refactored components
   - Ensure repository and service mocks match the updated interfaces

4. **Documentation**:
   - Complete README files for each component
   - Add usage examples and diagrams

These refactoring efforts aim to establish a clean, consistent architecture with proper separation of concerns and well-defined dependencies, making the codebase more maintainable and extensible. 