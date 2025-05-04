# Database Package

This package provides database connection management and utilities for the application.

## Features

* Connection pool initialization and management
* Transactions support
* Migration utilities
* Query helpers

## Usage

```go
import "github.com/yourorg/prayog-serviceability-service/pkg/database"

func main() {
    db, err := database.NewConnection(cfg.Database)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Use the database connection
    // ...
}
```

## Design Decisions

* Uses GORM as the ORM layer
* Connection pooling for improved performance
* Supports multiple database types (PostgreSQL, MySQL)
* Provides transaction management utilities

## Testing

The package includes utilities for setting up test databases. See test files for examples. 