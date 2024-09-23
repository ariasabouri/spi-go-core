// middlewares/handler-middleware.go

package middlewares

import (
	"log"
	"net/http"
	"spi-go-core/helpers"
	"spi-go-core/internal/encryption"
)

// ValidateConnection middleware checks if the connection is validated
func ValidateConnection(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the request ID from the headers
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" || !encryption.IsConnectionValidated(requestID) {
			log.Printf("Invalid or missing request ID, connection not validated. Request ID: %s", requestID)
			helpers.JSONError(w, "Connection is not validated", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
