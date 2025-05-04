package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/prayog/serviceability/internal/domain"
)

// RegisterServiceTypeRoutes registers all service type-related routes
func (h *Handler) RegisterServiceTypeRoutes(r chi.Router) {
	r.Route("/service-types", func(r chi.Router) {
		r.Get("/", h.ListServiceTypes)
		r.Post("/", h.CreateServiceType)
		r.Get("/{id}", h.GetServiceType)
		r.Put("/{id}", h.UpdateServiceType)
		r.Delete("/{id}", h.DeleteServiceType)
		r.Get("/code/{code}", h.GetServiceTypeByCode)
	})
}

// ListServiceTypes handles GET /service-types
func (h *Handler) ListServiceTypes(w http.ResponseWriter, r *http.Request) {
	serviceTypes, err := h.usecases.ServiceTypeService().List()
	if err != nil {
		RespondWithInternalError(w, "Failed to get service types: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, serviceTypes)
}

// GetServiceType handles GET /service-types/{id}
func (h *Handler) GetServiceType(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid service type ID")
		return
	}

	serviceType, err := h.usecases.ServiceTypeService().GetByID(uint(id))
	if err != nil {
		RespondWithNotFound(w, "Service type not found")
		return
	}

	RespondWithJSON(w, http.StatusOK, serviceType)
}

// GetServiceTypeByCode handles GET /service-types/code/{code}
func (h *Handler) GetServiceTypeByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		RespondWithValidationError(w, "Service type code is required")
		return
	}

	serviceType, err := h.usecases.ServiceTypeService().GetByCode(code)
	if err != nil {
		RespondWithNotFound(w, "Service type not found with code: "+code)
		return
	}

	RespondWithJSON(w, http.StatusOK, serviceType)
}

// CreateServiceType handles POST /service-types
func (h *Handler) CreateServiceType(w http.ResponseWriter, r *http.Request) {
	var serviceType domain.ServiceType
	if err := json.NewDecoder(r.Body).Decode(&serviceType); err != nil {
		RespondWithValidationError(w, "Invalid request body")
		return
	}

	if serviceType.Name == "" || serviceType.Code == "" {
		RespondWithValidationError(w, "Service type name and code are required")
		return
	}

	if err := h.usecases.ServiceTypeService().Create(&serviceType); err != nil {
		RespondWithInternalError(w, "Failed to create service type: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusCreated, serviceType)
}

// UpdateServiceType handles PUT /service-types/{id}
func (h *Handler) UpdateServiceType(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid service type ID")
		return
	}

	var serviceType domain.ServiceType
	if err := json.NewDecoder(r.Body).Decode(&serviceType); err != nil {
		RespondWithValidationError(w, "Invalid request body")
		return
	}

	serviceType.ID = uint(id)
	if err := h.usecases.ServiceTypeService().Update(&serviceType); err != nil {
		RespondWithInternalError(w, "Failed to update service type: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, serviceType)
}

// DeleteServiceType handles DELETE /service-types/{id}
func (h *Handler) DeleteServiceType(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid service type ID")
		return
	}

	if err := h.usecases.ServiceTypeService().Delete(uint(id)); err != nil {
		RespondWithInternalError(w, "Failed to delete service type: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Service type deleted successfully"})
}
