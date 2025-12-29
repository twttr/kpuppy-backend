package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRequestID_GeneratesNew(t *testing.T) {
	e := echo.New()
	e.Use(RequestID())

	var capturedID string
	e.GET("/test", func(c echo.Context) error {
		capturedID = GetRequestID(c.Request().Context())
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.NotEmpty(t, capturedID)
	assert.Equal(t, capturedID, rec.Header().Get(RequestIDHeader))
}

func TestRequestID_UsesExisting(t *testing.T) {
	e := echo.New()
	e.Use(RequestID())

	existingID := "existing-request-id-123"
	var capturedID string
	e.GET("/test", func(c echo.Context) error {
		capturedID = GetRequestID(c.Request().Context())
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, existingID)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, existingID, capturedID)
	assert.Equal(t, existingID, rec.Header().Get(RequestIDHeader))
}

func TestGetRequestID_EmptyContext(t *testing.T) {
	ctx := context.Background()
	id := GetRequestID(ctx)
	assert.Empty(t, id)
}

func TestGetRequestID_WithValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), RequestIDKey, "test-id")
	id := GetRequestID(ctx)
	assert.Equal(t, "test-id", id)
}
