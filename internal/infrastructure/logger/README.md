# Logger Package

This package provides structured logging capabilities for the application.

## Features

* Structured JSON logging
* Log level configuration
* Context-aware logging
* Correlation ID support for request tracing

## Usage

```go
import "github.com/yourorg/prayog-serviceability-service/pkg/logger"

func main() {
    log := logger.NewLogger(cfg.Logger)
    
    // Basic logging
    log.Info("Application started", logger.Field("port", cfg.Server.Port))
    
    // Error logging with structured fields
    if err != nil {
        log.Error("Failed to connect to service", 
            logger.Field("service", "auth"),
            logger.Field("error", err.Error()))
    }
    
    // Context-aware logging
    ctx := logger.WithCorrelationID(context.Background(), "request-123")
    log.InfoContext(ctx, "Processing request")
}
```

## Design Decisions

* Based on zap logger for high performance
* Structured logging for better searchability and analysis
* Support for different output formats (JSON, console)
* Correlation IDs for request tracing across services

## Testing

The package includes a test logger that captures log output for verification in tests. 