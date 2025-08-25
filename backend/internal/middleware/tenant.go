package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TenantContextKey string

const TenantIDKey TenantContextKey = "tenant_id"

// TenantMiddleware extracts tenant information from the request and adds it to context
func TenantMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID, err := extractTenantID(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing tenant identifier")
			}

			// Add tenant ID to request context
			ctx := context.WithValue(c.Request().Context(), TenantIDKey, tenantID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// extractTenantID extracts tenant ID from various sources (header, subdomain, path, etc.)
func extractTenantID(c echo.Context) (uuid.UUID, error) {
	// Method 1: From X-Tenant-ID header (recommended for API clients)
	tenantHeader := c.Request().Header.Get("X-Tenant-ID")
	if tenantHeader != "" {
		return uuid.Parse(tenantHeader)
	}

	// Method 2: From Authorization token claims (will be implemented later)
	// This requires the JWT middleware to run first and extract tenant from token

	// Method 3: From subdomain (e.g., tenant1.api.example.com)
	host := c.Request().Host
	if strings.Contains(host, ".") {
		subdomain := strings.Split(host, ".")[0]
		// This would require a lookup table from subdomain to tenant ID
		// For now, we'll skip this implementation
		_ = subdomain
	}

	// Method 4: From path parameter (e.g., /api/v1/tenants/{tenant_id}/items)
	tenantParam := c.Param("tenant_id")
	if tenantParam != "" {
		return uuid.Parse(tenantParam)
	}

	// Method 5: From query parameter (fallback, not recommended for production)
	tenantQuery := c.QueryParam("tenant_id")
	if tenantQuery != "" {
		return uuid.Parse(tenantQuery)
	}

	return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "Tenant identifier not found")
}

// GetTenantID retrieves the tenant ID from the request context
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	return tenantID, ok
}

// RequireTenant middleware ensures a valid tenant is present
func RequireTenant() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID, ok := GetTenantID(c.Request().Context())
			if !ok || tenantID == uuid.Nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Valid tenant identifier required")
			}
			return next(c)
		}
	}
}

// SetTenantID manually sets the tenant ID in context (useful for testing)
func SetTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}
