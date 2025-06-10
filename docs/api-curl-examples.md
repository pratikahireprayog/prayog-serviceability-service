# API Curl Examples

This document provides comprehensive curl examples for all Prayog Serviceability API endpoints. All examples include proper authentication headers and practical request/response scenarios.

## Table of Contents

- [Authentication](#authentication)
- [Health Check](#health-check)
- [Countries](#countries)
- [Region Types](#region-types)
- [Regions](#regions)
- [Districts](#districts)
- [Cities](#cities)
- [Areas](#areas)
- [Location Aliases (Nested)](#location-aliases-nested)
- [Location Aliases (Standalone)](#location-aliases-standalone)
- [Partner Location Coverage](#partner-location-coverage)
- [Serviceability](#serviceability)
- [Common Query Parameters](#common-query-parameters)

## Authentication

All API requests require a Bearer token in the Authorization header:

```bash
# Set your token as an environment variable for convenience
export API_TOKEN="your_jwt_token_here"
export BASE_URL="https://api.prayog.com"

# Alternative for local development
export BASE_URL="http://localhost:8080"
```

## Health Check

### Check API Health

```bash
# Health check (no authentication required)
curl -X GET \
  "${BASE_URL}/serviceability/ping" \
  -H "Content-Type: application/json"
```

**Response:**

```json
{
  "status": "success",
  "data": {
    "status": "up",
    "version": "1.0.0",
    "timestamp": "2024-01-15T10:30:00Z",
    "dependencies": {
      "database": "up",
      "cache": "up"
    }
  }
}
```

## Countries

### List All Countries

```bash
# Get all countries with pagination
curl -X GET \
  "${BASE_URL}/serviceability/v1/countries?limit=10&offset=0" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Create Country

```bash
# Create a new country
curl -X POST \
  "${BASE_URL}/serviceability/v1/countries" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "United States",
    "code": "US",
    "is_active": true
  }'
```

### Get Country by ID

```bash
# Get specific country by UUID
curl -X GET \
  "${BASE_URL}/serviceability/v1/countries/123e4567-e89b-12d3-a456-426614174000" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Get Country by Code

```bash
# Get country by country code
curl -X GET \
  "${BASE_URL}/serviceability/v1/countries/code/US" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Update Country

```bash
# Update existing country
curl -X PUT \
  "${BASE_URL}/serviceability/v1/countries/123e4567-e89b-12d3-a456-426614174000" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "United States of America",
    "code": "US",
    "is_active": true
  }'
```

### Delete Country

```bash
# Soft delete country
curl -X DELETE \
  "${BASE_URL}/serviceability/v1/countries/123e4567-e89b-12d3-a456-426614174000" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

## Region Types

### List All Region Types

```bash
# Get all region types
curl -X GET \
  "${BASE_URL}/serviceability/v1/region-types?limit=20&offset=0" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Create Region Type

```bash
# Create a new region type
curl -X POST \
  "${BASE_URL}/serviceability/v1/region-types" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "STATE",
    "name": "State",
    "description": "Administrative state or province",
    "is_active": true
  }'
```

### Get Region Type by Code

```bash
# Get region type by code
curl -X GET \
  "${BASE_URL}/serviceability/v1/region-types/STATE" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Update Region Type

```bash
# Update existing region type
curl -X PUT \
  "${BASE_URL}/serviceability/v1/region-types/STATE" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "STATE",
    "name": "State/Province",
    "description": "Administrative state, province, or territory",
    "is_active": true
  }'
```

### Delete Region Type

```bash
# Soft delete region type
curl -X DELETE \
  "${BASE_URL}/serviceability/v1/region-types/STATE" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

## Regions

### List All Regions

```bash
# Get all regions
curl -X GET \
  "${BASE_URL}/serviceability/v1/regions?limit=20&offset=0" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Create Region

```bash
# Create a new region
curl -X POST \
  "${BASE_URL}/serviceability/v1/regions" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "California",
    "code": "CA",
    "country_id": "123e4567-e89b-12d3-a456-426614174000",
    "region_type_code": "STATE",
    "is_active": true
  }'
```

### Get Region by ID

```bash
# Get specific region by UUID
curl -X GET \
  "${BASE_URL}/serviceability/v1/regions/456e7890-e89b-12d3-a456-426614174001" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Get Region by Code

```bash
# Get region by code
curl -X GET \
  "${BASE_URL}/serviceability/v1/regions/code/CA" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Get Regions by Country

```bash
# Get all regions for a specific country
curl -X GET \
  "${BASE_URL}/serviceability/v1/regions/country/123e4567-e89b-12d3-a456-426614174000" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Update Region

```bash
# Update existing region
curl -X PUT \
  "${BASE_URL}/serviceability/v1/regions/456e7890-e89b-12d3-a456-426614174001" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "California State",
    "code": "CA",
    "country_id": "123e4567-e89b-12d3-a456-426614174000",
    "region_type_code": "STATE",
    "is_active": true
  }'
```

### Delete Region

```bash
# Soft delete region
curl -X DELETE \
  "${BASE_URL}/serviceability/v1/regions/456e7890-e89b-12d3-a456-426614174001" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

## Districts

### List All Districts

```bash
# Get all districts
curl -X GET \
  "${BASE_URL}/serviceability/v1/districts?limit=20&offset=0" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Create District

```bash
# Create a new district
curl -X POST \
  "${BASE_URL}/serviceability/v1/districts" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Los Angeles County",
    "code": "LA",
    "region_id": "456e7890-e89b-12d3-a456-426614174001",
    "is_active": true
  }'
```

### Get District by ID

```bash
# Get specific district by UUID
curl -X GET \
  "${BASE_URL}/serviceability/v1/districts/789e0123-e89b-12d3-a456-426614174002" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Get District by Code

```bash
# Get district by code
curl -X GET \
  "${BASE_URL}/serviceability/v1/districts/code/LA" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Get Districts by Region

```bash
# Get all districts for a specific region
curl -X GET \
  "${BASE_URL}/serviceability/v1/districts/region/456e7890-e89b-12d3-a456-426614174001" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Update District

```bash
# Update existing district
curl -X PUT \
  "${BASE_URL}/serviceability/v1/districts/789e0123-e89b-12d3-a456-426614174002" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Los Angeles County District",
    "code": "LA",
    "region_id": "456e7890-e89b-12d3-a456-426614174001",
    "is_active": true
  }'
```

### Delete District

```bash
# Soft delete district
curl -X DELETE \
  "${BASE_URL}/serviceability/v1/districts/789e0123-e89b-12d3-a456-426614174002" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

## Cities

### List All Cities

```bash
# Get all cities
curl -X GET \
  "${BASE_URL}/serviceability/v1/cities?limit=20&offset=0" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Create City

```bash
# Create a new city
curl -X POST \
  "${BASE_URL}/serviceability/v1/cities" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Los Angeles",
    "code": "LA_CITY",
    "region_id": "456e7890-e89b-12d3-a456-426614174001",
    "is_active": true
  }'
```

### Get City by ID

```bash
# Get specific city by UUID
curl -X GET \
  "${BASE_URL}/serviceability/v1/cities/012e3456-e89b-12d3-a456-426614174003" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Get City by Code

```bash
# Get city by code
curl -X GET \
  "${BASE_URL}/serviceability/v1/cities/code/LA_CITY" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Get Cities by Region

```bash
# Get all cities for a specific region
curl -X GET \
  "${BASE_URL}/serviceability/v1/cities/region/456e7890-e89b-12d3-a456-426614174001" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Update City

```bash
# Update existing city
curl -X PUT \
  "${BASE_URL}/serviceability/v1/cities/012e3456-e89b-12d3-a456-426614174003" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Los Angeles City",
    "code": "LA_CITY",
    "region_id": "456e7890-e89b-12d3-a456-426614174001",
    "is_active": true
  }'
```

### Delete City

```bash
# Soft delete city
curl -X DELETE \
  "${BASE_URL}/serviceability/v1/cities/012e3456-e89b-12d3-a456-426614174003" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

## Areas

### List All Areas

```bash
# Get all areas
curl -X GET \
  "${BASE_URL}/serviceability/v1/areas?limit=20&offset=0" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Create Area

```bash
# Create a new area
curl -X POST \
  "${BASE_URL}/serviceability/v1/areas" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Hollywood",
    "code": "HOLLYWOOD",
    "city_id": "012e3456-e89b-12d3-a456-426614174003",
    "is_active": true
  }'
```

### Get Area by ID

```bash
# Get specific area by UUID
curl -X GET \
  "${BASE_URL}/serviceability/v1/areas/345e6789-e89b-12d3-a456-426614174004" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Get Area by Code

```bash
# Get area by code
curl -X GET \
  "${BASE_URL}/serviceability/v1/areas/code/HOLLYWOOD" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Get Areas by City

```bash
# Get all areas for a specific city
curl -X GET \
  "${BASE_URL}/serviceability/v1/areas/city/012e3456-e89b-12d3-a456-426614174003" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Update Area

```bash
# Update existing area
curl -X PUT \
  "${BASE_URL}/serviceability/v1/areas/345e6789-e89b-12d3-a456-426614174004" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Hollywood District",
    "code": "HOLLYWOOD",
    "city_id": "012e3456-e89b-12d3-a456-426614174003",
    "is_active": true
  }'
```

### Delete Area

```bash
# Soft delete area
curl -X DELETE \
  "${BASE_URL}/serviceability/v1/areas/345e6789-e89b-12d3-a456-426614174004" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

## Location Aliases (Nested)

### Create Country Alias

```bash
# Create alias for a country
curl -X POST \
  "${BASE_URL}/serviceability/v1/countries/123e4567-e89b-12d3-a456-426614174000/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type_code": "COUNTRY",
    "entity_type": "Country",
    "alias_name": "USA",
    "is_primary": true,
    "is_active": true
  }'
```

### Get Country Aliases

```bash
# Get all aliases for a country
curl -X GET \
  "${BASE_URL}/serviceability/v1/countries/123e4567-e89b-12d3-a456-426614174000/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Create Region Alias

```bash
# Create alias for a region
curl -X POST \
  "${BASE_URL}/serviceability/v1/regions/456e7890-e89b-12d3-a456-426614174001/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type_code": "REGION",
    "entity_type": "Region",
    "alias_name": "Cali",
    "is_primary": false,
    "is_active": true
  }'
```

### Get Region Aliases

```bash
# Get all aliases for a region
curl -X GET \
  "${BASE_URL}/serviceability/v1/regions/456e7890-e89b-12d3-a456-426614174001/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Create District Alias

```bash
# Create alias for a district
curl -X POST \
  "${BASE_URL}/serviceability/v1/districts/789e0123-e89b-12d3-a456-426614174002/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type_code": "DISTRICT",
    "entity_type": "District",
    "alias_name": "LA County",
    "is_primary": false,
    "is_active": true
  }'
```

### Get District Aliases

```bash
# Get all aliases for a district
curl -X GET \
  "${BASE_URL}/serviceability/v1/districts/789e0123-e89b-12d3-a456-426614174002/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Create City Alias

```bash
# Create alias for a city
curl -X POST \
  "${BASE_URL}/serviceability/v1/cities/012e3456-e89b-12d3-a456-426614174003/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type_code": "CITY",
    "entity_type": "City",
    "alias_name": "City of Angels",
    "is_primary": false,
    "is_active": true
  }'
```

### Get City Aliases

```bash
# Get all aliases for a city
curl -X GET \
  "${BASE_URL}/serviceability/v1/cities/012e3456-e89b-12d3-a456-426614174003/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Create Area Alias

```bash
# Create alias for an area
curl -X POST \
  "${BASE_URL}/serviceability/v1/areas/345e6789-e89b-12d3-a456-426614174004/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type_code": "AREA",
    "entity_type": "Area",
    "alias_name": "Tinseltown",
    "is_primary": false,
    "is_active": true
  }'
```

### Get Area Aliases

```bash
# Get all aliases for an area
curl -X GET \
  "${BASE_URL}/serviceability/v1/areas/345e6789-e89b-12d3-a456-426614174004/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Generic Location Aliases

```bash
# Create alias using generic location endpoint
curl -X POST \
  "${BASE_URL}/serviceability/v1/locations/123e4567-e89b-12d3-a456-426614174000/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "entity_type_code": "COUNTRY",
    "entity_type": "Country",
    "alias_name": "United States of America",
    "is_primary": false,
    "is_active": true
  }'

# Get aliases using generic location endpoint
curl -X GET \
  "${BASE_URL}/serviceability/v1/locations/123e4567-e89b-12d3-a456-426614174000/aliases" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

## Location Aliases (Standalone)

### List All Location Aliases

```bash
# Get all location aliases with pagination
curl -X GET \
  "${BASE_URL}/serviceability/v1/location-aliases?limit=20&offset=0" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Get Location Alias by ID

```bash
# Get specific location alias by UUID
curl -X GET \
  "${BASE_URL}/serviceability/v1/location-aliases/678e9012-e89b-12d3-a456-426614174005" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Update Location Alias

```bash
# Update existing location alias
curl -X PUT \
  "${BASE_URL}/serviceability/v1/location-aliases/678e9012-e89b-12d3-a456-426614174005" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "alias_name": "United States (Updated)",
    "is_primary": true,
    "is_active": true
  }'
```

### Delete Location Alias

```bash
# Soft delete location alias
curl -X DELETE \
  "${BASE_URL}/serviceability/v1/location-aliases/678e9012-e89b-12d3-a456-426614174005" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

## Partner Location Coverage

### List Partner Location Coverages

```bash
# Get all location coverages for a partner
curl -X GET \
  "${BASE_URL}/serviceability/v1/partners/550e8400-e29b-41d4-a716-446655440000/location-coverages" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"

# With filtering and pagination
curl -X GET \
  "${BASE_URL}/serviceability/v1/partners/550e8400-e29b-41d4-a716-446655440000/location-coverages?location_scope=CITY&zone_type=PRIMARY&limit=20&offset=0" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

**Response:**

```json
{
  "status": "success",
  "data": {
    "coverages": [
      {
        "id": "789e0123-e89b-12d3-a456-426655440000",
        "partner_id": "550e8400-e29b-41d4-a716-446655440000",
        "partner_code": "PARTNER_001",
        "location_scope": "CITY",
        "location_id": "123e4567-e89b-12d3-a456-426614174000",
        "location_code": "NYC",
        "zone_type": "PRIMARY",
        "is_active": true,
        "created_at": "2023-01-01T00:00:00Z",
        "updated_at": "2023-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "offset": 0,
      "limit": 20,
      "total": 1,
      "has_next": false,
      "has_previous": false
    }
  }
}
```

### Create Partner Location Coverage

```bash
# Create a new location coverage for a partner
curl -X POST \
  "${BASE_URL}/serviceability/v1/partners/550e8400-e29b-41d4-a716-446655440000/location-coverages" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "location_scope": "CITY",
    "location_id": "123e4567-e89b-12d3-a456-426614174000",
    "zone_type": "PRIMARY",
    "is_active": true
  }'

# Using location code instead of location ID
curl -X POST \
  "${BASE_URL}/serviceability/v1/partners/550e8400-e29b-41d4-a716-446655440000/location-coverages" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "location_scope": "CITY",
    "location_code": "NYC",
    "zone_type": "SECONDARY",
    "is_active": true
  }'
```

### Get Specific Partner Location Coverage

```bash
# Get specific location coverage by ID
curl -X GET \
  "${BASE_URL}/serviceability/v1/partners/550e8400-e29b-41d4-a716-446655440000/location-coverages/789e0123-e89b-12d3-a456-426655440000" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Update Partner Location Coverage

```bash
# Update existing location coverage
curl -X PUT \
  "${BASE_URL}/serviceability/v1/partners/550e8400-e29b-41d4-a716-446655440000/location-coverages/789e0123-e89b-12d3-a456-426655440000" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "zone_type": "BUFFER",
    "is_active": false
  }'
```

### Delete Partner Location Coverage

```bash
# Delete location coverage
curl -X DELETE \
  "${BASE_URL}/serviceability/v1/partners/550e8400-e29b-41d4-a716-446655440000/location-coverages/789e0123-e89b-12d3-a456-426655440000" \
  -H "Authorization: Bearer ${API_TOKEN}"
```

### Bulk Create Partner Location Coverages

```bash
# Create multiple location coverages in one request
curl -X POST \
  "${BASE_URL}/serviceability/v1/partners/550e8400-e29b-41d4-a716-446655440000/location-coverages/bulk" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "coverages": [
      {
        "location_scope": "CITY",
        "location_code": "NYC",
        "zone_type": "PRIMARY",
        "is_active": true
      },
      {
        "location_scope": "CITY",
        "location_code": "LA",
        "zone_type": "SECONDARY",
        "is_active": true
      },
      {
        "location_scope": "REGION",
        "location_code": "CA",
        "zone_type": "BUFFER",
        "is_active": true
      }
    ]
  }'
```

**Bulk Response:**

```json
{
  "status": "success",
  "data": {
    "created": [
      {
        "id": "111e1111-e89b-12d3-a456-426655440000",
        "partner_id": "550e8400-e29b-41d4-a716-446655440000",
        "location_scope": "CITY",
        "location_code": "NYC",
        "zone_type": "PRIMARY",
        "is_active": true,
        "created_at": "2023-01-01T00:00:00Z",
        "updated_at": "2023-01-01T00:00:00Z"
      },
      {
        "id": "222e2222-e89b-12d3-a456-426655440000",
        "partner_id": "550e8400-e29b-41d4-a716-446655440000",
        "location_scope": "CITY",
        "location_code": "LA",
        "zone_type": "SECONDARY",
        "is_active": true,
        "created_at": "2023-01-01T00:00:00Z",
        "updated_at": "2023-01-01T00:00:00Z"
      }
    ],
    "failed": [
      {
        "index": 2,
        "error": "Invalid location_code: CA not found for REGION scope",
        "request": {
          "location_scope": "REGION",
          "location_code": "CA",
          "zone_type": "BUFFER",
          "is_active": true
        }
      }
    ]
  }
}
```

## Serviceability

### Check Serviceability by Postal Code

```bash
# Check if a postal code is serviceable
curl -X GET \
  "${BASE_URL}/serviceability/v1/check/110001" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

**Response:**

```json
{
  "status": "success",
  "data": {
    "postal_code": "110001",
    "serviceable": true,
    "coverage_area": {
      "city": "New Delhi",
      "state": "Delhi",
      "country": "IN"
    },
    "services": [
      {
        "service_type": "STANDARD",
        "available": true,
        "estimated_delivery_days": 3
      },
      {
        "service_type": "EXPRESS",
        "available": true,
        "estimated_delivery_days": 1
      }
    ]
  }
}
```

### Bulk Serviceability Check

```bash
# Check multiple postal codes at once
curl -X POST \
  "${BASE_URL}/serviceability/v1/bulk-check" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "postal_codes": ["110001", "400001", "560001", "700001"],
    "service_types": ["STANDARD", "EXPRESS"]
  }'
```

**Response:**

```json
{
  "status": "success",
  "data": {
    "results": [
      {
        "postal_code": "110001",
        "serviceable": true,
        "services": [...]
      },
      {
        "postal_code": "400001",
        "serviceable": true,
        "services": [...]
      },
      {
        "postal_code": "560001",
        "serviceable": false,
        "reason": "Area not covered"
      },
      {
        "postal_code": "700001",
        "serviceable": true,
        "services": [...]
      }
    ],
    "summary": {
      "total_checked": 4,
      "serviceable_count": 3,
      "non_serviceable_count": 1
    }
  }
}
```

## Common Query Parameters

### Pagination Parameters

```bash
# Using pagination parameters
curl -X GET \
  "${BASE_URL}/serviceability/v1/countries?limit=50&offset=100" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Filtering Parameters

```bash
# Filter by active status (if supported)
curl -X GET \
  "${BASE_URL}/serviceability/v1/countries?is_active=true&limit=20" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

### Search Parameters

```bash
# Search by name (if supported)
curl -X GET \
  "${BASE_URL}/serviceability/v1/countries?search=united&limit=10" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

## Error Examples

### Validation Error

```bash
# This will return a validation error (missing required field)
curl -X POST \
  "${BASE_URL}/serviceability/v1/countries" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "US"
  }'
```

**Error Response:**

```json
{
  "status": "error",
  "message": "Field 'name' is required and cannot be empty",
  "code": "validation_error"
}
```

### Not Found Error

```bash
# This will return a not found error
curl -X GET \
  "${BASE_URL}/serviceability/v1/countries/00000000-0000-0000-0000-000000000000" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

**Error Response:**

```json
{
  "status": "error",
  "message": "Country with ID '00000000-0000-0000-0000-000000000000' not found",
  "code": "not_found"
}
```

### Authentication Error

```bash
# This will return an authentication error (missing token)
curl -X GET \
  "${BASE_URL}/serviceability/v1/countries" \
  -H "Content-Type: application/json"
```

**Error Response:**

```json
{
  "status": "error",
  "message": "Authentication token is required",
  "code": "authentication_error"
}
```

## Rate Limiting

When you hit rate limits, you'll receive headers indicating your usage:

```bash
# Check rate limit headers in response
curl -i -X GET \
  "${BASE_URL}/serviceability/v1/countries" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json"
```

**Rate Limit Headers:**

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## Environment Variables for Testing

Create a `.env` file for easier testing:

```bash
# .env file
API_TOKEN=your_jwt_token_here
BASE_URL=https://api.prayog.com

# For local development
# BASE_URL=http://localhost:8080

# Load environment variables
source .env

# Now you can use the examples above directly
```

## Postman Collection

These curl examples can be easily imported into Postman for GUI-based testing. Simply:

1. Copy any curl command
2. Open Postman
3. Click "Import" > "Raw text"
4. Paste the curl command
5. Postman will automatically create the request

---

This comprehensive curl reference covers all implemented API endpoints for the Prayog Serviceability API (Phase 1 - Non-Hub Features). For hub-related endpoints, refer to the Phase 2 documentation when available.
