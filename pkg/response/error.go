package response

import (
	"fmt"
	"net/http"
)

// AppError is custom application error
type AppError struct {
	Name       string      `json:"name"`
	Status     int         `json:"status"`
	Message    string      `json:"message"`
	ErrorsData interface{} `json:"errorsData,omitempty"`
	Err        error       `json:"-"` // Original error (not exposed in JSON)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (original: %v)", e.Name, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Name, e.Message)
}

// NewAppError creates a new application error
func NewAppError(name string, status int, message string, errorsData interface{}) *AppError {
	return &AppError{
		Name:       name,
		Status:     status,
		Message:    message,
		ErrorsData: errorsData,
	}
}

// Wrap wraps an existing error with AppError
func (e *AppError) Wrap(err error) *AppError {
	e.Err = err
	return e
}

// Common error constructors (like your Express.js throw patterns)

func ErrUnauthorized(message string) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return NewAppError("Error Auth", http.StatusUnauthorized, message, nil)
}

func ErrForbidden(message string) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return NewAppError("Error Forbidden", http.StatusForbidden, message, nil)
}

func ErrNotFound(message string) *AppError {
	if message == "" {
		message = "Resource not found"
	}
	return NewAppError("Error Not Found", http.StatusNotFound, message, nil)
}

func ErrBadRequest(message string, errorsData interface{}) *AppError {
	if message == "" {
		message = "Bad request"
	}
	return NewAppError("Error Validation", http.StatusBadRequest, message, errorsData)
}

func ErrValidation(message string, errorsData interface{}) *AppError {
	if message == "" {
		message = "Validation failed"
	}
	return NewAppError("Error Validation", http.StatusUnprocessableEntity, message, errorsData)
}

func ErrInternalServer(message string) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return NewAppError("Error Internal", http.StatusInternalServerError, message, nil)
}

func ErrConflict(message string) *AppError {
	if message == "" {
		message = "Resource conflict"
	}
	return NewAppError("Error Conflict", http.StatusConflict, message, nil)
}

func ErrServiceUnavailable(message string) *AppError {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	return NewAppError("Error Service", http.StatusServiceUnavailable, message, nil)
}
