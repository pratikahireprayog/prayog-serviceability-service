# Serviceability Service Database Schema

This document describes the database schema used in the Serviceability Service.

## Entity-Relationship Diagram

The database consists of the following main entity groups:
- Location entities (countries, regions, cities, areas, postal codes)
- Location aliases (translations/alternate names)
- Service entities (order types, service types, service availability)

## Tables

### Location Entities

#### countries
| Column Name   | Type         | Constraints       | Description                   |
|---------------|--------------|-------------------|-------------------------------|
| id            | uuid         | PK                | Unique identifier             |
| name          | varchar(100) | NOT NULL          | Country name                  |
| code          | varchar(10)  | NOT NULL, UNIQUE  | ISO country code (e.g., "US") |
| currency_code | varchar(3)   | -                 | ISO currency code (e.g., "USD") |
| phone_code    | varchar(10)  | -                 | International calling code    |
| is_active     | boolean      | NOT NULL, DEFAULT true | Status flag              |
| created_at    | timestamp    | NOT NULL          | Creation timestamp            |
| updated_at    | timestamp    | NOT NULL          | Last update timestamp         |

#### administrative_regions
| Column Name   | Type         | Constraints       | Description                   |
|---------------|--------------|-------------------|-------------------------------|
| id            | uuid         | PK                | Unique identifier             |
| country_id    | uuid         | FK, NOT NULL      | Reference to countries        |
| name          | varchar(100) | NOT NULL          | Region name                   |
| code          | varchar(20)  | NOT NULL          | Region code                   |
| region_type   | varchar(20)  | -                 | Type (STATE, PROVINCE, etc.)  |
| is_active     | boolean      | NOT NULL, DEFAULT true | Status flag              |
| created_at    | timestamp    | NOT NULL          | Creation timestamp            |
| updated_at    | timestamp    | NOT NULL          | Last update timestamp         |

#### cities
| Column Name   | Type         | Constraints       | Description                   |
|---------------|--------------|-------------------|-------------------------------|
| id            | uuid         | PK                | Unique identifier             |
| country_id    | uuid         | FK                | Reference to countries        |
| region_id     | uuid         | FK                | Reference to administrative_regions |
| name          | varchar(100) | NOT NULL          | City name                     |
| code          | varchar(20)  | -                 | City code                     |
| is_active     | boolean      | NOT NULL, DEFAULT true | Status flag              |
| created_at    | timestamp    | NOT NULL          | Creation timestamp            |
| updated_at    | timestamp    | NOT NULL          | Last update timestamp         |

#### areas
| Column Name   | Type         | Constraints       | Description                   |
|---------------|--------------|-------------------|-------------------------------|
| id            | uuid         | PK                | Unique identifier             |
| city_id       | uuid         | FK, NOT NULL      | Reference to cities           |
| name          | varchar(100) | NOT NULL          | Area name                     |
| area_type     | varchar(20)  | -                 | Type (NEIGHBORHOOD, etc.)     |
| geo_location  | jsonb        | -                 | Geographical coordinates      |
| is_active     | boolean      | NOT NULL, DEFAULT true | Status flag              |
| created_at    | timestamp    | NOT NULL          | Creation timestamp            |
| updated_at    | timestamp    | NOT NULL          | Last update timestamp         |

#### postal_codes
| Column Name   | Type         | Constraints       | Description                   |
|---------------|--------------|-------------------|-------------------------------|
| id            | uuid         | PK                | Unique identifier             |
| country_id    | uuid         | FK, NOT NULL      | Reference to countries        |
| region_id     | uuid         | FK                | Reference to administrative_regions |
| city_id       | uuid         | FK                | Reference to cities           |
| area_id       | uuid         | FK                | Reference to areas            |
| code          | varchar(20)  | NOT NULL          | Postal/zip code               |
| geo_location  | jsonb        | -                 | Geographical coordinates      |
| is_active     | boolean      | NOT NULL, DEFAULT true | Status flag              |
| created_at    | timestamp    | NOT NULL          | Creation timestamp            |
| updated_at    | timestamp    | NOT NULL          | Last update timestamp         |

### Location Aliases

#### country_aliases
| Column Name   | Type         | Constraints       | Description                   |
|---------------|--------------|-------------------|-------------------------------|
| id            | uuid         | PK                | Unique identifier             |
| country_id    | uuid         | FK, NOT NULL      | Reference to countries        |
| alias_name    | varchar(100) | NOT NULL          | Alternative country name      |
| language_code | varchar(10)  | NOT NULL          | ISO language code (e.g., "en") |
| is_primary    | boolean      | NOT NULL, DEFAULT false | Primary translation flag |
| is_active     | boolean      | NOT NULL, DEFAULT true | Status flag               |
| created_at    | timestamp    | NOT NULL          | Creation timestamp             |
| updated_at    | timestamp    | NOT NULL          | Last update timestamp          |

#### administrative_region_aliases
| Column Name   | Type         | Constraints       | Description                   |
|---------------|--------------|-------------------|-------------------------------|
| id            | uuid         | PK                | Unique identifier             |
| region_id     | uuid         | FK, NOT NULL      | Reference to administrative_regions |
| alias_name    | varchar(100) | NOT NULL          | Alternative region name       |
| language_code | varchar(10)  | NOT NULL          | ISO language code             |
| is_primary    | boolean      | NOT NULL, DEFAULT false | Primary translation flag |
| is_active     | boolean      | NOT NULL, DEFAULT true | Status flag               |
| created_at    | timestamp    | NOT NULL          | Creation timestamp             |
| updated_at    | timestamp    | NOT NULL          | Last update timestamp          |

#### city_aliases
| Column Name   | Type         | Constraints       | Description                   |
|---------------|--------------|-------------------|-------------------------------|
| id            | uuid         | PK                | Unique identifier             |
| city_id       | uuid         | FK, NOT NULL      | Reference to cities           |
| alias_name    | varchar(100) | NOT NULL          | Alternative city name         |
| language_code | varchar(10)  | NOT NULL          | ISO language code             |
| is_primary    | boolean      | NOT NULL, DEFAULT false | Primary translation flag |
| is_active     | boolean      | NOT NULL, DEFAULT true | Status flag               |
| created_at    | timestamp    | NOT NULL          | Creation timestamp             |
| updated_at    | timestamp    | NOT NULL          | Last update timestamp          |

#### area_aliases
| Column Name   | Type         | Constraints       | Description                   |
|---------------|--------------|-------------------|-------------------------------|
| id            | uuid         | PK                | Unique identifier             |
| area_id       | uuid         | FK, NOT NULL      | Reference to areas            |
| alias_name    | varchar(100) | NOT NULL          | Alternative area name         |
| language_code | varchar(10)  | NOT NULL          | ISO language code             |
| is_primary    | boolean      | NOT NULL, DEFAULT false | Primary translation flag |
| is_active     | boolean      | NOT NULL, DEFAULT true | Status flag               |
| created_at    | timestamp    | NOT NULL          | Creation timestamp             |
| updated_at    | timestamp    | NOT NULL          | Last update timestamp          |

### Service Entities

#### order_types
| Column Name   | Type         | Constraints       | Description                   |
|---------------|--------------|-------------------|-------------------------------|
| id            | uuid         | PK                | Unique identifier             |
| name          | varchar(100) | NOT NULL          | Order type name               |
| code          | varchar(20)  | NOT NULL, UNIQUE  | Order type code               |
| description   | varchar(500) | -                 | Order type description        |
| is_active     | boolean      | NOT NULL, DEFAULT true | Status flag              |
| created_at    | timestamp    | NOT NULL          | Creation timestamp            |
| updated_at    | timestamp    | NOT NULL          | Last update timestamp         |

#### service_types
| Column Name   | Type         | Constraints       | Description                   |
|---------------|--------------|-------------------|-------------------------------|
| id            | uuid         | PK                | Unique identifier             |
| name          | varchar(100) | NOT NULL          | Service type name             |
| code          | varchar(20)  | NOT NULL, UNIQUE  | Service type code             |
| sla_hours     | integer      | NOT NULL          | Service level agreement hours |
| description   | varchar(500) | -                 | Service type description      |
| is_active     | boolean      | NOT NULL, DEFAULT true | Status flag              |
| created_at    | timestamp    | NOT NULL          | Creation timestamp            |
| updated_at    | timestamp    | NOT NULL          | Last update timestamp         |

#### service_availabilities
| Column Name     | Type         | Constraints       | Description                   |
|-----------------|--------------|-------------------|-------------------------------|
| id              | uuid         | PK                | Unique identifier             |
| location_type   | varchar(20)  | NOT NULL          | Type (COUNTRY, CITY, etc.)    |
| location_id     | uuid         | NOT NULL          | ID of location entity         |
| order_type_id   | uuid         | FK, NOT NULL      | Reference to order_types      |
| service_type_id | uuid         | FK, NOT NULL      | Reference to service_types    |
| is_available    | boolean      | NOT NULL, DEFAULT false | Availability flag       |
| constraints     | jsonb        | -                 | Configuration parameters      |
| effective_from  | timestamp    | NOT NULL          | Start date of validity        |
| effective_to    | timestamp    | NOT NULL          | End date of validity          |
| is_active       | boolean      | NOT NULL, DEFAULT true | Status flag              |
| created_at      | timestamp    | NOT NULL          | Creation timestamp            |
| updated_at      | timestamp    | NOT NULL          | Last update timestamp         |

## Indexes

### Location Entities
- `countries`: `code` (UNIQUE), `currency_code`, `phone_code`, `is_active`
- `administrative_regions`: `country_id`, `code`, `region_type`, `is_active`
- `cities`: `country_id`, `region_id`, `code`, `is_active`
- `areas`: `city_id`, `area_type`, `is_active`
- `postal_codes`: `country_id`, `region_id`, `city_id`, `area_id`, `code`, `is_active`

### Location Aliases
- `country_aliases`: `country_id`, `language_code`, `is_primary`, `is_active`
- `administrative_region_aliases`: `region_id`, `language_code`, `is_primary`, `is_active`
- `city_aliases`: `city_id`, `language_code`, `is_primary`, `is_active`
- `area_aliases`: `area_id`, `language_code`, `is_primary`, `is_active`

### Service Entities
- `order_types`: `code` (UNIQUE), `is_active`
- `service_types`: `code` (UNIQUE), `is_active`
- `service_availabilities`: `location_type`, `location_id`, `order_type_id`, `service_type_id`, `is_available`, `effective_from`, `effective_to`, `is_active`
- `service_availabilities`: Compound index `(location_type, location_id, order_type_id, service_type_id)` for efficient lookups

## Foreign Key Constraints

- `administrative_regions.country_id` → `countries.id` (CASCADE)
- `cities.country_id` → `countries.id` (CASCADE)
- `cities.region_id` → `administrative_regions.id` (CASCADE)
- `areas.city_id` → `cities.id` (CASCADE)
- `postal_codes.country_id` → `countries.id` (CASCADE)
- `postal_codes.region_id` → `administrative_regions.id` (SET NULL)
- `postal_codes.city_id` → `cities.id` (SET NULL)
- `postal_codes.area_id` → `areas.id` (SET NULL)
- `country_aliases.country_id` → `countries.id` (CASCADE)
- `administrative_region_aliases.region_id` → `administrative_regions.id` (CASCADE)
- `city_aliases.city_id` → `cities.id` (CASCADE)
- `area_aliases.area_id` → `areas.id` (CASCADE)
- `service_availabilities.order_type_id` → `order_types.id` (CASCADE)
- `service_availabilities.service_type_id` → `service_types.id` (CASCADE)

## Data Types

### Custom Types
- `Point`: JSON structure for geographic coordinates
  ```json
  {
    "latitude": 19.1136,
    "longitude": 72.8697
  }
  ```

- `Constraints`: JSON structure for service availability constraints
  ```json
  {
    "min_order_value": 500,
    "max_weight_kg": 30,
    "delivery_hours": "09:00-19:00",
    "cutoff_time": "14:00"
  }
  ```

## Notes

1. All tables use UUIDs as primary keys with automatic generation on insert.
2. All tables include `created_at` and `updated_at` timestamps that are automatically managed.
3. All tables include an `is_active` flag for soft deletion/deactivation.
4. Alias tables enforce uniqueness of primary translations per language by using partial unique indexes.
5. The schema supports hierarchical location data (country → region → city → area → postal code) with flexibility. 