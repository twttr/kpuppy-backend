package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSentryCapture(t *testing.T) {
	e := echo.New()
	middleware := SentryCapture()

	t.Run("passes through to next handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerCalled := false
		handler := func(c echo.Context) error {
			handlerCalled = true
			return c.String(http.StatusOK, "success")
		}

		err := middleware(handler)(c)

		assert.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("sets request ID tag when present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := context.WithValue(req.Context(), RequestIDKey, "test-request-id")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		}

		err := middleware(handler)(c)

		assert.NoError(t, err)
	})

	t.Run("sets user ID when X-User-ID header present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-ID", "user-123")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		}

		err := middleware(handler)(c)

		assert.NoError(t, err)
	})
}

func TestSentryRecover(t *testing.T) {
	e := echo.New()
	middleware := SentryRecover()

	t.Run("passes through when no panic", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		}

		err := middleware(handler)(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("recovers from panic with error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var capturedErr error
		e.HTTPErrorHandler = func(err error, c echo.Context) {
			capturedErr = err
		}

		handler := func(c echo.Context) error {
			panic(errors.New("test panic"))
		}

		middleware(handler)(c)

		assert.NotNil(t, capturedErr)
	})

	t.Run("recovers from panic with string", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var capturedErr error
		e.HTTPErrorHandler = func(err error, c echo.Context) {
			capturedErr = err
		}

		handler := func(c echo.Context) error {
			panic("string panic")
		}

		middleware(handler)(c)

		assert.NotNil(t, capturedErr)
	})
}

func TestCaptureError(t *testing.T) {
	e := echo.New()

	t.Run("captures error without hub in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		CaptureError(c, errors.New("test error"))
	})
}
