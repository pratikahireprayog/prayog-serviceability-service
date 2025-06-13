# Prayog Serviceability Service API Documentation

## Overview

The Prayog Serviceability Service provides REST APIs for checking serviceability at postal code level, supporting both single and bulk operations. The service integrates with partner location coverage data to provide real-time serviceability information.

## Base URL

```
Production: https://apis.prayog.io/serviceability/v1
Sandbox: https://sandbox-apis.prayog.io/serviceability/v1
```

## Authentication

Some endpoints require authentication using Bearer tokens:

```bash
Authorization: Bearer YOUR_API_TOKEN
```

## API Endpoints

### 1. Check Single Postal Code Serviceability (GET)

Retrieves serviceability information for a specific postal code.

#### Endpoint

```
GET /check/{postal_code}
```

#### Parameters

| Parameter         | Type   | Location | Required | Description                                             |
| ----------------- | ------ | -------- | -------- | ------------------------------------------------------- |
| `postal_code`     | string | path     | Yes      | Destination postal code to check                        |
| `parcel_category` | string | query    | No       | Filter by parcel category (`ecomm`, `courier`, `cargo`) |
| `product_type`    | string | query    | No       | Filter by product type                                  |

#### Example Request

```bash
curl --request GET \
  --url https://sandbox-apis.prayog.io/serviceability/v1/check/385515 \
  --header 'authorization: Bearer $API_TOKEN' \
  --header 'content-type: application/json'
```

#### Success Response (200 OK)

```json
{
  "success": true,
  "is_serviceable": true,
  "data": {
    "destination_postal_code": "385515",
    "serviceability": [
      {
        "parcel_category_code": "ecomm",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "sdd",
            "tat_days": 1,
            "is_cod": true,
            "pickup": true,
            "delivery": true,
            "insurance": false,
            "product_types": {
              "express": true,
              "standard": false
            },
            "delivery_modes": {
              "air": true,
              "surface": false
            }
          },
          {
            "service_code": "ndd",
            "tat_days": 2,
            "is_cod": true,
            "pickup": true,
            "delivery": true,
            "insurance": true,
            "product_types": {
              "express": false,
              "standard": true
            },
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

#### Error Response - Postal Code Not Found (404 Not Found)

```json
{
  "success": false,
  "is_serviceable": false,
  "error": {
    "code": "POSTAL_CODE_NOT_FOUND",
    "message": "No serviceability data found for postal code",
    "details": "Postal code 385515 is not serviceable"
  }
}
```

---

### 2. Check Source to Destination Serviceability (POST)

Checks serviceability between source and destination postal codes.

#### Endpoint

```
POST /check
```

#### Request Body

| Field                     | Type   | Required | Description                                             |
| ------------------------- | ------ | -------- | ------------------------------------------------------- |
| `source_postal_code`      | string | No       | Source postal code (optional)                           |
| `destination_postal_code` | string | Yes      | Destination postal code                                 |
| `parcel_category`         | string | No       | Filter by parcel category (`ecomm`, `courier`, `cargo`) |
| `product_type`            | string | No       | Filter by product type                                  |

#### Example Request

```bash
curl --request POST \
  --url https://sandbox-apis.prayog.io/serviceability/v1/check \
  --header 'content-type: application/json' \
  --data '{
    "source_postal_code": "713333",
    "destination_postal_code": "385515",
    "parcel_category": "ecomm"
  }'
```

#### Success Response (200 OK)

```json
{
  "success": true,
  "is_serviceable": true,
  "data": {
    "source_postal_code": "713333",
    "destination_postal_code": "385515",
    "serviceability": [
      {
        "parcel_category_code": "ecomm",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "sdd",
            "tat_days": 1,
            "is_cod": true,
            "pickup": true,
            "delivery": true,
            "insurance": false,
            "product_types": {
              "express": true,
              "standard": false
            },
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

#### Error Response - Source Not Serviceable (200 OK)

```json
{
  "success": true,
  "is_serviceable": false,
  "data": {
    "code": "SOURCE_POSTAL_CODE_NOT_SERVICEABLE",
    "message": "Source postal code is not serviceable",
    "details": "Source postal code 713333 is not available in our service area"
  }
}
```

---

### 3. Bulk Serviceability Check

Check serviceability for multiple locations in a single request.

#### Endpoint

```
POST /bulk-check
```

#### Request Body

| Field      | Type  | Required | Description                                      |
| ---------- | ----- | -------- | ------------------------------------------------ |
| `requests` | array | Yes      | Array of serviceability check requests (max 100) |

#### Example Request

```bash
curl --request POST \
  --url https://sandbox-apis.prayog.io/serviceability/v1/bulk-check \
  --header 'content-type: application/json' \
  --data '{
    "requests": [
      {
        "postal_code": "110001",
        "country_code": "IN",
        "parcel_category": "ecomm"
      },
      {
        "pickup_postal_code": "110001",
        "delivery_postal_code": "560001",
        "country_code": "IN"
      }
    ]
  }'
```

#### Success Response (200 OK)

```json
{
  "success": true,
  "data": [
    {
      "success": true,
      "data": {
        "query_type": "generic_location",
        "location": {
          "postal_code": "110001",
          "country_code": "IN",
          "serviceability": [
            {
              "parcel_category": "ecom",
              "services": [
                {
                  "service_type": "SDD",
                  "operation_types": ["pickup", "delivery"],
                  "payment_modes": ["ONLINE", "COD"],
                  "delivery_modes": ["AIR"]
                }
              ]
            }
          ]
        }
      }
    },
    {
      "success": true,
      "data": {
        "query_type": "origin_destination",
        "pickup_location": {
          "postal_code": "110001",
          "country_code": "IN",
          "serviceability": [...]
        },
        "delivery_location": {
          "postal_code": "560001",
          "country_code": "IN",
          "serviceability": [...]
        }
      }
    }
  ]
}
```

#### Partial Failure Response (207 Multi-Status)

```json
{
  "success": false,
  "data": [
    {
      "success": true,
      "data": {
        "query_type": "generic_location",
        "location": {...}
      }
    },
    {
      "success": false,
      "error": {
        "code": "POSTAL_CODE_INACTIVE",
        "message": "Postal code is not active",
        "details": "Postal code 999999 is inactive"
      }
    }
  ],
  "error": {
    "code": "PARTIAL_FAILURE",
    "message": "Processed 2 requests with 1 errors"
  }
}
```

---

### 4. Service Health Check

Check if the serviceability service is available and ready.

#### Endpoint

```
GET /status
```

#### Example Request

```bash
curl --request GET \
  --url https://sandbox-apis.prayog.io/serviceability/v1/status
```

#### Success Response (200 OK)

```json
{
  "service": "serviceability",
  "status": "available",
  "message": "Serviceability API is ready",
  "database": "available",
  "features": {
    "basic_serviceability": "available",
    "postal_code_serviceability": "available",
    "location_management": "available",
    "partner_location_coverage": "available"
  }
}
```

---

## Response Fields Explained

### Serviceability Response Structure

| Field            | Type    | Description                                          |
| ---------------- | ------- | ---------------------------------------------------- |
| `success`        | boolean | Whether the request was processed successfully       |
| `is_serviceable` | boolean | Overall serviceability status (for postal code APIs) |
| `data`           | object  | Serviceability data (present on success)             |
| `error`          | object  | Error information (present on failure)               |

### Serviceability Data Fields

| Field                     | Type   | Description                        |
| ------------------------- | ------ | ---------------------------------- |
| `source_postal_code`      | string | Source postal code (POST API only) |
| `destination_postal_code` | string | Destination postal code            |
| `serviceability`          | array  | Array of parcel category services  |

### Parcel Category Service Fields

| Field                  | Type    | Description                                 |
| ---------------------- | ------- | ------------------------------------------- |
| `parcel_category_code` | string  | Category code (`ecomm`, `courier`, `cargo`) |
| `is_serviceable`       | boolean | Whether this category is serviceable        |
| `services`             | array   | Available services for this category        |

### Service Information Fields

| Field            | Type    | Description                                       |
| ---------------- | ------- | ------------------------------------------------- |
| `service_code`   | string  | Service type code (`sdd`, `ndd`, `express`, etc.) |
| `tat_days`       | integer | Turnaround time in days                           |
| `is_cod`         | boolean | Cash on delivery availability                     |
| `pickup`         | boolean | Pickup service availability                       |
| `delivery`       | boolean | Delivery service availability                     |
| `insurance`      | boolean | Insurance service availability                    |
| `product_types`  | object  | Map of product types and their availability       |
| `delivery_modes` | object  | Available delivery modes (air/surface)            |

---

## Error Codes

| Code                          | HTTP Status | Description                                       |
| ----------------------------- | ----------- | ------------------------------------------------- |
| `INVALID_REQUEST`             | 400         | Invalid request format or missing required fields |
| `VALIDATION_ERROR`            | 400         | Request validation failed                         |
| `INVALID_POSTAL_CODE`         | 400         | Invalid postal code format                        |
| `POSTAL_CODE_NOT_FOUND`       | 404         | Postal code not found in service area             |
| `POSTAL_CODE_INACTIVE`        | 422         | Postal code exists but is inactive                |
| `RATE_LIMIT_EXCEEDED`         | 429         | API rate limit exceeded                           |
| `SERVICEABILITY_CHECK_FAILED` | 500         | Internal service error                            |
| `SERVICE_UNAVAILABLE`         | 503         | Service temporarily unavailable                   |

---

## Rate Limiting

| Endpoint               | Rate Limit              |
| ---------------------- | ----------------------- |
| `/check/{postal_code}` | 100 requests per minute |
| `/check` (POST)        | 100 requests per minute |
| `/bulk-check`          | 20 requests per minute  |

Rate limit headers are included in responses:

- `X-RateLimit-Limit`: Maximum requests per window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Window reset time (Unix timestamp)

---

## Filtering and Scenarios

### 1. Parcel Category Filtering

When `parcel_category` is specified:

- Only serviceability for that category is returned
- If the category doesn't exist, an empty serviceability array is returned

```bash
# Only ecomm services
GET /check/385515?parcel_category=ecomm

# Only cargo services
GET /check/385515?parcel_category=cargo
```

### 2. Product Type Filtering

When `product_type` is specified:

- Services are filtered to show only those supporting the product type
- Product types are dynamically determined from database coverage

```bash
# Only express product type
GET /check/385515?product_type=express
```

### 3. Dynamic Response Generation

The API response is completely database-driven:

- **Parcel categories**: Determined by unique values in database
- **Service codes**: Based on actual service coverage data
- **Product types**: Aggregated from all available services
- **Array size**: Equals number of unique parcel categories found

### 4. Different Scenarios

#### Scenario 1: Multiple Categories Available

Database contains `ecomm`, `courier`, `cargo` → Response has 3 serviceability objects

#### Scenario 2: Single Category Available

Database contains only `ecomm` → Response has 1 serviceability object

#### Scenario 3: No Coverage Data

Database has no records → Response has empty serviceability array

#### Scenario 4: Source Not Serviceable (POST API)

Source postal code not found → Special error response indicating source issues

---

## Integration Examples

### JavaScript/Node.js

```javascript
// Single postal code check
const response = await fetch(
  "https://sandbox-apis.prayog.io/serviceability/v1/check/385515",
  {
    headers: {
      Authorization: "Bearer YOUR_API_TOKEN",
      "Content-Type": "application/json",
    },
  }
);
const data = await response.json();
console.log("Serviceable:", data.is_serviceable);

// Source to destination check
const postResponse = await fetch(
  "https://sandbox-apis.prayog.io/serviceability/v1/check",
  {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      source_postal_code: "713333",
      destination_postal_code: "385515",
      parcel_category: "ecomm",
    }),
  }
);
const postData = await postResponse.json();
```

### Python

```python
import requests

# Single postal code check
response = requests.get(
    'https://sandbox-apis.prayog.io/serviceability/v1/check/385515',
    headers={
        'Authorization': 'Bearer YOUR_API_TOKEN',
        'Content-Type': 'application/json'
    }
)
data = response.json()
print(f"Serviceable: {data['is_serviceable']}")

# Source to destination check
post_response = requests.post(
    'https://sandbox-apis.prayog.io/serviceability/v1/check',
    headers={'Content-Type': 'application/json'},
    json={
        'source_postal_code': '713333',
        'destination_postal_code': '385515',
        'parcel_category': 'ecomm'
    }
)
post_data = post_response.json()
```

---

## Best Practices

1. **Always check the `success` field** before processing response data
2. **Handle rate limiting** by implementing exponential backoff
3. **Cache responses** when appropriate to reduce API calls
4. **Use bulk endpoint** for multiple checks to improve efficiency
5. **Monitor error codes** to understand service health and data quality
6. **Implement timeout handling** for network resilience
7. **Validate postal codes** on client side before API calls

---

## Support

For API support and questions:

- Email: api-support@prayog.io
- Documentation: https://docs.prayog.io
- Status Page: https://status.prayog.io
