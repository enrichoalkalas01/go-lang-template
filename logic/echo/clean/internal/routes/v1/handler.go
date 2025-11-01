package v1

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// setupBaseRoutes sets up base routes (root, health, 404)
func (r *Router) setupBaseRoutes() {
	r.app.GET("/", r.rootHandler)
	r.app.GET("/health", r.healthHandler)

	// 404 handler
	r.app.RouteNotFound("/*", r.notFoundHandler)
}

// rootHandler handles root route
func (r *Router) rootHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Server is running...",
		"version": r.config.GetString("APP_VERSION"),
		"time":    time.Now().Format(time.RFC3339),
	})
}

// healthHandler handles health check (simple version, detailed in health.go)
func (r *Router) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Service is healthy",
	})
}

// notFoundHandler handles 404 not found
func (r *Router) notFoundHandler(c echo.Context) error {
	return c.JSON(http.StatusNotFound, map[string]interface{}{
		"status":  http.StatusNotFound,
		"error":   "Not Found",
		"message": "The requested route does not exist",
		"path":    c.Request().URL.Path,
		"method":  c.Request().Method,
	})
}
