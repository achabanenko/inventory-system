package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetInventory(c echo.Context) error {
	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       []interface{}{},
		Page:       1,
		PageSize:   20,
		TotalPages: 0,
		Total:      0,
	})
}

func (h *Handler) GetItemLocations(c echo.Context) error {
	itemID := c.Param("item_id")
	return c.JSON(http.StatusOK, map[string]interface{}{
		"item_id":   itemID,
		"locations": []interface{}{},
	})
}

func (h *Handler) GetMovements(c echo.Context) error {
	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       []interface{}{},
		Page:       1,
		PageSize:   20,
		TotalPages: 0,
		Total:      0,
	})
}
