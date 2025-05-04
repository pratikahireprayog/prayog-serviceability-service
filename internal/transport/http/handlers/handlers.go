package handlers

import (
	"net/http"

	"github.com/prayog/serviceability/internal/usecase"
)

// Factory creates and provides HTTP handlers
type Factory struct {
	useCases *usecase.UseCaseFactory
	version  string
}

// NewFactory creates a new handler factory
func NewFactory(useCases *usecase.UseCaseFactory, version string) *Factory {
	return &Factory{
		useCases: useCases,
		version:  version,
	}
}

// HealthHandler returns a health check handler
func (f *Factory) HealthHandler() http.HandlerFunc {
	return HealthHandler(f.version)
}

// ServiceabilityHandler returns a serviceability handler
func (f *Factory) ServiceabilityHandler() *ServiceabilityHandler {
	// The usecase.ServiceabilityService interface already matches what we need
	return NewServiceabilityHandler(f.useCases.ServiceabilityService())
}
