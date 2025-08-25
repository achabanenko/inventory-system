package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetAuditLogs(c echo.Context) error {
	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       []interface{}{},
		Page:       1,
		PageSize:   20,
		TotalPages: 0,
		Total:      0,
	})
}