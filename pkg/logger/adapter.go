package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// AsHTTPLogger returns an adapter that makes the Logger compatible with http.Logger interface
func (l *Logger) AsHTTPLogger() interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
} {
	return &httpLoggerAdapter{logger: l}
}

// httpLoggerAdapter adapts our zap Logger to the HTTP server's Logger interface
type httpLoggerAdapter struct {
	logger *Logger
}

// Info implements the HTTP server's Logger.Info method
func (a *httpLoggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) == 0 {
		a.logger.Info(msg)
		return
	}

	fields := convertToZapFields(keysAndValues)
	a.logger.Info(msg, fields...)
}

// Error implements the HTTP server's Logger.Error method
func (a *httpLoggerAdapter) Error(msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) == 0 {
		a.logger.Error(msg)
		return
	}

	fields := convertToZapFields(keysAndValues)
	a.logger.Error(msg, fields...)
}

// convertToZapFields converts a variadic list of key-value pairs to zap.Field slices
// expects alternating keys and values
func convertToZapFields(keysAndValues []interface{}) []zap.Field {
	if len(keysAndValues)%2 != 0 {
		// If odd number of arguments, add an "unknown" value
		keysAndValues = append(keysAndValues, "<unknown>")
	}

	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			key = fmt.Sprintf("arg%d", i)
		}
		fields = append(fields, zap.Any(key, keysAndValues[i+1]))
	}

	return fields
}
