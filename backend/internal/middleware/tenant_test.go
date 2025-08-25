package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestTenantMiddleware(t *testing.T) {
	e := echo.New()

	// Test tenant ID from header
	t.Run("Tenant ID from header", func(t *testing.T) {
		tenantID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Tenant-ID", tenantID.String())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := TenantMiddleware()(func(c echo.Context) error {
			extractedTenantID, ok := GetTenantID(c.Request().Context())
			assert.True(t, ok)
			assert.Equal(t, tenantID, extractedTenantID)
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// Test missing tenant ID
	t.Run("Missing tenant ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := TenantMiddleware()(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})

	// Test invalid tenant ID format
	t.Run("Invalid tenant ID format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Tenant-ID", "invalid-uuid")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := TenantMiddleware()(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})
}

func TestRequireTenant(t *testing.T) {
	e := echo.New()

	// Test with valid tenant context
	t.Run("Valid tenant context", func(t *testing.T) {
		tenantID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := context.WithValue(req.Context(), TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := RequireTenant()(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// Test without tenant context
	t.Run("Missing tenant context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := RequireTenant()(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})
}

func TestGetTenantID(t *testing.T) {
	tenantID := uuid.New()

	// Test with valid tenant context
	t.Run("Valid tenant context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), TenantIDKey, tenantID)
		extractedTenantID, ok := GetTenantID(ctx)
		assert.True(t, ok)
		assert.Equal(t, tenantID, extractedTenantID)
	})

	// Test without tenant context
	t.Run("Missing tenant context", func(t *testing.T) {
		ctx := context.Background()
		_, ok := GetTenantID(ctx)
		assert.False(t, ok)
	})

	// Test with wrong type in context
	t.Run("Wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), TenantIDKey, "not-a-uuid")
		_, ok := GetTenantID(ctx)
		assert.False(t, ok)
	})
}

func TestSetTenantID(t *testing.T) {
	tenantID := uuid.New()
	ctx := SetTenantID(context.Background(), tenantID)

	extractedTenantID, ok := GetTenantID(ctx)
	assert.True(t, ok)
	assert.Equal(t, tenantID, extractedTenantID)
}
