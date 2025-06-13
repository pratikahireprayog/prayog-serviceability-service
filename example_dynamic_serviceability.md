# Dynamic Serviceability Array Examples

This file demonstrates how the serviceability API dynamically creates the response array based on actual database content.

## Database Content Examples

### Example 1: Multiple Parcel Categories in DB

**Database Records for Postal Code 385515:**

```sql
SELECT partner_id, postal_code, parcel_category, service_type, delivery_mode
FROM partner_location_coverages
WHERE postal_code = '385515' AND is_active = true;
```

**Results:**
| partner_id | postal_code | parcel_category | service_type | delivery_mode |
|------------|-------------|-----------------|--------------|---------------|
| partner-1 | 385515 | ecom | express | air |
| partner-1 | 385515 | ecom | reverse | surface |
| partner-2 | 385515 | cargo | standard | surface |
| partner-3 | 385515 | courier | express | air |

**API Response:**

```json
{
  "success": true,
  "data": {
    "destination_postal_code": "385515",
    "product_types": {
      "express": true,
      "reverse": true,
      "standard": true
    },
    "serviceability": [
      {
        "parcel_category_code": "ecomm",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "sdd",
            "delivery_modes": { "air": true, "surface": false }
          },
          {
            "service_code": "reverse",
            "delivery_modes": { "air": false, "surface": true }
          }
        ]
      },
      {
        "parcel_category_code": "cargo",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "ndd",
            "delivery_modes": { "air": false, "surface": true }
          }
        ]
      },
      {
        "parcel_category_code": "courier",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "sdd",
            "delivery_modes": { "air": true, "surface": false }
          }
        ]
      }
    ]
  }
}
```

**Result**: 3 parcel categories in DB → 3 objects in serviceability array

### Example 2: Single Parcel Category in DB

**Database Records for Postal Code 110001:**

```sql
SELECT partner_id, postal_code, parcel_category, service_type, delivery_mode
FROM partner_location_coverages
WHERE postal_code = '110001' AND is_active = true;
```

**Results:**
| partner_id | postal_code | parcel_category | service_type | delivery_mode |
|------------|-------------|-----------------|--------------|---------------|
| partner-1 | 110001 | ecom | express | air |
| partner-2 | 110001 | ecom | standard | surface |

**API Response:**

```json
{
  "success": true,
  "data": {
    "destination_postal_code": "110001",
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
            "delivery_modes": { "air": true, "surface": false }
          },
          {
            "service_code": "ndd",
            "delivery_modes": { "air": false, "surface": true }
          }
        ]
      }
    ]
  }
}
```

**Result**: 1 parcel category in DB → 1 object in serviceability array

### Example 3: Custom Parcel Categories in DB

**Database Records for Postal Code 560001:**

```sql
SELECT partner_id, postal_code, parcel_category, service_type, delivery_mode
FROM partner_location_coverages
WHERE postal_code = '560001' AND is_active = true;
```

**Results:**
| partner_id | postal_code | parcel_category | service_type | delivery_mode |
|------------|-------------|-----------------|--------------|---------------|
| partner-1 | 560001 | pharma | urgent | air |
| partner-2 | 560001 | electronics | standard | surface |

**API Response:**

```json
{
  "success": true,
  "data": {
    "destination_postal_code": "560001",
    "product_types": {
      "urgent": true,
      "standard": true
    },
    "serviceability": [
      {
        "parcel_category_code": "pharma",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "urgent",
            "delivery_modes": { "air": true, "surface": false }
          }
        ]
      },
      {
        "parcel_category_code": "electronics",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "ndd",
            "delivery_modes": { "air": false, "surface": true }
          }
        ]
      }
    ]
  }
}
```

**Result**: Custom categories from DB → Exact categories appear in response

### Example 4: No Data in DB

**Database Records for Postal Code 999999:**

```sql
SELECT partner_id, postal_code, parcel_category, service_type, delivery_mode
FROM partner_location_coverages
WHERE postal_code = '999999' AND is_active = true;
```

**Results:** (No rows)

**API Response:**

```json
{
  "success": true,
  "data": {
    "destination_postal_code": "999999",
    "product_types": {},
    "serviceability": []
  }
}
```

**Result**: No data in DB → Empty serviceability array

## Key Points

1. **Completely Database-Driven**: The serviceability array reflects exactly what's in your database
2. **No Static Categories**: Categories come directly from the `parcel_category` field
3. **Custom Categories Supported**: Any parcel category value in your DB will appear in the response
4. **Dynamic Array Size**: Array length = number of unique parcel categories found
5. **Zero Configuration**: No need to predefine or configure parcel categories
