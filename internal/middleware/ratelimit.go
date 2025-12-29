package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twttr/kpuppy-backend/internal/domain"
	"golang.org/x/time/rate"
)

type RateLimiterConfig struct {
	Rate  rate.Limit
	Burst int
}

type RateLimiter struct {
	limiters map[string]*userLimiter
	mu       sync.RWMutex
	config   RateLimiterConfig
	cleanup  time.Duration
}

type userLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*userLimiter),
		config:   RateLimiterConfig{Rate: r, Burst: burst},
		cleanup:  time.Minute * 5,
	}

	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	for range ticker.C {
		rl.mu.Lock()
		for key, ul := range rl.limiters {
			if time.Since(ul.lastSeen) > rl.cleanup {
				delete(rl.limiters, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	ul, exists := rl.limiters[key]
	if !exists {
		ul = &userLimiter{
			limiter:  rate.NewLimiter(rl.config.Rate, rl.config.Burst),
			lastSeen: time.Now(),
		}
		rl.limiters[key] = ul
	}
	ul.lastSeen = time.Now()

	return ul.limiter
}

func (rl *RateLimiter) Allow(key string) bool {
	return rl.getLimiter(key).Allow()
}

func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := rl.getKey(c)
			limiter := rl.getLimiter(key)

			if !limiter.Allow() {
				return c.JSON(http.StatusTooManyRequests, domain.NewAPIError(domain.ErrRateLimitExceeded, domain.CodeRateLimitExceeded))
			}

			return next(c)
		}
	}
}

func (rl *RateLimiter) getKey(c echo.Context) string {
	if userID := c.Request().Header.Get("X-User-ID"); userID != "" {
		return "user:" + userID
	}
	return "ip:" + c.RealIP()
}

type MultiRateLimiter struct {
	limiters map[string]*RateLimiter
	defaults *RateLimiter
}

func NewMultiRateLimiter() *MultiRateLimiter {
	return &MultiRateLimiter{
		limiters: make(map[string]*RateLimiter),
		defaults: NewRateLimiter(rate.Every(time.Second), 20),
	}
}

func (mrl *MultiRateLimiter) AddLimiter(pattern string, r rate.Limit, burst int) {
	mrl.limiters[pattern] = NewRateLimiter(r, burst)
}

func (mrl *MultiRateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()
			method := c.Request().Method

			limiter := mrl.getLimiterForRoute(method, path)
			key := limiter.getKey(c)

			if !limiter.Allow(key) {
				return c.JSON(http.StatusTooManyRequests, domain.NewAPIError(domain.ErrRateLimitExceeded, domain.CodeRateLimitExceeded))
			}

			return next(c)
		}
	}
}

func (mrl *MultiRateLimiter) getLimiterForRoute(method, path string) *RateLimiter {
	key := method + ":" + path
	if limiter, ok := mrl.limiters[key]; ok {
		return limiter
	}
	return mrl.defaults
}

func CreateRateLimiters() *MultiRateLimiter {
	mrl := NewMultiRateLimiter()

	mrl.AddLimiter("POST:/users/provision", rate.Every(12*time.Second), 5)
	mrl.AddLimiter("POST:/content/:kinopubItemId/comments", rate.Every(6*time.Second), 3)
	mrl.AddLimiter("POST:/comments/:commentId/reply", rate.Every(6*time.Second), 3)
	mrl.AddLimiter("PATCH:/comments/:commentId", rate.Every(3*time.Second), 5)
	mrl.AddLimiter("DELETE:/comments/:commentId", rate.Every(3*time.Second), 5)

	return mrl
}
