package orchestrators

import (
    "context"
    "prayog-serviceability-service/internal/shared/models/v1"
)

// OrchestrationStrategy defines the contract for orchestration flows in V2
type OrchestrationStrategy interface {
    // Code returns the template/strategy code (e.g., "default", "smile_primary_np_extension")
    Code() string
    // Execute runs the orchestration flow and returns the V2 response
    Execute(ctx context.Context, req *models.ServiceabilityV2Request) (*models.ServiceabilityV2Response, error)
}

// StrategyConstructor is a factory function that returns a strategy instance
type StrategyConstructor func() OrchestrationStrategy


