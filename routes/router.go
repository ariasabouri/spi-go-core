// routes/router.go

package routes

import (
	"net/http"
	"spi-go-core/handlers"
	"spi-go-core/middlewares"
)

// Router holds the routing logic
type Router struct{}

// NewRouter creates a new Router
func NewRouter() *Router {
	return &Router{}
}

// RegisterRoutes registers routes directly with the necessary handlers and middleware
func (r *Router) RegisterRoutes() {
	// Handshake routes
	http.HandleFunc("/api/key-exchange", middlewares.OutputMiddleware(handlers.HandleKeyExchange))
	http.HandleFunc("/api/verify-message", middlewares.OutputMiddleware(middlewares.ValidateConnection(handlers.HandleMessageVerification)))
	http.HandleFunc("/api/handshake-success", middlewares.OutputMiddleware(middlewares.ValidateConnection(handlers.HandleSuccess)))

	// Protected routes (Require validated connection)
	http.HandleFunc("/api/exec", middlewares.OutputMiddleware(middlewares.ValidateConnection(handlers.HandleExecCommand)))

	// Root route
	http.HandleFunc("/", middlewares.OutputMiddleware(handlers.HandleRoot))
}
