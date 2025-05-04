package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/prayog/serviceability/internal/domain"
)

// RegisterOrderTypeRoutes registers all order type-related routes
func (h *Handler) RegisterOrderTypeRoutes(r chi.Router) {
	r.Route("/order-types", func(r chi.Router) {
		r.Get("/", h.ListOrderTypes)
		r.Post("/", h.CreateOrderType)
		r.Get("/{id}", h.GetOrderType)
		r.Put("/{id}", h.UpdateOrderType)
		r.Delete("/{id}", h.DeleteOrderType)
		r.Get("/code/{code}", h.GetOrderTypeByCode)
	})
}

// ListOrderTypes handles GET /order-types
func (h *Handler) ListOrderTypes(w http.ResponseWriter, r *http.Request) {
	orderTypes, err := h.usecases.OrderTypeService().List()
	if err != nil {
		RespondWithInternalError(w, "Failed to get order types: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, orderTypes)
}

// GetOrderType handles GET /order-types/{id}
func (h *Handler) GetOrderType(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid order type ID")
		return
	}

	orderType, err := h.usecases.OrderTypeService().GetByID(uint(id))
	if err != nil {
		RespondWithNotFound(w, "Order type not found")
		return
	}

	RespondWithJSON(w, http.StatusOK, orderType)
}

// GetOrderTypeByCode handles GET /order-types/code/{code}
func (h *Handler) GetOrderTypeByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		RespondWithValidationError(w, "Order type code is required")
		return
	}

	orderType, err := h.usecases.OrderTypeService().GetByCode(code)
	if err != nil {
		RespondWithNotFound(w, "Order type not found with code: "+code)
		return
	}

	RespondWithJSON(w, http.StatusOK, orderType)
}

// CreateOrderType handles POST /order-types
func (h *Handler) CreateOrderType(w http.ResponseWriter, r *http.Request) {
	var orderType domain.OrderType
	if err := json.NewDecoder(r.Body).Decode(&orderType); err != nil {
		RespondWithValidationError(w, "Invalid request body")
		return
	}

	if orderType.Name == "" || orderType.Code == "" {
		RespondWithValidationError(w, "Order type name and code are required")
		return
	}

	if err := h.usecases.OrderTypeService().Create(&orderType); err != nil {
		RespondWithInternalError(w, "Failed to create order type: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusCreated, orderType)
}

// UpdateOrderType handles PUT /order-types/{id}
func (h *Handler) UpdateOrderType(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid order type ID")
		return
	}

	var orderType domain.OrderType
	if err := json.NewDecoder(r.Body).Decode(&orderType); err != nil {
		RespondWithValidationError(w, "Invalid request body")
		return
	}

	orderType.ID = uint(id)
	if err := h.usecases.OrderTypeService().Update(&orderType); err != nil {
		RespondWithInternalError(w, "Failed to update order type: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, orderType)
}

// DeleteOrderType handles DELETE /order-types/{id}
func (h *Handler) DeleteOrderType(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		RespondWithValidationError(w, "Invalid order type ID")
		return
	}

	if err := h.usecases.OrderTypeService().Delete(uint(id)); err != nil {
		RespondWithInternalError(w, "Failed to delete order type: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Order type deleted successfully"})
}
