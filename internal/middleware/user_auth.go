package middleware

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twttr/kpuppy-backend/internal/domain"
)

// userIDContextKey is the echo context key for the authenticated user ID.
const userIDContextKey = "authenticated_user_id"

// UserResolver is implemented by the user service.
type UserResolver interface {
	GetByUserHash(ctx context.Context, userHash string) (*domain.User, error)
}

// UserAuth returns middleware that authenticates the caller via X-User-Hash header.
// It resolves the hash to an internal user ID and stores it in the request context.
// This prevents clients from spoofing arbitrary user IDs via X-User-ID (#7).
func UserAuth(resolver UserResolver) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userHash := c.Request().Header.Get("X-User-Hash")
			if userHash == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "X-User-Hash header required",
					"code":  string(domain.CodeUserNotProvisioned),
				})
			}

			user, err := resolver.GetByUserHash(c.Request().Context(), userHash)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "user not found",
					"code":  string(domain.CodeUserNotProvisioned),
				})
			}

			if user.IsBanned {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "user is banned",
					"code":  string(domain.CodeUserBanned),
				})
			}

			// Store verified user ID in context — handlers must use GetUserID(), not the header.
			c.Set(userIDContextKey, user.ID)
			return next(c)
		}
	}
}

// GetUserID retrieves the authenticated user ID set by UserAuth middleware.
// Returns ("", false) if called outside of authenticated context.
func GetUserID(c echo.Context) (string, bool) {
	val := c.Get(userIDContextKey)
	if val == nil {
		return "", false
	}
	id, ok := val.(string)
	return id, ok && id != ""
}
