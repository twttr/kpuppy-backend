package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type AdminAuth struct {
	Username     string
	PasswordHash string
}

func NewAdminAuth(username, passwordHash string) *AdminAuth {
	return &AdminAuth{
		Username:     username,
		PasswordHash: passwordHash,
	}
}

func (a *AdminAuth) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			username, password, ok := c.Request().BasicAuth()
			if !ok {
				return a.unauthorized(c)
			}

			if !a.validate(username, password) {
				return a.unauthorized(c)
			}

			return next(c)
		}
	}
}

func (a *AdminAuth) validate(username, password string) bool {
	if subtle.ConstantTimeCompare([]byte(username), []byte(a.Username)) != 1 {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password))
	return err == nil
}

func (a *AdminAuth) unauthorized(c echo.Context) error {
	c.Response().Header().Set("WWW-Authenticate", `Basic realm="Admin"`)
	return c.JSON(http.StatusUnauthorized, map[string]string{
		"error": "unauthorized",
		"code":  "UNAUTHORIZED",
	})
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
