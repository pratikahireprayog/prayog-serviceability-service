package dtos

// StandardResponse represents the base structure for all API responses
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// StandardListResponse represents the standardized structure for list responses with pagination
type StandardListResponse[T any] struct {
	Success    bool               `json:"success"`
	Message    string             `json:"message,omitempty"`
	Data       []T                `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
	Error      *ErrorInfo         `json:"error,omitempty"`
}

// StandardSingleResponse represents the standardized structure for single item responses
type StandardSingleResponse[T any] struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Data    T          `json:"data"`
	Error   *ErrorInfo `json:"error,omitempty"`
}

// ErrorInfo represents error details in responses
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// StandardErrorResponse represents the standardized structure for error responses
type StandardErrorResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Error   ErrorInfo `json:"error"`
}

// Helper functions for creating consistent error responses

// NewValidationErrorResponse creates a standard validation error response
func NewValidationErrorResponse(message string, details interface{}) StandardErrorResponse {
	return StandardErrorResponse{
		Success: false,
		Message: "Invalid request body",
		Error: ErrorInfo{
			Code:    "INVALID_REQUEST",
			Message: message,
			Details: details,
		},
	}
}

// NewNotFoundErrorResponse creates a standard not found error response
func NewNotFoundErrorResponse(message string) StandardErrorResponse {
	return StandardErrorResponse{
		Success: false,
		Message: "Resource not found",
		Error: ErrorInfo{
			Code:    "NOT_FOUND",
			Message: message,
		},
	}
}

// NewConflictErrorResponse creates a standard conflict error response
func NewConflictErrorResponse(message string) StandardErrorResponse {
	return StandardErrorResponse{
		Success: false,
		Message: "Resource conflict",
		Error: ErrorInfo{
			Code:    "CONFLICT",
			Message: message,
		},
	}
}

// NewInternalErrorResponse creates a standard internal server error response
func NewInternalErrorResponse(message string) StandardErrorResponse {
	return StandardErrorResponse{
		Success: false,
		Message: "Internal server error",
		Error: ErrorInfo{
			Code:    "INTERNAL_ERROR",
			Message: message,
		},
	}
}

// NewCustomErrorResponse creates a custom error response
func NewCustomErrorResponse(message, code, errorMessage string, details interface{}) StandardErrorResponse {
	return StandardErrorResponse{
		Success: false,
		Message: message,
		Error: ErrorInfo{
			Code:    code,
			Message: errorMessage,
			Details: details,
		},
	}
}
