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

# Prayog Serviceability Service API Documentation

## Overview

The Prayog Serviceability Service API provides endpoints to check service availability for postal codes and origin-destination pairs. This service enables you to determine if delivery services are available for specific locations or routes.

## Base URL

```
Production: https://api.prayog.com/api/v1/serviceability
Staging: https://staging-serviceability.prayog.com/api/v1/serviceability
Local: http://localhost:8080/api/v1/serviceability
```

## API Features

- ✅ **Single Location Check**: Check serviceability for a specific postal code
- ✅ **Origin-Destination Check**: Check serviceability between pickup and delivery locations
- ✅ **Bulk Operations**: Process multiple serviceability checks in a single request
- ✅ **Rate Limiting**: Built-in protection against abuse
- ✅ **Comprehensive Error Handling**: Detailed error responses with standardized error codes
- ✅ **Request Validation**: Input validation with clear error messages

## Quick Start

### 1. Single Location Check

Check if a postal code is serviceable:

```bash
curl -X POST \
  http://localhost:8080/api/v1/serviceability/check \
  -H "Content-Type: application/json" \
  -d '{
    "postal_code": "110001",
    "country_code": "IN",
    "service_types": ["STANDARD", "EXPRESS"]
  }'
```

**Response:**

```json
{
  "success": true,
  "data": {
    "serviceable": true,
    "services": [
      {
        "service_type": "STANDARD",
        "partner_id": "partner-123",
        "partner_name": "Express Logistics",
        "estimated_delivery_days": 3,
        "service_available": true
      },
      {
        "service_type": "EXPRESS",
        "partner_id": "partner-456",
        "partner_name": "FastTrack Delivery",
        "estimated_delivery_days": 1,
        "service_available": true
      }
    ],
    "location_info": {
      "postal_code": "110001",
      "city": "New Delhi",
      "state": "Delhi",
      "country": "IN"
    }
  }
}
```

### 2. Origin-Destination Check

Check serviceability between pickup and delivery locations:

```bash
curl -X POST \
  http://localhost:8080/api/v1/serviceability/check \
  -H "Content-Type: application/json" \
  -d '{
    "pickup_postal_code": "110001",
    "delivery_postal_code": "560001",
    "country_code": "IN",
    "service_types": ["STANDARD"]
  }'
```

### 3. Bulk Check

Process multiple checks in a single request:

```bash
curl -X POST \
  http://localhost:8080/api/v1/serviceability/bulk-check \
  -H "Content-Type: application/json" \
  -d '{
    "requests": [
      {
        "postal_code": "110001",
        "country_code": "IN",
        "service_types": ["STANDARD"]
      },
      {
        "pickup_postal_code": "110001",
        "delivery_postal_code": "560001",
        "country_code": "IN",
        "service_types": ["EXPRESS"]
      }
    ]
  }'
```

## API Endpoints

### POST `/check`

Check serviceability for a single location or route.

**Parameters:**

| Parameter              | Type   | Required      | Description                                     |
| ---------------------- | ------ | ------------- | ----------------------------------------------- |
| `postal_code`          | string | conditional\* | Postal code for single location check           |
| `pickup_postal_code`   | string | conditional\* | Pickup postal code for route check              |
| `delivery_postal_code` | string | conditional\* | Delivery postal code for route check            |
| `country_code`         | string | yes           | ISO 3166-1 alpha-2 country code (default: "IN") |
| `service_types`        | array  | no            | List of service types to check                  |
| `partner_ids`          | array  | no            | Specific partner IDs to check                   |

\*Either `postal_code` OR both `pickup_postal_code` and `delivery_postal_code` are required.

**Service Types:**

- `STANDARD` - Standard delivery
- `EXPRESS` - Express delivery
- `OVERNIGHT` - Overnight delivery
- `SAME_DAY` - Same day delivery

### POST `/bulk-check`

Process multiple serviceability checks in a single request.

**Parameters:**

| Parameter  | Type  | Required | Description                                      |
| ---------- | ----- | -------- | ------------------------------------------------ |
| `requests` | array | yes      | Array of serviceability check requests (max 100) |

Each request in the array follows the same format as the single check endpoint.

### GET `/status`

Check service health and availability.

## Rate Limiting

| Endpoint      | Limit        | Window            |
| ------------- | ------------ | ----------------- |
| `/check`      | 100 requests | per minute per IP |
| `/bulk-check` | 20 requests  | per minute per IP |
| `/status`     | No limit     | -                 |

When rate limits are exceeded, you'll receive a `429 Too Many Requests` response.

## Error Handling

All errors follow a standardized format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Additional error details"
  }
}
```

### Error Codes

| Code                          | HTTP Status | Description                           |
| ----------------------------- | ----------- | ------------------------------------- |
| `INVALID_REQUEST_BODY`        | 400         | Request body parsing failed           |
| `VALIDATION_ERROR`            | 400         | Input validation failed               |
| `INVALID_REQUEST_TYPE`        | 400         | Invalid combination of parameters     |
| `INVALID_POSTAL_CODE`         | 400         | Postal code format is invalid         |
| `INVALID_COUNTRY_CODE`        | 400         | Country code format is invalid        |
| `POSTAL_CODE_INACTIVE`        | 422         | Postal code is not active             |
| `REQUEST_TOO_LARGE`           | 413         | Request body exceeds 10MB limit       |
| `RATE_LIMIT_EXCEEDED`         | 429         | Rate limit exceeded                   |
| `INVALID_CONTENT_TYPE`        | 415         | Content-Type must be application/json |
| `SERVICEABILITY_CHECK_FAILED` | 500         | Service check failed                  |
| `SERVICE_UNAVAILABLE`         | 503         | Service temporarily unavailable       |
| `PARTIAL_FAILURE`             | 207         | Some bulk requests failed             |

## Request Examples

### Valid Request Patterns

#### Single Location Check

```json
{
  "postal_code": "110001",
  "country_code": "IN"
}
```

#### Origin-Destination Check

```json
{
  "pickup_postal_code": "110001",
  "delivery_postal_code": "560001",
  "country_code": "IN"
}
```

#### With Service Type Filtering

```json
{
  "postal_code": "110001",
  "country_code": "IN",
  "service_types": ["EXPRESS", "SAME_DAY"]
}
```

#### With Partner Filtering

```json
{
  "postal_code": "110001",
  "country_code": "IN",
  "partner_ids": ["partner-123", "partner-456"]
}
```

### Invalid Request Examples

#### ❌ Missing Required Fields

```json
{
  "country_code": "IN"
}
```

_Error: Must specify either postal_code or both pickup/delivery postal codes_

#### ❌ Mixed Request Types

```json
{
  "postal_code": "110001",
  "pickup_postal_code": "110002",
  "country_code": "IN"
}
```

_Error: Cannot specify both postal_code and pickup/delivery postal codes_

#### ❌ Incomplete Route

```json
{
  "pickup_postal_code": "110001",
  "country_code": "IN"
}
```

_Error: Both pickup_postal_code and delivery_postal_code are required_

## Response Examples

### Successful Response

```json
{
  "success": true,
  "data": {
    "serviceable": true,
    "services": [
      {
        "service_type": "STANDARD",
        "partner_id": "partner-123",
        "partner_name": "Express Logistics",
        "service_available": true,
        "estimated_delivery_days": 3,
        "additional_info": {
          "cutoff_time": "18:00",
          "pickup_available": true
        }
      }
    ],
    "location_info": {
      "postal_code": "110001",
      "city": "New Delhi",
      "state": "Delhi",
      "country": "IN",
      "coordinates": {
        "latitude": 28.6139,
        "longitude": 77.209
      }
    }
  }
}
```

### Non-Serviceable Location

```json
{
  "success": true,
  "data": {
    "serviceable": false,
    "services": [],
    "location_info": {
      "postal_code": "999999",
      "city": "Remote Area",
      "state": "Unknown",
      "country": "IN"
    }
  }
}
```

### Bulk Response

```json
{
  "success": true,
  "data": [
    {
      "success": true,
      "data": {
        "serviceable": true,
        "services": [...]
      }
    },
    {
      "success": false,
      "error": {
        "code": "POSTAL_CODE_INACTIVE",
        "message": "Postal code is not active"
      }
    }
  ]
}
```

## Best Practices

### 1. Request Optimization

- Use bulk requests for multiple checks to reduce API calls
- Cache results for frequently checked locations
- Include only necessary service types and partners

### 2. Error Handling

- Always check the `success` field before processing data
- Implement retry logic for 5xx errors with exponential backoff
- Handle rate limiting with appropriate delays

### 3. Performance

- Use connection pooling for multiple requests
- Implement timeouts (recommended: 30 seconds)
- Monitor response times and implement circuit breakers

### 4. Security

- Validate all input data before sending requests
- Use HTTPS in production
- Implement request signing if using API keys

## Code Examples

### JavaScript/Node.js

```javascript
const axios = require("axios");

class ServiceabilityClient {
  constructor(baseURL = "http://localhost:8080/api/v1/serviceability") {
    this.client = axios.create({
      baseURL,
      timeout: 30000,
      headers: { "Content-Type": "application/json" },
    });
  }

  async checkServiceability(request) {
    try {
      const response = await this.client.post("/check", request);
      return response.data;
    } catch (error) {
      if (error.response) {
        throw new Error(`API Error: ${error.response.data.error.message}`);
      }
      throw error;
    }
  }

  async bulkCheck(requests) {
    try {
      const response = await this.client.post("/bulk-check", { requests });
      return response.data;
    } catch (error) {
      if (error.response) {
        throw new Error(`API Error: ${error.response.data.error.message}`);
      }
      throw error;
    }
  }
}

// Usage
const client = new ServiceabilityClient();

// Single check
const result = await client.checkServiceability({
  postal_code: "110001",
  country_code: "IN",
  service_types: ["STANDARD", "EXPRESS"],
});

console.log("Serviceable:", result.data.serviceable);
```

### Python

```python
import requests
from typing import List, Dict, Any

class ServiceabilityClient:
    def __init__(self, base_url: str = "http://localhost:8080/api/v1/serviceability"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({"Content-Type": "application/json"})

    def check_serviceability(self, request: Dict[str, Any]) -> Dict[str, Any]:
        try:
            response = self.session.post(f"{self.base_url}/check", json=request, timeout=30)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise Exception(f"API Error: {e}")

    def bulk_check(self, requests: List[Dict[str, Any]]) -> Dict[str, Any]:
        try:
            response = self.session.post(
                f"{self.base_url}/bulk-check",
                json={"requests": requests},
                timeout=30
            )
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise Exception(f"API Error: {e}")

# Usage
client = ServiceabilityClient()

# Single check
result = client.check_serviceability({
    "postal_code": "110001",
    "country_code": "IN",
    "service_types": ["STANDARD", "EXPRESS"]
})

print(f"Serviceable: {result['data']['serviceable']}")
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type ServiceabilityClient struct {
    baseURL string
    client  *http.Client
}

type ServiceabilityRequest struct {
    PostalCode         *string  `json:"postal_code,omitempty"`
    PickupPostalCode   *string  `json:"pickup_postal_code,omitempty"`
    DeliveryPostalCode *string  `json:"delivery_postal_code,omitempty"`
    CountryCode        string   `json:"country_code"`
    ServiceTypes       []string `json:"service_types,omitempty"`
    PartnerIds         []string `json:"partner_ids,omitempty"`
}

func NewServiceabilityClient(baseURL string) *ServiceabilityClient {
    return &ServiceabilityClient{
        baseURL: baseURL,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *ServiceabilityClient) CheckServiceability(req ServiceabilityRequest) (map[string]interface{}, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := c.client.Post(
        c.baseURL+"/check",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result, nil
}

// Usage
func main() {
    client := NewServiceabilityClient("http://localhost:8080/api/v1/serviceability")

    postalCode := "110001"
    result, err := client.CheckServiceability(ServiceabilityRequest{
        PostalCode:   &postalCode,
        CountryCode:  "IN",
        ServiceTypes: []string{"STANDARD", "EXPRESS"},
    })

    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Result: %+v\n", result)
}
```

## Troubleshooting

### Common Issues

1. **400 Bad Request - Validation Error**

   - Check that required fields are provided
   - Ensure postal codes match the expected format
   - Verify country code is a valid 2-letter ISO code

2. **422 Unprocessable Entity**

   - The postal code may not be active in the system
   - Check if the location is supported

3. **429 Too Many Requests**

   - Implement exponential backoff
   - Consider using bulk endpoints for multiple checks
   - Distribute requests across multiple IPs if possible

4. **500 Internal Server Error**
   - Retry the request after a delay
   - Check service status at `/status` endpoint
   - Contact support if the issue persists

## Support

For API support and questions:

- Email: engineering@prayog.com
- Documentation: https://docs.prayog.com/serviceability
- Status Page: https://status.prayog.com

## Changelog

### v1.0.0 (Current)

- Initial release with single and bulk serviceability checks
- Comprehensive error handling and validation
- Rate limiting and security features
- OpenAPI 3.0 specification
