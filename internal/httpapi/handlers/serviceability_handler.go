package handlers

import (
	"encoding/json"
	"net/http"

	"prayog-serviceability-service/internal/httpapi/middleware"
	"prayog-serviceability-service/internal/service/usecase"

	"github.com/go-chi/chi/v5"
)

// ServiceabilityServiceInterface defines the minimal interface required for serviceability checks
type ServiceabilityServiceInterface interface {
	CheckServiceability(location, serviceType, orderType string) (bool, error)
}

// ServiceabilityHandler handles serviceability-related HTTP requests.
type ServiceabilityHandler struct {
	usecaseFactory usecase.Factory
}

// NewServiceabilityHandler creates a new serviceability handler.
func NewServiceabilityHandler(usecaseFactory usecase.Factory) *ServiceabilityHandler {
	return &ServiceabilityHandler{
		usecaseFactory: usecaseFactory,
	}
}

// RegisterRoutes registers the serviceability routes.
func (h *ServiceabilityHandler) RegisterRoutes(r chi.Router) {
	r.Get("/check/{postalCode}", h.CheckServiceability)
	r.Post("/bulk-check", middleware.RateLimiter(http.HandlerFunc(h.BulkCheckServiceability)).ServeHTTP)
}

// CheckServiceability checks if a location is serviceable.
func (h *ServiceabilityHandler) CheckServiceability(w http.ResponseWriter, r *http.Request) {
	postalCode := chi.URLParam(r, "postalCode")
	if postalCode == "" {
		http.Error(w, "Postal code is required", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual serviceability check
	isServiceable := postalCode != "00000" // Just a placeholder check

	response := map[string]interface{}{
		"postal_code":    postalCode,
		"is_serviceable": isServiceable,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BulkCheckServiceability checks if multiple locations are serviceable.
func (h *ServiceabilityHandler) BulkCheckServiceability(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PostalCodes []string `json:"postal_codes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results := make(map[string]bool)
	for _, code := range request.PostalCodes {
		// TODO: Implement actual serviceability check
		results[code] = code != "00000" // Just a placeholder check
	}

	response := map[string]interface{}{
		"results": results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
