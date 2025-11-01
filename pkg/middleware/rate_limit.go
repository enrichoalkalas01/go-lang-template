package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type RateLimitMiddleware struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	log      *zap.Logger
	limit    rate.Limit
	burst    int
}

func NewRateLimitMiddleware(log *zap.Logger, requestsPerMinute int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiters: make(map[string]*rate.Limiter),
		log:      log,
		limit:    rate.Limit(float64(requestsPerMinute) / 60.0), // per second
		burst:    requestsPerMinute,
	}
}

func (m *RateLimitMiddleware) getLimiter(ip string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, exists := m.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(m.limit, m.burst)
		m.limiters[ip] = limiter
	}

	return limiter
}

func (m *RateLimitMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := m.getLimiter(ip)

			if !limiter.Allow() {
				m.log.Warn("Rate limit exceeded",
					zap.String("ip", ip),
					zap.String("path", c.Request().URL.Path),
				)

				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error":   "Rate limit exceeded",
					"message": "Too many requests, please try again later",
				})
			}

			return next(c)
		}
	}
}

// CleanupOldLimiters removes inactive limiters (call periodically)
func (m *RateLimitMiddleware) CleanupOldLimiters() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			m.mu.Lock()
			m.limiters = make(map[string]*rate.Limiter)
			m.mu.Unlock()
			m.log.Info("Rate limiter cache cleared")
		}
	}()
}
