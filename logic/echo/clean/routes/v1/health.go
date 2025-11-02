package v1

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// setupHealthRoutes sets up health check routes
func (r *Router) setupHealthRoutes(v1 *echo.Group) {
	health := v1.Group("/health")

	// Simple health check
	health.GET("", r.healthCheckHandler)
	health.GET("/", r.healthCheckHandler)

	// Detailed health check
	health.GET("/detail", r.healthCheckDetailHandler)
	health.GET("/detail/", r.healthCheckDetailHandler)

	// Readiness check
	health.GET("/ready", r.readinessCheckHandler)
	health.GET("/ready/", r.readinessCheckHandler)

	// Liveness check
	health.GET("/live", r.livenessCheckHandler)
	health.GET("/live/", r.livenessCheckHandler)
}

// healthCheckHandler - Simple health check
func (r *Router) healthCheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Service is healthy",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// healthCheckDetailHandler - Detailed health check
func (r *Router) healthCheckDetailHandler(c echo.Context) error {
	// TODO: Check database connection, redis, etc.

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Service is healthy",
		"version": r.config.GetString("APP_VERSION"),
		"uptime":  "calculated_uptime", // TODO: Calculate actual uptime
		"checks": map[string]interface{}{
			"database": "ok",
			"redis":    "ok",
			"storage":  "ok",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// readinessCheckHandler - Kubernetes readiness probe
func (r *Router) readinessCheckHandler(c echo.Context) error {
	// TODO: Check if service is ready to accept traffic
	// Check database connection, required services, etc.

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ready",
	})
}

// livenessCheckHandler - Kubernetes liveness probe
func (r *Router) livenessCheckHandler(c echo.Context) error {
	// Simple check if service is alive
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "alive",
	})
}
