package handlers

import (
	"inventory/internal/middleware"
	"inventory/internal/services"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CreateTenantRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
	Slug string `json:"slug" validate:"required,min=1,max=50"`
}

type UpdateTenantRequest struct {
	Name   string  `json:"name" validate:"max=100"`
	Slug   string  `json:"slug" validate:"max=50"`
	Domain *string `json:"domain" validate:"omitempty,fqdn"`
}

// ListTenants returns all active tenants (system admin only)
func (h *Handler) ListTenants(c echo.Context) error {
	tenantService := services.NewTenantService(h.DB)
	tenants, err := tenantService.ListTenants(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": tenants,
	})
}

// CreateTenant creates a new tenant (system admin only)
func (h *Handler) CreateTenant(c echo.Context) error {
	var req CreateTenantRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tenantService := services.NewTenantService(h.DB)
	tenant, err := tenantService.CreateTenant(c.Request().Context(), req.Name, req.Slug)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": tenant,
	})
}

// GetTenant returns a specific tenant
func (h *Handler) GetTenant(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID")
	}

	tenantService := services.NewTenantService(h.DB)
	tenant, err := tenantService.GetTenantByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": tenant,
	})
}

// UpdateTenant updates a tenant
func (h *Handler) UpdateTenant(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID")
	}

	var req UpdateTenantRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tenantService := services.NewTenantService(h.DB)
	tenant, err := tenantService.UpdateTenant(c.Request().Context(), id, req.Name, req.Slug, req.Domain)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": tenant,
	})
}

// DeactivateTenant deactivates a tenant
func (h *Handler) DeactivateTenant(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID")
	}

	tenantService := services.NewTenantService(h.DB)
	if err := tenantService.DeactivateTenant(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// GetCurrentTenant returns the current user's tenant information
func (h *Handler) GetCurrentTenant(c echo.Context) error {
	tenantID, ok := middleware.GetTenantID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "No tenant context")
	}

	tenantService := services.NewTenantService(h.DB)
	tenant, err := tenantService.GetTenantByID(c.Request().Context(), tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": tenant,
	})
}
