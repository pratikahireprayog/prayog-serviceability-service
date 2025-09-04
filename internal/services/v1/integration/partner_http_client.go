package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"time"
)

// PartnerHTTPClient implements HTTPClient interface with timeout and retry logic
type PartnerHTTPClient struct {
	client        *http.Client
	retryAttempts int
	retryDelay    time.Duration
	logger        *log.Logger
}

// NewPartnerHTTPClient creates a new HTTP client with timeout and retry configuration
func NewPartnerHTTPClient(config PartnerValidationConfig, logger *log.Logger) HTTPClient {
	// Create HTTP client with timeout and connection pooling
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,  // Faster connection timeout
			KeepAlive: 60 * time.Second, // Longer keep-alive timeout
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,        // Faster TLS handshake
		ResponseHeaderTimeout: 10 * time.Second,       // Faster response timeout
		ExpectContinueTimeout: 500 * time.Millisecond, // Faster expect timeout
		MaxIdleConns:          10000,                  // 10x more idle connections
		MaxIdleConnsPerHost:   1000,                   // 100x more idle connections per host
		MaxConnsPerHost:       0,                      // 0 = unlimited connections per host
		IdleConnTimeout:       300 * time.Second,      // Longer idle timeout (5 minutes)
		DisableCompression:    false,                  // Keep compression for efficiency
		DisableKeepAlives:     false,                  // Keep alive for reuse
		ForceAttemptHTTP2:     true,                   // Use HTTP/2 for better performance
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.RequestTimeout,
	}

	return &PartnerHTTPClient{
		client:        client,
		retryAttempts: config.RetryAttempts,
		retryDelay:    config.RetryDelay,
		logger:        logger,
	}
}

// Get performs HTTP GET request with retry logic and timeout handling
func (c *PartnerHTTPClient) Get(ctx context.Context, url string) (*HTTPResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		// Check if context is already cancelled
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
		default:
		}

		// Log attempt
		if attempt > 0 {
			c.logger.Printf("HTTP GET retry attempt %d/%d for URL: %s", attempt, c.retryAttempts, url)
		} else {
			c.logger.Printf("HTTP GET request to URL: %s", url)
		}

		// Create request with context
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("User-Agent", "prayog-serviceability-service/1.0")
		req.Header.Set("Accept", "application/json")

		// Execute request
		startTime := time.Now()
		resp, err := c.client.Do(req)
		duration := time.Since(startTime)

		// Log request duration
		c.logger.Printf("HTTP GET request to %s completed in %v", url, duration)

		if err != nil {
			lastErr = err
			c.logger.Printf("HTTP GET request failed (attempt %d/%d): %v", attempt+1, c.retryAttempts+1, err)

			// Check if error is retryable
			if !c.isRetryableError(err) {
				c.logger.Printf("Non-retryable error, aborting retries: %v", err)
				return nil, fmt.Errorf("HTTP request failed: %w", err)
			}

			// Don't retry on last attempt
			if attempt < c.retryAttempts {
				delay := c.calculateRetryDelay(attempt)
				c.logger.Printf("Retrying in %v...", delay)

				// Wait with context cancellation support
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("request cancelled during retry wait: %w", ctx.Err())
				case <-time.After(delay):
					// Continue to next attempt
				}
			}
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			c.logger.Printf("Failed to read response body (attempt %d/%d): %v", attempt+1, c.retryAttempts+1, err)

			// Don't retry on last attempt
			if attempt < c.retryAttempts {
				delay := c.calculateRetryDelay(attempt)
				c.logger.Printf("Retrying in %v...", delay)

				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("request cancelled during retry wait: %w", ctx.Err())
				case <-time.After(delay):
					continue
				}
			}
			continue
		}

		// Create response headers map
		headers := make(map[string]string)
		for key, values := range resp.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}

		response := &HTTPResponse{
			StatusCode: resp.StatusCode,
			Body:       body,
			Headers:    headers,
		}

		c.logger.Printf("HTTP GET request successful: status=%d, body_size=%d", resp.StatusCode, len(body))

		// Check if we should retry based on status code
		if c.shouldRetryStatus(resp.StatusCode) {
			if attempt < c.retryAttempts {
				c.logger.Printf("Retryable status code %d, retrying...", resp.StatusCode)
				delay := c.calculateRetryDelay(attempt)

				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("request cancelled during retry wait: %w", ctx.Err())
				case <-time.After(delay):
					continue
				}
			} else {
				// Retries exhausted with retryable status code
				lastErr = fmt.Errorf("server returned retryable status %d", resp.StatusCode)
				continue
			}
		}

		return response, nil
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, fmt.Errorf("HTTP request failed after %d attempts: %w", c.retryAttempts+1, lastErr)
	}

	return nil, fmt.Errorf("HTTP request failed after %d attempts", c.retryAttempts+1)
}

// isRetryableError determines if an error is worth retrying
func (c *PartnerHTTPClient) isRetryableError(err error) bool {
	// Network errors are generally retryable
	if netErr, ok := err.(net.Error); ok {
		// Timeout errors are retryable
		if netErr.Timeout() {
			return true
		}
		// Temporary network errors are retryable
		if netErr.Temporary() {
			return true
		}
	}

	// Connection refused, DNS errors, etc. are retryable
	if _, ok := err.(*net.OpError); ok {
		return true
	}

	// Context cancellation is not retryable
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Default to not retryable for unknown errors
	return false
}

// shouldRetryStatus determines if an HTTP status code should trigger a retry
func (c *PartnerHTTPClient) shouldRetryStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusInternalServerError, // 500
		http.StatusBadGateway,                    // 502
		http.StatusServiceUnavailable,            // 503
		http.StatusGatewayTimeout,                // 504
		http.StatusInsufficientStorage,           // 507
		http.StatusNetworkAuthenticationRequired: // 511
		return true
	case http.StatusTooManyRequests: // 429
		return true
	default:
		return false
	}
}

// calculateRetryDelay calculates exponential backoff delay with jitter
func (c *PartnerHTTPClient) calculateRetryDelay(attempt int) time.Duration {
	// Exponential backoff: base_delay * 2^attempt
	baseDelay := c.retryDelay
	exponentialDelay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))

	// Cap the delay at 30 seconds
	maxDelay := 30 * time.Second
	if exponentialDelay > maxDelay {
		exponentialDelay = maxDelay
	}

	// Add jitter (±25% of the delay) to prevent thundering herd
	jitterRange := float64(exponentialDelay) * 0.25
	jitter := time.Duration(rand.Float64()*jitterRange*2 - jitterRange)

	finalDelay := exponentialDelay + jitter

	// Ensure minimum delay of 100ms
	minDelay := 100 * time.Millisecond
	if finalDelay < minDelay {
		finalDelay = minDelay
	}

	return finalDelay
}

// GetMetrics returns HTTP client metrics for monitoring
func (c *PartnerHTTPClient) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"retry_attempts": c.retryAttempts,
		"retry_delay":    c.retryDelay.String(),
		"timeout":        c.client.Timeout.String(),
	}
}
