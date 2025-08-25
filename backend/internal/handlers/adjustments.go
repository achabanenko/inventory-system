package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) ListAdjustments(c echo.Context) error {
	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       []interface{}{},
		Page:       1,
		PageSize:   20,
		TotalPages: 0,
		Total:      0,
	})
}

func (h *Handler) CreateAdjustment(c echo.Context) error {
	return c.JSON(http.StatusCreated, map[string]string{
		"message": "adjustment created",
	})
}

func (h *Handler) GetAdjustment(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{
		"id": id,
	})
}

func (h *Handler) ApproveAdjustment(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{
		"id":      id,
		"message": "adjustment approved",
	})
}
