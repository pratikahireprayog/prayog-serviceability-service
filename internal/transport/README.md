# Transport Layer

The transport layer is responsible for handling communication between the application and external systems, primarily through HTTP RESTful APIs.

## Structure

```
transport/
├── http/
│   ├── handlers/       # Request handlers for different resources
│   │   ├── health.go   # Health check endpoint handler
│   │   └── ...
│   ├── middleware/     # HTTP middleware components
│   ├── router.go       # Routes configuration
│   └── server.go       # HTTP server configuration and lifecycle
└── README.md
```

## HTTP Server

The HTTP server handles incoming requests through a series of components:

1. **Server**: Manages the HTTP server lifecycle (start, shutdown)
2. **Router**: Configures routes and attaches middleware
3. **Handlers**: Process specific resource requests
4. **Middleware**: Apply cross-cutting concerns (logging, auth, etc.)

## Adding New Endpoints

To add a new endpoint:

1. Create a handler in `http/handlers/` following the existing patterns
2. Add a `RegisterRoutes` method that connects route paths to handler methods
3. Update the router to use the new handler

## Available Endpoints

- **Health Check**: `GET /health` - Returns service status and version
- **Serviceability Check**: `GET /api/v1/serviceability/check` - Checks if a location is serviceable

## Middleware

The HTTP server uses the following middleware:

- **Request ID**: Assigns a unique ID to each request
- **Real IP**: Extracts the real client IP from headers
- **Recoverer**: Recovers from panics
- **Logger**: Logs request/response information
- **Response**: Standardizes response formats
- **CORS**: Handles cross-origin resource sharing (optional)
- **Auth**: Validates API keys (optional)
- **Rate Limiting**: Throttles excessive requests (optional)

## Responsibilities

* Handling HTTP requests
* Input validation
* Request parsing and transformation
* Response formatting
* Error handling at the API level
* Authentication and authorization
* Request/response logging

## Guidelines

* Keep this layer thin - no business logic should be here
* Use the usecase layer for all business operations
* Return appropriate HTTP status codes and consistent error responses
* Validate all input before passing to the usecase layer
* Use middleware for cross-cutting concerns
* Document API endpoints with comments for OpenAPI generation

## Potential Future Extensions

* gRPC transport
* WebSocket transport
* GraphQL transport 