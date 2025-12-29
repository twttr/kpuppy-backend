package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func SlogLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			req := c.Request()
			res := c.Response()

			attrs := []slog.Attr{
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.Int("status", res.Status),
				slog.Duration("latency", time.Since(start)),
				slog.String("ip", c.RealIP()),
				slog.String("user_agent", req.UserAgent()),
			}

			if requestID := GetRequestID(req.Context()); requestID != "" {
				attrs = append(attrs, slog.String("request_id", requestID))
			}

			if userID := req.Header.Get("X-User-ID"); userID != "" {
				attrs = append(attrs, slog.String("user_id", userID))
			}

			if err != nil {
				attrs = append(attrs, slog.String("error", err.Error()))
				logger.LogAttrs(req.Context(), slog.LevelError, "request failed", attrs...)
			} else if res.Status >= 500 {
				logger.LogAttrs(req.Context(), slog.LevelError, "request", attrs...)
			} else if res.Status >= 400 {
				logger.LogAttrs(req.Context(), slog.LevelWarn, "request", attrs...)
			} else {
				logger.LogAttrs(req.Context(), slog.LevelInfo, "request", attrs...)
			}

			return err
		}
	}
}
