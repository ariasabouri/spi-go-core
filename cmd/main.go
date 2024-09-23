// In main.go
package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"spi-go-core/config"
	"spi-go-core/handlers"
	"spi-go-core/routes"
)

func main() {
	// Load the configuration file
	// Get the absolute path of the root directory
	rootDir, err := os.Getwd() // Get the current working directory
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	// Check if config file exists
	configPath := filepath.Join(rootDir, "config.json") // Assuming it's in the root directory
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %v", err)
	} else {
		log.Println("Config file found")
	}

	cfg, err := config.LoadAppConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Check if encryption is enabled or disabled
	if cfg.EncryptionConfig.Enabled {
		log.Println("Encryption is enabled.")
		// Add your encryption logic here
	} else {
		log.Println("Encryption is disabled for testing purposes.")
		// Skip encryption for testing purposes
	}

	// Set allowed commands in handlers
	handlers.SetAllowedCommands(cfg.Commands.Allowed)

	// Register the routes
	router := routes.NewRouter(cfg)
	router.RegisterRoutes()

	// Server configuration based on TLS settings
	server := &http.Server{
		Addr:    ":8443",
		Handler: nil, // Uses the default HTTP handler
	}

	if cfg.Server.TLSEnabled {
		// Load server certificate and private key for TLS
		cert, err := tls.LoadX509KeyPair(cfg.Server.CertFile, cfg.Server.KeyFile)
		if err != nil {
			log.Fatalf("Failed to load server certificates: %v", err)
		}

		// Start HTTPS server with TLS
		server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
		log.Println("Starting secure server on https://localhost:", cfg.Server.Port)
		err = server.ListenAndServeTLS("", "")
	} else {
		// Start HTTP server without TLS
		log.Println("Starting server on http://localhost:", cfg.Server.Port)
		err = server.ListenAndServe()
	}

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
