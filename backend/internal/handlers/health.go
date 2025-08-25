package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) Ready(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := h.DB.PingContext(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "not ready",
			"error":  "database connection failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "ready",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}