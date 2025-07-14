package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"prayog-serviceability-service/internal/services/v2/orchestrators"
	"prayog-serviceability-service/test/v2/mocks"
)

// TestNewServiceabilityOrchestrator tests the constructor
func TestNewServiceabilityOrchestrator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		partnerFactory        *mocks.MockPartnerAdapterFactory
		partnerAttributeRepo  *mocks.MockPartnerAttributeMapRepository
		timeout               time.Duration
		returnOnlyServiceable bool
		expectedError         bool
		description           string
	}{
		{
			name:                  "Valid constructor with all dependencies",
			partnerFactory:        mocks.NewMockPartnerAdapterFactory(),
			partnerAttributeRepo:  mocks.NewMockPartnerAttributeMapRepository(),
			timeout:               30 * time.Second,
			returnOnlyServiceable: true,
			expectedError:         false,
			description:           "Should create orchestrator with all valid dependencies",
		},
		{
			name:                  "Valid constructor with returnOnlyServiceable false",
			partnerFactory:        mocks.NewMockPartnerAdapterFactory(),
			partnerAttributeRepo:  mocks.NewMockPartnerAttributeMapRepository(),
			timeout:               60 * time.Second,
			returnOnlyServiceable: false,
			expectedError:         false,
			description:           "Should create orchestrator with returnOnlyServiceable set to false",
		},
		{
			name:                  "Valid constructor with minimal timeout",
			partnerFactory:        mocks.NewMockPartnerAdapterFactory(),
			partnerAttributeRepo:  mocks.NewMockPartnerAttributeMapRepository(),
			timeout:               1 * time.Second,
			returnOnlyServiceable: true,
			expectedError:         false,
			description:           "Should create orchestrator with minimal timeout",
		},
		{
			name:                  "Valid constructor with zero timeout",
			partnerFactory:        mocks.NewMockPartnerAdapterFactory(),
			partnerAttributeRepo:  mocks.NewMockPartnerAttributeMapRepository(),
			timeout:               0,
			returnOnlyServiceable: true,
			expectedError:         false,
			description:           "Should create orchestrator with zero timeout (will use default)",
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			tt.partnerFactory.SetupDefaultAdapters()
			tt.partnerAttributeRepo.SetupDefaultData()

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				tt.partnerFactory,
				tt.partnerAttributeRepo,
				tt.timeout,
				tt.returnOnlyServiceable,
			)

			// Verify orchestrator was created
			assert.NotNil(t, orchestrator, "Orchestrator should not be nil")
			assert.Implements(t, (*orchestrators.ServiceabilityOrchestrator)(nil), orchestrator, "Should implement ServiceabilityOrchestrator interface")
		})
	}
}

// TestNewServiceabilityOrchestratorWithNilDependencies tests constructor with nil dependencies
func TestNewServiceabilityOrchestratorWithNilDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		partnerFactory        *mocks.MockPartnerAdapterFactory
		partnerAttributeRepo  *mocks.MockPartnerAttributeMapRepository
		timeout               time.Duration
		returnOnlyServiceable bool
		description           string
	}{
		{
			name:                  "Nil partner factory",
			partnerFactory:        nil,
			partnerAttributeRepo:  mocks.NewMockPartnerAttributeMapRepository(),
			timeout:               30 * time.Second,
			returnOnlyServiceable: true,
			description:           "Should create orchestrator even with nil partner factory",
		},
		{
			name:                  "Nil partner attribute repository",
			partnerFactory:        mocks.NewMockPartnerAdapterFactory(),
			partnerAttributeRepo:  nil,
			timeout:               30 * time.Second,
			returnOnlyServiceable: true,
			description:           "Should create orchestrator even with nil partner attribute repository",
		},
		{
			name:                  "Both dependencies nil",
			partnerFactory:        nil,
			partnerAttributeRepo:  nil,
			timeout:               30 * time.Second,
			returnOnlyServiceable: true,
			description:           "Should create orchestrator even with both dependencies nil",
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create orchestrator (should not panic)
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				tt.partnerFactory,
				tt.partnerAttributeRepo,
				tt.timeout,
				tt.returnOnlyServiceable,
			)

			// Verify orchestrator was created
			assert.NotNil(t, orchestrator, "Orchestrator should not be nil even with nil dependencies")
		})
	}
}

// TestNewServiceabilityOrchestratorTimeoutVariations tests different timeout values
func TestNewServiceabilityOrchestratorTimeoutVariations(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()
	mockFactory.SetupDefaultAdapters()
	mockRepo.SetupDefaultData()

	timeoutTests := []struct {
		name        string
		timeout     time.Duration
		description string
	}{
		{
			name:        "Nanosecond timeout",
			timeout:     1 * time.Nanosecond,
			description: "Should handle extremely small timeout",
		},
		{
			name:        "Microsecond timeout",
			timeout:     1 * time.Microsecond,
			description: "Should handle microsecond timeout",
		},
		{
			name:        "Millisecond timeout",
			timeout:     100 * time.Millisecond,
			description: "Should handle millisecond timeout",
		},
		{
			name:        "Second timeout",
			timeout:     5 * time.Second,
			description: "Should handle second timeout",
		},
		{
			name:        "Minute timeout",
			timeout:     2 * time.Minute,
			description: "Should handle minute timeout",
		},
		{
			name:        "Hour timeout",
			timeout:     1 * time.Hour,
			description: "Should handle hour timeout",
		},
	}

	for _, tt := range timeoutTests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				tt.timeout,
				true,
			)

			// Verify orchestrator was created
			require.NotNil(t, orchestrator, "Orchestrator should not be nil")
		})
	}
}

// TestNewServiceabilityOrchestratorReturnOnlyServiceableVariations tests different returnOnlyServiceable values
func TestNewServiceabilityOrchestratorReturnOnlyServiceableVariations(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()
	mockFactory.SetupDefaultAdapters()
	mockRepo.SetupDefaultData()

	returnOnlyServiceableTests := []struct {
		name                  string
		returnOnlyServiceable bool
		description           string
	}{
		{
			name:                  "Return only serviceable true",
			returnOnlyServiceable: true,
			description:           "Should create orchestrator with returnOnlyServiceable = true",
		},
		{
			name:                  "Return only serviceable false",
			returnOnlyServiceable: false,
			description:           "Should create orchestrator with returnOnlyServiceable = false",
		},
	}

	for _, tt := range returnOnlyServiceableTests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create orchestrator
			orchestrator := orchestrators.NewServiceabilityOrchestrator(
				mockFactory,
				mockRepo,
				30*time.Second,
				tt.returnOnlyServiceable,
			)

			// Verify orchestrator was created
			require.NotNil(t, orchestrator, "Orchestrator should not be nil")
		})
	}
}

// TestNewServiceabilityOrchestratorInterfaceCompatibility tests interface compatibility
func TestNewServiceabilityOrchestratorInterfaceCompatibility(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockFactory := mocks.NewMockPartnerAdapterFactory()
	mockRepo := mocks.NewMockPartnerAttributeMapRepository()
	mockFactory.SetupDefaultAdapters()
	mockRepo.SetupDefaultData()

	// Create orchestrator
	orchestrator := orchestrators.NewServiceabilityOrchestrator(
		mockFactory,
		mockRepo,
		30*time.Second,
		true,
	)

	// Verify interface compatibility
	require.NotNil(t, orchestrator)
	assert.Implements(t, (*orchestrators.ServiceabilityOrchestrator)(nil), orchestrator)

	// Verify it has the required methods (should compile if interface is implemented)
	require.NotNil(t, orchestrator.CheckServiceability)
	require.NotNil(t, orchestrator.BulkCheckServiceability)
}
