package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = uuid.New().String()
			}

			req.Header.Set(echo.HeaderXRequestID, id)
			res.Header().Set(echo.HeaderXRequestID, id)

			return next(c)
		}
	}
}