package resilience

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries         int              `yaml:"max_retries" json:"max_retries"`
	InitialDelay       time.Duration    `yaml:"initial_delay" json:"initial_delay"`
	MaxDelay           time.Duration    `yaml:"max_delay" json:"max_delay"`
	BackoffMultiplier  float64          `yaml:"backoff_multiplier" json:"backoff_multiplier"`
	Jitter             bool             `yaml:"jitter" json:"jitter"`
	RetryableErrorFunc func(error) bool `yaml:"-" json:"-"`
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:        3,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
		RetryableErrorFunc: func(err error) bool {
			// Default: retry all errors except context cancellation
			return err != context.Canceled && err != context.DeadlineExceeded
		},
	}
}

// RetryHandler implements retry logic with exponential backoff
type RetryHandler struct {
	policy RetryPolicy
	name   string
}

// NewRetryHandler creates a new retry handler
func NewRetryHandler(name string, policy RetryPolicy) *RetryHandler {
	if policy.RetryableErrorFunc == nil {
		policy.RetryableErrorFunc = DefaultRetryPolicy().RetryableErrorFunc
	}

	return &RetryHandler{
		policy: policy,
		name:   name,
	}
}

// Execute executes the operation with retry logic
func (rh *RetryHandler) Execute(ctx context.Context, operation func(ctx context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt <= rh.policy.MaxRetries; attempt++ {
		// First attempt doesn't need delay
		if attempt > 0 {
			delay := rh.calculateDelay(attempt)

			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return &RetryError{
					Attempts:    attempt,
					LastError:   lastErr,
					Reason:      "context cancelled during retry delay",
					ServiceName: rh.name,
				}
			}
		}

		err := operation(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !rh.policy.RetryableErrorFunc(err) {
			return &RetryError{
				Attempts:    attempt + 1,
				LastError:   err,
				Reason:      "error is not retryable",
				ServiceName: rh.name,
			}
		}

		// If this was the last attempt, don't retry
		if attempt == rh.policy.MaxRetries {
			break
		}
	}

	return &RetryError{
		Attempts:    rh.policy.MaxRetries + 1,
		LastError:   lastErr,
		Reason:      "max retries exceeded",
		ServiceName: rh.name,
	}
}

// calculateDelay calculates the delay for the given attempt with exponential backoff
func (rh *RetryHandler) calculateDelay(attempt int) time.Duration {
	delay := float64(rh.policy.InitialDelay) * math.Pow(rh.policy.BackoffMultiplier, float64(attempt-1))

	// Apply max delay
	if time.Duration(delay) > rh.policy.MaxDelay {
		delay = float64(rh.policy.MaxDelay)
	}

	// Add jitter if enabled
	if rh.policy.Jitter {
		// Add up to 20% jitter
		jitter := 0.8 + (0.4 * float64(time.Now().UnixNano()%100) / 100.0)
		delay = delay * jitter
	}

	return time.Duration(delay)
}

// GetPolicy returns the current retry policy
func (rh *RetryHandler) GetPolicy() RetryPolicy {
	return rh.policy
}

// UpdatePolicy updates the retry policy
func (rh *RetryHandler) UpdatePolicy(policy RetryPolicy) {
	if policy.RetryableErrorFunc == nil {
		policy.RetryableErrorFunc = rh.policy.RetryableErrorFunc
	}
	rh.policy = policy
}

// RetryError represents an error that occurred during retry attempts
type RetryError struct {
	Attempts    int
	LastError   error
	Reason      string
	ServiceName string
}

// Error implements error interface
func (e *RetryError) Error() string {
	return fmt.Sprintf("retry failed for service '%s' after %d attempts: %s (last error: %v)",
		e.ServiceName, e.Attempts, e.Reason, e.LastError)
}

// Unwrap returns the last error for error unwrapping
func (e *RetryError) Unwrap() error {
	return e.LastError
}

// IsRetryError checks if error is a retry error
func IsRetryError(err error) bool {
	_, ok := err.(*RetryError)
	return ok
}

// GetLastError extracts the last error from a retry error
func GetLastError(err error) error {
	if retryErr, ok := err.(*RetryError); ok {
		return retryErr.LastError
	}
	return err
}
