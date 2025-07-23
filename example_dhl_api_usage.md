# DHL API Integration Documentation

## Overview

The DHL adapter has been updated to use the new DHL rates API endpoint (`/mydhlapi/test/rates`) with basic authentication. This integration provides international shipping capabilities with detailed pickup and delivery information.

## Environment Variables

The following environment variables need to be set for DHL integration:

```bash
# DHL API Configuration
DHL_BASE_URL=https://express.api.dhl.com
DHL_USERNAME=your_dhl_username
DHL_PASSWORD=your_dhl_password
DHL_ACCOUNT_NUMBER=your_account_number
DHL_ENABLED=true
DHL_TIMEOUT=30s
DHL_MAX_RETRIES=3
DHL_RETRY_DELAY=2s
DHL_SANDBOX_MODE=true
```

## API Request Structure

The DHL adapter converts internal ServiceabilityV2Request to DHL's rates API format:

```json
{
  "customerDetails": {
    "shipperDetails": {
      "postalCode": "560086",
      "cityName": "Bangalore",
      "countryCode": "IN"
    },
    "receiverDetails": {
      "postalCode": "266001",
      "cityName": "QING DAO",
      "countryCode": "CN"
    }
  },
  "accounts": [
    {
      "typeCode": "shipper",
      "number": "533748932"
    }
  ],
  "productsAndServices": [
    {
      "productCode": "P",
      "localProductCode": "P"
    }
  ],
  "payerCountryCode": "IN",
  "plannedShippingDateAndTime": "2025-05-05T13:00:00GMT+05:30",
  "unitOfMeasurement": "metric",
  "isCustomsDeclarable": true,
  "estimatedDeliveryDate": {
    "isRequested": true,
    "typeCode": "QDDC"
  },
  "returnStandardProductsOnly": true,
  "packages": [
    {
      "weight": 0.5,
      "dimensions": {
        "length": 121,
        "width": 20,
        "height": 30
      }
    }
  ]
}
```

## Expected Response Format

The DHL adapter transforms the DHL API response into the following standardized format:

```json
{
  "partner_id": "uuid",
  "partner_code": "dhl",
  "capabilities": {
    "pickup_capabilities": {
      "next_business_day": false,
      "local_cutoff_date_and_time": "2025-05-05T12:30:00",
      "pickup_earliest": "10:00:00",
      "pickup_latest": "20:30:00",
      "pickup_cutoff_same_day_outbound_processing": "14:30:00",
      "origin_service_area_code": "BLR",
      "origin_facility_area_code": "YPU",
      "pickup_additional_days": 0,
      "pickup_day_of_week": 1
    },
    "delivery_capabilities": {
      "delivery_type_code": "QDDC",
      "estimated_delivery_date_and_time": "2025-05-12T23:59:00",
      "destination_service_area_code": "TAO",
      "destination_facility_area_code": "QDN",
      "delivery_additional_days": 0,
      "delivery_day_of_week": 1,
      "total_transit_days": 7
    }
  }
}
```

## Authentication

The DHL adapter uses Basic Authentication with the following headers:

- `Authorization: Basic <base64-encoded-credentials>`
- `Content-Type: application/json`
- `Accept: application/json`
- `Message-Reference: serviceability-<timestamp>`
- `Message-Reference-Date: <RFC1123-date>`
- `x-version: 2.12.0`

## Features

### Supported Request Types

- International shipping requests (different origin and destination countries)
- Parcel categories: "international", "ecomm", "courier"
- Package weight and dimensions conversion
- Multiple currency support

### Capabilities Mapping

The adapter maps DHL's response to standard capabilities:

- **Pickup Capabilities**: Business day scheduling, cutoff times, service area codes
- **Delivery Capabilities**: Delivery type codes, estimated dates, transit times
- **Services**: Product codes, names, TAT days, pricing information

### Error Handling

- Authentication failures
- API rate limiting
- Network timeouts
- Invalid request validation
- Service not available responses

## Usage Example

```go
// Create DHL adapter
config := config.DHLConfig{
    BaseURL:       "https://express.api.dhl.com",
    Username:      os.Getenv("DHL_USERNAME"),
    Password:      os.Getenv("DHL_PASSWORD"),
    AccountNumber: os.Getenv("DHL_ACCOUNT_NUMBER"),
    Enabled:       true,
    Timeout:       30 * time.Second,
    MaxRetries:    3,
    RetryDelay:    2 * time.Second,
}

adapter := dhl.NewAdapter(config)

// Initialize the adapter
ctx := context.Background()
if err := adapter.Initialize(ctx); err != nil {
    log.Fatal("Failed to initialize DHL adapter:", err)
}

// Create serviceability request
request := &models.ServiceabilityV2Request{
    CountryCode:           stringPtr("CN"),
    SourcePostalCode:      stringPtr("560086"),
    DestinationPostalCode: stringPtr("266001"),
    Package: &models.Package{
        Weight: &models.Weight{
            Value: 0.5,
            Unit:  "kg",
        },
        Dimensions: &models.Dimensions{
            Length: 30,
            Width:  20,
            Height: 15,
            Unit:   "cm",
        },
    },
}

// Check serviceability
result, err := adapter.CheckServiceability(ctx, request)
if err != nil {
    log.Fatal("Serviceability check failed:", err)
}

// Process results
if result.IsServiceable {
    fmt.Printf("DHL services available: %d\n", len(result.Services))
    fmt.Printf("Pickup capabilities: %+v\n", result.Capabilities["pickup_capabilities"])
    fmt.Printf("Delivery capabilities: %+v\n", result.Capabilities["delivery_capabilities"])
} else {
    fmt.Printf("Service not available: %s\n", result.Metadata["reason"])
}
```

## Testing

To test the DHL integration:

1. Set the required environment variables
2. Ensure the DHL account is active and has appropriate permissions
3. Use the sandbox mode for testing (`DHL_SANDBOX_MODE=true`)
4. Test with valid international postal codes
5. Verify the response format matches the expected structure

## Error Codes

Common error scenarios:

- **Authentication Failed**: Invalid credentials
- **Invalid Request**: Missing required fields
- **Service Not Available**: Route not supported
- **Rate Limit Exceeded**: Too many requests
- **Network Error**: Connection issues

## Configuration Tips

1. **Sandbox Mode**: Use `DHL_SANDBOX_MODE=true` for testing
2. **Timeouts**: Set appropriate timeout values based on network conditions
3. **Retries**: Configure retry policy for resilient operation
4. **Account Number**: Ensure the account number matches your DHL agreement
5. **Postal Codes**: Use valid international postal codes for testing

## Monitoring

Key metrics to monitor:

- Authentication success rate
- API response times
- Error rates by type
- Service availability percentage
- Request/response payload sizes
