package handlers

import (
	"net/http"

	"prayog-serviceability-service/internal/service/usecase"

	"github.com/go-chi/chi/v5"
)

// RegisterHandlers registers all HTTP handlers with the router.
func RegisterHandlers(r chi.Router, usecaseFactory usecase.Factory) {
	// Health check
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	// Register serviceability handlers
	serviceabilityHandler := NewServiceabilityHandler(usecaseFactory)
	serviceabilityHandler.RegisterRoutes(r)
}
