package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"service-golang/pkg/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ErrorHandlerMiddleware struct {
	log    *zap.Logger
	appEnv string
}

func NewErrorHandlerMiddleware(log *zap.Logger, appEnv string) *ErrorHandlerMiddleware {
	return &ErrorHandlerMiddleware{
		log:    log,
		appEnv: appEnv,
	}
}

// Handle is the error handler middleware
func (m *ErrorHandlerMiddleware) Handle() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		// Don't do anything if response already committed
		if c.Response().Committed {
			return
		}

		var (
			statusCode = http.StatusInternalServerError
			message    = "Internal Server Error"
			nameError  = ""
			errorsData interface{}
			stack      interface{}
		)

		// Check if it's our custom AppError
		if appErr, ok := err.(*response.AppError); ok {
			statusCode = appErr.Status
			message = appErr.Message
			nameError = appErr.Name
			errorsData = appErr.ErrorsData

			m.log.Error("Application Error",
				zap.Int("status", statusCode),
				zap.String("name", nameError),
				zap.String("message", message),
				zap.String("path", c.Request().URL.Path),
				zap.String("method", c.Request().Method),
				zap.String("ip", c.RealIP()),
				zap.Error(appErr.Err),
			)
		} else if echoErr, ok := err.(*echo.HTTPError); ok {
			// Echo's HTTP error
			statusCode = echoErr.Code
			message = fmt.Sprintf("%v", echoErr.Message)
			nameError = "Echo HTTP Error"

			m.log.Error("Echo HTTP Error",
				zap.Int("status", statusCode),
				zap.String("message", message),
				zap.String("path", c.Request().URL.Path),
				zap.String("method", c.Request().Method),
				zap.String("ip", c.RealIP()),
			)
		} else {
			// Unknown error
			message = err.Error()
			nameError = "Unknown Error"

			m.log.Error("Unknown Error",
				zap.String("error", message),
				zap.String("path", c.Request().URL.Path),
				zap.String("method", c.Request().Method),
				zap.String("ip", c.RealIP()),
			)
		}

		// Stack trace (only in development)
		if m.appEnv == "production" {
			stack = "🥞"
		} else {
			stack = string(debug.Stack())
			m.log.Error("Stack trace", zap.String("stack", fmt.Sprintf("%v", stack)))
		}

		// Build status text
		statusText := message
		if nameError != "" {
			statusText = fmt.Sprintf("%s | %s", message, nameError)
		}

		// Send error response
		errorResponse := response.ErrorResponse{
			Status:     statusCode,
			StatusCode: statusCode,
			StatusText: statusText,
			Message:    message,
			Errors:     errorsData,
			FetchDate:  time.Now().Format("2006-01-02 15:04:05"),
			SentryID:   nil, // Always nil since we're not using Sentry
		}

		// Log response
		m.log.Info("Error Response Sent",
			zap.Int("status", statusCode),
			zap.String("message", message),
			zap.String("path", c.Request().URL.Path),
		)

		c.JSON(statusCode, errorResponse)
	}
}
