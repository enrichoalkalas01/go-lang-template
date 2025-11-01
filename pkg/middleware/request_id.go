package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RequestIDMiddleware struct{}

func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

func (m *RequestIDMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			c.Response().Header().Set(echo.HeaderXRequestID, requestID)
			c.Set("request_id", requestID)

			return next(c)
		}
	}
}
