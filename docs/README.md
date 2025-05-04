# `/api` Directory

This directory contains API definitions for the serviceability service.

## Purpose

* Store OpenAPI/Swagger specifications
* Define API schemas
* Document API endpoints
* Serve as a single source of truth for API contracts

## Structure

* `/openapi` - OpenAPI specification files
* `/proto` - Protocol buffer definitions (if using gRPC)
* `/schemas` - JSON schema definitions

## Guidelines

* Keep API definitions up-to-date with implementation
* Document all endpoints, parameters, and responses
* Include examples for each endpoint
* Version APIs appropriately
* Generate client libraries from these definitions
* Validate API changes against these specifications

## Usage

The OpenAPI specification can be viewed using Swagger UI:

```
# Run Swagger UI (if you have Docker installed)
docker run -p 8080:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd)/api:/api swaggerapi/swagger-ui
```

Then navigate to http://localhost:8080 in your browser. 