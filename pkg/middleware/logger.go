package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type LoggerMiddleware struct {
	log *zap.Logger
}

func NewLoggerMiddleware(log *zap.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{log: log}
}

func (m *LoggerMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Log request details
			m.log.Info("HTTP Request",
				zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.String("query", c.Request().URL.RawQuery),
				zap.String("ip", c.RealIP()),
				zap.String("user_agent", c.Request().UserAgent()),
				zap.Int("status", c.Response().Status),
				zap.Duration("latency", time.Since(start)),
				zap.Error(err),
			)

			return err
		}
	}
}
