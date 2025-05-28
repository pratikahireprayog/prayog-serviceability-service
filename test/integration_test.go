package test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"prayog-serviceability-service/pkg/config"
	"prayog-serviceability-service/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestConfigAndLoggerIntegration(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tmpDir)

	// Create a test log config directly
	logConfig := &config.LogConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	}

	// Create a custom buffer for testing log output
	var buf bytes.Buffer
	testCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)
	testLogger := &logger.Logger{Logger: zap.New(testCore)}

	// Test that logger functions work correctly with different log levels
	testLogger.Info("Test info message", logger.String("key", "value"))
	testLogger.Debug("Test debug message", logger.Int("count", 42))
	testLogger.Warn("Test warn message", logger.Bool("warning", true))
	testLogger.Error("Test error message", logger.Error(assert.AnError))

	// Test structured logging output
	logs := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	assert.Equal(t, 4, len(logs), "Should have 4 log entries")

	// Validate each log entry format and structure
	logLevels := []string{"info", "debug", "warn", "error"}
	for i, logBytes := range logs {
		var logMap map[string]interface{}
		err := json.Unmarshal(logBytes, &logMap)
		require.NoError(t, err, "Log entry should be valid JSON")
		assert.Equal(t, logLevels[i], logMap["level"], "Log level should match")
	}

	// Test with actual logger
	actualLogger, err := logger.NewLogger(logConfig)
	assert.NoError(t, err, "Should be able to create logger from config")
	assert.NotNil(t, actualLogger, "Logger should not be nil")

	// Test with console format
	consoleConfig := &config.LogConfig{
		Level:      "info",
		Format:     "console",
		OutputPath: "stdout",
	}
	consoleLogger, err := logger.NewLogger(consoleConfig)
	assert.NoError(t, err, "Should be able to create logger with console format")
	assert.NotNil(t, consoleLogger, "Console logger should not be nil")

	// Test with file output
	tmpLogFile := filepath.Join(tmpDir, "test.log")
	fileConfig := &config.LogConfig{
		Level:      "info",
		Format:     "json",
		OutputPath: tmpLogFile,
	}
	fileLogger, err := logger.NewLogger(fileConfig)
	assert.NoError(t, err, "Should be able to create logger with file output")
	assert.NotNil(t, fileLogger, "File logger should not be nil")

	// Write to the file and verify it exists
	fileLogger.Info("Test file logging")
	_, err = os.Stat(tmpLogFile)
	assert.NoError(t, err, "Log file should exist")

	// Test with invalid file path (should error)
	invalidConfig := &config.LogConfig{
		Level:      "info",
		Format:     "json",
		OutputPath: "/nonexistent/directory/test.log",
	}
	_, err = logger.NewLogger(invalidConfig)
	assert.Error(t, err, "Should error with invalid file path")
}

func TestLoggerWithFieldsIntegration(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a custom core that writes to our buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	)

	// Create a logger with the custom core
	testLogger := &logger.Logger{Logger: zap.New(core)}

	// Test with fields from the logger package
	testLogger.Info("Structured log message",
		logger.String("string_field", "string_value"),
		logger.Int("int_field", 123),
		logger.Int64("int64_field", 123456789),
		logger.Float64("float_field", 123.456),
		logger.Bool("bool_field", true),
		logger.Error(assert.AnError),
		logger.Time("time_field", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
		logger.Duration("duration_field", 5*time.Second),
	)

	// Verify the log has all the fields
	var logMap map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logMap)
	require.NoError(t, err, "Log entry should be valid JSON")

	assert.Equal(t, "Structured log message", logMap["msg"])
	assert.Equal(t, "string_value", logMap["string_field"])
	assert.Equal(t, float64(123), logMap["int_field"])
	assert.Equal(t, float64(123456789), logMap["int64_field"])
	assert.Equal(t, 123.456, logMap["float_field"])
	assert.Equal(t, true, logMap["bool_field"])
	assert.Contains(t, logMap["error"], "assert.AnError")
	assert.Contains(t, logMap, "time_field", "Time field should exist")
	assert.Contains(t, logMap, "duration_field", "Duration field should exist")
}

func TestConfigErrorHandling(t *testing.T) {
	// Test with invalid log level (should fall back to info)
	invalidLevelConfig := &config.LogConfig{
		Level:      "warning",
		Format:     "json",
		OutputPath: "stdout",
	}

	loggerInst, err := logger.NewLogger(invalidLevelConfig)
	assert.NoError(t, err, "NewLogger should succeed with unknown log level (fallback to info)")
	assert.NotNil(t, loggerInst, "Logger should not be nil")

	// Test with permission issues for log file
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	// Try to write to a directory where we (likely) don't have permission
	restrictedConfig := &config.LogConfig{
		Level:      "info",
		Format:     "json",
		OutputPath: "/root/test.log",
	}

	// This should fail when trying to create a logger
	_, err = logger.NewLogger(restrictedConfig)
	assert.Error(t, err, "NewLogger should fail with permission issues")
}
