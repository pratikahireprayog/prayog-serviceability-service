package orchestrators

import (
    "context"
)

// OrchestratorFactory resolves a strategy code using JourneyTemplatesClient
// and returns a strategy instance.
type OrchestratorFactory interface {
    Resolve(ctx context.Context, parcelCategory *string) (OrchestrationStrategy, error)
}

type orchestratorFactory struct {
    client         JourneyTemplatesClient
    registry       map[string]StrategyConstructor
    defaultCode    string
}

// NewOrchestratorFactory creates a new factory
func NewOrchestratorFactory(client JourneyTemplatesClient, defaultCode string, registry map[string]StrategyConstructor) OrchestratorFactory {
    return &orchestratorFactory{
        client:      client,
        registry:    registry,
        defaultCode: defaultCode,
    }
}

// Resolve returns the strategy based on the template code. Falls back to default when needed.
func (f *orchestratorFactory) Resolve(ctx context.Context, parcelCategory *string) (OrchestrationStrategy, error) {
    code := f.defaultCode
    if parcelCategory != nil && *parcelCategory != "" && f.client != nil {
        if resp, err := f.client.GetTemplates(ctx, *parcelCategory); err == nil && resp != nil && resp.Success {
            if len(resp.Data) > 0 && resp.Data[0].Code != "" {
                code = resp.Data[0].Code
            }
        }
    }

    if ctor, ok := f.registry[code]; ok {
        return ctor(), nil
    }
    if ctor, ok := f.registry[f.defaultCode]; ok {
        return ctor(), nil
    }
    return nil, nil
}


