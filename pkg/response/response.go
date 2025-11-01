package response

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Success Response Structure
type SuccessResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Total   *int64      `json:"total,omitempty"`
}

// Error Response Structure
type ErrorResponse struct {
	Status     int         `json:"status"`
	StatusCode int         `json:"statusCode"`
	StatusText string      `json:"statusText"`
	Message    string      `json:"message"`
	Errors     interface{} `json:"errors,omitempty"`
	FetchDate  string      `json:"fetchDate"`
	SentryID   *string     `json:"sentryId,omitempty"`
}

type ResponseHandler struct {
	log *zap.Logger
}

func NewResponseHandler(log *zap.Logger) *ResponseHandler {
	return &ResponseHandler{log: log}
}

// Success sends a successful response
func (r *ResponseHandler) Success(c echo.Context, status int, message string, data interface{}, total *int64) error {
	response := SuccessResponse{
		Status:  status,
		Message: message,
		Data:    data,
		Total:   total,
	}

	r.log.Info("Response Success",
		zap.Int("status", status),
		zap.String("message", message),
		zap.String("path", c.Request().URL.Path),
	)

	return c.JSON(status, response)
}

// Failed sends a failed response
func (r *ResponseHandler) Failed(c echo.Context, status int, message string, errors interface{}) error {
	response := ErrorResponse{
		Status:     status,
		StatusCode: status,
		StatusText: message,
		Message:    message,
		Errors:     errors,
		FetchDate:  time.Now().Format("2006-01-02 15:04:05"),
	}

	r.log.Warn("Response Failed",
		zap.Int("status", status),
		zap.String("message", message),
		zap.String("path", c.Request().URL.Path),
		zap.Any("errors", errors),
	)

	return c.JSON(status, response)
}

// Helper functions for common use cases

// OK sends 200 success response
func (r *ResponseHandler) OK(c echo.Context, message string, data interface{}) error {
	return r.Success(c, 200, message, data, nil)
}

// Created sends 201 created response
func (r *ResponseHandler) Created(c echo.Context, message string, data interface{}) error {
	return r.Success(c, 201, message, data, nil)
}

// SuccessWithTotal sends success response with total count
func (r *ResponseHandler) SuccessWithTotal(c echo.Context, message string, data interface{}, total int64) error {
	return r.Success(c, 200, message, data, &total)
}

// BadRequest sends 400 error response
func (r *ResponseHandler) BadRequest(c echo.Context, message string, errors interface{}) error {
	return r.Failed(c, 400, message, errors)
}

// Unauthorized sends 401 error response
func (r *ResponseHandler) Unauthorized(c echo.Context, message string) error {
	return r.Failed(c, 401, message, nil)
}

// Forbidden sends 403 error response
func (r *ResponseHandler) Forbidden(c echo.Context, message string) error {
	return r.Failed(c, 403, message, nil)
}

// NotFound sends 404 error response
func (r *ResponseHandler) NotFound(c echo.Context, message string) error {
	return r.Failed(c, 404, message, nil)
}

// InternalServerError sends 500 error response
func (r *ResponseHandler) InternalServerError(c echo.Context, message string, errors interface{}) error {
	return r.Failed(c, 500, message, errors)
}
