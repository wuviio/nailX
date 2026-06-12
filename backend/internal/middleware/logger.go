package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// StructuredLogger は slog を使った構造化リクエストログミドルウェア
func StructuredLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			req := c.Request()
			res := c.Response()

			status := res.Status
			if he, ok := err.(*echo.HTTPError); ok {
				status = he.Code
			}

			logger.InfoContext(req.Context(), "request",
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.Int("status", status),
				slog.Duration("latency", time.Since(start)),
				slog.String("ip", c.RealIP()),
				slog.String("user_id", c.Get("user_id").(string)),
			)
			return err
		}
	}
}
