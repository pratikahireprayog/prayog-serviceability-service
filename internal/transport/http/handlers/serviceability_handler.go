package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prayog/serviceability/internal/transport/http/middleware"
	"github.com/prayog/serviceability/internal/usecase"
)

// ServiceabilityServiceInterface defines the minimal interface required for serviceability checks
type ServiceabilityServiceInterface interface {
	CheckServiceability(location, serviceType, orderType string) (bool, error)
}

// ServiceabilityHandler handles HTTP requests related to serviceability
type ServiceabilityHandler struct {
	service usecase.ServiceabilityService
}

// NewServiceabilityHandler creates a new serviceability handler
func NewServiceabilityHandler(service usecase.ServiceabilityService) *ServiceabilityHandler {
	return &ServiceabilityHandler{
		service: service,
	}
}

// CheckServiceability checks if a location is serviceable
func (h *ServiceabilityHandler) CheckServiceability(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	location := r.URL.Query().Get("location")
	serviceType := r.URL.Query().Get("service_type")
	orderType := r.URL.Query().Get("order_type")

	if location == "" || serviceType == "" {
		middleware.RespondWithValidationError(w, "Missing required parameters: location and service_type")
		return
	}

	// Call the service
	result, err := h.service.CheckServiceability(location, serviceType, orderType)
	if err != nil {
		middleware.RespondWithInternalError(w, "Failed to check serviceability: "+err.Error())
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, map[string]bool{
		"isServiceable": result,
	})
}

// RegisterRoutes registers all serviceability routes
func (h *ServiceabilityHandler) RegisterRoutes(r chi.Router) {
	r.Get("/check", h.CheckServiceability)
}
