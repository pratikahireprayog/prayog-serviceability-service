package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"prayog-serviceability-service/internal/infrastructure/resilience"
)

// HTTPTransport provides HTTP transport with resilience patterns
type HTTPTransport struct {
	httpClient     *http.Client
	circuitBreaker *resilience.CircuitBreaker
	retryHandler   *resilience.RetryHandler
	timeoutManager *resilience.TimeoutManager
	serviceName    string
	baseURL        string
	defaultHeaders map[string]string
	enableLogging  bool
}

// Config holds configuration for HTTP transport
type Config struct {
	ServiceName    string                   `yaml:"service_name" json:"service_name"`
	BaseURL        string                   `yaml:"base_url" json:"base_url"`
	Timeout        time.Duration            `yaml:"timeout" json:"timeout"`
	DefaultHeaders map[string]string        `yaml:"default_headers" json:"default_headers"`
	EnableLogging  bool                     `yaml:"enable_logging" json:"enable_logging"`
	CircuitBreaker resilience.Config        `yaml:"circuit_breaker" json:"circuit_breaker"`
	RetryPolicy    resilience.RetryPolicy   `yaml:"retry_policy" json:"retry_policy"`
	TimeoutConfig  resilience.TimeoutConfig `yaml:"timeout_config" json:"timeout_config"`
}

// NewHTTPTransport creates a new HTTP transport with resilience patterns
func NewHTTPTransport(config Config) *HTTPTransport {
	// Create HTTP client
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	// Create resilience components
	circuitBreaker := resilience.NewCircuitBreaker(config.CircuitBreaker)
	retryHandler := resilience.NewRetryHandler(config.ServiceName, config.RetryPolicy)
	timeoutManager := resilience.NewTimeoutManager(config.ServiceName, config.TimeoutConfig)

	return &HTTPTransport{
		httpClient:     httpClient,
		circuitBreaker: circuitBreaker,
		retryHandler:   retryHandler,
		timeoutManager: timeoutManager,
		serviceName:    config.ServiceName,
		baseURL:        strings.TrimSuffix(config.BaseURL, "/"),
		defaultHeaders: config.DefaultHeaders,
		enableLogging:  config.EnableLogging,
	}
}

// Request represents an HTTP request
type Request struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Body    interface{}       `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Query   map[string]string `json:"query,omitempty"`
}

// Response represents an HTTP response
type Response struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       interface{}         `json:"body,omitempty"`
}

// Execute performs an HTTP request with full resilience protection
func (ht *HTTPTransport) Execute(ctx context.Context, req Request, target interface{}) error {
	// Wrap the operation with circuit breaker
	return ht.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		// Wrap with retry logic
		return ht.retryHandler.Execute(ctx, func(ctx context.Context) error {
			// Wrap with timeout
			return ht.timeoutManager.WithTimeout(ctx, func(ctx context.Context) error {
				return ht.executeHTTPRequest(ctx, req, target)
			})
		})
	})
}

// executeHTTPRequest performs the actual HTTP request
func (ht *HTTPTransport) executeHTTPRequest(ctx context.Context, req Request, target interface{}) error {
	// Build URL
	url := ht.baseURL + "/" + strings.TrimPrefix(req.Path, "/")

	// Add query parameters
	if len(req.Query) > 0 {
		queryParams := make([]string, 0, len(req.Query))
		for key, value := range req.Query {
			queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
		}
		url = url + "?" + strings.Join(queryParams, "&")
	}

	// Prepare request body
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return NewHTTPError(0, "MARSHAL_ERROR", "Failed to marshal request body", err)
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return NewHTTPError(0, "REQUEST_CREATE_ERROR", "Failed to create HTTP request", err)
	}

	// Set default headers
	for key, value := range ht.defaultHeaders {
		httpReq.Header.Set(key, value)
	}

	// Set request-specific headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set content type for requests with body
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Log request if enabled
	if ht.enableLogging {
		ht.logRequest(req.Method, url, req.Headers)
	}

	// Execute HTTP request
	resp, err := ht.httpClient.Do(httpReq)
	if err != nil {
		if ht.enableLogging {
			ht.logError("HTTP request failed", err)
		}
		return NewHTTPError(0, "NETWORK_ERROR", "Network error occurred", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if ht.enableLogging {
			ht.logError("Failed to read response body", err)
		}
		return NewHTTPError(resp.StatusCode, "RESPONSE_READ_ERROR", "Failed to read response body", err)
	}

	// Log response if enabled
	if ht.enableLogging {
		ht.logResponse(resp.StatusCode, len(respBody))
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return NewHTTPErrorFromResponse(resp.StatusCode, string(respBody))
	}

	// Unmarshal response if target is provided
	if target != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, target); err != nil {
			return NewHTTPError(resp.StatusCode, "UNMARSHAL_ERROR", "Failed to unmarshal response", err)
		}
	}

	return nil
}

// GET performs a GET request
func (ht *HTTPTransport) GET(ctx context.Context, path string, target interface{}) error {
	return ht.Execute(ctx, Request{
		Method: "GET",
		Path:   path,
	}, target)
}

// POST performs a POST request
func (ht *HTTPTransport) POST(ctx context.Context, path string, body interface{}, target interface{}) error {
	return ht.Execute(ctx, Request{
		Method: "POST",
		Path:   path,
		Body:   body,
	}, target)
}

// PUT performs a PUT request
func (ht *HTTPTransport) PUT(ctx context.Context, path string, body interface{}, target interface{}) error {
	return ht.Execute(ctx, Request{
		Method: "PUT",
		Path:   path,
		Body:   body,
	}, target)
}

// DELETE performs a DELETE request
func (ht *HTTPTransport) DELETE(ctx context.Context, path string, target interface{}) error {
	return ht.Execute(ctx, Request{
		Method: "DELETE",
		Path:   path,
	}, target)
}

// GetCircuitBreakerState returns the current circuit breaker state
func (ht *HTTPTransport) GetCircuitBreakerState() resilience.State {
	return ht.circuitBreaker.GetState()
}

// GetMetrics returns transport metrics
func (ht *HTTPTransport) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"service_name":    ht.serviceName,
		"circuit_breaker": ht.circuitBreaker.GetMetrics(),
		"retry_policy":    ht.retryHandler.GetPolicy(),
		"timeout_config":  ht.timeoutManager.GetConfig(),
	}
}

// UpdateConfig updates transport configuration
func (ht *HTTPTransport) UpdateConfig(config Config) {
	ht.baseURL = strings.TrimSuffix(config.BaseURL, "/")
	ht.defaultHeaders = config.DefaultHeaders
	ht.enableLogging = config.EnableLogging
	ht.httpClient.Timeout = config.Timeout

	// Update resilience components
	ht.retryHandler.UpdatePolicy(config.RetryPolicy)
	ht.timeoutManager.UpdateConfig(config.TimeoutConfig)
}

// logRequest logs HTTP request details
func (ht *HTTPTransport) logRequest(method, url string, headers map[string]string) {
	fmt.Printf("[%s] HTTP %s %s (headers: %v)\n", ht.serviceName, method, url, headers)
}

// logResponse logs HTTP response details
func (ht *HTTPTransport) logResponse(statusCode, bodyLength int) {
	fmt.Printf("[%s] HTTP Response %d (body: %d bytes)\n", ht.serviceName, statusCode, bodyLength)
}

// logError logs error details
func (ht *HTTPTransport) logError(message string, err error) {
	fmt.Printf("[%s] ERROR: %s - %v\n", ht.serviceName, message, err)
}
