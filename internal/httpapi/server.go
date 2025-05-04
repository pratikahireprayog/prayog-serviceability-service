package httpapi

import (
	"context"
	"net/http"
	"time"

	"prayog-serviceability-service/internal/infrastructure/logger"
	"prayog-serviceability-service/internal/service/usecase"
)

// ServerConfig holds configuration for the HTTP server.
type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Logger       logger.HTTPLogger
	RouterConfig RouterConfig
}

// RouterConfig holds configuration for the HTTP router.
type RouterConfig struct {
	UsecaseFactory usecase.Factory
	Version        string
	Logger         logger.HTTPLogger
	EnableCORS     bool
	AuthEnabled    bool
	APIKeys        []string
	RateLimiting   bool
	Timeout        time.Duration
}

// Server represents an HTTP server.
type Server struct {
	server *http.Server
	router *Router
	logger logger.HTTPLogger
}

// NewServer creates a new HTTP server.
func NewServer(usecaseFactory usecase.Factory, log logger.Logger) (*Server, error) {
	httpLogger := log.AsHTTPLogger()

	routerConfig := RouterConfig{
		UsecaseFactory: usecaseFactory,
		Version:        "1.0.0", // TODO: Make this configurable
		Logger:         httpLogger,
		EnableCORS:     true,
		AuthEnabled:    false,
		APIKeys:        []string{}, // TODO: Load from config
		RateLimiting:   true,
		Timeout:        30 * time.Second,
	}

	router := NewRouter(routerConfig)

	serverConfig := &http.Server{
		Addr:         ":8080", // TODO: Make this configurable
		Handler:      router.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		server: serverConfig,
		router: router,
		logger: httpLogger,
	}, nil
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
