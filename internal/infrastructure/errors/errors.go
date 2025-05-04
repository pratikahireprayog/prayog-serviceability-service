package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with context.
type AppError struct {
	// HTTP status code
	StatusCode int `json:"-"`
	// Error code for client
	Code string `json:"code"`
	// User-friendly message
	Message string `json:"message"`
	// Internal error for logging
	Err error `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    message,
		Err:        err,
	}
}

// NewBadRequestError creates a new bad request error.
func NewBadRequestError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    message,
		Err:        err,
	}
}

// NewInternalServerError creates a new internal server error.
func NewInternalServerError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    message,
		Err:        err,
	}
}

// NewUnauthorizedError creates a new unauthorized error.
func NewUnauthorizedError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
		Err:        err,
	}
}

// NewForbiddenError creates a new forbidden error.
func NewForbiddenError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusForbidden,
		Code:       "FORBIDDEN",
		Message:    message,
		Err:        err,
	}
}
