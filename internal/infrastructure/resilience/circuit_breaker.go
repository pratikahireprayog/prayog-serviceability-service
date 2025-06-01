package resilience

import (
	"context"
	"sync"
	"time"
)

// State represents the state of a circuit breaker
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

// String returns string representation of circuit breaker state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern for any protocol
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            State
	failures         uint32
	requests         uint32
	lastFailureTime  time.Time
	maxRequests      uint32
	timeout          time.Duration
	interval         time.Duration
	failureThreshold uint32
	name             string
}

// Config holds configuration for circuit breaker
type Config struct {
	Name             string        `yaml:"name" json:"name"`
	MaxRequests      uint32        `yaml:"max_requests" json:"max_requests"`
	FailureThreshold uint32        `yaml:"failure_threshold" json:"failure_threshold"`
	Timeout          time.Duration `yaml:"timeout" json:"timeout"`
	Interval         time.Duration `yaml:"interval" json:"interval"`
}

// NewCircuitBreaker creates a new protocol-agnostic circuit breaker
func NewCircuitBreaker(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		name:             config.Name,
		maxRequests:      config.MaxRequests,
		timeout:          config.Timeout,
		interval:         config.Interval,
		failureThreshold: config.FailureThreshold,
	}
}

// Execute runs the given operation with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func(ctx context.Context) error) error {
	if !cb.CanRequest() {
		return &CircuitBreakerError{
			State: cb.GetState(),
			Name:  cb.name,
		}
	}

	err := operation(ctx)

	if err != nil {
		cb.OnFailure()
		return err
	}

	cb.OnSuccess()
	return nil
}

// CanRequest checks if a request can be made based on current state
func (cb *CircuitBreaker) CanRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.requests = 0
			return true
		}
		return false
	case StateHalfOpen:
		return cb.requests < cb.maxRequests
	default:
		return false
	}
}

// OnSuccess records a successful operation
func (cb *CircuitBreaker) OnSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.failures = 0
		cb.state = StateClosed
	}
}

// OnFailure records a failed operation
func (cb *CircuitBreaker) OnFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen || cb.failures >= cb.failureThreshold {
		cb.state = StateOpen
	}

	if cb.state == StateHalfOpen {
		cb.requests++
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns current metrics of the circuit breaker
func (cb *CircuitBreaker) GetMetrics() Metrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return Metrics{
		Name:             cb.name,
		State:            cb.state.String(),
		Failures:         cb.failures,
		Requests:         cb.requests,
		LastFailureTime:  cb.lastFailureTime,
		FailureThreshold: cb.failureThreshold,
		MaxRequests:      cb.maxRequests,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.requests = 0
	cb.lastFailureTime = time.Time{}
}

// Metrics represents circuit breaker metrics
type Metrics struct {
	Name             string    `json:"name"`
	State            string    `json:"state"`
	Failures         uint32    `json:"failures"`
	Requests         uint32    `json:"requests"`
	LastFailureTime  time.Time `json:"last_failure_time"`
	FailureThreshold uint32    `json:"failure_threshold"`
	MaxRequests      uint32    `json:"max_requests"`
}

// CircuitBreakerError represents an error when circuit breaker is open
type CircuitBreakerError struct {
	State State
	Name  string
}

// Error implements error interface
func (e *CircuitBreakerError) Error() string {
	return "circuit breaker '" + e.Name + "' is " + e.State.String()
}

// IsCircuitBreakerOpen checks if error is due to open circuit breaker
func IsCircuitBreakerOpen(err error) bool {
	_, ok := err.(*CircuitBreakerError)
	return ok
}
