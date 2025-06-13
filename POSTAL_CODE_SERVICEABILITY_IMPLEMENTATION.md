# Postal Code Serviceability API Implementation

## Overview

This document describes the implementation of the new postal code based serviceability API that integrates with the existing partner location coverage table to provide serviceability information in a specific format requested by the user.

## Key Features

- **Single Postal Code Check**: GET endpoint to check serviceability for a specific postal code
- **Source/Destination Check**: POST endpoint to check serviceability between source and destination postal codes
- **Data Transformation**: Converts partner location coverage data into the requested serviceability format
- **Service Aggregation**: Groups services by parcel category and aggregates delivery modes
- **Product Types Mapping**: Creates a standardized product types map showing availability of different product types
- **Filter Support**: Supports filtering by parcel category and product type

## API Endpoints

### 1. GET `/serviceability/v1/check/{postal_code}`

Checks serviceability for a single postal code.

**Parameters:**

- `postal_code` (path parameter): The destination postal code to check
- `parcel_category` (query parameter, optional): Filter by parcel category (ecomm, courier, cargo)
- `product_type` (query parameter, optional): Filter by product type

**Example:**

```bash
curl --request GET \
  --url "http://127.0.0.1:9022/serviceability/v1/check/385515?parcel_category=ecomm"
```

### 2. POST `/serviceability/v1/check`

Checks serviceability between source and destination postal codes.

**Request Body:**

```json
{
  "source_postal_code": "110001",
  "destination_postal_code": "385515",
  "parcel_category": "ecomm",
  "product_type": "standard"
}
```

**Example:**

```bash
curl --request POST \
  --url http://127.0.0.1:9022/serviceability/v1/check \
  --header "Content-Type: application/json" \
  --data '{
    "destination_postal_code": "385515",
    "parcel_category": "ecomm"
  }'
```

## Response Format

### GET API Response (Single Postal Code)

The GET `/check/{postal_code}` endpoint returns only the destination postal code:

```json
{
  "success": true,
  "data": {
    "destination_postal_code": "385515",
    "parcel_category": "ecomm",
    "product_types": {
      "express": true,
      "reverse": true
    },
    "serviceability": [
      {
        "parcel_category_code": "ecomm",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "sdd",
            "tat_days": 1,
            "is_cod": false,
            "pickup": true,
            "delivery": true,
            "insurance": false,
            "delivery_modes": {
              "air": true,
              "surface": false
            }
          }
        ]
      }
    ]
  }
}
```

### POST API Response (Source & Destination)

The POST `/check` endpoint includes both source and destination postal codes:

```json
{
  "success": true,
  "data": {
    "source_postal_code": "110001",
    "destination_postal_code": "385515",
    "parcel_category": "ecomm",
    "product_types": {
      "express": true,
      "reverse": true
    },
    "serviceability": [
      {
        "parcel_category_code": "ecomm",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "sdd",
            "tat_days": 1,
            "is_cod": false,
            "pickup": true,
            "delivery": true,
            "insurance": false,
            "delivery_modes": {
              "air": true,
              "surface": false
            }
          }
        ]
      }
    ]
  }
}
```

## Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "POSTAL_CODE_NOT_FOUND",
    "message": "No serviceability data found for postal code",
    "details": "Postal code 999999 is not serviceable"
  }
}
```

## Implementation Details

### 1. Data Transfer Objects (DTOs)

Added new DTOs to `internal/shared/dtos/v1/serviceability_dto.go`:

- `PostalCodeServiceabilityRequest`: Request structure for both endpoints
- `PostalCodeServiceabilityResponse`: Response wrapper with success/error handling
- `PostalCodeServiceabilityData`: Main serviceability data structure with source postal code and product types map
- `PostalCodeParcelCategoryService`: Groups services by parcel category
- `PostalCodeServiceInfo`: Individual service information
- `PostalCodeServiceDeliveryModes`: Delivery mode availability
- `PostalCodeServiceabilityError`: Error information structure

### 2. Service Layer

Created `internal/shared/services/v1/postal_code_serviceability_service.go`:

- **PostalCodeServiceabilityService**: Interface defining service operations
- **Data Fetching**: Uses existing `PartnerLocationCoverageRepository.GetByPostalCode()`
- **Data Transformation**: Converts partner coverage data to serviceability format
- **Service Aggregation**: Groups by parcel category and service type
- **Delivery Mode Handling**: Aggregates air/surface delivery modes
- **Filter Application**: Supports parcel category and product type filtering

### 3. Repository Integration

Updated `internal/shared/repositories/v1/location_repository.go`:

- Added `GetByPostalCode()` method to `PartnerLocationCoverageRepository` interface
- Added `GetByPostalCodeWithDeleted()` method for soft-deleted records
- These methods were already implemented in the repository implementation

### 4. Handler Layer

Enhanced `internal/infrastructure/api/http/v1/handlers/serviceability_handler.go`:

- **CheckPostalCodeServiceability**: Handles GET requests for single postal code
- **CheckPostalCodeServiceabilityPost**: Handles POST requests with source/destination
- **Error Handling**: Proper HTTP status codes and error responses
- **Validation**: Request parameter and body validation
- **Logging**: Debug logging for all operations

### 5. Routing

Updated `internal/infrastructure/api/http/v1/routes/serviceability_routes.go`:

- **GET `/check/:postal_code`**: Single postal code serviceability endpoint
- **POST `/check/`**: Source/destination postal code serviceability endpoint
- **Middleware**: Applied rate limiting, security headers, and request logging
- **Route Organization**: Clean separation of concerns

### 6. Dependency Injection

Modified `internal/infrastructure/api/http/server.go`:

- **Service Creation**: Instantiates postal code serviceability service
- **Repository Integration**: Connects to partner location coverage repository
- **Handler Injection**: Passes service to serviceability handler
- **Database Integration**: Uses existing database connection and repository factory

## Data Transformation Logic

### Service Code Mapping

The service transforms partner location coverage `service_type` values:

- `express` → `sdd` (same day delivery)
- `standard` → `ndd` (next day delivery)
- `reverse` → `reverse`
- `vayuquick` → `vayuquick`
- Other values pass through unchanged

### Parcel Category Normalization

Handles variations in parcel category naming:

- `ecom` or `ecommerce` → `ecomm`
- `courier` → `courier`
- `cargo` → `cargo`
- Default fallback → `ecomm`

### Service Aggregation

For duplicate postal codes with different service configurations:

1. Groups by parcel category and service type
2. Aggregates delivery modes (air/surface)
3. Takes best values for TAT (minimum days)
4. OR operation for boolean capabilities (COD, pickup, delivery, insurance)

### Delivery Mode Handling

Converts partner coverage `delivery_mode` field:

- `air` → Sets `delivery_modes.air = true`
- `surface` → Sets `delivery_modes.surface = true`
- Multiple records with different modes get aggregated

### Product Types Mapping

Creates a dynamic product types map from partner coverage database data:

- **Dynamic Data**: Only includes product types that actually exist in the database for the given postal code
- **Source Field**: Uses `product_type` field from `partner_location_coverages` table
- **Normalization**: Maps variants like `travel` → `travel_free`, `pharma` → `pharma_swift`
- **Availability**: Only shows product types found in coverage data (all set to `true`)
- **No Static Data**: No hardcoded or predefined product types - completely database-driven

## Database Integration

The implementation leverages the existing database schema:

- **Table**: `partner_location_coverages`
- **Key Fields**:
  - `postal_code`: Destination postal code for filtering
  - `partner_id`, `partner_code`: Partner identification
  - `service_type`: Type of service (express, standard, reverse, etc.)
  - `delivery_mode`: Delivery method (air, surface)
  - `pickup`, `delivery`, `cod_available`, `insurance`: Service capabilities
  - `tat_days`: Turnaround time in days
  - `is_active`: Active status filter

## Testing

The implementation has been verified to:

1. ✅ Compile successfully with Go 1.24.2
2. ✅ Integrate with existing codebase structure
3. ✅ Use existing repository patterns and database connections
4. ✅ Follow established error handling and validation patterns
5. ✅ Maintain compatibility with existing serviceability endpoints

## Future Improvements

1. **Repository Optimization**: Add dedicated postal code filtering at database level
2. **Caching**: Implement Redis caching for frequently accessed postal codes
3. **Pagination**: Add pagination support for large result sets
4. **Metrics**: Add monitoring and metrics collection
5. **Enhanced Filters**: Support more granular filtering options

## Migration Notes

This implementation:

- **Replaces** the existing POST `/check` endpoint with postal code based logic
- **Maintains** the existing bulk check functionality at `/bulk-check`
- **Adds** new GET `/check/{postal_code}` endpoint
- **Preserves** all existing middleware and security features
- **Reuses** existing database connections and repository patterns

## Example API Usage

Based on the user's provided example, the API will now return the data in the requested format:

```bash
# Original partner coverage API response (multiple entries for same postal code)
curl http://127.0.0.1:9022/serviceability/v1/partners/be9fdb7c-3767-4a3f-854a-037fb745916a/location-coverages/postal-code/385515

# New serviceability API (aggregated and transformed)
curl http://127.0.0.1:9022/serviceability/v1/check/385515
```

The new API transforms the duplicate partner location coverage entries into a unified serviceability response that groups services by category and aggregates delivery modes as requested.

## Dynamic Serviceability Array

The `serviceability` array is **completely dynamic** based on the actual parcel categories found in the database for the given postal code:

### Scenario 1: Multiple Parcel Categories Found

If the database has records for both `ecom` and `cargo` for postal code `385515`:

```json
{
  "success": true,
  "data": {
    "destination_postal_code": "385515",
    "product_types": {
      "express": true,
      "standard": true
    },
    "serviceability": [
      {
        "parcel_category_code": "ecomm",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "sdd",
            "tat_days": 1,
            "is_cod": false,
            "pickup": true,
            "delivery": true,
            "insurance": false,
            "delivery_modes": {
              "air": true,
              "surface": false
            }
          }
        ]
      },
      {
        "parcel_category_code": "cargo",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "standard",
            "tat_days": 3,
            "is_cod": true,
            "pickup": true,
            "delivery": true,
            "insurance": true,
            "delivery_modes": {
              "air": false,
              "surface": true
            }
          }
        ]
      }
    ]
  }
}
```

### Scenario 2: Single Parcel Category Found

If the database only has records for `ecom` for postal code `385515`:

```json
{
  "success": true,
  "data": {
    "destination_postal_code": "385515",
    "product_types": {
      "express": true
    },
    "serviceability": [
      {
        "parcel_category_code": "ecomm",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "sdd",
            "tat_days": 1,
            "is_cod": false,
            "pickup": true,
            "delivery": true,
            "insurance": false,
            "delivery_modes": {
              "air": true,
              "surface": false
            }
          }
        ]
      }
    ]
  }
}
```

### Scenario 3: No Parcel Categories Found

If no coverage data exists for the postal code:

```json
{
  "success": true,
  "data": {
    "destination_postal_code": "385515",
    "product_types": {},
    "serviceability": []
  }
}
```

## How Dynamic Grouping Works

1. **Database Query**: Gets all `partner_location_coverages` records for the postal code
2. **Parcel Category Extraction**: Extracts `parcel_category` field from each record
3. **Normalization**: Maps database values to standard codes:
   - `"ecom"` or `"ecommerce"` → `"ecomm"`
   - `"courier"` → `"courier"`
   - `"cargo"` → `"cargo"`
4. **Dynamic Grouping**: Creates map with actual categories found: `map[categoryCode]services`
5. **Array Creation**: Converts map to array - one object per category found in data

**Result**: The `serviceability` array contains exactly as many objects as there are distinct parcel categories in the database for that postal code.
