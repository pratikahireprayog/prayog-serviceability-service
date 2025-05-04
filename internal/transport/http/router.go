package http

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prayog/serviceability/internal/transport/http/handlers"
	"github.com/prayog/serviceability/internal/transport/http/middleware"
	"github.com/prayog/serviceability/internal/usecase"
)

// RouterConfig holds configuration for the router
type RouterConfig struct {
	UseCases     *usecase.UseCaseFactory
	Version      string
	Logger       Logger
	EnableCORS   bool
	AuthEnabled  bool
	APIKeys      []string
	RateLimiting bool
}

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(config RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Standard Chi middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Heartbeat("/ping"))

	// Add logging middleware
	if config.Logger != nil {
		// Create a standard logger for the middleware
		stdLogger := log.New(log.Writer(), "[HTTP] ", log.LstdFlags)
		r.Use(middleware.Logger(stdLogger))
	} else {
		r.Use(chiMiddleware.Logger)
	}

	// Add response standardization middleware
	r.Use(middleware.Response())

	// Add CORS if enabled
	if config.EnableCORS {
		r.Use(middleware.CORS([]string{"*"}))
	}

	// Add authentication if enabled
	if config.AuthEnabled {
		r.Use(middleware.Auth(middleware.AuthConfig{
			APIKeys: config.APIKeys,
		}))
	}

	// Rate limiting if enabled
	if config.RateLimiting {
		r.Use(chiMiddleware.Throttle(100))
	}

	// Create handler factory
	handlerFactory := handlers.NewFactory(config.UseCases, config.Version)

	// Health check endpoint
	r.Get("/health", handlerFactory.HealthHandler())

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Get serviceability handler
		serviceabilityHandler := handlerFactory.ServiceabilityHandler()

		// Register serviceability routes
		r.Route("/serviceability", func(r chi.Router) {
			serviceabilityHandler.RegisterRoutes(r)
		})
	})

	return r
}
