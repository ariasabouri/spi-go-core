package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"spi-go-core/routes"
)

func main() {
	// Load server certificate and private key
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Fatalf("Failed to load server certificates: %v", err)
	}

	// Set up TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Register the routes dynamically (routes package handles this)
	router := routes.NewRouter()
	router.RegisterRoutes()

	// Define the HTTPS server
	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
		Handler:   nil, // Uses the default HTTP handler
	}

	// Start the HTTPS server
	log.Println("Starting secure server on https://localhost:8443")
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("Failed to start HTTPS server: %v", err)
	}
}
