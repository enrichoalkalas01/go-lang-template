package middleware

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ValidationMiddleware struct {
	log       *zap.Logger
	validator *validator.Validate
}

func NewValidationMiddleware(log *zap.Logger) *ValidationMiddleware {
	return &ValidationMiddleware{
		log:       log,
		validator: validator.New(),
	}
}

// ValidateRequest validates request body against struct tags
func (m *ValidationMiddleware) ValidateRequest(model interface{}) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Bind request to model
			if err := c.Bind(model); err != nil {
				m.log.Warn("Failed to bind request",
					zap.Error(err),
					zap.String("path", c.Request().URL.Path),
				)
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"error":   "Bad Request",
					"message": "Invalid request format",
				})
			}

			// Validate model
			if err := m.validator.Struct(model); err != nil {
				validationErrors := make(map[string]string)
				for _, err := range err.(validator.ValidationErrors) {
					validationErrors[err.Field()] = m.getErrorMessage(err)
				}

				m.log.Warn("Validation failed",
					zap.Any("errors", validationErrors),
					zap.String("path", c.Request().URL.Path),
				)

				return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
					"error":   "Validation Failed",
					"message": "Request validation failed",
					"details": validationErrors,
				})
			}

			// Set validated model to context
			c.Set("validated_data", model)

			return next(c)
		}
	}
}

func (m *ValidationMiddleware) getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "gte":
		return "Value must be greater than or equal to " + err.Param()
	case "lte":
		return "Value must be less than or equal to " + err.Param()
	default:
		return "Invalid value"
	}
}

// GetValidator returns the validator instance for custom validations
func (m *ValidationMiddleware) GetValidator() *validator.Validate {
	return m.validator
}
