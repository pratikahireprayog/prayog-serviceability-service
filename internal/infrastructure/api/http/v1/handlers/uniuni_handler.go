package handlers

import (
    "github.com/gofiber/fiber/v2"

    "prayog-serviceability-service/internal/services/v2/partners/common"
    "prayog-serviceability-service/internal/services/v2/partners/uniuni"
    "prayog-serviceability-service/internal/shared/config"
    models "prayog-serviceability-service/internal/shared/models/v1"
)

type UniUniHandler struct {
    cfg config.UniUniConfig
}

func NewUniUniHandler(cfg config.UniUniConfig) *UniUniHandler {
    return &UniUniHandler{cfg: cfg}
}

type uniuniPostalCodeRequest struct {
    PostalCode string `json:"postal_code"`
}

// CheckPostalCode calls UniUni adapter and returns the raw partner response
func (h *UniUniHandler) CheckPostalCode(c *fiber.Ctx) error {
    var req uniuniPostalCodeRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
    }
    if req.PostalCode == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "postal_code is required"})
    }

    adapter := uniuni.NewAdapter(h.cfg)

    pc := req.PostalCode
    v2req := &models.ServiceabilityV2Request{PostalCode: &pc}
    result, _ := adapter.CheckServiceability(c.Context(), v2req, common.PartnerInfo{PartnerCode: "uniuni"})
    if result != nil && result.PartnerServices != nil {
        return c.Status(fiber.StatusOK).JSON(result.PartnerServices)
    }

    if result != nil && result.ErrorMessage != nil {
        return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
            "success": false,
            "error":   *result.ErrorMessage,
            "metadata": result.Metadata,
            "caps":     result.Capabilities,
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": false})
}


