package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"prayog-serviceability-service/internal/shared/config"
)

// EnvHandler handles environment variable requests
type EnvHandler struct {
	logger *logrus.Logger
}

// NewEnvHandler creates a new environment variables handler
func NewEnvHandler(logger *logrus.Logger) *EnvHandler {
	return &EnvHandler{
		logger: logger,
	}
}

// GetEnvVars returns all environment variables
// GET /serviceability/debug/env
func (h *EnvHandler) GetEnvVars(c *fiber.Ctx) error {
	h.logger.WithFields(logrus.Fields{
		"ip":   c.IP(),
		"path": c.Path(),
	}).Info("Environment variables API accessed")

	envVars := config.GetAllEnvVars()

	return c.JSON(envVars)
}

