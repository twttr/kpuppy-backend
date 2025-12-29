package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestAdminAuth_ValidCredentials(t *testing.T) {
	password := "testpassword"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	auth := NewAdminAuth("admin", hash)
	e := echo.New()

	handler := auth.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.SetBasicAuth("admin", password)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestAdminAuth_InvalidPassword(t *testing.T) {
	hash, err := HashPassword("correctpassword")
	require.NoError(t, err)

	auth := NewAdminAuth("admin", hash)
	e := echo.New()

	handler := auth.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.SetBasicAuth("admin", "wrongpassword")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAdminAuth_InvalidUsername(t *testing.T) {
	hash, err := HashPassword("testpassword")
	require.NoError(t, err)

	auth := NewAdminAuth("admin", hash)
	e := echo.New()

	handler := auth.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.SetBasicAuth("wronguser", "testpassword")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAdminAuth_MissingCredentials(t *testing.T) {
	hash, err := HashPassword("testpassword")
	require.NoError(t, err)

	auth := NewAdminAuth("admin", hash)
	e := echo.New()

	handler := auth.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Header().Get("WWW-Authenticate"), "Basic")
}

func TestAdminAuth_MalformedHeader(t *testing.T) {
	hash, err := HashPassword("testpassword")
	require.NoError(t, err)

	auth := NewAdminAuth("admin", hash)
	e := echo.New()

	handler := auth.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("malformed")))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
