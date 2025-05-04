package handlers

import (
	"github.com/prayog/serviceability/internal/usecase"
)

// UseCaseProvider defines the interface for accessing various use cases
type UseCaseProvider interface {
	GeoService() usecase.GeoService
	OrderTypeService() usecase.OrderTypeService
	ServiceTypeService() usecase.ServiceTypeService
	ServiceabilityService() usecase.ServiceabilityService
}

// Handler contains all HTTP handlers for the API
type Handler struct {
	usecases  UseCaseProvider
	apiPrefix string
}

// NewHandler creates a new Handler instance
func NewHandler(usecases UseCaseProvider, apiPrefix string) *Handler {
	return &Handler{
		usecases:  usecases,
		apiPrefix: apiPrefix,
	}
}
