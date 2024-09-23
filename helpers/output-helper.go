package helpers

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents the structure of error responses
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// JSONError sends a JSON error response
func JSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		Message: message,
		Code:    statusCode,
	}

	json.NewEncoder(w).Encode(errorResponse)
}
