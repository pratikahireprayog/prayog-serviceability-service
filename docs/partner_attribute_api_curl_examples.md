# Partner Attribute API - cURL Testing Examples

This document provides comprehensive cURL examples for testing the Partner Attribute API endpoints.

## Base URL

```
https://sandbox-apis.prayog.io/serviceability/v1
```

For local development:

```
http://localhost:8080/serviceability/v1
```

## Authentication

All API requests require proper authentication headers. Replace `{AUTH_TOKEN}` with your actual authentication token.

---

## 🏷️ Attribute Category API

### 1. Create Attribute Category

Create a new attribute category (e.g., 'Parcel Category', 'Service Type'):

```bash
curl -X POST \
  {BASE_URL}/attribute-categories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {AUTH_TOKEN}" \
  -d '{
    "code": "parcel_category",
    "name": "Parcel Category",
    "is_active": true
  }'
```

**Expected Response:**

```json
{
  "success": true,
  "message": "Attribute category created successfully",
  "data": {
    "id": "uuid-here",
    "code": "parcel_category",
    "name": "Parcel Category",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 2. Get All Attribute Categories

List all attribute categories with pagination:

```bash
curl -X GET \
  "{BASE_URL}/attribute-categories?offset=0&limit=10" \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 3. Get Attribute Category by ID

Retrieve a specific attribute category by its UUID:

```bash
curl -X GET \
  {BASE_URL}/attribute-categories/{category_id} \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 4. Get Attribute Category by Code

Retrieve a specific attribute category by its code:

```bash
curl -X GET \
  {BASE_URL}/attribute-categories/code/parcel_category \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 5. Update Attribute Category

Update an existing attribute category:

```bash
curl -X PUT \
  {BASE_URL}/attribute-categories/{category_id} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {AUTH_TOKEN}" \
  -d '{
    "name": "Updated Parcel Category",
    "is_active": false
  }'
```

### 6. Delete Attribute Category

Soft delete an attribute category:

```bash
curl -X DELETE \
  {BASE_URL}/attribute-categories/{category_id} \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 7. Get Attribute Category Statistics

Get statistics for an attribute category:

```bash
curl -X GET \
  {BASE_URL}/attribute-categories/{category_id}/stats \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

---

## 🏷️ Attribute API

**Note:** These endpoints will initially return "Service Unavailable" until the AttributeService is fully implemented.

### 1. Create Attribute

Create a new attribute within a category:

```bash
curl -X POST \
  {BASE_URL}/attributes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {AUTH_TOKEN}" \
  -d '{
    "category_id": "{category_uuid}",
    "code": "ecommerce",
    "name": "E-commerce",
    "is_active": true
  }'
```

### 2. Get All Attributes

List all attributes with pagination:

```bash
curl -X GET \
  "{BASE_URL}/attributes?offset=0&limit=10" \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 3. Get Attribute by ID

Retrieve a specific attribute:

```bash
curl -X GET \
  {BASE_URL}/attributes/{attribute_id} \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 4. Get Attribute by Code

Retrieve an attribute by its code:

```bash
curl -X GET \
  {BASE_URL}/attributes/code/ecommerce \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 5. Get Attributes by Category ID

Get all attributes within a specific category:

```bash
curl -X GET \
  {BASE_URL}/attribute-categories/{category_id}/attributes \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 6. Get Attributes by Category Code

Get all attributes within a category by category code:

```bash
curl -X GET \
  {BASE_URL}/attribute-categories/code/parcel_category/attributes \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 7. Update Attribute

Update an existing attribute:

```bash
curl -X PUT \
  {BASE_URL}/attributes/{attribute_id} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {AUTH_TOKEN}" \
  -d '{
    "name": "Updated E-commerce",
    "is_active": false
  }'
```

### 8. Delete Attribute

Soft delete an attribute:

```bash
curl -X DELETE \
  {BASE_URL}/attributes/{attribute_id} \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

---

## 🏷️ Partner Attribute Mapping API

**Note:** These endpoints will initially return "Service Unavailable" until the PartnerAttributeMapService is fully implemented.

### 1. Create Partner Attribute Mapping

Map an attribute to a partner:

```bash
curl -X POST \
  {BASE_URL}/partner-attribute-maps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {AUTH_TOKEN}" \
  -d '{
    "partner_code": "PARTNER_001",
    "attribute_id": "{attribute_uuid}",
    "attribute_code": "ecommerce",
    "is_active": true
  }'
```

### 2. Get All Partner Attribute Mappings

List all partner-attribute mappings:

```bash
curl -X GET \
  "{BASE_URL}/partner-attribute-maps?offset=0&limit=10" \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 3. Search Partner Attribute Mappings

Search mappings with filters:

```bash
curl -X POST \
  {BASE_URL}/partner-attribute-maps/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {AUTH_TOKEN}" \
  -d '{
    "partner_codes": ["PARTNER_001", "PARTNER_002"],
    "attribute_codes": ["ecommerce", "express"],
    "is_active": true,
    "pagination": {
      "offset": 0,
      "limit": 10
    }
  }'
```

### 4. Get Partner Attributes

Get all attributes mapped to a specific partner:

```bash
curl -X GET \
  {BASE_URL}/partners/PARTNER_001/attributes \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 5. Get Partners by Attribute

Get all partners that have a specific attribute:

```bash
curl -X GET \
  {BASE_URL}/attributes/code/ecommerce/partners \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 6. Get Partner Codes by Attribute

Get only partner codes for a specific attribute:

```bash
curl -X GET \
  {BASE_URL}/attributes/code/ecommerce/partner-codes \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 7. Bulk Create Partner Attribute Mappings

Create multiple mappings at once:

```bash
curl -X POST \
  {BASE_URL}/partner-attribute-maps/bulk \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {AUTH_TOKEN}" \
  -d '{
    "mappings": [
      {
        "partner_code": "PARTNER_001",
        "attribute_code": "ecommerce"
      },
      {
        "partner_code": "PARTNER_002",
        "attribute_code": "express"
      }
    ]
  }'
```

### 8. Delete Partner Attribute Mapping

Remove a partner-attribute mapping:

```bash
curl -X DELETE \
  {BASE_URL}/partner-attribute-maps/{mapping_id} \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

### 9. Bulk Delete Partner Attribute Mappings

Delete multiple mappings at once:

```bash
curl -X DELETE \
  {BASE_URL}/partner-attribute-maps/bulk \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {AUTH_TOKEN}" \
  -d '{
    "ids": ["{mapping_id_1}", "{mapping_id_2}"]
  }'
```

---

## 🔧 Testing Scenarios

### Scenario 1: Setup Parcel Category System

Create a complete parcel category attribute system:

```bash
# 1. Create parcel category
curl -X POST {BASE_URL}/attribute-categories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {AUTH_TOKEN}" \
  -d '{"code": "parcel_category", "name": "Parcel Category"}'

# 2. Create ecommerce attribute (will be unavailable initially)
curl -X POST {BASE_URL}/attributes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {AUTH_TOKEN}" \
  -d '{"category_id": "{category_id}", "code": "ecommerce", "name": "E-commerce"}'

# 3. Map partner to attribute (will be unavailable initially)
curl -X POST {BASE_URL}/partner-attribute-maps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {AUTH_TOKEN}" \
  -d '{"partner_code": "PARTNER_001", "attribute_code": "ecommerce"}'
```

### Scenario 2: Query Partners by Attribute

Filter partners by specific capabilities:

```bash
# Get all partners with ecommerce capability
curl -X GET {BASE_URL}/attributes/code/ecommerce/partner-codes \
  -H "Authorization: Bearer {AUTH_TOKEN}"

# Get detailed attributes for a specific partner
curl -X GET {BASE_URL}/partners/PARTNER_001/attributes/details \
  -H "Authorization: Bearer {AUTH_TOKEN}"
```

---

## 🚨 Error Responses

### Service Unavailable (503)

When services are not fully implemented:

```json
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Partner attribute features are temporarily unavailable - database connection required"
  }
}
```

### Validation Error (400)

Invalid request data:

```json
{
  "success": false,
  "message": "Invalid request",
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Attribute category code must be in snake_case format"
  }
}
```

### Not Found (404)

Resource not found:

```json
{
  "success": false,
  "message": "Resource not found",
  "error": {
    "code": "NOT_FOUND",
    "message": "Attribute category not found"
  }
}
```

### Conflict (409)

Resource already exists or cannot be deleted:

```json
{
  "success": false,
  "message": "Resource conflict",
  "error": {
    "code": "CONFLICT",
    "message": "Attribute category with code 'parcel_category' already exists"
  }
}
```

---

## 📝 Implementation Status

### ✅ Implemented

- **Database Models**: All three tables with soft delete support
- **Database Schema**: Complete with indexes and constraints (no migrations needed)
- **Repository Layer**: All repositories implemented
- **Attribute Category Service**: Fully functional
- **HTTP Handlers**: Complete with error handling
- **Routes Configuration**: Integrated into server
- **Validation**: Request/response validation

### 🚧 In Progress / Pending

- **Attribute Service**: Implementation pending
- **PartnerAttributeMap Service**: Implementation pending
- **Integration Testing**: Comprehensive test suite
- **API Documentation**: OpenAPI specification

### 🎯 Next Steps

1. Database schema is already set up (migration tools removed for safety)
2. Start the server: `make run`
3. Test attribute category endpoints (functional)
4. Implement remaining services for full functionality
5. Add comprehensive integration tests

---

## 📖 Related Documentation

- [Partner Attribute API Documentation](./SERVICEABILITY_API_DOCUMENTATION.md)
- [Database Schema Documentation](./db-schema.md)
- [Error Handling Guide](./error-handling.md)

Replace `{BASE_URL}` with your actual base URL and `{AUTH_TOKEN}` with your authentication token.
