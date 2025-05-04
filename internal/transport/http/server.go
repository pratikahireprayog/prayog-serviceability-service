package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Logger interface defines the logging methods the server needs
type Logger interface {
	// Info logs an informational message
	Info(msg string, keysAndValues ...interface{})
	// Error logs an error message
	Error(msg string, keysAndValues ...interface{})
}

// stdLogger adapts the standard log.Logger to our Logger interface
type stdLogger struct {
	logger *log.Logger
}

func (l *stdLogger) Info(msg string, _ ...interface{}) {
	l.logger.Println(msg)
}

func (l *stdLogger) Error(msg string, _ ...interface{}) {
	l.logger.Println("ERROR:", msg)
}

// ServerConfig holds all configuration for the HTTP server
type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Logger       Logger
	RouterConfig RouterConfig
}

// Server represents the HTTP server
type Server struct {
	server *http.Server
	logger Logger
}

// ServerOption defines a function to customize the server
type ServerOption func(*Server)

// WithLogger sets a custom logger for the server
func WithLogger(logger Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithReadTimeout sets the read timeout for the server
func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.server.ReadTimeout = timeout
	}
}

// WithWriteTimeout sets the write timeout for the server
func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.server.WriteTimeout = timeout
	}
}

// WithIdleTimeout sets the idle timeout for the server
func WithIdleTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.server.IdleTimeout = timeout
	}
}

// DefaultServerConfig returns a default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Logger:       &stdLogger{log.New(log.Writer(), "[HTTP] ", log.LstdFlags)},
	}
}

// NewServer creates a new HTTP server with configuration
func NewServer(config ServerConfig) *Server {
	// Create the router with the provided config
	handler := NewRouter(config.RouterConfig)

	srv := &Server{
		server: &http.Server{
			Addr:         config.Addr,
			Handler:      handler,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
		logger: config.Logger,
	}

	return srv
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info(fmt.Sprintf("Server starting on %s", s.server.Addr))
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Server is shutting down...")
	return s.server.Shutdown(ctx)
}
