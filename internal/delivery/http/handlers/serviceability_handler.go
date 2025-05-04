package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/prayog/serviceability/internal/domain"
)

// RegisterServiceabilityRoutes registers all serviceability-related routes
func (h *Handler) RegisterServiceabilityRoutes(r chi.Router) {
	// Service availability management
	r.Route("/service-availability", func(r chi.Router) {
		r.Get("/", h.ListServiceAvailabilities)
		r.Post("/", h.CreateServiceAvailability)
		r.Get("/{id}", h.GetServiceAvailability)
		r.Put("/{id}", h.UpdateServiceAvailability)
		r.Delete("/{id}", h.DeleteServiceAvailability)
		r.Get("/location/{type}/{id}", h.GetServiceAvailabilitiesByLocation)
		r.Get("/service/{serviceTypeId}", h.GetServiceAvailabilitiesByService)
		r.Get("/order-type/{orderTypeId}", h.GetServiceAvailabilitiesByOrderType)
	})

	// Serviceability check endpoints
	r.Route("/check", func(r chi.Router) {
		r.Post("/serviceability", h.CheckServiceability)
		r.Post("/serviceability/bulk", h.BulkCheckServiceability)
	})
}

// ListServiceAvailabilities handles GET /service-availability
func (h *Handler) ListServiceAvailabilities(w http.ResponseWriter, r *http.Request) {
	availabilities, err := h.usecases.ServiceabilityService().ListServiceAvailabilities()
	if err != nil {
		RespondWithInternalError(w, "Failed to get service availabilities: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, availabilities)
}

// GetServiceAvailability handles GET /service-availability/{id}
func (h *Handler) GetServiceAvailability(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid service availability ID")
		return
	}

	availability, err := h.usecases.ServiceabilityService().GetServiceAvailabilityByID(uint(id))
	if err != nil {
		RespondWithNotFound(w, "Service availability not found")
		return
	}

	RespondWithJSON(w, http.StatusOK, availability)
}

// GetServiceAvailabilitiesByLocation handles GET /service-availability/location/{type}/{id}
func (h *Handler) GetServiceAvailabilitiesByLocation(w http.ResponseWriter, r *http.Request) {
	locationType := chi.URLParam(r, "type")
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid location ID")
		return
	}

	if locationType == "" || locationType != "country" && locationType != "region" &&
		locationType != "city" && locationType != "area" && locationType != "postal_code" {
		RespondWithValidationError(w, "Invalid location type. Must be one of: country, region, city, area, postal_code")
		return
	}

	availabilities, err := h.usecases.ServiceabilityService().GetServiceAvailabilitiesByLocation(locationType, uint(id))
	if err != nil {
		RespondWithInternalError(w, "Failed to get service availabilities: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, availabilities)
}

// GetServiceAvailabilitiesByService handles GET /service-availability/service/{serviceTypeId}
func (h *Handler) GetServiceAvailabilitiesByService(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "serviceTypeId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid service type ID")
		return
	}

	availabilities, err := h.usecases.ServiceabilityService().GetServiceAvailabilitiesByService(uint(id))
	if err != nil {
		RespondWithInternalError(w, "Failed to get service availabilities: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, availabilities)
}

// GetServiceAvailabilitiesByOrderType handles GET /service-availability/order-type/{orderTypeId}
func (h *Handler) GetServiceAvailabilitiesByOrderType(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "orderTypeId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid order type ID")
		return
	}

	availabilities, err := h.usecases.ServiceabilityService().GetServiceAvailabilitiesByOrderType(uint(id))
	if err != nil {
		RespondWithInternalError(w, "Failed to get service availabilities: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, availabilities)
}

// CreateServiceAvailability handles POST /service-availability
func (h *Handler) CreateServiceAvailability(w http.ResponseWriter, r *http.Request) {
	var availability domain.ServiceAvailability
	if err := json.NewDecoder(r.Body).Decode(&availability); err != nil {
		RespondWithValidationError(w, "Invalid request body")
		return
	}

	// Validate location type
	if availability.LocationType != "country" && availability.LocationType != "region" &&
		availability.LocationType != "city" && availability.LocationType != "area" &&
		availability.LocationType != "postal_code" {
		RespondWithValidationError(w, "Invalid location type. Must be one of: country, region, city, area, postal_code")
		return
	}

	if err := h.usecases.ServiceabilityService().CreateServiceAvailability(&availability); err != nil {
		RespondWithInternalError(w, "Failed to create service availability: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusCreated, availability)
}

// UpdateServiceAvailability handles PUT /service-availability/{id}
func (h *Handler) UpdateServiceAvailability(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid service availability ID")
		return
	}

	var availability domain.ServiceAvailability
	if err := json.NewDecoder(r.Body).Decode(&availability); err != nil {
		RespondWithValidationError(w, "Invalid request body")
		return
	}

	// Validate location type
	if availability.LocationType != "country" && availability.LocationType != "region" &&
		availability.LocationType != "city" && availability.LocationType != "area" &&
		availability.LocationType != "postal_code" {
		RespondWithValidationError(w, "Invalid location type. Must be one of: country, region, city, area, postal_code")
		return
	}

	availability.ID = uint(id)
	if err := h.usecases.ServiceabilityService().UpdateServiceAvailability(&availability); err != nil {
		RespondWithInternalError(w, "Failed to update service availability: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, availability)
}

// DeleteServiceAvailability handles DELETE /service-availability/{id}
func (h *Handler) DeleteServiceAvailability(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid service availability ID")
		return
	}

	if err := h.usecases.ServiceabilityService().DeleteServiceAvailability(uint(id)); err != nil {
		RespondWithInternalError(w, "Failed to delete service availability: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Service availability deleted successfully"})
}

// CheckServiceability handles POST /check/serviceability
func (h *Handler) CheckServiceability(w http.ResponseWriter, r *http.Request) {
	var request domain.ServiceabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		RespondWithValidationError(w, "Invalid request body")
		return
	}

	if request.PostalCode == "" {
		RespondWithValidationError(w, "Postal code is required")
		return
	}

	if request.OrderTypeCode == "" {
		RespondWithValidationError(w, "Order type code is required")
		return
	}

	result, err := h.usecases.ServiceabilityService().CheckServiceability(request)
	if err != nil {
		RespondWithInternalError(w, "Failed to check serviceability: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, result)
}

// BulkCheckServiceability handles POST /check/serviceability/bulk
func (h *Handler) BulkCheckServiceability(w http.ResponseWriter, r *http.Request) {
	var requests []domain.ServiceabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		RespondWithValidationError(w, "Invalid request body")
		return
	}

	if len(requests) == 0 {
		RespondWithValidationError(w, "At least one serviceability request is required")
		return
	}

	// Basic validation of each request
	for i, req := range requests {
		if req.PostalCode == "" {
			RespondWithValidationError(w, "Postal code is required for request at index "+strconv.Itoa(i))
			return
		}
		if req.OrderTypeCode == "" {
			RespondWithValidationError(w, "Order type code is required for request at index "+strconv.Itoa(i))
			return
		}
	}

	results, err := h.usecases.ServiceabilityService().BulkCheckServiceability(requests)
	if err != nil {
		RespondWithInternalError(w, "Failed to bulk check serviceability: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, results)
}
