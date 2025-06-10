# Error Handling Guide

This guide provides comprehensive information about error handling in the Prayog Serviceability API, including error codes, response formats, and best practices for handling different error scenarios.

## Error Response Format

All API errors follow a consistent format to ensure predictable error handling:

```json
{
  "status": "error",
  "message": "Human-readable error description",
  "code": "MACHINE_READABLE_ERROR_CODE"
}
```

### Error Response Fields

| Field     | Type   | Description                                               |
| --------- | ------ | --------------------------------------------------------- |
| `status`  | string | Always "error" for error responses                        |
| `message` | string | Human-readable error description for developers and users |
| `code`    | string | Machine-readable error code for programmatic handling     |

## HTTP Status Codes

The API uses standard HTTP status codes to indicate the class of error:

### 2xx Success Codes

- **200 OK** - Successful GET/PUT operations
- **201 Created** - Successful POST operations
- **204 No Content** - Successful DELETE operations

### 4xx Client Error Codes

- **400 Bad Request** - Invalid request data or malformed request
- **401 Unauthorized** - Authentication required or invalid credentials
- **403 Forbidden** - Authentication valid but insufficient permissions
- **404 Not Found** - Requested resource does not exist
- **409 Conflict** - Resource already exists or constraint violation
- **413 Payload Too Large** - Request body exceeds size limits
- **415 Unsupported Media Type** - Invalid Content-Type header
- **422 Unprocessable Entity** - Request valid but cannot be processed
- **429 Too Many Requests** - Rate limit exceeded

### 5xx Server Error Codes

- **500 Internal Server Error** - Unexpected server error
- **502 Bad Gateway** - Upstream service error
- **503 Service Unavailable** - Service temporarily unavailable
- **504 Gateway Timeout** - Upstream service timeout

## Error Codes Reference

### Authentication & Authorization Errors

#### `authentication_error`

- **HTTP Status**: 401 Unauthorized
- **Description**: Authentication credentials missing or invalid
- **Common Causes**:
  - Missing Authorization header
  - Invalid or expired JWT token
  - Malformed Bearer token

```json
{
  "status": "error",
  "message": "Invalid or expired authentication token",
  "code": "authentication_error"
}
```

#### `authorization_error`

- **HTTP Status**: 403 Forbidden
- **Description**: Valid authentication but insufficient permissions
- **Common Causes**:
  - User lacks required permissions for the endpoint
  - API key has limited scope
  - Resource access restrictions

```json
{
  "status": "error",
  "message": "Insufficient permissions to access this resource",
  "code": "authorization_error"
}
```

### Validation Errors

#### `validation_error`

- **HTTP Status**: 400 Bad Request
- **Description**: Request validation failed
- **Common Causes**:
  - Missing required fields
  - Invalid field formats
  - Field value constraints violated
  - Invalid JSON structure

```json
{
  "status": "error",
  "message": "Field 'name' is required and cannot be empty",
  "code": "validation_error"
}
```

#### `invalid_uuid`

- **HTTP Status**: 400 Bad Request
- **Description**: Invalid UUID format in path parameter
- **Common Causes**:
  - Malformed UUID in URL path
  - Non-UUID value where UUID expected

```json
{
  "status": "error",
  "message": "Invalid UUID format for parameter 'id'",
  "code": "invalid_uuid"
}
```

### Resource Errors

#### `not_found`

- **HTTP Status**: 404 Not Found
- **Description**: Requested resource does not exist
- **Common Causes**:
  - Invalid resource ID
  - Resource has been deleted
  - Resource is inactive

```json
{
  "status": "error",
  "message": "Country with ID '123e4567-e89b-12d3-a456-426614174000' not found",
  "code": "not_found"
}
```

#### `conflict_error`

- **HTTP Status**: 409 Conflict
- **Description**: Resource already exists or constraint violation
- **Common Causes**:
  - Duplicate unique field values
  - Business rule violations
  - Foreign key constraint violations

```json
{
  "status": "error",
  "message": "Country with code 'US' already exists",
  "code": "conflict_error"
}
```

#### `inactive_resource`

- **HTTP Status**: 422 Unprocessable Entity
- **Description**: Resource exists but is inactive
- **Common Causes**:
  - Trying to use inactive entities
  - Operations on soft-deleted resources

```json
{
  "status": "error",
  "message": "Country 'US' is inactive and cannot be used",
  "code": "inactive_resource"
}
```

### Request Format Errors

#### `invalid_content_type`

- **HTTP Status**: 415 Unsupported Media Type
- **Description**: Invalid or missing Content-Type header
- **Common Causes**:
  - Missing Content-Type header on POST/PUT requests
  - Content-Type other than application/json

```json
{
  "status": "error",
  "message": "Content-Type must be 'application/json'",
  "code": "invalid_content_type"
}
```

#### `invalid_json`

- **HTTP Status**: 400 Bad Request
- **Description**: Malformed JSON in request body
- **Common Causes**:
  - Syntax errors in JSON
  - Incomplete JSON objects
  - Invalid Unicode characters

```json
{
  "status": "error",
  "message": "Invalid JSON in request body",
  "code": "invalid_json"
}
```

#### `payload_too_large`

- **HTTP Status**: 413 Payload Too Large
- **Description**: Request body exceeds maximum size
- **Common Causes**:
  - Large bulk operation requests
  - Excessive field values

```json
{
  "status": "error",
  "message": "Request body exceeds maximum size of 10MB",
  "code": "payload_too_large"
}
```

### Rate Limiting Errors

#### `rate_limit_exceeded`

- **HTTP Status**: 429 Too Many Requests
- **Description**: API rate limit exceeded
- **Common Causes**:
  - Too many requests from same IP
  - Burst traffic exceeding limits
  - Inadequate request distribution

```json
{
  "status": "error",
  "message": "Rate limit exceeded. Try again in 60 seconds",
  "code": "rate_limit_exceeded"
}
```

**Rate Limit Headers**:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1640995200
```

### Business Logic Errors

#### `primary_alias_conflict`

- **HTTP Status**: 409 Conflict
- **Description**: Attempting to create multiple primary aliases
- **Common Causes**:
  - Setting is_primary=true when primary alias exists
  - Business rule violations

```json
{
  "status": "error",
  "message": "Only one primary alias allowed per entity",
  "code": "primary_alias_conflict"
}
```

#### `invalid_entity_reference`

- **HTTP Status**: 422 Unprocessable Entity
- **Description**: Referenced entity does not exist or is invalid
- **Common Causes**:
  - Creating alias for non-existent location
  - Invalid entity type combinations

```json
{
  "status": "error",
  "message": "Referenced entity with ID '123e4567-e89b-12d3-a456-426614174000' does not exist",
  "code": "invalid_entity_reference"
}
```

### Server Errors

#### `internal_error`

- **HTTP Status**: 500 Internal Server Error
- **Description**: Unexpected server error
- **Common Causes**:
  - Database connection issues
  - Unhandled exceptions
  - Service dependency failures

```json
{
  "status": "error",
  "message": "An internal error occurred. Please try again later",
  "code": "internal_error"
}
```

#### `service_unavailable`

- **HTTP Status**: 503 Service Unavailable
- **Description**: Service temporarily unavailable
- **Common Causes**:
  - Scheduled maintenance
  - Service overload
  - Database maintenance

```json
{
  "status": "error",
  "message": "Service temporarily unavailable. Please try again later",
  "code": "service_unavailable"
}
```

## Error Handling Best Practices

### 1. Check Response Status

Always check the HTTP status code and response format:

```javascript
const response = await fetch("/api/v1/countries", {
  headers: { Authorization: "Bearer " + token },
});

if (!response.ok) {
  const error = await response.json();
  console.error("API Error:", error.code, error.message);
  throw new Error(error.message);
}

const data = await response.json();
```

### 2. Implement Retry Logic

For transient errors (5xx, 429), implement exponential backoff:

```javascript
async function apiCallWithRetry(url, options, maxRetries = 3) {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      const response = await fetch(url, options);

      if (response.ok) {
        return await response.json();
      }

      if (response.status >= 500 || response.status === 429) {
        if (attempt === maxRetries) {
          throw new Error("Max retries exceeded");
        }

        const delay = Math.pow(2, attempt) * 1000; // Exponential backoff
        await new Promise((resolve) => setTimeout(resolve, delay));
        continue;
      }

      // Client errors - don't retry
      const error = await response.json();
      throw new Error(error.message);
    } catch (error) {
      if (attempt === maxRetries) {
        throw error;
      }
    }
  }
}
```

### 3. Handle Rate Limits

Use rate limit headers to avoid hitting limits:

```javascript
function handleRateLimit(response) {
  const remaining = parseInt(response.headers.get("X-RateLimit-Remaining"));
  const reset = parseInt(response.headers.get("X-RateLimit-Reset"));

  if (remaining < 10) {
    const resetTime = new Date(reset * 1000);
    console.warn(`Rate limit approaching. Reset at: ${resetTime}`);
  }

  if (response.status === 429) {
    const retryAfter = response.headers.get("Retry-After");
    const delay = retryAfter ? parseInt(retryAfter) * 1000 : 60000;

    return new Promise((resolve) => setTimeout(resolve, delay));
  }
}
```

### 4. Validation Error Handling

Handle validation errors gracefully with user feedback:

```javascript
function handleValidationErrors(error) {
  if (error.code === "validation_error") {
    // Parse validation message for specific field errors
    const fieldMatch = error.message.match(/Field '(\w+)'/);
    if (fieldMatch) {
      const fieldName = fieldMatch[1];
      showFieldError(fieldName, error.message);
    } else {
      showGeneralError(error.message);
    }
  }
}
```

### 5. Circuit Breaker Pattern

Implement circuit breaker for resilient error handling:

```javascript
class CircuitBreaker {
  constructor(threshold = 5, timeout = 60000) {
    this.threshold = threshold;
    this.timeout = timeout;
    this.failures = 0;
    this.state = "CLOSED"; // CLOSED, OPEN, HALF_OPEN
    this.nextAttempt = 0;
  }

  async call(fn) {
    if (this.state === "OPEN") {
      if (Date.now() < this.nextAttempt) {
        throw new Error("Circuit breaker is OPEN");
      }
      this.state = "HALF_OPEN";
    }

    try {
      const result = await fn();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      throw error;
    }
  }

  onSuccess() {
    this.failures = 0;
    this.state = "CLOSED";
  }

  onFailure() {
    this.failures++;
    if (this.failures >= this.threshold) {
      this.state = "OPEN";
      this.nextAttempt = Date.now() + this.timeout;
    }
  }
}
```

## Error Logging and Monitoring

### Client-Side Error Logging

```javascript
function logApiError(error, context) {
  const errorData = {
    timestamp: new Date().toISOString(),
    error_code: error.code,
    message: error.message,
    context: context,
    user_agent: navigator.userAgent,
    url: window.location.href,
  };

  // Send to logging service
  fetch("/api/v1/logs/errors", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(errorData),
  });
}
```

### Error Context Information

Include helpful context when logging errors:

```javascript
try {
  const country = await createCountry({
    name: "United States",
    code: "US",
  });
} catch (error) {
  logApiError(error, {
    operation: "create_country",
    payload: { name: "United States", code: "US" },
    endpoint: "/api/v1/countries",
  });
  throw error;
}
```

## Testing Error Scenarios

### Unit Tests for Error Handling

```javascript
describe("API Error Handling", () => {
  test("should handle validation errors", async () => {
    const response = await fetch("/api/v1/countries", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code: "US" }), // Missing name
    });

    expect(response.status).toBe(400);
    const error = await response.json();
    expect(error.code).toBe("validation_error");
    expect(error.message).toContain("name");
  });

  test("should handle not found errors", async () => {
    const response = await fetch("/api/v1/countries/invalid-id");

    expect(response.status).toBe(404);
    const error = await response.json();
    expect(error.code).toBe("not_found");
  });
});
```

## Common Error Scenarios

### Creating Duplicate Resources

```bash
# First request succeeds
curl -X POST -H "Authorization: Bearer TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name": "United States", "code": "US"}' \
     /api/v1/countries

# Second request with same code fails
curl -X POST -H "Authorization: Bearer TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name": "USA", "code": "US"}' \
     /api/v1/countries

# Returns 409 Conflict with conflict_error code
```

### Invalid UUID in Path

```bash
curl -H "Authorization: Bearer TOKEN" \
     /api/v1/countries/invalid-uuid

# Returns 400 Bad Request with invalid_uuid code
```

### Missing Authentication

```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"name": "United States", "code": "US"}' \
     /api/v1/countries

# Returns 401 Unauthorized with authentication_error code
```

This comprehensive error handling guide ensures robust error management and helps developers build resilient applications with the Prayog Serviceability API.
