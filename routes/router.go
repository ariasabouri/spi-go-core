package routes

import (
	"log"
	"net/http"
	"spi-go-core/handlers"
	"spi-go-core/internal/config"
)

// Router holds the configuration and routing logic
type Router struct {
	Config config.AppConfig
}

// NewRouter creates a new Router and loads the configuration
func NewRouter(cfg *config.AppConfig) *Router {
	return &Router{
		Config: *cfg,
	}
}

// RegisterRoutes registers all routes dynamically based on the configuration
func (r *Router) RegisterRoutes() {
	handlers.RegisterHandlers() // Register all handlers

	for _, route := range r.Config.Routes {
		handlerFunc, exists := handlers.FunctionMap[route.Action]
		if !exists {
			log.Printf("Handler function for action %s not found", route.Action)
			continue
		}

		// Register the route
		switch route.Method {
		case "GET":
			http.HandleFunc(route.Path, handlerFunc)
		case "POST":
			http.HandleFunc(route.Path, handlerFunc)
		default:
			log.Printf("Unsupported HTTP method %s for path %s", route.Method, route.Path)
		}
	}
}
