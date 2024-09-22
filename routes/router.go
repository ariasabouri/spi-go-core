package routes

import (
	"log"
	"net/http"
	"spi-go-core/config"
	"spi-go-core/handlers"
)

type Router struct {
	Config *config.Config
}

// NewRouter creates a new instance of Router
func NewRouter() *Router {
	cfg, err := config.LoadConfig("routes.json")
	if err != nil {
		log.Fatalf("Failed to load route configuration: %v", err)
	}

	return &Router{
		Config: cfg,
	}
}

// RegisterRoutes registers all the routes from the configuration
func (r *Router) RegisterRoutes() {
	for _, route := range r.Config.Routes {
		handler := r.getHandler(route.Action)

		// Route based on method (GET, POST)
		switch route.Method {
		case "GET":
			http.HandleFunc(route.Path, handler)
		case "POST":
			http.HandleFunc(route.Path, handler)
		default:
			log.Printf("Unsupported HTTP method %s for path %s", route.Method, route.Path)
		}
	}
}

// getHandler returns the appropriate handler function for the action
func (r *Router) getHandler(action string) http.HandlerFunc {
	switch action {
	case "exec":
		return handlers.HandleExecCommand
	case "root":
		return handlers.HandleRoot
	default:
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Unknown action", http.StatusNotFound)
		}
	}
}
