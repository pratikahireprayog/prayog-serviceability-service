# Repository Layer

The repository layer implements the data access interfaces defined in the domain layer. Its main responsibilities are:

1. **Persistence**: Storing and retrieving domain entities
2. **Data Mapping**: Converting between domain entities and database models
3. **Query Execution**: Executing database queries and transactions
4. **Caching**: Optimizing data access through caching mechanisms (optional)

## Structure

```
repository/
├── gorm/                      # GORM implementation of repositories
│   ├── area_repository.go     # Area entity repository implementation
│   ├── city_repository.go     # City entity repository implementation
│   ├── country_repository.go  # Country entity repository implementation
│   ├── factory.go             # Factory for creating repository instances
│   ├── order_type_repository.go     # Order type repository implementation
│   ├── postal_code_repository.go    # Postal code repository implementation
│   ├── region_repository.go         # Region repository implementation
│   ├── repository.go                # Base repository implementation
│   └── service_availability_repository.go  # Service availability repository
└── README.md                  # This file
```

## Guidelines

- Repository implementations should follow the interfaces defined in the domain layer
- Database-specific logic should be contained within the repository implementations
- Repositories should return domain entities, not database models
- Error handling should translate database errors to domain-specific errors
- Connection management and transactions should be handled at this layer
- Use factory patterns to create repository instances

## Usage Example

```go
// Create a repository factory
factory := gorm.NewRepositoryFactory(db)

// Get a specific repository
countryRepo := factory.CountryRepository()

// Use the repository
countries, err := countryRepo.List(ctx)
``` 