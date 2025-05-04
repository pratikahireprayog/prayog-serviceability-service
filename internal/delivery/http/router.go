package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prayog/serviceability/internal/delivery/http/handlers"
	"github.com/prayog/serviceability/internal/usecase"
)

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(usecases *usecase.UseCaseFactory, version string) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	// Health check
	r.Get("/health", handlers.HealthHandler(version))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Create handler
		handler := handlers.NewHandler(usecases, "/api/v1")

		// Register all routes
		handler.RegisterGeoRoutes(r)
		handler.RegisterOrderTypeRoutes(r)
		handler.RegisterServiceTypeRoutes(r)
		handler.RegisterServiceabilityRoutes(r)
	})

	return r
}
