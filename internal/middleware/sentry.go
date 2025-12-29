package middleware

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

func SentryCapture() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			hub := sentry.GetHubFromContext(c.Request().Context())
			if hub == nil {
				hub = sentry.CurrentHub().Clone()
			}

			hub.Scope().SetRequest(c.Request())

			if requestID := GetRequestID(c.Request().Context()); requestID != "" {
				hub.Scope().SetTag("request_id", requestID)
			}

			if userID := c.Request().Header.Get("X-User-ID"); userID != "" {
				hub.Scope().SetUser(sentry.User{ID: userID})
			}

			ctx := sentry.SetHubOnContext(c.Request().Context(), hub)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func SentryRecover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					hub := sentry.GetHubFromContext(c.Request().Context())
					if hub != nil {
						hub.RecoverWithContext(c.Request().Context(), r)
					} else {
						sentry.CurrentHub().RecoverWithContext(c.Request().Context(), r)
					}

					c.Error(echo.NewHTTPError(http.StatusInternalServerError, err.Error()))
				}
			}()

			return next(c)
		}
	}
}

func CaptureError(ctx echo.Context, err error) {
	hub := sentry.GetHubFromContext(ctx.Request().Context())
	if hub != nil {
		hub.CaptureException(err)
	} else {
		sentry.CaptureException(err)
	}
}
