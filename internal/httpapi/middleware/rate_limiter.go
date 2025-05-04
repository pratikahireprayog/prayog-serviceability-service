package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter is a simple rate limiting middleware for Fiber.
func RateLimiter() fiber.Handler {
	var mu sync.Mutex
	lastSeen := make(map[string]time.Time)

	return func(c *fiber.Ctx) error {
		// Get IP address for rate limiting
		ip := c.IP()

		mu.Lock()
		lastTime, exists := lastSeen[ip]
		now := time.Now()

		// Allow 10 requests per second per IP
		if exists && now.Sub(lastTime) < time.Second/10 {
			mu.Unlock()
			return c.Status(fiber.StatusTooManyRequests).SendString("Rate limit exceeded")
		}

		lastSeen[ip] = now
		mu.Unlock()

		return c.Next()
	}
}
