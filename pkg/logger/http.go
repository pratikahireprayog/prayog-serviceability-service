package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPMiddleware wraps an http.Handler with logging functionality
func (l *Logger) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture status code
		rw := NewResponseWriter(w)

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Build request fields
		fields := []zap.Field{
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("ip", r.RemoteAddr),
			zap.Int("status", rw.statusCode),
			zap.Duration("latency", time.Since(start)),
			zap.String("user_agent", r.UserAgent()),
		}

		// Add trace ID if it exists
		if traceID := r.Header.Get("X-Trace-ID"); traceID != "" {
			fields = append(fields, zap.String("trace_id", traceID))
		}

		// Log at appropriate level based on status code
		msg := "Request completed"
		if rw.statusCode >= 500 {
			l.Error(msg, fields...)
		} else if rw.statusCode >= 400 {
			l.Warn(msg, fields...)
		} else {
			l.Info(msg, fields...)
		}
	})
}

// RecoveryMiddleware wraps an http.Handler with panic recovery
func (l *Logger) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				l.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("ip", r.RemoteAddr),
				)

				// Return internal server error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ResponseWriter is a wrapper for http.ResponseWriter that captures the status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code and passes it to the wrapped ResponseWriter
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
