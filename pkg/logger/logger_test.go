package logger

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/prayog/serviceability/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name       string
		config     config.LogConfig
		wantErr    bool
		wantLevel  zapcore.Level
		wantFormat string
	}{
		{
			name: "default config",
			config: config.LogConfig{
				Level:      "info",
				Format:     "json",
				OutputPath: "stdout",
			},
			wantErr:    false,
			wantLevel:  zapcore.InfoLevel,
			wantFormat: "json",
		},
		{
			name: "debug level",
			config: config.LogConfig{
				Level:      "debug",
				Format:     "json",
				OutputPath: "stdout",
			},
			wantErr:    false,
			wantLevel:  zapcore.DebugLevel,
			wantFormat: "json",
		},
		{
			name: "console format",
			config: config.LogConfig{
				Level:      "info",
				Format:     "console",
				OutputPath: "stdout",
			},
			wantErr:    false,
			wantLevel:  zapcore.InfoLevel,
			wantFormat: "console",
		},
		{
			name: "invalid level",
			config: config.LogConfig{
				Level:      "invalid",
				Format:     "json",
				OutputPath: "stdout",
			},
			wantErr: false, // We fallback to info level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(&tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, logger)
		})
	}
}

func TestLoggerOutput(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a custom core that writes to our buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	)

	// Create a logger with the custom core
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// Log a message
	logger.Info("test message", zap.String("key", "value"))

	// Verify the output contains the expected data
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	require.NoError(t, err, "log output should be valid JSON")

	assert.Equal(t, "test message", logData["msg"])
	assert.Equal(t, "value", logData["key"])
	assert.Equal(t, "info", logData["level"])
}

func TestLoggerLevels(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create temporary file for log output
	tmpFile, err := os.CreateTemp("", "logger_test")
	require.NoError(t, err, "Failed to create temp file")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Configure logger to write to both buffer and file
	cfg := &config.LogConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: tmpFile.Name(),
	}

	// Create logger
	logger, err := NewLogger(cfg)
	require.NoError(t, err, "Failed to create logger")

	// Replace the core to write to our buffer instead
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)
	logger.Logger = zap.New(core)

	// Log messages at different levels
	logger.Debug("debug message", zap.Int("count", 1))
	logger.Info("info message", zap.Int("count", 2))
	logger.Warn("warn message", zap.Int("count", 3))
	logger.Error("error message", zap.Int("count", 4))

	// Split the buffer by newlines to get individual log entries
	logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Equal(t, 4, len(logs), "Should have 4 log entries")

	// Parse each log entry and verify
	logLevels := []string{"debug", "info", "warn", "error"}
	for i, log := range logs {
		var logData map[string]interface{}
		err := json.Unmarshal([]byte(log), &logData)
		require.NoError(t, err, "log entry should be valid JSON")

		assert.Equal(t, logLevels[i], logData["level"], "Log level should match")
		assert.Equal(t, logLevels[i]+" message", logData["msg"], "Log message should match")
		assert.Equal(t, float64(i+1), logData["count"], "Log count should match")
	}
}

func TestHTTPMiddleware(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a custom core that writes to our buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	)

	// Create a logger with the custom core
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// Create a test handler that returns a 200 status code
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap the test handler with the HTTP middleware
	handler := logger.HTTPMiddleware(testHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Trace-ID", "test-trace-id")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, rr.Code, "Handler should return 200 OK")
	assert.Equal(t, "OK", rr.Body.String(), "Handler should return 'OK'")

	// Verify the log output
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	require.NoError(t, err, "log output should be valid JSON")

	assert.Equal(t, "Request completed", logData["msg"], "Log message should be 'Request completed'")
	assert.Equal(t, "GET", logData["method"], "Method should be GET")
	assert.Equal(t, "/test", logData["path"], "Path should be /test")
	assert.Equal(t, "param=value", logData["query"], "Query should be param=value")
	assert.Equal(t, float64(200), logData["status"], "Status should be 200")
	assert.Equal(t, "test-agent", logData["user_agent"], "User-Agent should be test-agent")
	assert.Equal(t, "test-trace-id", logData["trace_id"], "X-Trace-ID should be test-trace-id")
}

func TestRecoveryMiddleware(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a custom core that writes to our buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	)

	// Create a logger with the custom core
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// Create a test handler that panics
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap the test handler with the recovery middleware
	handler := logger.RecoveryMiddleware(testHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request (should not panic)
	handler.ServeHTTP(rr, req)

	// Verify the response
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Handler should return 500 Internal Server Error")
	assert.Equal(t, "Internal Server Error\n", rr.Body.String(), "Handler should return error message")

	// Verify the log output
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	require.NoError(t, err, "log output should be valid JSON")

	assert.Equal(t, "Panic recovered", logData["msg"], "Log message should be 'Panic recovered'")
	assert.Equal(t, "test panic", logData["error"], "Error should be 'test panic'")
	assert.Equal(t, "GET", logData["method"], "Method should be GET")
	assert.Equal(t, "/test", logData["path"], "Path should be /test")
}

func TestLoggerWithFields(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a custom core that writes to our buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	)

	// Create a logger with the custom core
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// Create a child logger with fields
	childLogger := logger.With(zap.String("component", "test"))

	// Log a message with the child logger
	childLogger.Info("child logger message", zap.Int("count", 42))

	// Verify the output contains the expected data
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	require.NoError(t, err, "log output should be valid JSON")

	assert.Equal(t, "child logger message", logData["msg"])
	assert.Equal(t, "test", logData["component"])
	assert.Equal(t, float64(42), logData["count"])
}

func TestHttpLoggerAdapter(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a custom core that writes to our buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	)

	// Create a logger with the custom core
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// Create adapter
	adapter := logger.AsHTTPLogger()

	// Use adapter to log messages
	adapter.Info("info message")
	buf.Reset() // Clear buffer

	adapter.Info("info with fields", "key1", "value1", "key2", 42)

	// Verify the output contains the expected data
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	require.NoError(t, err, "log output should be valid JSON")

	assert.Equal(t, "info with fields", logData["msg"])
	assert.Equal(t, "value1", logData["key1"])
	assert.Equal(t, float64(42), logData["key2"])

	// Test error logging
	buf.Reset()
	adapter.Error("error message", "err_key", "error_value")

	err = json.Unmarshal(buf.Bytes(), &logData)
	require.NoError(t, err, "log output should be valid JSON")

	assert.Equal(t, "error message", logData["msg"])
	assert.Equal(t, "error_value", logData["err_key"])
	assert.Equal(t, "error", logData["level"])
}

func TestResponseWriter(t *testing.T) {
	// Create a mock ResponseWriter
	mockRW := httptest.NewRecorder()

	// Create our ResponseWriter wrapper
	rw := NewResponseWriter(mockRW)

	// Initial status should be 200 OK
	assert.Equal(t, http.StatusOK, rw.statusCode, "Initial status code should be 200 OK")

	// Write a header with a different status code
	rw.WriteHeader(http.StatusNotFound)

	// Status code should be updated
	assert.Equal(t, http.StatusNotFound, rw.statusCode, "Status code should be updated to 404 Not Found")
	assert.Equal(t, http.StatusNotFound, mockRW.Code, "Underlying ResponseWriter status should also be 404")

	// Test writing content
	content := []byte("Not Found")
	_, err := rw.Write(content)
	assert.NoError(t, err, "Writing to ResponseWriter should not error")
	assert.Equal(t, string(content), mockRW.Body.String(), "Content should be written to underlying ResponseWriter")
}
