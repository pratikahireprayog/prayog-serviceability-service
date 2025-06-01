# Specification Service Enhancements for Serviceability API

## Overview

This document outlines the required enhancements to the Specification Service to support the comprehensive Serviceability API. The Serviceability API needs various catalog definitions and entity specifications to determine service availability across different partners and locations.

**Note**: Partner to location mapping is handled by the `partner_location_coverage` table in the Serviceability Service schema, not in the Specification Service, to avoid table overlap and maintain clear service boundaries.

## Required Catalog Definitions

### 1. PARCEL_CATEGORY Catalog

**Purpose**: Define the different types of parcel categories that can be shipped.

**Spec Definition**:

- **Name**: `PARCEL_CATEGORY`
- **Description**: "Catalog of supported parcel categories for logistics services"
- **Type**: Catalog

**Catalog Items**:

```json
[
  {
    "code": "ecom",
    "name": "E-commerce",
    "description": "E-commerce packages and shipments"
  },
  {
    "code": "courier",
    "name": "Courier",
    "description": "Express courier and document delivery"
  },
  {
    "code": "cargo",
    "name": "Cargo",
    "description": "Heavy cargo and freight shipments"
  }
]
```

### 2. SERVICE_TYPE Catalog

**Purpose**: Define the different service types available for shipments.

**Spec Definition**:

- **Name**: `SERVICE_TYPE`
- **Description**: "Catalog of logistics service types with delivery timeframes"
- **Type**: Catalog

**Catalog Items**:

```json
[
  {
    "code": "SDD",
    "name": "Same Day Delivery",
    "description": "Delivery within the same day"
  },
  {
    "code": "NDD",
    "name": "Next Day Delivery",
    "description": "Delivery by next business day"
  },
  {
    "code": "Standard",
    "name": "Standard Delivery",
    "description": "Standard delivery timeframe (2-5 days)"
  },
  {
    "code": "Express",
    "name": "Express Delivery",
    "description": "Express delivery (1-2 days)"
  },
  {
    "code": "vayuquick",
    "name": "VayuQuick",
    "description": "VayuQuick premium service"
  },
  {
    "code": "vayuquickpro",
    "name": "VayuQuick Pro",
    "description": "VayuQuick professional service"
  }
]
```

### 3. OPERATION_TYPE Catalog

**Purpose**: Define the types of operations supported (pickup, delivery, etc.).

**Spec Definition**:

- **Name**: `OPERATION_TYPE`
- **Description**: "Catalog of logistics operation types"
- **Type**: Catalog

**Catalog Items**:

```json
[
  {
    "code": "pickup",
    "name": "Pickup",
    "description": "Pickup from origin location"
  },
  {
    "code": "delivery",
    "name": "Delivery",
    "description": "Delivery to destination location"
  }
]
```

### 4. PAYMENT_MODE Catalog

**Purpose**: Define the payment methods supported for shipments.

**Spec Definition**:

- **Name**: `PAYMENT_MODE`
- **Description**: "Catalog of supported payment modes"
- **Type**: Catalog

**Catalog Items**:

```json
[
  {
    "code": "ONLINE",
    "name": "Online Payment",
    "description": "Pre-paid online transactions"
  },
  {
    "code": "COD",
    "name": "Cash on Delivery",
    "description": "Payment collected on delivery"
  }
]
```

### 5. DELIVERY_MODE Catalog

**Purpose**: Define the transportation modes used for delivery.

**Spec Definition**:

- **Name**: `DELIVERY_MODE`
- **Description**: "Catalog of transportation modes for delivery"
- **Type**: Catalog

**Catalog Items**:

```json
[
  {
    "code": "AIR",
    "name": "Air Transport",
    "description": "Air freight and express air delivery"
  },
  {
    "code": "SURFACE",
    "name": "Surface Transport",
    "description": "Road transport via trucks and vehicles"
  },
  {
    "code": "RAIL",
    "name": "Rail Transport",
    "description": "Railway freight and cargo"
  }
]
```

## Required Entity Type Specifications

### 1. PARTNER Entity Type

**Purpose**: Define specifications for partner service capabilities and preferences.

#### A. PARTNER_SERVICE_CAPABILITY Specification

**Spec Definition**:

- **Name**: `PARTNER_SERVICE_CAPABILITY`
- **Description**: "Defines service capabilities for logistics partners"
- **Entity Type**: `PARTNER`

**Specification Schema**:

```json
{
  "partner_id": "string",
  "service_capabilities": [
    {
      "service_type": "string (from SERVICE_TYPE catalog)",
      "parcel_category": "string (from PARCEL_CATEGORY catalog)",
      "operation_types": ["string (from OPERATION_TYPE catalog)"],
      "payment_modes": ["string (from PAYMENT_MODE catalog)"],
      "delivery_modes": ["string (from DELIVERY_MODE catalog)"],
      "is_active": "boolean",
      "effective_rating": "number",
      "capacity_limit": "number",
      "cost_multiplier": "number"
    }
  ]
}
```

**Example Specification Values**:

```json
{
  "partner_id": "1",
  "service_capabilities": [
    {
      "service_type": "SDD",
      "parcel_category": "ecom",
      "operation_types": ["pickup", "delivery"],
      "payment_modes": ["ONLINE", "COD"],
      "delivery_modes": ["AIR"],
      "is_active": true,
      "effective_rating": 4.5,
      "capacity_limit": 100,
      "cost_multiplier": 1.2
    }
  ]
}
```

#### B. PARTNER_SERVICE_PREFERENCES Specification

**Spec Definition**:

- **Name**: `PARTNER_SERVICE_PREFERENCES`
- **Description**: "Defines partner preferences and priorities for services"
- **Entity Type**: `PARTNER`

**Specification Schema**:

```json
{
  "partner_id": "string",
  "preferences": [
    {
      "service_type": "string (from SERVICE_TYPE catalog)",
      "parcel_category": "string (from PARCEL_CATEGORY catalog)",
      "is_preferred": "boolean",
      "priority_score": "number",
      "effective_rating_modifier": "number"
    }
  ]
}
```

### 2. LOCATION Entity Type

**Purpose**: Define location-specific service availability independently of partners.

**Note**: Partner location coverage is handled via the `partner_location_coverage` table in the Serviceability Service, which maps partners to locations at different scopes (postal_code, area, city, region, country) with zone types (PRIMARY, SECONDARY, BUFFER).

#### A. LOCATION_SERVICE_AVAILABILITY Specification

**Spec Definition**:

- **Name**: `LOCATION_SERVICE_AVAILABILITY`
- **Description**: "Defines service availability for specific locations"
- **Entity Type**: `LOCATION`

**Specification Schema**:

```json
{
  "location_id": "string",
  "location_type": "string (postal_code|area|city|region)",
  "available_services": [
    {
      "service_type": "string (from SERVICE_TYPE catalog)",
      "parcel_categories": ["string (from PARCEL_CATEGORY catalog)"],
      "operation_types": ["string (from OPERATION_TYPE catalog)"],
      "payment_modes": ["string (from PAYMENT_MODE catalog)"],
      "delivery_modes": ["string (from DELIVERY_MODE catalog)"],
      "is_active": "boolean",
      "restrictions": {
        "max_weight": "number",
        "max_dimensions": "object",
        "blocked_item_types": ["string"]
      }
    }
  ]
}
```

#### B. LOCATION_SERVICE_RESTRICTIONS Specification

**Spec Definition**:

- **Name**: `LOCATION_SERVICE_RESTRICTIONS`
- **Description**: "Defines restrictions for services in specific locations"
- **Entity Type**: `LOCATION`

**Specification Schema**:

```json
{
  "location_id": "string",
  "location_type": "string",
  "restrictions": [
    {
      "service_type": "string",
      "parcel_category": "string",
      "restriction_type": "string (weight|dimension|item_type|time)",
      "restriction_value": "any",
      "is_active": "boolean"
    }
  ]
}
```

## Integration Patterns

### 1. Serviceability Calculation Flow

1. **Fetch Catalogs**: Retrieve all required catalogs (PARCEL_CATEGORY, SERVICE_TYPE, etc.)
2. **Get Partner Coverage**: Query `partner_location_coverage` table in Serviceability Service for partner-location mappings
3. **Get Partner Capabilities**: Query PARTNER entity specifications for service capabilities
4. **Get Location Availability**: Query LOCATION entity specifications for location-specific rules
5. **Apply Business Logic**: Combine partner capabilities with location availability and coverage zones
6. **Filter by Preferences**: Apply partner preferences and ratings
7. **Format Response**: Structure the final serviceability response

### 2. Partner Location Mapping Flow

**In Serviceability Service (not Specification Service)**:

- Use `partner_location_coverage` table to determine which partners serve which locations
- Support different location scopes: POSTAL_CODE, AREA, CITY, REGION, COUNTRY
- Support zone types: PRIMARY, SECONDARY, BUFFER for priority-based selection

### 3. Caching Strategy

- **Catalog Data**: Cache for 2-4 hours (relatively static)
- **Partner Capabilities**: Cache for 30-60 minutes (can change frequently)
- **Partner Location Coverage**: Cache for 1-2 hours (changes occasionally)
- **Location Availability**: Cache for 1-2 hours (changes occasionally)

### 4. API Integration Points

#### Required Specification Service APIs:

```http
GET /api/v1/catalogs?spec_type=PARCEL_CATEGORY
GET /api/v1/catalogs?spec_type=SERVICE_TYPE
GET /api/v1/catalogs?spec_type=OPERATION_TYPE
GET /api/v1/catalogs?spec_type=PAYMENT_MODE
GET /api/v1/catalogs?spec_type=DELIVERY_MODE

GET /api/v1/entity-specifications?entity_type=PARTNER&entity_id={partner_id}&spec_name=PARTNER_SERVICE_CAPABILITY
GET /api/v1/entity-specifications?entity_type=PARTNER&entity_id={partner_id}&spec_name=PARTNER_SERVICE_PREFERENCES
GET /api/v1/entity-specifications?entity_type=LOCATION&entity_id={location_id}&spec_name=LOCATION_SERVICE_AVAILABILITY
```

#### Required Serviceability Service APIs (for partner coverage):

```http
GET /api/v1/partner-coverage?location_scope=POSTAL_CODE&location_id={postal_code_id}
GET /api/v1/partner-coverage?partner_id={partner_id}&location_scope=CITY
```

## Implementation Priority

### Phase 1: Core Catalogs

1. PARCEL_CATEGORY
2. SERVICE_TYPE
3. OPERATION_TYPE
4. PAYMENT_MODE
5. DELIVERY_MODE

### Phase 2: Partner Specifications

1. PARTNER_SERVICE_CAPABILITY
2. PARTNER_SERVICE_PREFERENCES (simplified without location coverage)

### Phase 3: Advanced Features

1. LOCATION_SERVICE_AVAILABILITY
2. LOCATION_SERVICE_RESTRICTIONS

**Note**: Partner location coverage is implemented via the `partner_location_coverage` table in the Serviceability Service schema.

## Validation Rules

### Catalog Validation

- All catalog codes must be unique within their catalog
- Codes should be alphanumeric and follow naming conventions
- Names and descriptions are required

### Entity Specification Validation

- Partner IDs must exist in Partner Service
- Location IDs must exist in Location hierarchy
- All referenced catalog values must exist in respective catalogs
- Numeric values (ratings, priorities) must be within valid ranges

## Migration Strategy

1. **Create Spec Definitions**: Add all required specification definitions
2. **Create Catalogs**: Populate all catalog data
3. **Test Integration**: Verify Serviceability API can fetch data
4. **Migrate Partner Data**: Convert existing partner capabilities to new format
5. **Add Location Rules**: Implement location-specific availability rules
6. **Performance Testing**: Ensure caching and performance requirements are met

## Monitoring and Maintenance

### Metrics to Track

- API response times for catalog fetches
- Cache hit rates for specification data
- Data consistency between services
- Business rule coverage and effectiveness

### Regular Maintenance

- Review and update catalog entries
- Monitor partner capability changes
- Validate location availability rules
- Performance optimization based on usage patterns

This design provides a comprehensive foundation for the Specification Service enhancements needed to support the Serviceability API's requirements for flexible, partner-aware, and location-specific service determination.
