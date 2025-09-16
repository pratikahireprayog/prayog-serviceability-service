package public_orchestrator

import (
	"context"

	orchestrators "prayog-serviceability-service/internal/services/v2/orchestrators"
	models "prayog-serviceability-service/internal/shared/models/v1"
)

// PublicServiceabilityOrchestrator provides a simplified boolean check for serviceability
// by delegating to the full V2 orchestrator and mapping its response to a boolean.
type PublicServiceabilityOrchestrator struct {
	base orchestrators.ServiceabilityOrchestrator
}

// NewPublicServiceabilityOrchestrator creates a new wrapper instance.
func NewPublicServiceabilityOrchestrator(base orchestrators.ServiceabilityOrchestrator) *PublicServiceabilityOrchestrator {
	return &PublicServiceabilityOrchestrator{base: base}
}

// CheckIsServiceable returns true if the underlying V2 orchestrator finds any serviceable partner.
func (p *PublicServiceabilityOrchestrator) CheckIsServiceable(ctx context.Context, req *models.ServiceabilityV2Request) (bool, error) {
	if p == nil || p.base == nil {
		return false, nil
	}
	resp, err := p.base.CheckServiceability(ctx, req)
	if err != nil {
		return false, err
	}
	return resp != nil && resp.Success, nil
}
