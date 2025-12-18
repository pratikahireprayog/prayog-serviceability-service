package middleware

import (
	tenantcontext "prayog-serviceability-service/internal/shared/context"

	"github.com/gofiber/fiber/v2"
)

// TenantMiddleware extracts tenant_id and user_id from request headers
// and stores them in the request context for use throughout the request lifecycle.
// Supports both standard headers (X-Tenant-ID, X-User-ID) and lowercase variants (tenantid, userid).
func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		// Extract tenant_id from headers (try both standard and lowercase variants)
		tenantID := c.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = c.Get("tenantid")
		}

		// Extract user_id from headers (try both standard and lowercase variants)
		userID := c.Get("X-User-ID")
		if userID == "" {
			userID = c.Get("userid")
		}

		// Store in context if present (optional - not required)
		if tenantID != "" {
			ctx = tenantcontext.WithTenantID(ctx, tenantID)
		}

		if userID != "" {
			ctx = tenantcontext.WithUserID(ctx, userID)
		}

		// Update the request context
		c.SetUserContext(ctx)

		// Continue to next handler
		return c.Next()
	}
}


