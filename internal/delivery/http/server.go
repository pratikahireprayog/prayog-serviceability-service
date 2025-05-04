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

// NewServer creates a new HTTP server
func NewServer(addr string, handler http.Handler, opts ...ServerOption) *Server {
	srv := &Server{
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		logger: &stdLogger{log.New(log.Writer(), "[HTTP] ", log.LstdFlags)},
	}

	// Apply options
	for _, opt := range opts {
		opt(srv)
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
