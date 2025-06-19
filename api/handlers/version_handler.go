package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// VersionHandler handles version-related HTTP requests
type VersionHandler struct{}

// NewVersionHandler creates a new version handler
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// GetVersion returns the service version information
func (h *VersionHandler) GetVersion(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"version": "1.0.0",
			"service": "prayog-serviceability-service",
			"commit":  "unknown",
		},
	})
}
