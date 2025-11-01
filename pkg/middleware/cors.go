package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type CORSMiddleware struct {
	allowOrigins []string
}

func NewCORSMiddleware(allowOrigins []string) *CORSMiddleware {
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"*"}
	}
	return &CORSMiddleware{allowOrigins: allowOrigins}
}

func (m *CORSMiddleware) Handle() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     m.allowOrigins,
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	})
}
