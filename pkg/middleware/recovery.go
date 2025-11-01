package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type RcoveryMiddleware struct {
	log *zap.Logger
}

func NewRecoveryMiddleware(log *zap.Logger) *RcoveryMiddleware {
	return &RcoveryMiddleware{log: log}
}

func (m *RcoveryMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					// Log panic dengan stack trace
					m.log.Error("PANICE RECOVERED",
						zap.String("error", err.Error()),
						zap.String("stack", string(debug.Stack())),
						zap.String("method", c.Request().Method),
						zap.String("path", c.Request().URL.Path),
					)

					// Return error response
					c.JSON(http.StatusInternalServerError, map[string]interface{}{
						"error":   "Internal Server Error",
						"message": "Something went wrong",
					})
				}
			}()

			return next(c)
		}
	}
}
