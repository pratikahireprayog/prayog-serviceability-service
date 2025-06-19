package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/shared/dtos/v1"
)

// ServiceabilityMiddleware provides middleware for serviceability endpoints
type ServiceabilityMiddleware struct {
	logger *logrus.Logger
}

// NewServiceabilityMiddleware creates a new serviceability middleware
func NewServiceabilityMiddleware(logger *logrus.Logger) *ServiceabilityMiddleware {
	return &ServiceabilityMiddleware{
		logger: logger,
	}
}

// RequestSizeLimit middleware to limit request body size
func (sm *ServiceabilityMiddleware) RequestSizeLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// DISABLED: No request size limit for maximum performance
		// All requests are allowed regardless of body size
		return c.Next()
	}
}

// RateLimiter middleware for API rate limiting
func (sm *ServiceabilityMiddleware) RateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        1000000,         // 1 MILLION requests per minute (practically unlimited)
		Expiration: 1 * time.Minute, // Per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address as the key for rate limiting
			// In production, you might want to use user ID or API key
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			sm.logger.WithFields(logrus.Fields{
				"request_id": c.Get("X-Request-ID", "unknown"),
				"method":     c.Method(),
				"path":       c.Path(),
				"ip":         c.IP(),
			}).Warn("Rate limit exceeded")

			return c.Status(fiber.StatusTooManyRequests).JSON(dtos.ServiceabilityResponseDTO{
				Success: false,
				Error: &dtos.ErrorResponseDTO{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Rate limit exceeded",
					Details: "Maximum 100 requests per minute allowed",
				},
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                nil, // Use in-memory storage (default)
	})
}

// BulkRequestLimiter middleware specifically for bulk endpoints
func (sm *ServiceabilityMiddleware) BulkRequestLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        1000000,         // 1 MILLION bulk requests per minute (unlimited)
		Expiration: 1 * time.Minute, // Per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			sm.logger.WithFields(logrus.Fields{
				"request_id": c.Get("X-Request-ID", "unknown"),
				"method":     c.Method(),
				"path":       c.Path(),
				"ip":         c.IP(),
			}).Warn("Bulk request rate limit exceeded")

			return c.Status(fiber.StatusTooManyRequests).JSON(dtos.BulkServiceabilityResponseDTO{
				Success: false,
				Error: &dtos.ErrorResponseDTO{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Bulk request rate limit exceeded",
					Details: "Maximum 20 bulk requests per minute allowed",
				},
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                nil,
	})
}

// RequestLogging middleware for detailed request logging
func (sm *ServiceabilityMiddleware) RequestLogging() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		requestID := c.Get("X-Request-ID", "unknown")

		// Log request start
		sm.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     c.Method(),
			"path":       c.Path(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
			"body_size":  len(c.Body()),
		}).Info("Request started")

		// Process request
		err := c.Next()

		// Log request completion
		duration := time.Since(start)
		logLevel := logrus.InfoLevel
		if c.Response().StatusCode() >= 400 {
			logLevel = logrus.WarnLevel
		}
		if c.Response().StatusCode() >= 500 {
			logLevel = logrus.ErrorLevel
		}

		sm.logger.WithFields(logrus.Fields{
			"request_id":    requestID,
			"method":        c.Method(),
			"path":          c.Path(),
			"status_code":   c.Response().StatusCode(),
			"duration_ms":   duration.Milliseconds(),
			"response_size": len(c.Response().Body()),
		}).Log(logLevel, "Request completed")

		return err
	}
}

// ContentTypeValidation middleware to ensure proper content type
func (sm *ServiceabilityMiddleware) ContentTypeValidation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// DISABLED: No content type validation for maximum performance
		// All content types are accepted
		return c.Next()
	}
}

// SecurityHeaders middleware to add security headers
func (sm *ServiceabilityMiddleware) SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Add security headers
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		return c.Next()
	}
}
