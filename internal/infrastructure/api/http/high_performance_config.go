package http

import (
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// HighPerformanceConfig returns a Fiber configuration optimized for maximum throughput
func HighPerformanceConfig(logger *logrus.Logger) fiber.Config {
	return fiber.Config{
		AppName:      "Prayog Serviceability Service - High Performance",
		ServerHeader: "", // Remove server header for performance

		// Timeouts - Aggressive for high throughput
		ReadTimeout:  2 * time.Second,  // Very aggressive
		WriteTimeout: 2 * time.Second,  // Very aggressive
		IdleTimeout:  15 * time.Second, // Quick connection recycling

		// Connection limits
		Concurrency: 2 * 1024 * 1024, // 2 million concurrent connections

		// Buffer sizes - Optimized for high throughput
		ReadBufferSize:  32 * 1024, // 32KB read buffer
		WriteBufferSize: 32 * 1024, // 32KB write buffer

		// Body limits
		BodyLimit: 10 * 1024 * 1024, // 10MB max body size

		// Performance optimizations
		DisableStartupMessage:     true,  // No startup message
		DisableDefaultDate:        true,  // No date header
		DisableDefaultContentType: true,  // No default content type
		DisableHeaderNormalizing:  true,  // No header normalization
		DisableKeepalive:          false, // Keep connections alive
		ReduceMemoryUsage:         false, // Don't reduce memory

		// Network optimizations
		Network:                 "tcp",
		EnableTrustedProxyCheck: false, // Disable proxy check for performance
		TrustedProxies:          []string{},
		ProxyHeader:             "",

		// Disable features that add overhead
		EnablePrintRoutes:            false,
		DisablePreParseMultipartForm: true, // Disable multipart parsing unless needed

		// Compression
		CompressedFileSuffix: ".fiber.gz",

		// Custom error handler
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Minimal error response for performance
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
				"code":  code,
			})
		},

		// Use default JSON encoder/decoder for performance
		// JSONEncoder and JSONDecoder are left as default (nil)
	}
}

// GetOptimalWorkerCount returns the optimal number of workers based on system resources
func GetOptimalWorkerCount() int {
	numCPU := runtime.NumCPU()

	// For I/O intensive workloads, use more workers than CPU cores
	// For CPU intensive workloads, use fewer workers

	if numCPU <= 4 {
		return numCPU * 4 // 4x multiplier for small systems
	} else if numCPU <= 16 {
		return numCPU * 3 // 3x multiplier for medium systems
	} else {
		return numCPU * 2 // 2x multiplier for large systems
	}
}

// ConfigureGCForHighLoad configures Go garbage collector for high load scenarios
func ConfigureGCForHighLoad(logger *logrus.Logger) {
	// Set GOGC to a higher value to reduce GC frequency
	// This uses more memory but reduces GC pauses
	runtime.GC()

	// Log current GC settings
	logger.WithFields(logrus.Fields{
		"GOMAXPROCS":   runtime.GOMAXPROCS(0),
		"NumCPU":       runtime.NumCPU(),
		"NumGoroutine": runtime.NumGoroutine(),
	}).Info("Go runtime configuration for high load")
}

// OptimizeForHighThroughput applies runtime optimizations
func OptimizeForHighThroughput(logger *logrus.Logger) {
	// Set GOMAXPROCS to match CPU count if not already set
	if runtime.GOMAXPROCS(0) == 1 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// Configure GC
	ConfigureGCForHighLoad(logger)

	logger.Info("Applied high throughput optimizations")
}
