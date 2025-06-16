# `/api` Directory

This directory contains API definitions for the serviceability service.

## Purpose

- Store OpenAPI/Swagger specifications
- Define API schemas
- Document API endpoints
- Serve as a single source of truth for API contracts

## Structure

- `/openapi` - OpenAPI specification files
- `/proto` - Protocol buffer definitions (if using gRPC)
- `/schemas` - JSON schema definitions

## Guidelines

- Keep API definitions up-to-date with implementation
- Document all endpoints, parameters, and responses
- Include examples for each endpoint
- Version APIs appropriately
- Generate client libraries from these definitions
- Validate API changes against these specifications

## Usage

The OpenAPI specification can be viewed using Swagger UI:

```
# Run Swagger UI (if you have Docker installed)
docker run -p 8080:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd)/api:/api swaggerapi/swagger-ui
```

Then navigate to http://localhost:8080 in your browser.

# Prayog Serviceability API Documentation

Welcome to the comprehensive documentation for the Prayog Serviceability API. This API provides robust geographical data management, location serviceability checking, and partner integration capabilities.

## 🚀 Quick Start

### Authentication

All API endpoints require authentication using Bearer tokens:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     https://api.prayog.com/serviceability/api/v1/countries
```

### Base URLs

- **Production**: `https://api.prayog.com/serviceability`
- **Staging**: `https://api.staging.prayog.com/serviceability`
- **Development**: `http://localhost:8080`

## 📚 API Documentation Structure

### Core Documentation Files

- **[OpenAPI Specification](./openapi/serviceability-api.yaml)** - Complete API specification in OpenAPI 3.0 format
- **[API Reference](./apis-docs.md)** - Detailed endpoint documentation with examples
- **[Database Schema](./db-schema.md)** - Database structure and relationships

### Feature-Specific Documentation

- **[Authentication Guide](#authentication)** - Authentication flows and middleware
- **[Geographical Entities](#geographical-entities)** - Countries, regions, districts, cities, areas
- **[Location Management](#location-management)** - Location types, aliases, and postal codes
- **[Partner Integration](#partner-integration)** - External partner service integration
- **[Error Handling](#error-handling)** - Error codes and handling patterns

## 🏗️ API Architecture

### Core Features

#### 1. Geographical Entity Management

Complete CRUD operations for hierarchical geographical data:

- **Countries** - ISO country management
- **Region Types** - Administrative region classifications
- **Regions** - Administrative regions (states, provinces)
- **Districts** - Administrative subdivisions
- **Cities** - City/municipality management
- **Areas** - Neighborhoods and local areas

#### 2. Location Management System

Flexible location classification and aliasing:

- **Location Types** - Customizable location classifications
- **Location Aliases** - Alternative names and identifiers for locations
- **Postal Codes** - Postal code management with location validation

#### 3. Partner Integration

External service integration capabilities:

- **Partner Validation** - External partner service validation
- **Location Coverage** - Partner service area management
- **Service Specifications** - Partner-specific service configurations

#### 4. Serviceability Engine

Real-time service area coverage checking:

- **Coverage Check** - Postal code serviceability validation
- **Bulk Operations** - Batch serviceability checking
- **Service Options** - Available services per location

## 🌐 API Endpoints Overview

### Geographical Entities

```
# Countries
GET    /serviceability/v1/countries                    # List countries (paginated)
POST   /serviceability/v1/countries                    # Create country
GET    /serviceability/v1/countries/{id}               # Get country by ID
PUT    /serviceability/v1/countries/{id}               # Update country
DELETE /serviceability/v1/countries/{id}               # Delete country
GET    /serviceability/v1/countries/code/{code}        # Get country by code

# Region Types
GET    /serviceability/v1/region-types                 # List region types
POST   /serviceability/v1/region-types                 # Create region type
GET    /serviceability/v1/region-types/{code}          # Get region type
PUT    /serviceability/v1/region-types/{code}          # Update region type
DELETE /serviceability/v1/region-types/{code}          # Delete region type

# Regions
GET    /serviceability/v1/regions                      # List regions
POST   /serviceability/v1/regions                      # Create region
GET    /serviceability/v1/regions/{id}                 # Get region by ID
PUT    /serviceability/v1/regions/{id}                 # Update region
DELETE /serviceability/v1/regions/{id}                 # Delete region
GET    /serviceability/v1/regions/code/{code}          # Get region by code
GET    /serviceability/v1/regions/country/{countryID}  # Get regions by country

# Districts
GET    /serviceability/v1/districts                    # List districts
POST   /serviceability/v1/districts                    # Create district
GET    /serviceability/v1/districts/{id}               # Get district by ID
PUT    /serviceability/v1/districts/{id}               # Update district
DELETE /serviceability/v1/districts/{id}               # Delete district
GET    /serviceability/v1/districts/code/{code}        # Get district by code
GET    /serviceability/v1/districts/region/{regionID}  # Get districts by region

# Cities
GET    /serviceability/v1/cities                       # List cities
POST   /serviceability/v1/cities                       # Create city
GET    /serviceability/v1/cities/{id}                  # Get city by ID
PUT    /serviceability/v1/cities/{id}                  # Update city
DELETE /serviceability/v1/cities/{id}                  # Delete city
GET    /serviceability/v1/cities/code/{code}           # Get city by code
GET    /serviceability/v1/cities/region/{regionID}     # Get cities by region

# Areas
GET    /serviceability/v1/areas                        # List areas
POST   /serviceability/v1/areas                        # Create area
GET    /serviceability/v1/areas/{id}                   # Get area by ID
PUT    /serviceability/v1/areas/{id}                   # Update area
DELETE /serviceability/v1/areas/{id}                   # Delete area
GET    /serviceability/v1/areas/code/{code}            # Get area by code
GET    /serviceability/v1/areas/city/{cityID}          # Get areas by city
```

### Location Management

```
# Location Aliases (Standalone)
GET    /serviceability/v1/location-aliases             # List all aliases
GET    /serviceability/v1/location-aliases/{id}        # Get alias by ID
PUT    /serviceability/v1/location-aliases/{id}        # Update alias
DELETE /serviceability/v1/location-aliases/{id}        # Delete alias

# Location Aliases (Nested)
GET    /serviceability/v1/countries/{id}/aliases       # Get country aliases
POST   /serviceability/v1/countries/{id}/aliases       # Create country alias
GET    /serviceability/v1/regions/{id}/aliases         # Get region aliases
POST   /serviceability/v1/regions/{id}/aliases         # Create region alias
GET    /serviceability/v1/districts/{id}/aliases       # Get district aliases
POST   /serviceability/v1/districts/{id}/aliases       # Create district alias
GET    /serviceability/v1/cities/{id}/aliases          # Get city aliases
POST   /serviceability/v1/cities/{id}/aliases          # Create city alias
GET    /serviceability/v1/areas/{id}/aliases           # Get area aliases
POST   /serviceability/v1/areas/{id}/aliases           # Create area alias

# Generic Location Aliases
GET    /serviceability/v1/locations/{id}/aliases       # Get location aliases
POST   /serviceability/v1/locations/{id}/aliases       # Create location alias
```

### Partner Location Coverage

```
# Partner coverage management
GET    /serviceability/v1/partners/{partner_id}/location-coverages          # List partner coverages
POST   /serviceability/v1/partners/{partner_id}/location-coverages          # Create coverage
GET    /serviceability/v1/partners/{partner_id}/location-coverages/{id}     # Get specific coverage
PUT    /serviceability/v1/partners/{partner_id}/location-coverages/{id}     # Update coverage
DELETE /serviceability/v1/partners/{partner_id}/location-coverages/{id}     # Delete coverage
POST   /serviceability/v1/partners/{partner_id}/location-coverages/bulk     # Bulk create coverages
```

### Serviceability

```
GET    /serviceability/v1/check/{postalCode}  # Check serviceability
POST   /serviceability/v1/bulk-check          # Bulk serviceability check
```

## 📝 Response Format

### Success Response

All successful API responses follow this structure:

```json
{
  "status": "success",
  "data": {
    // Response data here
  }
}
```

### Error Response

Error responses follow this structure:

```json
{
  "status": "error",
  "message": "Human-readable error message",
  "code": "machine_readable_error_code"
}
```

### Paginated Response

List endpoints return paginated responses:

```json
{
  "status": "success",
  "data": {
    "countries": [...],
    "pagination": {
      "offset": 0,
      "limit": 10,
      "total": 245
    }
  }
}
```

## 🔐 Authentication

### Bearer Token Authentication

Include your authentication token in the Authorization header:

```bash
Authorization: Bearer YOUR_JWT_TOKEN
```

### Authentication Flow

1. Obtain JWT token from authentication service
2. Include token in Authorization header for all requests
3. Token validation is performed by authentication middleware
4. Expired tokens return 401 Unauthorized

## ⚠️ Error Handling

### HTTP Status Codes

- **200 OK** - Successful GET/PUT operations
- **201 Created** - Successful POST operations
- **204 No Content** - Successful DELETE operations
- **400 Bad Request** - Invalid request data
- **401 Unauthorized** - Authentication required
- **404 Not Found** - Resource not found
- **409 Conflict** - Resource already exists or constraint violation
- **500 Internal Server Error** - Server error

### Error Codes

- `validation_error` - Request validation failed
- `not_found` - Requested resource not found
- `conflict_error` - Resource constraint violation
- `authentication_error` - Authentication failed
- `authorization_error` - Insufficient permissions
- `internal_error` - Internal server error

## 🧪 Testing & Examples

### Basic Country Operations

```bash
# Get all countries
curl -H "Authorization: Bearer TOKEN" \
     "https://api.prayog.com/serviceability/api/v1/countries?limit=5&offset=0"

# Create a country
curl -X POST \
     -H "Authorization: Bearer TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name": "United States", "code": "US", "is_active": true}' \
     "https://api.prayog.com/serviceability/api/v1/countries"

# Get country by ID
curl -H "Authorization: Bearer TOKEN" \
     "https://api.prayog.com/serviceability/api/v1/countries/123e4567-e89b-12d3-a456-426614174000"
```

### Location Alias Operations

```bash
# Create alias for a country
curl -X POST \
     -H "Authorization: Bearer TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "entity_type_code": "COUNTRY",
       "entity_type": "Country",
       "alias_name": "USA",
       "is_primary": true
     }' \
     "https://api.prayog.com/serviceability/api/v1/countries/123e4567-e89b-12d3-a456-426614174000/aliases"

# Get all aliases for a country
curl -H "Authorization: Bearer TOKEN" \
     "https://api.prayog.com/serviceability/api/v1/countries/123e4567-e89b-12d3-a456-426614174000/aliases"
```

### Serviceability Check

```bash
# Check serviceability for a postal code
curl -H "Authorization: Bearer TOKEN" \
     "https://api.prayog.com/serviceability/api/v1/serviceability/check/110001"

# Bulk serviceability check
curl -X POST \
     -H "Authorization: Bearer TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"postal_codes": ["110001", "400001", "560001"]}' \
     "https://api.prayog.com/serviceability/api/v1/serviceability/bulk-check"
```

## 🚀 Getting Started Checklist

1. **Obtain API credentials** from your account dashboard
2. **Test authentication** with the health check endpoint
3. **Explore geographical data** starting with countries
4. **Set up location aliases** for your specific use case
5. **Configure partner integrations** if needed
6. **Test serviceability checks** for your target areas

## 📊 Rate Limits

- **Standard requests**: 1000 requests per minute
- **Bulk operations**: 100 requests per minute
- **Authentication attempts**: 10 attempts per minute

Rate limit headers are included in all responses:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## 🔄 API Versioning

The API uses URL-based versioning:

- Current version: **v1**
- Base path: `/api/v1/`
- Backward compatibility is maintained for minor updates
- Breaking changes will result in a new version (v2, v3, etc.)

## 📞 Support

- **Documentation Issues**: [GitHub Issues](https://github.com/prayog/serviceability/issues)
- **API Support**: [api-support@prayog.com](mailto:api-support@prayog.com)
- **Status Page**: [status.prayog.com](https://status.prayog.com)

---

This documentation covers the **Phase 1** implementation focusing on non-hub features. Hub-related documentation will be added in Phase 2 when those features are implemented.
