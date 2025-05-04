# Config Package

This package provides configuration loading and management for the application.

## Features

* Environment variable loading
* Configuration validation
* Type-safe access to configuration values
* Default values for optional settings

## Usage

```go
import "github.com/yourorg/prayog-serviceability-service/pkg/config"

func main() {
    cfg, err := config.NewConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // Access configuration values
    dbURI := cfg.Database.URI
    port := cfg.Server.Port
}
```

## Testing

The package includes test utilities for mocking configuration in tests. See `config_test.go` for examples. 