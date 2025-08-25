package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) ListUsers(c echo.Context) error {
	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       []interface{}{},
		Page:       1,
		PageSize:   20,
		TotalPages: 0,
		Total:      0,
	})
}

func (h *Handler) CreateUser(c echo.Context) error {
	return c.JSON(http.StatusCreated, map[string]string{
		"message": "user created",
	})
}

func (h *Handler) GetUser(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{
		"id": id,
	})
}

func (h *Handler) UpdateUser(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{
		"id":      id,
		"message": "user updated",
	})
}

func (h *Handler) DisableUser(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{
		"id":      id,
		"message": "user disabled",
	})
}