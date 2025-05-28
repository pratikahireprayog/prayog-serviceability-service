package domain

import "fmt"

// ValidationError represents a domain validation error
type ValidationError struct {
	Message string
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		Message: message,
	}
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// NotFoundError represents a domain not found error
type NotFoundError struct {
	Entity  string
	ID      string
	Message string
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(entity, id string) *NotFoundError {
	return &NotFoundError{
		Entity:  entity,
		ID:      id,
		Message: fmt.Sprintf("%s with ID %s not found", entity, id),
	}
}

// Error implements the error interface
func (e *NotFoundError) Error() string {
	return e.Message
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// ConflictError represents a domain conflict error
type ConflictError struct {
	Entity  string
	ID      string
	Message string
}

// NewConflictError creates a new conflict error
func NewConflictError(entity, id, message string) *ConflictError {
	return &ConflictError{
		Entity:  entity,
		ID:      id,
		Message: message,
	}
}

// Error implements the error interface
func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict error: %s", e.Message)
}

// IsConflictError checks if an error is a conflict error
func IsConflictError(err error) bool {
	_, ok := err.(*ConflictError)
	return ok
}
