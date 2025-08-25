# Porter Hyperlocal Serviceability Integration

## Overview

Porter is a hyperlocal delivery partner that provides serviceability checks based on geographic boundaries. The Porter adapter supports **both postal codes and coordinates** for maximum flexibility.

## Key Features

- **Dual Input Support**: Accepts both postal codes AND coordinates
- **Coordinate Priority**: When both are provided, coordinates take priority (more accurate)
- **Geospatial Queries**: Uses PostGIS to check if locations fall within pickup boundaries
- **Dual Location Check**: Both source AND destination must be "INSIDE" boundaries
- **Hyperlocal Only**: Only activates for `parcel_category: "hyperlocal"`

## API Usage

### 1. Coordinates Only (Most Accurate)

```bash
curl -X POST \
  http://localhost:8080/serviceability/v2/check \
  -H "Content-Type: application/json" \
  -d '{
    "parcel_category": "hyperlocal",
    "source_latitude": 18.5913,
    "source_longitude": 73.7389,
    "destination_latitude": 19.0760,
    "destination_longitude": 72.8777,
    "country_code": "IN"
  }'
```

### 2. Postal Codes Only

```bash
curl -X POST \
  http://localhost:8080/serviceability/v2/check \
  -H "Content-Type: application/json" \
  -d '{
    "parcel_category": "hyperlocal",
    "source_postal_code": "411001",
    "destination_postal_code": "400001",
    "country_code": "IN"
  }'
```

### 3. Mixed Input (Coordinates + Postal Codes)

```bash
curl -X POST \
  http://localhost:8080/serviceability/v2/check \
  -H "Content-Type: application/json" \
  -d '{
    "parcel_category": "hyperlocal",
    "source_latitude": 18.5913,
    "source_longitude": 73.7389,
    "destination_postal_code": "400001",
    "country_code": "IN"
  }'
```

### 4. Complete Request with Packages

```bash
curl -X POST \
  http://localhost:8080/serviceability/v2/check \
  -H "Content-Type: application/json" \
  -d '{
    "parcel_category": "hyperlocal",
    "source_latitude": 18.5913,
    "source_longitude": 73.7389,
    "destination_latitude": 19.0760,
    "destination_longitude": 72.8777,
    "country_code": "IN",
    "packages": [
      {
        "weight": {
          "value": 1.0,
          "unit": "g"
        },
        "dimensions": {
          "length": 1,
          "width": 1,
          "height": 1,
          "unit": "cm"
        }
      }
    ]
  }'
```

## Request Fields

### Required Fields
- `parcel_category`: Must be `"hyperlocal"`
- Location data: Either coordinates OR postal codes for both source and destination

### Optional Fields
- `packages`: Package details (weight, dimensions)
- `country_code`: Defaults to "IN"
- `product_type`: Product type specification

### Location Input Options

#### Option 1: Coordinates (Direct)
```json
{
  "source_latitude": 18.5913,
  "source_longitude": 73.7389,
  "destination_latitude": 19.0760,
  "destination_longitude": 72.8777
}
```

#### Option 2: Postal Codes (Converted to Coordinates)
```json
{
  "source_postal_code": "411001",
  "destination_postal_code": "400001"
}
```

#### Option 3: Mixed (Coordinates take priority)
```json
{
  "source_latitude": 18.5913,
  "source_longitude": 73.7389,
  "destination_postal_code": "400001"
}
```

## Response Format

### Success Response (Serviceable)

```json
{
  "success": true,
  "partners": [
    {
      "partner_id": "porter_001",
      "partner_code": "porter",
      "partner_name": "Porter",
      "rating": 4.0,
      "services": [
        {
          "service_code": "hyperlocal_express",
          "service_name": "Hyperlocal Express",
          "tat_days": 1,
          "is_cod": true,
          "pickup": true,
          "delivery": true
        }
      ],
      "capabilities": {
        "is_serviceable": true,
        "pickup_available": true,
        "delivery_available": true,
        "cod_available": true,
        "insurance_available": false
      },
      "response_time": "50ms"
    }
  ]
}
```

### Success Response (Not Serviceable)

```json
{
  "success": true,
  "partners": [
    {
      "partner_id": "porter_001",
      "partner_code": "porter",
      "partner_name": "Porter",
      "rating": 4.0,
      "services": [],
      "capabilities": {
        "is_serviceable": false,
        "pickup_available": false,
        "delivery_available": false,
        "cod_available": false,
        "insurance_available": false
      },
      "error": "Location Not Serviceable: Either source or destination coordinates fall outside Porter's pickup boundaries",
      "response_time": "45ms"
    }
  ]
}
```

## Database Query

Porter uses the following PostGIS query to check serviceability:

```sql
SELECT
    CASE
        WHEN EXISTS (
            SELECT 1
            FROM pickup_boundaries
            WHERE ST_Contains(
                boundary,
                ST_GeomFromText(?, 4326)
            )
        )
        THEN 'INSIDE'
        ELSE 'OUTSIDE'
    END AS location_status;
```

### Query Execution
- **Executed twice**: Once for source, once for destination
- **Both must return "INSIDE"**: For Porter to be serviceable
- **Point format**: `POINT(longitude latitude)` in WGS84 (EPSG:4326)

## Error Scenarios

### 1. Invalid Parcel Category
```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Porter only supports hyperlocal parcel category"
  }
}
```

### 2. Missing Location Data
```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Porter requires either coordinates OR postal codes for both source and destination locations"
  }
}
```

### 3. Postal Code Conversion Failed
```json
{
  "success": false,
  "error": {
    "code": "GEOCODING_ERROR",
    "message": "Failed to get coordinates for source postal code 999999: postal code not found"
  }
}
```

### 4. Location Not Serviceable
```json
{
  "success": true,
  "partners": [
    {
      "error": "Location Not Serviceable: Either source or destination coordinates fall outside Porter's pickup boundaries"
    }
  ]
}
```

## Configuration

### Environment Variables
```bash
PORTER_ENABLED=true
PORTER_RATING=4.0
PORTER_TIMEOUT=30s
PORTER_MAX_RETRIES=3
PORTER_RETRY_DELAY=1s
```

### Database Requirements
- **Table**: `pickup_boundaries`
- **Column**: `boundary` (PostGIS geometry)
- **SRID**: 4326 (WGS84)

## Business Logic

### Serviceability Rules
1. **Parcel Category**: Must be `"hyperlocal"`
2. **Location Input**: Either coordinates OR postal codes for both locations
3. **Coordinate Priority**: When both provided, coordinates take priority
4. **Dual Check**: Both source AND destination must be "INSIDE" boundaries
5. **Geospatial Query**: Uses PostGIS `ST_Contains` function

### Coordinate Handling
- **Direct Coordinates**: Used immediately for database query
- **Postal Codes**: Converted to coordinates via geolocation service
- **Mixed Input**: Coordinates take priority over postal codes
- **Validation**: Ensures both source and destination have location data

### Performance Considerations
- **Database Connection**: Reuses existing database connection
- **Query Optimization**: Uses spatial indexes on `boundary` column
- **Caching**: Consider implementing coordinate caching for postal codes
- **Timeout**: Configurable timeout for database queries

## Testing

### Unit Tests
Run Porter adapter tests:
```bash
go test ./test/v2/unit/porter_adapter_test.go -v
```

### Integration Tests
Test with actual database:
```bash
# Test coordinates
curl -X POST http://localhost:8080/serviceability/v2/check \
  -H "Content-Type: application/json" \
  -d '{"parcel_category":"hyperlocal","source_latitude":18.5913,"source_longitude":73.7389,"destination_latitude":19.0760,"destination_longitude":72.8777,"country_code":"IN"}'

# Test postal codes
curl -X POST http://localhost:8080/serviceability/v2/check \
  -H "Content-Type: application/json" \
  -d '{"parcel_category":"hyperlocal","source_postal_code":"411001","destination_postal_code":"400001","country_code":"IN"}'
```

## Troubleshooting

### Common Issues
1. **"Porter does not support this request type"**: Check parcel category is "hyperlocal"
2. **"Missing location data"**: Ensure both source and destination have coordinates OR postal codes
3. **"Geolocation service error"**: Verify geolocation service is available for postal code conversion
4. **"Database connection failed"**: Check database connectivity and `pickup_boundaries` table

### Debug Logs
Enable debug logging to see:
- Coordinate extraction process
- Postal code to coordinate conversion
- Database query execution
- Serviceability results

## Migration Guide

### From V1 to V2
- Use `/serviceability/v2/check` endpoint
- Include `parcel_category: "hyperlocal"` for Porter
- Support both coordinates and postal codes
- Enhanced error handling and response format

### Backward Compatibility
- V1 endpoints remain available
- Porter only available in V2
- Coordinate fields are optional in V2 request model
