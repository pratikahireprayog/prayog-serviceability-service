# `/internal` Directory

The `/internal` directory contains private application code that cannot be imported by external applications. This is the place for code that's specific to this application.

## Structure

The internal code follows a Clean Architecture pattern with the following layers:

### `/domain`

Contains all the core business entities and interfaces that are essential to the business rules. This is the most stable and framework-independent layer.

* Business entities like Country, AdministrativeRegion, City, etc.
* Repository interfaces
* Domain services and interfaces
* Value objects and custom types

### `/repository`

Contains the data access implementations for the repository interfaces defined in the domain layer. This is where we interact with the database or external services.

* GORM implementations of repository interfaces
* Database connection and transaction management
* Data mappers and transformers

### `/usecase`

Contains the application business rules and use cases. This layer orchestrates the flow of data to and from the entities and implements business rules that are specific to the application.

* Application services
* Business use cases
* Data transformations
* Input/output boundary objects

### `/delivery`

Contains the presentation layer components like HTTP handlers. This is where we handle HTTP requests, validate input, and format responses.

* HTTP handlers
* Middleware
* Input validation
* Response formatting

## Dependencies

The dependencies between layers should flow inward, with domain at the center:

```
delivery -> usecase -> domain
             ↑
repository --┘
```

This ensures that the domain layer is independent of frameworks and external concerns. 