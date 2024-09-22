package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"spi-go-core/config"
)

type ExecHandler struct {
	Config *config.AppConfig
}

// Global variable for allowed commands
var allowedCommands = map[string]bool{}

// SetAllowedCommands updates the allowed commands based on config
func SetAllowedCommands(commands []string) {
	for _, cmd := range commands {
		allowedCommands[cmd] = true
	}
}

// HandleExecCommand handles the POST request to execute a command
func HandleExecCommand(w http.ResponseWriter, r *http.Request) {
	handler := ExecHandler{}
	// Read the request body
	encryptedCommand, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	// Decrypt the command
	var commandStr string
	// Check if encryption is enabled in the config
	if handler.Config.EncryptionConfig.Enabled {
		commandStr, err = decryptCommand(encryptedCommand)
		if err != nil {
			// Handle the error, return or log it as needed
			http.Error(w, "Failed to decrypt command", http.StatusBadRequest)
			return
		}
	} else {
		// If encryption is not enabled, use the original command
		commandStr = string(encryptedCommand)
	}
	if err != nil {
		http.Error(w, "Failed to decrypt command", http.StatusBadRequest)
		return
	}

	// Extract and validate the command
	cmd := extractCommand(commandStr)
	if !isWhitelistedCommand(cmd) {
		http.Error(w, fmt.Sprintf("Command '%s' is not allowed", cmd), http.StatusForbidden)
		return
	}

	// Execute the system command
	out, err := execCommand(cmd)
	if err != nil {
		http.Error(w, fmt.Sprintf("Command execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Send the output back
	w.Write([]byte(out))
}

// decryptCommand decrypts the encrypted command using the server's private key
func decryptCommand(encryptedCommand []byte) (string, error) {
	// Load the private key
	privateKey, err := loadPrivateKey("certs/private_key.pem")
	if err != nil {
		return "", err
	}

	// Decrypt the command using the private key
	decryptedBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedCommand)
	if err != nil {
		return "", err
	}

	return string(decryptedBytes), nil
}

// loadPrivateKey loads the private RSA key from a file
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decode the PEM block
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	// Parse the private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// extractCommand extracts the base command from the input (e.g., "ls -la" -> "ls")
func extractCommand(fullCommand string) string {
	var cmd string
	_, err := fmt.Sscanf(fullCommand, "%s", &cmd)
	if err != nil {
		return ""
	}
	return cmd
}

// isWhitelistedCommand checks if the given command is allowed
func isWhitelistedCommand(command string) bool {
	_, exists := allowedCommands[command]
	return exists
}

// execCommand executes the system command and returns the output or error
func execCommand(command string) (string, error) {
	// This is a placeholder. In reality, you'd execute the command and capture its output.
	return fmt.Sprintf("Executed: %s", command), nil
}
