// middlewares/response-middleware.go

package middlewares

import (
	"encoding/json"
	"log"
	"net/http"
)

type ResponseWrapper struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (rw *ResponseWrapper) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *ResponseWrapper) Write(body []byte) (int, error) {
	rw.Body = body
	return rw.ResponseWriter.Write(body)
}

func OutputMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Wrap the ResponseWriter
		wrapper := &ResponseWrapper{ResponseWriter: w, StatusCode: http.StatusOK}

		// Call the next handler
		next(wrapper, r)

		// Process the response
		if wrapper.StatusCode >= 400 {
			// Handle error responses
			//jsonError(w, "An error occurred", wrapper.StatusCode)
		} else {
			// You can process successful responses if needed
			log.Printf("Status Code: %d, Response Body: %s", wrapper.StatusCode, wrapper.Body)
		}
	}
}

func jsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
