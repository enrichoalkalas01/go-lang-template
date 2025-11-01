package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	log       *zap.Logger
	jwtSecret string
}

type JWTClaims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func NewAuthMiddleware(log *zap.Logger, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		log:       log,
		jwtSecret: jwtSecret,
	}
}

func (m *AuthMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				m.log.Warn("Missing authorization header",
					zap.String("path", c.Request().URL.Path),
				)
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "Unauthorized",
					"message": "Missing authorization token",
				})
			}

			// Extract token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "Unauthorized",
					"message": "Invalid authorization format. Use: Bearer <token>",
				})
			}

			// Parse and validate token
			token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(m.jwtSecret), nil
			})

			if err != nil {
				m.log.Warn("Invalid JWT token",
					zap.Error(err),
					zap.String("path", c.Request().URL.Path),
				)
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "Unauthorized",
					"message": "Invalid or expired token",
				})
			}

			// Get claims
			claims, ok := token.Claims.(*JWTClaims)
			if !ok || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "Unauthorized",
					"message": "Invalid token claims",
				})
			}

			// Set user info to context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("roles", claims.Roles)

			m.log.Info("Authentication successful",
				zap.String("user_id", claims.UserID),
				zap.String("email", claims.Email),
			)

			return next(c)
		}
	}
}

// GenerateToken - Helper function to generate JWT token
func (m *AuthMiddleware) GenerateToken(userID, email string, roles []string) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.jwtSecret))
}
