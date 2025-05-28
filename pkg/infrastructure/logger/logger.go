package logger

import (
	"go.uber.org/zap"
)

// Logger represents a logger instance.
type Logger interface {
	// Info logs an informational message.
	Info(msg string, fields ...zap.Field)
	// Error logs an error message.
	Error(msg string, fields ...zap.Field)
	// Debug logs a debug message.
	Debug(msg string, fields ...zap.Field)
	// Warn logs a warning message.
	Warn(msg string, fields ...zap.Field)
	// Fatal logs a fatal message and terminates the program.
	Fatal(msg string, fields ...zap.Field)
	// With returns a logger with the given fields.
	With(fields ...zap.Field) Logger
	// AsHTTPLogger returns an HTTP logger.
	AsHTTPLogger() HTTPLogger
	// Sync flushes any buffered log entries.
	Sync() error
}

// HTTPLogger represents a logger for HTTP operations.
type HTTPLogger interface {
	// LogRequest logs an HTTP request.
	LogRequest(method, path, status, latency string)
	// LogError logs an HTTP error.
	LogError(err error, method, path string)
}

// zapLogger is a zap implementation of the Logger interface.
type zapLogger struct {
	logger *zap.Logger
}

// zapHTTPLogger is a zap implementation of the HTTPLogger interface.
type zapHTTPLogger struct {
	logger *zap.Logger
}

// NewLogger creates a new logger instance.
func NewLogger() (Logger, error) {
	zapLog, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &zapLogger{logger: zapLog}, nil
}

// Info logs an informational message.
func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Error logs an error message.
func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Debug logs a debug message.
func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Warn logs a warning message.
func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Fatal logs a fatal message and terminates the program.
func (l *zapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// With returns a logger with the given fields.
func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

// AsHTTPLogger returns an HTTP logger.
func (l *zapLogger) AsHTTPLogger() HTTPLogger {
	return &zapHTTPLogger{logger: l.logger}
}

// Sync flushes any buffered log entries.
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// LogRequest logs an HTTP request.
func (l *zapHTTPLogger) LogRequest(method, path, status, latency string) {
	l.logger.Info("HTTP Request",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("status", status),
		zap.String("latency", latency),
	)
}

// LogError logs an HTTP error.
func (l *zapHTTPLogger) LogError(err error, method, path string) {
	l.logger.Error("HTTP Error",
		zap.Error(err),
		zap.String("method", method),
		zap.String("path", path),
	)
}

// Field types for structured logging
var (
	String   = zap.String
	Int      = zap.Int
	Int64    = zap.Int64
	Float64  = zap.Float64
	Bool     = zap.Bool
	Any      = zap.Any
	Error    = zap.Error
	Time     = zap.Time
	Duration = zap.Duration
)
