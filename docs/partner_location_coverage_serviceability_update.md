# Partner Location Coverage Serviceability Enhancement

## Overview

This document outlines the comprehensive updates made to the Partner Location Coverage system to include detailed serviceability fields, enabling operations at the postal code level with full service capability tracking.

## Database Changes

### Migration Script

- **File**: `scripts/add_serviceability_columns_to_partner_location_coverage.sql`
- **Purpose**: Adds 12 new serviceability columns to the existing `partner_location_coverage` table

### New Columns Added

1. **Geographic Context**:

   - `country_code` (varchar(10)) - Country code for postal code location

2. **Core Serviceability Attributes**:

   - `product_type` (varchar(50)) - Product type (e.g., travel_free)
   - `parcel_category` (varchar(50)) - Parcel category (e.g., ecomm, courier, cargo)
   - `service_type` (varchar(50)) - Service type (e.g., sdd, ndd, standard)
   - `tat_days` (integer) - Turn Around Time in days for delivery

3. **Service Capabilities**:

   - `pickup` (boolean, default: false) - Whether pickup service is available
   - `delivery` (boolean, default: false) - Whether delivery service is available
   - `delivery_mode` (varchar(50)) - Delivery mode (air, surface, rail)
   - `cod_available` (boolean, default: false) - Whether Cash on Delivery is available
   - `insurance` (boolean, default: false) - Whether insurance is available

4. **Weight Constraints**:
   - `min_weight_kg` (numeric(10,3)) - Minimum weight supported in kilograms
   - `max_weight_kg` (numeric(10,3)) - Maximum weight supported in kilograms

### New Indexes

- Individual indexes on all new serviceability columns
- Composite indexes for common serviceability queries
- Enhanced performance for filtering operations

## Model Updates

### PartnerLocationCoverage Model

- **File**: `internal/shared/models/v1/location_models.go`
- **Changes**: Added all 12 new serviceability fields with appropriate GORM tags
- **Validation**: Updated business rules to accommodate new fields

### PartnerLocationCoverageFilters Model

- **Changes**: Added filtering capabilities for all new serviceability fields
- **Purpose**: Enable advanced filtering and querying based on service capabilities

## DTO Updates

### CreatePartnerLocationCoverageRequest

- **File**: `internal/shared/dtos/v1/partner_location_coverage_dto.go`
- **Changes**: Added all serviceability fields with proper validation tags
- **Validation**: Includes min/max constraints, enum validation for delivery modes

### UpdatePartnerLocationCoverageRequest

- **Changes**: Added all serviceability fields for update operations
- **Support**: Allows partial updates of serviceability attributes

### PartnerLocationCoverageResponse

- **Changes**: Includes all new serviceability fields in API responses
- **Documentation**: Complete field documentation with examples

### PartnerLocationCoverageFiltersRequest

- **Changes**: Added filtering support for all serviceability attributes
- **Features**: Enables complex queries based on service capabilities

## Service Layer Updates

### PartnerLocationCoverageService

- **File**: `internal/shared/services/v1/partner_location_coverage_service.go`

#### Create Method

- Added handling for all 12 new serviceability fields
- Proper string trimming and case normalization
- Default value assignment for boolean fields

#### Update Method

- Full support for updating serviceability attributes
- Preserves existing values when not specified in update request
- Handles null value assignments for optional fields

#### Filter Processing

- Enhanced `filterCoverages` method to support serviceability filtering
- Case-insensitive filtering for string fields
- Range-based filtering for weight constraints

#### Response Mapping

- Updated `PartnerLocationCoverageToResponse` function
- Complete mapping of all serviceability fields

### GetByPartnerID Method

- Enhanced filter mapping from DTO to model filters
- Support for all new serviceability filter parameters
- Maintained backward compatibility

## Repository Layer Updates

### PartnerLocationCoverageRepository

- **File**: `internal/shared/repositories/v1/partner_location_coverage_repository.go`

#### GetByFilters Method

- Added database-level filtering for all serviceability fields
- Optimized query performance with proper indexing
- Case-sensitive/insensitive handling as appropriate

#### Filter Support

- Country code normalization (uppercase)
- Delivery mode normalization (lowercase)
- Weight range filtering (>= for min, <= for max)
- Boolean field exact matching

## API Enhancements

### Postal Code Level Operations

- All operations now work effectively at postal code level
- Support for updating postal code data along with partner data
- Enhanced serviceability lookup capabilities

### Filter Capabilities

The API now supports filtering by:

- **Geographic**: Country code
- **Service Types**: Product type, parcel category, service type
- **Performance**: TAT days
- **Capabilities**: Pickup, delivery, COD, insurance
- **Delivery**: Delivery mode preferences
- **Weight**: Min/max weight constraints

### Validation

- Comprehensive validation for all new fields
- Enum validation for delivery modes (air, surface, rail)
- Range validation for weight and TAT constraints
- String length validation for all text fields

## Usage Examples

### Creating Coverage with Serviceability

```json
{
  "postal_code": "110001",
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
  "max_weight_kg": 50.0
}
```

### Filtering by Serviceability

```json
{
  "parcel_category": "ecomm",
  "service_type": "sdd",
  "pickup": true,
  "cod_available": true,
  "max_weight_kg": 25.0
}
```

## Migration Instructions

1. **Run Database Migration**:

   ```bash
   psql -d your_database -f scripts/add_serviceability_columns_to_partner_location_coverage.sql
   ```

2. **Update Application**:

   - Deploy updated models, DTOs, services, and repositories
   - Test API endpoints with new serviceability fields

3. **Data Population**:
   - Existing records will have null values for new fields
   - Use bulk update APIs to populate serviceability data
   - Consider setting default values for boolean fields

## Backward Compatibility

- All new fields are optional (nullable)
- Existing API calls continue to work
- Old filters and queries remain functional
- Response format is extended, not changed

## Performance Considerations

- New indexes optimize serviceability queries
- Composite indexes handle common filter combinations
- Database query performance maintained or improved
- Memory usage increase minimal due to nullable fields

## Future Enhancements

- Real-time serviceability checking
- Dynamic TAT calculation based on postal codes
- Integration with external logistics APIs
- Bulk serviceability import/export features
