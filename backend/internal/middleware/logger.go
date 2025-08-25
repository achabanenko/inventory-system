package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			
			log.Info().
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Msg("Request received")

			err := next(c)

			req := c.Request()
			res := c.Response()

			fields := map[string]interface{}{
				"method":     req.Method,
				"path":       req.URL.Path,
				"status":     res.Status,
				"latency_ms": time.Since(start).Milliseconds(),
				"ip":         c.RealIP(),
				"user_agent": req.UserAgent(),
			}

			if reqID := c.Request().Header.Get(echo.HeaderXRequestID); reqID != "" {
				fields["request_id"] = reqID
			}

			if err != nil {
				fields["error"] = err.Error()
				log.Error().Fields(fields).Msg("Request failed")
			} else {
				log.Info().Fields(fields).Msg("Request completed")
			}

			return err
		}
	}
}