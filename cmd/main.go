// In main.go
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	_ "embed"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"spi-go-core/handlers"
	"spi-go-core/internal/config"
	"spi-go-core/internal/ui"
	"spi-go-core/routes"
)

// go:embed embedded/typescript-app.tar.gz
var tsAppCompressed []byte

// Extracts the TypeScript project to a temporary directory
func extractTSApp() (string, error) {
	tempDir, err := os.MkdirTemp("", "typescript-app")
	if err != nil {
		return "", err
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(tsAppCompressed))
	if err != nil {
		return "", err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			break
		}

		filePath := filepath.Join(tempDir, header.Name)
		if header.Typeflag == tar.TypeDir {
			os.MkdirAll(filePath, 0755)
		} else if header.Typeflag == tar.TypeReg {
			file, err := os.Create(filePath)
			if err != nil {
				return "", err
			}
			_, err = ioutil.ReadAll(tarReader)
			file.Close()
		}
	}
	return tempDir, nil
}

// Runs the TypeScript project as a subprocess
func runTSApp(tsAppDir string) error {
	cmd := exec.Command("node", filepath.Join(tsAppDir, "dist/index.js"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

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
	} else {
		log.Printf("Loaded configuration from %s", configPath)
		log.Printf("Private key: %s", config.GlobalConfig.Encryption.PrivateKey)
		log.Printf("Public key: %s", config.GlobalConfig.Encryption.PublicKey)
	}

	// Start the UI when setting up the environment
	if cfg.UI.Enabled {
		log.Println("Starting UI..")
		ui.StartUI()
	}

	// Check if encryption is enabled or disabled
	if cfg.Encryption.Enabled {
		log.Println("Encryption is enabled.")
		// Add your encryption logic here
	} else {
		log.Println("Encryption is disabled for testing purposes.")
		// Skip encryption for testing purposes
	}

	// Set allowed commands in handlers
	handlers.SetAllowedCommands(cfg.Commands.Allowed)

	// Register the routes
	router := routes.NewRouter()
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

	// Log file for environment setup
	logFile, err := os.OpenFile("install.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()

	// Step 1: Install Node.js using EnvironmentSetupHandler
	log.Println("Starting environment setup...")
	err = handlers.EnvironmentSetupHandler(logFile)
	if err != nil {
		log.Fatalf("Environment setup failed: %v", err)
	}

	// Step 2: Extract the TypeScript app
	log.Println("Extracting TypeScript app...")
	tsAppDir, err := extractTSApp()
	if err != nil {
		log.Fatalf("Failed to extract TypeScript app: %v", err)
	}

	// Step 3: Run the TypeScript app as a subprocess
	log.Println("Running TypeScript app...")
	err = runTSApp(tsAppDir)
	if err != nil {
		log.Fatalf("Failed to run TypeScript app: %v", err)
	}

	log.Println("Process completed successfully.")
}
