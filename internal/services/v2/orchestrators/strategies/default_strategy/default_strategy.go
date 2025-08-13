package defaultstrategy

import (
    "context"
    "prayog-serviceability-service/internal/shared/models/v1"
)

// DefaultStrategy delegates to the existing default orchestration flow.
type DefaultStrategy struct {
    ExecuteFunc func(ctx context.Context, req *models.ServiceabilityV2Request) (*models.ServiceabilityV2Response, error)
}

func (s *DefaultStrategy) Code() string { return "default" }

func (s *DefaultStrategy) Execute(ctx context.Context, req *models.ServiceabilityV2Request) (*models.ServiceabilityV2Response, error) {
    if s.ExecuteFunc == nil {
        return &models.ServiceabilityV2Response{Success: false}, nil
    }
    return s.ExecuteFunc(ctx, req)
}

// No explicit interface assertion to avoid import cycles; methods satisfy the interface.


