package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CustomFilterMiddleware struct {
	log *zap.Logger
}

func NewCustomFilterMiddleware(log *zap.Logger) *CustomFilterMiddleware {
	return &CustomFilterMiddleware{log: log}
}

// Example: Block requests from specific user agents
func (m *CustomFilterMiddleware) BlockBadUserAgents() echo.MiddlewareFunc {
	blockedAgents := []string{"bot", "crawler", "spider"}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userAgent := strings.ToLower(c.Request().UserAgent())

			for _, blocked := range blockedAgents {
				if strings.Contains(userAgent, blocked) {
					m.log.Warn("Blocked bad user agent",
						zap.String("user_agent", userAgent),
						zap.String("ip", c.RealIP()),
					)
					return echo.NewHTTPError(403, "Forbidden")
				}
			}

			return next(c)
		}
	}
}

// Example: Add custom headers to response
func (m *CustomFilterMiddleware) AddCustomHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Custom-Header", "MyApp")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			return next(c)
		}
	}
}

// Example: Request body size limit
func (m *CustomFilterMiddleware) LimitRequestSize(maxSize int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().ContentLength > maxSize {
				m.log.Warn("Request body too large",
					zap.Int64("size", c.Request().ContentLength),
					zap.Int64("max_size", maxSize),
				)
				return echo.NewHTTPError(413, "Request body too large")
			}

			return next(c)
		}
	}
}

// Example: Sanitize input
func (m *CustomFilterMiddleware) SanitizeInput() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Sanitize query parameters
			for key, values := range c.QueryParams() {
				for i, value := range values {
					// Remove potentially dangerous characters
					sanitized := strings.ReplaceAll(value, "<", "")
					sanitized = strings.ReplaceAll(sanitized, ">", "")
					sanitized = strings.ReplaceAll(sanitized, "script", "")
					c.QueryParams()[key][i] = sanitized
				}
			}

			return next(c)
		}
	}
}
