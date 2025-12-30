package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// LogoHandler handles logo management operations
type LogoHandler struct {
	logger    *logrus.Logger
	uploadDir string
}

// NewLogoHandler creates a new LogoHandler
func NewLogoHandler(logger *logrus.Logger) *LogoHandler {
	// ensure assets/logos exists
	path := "./assets/logos"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			logger.WithError(err).Error("Failed to create logos directory")
		}
	}
	
	return &LogoHandler{
		logger:    logger,
		uploadDir: path,
	}
}

// UploadLogo handles uploading a new logo
func (h *LogoHandler) UploadLogo(c *fiber.Ctx) error {
	file, err := c.FormFile("logo")
	if err != nil {
		h.logger.WithError(err).Error("Failed to get file from request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File upload failed",
		})
	}
	
	// Optional: Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".svg" && ext != ".webp" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file type. Allowed: png, jpg, jpeg, svg, webp",
		})
	}

	filename := filepath.Base(file.Filename)
	path := filepath.Join(h.uploadDir, filename)

	if err := c.SaveFile(file, path); err != nil {
		h.logger.WithError(err).Error("Failed to save logo")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	h.logger.WithField("filename", filename).Info("Logo uploaded successfully")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Logo uploaded successfully",
		"filename": filename,
		"url":      fmt.Sprintf("/logos/%s", filename),
	})
}

// UpdateLogo updates an existing logo
func (h *LogoHandler) UpdateLogo(c *fiber.Ctx) error {
	// effectively same as upload for file replacement
	return h.UploadLogo(c)
}

// DeleteLogo deletes a logo
func (h *LogoHandler) DeleteLogo(c *fiber.Ctx) error {
	filename := c.Params("filename")
	if filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Filename is required",
		})
	}

	// Sanitize filename
	filename = filepath.Base(filename)
	path := filepath.Join(h.uploadDir, filename)

	// Check if exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Logo not found",
		})
	}

	if err := os.Remove(path); err != nil {
		h.logger.WithError(err).Error("Failed to delete logo")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete logo",
		})
	}

	h.logger.WithField("filename", filename).Info("Logo deleted successfully")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logo deleted successfully",
	})
}

// ListLogos lists all available logos
func (h *LogoHandler) ListLogos(c *fiber.Ctx) error {
	entries, err := os.ReadDir(h.uploadDir)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list logos")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list logos",
		})
	}

	logos := make([]map[string]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			logos = append(logos, map[string]string{
				"filename": entry.Name(),
				"url":      fmt.Sprintf("/logos/%s", entry.Name()),
			})
		}
	}

	return c.JSON(fiber.Map{
		"logos": logos,
	})
}
