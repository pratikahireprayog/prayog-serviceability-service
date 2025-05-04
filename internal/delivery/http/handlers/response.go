package handlers

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents a standard error response structure
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// SuccessResponse represents a standard success response structure
type SuccessResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

// RespondWithError sends a standardized JSON error response
func RespondWithError(w http.ResponseWriter, code int, message string, errorCode string) {
	response := ErrorResponse{
		Status:  "error",
		Message: message,
		Code:    errorCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RespondWithJSON sends a standardized JSON success response
func RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	response := SuccessResponse{
		Status: "success",
		Data:   data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to encode response", "encode_error")
	}
}

// RespondWithValidationError responds with a 400 Bad Request and validation details
func RespondWithValidationError(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusBadRequest, message, "validation_error")
}

// RespondWithNotFound responds with a 404 Not Found error
func RespondWithNotFound(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusNotFound, message, "not_found")
}

// RespondWithInternalError responds with a 500 Internal Server Error
func RespondWithInternalError(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusInternalServerError, message, "internal_error")
}
