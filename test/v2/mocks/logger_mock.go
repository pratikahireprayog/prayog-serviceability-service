package mocks

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LogEntry represents a log entry captured by the mock logger
type LogEntry struct {
	Level     logrus.Level
	Message   string
	Fields    logrus.Fields
	Time      time.Time
	Formatted string
}

// MockLogger is a mock implementation of logrus.FieldLogger for testing
type MockLogger struct {
	entries   []LogEntry
	level     logrus.Level
	formatter logrus.Formatter
	mu        sync.RWMutex
}

// NewMockLogger creates a new mock logger instance
func NewMockLogger() *MockLogger {
	return &MockLogger{
		entries:   make([]LogEntry, 0),
		level:     logrus.InfoLevel,
		formatter: &logrus.TextFormatter{},
	}
}

// Debug logs a message at Debug level
func (m *MockLogger) Debug(args ...interface{}) {
	m.log(logrus.DebugLevel, fmt.Sprint(args...), nil)
}

// Debugf logs a formatted message at Debug level
func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.log(logrus.DebugLevel, fmt.Sprintf(format, args...), nil)
}

// Info logs a message at Info level
func (m *MockLogger) Info(args ...interface{}) {
	m.log(logrus.InfoLevel, fmt.Sprint(args...), nil)
}

// Infof logs a formatted message at Info level
func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.log(logrus.InfoLevel, fmt.Sprintf(format, args...), nil)
}

// Warn logs a message at Warn level
func (m *MockLogger) Warn(args ...interface{}) {
	m.log(logrus.WarnLevel, fmt.Sprint(args...), nil)
}

// Warnf logs a formatted message at Warn level
func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.log(logrus.WarnLevel, fmt.Sprintf(format, args...), nil)
}

// Error logs a message at Error level
func (m *MockLogger) Error(args ...interface{}) {
	m.log(logrus.ErrorLevel, fmt.Sprint(args...), nil)
}

// Errorf logs a formatted message at Error level
func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.log(logrus.ErrorLevel, fmt.Sprintf(format, args...), nil)
}

// Fatal logs a message at Fatal level
func (m *MockLogger) Fatal(args ...interface{}) {
	m.log(logrus.FatalLevel, fmt.Sprint(args...), nil)
}

// Fatalf logs a formatted message at Fatal level
func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.log(logrus.FatalLevel, fmt.Sprintf(format, args...), nil)
}

// Panic logs a message at Panic level
func (m *MockLogger) Panic(args ...interface{}) {
	m.log(logrus.PanicLevel, fmt.Sprint(args...), nil)
}

// Panicf logs a formatted message at Panic level
func (m *MockLogger) Panicf(format string, args ...interface{}) {
	m.log(logrus.PanicLevel, fmt.Sprintf(format, args...), nil)
}

// WithField adds a field to the logger
func (m *MockLogger) WithField(key string, value interface{}) *logrus.Entry {
	return m.toLogrus().WithField(key, value)
}

// WithFields adds multiple fields to the logger
func (m *MockLogger) WithFields(fields logrus.Fields) *logrus.Entry {
	return m.toLogrus().WithFields(fields)
}

// WithError adds an error field to the logger
func (m *MockLogger) WithError(err error) *logrus.Entry {
	return m.toLogrus().WithError(err)
}

// SetLevel sets the logging level
func (m *MockLogger) SetLevel(level logrus.Level) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.level = level
}

// GetLevel returns the current logging level
func (m *MockLogger) GetLevel() logrus.Level {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.level
}

// SetFormatter sets the log formatter
func (m *MockLogger) SetFormatter(formatter logrus.Formatter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.formatter = formatter
}

// log is the internal logging method
func (m *MockLogger) log(level logrus.Level, message string, fields logrus.Fields) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we should log this level
	if level > m.level {
		return
	}

	// Create log entry
	entry := LogEntry{
		Level:   level,
		Message: message,
		Fields:  fields,
		Time:    time.Now(),
	}

	// Format the entry
	logrusEntry := &logrus.Entry{
		Logger:  m.toLogrus(),
		Level:   level,
		Message: message,
		Time:    entry.Time,
		Data:    fields,
	}

	if m.formatter != nil {
		if formatted, err := m.formatter.Format(logrusEntry); err == nil {
			entry.Formatted = string(formatted)
		}
	}

	// Add to entries
	m.entries = append(m.entries, entry)
}

// toLogrus converts the mock logger to a logrus logger for compatibility
func (m *MockLogger) toLogrus() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(m.level)
	if m.formatter != nil {
		logger.SetFormatter(m.formatter)
	}
	return logger
}

// GetEntries returns all log entries
func (m *MockLogger) GetEntries() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent concurrent modification
	entries := make([]LogEntry, len(m.entries))
	copy(entries, m.entries)
	return entries
}

// GetEntriesForLevel returns all log entries for a specific level
func (m *MockLogger) GetEntriesForLevel(level logrus.Level) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []LogEntry
	for _, entry := range m.entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// GetLastEntry returns the last log entry, or nil if no entries
func (m *MockLogger) GetLastEntry() *LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.entries) == 0 {
		return nil
	}

	entry := m.entries[len(m.entries)-1]
	return &entry
}

// HasEntry checks if a log entry with the given message exists
func (m *MockLogger) HasEntry(message string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entry := range m.entries {
		if entry.Message == message {
			return true
		}
	}
	return false
}

// HasEntryWithLevel checks if a log entry with the given level and message exists
func (m *MockLogger) HasEntryWithLevel(level logrus.Level, message string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entry := range m.entries {
		if entry.Level == level && entry.Message == message {
			return true
		}
	}
	return false
}

// HasEntryContaining checks if any log entry contains the given text
func (m *MockLogger) HasEntryContaining(text string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entry := range m.entries {
		if strings.Contains(entry.Message, text) {
			return true
		}
	}
	return false
}

// HasErrorWithMessage checks if an error log entry with the given message exists
func (m *MockLogger) HasErrorWithMessage(message string) bool {
	return m.HasEntryWithLevel(logrus.ErrorLevel, message)
}

// HasWarningWithMessage checks if a warning log entry with the given message exists
func (m *MockLogger) HasWarningWithMessage(message string) bool {
	return m.HasEntryWithLevel(logrus.WarnLevel, message)
}

// HasInfoWithMessage checks if an info log entry with the given message exists
func (m *MockLogger) HasInfoWithMessage(message string) bool {
	return m.HasEntryWithLevel(logrus.InfoLevel, message)
}

// HasDebugWithMessage checks if a debug log entry with the given message exists
func (m *MockLogger) HasDebugWithMessage(message string) bool {
	return m.HasEntryWithLevel(logrus.DebugLevel, message)
}

// CountEntries returns the total number of log entries
func (m *MockLogger) CountEntries() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

// CountEntriesForLevel returns the number of log entries for a specific level
func (m *MockLogger) CountEntriesForLevel(level logrus.Level) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, entry := range m.entries {
		if entry.Level == level {
			count++
		}
	}
	return count
}

// Clear removes all log entries
func (m *MockLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = make([]LogEntry, 0)
}

// Reset clears all entries and resets the logger to default state
func (m *MockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = make([]LogEntry, 0)
	m.level = logrus.InfoLevel
	m.formatter = &logrus.TextFormatter{}
}

// String returns a string representation of all log entries
func (m *MockLogger) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var builder strings.Builder
	for i, entry := range m.entries {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf("[%s] %s: %s",
			entry.Time.Format("2006-01-02 15:04:05"),
			entry.Level.String(),
			entry.Message))

		if len(entry.Fields) > 0 {
			builder.WriteString(" ")
			builder.WriteString(fmt.Sprintf("%+v", entry.Fields))
		}
	}
	return builder.String()
}

