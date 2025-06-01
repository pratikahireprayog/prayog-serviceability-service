package resilience

import (
	"context"
	"fmt"
	"time"
)

// TimeoutConfig defines timeout configuration
type TimeoutConfig struct {
	RequestTimeout    time.Duration `yaml:"request_timeout" json:"request_timeout"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout" json:"connection_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
}

// DefaultTimeoutConfig returns sensible default timeout configuration
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		RequestTimeout:    30 * time.Second,
		ConnectionTimeout: 10 * time.Second,
		IdleTimeout:       90 * time.Second,
	}
}

// TimeoutManager manages timeouts for operations
type TimeoutManager struct {
	config      TimeoutConfig
	serviceName string
}

// NewTimeoutManager creates a new timeout manager
func NewTimeoutManager(serviceName string, config TimeoutConfig) *TimeoutManager {
	return &TimeoutManager{
		config:      config,
		serviceName: serviceName,
	}
}

// WithTimeout executes an operation with timeout protection
func (tm *TimeoutManager) WithTimeout(ctx context.Context, operation func(ctx context.Context) error) error {
	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, tm.config.RequestTimeout)
	defer cancel()

	// Channel to receive operation result
	resultChan := make(chan error, 1)

	// Execute operation in goroutine
	go func() {
		resultChan <- operation(timeoutCtx)
	}()

	// Wait for either completion or timeout
	select {
	case err := <-resultChan:
		return err
	case <-timeoutCtx.Done():
		return &TimeoutError{
			Timeout:     tm.config.RequestTimeout,
			ServiceName: tm.serviceName,
			Operation:   "request",
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WithCustomTimeout executes an operation with custom timeout
func (tm *TimeoutManager) WithCustomTimeout(ctx context.Context, timeout time.Duration, operation func(ctx context.Context) error) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultChan := make(chan error, 1)

	go func() {
		resultChan <- operation(timeoutCtx)
	}()

	select {
	case err := <-resultChan:
		return err
	case <-timeoutCtx.Done():
		return &TimeoutError{
			Timeout:     timeout,
			ServiceName: tm.serviceName,
			Operation:   "custom_request",
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetConfig returns the current timeout configuration
func (tm *TimeoutManager) GetConfig() TimeoutConfig {
	return tm.config
}

// UpdateConfig updates the timeout configuration
func (tm *TimeoutManager) UpdateConfig(config TimeoutConfig) {
	tm.config = config
}

// CreateTimeoutContext creates a context with request timeout
func (tm *TimeoutManager) CreateTimeoutContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, tm.config.RequestTimeout)
}

// CreateConnectionContext creates a context with connection timeout
func (tm *TimeoutManager) CreateConnectionContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, tm.config.ConnectionTimeout)
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	Timeout     time.Duration
	ServiceName string
	Operation   string
}

// Error implements error interface
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout after %v for %s operation on service '%s'",
		e.Timeout, e.Operation, e.ServiceName)
}

// IsTimeout returns true for timeout errors
func (e *TimeoutError) IsTimeout() bool {
	return true
}

// IsTimeoutError checks if error is a timeout error
func IsTimeoutError(err error) bool {
	if timeoutErr, ok := err.(*TimeoutError); ok {
		return timeoutErr.IsTimeout()
	}
	return err == context.DeadlineExceeded
}

// GetTimeoutDuration extracts timeout duration from timeout error
func GetTimeoutDuration(err error) time.Duration {
	if timeoutErr, ok := err.(*TimeoutError); ok {
		return timeoutErr.Timeout
	}
	return 0
}
