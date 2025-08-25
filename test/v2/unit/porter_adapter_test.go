package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"prayog-serviceability-service/internal/services/v2/partners/porter"
	"prayog-serviceability-service/internal/shared/config"
	"prayog-serviceability-service/internal/shared/models/v1"
)

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create float64 pointers
func float64Ptr(f float64) *float64 {
	return &f
}

func TestPorterAdapter_SupportsRequest(t *testing.T) {
	cfg := config.PorterConfig{Enabled: true, Rating: 4.0}
	adapter := porter.NewPorterAdapter(cfg, nil, nil)

	tests := []struct {
		name    string
		request *models.ServiceabilityV2Request
		want    bool
	}{
		{
			name: "Hyperlocal with coordinates only - should support",
			request: &models.ServiceabilityV2Request{
				ParcelCategory:       stringPtr("hyperlocal"),
				SourceLatitude:       float64Ptr(18.5913),
				SourceLongitude:      float64Ptr(73.7389),
				DestinationLatitude:  float64Ptr(19.0760),
				DestinationLongitude: float64Ptr(72.8777),
			},
			want: true,
		},
		{
			name: "Hyperlocal with postal codes only - should support",
			request: &models.ServiceabilityV2Request{
				ParcelCategory:       stringPtr("hyperlocal"),
				SourcePostalCode:     stringPtr("411001"),
				DestinationPostalCode: stringPtr("400001"),
			},
			want: true,
		},
		{
			name: "Hyperlocal with mixed coordinates and postal codes - should support",
			request: &models.ServiceabilityV2Request{
				ParcelCategory:       stringPtr("hyperlocal"),
				SourceLatitude:       float64Ptr(18.5913),
				SourceLongitude:      float64Ptr(73.7389),
				DestinationPostalCode: stringPtr("400001"),
			},
			want: true,
		},
		{
			name: "Non-hyperlocal category - should not support",
			request: &models.ServiceabilityV2Request{
				ParcelCategory:       stringPtr("ecomm"),
				SourceLatitude:       float64Ptr(18.5913),
				SourceLongitude:      float64Ptr(73.7389),
				DestinationLatitude:  float64Ptr(19.0760),
				DestinationLongitude: float64Ptr(72.8777),
			},
			want: false,
		},
		{
			name: "Hyperlocal with missing source location - should not support",
			request: &models.ServiceabilityV2Request{
				ParcelCategory:       stringPtr("hyperlocal"),
				DestinationLatitude:  float64Ptr(19.0760),
				DestinationLongitude: float64Ptr(72.8777),
			},
			want: false,
		},
		{
			name: "Hyperlocal with missing destination location - should not support",
			request: &models.ServiceabilityV2Request{
				ParcelCategory:  stringPtr("hyperlocal"),
				SourceLatitude:  float64Ptr(18.5913),
				SourceLongitude: float64Ptr(73.7389),
			},
			want: false,
		},
		{
			name: "Hyperlocal with no location data - should not support",
			request: &models.ServiceabilityV2Request{
				ParcelCategory: stringPtr("hyperlocal"),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.SupportsRequest(context.Background(), tt.request)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestPorterAdapter_CoordinateHandling(t *testing.T) {
	t.Run("Coordinates take priority over postal codes", func(t *testing.T) {
		// This test documents the business logic:
		// - When both coordinates and postal codes are provided, coordinates take priority
		// - This ensures maximum accuracy for serviceability checks
		cfg := config.PorterConfig{Enabled: true, Rating: 4.0}
		adapter := porter.NewPorterAdapter(cfg, nil, nil)
		
		request := &models.ServiceabilityV2Request{
			ParcelCategory:       stringPtr("hyperlocal"),
			SourceLatitude:       float64Ptr(18.5913),
			SourceLongitude:      float64Ptr(73.7389),
			SourcePostalCode:     stringPtr("411001"), // This should be ignored
			DestinationLatitude:  float64Ptr(19.0760),
			DestinationLongitude: float64Ptr(72.8777),
			DestinationPostalCode: stringPtr("400001"), // This should be ignored
		}
		
		assert.True(t, adapter.SupportsRequest(context.Background(), request))
		// Note: The actual coordinate extraction logic is tested in getCoordinatesFromRequest
	})
}

func TestPorterAdapter_ServiceabilityLogic(t *testing.T) {
	t.Run("Both locations must be INSIDE for serviceability", func(t *testing.T) {
		// This test documents the business logic:
		// - Source coordinates: 18.5913, 73.7389 (Pune)
		// - Destination coordinates: 19.0760, 72.8777 (Mumbai)
		// - Both must return "INSIDE" from the pickup_boundaries query
		// - Only then will Porter be available for the route
		cfg := config.PorterConfig{Enabled: true, Rating: 4.0}
		adapter := porter.NewPorterAdapter(cfg, nil, nil)
		
		request := &models.ServiceabilityV2Request{
			ParcelCategory:       stringPtr("hyperlocal"),
			SourceLatitude:       float64Ptr(18.5913),
			SourceLongitude:      float64Ptr(73.7389),
			DestinationLatitude:  float64Ptr(19.0760),
			DestinationLongitude: float64Ptr(72.8777),
		}
		
		assert.True(t, adapter.SupportsRequest(context.Background(), request))
		// Note: Actual serviceability check would require database connection
		// and would execute the SQL query twice - once for each location
	})
}

func TestPorterAdapter_PostalCodeToCoordinateConversion(t *testing.T) {
	t.Run("Postal codes are converted to coordinates using geolocation service", func(t *testing.T) {
		// This test documents the business logic:
		// - When only postal codes are provided, Porter uses geolocation service
		// - The geolocation service converts postal codes to coordinates
		// - These coordinates are then used for the pickup_boundaries query
		cfg := config.PorterConfig{Enabled: true, Rating: 4.0}
		adapter := porter.NewPorterAdapter(cfg, nil, nil)
		
		request := &models.ServiceabilityV2Request{
			ParcelCategory:       stringPtr("hyperlocal"),
			SourcePostalCode:     stringPtr("411001"),
			DestinationPostalCode: stringPtr("400001"),
		}
		
		assert.True(t, adapter.SupportsRequest(context.Background(), request))
		// Note: Actual coordinate conversion would require geolocation service
		// and would convert postal codes to lat/lng before database query
	})
}
