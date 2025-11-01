package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type RBACMiddleware struct {
	log *zap.Logger
}

func NewRBACMiddleware(log *zap.Logger) *RBACMiddleware {
	return &RBACMiddleware{log: log}
}

// RequireRole checks if user has specific role
func (m *RBACMiddleware) RequireRole(requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user roles from context (set by auth middleware)
			userRoles, ok := c.Get("roles").([]string)
			if !ok {
				m.log.Warn("No roles found in context",
					zap.String("path", c.Request().URL.Path),
				)
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"error":   "Forbidden",
					"message": "Access denied",
				})
			}

			// Check if user has required role
			hasRole := false
			for _, requiredRole := range requiredRoles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				m.log.Warn("Insufficient permissions",
					zap.Strings("user_roles", userRoles),
					zap.Strings("required_roles", requiredRoles),
					zap.String("path", c.Request().URL.Path),
				)
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"error":   "Forbidden",
					"message": "You don't have permission to access this resource",
				})
			}

			m.log.Info("Authorization successful",
				zap.Strings("user_roles", userRoles),
				zap.Strings("required_roles", requiredRoles),
			)

			return next(c)
		}
	}
}

// RequirePermission checks if user has specific permission
func (m *RBACMiddleware) RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user permissions from context
			userPermissions, ok := c.Get("permissions").([]string)
			if !ok {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"error":   "Forbidden",
					"message": "Access denied",
				})
			}

			// Check if user has required permission
			hasPermission := false
			for _, userPermission := range userPermissions {
				if userPermission == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				m.log.Warn("Insufficient permissions",
					zap.Strings("user_permissions", userPermissions),
					zap.String("required_permission", permission),
					zap.String("path", c.Request().URL.Path),
				)
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"error":   "Forbidden",
					"message": "You don't have permission to access this resource",
				})
			}

			return next(c)
		}
	}
}
