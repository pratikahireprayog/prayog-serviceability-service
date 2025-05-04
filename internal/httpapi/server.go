package httpapi

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	httplogger "prayog-serviceability-service/internal/infrastructure/logger"
)

// ServerConfig holds configuration for the HTTP server.
type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Logger       httplogger.HTTPLogger
	RouterConfig RouterConfig
}

// RouterConfig holds configuration for the HTTP router.
type RouterConfig struct {
	UsecaseFactory usecase.Factory
	Version        string
	Logger         httplogger.HTTPLogger
	EnableCORS     bool
	AuthEnabled    bool
	APIKeys        []string
	RateLimiting   bool
	Timeout        time.Duration
}

// Server represents an HTTP server using Fiber.
type Server struct {
	app    *fiber.App
	router *Router
	logger httplogger.HTTPLogger
}

// NewServer creates a new HTTP server.
func NewServer(usecaseFactory usecase.Factory, log httplogger.Logger) (*Server, error) {
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

	// Create Fiber app with settings
	app := fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	// Add global middleware
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(recover.New())

	// CORS configuration if enabled
	if routerConfig.EnableCORS {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     "*", // TODO: Configure this in production
			AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders:     "Accept,Authorization,Content-Type,X-CSRF-Token",
			ExposeHeaders:    "Link",
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}))
	}

	// Create Router with the app
	router := NewRouter(app, routerConfig)

	return &Server{
		app:    app,
		router: router,
		logger: httpLogger,
	}, nil
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	return s.app.Listen(":8080") // TODO: Make this configurable
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.Shutdown()
}
