package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"prayog-serviceability-service/internal/httpapi/handlers"
	apimiddleware "prayog-serviceability-service/internal/httpapi/middleware"
)

// Router handles HTTP routing for the application.
type Router struct {
	config RouterConfig
	router *chi.Mux
}

// NewRouter creates a new HTTP router.
func NewRouter(config RouterConfig) *Router {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(config.Timeout))

	// CORS configuration if enabled
	if config.EnableCORS {
		corsMiddleware := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"}, // TODO: Configure this in production
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		})
		r.Use(corsMiddleware.Handler)
	}

	// Rate limiting if enabled
	if config.RateLimiting {
		r.Use(apimiddleware.RateLimiter)
	}

	// Auth middleware if enabled
	if config.AuthEnabled {
		r.Use(apimiddleware.Auth(config.APIKeys))
	}

	// Health check endpoints
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(config.Version))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/serviceability", func(r chi.Router) {
			handlers.RegisterHandlers(r, config.UsecaseFactory)
		})
	})

	return &Router{
		config: config,
		router: r,
	}
}

// Handler returns the HTTP handler for the router.
func (r *Router) Handler() http.Handler {
	return r.router
}
