package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(rate.Every(time.Second), 2)

	assert.True(t, rl.Allow("user1"))
	assert.True(t, rl.Allow("user1"))
	assert.False(t, rl.Allow("user1"))

	assert.True(t, rl.Allow("user2"))
}

func TestRateLimiter_DifferentUsers(t *testing.T) {
	rl := NewRateLimiter(rate.Every(time.Second), 1)

	assert.True(t, rl.Allow("user1"))
	assert.False(t, rl.Allow("user1"))

	assert.True(t, rl.Allow("user2"))
	assert.False(t, rl.Allow("user2"))

	assert.True(t, rl.Allow("user3"))
}

func TestRateLimiter_Middleware_AllowsRequest(t *testing.T) {
	e := echo.New()
	rl := NewRateLimiter(rate.Every(time.Second), 5)

	handler := rl.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestRateLimiter_Middleware_BlocksExcessiveRequests(t *testing.T) {
	e := echo.New()
	rl := NewRateLimiter(rate.Every(time.Minute), 2)

	handler := rl.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-ID", "test-user")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestRateLimiter_Middleware_UsesIPWhenNoUserID(t *testing.T) {
	e := echo.New()
	rl := NewRateLimiter(rate.Every(time.Minute), 1)

	handler := rl.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	err = handler(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
}

func TestMultiRateLimiter_DifferentLimitsPerRoute(t *testing.T) {
	mrl := NewMultiRateLimiter()
	mrl.AddLimiter("POST:/strict", rate.Every(time.Minute), 1)
	mrl.AddLimiter("POST:/relaxed", rate.Every(time.Minute), 5)

	e := echo.New()

	strictHandler := mrl.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "strict")
	})

	req := httptest.NewRequest(http.MethodPost, "/strict", nil)
	req.Header.Set("X-User-ID", "user1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/strict")
	err := strictHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	req2 := httptest.NewRequest(http.MethodPost, "/strict", nil)
	req2.Header.Set("X-User-ID", "user1")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.SetPath("/strict")
	err = strictHandler(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
}

func TestRateLimiter_BurstAllowed(t *testing.T) {
	rl := NewRateLimiter(rate.Every(time.Hour), 5)

	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow("burst-user"), "request %d should be allowed", i+1)
	}

	assert.False(t, rl.Allow("burst-user"), "6th request should be blocked")
}

func TestCreateRateLimiters(t *testing.T) {
	mrl := CreateRateLimiters()

	assert.NotNil(t, mrl)
	assert.NotNil(t, mrl.defaults)
	assert.NotEmpty(t, mrl.limiters)
}
