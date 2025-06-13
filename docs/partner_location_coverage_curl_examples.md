# Partner Location Coverage API - Curl Examples

## Base Configuration

```bash
# Set your base URL and partner identifiers
BASE_URL="http://localhost:9022/serviceability/v1"
PARTNER_ID="550e8400-e29b-41d4-a716-446655440000"
PARTNER_CODE="SHIPYAARI_001"
POSTAL_CODE="110001"
```

## 1. Create Partner Location Coverage

### Basic Coverage Creation (Using Partner ID)

```bash
curl -X POST \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "partner_code": "SHIPYAARI_001",
    "postal_code": "110001",
    "zone_type": "PRIMARY",
    "is_active": true
  }'
```

### Basic Coverage Creation (Using Partner Code)

```bash
curl -X POST \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "postal_code": "110001",
    "zone_type": "PRIMARY",
    "is_active": true
  }'
```

### Complete Coverage Creation with All Serviceability Fields

```bash
curl -X POST \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "partner_code": "SHIPYAARI_001",
    "postal_code": "110001",
    "zone_type": "PRIMARY",
    "country_code": "IN",
    "product_type": "travel_free",
    "parcel_category": "ecomm",
    "service_type": "sdd",
    "tat_days": 1,
    "pickup": true,
    "delivery": true,
    "delivery_mode": "surface",
    "cod_available": true,
    "insurance": true,
    "min_weight_kg": 0.1,
    "max_weight_kg": 50.0,
    "is_active": true
  }'
```

### Express Delivery Service Coverage

```bash
curl -X POST \
  "${BASE_URL}/partners/code/EXPRESS_PARTNER/location-coverages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "postal_code": "400001",
    "zone_type": "PRIMARY",
    "country_code": "IN",
    "product_type": "express",
    "parcel_category": "courier",
    "service_type": "sdd",
    "tat_days": 0,
    "pickup": true,
    "delivery": true,
    "delivery_mode": "air",
    "cod_available": false,
    "insurance": true,
    "min_weight_kg": 0.05,
    "max_weight_kg": 5.0,
    "is_active": true
  }'
```

## 2. Update Partner Location Coverage (By Partner + Postal Code)

### Partial Update - Service Capabilities Only (Using Partner ID)

```bash
curl -X PUT \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages/postal-code/${POSTAL_CODE}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "pickup": false,
    "delivery": true,
    "cod_available": true,
    "insurance": false
  }'
```

### Partial Update - Service Capabilities Only (Using Partner Code)

```bash
curl -X PUT \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages/postal-code/${POSTAL_CODE}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "pickup": false,
    "delivery": true,
    "cod_available": true,
    "insurance": false
  }'
```

### Complete Update with All Fields

```bash
curl -X PUT \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages/postal-code/110002" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "partner_code": "UPDATED_PARTNER",
    "zone_type": "SECONDARY",
    "country_code": "IN",
    "product_type": "premium",
    "parcel_category": "cargo",
    "service_type": "ndd",
    "tat_days": 2,
    "pickup": true,
    "delivery": true,
    "delivery_mode": "rail",
    "cod_available": true,
    "insurance": true,
    "min_weight_kg": 1.0,
    "max_weight_kg": 100.0,
    "is_active": true
  }'
```

### Update TAT and Weight Constraints

```bash
curl -X PUT \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages/postal-code/${POSTAL_CODE}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "tat_days": 3,
    "min_weight_kg": 0.5,
    "max_weight_kg": 75.0,
    "delivery_mode": "surface"
  }'
```

## 3. Get Partner Location Coverage by Partner + Postal Code

### Get Coverage by Partner ID + Postal Code

```bash
curl -X GET \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages/postal-code/${POSTAL_CODE}" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Get Coverage by Partner Code + Postal Code

```bash
curl -X GET \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages/postal-code/${POSTAL_CODE}" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 4. List Partner Location Coverages with Filters

### Basic List (No Filters) - Using Partner ID

```bash
curl -X GET \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Basic List (No Filters) - Using Partner Code

```bash
curl -X GET \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### With Pagination

```bash
curl -X GET \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages?limit=20&offset=0" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Filter by Postal Code

```bash
curl -X GET \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages?postal_code=110001" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Filter by Service Type and Capabilities

```bash
curl -X GET \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages?service_type=sdd&pickup=true&cod_available=true" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Filter by Parcel Category and Weight Range

```bash
curl -X GET \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages?parcel_category=ecomm&max_weight_kg=25.0" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Complex Filter - Express Services Only

```bash
curl -X GET \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages?service_type=sdd&tat_days=1&delivery_mode=air&insurance=true" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Filter by Country and Product Type

```bash
curl -X GET \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages?country_code=IN&product_type=travel_free&is_active=true" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Filter with URL Encoding for Multiple Parameters

```bash
curl -X GET \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages?parcel_category=ecomm&service_type=sdd&pickup=true&delivery=true&cod_available=true&limit=10&offset=0" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 5. Bulk Create Partner Location Coverages

### Create Multiple Coverages for Different Postal Codes (Using Partner ID)

```bash
curl -X POST \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages/bulk" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "coverages": [
      {
        "partner_code": "BULK_PARTNER",
        "postal_code": "110001",
        "zone_type": "PRIMARY",
        "country_code": "IN",
        "product_type": "standard",
        "parcel_category": "ecomm",
        "service_type": "sdd",
        "tat_days": 1,
        "pickup": true,
        "delivery": true,
        "delivery_mode": "surface",
        "cod_available": true,
        "insurance": true,
        "min_weight_kg": 0.1,
        "max_weight_kg": 25.0,
        "is_active": true
      },
      {
        "partner_code": "BULK_PARTNER",
        "postal_code": "110002",
        "zone_type": "SECONDARY",
        "country_code": "IN",
        "product_type": "standard",
        "parcel_category": "ecomm",
        "service_type": "ndd",
        "tat_days": 2,
        "pickup": false,
        "delivery": true,
        "delivery_mode": "surface",
        "cod_available": true,
        "insurance": false,
        "min_weight_kg": 0.5,
        "max_weight_kg": 50.0,
        "is_active": true
      },
      {
        "partner_code": "BULK_PARTNER",
        "postal_code": "400001",
        "zone_type": "PRIMARY",
        "country_code": "IN",
        "product_type": "express",
        "parcel_category": "courier",
        "service_type": "sdd",
        "tat_days": 0,
        "pickup": true,
        "delivery": true,
        "delivery_mode": "air",
        "cod_available": false,
        "insurance": true,
        "min_weight_kg": 0.1,
        "max_weight_kg": 10.0,
        "is_active": true
      }
    ]
  }'
```

### Bulk Create Using Partner Code

```bash
curl -X POST \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages/bulk" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "coverages": [
      {
        "postal_code": "110001",
        "zone_type": "PRIMARY",
        "service_type": "sdd",
        "tat_days": 1,
        "pickup": true,
        "delivery": true,
        "is_active": true
      },
      {
        "postal_code": "110002",
        "zone_type": "SECONDARY",
        "service_type": "ndd",
        "tat_days": 2,
        "pickup": false,
        "delivery": true,
        "is_active": true
      }
    ]
  }'
```

## 6. Delete Partner Location Coverage (By Partner + Postal Code)

### Delete by Partner ID + Postal Code

```bash
curl -X DELETE \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages/postal-code/${POSTAL_CODE}" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Delete by Partner Code + Postal Code

```bash
curl -X DELETE \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages/postal-code/${POSTAL_CODE}" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 7. Check Coverage Availability

### Check if Partner Covers Specific Postal Code

```bash
curl -X GET \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages/postal-code/${POSTAL_CODE}/check" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Check Coverage with Service Requirements

```bash
curl -X GET \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages/postal-code/${POSTAL_CODE}/check?service_type=sdd&pickup=true&cod_available=true" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Response Examples

### Successful Create Response

```json
{
  "status": "success",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "partner_id": "550e8400-e29b-41d4-a716-446655440000",
    "partner_code": "SHIPYAARI_001",
    "postal_code": "110001",
    "postal_code_id": "770e8400-e29b-41d4-a716-446655440002",
    "zone_type": "PRIMARY",
    "country_code": "IN",
    "product_type": "travel_free",
    "parcel_category": "ecomm",
    "service_type": "sdd",
    "tat_days": 1,
    "pickup": true,
    "delivery": true,
    "delivery_mode": "surface",
    "cod_available": true,
    "insurance": true,
    "min_weight_kg": 0.1,
    "max_weight_kg": 50.0,
    "is_active": true,
    "created_at": "2023-12-01T10:00:00Z",
    "updated_at": "2023-12-01T10:00:00Z"
  }
}
```

### Get Coverage by Partner + Postal Code Response

```json
{
  "status": "success",
  "data": {
    "partner_id": "550e8400-e29b-41d4-a716-446655440000",
    "partner_code": "SHIPYAARI_001",
    "postal_code": "110001",
    "zone_type": "PRIMARY",
    "country_code": "IN",
    "service_type": "sdd",
    "tat_days": 1,
    "pickup": true,
    "delivery": true,
    "delivery_mode": "surface",
    "cod_available": true,
    "insurance": true,
    "min_weight_kg": 0.1,
    "max_weight_kg": 50.0,
    "is_active": true
  }
}
```

### Coverage Check Response

```json
{
  "status": "success",
  "data": {
    "covered": true,
    "partner_code": "SHIPYAARI_001",
    "postal_code": "110001",
    "service_capabilities": {
      "pickup": true,
      "delivery": true,
      "cod_available": true,
      "insurance": true,
      "service_type": "sdd",
      "tat_days": 1,
      "delivery_mode": "surface",
      "weight_range": {
        "min_kg": 0.1,
        "max_kg": 50.0
      }
    }
  }
}
```

### List Response with Pagination

```json
{
  "status": "success",
  "data": {
    "coverages": [
      {
        "postal_code": "110001",
        "zone_type": "PRIMARY",
        "service_type": "sdd",
        "tat_days": 1,
        "pickup": true,
        "delivery": true,
        "cod_available": true,
        "is_active": true
      },
      {
        "postal_code": "110002",
        "zone_type": "SECONDARY",
        "service_type": "ndd",
        "tat_days": 2,
        "pickup": false,
        "delivery": true,
        "cod_available": true,
        "is_active": true
      }
    ],
    "pagination": {
      "offset": 0,
      "limit": 10,
      "total": 2,
      "has_next": false,
      "has_previous": false
    }
  }
}
```

### Bulk Create Response

```json
{
  "status": "success",
  "data": {
    "created": [
      {
        "postal_code": "110001",
        "service_type": "sdd",
        "status": "created"
      },
      {
        "postal_code": "110002",
        "service_type": "ndd",
        "status": "created"
      }
    ],
    "updated": [
      {
        "postal_code": "400001",
        "service_type": "sdd",
        "status": "updated",
        "message": "existing coverage updated with new serviceability data"
      }
    ],
    "failed": [
      {
        "postal_code": "500001",
        "error": "invalid postal code format",
        "request": {
          "postal_code": "500001",
          "service_type": "sdd"
        }
      }
    ]
  }
}
```

## API Endpoint Structure

### Partner ID Based Operations

- `POST /partners/{partner_id}/location-coverages` - Create coverage
- `GET /partners/{partner_id}/location-coverages` - List all coverages
- `GET /partners/{partner_id}/location-coverages/postal-code/{postal_code}` - Get specific coverage
- `PUT /partners/{partner_id}/location-coverages/postal-code/{postal_code}` - Update coverage
- `DELETE /partners/{partner_id}/location-coverages/postal-code/{postal_code}` - Delete coverage
- `POST /partners/{partner_id}/location-coverages/bulk` - Bulk create
- `GET /partners/{partner_id}/location-coverages/postal-code/{postal_code}/check` - Check coverage

### Partner Code Based Operations

- `POST /partners/code/{partner_code}/location-coverages` - Create coverage
- `GET /partners/code/{partner_code}/location-coverages` - List all coverages
- `GET /partners/code/{partner_code}/location-coverages/postal-code/{postal_code}` - Get specific coverage
- `PUT /partners/code/{partner_code}/location-coverages/postal-code/{postal_code}` - Update coverage
- `DELETE /partners/code/{partner_code}/location-coverages/postal-code/{postal_code}` - Delete coverage
- `POST /partners/code/{partner_code}/location-coverages/bulk` - Bulk create
- `GET /partners/code/{partner_code}/location-coverages/postal-code/{postal_code}/check` - Check coverage

## Query Parameter Reference

### Available Filter Parameters

- `postal_code` - Filter by specific postal code
- `postal_code_id` - Filter by postal code UUID
- `zone_type` - Filter by zone type (PRIMARY, SECONDARY, BUFFER)
- `country_code` - Filter by country code
- `product_type` - Filter by product type
- `parcel_category` - Filter by parcel category (ecomm, courier, cargo)
- `service_type` - Filter by service type (sdd, ndd, standard)
- `tat_days` - Filter by specific TAT days
- `pickup` - Filter by pickup availability (true/false)
- `delivery` - Filter by delivery availability (true/false)
- `delivery_mode` - Filter by delivery mode (air, surface, rail)
- `cod_available` - Filter by COD availability (true/false)
- `insurance` - Filter by insurance availability (true/false)
- `min_weight_kg` - Filter by minimum weight capability
- `max_weight_kg` - Filter by maximum weight capability
- `is_active` - Filter by active status (true/false)
- `limit` - Number of records to return (1-100, default: 10)
- `offset` - Number of records to skip (default: 0)

### URL Encoding Examples

```bash
# For boolean values
pickup=true
delivery=false

# For numeric values
tat_days=1
min_weight_kg=0.5
max_weight_kg=25.0

# For string values (no encoding needed for simple values)
service_type=sdd
delivery_mode=surface
```

## Error Handling Examples

### Validation Error Response

```json
{
  "status": "error",
  "message": "Validation failed",
  "errors": [
    {
      "field": "tat_days",
      "message": "must be between 0 and 365"
    },
    {
      "field": "delivery_mode",
      "message": "must be one of: air, surface, rail"
    }
  ]
}
```

### Not Found Error

```json
{
  "status": "error",
  "message": "partner location coverage not found for postal code: 110001"
}
```

### Partner Not Found Error

```json
{
  "status": "error",
  "message": "partner not found with code: INVALID_PARTNER"
}
```

## Testing Scripts

### Complete Test Script for Partner ID Based Operations

```bash
#!/bin/bash

BASE_URL="http://localhost:9022/serviceability/v1"
PARTNER_ID="550e8400-e29b-41d4-a716-446655440000"
POSTAL_CODE="110001"

# Test Create
echo "Testing Create Coverage..."
curl -s -X POST \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages" \
  -H "Content-Type: application/json" \
  -d '{
    "postal_code": "'${POSTAL_CODE}'",
    "service_type": "sdd",
    "pickup": true,
    "delivery": true,
    "cod_available": true
  }' | jq

# Test Get by Partner + Postal Code
echo "Testing Get Coverage by Partner + Postal Code..."
curl -s -X GET \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages/postal-code/${POSTAL_CODE}" | jq

# Test Update
echo "Testing Update Coverage..."
curl -s -X PUT \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages/postal-code/${POSTAL_CODE}" \
  -H "Content-Type: application/json" \
  -d '{"tat_days": 2, "insurance": true}' | jq

# Test Check Coverage
echo "Testing Check Coverage..."
curl -s -X GET \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages/postal-code/${POSTAL_CODE}/check?service_type=sdd&pickup=true" | jq

# Test List with Filters
echo "Testing List with Filters..."
curl -s -X GET \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages?service_type=sdd&pickup=true" | jq

# Test Delete
echo "Testing Delete Coverage..."
curl -s -X DELETE \
  "${BASE_URL}/partners/${PARTNER_ID}/location-coverages/postal-code/${POSTAL_CODE}"
```

### Complete Test Script for Partner Code Based Operations

```bash
#!/bin/bash

BASE_URL="http://localhost:9022/serviceability/v1"
PARTNER_CODE="SHIPYAARI_001"
POSTAL_CODE="110001"

# Test Create
echo "Testing Create Coverage with Partner Code..."
curl -s -X POST \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages" \
  -H "Content-Type: application/json" \
  -d '{
    "postal_code": "'${POSTAL_CODE}'",
    "service_type": "sdd",
    "pickup": true,
    "delivery": true,
    "cod_available": true
  }' | jq

# Test Get by Partner Code + Postal Code
echo "Testing Get Coverage by Partner Code + Postal Code..."
curl -s -X GET \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages/postal-code/${POSTAL_CODE}" | jq

# Test Update
echo "Testing Update Coverage with Partner Code..."
curl -s -X PUT \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages/postal-code/${POSTAL_CODE}" \
  -H "Content-Type: application/json" \
  -d '{"tat_days": 2, "insurance": true}' | jq

# Test Check Coverage
echo "Testing Check Coverage with Partner Code..."
curl -s -X GET \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages/postal-code/${POSTAL_CODE}/check" | jq

# Test List with Filters
echo "Testing List with Partner Code..."
curl -s -X GET \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages?pickup=true" | jq

# Test Delete
echo "Testing Delete Coverage with Partner Code..."
curl -s -X DELETE \
  "${BASE_URL}/partners/code/${PARTNER_CODE}/location-coverages/postal-code/${POSTAL_CODE}"
```
