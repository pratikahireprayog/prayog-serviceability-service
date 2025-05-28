package handlers

import (
	"prayog-serviceability-service/pkg/version"

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
		"data":   version.VersionInfo(),
	})
}
