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
