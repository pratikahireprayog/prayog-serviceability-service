package middleware

import (
	tenantcontext "prayog-serviceability-service/internal/shared/context"

	"github.com/gofiber/fiber/v2"
)

// TenantMiddleware extracts tenant_id and user_id from request headers
// and stores them in the request context for use throughout the request lifecycle.
// Supports case-insensitive header extraction (X-Tenant-ID, x-tenant-id, tenantid, etc.)
func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		// Extract tenant_id from headers (case-insensitive - try multiple variations)
		tenantID := getHeaderCaseInsensitive(c, []string{
			"X-Tenant-ID",
			"x-tenant-id",
			"X-TENANT-ID",
			"tenantid",
			"TenantID",
			"TENANTID",
		})

		// Extract user_id from headers (case-insensitive - try multiple variations)
		userID := getHeaderCaseInsensitive(c, []string{
			"X-User-ID",
			"x-user-id",
			"X-USER-ID",
			"userid",
			"UserID",
			"USERID",
		})

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

// getHeaderCaseInsensitive tries multiple header name variations and returns the first non-empty value
func getHeaderCaseInsensitive(c *fiber.Ctx, headerNames []string) string {
	for _, name := range headerNames {
		if value := c.Get(name); value != "" {
			return value
		}
	}
	return ""
}


