package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/infrastructure/db"
	integrationServices "prayog-serviceability-service/internal/services/v1/integration"
	"prayog-serviceability-service/internal/shared/config"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	dbManager          *db.DatabaseManager
	integrationFactory *integrationServices.IntegrationFactory
	configManager      *config.ConfigManager
	logger             *logrus.Logger
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(
	dbManager *db.DatabaseManager,
	integrationFactory *integrationServices.IntegrationFactory,
	configManager *config.ConfigManager,
	logger *logrus.Logger,
) *HealthHandler {
	return &HealthHandler{
		dbManager:          dbManager,
		integrationFactory: integrationFactory,
		configManager:      configManager,
		logger:             logger,
	}
}

// HealthResponse represents the health check response structure
type HealthResponse struct {
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	Database    HealthCheckResult      `json:"database"`
	Integration map[string]interface{} `json:"integration"`
	Uptime      string                 `json:"uptime"`
}

// HealthCheckResult represents individual component health status
type HealthCheckResult struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Ping handles GET /ping (simple health check)
func (h *HealthHandler) Ping(c *fiber.Ctx) error {
	h.logger.Debug("Health ping check requested")

	return c.JSON(fiber.Map{
		"status":    "healthy",
		"message":   "Service is running",
		"timestamp": time.Now().UTC(),
		"service":   "prayog-serviceability-service",
	})
}

// Health handles GET /health (comprehensive health check)
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	h.logger.Debug("Comprehensive health check requested")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Initialize response
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Service:   "prayog-serviceability-service",
		Version:   "1.0.0",   // TODO: Get from version package
		Uptime:    "unknown", // TODO: Calculate uptime
	}

	// Check database health
	dbHealth := h.checkDatabaseHealth(ctx)
	response.Database = dbHealth

	// Check integration health
	integrationHealth := h.checkIntegrationHealth()
	response.Integration = integrationHealth

	// Determine overall status
	if dbHealth.Status != "healthy" {
		response.Status = "unhealthy"
	} else {
		// Check if any integration is unhealthy
		for _, status := range integrationHealth {
			if str, ok := status.(string); ok && str != "healthy" && !contains(str, "healthy") {
				response.Status = "degraded"
				break
			}
		}
	}

	// Return appropriate status code
	statusCode := fiber.StatusOK
	if response.Status == "unhealthy" {
		statusCode = fiber.StatusServiceUnavailable
	} else if response.Status == "degraded" {
		statusCode = fiber.StatusPartialContent
	}

	return c.Status(statusCode).JSON(response)
}

// Ready handles GET /ready (readiness probe for Kubernetes)
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	h.logger.Debug("Readiness check requested")

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// Check critical dependencies for readiness
	dbHealth := h.checkDatabaseHealth(ctx)

	if dbHealth.Status != "healthy" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":    "not ready",
			"message":   "Database not available",
			"timestamp": time.Now().UTC(),
		})
	}

	return c.JSON(fiber.Map{
		"status":    "ready",
		"message":   "Service is ready to accept traffic",
		"timestamp": time.Now().UTC(),
	})
}

// Live handles GET /live (liveness probe for Kubernetes)
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	h.logger.Debug("Liveness check requested")

	// Simple liveness check - if we can respond, we're alive
	return c.JSON(fiber.Map{
		"status":    "alive",
		"message":   "Service is alive",
		"timestamp": time.Now().UTC(),
	})
}

// checkDatabaseHealth performs database health checks
func (h *HealthHandler) checkDatabaseHealth(ctx context.Context) HealthCheckResult {
	if h.dbManager == nil {
		return HealthCheckResult{
			Status:  "unhealthy",
			Message: "Database manager not initialized",
		}
	}

	// Check database connectivity
	if err := h.dbManager.HealthCheck(ctx); err != nil {
		h.logger.WithError(err).Error("Database health check failed")
		return HealthCheckResult{
			Status:  "unhealthy",
			Message: "Database connectivity failed",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	// Get connection stats
	stats, err := h.dbManager.GetConnectionStats(ctx)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get database connection stats")
		stats = map[string]interface{}{
			"error": "Failed to retrieve stats",
		}
	}

	return HealthCheckResult{
		Status:  "healthy",
		Message: "Database is healthy",
		Details: stats,
	}
}

// checkIntegrationHealth checks the health of all integration services
func (h *HealthHandler) checkIntegrationHealth() map[string]interface{} {
	if h.integrationFactory == nil {
		return map[string]interface{}{
			"error": "Integration factory not initialized",
		}
	}

	// Get integration health status
	healthStatus := h.integrationFactory.GetIntegrationHealthStatus()

	// Convert map[string]string to map[string]interface{}
	result := make(map[string]interface{})
	for key, value := range healthStatus {
		result[key] = value
	}

	return result
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}
